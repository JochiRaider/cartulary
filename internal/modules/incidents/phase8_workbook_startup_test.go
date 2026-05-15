package incidents_test

import (
	"context"
	"database/sql"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestPhase8_WorkbookPreferencePointers_U_8_05(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase8-workbook-prefs-u-8-05")
	adminLogin, adminID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-u-8-05-incident",
		"incident_key":  "IR-U805",
		"title":         "Phase 8 workbook preferences",
	})
	incidentID := incident["incident_id"].(string)

	viewerID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase8-u805-viewer@example.test", "Phase8 U805 Viewer", "Phase8U805Viewer1!", false, false, true)
	viewerSession, viewerCSRF := phase2test.LoginLocalUser(t, harness.Server, "phase8-u805-viewer@example.test", "Phase8U805Viewer1!")
	secondAdminID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase8-u805-admin2@example.test", "Phase8 U805 Admin2", "Phase8U805Admin21!", false, false, true)
	secondAdminSession, secondAdminCSRF := phase2test.LoginLocalUser(t, harness.Server, "phase8-u805-admin2@example.test", "Phase8U805Admin21!")
	otherID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase8-u805-other@example.test", "Phase8 U805 Other", "Phase8U805Other1!", false, false, true)
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-u-8-05-viewer-membership", "user_id": viewerID, "role": "viewer"})
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-u-8-05-admin2-membership", "user_id": secondAdminID, "role": "admin"})

	homeSavedViewID := "00000000-0000-0000-0000-000000008501"
	defaultSavedViewID := "00000000-0000-0000-0000-000000008502"
	hiddenSavedViewID := "00000000-0000-0000-0000-000000008503"
	seedPhase8SavedView(t, harness.DB, homeSavedViewID, incidentID, timeline.TimelineViewSchemaID, "private", "Viewer home", viewerID, "2026-05-14T18:00:00Z")
	seedPhase8SavedView(t, harness.DB, defaultSavedViewID, incidentID, "cartulary.view.hosts.v1", "shared", "Incident default", adminID, "2026-05-14T18:01:00Z")
	seedPhase8SavedView(t, harness.DB, hiddenSavedViewID, incidentID, "cartulary.view.identities.v1", "private", "Hidden other-user home", otherID, "2026-05-14T18:02:00Z")

	viewerHome := putUserWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, viewerSession, viewerCSRF, map[string]any{
		"home_sheet_ref": map[string]any{"kind": "saved_view", "id": homeSavedViewID},
	})
	requireSheetRef(t, viewerHome["home_sheet_ref"], "saved_view", homeSavedViewID)

	adminDefaultBefore := getDefaultWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, adminLogin.SessionCookie)
	if adminDefaultBefore["default_sheet_ref"] != nil {
		t.Fatalf("caller-owned preference route must not mutate default_sheet_ref: %#v", adminDefaultBefore)
	}

	adminDefault := putDefaultWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, adminLogin.SessionCookie, adminLogin.CSRFCookie, map[string]any{
		"default_sheet_ref": map[string]any{"kind": "saved_view", "id": defaultSavedViewID},
	})
	requireSheetRef(t, adminDefault["default_sheet_ref"], "saved_view", defaultSavedViewID)
	if adminDefault["updated_by_user_id"] != adminID {
		t.Fatalf("default preference update must attribute admin actor: %#v", adminDefault)
	}

	viewerHomeAfterDefault := getUserWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, viewerSession)
	requireSheetRef(t, viewerHomeAfterDefault["home_sheet_ref"], "saved_view", homeSavedViewID)

	viewerDefault := phase2test.DoJSON(
		t,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		map[string]any{"default_sheet_ref": map[string]any{"kind": "view_schema", "id": timeline.TimelineViewSchemaID}},
		phase2test.WithCookies(viewerSession, viewerCSRF),
		phase2test.WithHeader(authn.CSRFHeaderName, viewerCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, viewerDefault, http.StatusForbidden, "authorization_denied")
	defaultAfterDenied := getDefaultWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, adminLogin.SessionCookie)
	requireSheetRef(t, defaultAfterDenied["default_sheet_ref"], "saved_view", defaultSavedViewID)

	viewerHomeNoOp := putUserWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, viewerSession, viewerCSRF, map[string]any{
		"home_sheet_ref": map[string]any{"id": homeSavedViewID, "kind": "saved_view"},
	})
	if viewerHomeNoOp["updated_at"] != viewerHome["updated_at"] {
		t.Fatalf("structurally valid user preference no-op must preserve updated_at: before=%#v after=%#v", viewerHome, viewerHomeNoOp)
	}

	defaultNoOp := putDefaultWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, secondAdminSession, secondAdminCSRF, map[string]any{
		"default_sheet_ref": map[string]any{"id": defaultSavedViewID, "kind": "saved_view"},
	})
	if defaultNoOp["updated_at"] != adminDefault["updated_at"] || defaultNoOp["updated_by_user_id"] != adminID {
		t.Fatalf("structurally valid default preference no-op must preserve updated_at and updated_by_user_id: before=%#v after=%#v", adminDefault, defaultNoOp)
	}

	putUserWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, viewerSession, viewerCSRF, map[string]any{
		"home_sheet_ref": map[string]any{"kind": "saved_view", "id": hiddenSavedViewID},
	})
	hiddenHomeStartup := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, viewerSession)
	requireStartupSelection(t, hiddenHomeStartup, "default", "saved_view", defaultSavedViewID, "cartulary.view.hosts.v1")
	requireClearedPointers(t, hiddenHomeStartup, map[string]string{"source": "home", "reason_code": "saved_view_not_visible"})
	if homePrefs := getUserWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, viewerSession); homePrefs["home_sheet_ref"] != nil {
		t.Fatalf("startup must persistently clear hidden home pointer before fallback continues: %#v", homePrefs)
	}
}

func TestPhase8_WorkbookStartupFallback_I_8_02(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase8-workbook-startup-i-8-02")
	adminLogin, adminID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-i-8-02-incident",
		"incident_key":  "IR-I802",
		"title":         "Phase 8 workbook startup",
	})
	incidentID := incident["incident_id"].(string)

	viewerID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase8-i802-viewer@example.test", "Phase8 I802 Viewer", "Phase8I802Viewer1!", false, false, true)
	viewerSession, viewerCSRF := phase2test.LoginLocalUser(t, harness.Server, "phase8-i802-viewer@example.test", "Phase8I802Viewer1!")
	otherID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase8-i802-other@example.test", "Phase8 I802 Other", "Phase8I802Other1!", false, false, true)
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-i-8-02-viewer-membership", "user_id": viewerID, "role": "viewer"})

	homeSavedViewID := "00000000-0000-0000-0000-000000008601"
	defaultSavedViewID := "00000000-0000-0000-0000-000000008602"
	hiddenSavedViewID := "00000000-0000-0000-0000-000000008603"
	invalidSchemaSavedViewID := "00000000-0000-0000-0000-000000008604"
	deletedSavedViewID := "00000000-0000-0000-0000-000000008605"
	missingSavedViewID := "00000000-0000-0000-0000-000000008606"
	seedPhase8SavedView(t, harness.DB, homeSavedViewID, incidentID, "cartulary.view.hosts.v1", "private", "Viewer hosts", viewerID, "2026-05-14T19:00:00Z")
	seedPhase8SavedView(t, harness.DB, defaultSavedViewID, incidentID, "cartulary.view.evidence.v1", "shared", "Shared evidence", adminID, "2026-05-14T19:01:00Z")
	seedPhase8SavedView(t, harness.DB, hiddenSavedViewID, incidentID, "cartulary.view.identities.v1", "private", "Other private hidden", otherID, "2026-05-14T19:02:00Z")
	seedPhase8SavedView(t, harness.DB, invalidSchemaSavedViewID, incidentID, "cartulary.view.unknown.v1", "shared", "Invalid schema", adminID, "2026-05-14T19:03:00Z")
	seedPhase8SavedView(t, harness.DB, deletedSavedViewID, incidentID, timeline.TimelineViewSchemaID, "shared", "Deleted default", adminID, "2026-05-14T19:04:00Z")

	putUserWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, viewerSession, viewerCSRF, map[string]any{
		"home_sheet_ref": map[string]any{"kind": "saved_view", "id": homeSavedViewID},
	})
	putDefaultWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, adminLogin.SessionCookie, adminLogin.CSRFCookie, map[string]any{
		"default_sheet_ref": map[string]any{"kind": "saved_view", "id": defaultSavedViewID},
	})

	explicit := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, "view_schema_id="+timeline.TimelineViewSchemaID, viewerSession)
	requireStartupSelection(t, explicit, "explicit", "view_schema", timeline.TimelineViewSchemaID, timeline.TimelineViewSchemaID)
	requireClearedPointers(t, explicit)

	home := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, viewerSession)
	requireStartupSelection(t, home, "home", "saved_view", homeSavedViewID, "cartulary.view.hosts.v1")
	if home["selected_saved_view"] == nil {
		t.Fatalf("saved-view startup selection must include selected_saved_view: %#v", home)
	}

	putUserWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, viewerSession, viewerCSRF, map[string]any{
		"home_sheet_ref": map[string]any{"kind": "saved_view", "id": hiddenSavedViewID},
	})
	hiddenHome := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, viewerSession)
	requireStartupSelection(t, hiddenHome, "default", "saved_view", defaultSavedViewID, "cartulary.view.evidence.v1")
	requireClearedPointers(t, hiddenHome, map[string]string{"source": "home", "reason_code": "saved_view_not_visible"})
	if homePrefs := getUserWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, viewerSession); homePrefs["home_sheet_ref"] != nil {
		t.Fatalf("hidden home pointer must be cleared before fallback continues: %#v", homePrefs)
	}

	putUserWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, viewerSession, viewerCSRF, map[string]any{
		"home_sheet_ref": map[string]any{"kind": "saved_view", "id": invalidSchemaSavedViewID},
	})
	invalidHome := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, viewerSession)
	requireStartupSelection(t, invalidHome, "default", "saved_view", defaultSavedViewID, "cartulary.view.evidence.v1")
	requireClearedPointers(t, invalidHome, map[string]string{"source": "home", "reason_code": "unknown_view_schema"})

	putUserWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, viewerSession, viewerCSRF, map[string]any{"home_sheet_ref": nil})
	putDefaultWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, adminLogin.SessionCookie, adminLogin.CSRFCookie, map[string]any{
		"default_sheet_ref": map[string]any{"kind": "saved_view", "id": deletedSavedViewID},
	})
	if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM saved_views WHERE saved_view_id = $1::uuid`, deletedSavedViewID); err != nil {
		t.Fatalf("delete saved view used by default pointer: %v", err)
	}
	deletedDefault := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, viewerSession)
	requireStartupSelection(t, deletedDefault, "timeline", "view_schema", timeline.TimelineViewSchemaID, timeline.TimelineViewSchemaID)
	requireClearedPointers(t, deletedDefault, map[string]string{"source": "default", "reason_code": "saved_view_not_found"})
	if defaultPrefs := getDefaultWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, adminLogin.SessionCookie); defaultPrefs["default_sheet_ref"] != nil {
		t.Fatalf("deleted default pointer must be cleared before fallback continues: %#v", defaultPrefs)
	}

	putDefaultWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, adminLogin.SessionCookie, adminLogin.CSRFCookie, map[string]any{
		"default_sheet_ref": map[string]any{"kind": "view_schema", "id": "cartulary.view.task_requests.v1"},
	})
	baseSurface := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, viewerSession)
	requireStartupSelection(t, baseSurface, "default", "view_schema", "cartulary.view.task_requests.v1", "cartulary.view.task_requests.v1")

	explicitInvalid := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, "sheet_ref_kind=saved_view&sheet_ref_id="+missingSavedViewID, viewerSession)
	requireStartupSelection(t, explicitInvalid, "default", "view_schema", "cartulary.view.task_requests.v1", "cartulary.view.task_requests.v1")
	requireClearedPointers(t, explicitInvalid)

	dualSelector := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-startup?view_schema_id="+timeline.TimelineViewSchemaID+"&sheet_ref_kind=view_schema&sheet_ref_id=cartulary.view.hosts.v1",
		nil,
		phase2test.WithCookies(viewerSession),
	)
	httptestx.RequireErrorEnvelope(t, dualSelector, http.StatusBadRequest, "invalid_startup_request")
}

func TestPhase8_WorkbookStartupBaseSurfaceDoesNotRequireSavedView_I_8_02(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase8-workbook-startup-base-surface-i-8-02")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-i-8-02-base-incident",
		"incident_key":  "IR-I802-BASE",
		"title":         "Phase 8 base startup",
	})
	incidentID := incident["incident_id"].(string)

	viewerID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase8-i802-base-viewer@example.test", "Phase8 I802 Base Viewer", "Phase8I802BaseViewer1!", false, false, true)
	viewerSession, _ := phase2test.LoginLocalUser(t, harness.Server, "phase8-i802-base-viewer@example.test", "Phase8I802BaseViewer1!")
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-i-8-02-base-viewer-membership", "user_id": viewerID, "role": "viewer"})
	putDefaultWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, adminLogin.SessionCookie, adminLogin.CSRFCookie, map[string]any{
		"default_sheet_ref": map[string]any{"kind": "view_schema", "id": "cartulary.view.task_requests.v1"},
	})

	listResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views", nil, phase2test.WithCookies(viewerSession))
	savedViews := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)["saved_views"].([]any)
	if len(savedViews) != 0 {
		t.Fatalf("test setup expected no saved-view resources, got %#v", savedViews)
	}
	startup := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, viewerSession)
	requireStartupSelection(t, startup, "default", "view_schema", "cartulary.view.task_requests.v1", "cartulary.view.task_requests.v1")
}

func seedPhase8SavedView(t testing.TB, db *sql.DB, savedViewID string, incidentID string, viewSchemaID string, scope string, name string, ownerUserID string, updatedAt string) {
	t.Helper()
	ownerExpr := any(nil)
	if ownerUserID != "" {
		ownerExpr = ownerUserID
	}
	ts, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		t.Fatalf("parse seed timestamp: %v", err)
	}
	layoutJSON := []byte(`{"layout_schema_id":"cartulary.layout.v1","column_order":[],"hidden_field_keys":[],"column_widths":[]}`)
	if _, ok := viewschema.Lookup(viewSchemaID); ok {
		layout, layoutErr := viewschema.DefaultLayout(viewSchemaID)
		if layoutErr != nil {
			t.Fatalf("build seed saved-view layout: %+v", layoutErr)
		}
		layoutJSON = layout
	}
	if _, err := db.ExecContext(context.Background(), `
	INSERT INTO saved_views (
	    saved_view_id, incident_id, view_schema_id, scope, display_name, query_json, layout_json,
	    owner_user_id, created_at, updated_at, saved_view_version
	)
	VALUES ($1, $2, $3, $4, $5, '{"sort":[],"filters":[]}'::jsonb, $6::jsonb, $7, $8, $8, 1)
	`, savedViewID, incidentID, viewSchemaID, scope, name, layoutJSON, ownerExpr, ts); err != nil {
		t.Fatalf("seed saved view %s: %v", name, err)
	}
}

func putUserWorkbookPreferences(t testing.TB, baseURL string, incidentID string, session *http.Cookie, csrf *http.Cookie, body map[string]any) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(
		t,
		http.MethodPut,
		baseURL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me",
		body,
		phase2test.WithCookies(session, csrf),
		phase2test.WithHeader(authn.CSRFHeaderName, csrf.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func putDefaultWorkbookPreferences(t testing.TB, baseURL string, incidentID string, session *http.Cookie, csrf *http.Cookie, body map[string]any) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(
		t,
		http.MethodPut,
		baseURL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		body,
		phase2test.WithCookies(session, csrf),
		phase2test.WithHeader(authn.CSRFHeaderName, csrf.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func getUserWorkbookPreferences(t testing.TB, baseURL string, incidentID string, session *http.Cookie) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(t, http.MethodGet, baseURL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, phase2test.WithCookies(session))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func getDefaultWorkbookPreferences(t testing.TB, baseURL string, incidentID string, session *http.Cookie) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(t, http.MethodGet, baseURL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default", nil, phase2test.WithCookies(session))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func getWorkbookStartup(t testing.TB, baseURL string, incidentID string, args ...any) map[string]any {
	t.Helper()
	query := ""
	session := args[len(args)-1].(*http.Cookie)
	if len(args) == 2 {
		query = "?" + args[0].(string)
	}
	resp := phase2test.DoJSON(t, http.MethodGet, baseURL+"/api/v1/incidents/"+incidentID+"/workbook-startup"+query, nil, phase2test.WithCookies(session))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected startup status: got %d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func requireSheetRef(t testing.TB, value any, wantKind string, wantID string) {
	t.Helper()
	ref, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected sheet_ref object, got %#v", value)
	}
	if ref["kind"] != wantKind || ref["id"] != wantID {
		t.Fatalf("unexpected sheet_ref: got %#v want kind=%q id=%q", ref, wantKind, wantID)
	}
}

func requireStartupSelection(t testing.TB, data map[string]any, wantSource string, wantKind string, wantID string, wantViewSchemaID string) {
	t.Helper()
	if data["source"] != wantSource || data["selected_view_schema_id"] != wantViewSchemaID {
		t.Fatalf("unexpected startup selection: got %#v want source=%q view_schema_id=%q", data, wantSource, wantViewSchemaID)
	}
	requireSheetRef(t, data["selected_sheet_ref"], wantKind, wantID)
}

func requireClearedPointers(t testing.TB, data map[string]any, want ...map[string]string) {
	t.Helper()
	if want == nil {
		want = []map[string]string{}
	}
	raw, ok := data["cleared_pointers"].([]any)
	if !ok {
		t.Fatalf("cleared_pointers must be an array: %#v", data)
	}
	got := make([]map[string]string, 0, len(raw))
	for _, item := range raw {
		entry := item.(map[string]any)
		got = append(got, map[string]string{
			"source":      entry["source"].(string),
			"reason_code": entry["reason_code"].(string),
		})
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected cleared pointers: got %#v want %#v in payload %#v", got, want, data)
	}
}
