package reference_data_test

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestReferencePackFamilyAdmissionOrder_Integration drives the Reference Pack
// family through central admission in its own order (Core 01 §17.4,
// REQ-01-707): session authentication, current deployment_admin, structural
// canonicality, method and operation selection, the operation's cookie CSRF,
// then the query before path parameters. An unauthenticated caller gets the
// ordinary authentication failure and a bootstrap token is not allowed on any
// family route; a valid non-admin session is denied before canonicality,
// unknown descendants, method handling and CSRF, without sliding its session;
// an admin receives the enveloped route_not_found and method_not_allowed
// rejections, a HEAD 405 without a body, CSRF before query errors and query
// errors before a malformed pack_key. A session found expired at the gate is
// revoked and published to its sockets. No rejection changes pack state.
func TestReferencePackFamilyAdmissionOrder_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	h := startReferencePackServer(t, runtime, "reference-family-admission-order")
	base := h.Server.HTTP.URL
	admin, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	const password = "ReferenceFamilyOrder1!"
	member := flowtest.SeedLocalUserRecord(t, h.DB, "family-order-member@example.test", "Family order member", password, false, false, true)
	memberLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, member.Email, password, nil)
	expiring := flowtest.SeedLocalUserRecord(t, h.DB, "family-order-expiring@example.test", "Family order expiring", password, false, false, true)
	expiringLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, expiring.Email, password, nil)
	setup := flowtest.SeedLocalUserRecord(t, h.DB, "family-order-setup@example.test", "Family order setup", password, true, false, true)
	bootstrapToken := flowtest.RequireBootstrapLogin(t, http.DefaultClient, base, setup.Email, password)
	// The expiring caller's socket attaches to an incident it is a member of.
	incidentID := appsupport.MustUUID(t, appsupport.CreateIncident(t, h.Server, admin, map[string]any{
		"client_txn_id": "txn-reference-family-order-incident",
		"incident_key":  "IR-REFERENCE-FAMILY-ORDER",
		"title":         "Reference Pack family admission order",
	})["incident_id"].(string))
	appsupport.SeedMembershipSQL(t, h.DB, incidentID, expiring.ID, "viewer", adminID)

	// A redirect is never followed, so a 3xx would surface as itself.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	send := func(t testing.TB, method, path string, options ...func(*http.Request)) *http.Response {
		t.Helper()
		return httptestx.DoRawJSON(t, client, method, base+path, "{", options...)
	}
	cookie := func(login authcookietest.AuthCookies) func(*http.Request) {
		return httptestx.WithCookies(login.Session)
	}
	csrf := func(login authcookietest.AuthCookies) func(*http.Request) {
		return func(r *http.Request) {
			r.AddCookie(login.CSRF)
			r.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
		}
	}
	bearer := func(token string) func(*http.Request) {
		return httptestx.WithHeader("Authorization", "Bearer "+token)
	}
	requireRejection := func(t testing.TB, response *http.Response, status int, code string, details map[string]any) map[string]any {
		t.Helper()
		if location := response.Header.Get("Location"); location != "" {
			t.Fatalf("rejection redirected to %s", location)
		}
		body := httptestx.RequireErrorEnvelope(t, response, status, code)
		if got := httptestx.RequireErrorDetails(t, body); !reflect.DeepEqual(got, details) {
			t.Fatalf("%s details = %#v, want %#v", code, got, details)
		}
		return body
	}
	requireMethodNotAllowed := func(t testing.TB, response *http.Response, allow string) {
		t.Helper()
		if got := response.Header.Get("Allow"); got != allow {
			t.Fatalf("Allow = %q, want %q", got, allow)
		}
		requireRejection(t, response, http.StatusMethodNotAllowed, "method_not_allowed", map[string]any{"allowed_methods": []any{allow}})
	}
	sessionActivity := func(t testing.TB, userID string) time.Time {
		t.Helper()
		var activity time.Time
		if err := h.DB.QueryRow(`SELECT last_qualifying_activity_at FROM user_sessions WHERE user_id=$1 AND revoked_at IS NULL`, userID).Scan(&activity); err != nil {
			t.Fatal(err)
		}
		return activity
	}

	const root = "/api/v1/reference-packs"
	// Requests every caller below the admin gate is refused, whatever their
	// structure, method or credentials beyond the session.
	gated := []struct{ method, path string }{
		{http.MethodGet, root},
		{http.MethodGet, root + "//type_registry.host/1"},
		{http.MethodGet, "/api/v1/other/../reference-packs"},
		{http.MethodPost, root + "/../users"},
		{http.MethodGet, root + "/unknown/child/extra"},
		{http.MethodPatch, root},
		{http.MethodDelete, root + "/type_registry.host/1"},
		{http.MethodPost, root + "/import?limit=1"},
		{http.MethodPost, root + "/TYPE_REGISTRY/1/reverify?private-query-value=secret"},
	}
	before := packAdmissionState(t, h)

	t.Run("an unauthenticated caller receives the ordinary authentication failure", func(t *testing.T) {
		for _, tc := range gated {
			requireRejection(t, send(t, tc.method, tc.path), http.StatusUnauthorized, "session_required", map[string]any{})
		}
	})
	t.Run("a bootstrap token is not allowed on any family route", func(t *testing.T) {
		for _, tc := range gated {
			requireRejection(t, send(t, tc.method, tc.path, bearer(bootstrapToken)), http.StatusConflict, "credential_bootstrap_rejected", map[string]any{"reason_code": "not_allowed_for_route"})
		}
	})
	t.Run("a non-admin session is denied before every route check without CSRF", func(t *testing.T) {
		activity := sessionActivity(t, member.ID.String())
		denied := map[string]any{"required_capability": "deployment_admin"}
		for _, tc := range gated {
			requireRejection(t, send(t, tc.method, tc.path, cookie(memberLogin)), http.StatusForbidden, "authorization_denied", denied)
			requireRejection(t, send(t, tc.method, tc.path, bearer(memberLogin.Session.Value)), http.StatusForbidden, "authorization_denied", denied)
		}
		if response := send(t, http.MethodHead, root, cookie(memberLogin)); response.StatusCode != http.StatusForbidden {
			t.Fatalf("non-admin HEAD = %d", response.StatusCode)
		}
		if after := sessionActivity(t, member.ID.String()); !after.Equal(activity) {
			t.Fatalf("authorization denial slid the session: %s -> %s", activity, after)
		}
	})
	t.Run("an admin receives the enveloped structural, route and method rejections", func(t *testing.T) {
		for _, path := range []string{root + "//type_registry.host/1", "/api/v1/other/../reference-packs", root + "/../users", root + "/unknown/..", root + "/unknown/child/extra", root + "/"} {
			for _, method := range []string{http.MethodGet, http.MethodPost} {
				requireRejection(t, send(t, method, path+"?invalid=%zz", cookie(admin)), http.StatusNotFound, "route_not_found", map[string]any{})
			}
		}
		requireMethodNotAllowed(t, send(t, http.MethodPatch, root+"?invalid=%zz", cookie(admin)), "GET")
		requireMethodNotAllowed(t, send(t, http.MethodGet, root+"/import", cookie(admin)), "POST")
		requireMethodNotAllowed(t, send(t, http.MethodDelete, root+"/type_registry.host/1", cookie(admin)), "GET")
		requireMethodNotAllowed(t, send(t, http.MethodGet, root+"/type_registry.host/1/reverify", cookie(admin)), "POST")
		head := send(t, http.MethodHead, root, cookie(admin))
		if head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != "GET" || head.ContentLength > 0 {
			t.Fatalf("HEAD = %d Allow=%q Content-Length=%d", head.StatusCode, head.Header.Get("Allow"), head.ContentLength)
		}
	})
	t.Run("an admin's cookie CSRF precedes the query, which precedes path parameters", func(t *testing.T) {
		query := "?private-query-value=secret"
		unknownMember := map[string]any{"field": "query", "reason_code": "unknown_query_member"}
		for _, path := range []string{root + "/refresh", root + "/import", root + "/TYPE_REGISTRY/1/reverify", root + "/type_registry.host/1/activate"} {
			requireRejection(t, send(t, http.MethodPost, path+query, cookie(admin)), http.StatusForbidden, "csrf_verification_failed", map[string]any{})
			requireRejection(t, send(t, http.MethodPost, path+query, cookie(admin), csrf(admin)), http.StatusBadRequest, "invalid_reference_pack_request", unknownMember)
			// A bearer session needs no CSRF proof.
			requireRejection(t, send(t, http.MethodPost, path+query, bearer(admin.Session.Value)), http.StatusBadRequest, "invalid_reference_pack_request", unknownMember)
		}
		for _, path := range []string{root + "/TYPE_REGISTRY/1", root + "/validation-summaries/absent"} {
			requireRejection(t, send(t, http.MethodGet, path+"?limit=1", cookie(admin)), http.StatusBadRequest, "invalid_pagination_request", map[string]any{"reason_code": "pagination_not_supported"})
		}
		// Without the query, the malformed pack_key receives the declared
		// not-found response, identical to a well-formed key naming no pack.
		missing := requireRejection(t, postAction(t, h, admin, root+"/enrichment.absent/1/reverify", "family-order-missing", ""), http.StatusNotFound, "reference_pack_not_found", map[string]any{})
		malformed := requireRejection(t, postAction(t, h, admin, root+"/TYPE_REGISTRY/1/reverify", "family-order-malformed", ""), http.StatusNotFound, "reference_pack_not_found", map[string]any{})
		for _, body := range []map[string]any{missing, malformed} {
			delete(body["error"].(map[string]any), "request_id")
		}
		if !reflect.DeepEqual(missing, malformed) {
			t.Fatalf("malformed pack_key response %v differs from the missing-pack response %v", malformed, missing)
		}
		for _, path := range []string{root + "/TYPE_REGISTRY/1", root + "/type_registry.host/a%2Fb", root + "/validation-summaries/absent"} {
			requireRejection(t, send(t, http.MethodGet, path, cookie(admin)), http.StatusNotFound, "reference_pack_not_found", map[string]any{})
		}
	})
	t.Run("an expired session is revoked and published to its sockets", func(t *testing.T) {
		socket := incidentwstest.ConnectAndHello(t, base, incidentID.String(), incidentwstest.ConnectOptions{SessionToken: expiringLogin.Session.Value})
		defer socket.Close(websocket.StatusNormalClosure, "integration_cleanup")
		// Expire the session in storage only: the socket still holds the
		// expiry it saw at connect time, so only the gate's publication can
		// revoke it now.
		if _, err := h.DB.Exec(`UPDATE user_sessions SET session_expires_at=authenticated_at-interval '1 second', idle_expires_at=authenticated_at-interval '1 second' WHERE user_id=$1 AND revoked_at IS NULL`, expiring.ID); err != nil {
			t.Fatal(err)
		}
		requireRejection(t, send(t, http.MethodGet, root, cookie(expiringLogin)), http.StatusUnauthorized, "session_required", map[string]any{})
		incidentwstest.ExpectSessionRevoked(t, socket, "session_expired")
		var reason string
		if err := h.DB.QueryRow(`SELECT revoke_reason_code FROM user_sessions WHERE user_id=$1`, expiring.ID).Scan(&reason); err != nil || reason != "session_expired" {
			t.Fatalf("revoke_reason_code = %q, %v", reason, err)
		}
	})
	if after := packAdmissionState(t, h); after != before {
		t.Error("rejected requests changed pack or admission state")
	}
}
