package merge

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LinkEffectMutation struct {
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
}

type RepointLinksCommand struct {
	IncidentID       uuid.UUID
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	ActorUserID      uuid.UUID
	Now              time.Time
}

type RepointLinksResult struct {
	Mutations                 []LinkEffectMutation
	RepointedCount            int
	DedupedCount              int
	LinkTypesBySourceRecordID map[uuid.UUID][]string
}

type RepointTagsCommand struct {
	IncidentID       uuid.UUID
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	ActorUserID      uuid.UUID
	Now              time.Time
}

type RepointTagsResult struct {
	Mutations      []LinkEffectMutation
	RepointedCount int
	DedupedCount   int
}

type LinkEffectsPort interface {
	RepointLinksTx(context.Context, pgx.Tx, RepointLinksCommand) (RepointLinksResult, error)
	RepointTagsTx(context.Context, pgx.Tx, RepointTagsCommand) (RepointTagsResult, error)
}

func mergeMutationsFromLinkEffects(mutations []LinkEffectMutation) []mergeMutation {
	result := make([]mergeMutation, len(mutations))
	for index, mutation := range mutations {
		result[index] = mergeMutation{
			TargetKind:      mutation.TargetKind,
			TargetID:        mutation.TargetID,
			OperationKind:   mutation.OperationKind,
			BeforeVersionID: mutation.BeforeVersionID,
			AfterVersionID:  mutation.AfterVersionID,
			BeforeValue:     mutation.BeforeValue,
			AfterValue:      mutation.AfterValue,
		}
	}
	return result
}
