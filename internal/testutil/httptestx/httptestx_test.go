package httptestx

import (
	"context"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestHarnessBootsServerAndAssertsEnvelopes(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "httptestx")

	s3Harness := s3test.Start(t)
	bucket, err := s3Harness.BootstrapBucket(context.Background(), "httptestx")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}

	server := StartServer(t, ServerOptions{Env: env, TestRouteMode: TestRouteModeHarnessOwned})

	successResp := Do(t, server.HTTP.Client(), NewJSONRequest(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/success", nil))
	successBody := RequireSuccessEnvelope(t, successResp, http.StatusOK)
	data := successBody["data"].(map[string]any)
	if data["service"] != "bootstrap" || data["status"] != "ok" {
		t.Fatalf("unexpected success payload: %#v", data)
	}

	errorResp := Do(t, server.HTTP.Client(), NewJSONRequest(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/error", nil))
	errorBody := RequireErrorEnvelope(t, errorResp, http.StatusServiceUnavailable, "bootstrap_error")
	errorDetails := errorBody["error"].(map[string]any)["details"].(map[string]any)
	if errorDetails["reason_code"] != "bootstrap_unavailable" {
		t.Fatalf("unexpected error details: %#v", errorDetails)
	}
}

func TestStartServerHonorsTestRouteMode(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "httptestx-test-route-mode")

	s3Harness := s3test.Start(t)
	bucket, err := s3Harness.BootstrapBucket(context.Background(), "httptestx-test-route-mode")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}

	t.Run("harness owned", func(t *testing.T) {
		server := StartServer(t, ServerOptions{Env: env, TestRouteMode: TestRouteModeHarnessOwned})
		req := NewJSONRequest(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/clock/state", nil)
		req.Header.Set("X-Cartulary-Test-Route-Token", TestRouteToken)

		resp := Do(t, server.HTTP.Client(), req)
		RequireSuccessEnvelope(t, resp, http.StatusOK)
	})

	t.Run("disabled", func(t *testing.T) {
		server := StartServer(t, ServerOptions{
			Env:           env,
			TestRouteMode: TestRouteModeDisabled,
		})
		req := NewJSONRequest(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/clock/state", nil)
		req.Header.Set("X-Cartulary-Test-Route-Token", TestRouteToken)

		resp := Do(t, server.HTTP.Client(), req)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("disabled test route mode must leave clock routes unavailable: got %d want %d", resp.StatusCode, http.StatusNotFound)
		}
		_ = resp.Body.Close()
	})

	t.Run("custom env", func(t *testing.T) {
		customEnv := make(map[string]string, len(env)+3)
		for key, value := range env {
			customEnv[key] = value
		}
		customEnv["CARTULARY_ENABLE_TEST_ROUTES"] = "1"
		customEnv["CARTULARY_TEST_RUNTIME_MARKER"] = "harness-owned"
		customEnv["CARTULARY_TEST_ROUTE_TOKEN"] = TestRouteToken

		server := StartServer(t, ServerOptions{
			Env:           customEnv,
			TestRouteMode: TestRouteModeCustomEnv,
		})
		req := NewJSONRequest(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/clock/state", nil)
		req.Header.Set("X-Cartulary-Test-Route-Token", TestRouteToken)

		resp := Do(t, server.HTTP.Client(), req)
		RequireSuccessEnvelope(t, resp, http.StatusOK)
	})
}

func TestApplyTestRouteModeRejectsMissingAndUnknownModes(t *testing.T) {
	for _, mode := range []TestRouteMode{"", "unknown"} {
		env := map[string]string{}
		if err := applyTestRouteMode(env, mode); err == nil {
			t.Fatalf("expected mode %q to fail", mode)
		}
		if len(env) != 0 {
			t.Fatalf("invalid mode %q mutated env: %#v", mode, env)
		}
	}
}

func TestRequireAuthCookies(t *testing.T) {
	authCookies := RequireAuthCookies(t, []*http.Cookie{
		{
			Name:     authn.SessionCookieName,
			Value:    "session-token",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		},
		{
			Name:     authn.CSRFCookieName,
			Value:    "csrf-token",
			Path:     "/",
			HttpOnly: false,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		},
	})

	if authCookies.Session == nil || authCookies.Session.Value != "session-token" {
		t.Fatalf("unexpected session cookie: %#v", authCookies.Session)
	}
	if authCookies.CSRF == nil || authCookies.CSRF.Value != "csrf-token" {
		t.Fatalf("unexpected csrf cookie: %#v", authCookies.CSRF)
	}
}
