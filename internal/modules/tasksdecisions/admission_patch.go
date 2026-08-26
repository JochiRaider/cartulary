package tasksdecisions

import (
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func AdmitPatchJSON(reader io.Reader) (PatchRequest, *AdmissionFailure) {
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return PatchRequest{}, invalidAdmission("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"view_schema_id": {}, "base_row_version": {}, "client_txn_id": {}, "changes": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return PatchRequest{}, invalidAdmission(key, "unknown_field")
		}
	}
	var request PatchRequest
	if value, present := raw["view_schema_id"]; !present {
		return PatchRequest{}, invalidAdmission("view_schema_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ViewSchemaID) != nil || !isMutationView(request.ViewSchemaID) {
		return PatchRequest{}, invalidAdmission("view_schema_id", "invalid_view_schema_id")
	}
	if value, present := raw["base_row_version"]; !present {
		return PatchRequest{}, invalidAdmission("base_row_version", "missing_required_field")
	} else if json.Unmarshal(value, &request.BaseRowVersion) != nil || request.BaseRowVersion < 1 {
		return PatchRequest{}, invalidAdmission("base_row_version", "invalid_base_row_version")
	}
	if value, present := raw["client_txn_id"]; !present {
		return PatchRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return PatchRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	}
	var rawChanges []json.RawMessage
	if value, present := raw["changes"]; !present {
		return PatchRequest{}, invalidAdmission("changes", "missing_required_field")
	} else if json.Unmarshal(value, &rawChanges) != nil {
		return PatchRequest{}, invalidAdmission("changes", "invalid_value")
	}
	if len(rawChanges) == 0 {
		return PatchRequest{}, invalidAdmission("changes", "empty_changes")
	}
	if len(rawChanges) > maxMutationPatchChanges {
		return PatchRequest{}, invalidCountAdmission(
			"changes", "change_count_exceeded", len(rawChanges), maxMutationPatchChanges, "",
		)
	}
	seen := make(map[string]struct{}, len(rawChanges))
	for _, rawChange := range rawChanges {
		change, apiErr := decodeMutationPatchChange(request.ViewSchemaID, rawChange)
		if apiErr != nil {
			return PatchRequest{}, apiErr
		}
		if _, duplicate := seen[change.FieldKey]; duplicate {
			return PatchRequest{}, invalidAdmission("changes", "duplicate_field_key")
		}
		seen[change.FieldKey] = struct{}{}
		request.Changes = append(request.Changes, change)
	}
	slices.SortFunc(request.Changes, func(left, right PatchChange) int {
		return strings.Compare(left.FieldKey, right.FieldKey)
	})
	return request, nil
}

func decodeMutationPatchChange(viewSchemaID string, raw json.RawMessage) (PatchChange, *AdmissionFailure) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return PatchChange{}, invalidAdmission("changes", "invalid_change")
	}
	allowed := map[string]struct{}{"field_key": {}, "value": {}, "action_payload": {}}
	for key := range object {
		if _, admitted := allowed[key]; !admitted {
			return PatchChange{}, invalidAdmission("changes", "unknown_field")
		}
	}
	var fieldKey string
	if value, present := object["field_key"]; !present {
		return PatchChange{}, invalidAdmission("changes", "missing_field_key")
	} else if json.Unmarshal(value, &fieldKey) != nil {
		return PatchChange{}, invalidAdmission("field_key", "invalid_value")
	}
	field, ok := viewschema.LookupField(viewSchemaID, fieldKey)
	if !ok || !field.Writable {
		return PatchChange{}, invalidAdmission(fieldKey, "unsupported_field_key")
	}
	value, hasValue := object["value"]
	actionPayload, hasActionPayload := object["action_payload"]
	if hasValue == hasActionPayload {
		return PatchChange{}, invalidAdmission("changes", "invalid_change")
	}
	change := PatchChange{FieldKey: fieldKey}
	if field.ConflictResolutionClass == "collection_review" {
		if !hasActionPayload {
			return PatchChange{}, invalidAdmission("action_payload", "missing_required_field")
		}
		payload, apiErr := decodeMutationCollectionPayload(fieldKey, actionPayload)
		if apiErr != nil {
			return PatchChange{}, apiErr
		}
		change.Collection, change.CanonicalValue = &payload, canonicalMutationCollection(payload)
		return change, nil
	}
	if !hasValue {
		return PatchChange{}, invalidAdmission("value", "missing_required_field")
	}
	direct, canonical, apiErr := decodeMutationValue(fieldKey, field, value, true)
	if apiErr != nil {
		return PatchChange{}, apiErr
	}
	change.Value, change.CanonicalValue = &direct, canonical
	return change, nil
}
