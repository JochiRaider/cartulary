package harnessruntime

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime/lockwaits"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func requireStallDiagnosticsContract(t *testing.T, diagnostics StallDiagnostics) {
	t.Helper()
	encoded, err := json.Marshal(diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	value, err := harnessschema.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := harnessschema.Validate(stallDiagnosticsSchemaID, value); err != nil {
		t.Fatalf("stall diagnostics violate their schema: %v\n%s", err, encoded)
	}
}

func TestStallDiagnosticsRouteRequiresHarnessAuthorization(t *testing.T) {
	disabled := startTestRuntimeIdentityServer(t, map[string]string{}, RegisterStallDiagnosticsRoute())
	resp := doTestRuntimeResetRequest(t, disabled.Client(), newTestRuntimeResetJSONRequest(t, http.MethodGet, disabled.URL+stallDiagnosticsPath, nil))
	requireTestRuntimeResetStatus(t, resp, http.StatusNotFound)
	resp.Body.Close()

	server := startTestRuntimeIdentityServer(t, testRuntimeEnabledEnv(), RegisterStallDiagnosticsRoute())
	requireTestRuntimeResetErrorEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+stallDiagnosticsPath, nil)), http.StatusForbidden, "test_route_forbidden")
	requireTestRuntimeResetErrorEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+stallDiagnosticsPath, nil))), http.StatusMethodNotAllowed, "method_not_allowed")

	success := doTestRuntimeResetRequest(t, server.Client(), authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+stallDiagnosticsPath, nil)))
	body := requireTestRuntimeResetSuccessEnvelope(t, success, http.StatusOK)
	encoded, err := json.Marshal(body["data"])
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics StallDiagnostics
	if err := json.Unmarshal(encoded, &diagnostics); err != nil {
		t.Fatal(err)
	}
	requireStallDiagnosticsContract(t, diagnostics)
	// Without a database the observation says so instead of failing.
	if diagnostics.DatabasePool != nil || diagnostics.LockWaits.Status != "unavailable" || diagnostics.LockWaits.Reason == nil || *diagnostics.LockWaits.Reason != "no_database" {
		t.Fatalf("database observation without a pool = %#v %#v", diagnostics.DatabasePool, diagnostics.LockWaits)
	}
	if diagnostics.Goroutines.Count < 1 || !strings.HasPrefix(diagnostics.Goroutines.Profile, "goroutine profile: total ") || !strings.Contains(diagnostics.Goroutines.Profile, "harnessruntime.TestStallDiagnosticsRouteRequiresHarnessAuthorization") {
		t.Fatalf("goroutine profile does not name this test's stack: %#v", diagnostics.Goroutines.Count)
	}
	if strings.Contains(diagnostics.Goroutines.Profile, "# labels:") {
		t.Fatal("goroutine profile retained label values")
	}
	requireNoPermissiveTestRuntimeCORS(t, success)
}

// TestStallDiagnosticsObserveLockWaitsAndPoolExhaustion holds a row lock that
// blocks a second session, then exhausts a one-connection pool: the first is
// an observed lock wait, the second an unavailable snapshot whose pool
// counters show every connection acquired.
func TestStallDiagnosticsObserveLockWaitsAndPoolExhaustion(t *testing.T) {
	testDB := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "stall-diagnostics")
	ctx := t.Context()
	sessions, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sessions.Close() })
	if _, err := sessions.ExecContext(ctx, `CREATE TABLE stall_probe (id int PRIMARY KEY); INSERT INTO stall_probe VALUES (1)`); err != nil {
		t.Fatal(err)
	}
	holder, err := sessions.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = holder.Rollback() })
	if _, err := holder.ExecContext(ctx, `SELECT id FROM stall_probe WHERE id = 1 FOR UPDATE`); err != nil {
		t.Fatal(err)
	}
	waiterDone := make(chan error, 1)
	go func() {
		// A leading comment hides the keyword: the query reduces it to "other".
		_, err := sessions.ExecContext(ctx, `/* stall-probe */ UPDATE stall_probe SET id = 1 WHERE id = 1`)
		waiterDone <- err
	}()

	config, err := pgxpool.ParseConfig(testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	config.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	var observed StallDiagnostics
	deadline := time.Now().Add(10 * time.Second)
	for {
		observed = ObserveStall(ctx, pool)
		if observed.LockWaits.Snapshot != nil && observed.LockWaits.Snapshot.Blocked() > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("lock wait never observed: %#v", observed.LockWaits)
		}
		time.Sleep(50 * time.Millisecond)
	}
	requireStallDiagnosticsContract(t, observed)
	var blocked lockwaits.Session
	for _, session := range observed.LockWaits.Snapshot.Sessions {
		if len(session.BlockingPIDs) > 0 {
			blocked = session
		}
	}
	if blocked.StatementKind != "other" || blocked.WaitEventType != "Lock" {
		t.Fatalf("blocked session = %#v", blocked)
	}
	if err := holder.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := <-waiterDone; err != nil {
		t.Fatal(err)
	}

	held, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	exhausted := ObserveStall(ctx, pool)
	held.Release()
	requireStallDiagnosticsContract(t, exhausted)
	if exhausted.LockWaits.Status != "unavailable" || *exhausted.LockWaits.Reason != "pool_acquire_timeout" {
		t.Fatalf("exhausted pool observation = %#v", exhausted.LockWaits)
	}
	if pool := exhausted.DatabasePool; pool == nil || pool.MaxConns != 1 || pool.AcquiredConns != 1 || pool.EmptyAcquireCount < 1 {
		t.Fatalf("exhausted pool counters = %#v", exhausted.DatabasePool)
	}
}
