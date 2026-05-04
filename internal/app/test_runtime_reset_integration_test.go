package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestTestRuntimeResetRouteDisabledByDefault(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PreparePackageDatabaseT(t, "test-runtime-reset-disabled")
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
}

func TestTestRuntimeResetRouteClearsStateAndRestoresBootstrap(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PreparePackageDatabaseT(t, "test-runtime-reset")
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

	resp := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil))
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

	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM goose_db_version`, beforeGooseVersions)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 1)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM incidents`, 0)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM records`, 0)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM user_sessions`, 0)
	requireSQLCountEqual(t, db, `SELECT COUNT(*) FROM route_idempotency`, 0)

	loginResp := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/auth/login", map[string]any{
		"username": "bootstrap-admin@example.test",
		"password": "BootstrapPass1!",
	}))
	requireTestRuntimeResetErrorEnvelope(t, loginResp, http.StatusUnauthorized, "mfa_setup_required")
}

func startTestRuntimeResetServer(t testing.TB, env map[string]string, routes []httpapi.RouteRegistrar) (*Runtime, *httptest.Server) {
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
	configtest.BindPostgresEnvToDatabaseRoot(t, tempRoots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], effectiveEnv)

	cfg := configtest.LoadEffectiveFixture(t, []string{"config", "valid.toml"}, effectiveEnv)
	clock := httpapi.NewTestClock()
	runtime, err := NewRuntime(context.Background(), cfg, Options{
		Env: effectiveEnv,
		Now: clock.Now,
		HTTP: httpapi.Options{
			AdditionalRoutes: append([]httpapi.RouteRegistrar{httpapi.RegisterTestClockRoutes(clock)}, routes...),
		},
	})
	if err != nil {
		t.Fatalf("start reset test runtime: %v", err)
	}
	server := httptest.NewServer(runtime.Handler)
	t.Cleanup(func() {
		server.Close()
		runtime.Close()
	})
	return runtime, server
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
	bucket, err := h.BootstrapBucket(context.Background(), prefix)
	if err != nil {
		t.Fatalf("bootstrap reset test bucket: %v", err)
	}
	t.Cleanup(func() {
		if err := h.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup reset test bucket: %v", err)
		}
	})
	return bucket
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
