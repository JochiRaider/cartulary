package incidents_test

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/mutationtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentLifecycleCloseReopenAuthorizationReplayAndTransitions_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "incident-lifecycle-close-reopen")
	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-incident-lifecycle-create",
		"incident_key":  "IR-LIFECYCLE",
		"title":         "Lifecycle contract",
	})
	incidentID := incident["incident_id"].(string)
	path := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID

	noSession := httptestx.DoJSON(
		t,
		http.MethodPost,
		path+"/close",
		map[string]any{
			"base_incident_version": 1,
			"client_txn_id":         "txn-lifecycle-no-session",
			"reason":                "unauthenticated",
		},
	)
	httptestx.RequireErrorEnvelope(t, noSession, http.StatusUnauthorized, "session_required")

	noCSRF := httptestx.DoJSON(
		t,
		http.MethodPost,
		path+"/close",
		map[string]any{
			"base_incident_version": 1,
			"client_txn_id":         "txn-lifecycle-no-csrf",
			"reason":                "missing csrf",
		},
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, noCSRF, http.StatusForbidden, "csrf_verification_failed")

	const viewerSecretBase32 = "JBSWY3DPEHPK3PXP"
	viewerID := flowtest.SeedLocalUserWithActiveTOTP(
		t,
		harness.DB,
		"incident-lifecycle-viewer@example.test",
		"Incident Lifecycle Viewer",
		"IncidentLifecycleViewer1!",
		true,
		false,
		viewerSecretBase32,
	)
	viewerLogin := flowtest.LoginLocalUserWithSecondFactor(
		t,
		harness.Server.HTTP.URL,
		"incident-lifecycle-viewer@example.test",
		"IncidentLifecycleViewer1!",
		flowtest.GenerateTOTPCode(t, viewerSecretBase32),
	)
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-lifecycle-viewer-membership",
		"user_id":       viewerID,
		"role":          "viewer",
	})
	viewerDenied := httptestx.DoJSON(
		t,
		http.MethodPost,
		path+"/close",
		map[string]any{"unexpected": true},
		httptestx.WithCookies(viewerLogin.SessionCookie, viewerLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, viewerDenied, http.StatusForbidden, "authorization_denied")

	const deploymentAdminSecretBase32 = "JBSWY3DPEHPK3QAA"
	flowtest.SeedLocalUserWithActiveTOTP(
		t,
		harness.DB,
		"incident-lifecycle-nonmember@example.test",
		"Incident Lifecycle Nonmember",
		"IncidentLifecycleNonmember1!",
		true,
		true,
		deploymentAdminSecretBase32,
	)
	deploymentAdminLogin := flowtest.LoginLocalUserWithSecondFactor(
		t,
		harness.Server.HTTP.URL,
		"incident-lifecycle-nonmember@example.test",
		"IncidentLifecycleNonmember1!",
		flowtest.GenerateTOTPCode(t, deploymentAdminSecretBase32),
	)
	hidden := httptestx.DoJSON(
		t,
		http.MethodPost,
		path+"/reopen",
		map[string]any{"unexpected": true},
		httptestx.WithCookies(deploymentAdminLogin.SessionCookie, deploymentAdminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, deploymentAdminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, hidden, http.StatusNotFound, "incident_not_found")

	closeTime := time.Date(2026, 7, 25, 18, 45, 0, 0, time.UTC)
	httptestx.SetClockFixed(t, harness.Server, closeTime)
	closeRequest := map[string]any{
		"base_incident_version": 1,
		"client_txn_id":         "txn-lifecycle-close",
		"reason":                "  Close after review.\r\nReady.  ",
	}
	closeResponse := lifecycleRequest(t, path+"/close", closeRequest, adminLogin)
	closeData := httptestx.RequireSuccessEnvelope(t, closeResponse, http.StatusOK)["data"].(map[string]any)
	if closeData["status"] != "closed" ||
		closeData["incident_version"] != float64(2) ||
		closeData["closed_at"] != closeTime.Format(time.RFC3339) ||
		closeData["updated_at"] != closeTime.Format(time.RFC3339) ||
		closeData["updated_by_user_id"] != adminID {
		t.Fatalf("unexpected close resource: %#v", closeData)
	}

	replayRequest := map[string]any{
		"base_incident_version": 1,
		"client_txn_id":         "txn-lifecycle-close",
		"reason":                "Close after review.\nReady.",
	}
	replayData := httptestx.RequireSuccessEnvelope(
		t,
		lifecycleRequest(t, path+"/close", replayRequest, adminLogin),
		http.StatusOK,
	)["data"].(map[string]any)
	if !reflect.DeepEqual(closeData, replayData) {
		t.Fatalf("close replay changed the committed resource: first=%#v replay=%#v", closeData, replayData)
	}
	divergent := lifecycleRequest(t, path+"/close", map[string]any{
		"base_incident_version": 1,
		"client_txn_id":         "txn-lifecycle-close",
		"reason":                "different",
	}, adminLogin)
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")

	repeatedClose := lifecycleRequest(t, path+"/close", map[string]any{
		"base_incident_version": 2,
		"client_txn_id":         "txn-lifecycle-close-again",
		"reason":                "close twice",
	}, adminLogin)
	repeatedCloseBody := httptestx.RequireErrorEnvelope(t, repeatedClose, http.StatusConflict, "illegal_transition")
	requireLifecycleReasonCode(t, repeatedCloseBody, "incident_already_closed")

	staleReopen := lifecycleRequest(t, path+"/reopen", map[string]any{
		"base_incident_version": 1,
		"client_txn_id":         "txn-lifecycle-reopen-stale",
		"reason":                "stale reopen",
	}, adminLogin)
	httptestx.RequireErrorEnvelope(t, staleReopen, http.StatusConflict, "incident_version_conflict")

	reopenTime := closeTime.Add(time.Minute)
	httptestx.SetClockFixed(t, harness.Server, reopenTime)
	reopenResponse := lifecycleRequest(t, path+"/reopen", map[string]any{
		"base_incident_version": 2,
		"client_txn_id":         "txn-lifecycle-reopen",
		"reason":                "Resume response.",
	}, adminLogin)
	reopenData := httptestx.RequireSuccessEnvelope(t, reopenResponse, http.StatusOK)["data"].(map[string]any)
	if reopenData["status"] != "active" ||
		reopenData["incident_version"] != float64(3) ||
		reopenData["closed_at"] != nil ||
		reopenData["updated_at"] != reopenTime.Format(time.RFC3339) ||
		reopenData["updated_by_user_id"] != adminID {
		t.Fatalf("unexpected reopen resource: %#v", reopenData)
	}

	repeatedReopen := lifecycleRequest(t, path+"/reopen", map[string]any{
		"base_incident_version": 3,
		"client_txn_id":         "txn-lifecycle-reopen-again",
		"reason":                "reopen twice",
	}, adminLogin)
	repeatedReopenBody := httptestx.RequireErrorEnvelope(t, repeatedReopen, http.StatusConflict, "illegal_transition")
	requireLifecycleReasonCode(t, repeatedReopenBody, "incident_not_closed")

	invalid := lifecycleRequest(t, path+"/close", map[string]any{
		"base_incident_version": 3,
		"client_txn_id":         "txn-lifecycle-invalid",
		"reason":                nil,
	}, adminLogin)
	invalidBody := httptestx.RequireErrorEnvelope(t, invalid, http.StatusBadRequest, "invalid_incident_lifecycle_request")
	requireLifecycleReasonCode(t, invalidBody, "field_not_nullable")

	events := mutationtest.LookupOwnerMutations(
		t,
		mutationtest.SQLDatabase(harness.DB),
		mutationtest.MutationSelector{IncidentID: incidentID},
		mutationtest.MutationOwnerIncidentResource,
	)
	lifecycleEvents := make([]mutationtest.AuditEventRecord, 0, 2)
	for _, event := range events {
		if event.EventKind == "incident_close" || event.EventKind == "incident_reopen" {
			lifecycleEvents = append(lifecycleEvents, event)
		}
	}
	if len(lifecycleEvents) != 2 ||
		lifecycleEvents[0].EventKind != "incident_close" ||
		lifecycleEvents[0].ReasonCode != "Close after review.\nReady." ||
		lifecycleEvents[1].EventKind != "incident_reopen" ||
		lifecycleEvents[1].ReasonCode != "Resume response." {
		t.Fatalf("unexpected lifecycle audit events: %#v", lifecycleEvents)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key IN ('incidents.close', 'incidents.reopen')
   AND scope_key = $1
`, incidentID); got != 2 {
		t.Fatalf("expected exactly two committed lifecycle idempotency rows, got %d", got)
	}

	bearerIncident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-incident-lifecycle-bearer-create",
		"incident_key":  "IR-LIFECYCLE-BEARER",
		"title":         "Lifecycle bearer contract",
	})
	bearerClose := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+bearerIncident["incident_id"].(string)+"/close",
		map[string]any{
			"base_incident_version": 1,
			"client_txn_id":         "txn-lifecycle-bearer-close",
			"reason":                "Bearer session requires no CSRF proof.",
		},
		httptestx.WithHeader("Authorization", "Bearer "+adminLogin.SessionCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, bearerClose, http.StatusOK)
}

func lifecycleRequest(
	t testing.TB,
	url string,
	body map[string]any,
	login flowtest.LoginResult,
) *http.Response {
	t.Helper()
	return httptestx.DoJSON(
		t,
		http.MethodPost,
		url,
		body,
		httptestx.WithCookies(login.SessionCookie, login.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func requireLifecycleReasonCode(t testing.TB, body map[string]any, want string) {
	t.Helper()
	errorObject := body["error"].(map[string]any)
	details := errorObject["details"].(map[string]any)
	if details["reason_code"] != want {
		t.Fatalf("reason_code got %#v want %q", details["reason_code"], want)
	}
}
