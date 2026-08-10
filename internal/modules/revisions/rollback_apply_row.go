package revisions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func (a rollbackTransactionalApplier) applyRowRestorePlanTx(ctx context.Context, tx pgx.Tx, actor ActorID, record rollbackRecordEnvelope, plan rollbackPlan, request RollbackRequest, requestID string, now time.Time) (rollbackApplyResult, error) {
	if err := a.ensureExpectedVersionsTx(ctx, tx, plan); err != nil {
		return rollbackApplyResult{}, err
	}
	provider, ok := a.deleteRestoreSources.Source(record.RecordType)
	if !ok {
		return rollbackApplyResult{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	viewSchemaID, err := provider.ViewSchemaID(ctx, tx, record.RecordID)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	beforeSnapshot, err := a.publication.captureRecordSnapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	beforeLiveRecord, err := a.loadLiveRecordTx(ctx, tx, viewSchemaID, record.RecordID, provider)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	nextRowVersion, err := a.envelopes.AdvanceVersionTx(ctx, tx, record.RecordID, actor.UUID(), now)
	if err != nil {
		return rollbackApplyResult{}, adaptEnvelopeError(err)
	}
	targetKind, err := a.targetSemantics.defaultRowTargetKind(record.RecordType)
	if err != nil {
		return rollbackApplyResult{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err := a.restoreRollbackSourceTx(ctx, tx, targetKind, record.RecordType, record.RecordID, actor.UUID(), now, nextRowVersion, plan.RestoreSnapshot); err != nil {
		return rollbackApplyResult{}, err
	}
	if err := a.publication.rebuildProjectionsTx(ctx, tx, record.IncidentID); err != nil {
		return rollbackApplyResult{}, err
	}
	afterSnapshot, err := a.publication.captureRecordSnapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	afterLiveRecord, err := a.loadLiveRecordTx(ctx, tx, viewSchemaID, record.RecordID, provider)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	changeSetID, err := a.publication.appendChangeSetTx(ctx, tx, AppendChangeSetParams{
		IncidentID:  record.IncidentID,
		ActorUserID: actor.UUID(),
		Source:      "rollback",
		Reason:      request.Reason,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now,
	})
	if err != nil {
		return rollbackApplyResult{}, err
	}
	beforeVersionID := fmt.Sprintf("%s:%s:%d", targetKind, record.RecordID, record.RowVersion)
	afterVersionID := fmt.Sprintf("%s:%s:%d", targetKind, record.RecordID, nextRowVersion)
	if err := a.publication.appendRecordMutationTx(ctx, tx, AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      targetKind,
		RecordID:        record.RecordID,
		OperationKind:   "row_restore",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  &beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return rollbackApplyResult{}, err
	}
	if err := a.publication.appendRecordRevisionAndIntentTx(ctx, tx, AppendRecordRevisionParams{
		ChangeSetID:    changeSetID,
		RecordID:       record.RecordID,
		RowVersion:     nextRowVersion,
		BeforeSnapshot: &beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		LiveChange: LiveRecordChange{
			BeforeValue: beforeLiveRecord,
			AfterValue:  afterLiveRecord,
		},
	}); err != nil {
		return rollbackApplyResult{}, err
	}
	change := RollbackRecordChange{
		RecordID:         record.RecordID,
		RowVersion:       nextRowVersion,
		ChangeSetID:      changeSetID,
		ViewSchemaID:     viewSchemaID,
		ChangedFieldKeys: rollbackChangedFieldKeys(beforeLiveRecord, afterLiveRecord),
	}
	return rollbackApplyResult{ChangeSetID: changeSetID, Changes: []RollbackRecordChange{change}}, nil
}

func (a rollbackTransactionalApplier) applyRowBackedRollbackMutationTx(ctx context.Context, tx pgx.Tx, actor ActorID, target rollbackMutationTarget, changeSetID uuid.UUID, sequenceNo *int, now time.Time, nextVersions map[uuid.UUID]int64) (uuid.UUID, []string, error) {
	targetRecordID, err := uuid.Parse(target.TargetID)
	if err != nil {
		return uuid.UUID{}, nil, ErrRollbackTargetNotFound
	}
	targetRecord, err := a.repository.loadRollbackRecordEnvelopeTx(ctx, tx, targetRecordID, false)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	if _, err := a.targetSemantics.rowProvider(target.TargetKind, targetRecord.RecordType); err != nil {
		return uuid.UUID{}, nil, err
	}
	recordType := targetRecord.RecordType
	if _, ok := a.deleteRestoreSources.Source(recordType); !ok {
		return uuid.UUID{}, nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	beforeSnapshot, err := a.publication.captureRecordSnapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	nextRowVersion, ok := nextVersions[targetRecordID]
	if !ok {
		return uuid.UUID{}, nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err := a.restoreRollbackSourceTx(ctx, tx, target.TargetKind, recordType, targetRecordID, actor.UUID(), now, nextRowVersion, target.BeforeValue); err != nil {
		return uuid.UUID{}, nil, err
	}
	afterSnapshot, err := a.publication.captureRecordSnapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	beforeVersionID := fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, targetRecord.RowVersion)
	afterVersionID := fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, nextRowVersion)
	if err := a.publication.appendRecordMutationTx(ctx, tx, AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      *sequenceNo,
		TargetKind:      target.TargetKind,
		RecordID:        targetRecordID,
		OperationKind:   "rollback",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  &beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return uuid.UUID{}, nil, err
	}
	(*sequenceNo)++
	return targetRecordID, rollbackChangedFieldKeys(target.BeforeValue, target.AfterValue), nil
}

func (a rollbackTransactionalApplier) applyRowBackedRollbackTx(ctx context.Context, tx pgx.Tx, actor ActorID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) (RollbackRecordChange, error) {
	target := plan.Target
	targetRecordID, err := uuid.Parse(target.TargetID)
	if err != nil {
		return RollbackRecordChange{}, ErrRollbackTargetNotFound
	}
	targetRecord, err := a.repository.loadRollbackRecordEnvelopeTx(ctx, tx, targetRecordID, false)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	if _, err := a.targetSemantics.rowProvider(target.TargetKind, targetRecord.RecordType); err != nil {
		return RollbackRecordChange{}, err
	}
	recordType := targetRecord.RecordType
	provider, ok := a.deleteRestoreSources.Source(recordType)
	if !ok {
		return RollbackRecordChange{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	viewSchemaID, err := provider.ViewSchemaID(ctx, tx, targetRecordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	beforeSnapshot, err := a.publication.captureRecordSnapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	beforeLiveRecord, err := a.loadLiveRecordTx(ctx, tx, viewSchemaID, targetRecordID, provider)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	nextRowVersion, err := a.envelopes.AdvanceVersionTx(ctx, tx, targetRecordID, actor.UUID(), now)
	if err != nil {
		return RollbackRecordChange{}, adaptEnvelopeError(err)
	}
	if err := a.restoreRollbackSourceTx(ctx, tx, target.TargetKind, recordType, targetRecordID, actor.UUID(), now, nextRowVersion, target.BeforeValue); err != nil {
		return RollbackRecordChange{}, err
	}
	if err := a.publication.rebuildProjectionsTx(ctx, tx, targetRecord.IncidentID); err != nil {
		return RollbackRecordChange{}, err
	}
	afterSnapshot, err := a.publication.captureRecordSnapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	afterLiveRecord, err := a.loadLiveRecordTx(ctx, tx, viewSchemaID, targetRecordID, provider)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	beforeVersionID := fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, targetRecord.RowVersion)
	afterVersionID := fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, nextRowVersion)
	if err := a.publication.appendRecordMutationTx(ctx, tx, AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      *sequenceNo,
		TargetKind:      target.TargetKind,
		RecordID:        targetRecordID,
		OperationKind:   "rollback",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  &beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	(*sequenceNo)++
	if err := a.publication.appendRecordRevisionAndIntentTx(ctx, tx, AppendRecordRevisionParams{
		ChangeSetID:    changeSetID,
		RecordID:       targetRecordID,
		RowVersion:     nextRowVersion,
		BeforeSnapshot: &beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		LiveChange: LiveRecordChange{
			BeforeValue: beforeLiveRecord,
			AfterValue:  afterLiveRecord,
		},
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	return RollbackRecordChange{
		RecordID:         targetRecordID,
		RowVersion:       nextRowVersion,
		ChangeSetID:      changeSetID,
		ViewSchemaID:     viewSchemaID,
		ChangedFieldKeys: rollbackChangedFieldKeys(beforeLiveRecord, afterLiveRecord),
	}, nil
}

func (a rollbackTransactionalApplier) restoreRollbackSourceTx(ctx context.Context, tx pgx.Tx, targetKind string, recordType string, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64, retainedValue map[string]any) error {
	provider, err := a.targetSemantics.rowProvider(targetKind, recordType)
	if err != nil {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err := provider.ValidateRollbackValue(retainedValue); err != nil {
		return adaptRowRollbackProviderError(err)
	}
	return adaptRowRollbackProviderError(provider.RestoreTx(ctx, tx, rollbackcontract.RestoreRequest{
		RecordID:       recordID,
		ActorUserID:    actorUserID,
		Now:            now.UTC(),
		NextRowVersion: rowVersion,
		RetainedValue:  retainedValue,
	}))
}
