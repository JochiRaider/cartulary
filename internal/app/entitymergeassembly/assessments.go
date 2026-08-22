package entitymergeassembly

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	entitymerge "github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type assessmentEffectsAdapter struct {
	effects *assessments.MergeEffects
}

func NewAssessmentEffects(effects *assessments.MergeEffects) (entitymerge.AssessmentEffectsPort, error) {
	if effects == nil {
		return nil, fmt.Errorf("compose entity merge assessment effects: concrete effects are required")
	}
	return assessmentEffectsAdapter{effects: effects}, nil
}

func (a assessmentEffectsAdapter) LoadProtectedRecordIDsTx(
	ctx context.Context,
	tx pgx.Tx,
	command entitymerge.AssessmentProtectedSetCommand,
) ([]uuid.UUID, error) {
	recordIDs, err := a.effects.LoadProtectedRecordIDsTx(
		ctx,
		tx,
		command.IncidentID,
		command.RecordType,
		command.LoserRecordID,
	)
	return append([]uuid.UUID(nil), recordIDs...), err
}

func (a assessmentEffectsAdapter) RepointTx(
	ctx context.Context,
	tx pgx.Tx,
	command entitymerge.AssessmentRepointCommand,
) (entitymerge.AssessmentRepointResult, error) {
	protectedRecordIDs := append([]uuid.UUID(nil), command.ProtectedRecordIDs...)
	protectedRecordSet := make(map[uuid.UUID]struct{}, len(protectedRecordIDs))
	for _, recordID := range protectedRecordIDs {
		protectedRecordSet[recordID] = struct{}{}
	}
	mutations, repointedCount, err := a.effects.RepointTx(
		ctx,
		tx,
		command.IncidentID,
		command.RecordType,
		command.SurvivorRecordID,
		command.LoserRecordID,
		protectedRecordSet,
		command.Now,
	)
	if err != nil {
		var changed *assessments.MergeProtectedSetChangedError
		if errors.As(err, &changed) {
			return entitymerge.AssessmentRepointResult{}, &entitymerge.AssessmentProtectedSetChangedError{RecordID: changed.RecordID}
		}
		return entitymerge.AssessmentRepointResult{}, err
	}
	result := entitymerge.AssessmentRepointResult{
		Mutations:      make([]entitymerge.AssessmentMutation, 0, len(mutations)),
		RepointedCount: repointedCount,
	}
	for _, mutation := range mutations {
		result.Mutations = append(result.Mutations, entitymerge.AssessmentMutation{
			TargetKind:     mutation.TargetKind,
			TargetID:       mutation.TargetID,
			OperationKind:  mutation.OperationKind,
			BeforeValue:    cloneAssessmentValue(mutation.BeforeValue),
			AfterValue:     cloneAssessmentValue(mutation.AfterValue),
			BeforeSnapshot: cloneSnapshot(mutation.BeforeSnapshot),
			AfterSnapshot:  cloneSnapshot(mutation.AfterSnapshot),
		})
	}
	return result, nil
}

func cloneSnapshot(snapshot *revisions.RecordSnapshot) *revisions.RecordSnapshot {
	if snapshot == nil {
		return nil
	}
	cloned := *snapshot
	return &cloned
}

func cloneAssessmentValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		cloned := make(map[string]any, len(typed))
		for key, item := range typed {
			cloned[key] = cloneAssessmentValue(item)
		}
		return cloned
	case []any:
		cloned := make([]any, len(typed))
		for index, item := range typed {
			cloned[index] = cloneAssessmentValue(item)
		}
		return cloned
	case []string:
		return append([]string(nil), typed...)
	case []uuid.UUID:
		return append([]uuid.UUID(nil), typed...)
	case *int:
		if typed == nil {
			return (*int)(nil)
		}
		cloned := *typed
		return &cloned
	case *string:
		if typed == nil {
			return (*string)(nil)
		}
		cloned := *typed
		return &cloned
	case *time.Time:
		if typed == nil {
			return (*time.Time)(nil)
		}
		cloned := *typed
		return &cloned
	case *uuid.UUID:
		if typed == nil {
			return (*uuid.UUID)(nil)
		}
		cloned := *typed
		return &cloned
	default:
		return value
	}
}
