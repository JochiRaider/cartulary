package harnessruntime

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httpapiextensions"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

const testRuntimeResetToken = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"

func TestTestRuntimeResetRouteDisabledByDefault(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "test-runtime-reset-disabled")
	s3Harness := s3test.Start(t)
	bucket := prepareTestRuntimeResetBucket(t, s3Harness, "test-runtime-reset-disabled")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	_, server := startTestRuntimeResetServer(t, env, nil)
	resp := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil))
	defer resp.Body.Close()
	requireTestRuntimeResetStatus(t, resp, http.StatusNotFound)

	identityResp := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+"/api/v1/test/runtime/identity", nil))
	defer identityResp.Body.Close()
	requireTestRuntimeResetStatus(t, identityResp, http.StatusNotFound)
}

func TestTestRuntimeIdentityRouteRequiresHarnessAuthorization(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "test-runtime-identity")
	s3Harness := s3test.Start(t)
	bucket := prepareTestRuntimeResetBucket(t, s3Harness, "test-runtime-identity")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	_, server := startTestRuntimeResetServer(t, env, []httpapi.RouteRegistrar{RegisterTestRuntimeResetRoute()})

	missing := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+"/api/v1/test/runtime/identity", nil))
	requireTestRuntimeResetErrorEnvelope(t, missing, http.StatusForbidden, "test_route_forbidden")

	wrong := newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+"/api/v1/test/runtime/identity", nil)
	wrong.Header.Set(testRouteTokenHeader, "ABCDEFGabcdefghijklmnopqrstuvwxyz0123456789")
	requireTestRuntimeResetErrorEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), wrong), http.StatusForbidden, "test_route_forbidden")

	success := doTestRuntimeResetRequest(t, server.Client(), authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+"/api/v1/test/runtime/identity", nil)))
	body := requireTestRuntimeResetSuccessEnvelope(t, success, http.StatusOK)
	data := body["data"].(map[string]any)
	if data["schema_id"] != testRuntimeIdentitySchemaID {
		t.Fatalf("unexpected identity schema id: %#v", data["schema_id"])
	}
	if data["runtime_marker"] != testRuntimeMarkerValue {
		t.Fatalf("unexpected runtime marker: %#v", data["runtime_marker"])
	}
	if data["test_routes_enabled"] != true {
		t.Fatalf("identity route must report test routes enabled: %#v", data)
	}
	if _, ok := data["server_pid"].(float64); !ok {
		t.Fatalf("identity route must report server_pid: %#v", data)
	}
	requireNoPermissiveTestRuntimeCORS(t, success)
}

func TestTestRuntimeRoutesRejectNonHarnessOriginAndHost(t *testing.T) {
	service := &testRuntimeResetService{
		guard: httpapi.TestRouteGuard{
			Token:        testRuntimeResetToken,
			ExpectedHost: "127.0.0.1:8080",
			AllowedOrigins: map[string]struct{}{
				"http://127.0.0.1:8080": {},
				"http://127.0.0.1:4173": {},
			},
		},
	}

	wrongHost := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, "http://evil.example.test/api/v1/test/runtime/identity", nil))
	wrongHost.Host = "evil.example.test"
	wrongHostRecorder := httptest.NewRecorder()
	service.handleIdentity(wrongHostRecorder, wrongHost)
	requireTestRuntimeResetErrorEnvelope(t, wrongHostRecorder.Result(), http.StatusForbidden, "test_route_forbidden")

	missingOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, "http://127.0.0.1:8080/api/v1/test/runtime/identity", nil))
	missingOriginRecorder := httptest.NewRecorder()
	service.handleIdentity(missingOriginRecorder, missingOrigin)
	requireTestRuntimeResetErrorEnvelope(t, missingOriginRecorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, "http://127.0.0.1:8080/api/v1/test/runtime/identity", nil))
	wrongOrigin.Header.Set("Origin", "http://evil.example.test")
	wrongOriginRecorder := httptest.NewRecorder()
	service.handleIdentity(wrongOriginRecorder, wrongOrigin)
	requireTestRuntimeResetErrorEnvelope(t, wrongOriginRecorder.Result(), http.StatusForbidden, "test_route_forbidden")

	malformedOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, "http://127.0.0.1:8080/api/v1/test/runtime/identity", nil))
	malformedOrigin.Header.Set("Origin", "http://127.0.0.1:4173/path")
	malformedOriginRecorder := httptest.NewRecorder()
	service.handleIdentity(malformedOriginRecorder, malformedOrigin)
	requireTestRuntimeResetErrorEnvelope(t, malformedOriginRecorder.Result(), http.StatusForbidden, "test_route_forbidden")

	service.resetMu.Lock()
	missingResetOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/reset", nil))
	missingResetOriginRecorder := httptest.NewRecorder()
	service.handleReset(missingResetOriginRecorder, missingResetOrigin)
	requireTestRuntimeResetErrorEnvelope(t, missingResetOriginRecorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongResetOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/reset", nil))
	wrongResetOrigin.Header.Set("Origin", "http://evil.example.test")
	wrongResetOriginRecorder := httptest.NewRecorder()
	service.handleReset(wrongResetOriginRecorder, wrongResetOrigin)
	requireTestRuntimeResetErrorEnvelope(t, wrongResetOriginRecorder.Result(), http.StatusForbidden, "test_route_forbidden")

	allowedResetOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/reset", nil))
	allowedResetOrigin.Header.Set("Origin", "http://127.0.0.1:8080")
	allowedResetOriginRecorder := httptest.NewRecorder()
	service.handleReset(allowedResetOriginRecorder, allowedResetOrigin)
	service.resetMu.Unlock()
	requireTestRuntimeResetErrorEnvelope(t, allowedResetOriginRecorder.Result(), http.StatusConflict, "test_runtime_reset_in_progress")

	allowedAPIOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, "http://127.0.0.1:8080/api/v1/test/runtime/identity", nil))
	allowedAPIOrigin.Header.Set("Origin", "http://127.0.0.1:8080")
	allowedAPIOriginRecorder := httptest.NewRecorder()
	service.handleIdentity(allowedAPIOriginRecorder, allowedAPIOrigin)
	allowedAPIOriginResp := allowedAPIOriginRecorder.Result()
	requireTestRuntimeResetStatus(t, allowedAPIOriginResp, http.StatusOK)
	requireNoPermissiveTestRuntimeCORS(t, allowedAPIOriginResp)
	_ = readTestRuntimeResetJSONBody(t, allowedAPIOriginResp)

	allowedOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, "http://127.0.0.1:8080/api/v1/test/runtime/identity", nil))
	allowedOrigin.Header.Set("Origin", "http://127.0.0.1:4173")
	allowedOriginRecorder := httptest.NewRecorder()
	service.handleIdentity(allowedOriginRecorder, allowedOrigin)
	allowedOriginResp := allowedOriginRecorder.Result()
	requireTestRuntimeResetStatus(t, allowedOriginResp, http.StatusOK)
	body := readTestRuntimeResetJSONBody(t, allowedOriginResp)
	if data, ok := body["data"].(map[string]any); !ok || data["schema_id"] != testRuntimeIdentitySchemaID {
		t.Fatalf("expected identity success envelope, got %#v", body)
	}
	requireNoPermissiveTestRuntimeCORS(t, allowedOriginResp)
}

func TestTestRuntimeResetRouteRejectsInvalidHarnessPredicates(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "test-runtime-invalid-predicates")
	s3Harness := s3test.Start(t)
	bucket := prepareTestRuntimeResetBucket(t, s3Harness, "test-runtime-invalid-predicates")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	wrongMarker := cloneTestRuntimeResetEnv(env)
	wrongMarker[testRoutesEnabledEnv] = "1"
	wrongMarker[testRouteTokenEnv] = testRuntimeResetToken
	wrongMarker[testRuntimeMarkerEnv] = "not-owned"
	requireTestRuntimeResetStartupError(t, wrongMarker, "must be \"harness-owned\"")

	weakToken := cloneTestRuntimeResetEnv(env)
	weakToken[testRoutesEnabledEnv] = "1"
	weakToken[testRuntimeMarkerEnv] = testRuntimeMarkerValue
	weakToken[testRouteTokenEnv] = "short"
	requireTestRuntimeResetStartupError(t, weakToken, "must be a visible ASCII token")
}

func TestTestRuntimeResetRouteClearsStateAndRestoresBootstrap(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "test-runtime-reset")
	s3Harness := s3test.Start(t)
	bucket := prepareTestRuntimeResetBucket(t, s3Harness, "test-runtime-reset")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	runtime, server := startTestRuntimeResetServer(t, env, []httpapi.RouteRegistrar{RegisterTestRuntimeResetRoute()})
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open reset test database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	beforeGooseVersions := requireSQLCount(t, db, `SELECT COUNT(*) FROM goose_db_version`)
	seedTestRuntimeResetRows(t, db)
	if err := runtime.ObjectStore.PutObject(context.Background(), "reset/proof.txt", bytes.NewReader([]byte("proof")), int64(len("proof")), "text/plain"); err != nil {
		t.Fatalf("put reset proof object: %v", err)
	}

	requireTestRuntimeResetForbiddenOrInvalidRequests(t, server)

	resp := doTestRuntimeResetRequest(t, server.Client(), authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil)))
	body := requireTestRuntimeResetSuccessEnvelope(t, resp, http.StatusOK)
	data := body["data"].(map[string]any)
	if data["schema_id"] != testRuntimeResetSchemaID {
		t.Fatalf("unexpected schema id: %#v", data["schema_id"])
	}
	if data["migration_metadata_preserved"] != true {
		t.Fatalf("expected migration metadata preservation evidence, got %#v", data)
	}
	if data["bootstrap_admin_restored"] != true {
		t.Fatalf("expected bootstrap restoration evidence, got %#v", data)
	}
	if data["object_count_removed"] != float64(1) || data["object_count_after"] != float64(0) {
		t.Fatalf("unexpected object reset evidence: %#v", data)
	}
	if data["partial_failure"] != false {
		t.Fatalf("reset success must not report partial failure: %#v", data)
	}

	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM goose_db_version`, beforeGooseVersions)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 1)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM incidents`, 0)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM records`, 0)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM user_sessions`, 0)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM route_idempotency`, 0)

}

func TestTestRuntimeResetRouteClearsRegisteredTestClock(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "test-runtime-reset-clock")
	s3Harness := s3test.Start(t)
	bucket := prepareTestRuntimeResetBucket(t, s3Harness, "test-runtime-reset-clock")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	_, server := startTestRuntimeResetServer(t, env, []httpapi.RouteRegistrar{RegisterTestRuntimeResetRoute()})

	fixed := "2035-01-01T00:00:00Z"
	setClock := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/clock/set", map[string]any{
		"fixed_now": fixed,
	}))
	setResponse := requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), setClock), http.StatusOK)
	if data := setResponse["data"].(map[string]any); data["mode"] != "fixed" || data["now"] != fixed {
		t.Fatalf("unexpected fixed clock response: %#v", data)
	}

	reset := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), reset), http.StatusOK)

	state := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+"/api/v1/test/clock/state", nil))
	stateResponse := requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), state), http.StatusOK)
	data := stateResponse["data"].(map[string]any)
	if data["mode"] != "wall" {
		t.Fatalf("runtime reset must restore wall clock mode, got %#v", data)
	}
	if data["offset_seconds"] != float64(0) {
		t.Fatalf("runtime reset must clear clock offset, got %#v", data)
	}
	if _, ok := data["fixed_now"]; ok {
		t.Fatalf("runtime reset must clear fixed clock, got %#v", data)
	}
}

func TestTestRuntimeResetRouteRejectsConcurrentReset(t *testing.T) {
	service := &testRuntimeResetService{
		guard: httpapi.TestRouteGuard{Token: testRuntimeResetToken},
	}
	service.resetMu.Lock()
	defer service.resetMu.Unlock()

	req := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "/api/v1/test/runtime/reset", nil))
	recorder := httptest.NewRecorder()
	service.handleReset(recorder, req)

	resp := recorder.Result()
	requireTestRuntimeResetErrorEnvelope(t, resp, http.StatusConflict, "test_runtime_reset_in_progress")
}

func requireTestRuntimeResetForbiddenOrInvalidRequests(t testing.TB, server *httptest.Server) {
	t.Helper()
	missing := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil))
	requireTestRuntimeResetErrorEnvelope(t, missing, http.StatusForbidden, "test_route_forbidden")

	wrong := newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil)
	wrong.Header.Set(testRouteTokenHeader, "ABCDEFGabcdefghijklmnopqrstuvwxyz0123456789")
	requireTestRuntimeResetErrorEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), wrong), http.StatusForbidden, "test_route_forbidden")

	invalidBody := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", map[string]any{
		"unexpected": true,
	}))
	requireTestRuntimeResetErrorEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), invalidBody), http.StatusBadRequest, "invalid_test_reset_request")
}

type testRuntimeResetRuntime struct {
	ObjectStore objectstore.Store
}

func startTestRuntimeResetServer(t testing.TB, env map[string]string, routes []httpapi.RouteRegistrar) (*testRuntimeResetRuntime, *httptest.Server) {
	t.Helper()
	return startTestRuntimeResetServerWithHTTPDeps(t, env, routes, httpapi.DependencySet{})
}

func startTestRuntimeResetServerWithHTTPDeps(t testing.TB, env map[string]string, routes []httpapi.RouteRegistrar, deps httpapi.DependencySet) (*testRuntimeResetRuntime, *httptest.Server) {
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
	if _, exists := effectiveEnv[testRoutesEnabledEnv]; !exists {
		effectiveEnv[testRoutesEnabledEnv] = "1"
	}
	if _, exists := effectiveEnv[testRuntimeMarkerEnv]; !exists {
		effectiveEnv[testRuntimeMarkerEnv] = testRuntimeMarkerValue
	}
	if _, exists := effectiveEnv[testRouteTokenEnv]; !exists {
		effectiveEnv[testRouteTokenEnv] = testRuntimeResetToken
	}
	configtest.BindPostgresEnvToDatabaseRoot(t, tempRoots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], effectiveEnv, postgres.PurposeRuntime)

	cfg := configtest.LoadFixture(t, []string{"config", "valid.toml"}, effectiveEnv).Deployment()
	clock := httpapi.NewTestClock()
	if deps.ModuleOverrides == nil {
		deps.ModuleOverrides = map[string]any{}
	} else {
		overrides := make(map[string]any, len(deps.ModuleOverrides)+1)
		for key, value := range deps.ModuleOverrides {
			overrides[key] = value
		}
		deps.ModuleOverrides = overrides
	}
	deps.ModuleOverrides[testClockModuleOverrideKey] = clock
	ctx := context.Background()
	pool, err := appsupport.OpenPostgres(ctx, cfg, effectiveEnv)
	if err != nil {
		t.Fatalf("open reset test postgres fixture: %v", err)
	}
	store, err := appsupport.OpenObjectStore(ctx, cfg, effectiveEnv)
	if err != nil {
		pool.Close()
		t.Fatalf("open reset test object-store fixture: %v", err)
	}
	bootstrapSettings := bootstrap.Settings{ManifestPath: cfg.Bootstrap.FirstAdminManifestPath}
	if err := bootstrap.Preflight(ctx, bootstrapSettings, pool); err != nil {
		pool.Close()
		_ = store.Close()
		t.Fatalf("bootstrap reset test runtime: %v", err)
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
		t.Fatalf("open reset test Recovery postgres fixture: %v", err)
	}
	deps.TestResetDatabase = func(ctx context.Context) error {
		return ResetDatabase(ctx, recoveryPool, deps.TestResetBootstrap)
	}
	deps.Env = effectiveEnv
	deps.Postgres = pool
	deps.ObjectStore = store
	deps.Now = clock.Now
	handler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies:     testHTTPDependencies(deps),
		AdditionalRoutes: append([]httpapi.RouteRegistrar{httpapi.RegisterTestClockRoutes(clock)}, routes...),
	})
	if err != nil {
		pool.Close()
		_ = store.Close()
		t.Fatalf("start reset test runtime: %v", err)
	}
	runtime := &testRuntimeResetRuntime{ObjectStore: store}
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
		recoveryPool.Close()
		pool.Close()
		_ = store.Close()
	})
	return runtime, server
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
	req.Header.Set(testRouteTokenHeader, testRuntimeResetToken)
	return req
}

func cloneTestRuntimeResetEnv(env map[string]string) map[string]string {
	clone := make(map[string]string, len(env))
	for key, value := range env {
		clone[key] = value
	}
	return clone
}

func requireTestRuntimeResetStartupError(t testing.TB, env map[string]string, want string) {
	t.Helper()
	registrar := RegisterTestRuntimeResetRoute()
	err := registrar(http.NewServeMux(), httpapi.DependencySet{Env: env})
	if err == nil {
		t.Fatalf("expected route registration to fail with %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("unexpected route registration error: got %q want containing %q", err.Error(), want)
	}
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
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unexpected status: got %d want %d; read redacted response: %v", resp.StatusCode, want, err)
		}
		t.Fatalf("unexpected status: got %d want %d; redacted response=%s", resp.StatusCode, want, body)
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

func requireNoPermissiveTestRuntimeCORS(t testing.TB, resp *http.Response) {
	t.Helper()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("test runtime route must not emit permissive CORS, got Access-Control-Allow-Origin=%q", got)
	}
}

func prepareTestRuntimeResetBucket(t testing.TB, h *s3test.Harness, prefix string) string {
	t.Helper()
	return h.BootstrapBucketT(t, prefix)
}

func seedTestRuntimeResetRows(t testing.TB, db *sql.DB) {
	t.Helper()
	var userID string
	if err := db.QueryRowContext(context.Background(), `SELECT id::text FROM users WHERE is_deployment_admin = true LIMIT 1`).Scan(&userID); err != nil {
		t.Fatalf("lookup bootstrap user: %v", err)
	}
	var incidentID string
	if err := db.QueryRowContext(context.Background(), `
		INSERT INTO incidents (incident_key, incident_key_canonical, title, status, created_by_user_id, updated_by_user_id)
		VALUES ('RESET-1', 'reset-1', 'Reset proof incident', 'active', $1, $1)
		RETURNING id::text
	`, userID).Scan(&incidentID); err != nil {
		t.Fatalf("seed incident: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
		VALUES ($1, 'timeline_event', $2, $2)
	`, incidentID, userID); err != nil {
		t.Fatalf("seed record: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO user_sessions (
			user_id,
			token_fingerprint,
			authenticated_at,
			last_qualifying_activity_at,
			idle_expires_at,
			absolute_expires_at,
			session_expires_at
		)
		VALUES ($1, decode('001122', 'hex'), now(), now(), now() + interval '1 hour', now() + interval '1 hour', now() + interval '1 hour')
	`, userID); err != nil {
		t.Fatalf("seed user session: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO route_idempotency (route_key, scope_key, client_txn_id, actor_user_id, request_hash, status_code, response_json)
		VALUES ('test.reset', 'scope', 'txn-reset', $1, decode('334455', 'hex'), 200, '{}'::jsonb)
	`, userID); err != nil {
		t.Fatalf("seed route idempotency: %v", err)
	}
}

func requireSQLCount(t testing.TB, db *sql.DB, query string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), query).Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	return count
}

func requireSQLCountEqual(t testing.TB, db *sql.DB, query string, want int) {
	t.Helper()
	if got := requireSQLCount(t, db, query); got != want {
		t.Fatalf("unexpected count for %q: got %d want %d", query, got, want)
	}
}
