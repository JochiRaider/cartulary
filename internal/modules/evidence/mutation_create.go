package evidence

// Evidence create orchestration.
import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (f *mutationFacade) Create(ctx context.Context, command CreateCommand) (MutationResult, error) {
	if !command.Admission.valid() || command.ActorUserID == uuid.Nil || command.IncidentID == uuid.Nil {
		return MutationResult{}, &ValidationError{Field: "payload", ReasonCode: "invalid_value"}
	}
	request := command.Admission.requestValue()
	requestHash := command.Admission.requestHash()
	idempotencyKey := IdempotencyKey{
		OperationID: OperationCreate,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.IncidentID.String() + ":" + request.ViewSchemaID,
		ClientTxnID: request.ClientTxnID,
	}
	if stored, found, err := f.replayStoredMutation(ctx, idempotencyKey, requestHash, StoredMutationCreate); err != nil {
		return MutationResult{}, err
	} else if found {
		return mutationResultFromStored(stored, request.ClientTxnID), nil
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

	if err := f.incidentAccess.RequireOpenTx(ctx, tx, command.IncidentID); err != nil {
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
		ActorUserID:   command.ActorUserID,
		ViewSchemaID:  request.ViewSchemaID,
		ClientTxnID:   request.ClientTxnID,
		RequestID:     command.RequestID,
		Source:        string(OperationCreate),
		MutationOrder: 1,
		Values:        request.Values,
		Now:           command.Now,
	}, createParams)
	if err != nil {
		if request.InitialObjectBlobID != nil && isEvidenceBlobUniqueViolation(err) {
			return MutationResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: errBlobNotAttachable}
		}
		return MutationResult{}, err
	}
	stored := NewStoredCreateResult(StoredMutationPayload{
		ViewSchemaID: request.ViewSchemaID, IncidentID: command.IncidentID,
		RecordID: result.recordID, RowVersion: 1, ChangeSetID: uuidPointer(result.changeSetID), Row: result.row,
	})
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, requestHash, stored); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit evidence create transaction: %w", err)
	}
	return MutationResult{
		Row:              cloneStringAnyMap(result.row),
		Outcome:          MutationOutcomeCreated,
		IncidentID:       command.IncidentID,
		RecordID:         result.recordID,
		ChangeSetID:      uuidPointer(result.changeSetID),
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: result.changedFieldKeys,
	}, nil
}

func mutationResultFromStored(stored StoredMutationPayload, clientTxnID string) MutationResult {
	return MutationResult{
		Row: cloneStringAnyMap(stored.Row), Outcome: MutationOutcomeReplayed,
		IncidentID: stored.IncidentID, RecordID: stored.RecordID, ChangeSetID: cloneUUIDPointer(stored.ChangeSetID),
		ClientTxnID: clientTxnID, RowVersion: stored.RowVersion, ViewSchemaID: stored.ViewSchemaID,
	}
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	result := value
	return &result
}

func cloneUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	return uuidPointer(*value)
}
