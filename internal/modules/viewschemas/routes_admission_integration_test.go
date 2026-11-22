package viewschemas_test

// View-schema and extension discovery route admission order (Core 01
// REQ-01-707, Core 04 AC-617).
import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestDiscoveryRouteAdmissionOrder_Integration drives the view-schema and
// extension discovery operations through central admission. Authentication
// precedes view_schema_id and every query check; a malformed view_schema_id
// receives view_schema_not_found, identical to a well-formed unknown one; an
// encoded slash stays inside its segment; undeclared methods and HEAD are
// enveloped 405s. Discovery is read-only, so no request commits an effect.
func TestDiscoveryRouteAdmissionOrder_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "discovery-admission-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	base := harness.Server.HTTP.URL
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	// A redirect is never followed, so a 3xx would surface as itself.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	get := func(t testing.TB, method string, path string, login *authcookietest.AuthCookies) *http.Response {
		t.Helper()
		var options []func(*http.Request)
		if login != nil {
			options = append(options, httptestx.WithCookies(login.Session))
		}
		return httptestx.DoJSON(t, client, method, base+path, nil, options...)
	}
	requireAlike := func(t testing.TB, malformed, missing *http.Response) {
		t.Helper()
		bodies := make([]map[string]any, 0, 2)
		for _, response := range []*http.Response{malformed, missing} {
			body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "view_schema_not_found")
			delete(body["error"].(map[string]any), "request_id")
			bodies = append(bodies, body)
		}
		if !reflect.DeepEqual(bodies[0], bodies[1]) {
			t.Fatalf("malformed-identifier response %v differs from unknown-schema response %v", bodies[0], bodies[1])
		}
	}
	requireMethodNotAllowed := func(t testing.TB, response *http.Response, allow string) {
		t.Helper()
		if got := response.Header.Get("Allow"); got != allow {
			t.Fatalf("Allow = %q, want %q", got, allow)
		}
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusMethodNotAllowed, "method_not_allowed")
		if details := httptestx.RequireErrorDetails(t, body); !reflect.DeepEqual(details, map[string]any{"allowed_methods": []any{allow}}) {
			t.Fatalf("details = %#v, want allowed_methods [%s]", details, allow)
		}
	}
	const known = "/api/v1/view-schemas/cartulary.view.timeline.v2"
	const unknown = "/api/v1/view-schemas/cartulary.view.not_real.v1"
	httptestx.RequireSuccessEnvelope(t, get(t, http.MethodGet, known, &login), http.StatusOK)

	t.Run("authentication precedes path parameters and queries", func(t *testing.T) {
		for _, path := range []string{"/api/v1/view-schemas/NOT_A_SCHEMA", known + "?limit=1", "/api/v1/view-schemas?limit=0", "/api/v1/extensions?limit=1", "/api/v1/extensions?cursor_token=opaque"} {
			httptestx.RequireErrorEnvelope(t, get(t, http.MethodGet, path, nil), http.StatusUnauthorized, "session_required")
		}
	})
	t.Run("a malformed view_schema_id is an unknown schema", func(t *testing.T) {
		for _, malformed := range []string{"NOT_A_SCHEMA", "cartulary.view.notes", "cartulary.view.notes.v0", "cartulary..view.v1", "%20cartulary.view.notes.v1", strings.Repeat("a", 300) + ".v1"} {
			requireAlike(t, get(t, http.MethodGet, "/api/v1/view-schemas/"+malformed, &login), get(t, http.MethodGet, unknown, &login))
		}
	})
	t.Run("an encoded slash stays inside view_schema_id", func(t *testing.T) {
		requireAlike(t, get(t, http.MethodGet, known+"%2Fx", &login), get(t, http.MethodGet, unknown, &login))
		requireAlike(t, get(t, http.MethodGet, "/api/v1/view-schemas/cartulary%2Fview.timeline.v2", &login), get(t, http.MethodGet, unknown, &login))
	})
	t.Run("route resolution rejections are enveloped", func(t *testing.T) {
		for _, path := range []string{"/api/v1/view-schemas", known, "/api/v1/extensions"} {
			head := get(t, http.MethodHead, path, &login)
			body, err := io.ReadAll(head.Body)
			_ = head.Body.Close()
			if err != nil || head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != http.MethodGet || len(body) != 0 {
				t.Fatalf("HEAD %s = %d Allow=%q body=%q err=%v, want 405 Allow GET without a body", path, head.StatusCode, head.Header.Get("Allow"), body, err)
			}
			requireMethodNotAllowed(t, get(t, http.MethodDelete, path, &login), http.MethodGet)
		}
		for _, path := range []string{known + "/fields", "/api/v1/extensions/import"} {
			httptestx.RequireErrorEnvelope(t, get(t, http.MethodGet, path, &login), http.StatusNotFound, "route_not_found")
		}
	})
	t.Run("an authenticated caller meets the query checks", func(t *testing.T) {
		for _, path := range []string{known + "?limit=1", unknown + "?limit=1", "/api/v1/view-schemas?limit=0", "/api/v1/extensions?limit=1", "/api/v1/extensions?cursor_token=opaque"} {
			httptestx.RequireErrorEnvelope(t, get(t, http.MethodGet, path, &login), http.StatusBadRequest, "invalid_pagination_request")
		}
	})
}
