package revisions

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

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
