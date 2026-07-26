package timelineassembly

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	entityfacts "github.com/JochiRaider/cartulary/internal/modules/entities/timelinefacts"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

type collectionReadAdapter struct {
	entities entityfacts.Reader
	links    links.TimelineFactReader
	evidence evidence.TimelineFactReader
}

func newCollectionReadAdapter() collectionReadAdapter {
	return collectionReadAdapter{
		entities: entityfacts.Reader{},
		links:    links.TimelineFactReader{},
		evidence: evidence.TimelineFactReader{},
	}
}

func (adapter collectionReadAdapter) LoadTimelineCollectionFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) (workbookprojection.CollectionFacts, error) {
	mentions, err := adapter.entities.LoadMentionsTx(ctx, tx, recordID)
	if err != nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline entity facts: %w", err)
	}
	linkFacts, err := adapter.links.LoadTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline link facts: %w", err)
	}
	evidenceFacts, err := adapter.evidence.LoadTx(ctx, tx, incidentID, linkFacts.AttachedEvidenceIDs)
	if err != nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline evidence facts: %w", err)
	}
	return workbookprojection.CollectionFacts{
		Mentions:            mentions,
		ResolvedLinks:       linkFacts.ResolvedLinks,
		Tags:                linkFacts.Tags,
		AttachedEvidence:    evidenceFacts,
		ReplacementRecordID: linkFacts.ReplacementRecordID,
	}, nil
}
