package savedviews_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestSavedViewIndependentReadContract_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "saved-view-independent-reads")
	admin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{"client_txn_id": "svd-incident", "incident_key": "IR-SVD", "title": "Saved view discovery"})
	id := incident["incident_id"].(string)
	viewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "svd-viewer@example.test", "Viewer", "SavedViewDiscovery1!", false, false, true)
	viewer, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "svd-viewer@example.test", "SavedViewDiscovery1!", nil)
	scenariotest.CreateMembership(t, harness.Server, admin, id, map[string]any{"client_txn_id": "svd-member", "user_id": viewerID, "role": "viewer"})
	const schema = "cartulary.view.timeline.v2"
	for i := 0; i < 101; i++ {
		seedSavedView(t, harness.DB, fmt.Sprintf("00000000-0000-4000-8000-%012d", i+1), id, "cartulary.view.notes.v1", "shared", fmt.Sprintf("Other %d", i), adminID, "2026-08-02T00:00:00Z")
	}
	ids := []string{"00000000-0000-4000-8000-000000000201", "00000000-0000-4000-8000-000000000202", "00000000-0000-4000-8000-000000000203"}
	seedSavedView(t, harness.DB, ids[0], id, schema, "private", "Owner private", viewerID, "2026-08-01T00:00:00Z")
	seedSavedView(t, harness.DB, ids[1], id, schema, "system", "System", "", "2026-08-01T00:00:00Z")
	seedSavedView(t, harness.DB, ids[2], id, schema, "shared", "Shared", adminID, "2026-08-01T00:00:00Z")
	hidden := "00000000-0000-4000-8000-000000000204"
	seedSavedView(t, harness.DB, hidden, id, schema, "private", "Hidden", adminID, "2026-08-03T00:00:00Z")
	endpoint := harness.Server.HTTP.URL + "/api/v1/incidents/" + id + "/saved-views"
	read := func(suffix string, cookie *http.Cookie) map[string]any {
		response := httptestx.DoJSON(t, http.MethodGet, endpoint+suffix, nil, httptestx.WithCookies(cookie))
		if response.Header.Get("Cache-Control") != "no-store" {
			t.Fatal("read must be no-store")
		}
		return httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
	}
	unfiltered := read("", viewer)["data"].(map[string]any)["saved_views"].([]any)
	if len(unfiltered) != 100 || unfiltered[0].(map[string]any)["view_schema_id"] == schema {
		t.Fatal("fixture must place active schema beyond first unfiltered page")
	}
	first := read("?view_schema_id="+schema+"&limit=2", viewer)
	rows := first["data"].(map[string]any)["saved_views"].([]any)
	if !reflect.DeepEqual(savedViewNames(rows), []string{"Owner private", "System"}) {
		t.Fatalf("first page: %v", rows)
	}
	cursor := first["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor"].(string)
	second := read("?view_schema_id="+schema+"&cursor_token="+url.QueryEscape(cursor), viewer)
	if !reflect.DeepEqual(savedViewNames(second["data"].(map[string]any)["saved_views"].([]any)), []string{"Shared"}) {
		t.Fatal("wrong continuation")
	}
	if second["meta"].(map[string]any)["paging"].(map[string]any)["has_more"] != false {
		t.Fatal("hidden resource affected continuation")
	}
	for i, resourceID := range ids {
		got := read("/"+resourceID, viewer)["data"].(map[string]any)
		if got["saved_view_id"] != resourceID {
			t.Fatal("unrelated detail")
		}
		if i < 2 && !reflect.DeepEqual(got, rows[i]) {
			t.Fatal("list/detail envelope disagreement")
		}
	}
	missing := "00000000-0000-4000-8000-000000009999"
	for _, resourceID := range []string{hidden, missing} {
		resp := httptestx.DoJSON(t, http.MethodGet, endpoint+"/"+resourceID, nil, httptestx.WithCookies(viewer))
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusNotFound, "saved_view_not_found")
		if len(body["error"].(map[string]any)["details"].(map[string]any)) != 0 {
			t.Fatal("concealment details leaked")
		}
	}
	read("/"+hidden, admin.SessionCookie)
	// Reads do not repair an invalid home pointer. Startup retains that ownership.
	if _, err := harness.DB.ExecContext(context.Background(), `INSERT INTO user_workbook_preferences (incident_id,user_id,home_sheet_ref,updated_at) VALUES ($1,$2,jsonb_build_object('kind','saved_view','id',$3::text),now()) ON CONFLICT (incident_id,user_id) DO UPDATE SET home_sheet_ref=excluded.home_sheet_ref`, id, viewerID, missing); err != nil {
		t.Fatal(err)
	}
	read("/"+ids[0], viewer)
	var pointer string
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT home_sheet_ref->>'id' FROM user_workbook_preferences WHERE incident_id=$1 AND user_id=$2`, id, viewerID).Scan(&pointer); err != nil || pointer != missing {
		t.Fatalf("read repaired pointer: %s %v", pointer, err)
	}
	read("?view_schema_id="+schema, viewer)
	// Query errors remain public contract errors, with no identifying details.
	for _, tc := range []struct{ suffix, code, reason string }{
		{"?scope=private&scope=shared", "invalid_list_query", "duplicate_query_member"},
		{"?view_schema_id=", "invalid_list_query", "invalid_filter_value"},
		{"?view_schema_id=%FF", "invalid_saved_view_read_request", "malformed_query"},
		{"/" + ids[0] + "?limit=", "invalid_pagination_request", "pagination_not_supported"},
		{"/" + ids[0] + "?view_schema_id=" + schema, "invalid_saved_view_read_request", "unknown_query_member"},
		{"?view_schema_id=cartulary.view.notes.v1&cursor_token=" + url.QueryEscape(cursor), "invalid_pagination_request", "cursor_query_mismatch"},
	} {
		resp := httptestx.DoJSON(t, http.MethodGet, endpoint+tc.suffix, nil, httptestx.WithCookies(viewer))
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, tc.code)
		if body["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != tc.reason {
			t.Fatalf("wrong read reason for %s: %v", tc.suffix, body)
		}
	}
	httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.MethodGet, endpoint+"/"+ids[0], nil), http.StatusUnauthorized, "session_required")
	other := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{"client_txn_id": "svd-other", "incident_key": "IR-SVD-OTHER", "title": "Other incident"})
	foreign := "00000000-0000-4000-8000-000000000301"
	seedSavedView(t, harness.DB, foreign, other["incident_id"].(string), schema, "shared", "Foreign incident resource", adminID, "2026-08-01T00:00:00Z")
	httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.MethodGet, endpoint+"/"+foreign, nil, httptestx.WithCookies(viewer)), http.StatusNotFound, "saved_view_not_found")
	closed := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+id+"/close", map[string]any{"base_incident_version": incident["incident_version"], "client_txn_id": "svd-close", "reason": "Read contract closed incident coverage"}, httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value))
	httptestx.RequireSuccessEnvelope(t, closed, http.StatusOK)
	read("/"+ids[0], viewer)
	read("?view_schema_id="+schema, viewer)
	// Closing the incident does not freeze its saved configuration objects.
	created := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, endpoint,
		map[string]any{"display_name": "Closed configuration", "view_schema_id": schema, "query_json": map[string]any{}},
		httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value)), http.StatusCreated)["data"].(map[string]any)
	createdID := created["saved_view_id"].(string)
	updated := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPatch, endpoint+"/"+createdID,
		map[string]any{"display_name": "Closed configuration revised", "base_saved_view_version": created["saved_view_version"]},
		httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value)), http.StatusOK)["data"].(map[string]any)
	if !reflect.DeepEqual(read("/"+createdID, admin.SessionCookie)["data"], updated) {
		t.Fatal("closed configuration detail did not observe update")
	}
	deleted := httptestx.DoJSON(t, http.MethodDelete, endpoint+"/"+createdID,
		nil,
		httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value))
	httptestx.RequireSuccessEnvelope(t, deleted, http.StatusOK)
	httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.MethodGet, endpoint+"/"+createdID, nil, httptestx.WithCookies(admin.SessionCookie)), http.StatusNotFound, "saved_view_not_found")
	// Current role controls visibility on both routes, even on a closed incident.
	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role='admin' WHERE incident_id=$1 AND user_id=$2`, id, viewerID); err != nil {
		t.Fatal(err)
	}
	read("/"+hidden, viewer)
	elevated := read("?view_schema_id="+schema, viewer)["data"].(map[string]any)["saved_views"].([]any)
	if len(elevated) != 4 {
		t.Fatalf("admin visibility: %v", elevated)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role='viewer' WHERE incident_id=$1 AND user_id=$2`, id, viewerID); err != nil {
		t.Fatal(err)
	}
	httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.MethodGet, endpoint+"/"+hidden, nil, httptestx.WithCookies(viewer)), http.StatusNotFound, "saved_view_not_found")

	// Current membership is evaluated again for an existing cursor and detail.
	if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, id, viewerID); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"/" + ids[0], "?view_schema_id=" + schema + "&cursor_token=" + url.QueryEscape(cursor)} {
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.MethodGet, endpoint+suffix, nil, httptestx.WithCookies(viewer)), http.StatusNotFound, "saved_view_not_found")
	}
}
