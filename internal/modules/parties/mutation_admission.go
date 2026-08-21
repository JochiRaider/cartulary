package parties

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const maxPatchChanges = 32

type ConflictClaims struct {
	RecordID          uuid.UUID
	ViewSchemaID      string
	FieldKey          string
	CurrentRowVersion int64
}

type ConflictResolveRequest struct {
	ConflictToken  string
	ResolutionKind string
	ClientTxnID    string
	Patch          *PatchRequest
	CanonicalValue any
}

// DecodeCreateRequest admits the fixed Party create surface. Party field
// policy and normalization stay beside the Party mutation owner.
func DecodeCreateRequest(reader io.Reader) (CreateRequest, *httpapi.APIError) {
	schema, ok := viewschema.Lookup(ViewSchemaID)
	if !ok || !schema.CreateCapable {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return CreateRequest{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable || field.CreateWritable {
			allowed[fieldKey] = struct{}{}
		}
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return CreateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	request := CreateRequest{ViewSchemaID: ViewSchemaID, Values: map[string]FieldValue{}}
	if value, present := raw["client_txn_id"]; !present {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	for fieldKey, field := range schema.Fields() {
		value, present := raw[fieldKey]
		if !present {
			continue
		}
		admitted, _, apiErr := decodePartyValue(fieldKey, field, value, false)
		if apiErr != nil {
			return CreateRequest{}, apiErr
		}
		request.Values[fieldKey] = admitted
	}
	if err := ValidateCreateParams(CreateParams{Values: request.Values}); err != nil {
		var validation *ValidationError
		if errors.As(err, &validation) {
			return CreateRequest{}, invalidMutationPayload(validation.Field, validation.ReasonCode)
		}
		return CreateRequest{}, invalidMutationPayload("payload", "invalid_value")
	}
	return request, nil
}

// DecodePatchRequest admits only Party patch fields and returns a canonical,
// field-key-sorted owner request.
func DecodePatchRequest(reader io.Reader) (PatchRequest, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return PatchRequest{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"view_schema_id": {}, "base_row_version": {}, "client_txn_id": {}, "changes": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return PatchRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	var request PatchRequest
	if value, present := raw["view_schema_id"]; !present {
		return PatchRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ViewSchemaID) != nil || request.ViewSchemaID != ViewSchemaID {
		return PatchRequest{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	if value, present := raw["base_row_version"]; !present {
		return PatchRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if json.Unmarshal(value, &request.BaseRowVersion) != nil || request.BaseRowVersion < 1 {
		return PatchRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, present := raw["client_txn_id"]; !present {
		return PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return PatchRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	rawChanges, apiErr := decodeRawChanges(raw["changes"])
	if apiErr != nil {
		return PatchRequest{}, apiErr
	}
	seen := make(map[string]struct{}, len(rawChanges))
	for _, rawChange := range rawChanges {
		change, apiErr := decodePartyPatchChange(rawChange)
		if apiErr != nil {
			return PatchRequest{}, apiErr
		}
		if _, duplicate := seen[change.FieldKey]; duplicate {
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

func DecodeConflictResolveRequest(
	reader io.Reader,
	token string,
	claims ConflictClaims,
) (ConflictResolveRequest, *httpapi.APIError) {
	if claims.RecordID == uuid.Nil || claims.ViewSchemaID != ViewSchemaID || claims.CurrentRowVersion < 1 {
		return ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "invalid_value")
	}
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return ConflictResolveRequest{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"conflict_token": {}, "resolution_kind": {}, "client_txn_id": {}, "resolved_value": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return ConflictResolveRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	request := ConflictResolveRequest{ConflictToken: token}
	if value, present := raw["conflict_token"]; !present {
		return ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "missing_required_field")
	} else if json.Unmarshal(value, &request.ConflictToken) != nil || request.ConflictToken != token {
		return ConflictResolveRequest{}, invalidMutationPayload("conflict_token", "invalid_value")
	}
	if value, present := raw["resolution_kind"]; !present {
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "missing_required_field")
	} else if json.Unmarshal(value, &request.ResolutionKind) != nil {
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	switch request.ResolutionKind {
	case "keep_saved", "use_unsaved", "merged_value":
	default:
		return ConflictResolveRequest{}, invalidMutationPayload("resolution_kind", "invalid_value")
	}
	if value, present := raw["client_txn_id"]; !present {
		return ConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ConflictResolveRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	resolvedValue, present := raw["resolved_value"]
	if request.ResolutionKind == "keep_saved" {
		if present {
			return ConflictResolveRequest{}, invalidMutationPayload("resolved_value", "forbidden_field")
		}
		return request, nil
	}
	if !present {
		return ConflictResolveRequest{}, invalidMutationPayload("resolved_value", "missing_required_field")
	}
	field, ok := viewschema.LookupField(ViewSchemaID, claims.FieldKey)
	if !ok || !field.Writable {
		return ConflictResolveRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
	}
	value, canonical, apiErr := decodePartyValue(claims.FieldKey, field, resolvedValue, true)
	if apiErr != nil {
		return ConflictResolveRequest{}, apiErr
	}
	request.Patch = &PatchRequest{
		ViewSchemaID: ViewSchemaID, BaseRowVersion: claims.CurrentRowVersion,
		ClientTxnID: request.ClientTxnID,
		Changes:     []PatchChange{{FieldKey: claims.FieldKey, Value: &value, CanonicalValue: canonical}},
	}
	request.CanonicalValue = canonical
	return request, nil
}

func CreateRequestHash(request CreateRequest) []byte {
	values := make(map[string]any, len(request.Values))
	for fieldKey, value := range request.Values {
		values[fieldKey] = canonicalPartyValue(value)
	}
	return hashMutationPayload(map[string]any{
		"view_schema_id": ViewSchemaID,
		"values":         values,
		"collection_ops": map[string]any{},
		"create_inputs":  map[string]any{},
	})
}

func PatchRequestHash(request PatchRequest) []byte {
	changes := make([]map[string]any, 0, len(request.Changes))
	for _, change := range request.Changes {
		changes = append(changes, map[string]any{"field_key": change.FieldKey, "value": change.CanonicalValue})
	}
	return hashMutationPayload(map[string]any{
		"view_schema_id": ViewSchemaID, "base_row_version": request.BaseRowVersion, "changes": changes,
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

func decodeRawChanges(raw json.RawMessage) ([]json.RawMessage, *httpapi.APIError) {
	if raw == nil {
		return nil, invalidMutationPayload("changes", "missing_required_field")
	}
	var changes []json.RawMessage
	if json.Unmarshal(raw, &changes) != nil {
		return nil, invalidMutationPayload("changes", "invalid_value")
	}
	if len(changes) == 0 {
		return nil, invalidMutationPayload("changes", "empty_changes")
	}
	if len(changes) > maxPatchChanges {
		return nil, invalidMutationPayloadWithDetails("changes", "change_count_exceeded", map[string]any{
			"requested_count": len(changes), "max_count": maxPatchChanges,
		})
	}
	return changes, nil
}

func decodePartyPatchChange(raw json.RawMessage) (PatchChange, *httpapi.APIError) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}
	allowed := map[string]struct{}{"field_key": {}, "value": {}, "action_payload": {}}
	for key := range object {
		if _, admitted := allowed[key]; !admitted {
			return PatchChange{}, invalidMutationPayload("changes", "unknown_field")
		}
	}
	var fieldKey string
	if value, present := object["field_key"]; !present {
		return PatchChange{}, invalidMutationPayload("changes", "missing_field_key")
	} else if json.Unmarshal(value, &fieldKey) != nil {
		return PatchChange{}, invalidMutationPayload("field_key", "invalid_value")
	}
	field, ok := viewschema.LookupField(ViewSchemaID, fieldKey)
	if !ok || !field.Writable {
		return PatchChange{}, invalidMutationPayload(fieldKey, "unsupported_field_key")
	}
	value, hasValue := object["value"]
	_, hasActionPayload := object["action_payload"]
	if hasValue == hasActionPayload {
		return PatchChange{}, invalidMutationPayload("changes", "invalid_change")
	}
	if !hasValue {
		return PatchChange{}, invalidMutationPayload("value", "missing_required_field")
	}
	admitted, canonical, apiErr := decodePartyValue(fieldKey, field, value, true)
	if apiErr != nil {
		return PatchChange{}, apiErr
	}
	return PatchChange{FieldKey: fieldKey, Value: &admitted, CanonicalValue: canonical}, nil
}

func decodePartyValue(
	fieldKey string,
	field viewschema.Field,
	raw json.RawMessage,
	patch bool,
) (FieldValue, any, *httpapi.APIError) {
	if string(raw) == "null" {
		if !patch || field.Clearable {
			return FieldValue{}, nil, nil
		}
		return FieldValue{}, nil, invalidMutationPayload(fieldKey, "field_not_nullable")
	}
	var rawText string
	if json.Unmarshal(raw, &rawText) != nil {
		return FieldValue{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
	}
	var normalized string
	var ok bool
	if field.StringContractID != nil && *field.StringContractID == "multiline_body_v1" {
		normalized, ok = fieldnorm.NormalizeNote(rawText)
	} else {
		normalized, ok = fieldnorm.NormalizeLine(rawText)
	}
	if !ok {
		return FieldValue{}, nil, invalidMutationPayload(fieldKey, "invalid_value")
	}
	return FieldValue{Text: &normalized}, normalized, nil
}

func canonicalPartyValue(value FieldValue) any {
	switch {
	case value.Text != nil:
		return *value.Text
	case value.Timestamp != nil:
		return value.Timestamp.UTC().Format(time.RFC3339Nano)
	case value.UUID != nil:
		return value.UUID.String()
	case value.Number != nil:
		return *value.Number
	case value.Bool != nil:
		return *value.Bool
	default:
		return nil
	}
}

func hashMutationPayload(payload any) []byte {
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return append([]byte(nil), sum[:]...)
}

func invalidMutationPayload(field string, reasonCode string) *httpapi.APIError {
	return invalidMutationPayloadWithDetails(field, reasonCode, nil)
}

func invalidMutationPayloadWithDetails(field string, reasonCode string, extra map[string]any) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	for key, value := range extra {
		details[key] = value
	}
	return &httpapi.APIError{
		Status: http.StatusBadRequest, Code: "invalid_mutation_payload",
		Message: "invalid mutation payload", Details: details,
	}
}
