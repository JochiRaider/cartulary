package artifacts

import (
	"crypto/sha256"
	"encoding/json"
	"slices"
	"time"

	"github.com/google/uuid"
)

func createAdmissionHash(request createRequest) [sha256.Size]byte {
	return hashArtifactMutationPayload(map[string]any{
		"view_schema_id": request.ViewSchemaID,
		"values":         canonicalArtifactValues(request.Values),
		"collection_ops": canonicalArtifactCollections(request.Collections),
		"create_inputs":  map[string]any{},
	})
}

func patchAdmissionHash(request patchRequest) [sha256.Size]byte {
	changes := make([]map[string]any, 0, len(request.Changes))
	for _, change := range request.Changes {
		changes = append(changes, map[string]any{"field_key": change.FieldKey, "value": change.CanonicalValue})
	}
	return hashArtifactMutationPayload(map[string]any{
		"view_schema_id": request.ViewSchemaID, "base_row_version": request.BaseRowVersion, "changes": changes,
	})
}

func conflictResolutionAdmissionHash(context ConflictAdmissionContext, request conflictResolveRequest) [sha256.Size]byte {
	return hashArtifactMutationPayload(map[string]any{
		"conflict_token": request.ConflictToken, "resolution_kind": request.ResolutionKind,
		"record_id": context.RecordID, "view_schema_id": context.ViewSchemaID,
		"field_key": context.FieldKey, "current_row_version": context.CurrentRowVersion,
		"resolved_value": request.CanonicalValue,
	})
}

func contextualNoteAdmissionHash(sourceRecordID uuid.UUID, request createRequest) []byte {
	hash := hashArtifactMutationPayload(map[string]any{
		"source_record_id": sourceRecordID.String(), "view_schema_id": NotesViewSchemaID,
		"values": canonicalArtifactValues(request.Values), "collection_ops": canonicalArtifactCollections(request.Collections),
	})
	return append([]byte(nil), hash[:]...)
}

func canonicalArtifactValues(values map[string]fieldValue) map[string]any {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := map[string]any{}
	for _, key := range keys {
		value := canonicalArtifactValue(values[key])
		if value != nil {
			result[key] = value
		}
	}
	return result
}

func canonicalArtifactValue(value fieldValue) any {
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

func canonicalArtifactCollections(collections map[string]collectionActionPayload) map[string]any {
	keys := make([]string, 0, len(collections))
	for key := range collections {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := map[string]any{}
	for _, key := range keys {
		result[key] = canonicalArtifactCollectionPayload(collections[key])
	}
	return result
}

func canonicalArtifactCollectionPayload(payload collectionActionPayload) map[string]any {
	actions := make([]map[string]any, 0, len(payload.Actions))
	for _, action := range payload.Actions {
		entry := map[string]any{"op": action.Op}
		if action.LinkedRecordID != nil {
			entry["linked_record_id"] = action.LinkedRecordID.String()
		}
		if action.PartyID != nil {
			entry["party_id"] = action.PartyID.String()
		}
		if action.ItemRef != "" {
			entry["item_ref"] = action.ItemRef
		}
		if action.Op == "add_tag" && action.RawText != "" {
			entry["tag_name"] = action.RawText
		}
		if action.Op == "add_token" && action.NormalizedText != "" {
			entry["raw_text"] = action.NormalizedText
		}
		if action.Op == "add_risk_ref" && action.NormalizedText != "" {
			entry["risk_ref_text"] = action.NormalizedText
		}
		actions = append(actions, entry)
	}
	return map[string]any{"kind": "collection_actions_v1", "actions": actions}
}

func hashArtifactMutationPayload(payload any) [sha256.Size]byte {
	data, _ := json.Marshal(payload)
	return sha256.Sum256(data)
}
