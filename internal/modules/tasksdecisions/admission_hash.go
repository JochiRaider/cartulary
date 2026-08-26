package tasksdecisions

import (
	"crypto/sha256"
	"encoding/json"
	"slices"
	"time"
)

func CreateRequestHash(request CreateRequest) []byte {
	return hashMutationPayload(map[string]any{
		"view_schema_id": request.ViewSchemaID,
		"values":         canonicalMutationValues(request.Values),
		"collection_ops": canonicalMutationCollections(request.Collections),
		"create_inputs":  map[string]any{},
	})
}

func PatchRequestHash(request PatchRequest) []byte {
	changes := make([]map[string]any, 0, len(request.Changes))
	for _, change := range request.Changes {
		changes = append(changes, map[string]any{"field_key": change.FieldKey, "value": change.CanonicalValue})
	}
	return hashMutationPayload(map[string]any{
		"view_schema_id": request.ViewSchemaID, "base_row_version": request.BaseRowVersion, "changes": changes,
	})
}

func ConflictResolveRequestHash(claims ConflictClaims, request ConflictResolveRequest) []byte {
	return hashMutationPayload(map[string]any{
		"conflict_token": request.ConflictToken, "resolution_kind": request.ResolutionKind,
		"record_id": claims.RecordID, "view_schema_id": claims.ViewSchemaID,
		"field_key": claims.FieldKey, "current_row_version": claims.CurrentRowVersion,
		"resolved_value": request.CanonicalValue,
	})
}

func SupersedeRequestHash(request SupersedeRequest) []byte {
	return hashMutationPayload(map[string]any{
		"base_row_version": request.BaseRowVersion, "reason": request.Reason,
		"replacement_record_id": request.ReplacementRecordID,
	})
}

func canonicalMutationValues(values map[string]FieldValue) map[string]any {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := map[string]any{}
	for _, key := range keys {
		if value := canonicalMutationValue(values[key]); value != nil {
			result[key] = value
		}
	}
	return result
}

func canonicalMutationValue(value FieldValue) any {
	switch {
	case value.Timestamp != nil:
		return value.Timestamp.UTC().Format(time.RFC3339Nano)
	case value.UUID != nil:
		return value.UUID.String()
	case value.Text != nil:
		return *value.Text
	case value.Number != nil:
		return *value.Number
	case value.Bool != nil:
		return *value.Bool
	default:
		return nil
	}
}

func canonicalMutationCollections(collections map[string]CollectionActionPayload) map[string]any {
	keys := make([]string, 0, len(collections))
	for key := range collections {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := map[string]any{}
	for _, key := range keys {
		result[key] = canonicalMutationCollection(collections[key])
	}
	return result
}

func canonicalMutationCollection(payload CollectionActionPayload) map[string]any {
	actions := make([]map[string]any, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		entry := map[string]any{"op": action.Op}
		if action.LinkedRecordID != nil {
			entry["linked_record_id"] = action.LinkedRecordID.String()
		}
		if action.ItemRef != "" {
			entry["item_ref"] = action.ItemRef
		}
		actions = append(actions, entry)
	}
	return map[string]any{"kind": "collection_actions_v1", "actions": actions}
}

func hashMutationPayload(value any) []byte {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return append([]byte(nil), sum[:]...)
}
