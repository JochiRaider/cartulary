package harnesscontrol

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/httpapiextensions"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

const testRuntimeResetToken = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"

func testRuntimeEnabledEnv() map[string]string {
	return map[string]string{
		httpapi.TestRoutesEnabledEnv: "1",
		httpapi.TestRuntimeMarkerEnv: httpapi.TestRuntimeMarkerValue,
		httpapi.TestRouteTokenEnv:    testRuntimeResetToken,
	}
}

func startTestRuntimeResetServerWithHTTPDeps(t testing.TB, env map[string]string, routes []httpapi.RouteRegistrar, deps httpapi.DependencySet) *httptest.Server {
	t.Helper()
	effectiveEnv := make(map[string]string, len(env)+8)
	for key, value := range env {
		effectiveEnv[key] = value
	}
	tempRoots := configtest.SetupTempRoots(t)
	for key, value := range tempRoots.Paths {
		if _, exists := effectiveEnv[key]; !exists {
			effectiveEnv[key] = value
		}
	}
	if _, exists := effectiveEnv[httpapi.TestRoutesEnabledEnv]; !exists {
		effectiveEnv[httpapi.TestRoutesEnabledEnv] = "1"
	}
	if _, exists := effectiveEnv[httpapi.TestRuntimeMarkerEnv]; !exists {
		effectiveEnv[httpapi.TestRuntimeMarkerEnv] = httpapi.TestRuntimeMarkerValue
	}
	if _, exists := effectiveEnv[httpapi.TestRouteTokenEnv]; !exists {
		effectiveEnv[httpapi.TestRouteTokenEnv] = testRuntimeResetToken
	}
	configtest.BindPostgresEnvToDatabaseRoot(t, tempRoots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], effectiveEnv, postgres.PurposeRuntime)

	cfg := configtest.LoadFixture(t, []string{"config", "valid.toml"}, effectiveEnv).Deployment()
	ctx := context.Background()
	pool, err := appsupport.OpenPostgres(ctx, cfg, effectiveEnv)
	if err != nil {
		t.Fatalf("open Network Flow reset postgres fixture: %v", err)
	}
	store, err := appsupport.OpenObjectStore(ctx, cfg, effectiveEnv)
	if err != nil {
		pool.Close()
		t.Fatalf("open Network Flow reset object-store fixture: %v", err)
	}
	bootstrapSettings := bootstrap.Settings{ManifestPath: cfg.Bootstrap.FirstAdminManifestPath}
	if err := bootstrap.Preflight(ctx, bootstrapSettings, pool); err != nil {
		pool.Close()
		_ = store.Close()
		t.Fatalf("bootstrap Network Flow reset fixture: %v", err)
	}
	deps.TestResetBootstrap = func(ctx context.Context, tx pgx.Tx) error {
		return bootstrap.PreflightTx(ctx, bootstrapSettings, tx)
	}
	recoveryPool, err := postgres.Setup(ctx, postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          effectiveEnv[suiteservices.PostgresDSNEnv],
		Purpose:      postgres.PurposeRecovery,
		ExpectedRole: "cartulary_recovery",
	})
	if err != nil {
		pool.Close()
		_ = store.Close()
		t.Fatalf("open Network Flow reset Recovery fixture: %v", err)
	}
	deps.TestResetDatabase = func(ctx context.Context) error {
		return harnessruntime.ResetDatabase(ctx, recoveryPool, deps.TestResetBootstrap)
	}
	deps.Env = effectiveEnv
	deps.Postgres = pool
	deps.ObjectStore = store
	handler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies:     testHTTPDependencies(deps),
		AdditionalRoutes: routes,
	})
	if err != nil {
		pool.Close()
		_ = store.Close()
		t.Fatalf("start Network Flow harness-control reset runtime: %v", err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
		recoveryPool.Close()
		pool.Close()
		_ = store.Close()
	})
	return server
}

func testHTTPDependencies(deps httpapi.DependencySet) httpapi.DependencySet {
	return httpapiextensions.New(nil).Dependencies(deps)
}

func newTestRuntimeResetJSONRequest(t testing.TB, method string, url string, body any) *http.Request {
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

func authorizeTestRuntimeResetRequest(req *http.Request) *http.Request {
	req.Header.Set(httpapi.TestRouteTokenHeader, testRuntimeResetToken)
	return req
}

func doTestRuntimeResetRequest(t testing.TB, client *http.Client, req *http.Request) *http.Response {
	t.Helper()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func requireTestRuntimeResetStatus(t testing.TB, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, want)
	}
}

func readTestRuntimeResetJSONBody(t testing.TB, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return body
}

func requireTestRuntimeResetSuccessEnvelope(t testing.TB, resp *http.Response, want int) map[string]any {
	t.Helper()
	requireTestRuntimeResetStatus(t, resp, want)
	if resp.Header.Get(httpapi.RequestIDHeader) == "" {
		t.Fatal("missing request id header")
	}
	body := readTestRuntimeResetJSONBody(t, resp)
	if _, ok := body["data"].(map[string]any); !ok {
		t.Fatalf("expected success envelope data object, got %T", body["data"])
	}
	return body
}

func requireTestRuntimeResetErrorEnvelope(t testing.TB, resp *http.Response, want int, wantCode string) map[string]any {
	t.Helper()
	requireTestRuntimeResetStatus(t, resp, want)
	body := readTestRuntimeResetJSONBody(t, resp)
	errorValue, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error envelope, got %#v", body)
	}
	if errorValue["code"] != wantCode {
		t.Fatalf("unexpected error code: got %#v want %q", errorValue["code"], wantCode)
	}
	return body
}

func prepareTestRuntimeResetBucket(t testing.TB, h *s3test.Harness, prefix string) string {
	t.Helper()
	return h.BootstrapBucketT(t, prefix)
}
