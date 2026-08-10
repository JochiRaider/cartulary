package revisions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type rollbackTransactionalApplier struct {
	*commandStore
	repository  rollbackQueryRepository
	publication rollbackPublicationService
}

func (a rollbackTransactionalApplier) applyRollbackPlanTx(ctx context.Context, tx pgx.Tx, actor ActorID, record rollbackRecordEnvelope, plan rollbackPlan, request RollbackRequest, requestID string, now time.Time) (rollbackApplyResult, error) {
	if err := a.ensureExpectedVersionsTx(ctx, tx, plan); err != nil {
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
	sequenceNo := 1
	var changes []RollbackRecordChange
	if plan.WholeSet {
		changes, err = a.applyChangeSetRollbackPlanTx(ctx, tx, actor, record.IncidentID, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		return rollbackApplyResult{ChangeSetID: changeSetID, Changes: changes}, nil
	}
	dispatch, err := a.targetSemantics.dispatchClass(plan.Target.TargetKind)
	if err != nil {
		return rollbackApplyResult{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	switch dispatch {
	case RollbackDispatchRow:
		change, err := a.applyRowBackedRollbackTx(ctx, tx, actor, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		changes = append(changes, change)
	case RollbackDispatchNonRow:
		nonRowChanges, err := a.applyNonRowRollbackTx(ctx, tx, actor, record.IncidentID, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		changes = append(changes, nonRowChanges...)
	default:
		return rollbackApplyResult{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	return rollbackApplyResult{ChangeSetID: changeSetID, Changes: changes}, nil
}

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
	if err := a.publication.appendCapturedRecordMutationTx(ctx, tx, AppendCapturedRecordMutationParams{
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
	if err := a.publication.appendCapturedRecordRevisionTx(ctx, tx, AppendCapturedRecordRevisionParams{
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

func (a rollbackTransactionalApplier) applyChangeSetRollbackPlanTx(ctx context.Context, tx pgx.Tx, actor ActorID, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	beforeSnapshots, err := a.snapshotRollbackAffectedRecordsTx(ctx, tx, plan.Affected)
	if err != nil {
		return nil, err
	}
	nextVersions := make(map[uuid.UUID]int64, len(plan.Affected))
	for _, recordID := range plan.Affected {
		nextRowVersion, err := a.envelopes.AdvanceVersionTx(ctx, tx, recordID, actor.UUID(), now)
		if err != nil {
			return nil, adaptEnvelopeError(err)
		}
		nextVersions[recordID] = nextRowVersion
	}

	for _, step := range plan.ApplyOrder {
		target := step.Target
		dispatch, err := a.targetSemantics.dispatchClass(target.TargetKind)
		if err != nil {
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		switch dispatch {
		case RollbackDispatchRow:
			_, _, err := a.applyRowBackedRollbackMutationTx(ctx, tx, actor, target, changeSetID, sequenceNo, now, nextVersions)
			if err != nil {
				return nil, err
			}
		case RollbackDispatchNonRow:
			_, err := a.applyNonRowRollbackMutationTx(ctx, tx, actor, incidentID, target, changeSetID, sequenceNo, now)
			if err != nil {
				return nil, err
			}
		default:
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
	}
	if err := a.publication.rebuildProjectionsTx(ctx, tx, incidentID); err != nil {
		return nil, err
	}

	changes := make([]RollbackRecordChange, 0, len(plan.Affected))
	for _, recordID := range plan.Affected {
		change, err := a.insertRollbackRecordRevisionSnapshotTx(ctx, tx, recordID, changeSetID, beforeSnapshots[recordID], nextVersions[recordID])
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, nil
}

func (a rollbackTransactionalApplier) ensureExpectedVersionsTx(ctx context.Context, tx pgx.Tx, plan rollbackPlan) error {
	for _, recordID := range canonicalRecordIDs(plan.Affected) {
		record, err := a.repository.loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
		if err != nil {
			return err
		}
		if record.RowVersion != plan.ExpectedVersions[recordID] {
			return &RollbackPreconditionError{ReasonCode: "stale_target"}
		}
	}
	return nil
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
	if err := a.publication.appendCapturedRecordMutationTx(ctx, tx, AppendCapturedRecordMutationParams{
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

func (a rollbackTransactionalApplier) applyNonRowRollbackMutationTx(ctx context.Context, tx pgx.Tx, actor ActorID, incidentID uuid.UUID, target rollbackMutationTarget, changeSetID uuid.UUID, sequenceNo *int, now time.Time) (rollbackcontract.ApplyInverseResult, error) {
	result, err := a.executeNonRowInverseTx(ctx, tx, actor, incidentID, target, now)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	if err := a.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, target, result.BeforeValue, result.AfterValue); err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	return result, nil
}

func (a rollbackTransactionalApplier) executeNonRowInverseTx(ctx context.Context, tx pgx.Tx, actor ActorID, incidentID uuid.UUID, target rollbackMutationTarget, now time.Time) (rollbackcontract.ApplyInverseResult, error) {
	provider, err := a.targetSemantics.nonRowProvider(target.TargetKind)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	contractTarget := nonRowContractTarget(incidentID, target)
	descriptor, err := provider.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{Target: contractTarget})
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, adaptRowRollbackProviderError(err)
	}
	result, err := provider.ApplyInverseTx(ctx, tx, rollbackcontract.ApplyInverseRequest{
		Target:      contractTarget,
		ActorUserID: actor.UUID(),
		Now:         now,
	})
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, adaptRowRollbackProviderError(err)
	}
	if err := validateNonRowApplyResult(descriptor, result); err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	return result, nil
}

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
	if err := a.publication.appendCapturedRecordRevisionTx(ctx, tx, AppendCapturedRecordRevisionParams{
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
	if err := a.publication.appendCapturedRecordMutationTx(ctx, tx, AppendCapturedRecordMutationParams{
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
	if err := a.publication.appendCapturedRecordRevisionTx(ctx, tx, AppendCapturedRecordRevisionParams{
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

func (a rollbackTransactionalApplier) applyNonRowRollbackTx(ctx context.Context, tx pgx.Tx, actor ActorID, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	beforeSnapshots, err := a.snapshotRollbackAffectedRecordsTx(ctx, tx, plan.Affected)
	if err != nil {
		return nil, err
	}
	targets := append([]rollbackMutationTarget{plan.Target}, plan.Companion...)
	results := make([]rollbackcontract.ApplyInverseResult, 0, len(targets))
	for index, target := range targets {
		result, err := a.executeNonRowInverseTx(ctx, tx, actor, incidentID, target, now)
		if err != nil {
			return nil, err
		}
		if index == 0 {
			if err := validateNonRowApplyResult(rollbackcontract.TargetDescriptor{AffectedRecordIDs: plan.Affected}, result); err != nil {
				return nil, err
			}
		}
		results = append(results, result)
	}
	nextVersions, err := a.advanceRollbackAffectedRecordsTx(ctx, tx, actor, plan.Affected, now)
	if err != nil {
		return nil, err
	}
	if err := a.publication.rebuildProjectionsTx(ctx, tx, incidentID); err != nil {
		return nil, err
	}
	for index, target := range targets {
		result := results[index]
		if err := a.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, target, result.BeforeValue, result.AfterValue); err != nil {
			return nil, err
		}
	}
	changes := make([]RollbackRecordChange, 0, len(plan.Affected))
	for _, recordID := range plan.Affected {
		change, err := a.insertRollbackRecordRevisionSnapshotTx(ctx, tx, recordID, changeSetID, beforeSnapshots[recordID], nextVersions[recordID])
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, nil
}

type rollbackAffectedRecordSnapshot struct {
	captured     CapturedRecordSnapshot
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
