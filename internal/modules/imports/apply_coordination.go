package imports

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

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

func (s *Service) applyUnit(
	ctx context.Context,
	execution jobs.Execution,
	actor authn.UserRecord,
	start ApplyStartResult,
	unitID uuid.UUID,
) (unitApplyOutcome, error) {
	if existing, err := s.store.getUnitOutcome(ctx, start.ImportSessionID, unitID); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrNotFound) {
		return unitApplyOutcome{}, err
	}
	tx, err := s.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return unitApplyOutcome{}, fmt.Errorf("begin import apply unit transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := s.store.ensureApplyJobRunnableTx(ctx, tx, execution); err != nil {
		return unitApplyOutcome{}, err
	}
	unit, err := s.store.lockApplyUnitTx(ctx, tx, start, unitID)
	if err != nil {
		if existing, findErr := findUnitOutcome(ctx, tx, start.ImportSessionID, unitID); findErr == nil {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				return unitApplyOutcome{}, commitErr
			}
			return existing, nil
		}
		return unitApplyOutcome{}, err
	}
	if _, err := s.incidentAccess.CheckTx(
		ctx,
		tx,
		start.IncidentID,
		actor.ID,
		admission.Requirement{
			AllowedRoles: admission.RolesEditorReviewerAdmin,
			Lifecycle:    admission.LifecycleOpen,
		},
	); err != nil {
		return unitApplyOutcome{}, err
	}
	currentActor, err := s.authStore.GetUserByIDForUpdateTx(ctx, tx, actor.ID)
	if err != nil || !currentActor.IsActive {
		return unitApplyOutcome{}, errImportActorUnauthorized
	}
	target, ok := lookupApprovedImportTarget(unit.ApprovedMapping)
	if !ok || !target.importable(s.extensionProfileClaimed) {
		if unit.ApprovedMapping.targetKindOrDefault() == ImportTargetKindViewSchema {
			return unitApplyOutcome{}, importApplyBlockedError("target_view_schema_not_importable")
		}
		return unitApplyOutcome{}, importApplyBlockedError("target_kind_not_importable")
	}
	var commit appliedUnitCommit
	if unit.ApprovedMapping.targetKindOrDefault() != ImportTargetKindViewSchema {
		if !target.ownerApplyFacadeAvailable() {
			return unitApplyOutcome{}, importApplyBlockedError("owner_apply_contract_unavailable")
		}
		commit, err = s.applyExtensionOwnerUnitTx(ctx, tx, currentActor, start, unit, target)
	} else {
		if !target.ownerCreateFacadeAvailable() {
			return unitApplyOutcome{}, importApplyBlockedError("owner_create_contract_unavailable")
		}
		commit, err = s.applyGenericOwnerUnitTx(ctx, tx, currentActor, start, unit, target)
	}
	if err != nil {
		return unitApplyOutcome{}, err
	}
	outcome, err := s.store.insertAppliedUnitOutcomeTx(
		ctx,
		tx,
		start,
		unit,
		target,
		currentActor.ID,
		commit,
		s.now().UTC(),
	)
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
