package timelinefactassembly

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/collectionfacts"
)

type linkReader struct {
	source links.FactReader
}

func NewLinkReader() collectionfacts.LinkReader {
	return linkReader{source: links.FactReader{}}
}

func (reader linkReader) LoadTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) (collectionfacts.LinkFacts, error) {
	facts, err := reader.source.LoadRecordTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return collectionfacts.LinkFacts{}, err
	}
	outbound := make([]links.RecordLinkFact, 0)
	tags := make([]links.RecordTagFact, 0)
	var replacement *links.RecordLinkFact
	for _, fact := range facts.RecordLinks {
		if fact.SrcRecordID == recordID {
			switch fact.LinkType {
			case links.LinkTypeObservedOnHost, links.LinkTypeObservedAsIdentity, links.LinkTypeAttachedEvidence:
				outbound = append(outbound, fact)
			}
		}
		if fact.DstRecordID == recordID && fact.LinkType == links.LinkTypeSupersedes &&
			(replacement == nil || replacement.CreatedAt.Before(fact.CreatedAt) ||
				(replacement.CreatedAt.Equal(fact.CreatedAt) && replacement.RecordLinkID.String() < fact.RecordLinkID.String())) {
			copy := fact
			replacement = &copy
		}
	}
	for _, fact := range facts.RecordTags {
		if fact.RecordID == recordID {
			tags = append(tags, fact)
		}
	}
	sort.Slice(outbound, func(i, j int) bool {
		if outbound[i].LinkType != outbound[j].LinkType {
			return outbound[i].LinkType < outbound[j].LinkType
		}
		if !outbound[i].CreatedAt.Equal(outbound[j].CreatedAt) {
			return outbound[i].CreatedAt.Before(outbound[j].CreatedAt)
		}
		return outbound[i].RecordLinkID.String() < outbound[j].RecordLinkID.String()
	})
	sort.Slice(tags, func(i, j int) bool {
		if tags[i].NormalizedTagName != tags[j].NormalizedTagName {
			return tags[i].NormalizedTagName < tags[j].NormalizedTagName
		}
		return tags[i].RecordTagID.String() < tags[j].RecordTagID.String()
	})
	result := collectionfacts.LinkFacts{
		ResolvedLinks:       []collectionfacts.LinkFact{},
		Tags:                make([]collectionfacts.TagFact, len(tags)),
		AttachedEvidenceIDs: []uuid.UUID{},
	}
	if replacement != nil {
		replacementRecordID := replacement.SrcRecordID
		result.ReplacementRecordID = &replacementRecordID
	}
	for _, fact := range outbound {
		if fact.LinkType == links.LinkTypeAttachedEvidence {
			result.AttachedEvidenceIDs = append(result.AttachedEvidenceIDs, fact.DstRecordID)
			continue
		}
		mapped := collectionfacts.LinkFact{
			TargetRecordID: fact.DstRecordID,
			LinkType:       fact.LinkType.String(),
			Provenance:     fact.Provenance.String(),
		}
		if fact.Confidence != nil {
			confidence := *fact.Confidence
			mapped.Confidence = &confidence
		}
		result.ResolvedLinks = append(result.ResolvedLinks, mapped)
	}
	for index, fact := range tags {
		result.Tags[index] = collectionfacts.TagFact{
			RecordTagID: fact.RecordTagID,
			TagName:     fact.TagName,
		}
	}
	return result, nil
}
