package tasksdecisions

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
)

func (f *MutationFacade) Patch(ctx context.Context, command PatchCommand) (MutationResult, error) {
	request := command.Request
	idempotencyKey := IdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.RecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.idempotency.Get(ctx, idempotencyKey, command.RequestHash); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return MutationResult{}, ErrClientTxnConflict
		}
		if existing.Result.Kind() != StoredMutationPatch {
			return MutationResult{}, ErrStoredMutationKindMismatch
		}
		stored, ok := existing.Result.RowMutationResult()
		if !ok || stored.ViewSchemaID != request.ViewSchemaID || stored.RecordID != command.RecordID {
			return MutationResult{}, ErrStoredMutationKindMismatch
		}
		return MutationResult{Row: stored.Row, Replayed: true, RecordID: command.RecordID, ChangeSetID: stored.ChangeSetID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
		return MutationResult{}, fmt.Errorf("query task/decision patch idempotency: %w", err)
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin task/decision patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadSupersedeRecordMetaForUpdateTx(ctx, tx, f.recordStore, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if !recordTypeMatchesView(f.catalog, meta.RecordType, request.ViewSchemaID) {
		return MutationResult{}, pgx.ErrNoRows
	}
	if err := f.incidentAccess.RequireOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return MutationResult{}, err
	}
	effectiveBeforeVersion := request.BaseRowVersion
	if meta.RowVersion != request.BaseRowVersion {
		if meta.RowVersion < request.BaseRowVersion {
			return MutationResult{}, &RowVersionConflictError{RecordID: command.RecordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
		}
		windowRows, err := f.revisions.LoadRevisionWindowTx(ctx, tx, command.RecordID, request.BaseRowVersion, meta.RowVersion)
		if err != nil {
			return MutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		fieldDescriptors, err := f.conflictFields.ResolveViewSchema(request.ViewSchemaID)
		if err != nil {
			return MutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		projector, ok := taskDecisionConflictSnapshotProjector(f.catalog, request.ViewSchemaID)
		if !ok {
			return MutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, conflicttokens.ErrInvalidSnapshotProjector)
		}
		window, err := conflicttokens.BuildCanonicalPatchConflictWindow(command.RecordID, request.BaseRowVersion, meta.RowVersion, windowRows, fieldDescriptors, projector)
		if err != nil {
			return MutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		if change, changed, ok := overlappingPatchChange(request.Changes, window.ChangedFields); ok {
			current, err := f.loadProjectionRowTx(ctx, tx, request.ViewSchemaID, command.RecordID)
			if err != nil {
				return MutationResult{}, err
			}
			conflictPayload, err := buildSameFieldConflict(sameFieldConflictParams{
				RouteKey:          command.ConflictRouteKey,
				RecordID:          command.RecordID,
				ViewSchemaID:      request.ViewSchemaID,
				BaseRowVersion:    request.BaseRowVersion,
				CurrentRowVersion: meta.RowVersion,
				RequestHash:       command.RequestHash,
				Window:            window,
				Change:            change,
				Changed:           changed,
				CurrentRow:        current,
				FieldDescriptors:  fieldDescriptors,
				Codec:             f.conflictTokens,
			})
			if err != nil {
				return MutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
			}
			return MutationResult{}, &SameFieldConflictError{Conflict: conflictPayload}
		}
		effectiveBeforeVersion = meta.RowVersion
	}
	beforeRow, err := f.loadProjectionRowTx(ctx, tx, request.ViewSchemaID, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	beforeSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := validatePatchReferencesTx(ctx, tx, f.catalog, f.linkStore, meta.IncidentID, request); err != nil {
		return MutationResult{}, err
	}
	changed, collectionMutations, err := f.applyPatchTx(ctx, tx, meta.IncidentID, command.RecordID, command.ActorUserID, request, command.Now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if !changed {
		return MutationResult{}, &ValidationError{Field: "changes", ReasonCode: "no_effective_change"}
	}
	rowVersion, err := f.recordStore.AdvanceVersionTx(ctx, tx, command.RecordID, command.ActorUserID, command.Now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if err := f.touchSourceRowTx(ctx, tx, request.ViewSchemaID, command.RecordID, command.Now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := f.refreshRowTx(ctx, tx, request.ViewSchemaID, command.RecordID); err != nil {
		return MutationResult{}, err
	}
	afterRow, err := f.loadProjectionRowTx(ctx, tx, request.ViewSchemaID, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	afterSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := f.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  meta.IncidentID,
		ActorUserID: command.ActorUserID,
		Source:      command.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   command.Now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	beforeVersionID := supersedeVersionID(command.RecordID, request.BaseRowVersion)
	if effectiveBeforeVersion != request.BaseRowVersion {
		beforeVersionID = supersedeVersionID(command.RecordID, effectiveBeforeVersion)
	}
	afterVersionID := supersedeVersionID(command.RecordID, rowVersion)
	if err := f.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		RecordID:        command.RecordID,
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  &beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := f.appendRecordLinkMutationsTx(ctx, tx, changeSetID, 2, collectionMutations); err != nil {
		return MutationResult{}, err
	}
	changedFields := changedFieldKeys(beforeRow, afterRow)
	if err := f.revisions.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
		ChangeSetID:    changeSetID,
		RecordID:       command.RecordID,
		RowVersion:     rowVersion,
		BeforeSnapshot: &beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		ConflictFacts:  taskDecisionRevisionFacts(beforeRow, afterRow, changedFields),
	}); err != nil {
		return MutationResult{}, err
	}
	if err := f.appendTaskDecisionRecordChangedTx(ctx, tx, meta.IncidentID, command.ActorUserID, request.ClientTxnID, changeSetID, command.RecordID, rowVersion, 0, command.Now, request.ViewSchemaID, afterRow, changedFields); err != nil {
		return MutationResult{}, err
	}
	storedResult := NewStoredPatchResult(StoredRowMutationResult{
		ViewSchemaID: request.ViewSchemaID,
		RecordID:     command.RecordID,
		ChangeSetID:  changeSetID,
		Row:          afterRow,
	})
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, command.RequestHash, storedResult); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit task/decision patch transaction: %w", err)
	}
	return MutationResult{
		Row:              afterRow,
		IncidentID:       meta.IncidentID,
		RecordID:         command.RecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFields,
	}, nil
}

func validatePatchReferencesTx(
	ctx context.Context,
	tx pgx.Tx,
	catalog *sourcecatalog.Catalog,
	linkStore LinkCapability,
	incidentID uuid.UUID,
	request PatchRequest,
) error {
	for _, change := range request.Changes {
		if change.Value != nil && change.Value.UUID != nil && isMemberUserReferenceField(catalog, change.FieldKey) {
			if err := validateIncidentMemberUserTx(ctx, tx, incidentID, *change.Value.UUID, change.FieldKey); err != nil {
				return err
			}
		}
		if change.Value != nil && change.Value.UUID != nil {
			if err := validateDirectReferenceTx(ctx, tx, catalog, incidentID, change.FieldKey, *change.Value.UUID); err != nil {
				return err
			}
		}
		if change.Collection != nil {
			if err := validateCollectionPayloadTx(ctx, tx, linkStore, request.ViewSchemaID, incidentID, change.FieldKey, *change.Collection); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *MutationFacade) applyPatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, request PatchRequest, now time.Time) (bool, []links.RecordLinkMutation, error) {
	changed := false
	collectionMutations := make([]links.RecordLinkMutation, 0)
	var beforeTask policy.TaskLifecycleState
	var beforeDecisionStatus string
	var err error
	if request.ViewSchemaID == TaskRequestsViewSchemaID && touchesAnyField(request.Changes, "task.status", "task.blocked_reason", "task.completed_at", "task.owner_user_id") {
		beforeTask, err = tasksource.LoadTaskLifecycleStateTx(ctx, tx, recordID)
		if err != nil {
			return false, nil, err
		}
	}
	if request.ViewSchemaID == DecisionsViewSchemaID {
		if err := validateDecisionMachineConsistentTx(ctx, tx, recordID); err != nil {
			return false, nil, err
		}
		if touchesField(request.Changes, "decision.status") {
			beforeDecisionStatus, err = tasksource.LoadDecisionStatusTx(ctx, tx, recordID)
			if err != nil {
				return false, nil, err
			}
		}
	}
	for _, change := range request.Changes {
		if change.Value != nil {
			applied, mutations, err := f.applyDirectChangeTx(ctx, tx, incidentID, recordID, actorID, request.ViewSchemaID, change, now)
			if err != nil {
				return false, nil, err
			}
			changed = changed || applied
			collectionMutations = append(collectionMutations, mutations...)
			continue
		}
		if change.Collection != nil {
			applied, mutations, err := f.applyCollectionPayloadTx(ctx, tx, request.ViewSchemaID, incidentID, recordID, actorID, change.FieldKey, *change.Collection, now)
			if err != nil {
				return false, nil, err
			}
			changed = changed || applied
			collectionMutations = append(collectionMutations, mutations...)
		}
	}
	if request.ViewSchemaID == TaskRequestsViewSchemaID && touchesAnyField(request.Changes, "task.status", "task.blocked_reason", "task.completed_at", "task.owner_user_id") {
		applied, err := tasksource.NormalizeTaskLifecycleTx(ctx, tx, recordID, beforeTask, touchesField(request.Changes, "task.completed_at"), now)
		if err != nil {
			return false, nil, err
		}
		changed = changed || applied
	}
	if request.ViewSchemaID == DecisionsViewSchemaID && touchesField(request.Changes, "decision.status") {
		afterDecisionStatus, err := tasksource.LoadDecisionStatusTx(ctx, tx, recordID)
		if err != nil {
			return false, nil, err
		}
		if err := policy.ValidateDecisionStatusTransition(beforeDecisionStatus, afterDecisionStatus); err != nil {
			return false, nil, err
		}
		if err := validateDecisionMachineConsistentTx(ctx, tx, recordID); err != nil {
			return false, nil, err
		}
	}
	return changed, collectionMutations, nil
}

func (f *MutationFacade) applyDirectChangeTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, viewSchemaID string, change PatchChange, now time.Time) (bool, []links.RecordLinkMutation, error) {
	if err := validateDirectPatchChange(f.catalog, viewSchemaID, change.FieldKey, *change.Value); err != nil {
		return false, nil, err
	}
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		return applyTaskDirectChangeTx(ctx, tx, f.catalog, f.linkStore, incidentID, recordID, actorID, change.FieldKey, *change.Value, now)
	case DecisionsViewSchemaID:
		changed, err := tasksource.ApplyDecisionDirectChangeTx(ctx, tx, f.catalog, recordID, change.FieldKey, *change.Value, now)
		return changed, nil, err
	default:
		return false, nil, &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
}

func validateDirectPatchChange(catalog *sourcecatalog.Catalog, viewSchemaID string, fieldKey string, value FieldValue) error {
	field, ok := catalog.Field(fieldKey)
	if !ok || field.ViewSchemaID != viewSchemaID || field.Kind != sourcecatalog.FieldKindDirect {
		return &ValidationError{Field: fieldKey, ReasonCode: "unsupported_field_key"}
	}
	if field.View.ReadKind == "enum" && value.Text != nil && !slices.Contains(field.View.EnumValues, *value.Text) {
		return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
	}
	return nil
}
