package savedviews

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type CreateRequest struct {
	ViewSchemaID string
	DisplayName  string
	Scope        Scope
	QueryJSON    []byte
	LayoutJSON   []byte
}

type OptionalString struct {
	Present bool
	Value   string
}

type OptionalScope struct {
	Present bool
	Value   Scope
}

type OptionalJSON struct {
	Present bool
	Value   []byte
}

type PatchRequest struct {
	BaseSavedViewVersion int64
	DisplayName          OptionalString
	Scope                OptionalScope
	QueryJSON            OptionalJSON
	LayoutJSON           OptionalJSON
}

func DecodeCreateRequest(reader io.Reader) (CreateRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return CreateRequest{}, apiErr
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
			return CreateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request CreateRequest
	if value, ok := raw["view_schema_id"]; !ok {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || strings.TrimSpace(request.ViewSchemaID) == "" {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "invalid_value")
	}
	if _, ok := viewschema.Lookup(request.ViewSchemaID); !ok {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}

	if value, ok := raw["display_name"]; !ok {
		return CreateRequest{}, invalidMutationPayload("display_name", "missing_required_field")
	} else if string(value) == "null" {
		return CreateRequest{}, invalidMutationPayload("display_name", "field_not_nullable")
	} else {
		var displayName string
		if err := json.Unmarshal(value, &displayName); err != nil {
			return CreateRequest{}, invalidMutationPayload("display_name", "invalid_value")
		}
		normalized, ok := NormalizeDisplayName(displayName)
		if !ok {
			return CreateRequest{}, invalidMutationPayload("display_name", "invalid_value")
		}
		request.DisplayName = normalized
	}

	scopeRaw, hasScope := raw["scope"]
	switch {
	case !hasScope:
		request.Scope = ScopePrivate
	case string(scopeRaw) == "null":
		return CreateRequest{}, invalidMutationPayload("scope", "field_not_nullable")
	default:
		var scopeValue string
		if err := json.Unmarshal(scopeRaw, &scopeValue); err != nil {
			return CreateRequest{}, invalidMutationPayload("scope", "invalid_scope")
		}
		scope, ok := ParseScope(scopeValue)
		if !ok {
			return CreateRequest{}, invalidMutationPayload("scope", "invalid_scope")
		}
		if !IsOrdinaryCreateScope(scope) {
			return CreateRequest{}, invalidMutationPayload("scope", "system_scope_forbidden")
		}
		request.Scope = scope
	}

	queryRaw, ok := raw["query_json"]
	if !ok {
		return CreateRequest{}, invalidMutationPayload("query_json", "missing_required_field")
	}
	queryJSON, validationErr := viewquery.NormalizePersisted(queryRaw, request.ViewSchemaID)
	if validationErr != nil {
		return CreateRequest{}, invalidMutationPayloadFromQuery(validationErr)
	}
	request.QueryJSON = queryJSON

	layoutRaw := json.RawMessage(nil)
	if value, ok := raw["layout_json"]; ok {
		layoutRaw = value
	}
	layoutJSON, layoutErr := viewschema.NormalizeLayout(layoutRaw, request.ViewSchemaID)
	if layoutErr != nil {
		return CreateRequest{}, invalidMutationPayload(layoutErr.Field, layoutErr.ReasonCode)
	}
	request.LayoutJSON = layoutJSON
	return request, nil
}

func DecodePatchRequest(reader io.Reader, viewSchemaID string) (PatchRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return PatchRequest{}, apiErr
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
			return PatchRequest{}, invalidMutationPayload(key, "server_managed_field")
		}
		if _, ok := allowed[key]; !ok {
			return PatchRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request PatchRequest
	if value, ok := raw["base_saved_view_version"]; !ok {
		return PatchRequest{}, invalidMutationPayload("base_saved_view_version", "missing_required_field")
	} else if string(value) == "null" {
		return PatchRequest{}, invalidMutationPayload("base_saved_view_version", "field_not_nullable")
	} else if err := json.Unmarshal(value, &request.BaseSavedViewVersion); err != nil || request.BaseSavedViewVersion < 1 {
		return PatchRequest{}, invalidMutationPayload("base_saved_view_version", "invalid_value")
	}

	if value, ok := raw["display_name"]; ok {
		if string(value) == "null" {
			return PatchRequest{}, invalidMutationPayload("display_name", "field_not_nullable")
		}
		var displayName string
		if err := json.Unmarshal(value, &displayName); err != nil {
			return PatchRequest{}, invalidMutationPayload("display_name", "invalid_value")
		}
		normalized, ok := NormalizeDisplayName(displayName)
		if !ok {
			return PatchRequest{}, invalidMutationPayload("display_name", "invalid_value")
		}
		request.DisplayName = OptionalString{Present: true, Value: normalized}
	}

	if value, ok := raw["scope"]; ok {
		if string(value) == "null" {
			return PatchRequest{}, invalidMutationPayload("scope", "field_not_nullable")
		}
		var scopeValue string
		if err := json.Unmarshal(value, &scopeValue); err != nil {
			return PatchRequest{}, invalidMutationPayload("scope", "invalid_scope")
		}
		scope, ok := ParseScope(scopeValue)
		if !ok {
			return PatchRequest{}, invalidMutationPayload("scope", "invalid_scope")
		}
		if !IsOrdinaryCreateScope(scope) {
			return PatchRequest{}, invalidMutationPayload("scope", "system_scope_forbidden")
		}
		request.Scope = OptionalScope{Present: true, Value: scope}
	}

	if value, ok := raw["query_json"]; ok {
		if string(value) == "null" {
			return PatchRequest{}, invalidMutationPayload("query_json", "field_not_nullable")
		}
		queryJSON, validationErr := viewquery.NormalizePersisted(value, viewSchemaID)
		if validationErr != nil {
			return PatchRequest{}, invalidMutationPayloadFromQuery(validationErr)
		}
		request.QueryJSON = OptionalJSON{Present: true, Value: queryJSON}
	}

	if value, ok := raw["layout_json"]; ok {
		if string(value) == "null" {
			return PatchRequest{}, invalidMutationPayload("layout_json", "field_not_nullable")
		}
		layoutJSON, layoutErr := viewschema.NormalizeLayout(value, viewSchemaID)
		if layoutErr != nil {
			return PatchRequest{}, invalidMutationPayload(layoutErr.Field, layoutErr.ReasonCode)
		}
		request.LayoutJSON = OptionalJSON{Present: true, Value: layoutJSON}
	}
	return request, nil
}

func ApplyPatch(current Record, request PatchRequest, updatedAt time.Time) (Record, bool, error) {
	next := current
	if request.DisplayName.Present {
		next.DisplayName = request.DisplayName.Value
	}
	if request.Scope.Present {
		next.Scope = request.Scope.Value
	}
	if request.QueryJSON.Present {
		next.QueryJSON = append([]byte(nil), request.QueryJSON.Value...)
	}
	if request.LayoutJSON.Present {
		next.LayoutJSON = append([]byte(nil), request.LayoutJSON.Value...)
	}

	sameQuery, err := jsonStructurallyEqual(current.QueryJSON, next.QueryJSON)
	if err != nil {
		return Record{}, false, err
	}
	sameLayout, err := jsonStructurallyEqual(current.LayoutJSON, next.LayoutJSON)
	if err != nil {
		return Record{}, false, err
	}
	if current.DisplayName == next.DisplayName &&
		current.Scope == next.Scope &&
		sameQuery &&
		sameLayout {
		return current, false, nil
	}

	next.UpdatedAt = updatedAt.UTC()
	next.SavedViewVersion = current.SavedViewVersion + 1
	return next, true, nil
}

func BuildResource(record Record) map[string]any {
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

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *auth.APIError) {
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

func jsonStructurallyEqual(left []byte, right []byte) (bool, error) {
	var leftValue any
	var rightValue any
	leftDecoder := json.NewDecoder(bytes.NewReader(left))
	leftDecoder.UseNumber()
	if err := leftDecoder.Decode(&leftValue); err != nil {
		return false, err
	}
	rightDecoder := json.NewDecoder(bytes.NewReader(right))
	rightDecoder.UseNumber()
	if err := rightDecoder.Decode(&rightValue); err != nil {
		return false, err
	}
	return reflect.DeepEqual(leftValue, rightValue), nil
}

func invalidMutationPayloadFromQuery(validationErr *viewquery.ValidationError) *auth.APIError {
	field := validationErr.Field
	switch field {
	case "":
		field = "query_json"
	case "sort", "filters", "group_by":
		field = "query_json." + field
	}
	if validationErr.FilterIndex != nil {
		field = "query_json.filters[" + itoa(*validationErr.FilterIndex) + "]"
		if validationErr.FieldKey != "" {
			field += ".field_key"
		}
	}
	details := map[string]any{
		"field":       field,
		"reason_code": validationErr.ReasonCode,
	}
	if validationErr.RequestedCount != nil {
		details["requested_count"] = *validationErr.RequestedCount
	}
	if validationErr.MaxCount != nil {
		details["max_count"] = *validationErr.MaxCount
	}
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func invalidMutationPayload(field string, reasonCode string) *auth.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func invalidPaginationRequest(reasonCode string) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_pagination_request",
		Message: "invalid pagination request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func incidentNotFoundError() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func savedViewNotFoundError() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "saved_view_not_found", Details: map[string]any{}}
}

func authorizationDeniedError() *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusForbidden,
		Code:    "authorization_denied",
		Message: "authorization denied",
		Details: map[string]any{},
	}
}

func savedViewVersionConflictError(conflict *SavedViewVersionConflictError) *auth.APIError {
	details := map[string]any{}
	if conflict != nil {
		details = conflict.Details()
	}
	return &auth.APIError{Status: http.StatusConflict, Code: "saved_view_version_conflict", Details: details}
}

func internalAPIError(err error) *auth.APIError {
	return &auth.APIError{
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
