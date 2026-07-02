package incidents

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
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

type CreateIncidentRequest struct {
	ClientTxnID            string
	IncidentKey            string
	Title                  string
	Description            *string
	Severity               *string
	TLP                    *string
	CurrentPhase           *string
	PrimaryExternalCaseRef *string
}

type MembershipCreateRequest struct {
	ClientTxnID string
	UserID      *uuid.UUID
	Email       *string
	Role        string
}

type MembershipPatchRequest struct {
	BaseMembershipVersion int64
	Role                  string
}

type MembershipDeleteRequest struct {
	BaseMembershipVersion int64
}

type OptionalNullableString struct {
	Present bool
	Value   *string
}

type IncidentPatchRequest struct {
	BaseIncidentVersion    int64
	Description            OptionalNullableString
	Severity               OptionalNullableString
	TLP                    OptionalNullableString
	CurrentPhase           OptionalNullableString
	PrimaryExternalCaseRef OptionalNullableString
}

type IncidentLifecycleRequest struct {
	BaseIncidentVersion int64
	ClientTxnID         string
	Reason              string
}

func DecodeIncidentCreateRequest(reader io.Reader) (CreateIncidentRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidIncidentCreate)
	if apiErr != nil {
		return CreateIncidentRequest{}, apiErr
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
			return CreateIncidentRequest{}, invalidIncidentCreate(key, "collaborator_seeding_not_supported")
		}
		if _, ok := serverManaged[key]; ok {
			return CreateIncidentRequest{}, invalidIncidentCreate(key, "server_managed_field")
		}
		if _, ok := allowed[key]; !ok {
			return CreateIncidentRequest{}, invalidIncidentCreate(key, "unknown_field")
		}
	}

	var request CreateIncidentRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateIncidentRequest{}, invalidIncidentCreate("client_txn_id", "missing_required_field")
	}

	if value, ok := raw["incident_key"]; !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("incident_key", "missing_required_field")
	} else {
		normalized, ok := normalizeIncidentKeyValue(value)
		if !ok {
			return CreateIncidentRequest{}, invalidIncidentCreate("incident_key", "invalid_value")
		}
		request.IncidentKey = normalized
	}

	if value, ok := raw["title"]; !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("title", "missing_required_field")
	} else {
		normalized, ok := normalizeTitleValue(value)
		if !ok {
			return CreateIncidentRequest{}, invalidIncidentCreate("title", "invalid_value")
		}
		request.Title = normalized
	}

	var ok bool
	if request.Description, ok = normalizeNullableNoteField(raw, "description"); !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("description", "invalid_value")
	}
	if request.Severity, ok = normalizeNullableIncidentMetadataField(raw, "severity"); !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("severity", "invalid_value")
	}
	if request.TLP, ok = normalizeNullableTLPField(raw, "tlp"); !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("tlp", "invalid_value")
	}
	if request.CurrentPhase, ok = normalizeNullableIncidentMetadataField(raw, "current_phase"); !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("current_phase", "invalid_value")
	}
	if request.PrimaryExternalCaseRef, ok = normalizeNullableIncidentMetadataField(raw, "primary_external_case_ref"); !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("primary_external_case_ref", "invalid_value")
	}

	return request, nil
}

func DecodeIncidentPatchRequest(reader io.Reader) (IncidentPatchRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidIncidentPatch)
	if apiErr != nil {
		return IncidentPatchRequest{}, apiErr
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
			return IncidentPatchRequest{}, invalidIncidentPatch(key, "forbidden_field")
		}
		if _, ok := allowed[key]; !ok {
			return IncidentPatchRequest{}, invalidIncidentPatch(key, "unknown_field")
		}
	}

	var request IncidentPatchRequest
	if value, ok := raw["base_incident_version"]; !ok {
		return IncidentPatchRequest{}, invalidIncidentPatch("base_incident_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseIncidentVersion); err != nil || request.BaseIncidentVersion < 1 {
		return IncidentPatchRequest{}, invalidIncidentPatch("base_incident_version", "invalid_base_incident_version")
	}

	var ok bool
	if request.Description, ok = decodeOptionalNullableNoteField(raw, "description"); !ok {
		return IncidentPatchRequest{}, invalidIncidentPatch("description", "invalid_value")
	}
	if request.Severity, ok = decodeOptionalNullableIncidentMetadataField(raw, "severity"); !ok {
		return IncidentPatchRequest{}, invalidIncidentPatch("severity", "invalid_value")
	}
	if request.TLP, ok = decodeOptionalNullableTLPField(raw, "tlp"); !ok {
		return IncidentPatchRequest{}, invalidIncidentPatch("tlp", "invalid_value")
	}
	if request.CurrentPhase, ok = decodeOptionalNullableIncidentMetadataField(raw, "current_phase"); !ok {
		return IncidentPatchRequest{}, invalidIncidentPatch("current_phase", "invalid_value")
	}
	if request.PrimaryExternalCaseRef, ok = decodeOptionalNullableIncidentMetadataField(raw, "primary_external_case_ref"); !ok {
		return IncidentPatchRequest{}, invalidIncidentPatch("primary_external_case_ref", "invalid_value")
	}
	return request, nil
}

func DecodeIncidentLifecycleRequest(reader io.Reader) (IncidentLifecycleRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return IncidentLifecycleRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_incident_version": {},
		"client_txn_id":         {},
		"reason":                {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return IncidentLifecycleRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request IncidentLifecycleRequest
	if value, ok := raw["base_incident_version"]; !ok {
		return IncidentLifecycleRequest{}, invalidMutationPayload("base_incident_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseIncidentVersion); err != nil || request.BaseIncidentVersion < 1 {
		return IncidentLifecycleRequest{}, invalidMutationPayload("base_incident_version", "invalid_base_incident_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return IncidentLifecycleRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return IncidentLifecycleRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["reason"]; !ok {
		return IncidentLifecycleRequest{}, invalidMutationPayload("reason", "missing_required_field")
	} else {
		var reason string
		if err := json.Unmarshal(value, &reason); err != nil {
			return IncidentLifecycleRequest{}, invalidMutationPayload("reason", "field_not_nullable")
		}
		normalized := normalizeReasonLine(reason)
		if normalized == "" {
			return IncidentLifecycleRequest{}, invalidMutationPayload("reason", "invalid_value")
		}
		request.Reason = normalized
	}
	return request, nil
}

func DecodeMembershipCreateRequest(reader io.Reader) (MembershipCreateRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return MembershipCreateRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"client_txn_id": {},
		"user_id":       {},
		"email":         {},
		"role":          {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return MembershipCreateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request MembershipCreateRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return MembershipCreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return MembershipCreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	if value, ok := raw["user_id"]; ok {
		var parsed string
		if err := json.Unmarshal(value, &parsed); err != nil {
			return MembershipCreateRequest{}, invalidMutationPayload("user_id", "invalid_user_id")
		}
		userID, err := uuid.Parse(parsed)
		if err != nil {
			return MembershipCreateRequest{}, invalidMutationPayload("user_id", "invalid_user_id")
		}
		request.UserID = &userID
	}

	if value, ok := raw["email"]; ok {
		var rawEmail string
		if err := json.Unmarshal(value, &rawEmail); err != nil {
			return MembershipCreateRequest{}, invalidMutationPayload("email", "invalid_email")
		}
		normalized, _, ok := authn.NormalizeEmailAddress(rawEmail)
		if !ok {
			return MembershipCreateRequest{}, invalidMutationPayload("email", "invalid_email")
		}
		request.Email = &normalized
	}

	if (request.UserID == nil && request.Email == nil) || (request.UserID != nil && request.Email != nil) {
		return MembershipCreateRequest{}, invalidMutationPayload("user_id", "exactly_one_target_selector")
	}

	if value, ok := raw["role"]; !ok {
		return MembershipCreateRequest{}, invalidMutationPayload("role", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Role); err != nil || !slices.Contains(membershipRoles, request.Role) {
		return MembershipCreateRequest{}, invalidMutationPayload("role", "invalid_role")
	}

	return request, nil
}

func DecodeMembershipPatchRequest(reader io.Reader) (MembershipPatchRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return MembershipPatchRequest{}, apiErr
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
			return MembershipPatchRequest{}, invalidMutationPayload(key, "forbidden_field")
		}
		if _, ok := allowed[key]; !ok {
			return MembershipPatchRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request MembershipPatchRequest
	if value, ok := raw["base_membership_version"]; !ok {
		return MembershipPatchRequest{}, invalidMutationPayload("base_membership_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseMembershipVersion); err != nil || request.BaseMembershipVersion < 1 {
		return MembershipPatchRequest{}, invalidMutationPayload("base_membership_version", "invalid_base_membership_version")
	}

	if value, ok := raw["role"]; !ok {
		return MembershipPatchRequest{}, invalidMutationPayload("role", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Role); err != nil || !slices.Contains(membershipRoles, request.Role) {
		return MembershipPatchRequest{}, invalidMutationPayload("role", "invalid_role")
	}
	return request, nil
}

func DecodeMembershipDeleteRequest(reader io.Reader) (MembershipDeleteRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return MembershipDeleteRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_membership_version": {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return MembershipDeleteRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request MembershipDeleteRequest
	if value, ok := raw["base_membership_version"]; !ok {
		return MembershipDeleteRequest{}, invalidMutationPayload("base_membership_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseMembershipVersion); err != nil || request.BaseMembershipVersion < 1 {
		return MembershipDeleteRequest{}, invalidMutationPayload("base_membership_version", "invalid_base_membership_version")
	}
	return request, nil
}

func BuildIncidentResource(record IncidentRecord) map[string]any {
	return map[string]any{
		"incident_id":               record.ID,
		"incident_key":              record.IncidentKey,
		"title":                     record.Title,
		"description":               record.Description,
		"status":                    record.Status,
		"severity":                  record.Severity,
		"tlp":                       record.TLP,
		"current_phase":             record.CurrentPhase,
		"primary_external_case_ref": record.PrimaryExternalCaseRef,
		"created_by_user_id":        record.CreatedByUserID,
		"created_at":                record.CreatedAt,
		"updated_at":                record.UpdatedAt,
		"updated_by_user_id":        record.UpdatedByUserID,
		"incident_version":          record.IncidentVersion,
		"closed_at":                 record.ClosedAt,
	}
}

func BuildMembershipResource(record MembershipRecord) map[string]any {
	return map[string]any{
		"incident_id":        record.IncidentID,
		"user_id":            record.UserID,
		"display_name":       record.DisplayName,
		"role":               record.Role,
		"joined_at":          record.JoinedAt,
		"added_by_user_id":   record.AddedByUserID,
		"updated_at":         record.UpdatedAt,
		"updated_by_user_id": record.UpdatedByUserID,
		"membership_version": record.MembershipVersion,
	}
}

func WouldLeaveNoIncidentAdmins(currentRole string, adminCount int, nextRole *string, deleting bool) bool {
	if currentRole != "admin" {
		return false
	}
	if !deleting && nextRole != nil && *nextRole == "admin" {
		return false
	}
	return adminCount <= 1
}

func invalidIncidentCreate(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_incident_create",
		Message: "invalid incident create request",
		Details: details,
	}
}

func invalidIncidentPatch(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_incident_patch",
		Message: "invalid incident patch request",
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

func invalidListQuery(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_list_query",
		Message: "invalid list query",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func incidentNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func userNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "user_not_found", Details: map[string]any{}}
}

func incidentKeyConflictError(incidentKeyCanonical string) *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusConflict,
		Code:   "incident_key_conflict",
		Details: map[string]any{
			"field":                  "incident_key",
			"incident_key_canonical": incidentKeyCanonical,
		},
	}
}

func incidentVersionConflictError(conflict *IncidentVersionConflictError) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_version_conflict", Details: conflict.Details()}
}

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "incident_closed",
		Message: "incident closed",
		Details: map[string]any{},
	}
}

func incidentIllegalTransitionError(action string) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "illegal_transition",
		Message: "illegal transition",
		Details: map[string]any{
			"action": action,
		},
	}
}

func membershipNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "membership_not_found", Details: map[string]any{}}
}

func membershipExistsUsePatchError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "membership_exists_use_patch", Details: map[string]any{}}
}

func membershipVersionConflictError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "membership_version_conflict", Details: map[string]any{}}
}

func lastIncidentAdminError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "last_incident_admin", Details: map[string]any{}}
}

func userInactiveError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "user_inactive", Details: map[string]any{}}
}

func authorizationDeniedError(requiredRole string) *httpapi.APIError {
	details := map[string]any{}
	if requiredRole != "" {
		details["required_role"] = requiredRole
	}
	return &httpapi.APIError{
		Status:  http.StatusForbidden,
		Code:    "authorization_denied",
		Message: "authorization denied",
		Details: details,
	}
}

func requiredRoleDescription(roles ...string) string {
	if len(roles) == 0 {
		return ""
	}
	if len(roles) == 1 {
		return roles[0]
	}
	return strings.Join(roles, "|")
}

func hashRequestPayload(payload any) []byte {
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
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

func decodeObject(reader io.Reader, invalid func(string, string) *httpapi.APIError) (map[string]json.RawMessage, *httpapi.APIError) {
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

func decodeOptionalNullableIncidentMetadataField(raw map[string]json.RawMessage, field string) (OptionalNullableString, bool) {
	value, ok := raw[field]
	if !ok {
		return OptionalNullableString{}, true
	}
	if string(value) == "null" {
		return OptionalNullableString{Present: true, Value: nil}, true
	}
	var line string
	if err := json.Unmarshal(value, &line); err != nil {
		return OptionalNullableString{}, false
	}
	normalized, ok := normalizeSingleLine(line, singleLineConstraint{maxRunes: incidentMetadataMaxRunes})
	if !ok {
		return OptionalNullableString{}, false
	}
	if normalized == "" {
		return OptionalNullableString{Present: true, Value: nil}, true
	}
	return OptionalNullableString{Present: true, Value: &normalized}, true
}

func decodeOptionalNullableTLPField(raw map[string]json.RawMessage, field string) (OptionalNullableString, bool) {
	value, ok := raw[field]
	if !ok {
		return OptionalNullableString{}, true
	}
	if string(value) == "null" {
		return OptionalNullableString{Present: true, Value: nil}, true
	}
	normalized, ok := normalizeTLPValue(value)
	if !ok {
		return OptionalNullableString{}, false
	}
	return OptionalNullableString{Present: true, Value: &normalized}, true
}

func decodeOptionalNullableNoteField(raw map[string]json.RawMessage, field string) (OptionalNullableString, bool) {
	value, ok := raw[field]
	if !ok {
		return OptionalNullableString{}, true
	}
	if string(value) == "null" {
		return OptionalNullableString{Present: true, Value: nil}, true
	}
	var note string
	if err := json.Unmarshal(value, &note); err != nil {
		return OptionalNullableString{}, false
	}
	normalized, ok := normalizeNote(note, multilineBodyMaxRunes)
	if !ok {
		return OptionalNullableString{}, false
	}
	if normalized == "" {
		return OptionalNullableString{Present: true, Value: nil}, true
	}
	return OptionalNullableString{Present: true, Value: &normalized}, true
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

func normalizeReasonLine(raw string) string {
	normalized, ok := normalizeNote(raw, reasonNoteMaxRunes)
	if !ok || normalized == "" {
		return ""
	}
	return normalized
}
