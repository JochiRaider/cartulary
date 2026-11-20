package httpapi

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

const testRouteGuardToken = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"

func TestNewTestRouteGuardRequiresHarnessOwnedRuntime(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  map[string]string
	}{
		{
			name: "disabled",
			env:  map[string]string{},
		},
		{
			name: "missing marker",
			env: map[string]string{
				TestRoutesEnabledEnv: "1",
				TestRouteTokenEnv:    testRouteGuardToken,
			},
		},
		{
			name: "weak token",
			env: map[string]string{
				TestRoutesEnabledEnv: "1",
				TestRuntimeMarkerEnv: TestRuntimeMarkerValue,
				TestRouteTokenEnv:    "short",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewTestRouteGuard(tc.env); err == nil {
				t.Fatal("expected guard setup to fail")
			}
		})
	}
}

func TestTestRouteGuardAuthorization(t *testing.T) {
	guard, err := NewTestRouteGuard(map[string]string{
		TestRoutesEnabledEnv:       "1",
		TestRuntimeMarkerEnv:       TestRuntimeMarkerValue,
		TestRouteTokenEnv:          testRouteGuardToken,
		TestRuntimeAPIOriginEnv:    "http://127.0.0.1:8080",
		TestRuntimePublicOriginEnv: "http://127.0.0.1:4173",
	})
	if err != nil {
		t.Fatalf("new test route guard: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/v1/test/clock/set", nil)
	req.Host = "127.0.0.1:8080"
	req.Header.Set(TestRouteTokenHeader, testRouteGuardToken)
	req.Header.Set("Origin", "http://127.0.0.1:4173")
	if !guard.Authorized(req) {
		t.Fatal("expected matching token to authorize")
	}
	if !guard.AllowedRequestBoundary(req) {
		t.Fatal("expected matching host and origin to satisfy boundary")
	}

	wrongToken := req.Clone(req.Context())
	wrongToken.Header.Set(TestRouteTokenHeader, "ABCDEFGabcdefghijklmnopqrstuvwxyz0123456789")
	if guard.Authorized(wrongToken) {
		t.Fatal("wrong token authorized")
	}

	wrongHost := req.Clone(req.Context())
	wrongHost.Host = "evil.example.test"
	if guard.AllowedRequestBoundary(wrongHost) {
		t.Fatal("wrong host passed boundary")
	}

	missingOrigin := req.Clone(req.Context())
	missingOrigin.Header.Del("Origin")
	if guard.AllowedRequestBoundary(missingOrigin) {
		t.Fatal("missing origin passed boundary when origins are configured")
	}
}

func TestValidTestRouteTokenRejectsClosedWeakAttachTokens(t *testing.T) {
	if !ValidTestRouteToken(testRouteGuardToken) {
		t.Fatal("expected fixture token to be valid")
	}
	for _, token := range []string{
		"short",
		"token",
		"password",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"0123456789abcdefghijklmnopqrstuvwxyzABCDEF\n",
	} {
		if ValidTestRouteToken(token) {
			t.Fatalf("expected token %q to be invalid", token)
		}
	}
	// Exact length bounds: 43..512 printable, non-repeated bytes.
	varied := func(n int) string {
		return strings.Repeat(testRouteGuardToken, 512/len(testRouteGuardToken)+1)[:n]
	}
	for n, want := range map[int]bool{42: false, 43: true, 512: true, 513: false} {
		if got := ValidTestRouteToken(varied(n)); got != want {
			t.Fatalf("token of %d bytes valid = %v, want %v", n, got, want)
		}
	}
}

func TestTestRouteGuardProtectWritesForbiddenWithoutCORS(t *testing.T) {
	guard := TestRouteGuard{Token: testRouteGuardToken}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test/clock/set", nil)
	recorder := httptest.NewRecorder()

	guard.Protect(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler must not be called")
	})(recorder, req)

	resp := recorder.Result()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("test route guard must not emit permissive CORS, got %q", got)
	}
}

// TestTestRouteGuardProtectMethod covers the single-method form of Protect
// used by harness touch routes: the guard still decides first, then any other
// method, HEAD included, receives the enveloped 405 method_not_allowed with
// Allow (headers only for HEAD) and never reaches the handler.
func TestTestRouteGuardProtectMethod(t *testing.T) {
	guard := TestRouteGuard{Token: testRouteGuardToken}
	calls := 0
	protected := guard.ProtectMethod(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})
	serve := func(method string, token bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/v1/test/auth/touch", nil)
		if token {
			req.Header.Set(TestRouteTokenHeader, testRouteGuardToken)
		}
		recorder := httptest.NewRecorder()
		withRequestID(protected, nil).ServeHTTP(recorder, req)
		return recorder
	}
	if recorder := serve(http.MethodPost, false); recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), `"code":"test_route_forbidden"`) {
		t.Fatalf("missing token = %d %s, want the guard's 403 first", recorder.Code, recorder.Body)
	}
	requireMethodNotAllowedEnvelope(t, serve(http.MethodPost, true), "GET")
	if recorder := serve(http.MethodHead, true); recorder.Code != http.StatusMethodNotAllowed || recorder.Header().Get("Allow") != "GET" || recorder.Body.Len() != 0 {
		t.Fatalf("HEAD = %d Allow=%q body=%q", recorder.Code, recorder.Header().Get("Allow"), recorder.Body)
	}
	if calls != 0 {
		t.Fatal("a rejected method reached the handler")
	}
	if recorder := serve(http.MethodGet, true); recorder.Code != http.StatusOK || calls != 1 {
		t.Fatalf("GET = %d calls=%d", recorder.Code, calls)
	}

	// BindTestRoute binds a test-only harness route with the same order and
	// records it in the route registry (TH-HARNESS-REQ-450).
	t.Run("BindTestRoute guards, rejects other methods and records the route", func(t *testing.T) {
		registry, err := NewRouteRegistry(nil)
		if err != nil {
			t.Fatal(err)
		}
		deps := DependencySet{PublicRoutes: registry}
		mux := http.NewServeMux()
		bound := 0
		const path = "/api/v1/test/things/{thing_id}/arm"
		if err := BindTestRoute(mux, deps, guard, http.MethodPost, path, func(w http.ResponseWriter, r *http.Request) {
			bound++
			w.WriteHeader(http.StatusNoContent)
		}); err != nil {
			t.Fatal(err)
		}
		serveMux := func(method string, token bool) *httptest.ResponseRecorder {
			req := httptest.NewRequest(method, "/api/v1/test/things/x/arm", nil)
			if token {
				req.Header.Set(TestRouteTokenHeader, testRouteGuardToken)
			}
			recorder := httptest.NewRecorder()
			withRequestID(mux, nil).ServeHTTP(recorder, req)
			return recorder
		}
		if recorder := serveMux(http.MethodGet, false); recorder.Code != http.StatusForbidden {
			t.Fatalf("GET without token = %d, want the guard's 403 first", recorder.Code)
		}
		requireMethodNotAllowedEnvelope(t, serveMux(http.MethodGet, true), "POST")
		if recorder := serveMux(http.MethodPost, true); recorder.Code != http.StatusNoContent || bound != 1 {
			t.Fatalf("POST = %d calls=%d", recorder.Code, bound)
		}
		if got := registry.SupplementaryRoutes(); !reflect.DeepEqual(got, []SupplementaryRoute{{Kind: SupplementaryTestRoute, Method: http.MethodPost, Path: path}}) {
			t.Fatalf("recorded routes = %#v", got)
		}
		if err := BindTestRoute(http.NewServeMux(), deps, guard, http.MethodGet, path, func(http.ResponseWriter, *http.Request) {}); err == nil {
			t.Fatal("a second registration of the same path was accepted")
		}
		for _, invalid := range []struct{ method, path string }{
			{http.MethodPost, "/api/v1/things/x"},
			{http.MethodPost, "/ws/v1/incidents/{incident_id}"},
			{"", "/api/v1/test/empty-method"},
			{"POST /x", "/api/v1/test/spaced-method"},
		} {
			if err := BindTestRoute(http.NewServeMux(), deps, guard, invalid.method, invalid.path, func(http.ResponseWriter, *http.Request) {}); err == nil {
				t.Fatalf("BindTestRoute accepted %s %q", invalid.method, invalid.path)
			}
		}
	})
}
