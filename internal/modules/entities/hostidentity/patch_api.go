package hostidentity

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mutationadmission"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const maxPatchChanges = 32

func DecodePatchRequest(reader io.Reader) (PatchRequest, *mutationadmission.Failure) {
	raw, failure := decodeObject(reader)
	if failure != nil {
		return PatchRequest{}, failure
	}
	allowed := map[string]struct{}{
		"view_schema_id":   {},
		"base_row_version": {},
		"client_txn_id":    {},
		"changes":          {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return PatchRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request PatchRequest
	if value, ok := raw["view_schema_id"]; !ok {
		return PatchRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || !isEntityPatchSurface(request.ViewSchemaID) {
		return PatchRequest{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	if value, ok := raw["base_row_version"]; !ok {
		return PatchRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return PatchRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	value, ok := raw["changes"]
	if !ok {
		return PatchRequest{}, invalidMutationPayload("changes", "missing_required_field")
	}
	var rawChanges []json.RawMessage
	if err := json.Unmarshal(value, &rawChanges); err != nil {
		return PatchRequest{}, invalidMutationPayload("changes", "invalid_value")
	}
	if len(rawChanges) == 0 {
		return PatchRequest{}, invalidMutationPayload("changes", "empty_changes")
	}
	if len(rawChanges) > maxPatchChanges {
		return PatchRequest{}, mutationadmission.NewLimit(
			"changes",
			mutationadmission.ReasonChangeCountExceeded,
			len(rawChanges),
			maxPatchChanges,
			"",
		)
	}
	seen := map[string]struct{}{}
	for _, rawChange := range rawChanges {
		change, apiErr := decodeEntityPatchChange(request.ViewSchemaID, rawChange)
		if apiErr != nil {
			return PatchRequest{}, apiErr
		}
		if _, ok := seen[change.FieldKey]; ok {
			return PatchRequest{}, invalidMutationPayload("changes", "duplicate_field_key")
		}
		seen[change.FieldKey] = struct{}{}
		request.Changes = append(request.Changes, change)
	}
	slices.SortFunc(request.Changes, func(left PatchChange, right PatchChange) int {
		return strings.Compare(left.FieldKey, right.FieldKey)
	})
	return request, nil
}

func PatchRequestHash(request PatchRequest) []byte {
	changes := make([]map[string]any, 0, len(request.Changes))
	for _, change := range request.Changes {
		entry := map[string]any{"field_key": change.FieldKey}
		if change.CollectionActions != nil {
			entry["action_payload"] = canonicalAliasActions(change.CollectionActions)
		} else {
			entry["value"] = canonicalPatchValue(change.Value)
		}
		changes = append(changes, entry)
	}
	data, _ := json.Marshal(map[string]any{
		"view_schema_id":   request.ViewSchemaID,
		"base_row_version": request.BaseRowVersion,
		"changes":          changes,
	})
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}

func decodeEntityPatchChange(viewSchemaID string, raw json.RawMessage) (PatchChange, *mutationadmission.Failure) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}
	allowed := map[string]struct{}{"field_key": {}, "value": {}, "action_payload": {}}
	for key := range object {
		if _, ok := allowed[key]; !ok {
			return PatchChange{}, invalidMutationPayload("changes", "unknown_field")
		}
	}
	fieldValue, ok := object["field_key"]
	if !ok {
		return PatchChange{}, invalidMutationPayload("changes", "missing_field_key")
	}
	var fieldKey string
	if err := json.Unmarshal(fieldValue, &fieldKey); err != nil {
		return PatchChange{}, invalidMutationPayload("field_key", "invalid_value")
	}
	descriptor, ok := entityFields.lookup(viewSchemaID, fieldKey)
	if !ok || !descriptor.owner.Writable || descriptor.patch == entityFieldPatchNone {
		return PatchChange{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	field := descriptor.owner
	value, hasValue := object["value"]
	_, hasActionPayload := object["action_payload"]
	if hasValue == hasActionPayload {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}
	if hasActionPayload {
		if descriptor.patch != entityFieldPatchCollection {
			return PatchChange{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		actions, ok := decodeAliasPatchActionPayload(viewSchemaID, fieldKey, object["action_payload"])
		if !ok {
			return PatchChange{}, invalidMutationPayload(fieldKey, "invalid_value")
		}
		return PatchChange{FieldKey: fieldKey, CollectionActions: actions}, nil
	}
	if !hasValue || descriptor.patch == entityFieldPatchCollection {
		return PatchChange{}, invalidMutationPayload("value", "missing_required_field")
	}
	decoded, apiErr := decodeEntityPatchValue(fieldKey, field, value)
	if apiErr != nil {
		return PatchChange{}, apiErr
	}
	return PatchChange{FieldKey: fieldKey, Value: decoded}, nil
}

func decodeAliasPatchActionPayload(viewSchemaID string, fieldKey string, value json.RawMessage) ([]CollectionAction, bool) {
	descriptor, ok := entityFields.lookup(viewSchemaID, fieldKey)
	if !ok || descriptor.patch != entityFieldPatchCollection {
		return nil, false
	}
	var payload struct {
		Kind    string                       `json:"kind"`
		Actions []map[string]json.RawMessage `json:"actions"`
	}
	if err := json.Unmarshal(value, &payload); err != nil || payload.Kind != "collection_actions_v1" || len(payload.Actions) == 0 || len(payload.Actions) > maxCollectionActions {
		return nil, false
	}
	actions := make([]CollectionAction, 0, len(payload.Actions))
	for _, rawAction := range payload.Actions {
		var op string
		if err := json.Unmarshal(rawAction["op"], &op); err != nil {
			return nil, false
		}
		switch op {
		case "add_alias":
			if len(rawAction) != 2 {
				return nil, false
			}
			var rawText string
			if err := json.Unmarshal(rawAction["alias_text"], &rawText); err != nil {
				return nil, false
			}
			normalized, ok := fieldnorm.NormalizeAliasText(rawText)
			if !ok {
				return nil, false
			}
			actions = append(actions, CollectionAction{Op: op, RawText: normalized, NormalizedText: normalized})
		case "remove_alias":
			if len(rawAction) != 2 {
				return nil, false
			}
			var itemRef string
			if err := json.Unmarshal(rawAction["item_ref"], &itemRef); err != nil {
				return nil, false
			}
			if _, err := parseEntityAliasItemRef(itemRef); err != nil {
				return nil, false
			}
			actions = append(actions, CollectionAction{Op: op, ItemRef: itemRef})
		default:
			return nil, false
		}
	}
	return actions, true
}

func canonicalAliasActions(actions []CollectionAction) map[string]any {
	values := make([]map[string]any, 0, len(actions))
	for _, action := range actions {
		value := map[string]any{"op": action.Op}
		if action.Op == "add_alias" {
			value["alias_text"] = action.NormalizedText
		} else {
			value["item_ref"] = action.ItemRef
		}
		values = append(values, value)
	}
	return map[string]any{"kind": "collection_actions_v1", "actions": values}
}

func decodeEntityPatchValue(fieldKey string, field viewschema.Field, value json.RawMessage) (*string, *mutationadmission.Failure) {
	if string(value) == "null" {
		if field.Clearable {
			return nil, nil
		}
		return nil, invalidMutationPayload(fieldKey, "field_not_nullable")
	}
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return nil, invalidMutationPayload(fieldKey, "invalid_value")
	}
	normalized, ok := fieldnorm.NormalizeLine(raw)
	if !ok {
		return nil, invalidMutationPayload(fieldKey, "invalid_value")
	}
	return &normalized, nil
}

func canonicalPatchValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func isEntityPatchSurface(viewSchemaID string) bool {
	return viewSchemaID == entitycontract.HostsViewSchemaID || viewSchemaID == entitycontract.IdentitiesViewSchemaID
}
