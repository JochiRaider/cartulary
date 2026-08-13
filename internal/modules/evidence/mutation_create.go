package evidence

// Evidence create orchestration.
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (f *mutationFacade) Create(ctx context.Context, command CreateCommand) (MutationResult, error) {
	request := command.Request
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.Actor.ID,
		ScopeKey:    command.IncidentID.String() + ":" + request.ViewSchemaID,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed evidence create payload: %w", err)
		}
		recordID, err := extractPayloadUUID(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, IncidentID: command.IncidentID, RecordID: recordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query evidence create idempotency: %w", err)
	}
	createParams := createParams{
		Values:                 request.Values,
		InitialBlobWasSupplied: request.InitialObjectBlobID != nil,
	}
	if err := validateCreateParams(createParams); err != nil {
		return MutationResult{}, err
	}
	var observed *observedObject
	if request.InitialObjectBlobID != nil {
		var observeErr error
		observed, observeErr = f.observeInitialBlob(ctx, command.IncidentID, *request.InitialObjectBlobID)
		if observeErr != nil {
			return MutationResult{}, observeErr
		}
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin evidence create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, command.IncidentID); err != nil {
		return MutationResult{}, err
	}
	if request.InitialObjectBlobID != nil {
		initialBlob, commitRejection, finalizeErr := f.finalizeInitialBlobTx(
			ctx,
			tx,
			command.IncidentID,
			*request.InitialObjectBlobID,
			observed,
			command.Now.UTC(),
		)
		if finalizeErr != nil {
			if commitRejection {
				if err := tx.Commit(ctx); err != nil {
					return MutationResult{}, fmt.Errorf("commit rejected evidence blob finalization: %w", err)
				}
			}
			return MutationResult{}, finalizeErr
		}
		createParams.InitialBlob = initialBlob
		createParams.InitialBlobFinalized = true
	}
	if err := validateCreateParams(createParams); err != nil {
		return MutationResult{}, err
	}
	result, err := f.mutations.createTx(ctx, tx, evidenceCreateTxCommand{
		IncidentID:    command.IncidentID,
		ActorUserID:   command.Actor.ID,
		ViewSchemaID:  request.ViewSchemaID,
		ClientTxnID:   request.ClientTxnID,
		RequestID:     command.RequestID,
		Source:        command.RouteKey,
		MutationOrder: 1,
		Values:        request.Values,
		Now:           command.Now,
	}, createParams)
	if err != nil {
		if request.InitialObjectBlobID != nil && isEvidenceBlobUniqueViolation(err) {
			return MutationResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
		}
		return MutationResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, command.RequestHash, http.StatusCreated, result.payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit evidence create transaction: %w", err)
	}
	return MutationResult{
		Payload:          result.payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       command.IncidentID,
		RecordID:         result.recordID,
		ChangeSetID:      result.changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: result.changedFieldKeys,
	}, nil
}
