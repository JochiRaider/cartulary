package savedviews

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func decodeCreateRequest(reader io.Reader) (createRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return createRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"view_schema_id": {},
		"display_name":   {},
		"query_json":     {},
		"layout_json":    {},
		"scope":          {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return createRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	scopeRaw, hasScope := raw["scope"]
	requestScope := scopePrivate
	switch {
	case !hasScope:
	case string(scopeRaw) == "null":
		return createRequest{}, invalidMutationPayload("scope", "field_not_nullable")
	default:
		var scopeValue string
		if err := json.Unmarshal(scopeRaw, &scopeValue); err != nil {
			return createRequest{}, invalidMutationPayload("scope", "invalid_scope")
		}
		parsedScope, policyErr := normalizeOrdinaryScope(scopeValue)
		if policyErr != nil {
			return createRequest{}, invalidMutationPayloadFromPolicy(policyErr)
		}
		requestScope = parsedScope
	}

	return decodeCreateFields(raw, requestScope)
}

func decodeSystemFixtureRequest(reader io.Reader) (createRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return createRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"view_schema_id": {},
		"display_name":   {},
		"query_json":     {},
		"layout_json":    {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return createRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	return decodeCreateFields(raw, scopeSystem)
}

func decodeCreateFields(raw map[string]json.RawMessage, requestScope scope) (createRequest, *httpapi.APIError) {
	var request createRequest
	request.Scope = requestScope
	if value, ok := raw["view_schema_id"]; !ok {
		return createRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || strings.TrimSpace(request.ViewSchemaID) == "" {
		return createRequest{}, invalidMutationPayload("view_schema_id", "invalid_value")
	}
	if policyErr := validateViewSchemaID(request.ViewSchemaID); policyErr != nil {
		return createRequest{}, invalidMutationPayloadFromPolicy(policyErr)
	}

	if value, ok := raw["display_name"]; !ok {
		return createRequest{}, invalidMutationPayload("display_name", "missing_required_field")
	} else if string(value) == "null" {
		return createRequest{}, invalidMutationPayload("display_name", "field_not_nullable")
	} else {
		var displayName string
		if err := json.Unmarshal(value, &displayName); err != nil {
			return createRequest{}, invalidMutationPayload("display_name", "invalid_value")
		}
		normalized, policyErr := normalizeMutationDisplayName(displayName)
		if policyErr != nil {
			return createRequest{}, invalidMutationPayloadFromPolicy(policyErr)
		}
		request.DisplayName = normalized
	}

	queryRaw, ok := raw["query_json"]
	if !ok {
		return createRequest{}, invalidMutationPayload("query_json", "missing_required_field")
	}
	queryJSON, policyErr := normalizeMutationQuery(queryRaw, request.ViewSchemaID)
	if policyErr != nil {
		return createRequest{}, invalidMutationPayloadFromPolicy(policyErr)
	}
	request.QueryJSON = queryJSON

	layoutRaw := json.RawMessage(nil)
	if value, ok := raw["layout_json"]; ok {
		layoutRaw = value
	}
	layoutJSON, policyErr := normalizeMutationLayout(layoutRaw, request.ViewSchemaID)
	if policyErr != nil {
		return createRequest{}, invalidMutationPayloadFromPolicy(policyErr)
	}
	request.LayoutJSON = layoutJSON
	return request, nil
}

func decodePatchRequest(reader io.Reader, viewSchemaID string) (patchRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return patchRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"base_saved_view_version": {},
		"display_name":            {},
		"query_json":              {},
		"layout_json":             {},
		"scope":                   {},
	}
	serverManaged := map[string]struct{}{
		"incident_id":        {},
		"saved_view_id":      {},
		"view_schema_id":     {},
		"owner_user_id":      {},
		"created_at":         {},
		"updated_at":         {},
		"saved_view_version": {},
	}
	for key := range raw {
		if _, ok := serverManaged[key]; ok {
			return patchRequest{}, invalidMutationPayload(key, "server_managed_field")
		}
		if _, ok := allowed[key]; !ok {
			return patchRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request patchRequest
	if value, ok := raw["base_saved_view_version"]; !ok {
		return patchRequest{}, invalidMutationPayload("base_saved_view_version", "missing_required_field")
	} else if string(value) == "null" {
		return patchRequest{}, invalidMutationPayload("base_saved_view_version", "field_not_nullable")
	} else if err := json.Unmarshal(value, &request.BaseSavedViewVersion); err != nil || request.BaseSavedViewVersion < 1 {
		return patchRequest{}, invalidMutationPayload("base_saved_view_version", "invalid_value")
	}

	if value, ok := raw["display_name"]; ok {
		if string(value) == "null" {
			return patchRequest{}, invalidMutationPayload("display_name", "field_not_nullable")
		}
		var displayName string
		if err := json.Unmarshal(value, &displayName); err != nil {
			return patchRequest{}, invalidMutationPayload("display_name", "invalid_value")
		}
		normalized, policyErr := normalizeMutationDisplayName(displayName)
		if policyErr != nil {
			return patchRequest{}, invalidMutationPayloadFromPolicy(policyErr)
		}
		request.DisplayName = optionalString{Present: true, Value: normalized}
	}

	if value, ok := raw["scope"]; ok {
		if string(value) == "null" {
			return patchRequest{}, invalidMutationPayload("scope", "field_not_nullable")
		}
		var scopeValue string
		if err := json.Unmarshal(value, &scopeValue); err != nil {
			return patchRequest{}, invalidMutationPayload("scope", "invalid_scope")
		}
		scope, policyErr := normalizeOrdinaryScope(scopeValue)
		if policyErr != nil {
			return patchRequest{}, invalidMutationPayloadFromPolicy(policyErr)
		}
		request.Scope = optionalScope{Present: true, Value: scope}
	}

	if value, ok := raw["query_json"]; ok {
		if string(value) == "null" {
			return patchRequest{}, invalidMutationPayload("query_json", "field_not_nullable")
		}
		queryJSON, policyErr := normalizeMutationQuery(value, viewSchemaID)
		if policyErr != nil {
			return patchRequest{}, invalidMutationPayloadFromPolicy(policyErr)
		}
		request.QueryJSON = optionalJSON{Present: true, Value: queryJSON}
	}

	if value, ok := raw["layout_json"]; ok {
		if string(value) == "null" {
			return patchRequest{}, invalidMutationPayload("layout_json", "field_not_nullable")
		}
		layoutJSON, policyErr := normalizeMutationLayout(value, viewSchemaID)
		if policyErr != nil {
			return patchRequest{}, invalidMutationPayloadFromPolicy(policyErr)
		}
		request.LayoutJSON = optionalJSON{Present: true, Value: layoutJSON}
	}
	return request, nil
}

func buildResource(record savedViewRecord) map[string]any {
	return map[string]any{
		"saved_view_id":      record.SavedViewID,
		"incident_id":        record.IncidentID,
		"view_schema_id":     record.ViewSchemaID,
		"scope":              record.Scope,
		"display_name":       record.DisplayName,
		"query_json":         decodeJSON(record.QueryJSON),
		"layout_json":        decodeJSON(record.LayoutJSON),
		"owner_user_id":      record.OwnerUserID,
		"created_at":         record.CreatedAt,
		"updated_at":         record.UpdatedAt,
		"saved_view_version": record.SavedViewVersion,
	}
}

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	if raw == nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
}

func decodeJSON(raw []byte) any {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	return value
}

func invalidMutationPayloadFromPolicy(policyErr *mutationPolicyError) *httpapi.APIError {
	field := policyErr.Field
	switch field {
	case "":
		field = "query_json"
	case "sort", "filters", "group_by":
		field = "query_json." + field
	}
	if policyErr.FilterIndex != nil {
		field = "query_json.filters[" + itoa(*policyErr.FilterIndex) + "]"
		if policyErr.FieldKey != "" {
			field += ".field_key"
		}
	}
	details := map[string]any{
		"field":       field,
		"reason_code": policyErr.ReasonCode,
	}
	if policyErr.RequestedCount != nil {
		details["requested_count"] = *policyErr.RequestedCount
	}
	if policyErr.MaxCount != nil {
		details["max_count"] = *policyErr.MaxCount
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func invalidMutationPayload(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func invalidPaginationRequest(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_pagination_request",
		Message: "invalid pagination request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func savedViewNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "saved_view_not_found", Details: map[string]any{}}
}

func authorizationDeniedError() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusForbidden,
		Code:    "authorization_denied",
		Message: "authorization denied",
		Details: map[string]any{},
	}
}

func savedViewVersionConflictAPIError(conflict *savedViewVersionConflictError) *httpapi.APIError {
	details := map[string]any{}
	if conflict != nil {
		details = conflict.Details()
	}
	return &httpapi.APIError{Status: http.StatusConflict, Code: "saved_view_version_conflict", Details: details}
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buffer [20]byte
	i := len(buffer)
	for value > 0 {
		i--
		buffer[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buffer[i:])
}
