package processlease

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestPostgresApplicationProcessLease_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "extensions-process-lease")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	warmCtx, cancelWarm := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelWarm()
	if err := warmPostgresLeaseSessions(warmCtx, pool, 2); err != nil {
		t.Fatalf("warm postgres process lease sessions: %v", err)
	}
	backend := PostgresBackend{Pool: pool}

	first, err := Acquire(context.Background(), backend, 100*time.Millisecond, 40*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire first process lease: %v", err)
	}
	if _, err := Acquire(context.Background(), backend, 20*time.Millisecond, 40*time.Millisecond); !errors.Is(err, ErrApplicationProcessActive) {
		t.Fatalf("overlapping process acquisition = %v", err)
	}
	if err := first.Release(context.Background()); err != nil {
		t.Fatalf("release first process lease: %v", err)
	}

	crashed, err := Acquire(context.Background(), backend, 100*time.Millisecond, 40*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire crash lease: %v", err)
	}
	crashed.Close()
	afterCrash, err := Acquire(context.Background(), backend, 100*time.Millisecond, 40*time.Millisecond)
	if err != nil {
		t.Fatalf("database session close did not release crash lease: %v", err)
	}
	if err := afterCrash.Release(context.Background()); err != nil {
		t.Fatal(err)
	}

	lost, err := Acquire(context.Background(), backend, 100*time.Millisecond, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	monitorCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lost.StartMonitor(monitorCtx)
	session := lost.session.(*postgresSession)
	if err := session.connection.Conn().Close(context.Background()); err != nil {
		t.Fatalf("close lease backend session: %v", err)
	}
	waitForPostgresLeaseState(t, lost, StateLost)
	if err := lost.Release(context.Background()); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("lost process lease was releasable/reacquirable: %v", err)
	}
}

func warmPostgresLeaseSessions(ctx context.Context, pool *pgxpool.Pool, count int) error {
	connections := make([]*pgxpool.Conn, 0, count)
	defer func() {
		for _, connection := range connections {
			connection.Release()
		}
	}()

	for index := 0; index < count; index++ {
		connection, err := pool.Acquire(ctx)
		if err != nil {
			return fmt.Errorf("acquire session %d of %d: %w", index+1, count, err)
		}
		connections = append(connections, connection)
	}
	return nil
}

func waitForPostgresLeaseState(t *testing.T, lease *Lease, want State) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if lease.State() == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("lease state = %s; want %s", lease.State(), want)
}
