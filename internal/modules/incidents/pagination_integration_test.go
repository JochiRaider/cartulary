package incidents_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentListUsesLiveFirstPageIndependentlyOfTestClock(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix:        "incident_membership-pagination-live-first-page",
		TestRouteMode: httptestx.TestRouteModeHarnessOwned,
	})

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-incident_membership-live-first-create",
		"incident_key":  "IR-PAGINATION-LIVE",
		"title":         "Live First Page",
	})
	incidentID := incident["incident_id"].(string)

	patchResp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "TLP:AMBER",
			"current_phase":         "containment",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	flowtest.WithClockOffset(t, http.DefaultClient, harness.Server.HTTP.URL, -600)

	listResp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		nil,
		httptestx.WithCookies(adminLogin.Session),
	)
	listBody := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)
	incidents := listBody["data"].(map[string]any)["incidents"].([]any)
	incidentRow := findByKey(t, incidents, "incident_id", incidentID)
	if incidentRow["tlp"] != "TLP:AMBER" || incidentRow["current_phase"] != "containment" {
		t.Fatalf("expected patched incident to remain visible on first page, got %#v", incidentRow)
	}
}

func TestIncidentListSearchStatusAndCursorScope(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "incident_membership-pagination-incidents-search-status", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	first := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id":             "txn-support-incident_membership-search-first",
		"incident_key":              "IR-SEARCH-FIRST",
		"title":                     "First Incident",
		"severity":                  "high",
		"primary_external_case_ref": "CASE-SEARCH-1",
	})
	firstID := first["incident_id"].(string)

	httptestx.SetClockAfter(t, harness.Server, mustParseTimestamp(t, first["updated_at"]), time.Second)
	scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-incident_membership-search-second",
		"incident_key":  "IR-SEARCH-SECOND",
		"title":         "Second Incident",
		"severity":      "low",
	})

	searchResp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1&search="+url.QueryEscape("CASE-SEARCH-1")+"&status=active",
		nil,
		httptestx.WithCookies(adminLogin.Session),
	)
	searchBody := httptestx.RequireSuccessEnvelope(t, searchResp, http.StatusOK)
	searchRows := searchBody["data"].(map[string]any)["incidents"].([]any)
	if len(searchRows) != 1 || searchRows[0].(map[string]any)["incident_id"] != firstID {
		t.Fatalf("search must evaluate the authorized collection before pagination, got %#v", searchRows)
	}

	firstPage := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1&status=active",
		nil,
		httptestx.WithCookies(adminLogin.Session),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	nextCursor := requireNextCursor(t, firstPageBody)

	mismatchedCursor := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?cursor_token="+url.QueryEscape(nextCursor)+"&status=closed",
		nil,
		httptestx.WithCookies(adminLogin.Session),
	)
	mismatchBody := httptestx.RequireErrorEnvelope(t, mismatchedCursor, http.StatusBadRequest, "invalid_pagination_request")
	details := mismatchBody["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "cursor_query_mismatch" {
		t.Fatalf("expected cursor_query_mismatch for changed status scope, got %#v", details)
	}
}

func TestIncidentListContinuationUsesLiveMembershipQuery(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "incident_membership-pagination-incidents", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	first := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-incident_membership-incidents-first",
		"incident_key":  "IR-PAGINATION-FIRST",
		"title":         "First Incident",
	})
	firstID := first["incident_id"].(string)

	httptestx.SetClockAfter(t, harness.Server, mustParseTimestamp(t, first["updated_at"]), time.Second)
	second := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-incident_membership-incidents-second",
		"incident_key":  "IR-PAGINATION-SECOND",
		"title":         "Second Incident",
	})
	secondID := second["incident_id"].(string)
	httptestx.SetClockAfter(t, harness.Server, mustParseTimestamp(t, second["updated_at"]), time.Second)

	firstPage := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1",
		nil,
		httptestx.WithCookies(adminLogin.Session),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	firstPageRows := firstPageBody["data"].(map[string]any)["incidents"].([]any)
	if len(firstPageRows) != 1 {
		t.Fatalf("expected one incident on first page, got %#v", firstPageRows)
	}
	if firstPageRows[0].(map[string]any)["incident_id"] != secondID {
		t.Fatalf("expected newest incident on first page, got %#v", firstPageRows[0])
	}
	nextCursor := requireNextCursor(t, firstPageBody)

	patchResp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+firstID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "TLP:AMBER",
			"current_phase":         "containment",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	continued := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?cursor_token="+url.QueryEscape(nextCursor),
		nil,
		httptestx.WithCookies(adminLogin.Session),
	)
	continuedBody := httptestx.RequireSuccessEnvelope(t, continued, http.StatusOK)
	continuedRows := continuedBody["data"].(map[string]any)["incidents"].([]any)
	if len(continuedRows) != 0 {
		t.Fatalf("expected updated incident to move outside anchored continuation page, got %#v", continuedRows)
	}

	fresh := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1",
		nil,
		httptestx.WithCookies(adminLogin.Session),
	)
	freshBody := httptestx.RequireSuccessEnvelope(t, fresh, http.StatusOK)
	freshRows := freshBody["data"].(map[string]any)["incidents"].([]any)
	if len(freshRows) != 1 {
		t.Fatalf("expected one incident on refreshed first page, got %#v", freshRows)
	}
	liveIncident := freshRows[0].(map[string]any)
	if liveIncident["incident_id"] != firstID || liveIncident["tlp"] != "TLP:AMBER" || liveIncident["current_phase"] != "containment" {
		t.Fatalf("expected fresh request to reflect live ordering and payload, got %#v", liveIncident)
	}
}

func TestIncidentListContinuationOmitsRevokedMembership(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "incident_membership-pagination-revoked-membership", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	viewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "pagination-viewer@example.test", "Pagination Viewer", "PaginationViewer1!", false, false, true)
	viewerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "pagination-viewer@example.test", "PaginationViewer1!", nil)

	first := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-incident_membership-revoke-first",
		"incident_key":  "IR-PAGINATION-REVOKE-FIRST",
		"title":         "Revoked First Incident",
	})
	firstID := first["incident_id"].(string)
	scenariotest.CreateMembership(t, harness.Server, adminLogin, firstID, map[string]any{
		"client_txn_id": "txn-support-incident_membership-revoke-first-membership",
		"user_id":       viewerID,
		"role":          "viewer",
	})

	httptestx.SetClockAfter(t, harness.Server, mustParseTimestamp(t, first["updated_at"]), time.Second)
	second := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-incident_membership-revoke-second",
		"incident_key":  "IR-PAGINATION-REVOKE-SECOND",
		"title":         "Visible Second Incident",
	})
	secondID := second["incident_id"].(string)
	scenariotest.CreateMembership(t, harness.Server, adminLogin, secondID, map[string]any{
		"client_txn_id": "txn-support-incident_membership-revoke-second-membership",
		"user_id":       viewerID,
		"role":          "viewer",
	})

	firstPage := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1",
		nil,
		httptestx.WithCookies(viewerLogin.Session),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	firstPageRows := firstPageBody["data"].(map[string]any)["incidents"].([]any)
	if len(firstPageRows) != 1 || firstPageRows[0].(map[string]any)["incident_id"] != secondID {
		t.Fatalf("expected viewer first page to return second incident, got %#v", firstPageRows)
	}
	nextCursor := requireNextCursor(t, firstPageBody)

	scenariotest.DeleteMembership(t, harness.Server, adminLogin, firstID, viewerID, map[string]any{
		"base_membership_version": 1,
	})

	continued := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?cursor_token="+url.QueryEscape(nextCursor),
		nil,
		httptestx.WithCookies(viewerLogin.Session),
	)
	continuedBody := httptestx.RequireSuccessEnvelope(t, continued, http.StatusOK)
	continuedRows := continuedBody["data"].(map[string]any)["incidents"].([]any)
	if len(continuedRows) != 0 {
		t.Fatalf("expected revoked incident to be omitted from continuation, got %#v", continuedRows)
	}
}

func TestMembershipListContinuationUsesLiveRows(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "incident_membership-pagination-memberships", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-incident_membership-memberships-incident",
		"incident_key":  "IR-PAGINATION-MEMBERSHIPS",
		"title":         "Membership Pagination",
	})
	incidentID := incident["incident_id"].(string)

	memberOneID := flowtest.SeedActor(t, harness.DB, flowtest.ActorAttributes{Email: "membership-one@example.test", DisplayName: "Membership One"}).ID.String()
	memberTwoID := flowtest.SeedActor(t, harness.DB, flowtest.ActorAttributes{Email: "membership-two@example.test", DisplayName: "Membership Two"}).ID.String()

	if memberOneID > memberTwoID {
		memberOneID, memberTwoID = memberTwoID, memberOneID
	}

	createMembership(t, harness, adminLogin, incidentID, "txn-support-incident_membership-membership-one", memberOneID, "viewer")
	memberTwo := createMembership(t, harness, adminLogin, incidentID, "txn-support-incident_membership-membership-two", memberTwoID, "viewer")

	tied := time.Now().UTC().Truncate(time.Microsecond).Add(time.Second)
	if _, err := harness.DB.ExecContext(t.Context(), `UPDATE incident_memberships SET joined_at=$2 WHERE incident_id=$1 AND user_id IN ($3,$4)`, incidentID, tied, memberOneID, memberTwoID); err != nil {
		t.Fatal(err)
	}

	firstPage := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships?limit=2",
		nil,
		httptestx.WithCookies(adminLogin.Session),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	nextCursor := requireNextCursor(t, firstPageBody)

	patchResp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+memberTwoID,
		map[string]any{
			"base_membership_version": memberTwo["membership_version"],
			"role":                    "reviewer",
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	// Removing a row preceding the cursor must not skip the next live row.
	removed := httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+memberOneID, map[string]any{"base_membership_version": 1}, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
	if removed.StatusCode != 204 {
		t.Fatalf("delete prior row: %d", removed.StatusCode)
	}
	_ = removed.Body.Close()
	beforeID := flowtest.SeedActor(t, harness.DB, flowtest.ActorAttributes{Email: "before-cursor@example.test", DisplayName: "Before cursor"}).ID.String()
	afterID := flowtest.SeedActor(t, harness.DB, flowtest.ActorAttributes{Email: "after-cursor@example.test", DisplayName: "After cursor"}).ID.String()
	createMembership(t, harness, adminLogin, incidentID, "before-cursor", beforeID, "viewer")
	createMembership(t, harness, adminLogin, incidentID, "after-cursor", afterID, "viewer")
	for id, joined := range map[string]time.Time{beforeID: tied.Add(-time.Second), afterID: tied.Add(time.Second)} {
		if _, err := harness.DB.ExecContext(t.Context(), `UPDATE incident_memberships SET joined_at=$3 WHERE incident_id=$1 AND user_id=$2`, incidentID, id, joined); err != nil {
			t.Fatal(err)
		}
	}

	continued := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships?cursor_token="+url.QueryEscape(nextCursor),
		nil,
		httptestx.WithCookies(adminLogin.Session),
	)
	continuedBody := httptestx.RequireSuccessEnvelope(t, continued, http.StatusOK)
	continuedRows := continuedBody["data"].(map[string]any)["memberships"].([]any)
	if len(continuedRows) != 2 {
		t.Fatalf("expected two memberships on continuation page, got %#v", continuedRows)
	}
	liveContinuedMembership := continuedRows[0].(map[string]any)
	if liveContinuedMembership["user_id"] != memberTwoID || liveContinuedMembership["role"] != "reviewer" {
		t.Fatalf("expected live membership payload, got %#v", liveContinuedMembership)
	}

	if continuedRows[1].(map[string]any)["user_id"] != afterID {
		t.Fatalf("insertion after cursor missing: %#v", continuedRows)
	}
	if continuedBody["meta"].(map[string]any)["paging"].(map[string]any)["has_more"] != false {
		t.Fatal("exhausted page has continuation")
	}

	fresh := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		nil,
		httptestx.WithCookies(adminLogin.Session),
	)
	freshBody := httptestx.RequireSuccessEnvelope(t, fresh, http.StatusOK)
	liveMembership := findByKey(t, freshBody["data"].(map[string]any)["memberships"].([]any), "user_id", memberTwoID)
	if liveMembership["role"] != "reviewer" {
		t.Fatalf("expected fresh membership list to reflect live role, got %#v", liveMembership)
	}
	for _, query := range []string{"limit=1&limit=2", "%xx=x", "%ff=value", "search=no", "offset=1", "limit="} {
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships?"+query, nil, httptestx.WithCookies(adminLogin.Session))
		if response.StatusCode != 400 {
			t.Fatalf("query %q status=%d", query, response.StatusCode)
		}
		_ = response.Body.Close()
	}
	if _, err := harness.DB.ExecContext(t.Context(), `DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, incidentID, adminID); err != nil {
		t.Fatal(err)
	}
	denied := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships?cursor_token="+url.QueryEscape(nextCursor)+"&limit=", nil, httptestx.WithCookies(adminLogin.Session))
	httptestx.RequireErrorEnvelope(t, denied, 404, "incident_not_found")

}

func createMembership(
	t testing.TB,
	harness *appsupport.ServerHarness,
	adminLogin authcookietest.AuthCookies,
	incidentID string,
	clientTxnID string,
	userID string,
	role string,
) map[string]any {
	t.Helper()

	resp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": clientTxnID,
			"user_id":       userID,
			"role":          role,
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func mustParseTimestamp(t testing.TB, value any) time.Time {
	t.Helper()
	text, ok := value.(string)
	if !ok {
		t.Fatalf("timestamp value got %T want string", value)
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		t.Fatalf("parse timestamp %q: %v", text, err)
	}
	return parsed
}

func requireNextCursor(t testing.TB, envelope map[string]any) string {
	t.Helper()

	meta, ok := envelope["meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected meta object, got %#v", envelope["meta"])
	}
	paging, ok := meta["paging"].(map[string]any)
	if !ok {
		t.Fatalf("expected paging object, got %#v", meta["paging"])
	}
	token, ok := paging["next_cursor"].(string)
	if !ok || token == "" {
		t.Fatalf("expected next cursor, got %#v", paging["next_cursor"])
	}
	return token
}

func findByKey(t testing.TB, rows []any, key string, value string) map[string]any {
	t.Helper()

	for _, row := range rows {
		candidate, ok := row.(map[string]any)
		if ok && candidate[key] == value {
			return candidate
		}
	}
	t.Fatalf("expected row with %s=%s, got %#v", key, value, rows)
	return nil
}
