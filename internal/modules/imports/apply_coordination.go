package imports

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

var (
	errImportActorUnauthorized = errors.New("imports: actor is not currently authorized")
	errImportUnitCanceled      = errors.New("imports: unit apply canceled")
	errUnitCommitIndeterminate = errors.New("imports: unit commit is indeterminate")
)

func (s *service) applyUnit(
	ctx context.Context,
	execution jobs.Execution,
	actor authn.UserRecord,
	start applyStartResult,
	unitID uuid.UUID,
) (outcome unitApplyOutcome, retErr error) {
	if existing, err := s.store.getUnitOutcome(ctx, start.ImportSessionID, unitID); err == nil {
		return existing, nil
	} else if !errors.Is(err, errNotFound) {
		return unitApplyOutcome{}, err
	}
	ctx, cancelWork := context.WithCancelCause(ctx)
	defer cancelWork(nil)
	defer func() {
		if cause := context.Cause(ctx); cause != nil && retErr != nil && !errors.Is(retErr, errUnitCommitIndeterminate) && !crossownertransaction.IsFatalIntegrity(retErr) {
			retErr = cause
		}
	}()
	stopObservation := s.observeApplyCancellation(ctx, cancelWork, execution)
	defer stopObservation()
	var targetKind string
	if err := s.store.db.QueryRow(ctx, `SELECT target_kind FROM import_apply_unit_plans
WHERE import_session_id=$1 AND import_unit_id=$2 AND apply_job_id=$3`, start.ImportSessionID, unitID, start.Job.JobID).Scan(&targetKind); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return unitApplyOutcome{}, importApplyBlockedError("unit_not_ready")
		}
		return unitApplyOutcome{}, err
	}
	if targetKind != ImportTargetKindViewSchema {
		return s.applyAnalyticalUnit(ctx, execution, actor, start, unitID, stopObservation)
	}
	tx, err := s.applyTransactions.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return unitApplyOutcome{}, fmt.Errorf("begin import apply unit transaction: %w", err)
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()
	unit, target, currentActor, err := s.validateUnitTx(ctx, tx, actor, start, unitID)
	if err != nil {
		if existing, findErr := findUnitOutcome(ctx, tx, start.ImportSessionID, unitID); findErr == nil {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				return unitApplyOutcome{}, commitErr
			}
			return existing, nil
		}
		return unitApplyOutcome{}, err
	}
	if target.TargetKind != ImportTargetKindViewSchema {
		return unitApplyOutcome{}, importApplyBlockedError("source_changed")
	}
	commit, err := s.applyGenericOwnerUnitTx(ctx, tx, currentActor, start, unit, target)
	if err != nil {
		return unitApplyOutcome{}, err
	}
	outcome, err = s.publishUnitTx(ctx, tx, execution, actor, start, unit, target, commit, stopObservation)
	if err != nil {
		return unitApplyOutcome{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		if recovered, findErr := s.store.getUnitOutcome(
			context.WithoutCancel(ctx),
			start.ImportSessionID,
			unitID,
		); findErr == nil {
			return recovered, nil
		}
		return unitApplyOutcome{}, fmt.Errorf(
			"%w: commit import apply unit transaction: %v",
			errUnitCommitIndeterminate,
			err,
		)
	}
	return outcome, nil
}

// validateUnitTx and publishUnitTx are the common Core unit-commit boundaries
// for view-schema owner creation and analytical participant execution.
func (s *service) validateUnitTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, start applyStartResult, unitID uuid.UUID) (applyUnitData, importTarget, authn.UserRecord, error) {
	unit, err := s.store.lockApplyUnitTx(ctx, tx, start, unitID)
	if err != nil {
		return applyUnitData{}, importTarget{}, authn.UserRecord{}, err
	}
	if _, err := s.incidentAccess.CheckTx(ctx, tx, start.IncidentID, actor.ID, admission.Requirement{AllowedRoles: admission.RolesEditorReviewerAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
		return applyUnitData{}, importTarget{}, authn.UserRecord{}, err
	}
	currentActor, err := s.authStore.GetUserByIDTx(ctx, tx, actor.ID)
	if err != nil || !currentActor.IsActive {
		return applyUnitData{}, importTarget{}, authn.UserRecord{}, errImportActorUnauthorized
	}
	target, ok := lookupApprovedImportTarget(unit.approvedMapping)
	if !ok || !target.importable(s.extensionProfileClaimed) {
		reason := "target_kind_not_importable"
		if unit.approvedMapping.targetKindOrDefault() == ImportTargetKindViewSchema {
			reason = "target_view_schema_not_importable"
		}
		return applyUnitData{}, importTarget{}, authn.UserRecord{}, importApplyBlockedError(reason)
	}
	if target.TargetKind == ImportTargetKindViewSchema && !target.ownerCreateFacadeAvailable() {
		return applyUnitData{}, importTarget{}, authn.UserRecord{}, importApplyBlockedError("owner_create_contract_unavailable")
	}
	if target.TargetKind != ImportTargetKindViewSchema && !target.ownerApplyFacadeAvailable() {
		return applyUnitData{}, importTarget{}, authn.UserRecord{}, importApplyBlockedError("owner_apply_contract_unavailable")
	}
	return unit, target, currentActor, nil
}

func (s *service) publishUnitTx(ctx context.Context, tx pgx.Tx, execution jobs.Execution, actor authn.UserRecord, start applyStartResult, unit applyUnitData, target importTarget, commit appliedUnitCommit, stopObservation func()) (unitApplyOutcome, error) {
	stopObservation()
	if cause := context.Cause(ctx); cause != nil {
		return unitApplyOutcome{}, cause
	}
	// The Jobs lock is acquired only after expensive owner work.
	if err := s.store.ensureApplyJobRunnableTx(ctx, tx, execution); err != nil {
		return unitApplyOutcome{}, err
	}
	currentActor, err := s.authStore.GetUserByIDForUpdateTx(ctx, tx, actor.ID)
	if err != nil || !currentActor.IsActive {
		return unitApplyOutcome{}, errImportActorUnauthorized
	}
	now := s.now().UTC()
	if err := commitSourceTx(ctx, tx, start, unit, currentActor.ID, now); err != nil {
		return unitApplyOutcome{}, err
	}
	return s.store.insertAppliedUnitOutcomeTx(ctx, tx, start, unit, target, currentActor.ID, commit, now)
}

func sourceRowCellsByOrdinal(sourceRow map[string]any) map[int]map[string]any {
	cellsByOrdinal := map[int]map[string]any{}
	if cells, ok := sourceRow["cells"].([]any); ok {
		for _, rawCell := range cells {
			cell, ok := rawCell.(map[string]any)
			if !ok {
				continue
			}
			ordinal, ok := intFromAny(cell["source_column_ordinal"])
			if ok {
				cellsByOrdinal[ordinal] = cell
			}
		}
		return cellsByOrdinal
	}
	if cells, ok := sourceRow["cells"].([]map[string]any); ok {
		for _, cell := range cells {
			ordinal, ok := intFromAny(cell["source_column_ordinal"])
			if ok {
				cellsByOrdinal[ordinal] = cell
			}
		}
	}
	return cellsByOrdinal
}
