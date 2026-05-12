package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const testRouteGuardToken = "0123456789abcdef0123456789abcdef"

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
	wrongToken.Header.Set(TestRouteTokenHeader, "wrong-token-wrong-token-wrong-token")
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
