package processlease

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"

	"github.com/JochiRaider/cartulary/internal/testutil/testdeadline"
)

// Leases are held for the process lifetime, so they must not consume the
// request pool. Each lease session is still initialized by the pool's own
// admission hook and holds its advisory lock on its own backend.
func TestPostgresLeaseSessionsUseDedicatedConnections_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "process-lease-dedicated-sessions")
	ctx, cancel := testdeadline.Context(t)
	defer cancel()
	const admittedName = "cartulary-lease-admission-hook"
	config, err := pgxpool.ParseConfig(testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	config.MaxConns = 2
	config.AfterConnect = func(ctx context.Context, connection *pgx.Conn) error {
		_, err := connection.Exec(ctx, "SET application_name = '"+admittedName+"'")
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	application, err := AcquireApplicationProcess(ctx, pool, 5*time.Second, time.Second)
	if err != nil {
		t.Fatalf("acquire application process lease: %v", err)
	}
	t.Cleanup(application.Close)
	serving, err := AcquireApplicationRecoveryServing(ctx, pool, 5*time.Second, time.Second)
	if err != nil {
		t.Fatalf("acquire recovery serving lease: %v", err)
	}
	t.Cleanup(serving.Close)
	if stat := pool.Stat(); stat.AcquiredConns() != 0 || stat.TotalConns() != 0 {
		t.Fatalf("held leases drew from the request pool: acquired=%d total=%d max=%d", stat.AcquiredConns(), stat.TotalConns(), stat.MaxConns())
	}

	// The request pool still serves its complete capacity while both leases are held.
	acquireCtx, cancelAcquire := context.WithTimeout(ctx, 5*time.Second)
	defer cancelAcquire()
	pooled := make(map[uint32]bool, config.MaxConns)
	connections := make([]*pgxpool.Conn, 0, config.MaxConns)
	t.Cleanup(func() {
		for _, connection := range connections {
			connection.Release()
		}
	})
	for index := range int(config.MaxConns) {
		connection, err := pool.Acquire(acquireCtx)
		if err != nil {
			t.Fatalf("request pool connection %d of %d unavailable while leases are held: %v", index+1, config.MaxConns, err)
		}
		connections = append(connections, connection)
		pooled[connection.Conn().PgConn().PID()] = true
	}

	leases := map[string]*Lease{"application process": application.Lease, "recovery serving": serving.Lease}
	sessions := map[string]bool{}
	for purpose, lease := range leases {
		identity := lease.SessionIdentity()
		pid, err := strconv.ParseUint(identity, 10, 32)
		if err != nil {
			t.Fatalf("%s lease identity %q: %v", purpose, identity, err)
		}
		if pooled[uint32(pid)] || sessions[identity] {
			t.Fatalf("%s lease session %s is not a dedicated connection", purpose, identity)
		}
		sessions[identity] = true
		var name string
		var advisoryLocks int
		if err := connections[0].QueryRow(ctx, `
SELECT application_name,
       (SELECT count(*) FROM pg_locks WHERE locktype = 'advisory' AND granted AND pid = $1)
  FROM pg_stat_activity
 WHERE pid = $1`, int32(pid)).Scan(&name, &advisoryLocks); err != nil {
			t.Fatalf("observe %s lease session: %v", purpose, err)
		}
		if name != admittedName || advisoryLocks != 1 {
			t.Fatalf("%s lease session application_name=%q advisory_locks=%d; want admitted session holding one lock", purpose, name, advisoryLocks)
		}
	}

	if err := serving.Release(ctx); err != nil {
		t.Fatalf("release recovery serving lease: %v", err)
	}
	if err := application.Release(ctx); err != nil {
		t.Fatalf("release application process lease: %v", err)
	}
	if stat := pool.Stat(); stat.AcquiredConns() != config.MaxConns || stat.TotalConns() != config.MaxConns {
		t.Fatalf("lease release changed request pool: acquired=%d total=%d", stat.AcquiredConns(), stat.TotalConns())
	}
}
