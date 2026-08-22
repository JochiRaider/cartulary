package links

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links/internal/mergeeffects"
)

type MergeMutation struct {
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
}

type RepointMergedLinksCommand struct {
	IncidentID       uuid.UUID
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	ActorUserID      uuid.UUID
	Now              time.Time
}

type RepointMergedLinksResult struct {
	Mutations                 []MergeMutation
	RepointedCount            int
	DedupedCount              int
	LinkTypesBySourceRecordID map[uuid.UUID][]string
}

type RepointMergedTagsCommand struct {
	IncidentID       uuid.UUID
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	ActorUserID      uuid.UUID
	Now              time.Time
}

type RepointMergedTagsResult struct {
	Mutations      []MergeMutation
	RepointedCount int
	DedupedCount   int
}

func (s *Store) RepointMergedLinksTx(
	ctx context.Context,
	tx pgx.Tx,
	command RepointMergedLinksCommand,
) (RepointMergedLinksResult, error) {
	result, err := mergeeffects.RepointLinksTx(ctx, tx, mergeeffects.RepointLinksCommand(command), mergeeffects.LinkDependencies{
		Validate: func(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType string, provenance string, confidence *int) error {
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
		Mutations:                 mergeMutations(result.Mutations),
		RepointedCount:            result.RepointedCount,
		DedupedCount:              result.DedupedCount,
		LinkTypesBySourceRecordID: result.LinkTypesBySourceRecordID,
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
		Mutations:      mergeMutations(result.Mutations),
		RepointedCount: result.RepointedCount,
		DedupedCount:   result.DedupedCount,
	}, nil
}

func mergeMutations(mutations []mergeeffects.Mutation) []MergeMutation {
	result := make([]MergeMutation, len(mutations))
	for index, mutation := range mutations {
		result[index] = MergeMutation(mutation)
	}
	return result
}
