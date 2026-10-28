package incidents_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// The controller changes membership through real routes. The subject keeps the
// same authenticated session for every request, including after downgrade/removal.
// This fixture contains no operation targets or expected permission policy.
type authorizationFixture struct {
	harness         *appsupport.ServerHarness
	controller      authcookietest.AuthCookies
	subject         authcookietest.AuthCookies
	subjectID       string
	deploymentAdmin authcookietest.AuthCookies
	incident        map[string]any
	incidentID      string
	membership      map[string]any
}

func newAuthorizationAccounts(t testing.TB, prefix string) *authorizationFixture {
	t.Helper()
	harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{
		Prefix: prefix, TestRouteMode: httptestx.TestRouteModeDisabled,
	})
	controller, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	subject, subjectID := loginAuthorizationActor(t, harness, false)
	return &authorizationFixture{harness: harness, controller: controller, subject: subject, subjectID: subjectID}
}

func newIncidentAuthorization(t testing.TB, prefix string) *authorizationFixture {
	t.Helper()
	f := newAuthorizationAccounts(t, prefix)
	incident := scenariotest.CreateIncident(t, f.harness.Server, f.controller, map[string]any{
		"client_txn_id": uuid.NewString(), "incident_key": "IR-" + uuid.NewString(), "title": "Authorization incident",
	})
	f.incident = incident
	f.incidentID = incident["incident_id"].(string)
	f.deploymentAdmin, _ = loginAuthorizationActor(t, f.harness, true)
	return f
}

func loginAuthorizationActor(t testing.TB, harness *appsupport.ServerHarness, deploymentAdmin bool) (authcookietest.AuthCookies, string) {
	t.Helper()
	email := uuid.NewString() + "@example.test"
	const password = "AuthorizationSubject1!"
	id := flowtest.SeedLocalUserFlags(t, harness.DB, email, "Authorization actor", password, false, deploymentAdmin, true)
	login := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, email, password, nil)
	return login, id
}

func (f *authorizationFixture) setSubjectRole(t testing.TB, role string) {
	t.Helper()
	if role == "" {
		if f.membership != nil {
			resp := authorizationWrite(t, f, f.controller, http.MethodDelete,
				"/api/v1/incidents/"+f.incidentID+"/memberships/"+f.subjectID,
				map[string]any{"base_membership_version": f.membership["membership_version"]})
			requireAuthorizationResponse(t, resp, 204, "")
			f.membership = nil
		}
		return
	}
	if f.membership == nil {
		f.membership = scenariotest.CreateMembership(t, f.harness.Server, f.controller, f.incidentID,
			map[string]any{"client_txn_id": uuid.NewString(), "user_id": f.subjectID, "role": role})
		return
	}
	f.membership = scenariotest.PatchMembership(t, f.harness.Server, f.controller, f.incidentID, f.subjectID,
		map[string]any{"base_membership_version": f.membership["membership_version"], "role": role})
}

func authorizationWrite(t testing.TB, f *authorizationFixture, actor authcookietest.AuthCookies, method, path string, body any) *http.Response {
	t.Helper()
	return httptestx.DoJSON(t, http.DefaultClient, method, f.harness.Server.HTTP.URL+path, body,
		httptestx.WithCookies(actor.Session, actor.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRF.Value))
}

func requireWriteAuthentication(t testing.TB, f *authorizationFixture, method, path string, body any) {
	t.Helper()
	anonymous := httptestx.DoJSON(t, http.DefaultClient, method, f.harness.Server.HTTP.URL+path, body)
	httptestx.RequireErrorEnvelope(t, anonymous, 401, "session_required")
	missingCSRF := httptestx.DoJSON(t, http.DefaultClient, method, f.harness.Server.HTTP.URL+path, body,
		httptestx.WithCookies(f.controller.Session, f.controller.CSRF))
	httptestx.RequireErrorEnvelope(t, missingCSRF, 403, "csrf_verification_failed")
}

func requireReadAuthentication(t testing.TB, f *authorizationFixture, method, path string, body any) {
	t.Helper()
	resp := httptestx.DoJSON(t, http.DefaultClient, method, f.harness.Server.HTTP.URL+path, body)
	httptestx.RequireErrorEnvelope(t, resp, 401, "session_required")
}

func requireAuthorizationResponse(t testing.TB, resp *http.Response, status int, code string) map[string]any {
	t.Helper()
	if code != "" {
		httptestx.RequireErrorEnvelope(t, resp, status, code)
		return nil
	}
	if status == http.StatusNoContent {
		httptestx.RequireNoContent(t, resp)
		httptestx.RequireRequestID(t, resp)
		return nil
	}
	return httptestx.RequireSuccessEnvelope(t, resp, status)["data"].(map[string]any)
}

func authorizationVersion(t testing.TB, value any) int64 {
	t.Helper()
	version, ok := value.(float64)
	if !ok || version < 1 || version != float64(int64(version)) {
		t.Fatalf("invalid response version: %v", value)
	}
	return int64(version)
}
