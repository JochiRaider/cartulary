package workbookprojection

import (
	"encoding/json"

	"github.com/google/uuid"
)

type MentionFact struct {
	MentionID        uuid.UUID
	EntityType       string
	SourceFieldKey   string
	RawText          string
	ResolutionStatus string
	RowVersion       int64
	ResolvedRecordID *uuid.UUID
	ResolutionMethod *string
	MatchedAliasText *string
}

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

type EvidenceFact struct {
	RecordID       uuid.UUID
	Title          string
	LifecycleState string
	UploadState    string
}

type CollectionFacts struct {
	Mentions            []MentionFact
	ResolvedLinks       []LinkFact
	Tags                []TagFact
	AttachedEvidence    []EvidenceFact
	ReplacementRecordID *uuid.UUID
}

type MentionRef struct {
	ItemRef           string     `json:"item_ref"`
	ItemKind          string     `json:"item_kind"`
	EntityType        string     `json:"entity_type"`
	DisplayText       string     `json:"display_text"`
	RawText           string     `json:"raw_text"`
	MentionRowVersion int64      `json:"mention_row_version"`
	ResolvedRecordID  *uuid.UUID `json:"resolved_record_id,omitempty"`
	ResolutionMethod  *string    `json:"resolution_method,omitempty"`
	AutoResolved      bool       `json:"auto_resolved,omitempty"`
	Provenance        *string    `json:"provenance,omitempty"`
	Confidence        *int       `json:"confidence,omitempty"`
	MatchedAliasText  *string    `json:"matched_alias_text,omitempty"`
}

func (item MentionRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(mentionRefToMap(item))
}

type EvidenceRef struct {
	ItemRef        string    `json:"item_ref"`
	ItemKind       string    `json:"item_kind"`
	DisplayText    string    `json:"display_text"`
	LinkedRecordID uuid.UUID `json:"linked_record_id"`
}

type TagRef struct {
	ItemRef     string    `json:"item_ref"`
	ItemKind    string    `json:"item_kind"`
	DisplayText string    `json:"display_text"`
	TagID       uuid.UUID `json:"tag_id"`
}

func ApplyCollectionFacts(record *DerivedRecord, facts CollectionFacts) {
	if record == nil {
		return
	}
	record.ReplacementRecordID = cloneUUIDPointer(facts.ReplacementRecordID)
	record.HostRefs = make([]MentionRef, 0)
	record.IdentityRefs = make([]MentionRef, 0)
	record.AttachedEvidence = make([]EvidenceRef, 0, len(facts.AttachedEvidence))
	record.Tags = make([]TagRef, 0, len(facts.Tags))
	record.HasUnresolvedMentions = false

	for _, mention := range facts.Mentions {
		item := MentionRef{
			ItemRef:           "entity_mention:" + mention.MentionID.String(),
			EntityType:        mention.EntityType,
			DisplayText:       mention.RawText,
			RawText:           mention.RawText,
			MentionRowVersion: mention.RowVersion,
		}
		if mention.ResolutionStatus == "resolved" && mention.ResolvedRecordID != nil {
			item.ItemKind = "resolved_ref"
			item.ResolvedRecordID = cloneUUIDPointer(mention.ResolvedRecordID)
			item.ResolutionMethod = cloneStringPointer(mention.ResolutionMethod)
			item.MatchedAliasText = cloneStringPointer(mention.MatchedAliasText)
			item.AutoResolved = mention.ResolutionMethod != nil && *mention.ResolutionMethod == "auto_match"
			if linkType, ok := collectionLinkType(mention.SourceFieldKey); ok {
				for _, link := range facts.ResolvedLinks {
					if link.TargetRecordID == *mention.ResolvedRecordID && link.LinkType == linkType {
						provenance := link.Provenance
						item.Provenance = &provenance
						if link.Confidence != nil {
							confidence := *link.Confidence
							item.Confidence = &confidence
						}
						break
					}
				}
			}
		} else {
			item.ItemKind = "unresolved_mention"
			record.HasUnresolvedMentions = true
		}
		switch mention.EntityType {
		case "host":
			record.HostRefs = append(record.HostRefs, item)
		case "identity":
			record.IdentityRefs = append(record.IdentityRefs, item)
		}
	}

	for _, tag := range facts.Tags {
		record.Tags = append(record.Tags, TagRef{
			ItemRef:     "record_tag:" + record.RecordID.String() + ":" + tag.RecordTagID.String(),
			ItemKind:    "tag",
			DisplayText: tag.TagName,
			TagID:       tag.RecordTagID,
		})
	}

	record.EvidenceCount = 0
	for _, evidence := range facts.AttachedEvidence {
		record.AttachedEvidence = append(record.AttachedEvidence, EvidenceRef{
			ItemRef:        "record_ref:" + evidence.RecordID.String(),
			ItemKind:       "record_ref",
			DisplayText:    evidence.Title,
			LinkedRecordID: evidence.RecordID,
		})
		if evidence.UploadState == "available" && (evidence.LifecycleState == "available" || evidence.LifecycleState == "released") {
			record.EvidenceCount++
		}
	}
	record.HasEvidence = record.EvidenceCount > 0
}

func collectionLinkType(sourceFieldKey string) (string, bool) {
	switch sourceFieldKey {
	case "timeline.host_refs":
		return "observed_on_host", true
	case "timeline.identity_refs":
		return "observed_as_identity", true
	default:
		return "", false
	}
}

func mentionRefsToMaps(items []MentionRef) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, mentionRefToMap(item))
	}
	return result
}

func mentionRefToMap(item MentionRef) map[string]any {
	value := map[string]any{
		"item_ref":            item.ItemRef,
		"item_kind":           item.ItemKind,
		"entity_type":         item.EntityType,
		"display_text":        item.DisplayText,
		"raw_text":            item.RawText,
		"mention_row_version": item.MentionRowVersion,
	}
	if item.ResolvedRecordID != nil {
		value["resolved_record_id"] = item.ResolvedRecordID.String()
	}
	if item.ResolutionMethod != nil {
		value["resolution_method"] = *item.ResolutionMethod
	}
	if item.AutoResolved {
		value["auto_resolved"] = true
	}
	if item.Provenance != nil {
		value["provenance"] = *item.Provenance
		if item.Confidence == nil {
			value["confidence"] = nil
		}
	}
	if item.Confidence != nil {
		value["confidence"] = *item.Confidence
	}
	if item.MatchedAliasText != nil {
		value["matched_alias_text"] = *item.MatchedAliasText
	}
	return value
}

func evidenceRefsToMaps(items []EvidenceRef) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"item_ref":         item.ItemRef,
			"item_kind":        item.ItemKind,
			"display_text":     item.DisplayText,
			"linked_record_id": item.LinkedRecordID.String(),
		})
	}
	return result
}

func tagRefsToMaps(items []TagRef) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"item_ref":     item.ItemRef,
			"item_kind":    item.ItemKind,
			"display_text": item.DisplayText,
			"tag_id":       item.TagID.String(),
		})
	}
	return result
}
