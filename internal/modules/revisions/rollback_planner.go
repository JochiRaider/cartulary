package revisions

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type rollbackPlanner struct {
	nonRowProviders *NonRowProviderCatalog
}

func (p rollbackPlanner) finalize(plan rollbackPlan, envelopes map[uuid.UUID]rollbackRecordEnvelope) (rollbackPlan, error) {
	plan.Affected = canonicalRecordIDs(plan.Affected)
	plan.ExpectedVersions = make(map[uuid.UUID]int64, len(plan.Affected))
	for _, recordID := range plan.Affected {
		envelope, ok := envelopes[recordID]
		if !ok {
			return rollbackPlan{}, ErrRollbackTargetNotFound
		}
		plan.ExpectedVersions[recordID] = envelope.RowVersion
	}

	targets := make([]rollbackMutationTarget, 0, len(plan.Targets)+len(plan.Companion)+1)
	switch {
	case plan.WholeSet:
		for index := len(plan.Targets) - 1; index >= 0; index-- {
			targets = append(targets, plan.Targets[index])
		}
	default:
		targets = append(targets, plan.Target)
		targets = append(targets, plan.Companion...)
	}
	plan.ApplyOrder = make([]rollbackPlanStep, 0, len(targets))
	for index, target := range targets {
		providerID := "nonrow/" + target.TargetKind
		if firstClassRollbackTargetKind(target.TargetKind) {
			providerID = "row/" + target.TargetKind
			if target.TargetKind == "record" {
				providerID = "row/" + plan.RecordType
			}
		}
		plan.ApplyOrder = append(plan.ApplyOrder, rollbackPlanStep{
			Order:            index + 1,
			TargetIdentity:   fmt.Sprintf("%s:%s:%s:%d", target.TargetKind, target.TargetID, target.ChangeSetID, target.SequenceNo),
			ProviderID:       providerID,
			Target:           target,
			ChangedFieldKeys: rollbackChangedFieldKeys(target.AfterValue, target.BeforeValue),
		})
	}
	if err := p.validateFinalizedPlan(plan); err != nil {
		return rollbackPlan{}, err
	}
	return plan, nil
}

func (p rollbackPlanner) validateFinalizedPlan(plan rollbackPlan) error {
	if err := p.validateRollbackPlan(plan); err != nil {
		return err
	}
	if len(plan.Affected) == 0 || len(plan.ExpectedVersions) != len(plan.Affected) || len(plan.ApplyOrder) == 0 {
		return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	for index, step := range plan.ApplyOrder {
		if step.Order != index+1 || step.TargetIdentity == "" || step.ProviderID == "" {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
	}
	return nil
}

func (p rollbackPlanner) validateRollbackPlan(plan rollbackPlan) error {
	if plan.DeferredErr != nil {
		return plan.DeferredErr
	}
	if plan.RestoreRevisionNo > 0 {
		if len(plan.Affected) != 1 || plan.RestoreSnapshot == nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		return nil
	}
	if plan.RequiresChangeSet {
		return &RollbackPreconditionError{ReasonCode: "entry_requires_change_set"}
	}
	if plan.WholeSet {
		if len(plan.Targets) == 0 {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		for _, target := range plan.Targets {
			if err := p.validateRollbackTarget(target); err != nil {
				return err
			}
		}
		return nil
	}
	return p.validateRollbackTarget(plan.Target)
}

func deferableRollbackProviderError(err error) bool {
	if errors.Is(err, ErrRollbackTargetNotFound) {
		return true
	}
	var precondition *RollbackPreconditionError
	return errors.As(err, &precondition)
}

func (p rollbackPlanner) validateRollbackTarget(target rollbackMutationTarget) error {
	if firstClassRollbackTargetKind(target.TargetKind) {
		if target.BeforeValue == nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		return nil
	}
	if _, ok := p.nonRowProviders.Provider(target.TargetKind); ok {
		return nil
	}
	return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
}

func rollbackRecordTypeForTarget(target rollbackMutationTarget, currentRecordType string) (string, error) {
	if currentRecordType == "" {
		return "", &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	if target.TargetKind != "record" && rollbackMutationTargetKindForRecordType(currentRecordType) != target.TargetKind {
		return "", &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	return currentRecordType, nil
}

func rollbackMutationTargetKindForRecordType(recordType string) string {
	switch recordType {
	case "timeline_event":
		return "timeline_record"
	case "host", "identity", "indicator", "assessment", "evidence":
		return recordType
	default:
		return "record"
	}
}

func nonRowContractTarget(incidentID uuid.UUID, target rollbackMutationTarget) rollbackcontract.NonRowTarget {
	return rollbackcontract.NonRowTarget{
		IncidentID:    incidentID,
		ChangeSetID:   target.ChangeSetID,
		SequenceNo:    target.SequenceNo,
		TargetKind:    target.TargetKind,
		TargetID:      target.TargetID,
		OperationKind: target.OperationKind,
		BeforeValue:   target.BeforeValue,
		AfterValue:    target.AfterValue,
	}
}

func affectedRecordsForRollbackTarget(target rollbackMutationTarget, fallback uuid.UUID) ([]uuid.UUID, error) {
	if !firstClassRollbackTargetKind(target.TargetKind) {
		return []uuid.UUID{fallback}, nil
	}
	recordID, err := uuid.Parse(target.TargetID)
	if err != nil || recordID == uuid.Nil {
		return nil, ErrRollbackTargetNotFound
	}
	return []uuid.UUID{recordID}, nil
}

func firstClassRollbackTargetKind(targetKind string) bool {
	switch targetKind {
	case "record", "timeline_record", "host", "identity", "indicator", "assessment", "evidence":
		return true
	default:
		return false
	}
}
