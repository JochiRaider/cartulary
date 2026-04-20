package httptestx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	stdhttptest "net/http/httptest"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

type Server struct {
	Runtime *app.Runtime
	HTTP    *stdhttptest.Server
	Clock   *httpapi.TestClock
}

type ServerOptions struct {
	Config           config.Config
	Env              map[string]string
	AdditionalRoutes []httpapi.RouteRegistrar
}

func StartServer(t testing.TB, options ServerOptions) *Server {
	t.Helper()

	env := make(map[string]string, len(options.Env))
	for key, value := range options.Env {
		env[key] = value
	}

	cfg := options.Config
	if cfg.ConfigSchemaID == "" {
		tempRoots := configtest.SetupTempRoots(t)
		for key, value := range tempRoots.Paths {
			if _, exists := env[key]; !exists {
				env[key] = value
			}
		}
		if _, exists := env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"]; !exists {
			env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")
		}
		cfg = configtest.LoadEffectiveFixture(t, []string{"config", "valid.toml"}, env)
	}

	clock := httpapi.NewTestClock()
	routes := append([]httpapi.RouteRegistrar{RegisterBootstrapRoutes(), httpapi.RegisterTestClockRoutes(clock)}, options.AdditionalRoutes...)
	runtime, err := app.NewRuntime(context.Background(), cfg, app.Options{
		Env: env,
		Now: clock.Now,
		HTTP: httpapi.Options{
			AdditionalRoutes: routes,
		},
	})
	if err != nil {
		t.Fatalf("start app runtime: %v", err)
	}

	server := &Server{
		Runtime: runtime,
		HTTP:    stdhttptest.NewServer(runtime.Handler),
		Clock:   clock,
	}
	t.Cleanup(func() {
		server.Close()
	})

	return server
}

func (s *Server) Close() {
	if s == nil {
		return
	}
	if s.HTTP != nil {
		s.HTTP.Close()
	}
	if s.Runtime != nil {
		s.Runtime.Close()
	}
}

func NewJSONRequest(t testing.TB, method string, url string, body any) *http.Request {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func NewAuthenticatedJSONRequest(t testing.TB, method string, url string, token string, body any) *http.Request {
	t.Helper()

	req := NewJSONRequest(t, method, url, body)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func Do(t testing.TB, client *http.Client, req *http.Request) *http.Response {
	t.Helper()

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func ReadJSONBody(t testing.TB, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()

	var payload map[string]any
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return payload
}

func RequireStatus(t testing.TB, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, want)
	}
}

func RequireRequestID(t testing.TB, resp *http.Response) string {
	t.Helper()
	requestID := resp.Header.Get(httpapi.RequestIDHeader)
	if requestID == "" {
		t.Fatal("missing request id header")
	}
	return requestID
}

func RequireSuccessEnvelope(t testing.TB, resp *http.Response, wantStatus int) map[string]any {
	t.Helper()
	RequireStatus(t, resp, wantStatus)
	requestID := RequireRequestID(t, resp)
	body := ReadJSONBody(t, resp)

	metaValue, ok := body["meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected success envelope meta object, got %T", body["meta"])
	}
	if metaValue["request_id"] != requestID {
		t.Fatalf("body meta.request_id mismatch: got %v want %s", metaValue["request_id"], requestID)
	}
	if _, ok := body["data"].(map[string]any); !ok {
		t.Fatalf("expected success envelope data object, got %T", body["data"])
	}
	return body
}

func RequireErrorEnvelope(t testing.TB, resp *http.Response, wantStatus int, wantCode string) map[string]any {
	t.Helper()
	RequireStatus(t, resp, wantStatus)
	requestID := RequireRequestID(t, resp)
	body := ReadJSONBody(t, resp)

	errorValue, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %T", body["error"])
	}
	if errorValue["request_id"] != requestID {
		t.Fatalf("error.request_id mismatch: got %v want %s", errorValue["request_id"], requestID)
	}
	if errorValue["code"] != wantCode {
		t.Fatalf("unexpected error code: got %v want %s", errorValue["code"], wantCode)
	}
	if _, ok := errorValue["retryable"].(bool); !ok {
		t.Fatalf("expected error.retryable boolean, got %T", errorValue["retryable"])
	}
	if _, ok := errorValue["details"].(map[string]any); !ok {
		t.Fatalf("expected error details object, got %T", errorValue["details"])
	}
	return body
}
