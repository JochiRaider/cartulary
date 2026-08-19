package evidence

// Evidence patch and idempotent replay orchestration.
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (f *mutationFacade) Patch(ctx context.Context, command PatchCommand) (MutationResult, error) {
	request := command.Request
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.Actor.ID,
		ScopeKey:    command.RecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed evidence patch payload: %w", err)
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: command.RecordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query evidence patch idempotency: %w", err)
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin evidence patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadEvidenceRecordMetaForUpdateTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if meta.RecordType != "evidence" {
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
		windowRows, err := f.revisionHistory.LoadRevisionWindowTx(ctx, tx, command.RecordID, request.BaseRowVersion, meta.RowVersion)
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
		if change, changed, ok := overlappingEvidencePatchChange(request.Changes, window.ChangedFields); ok {
			current, err := f.projectionRows.LoadEvidenceTx(ctx, tx, command.RecordID)
			if err != nil {
				return MutationResult{}, err
			}
			conflictPayload, err := buildEvidenceSameFieldConflict(evidenceSameFieldConflictParams{
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
	beforeRow, err := f.projectionRows.LoadEvidenceTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	beforeSnapshot, err := f.revisionAppender.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := validateEvidencePatchReferencesTx(ctx, tx, meta.IncidentID, request); err != nil {
		return MutationResult{}, err
	}
	if err := f.validateLifecyclePatchTx(ctx, tx, command.RecordID, request); err != nil {
		return MutationResult{}, err
	}
	changed, err := f.applyPatchTx(ctx, tx, command.RecordID, request, command.Now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if !changed {
		return MutationResult{}, &ValidationError{Field: "changes", ReasonCode: "no_effective_change"}
	}
	rowVersion, err := f.recordStore.AdvanceVersionTx(ctx, tx, command.RecordID, command.Actor.ID, command.Now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if err := f.sourceMutations.touchRowTx(ctx, tx, command.RecordID, command.Now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := f.projectionRows.RefreshEvidenceTx(ctx, tx, command.RecordID); err != nil {
		return MutationResult{}, err
	}
	afterRow, err := f.projectionRows.LoadEvidenceTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	afterSnapshot, err := f.revisionAppender.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := f.revisionAppender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  meta.IncidentID,
		ActorUserID: command.Actor.ID,
		Source:      command.RouteKey,
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
	if err := f.revisionAppender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
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
	if err := f.revisionAppender.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID:    changeSetID,
		RecordID:       command.RecordID,
		RowVersion:     rowVersion,
		BeforeSnapshot: &beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		LiveChange:     revisions.LiveRecordChange{BeforeValue: beforeRow, AfterValue: afterRow},
	}); err != nil {
		return MutationResult{}, err
	}
	patchChangedFieldKeys := changedFieldKeys(beforeRow, afterRow)
	var affectedChanges []attachRecordChange
	if slices.Contains(patchChangedFieldKeys, "evidence.lifecycle_state") {
		affectedChanges, err = refreshEvidenceSupportProjectionsTx(
			ctx,
			tx,
			f.supportEffects,
			meta.IncidentID,
			command.RecordID,
		)
		if err != nil {
			return MutationResult{}, err
		}
	}
	if err := appendEvidenceRecordChangeIntentsTx(
		ctx,
		tx,
		f.collaboration,
		meta.IncidentID,
		command.Actor.ID,
		request.ClientTxnID,
		changeSetID,
		attachRecordChange{
			RecordID:         command.RecordID,
			RowVersion:       rowVersion,
			ViewSchemaID:     request.ViewSchemaID,
			ChangedFieldKeys: patchChangedFieldKeys,
		},
		afterRow,
		affectedChanges,
		command.Now.UTC(),
	); err != nil {
		return MutationResult{}, err
	}
	payload := buildMutationPayload(request.ViewSchemaID, changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, command.RequestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit evidence patch transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       meta.IncidentID,
		RecordID:         command.RecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: patchChangedFieldKeys,
	}, nil
}
