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
	switch plan.Target.TargetKind {
	case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
		change, err := a.applyRowBackedRollbackTx(ctx, tx, actor, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		changes = append(changes, change)
	case "record_link", "entity_alias", "entity_preserved_identifier", "indicator_observation", "indicator_state_interval":
		linkChanges, err := a.applyRecordLinkRollbackTx(ctx, tx, actor, record.IncidentID, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		changes = append(changes, linkChanges...)
	case "entity_mention":
		mentionChanges, err := a.applyMentionRollbackTx(ctx, tx, actor, record.IncidentID, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		changes = append(changes, mentionChanges...)
	case "record_tag":
		tagChanges, err := a.applyRecordTagRollbackTx(ctx, tx, actor, record.IncidentID, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		changes = append(changes, tagChanges...)
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
	beforeSnapshot, err := a.snapshotRecordTx(ctx, tx, record.RecordID, provider)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	nextRowVersion, err := a.envelopes.AdvanceVersionTx(ctx, tx, record.RecordID, actor.UUID(), now)
	if err != nil {
		return rollbackApplyResult{}, adaptEnvelopeError(err)
	}
	if err := a.restoreRollbackSourceTx(ctx, tx, record.RecordType, record.RecordID, actor.UUID(), now, nextRowVersion, plan.RestoreSnapshot); err != nil {
		return rollbackApplyResult{}, err
	}
	if err := a.publication.rebuildProjectionsTx(ctx, tx, record.IncidentID); err != nil {
		return rollbackApplyResult{}, err
	}
	afterSnapshot, err := a.snapshotRecordTx(ctx, tx, record.RecordID, provider)
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
	beforeVersionID := fmt.Sprintf("record:%s:%d", record.RecordID, record.RowVersion)
	afterVersionID := fmt.Sprintf("record:%s:%d", record.RecordID, nextRowVersion)
	targetKind := rollbackMutationTargetKindForRecordType(record.RecordType)
	if targetKind != "record" {
		beforeVersionID = fmt.Sprintf("%s:%s:%d", targetKind, record.RecordID, record.RowVersion)
		afterVersionID = fmt.Sprintf("%s:%s:%d", targetKind, record.RecordID, nextRowVersion)
	}
	if err := a.publication.appendMutationTx(ctx, tx, AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      targetKind,
		TargetID:        record.RecordID.String(),
		OperationKind:   "row_restore",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeSnapshot,
		AfterValue:      afterSnapshot,
	}); err != nil {
		return rollbackApplyResult{}, err
	}
	if err := a.publication.appendRecordRevisionTx(ctx, tx, AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    record.RecordID,
		RowVersion:  nextRowVersion,
		BeforeValue: beforeSnapshot,
		AfterValue:  afterSnapshot,
	}); err != nil {
		return rollbackApplyResult{}, err
	}
	viewSchemaID, err := provider.ViewSchemaID(ctx, tx, record.RecordID)
	if err != nil {
		return rollbackApplyResult{}, err
	}
	change := RollbackRecordChange{
		RecordID:         record.RecordID,
		RowVersion:       nextRowVersion,
		ChangeSetID:      changeSetID,
		ViewSchemaID:     viewSchemaID,
		ChangedFieldKeys: rollbackChangedFieldKeys(beforeSnapshot, afterSnapshot),
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
		switch target.TargetKind {
		case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
			_, _, err := a.applyRowBackedRollbackMutationTx(ctx, tx, actor, target, changeSetID, sequenceNo, now, nextVersions)
			if err != nil {
				return nil, err
			}
		case "record_link", "entity_mention", "record_tag", "entity_preserved_identifier", "entity_alias", "indicator_observation", "indicator_state_interval":
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
	recordType, err := rollbackRecordTypeForTarget(target, targetRecord.RecordType)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	provider, ok := a.deleteRestoreSources.Source(recordType)
	if !ok {
		return uuid.UUID{}, nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	beforeSnapshot, err := provider.SnapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	nextRowVersion, ok := nextVersions[targetRecordID]
	if !ok {
		return uuid.UUID{}, nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if err := a.restoreRollbackSourceTx(ctx, tx, recordType, targetRecordID, actor.UUID(), now, nextRowVersion, target.BeforeValue); err != nil {
		return uuid.UUID{}, nil, err
	}
	afterSnapshot, err := provider.SnapshotTx(ctx, tx, targetRecordID)
	if err != nil {
		return uuid.UUID{}, nil, err
	}
	beforeVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, targetRecord.RowVersion)
	afterVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, nextRowVersion)
	if target.TargetKind != "record" {
		beforeVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, targetRecord.RowVersion)
		afterVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, nextRowVersion)
	}
	if err := a.publication.appendMutationTx(ctx, tx, AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      *sequenceNo,
		TargetKind:      target.TargetKind,
		TargetID:        target.TargetID,
		OperationKind:   "rollback",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeSnapshot,
		AfterValue:      afterSnapshot,
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
	provider, ok := a.nonRowRollbackProviders.Provider(target.TargetKind)
	if !ok {
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

func (a rollbackTransactionalApplier) insertRollbackRecordRevisionSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changeSetID uuid.UUID, beforeSnapshot map[string]any, rowVersion int64) (RollbackRecordChange, error) {
	record, err := a.repository.loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	provider, ok := a.deleteRestoreSources.Source(record.RecordType)
	if !ok {
		return RollbackRecordChange{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	afterSnapshot, err := a.snapshotRecordTx(ctx, tx, recordID, provider)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	if err := a.publication.appendRecordRevisionTx(ctx, tx, AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  rowVersion,
		BeforeValue: beforeSnapshot,
		AfterValue:  afterSnapshot,
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	viewSchemaID, err := provider.ViewSchemaID(ctx, tx, recordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	return RollbackRecordChange{
		RecordID:         recordID,
		RowVersion:       rowVersion,
		ChangeSetID:      changeSetID,
		ViewSchemaID:     viewSchemaID,
		ChangedFieldKeys: rollbackChangedFieldKeys(beforeSnapshot, afterSnapshot),
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
	recordType, err := rollbackRecordTypeForTarget(target, targetRecord.RecordType)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	provider, ok := a.deleteRestoreSources.Source(recordType)
	if !ok {
		return RollbackRecordChange{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	beforeSnapshot, err := a.snapshotRecordTx(ctx, tx, targetRecordID, provider)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	nextRowVersion, err := a.envelopes.AdvanceVersionTx(ctx, tx, targetRecordID, actor.UUID(), now)
	if err != nil {
		return RollbackRecordChange{}, adaptEnvelopeError(err)
	}
	if err := a.restoreRollbackSourceTx(ctx, tx, recordType, targetRecordID, actor.UUID(), now, nextRowVersion, target.BeforeValue); err != nil {
		return RollbackRecordChange{}, err
	}
	if err := a.publication.rebuildProjectionsTx(ctx, tx, targetRecord.IncidentID); err != nil {
		return RollbackRecordChange{}, err
	}
	afterSnapshot, err := a.snapshotRecordTx(ctx, tx, targetRecordID, provider)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	beforeVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, targetRecord.RowVersion)
	afterVersionID := fmt.Sprintf("record:%s:%d", targetRecordID, nextRowVersion)
	if target.TargetKind != "record" {
		beforeVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, targetRecord.RowVersion)
		afterVersionID = fmt.Sprintf("%s:%s:%d", target.TargetKind, targetRecordID, nextRowVersion)
	}
	if err := a.publication.appendMutationTx(ctx, tx, AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      *sequenceNo,
		TargetKind:      target.TargetKind,
		TargetID:        target.TargetID,
		OperationKind:   "rollback",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeSnapshot,
		AfterValue:      afterSnapshot,
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	(*sequenceNo)++
	if err := a.publication.appendRecordRevisionTx(ctx, tx, AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    targetRecordID,
		RowVersion:  nextRowVersion,
		BeforeValue: beforeSnapshot,
		AfterValue:  afterSnapshot,
	}); err != nil {
		return RollbackRecordChange{}, err
	}
	viewSchemaID, err := provider.ViewSchemaID(ctx, tx, targetRecordID)
	if err != nil {
		return RollbackRecordChange{}, err
	}
	return RollbackRecordChange{
		RecordID:         targetRecordID,
		RowVersion:       nextRowVersion,
		ChangeSetID:      changeSetID,
		ViewSchemaID:     viewSchemaID,
		ChangedFieldKeys: rollbackChangedFieldKeys(beforeSnapshot, afterSnapshot),
	}, nil
}

func (a rollbackTransactionalApplier) insertRollbackMutationTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, sequenceNo *int, target rollbackMutationTarget, beforeValue any, afterValue any) error {
	if err := a.publication.appendMutationTx(ctx, tx, AppendMutationParams{
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

func (a rollbackTransactionalApplier) applyRecordLinkRollbackTx(ctx context.Context, tx pgx.Tx, actor ActorID, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	return a.applyNonRowRollbackTx(ctx, tx, actor, incidentID, plan, changeSetID, sequenceNo, now)
}

func (a rollbackTransactionalApplier) applyMentionRollbackTx(ctx context.Context, tx pgx.Tx, actor ActorID, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	target := plan.Target
	beforeSnapshots, err := a.snapshotRollbackAffectedRecordsTx(ctx, tx, plan.Affected)
	if err != nil {
		return nil, err
	}
	mentionResult, err := a.executeNonRowInverseTx(ctx, tx, actor, incidentID, target, now)
	if err != nil {
		return nil, err
	}
	if err := validateNonRowApplyResult(rollbackcontract.TargetDescriptor{AffectedRecordIDs: plan.Affected}, mentionResult); err != nil {
		return nil, err
	}
	companionResults := make([]rollbackcontract.ApplyInverseResult, 0, len(plan.Companion))
	for _, companion := range plan.Companion {
		result, err := a.executeNonRowInverseTx(ctx, tx, actor, incidentID, companion, now)
		if err != nil {
			return nil, err
		}
		companionResults = append(companionResults, result)
	}
	nextVersions, err := a.advanceRollbackAffectedRecordsTx(ctx, tx, actor, plan.Affected, now)
	if err != nil {
		return nil, err
	}
	if err := a.publication.rebuildProjectionsTx(ctx, tx, incidentID); err != nil {
		return nil, err
	}
	if err := a.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, target, mentionResult.BeforeValue, mentionResult.AfterValue); err != nil {
		return nil, err
	}
	for index, companion := range plan.Companion {
		result := companionResults[index]
		if err := a.insertRollbackMutationTx(ctx, tx, changeSetID, sequenceNo, companion, result.BeforeValue, result.AfterValue); err != nil {
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

func (a rollbackTransactionalApplier) applyRecordTagRollbackTx(ctx context.Context, tx pgx.Tx, actor ActorID, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	return a.applyNonRowRollbackTx(ctx, tx, actor, incidentID, plan, changeSetID, sequenceNo, now)
}

func (a rollbackTransactionalApplier) applyNonRowRollbackTx(ctx context.Context, tx pgx.Tx, actor ActorID, incidentID uuid.UUID, plan rollbackPlan, changeSetID uuid.UUID, sequenceNo *int, now time.Time) ([]RollbackRecordChange, error) {
	beforeSnapshots, err := a.snapshotRollbackAffectedRecordsTx(ctx, tx, plan.Affected)
	if err != nil {
		return nil, err
	}
	result, err := a.applyNonRowRollbackMutationTx(ctx, tx, actor, incidentID, plan.Target, changeSetID, sequenceNo, now)
	if err != nil {
		return nil, err
	}
	if err := validateNonRowApplyResult(rollbackcontract.TargetDescriptor{AffectedRecordIDs: plan.Affected}, result); err != nil {
		return nil, err
	}
	nextVersions, err := a.advanceRollbackAffectedRecordsTx(ctx, tx, actor, plan.Affected, now)
	if err != nil {
		return nil, err
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

func (a rollbackTransactionalApplier) snapshotRollbackAffectedRecordsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID]map[string]any, error) {
	snapshots := make(map[uuid.UUID]map[string]any, len(recordIDs))
	for _, recordID := range recordIDs {
		record, err := a.repository.loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
		if err != nil {
			return nil, err
		}
		provider, ok := a.deleteRestoreSources.Source(record.RecordType)
		if !ok {
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		snapshot, err := a.snapshotRecordTx(ctx, tx, recordID, provider)
		if err != nil {
			return nil, err
		}
		snapshots[recordID] = snapshot
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

func (a rollbackTransactionalApplier) restoreRollbackSourceTx(ctx context.Context, tx pgx.Tx, recordType string, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, rowVersion int64, retainedValue map[string]any) error {
	provider, ok := a.rowRollbackProviders.Provider(recordType)
	if !ok {
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
