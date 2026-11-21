package timelinefactassembly

import (
	"context"
	"slices"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	linkfacts "github.com/JochiRaider/cartulary/internal/modules/links/facts"
	linkvocabulary "github.com/JochiRaider/cartulary/internal/modules/links/vocabulary"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/collectionfacts"
)

type linkReader struct {
	source linkfacts.FactReader
}

func NewLinkReader() collectionfacts.LinkReader {
	return linkReader{source: linkfacts.FactReader{}}
}

func (reader linkReader) LoadTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordIDs []uuid.UUID,
) (map[uuid.UUID]collectionfacts.LinkFacts, error) {
	facts, err := reader.source.LoadRecordsTx(ctx, tx, incidentID, recordIDs)
	if err != nil {
		return nil, err
	}
	outbound := make(map[uuid.UUID][]linkfacts.RecordLinkFact)
	replacements := make(map[uuid.UUID]linkfacts.RecordLinkFact)
	for _, fact := range facts.RecordLinks {
		switch fact.LinkType {
		case linkvocabulary.LinkTypeObservedOnHost, linkvocabulary.LinkTypeObservedAsIdentity, linkvocabulary.LinkTypeAttachedEvidence:
			outbound[fact.SrcRecordID] = append(outbound[fact.SrcRecordID], fact)
		case linkvocabulary.LinkTypeSupersedes:
			replacement, found := replacements[fact.DstRecordID]
			if !found || replacement.CreatedAt.Before(fact.CreatedAt) ||
				(replacement.CreatedAt.Equal(fact.CreatedAt) && replacement.RecordLinkID.String() < fact.RecordLinkID.String()) {
				replacements[fact.DstRecordID] = fact
			}
		}
	}
	tags := make(map[uuid.UUID][]linkfacts.RecordTagFact)
	for _, fact := range facts.RecordTags {
		tags[fact.RecordID] = append(tags[fact.RecordID], fact)
	}
	result := make(map[uuid.UUID]collectionfacts.LinkFacts, len(recordIDs))
	for _, recordID := range recordIDs {
		var replacement *linkfacts.RecordLinkFact
		if fact, found := replacements[recordID]; found {
			replacement = &fact
		}
		result[recordID] = recordLinkFacts(outbound[recordID], tags[recordID], replacement)
	}
	return result, nil
}

// recordLinkFacts maps one record's outbound collection links, tags and
// newest replacing supersession to Timeline collection facts.
func recordLinkFacts(
	outbound []linkfacts.RecordLinkFact,
	tags []linkfacts.RecordTagFact,
	replacement *linkfacts.RecordLinkFact,
) collectionfacts.LinkFacts {
	outbound = slices.Clone(outbound)
	tags = slices.Clone(tags)
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
		if fact.LinkType == linkvocabulary.LinkTypeAttachedEvidence {
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
	return result
}
