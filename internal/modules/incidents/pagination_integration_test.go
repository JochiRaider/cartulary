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
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentListUsesLiveFirstPageIndependentlyOfTestClock(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix:        "incident_membership-pagination-live-first-page",
		TestRouteMode: httptestx.TestRouteModeHarnessOwned,
	})

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-incident_membership-live-first-create",
		"incident_key":  "IR-PAGINATION-LIVE",
		"title":         "Live First Page",
	})
	incidentID := incident["incident_id"].(string)

	patchResp := httptestx.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "TLP:AMBER",
			"current_phase":         "containment",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	flowtest.WithClockOffset(t, harness.Server.HTTP.URL, -600)

	listResp := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
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
	harness := runtime.StartDefaultServer(t, "incident_membership-pagination-incidents-search-status")

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
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
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1&search="+url.QueryEscape("CASE-SEARCH-1")+"&status=active",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	searchBody := httptestx.RequireSuccessEnvelope(t, searchResp, http.StatusOK)
	searchRows := searchBody["data"].(map[string]any)["incidents"].([]any)
	if len(searchRows) != 1 || searchRows[0].(map[string]any)["incident_id"] != firstID {
		t.Fatalf("search must evaluate the authorized collection before pagination, got %#v", searchRows)
	}

	firstPage := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1&status=active",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	nextCursor := requireNextCursor(t, firstPageBody)

	mismatchedCursor := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?cursor_token="+url.QueryEscape(nextCursor)+"&status=closed",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	mismatchBody := httptestx.RequireErrorEnvelope(t, mismatchedCursor, http.StatusBadRequest, "invalid_pagination_request")
	details := mismatchBody["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "cursor_query_mismatch" {
		t.Fatalf("expected cursor_query_mismatch for changed status scope, got %#v", details)
	}
}

func TestIncidentListContinuationUsesLiveMembershipQuery(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "incident_membership-pagination-incidents")

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
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
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
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
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+firstID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "TLP:AMBER",
			"current_phase":         "containment",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	continued := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?cursor_token="+url.QueryEscape(nextCursor),
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	continuedBody := httptestx.RequireSuccessEnvelope(t, continued, http.StatusOK)
	continuedRows := continuedBody["data"].(map[string]any)["incidents"].([]any)
	if len(continuedRows) != 0 {
		t.Fatalf("expected updated incident to move outside anchored continuation page, got %#v", continuedRows)
	}

	fresh := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
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
	harness := runtime.StartDefaultServer(t, "incident_membership-pagination-revoked-membership")

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	viewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "pagination-viewer@example.test", "Pagination Viewer", "PaginationViewer1!", false, false, true)
	viewerSession, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "pagination-viewer@example.test", "PaginationViewer1!", nil)

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
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1",
		nil,
		httptestx.WithCookies(viewerSession),
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
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?cursor_token="+url.QueryEscape(nextCursor),
		nil,
		httptestx.WithCookies(viewerSession),
	)
	continuedBody := httptestx.RequireSuccessEnvelope(t, continued, http.StatusOK)
	continuedRows := continuedBody["data"].(map[string]any)["incidents"].([]any)
	if len(continuedRows) != 0 {
		t.Fatalf("expected revoked incident to be omitted from continuation, got %#v", continuedRows)
	}
}

func TestMembershipListContinuationUsesLiveRows(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "incident_membership-pagination-memberships")

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-incident_membership-memberships-incident",
		"incident_key":  "IR-PAGINATION-MEMBERSHIPS",
		"title":         "Membership Pagination",
	})
	incidentID := incident["incident_id"].(string)

	memberOneID := flowtest.SeedLocalUserFlags(t, harness.DB, "membership-one@example.test", "Membership One", "MembershipOne1!", false, false, true)
	memberTwoID := flowtest.SeedLocalUserFlags(t, harness.DB, "membership-two@example.test", "Membership Two", "MembershipTwo1!", false, false, true)

	createMembership(t, harness, adminLogin, incidentID, "txn-support-incident_membership-membership-one", memberOneID, "viewer")
	memberTwo := createMembership(t, harness, adminLogin, incidentID, "txn-support-incident_membership-membership-two", memberTwoID, "viewer")

	firstPage := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships?limit=2",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	nextCursor := requireNextCursor(t, firstPageBody)

	patchResp := httptestx.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+memberTwoID,
		map[string]any{
			"base_membership_version": memberTwo["membership_version"],
			"role":                    "reviewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	continued := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships?cursor_token="+url.QueryEscape(nextCursor),
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	continuedBody := httptestx.RequireSuccessEnvelope(t, continued, http.StatusOK)
	continuedRows := continuedBody["data"].(map[string]any)["memberships"].([]any)
	if len(continuedRows) != 1 {
		t.Fatalf("expected one membership on continuation page, got %#v", continuedRows)
	}
	liveContinuedMembership := continuedRows[0].(map[string]any)
	if liveContinuedMembership["user_id"] != memberTwoID || liveContinuedMembership["role"] != "reviewer" {
		t.Fatalf("expected live membership payload, got %#v", liveContinuedMembership)
	}

	fresh := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	freshBody := httptestx.RequireSuccessEnvelope(t, fresh, http.StatusOK)
	liveMembership := findByKey(t, freshBody["data"].(map[string]any)["memberships"].([]any), "user_id", memberTwoID)
	if liveMembership["role"] != "reviewer" {
		t.Fatalf("expected fresh membership list to reflect live role, got %#v", liveMembership)
	}
}

func createMembership(
	t testing.TB,
	harness *appsupport.ServerHarness,
	adminLogin flowtest.LoginResult,
	incidentID string,
	clientTxnID string,
	userID string,
	role string,
) map[string]any {
	t.Helper()

	resp := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": clientTxnID,
			"user_id":       userID,
			"role":          role,
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
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
