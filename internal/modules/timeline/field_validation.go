package timeline

import (
	"encoding/json"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

func normalizeFieldTextValue(fieldKey string, value json.RawMessage) (*string, bool) {
	if !mutationpolicy.IsDirectWritableField(fieldKey) {
		return nil, false
	}
	if string(value) == "null" {
		return nil, true
	}
	var rawValue string
	if err := json.Unmarshal(value, &rawValue); err != nil || !mutationpolicy.IsValidVisibleText(rawValue) {
		return nil, false
	}
	return &rawValue, true
}

func normalizeCollectionToken(fieldKey string, rawText string) (string, bool) {
	if isTimelineTagCollection(fieldKey) {
		_, normalized, ok := fieldnorm.NormalizeTagLabel(rawText)
		return normalized, ok
	}
	return fieldnorm.NormalizeMentionToken(rawText)
}

func timelineCollectionPolicy(fieldKey string) (CollectionPolicy, bool) {
	policy, ok := LookupCollectionPolicy(fieldKey)
	if !ok {
		return CollectionPolicy{}, false
	}
	if policy.Family == CollectionFamilyMentionOrigin {
		return policy, true
	}
	if policy.AllowsLinksCollectionMutation() &&
		(fieldKey == "timeline.tags" || fieldKey == "timeline.attached_evidence_ids") {
		return policy, true
	}
	return CollectionPolicy{}, false
}

func isTimelineMentionCollection(fieldKey string) bool {
	policy, ok := timelineCollectionPolicy(fieldKey)
	return ok && policy.Family == CollectionFamilyMentionOrigin
}

func isTimelineTagCollection(fieldKey string) bool {
	policy, ok := timelineCollectionPolicy(fieldKey)
	return ok && policy.Family == CollectionFamilyRecordTag
}

func isTimelineAttachedEvidenceCollection(fieldKey string) bool {
	policy, ok := timelineCollectionPolicy(fieldKey)
	return ok && policy.Family == CollectionFamilyRecordRef && policy.LinkType == "attached_evidence"
}
