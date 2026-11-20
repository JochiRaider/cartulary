package savedviews_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestSavedViewIndependentReadContract_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "saved-view-independent-reads", TestRouteMode: httptestx.TestRouteModeDisabled})
	admin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{"client_txn_id": "svd-incident", "incident_key": "IR-SVD", "title": "Saved view discovery"})
	id := incident["incident_id"].(string)
	viewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "svd-viewer@example.test", "Viewer", "SavedViewDiscovery1!", false, false, true)
	viewerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "svd-viewer@example.test", "SavedViewDiscovery1!", nil)
	scenariotest.CreateMembership(t, harness.Server, admin, id, map[string]any{"client_txn_id": "svd-member", "user_id": viewerID, "role": "viewer"})
	const schema = "cartulary.view.timeline.v2"
	for i := 0; i < 101; i++ {
		seedSavedView(t, harness.DB, fmt.Sprintf("00000000-0000-4000-8000-%012d", i+1), id, "cartulary.view.notes.v1", "shared", fmt.Sprintf("Other %d", i), adminID.String(), "2026-08-02T00:00:00Z")
	}
	ids := []string{"00000000-0000-4000-8000-000000000201", "00000000-0000-4000-8000-000000000202", "00000000-0000-4000-8000-000000000203"}
	seedSavedView(t, harness.DB, ids[0], id, schema, "private", "Owner private", viewerID, "2026-08-01T00:00:00Z")
	seedSavedView(t, harness.DB, ids[1], id, schema, "system", "System", "", "2026-08-01T00:00:00Z")
	seedSavedView(t, harness.DB, ids[2], id, schema, "shared", "Shared", adminID.String(), "2026-08-01T00:00:00Z")
	hidden := "00000000-0000-4000-8000-000000000204"
	seedSavedView(t, harness.DB, hidden, id, schema, "private", "Hidden", adminID.String(), "2026-08-03T00:00:00Z")
	endpoint := harness.Server.HTTP.URL + "/api/v1/incidents/" + id + "/saved-views"
	read := func(suffix string, cookie *http.Cookie) map[string]any {
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, endpoint+suffix, nil, httptestx.WithCookies(cookie))
		if response.Header.Get("Cache-Control") != "no-store" {
			t.Fatal("read must be no-store")
		}
		return httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
	}
	type storedState struct {
		Layout, Updated string
		Version         int64
	}
	readStored := func(resourceID string) storedState {
		var state storedState
		if err := harness.DB.QueryRowContext(context.Background(), `SELECT layout_json::text,updated_at::text,saved_view_version FROM saved_views WHERE saved_view_id=$1`, resourceID).Scan(&state.Layout, &state.Updated, &state.Version); err != nil {
			t.Fatal(err)
		}
		return state
	}
	beforePrivate := readStored(ids[0])
	beforeSystem := readStored(ids[1])

	unfiltered := read("", viewerLogin.Session)["data"].(map[string]any)["saved_views"].([]any)
	if len(unfiltered) != 100 || unfiltered[0].(map[string]any)["view_schema_id"] == schema {
		t.Fatal("fixture must place active schema beyond first unfiltered page")
	}
	first := read("?view_schema_id="+schema+"&limit=2", viewerLogin.Session)
	rows := first["data"].(map[string]any)["saved_views"].([]any)
	if !reflect.DeepEqual(savedViewNames(rows), []string{"Owner private", "System"}) {
		t.Fatalf("first page: %v", rows)
	}
	cursor := first["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor"].(string)
	second := read("?view_schema_id="+schema+"&cursor_token="+url.QueryEscape(cursor), viewerLogin.Session)
	if !reflect.DeepEqual(savedViewNames(second["data"].(map[string]any)["saved_views"].([]any)), []string{"Shared"}) {
		t.Fatal("wrong continuation")
	}
	if second["meta"].(map[string]any)["paging"].(map[string]any)["has_more"] != false {
		t.Fatal("hidden resource affected continuation")
	}
	for i, resourceID := range ids {
		got := read("/"+resourceID, viewerLogin.Session)["data"].(map[string]any)
		if got["saved_view_id"] != resourceID {
			t.Fatal("unrelated detail")
		}
		if i < 2 && !reflect.DeepEqual(got, rows[i]) {
			t.Fatal("list/detail envelope disagreement")
		}
	}
	for _, row := range rows {
		layout := row.(map[string]any)["layout_json"].(map[string]any)
		if layout["layout_schema_id"] != "cartulary.layout.v2" || layout["frozen_through_field_key"] != nil {
			t.Fatal("current layout is incomplete", layout)
		}
	}
	if readStored(ids[0]) != beforePrivate || readStored(ids[1]) != beforeSystem {
		t.Fatal("read rewrote stored resource")
	}
	// Canonical no-op PATCH leaves current stored bytes and versions untouched.
	currentLayout := rows[0].(map[string]any)["layout_json"]
	patchSavedViewHTTP(t, harness.Server.HTTP.URL, id, ids[0], viewerLogin.Session, viewerLogin.CSRF, map[string]any{"base_saved_view_version": beforePrivate.Version, "layout_json": currentLayout})
	if readStored(ids[0]) != beforePrivate {
		t.Fatal("canonical no-op rewrote stored layout")
	}

	missing := "00000000-0000-4000-8000-000000009999"
	for _, resourceID := range []string{hidden, missing} {
		resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, endpoint+"/"+resourceID, nil, httptestx.WithCookies(viewerLogin.Session))
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusNotFound, "saved_view_not_found")
		if len(body["error"].(map[string]any)["details"].(map[string]any)) != 0 {
			t.Fatal("concealment details leaked")
		}
	}
	read("/"+hidden, admin.Session)
	// Reads do not repair an invalid home pointer. Startup retains that ownership.
	if _, err := harness.DB.ExecContext(context.Background(), `INSERT INTO user_workbook_preferences (incident_id,user_id,home_sheet_ref,updated_at) VALUES ($1,$2,jsonb_build_object('kind','saved_view','id',$3::text),now()) ON CONFLICT (incident_id,user_id) DO UPDATE SET home_sheet_ref=excluded.home_sheet_ref`, id, viewerID, missing); err != nil {
		t.Fatal(err)
	}
	read("/"+ids[0], viewerLogin.Session)
	var pointer string
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT home_sheet_ref->>'id' FROM user_workbook_preferences WHERE incident_id=$1 AND user_id=$2`, id, viewerID).Scan(&pointer); err != nil || pointer != missing {
		t.Fatalf("read repaired pointer: %s %v", pointer, err)
	}
	read("?view_schema_id="+schema, viewerLogin.Session)
	// Query errors remain public contract errors, with no identifying details.
	for _, tc := range []struct{ suffix, code, reason string }{
		{"?scope=private&scope=shared", "invalid_list_query", "duplicate_query_member"},
		{"?view_schema_id=", "invalid_list_query", "invalid_filter_value"},
		{"?view_schema_id=%FF", "invalid_saved_view_read_request", "malformed_query"},
		{"/" + ids[0] + "?limit=", "invalid_pagination_request", "pagination_not_supported"},
		{"/" + ids[0] + "?view_schema_id=" + schema, "invalid_saved_view_read_request", "unknown_query_member"},
		{"?view_schema_id=cartulary.view.notes.v1&cursor_token=" + url.QueryEscape(cursor), "invalid_pagination_request", "cursor_query_mismatch"},
	} {
		resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, endpoint+tc.suffix, nil, httptestx.WithCookies(viewerLogin.Session))
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, tc.code)
		if body["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != tc.reason {
			t.Fatalf("wrong read reason for %s: %v", tc.suffix, body)
		}
	}
	httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, endpoint+"/"+ids[0], nil), http.StatusUnauthorized, "session_required")
	other := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{"client_txn_id": "svd-other", "incident_key": "IR-SVD-OTHER", "title": "Other incident"})
	foreign := "00000000-0000-4000-8000-000000000301"
	seedSavedView(t, harness.DB, foreign, other["incident_id"].(string), schema, "shared", "Foreign incident resource", adminID.String(), "2026-08-01T00:00:00Z")
	httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, endpoint+"/"+foreign, nil, httptestx.WithCookies(viewerLogin.Session)), http.StatusNotFound, "saved_view_not_found")
	closed := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+id+"/close", map[string]any{"base_incident_version": incident["incident_version"], "client_txn_id": "svd-close", "reason": "Read contract closed incident coverage"}, httptestx.WithCookies(admin.Session, admin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRF.Value))
	httptestx.RequireSuccessEnvelope(t, closed, http.StatusOK)
	read("/"+ids[0], viewerLogin.Session)
	read("?view_schema_id="+schema, viewerLogin.Session)
	// Closing the incident does not freeze its saved configuration objects.
	created := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, endpoint,
		map[string]any{"display_name": "Closed configuration", "view_schema_id": schema, "query_json": map[string]any{}},
		httptestx.WithCookies(admin.Session, admin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRF.Value)), http.StatusCreated)["data"].(map[string]any)
	createdID := created["saved_view_id"].(string)
	updated := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, endpoint+"/"+createdID,
		map[string]any{"display_name": "Closed configuration revised", "base_saved_view_version": created["saved_view_version"]},
		httptestx.WithCookies(admin.Session, admin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRF.Value)), http.StatusOK)["data"].(map[string]any)
	if !reflect.DeepEqual(read("/"+createdID, admin.Session)["data"], updated) {
		t.Fatal("closed configuration detail did not observe update")
	}
	deleted := httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, endpoint+"/"+createdID,
		nil,
		httptestx.WithCookies(admin.Session, admin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRF.Value))
	httptestx.RequireSuccessEnvelope(t, deleted, http.StatusOK)
	httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, endpoint+"/"+createdID, nil, httptestx.WithCookies(admin.Session)), http.StatusNotFound, "saved_view_not_found")
	// Current role controls visibility on both routes, even on a closed incident.
	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role='admin' WHERE incident_id=$1 AND user_id=$2`, id, viewerID); err != nil {
		t.Fatal(err)
	}
	read("/"+hidden, viewerLogin.Session)
	elevated := read("?view_schema_id="+schema, viewerLogin.Session)["data"].(map[string]any)["saved_views"].([]any)
	if len(elevated) != 4 {
		t.Fatalf("admin visibility: %v", elevated)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role='viewer' WHERE incident_id=$1 AND user_id=$2`, id, viewerID); err != nil {
		t.Fatal(err)
	}
	httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, endpoint+"/"+hidden, nil, httptestx.WithCookies(viewerLogin.Session)), http.StatusNotFound, "saved_view_not_found")

	t.Run("invalid stored layouts fail without repair or preference changes", func(t *testing.T) {
		legacy := savedViewLayoutWith(t, func(layout map[string]any) {
			layout["layout_schema_id"] = "cartulary.layout.v1"
			delete(layout, "frozen_through_field_key")
		})
		unknown := savedViewLayoutWith(t, func(layout map[string]any) { layout["layout_schema_id"] = "cartulary.layout.v99" })
		missing := savedViewLayoutWith(t, func(layout map[string]any) { delete(layout, "frozen_through_field_key") })
		for _, invalid := range []any{legacy, unknown, missing, map[string]any{}, nil, []any{}} {
			invalidBytes, _ := json.Marshal(invalid)
			if _, err := harness.DB.ExecContext(context.Background(), `UPDATE saved_views SET layout_json=$2 WHERE saved_view_id=$1`, ids[0], invalidBytes); err != nil {
				t.Fatal(err)
			}
			if _, err := harness.DB.ExecContext(context.Background(), `UPDATE user_workbook_preferences SET home_sheet_ref=jsonb_build_object('kind','saved_view','id',$3::text) WHERE incident_id=$1 AND user_id=$2`, id, viewerID, ids[0]); err != nil {
				t.Fatal(err)
			}
			before := readStored(ids[0])
			var preferenceBefore, preferenceAfter string
			preferenceSQL := `SELECT home_sheet_ref::text || updated_at::text FROM user_workbook_preferences WHERE incident_id=$1 AND user_id=$2`
			if err := harness.DB.QueryRowContext(context.Background(), preferenceSQL, id, viewerID).Scan(&preferenceBefore); err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{endpoint + "/" + ids[0], endpoint + "?view_schema_id=" + schema, harness.Server.HTTP.URL + "/api/v1/incidents/" + id + "/workbook-startup"} {
				response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, path, nil, httptestx.WithCookies(viewerLogin.Session))
				httptestx.RequireErrorEnvelope(t, response, http.StatusInternalServerError, "internal_error")
			}
			response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, endpoint+"/"+ids[0], map[string]any{"base_saved_view_version": before.Version, "layout_json": currentLayout}, httptestx.WithCookies(viewerLogin.Session, viewerLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRF.Value))
			httptestx.RequireErrorEnvelope(t, response, http.StatusInternalServerError, "internal_error")
			if err := harness.DB.QueryRowContext(context.Background(), preferenceSQL, id, viewerID).Scan(&preferenceAfter); err != nil {
				t.Fatal(err)
			}
			if readStored(ids[0]) != before || preferenceAfter != preferenceBefore {
				t.Fatal("invalid read/update repaired stored state or preferences")
			}
		}
		valid, _ := json.Marshal(currentLayout)
		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE saved_views SET layout_json=$2 WHERE saved_view_id=$1`, ids[0], valid); err != nil {
			t.Fatal(err)
		}
	})

	// Current membership is evaluated again for an existing cursor and detail.
	if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, id, viewerID); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"/" + ids[0], "?view_schema_id=" + schema + "&cursor_token=" + url.QueryEscape(cursor)} {
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, endpoint+suffix, nil, httptestx.WithCookies(viewerLogin.Session)), http.StatusNotFound, "saved_view_not_found")
	}
}

func TestSavedViewTimestampCompatibility_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "saved-view-timestamp-compatibility", TestRouteMode: httptestx.TestRouteModeDisabled})
	admin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{"client_txn_id": "timestamp-incident", "incident_key": "IR-TIMESTAMP", "title": "Timestamp compatibility"})
	id := incident["incident_id"].(string)
	const schema = "cartulary.view.notes.v1"
	const savedID = "00000000-0000-4000-8000-000000008201"
	seedSavedView(t, harness.DB, savedID, id, schema, "private", "Older timestamp range", adminID.String(), "2026-08-01T00:00:00Z")
	query := func(lower, upper string) map[string]any {
		return map[string]any{"sort": []any{}, "filters": []any{map[string]any{"field_key": "note.updated_at", "op": "range", "arg": map[string]any{"gte": lower, "lte": upper}}}}
	}
	invalid := query("2026-04-18T00:00:00.1Z", "2026-04-18T00:00:00Z")
	raw, err := json.Marshal(invalid)
	if err != nil {
		t.Fatal(err)
	}
	// This represents a row persisted before chronological range validation.
	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE saved_views SET query_json=$2 WHERE saved_view_id=$1`, savedID, raw); err != nil {
		t.Fatal(err)
	}
	endpoint := harness.Server.HTTP.URL + "/api/v1/incidents/" + id
	request := func(method, path string, body any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, method, endpoint+path, body, httptestx.WithCookies(admin.Session, admin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRF.Value))
	}
	read := func() map[string]any {
		return httptestx.RequireSuccessEnvelope(t, request(http.MethodGet, "/saved-views/"+savedID, nil), http.StatusOK)["data"].(map[string]any)
	}
	before := read()
	if !reflect.DeepEqual(before["query_json"], invalid) {
		t.Fatal("read changed the older predicate", before)
	}
	startup := httptestx.RequireSuccessEnvelope(t, request(http.MethodGet, "/workbook-startup?sheet_ref_kind=saved_view&sheet_ref_id="+savedID, nil), http.StatusOK)["data"].(map[string]any)
	if !reflect.DeepEqual(startup["selected_saved_view"].(map[string]any)["query_json"], invalid) {
		t.Fatal("startup changed the older predicate", startup)
	}
	failure := httptestx.RequireErrorEnvelope(t, request(http.MethodPost, "/views/"+schema+"/query", invalid), http.StatusBadRequest, "invalid_view_query")
	if failure["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "invalid_filter_operand" {
		t.Fatal(failure)
	}
	httptestx.RequireErrorEnvelope(t, request(http.MethodPost, "/saved-views", map[string]any{"view_schema_id": schema, "display_name": "Invalid copy", "query_json": invalid}), http.StatusBadRequest, "invalid_mutation_payload")
	httptestx.RequireErrorEnvelope(t, request(http.MethodPatch, "/saved-views/"+savedID, map[string]any{"base_saved_view_version": before["saved_view_version"], "query_json": invalid}), http.StatusBadRequest, "invalid_mutation_payload")
	if !reflect.DeepEqual(read(), before) {
		t.Fatal("rejection rewrote or removed the stored query")
	}
	valid := query("2026-04-17T20:00:00.000000001-04:00", "2026-04-18T00:00:00.1Z")
	expected := query("2026-04-18T00:00:00.000000001Z", "2026-04-18T00:00:00.1Z")
	updated := httptestx.RequireSuccessEnvelope(t, request(http.MethodPatch, "/saved-views/"+savedID, map[string]any{"base_saved_view_version": before["saved_view_version"], "query_json": valid}), http.StatusOK)["data"].(map[string]any)
	if !reflect.DeepEqual(updated["query_json"], expected) || !reflect.DeepEqual(read(), updated) {
		t.Fatal("explicit correction lost canonical precision", updated)
	}
	httptestx.RequireSuccessEnvelope(t, request(http.MethodPost, "/views/"+schema+"/query", updated["query_json"]), http.StatusOK)
	created := httptestx.RequireSuccessEnvelope(t, request(http.MethodPost, "/saved-views", map[string]any{"view_schema_id": schema, "display_name": "Valid fractional query", "query_json": valid}), http.StatusCreated)["data"].(map[string]any)
	if !reflect.DeepEqual(created["query_json"], expected) {
		t.Fatal("create lost instant meaning", created)
	}
}

// TestSavedViewRouteAdmissionOrder_Integration covers Core 01 REQ-01-707 and
// AC-617 for the saved-view routes: route resolution rejects undeclared
// methods and HEAD with registered envelopes; authentication and CSRF precede
// path parameters; a malformed or encoded-slash incident_id or saved_view_id
// receives the declared saved_view_not_found, identical to a missing
// resource's, and never reaches another operation; and the incident's
// visibility precedes the query and the request body (REQ-01-704), with no
// committed effect. The saved-view routes declare no incident role gate.
func TestSavedViewRouteAdmissionOrder_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "saved-view-admission-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	base := harness.Server.HTTP.URL
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	incidentID := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-saved-view-admission-order-incident",
		"incident_key":  "IR-SAVED-VIEW-ADMISSION-ORDER",
		"title":         "Saved-view route admission order",
	})["incident_id"].(string)
	outsiderPassword := "SavedViewAdmissionOutsider1!"
	flowtest.SeedLocalUserFlags(t, harness.DB, "saved-view-admission-outsider@example.test", "Saved View Admission Outsider", outsiderPassword, false, false, true)
	outsiderLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, "saved-view-admission-outsider@example.test", outsiderPassword, nil)
	// A redirect is never followed, so a 3xx would surface as itself.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	send := func(method string, path string, login authcookietest.AuthCookies, body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, client, method, base+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	requireMethodNotAllowed := func(t testing.TB, response *http.Response, allow ...string) {
		t.Helper()
		if got := response.Header.Get("Allow"); got != strings.Join(allow, ", ") {
			t.Fatalf("Allow = %q, want %q", got, strings.Join(allow, ", "))
		}
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusMethodNotAllowed, "method_not_allowed")
		want := make([]any, 0, len(allow))
		for _, method := range allow {
			want = append(want, method)
		}
		if details := httptestx.RequireErrorDetails(t, body); !reflect.DeepEqual(details, map[string]any{"allowed_methods": want}) {
			t.Fatalf("details = %#v, want allowed_methods %v", details, allow)
		}
	}
	incidentPath := "/api/v1/incidents/" + incidentID
	collection := incidentPath + "/saved-views"
	created := httptestx.RequireSuccessEnvelope(t, send(http.MethodPost, collection, admin, map[string]any{
		"display_name": "Admission order", "view_schema_id": "cartulary.view.timeline.v2", "query_json": map[string]any{}, "scope": "shared",
	}), http.StatusCreated)["data"].(map[string]any)
	savedViewID := created["saved_view_id"].(string)
	item := collection + "/" + savedViewID
	missingItem := collection + "/" + uuid.NewString()
	patchBody := map[string]any{"base_saved_view_version": created["saved_view_version"], "display_name": "Admission order revised"}
	// Every route, with a body its handler accepts up to the addressed
	// resource's classification.
	routes := []struct {
		name, method string
		item         bool
		body         any
	}{
		{name: "list", method: http.MethodGet},
		{name: "create", method: http.MethodPost, body: map[string]any{}},
		{name: "get", method: http.MethodGet, item: true},
		{name: "patch", method: http.MethodPatch, item: true, body: patchBody},
		{name: "delete", method: http.MethodDelete, item: true},
	}
	pathFor := func(incident string, savedView string, itemRoute bool) string {
		path := "/api/v1/incidents/" + incident + "/saved-views"
		if itemRoute {
			path += "/" + savedView
		}
		return path
	}
	httptestx.RequireSuccessEnvelope(t, send(http.MethodGet, pathFor(strings.ToUpper(incidentID), strings.ToUpper(savedViewID), true), admin, nil), http.StatusOK)
	before := savedViewAdmissionEffects(t, harness.DB)

	t.Run("authentication and CSRF precede path parameters", func(t *testing.T) {
		for _, route := range routes {
			for _, path := range []string{pathFor("not-an-incident-id", savedViewID, route.item), pathFor(incidentID, "not-a-saved-view-id", route.item)} {
				httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, route.method, base+path, route.body), http.StatusUnauthorized, "session_required")
				if route.method != http.MethodGet {
					withoutCSRF := httptestx.DoJSON(t, client, route.method, base+path, route.body, httptestx.WithCookies(admin.Session))
					httptestx.RequireErrorEnvelope(t, withoutCSRF, http.StatusForbidden, "csrf_verification_failed")
				}
			}
		}
	})
	t.Run("a malformed identifier is a missing resource", func(t *testing.T) {
		for _, malformed := range []string{"not-an-id", strings.ReplaceAll(savedViewID, "-", ""), "%7B" + savedViewID + "%7D"} {
			for _, route := range routes {
				requireSavedViewResponsesAlike(t,
					send(route.method, pathFor(malformed, savedViewID, route.item), admin, route.body),
					send(route.method, pathFor(uuid.NewString(), savedViewID, route.item), admin, route.body))
				if route.item {
					requireSavedViewResponsesAlike(t,
						send(route.method, pathFor(incidentID, malformed, true), admin, route.body),
						send(route.method, missingItem, admin, route.body))
				}
			}
		}
	})
	t.Run("an encoded slash stays inside the identifier", func(t *testing.T) {
		requireSavedViewResponsesAlike(t, send(http.MethodGet, item+"%2Fx", admin, nil), send(http.MethodGet, missingItem, admin, nil))
		requireSavedViewResponsesAlike(t, send(http.MethodPatch, item+"%2Fx", admin, patchBody), send(http.MethodPatch, missingItem, admin, patchBody))
		requireSavedViewResponsesAlike(t, send(http.MethodDelete, item+"%2Fx", admin, nil), send(http.MethodDelete, missingItem, admin, nil))
		// The incident read template: the encoded slash never selects the saved-view list.
		httptestx.RequireErrorEnvelope(t, send(http.MethodGet, incidentPath+"%2Fsaved-views", admin, nil), http.StatusNotFound, "incident_not_found")
		requireMethodNotAllowed(t, send(http.MethodPost, incidentPath+"%2Fsaved-views", admin, map[string]any{}), http.MethodGet, http.MethodPatch)
		httptestx.RequireErrorEnvelope(t, send(http.MethodGet, incidentPath+"/saved-views%2F"+savedViewID, admin, nil), http.StatusNotFound, "route_not_found")
	})
	t.Run("route resolution rejections are enveloped", func(t *testing.T) {
		for path, allow := range map[string]string{collection: "GET, POST", item: "DELETE, GET, PATCH"} {
			head := httptestx.DoJSON(t, client, http.MethodHead, base+path, nil, httptestx.WithCookies(admin.Session))
			body, err := io.ReadAll(head.Body)
			_ = head.Body.Close()
			if err != nil || head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != allow || len(body) != 0 {
				t.Fatalf("HEAD %s = %d Allow=%q body=%q err=%v, want 405 Allow %s without a body", path, head.StatusCode, head.Header.Get("Allow"), body, err, allow)
			}
		}
		requireMethodNotAllowed(t, send(http.MethodPut, collection, admin, map[string]any{}), http.MethodGet, http.MethodPost)
		requireMethodNotAllowed(t, send(http.MethodPost, item, admin, map[string]any{}), http.MethodDelete, http.MethodGet, http.MethodPatch)
		httptestx.RequireErrorEnvelope(t, send(http.MethodGet, item+"/unknown", admin, nil), http.StatusNotFound, "route_not_found")
	})
	t.Run("every resolved read response is no-store", func(t *testing.T) {
		// Core 01 REQ-01-144B: authentication, path-parameter, visibility and
		// query failures carry no-store like the success; the mutations do not.
		requireNoStore := func(t testing.TB, response *http.Response, want string) {
			t.Helper()
			defer response.Body.Close()
			if got := response.Header.Get("Cache-Control"); got != want {
				t.Fatalf("%s %s Cache-Control = %q, want %q", response.Request.Method, response.Request.URL.Path, got, want)
			}
		}
		for _, path := range []string{collection, item} {
			requireNoStore(t, httptestx.DoJSON(t, client, http.MethodGet, base+path, nil), "no-store")
			requireNoStore(t, send(http.MethodGet, path, admin, nil), "no-store")
			requireNoStore(t, send(http.MethodGet, path+"?limit=0&page=1", admin, nil), "no-store")
			requireNoStore(t, send(http.MethodGet, path, outsiderLogin, nil), "no-store")
		}
		requireNoStore(t, send(http.MethodGet, pathFor("not-an-incident", savedViewID, false), admin, nil), "no-store")
		requireNoStore(t, send(http.MethodGet, pathFor(incidentID, "not-a-saved-view", true), admin, nil), "no-store")
		requireNoStore(t, send(http.MethodGet, missingItem, admin, nil), "no-store")
		requireNoStore(t, send(http.MethodPatch, missingItem, admin, patchBody), "")
	})
	t.Run("visibility precedes route handling", func(t *testing.T) {
		for _, route := range routes {
			body := any([]any{1})
			if route.method == http.MethodGet || route.method == http.MethodDelete {
				body = nil
			}
			requireSavedViewResponsesAlike(t,
				send(route.method, pathFor(incidentID, savedViewID, route.item)+"?limit=0&view_schema_id=%FF", outsiderLogin, body),
				send(route.method, pathFor(uuid.NewString(), savedViewID, route.item)+"?limit=0&view_schema_id=%FF", outsiderLogin, body))
		}
		httptestx.RequireErrorEnvelope(t, send(http.MethodGet, collection+"?limit=0", admin, nil), http.StatusBadRequest, "invalid_pagination_request")
		httptestx.RequireErrorEnvelope(t, send(http.MethodGet, item+"?limit=1", admin, nil), http.StatusBadRequest, "invalid_pagination_request")
		httptestx.RequireErrorEnvelope(t, send(http.MethodPost, collection, admin, []any{1}), http.StatusBadRequest, "invalid_mutation_payload")
	})
	if after := savedViewAdmissionEffects(t, harness.DB); after != before {
		t.Fatalf("rejected saved-view admission requests committed an effect: before=%s after=%s", before, after)
	}
}

// requireSavedViewResponsesAlike requires the response for an existing
// resource to match the response for a missing one in everything but its
// request identity: the 404 saved_view_not_found envelope (REQ-01-704).
func requireSavedViewResponsesAlike(t testing.TB, existing, missing *http.Response) {
	t.Helper()
	bodies := make([]map[string]any, 0, 2)
	for _, response := range []*http.Response{existing, missing} {
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "saved_view_not_found")
		delete(body["error"].(map[string]any), "request_id")
		bodies = append(bodies, body)
	}
	if !reflect.DeepEqual(bodies[0], bodies[1]) {
		t.Fatalf("existing-resource response %v differs from missing-resource response %v", bodies[0], bodies[1])
	}
	if got, want := existing.Header.Get("Cache-Control"), missing.Header.Get("Cache-Control"); got != want {
		t.Fatalf("existing-resource Cache-Control %q differs from missing-resource Cache-Control %q", got, want)
	}
}

// savedViewAdmissionEffects renders the saved-view, preference, receipt and
// Collaboration state a rejected saved-view request could commit.
func savedViewAdmissionEffects(t testing.TB, db *sql.DB) string {
	t.Helper()
	var effects string
	if err := db.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'saved_views',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY saved_view_id),'[]') FROM saved_views v),
 'user_preferences',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY incident_id,user_id),'[]') FROM user_workbook_preferences p),
 'receipts',(SELECT count(*) FROM route_idempotency)
 )::text`).Scan(&effects); err != nil {
		t.Fatal(err)
	}
	return effects + " collaboration_intents=" + strconv.Itoa(collaborationsupport.CountIntents(t, db, collaborationsupport.IntentSelector{}))
}
