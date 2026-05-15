package incidents

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var membershipRoles = []string{"viewer", "editor", "reviewer", "admin"}

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

type UserWorkbookPreferencesPutRequest struct {
	HomeSheetRef []byte
}

type DefaultWorkbookPreferencesPutRequest struct {
	DefaultSheetRef []byte
}

type OptionalNullableString struct {
	Present bool
	Value   *string
}

type IncidentPatchRequest struct {
	BaseIncidentVersion    int64
	TLP                    OptionalNullableString
	CurrentPhase           OptionalNullableString
	PrimaryExternalCaseRef OptionalNullableString
}

func DecodeIncidentCreateRequest(reader io.Reader) (CreateIncidentRequest, *auth.APIError) {
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
		normalized, ok := normalizeLineValue(value)
		if !ok {
			return CreateIncidentRequest{}, invalidIncidentCreate("incident_key", "invalid_value")
		}
		request.IncidentKey = normalized
	}

	if value, ok := raw["title"]; !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("title", "missing_required_field")
	} else {
		normalized, ok := normalizeLineValue(value)
		if !ok {
			return CreateIncidentRequest{}, invalidIncidentCreate("title", "invalid_value")
		}
		request.Title = normalized
	}

	var ok bool
	if request.Description, ok = normalizeNullableNoteField(raw, "description"); !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("description", "invalid_value")
	}
	if request.Severity, ok = normalizeNullableLineField(raw, "severity"); !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("severity", "invalid_value")
	}
	if request.TLP, ok = normalizeNullableLineField(raw, "tlp"); !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("tlp", "invalid_value")
	}
	if request.CurrentPhase, ok = normalizeNullableLineField(raw, "current_phase"); !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("current_phase", "invalid_value")
	}
	if request.PrimaryExternalCaseRef, ok = normalizeNullableLineField(raw, "primary_external_case_ref"); !ok {
		return CreateIncidentRequest{}, invalidIncidentCreate("primary_external_case_ref", "invalid_value")
	}

	return request, nil
}

func DecodeIncidentPatchRequest(reader io.Reader) (IncidentPatchRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader, invalidIncidentPatch)
	if apiErr != nil {
		return IncidentPatchRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_incident_version":     {},
		"tlp":                       {},
		"current_phase":             {},
		"primary_external_case_ref": {},
	}
	forbidden := map[string]struct{}{
		"incident_id":                  {},
		"incident_key":                 {},
		"title":                        {},
		"description":                  {},
		"status":                       {},
		"severity":                     {},
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
	if request.TLP, ok = decodeOptionalNullableLineField(raw, "tlp"); !ok {
		return IncidentPatchRequest{}, invalidIncidentPatch("tlp", "invalid_value")
	}
	if request.CurrentPhase, ok = decodeOptionalNullableLineField(raw, "current_phase"); !ok {
		return IncidentPatchRequest{}, invalidIncidentPatch("current_phase", "invalid_value")
	}
	if request.PrimaryExternalCaseRef, ok = decodeOptionalNullableLineField(raw, "primary_external_case_ref"); !ok {
		return IncidentPatchRequest{}, invalidIncidentPatch("primary_external_case_ref", "invalid_value")
	}
	return request, nil
}

func DecodeMembershipCreateRequest(reader io.Reader) (MembershipCreateRequest, *auth.APIError) {
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

func DecodeMembershipPatchRequest(reader io.Reader) (MembershipPatchRequest, *auth.APIError) {
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

func DecodeMembershipDeleteRequest(reader io.Reader) (MembershipDeleteRequest, *auth.APIError) {
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

func DecodeUserWorkbookPreferencesPutRequest(reader io.Reader) (UserWorkbookPreferencesPutRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return UserWorkbookPreferencesPutRequest{}, apiErr
	}

	for key := range raw {
		if key != "home_sheet_ref" {
			return UserWorkbookPreferencesPutRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	value, ok := raw["home_sheet_ref"]
	if !ok {
		return UserWorkbookPreferencesPutRequest{}, invalidMutationPayload("home_sheet_ref", "missing_required_field")
	}
	sheetRef, apiErr := canonicalSheetRef(value, "home_sheet_ref")
	if apiErr != nil {
		return UserWorkbookPreferencesPutRequest{}, apiErr
	}
	return UserWorkbookPreferencesPutRequest{HomeSheetRef: sheetRef}, nil
}

func DecodeDefaultWorkbookPreferencesPutRequest(reader io.Reader) (DefaultWorkbookPreferencesPutRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return DefaultWorkbookPreferencesPutRequest{}, apiErr
	}

	for key := range raw {
		if key != "default_sheet_ref" {
			return DefaultWorkbookPreferencesPutRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	value, ok := raw["default_sheet_ref"]
	if !ok {
		return DefaultWorkbookPreferencesPutRequest{}, invalidMutationPayload("default_sheet_ref", "missing_required_field")
	}
	sheetRef, apiErr := canonicalSheetRef(value, "default_sheet_ref")
	if apiErr != nil {
		return DefaultWorkbookPreferencesPutRequest{}, apiErr
	}
	return DefaultWorkbookPreferencesPutRequest{DefaultSheetRef: sheetRef}, nil
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

func BuildDefaultWorkbookPreferencesResource(record IncidentWorkbookPreferencesRecord) map[string]any {
	return map[string]any{
		"incident_id":        record.IncidentID,
		"default_sheet_ref":  decodeOptionalJSON(record.DefaultSheetRef),
		"created_at":         record.CreatedAt,
		"updated_at":         record.UpdatedAt,
		"updated_by_user_id": record.UpdatedByUserID,
	}
}

func BuildUserWorkbookPreferencesResource(record UserWorkbookPreferencesRecord) map[string]any {
	return map[string]any{
		"incident_id":    record.IncidentID,
		"user_id":        record.UserID,
		"home_sheet_ref": decodeOptionalJSON(record.HomeSheetRef),
		"created_at":     record.CreatedAt,
		"updated_at":     record.UpdatedAt,
	}
}

func BuildWorkbookStartupResource(record WorkbookStartupRecord) map[string]any {
	cleared := make([]map[string]any, 0, len(record.ClearedPointers))
	for _, pointer := range record.ClearedPointers {
		cleared = append(cleared, map[string]any{
			"source":      pointer.Source,
			"sheet_ref":   decodeOptionalJSON(pointer.SheetRef),
			"reason_code": pointer.ReasonCode,
		})
	}
	var savedView any
	if record.SelectedSavedView != nil {
		savedView = BuildStartupSavedViewResource(*record.SelectedSavedView)
	}
	return map[string]any{
		"incident_id":             record.IncidentID,
		"selected_sheet_ref":      decodeOptionalJSON(record.SelectedSheetRef),
		"selected_view_schema_id": record.SelectedViewSchemaID,
		"selected_saved_view":     savedView,
		"source":                  record.Source,
		"cleared_pointers":        cleared,
		"home_sheet_ref":          decodeOptionalJSON(record.HomeSheetRef),
		"default_sheet_ref":       decodeOptionalJSON(record.DefaultSheetRef),
	}
}

func BuildStartupSavedViewResource(record StartupSavedViewRecord) map[string]any {
	return map[string]any{
		"saved_view_id":      record.SavedViewID,
		"incident_id":        record.IncidentID,
		"view_schema_id":     record.ViewSchemaID,
		"scope":              record.Scope,
		"display_name":       record.DisplayName,
		"query_json":         decodeOptionalJSON(record.QueryJSON),
		"layout_json":        decodeOptionalJSON(record.LayoutJSON),
		"owner_user_id":      record.OwnerUserID,
		"created_at":         record.CreatedAt,
		"updated_at":         record.UpdatedAt,
		"saved_view_version": record.SavedViewVersion,
	}
}

func BuildExtensionResource(profile httpapi.ExtensionProfile) map[string]any {
	return map[string]any{
		"profile_id":     profile.ProfileID,
		"claimed":        profile.Claimed,
		"route_families": append([]string(nil), profile.RouteFamilies...),
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

func invalidIncidentCreate(field string, reasonCode string) *auth.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_incident_create",
		Message: "invalid incident create request",
		Details: details,
	}
}

func invalidIncidentPatch(field string, reasonCode string) *auth.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_incident_patch",
		Message: "invalid incident patch request",
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

func userNotFoundError() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "user_not_found", Details: map[string]any{}}
}

func incidentKeyConflictError(incidentKeyCanonical string) *auth.APIError {
	return &auth.APIError{
		Status: http.StatusConflict,
		Code:   "incident_key_conflict",
		Details: map[string]any{
			"field":                  "incident_key",
			"incident_key_canonical": incidentKeyCanonical,
		},
	}
}

func incidentVersionConflictError(conflict *IncidentVersionConflictError) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "incident_version_conflict", Details: conflict.Details()}
}

func membershipNotFoundError() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "membership_not_found", Details: map[string]any{}}
}

func membershipExistsUsePatchError() *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "membership_exists_use_patch", Details: map[string]any{}}
}

func membershipVersionConflictError() *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "membership_version_conflict", Details: map[string]any{}}
}

func lastIncidentAdminError() *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "last_incident_admin", Details: map[string]any{}}
}

func userInactiveError() *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "user_inactive", Details: map[string]any{}}
}

func authorizationDeniedError(requiredRole string) *auth.APIError {
	details := map[string]any{}
	if requiredRole != "" {
		details["required_role"] = requiredRole
	}
	return &auth.APIError{
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

func decodeOptionalJSON(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	return value
}

func decodeObject(reader io.Reader, invalid func(string, string) *auth.APIError) (map[string]json.RawMessage, *auth.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalid("", "request_not_object")
	}
	return raw, nil
}

func canonicalSheetRef(value json.RawMessage, field string) ([]byte, *auth.APIError) {
	if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return nil, nil
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(value, &object); err != nil {
		return nil, invalidMutationPayload(field, "invalid_sheet_ref")
	}
	allowed := map[string]struct{}{
		"kind": {},
		"id":   {},
	}
	for key := range object {
		if _, ok := allowed[key]; !ok {
			return nil, invalidMutationPayload(field+"."+key, "unknown_field")
		}
	}

	var kind string
	if rawKind, ok := object["kind"]; !ok {
		return nil, invalidMutationPayload(field+".kind", "missing_required_field")
	} else if err := json.Unmarshal(rawKind, &kind); err != nil || strings.TrimSpace(kind) == "" {
		return nil, invalidMutationPayload(field+".kind", "invalid_sheet_ref")
	}
	kind = strings.TrimSpace(kind)
	var id string
	if rawID, ok := object["id"]; !ok {
		return nil, invalidMutationPayload(field+".id", "missing_required_field")
	} else if err := json.Unmarshal(rawID, &id); err != nil || strings.TrimSpace(id) == "" {
		return nil, invalidMutationPayload(field+".id", "invalid_sheet_ref")
	}
	id = strings.TrimSpace(id)

	if apiErr := resolveWorkbookPreferenceSheetRef(kind, id, field); apiErr != nil {
		return nil, apiErr
	}

	canonical, err := json.Marshal(struct {
		Kind string `json:"kind"`
		ID   string `json:"id"`
	}{
		Kind: kind,
		ID:   id,
	})
	if err != nil {
		return nil, internalAPIError(err)
	}
	return canonical, nil
}

func resolveWorkbookPreferenceSheetRef(kind string, id string, field string) *auth.APIError {
	switch kind {
	case "view_schema":
		if _, ok := viewschema.Lookup(id); !ok {
			return invalidMutationPayload(field+".id", "unknown_view_schema")
		}
		return nil
	case "saved_view":
		if _, err := uuid.Parse(id); err != nil {
			return invalidMutationPayload(field+".id", "invalid_saved_view_id")
		}
		return nil
	default:
		return invalidMutationPayload(field+".kind", "unsupported_sheet_ref_kind")
	}
}

func normalizeLineValue(value json.RawMessage) (string, bool) {
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return "", false
	}
	return normalizeLine(raw)
}

func normalizeNullableLineField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok || string(value) == "null" {
		return nil, true
	}
	normalized, ok := normalizeLineValue(value)
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
	normalized, ok := normalizeNote(note)
	if !ok {
		return nil, false
	}
	return &normalized, true
}

func decodeOptionalNullableLineField(raw map[string]json.RawMessage, field string) (OptionalNullableString, bool) {
	value, ok := raw[field]
	if !ok {
		return OptionalNullableString{}, true
	}
	if string(value) == "null" {
		return OptionalNullableString{Present: true, Value: nil}, true
	}
	normalized, ok := normalizeLineValue(value)
	if !ok {
		return OptionalNullableString{}, false
	}
	return OptionalNullableString{Present: true, Value: &normalized}, true
}

func normalizeLine(raw string) (string, bool) {
	normalized := norm.NFC.String(strings.TrimFunc(raw, unicode.IsSpace))
	if normalized == "" {
		return "", false
	}
	for _, r := range normalized {
		if unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r) {
			return "", false
		}
	}
	return normalized, true
}

func normalizeNote(raw string) (string, bool) {
	normalized := norm.NFC.String(raw)
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimFunc(normalized, unicode.IsSpace)
	if normalized == "" {
		return "", false
	}
	for _, r := range normalized {
		switch {
		case r == '\n' || r == '\t':
		case unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r):
			return "", false
		}
	}
	return normalized, true
}
