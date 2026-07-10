package hostidentity

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const maxPatchChanges = 32

func DecodePatchRequest(reader io.Reader) (PatchRequest, *httpapi.APIError) {
	raw, apiErr := decodePatchObject(reader)
	if apiErr != nil {
		return PatchRequest{}, apiErr
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
		return PatchRequest{}, invalidMutationPayloadWithDetails("changes", "change_count_exceeded", map[string]any{
			"requested_count": len(rawChanges),
			"max_count":       maxPatchChanges,
		})
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
		changes = append(changes, map[string]any{
			"field_key": change.FieldKey,
			"value":     canonicalPatchValue(change.Value),
		})
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

func decodeEntityPatchChange(viewSchemaID string, raw json.RawMessage) (PatchChange, *httpapi.APIError) {
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
	field, ok := viewschema.LookupField(viewSchemaID, fieldKey)
	if !ok || !field.Writable || !isEntityDirectPatchField(viewSchemaID, fieldKey) {
		return PatchChange{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	value, hasValue := object["value"]
	_, hasActionPayload := object["action_payload"]
	if hasValue == hasActionPayload {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}
	if !hasValue {
		return PatchChange{}, invalidMutationPayload("value", "missing_required_field")
	}
	decoded, apiErr := decodeEntityPatchValue(fieldKey, field, value)
	if apiErr != nil {
		return PatchChange{}, apiErr
	}
	return PatchChange{FieldKey: fieldKey, Value: decoded}, nil
}

func decodeEntityPatchValue(fieldKey string, field viewschema.Field, value json.RawMessage) (*string, *httpapi.APIError) {
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
	return viewSchemaID == HostsViewSchemaID || viewSchemaID == IdentitiesViewSchemaID
}

func isEntityDirectPatchField(viewSchemaID string, fieldKey string) bool {
	switch viewSchemaID {
	case HostsViewSchemaID:
		switch fieldKey {
		case "host.display_name", "host.hostname", "host.aad_device_id", "host.fqdn",
			"host.location", "host.os_platform", "host.business_owner", "host.criticality", "host.containment_status":
			return true
		default:
			return false
		}
	case IdentitiesViewSchemaID:
		switch fieldKey {
		case "identity.display_name", "identity.aad_object_id", "identity.sid", "identity.upn", "identity.email", "identity.sam_account_name",
			"identity.privilege_level", "identity.mfa_state", "identity.reset_status":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func decodePatchObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
}

func invalidMutationPayloadWithDetails(field string, reasonCode string, extra map[string]any) *httpapi.APIError {
	apiErr := invalidMutationPayload(field, reasonCode)
	for key, value := range extra {
		apiErr.Details[key] = value
	}
	return apiErr
}
