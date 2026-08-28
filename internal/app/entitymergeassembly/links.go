package entitymergeassembly

import (
	"context"

	"github.com/jackc/pgx/v5"

	entitymerge "github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/modules/links"
)

type linkEffects struct {
	store *links.Store
}

func NewLinkEffects() entitymerge.LinkEffectsPort {
	return linkEffects{store: links.NewStore()}
}

func (effects linkEffects) RepointLinksTx(
	ctx context.Context,
	tx pgx.Tx,
	command entitymerge.RepointLinksCommand,
) (entitymerge.RepointLinksResult, error) {
	result, err := effects.store.RepointMergedLinksTx(ctx, tx, links.RepointMergedLinksCommand{
		IncidentID:       command.IncidentID,
		SurvivorRecordID: command.SurvivorRecordID,
		LoserRecordID:    command.LoserRecordID,
		ActorUserID:      command.ActorUserID,
		Now:              command.Now,
	})
	if err != nil {
		return entitymerge.RepointLinksResult{}, err
	}
	return entitymerge.RepointLinksResult{
		Mutations:                 linkMutations(result.Mutations()),
		RepointedCount:            result.RepointedCount,
		DedupedCount:              result.DedupedCount,
		LinkTypesBySourceRecordID: result.LinkTypesBySourceRecordID(),
	}, nil
}

func (effects linkEffects) RepointTagsTx(
	ctx context.Context,
	tx pgx.Tx,
	command entitymerge.RepointTagsCommand,
) (entitymerge.RepointTagsResult, error) {
	result, err := effects.store.RepointMergedTagsTx(ctx, tx, links.RepointMergedTagsCommand{
		IncidentID:       command.IncidentID,
		SurvivorRecordID: command.SurvivorRecordID,
		LoserRecordID:    command.LoserRecordID,
		ActorUserID:      command.ActorUserID,
		Now:              command.Now,
	})
	if err != nil {
		return entitymerge.RepointTagsResult{}, err
	}
	return entitymerge.RepointTagsResult{
		Mutations:      linkMutations(result.Mutations()),
		RepointedCount: result.RepointedCount,
		DedupedCount:   result.DedupedCount,
	}, nil
}

func linkMutations(mutations []links.Mutation) []entitymerge.LinkEffectMutation {
	result := make([]entitymerge.LinkEffectMutation, len(mutations))
	for index, mutation := range mutations {
		result[index] = entitymerge.LinkEffectMutation{
			TargetKind:    mutation.TargetKind(),
			TargetID:      mutation.TargetID(),
			OperationKind: mutation.OperationKind(),
			BeforeValue:   mutation.BeforeValue(),
			AfterValue:    mutation.AfterValue(),
		}
	}
	return result
}
