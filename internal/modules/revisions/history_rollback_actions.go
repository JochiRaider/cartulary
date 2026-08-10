package revisions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type historyRollbackActionEvaluator struct {
	repository rollbackQueryRepository
	planner    rollbackPlanner
}

func newHistoryRollbackActionEvaluator(commands *commandStore) historyRollbackActionEvaluator {
	return historyRollbackActionEvaluator{
		repository: rollbackQueryRepository{store: commands},
		planner:    rollbackPlanner{targetSemantics: commands.targetSemantics},
	}
}

func (e historyRollbackActionEvaluator) DecorateTx(ctx context.Context, tx pgx.Tx, record RecordHistoryRecord, items []RecordHistoryItem) error {
	rollbackRecord := rollbackRecordEnvelope{
		IncidentID:      record.IncidentID,
		RecordID:        record.RecordID,
		RecordType:      record.RecordType,
		RowVersion:      record.RowVersion,
		DeletedAt:       record.DeletedAt,
		DeletedByUserID: record.DeletedByID,
	}
	changeSetExecutable := make(map[uuid.UUID]bool)
	changeSetChecked := make(map[uuid.UUID]bool)
	for index := range items {
		items[index].AvailableRollbackActions = nil
		if rollbackRecord.DeletedAt != nil {
			items[index].RevisionNo = nil
			items[index].Reversible = false
			continue
		}
		if items[index].HistoryEntryRef != nil {
			executable, err := e.historyEntryRollbackExecutableTx(ctx, tx, rollbackRecord, *items[index].HistoryEntryRef)
			if err != nil {
				return err
			}
			if executable {
				items[index].AvailableRollbackActions = append(items[index].AvailableRollbackActions, "history_entry")
			}
		}
		if !changeSetChecked[items[index].ChangeSetID] {
			executable, err := e.changeSetRollbackExecutableTx(ctx, tx, rollbackRecord, items[index].ChangeSetID)
			if err != nil {
				return err
			}
			changeSetChecked[items[index].ChangeSetID] = true
			changeSetExecutable[items[index].ChangeSetID] = executable
		}
		if changeSetExecutable[items[index].ChangeSetID] {
			items[index].AvailableRollbackActions = append(items[index].AvailableRollbackActions, "change_set")
		}
		if items[index].RevisionNo != nil {
			executable, err := e.rowRestoreExecutableTx(ctx, tx, rollbackRecord, *items[index].RevisionNo)
			if err != nil {
				return err
			}
			if executable {
				items[index].AvailableRollbackActions = append(items[index].AvailableRollbackActions, "row_restore")
			} else {
				items[index].RevisionNo = nil
			}
		}
		items[index].Reversible = len(items[index].AvailableRollbackActions) > 0
	}
	return nil
}

func (e historyRollbackActionEvaluator) historyEntryRollbackExecutableTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, historyEntryRef string) (bool, error) {
	plan, err := e.repository.loadHistoryEntryRollbackPlanTx(ctx, tx, record, historyEntryRef)
	if err != nil {
		if errors.Is(err, ErrRollbackTargetNotFound) || errors.Is(err, ErrRollbackPreconditionFailed) {
			return false, nil
		}
		return false, err
	}
	return e.rollbackPlanExecutableTx(ctx, tx, plan)
}

func (e historyRollbackActionEvaluator) changeSetRollbackExecutableTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, changeSetID uuid.UUID) (bool, error) {
	plan, err := e.repository.loadChangeSetRollbackPlanTx(ctx, tx, record, changeSetID.String())
	if err != nil {
		if errors.Is(err, ErrRollbackTargetNotFound) || errors.Is(err, ErrRollbackPreconditionFailed) {
			return false, nil
		}
		return false, err
	}
	return e.rollbackPlanExecutableTx(ctx, tx, plan)
}

func (e historyRollbackActionEvaluator) rollbackPlanExecutableTx(ctx context.Context, tx pgx.Tx, plan rollbackPlan) (bool, error) {
	if err := e.planner.validateRollbackPlan(plan); err != nil {
		if errors.Is(err, ErrRollbackPreconditionFailed) {
			return false, nil
		}
		return false, err
	}
	if err := e.repository.ensurePlanCurrentTx(ctx, tx, plan); err != nil {
		if errors.Is(err, ErrRollbackPreconditionFailed) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (e historyRollbackActionEvaluator) rowRestoreExecutableTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, revisionNo int64) (bool, error) {
	_, err := e.repository.loadRowRestorePlanTx(ctx, tx, record, revisionNo)
	if err != nil {
		if errors.Is(err, ErrRollbackTargetNotFound) || errors.Is(err, ErrRollbackPreconditionFailed) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
