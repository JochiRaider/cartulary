package harnessruntime

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httpapiextensions"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

const testRuntimeResetToken = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"

func TestTestRuntimeIdentityRouteDisabledByDefault(t *testing.T) {
	server := startTestRuntimeIdentityServer(t, map[string]string{}, RegisterTestRuntimeIdentityRoute())
	resp := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+"/api/v1/test/runtime/identity", nil))
	defer resp.Body.Close()
	requireTestRuntimeResetStatus(t, resp, http.StatusNotFound)
}

func TestTestRuntimeIdentityRouteRequiresHarnessAuthorization(t *testing.T) {
	server := startTestRuntimeIdentityServer(t, testRuntimeEnabledEnv(), RegisterTestRuntimeIdentityRoute())

	missing := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+"/api/v1/test/runtime/identity", nil))
	requireTestRuntimeResetErrorEnvelope(t, missing, http.StatusForbidden, "test_route_forbidden")

	wrong := newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+"/api/v1/test/runtime/identity", nil)
	wrong.Header.Set(testRouteTokenHeader, "ABCDEFGabcdefghijklmnopqrstuvwxyz0123456789")
	requireTestRuntimeResetErrorEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), wrong), http.StatusForbidden, "test_route_forbidden")

	success := doTestRuntimeResetRequest(t, server.Client(), authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+"/api/v1/test/runtime/identity", nil)))
	body := requireTestRuntimeResetSuccessEnvelope(t, success, http.StatusOK)
	data := body["data"].(map[string]any)
	if data["schema_id"] != testRuntimeIdentitySchemaID || data["runtime_marker"] != testRuntimeMarkerValue || data["test_routes_enabled"] != true {
		t.Fatalf("unexpected runtime identity: %#v", data)
	}
	if _, ok := data["server_pid"].(float64); !ok {
		t.Fatalf("identity route must report server_pid: %#v", data)
	}
	requireNoPermissiveTestRuntimeCORS(t, success)
}

func TestTestRuntimeIdentityRouteRejectsNonHarnessOriginAndHost(t *testing.T) {
	guard := httpapi.TestRouteGuard{
		Token:        testRuntimeResetToken,
		ExpectedHost: "127.0.0.1:8080",
		AllowedOrigins: map[string]struct{}{
			"http://127.0.0.1:8080": {},
			"http://127.0.0.1:4173": {},
		},
	}

	for _, tc := range []struct {
		name   string
		host   string
		origin string
	}{
		{name: "wrong-host", host: "evil.example.test", origin: "http://127.0.0.1:4173"},
		{name: "missing-origin", host: "127.0.0.1:8080"},
		{name: "wrong-origin", host: "127.0.0.1:8080", origin: "http://evil.example.test"},
		{name: "malformed-origin", host: "127.0.0.1:8080", origin: "http://127.0.0.1:4173/path"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, "http://127.0.0.1:8080/api/v1/test/runtime/identity", nil))
			req.Host = tc.host
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			recorder := httptest.NewRecorder()
			handleTestRuntimeIdentity(recorder, req, guard)
			requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")
		})
	}

	allowed := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, "http://127.0.0.1:8080/api/v1/test/runtime/identity", nil))
	allowed.Host = "127.0.0.1:8080"
	allowed.Header.Set("Origin", "http://127.0.0.1:4173")
	recorder := httptest.NewRecorder()
	handleTestRuntimeIdentity(recorder, allowed, guard)
	resp := recorder.Result()
	requireTestRuntimeResetStatus(t, resp, http.StatusOK)
	requireNoPermissiveTestRuntimeCORS(t, resp)
	_ = readTestRuntimeResetJSONBody(t, resp)
}

func TestTestRuntimeIdentityRouteRejectsInvalidHarnessPredicates(t *testing.T) {
	wrongMarker := testRuntimeEnabledEnv()
	wrongMarker[testRuntimeMarkerEnv] = "not-owned"
	requireTestRuntimeIdentityStartupError(t, wrongMarker, "must be \"harness-owned\"")

	weakToken := testRuntimeEnabledEnv()
	weakToken[testRouteTokenEnv] = "short"
	requireTestRuntimeIdentityStartupError(t, weakToken, "must be a visible ASCII token")
}

func TestTestRuntimeResetRouteIsRetired(t *testing.T) {
	server := startTestRuntimeIdentityServer(t, testRuntimeEnabledEnv(), RegisterTestRuntimeIdentityRoute())
	resp := doTestRuntimeResetRequest(t, server.Client(), authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil)))
	defer resp.Body.Close()
	requireTestRuntimeResetStatus(t, resp, http.StatusNotFound)
}

func TestResetDatabaseRestoresBootstrapAndPreservesMigrationMetadata(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "recovery-reset")
	env := testDB.Env()
	tempRoots := configtest.SetupTempRoots(t)
	for key, value := range tempRoots.Paths {
		env[key] = value
	}
	configtest.BindPostgresEnvToDatabaseRoot(t, tempRoots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], env, postgres.PurposeRuntime)
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")
	cfg := configtest.LoadFixture(t, []string{"config", "valid.toml"}, env).Deployment()

	ctx := context.Background()
	runtimePool, err := appsupport.OpenPostgres(ctx, cfg, env)
	if err != nil {
		t.Fatalf("open runtime database: %v", err)
	}
	t.Cleanup(runtimePool.Close)
	bootstrapSettings := bootstrap.Settings{ManifestPath: cfg.Bootstrap.FirstAdminManifestPath}
	if err := bootstrap.Preflight(ctx, bootstrapSettings, runtimePool); err != nil {
		t.Fatalf("bootstrap reset fixture: %v", err)
	}
	recoveryPool, err := postgres.Setup(ctx, postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          env[suiteservices.PostgresDSNEnv],
		Purpose:      postgres.PurposeRecovery,
		ExpectedRole: "cartulary_recovery",
	})
	if err != nil {
		t.Fatalf("open Recovery database: %v", err)
	}
	t.Cleanup(recoveryPool.Close)
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open reset assertion database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	beforeGooseVersions := requireSQLCount(t, db, `SELECT COUNT(*) FROM goose_db_version`)
	beforeLineage := requireSQLCount(t, db, `SELECT COUNT(*) FROM schema_migration_lineage`)
	seedTestRuntimeResetRows(t, db)
	result, err := ResetDatabase(ctx, recoveryPool, func(ctx context.Context, tx pgx.Tx) error {
		return bootstrap.PreflightTx(ctx, bootstrapSettings, tx)
	})
	if err != nil {
		t.Fatalf("reset database: %v", err)
	}
	if !result.MigrationMetadataPreserved || !result.BootstrapAdminRestored || result.MutableTableCount != len(result.TablesReset) {
		t.Fatalf("unexpected reset proof: %#v", result)
	}
	if !sortStringsAreStrict(result.TablesReset) {
		t.Fatalf("mutable table proof must be sorted and unique: %#v", result.TablesReset)
	}
	if len(result.TableCounts) != len(result.TablesReset) {
		t.Fatalf("mutable table counts must cover every reset table: %#v", result.TableCounts)
	}
	for index, table := range result.TablesReset {
		if result.TableCounts[index].Table != table || result.TableCounts[index].Rows < 0 {
			t.Fatalf("mutable table counts must remain sorted and non-negative: %#v", result.TableCounts)
		}
	}
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM goose_db_version`, beforeGooseVersions)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM schema_migration_lineage`, beforeLineage)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 1)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM incidents`, 0)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM records`, 0)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM user_sessions`, 0)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM route_idempotency`, 0)

	recoveryPool.Close()
	if _, err := db.ExecContext(ctx, "ALTER DATABASE "+pgx.Identifier{testDB.Name}.Sanitize()+" SET lock_timeout = '100ms'"); err != nil {
		t.Fatalf("configure isolated contention lock timeout: %v", err)
	}
	contentionRecovery, err := postgres.Setup(ctx, postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          env[suiteservices.PostgresDSNEnv],
		Purpose:      postgres.PurposeRecovery,
		ExpectedRole: "cartulary_recovery",
	})
	if err != nil {
		t.Fatalf("open contention Recovery database: %v", err)
	}
	t.Cleanup(contentionRecovery.Close)
	lockTransaction, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin conflicting transaction: %v", err)
	}
	t.Cleanup(func() { _ = lockTransaction.Rollback() })
	if _, err := lockTransaction.ExecContext(ctx, `LOCK TABLE incidents IN ACCESS SHARE MODE`); err != nil {
		t.Fatalf("hold reset-conflicting table lock: %v", err)
	}
	_, err = ResetDatabase(ctx, contentionRecovery, func(ctx context.Context, tx pgx.Tx) error {
		return bootstrap.PreflightTx(ctx, bootstrapSettings, tx)
	})
	var contentionFailure *DatabaseResetFailure
	if !errors.As(err, &contentionFailure) {
		t.Fatalf("contention reset must return a typed failure: %v", err)
	}
	if contentionFailure.Stage() != "recovery_reset_truncate_failed" || contentionFailure.SQLState() != "55P03" {
		t.Fatalf("unexpected contention diagnostic: stage=%q sqlstate=%q", contentionFailure.Stage(), contentionFailure.SQLState())
	}
}

func TestDatabaseResetFailurePreservesCauseAndExposesOnlyStageAndSQLState(t *testing.T) {
	cause := &pgconn.PgError{Code: "55P03", Message: "relation secret_table contains secret-row"}
	failure := NewDatabaseResetFailure("recovery_reset_truncate_failed", cause)
	if failure.Error() != "recovery_reset_truncate_failed" || failure.Stage() != "recovery_reset_truncate_failed" {
		t.Fatalf("unexpected public reset failure: %q", failure.Error())
	}
	if failure.SQLState() != "55P03" {
		t.Fatalf("unexpected SQLSTATE: %q", failure.SQLState())
	}
	if strings.Contains(failure.Error(), "secret") {
		t.Fatalf("public reset failure leaked wrapped error: %q", failure.Error())
	}
	var postgresError *pgconn.PgError
	if !errors.As(failure, &postgresError) || postgresError != cause {
		t.Fatal("wrapped PostgreSQL cause must remain available internally")
	}
}

func startTestRuntimeIdentityServer(t testing.TB, env map[string]string, routes ...httpapi.RouteRegistrar) *httptest.Server {
	t.Helper()
	handler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies:     testHTTPDependencies(httpapi.DependencySet{Env: env}),
		AdditionalRoutes: routes,
	})
	if err != nil {
		t.Fatalf("start test runtime identity server: %v", err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
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
	req.Header.Set(testRouteTokenHeader, testRuntimeResetToken)
	return req
}

func testRuntimeEnabledEnv() map[string]string {
	return map[string]string{
		testRoutesEnabledEnv: "1",
		testRuntimeMarkerEnv: testRuntimeMarkerValue,
		testRouteTokenEnv:    testRuntimeResetToken,
	}
}

func requireTestRuntimeIdentityStartupError(t testing.TB, env map[string]string, want string) {
	t.Helper()
	err := RegisterTestRuntimeIdentityRoute()(http.NewServeMux(), httpapi.DependencySet{Env: env})
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("unexpected route registration error: got %v want containing %q", err, want)
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
		INSERT INTO user_sessions (user_id, token_fingerprint, authenticated_at, last_qualifying_activity_at, idle_expires_at, absolute_expires_at, session_expires_at)
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

func sortStringsAreStrict(values []string) bool {
	for index := 1; index < len(values); index++ {
		if values[index-1] >= values[index] {
			return false
		}
	}
	return true
}
