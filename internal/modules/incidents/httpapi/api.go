package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

var membershipRoles = []string{"viewer", "editor", "reviewer", "admin"}

const (
	incidentKeyMaxBytes      = 128
	singleLineTitleMaxRunes  = 512
	incidentMetadataMaxRunes = 128
	multilineBodyMaxRunes    = 16384
	reasonNoteMaxRunes       = 4096
)

var canonicalTLPTokens = map[string]struct{}{
	"TLP:CLEAR":        {},
	"TLP:GREEN":        {},
	"TLP:AMBER":        {},
	"TLP:AMBER+STRICT": {},
	"TLP:RED":          {},
}

func decodeIncidentCreateRequest(reader io.Reader) (incidents.CreateIncidentRequest, *platformhttpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidIncidentCreate)
	if apiErr != nil {
		return incidents.CreateIncidentRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"client_txn_id":             {},
		"incident_key":              {},
		"title":                     {},
		"description":               {},
		"severity":                  {},
		"tlp":                       {},
		"current_phase":             {},
		"primary_external_case_ref": {},
	}
	serverManaged := map[string]struct{}{
		"incident_id":        {},
		"status":             {},
		"incident_version":   {},
		"closed_at":          {},
		"created_at":         {},
		"updated_at":         {},
		"created_by_user_id": {},
		"updated_by_user_id": {},
	}
	for key := range raw {
		if key == "initial_memberships" {
			return incidents.CreateIncidentRequest{}, invalidIncidentCreate(key, "collaborator_seeding_not_supported")
		}
		if _, ok := serverManaged[key]; ok {
			return incidents.CreateIncidentRequest{}, invalidIncidentCreate(key, "server_managed_field")
		}
		if _, ok := allowed[key]; !ok {
			return incidents.CreateIncidentRequest{}, invalidIncidentCreate(key, "unknown_field")
		}
	}

	var request incidents.CreateIncidentRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return incidents.CreateIncidentRequest{}, invalidIncidentCreate("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return incidents.CreateIncidentRequest{}, invalidIncidentCreate("client_txn_id", "missing_required_field")
	}

	if value, ok := raw["incident_key"]; !ok {
		return incidents.CreateIncidentRequest{}, invalidIncidentCreate("incident_key", "missing_required_field")
	} else {
		normalized, ok := normalizeIncidentKeyValue(value)
		if !ok {
			return incidents.CreateIncidentRequest{}, invalidIncidentCreate("incident_key", "invalid_value")
		}
		request.IncidentKey = normalized
	}

	if value, ok := raw["title"]; !ok {
		return incidents.CreateIncidentRequest{}, invalidIncidentCreate("title", "missing_required_field")
	} else {
		normalized, ok := normalizeTitleValue(value)
		if !ok {
			return incidents.CreateIncidentRequest{}, invalidIncidentCreate("title", "invalid_value")
		}
		request.Title = normalized
	}

	var ok bool
	if request.Description, ok = normalizeNullableNoteField(raw, "description"); !ok {
		return incidents.CreateIncidentRequest{}, invalidIncidentCreate("description", "invalid_value")
	}
	if request.Severity, ok = normalizeNullableIncidentMetadataField(raw, "severity"); !ok {
		return incidents.CreateIncidentRequest{}, invalidIncidentCreate("severity", "invalid_value")
	}
	if request.TLP, ok = normalizeNullableTLPField(raw, "tlp"); !ok {
		return incidents.CreateIncidentRequest{}, invalidIncidentCreate("tlp", "invalid_value")
	}
	if request.CurrentPhase, ok = normalizeNullableIncidentMetadataField(raw, "current_phase"); !ok {
		return incidents.CreateIncidentRequest{}, invalidIncidentCreate("current_phase", "invalid_value")
	}
	if request.PrimaryExternalCaseRef, ok = normalizeNullableIncidentMetadataField(raw, "primary_external_case_ref"); !ok {
		return incidents.CreateIncidentRequest{}, invalidIncidentCreate("primary_external_case_ref", "invalid_value")
	}

	return request, nil
}

func decodeIncidentPatchRequest(reader io.Reader) (incidents.IncidentPatchRequest, *platformhttpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidIncidentPatch)
	if apiErr != nil {
		return incidents.IncidentPatchRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_incident_version":     {},
		"description":               {},
		"severity":                  {},
		"tlp":                       {},
		"current_phase":             {},
		"primary_external_case_ref": {},
	}
	forbidden := map[string]struct{}{
		"incident_id":                  {},
		"incident_key":                 {},
		"title":                        {},
		"status":                       {},
		"created_at":                   {},
		"created_by_user_id":           {},
		"updated_at":                   {},
		"updated_by_user_id":           {},
		"closed_at":                    {},
		"incident_version":             {},
		"memberships":                  {},
		"saved_views":                  {},
		"saved_view":                   {},
		"workbook_preferences":         {},
		"default_workbook_preferences": {},
		"user_workbook_preferences":    {},
	}
	for key := range raw {
		if _, ok := forbidden[key]; ok {
			return incidents.IncidentPatchRequest{}, invalidIncidentPatch(key, "forbidden_field")
		}
		if _, ok := allowed[key]; !ok {
			return incidents.IncidentPatchRequest{}, invalidIncidentPatch(key, "unknown_field")
		}
	}

	var request incidents.IncidentPatchRequest
	if value, ok := raw["base_incident_version"]; !ok {
		return incidents.IncidentPatchRequest{}, invalidIncidentPatch("base_incident_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseIncidentVersion); err != nil || request.BaseIncidentVersion < 1 {
		return incidents.IncidentPatchRequest{}, invalidIncidentPatch("base_incident_version", "invalid_base_incident_version")
	}

	var ok bool
	if request.Description, ok = decodeOptionalNullableNoteField(raw, "description"); !ok {
		return incidents.IncidentPatchRequest{}, invalidIncidentPatch("description", "invalid_value")
	}
	if request.Severity, ok = decodeOptionalNullableIncidentMetadataField(raw, "severity"); !ok {
		return incidents.IncidentPatchRequest{}, invalidIncidentPatch("severity", "invalid_value")
	}
	if request.TLP, ok = decodeOptionalNullableTLPField(raw, "tlp"); !ok {
		return incidents.IncidentPatchRequest{}, invalidIncidentPatch("tlp", "invalid_value")
	}
	if request.CurrentPhase, ok = decodeOptionalNullableIncidentMetadataField(raw, "current_phase"); !ok {
		return incidents.IncidentPatchRequest{}, invalidIncidentPatch("current_phase", "invalid_value")
	}
	if request.PrimaryExternalCaseRef, ok = decodeOptionalNullableIncidentMetadataField(raw, "primary_external_case_ref"); !ok {
		return incidents.IncidentPatchRequest{}, invalidIncidentPatch("primary_external_case_ref", "invalid_value")
	}
	return request, nil
}

func decodeIncidentLifecycleRequest(reader io.Reader) (incidents.IncidentLifecycleRequest, *platformhttpapi.APIError) {
	raw, err := platformhttpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest("", "request_not_object")
	}

	allowed := map[string]struct{}{
		"base_incident_version": {},
		"client_txn_id":         {},
		"reason":                {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest(key, "unknown_field")
		}
	}

	var request incidents.IncidentLifecycleRequest
	if value, ok := raw["base_incident_version"]; !ok {
		return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest("base_incident_version", "missing_required_field")
	} else if isJSONNull(value) {
		return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest("base_incident_version", "field_not_nullable")
	} else if err := json.Unmarshal(value, &request.BaseIncidentVersion); err != nil || request.BaseIncidentVersion < 1 {
		return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest("base_incident_version", "invalid_base_incident_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest("client_txn_id", "missing_required_field")
	} else if isJSONNull(value) {
		return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest("client_txn_id", "field_not_nullable")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest("client_txn_id", "invalid_client_txn_id")
	}
	if value, ok := raw["reason"]; !ok {
		return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest("reason", "missing_required_field")
	} else if isJSONNull(value) {
		return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest("reason", "field_not_nullable")
	} else {
		var reason string
		if err := json.Unmarshal(value, &reason); err != nil {
			return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest("reason", "invalid_reason")
		}
		normalized, reasonCode := normalizeIncidentLifecycleReason(reason)
		if reasonCode != "" {
			return incidents.IncidentLifecycleRequest{}, invalidIncidentLifecycleRequest("reason", reasonCode)
		}
		request.Reason = normalized
	}
	return request, nil
}

func decodeMembershipCreateRequest(reader io.Reader) (incidents.MembershipCreateRequest, *platformhttpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return incidents.MembershipCreateRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"client_txn_id": {},
		"user_id":       {},
		"email":         {},
		"role":          {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return incidents.MembershipCreateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request incidents.MembershipCreateRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return incidents.MembershipCreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return incidents.MembershipCreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	if value, ok := raw["user_id"]; ok {
		var parsed string
		if err := json.Unmarshal(value, &parsed); err != nil {
			return incidents.MembershipCreateRequest{}, invalidMutationPayload("user_id", "invalid_user_id")
		}
		userID, err := uuid.Parse(parsed)
		if err != nil {
			return incidents.MembershipCreateRequest{}, invalidMutationPayload("user_id", "invalid_user_id")
		}
		request.UserID = &userID
	}

	if value, ok := raw["email"]; ok {
		var rawEmail string
		if err := json.Unmarshal(value, &rawEmail); err != nil {
			return incidents.MembershipCreateRequest{}, invalidMutationPayload("email", "invalid_email")
		}
		normalized, _, ok := authn.NormalizeEmailAddress(rawEmail)
		if !ok {
			return incidents.MembershipCreateRequest{}, invalidMutationPayload("email", "invalid_email")
		}
		request.Email = &normalized
	}

	if (request.UserID == nil && request.Email == nil) || (request.UserID != nil && request.Email != nil) {
		return incidents.MembershipCreateRequest{}, invalidMutationPayload("user_id", "exactly_one_target_selector")
	}

	if value, ok := raw["role"]; !ok {
		return incidents.MembershipCreateRequest{}, invalidMutationPayload("role", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Role); err != nil || !slices.Contains(membershipRoles, request.Role) {
		return incidents.MembershipCreateRequest{}, invalidMutationPayload("role", "invalid_role")
	}

	return request, nil
}

func decodeMembershipPatchRequest(reader io.Reader) (incidents.MembershipPatchRequest, *platformhttpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return incidents.MembershipPatchRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_membership_version": {},
		"role":                    {},
	}
	forbidden := map[string]struct{}{
		"incident_id":        {},
		"user_id":            {},
		"joined_at":          {},
		"added_by_user_id":   {},
		"updated_at":         {},
		"updated_by_user_id": {},
		"membership_version": {},
	}
	for key := range raw {
		if _, ok := forbidden[key]; ok {
			return incidents.MembershipPatchRequest{}, invalidMutationPayload(key, "forbidden_field")
		}
		if _, ok := allowed[key]; !ok {
			return incidents.MembershipPatchRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request incidents.MembershipPatchRequest
	if value, ok := raw["base_membership_version"]; !ok {
		return incidents.MembershipPatchRequest{}, invalidMutationPayload("base_membership_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseMembershipVersion); err != nil || request.BaseMembershipVersion < 1 {
		return incidents.MembershipPatchRequest{}, invalidMutationPayload("base_membership_version", "invalid_base_membership_version")
	}

	if value, ok := raw["role"]; !ok {
		return incidents.MembershipPatchRequest{}, invalidMutationPayload("role", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Role); err != nil || !slices.Contains(membershipRoles, request.Role) {
		return incidents.MembershipPatchRequest{}, invalidMutationPayload("role", "invalid_role")
	}
	return request, nil
}

func decodeMembershipDeleteRequest(reader io.Reader) (incidents.MembershipDeleteRequest, *platformhttpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return incidents.MembershipDeleteRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_membership_version": {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return incidents.MembershipDeleteRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request incidents.MembershipDeleteRequest
	if value, ok := raw["base_membership_version"]; !ok {
		return incidents.MembershipDeleteRequest{}, invalidMutationPayload("base_membership_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseMembershipVersion); err != nil || request.BaseMembershipVersion < 1 {
		return incidents.MembershipDeleteRequest{}, invalidMutationPayload("base_membership_version", "invalid_base_membership_version")
	}
	return request, nil
}

func invalidIncidentCreate(field string, reasonCode string) *platformhttpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &platformhttpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_incident_create",
		Message: "invalid incident create request",
		Details: details,
	}
}

func invalidIncidentPatch(field string, reasonCode string) *platformhttpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &platformhttpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_incident_patch",
		Message: "invalid incident patch request",
		Details: details,
	}
}

func invalidMutationPayload(field string, reasonCode string) *platformhttpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &platformhttpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func invalidIncidentLifecycleRequest(field string, reasonCode string) *platformhttpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &platformhttpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_incident_lifecycle_request",
		Message: "invalid incident lifecycle request",
		Details: details,
	}
}

func invalidPaginationRequest(reasonCode string) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_pagination_request",
		Message: "invalid pagination request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func invalidListQuery(reasonCode string) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_list_query",
		Message: "invalid list query",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func incidentNotFoundError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func userNotFoundError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "user_not_found", Details: map[string]any{}}
}

func incidentKeyConflictError(incidentKeyCanonical string) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status: http.StatusConflict,
		Code:   "incident_key_conflict",
		Details: map[string]any{
			"field":                  "incident_key",
			"incident_key_canonical": incidentKeyCanonical,
		},
	}
}

func incidentVersionConflictError(conflict *incidents.IncidentVersionConflictError) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "incident_version_conflict", Details: conflict.Details()}
}

func incidentClosedError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "incident_closed",
		Message: "incident closed",
		Details: map[string]any{},
	}
}

func incidentIllegalTransitionError(action string) *platformhttpapi.APIError {
	reasonCode := "incident_not_closed"
	if action == "close" {
		reasonCode = "incident_already_closed"
	}
	return &platformhttpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "illegal_transition",
		Message: "illegal transition",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func membershipNotFoundError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "membership_not_found", Details: map[string]any{}}
}

func membershipExistsUsePatchError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "membership_exists_use_patch", Details: map[string]any{}}
}

func membershipVersionConflictError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "membership_version_conflict", Details: map[string]any{}}
}

func lastIncidentAdminError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "last_incident_admin", Details: map[string]any{}}
}

func userInactiveError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "user_inactive", Details: map[string]any{}}
}

func authorizationDeniedError(requiredRole string) *platformhttpapi.APIError {
	details := map[string]any{}
	if requiredRole != "" {
		details["required_role"] = requiredRole
	}
	return &platformhttpapi.APIError{
		Status:  http.StatusForbidden,
		Code:    "authorization_denied",
		Message: "authorization denied",
		Details: details,
	}
}

func decodeObject(reader io.Reader, invalid func(string, string) *platformhttpapi.APIError) (map[string]json.RawMessage, *platformhttpapi.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalid("", "request_not_object")
	}
	return raw, nil
}

func normalizeIncidentKeyValue(value json.RawMessage) (string, bool) {
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return "", false
	}
	normalized, ok := normalizeSingleLine(raw, singleLineConstraint{maxBytes: incidentKeyMaxBytes})
	if !ok || normalized == "" {
		return "", false
	}
	return normalized, true
}

func normalizeTitleValue(value json.RawMessage) (string, bool) {
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return "", false
	}
	normalized, ok := normalizeSingleLine(raw, singleLineConstraint{maxRunes: singleLineTitleMaxRunes})
	if !ok || normalized == "" {
		return "", false
	}
	return normalized, true
}

func normalizeNullableIncidentMetadataField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok || string(value) == "null" {
		return nil, true
	}
	var line string
	if err := json.Unmarshal(value, &line); err != nil {
		return nil, false
	}
	normalized, ok := normalizeSingleLine(line, singleLineConstraint{maxRunes: incidentMetadataMaxRunes})
	if !ok {
		return nil, false
	}
	if normalized == "" {
		return nil, true
	}
	return &normalized, true
}

func normalizeNullableTLPField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok || string(value) == "null" {
		return nil, true
	}
	normalized, ok := normalizeTLPValue(value)
	if !ok {
		return nil, false
	}
	return &normalized, true
}

func normalizeNullableNoteField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok || string(value) == "null" {
		return nil, true
	}
	var note string
	if err := json.Unmarshal(value, &note); err != nil {
		return nil, false
	}
	normalized, ok := normalizeNote(note, multilineBodyMaxRunes)
	if !ok {
		return nil, false
	}
	if normalized == "" {
		return nil, true
	}
	return &normalized, true
}

func decodeOptionalNullableIncidentMetadataField(raw map[string]json.RawMessage, field string) (incidents.OptionalNullableString, bool) {
	value, ok := raw[field]
	if !ok {
		return incidents.OptionalNullableString{}, true
	}
	if string(value) == "null" {
		return incidents.OptionalNullableString{Present: true, Value: nil}, true
	}
	var line string
	if err := json.Unmarshal(value, &line); err != nil {
		return incidents.OptionalNullableString{}, false
	}
	normalized, ok := normalizeSingleLine(line, singleLineConstraint{maxRunes: incidentMetadataMaxRunes})
	if !ok {
		return incidents.OptionalNullableString{}, false
	}
	if normalized == "" {
		return incidents.OptionalNullableString{Present: true, Value: nil}, true
	}
	return incidents.OptionalNullableString{Present: true, Value: &normalized}, true
}

func decodeOptionalNullableTLPField(raw map[string]json.RawMessage, field string) (incidents.OptionalNullableString, bool) {
	value, ok := raw[field]
	if !ok {
		return incidents.OptionalNullableString{}, true
	}
	if string(value) == "null" {
		return incidents.OptionalNullableString{Present: true, Value: nil}, true
	}
	normalized, ok := normalizeTLPValue(value)
	if !ok {
		return incidents.OptionalNullableString{}, false
	}
	return incidents.OptionalNullableString{Present: true, Value: &normalized}, true
}

func decodeOptionalNullableNoteField(raw map[string]json.RawMessage, field string) (incidents.OptionalNullableString, bool) {
	value, ok := raw[field]
	if !ok {
		return incidents.OptionalNullableString{}, true
	}
	if string(value) == "null" {
		return incidents.OptionalNullableString{Present: true, Value: nil}, true
	}
	var note string
	if err := json.Unmarshal(value, &note); err != nil {
		return incidents.OptionalNullableString{}, false
	}
	normalized, ok := normalizeNote(note, multilineBodyMaxRunes)
	if !ok {
		return incidents.OptionalNullableString{}, false
	}
	if normalized == "" {
		return incidents.OptionalNullableString{Present: true, Value: nil}, true
	}
	return incidents.OptionalNullableString{Present: true, Value: &normalized}, true
}

type singleLineConstraint struct {
	maxBytes int
	maxRunes int
}

func normalizeSingleLine(raw string, constraint singleLineConstraint) (string, bool) {
	normalized := norm.NFC.String(strings.TrimFunc(raw, unicode.IsSpace))
	for _, r := range normalized {
		if unicode.Is(unicode.Cc, r) {
			return "", false
		}
	}
	if constraint.maxBytes > 0 && len([]byte(normalized)) > constraint.maxBytes {
		return "", false
	}
	if constraint.maxRunes > 0 && len([]rune(normalized)) > constraint.maxRunes {
		return "", false
	}
	return normalized, true
}

func normalizeNote(raw string, maxRunes int) (string, bool) {
	normalized := norm.NFC.String(raw)
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimFunc(normalized, unicode.IsSpace)
	for _, r := range normalized {
		switch {
		case r == '\n' || r == '\t':
		case unicode.Is(unicode.Cc, r):
			return "", false
		}
	}
	if maxRunes > 0 && len([]rune(normalized)) > maxRunes {
		return "", false
	}
	return normalized, true
}

func normalizeTLPValue(value json.RawMessage) (string, bool) {
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return "", false
	}
	if _, ok := canonicalTLPTokens[raw]; !ok {
		return "", false
	}
	return raw, true
}

func normalizeIncidentLifecycleReason(raw string) (string, string) {
	normalized := norm.NFC.String(raw)
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimFunc(normalized, unicode.IsSpace)
	if normalized == "" {
		return "", "reason_empty_after_normalization"
	}
	if utf8.RuneCountInString(normalized) > reasonNoteMaxRunes {
		return "", "reason_too_long"
	}
	for _, r := range normalized {
		switch {
		case r == '\n' || r == '\t':
		case unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r):
			return "", "control_character_not_allowed"
		}
	}
	return normalized, ""
}

func isJSONNull(value json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(value), []byte("null"))
}
