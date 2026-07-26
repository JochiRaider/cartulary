package timeline

import (
	"crypto/sha256"
	"encoding/json"
)

func canonicalChangeValue(change PatchChange) any {
	return derefString(change.TextValue)
}

func canonicalCollectionActionPayload(payload *CollectionActionPayload) any {
	if payload == nil {
		return nil
	}
	actions := make([]map[string]any, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		entry := map[string]any{"op": action.Op}
		if action.Op == "add_tag" && action.RawText != "" {
			entry["tag_name"] = action.RawText
		} else if action.RawText != "" {
			entry["raw_text"] = action.NormalizedText
		}
		if action.ItemRef != "" {
			entry["item_ref"] = action.ItemRef
		}
		if action.ResolvedRecord != nil {
			entry["resolved_record_id"] = action.ResolvedRecord.String()
		}
		if action.LinkedRecordID != nil {
			entry["linked_record_id"] = action.LinkedRecordID.String()
		}
		actions = append(actions, entry)
	}
	return map[string]any{
		"kind":    "collection_actions_v1",
		"actions": actions,
	}
}

func hashesEqual(left []byte, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func hashCanonicalValue(value any) []byte {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return append([]byte(nil), sum[:]...)
}
