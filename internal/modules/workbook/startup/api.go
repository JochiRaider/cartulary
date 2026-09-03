package startup

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	TimelineViewSchemaID = "cartulary.view.timeline.v2"
	SourceExplicit       = "explicit"
	SourceHome           = "home"
	SourceDefault        = "default"
	SourceTimeline       = "timeline"
)

type SheetRef struct {
	Kind               string `json:"kind"`
	ID                 string `json:"id,omitempty"`
	ExtensionProfileID string `json:"extension_profile_id,omitempty"`
	WorkspaceKey       string `json:"workspace_key,omitempty"`
}

type DefaultPreferencesRecord struct {
	IncidentID      uuid.UUID
	DefaultSheetRef []byte
	CreatedAt       time.Time
	UpdatedAt       time.Time
	UpdatedByUserID *uuid.UUID
}

type UserPreferencesRecord struct {
	IncidentID   uuid.UUID
	UserID       uuid.UUID
	HomeSheetRef []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SavedViewRecord struct {
	SavedViewID      uuid.UUID
	IncidentID       uuid.UUID
	ViewSchemaID     string
	Scope            string
	DisplayName      string
	QueryJSON        []byte
	LayoutJSON       []byte
	OwnerUserID      *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	SavedViewVersion int64
}

type ClearedPointer struct {
	Source     string
	SheetRef   []byte
	ReasonCode string
}

type Record struct {
	IncidentID                     uuid.UUID
	SelectedSheetRef               []byte
	SelectedViewSchemaID           *string
	SelectedSavedView              *SavedViewRecord
	Source                         string
	ClearedPointers                []ClearedPointer
	HomeSheetRef                   []byte
	DefaultSheetRef                []byte
	ExtensionWorkspaceAvailability ExtensionWorkspaceAvailability
}

type DefaultPreferencesPutRequest struct {
	DefaultSheetRef []byte
}

type UserPreferencesPutRequest struct {
	HomeSheetRef []byte
}

func DecodeDefaultPreferencesPutRequest(reader io.Reader) (DefaultPreferencesPutRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return DefaultPreferencesPutRequest{}, apiErr
	}
	for key := range raw {
		if key != "default_sheet_ref" {
			return DefaultPreferencesPutRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	value, ok := raw["default_sheet_ref"]
	if !ok {
		return DefaultPreferencesPutRequest{}, invalidMutationPayload("default_sheet_ref", "missing_required_field")
	}
	sheetRef, apiErr := canonicalSheetRef(value, "default_sheet_ref")
	if apiErr != nil {
		return DefaultPreferencesPutRequest{}, apiErr
	}
	return DefaultPreferencesPutRequest{DefaultSheetRef: sheetRef}, nil
}

func DecodeUserPreferencesPutRequest(reader io.Reader) (UserPreferencesPutRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return UserPreferencesPutRequest{}, apiErr
	}
	for key := range raw {
		if key != "home_sheet_ref" {
			return UserPreferencesPutRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	value, ok := raw["home_sheet_ref"]
	if !ok {
		return UserPreferencesPutRequest{}, invalidMutationPayload("home_sheet_ref", "missing_required_field")
	}
	sheetRef, apiErr := canonicalSheetRef(value, "home_sheet_ref")
	if apiErr != nil {
		return UserPreferencesPutRequest{}, apiErr
	}
	return UserPreferencesPutRequest{HomeSheetRef: sheetRef}, nil
}

func ParseExplicitSheetRef(query url.Values) ([]byte, *httpapi.APIError) {
	sheetRefKind := strings.TrimSpace(query.Get("sheet_ref_kind"))
	sheetRefID := strings.TrimSpace(query.Get("sheet_ref_id"))
	extensionProfileID := strings.TrimSpace(query.Get("extension_profile_id"))
	hasSheetRef := sheetRefKind != "" || sheetRefID != "" || extensionProfileID != ""
	for key := range query {
		switch key {
		case "sheet_ref_kind", "sheet_ref_id", "extension_profile_id":
		default:
			return nil, invalidStartupRequest(key, "unknown_field")
		}
	}
	if !hasSheetRef {
		return nil, nil
	}
	if sheetRefKind == "" {
		return nil, invalidStartupRequest("sheet_ref_kind", "missing_required_field")
	}
	if sheetRefID == "" && sheetRefKind != "extension_workspace" {
		return nil, invalidStartupRequest("sheet_ref_id", "missing_required_field")
	}
	if sheetRefKind == "extension_workspace" {
		if !validExtensionToken(extensionProfileID) {
			return nil, invalidStartupRequest("extension_profile_id", "invalid_extension_profile_id")
		}
		if !validExtensionToken(sheetRefID) {
			return nil, invalidStartupRequest("sheet_ref_id", "invalid_extension_workspace_key")
		}
		return canonicalStartupSheetRef(SheetRef{
			Kind:               sheetRefKind,
			ExtensionProfileID: extensionProfileID,
			WorkspaceKey:       sheetRefID,
		})
	}
	if extensionProfileID != "" {
		return nil, invalidStartupRequest("extension_profile_id", "unknown_field")
	}
	return canonicalStartupSheetRef(SheetRef{Kind: sheetRefKind, ID: sheetRefID})
}

func BuildDefaultPreferencesResource(record DefaultPreferencesRecord) map[string]any {
	return map[string]any{
		"incident_id":        record.IncidentID,
		"default_sheet_ref":  decodeOptionalJSON(record.DefaultSheetRef),
		"created_at":         record.CreatedAt,
		"updated_at":         record.UpdatedAt,
		"updated_by_user_id": record.UpdatedByUserID,
	}
}

func BuildUserPreferencesResource(record UserPreferencesRecord) map[string]any {
	return map[string]any{
		"incident_id":    record.IncidentID,
		"user_id":        record.UserID,
		"home_sheet_ref": decodeOptionalJSON(record.HomeSheetRef),
		"created_at":     record.CreatedAt,
		"updated_at":     record.UpdatedAt,
	}
}

func BuildStartupResource(record Record) map[string]any {
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
		savedView = BuildSavedViewResource(*record.SelectedSavedView)
	}
	availability := record.ExtensionWorkspaceAvailability
	if availability.SchemaID == "" {
		availability = ExtensionWorkspaceAvailability{
			SchemaID:   ExtensionWorkspaceAvailabilitySchemaID,
			IncidentID: record.IncidentID.String(),
			Workspaces: []ExtensionWorkspaceAvailabilityRow{},
		}
	}
	return map[string]any{
		"incident_id":                      record.IncidentID,
		"selected_sheet_ref":               decodeOptionalJSON(record.SelectedSheetRef),
		"selected_view_schema_id":          record.SelectedViewSchemaID,
		"selected_saved_view":              savedView,
		"source":                           record.Source,
		"cleared_pointers":                 cleared,
		"home_sheet_ref":                   decodeOptionalJSON(record.HomeSheetRef),
		"default_sheet_ref":                decodeOptionalJSON(record.DefaultSheetRef),
		"extension_workspace_availability": availability,
	}
}

func BuildSavedViewResource(record SavedViewRecord) map[string]any {
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

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
}

func canonicalSheetRef(value json.RawMessage, field string) ([]byte, *httpapi.APIError) {
	if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return nil, nil
	}
	object, err := httpapi.DecodeStrictJSONObject(bytes.NewReader(value))
	if err != nil {
		return nil, invalidMutationPayload(field, "invalid_sheet_ref")
	}
	var kind string
	if rawKind, ok := object["kind"]; !ok {
		return nil, invalidMutationPayload(field+".kind", "missing_required_field")
	} else if err := json.Unmarshal(rawKind, &kind); err != nil || strings.TrimSpace(kind) == "" {
		return nil, invalidMutationPayload(field+".kind", "invalid_sheet_ref")
	}
	kind = strings.TrimSpace(kind)
	switch kind {
	case "view_schema", "saved_view":
		if apiErr := rejectUnknownSheetRefMembers(object, field, "kind", "id"); apiErr != nil {
			return nil, apiErr
		}
		id, apiErr := requiredSheetRefString(object, "id", field, "invalid_sheet_ref")
		if apiErr != nil {
			return nil, apiErr
		}
		if apiErr := resolveSheetRef(kind, id, field); apiErr != nil {
			return nil, apiErr
		}
		return mustSheetRefJSON(SheetRef{Kind: kind, ID: id}), nil
	case "extension_workspace":
		if apiErr := rejectUnknownSheetRefMembers(object, field, "kind", "extension_profile_id", "workspace_key"); apiErr != nil {
			return nil, apiErr
		}
		profileID, apiErr := requiredSheetRefString(object, "extension_profile_id", field, "invalid_extension_profile_id")
		if apiErr != nil {
			return nil, apiErr
		}
		if !validExtensionToken(profileID) {
			return nil, invalidMutationPayload(field+".extension_profile_id", "invalid_extension_profile_id")
		}
		workspaceKey, apiErr := requiredSheetRefString(object, "workspace_key", field, "invalid_extension_workspace_key")
		if apiErr != nil {
			return nil, apiErr
		}
		if !validExtensionToken(workspaceKey) {
			return nil, invalidMutationPayload(field+".workspace_key", "invalid_extension_workspace_key")
		}
		return mustSheetRefJSON(SheetRef{
			Kind:               kind,
			ExtensionProfileID: profileID,
			WorkspaceKey:       workspaceKey,
		}), nil
	default:
		return nil, invalidMutationPayload(field+".kind", "unsupported_sheet_ref_kind")
	}
}

func rejectUnknownSheetRefMembers(object map[string]json.RawMessage, field string, allowed ...string) *httpapi.APIError {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	for key := range object {
		if _, ok := allowedSet[key]; !ok {
			return invalidMutationPayload(field+"."+key, "unknown_field")
		}
	}
	return nil
}

func requiredSheetRefString(object map[string]json.RawMessage, key string, field string, invalidReason string) (string, *httpapi.APIError) {
	raw, ok := object[key]
	if !ok {
		missingReason := "missing_required_field"
		if strings.HasPrefix(invalidReason, "invalid_extension_") {
			missingReason = invalidReason
		}
		return "", invalidMutationPayload(field+"."+key, missingReason)
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil || strings.TrimSpace(value) == "" {
		return "", invalidMutationPayload(field+"."+key, invalidReason)
	}
	return strings.TrimSpace(value), nil
}

func resolveSheetRef(kind string, id string, field string) *httpapi.APIError {
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

func canonicalStartupSheetRef(ref SheetRef) ([]byte, *httpapi.APIError) {
	ref.Kind = strings.TrimSpace(ref.Kind)
	ref.ID = strings.TrimSpace(ref.ID)
	ref.ExtensionProfileID = strings.TrimSpace(ref.ExtensionProfileID)
	ref.WorkspaceKey = strings.TrimSpace(ref.WorkspaceKey)
	switch ref.Kind {
	case "view_schema":
	case "saved_view":
		if _, err := uuid.Parse(ref.ID); err != nil {
			return nil, invalidStartupRequest("sheet_ref_id", "invalid_saved_view_id")
		}
	case "extension_workspace":
		if !validExtensionToken(ref.ExtensionProfileID) {
			return nil, invalidStartupRequest("extension_profile_id", "invalid_extension_profile_id")
		}
		if !validExtensionToken(ref.WorkspaceKey) {
			return nil, invalidStartupRequest("sheet_ref_id", "invalid_extension_workspace_key")
		}
		return mustSheetRefJSON(ref), nil
	default:
		return nil, invalidStartupRequest("sheet_ref_kind", "unsupported_sheet_ref_kind")
	}
	if ref.ID == "" {
		return nil, invalidStartupRequest("sheet_ref_id", "missing_required_field")
	}
	return mustSheetRefJSON(ref), nil
}

func decodeSheetRef(raw []byte) (SheetRef, string) {
	if len(raw) == 0 {
		return SheetRef{}, "invalid_sheet_ref"
	}
	object, err := httpapi.DecodeStrictJSONObject(bytes.NewReader(raw))
	if err != nil {
		return SheetRef{}, "invalid_sheet_ref"
	}
	var kind string
	if rawKind, ok := object["kind"]; !ok || json.Unmarshal(rawKind, &kind) != nil || strings.TrimSpace(kind) == "" {
		return SheetRef{}, "invalid_sheet_ref"
	}
	kind = strings.TrimSpace(kind)
	switch kind {
	case "view_schema", "saved_view":
		if !sheetRefHasOnlyMembers(object, "kind", "id") {
			return SheetRef{}, "invalid_sheet_ref"
		}
		var id string
		if rawID, ok := object["id"]; !ok || json.Unmarshal(rawID, &id) != nil || strings.TrimSpace(id) == "" {
			if kind == "saved_view" {
				return SheetRef{}, "invalid_saved_view_id"
			}
			return SheetRef{}, "invalid_sheet_ref"
		}
		return SheetRef{Kind: kind, ID: strings.TrimSpace(id)}, ""
	case "extension_workspace":
		if !sheetRefHasOnlyMembers(object, "kind", "extension_profile_id", "workspace_key") {
			return SheetRef{}, "invalid_sheet_ref"
		}
		var profileID string
		if rawProfileID, ok := object["extension_profile_id"]; !ok || json.Unmarshal(rawProfileID, &profileID) != nil || !validExtensionToken(strings.TrimSpace(profileID)) {
			return SheetRef{}, "invalid_extension_profile_id"
		}
		var workspaceKey string
		if rawWorkspaceKey, ok := object["workspace_key"]; !ok || json.Unmarshal(rawWorkspaceKey, &workspaceKey) != nil || !validExtensionToken(strings.TrimSpace(workspaceKey)) {
			return SheetRef{}, "invalid_extension_workspace_key"
		}
		return SheetRef{
			Kind:               kind,
			ExtensionProfileID: strings.TrimSpace(profileID),
			WorkspaceKey:       strings.TrimSpace(workspaceKey),
		}, ""
	default:
		return SheetRef{}, "unsupported_sheet_ref_kind"
	}
}

func sheetRefHasOnlyMembers(object map[string]json.RawMessage, allowed ...string) bool {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	if len(object) != len(allowedSet) {
		return false
	}
	for key := range object {
		if _, ok := allowedSet[key]; !ok {
			return false
		}
	}
	return true
}

var extensionTokenPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

func validExtensionToken(value string) bool {
	return extensionTokenPattern.MatchString(value)
}

func mustSheetRefJSON(ref SheetRef) []byte {
	payload, err := json.Marshal(ref)
	if err != nil {
		panic(err)
	}
	return payload
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

func invalidStartupRequest(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_startup_request",
		Message: "invalid workbook startup request",
		Details: details,
	}
}
