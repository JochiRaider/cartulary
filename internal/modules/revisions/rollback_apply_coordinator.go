package revisions

import (
	"context"
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
	case rollbackcontract.DispatchRow:
		change, err := a.applyRowBackedRollbackTx(ctx, tx, actor, plan, changeSetID, &sequenceNo, now)
		if err != nil {
			return rollbackApplyResult{}, err
		}
		changes = append(changes, change)
	case rollbackcontract.DispatchNonRow:
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
	if err := a.syncRollbackEnvelopeMirrorsTx(ctx, tx, plan.Affected); err != nil {
		return nil, err
	}

	for _, step := range plan.ApplyOrder {
		target := step.Target
		dispatch, err := a.targetSemantics.dispatchClass(target.TargetKind)
		if err != nil {
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		switch dispatch {
		case rollbackcontract.DispatchRow:
			_, _, err := a.applyRowBackedRollbackMutationTx(ctx, tx, actor, target, changeSetID, sequenceNo, now, nextVersions)
			if err != nil {
				return nil, err
			}
		case rollbackcontract.DispatchNonRow:
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
