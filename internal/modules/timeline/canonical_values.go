package timeline

import (
	"encoding/json"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/valuecodec"
)

func (change PatchChange) CanonicalValue() any {
	if change.ActionPayload != nil {
		return change.ActionPayload.CanonicalValue()
	}
	return valuecodec.OptionalString(change.TextValue)
}

func (payload *CollectionActionPayload) CanonicalValue() any {
	if payload == nil {
		return nil
	}
	actions := make([]map[string]any, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		actions = append(actions, action.CanonicalValue())
	}
	return map[string]any{
		"kind":    "collection_actions_v1",
		"actions": actions,
	}
}

func (action CollectionAction) CanonicalValue() map[string]any {
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
	return entry
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
