package collectionfacts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

type MentionReader interface {
	LoadMentionsTx(context.Context, pgx.Tx, uuid.UUID) ([]workbookprojection.MentionFact, error)
}

type LinkReader interface {
	LoadTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (links.TimelineFacts, error)
}

type EvidenceReader interface {
	LoadTx(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) ([]evidence.TimelineFact, error)
}

// Reader loads Timeline collection facts in their authoritative dependency
// order using the caller-owned transaction. It owns neither authorization nor
// transaction lifecycle.
type Reader struct {
	mentions MentionReader
	links    LinkReader
	evidence EvidenceReader
}

func New(mentions MentionReader, links LinkReader, evidence EvidenceReader) Reader {
	return Reader{mentions: mentions, links: links, evidence: evidence}
}

func (reader Reader) LoadTimelineCollectionFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) (workbookprojection.CollectionFacts, error) {
	if reader.mentions == nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline collection facts: mention reader is required")
	}
	if reader.links == nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline collection facts: link reader is required")
	}
	if reader.evidence == nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline collection facts: evidence reader is required")
	}

	mentions, err := reader.mentions.LoadMentionsTx(ctx, tx, recordID)
	if err != nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline entity facts: %w", err)
	}
	linkFacts, err := reader.links.LoadTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline link facts: %w", err)
	}
	evidenceFacts, err := reader.evidence.LoadTx(ctx, tx, incidentID, linkFacts.AttachedEvidenceIDs)
	if err != nil {
		return workbookprojection.CollectionFacts{}, fmt.Errorf("load Timeline evidence facts: %w", err)
	}
	attachedEvidence := make([]workbookprojection.EvidenceFact, len(evidenceFacts))
	for index, fact := range evidenceFacts {
		attachedEvidence[index] = workbookprojection.EvidenceFact{
			RecordID:       fact.RecordID,
			Title:          fact.Title,
			LifecycleState: fact.LifecycleState,
			UploadState:    fact.UploadState,
		}
	}
	return workbookprojection.CollectionFacts{
		Mentions:            mentions,
		ResolvedLinks:       linkFacts.ResolvedLinks,
		Tags:                linkFacts.Tags,
		AttachedEvidence:    attachedEvidence,
		ReplacementRecordID: linkFacts.ReplacementRecordID,
	}, nil
}
