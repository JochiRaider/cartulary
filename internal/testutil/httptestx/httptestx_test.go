package httptestx

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestHarnessBootsServerAndAssertsEnvelopes(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "httptestx")

	s3Harness := s3test.Start(t)
	bucket := s3Harness.BootstrapBucketT(t, "httptestx")

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
	bucket := s3Harness.BootstrapBucketT(t, "httptestx-test-route-mode")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	// These runtimes reopen one database. Its retained Base content must outlive
	// each child runtime, just as it does across a deployment process restart.
	env["CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH"] = t.TempDir()

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

// TestStartServerRecordsLifecycleFatal covers TH-HARNESS-REQ-863 and AC-158
// for an in-process test server: losing its application-process lease forces
// a lifecycle fatal, which writes exactly one server_fatal record with its
// reason code; a request that then meets the 503 keeps the body naming the
// fatal, and a poll stops on it instead of waiting to its deadline. The record
// goes to a private run root, so this deliberate fatal does not fail the run.
func TestStartServerRecordsLifecycleFatal(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "httptestx-fatal")
	s3Harness := s3test.Start(t)
	bucket := s3Harness.BootstrapBucketT(t, "httptestx-fatal")
	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__TIMEOUTS__EXTENSIONS__PROCESS_LEASE_LOSS_DETECTION_SECONDS"] = "1"
	lookup, stream := privateEvidenceRun(t)
	server := StartServer(t, ServerOptions{Env: env, TestRouteMode: TestRouteModeHarnessOwned, evidenceLookup: lookup})

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	// The isolated database has exactly one application-process lease
	// session; never infer an owner from a global PID.
	var sessions, leasePID int
	if err := db.QueryRowContext(t.Context(), `
SELECT count(DISTINCT l.pid), COALESCE(min(l.pid),0)
FROM pg_catalog.pg_locks l
JOIN pg_catalog.pg_stat_activity a ON a.pid=l.pid
WHERE l.locktype='advisory' AND l.mode='ExclusiveLock' AND l.granted
  AND l.objsubid=1 AND pg_has_role(a.usesysid,'cartulary_runtime','MEMBER')
  AND l.database=(SELECT oid FROM pg_catalog.pg_database WHERE datname=current_database())
`).Scan(&sessions, &leasePID); err != nil || sessions != 1 {
		t.Fatalf("ambiguous application lease: sessions=%d error=%v", sessions, err)
	}
	var terminated bool
	if err := db.QueryRowContext(t.Context(), `SELECT pg_terminate_backend($1)`, leasePID).Scan(&terminated); err != nil || !terminated {
		t.Fatalf("terminate lease session: terminated=%v error=%v", terminated, err)
	}
	select {
	case <-server.runtime.FatalDone():
	case <-time.After(30 * time.Second):
		t.Fatal("lease loss did not reach a lifecycle fatal")
	}

	response := Do(t, server.HTTP.Client(), NewJSONRequest(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/success", nil))
	if _, err := readResponseBody(response, http.StatusOK); err == nil || !strings.Contains(err.Error(), "got 503") || !strings.Contains(err.Error(), "extension_integrity_failure") {
		t.Fatalf("request after the fatal lost its body: %v", err)
	}
	if _, err := pollJob(t.Context(), server.HTTP.URL, "00000000-0000-0000-0000-000000000000", PollOptions{Timeout: 30 * time.Second, Client: server.HTTP.Client()}, []string{"succeeded"}); err == nil || !strings.Contains(err.Error(), "stopped on a non-retryable response") || !strings.Contains(err.Error(), "extension_integrity_failure") {
		t.Fatalf("poll waited through a fatal server: %v", err)
	}

	server.Close()
	records := fatalRecords(t, stream)
	if len(records) != 1 || records[0]["kind"] != "server_fatal" || records[0]["reason_code"] != "application_process_lease_lost" || records[0]["process_kind"] != "in_process" {
		t.Fatalf("fatal records = %v, want one application_process_lease_lost", records)
	}
}

// TestLockWaitSnapshotNamesBlockedAndBlockingSessions covers TH-HARNESS-REQ-864
// and AC-159: a request held behind a lock is recorded as a blocked session
// naming its blocker, with wait events and statement kinds, and no statement
// text or parameters, in the run root.
func TestLockWaitSnapshotNamesBlockedAndBlockingSessions(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "httptestx-lock-wait")
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	holder, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Close()
	waiter, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer waiter.Close()
	var holderPID, waiterPID int64
	if err := holder.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&holderPID); err != nil {
		t.Fatal(err)
	}
	if err := waiter.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&waiterPID); err != nil {
		t.Fatal(err)
	}
	if _, err := holder.ExecContext(ctx, `SELECT pg_advisory_lock(864)`); err != nil {
		t.Fatal(err)
	}
	waited := make(chan error, 1)
	go func() {
		_, err := waiter.ExecContext(ctx, `SELECT /* sentinel-statement-text */ pg_advisory_lock($1), $2::text`, 864, "sentinel-parameter-value")
		waited <- err
	}()
	var snapshot LockWaitSnapshot
	for {
		snapshot, err = SnapshotLockWaits(ctx, db)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Blocked() > 0 || ctx.Err() != nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	byPID := map[int64]LockWaitSession{}
	for _, session := range snapshot.Sessions {
		byPID[session.PID] = session
	}
	blocked, blocking := byPID[waiterPID], byPID[holderPID]
	if len(blocked.BlockingPIDs) != 1 || blocked.BlockingPIDs[0] != holderPID || blocked.WaitEventType != "Lock" || blocked.WaitEvent != "advisory" || blocked.StatementKind != "select" || blocked.State != "active" {
		t.Fatalf("blocked session = %+v in %+v", blocked, snapshot)
	}
	if blocking.PID != holderPID || len(blocking.BlockingPIDs) != 0 || blocking.StatementKind != "select" || blocking.BackendType != "client backend" {
		t.Fatalf("blocking session = %+v in %+v", blocking, snapshot)
	}

	lookup, _ := privateEvidenceRun(t)
	path, _, err := recordLockWaitsTo(ctx, db, lookup)
	if err != nil {
		t.Fatal(err)
	}
	results, _ := lookup("CARTULARY_TEST_RESULTS_DIR")
	file := filepath.Join(results, "run", filepath.FromSlash(path))
	info, err := os.Lstat(file)
	if err != nil || info.Mode().Perm() != 0o600 || !strings.HasPrefix(path, lockWaitDirectory+"/") {
		t.Fatalf("snapshot file %s = %v, %v", path, info, err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var recorded LockWaitSnapshot
	if err := json.Unmarshal(data, &recorded); err != nil || recorded.Blocked() == 0 {
		t.Fatalf("recorded snapshot = %s, %v", data, err)
	}
	for _, text := range []string{"sentinel", "pg_advisory_lock", "/*", testDB.Name} {
		if strings.Contains(string(data), text) {
			t.Fatalf("snapshot carries statement text or identity %q: %s", text, data)
		}
	}

	if _, err := holder.ExecContext(ctx, `SELECT pg_advisory_unlock(864)`); err != nil {
		t.Fatal(err)
	}
	if err := <-waited; err != nil {
		t.Fatal(err)
	}
	if _, err := waiter.ExecContext(ctx, `SELECT pg_advisory_unlock(864)`); err != nil {
		t.Fatal(err)
	}
}
