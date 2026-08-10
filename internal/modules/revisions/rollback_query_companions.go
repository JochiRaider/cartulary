package revisions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func (r rollbackQueryRepository) loadRollbackSiblingTargetsTx(ctx context.Context, tx pgx.Tx, target rollbackMutationTarget) ([]rollbackMutationTarget, error) {
	rows, err := tx.Query(ctx, `
SELECT csm.change_set_id,
       cs.created_at,
       cs.source,
       csm.sequence_no,
       csm.target_kind,
       csm.target_id,
       csm.operation_kind,
       csm.before_value,
       csm.after_value
  FROM change_set_mutations csm
  JOIN change_sets cs
    ON cs.change_set_id = csm.change_set_id
 WHERE csm.change_set_id = $1
   AND csm.sequence_no <> $2
 ORDER BY csm.sequence_no ASC
`, target.ChangeSetID, target.SequenceNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var siblings []rollbackMutationTarget
	for rows.Next() {
		var (
			sibling   rollbackMutationTarget
			beforeRaw []byte
			afterRaw  []byte
		)
		if err := rows.Scan(&sibling.ChangeSetID, &sibling.CreatedAt, &sibling.Source, &sibling.SequenceNo, &sibling.TargetKind, &sibling.TargetID, &sibling.OperationKind, &beforeRaw, &afterRaw); err != nil {
			return nil, err
		}
		before, err := decodeRollbackValue(beforeRaw)
		if err != nil {
			return nil, err
		}
		after, err := decodeRollbackValue(afterRaw)
		if err != nil {
			return nil, err
		}
		sibling.BeforeValue = before
		sibling.AfterValue = after
		siblings = append(siblings, sibling)
	}
	return siblings, rows.Err()
}

func (r rollbackQueryRepository) describeRollbackTargetTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	target rollbackMutationTarget,
	addressedRecordID uuid.UUID,
	siblings []rollbackMutationTarget,
) (rollbackcontract.TargetDescriptor, error) {
	dispatch, err := r.store.targetSemantics.dispatchClass(target.TargetKind)
	if err != nil {
		return rollbackcontract.TargetDescriptor{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	switch dispatch {
	case rollbackcontract.DispatchRow:
		history, err := r.store.targetSemantics.DescribeMutation(StoredMutation{
			TargetKind:  target.TargetKind,
			TargetID:    target.TargetID,
			BeforeValue: target.BeforeValue,
			AfterValue:  target.AfterValue,
		})
		if err != nil {
			return rollbackcontract.TargetDescriptor{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		return rollbackcontract.TargetDescriptor{AffectedRecordIDs: history.HistoryRecordIDs}, nil
	case rollbackcontract.DispatchNonRow:
		provider, err := r.store.targetSemantics.nonRowProvider(target.TargetKind)
		if err != nil {
			return rollbackcontract.TargetDescriptor{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		siblingFacts, err := r.rollbackSiblingFacts(siblings)
		if err != nil {
			return rollbackcontract.TargetDescriptor{}, err
		}
		return provider.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{
			Target:            nonRowContractTarget(incidentID, target),
			AddressedRecordID: addressedRecordID,
			SiblingTargets:    siblingFacts,
		})
	default:
		return rollbackcontract.TargetDescriptor{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
}

func (r rollbackQueryRepository) rollbackSiblingFacts(targets []rollbackMutationTarget) ([]rollbackcontract.SiblingTarget, error) {
	result := make([]rollbackcontract.SiblingTarget, 0, len(targets))
	for _, target := range targets {
		dispatch, err := r.store.targetSemantics.dispatchClass(target.TargetKind)
		if err != nil {
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		result = append(result, rollbackcontract.SiblingTarget{
			TargetReference: rollbackcontract.TargetReference{TargetKind: target.TargetKind, TargetID: target.TargetID},
			SequenceNo:      target.SequenceNo,
			DispatchClass:   dispatch,
		})
	}
	return result, nil
}

func (r rollbackQueryRepository) historyEntryRequiresChangeSetTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	target rollbackMutationTarget,
	descriptor rollbackcontract.TargetDescriptor,
	siblings []rollbackMutationTarget,
) (bool, error) {
	if descriptor.RequiresWholeChangeSet {
		return true, nil
	}
	dispatch, err := r.store.targetSemantics.dispatchClass(target.TargetKind)
	if err != nil {
		return false, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	targetReference := rollbackcontract.TargetReference{TargetKind: target.TargetKind, TargetID: target.TargetID}
	allTargets := append([]rollbackMutationTarget{target}, siblings...)
	for _, sibling := range siblings {
		siblingDescriptor, describeErr := r.describeRollbackTargetTx(
			ctx,
			tx,
			incidentID,
			sibling,
			uuid.Nil,
			rollbackTargetsExcept(allTargets, sibling.SequenceNo),
		)
		if describeErr != nil {
			adapted := adaptRowRollbackProviderError(describeErr)
			var precondition *RollbackPreconditionError
			if !errors.Is(adapted, ErrRollbackTargetNotFound) && !errors.As(adapted, &precondition) {
				return false, adapted
			}
		}
		if siblingDescriptor.RequiresWholeChangeSet ||
			(dispatch == rollbackcontract.DispatchRow && containsTargetReference(siblingDescriptor.RequiresWholeChangeSetWith, targetReference)) {
			return true, nil
		}
	}
	return false, nil
}

func selectRollbackCompanions(references []rollbackcontract.TargetReference, siblings []rollbackMutationTarget) ([]rollbackMutationTarget, error) {
	selected := make([]rollbackMutationTarget, 0, len(references))
	seen := make(map[int]struct{}, len(references))
	for _, reference := range references {
		matched := false
		for _, sibling := range siblings {
			if sibling.TargetKind != reference.TargetKind || sibling.TargetID != reference.TargetID {
				continue
			}
			if _, duplicate := seen[sibling.SequenceNo]; duplicate {
				return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
			}
			seen[sibling.SequenceNo] = struct{}{}
			selected = append(selected, sibling)
			matched = true
			break
		}
		if !matched {
			return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
	}
	return selected, nil
}

func containsTargetReference(values []rollbackcontract.TargetReference, target rollbackcontract.TargetReference) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func rollbackTargetsExcept(targets []rollbackMutationTarget, sequenceNo int) []rollbackMutationTarget {
	result := make([]rollbackMutationTarget, 0, len(targets))
	for _, target := range targets {
		if target.SequenceNo != sequenceNo {
			result = append(result, target)
		}
	}
	return result
}
