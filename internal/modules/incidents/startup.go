package incidents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	startupTimelineViewSchemaID = "cartulary.view.timeline.v2"
	startupSourceExplicit       = "explicit"
	startupSourceHome           = "home"
	startupSourceDefault        = "default"
	startupSourceTimeline       = "timeline"
)

type StartupCandidate struct {
	SheetRef map[string]string
	Valid    bool
}

type WorkbookSheetRef struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type StartupSavedViewRecord struct {
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

type ClearedStartupPointer struct {
	Source     string
	SheetRef   []byte
	ReasonCode string
}

type WorkbookStartupRecord struct {
	IncidentID           uuid.UUID
	SelectedSheetRef     []byte
	SelectedViewSchemaID string
	SelectedSavedView    *StartupSavedViewRecord
	Source               string
	ClearedPointers      []ClearedStartupPointer
	HomeSheetRef         []byte
	DefaultSheetRef      []byte
}

func SelectStartupSheet(explicit StartupCandidate, home StartupCandidate, defaults StartupCandidate) (map[string]string, []string) {
	cleared := []string{}
	for _, candidate := range []struct {
		name string
		ref  StartupCandidate
	}{
		{name: startupSourceExplicit, ref: explicit},
		{name: startupSourceHome, ref: home},
		{name: startupSourceDefault, ref: defaults},
	} {
		if candidate.ref.SheetRef == nil {
			continue
		}
		if candidate.ref.Valid {
			return cloneSheetRef(candidate.ref.SheetRef), cleared
		}
		cleared = append(cleared, candidate.name)
	}
	return map[string]string{"kind": "view_schema", "id": startupTimelineViewSchemaID}, cleared
}

func ParseStartupExplicitSheetRef(query url.Values) ([]byte, *httpapi.APIError) {
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
		return canonicalStartupSheetRef(WorkbookSheetRef{Kind: "view_schema", ID: viewSchemaID})
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
	return canonicalStartupSheetRef(WorkbookSheetRef{Kind: sheetRefKind, ID: sheetRefID})
}

func (s *Store) ResolveWorkbookStartup(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, membershipRole string, explicitSheetRef []byte, now time.Time) (WorkbookStartupRecord, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkbookStartupRecord{}, fmt.Errorf("begin workbook startup transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	homeRef, err := getUserWorkbookPreferenceRefTx(ctx, tx, incidentID, userID)
	if err != nil {
		return WorkbookStartupRecord{}, err
	}
	defaultRef, err := getDefaultWorkbookPreferenceRefTx(ctx, tx, incidentID)
	if err != nil {
		return WorkbookStartupRecord{}, err
	}

	record := WorkbookStartupRecord{
		IncidentID:      incidentID,
		HomeSheetRef:    cloneBytes(homeRef),
		DefaultSheetRef: cloneBytes(defaultRef),
	}
	for _, candidate := range []struct {
		source string
		ref    []byte
	}{
		{source: startupSourceExplicit, ref: explicitSheetRef},
		{source: startupSourceHome, ref: homeRef},
		{source: startupSourceDefault, ref: defaultRef},
	} {
		if len(candidate.ref) == 0 {
			continue
		}
		resolved, invalidReason, err := s.resolveStartupCandidate(ctx, tx, incidentID, userID, membershipRole, candidate.ref)
		if err != nil {
			return WorkbookStartupRecord{}, err
		}
		if invalidReason == "" {
			record.SelectedSheetRef = cloneBytes(candidate.ref)
			record.SelectedViewSchemaID = resolved.ViewSchemaID
			record.SelectedSavedView = resolved.SavedView
			record.Source = candidate.source
			if err := tx.Commit(ctx); err != nil {
				return WorkbookStartupRecord{}, fmt.Errorf("commit workbook startup transaction: %w", err)
			}
			return record, nil
		}
		if candidate.source == startupSourceExplicit {
			continue
		}
		record.ClearedPointers = append(record.ClearedPointers, ClearedStartupPointer{
			Source:     candidate.source,
			SheetRef:   cloneBytes(candidate.ref),
			ReasonCode: invalidReason,
		})
		if candidate.source == startupSourceHome {
			if err := clearUserWorkbookPreferenceRefTx(ctx, tx, incidentID, userID, now); err != nil {
				return WorkbookStartupRecord{}, err
			}
			homeRef = nil
			record.HomeSheetRef = nil
		}
		if candidate.source == startupSourceDefault {
			if err := clearDefaultWorkbookPreferenceRefTx(ctx, tx, incidentID, now); err != nil {
				return WorkbookStartupRecord{}, err
			}
			defaultRef = nil
			record.DefaultSheetRef = nil
		}
	}

	record.SelectedSheetRef = mustSheetRefJSON(WorkbookSheetRef{Kind: "view_schema", ID: startupTimelineViewSchemaID})
	record.SelectedViewSchemaID = startupTimelineViewSchemaID
	record.Source = startupSourceTimeline
	if err := tx.Commit(ctx); err != nil {
		return WorkbookStartupRecord{}, fmt.Errorf("commit workbook startup fallback transaction: %w", err)
	}
	return record, nil
}

type resolvedStartupCandidate struct {
	ViewSchemaID string
	SavedView    *StartupSavedViewRecord
}

func (s *Store) resolveStartupCandidate(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID, membershipRole string, raw []byte) (resolvedStartupCandidate, string, error) {
	ref, err := decodeWorkbookSheetRef(raw)
	if err != nil {
		return resolvedStartupCandidate{}, "invalid_sheet_ref", nil
	}
	switch ref.Kind {
	case "view_schema":
		if _, ok := viewschema.Lookup(ref.ID); !ok {
			return resolvedStartupCandidate{}, "unknown_view_schema", nil
		}
		return resolvedStartupCandidate{ViewSchemaID: ref.ID}, "", nil
	case "saved_view":
		savedViewID, err := uuid.Parse(ref.ID)
		if err != nil {
			return resolvedStartupCandidate{}, "invalid_saved_view_id", nil
		}
		savedView, found, err := getStartupSavedViewTx(ctx, tx, incidentID, savedViewID)
		if err != nil {
			return resolvedStartupCandidate{}, "", err
		}
		if !found {
			return resolvedStartupCandidate{}, "saved_view_not_found", nil
		}
		if !startupSavedViewVisible(savedView, userID, membershipRole) {
			return resolvedStartupCandidate{}, "saved_view_not_visible", nil
		}
		if _, ok := viewschema.Lookup(savedView.ViewSchemaID); !ok {
			return resolvedStartupCandidate{}, "unknown_view_schema", nil
		}
		return resolvedStartupCandidate{ViewSchemaID: savedView.ViewSchemaID, SavedView: &savedView}, "", nil
	default:
		return resolvedStartupCandidate{}, "unsupported_sheet_ref_kind", nil
	}
}

func getUserWorkbookPreferenceRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID) ([]byte, error) {
	var ref []byte
	err := tx.QueryRow(ctx, `
	SELECT home_sheet_ref
	  FROM user_workbook_preferences
	 WHERE incident_id = $1
	   AND user_id = $2
	 FOR UPDATE
	`, incidentID, userID).Scan(&ref)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query startup user preference: %w", err)
	}
	return ref, nil
}

func getDefaultWorkbookPreferenceRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) ([]byte, error) {
	var ref []byte
	err := tx.QueryRow(ctx, `
	SELECT default_sheet_ref
	  FROM incident_workbook_preferences
	 WHERE incident_id = $1
	 FOR UPDATE
	`, incidentID).Scan(&ref)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query startup default preference: %w", err)
	}
	return ref, nil
}

func clearUserWorkbookPreferenceRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `
	UPDATE user_workbook_preferences
	   SET home_sheet_ref = NULL,
	       updated_at = $3
	 WHERE incident_id = $1
	   AND user_id = $2
	`, incidentID, userID, now.UTC()); err != nil {
		return fmt.Errorf("clear startup home pointer: %w", err)
	}
	return nil
}

func clearDefaultWorkbookPreferenceRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `
	UPDATE incident_workbook_preferences
	   SET default_sheet_ref = NULL,
	       updated_at = $2
	 WHERE incident_id = $1
	`, incidentID, now.UTC()); err != nil {
		return fmt.Errorf("clear startup default pointer: %w", err)
	}
	return nil
}

func getStartupSavedViewTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, savedViewID uuid.UUID) (StartupSavedViewRecord, bool, error) {
	var record StartupSavedViewRecord
	err := tx.QueryRow(ctx, `
	SELECT saved_view_id, incident_id, view_schema_id, scope, display_name, query_json, layout_json,
	       owner_user_id, created_at, updated_at, saved_view_version
	  FROM saved_views
	 WHERE incident_id = $1
	   AND saved_view_id = $2
	`, incidentID, savedViewID).Scan(
		&record.SavedViewID,
		&record.IncidentID,
		&record.ViewSchemaID,
		&record.Scope,
		&record.DisplayName,
		&record.QueryJSON,
		&record.LayoutJSON,
		&record.OwnerUserID,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.SavedViewVersion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return StartupSavedViewRecord{}, false, nil
	}
	if err != nil {
		return StartupSavedViewRecord{}, false, fmt.Errorf("query startup saved view: %w", err)
	}
	return record, true, nil
}

func startupSavedViewVisible(record StartupSavedViewRecord, userID uuid.UUID, membershipRole string) bool {
	if record.Scope == "shared" || record.Scope == "system" || membershipRole == "admin" {
		return true
	}
	return record.OwnerUserID != nil && *record.OwnerUserID == userID
}

func canonicalStartupSheetRef(ref WorkbookSheetRef) ([]byte, *httpapi.APIError) {
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

func decodeWorkbookSheetRef(raw []byte) (WorkbookSheetRef, error) {
	var ref WorkbookSheetRef
	if len(raw) == 0 {
		return ref, errors.New("empty sheet ref")
	}
	if err := json.Unmarshal(raw, &ref); err != nil {
		return ref, err
	}
	ref.Kind = strings.TrimSpace(ref.Kind)
	ref.ID = strings.TrimSpace(ref.ID)
	if ref.Kind == "" || ref.ID == "" {
		return ref, errors.New("missing sheet ref member")
	}
	return ref, nil
}

func mustSheetRefJSON(ref WorkbookSheetRef) []byte {
	payload, err := json.Marshal(ref)
	if err != nil {
		panic(err)
	}
	return payload
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
		Message: "invalid startup request",
		Details: details,
	}
}

func cloneSheetRef(value map[string]string) map[string]string {
	if value == nil {
		return nil
	}
	cloned := make(map[string]string, len(value))
	for key, item := range value {
		cloned[key] = item
	}
	return cloned
}

func cloneBytes(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	cloned := make([]byte, len(value))
	copy(cloned, value)
	return cloned
}
