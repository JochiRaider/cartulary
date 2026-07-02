package startup

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
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
	Kind string `json:"kind"`
	ID   string `json:"id"`
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
	IncidentID           uuid.UUID
	SelectedSheetRef     []byte
	SelectedViewSchemaID string
	SelectedSavedView    *SavedViewRecord
	Source               string
	ClearedPointers      []ClearedPointer
	HomeSheetRef         []byte
	DefaultSheetRef      []byte
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
	viewSchemaID := strings.TrimSpace(query.Get("view_schema_id"))
	sheetRefKind := strings.TrimSpace(query.Get("sheet_ref_kind"))
	sheetRefID := strings.TrimSpace(query.Get("sheet_ref_id"))
	hasViewSchema := viewSchemaID != ""
	hasSheetRef := sheetRefKind != "" || sheetRefID != ""
	if hasViewSchema && hasSheetRef {
		return nil, invalidStartupRequest("sheet_ref", "ambiguous_explicit_sheet_ref")
	}
	for key := range query {
		switch key {
		case "view_schema_id", "sheet_ref_kind", "sheet_ref_id":
		default:
			return nil, invalidStartupRequest(key, "unknown_field")
		}
	}
	if hasViewSchema {
		return canonicalStartupSheetRef(SheetRef{Kind: "view_schema", ID: viewSchemaID})
	}
	if !hasSheetRef {
		return nil, nil
	}
	if sheetRefKind == "" {
		return nil, invalidStartupRequest("sheet_ref_kind", "missing_required_field")
	}
	if sheetRefID == "" {
		return nil, invalidStartupRequest("sheet_ref_id", "missing_required_field")
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
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
}

func canonicalSheetRef(value json.RawMessage, field string) ([]byte, *httpapi.APIError) {
	if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return nil, nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(value, &object); err != nil {
		return nil, invalidMutationPayload(field, "invalid_sheet_ref")
	}
	for key := range object {
		if key != "kind" && key != "id" {
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
	if apiErr := resolveSheetRef(kind, id, field); apiErr != nil {
		return nil, apiErr
	}
	return mustSheetRefJSON(SheetRef{Kind: kind, ID: id}), nil
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
	switch ref.Kind {
	case "view_schema":
	case "saved_view":
		if _, err := uuid.Parse(ref.ID); err != nil {
			return nil, invalidStartupRequest("sheet_ref_id", "invalid_saved_view_id")
		}
	default:
		return nil, invalidStartupRequest("sheet_ref_kind", "unsupported_sheet_ref_kind")
	}
	if ref.ID == "" {
		return nil, invalidStartupRequest("sheet_ref_id", "missing_required_field")
	}
	return mustSheetRefJSON(ref), nil
}

func decodeSheetRef(raw []byte) (SheetRef, error) {
	var ref SheetRef
	if len(raw) == 0 {
		return ref, errEmptySheetRef
	}
	if err := json.Unmarshal(raw, &ref); err != nil {
		return ref, err
	}
	ref.Kind = strings.TrimSpace(ref.Kind)
	ref.ID = strings.TrimSpace(ref.ID)
	if ref.Kind == "" || ref.ID == "" {
		return ref, errMissingSheetRefMember
	}
	return ref, nil
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
