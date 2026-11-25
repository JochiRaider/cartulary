package harnessruntime

import (
	"database/sql"
	"encoding/json"
	"errors"
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

type lockWaitRows struct {
	rows [][]any
	err  error
}

func (r *lockWaitRows) Next() bool { return len(r.rows) > 0 }
func (r *lockWaitRows) Err() error { return r.err }
func (r *lockWaitRows) Scan(dest ...any) error {
	row := r.rows[0]
	r.rows = r.rows[1:]
	for i, value := range row {
		switch target := dest[i].(type) {
		case *int64:
			*target = value.(int64)
		case *string:
			*target = value.(string)
		case *int:
			*target = value.(int)
		}
	}
	return nil
}

func TestLockWaitCollectionKeepsOnlyClosedVocabulary(t *testing.T) {
	rows := &lockWaitRows{rows: [][]any{
		{int64(11), "client backend", "active", "Lock", "transactionid", "update", "12,13", int64(3100), 70},
		{int64(12), "client backend", "idle in transaction", "Client", "ClientRead", "/* private comment", "", int64(9000), 70},
		{int64(13), "client backend; DROP", "active", "Lock", "relation\nprivate", "", "", int64(0), 70},
	}}
	snapshot, err := lockwaits.Collect(rows, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Blocked() != 1 || snapshot.OmittedSessionCount != 67 || snapshot.CapturedAt.Location() != time.UTC {
		t.Fatalf("snapshot counts = blocked %d omitted %d", snapshot.Blocked(), snapshot.OmittedSessionCount)
	}
	first, second, third := snapshot.Sessions[0], snapshot.Sessions[1], snapshot.Sessions[2]
	if first.StatementKind != "update" || len(first.BlockingPIDs) != 2 || first.BlockingPIDs[1] != 13 {
		t.Fatalf("blocked session = %#v", first)
	}
	if second.StatementKind != "other" || len(second.BlockingPIDs) != 0 {
		t.Fatalf("an unreadable statement must be other: %#v", second)
	}
	if third.BackendType != "other" || third.WaitEvent != "other" || third.StatementKind != "none" {
		t.Fatalf("names outside the server vocabulary must be other: %#v", third)
	}
	failing := &lockWaitRows{err: errors.New("cursor failed")}
	if _, err := lockwaits.Collect(failing, time.Now()); err == nil {
		t.Fatal("a failed cursor produced a snapshot")
	}
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
		_, err := sessions.ExecContext(ctx, `UPDATE stall_probe SET id = 1 WHERE id = 1`)
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
	if blocked.StatementKind != "update" || blocked.WaitEventType != "Lock" {
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
