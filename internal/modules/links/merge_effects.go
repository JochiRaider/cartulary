package links

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links/internal/mergeeffects"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/mutationvalue"
)

type RepointMergedLinksCommand struct {
	IncidentID       uuid.UUID
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	ActorUserID      uuid.UUID
	Now              time.Time
}

type RepointMergedLinksResult struct {
	mutations                 []Mutation
	RepointedCount            int
	DedupedCount              int
	linkTypesBySourceRecordID map[uuid.UUID][]string
}

func (result RepointMergedLinksResult) Mutations() []Mutation {
	return mutationvalue.Copy(result.mutations)
}

func (result RepointMergedLinksResult) LinkTypesBySourceRecordID() map[uuid.UUID][]string {
	return cloneLinkTypeInvalidations(result.linkTypesBySourceRecordID)
}

type RepointMergedTagsCommand struct {
	IncidentID       uuid.UUID
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	ActorUserID      uuid.UUID
	Now              time.Time
}

type RepointMergedTagsResult struct {
	mutations      []Mutation
	RepointedCount int
	DedupedCount   int
}

func (result RepointMergedTagsResult) Mutations() []Mutation {
	return mutationvalue.Copy(result.mutations)
}

func (s *Store) RepointMergedLinksTx(
	ctx context.Context,
	tx pgx.Tx,
	command RepointMergedLinksCommand,
) (RepointMergedLinksResult, error) {
	result, err := mergeeffects.RepointLinksTx(ctx, tx, mergeeffects.RepointLinksCommand(command), mergeeffects.LinkDependencies{
		Validate: func(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType LinkType, provenance LinkProvenance, confidence *int) error {
			if err := validateRecordLinkCommand(linkType, provenance, confidence, srcRecordID, dstRecordID); err != nil {
				return err
			}
			return validateActiveLinkEndpointsTx(ctx, tx, incidentID, srcRecordID, dstRecordID)
		},
		Tombstone: func(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID, actorUserID uuid.UUID, now time.Time) (*time.Time, error) {
			state, err := tombstoneRecordLinkStateTx(ctx, tx, recordLinkID, actorUserID, now)
			return state.deletedAt, err
		},
	})
	if err != nil {
		return RepointMergedLinksResult{}, err
	}
	return RepointMergedLinksResult{
		mutations:                 mutationvalue.Copy(result.Mutations),
		RepointedCount:            result.RepointedCount,
		DedupedCount:              result.DedupedCount,
		linkTypesBySourceRecordID: cloneLinkTypeInvalidations(result.LinkTypesBySourceRecordID),
	}, nil
}

func (s *Store) RepointMergedTagsTx(
	ctx context.Context,
	tx pgx.Tx,
	command RepointMergedTagsCommand,
) (RepointMergedTagsResult, error) {
	result, err := mergeeffects.RepointTagsTx(ctx, tx, mergeeffects.RepointTagsCommand(command))
	if err != nil {
		return RepointMergedTagsResult{}, err
	}
	return RepointMergedTagsResult{
		mutations:      mutationvalue.Copy(result.Mutations),
		RepointedCount: result.RepointedCount,
		DedupedCount:   result.DedupedCount,
	}, nil
}

func cloneLinkTypeInvalidations(source map[uuid.UUID][]string) map[uuid.UUID][]string {
	if source == nil {
		return nil
	}
	result := make(map[uuid.UUID][]string, len(source))
	for recordID, linkTypes := range source {
		result[recordID] = append([]string(nil), linkTypes...)
	}
	return result
}
