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
	testDB := postgresHarness.PreparePackageDatabaseT(t, "httptestx")

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

	server := StartServer(t, ServerOptions{Env: env})

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
