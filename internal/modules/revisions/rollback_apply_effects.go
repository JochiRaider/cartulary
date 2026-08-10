package revisions

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (a rollbackTransactionalApplier) insertRollbackRecordRevisionSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changeSetID uuid.UUID, beforeSnapshot rollbackAffectedRecordSnapshot, rowVersion int64) (RollbackRecordChange, error) {
	record, err := a.repository.loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	provider, ok := a.deleteRestoreSources.Source(record.RecordType)
	if !ok {
		return RollbackRecordChange{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	afterSnapshot, err := a.publication.captureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	afterLiveRecord, err := a.loadLiveRecordTx(ctx, tx, beforeSnapshot.viewSchemaID, recordID, provider)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	if err := a.publication.appendRecordRevisionAndIntentTx(ctx, tx, AppendRecordRevisionParams{
		ChangeSetID:    changeSetID,
		RecordID:       recordID,
		RowVersion:     rowVersion,
		BeforeSnapshot: &beforeSnapshot.captured,
		AfterSnapshot:  &afterSnapshot,
		LiveChange: LiveRecordChange{
			BeforeValue: beforeSnapshot.live,
			AfterValue:  afterLiveRecord,
		},
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	return RollbackRecordChange{
		RecordID:         recordID,
		RowVersion:       rowVersion,
		ChangeSetID:      changeSetID,
		ViewSchemaID:     beforeSnapshot.viewSchemaID,
		ChangedFieldKeys: rollbackChangedFieldKeys(beforeSnapshot.live, afterLiveRecord),
	}, nil
}

func (a rollbackTransactionalApplier) insertRollbackMutationTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, sequenceNo *int, target rollbackMutationTarget, beforeValue any, afterValue any) error {
	if err := a.publication.appendNonRowMutationTx(ctx, tx, AppendNonRowMutationParams{
		ChangeSetID:   changeSetID,
		SequenceNo:    *sequenceNo,
		TargetKind:    target.TargetKind,
		TargetID:      target.TargetID,
		OperationKind: "rollback",
		BeforeValue:   beforeValue,
		AfterValue:    afterValue,
	}); err != nil {
		return err
	}
	(*sequenceNo)++
	return nil
}

type rollbackAffectedRecordSnapshot struct {
	captured     RecordSnapshot
	live         map[string]any
	viewSchemaID string
}

func (a rollbackTransactionalApplier) snapshotRollbackAffectedRecordsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID]rollbackAffectedRecordSnapshot, error) {
	snapshots := make(map[uuid.UUID]rollbackAffectedRecordSnapshot, len(recordIDs))
	for _, recordID := range recordIDs {
		record, err := a.repository.loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
		if err != nil {
			return nil, err
		}
		provider, ok := a.deleteRestoreSources.Source(record.RecordType)
		if !ok {
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		viewSchemaID, err := provider.ViewSchemaID(ctx, tx, recordID)
		if err != nil {
			return nil, err
		}
		captured, err := a.publication.captureRecordSnapshotTx(ctx, tx, recordID)
		if err != nil {
			return nil, err
		}
		live, err := a.loadLiveRecordTx(ctx, tx, viewSchemaID, recordID, provider)
		if err != nil {
			return nil, err
		}
		snapshots[recordID] = rollbackAffectedRecordSnapshot{
			captured:     captured,
			live:         live,
			viewSchemaID: viewSchemaID,
		}
	}
	return snapshots, nil
}

func (a rollbackTransactionalApplier) advanceRollbackAffectedRecordsTx(ctx context.Context, tx pgx.Tx, actor ActorID, recordIDs []uuid.UUID, now time.Time) (map[uuid.UUID]int64, error) {
	nextVersions := make(map[uuid.UUID]int64, len(recordIDs))
	for _, recordID := range recordIDs {
		nextRowVersion, err := a.envelopes.AdvanceVersionTx(ctx, tx, recordID, actor.UUID(), now)
		if err != nil {
			return nil, adaptEnvelopeError(err)
		}
		nextVersions[recordID] = nextRowVersion
	}
	return nextVersions, nil
}
