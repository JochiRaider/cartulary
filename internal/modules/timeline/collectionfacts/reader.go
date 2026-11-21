package collectionfacts

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidencefacts "github.com/JochiRaider/cartulary/internal/modules/evidence/timelinefacts"
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
	// LoadMentionsTx returns the active mentions of each record that has any,
	// in each record's authoritative order.
	LoadMentionsTx(context.Context, pgx.Tx, []uuid.UUID) (map[uuid.UUID][]workbookprojection.MentionFact, error)
}

type LinkReader interface {
	// LoadTx returns the link facts of every requested record.
	LoadTx(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) (map[uuid.UUID]LinkFacts, error)
}

type EvidenceReader interface {
	LoadTx(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) ([]evidencefacts.TimelineFact, error)
}

// Reader loads Timeline collection facts in their authoritative dependency
// order using the caller-owned transaction. It owns neither authorization nor
// transaction lifecycle. Each fact source is read once per call, however many
// records the call names, so a projection page costs a fixed number of reads.
type Reader struct {
	mentions MentionReader
	links    LinkReader
	evidence EvidenceReader
}

func New(mentions MentionReader, links LinkReader, evidence EvidenceReader) Reader {
	return Reader{mentions: mentions, links: links, evidence: evidence}
}

// LoadTimelineCollectionFactsTx loads one record's facts.
func (reader Reader) LoadTimelineCollectionFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) (workbookprojection.CollectionFacts, error) {
	facts, err := reader.LoadTimelineCollectionFactsForRecordsTx(ctx, tx, incidentID, []uuid.UUID{recordID})
	if err != nil {
		return workbookprojection.CollectionFacts{}, err
	}
	return facts[recordID], nil
}

// LoadTimelineCollectionFactsForRecordsTx loads the facts of every requested
// record of one incident with one read per fact source.
func (reader Reader) LoadTimelineCollectionFactsForRecordsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordIDs []uuid.UUID,
) (map[uuid.UUID]workbookprojection.CollectionFacts, error) {
	if reader.mentions == nil {
		return nil, fmt.Errorf("load Timeline collection facts: mention reader is required")
	}
	if reader.links == nil {
		return nil, fmt.Errorf("load Timeline collection facts: link reader is required")
	}
	if reader.evidence == nil {
		return nil, fmt.Errorf("load Timeline collection facts: evidence reader is required")
	}

	mentions, err := reader.mentions.LoadMentionsTx(ctx, tx, recordIDs)
	if err != nil {
		return nil, fmt.Errorf("load Timeline entity facts: %w", err)
	}
	linkFacts, err := reader.links.LoadTx(ctx, tx, incidentID, recordIDs)
	if err != nil {
		return nil, fmt.Errorf("load Timeline link facts: %w", err)
	}
	attachedIDs := make([]uuid.UUID, 0)
	attachedSeen := make(map[uuid.UUID]bool)
	for _, recordID := range recordIDs {
		for _, evidenceID := range linkFacts[recordID].AttachedEvidenceIDs {
			if !attachedSeen[evidenceID] {
				attachedSeen[evidenceID] = true
				attachedIDs = append(attachedIDs, evidenceID)
			}
		}
	}
	evidenceFacts, err := reader.evidence.LoadTx(ctx, tx, incidentID, attachedIDs)
	if err != nil {
		return nil, fmt.Errorf("load Timeline evidence facts: %w", err)
	}
	// The Evidence reader orders facts for display; each record keeps that
	// order over its own attachments.
	evidencePosition := make(map[uuid.UUID]int, len(evidenceFacts))
	for index, fact := range evidenceFacts {
		evidencePosition[fact.RecordID] = index
	}

	result := make(map[uuid.UUID]workbookprojection.CollectionFacts, len(recordIDs))
	for _, recordID := range recordIDs {
		links := linkFacts[recordID]
		positions := make([]int, 0, len(links.AttachedEvidenceIDs))
		for _, evidenceID := range links.AttachedEvidenceIDs {
			if position, ok := evidencePosition[evidenceID]; ok {
				positions = append(positions, position)
			}
		}
		slices.Sort(positions)
		positions = slices.Compact(positions)
		attachedEvidence := make([]workbookprojection.EvidenceFact, len(positions))
		for index, position := range positions {
			fact := evidenceFacts[position]
			attachedEvidence[index] = workbookprojection.EvidenceFact{
				RecordID:       fact.RecordID,
				Title:          fact.Title,
				LifecycleState: fact.LifecycleState,
				UploadState:    fact.UploadState,
			}
		}
		resolvedLinks := make([]workbookprojection.LinkFact, len(links.ResolvedLinks))
		for index, fact := range links.ResolvedLinks {
			resolvedLinks[index] = workbookprojection.LinkFact{
				TargetRecordID: fact.TargetRecordID,
				LinkType:       fact.LinkType,
				Provenance:     fact.Provenance,
				Confidence:     fact.Confidence,
			}
		}
		tags := make([]workbookprojection.TagFact, len(links.Tags))
		for index, fact := range links.Tags {
			tags[index] = workbookprojection.TagFact{
				RecordTagID: fact.RecordTagID,
				TagName:     fact.TagName,
			}
		}
		recordMentions := mentions[recordID]
		if recordMentions == nil {
			recordMentions = []workbookprojection.MentionFact{}
		}
		result[recordID] = workbookprojection.CollectionFacts{
			Mentions:            recordMentions,
			ResolvedLinks:       resolvedLinks,
			Tags:                tags,
			AttachedEvidence:    attachedEvidence,
			ReplacementRecordID: links.ReplacementRecordID,
		}
	}
	return result, nil
}
