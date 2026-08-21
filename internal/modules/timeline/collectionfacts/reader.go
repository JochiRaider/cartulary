package collectionfacts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

type LinkFact struct {
	TargetRecordID uuid.UUID
	LinkType       string
	Provenance     string
	Confidence     *int
}

type TagFact struct {
	RecordTagID uuid.UUID
	TagName     string
}

type LinkFacts struct {
	ResolvedLinks       []LinkFact
	Tags                []TagFact
	AttachedEvidenceIDs []uuid.UUID
	ReplacementRecordID *uuid.UUID
}

type MentionReader interface {
	LoadMentionsTx(context.Context, pgx.Tx, uuid.UUID) ([]workbookprojection.MentionFact, error)
}

type LinkReader interface {
	LoadTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (LinkFacts, error)
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
	resolvedLinks := make([]workbookprojection.LinkFact, len(linkFacts.ResolvedLinks))
	for index, fact := range linkFacts.ResolvedLinks {
		resolvedLinks[index] = workbookprojection.LinkFact{
			TargetRecordID: fact.TargetRecordID,
			LinkType:       fact.LinkType,
			Provenance:     fact.Provenance,
			Confidence:     fact.Confidence,
		}
	}
	tags := make([]workbookprojection.TagFact, len(linkFacts.Tags))
	for index, fact := range linkFacts.Tags {
		tags[index] = workbookprojection.TagFact{
			RecordTagID: fact.RecordTagID,
			TagName:     fact.TagName,
		}
	}
	return workbookprojection.CollectionFacts{
		Mentions:            mentions,
		ResolvedLinks:       resolvedLinks,
		Tags:                tags,
		AttachedEvidence:    attachedEvidence,
		ReplacementRecordID: linkFacts.ReplacementRecordID,
	}, nil
}
