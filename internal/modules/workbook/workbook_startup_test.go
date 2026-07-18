package workbook_test

import (
	"context"
	"database/sql"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestWorkbookPreferencePointers_Unit(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase8-workbook-prefs-u-8-05")
	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-u-8-05-incident",
		"incident_key":  "IR-U805",
		"title":         "Workbook query workbook preferences",
	})
	incidentID := incident["incident_id"].(string)

	viewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "phase8-u805-viewer@example.test", "Phase8 U805 Viewer", "Phase8U805Viewer1!", false, false, true)
	viewerSession, viewerCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "phase8-u805-viewer@example.test", "Phase8U805Viewer1!", nil)
	secondAdminID := flowtest.SeedLocalUserFlags(t, harness.DB, "phase8-u805-admin2@example.test", "Phase8 U805 Admin2", "Phase8U805Admin21!", false, false, true)
	secondAdminSession, secondAdminCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "phase8-u805-admin2@example.test", "Phase8U805Admin21!", nil)
	otherID := flowtest.SeedLocalUserFlags(t, harness.DB, "phase8-u805-other@example.test", "Phase8 U805 Other", "Phase8U805Other1!", false, false, true)
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-u-8-05-viewer-membership", "user_id": viewerID, "role": "viewer"})
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-u-8-05-admin2-membership", "user_id": secondAdminID, "role": "admin"})

	homeSavedViewID := "00000000-0000-0000-0000-000000008501"
	defaultSavedViewID := "00000000-0000-0000-0000-000000008502"
	hiddenSavedViewID := "00000000-0000-0000-0000-000000008503"
	seedSavedView(t, harness.DB, homeSavedViewID, incidentID, timeline.TimelineViewSchemaID, "private", "Viewer home", viewerID, "2026-05-14T18:00:00Z")
	seedSavedView(t, harness.DB, defaultSavedViewID, incidentID, "cartulary.view.hosts.v1", "shared", "Incident default", adminID, "2026-05-14T18:01:00Z")
	seedSavedView(t, harness.DB, hiddenSavedViewID, incidentID, "cartulary.view.identities.v1", "private", "Hidden other-user home", otherID, "2026-05-14T18:02:00Z")

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

	viewerDefault := httptestx.DoJSON(
		t,
		http.MethodPut,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		map[string]any{"default_sheet_ref": map[string]any{"kind": "view_schema", "id": timeline.TimelineViewSchemaID}},
		httptestx.WithCookies(viewerSession, viewerCSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, viewerCSRF.Value),
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

func TestWorkbookStartupFallback_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	profiles := httpapi.CurrentExtensionProfiles()
	for index := range profiles {
		if profiles[index].ProfileID == "network_flow_activity" {
			profiles[index].Claimed = true
		}
	}
	harness := runtime.StartServerWithDependencies(t, "phase8-workbook-startup-i-8-02", httpapi.DependencySet{
		ExtensionProfiles: profiles,
		ModuleOverrides: map[string]any{
			networkflow.KeyRingsOverrideKey: NetworkFlowHarnessKeyRings(t),
		},
	})
	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-i-8-02-incident",
		"incident_key":  "IR-I802",
		"title":         "Workbook query workbook startup",
	})
	incidentID := incident["incident_id"].(string)

	viewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "phase8-i802-viewer@example.test", "Phase8 I802 Viewer", "Phase8I802Viewer1!", false, false, true)
	viewerSession, viewerCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "phase8-i802-viewer@example.test", "Phase8I802Viewer1!", nil)
	otherID := flowtest.SeedLocalUserFlags(t, harness.DB, "phase8-i802-other@example.test", "Phase8 I802 Other", "Phase8I802Other1!", false, false, true)
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-i-8-02-viewer-membership", "user_id": viewerID, "role": "viewer"})

	homeSavedViewID := "00000000-0000-0000-0000-000000008601"
	defaultSavedViewID := "00000000-0000-0000-0000-000000008602"
	hiddenSavedViewID := "00000000-0000-0000-0000-000000008603"
	invalidSchemaSavedViewID := "00000000-0000-0000-0000-000000008604"
	deletedSavedViewID := "00000000-0000-0000-0000-000000008605"
	missingSavedViewID := "00000000-0000-0000-0000-000000008606"
	seedSavedView(t, harness.DB, homeSavedViewID, incidentID, "cartulary.view.hosts.v1", "private", "Viewer hosts", viewerID, "2026-05-14T19:00:00Z")
	seedSavedView(t, harness.DB, defaultSavedViewID, incidentID, "cartulary.view.evidence.v1", "shared", "Shared evidence", adminID, "2026-05-14T19:01:00Z")
	seedSavedView(t, harness.DB, hiddenSavedViewID, incidentID, "cartulary.view.identities.v1", "private", "Other private hidden", otherID, "2026-05-14T19:02:00Z")
	seedSavedView(t, harness.DB, invalidSchemaSavedViewID, incidentID, "cartulary.view.unknown.v1", "shared", "Invalid schema", adminID, "2026-05-14T19:03:00Z")
	seedSavedView(t, harness.DB, deletedSavedViewID, incidentID, timeline.TimelineViewSchemaID, "shared", "Deleted default", adminID, "2026-05-14T19:04:00Z")

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

	dualSelector := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-startup?view_schema_id="+timeline.TimelineViewSchemaID+"&sheet_ref_kind=view_schema&sheet_ref_id=cartulary.view.hosts.v1",
		nil,
		httptestx.WithCookies(viewerSession),
	)
	httptestx.RequireErrorEnvelope(t, dualSelector, http.StatusBadRequest, "invalid_startup_request")

	extensionRef := map[string]any{
		"kind":                 "extension_workspace",
		"extension_profile_id": "network_flow_activity",
		"workspace_key":        "network_analysis",
	}
	extensionHome := putUserWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, viewerSession, viewerCSRF, map[string]any{
		"home_sheet_ref": extensionRef,
	})
	requireExtensionSheetRef(t, extensionHome["home_sheet_ref"])

	extensionStartup := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, viewerSession)
	requireExtensionStartupSelection(t, extensionStartup, "home")
	var networkFlowTableCount int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM network_flow_tables WHERE incident_id = $1::uuid`, incidentID).Scan(&networkFlowTableCount); err != nil {
		t.Fatalf("count Network Flow tables: %v", err)
	}
	if networkFlowTableCount != 0 {
		t.Fatalf("extension workspace startup must not require Network Flow tables: count=%d", networkFlowTableCount)
	}

	explicitExtension := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, "sheet_ref_kind=extension_workspace&sheet_ref_id=network_analysis&extension_profile_id=network_flow_activity", viewerSession)
	requireExtensionStartupSelection(t, explicitExtension, "explicit")

	legacyWorkspaceAlias := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-startup?sheet_ref_kind=extension_workspace&sheet_ref_id=network_analysis&extension_profile_id=network_flow_activity&workspace_key=network_analysis",
		nil,
		httptestx.WithCookies(viewerSession),
	)
	httptestx.RequireErrorEnvelope(t, legacyWorkspaceAlias, http.StatusBadRequest, "invalid_startup_request")
}

func NetworkFlowHarnessKeyRings(t testing.TB) *networkflow.KeyRings {
	t.Helper()
	rings, err := networkflow.ParseKeyRings([]byte(`{
  "schema_id":"cartulary.network_flow_key_rings.v1",
  "cursor_key_ring":{"algorithm":"aes_256_gcm_v1","keys":[{"cursor_key_id":"phase8-harness-cursor","state":"active","secret_ref":{"kind":"env","name":"phase8-harness-cursor"}}]},
  "safe_digest_key_ring":{"algorithm":"hmac_sha256_v1","keys":[{"safe_digest_key_id":"phase8-harness-safe","state":"active","secret_ref":{"kind":"env","name":"phase8-harness-safe"}}]}
}`), map[string]string{
		"CARTULARY_SECRET_PHASE8_HARNESS_CURSOR": "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_PHASE8_HARNESS_SAFE":   "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}, time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("parse Workbook query Network Flow harness key rings: %v", err)
	}
	return rings
}

func TestWorkbookStartupBaseSurfaceDoesNotRequireSavedView_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase8-workbook-startup-base-surface-i-8-02")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-i-8-02-base-incident",
		"incident_key":  "IR-I802-BASE",
		"title":         "Workbook query base startup",
	})
	incidentID := incident["incident_id"].(string)

	viewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "phase8-i802-base-viewer@example.test", "Phase8 I802 Base Viewer", "Phase8I802BaseViewer1!", false, false, true)
	viewerSession, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "phase8-i802-base-viewer@example.test", "Phase8I802BaseViewer1!", nil)
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-i-8-02-base-viewer-membership", "user_id": viewerID, "role": "viewer"})
	putDefaultWorkbookPreferences(t, harness.Server.HTTP.URL, incidentID, adminLogin.SessionCookie, adminLogin.CSRFCookie, map[string]any{
		"default_sheet_ref": map[string]any{"kind": "view_schema", "id": "cartulary.view.task_requests.v1"},
	})

	listResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views", nil, httptestx.WithCookies(viewerSession))
	savedViews := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)["saved_views"].([]any)
	if len(savedViews) != 0 {
		t.Fatalf("test setup expected no saved-view resources, got %#v", savedViews)
	}
	startup := getWorkbookStartup(t, harness.Server.HTTP.URL, incidentID, viewerSession)
	requireStartupSelection(t, startup, "default", "view_schema", "cartulary.view.task_requests.v1", "cartulary.view.task_requests.v1")
}

func seedSavedView(t testing.TB, db *sql.DB, savedViewID string, incidentID string, viewSchemaID string, scope string, name string, ownerUserID string, updatedAt string) {
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
	resp := httptestx.DoJSON(
		t,
		http.MethodPut,
		baseURL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me",
		body,
		httptestx.WithCookies(session, csrf),
		httptestx.WithHeader(authn.CSRFHeaderName, csrf.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func putDefaultWorkbookPreferences(t testing.TB, baseURL string, incidentID string, session *http.Cookie, csrf *http.Cookie, body map[string]any) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(
		t,
		http.MethodPut,
		baseURL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default",
		body,
		httptestx.WithCookies(session, csrf),
		httptestx.WithHeader(authn.CSRFHeaderName, csrf.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func getUserWorkbookPreferences(t testing.TB, baseURL string, incidentID string, session *http.Cookie) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(t, http.MethodGet, baseURL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/me", nil, httptestx.WithCookies(session))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func getDefaultWorkbookPreferences(t testing.TB, baseURL string, incidentID string, session *http.Cookie) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(t, http.MethodGet, baseURL+"/api/v1/incidents/"+incidentID+"/workbook-preferences/default", nil, httptestx.WithCookies(session))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func getWorkbookStartup(t testing.TB, baseURL string, incidentID string, args ...any) map[string]any {
	t.Helper()
	query := ""
	session := args[len(args)-1].(*http.Cookie)
	if len(args) == 2 {
		query = "?" + args[0].(string)
	}
	resp := httptestx.DoJSON(t, http.MethodGet, baseURL+"/api/v1/incidents/"+incidentID+"/workbook-startup"+query, nil, httptestx.WithCookies(session))
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

func requireExtensionSheetRef(t testing.TB, value any) {
	t.Helper()
	ref, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected extension sheet_ref object, got %#v", value)
	}
	if ref["kind"] != "extension_workspace" || ref["extension_profile_id"] != "network_flow_activity" || ref["workspace_key"] != "network_analysis" {
		t.Fatalf("unexpected extension sheet_ref: %#v", ref)
	}
	if _, hasID := ref["id"]; hasID {
		t.Fatalf("extension sheet_ref must not include id: %#v", ref)
	}
}

func requireExtensionStartupSelection(t testing.TB, data map[string]any, wantSource string) {
	t.Helper()
	if data["source"] != wantSource || data["selected_view_schema_id"] != nil || data["selected_saved_view"] != nil {
		t.Fatalf("unexpected extension startup selection: got %#v want source=%q with null base identities", data, wantSource)
	}
	requireExtensionSheetRef(t, data["selected_sheet_ref"])
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
