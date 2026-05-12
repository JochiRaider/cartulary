package incidents_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase1test"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestSupportPhase2_IncidentListUsesLiveFirstPageIndependentlyOfTestClock(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-pagination-live-first-page")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-phase2-live-first-create",
		"incident_key":  "IR-PAGINATION-LIVE",
		"title":         "Live First Page",
	})
	incidentID := incident["incident_id"].(string)

	patchResp := phase2test.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "amber",
			"current_phase":         "containment",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	phase1test.WithClockOffset(t, harness.Server.HTTP.URL, -600)

	listResp := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents",
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie),
	)
	listBody := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)
	incidents := listBody["data"].(map[string]any)["incidents"].([]any)
	incidentRow := findByKey(t, incidents, "incident_id", incidentID)
	if incidentRow["tlp"] != "amber" || incidentRow["current_phase"] != "containment" {
		t.Fatalf("expected patched incident to remain visible on first page, got %#v", incidentRow)
	}
}

func TestSupportPhase2_IncidentListContinuationUsesLiveMembershipQuery(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-pagination-incidents")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	first := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-phase2-incidents-first",
		"incident_key":  "IR-PAGINATION-FIRST",
		"title":         "First Incident",
	})
	firstID := first["incident_id"].(string)

	httptestx.SetClockAfter(t, harness.Server, mustParseTimestamp(t, first["updated_at"]), time.Second)
	second := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-phase2-incidents-second",
		"incident_key":  "IR-PAGINATION-SECOND",
		"title":         "Second Incident",
	})
	secondID := second["incident_id"].(string)
	httptestx.SetClockAfter(t, harness.Server, mustParseTimestamp(t, second["updated_at"]), time.Second)

	firstPage := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1",
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie),
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

	patchResp := phase2test.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+firstID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "amber",
			"current_phase":         "containment",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	continued := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?cursor_token="+url.QueryEscape(nextCursor),
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie),
	)
	continuedBody := httptestx.RequireSuccessEnvelope(t, continued, http.StatusOK)
	continuedRows := continuedBody["data"].(map[string]any)["incidents"].([]any)
	if len(continuedRows) != 0 {
		t.Fatalf("expected updated incident to move outside anchored continuation page, got %#v", continuedRows)
	}

	fresh := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1",
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie),
	)
	freshBody := httptestx.RequireSuccessEnvelope(t, fresh, http.StatusOK)
	freshRows := freshBody["data"].(map[string]any)["incidents"].([]any)
	if len(freshRows) != 1 {
		t.Fatalf("expected one incident on refreshed first page, got %#v", freshRows)
	}
	liveIncident := freshRows[0].(map[string]any)
	if liveIncident["incident_id"] != firstID || liveIncident["tlp"] != "amber" || liveIncident["current_phase"] != "containment" {
		t.Fatalf("expected fresh request to reflect live ordering and payload, got %#v", liveIncident)
	}
}

func TestSupportPhase2_IncidentListContinuationOmitsRevokedMembership(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-pagination-revoked-membership")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	viewerID := phase2test.SeedLocalUserFlags(t, harness.DB, "pagination-viewer@example.test", "Pagination Viewer", "PaginationViewer1!", false, false, true)
	viewerSession, _ := phase2test.LoginLocalUser(t, harness.Server, "pagination-viewer@example.test", "PaginationViewer1!")

	first := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-phase2-revoke-first",
		"incident_key":  "IR-PAGINATION-REVOKE-FIRST",
		"title":         "Revoked First Incident",
	})
	firstID := first["incident_id"].(string)
	phase2test.CreateMembership(t, harness.Server, adminLogin, firstID, map[string]any{
		"client_txn_id": "txn-support-phase2-revoke-first-membership",
		"user_id":       viewerID,
		"role":          "viewer",
	})

	httptestx.SetClockAfter(t, harness.Server, mustParseTimestamp(t, first["updated_at"]), time.Second)
	second := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-phase2-revoke-second",
		"incident_key":  "IR-PAGINATION-REVOKE-SECOND",
		"title":         "Visible Second Incident",
	})
	secondID := second["incident_id"].(string)
	phase2test.CreateMembership(t, harness.Server, adminLogin, secondID, map[string]any{
		"client_txn_id": "txn-support-phase2-revoke-second-membership",
		"user_id":       viewerID,
		"role":          "viewer",
	})

	firstPage := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?limit=1",
		nil,
		phase2test.WithCookies(viewerSession),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	firstPageRows := firstPageBody["data"].(map[string]any)["incidents"].([]any)
	if len(firstPageRows) != 1 || firstPageRows[0].(map[string]any)["incident_id"] != secondID {
		t.Fatalf("expected viewer first page to return second incident, got %#v", firstPageRows)
	}
	nextCursor := requireNextCursor(t, firstPageBody)

	phase2test.DeleteMembership(t, harness.Server, adminLogin, firstID, viewerID, map[string]any{
		"base_membership_version": 1,
	})

	continued := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents?cursor_token="+url.QueryEscape(nextCursor),
		nil,
		phase2test.WithCookies(viewerSession),
	)
	continuedBody := httptestx.RequireSuccessEnvelope(t, continued, http.StatusOK)
	continuedRows := continuedBody["data"].(map[string]any)["incidents"].([]any)
	if len(continuedRows) != 0 {
		t.Fatalf("expected revoked incident to be omitted from continuation, got %#v", continuedRows)
	}
}

func TestSupportPhase2_MembershipListContinuationUsesLiveRows(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-pagination-memberships")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-phase2-memberships-incident",
		"incident_key":  "IR-PAGINATION-MEMBERSHIPS",
		"title":         "Membership Pagination",
	})
	incidentID := incident["incident_id"].(string)

	memberOneID := phase2test.SeedLocalUserFlags(t, harness.DB, "membership-one@example.test", "Membership One", "MembershipOne1!", false, false, true)
	memberTwoID := phase2test.SeedLocalUserFlags(t, harness.DB, "membership-two@example.test", "Membership Two", "MembershipTwo1!", false, false, true)

	createMembership(t, harness, adminLogin, incidentID, "txn-support-phase2-membership-one", memberOneID, "viewer")
	memberTwo := createMembership(t, harness, adminLogin, incidentID, "txn-support-phase2-membership-two", memberTwoID, "viewer")

	firstPage := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships?limit=2",
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	nextCursor := requireNextCursor(t, firstPageBody)

	patchResp := phase2test.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+memberTwoID,
		map[string]any{
			"base_membership_version": memberTwo["membership_version"],
			"role":                    "reviewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	continued := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships?cursor_token="+url.QueryEscape(nextCursor),
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie),
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

	fresh := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie),
	)
	freshBody := httptestx.RequireSuccessEnvelope(t, fresh, http.StatusOK)
	liveMembership := findByKey(t, freshBody["data"].(map[string]any)["memberships"].([]any), "user_id", memberTwoID)
	if liveMembership["role"] != "reviewer" {
		t.Fatalf("expected fresh membership list to reflect live role, got %#v", liveMembership)
	}
}

func createMembership(
	t testing.TB,
	harness *phase2test.ServerHarness,
	adminLogin phase2test.LoginResult,
	incidentID string,
	clientTxnID string,
	userID string,
	role string,
) map[string]any {
	t.Helper()

	resp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": clientTxnID,
			"user_id":       userID,
			"role":          role,
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
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
