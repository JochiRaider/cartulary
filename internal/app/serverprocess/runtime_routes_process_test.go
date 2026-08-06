package serverprocess

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestHarnessRuntimeRoutesEnableOnlyForExactOneInServerProcess(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "omitted"},
		{name: "non-exact-true", value: "true"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			routeEnv := map[string]string{
				"CARTULARY_TEST_RUNTIME_MARKER": "harness-owned",
				"CARTULARY_TEST_ROUTE_TOKEN":    httptestx.TestRouteToken,
			}
			if tc.value != "" {
				routeEnv["CARTULARY_ENABLE_TEST_ROUTES"] = tc.value
			}
			server := startHarnessRuntimeServerProcess(t, "test-runtime-enable-value-"+tc.name, routeEnv)
			requireHarnessRuntimeStatus(t, server, http.MethodGet, "/api/v1/test/runtime/identity", nil, "", "", "", http.StatusNotFound)
		})
	}

	t.Run("exact-one", func(t *testing.T) {
		addr := reserveHarnessRuntimeProcessAddress(t)
		publicOrigin := "http://127.0.0.1:4173"
		server := startHarnessRuntimeServerProcess(t, "test-runtime-enable-value-exact-one", map[string]string{
			"CARTULARY_HTTP_ADDR":             addr,
			"CARTULARY_ENABLE_TEST_ROUTES":    "1",
			"CARTULARY_TEST_RUNTIME_MARKER":   "harness-owned",
			"CARTULARY_TEST_ROUTE_TOKEN":      httptestx.TestRouteToken,
			"CARTULARY_WEB_E2E_API_ORIGIN":    "http://" + addr,
			"CARTULARY_WEB_E2E_PUBLIC_ORIGIN": publicOrigin,
		})
		requireHarnessRuntimeStatus(t, server, http.MethodGet, "/api/v1/test/runtime/identity", nil, httptestx.TestRouteToken, publicOrigin, "", http.StatusOK)
	})
}

func TestHarnessRuntimeRoutesFailClosedDuringServerStartup(t *testing.T) {
	for _, tc := range []struct {
		name     string
		routeEnv map[string]string
	}{
		{
			name: "missing-marker",
			routeEnv: map[string]string{
				"CARTULARY_ENABLE_TEST_ROUTES": "1",
				"CARTULARY_TEST_ROUTE_TOKEN":   httptestx.TestRouteToken,
			},
		},
		{
			name: "missing-token",
			routeEnv: map[string]string{
				"CARTULARY_ENABLE_TEST_ROUTES":  "1",
				"CARTULARY_TEST_RUNTIME_MARKER": "harness-owned",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := harnessRuntimeServerProcessEnv(t, "test-runtime-fail-closed-"+tc.name)
			for key, value := range tc.routeEnv {
				env[key] = value
			}

			server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
			if err := server.WaitForExit(t); err == nil {
				t.Fatal("expected cmd/server to exit before readiness when test-route guard setup is incomplete")
			}
		})
	}
}

func TestHarnessRuntimeRoutesPreserveServerProcessSecurityAndReset(t *testing.T) {
	addr := reserveHarnessRuntimeProcessAddress(t)
	apiOrigin := "http://" + addr
	publicOrigin := "http://127.0.0.1:4173"
	server := startHarnessRuntimeServerProcess(t, "test-runtime-security", map[string]string{
		"CARTULARY_HTTP_ADDR":             addr,
		"CARTULARY_ENABLE_TEST_ROUTES":    "1",
		"CARTULARY_TEST_RUNTIME_MARKER":   "harness-owned",
		"CARTULARY_TEST_ROUTE_TOKEN":      httptestx.TestRouteToken,
		"CARTULARY_WEB_E2E_API_ORIGIN":    apiOrigin,
		"CARTULARY_WEB_E2E_PUBLIC_ORIGIN": publicOrigin,
	})

	missingToken := doHarnessRuntimeJSON(t, server, http.MethodGet, "/api/v1/test/runtime/identity", nil, "", publicOrigin, "")
	requireHarnessRuntimeForbiddenWithoutCORS(t, missingToken)

	wrongToken := doHarnessRuntimeJSON(t, server, http.MethodGet, "/api/v1/test/runtime/identity", nil, "ABCDEFGabcdefghijklmnopqrstuvwxyz0123456789", publicOrigin, "")
	requireHarnessRuntimeForbiddenWithoutCORS(t, wrongToken)

	wrongOrigin := doHarnessRuntimeJSON(t, server, http.MethodGet, "/api/v1/test/runtime/identity", nil, httptestx.TestRouteToken, "http://evil.example.test", "")
	requireHarnessRuntimeForbiddenWithoutCORS(t, wrongOrigin)

	wrongHost := doHarnessRuntimeJSON(t, server, http.MethodGet, "/api/v1/test/runtime/identity", nil, httptestx.TestRouteToken, publicOrigin, "evil.example.test")
	requireHarnessRuntimeForbiddenWithoutCORS(t, wrongHost)

	identity := doHarnessRuntimeJSON(t, server, http.MethodGet, "/api/v1/test/runtime/identity", nil, httptestx.TestRouteToken, publicOrigin, "")
	httptestx.RequireSuccessEnvelope(t, identity, http.StatusOK)

	setClock := doHarnessRuntimeJSON(t, server, http.MethodPost, "/api/v1/test/clock/set", map[string]any{
		"fixed_now": "2035-01-01T00:00:00Z",
	}, httptestx.TestRouteToken, publicOrigin, "")
	httptestx.RequireSuccessEnvelope(t, setClock, http.StatusOK)

	reset := doHarnessRuntimeJSON(t, server, http.MethodPost, "/api/v1/test/runtime/reset", nil, httptestx.TestRouteToken, publicOrigin, "")
	httptestx.RequireSuccessEnvelope(t, reset, http.StatusOK)

	clockState := doHarnessRuntimeJSON(t, server, http.MethodGet, "/api/v1/test/clock/state", nil, httptestx.TestRouteToken, publicOrigin, "")
	stateBody := httptestx.RequireSuccessEnvelope(t, clockState, http.StatusOK)
	state := stateBody["data"].(map[string]any)
	if state["mode"] != "wall" {
		t.Fatalf("runtime reset must restore wall clock mode, got %#v", state)
	}
	if state["offset_seconds"] != float64(0) {
		t.Fatalf("runtime reset must clear clock offset, got %#v", state)
	}
	if _, ok := state["fixed_now"]; ok {
		t.Fatalf("runtime reset must clear fixed clock, got %#v", state)
	}
}

func startHarnessRuntimeServerProcess(t testing.TB, prefix string, routeEnv map[string]string) *processtest.Server {
	t.Helper()

	env := harnessRuntimeServerProcessEnv(t, prefix)
	for key, value := range routeEnv {
		env[key] = value
	}

	server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
	t.Cleanup(func() {
		server.Stop(t)
	})
	server.WaitForReady(t)
	return server
}

func harnessRuntimeServerProcessEnv(t testing.TB, prefix string) map[string]string {
	t.Helper()

	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
	bucket := BucketName(prefix)
	t.Cleanup(func() {
		cleanupBucket(t, s3Harness, bucket)
	})

	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	return ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))
}

func reserveHarnessRuntimeProcessAddress(t testing.TB) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve process address: %v", err)
	}
	defer listener.Close()
	return listener.Addr().String()
}

func requireHarnessRuntimeStatus(t testing.TB, server *processtest.Server, method string, path string, body any, token string, origin string, host string, want int) {
	t.Helper()

	resp := doHarnessRuntimeJSON(t, server, method, path, body, token, origin, host)
	defer resp.Body.Close()
	if resp.StatusCode != want {
		t.Fatalf("unexpected status for %s %s: got %d want %d", method, path, resp.StatusCode, want)
	}
}

func requireHarnessRuntimeForbiddenWithoutCORS(t testing.TB, resp *http.Response) {
	t.Helper()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("test route guard must not emit permissive CORS, got %q", got)
	}
	httptestx.RequireErrorEnvelope(t, resp, http.StatusForbidden, "test_route_forbidden")
}

func doHarnessRuntimeJSON(t testing.TB, server *processtest.Server, method string, path string, body any, token string, origin string, host string) *http.Response {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequest(method, server.BaseURL+path, reader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Cartulary-Test-Route-Token", token)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if host != "" {
		req.Host = host
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do %s %s: %v", method, path, err)
	}
	return resp
}
