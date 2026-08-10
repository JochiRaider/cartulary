package revisions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type rollbackQueryRepository struct {
	store *commandStore
}

func (r rollbackQueryRepository) loadRollbackRecordEnvelopesTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID]rollbackRecordEnvelope, error) {
	envelopes := make(map[uuid.UUID]rollbackRecordEnvelope, len(recordIDs))
	for _, recordID := range canonicalRecordIDs(recordIDs) {
		envelope, err := r.loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
		if err != nil {
			return nil, err
		}
		envelopes[recordID] = envelope
	}
	return envelopes, nil
}

func (r rollbackQueryRepository) loadRollbackProtectedSetTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, target RollbackTarget) (rollbackProtectedSet, error) {
	fallback := rollbackProtectedSet{Affected: []uuid.UUID{record.RecordID}}
	switch target.Kind {
	case "history_entry":
		mutation, err := loadHistoryEntryRollbackTargetTx(ctx, tx, record, target.HistoryEntryRef)
		if errors.Is(err, ErrRollbackTargetNotFound) {
			fallback.DeferredErr = ErrRollbackTargetNotFound
			return fallback, nil
		}
		if err != nil {
			return rollbackProtectedSet{}, err
		}
		affected, err := r.affectedRecordsForRollbackTargetTx(ctx, tx, record.IncidentID, mutation, record.RecordID)
		if err != nil {
			fallback.DeferredErr = ErrRollbackTargetNotFound
			return fallback, nil
		}
		return rollbackProtectedSet{Affected: affected}, nil
	case "change_set":
		plan, err := r.loadChangeSetRollbackPlanTx(ctx, tx, record, target.ChangeSetID)
		if errors.Is(err, ErrRollbackTargetNotFound) {
			fallback.DeferredErr = ErrRollbackTargetNotFound
			return fallback, nil
		}
		if err != nil {
			return rollbackProtectedSet{}, err
		}
		return rollbackProtectedSet{Affected: plan.Affected}, nil
	case "row_restore":
		return fallback, nil
	default:
		fallback.DeferredErr = &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		return fallback, nil
	}
}

func (r rollbackQueryRepository) loadRollbackRecordEnvelopeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, forUpdate bool) (rollbackRecordEnvelope, error) {
	record, err := r.store.envelopes.LoadEnvelopeTx(ctx, tx, recordID, forUpdate)
	if err != nil {
		return rollbackRecordEnvelope{}, adaptEnvelopeError(err)
	}
	return record, nil
}
