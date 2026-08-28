package tasksdecisions

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
)

func (f *MutationFacade) Create(ctx context.Context, command CreateCommand) (MutationResult, error) {
	request := command.Request
	idempotencyKey := IdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.IncidentID.String() + ":" + request.ViewSchemaID,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.idempotency.Get(ctx, idempotencyKey, command.RequestHash); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return MutationResult{}, ErrClientTxnConflict
		}
		if existing.Result.Kind() != StoredMutationCreate {
			return MutationResult{}, ErrStoredMutationKindMismatch
		}
		stored, ok := existing.Result.RowMutationResult()
		if !ok || stored.ViewSchemaID != request.ViewSchemaID {
			return MutationResult{}, ErrStoredMutationKindMismatch
		}
		return MutationResult{Row: stored.Row, Replayed: true, IncidentID: command.IncidentID, RecordID: stored.RecordID, ChangeSetID: stored.ChangeSetID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
		return MutationResult{}, fmt.Errorf("query task/decision create idempotency: %w", err)
	}
	if err := validateCreateRequest(request); err != nil {
		return MutationResult{}, err
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin task/decision create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := f.incidentAccess.RequireOpenTx(ctx, tx, command.IncidentID); err != nil {
		return MutationResult{}, err
	}
	if err := validateCreateReferencesTx(ctx, tx, f.catalog, f.linkStore, command.IncidentID, request); err != nil {
		return MutationResult{}, err
	}
	now := command.Now.UTC()
	surface, ok := f.catalog.SurfaceByViewID(request.ViewSchemaID)
	if !ok {
		return MutationResult{}, &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
	recordID, err := f.recordStore.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      command.IncidentID,
		RecordType:      surface.RecordType,
		CreatedByUserID: command.ActorUserID,
		CreatedAt:       now,
		UpdatedByUserID: command.ActorUserID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return MutationResult{}, err
	}
	directLinkMutations := make([]links.Mutation, 0)
	switch request.ViewSchemaID {
	case TaskRequestsViewSchemaID:
		if err := tasksource.InsertTaskRequestTx(ctx, tx, recordID, command.IncidentID, command.ActorUserID, policy.TaskCreateParams{Values: request.Values}, now); err != nil {
			return MutationResult{}, err
		}
		if value, ok := request.Values[taskDecisionRecordFieldKey]; ok && value.UUID != nil {
			directLinkMutations, err = syncTaskDecisionReferenceTx(ctx, tx, f.catalog, f.linkStore, command.IncidentID, recordID, command.ActorUserID, value.UUID, now)
			if err != nil {
				return MutationResult{}, err
			}
		}
	case DecisionsViewSchemaID:
		if err := tasksource.InsertDecisionTx(ctx, tx, recordID, command.IncidentID, command.ActorUserID, policy.DecisionCreateParams{Values: request.Values}, now); err != nil {
			return MutationResult{}, err
		}
	default:
		return MutationResult{}, &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
	collectionMutations, err := f.applyCollectionPayloadsTx(ctx, tx, request.ViewSchemaID, command.IncidentID, recordID, command.ActorUserID, request.Collections, now)
	if err != nil {
		return MutationResult{}, err
	}
	collectionMutations = append(directLinkMutations, collectionMutations...)
	afterSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := f.refreshRowTx(ctx, tx, request.ViewSchemaID, recordID); err != nil {
		return MutationResult{}, err
	}
	row, err := f.loadProjectionRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := f.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  command.IncidentID,
		ActorUserID: command.ActorUserID,
		Source:      command.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   now,
	})
	if err != nil {
		return MutationResult{}, err
	}
	afterVersionID := supersedeVersionID(recordID, 1)
	if err := f.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		RecordID:       recordID,
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterSnapshot:  &afterSnapshot,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := f.appendRecordLinkMutationsTx(ctx, tx, changeSetID, 2, collectionMutations); err != nil {
		return MutationResult{}, err
	}
	changedFields := changedFieldKeys(nil, row)
	if err := f.revisions.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
		ChangeSetID:   changeSetID,
		RecordID:      recordID,
		RowVersion:    1,
		AfterSnapshot: &afterSnapshot,
		ConflictFacts: taskDecisionRevisionFacts(nil, row, changedFields),
	}); err != nil {
		return MutationResult{}, err
	}
	if err := f.appendTaskDecisionRecordChangedTx(ctx, tx, command.IncidentID, command.ActorUserID, request.ClientTxnID, changeSetID, recordID, 1, 0, command.Now, request.ViewSchemaID, row, changedFields); err != nil {
		return MutationResult{}, err
	}
	storedResult := NewStoredCreateResult(StoredRowMutationResult{
		ViewSchemaID: request.ViewSchemaID,
		RecordID:     recordID,
		ChangeSetID:  changeSetID,
		Row:          row,
	})
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, command.RequestHash, storedResult); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit task/decision create transaction: %w", err)
	}
	return MutationResult{
		Row:              row,
		Created:          true,
		IncidentID:       command.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFields,
	}, nil
}

func validateCreateRequest(request CreateRequest) error {
	switch request.ViewSchemaID {
	case TaskRequestsViewSchemaID:
		return policy.ValidateTaskCreateParams(policy.TaskCreateParams{Values: request.Values})
	case DecisionsViewSchemaID:
		return policy.ValidateDecisionCreateParams(policy.DecisionCreateParams{Values: request.Values})
	default:
		return &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
}

func validateCreateReferencesTx(
	ctx context.Context,
	tx pgx.Tx,
	catalog *sourcecatalog.Catalog,
	linkStore LinkCapability,
	incidentID uuid.UUID,
	request CreateRequest,
) error {
	for fieldKey, value := range request.Values {
		if value.UUID != nil && isMemberUserReferenceField(catalog, fieldKey) {
			if err := validateIncidentMemberUserTx(ctx, tx, incidentID, *value.UUID, fieldKey); err != nil {
				return err
			}
		}
		if value.UUID != nil {
			if err := validateDirectReferenceTx(ctx, tx, catalog, incidentID, fieldKey, *value.UUID); err != nil {
				return err
			}
		}
	}
	for fieldKey, payload := range request.Collections {
		if err := validateCollectionPayloadTx(ctx, tx, linkStore, request.ViewSchemaID, incidentID, fieldKey, payload); err != nil {
			return err
		}
	}
	return nil
}
