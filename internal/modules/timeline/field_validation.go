package timeline

import (
	"encoding/json"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
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

func isTimelineMentionCollection(fieldKey string) bool {
	policy, ok := LookupCollectionPolicy(fieldKey)
	return ok && policy.Family == CollectionFamilyMentionOrigin
}

func isTimelineTagCollection(fieldKey string) bool {
	policy, ok := LookupCollectionPolicy(fieldKey)
	return ok && policy.Family == CollectionFamilyRecordTag
}

func isTimelineAttachedEvidenceCollection(fieldKey string) bool {
	policy, ok := LookupCollectionPolicy(fieldKey)
	return ok && policy.IsAttachedEvidence()
}
