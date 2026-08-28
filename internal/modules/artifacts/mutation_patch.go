package artifacts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

func (f *MutationFacade) Patch(ctx context.Context, command PatchCommand) (MutationResult, error) {
	return f.patch(ctx, command, OperationPatch)
}

func (f *MutationFacade) patch(ctx context.Context, command PatchCommand, operationID OperationID) (MutationResult, error) {
	if !command.Admission.valid() {
		return MutationResult{}, &ValidationError{Field: "payload", ReasonCode: "invalid_value"}
	}
	request := command.Admission.requestValue()
	requestHash := command.Admission.requestHash()
	if operationID != OperationPatch && operationID != OperationConflictResolve {
		return MutationResult{}, ErrStoredMutationKindMismatch
	}
	idempotencyKey := IdempotencyKey{
		OperationID: operationID,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.RecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	stored, replayed, err := f.replayStoredMutation(ctx, idempotencyKey, requestHash, "patch", storedMutationExpectation{
		kind: StoredMutationPatch, viewSchemaID: request.ViewSchemaID, recordID: &command.RecordID,
	})
	if err != nil {
		return MutationResult{}, err
	}
	if replayed {
		return MutationResult{
			Row: stored.Row, Outcome: MutationOutcomeReplayed, IncidentID: stored.IncidentID,
			RecordID: command.RecordID, ChangeSetID: cloneUUIDPointer(stored.ChangeSetID),
			ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID,
			RowVersion: stored.RowVersion,
		}, nil
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin artifact patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := f.loadArtifactRecordMetaForUpdateTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if meta.RecordType != "artifact" {
		return MutationResult{}, pgx.ErrNoRows
	}
	if err := validateArtifactViewRecordTx(ctx, tx, command.RecordID, request.ViewSchemaID); err != nil {
		return MutationResult{}, err
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
		window, err := conflicttokens.BuildCanonicalPatchConflictWindow(command.RecordID, request.BaseRowVersion, meta.RowVersion, windowRows, fieldDescriptors, f.conflictSnapshots)
		if err != nil {
			return MutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		if change, changed, ok := overlappingArtifactPatchChange(request.Changes, window.ChangedFields); ok {
			current, err := f.source.projections.LoadArtifactTx(ctx, tx, request.ViewSchemaID, command.RecordID)
			if err != nil {
				return MutationResult{}, err
			}
			conflictPayload, err := buildArtifactSameFieldConflict(artifactSameFieldConflictParams{
				RouteKey:          string(OperationConflictResolve),
				RecordID:          command.RecordID,
				ViewSchemaID:      request.ViewSchemaID,
				BaseRowVersion:    request.BaseRowVersion,
				CurrentRowVersion: meta.RowVersion,
				RequestHash:       requestHash,
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
	beforeRow, err := f.source.projections.LoadArtifactTx(ctx, tx, request.ViewSchemaID, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	beforeSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := validateArtifactPatchReferencesTx(ctx, tx, f.memberReferences, f.linkStore, meta.IncidentID, request); err != nil {
		return MutationResult{}, err
	}
	changed, collectionMutations, err := f.applyPatchTx(ctx, tx, meta.IncidentID, command.RecordID, command.ActorUserID, request, command.Now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if !changed {
		return MutationResult{}, &ValidationError{Field: "changes", ReasonCode: "no_effective_change"}
	}
	rowVersion, err := f.recordEnvelopes.AdvanceVersionTx(ctx, tx, command.RecordID, command.ActorUserID, command.Now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if err := f.source.rows.touchRowTx(ctx, tx, command.RecordID, command.Now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := f.source.projections.RefreshArtifactTx(ctx, tx, command.RecordID); err != nil {
		return MutationResult{}, err
	}
	afterRow, err := f.source.projections.LoadArtifactTx(ctx, tx, request.ViewSchemaID, command.RecordID)
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
		Source:      string(operationID),
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   command.Now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	beforeVersionID := workbookVersionID(command.RecordID, request.BaseRowVersion)
	if effectiveBeforeVersion != request.BaseRowVersion {
		beforeVersionID = workbookVersionID(command.RecordID, effectiveBeforeVersion)
	}
	afterVersionID := workbookVersionID(command.RecordID, rowVersion)
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
	if _, err := f.appendCollectionMutationsTx(ctx, tx, changeSetID, 2, collectionMutations); err != nil {
		return MutationResult{}, err
	}
	changedFields := changedFieldKeys(beforeRow, afterRow)
	if err := f.revisions.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
		ChangeSetID:    changeSetID,
		RecordID:       command.RecordID,
		RowVersion:     rowVersion,
		BeforeSnapshot: &beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		ConflictFacts:  artifactRevisionFacts(beforeRow, afterRow, changedFields),
	}); err != nil {
		return MutationResult{}, err
	}
	if err := f.appendRecordChangedTx(ctx, tx, meta.IncidentID, command.ActorUserID, request.ClientTxnID, changeSetID, command.RecordID, rowVersion, 0, command.Now, request.ViewSchemaID, afterRow, changedFields); err != nil {
		return MutationResult{}, err
	}
	storedResult := NewStoredPatchResult(StoredMutationPayload{
		ViewSchemaID: request.ViewSchemaID, IncidentID: meta.IncidentID, RecordID: command.RecordID,
		RowVersion: rowVersion, ChangeSetID: uuidPointer(changeSetID), Row: afterRow,
	})
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, requestHash, storedResult); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit artifact patch transaction: %w", err)
	}
	return MutationResult{
		Row:              afterRow,
		Outcome:          MutationOutcomeUpdated,
		IncidentID:       meta.IncidentID,
		RecordID:         command.RecordID,
		ChangeSetID:      uuidPointer(changeSetID),
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFields,
	}, nil
}

func (f *MutationFacade) applyPatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, request patchRequest, now time.Time) (bool, links.CollectionMutationResult, error) {
	changed := false
	mutations := links.CollectionMutationResult{}
	for _, change := range request.Changes {
		if change.Value != nil {
			policy, ok := lookupArtifactSourceField(change.FieldKey)
			if !ok || policy.ViewSchemaID != request.ViewSchemaID ||
				policy.Kind != sourcecatalog.FieldKindDirect || (!policy.View.Writable && !policy.View.CreateWritable) {
				return false, links.CollectionMutationResult{}, &ValidationError{Field: change.FieldKey, ReasonCode: "unsupported_field_key"}
			}
			if err := validateDirectPatchChange(change.FieldKey, *change.Value); err != nil {
				return false, links.CollectionMutationResult{}, err
			}
			applied, err := f.source.rows.applyDirectChangeTx(ctx, tx, recordID, change.FieldKey, *change.Value, now)
			if err != nil && strings.Contains(err.Error(), "unsupported field key") {
				return false, links.CollectionMutationResult{}, &ValidationError{Field: change.FieldKey, ReasonCode: "unsupported_field_key"}
			}
			if err != nil {
				return false, links.CollectionMutationResult{}, err
			}
			changed = changed || applied
			continue
		}
		if change.Collection != nil {
			applied, collectionResult, err := f.applyCollectionTx(ctx, tx, incidentID, recordID, actorID, request.ViewSchemaID, change.FieldKey, *change.Collection, now)
			if err != nil {
				return false, links.CollectionMutationResult{}, err
			}
			changed = changed || applied
			mutations.RecordLinks = append(mutations.RecordLinks, collectionResult.RecordLinks...)
			mutations.RecordTags = append(mutations.RecordTags, collectionResult.RecordTags...)
		}
	}
	if request.ViewSchemaID == FindingsViewSchemaID && touchesArtifactField(request.Changes, "finding.state") {
		applied, err := f.source.rows.normalizeFindingLifecycleTx(ctx, tx, recordID, now)
		if err != nil {
			return false, links.CollectionMutationResult{}, err
		}
		changed = changed || applied
	}
	return changed, mutations, nil
}
