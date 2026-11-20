package collaboration_test

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	incidentscenariotest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

// TestIncidentSocketUpgradeAdmissionOrder_Integration covers Core 01
// REQ-01-708 and Core 04 AC-131: before the upgrade, route resolution comes
// first, so an unknown /ws/v1/ path receives the enveloped 404
// route_not_found and any method but GET on the socket route, HEAD included,
// the enveloped 405 method_not_allowed with Allow: GET; then a request carrying
// an Origin other than the configured application origin receives the
// enveloped 403 csrf_verification_failed, with or without a session cookie;
// then session authentication precedes the incident identifier, so a request
// without a session receives 401 session_required even for a malformed
// identifier; an authenticated malformed identifier receives the same
// enveloped 404 incident_not_found as a missing incident and as an incident the
// caller cannot see; and a session the upgrade finds expired is published to
// that session's other sockets (REQ-01-277), because the upgrade authenticates
// through the central authenticator.
func TestIncidentSocketUpgradeAdmissionOrder_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness, admin, incidentID := setupSocketIncident(t, runtime, "collaboration-upgrade-admission")
	base := harness.Server.HTTP.URL
	malformed := []string{"not-a-uuid", strings.ReplaceAll(incidentID, "-", ""), "%7B" + incidentID + "%7D", incidentID + "0", "a%2Fb"}
	const untrustedOrigin = "https://untrusted.example.test"

	t.Run("route resolution precedes the Origin check and authentication", func(t *testing.T) {
		route := "/ws/v1/incidents/" + incidentID
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodHead} {
			for _, cookies := range [][]*http.Cookie{nil, {admin.Session}} {
				response := socketRouteRequest(t, method, base+route, untrustedOrigin, cookies...)
				requireSocketRouteRejection(t, method, route, response, http.StatusMethodNotAllowed, "method_not_allowed", "GET")
			}
		}
		unknown := []string{
			"/ws/v1", "/ws/v1/", "/ws/v1/incidents", "/ws/v1/incidents/", route + "/presence", "/ws/v1/sessions/" + incidentID,
			"/ws/v1//incidents/" + incidentID, "/ws/v1/incidents/x/../" + incidentID, "/ws/v1/./incidents/" + incidentID,
		}
		for _, path := range unknown {
			for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost} {
				response := socketRouteRequest(t, method, base+path, untrustedOrigin, admin.Session)
				requireSocketRouteRejection(t, method, path, response, http.StatusNotFound, "route_not_found", "")
			}
		}
	})
	t.Run("an untrusted Origin is rejected before authentication", func(t *testing.T) {
		for name, options := range map[string]incidentwstest.ConnectOptions{
			"cookie session":          {Cookies: []*http.Cookie{admin.Session}, Origin: untrustedOrigin},
			"bearer session":          {SessionToken: admin.Session.Value, Origin: untrustedOrigin},
			"no session":              {Origin: untrustedOrigin},
			"opaque origin":           {SessionToken: admin.Session.Value, Origin: "null"},
			"configured host on http": {SessionToken: admin.Session.Value, Origin: "http" + strings.TrimPrefix(harness.Server.Config.Application.PublicOrigin, "https")},
		} {
			for _, id := range []string{incidentID, "not-a-uuid"} {
				_, response, err := incidentwstest.TryConnect(base, id, options)
				if err == nil || response == nil {
					t.Fatalf("%s upgrade of %s succeeded or lost its response: %v", name, id, err)
				}
				body := httptestx.RequireErrorEnvelope(t, response, http.StatusForbidden, "csrf_verification_failed")
				if details, _ := body["error"].(map[string]any)["details"].(map[string]any); len(details) != 0 {
					t.Fatalf("%s Origin rejection details = %v, want empty", name, details)
				}
			}
		}
		// Controls: the configured origin and a request without Origin reach
		// authentication, and an authenticated configured-origin client connects.
		trusted := harness.Server.Config.Application.PublicOrigin
		incidentwstest.RequireDialErrorEnvelope(t, base, incidentID, incidentwstest.ConnectOptions{Origin: trusted}, http.StatusUnauthorized, "session_required")
		incidentwstest.RequireDialErrorEnvelope(t, base, incidentID, incidentwstest.ConnectOptions{}, http.StatusUnauthorized, "session_required")
		socket := incidentwstest.ConnectAndHello(t, base, incidentID, incidentwstest.ConnectOptions{SessionToken: admin.Session.Value, Origin: trusted})
		socket.Close(wstest.StatusNormalClosure, "integration_cleanup")
	})
	t.Run("authentication precedes the incident identifier", func(t *testing.T) {
		for _, id := range append([]string{incidentID}, malformed...) {
			incidentwstest.RequireDialErrorEnvelope(t, base, id, incidentwstest.ConnectOptions{}, http.StatusUnauthorized, "session_required")
		}
	})
	t.Run("a malformed identifier receives the hidden-incident response", func(t *testing.T) {
		const password = "UpgradeOutsiderPass1!"
		outsider := flowtest.SeedLocalUserRecord(t, harness.DB, "upgrade-outsider@example.test", "Upgrade outsider", password, false, false, true)
		outsiderLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, outsider.Email, password, nil)
		hidden := dialRejection(t, base, incidentID, outsiderLogin.Session.Value)
		missing := dialRejection(t, base, uuid.NewString(), admin.Session.Value)
		if !reflect.DeepEqual(hidden, missing) {
			t.Fatalf("hidden incident %v differs from missing incident %v", hidden, missing)
		}
		for _, id := range malformed {
			if got := dialRejection(t, base, id, admin.Session.Value); !reflect.DeepEqual(got, hidden) {
				t.Fatalf("malformed %q response %v differs from the hidden-incident response %v", id, got, hidden)
			}
		}
	})
	t.Run("an expired session found at the upgrade is published to its sockets", func(t *testing.T) {
		const password = "UpgradeExpiringPass1!"
		member := flowtest.SeedLocalUserRecord(t, harness.DB, "upgrade-expiring@example.test", "Upgrade expiring", password, false, false, true)
		incidentscenariotest.CreateMembershipForUser(t, harness.Server, admin, incidentID, member.ID.String(), member.Email, "viewer")
		login := flowtest.LoginLocalUser(t, http.DefaultClient, base, member.Email, password, nil)
		socket := incidentwstest.ConnectAndHello(t, base, incidentID, incidentwstest.ConnectOptions{SessionToken: login.Session.Value})
		defer socket.Close(wstest.StatusNormalClosure, "integration_cleanup")
		// Expire the session in storage only: the open socket still holds the
		// expiry it saw at connect time, so only the upgrade's publication can
		// revoke it now.
		if _, err := harness.DB.Exec(`UPDATE user_sessions SET session_expires_at=authenticated_at-interval '1 second', idle_expires_at=authenticated_at-interval '1 second' WHERE user_id=$1 AND revoked_at IS NULL`, member.ID); err != nil {
			t.Fatal(err)
		}
		incidentwstest.RequireDialErrorEnvelope(t, base, incidentID, incidentwstest.ConnectOptions{SessionToken: login.Session.Value}, http.StatusUnauthorized, "session_required")
		incidentwstest.ExpectSessionRevoked(t, socket, "session_expired")
		var reason string
		if err := harness.DB.QueryRow(`SELECT revoke_reason_code FROM user_sessions WHERE user_id=$1`, member.ID).Scan(&reason); err != nil || reason != "session_expired" {
			t.Fatalf("revoke_reason_code = %q, %v", reason, err)
		}
	})
}

// socketRouteRequest sends one plain HTTP request to a /ws path, carrying
// origin and cookies, without following a redirect.
func socketRouteRequest(t *testing.T, method, target, origin string, cookies ...*http.Cookie) *http.Response {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), method, target, nil)
	if err != nil {
		t.Fatalf("build %s %s: %v", method, target, err)
	}
	request.Header.Set("Origin", origin)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	t.Cleanup(func() { _ = response.Body.Close() })
	return response
}

// requireSocketRouteRejection requires REQ-01-708 step 1's enveloped
// rejection; a HEAD response carries its headers only.
func requireSocketRouteRejection(t *testing.T, method, path string, response *http.Response, status int, code, allow string) {
	t.Helper()
	if got := response.Header.Get("Allow"); got != allow {
		t.Fatalf("%s %s Allow = %q, want %q (status %d)", method, path, got, allow, response.StatusCode)
	}
	if method == http.MethodHead {
		body, err := io.ReadAll(response.Body)
		if err != nil || response.StatusCode != status || len(body) != 0 || response.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("HEAD %s = %d %q body=%q err=%v, want %d application/json without a body", path, response.StatusCode, response.Header.Get("Content-Type"), body, err, status)
		}
		return
	}
	body := httptestx.RequireErrorEnvelope(t, response, status, code)
	want := map[string]any{}
	if allow != "" {
		want["allowed_methods"] = []any{allow}
	}
	if details := body["error"].(map[string]any)["details"]; !reflect.DeepEqual(details, want) {
		t.Fatalf("%s %s details = %v, want %v", method, path, details, want)
	}
}

// dialRejection returns the upgrade's error envelope without its request ID.
func dialRejection(t *testing.T, base, incidentID, sessionToken string) map[string]any {
	t.Helper()
	_, response, err := incidentwstest.TryConnect(base, incidentID, incidentwstest.ConnectOptions{SessionToken: sessionToken})
	if err == nil || response == nil {
		t.Fatalf("dial %s succeeded or lost its response: %v", incidentID, err)
	}
	body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "incident_not_found")
	delete(body["error"].(map[string]any), "request_id")
	return body
}
