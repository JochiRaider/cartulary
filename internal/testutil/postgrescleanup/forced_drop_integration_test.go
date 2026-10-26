package postgrescleanup_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescatalog"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestOwnedDatabaseCleanupIsIdempotent(t *testing.T) {
	harness := pgtest.Start(t)
	database := harness.NewDatabaseT(t, "single_drop")
	handle, err := sql.Open("pgx", database.DSN)
	if err != nil {
		t.Fatal(err)
	}
	if err := handle.PingContext(context.Background()); err != nil {
		_ = handle.Close()
		t.Fatal(err)
	}
	if err := handle.Close(); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := postgrescleanup.DropOwnedDatabase(context.Background(), harness.AdminDSN(), database.Name); err != nil {
			t.Fatalf("single owned drop: %v", err)
		}
	}
}

func TestConcurrentActiveDatabaseCleanupUsesTargetScopedForcedAdmission(t *testing.T) {
	harness := pgtest.Start(t)
	const databaseCount = 4
	databases := make([]*pgtest.TestDatabase, 0, databaseCount)
	handles := make([]*sql.DB, 0, databaseCount)
	for index := range databaseCount {
		database := harness.NewDatabaseT(t, fmt.Sprintf("forced_reader_writer_%d", index))
		handle, err := sql.Open("pgx", database.DSN)
		if err != nil {
			t.Fatalf("open active database %s: %v", database.Name, err)
		}
		if err := handle.PingContext(context.Background()); err != nil {
			_ = handle.Close()
			t.Fatalf("establish active database %s: %v", database.Name, err)
		}
		databases = append(databases, database)
		handles = append(handles, handle)
	}
	t.Cleanup(func() {
		for _, handle := range handles {
			_ = handle.Close()
		}
	})

	start := make(chan struct{})
	errorsByDatabase := make(chan error, databaseCount)
	var wait sync.WaitGroup
	for _, database := range databases {
		wait.Add(1)
		go func(database *pgtest.TestDatabase) {
			defer wait.Done()
			<-start
			err := postgrescleanup.DropOwnedDatabase(
				context.Background(),
				harness.AdminDSN(),
				database.Name)
			if err != nil {
				errorsByDatabase <- fmt.Errorf("drop %s: %w", database.Name, err)
				return
			}
		}(database)
	}
	close(start)
	wait.Wait()
	close(errorsByDatabase)
	for err := range errorsByDatabase {
		t.Error(err)
	}

	for index, handle := range handles {
		probeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		err := handle.PingContext(probeCtx)
		cancel()
		if err == nil {
			t.Errorf("database %s remained reachable after forced cleanup", databases[index].Name)
		}
	}
}

func TestCatalogMutationSerializesOneTargetWithoutBlockingAnotherTarget(t *testing.T) {
	harness := pgtest.Start(t)
	databaseName := fmt.Sprintf("ct_catalog_coordination_%d", time.Now().UnixNano())
	differentDatabaseName := databaseName + "_other"
	// The query session must own admission itself. Otherwise cancellation can
	// close its connection while a different session releases target exclusion
	// before PostgreSQL has physically stopped the original catalog operation.
	if err := postgrescatalog.WithMutation(context.Background(), harness.AdminDSN(), databaseName, 5*time.Second, func(admin *sql.Conn) error {
		var held bool
		if err := admin.QueryRowContext(context.Background(), `SELECT EXISTS (SELECT 1 FROM pg_catalog.pg_locks WHERE locktype = 'advisory' AND pid = pg_backend_pid() AND granted)`).Scan(&held); err != nil {
			return err
		}
		if !held {
			return fmt.Errorf("catalog operation session does not hold its target admission")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var cancelledOperation error
	if err := postgrescatalog.WithMutation(context.Background(), harness.AdminDSN(), databaseName, 5*time.Second, func(admin *sql.Conn) error {
		operationCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_, cancelledOperation = admin.ExecContext(operationCtx, `SELECT pg_sleep(60)`)
		// This probe separates the cancelled operation from lock settlement.
		// Server-side exclusion must remain authoritative after client retirement.
		return nil
	}); err != nil {
		t.Fatalf("cancelled operation changed admission outcome: %v", err)
	}
	if !errors.Is(cancelledOperation, context.DeadlineExceeded) {
		t.Fatalf("operation did not preserve cancellation: %v", cancelledOperation)
	}
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- postgrescatalog.WithMutation(
			context.Background(),
			harness.AdminDSN(),
			databaseName,
			5*time.Second,
			func(*sql.Conn) error {
				close(firstEntered)
				<-releaseFirst
				return nil
			},
		)
	}()
	<-firstEntered

	sameTargetEntered := make(chan struct{})
	sameTargetDone := make(chan error, 1)
	go func() {
		sameTargetDone <- postgrescatalog.WithMutation(
			context.Background(),
			harness.AdminDSN(),
			databaseName,
			5*time.Second,
			func(admin *sql.Conn) error {
				close(sameTargetEntered)
				_, err := admin.ExecContext(
					context.Background(),
					fmt.Sprintf(`CREATE DATABASE "%s"`, databaseName),
				)
				return err
			},
		)
	}()
	select {
	case <-sameTargetEntered:
		t.Fatal("same-target database creation entered while its catalog mutation was active")
	case <-time.After(100 * time.Millisecond):
	}

	differentTargetEntered := make(chan struct{})
	differentTargetDone := make(chan error, 1)
	go func() {
		differentTargetDone <- postgrescatalog.WithMutation(
			context.Background(),
			harness.AdminDSN(),
			differentDatabaseName,
			5*time.Second,
			func(admin *sql.Conn) error {
				close(differentTargetEntered)
				_, err := admin.ExecContext(
					context.Background(),
					fmt.Sprintf(`CREATE DATABASE "%s"`, differentDatabaseName),
				)
				return err
			},
		)
	}()
	select {
	case <-differentTargetEntered:
	case <-time.After(5 * time.Second):
		t.Fatal("different-target catalog mutation did not overlap the held target")
	}
	if err := <-differentTargetDone; err != nil {
		t.Fatalf("different-target database creation: %v", err)
	}

	close(releaseFirst)
	if err := <-firstDone; err != nil {
		t.Fatalf("first catalog mutation: %v", err)
	}
	if err := <-sameTargetDone; err != nil {
		t.Fatalf("create database after same-target mutation: %v", err)
	}
	select {
	case <-sameTargetEntered:
	default:
		t.Fatal("same-target database creation never acquired catalog admission")
	}

	if err := postgrescleanup.DropOwnedDatabase(context.Background(), harness.AdminDSN(), databaseName); err != nil {
		t.Fatalf("clean coordinated database: %v", err)
	}
	if err := postgrescleanup.DropOwnedDatabase(context.Background(), harness.AdminDSN(), differentDatabaseName); err != nil {
		t.Fatalf("clean different-target database: %v", err)
	}
}

func TestOwnedDelayedBarrierParticipantIsRedacted(t *testing.T) {
	if suiteservices.LookupEnvValue(nil, suiteservices.PGAdminDSNEnv) != "" {
		t.Fatal("self-owned lifecycle selection acquired an unused broker cluster")
	}
	// The entire cluster belongs to this fixture; never stop a participant in
	// a borrowed suite service. The peer connection remains owned until resumed.
	harness, err := pgtest.StartOwned(t.Context(), pgtest.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := harness.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	database := harness.Fixture().NewDatabaseT(t, "barrier_participant")
	peer, err := pgconn.Connect(t.Context(), harness.AdminDSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = peer.Close(ctx)
	})
	peer.Exec(t.Context(), `SELECT pg_sleep(60)`)
	admin, err := sql.Open("pgx", harness.AdminDSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close() })
	ready, cancelReady := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancelReady()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var active bool
		if err := admin.QueryRowContext(ready, `SELECT EXISTS(SELECT 1 FROM pg_catalog.pg_stat_activity WHERE pid = $1 AND state = 'active' AND wait_event = 'PgSleep')`, peer.PID()).Scan(&active); err != nil {
			t.Fatal(err)
		}
		if active {
			break
		}
		select {
		case <-ready.Done():
			t.Fatal("owned peer did not enter active statement")
		case <-ticker.C:
		}
	}
	resumed := false
	resume := func() {
		if resumed {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		code, _, err := harness.Container.Exec(ctx, []string{"kill", "-CONT", strconv.FormatUint(uint64(peer.PID()), 10)})
		if err != nil || code != 0 {
			t.Error("could not resume exact owned fixture backend")
			return
		}
		resumed = true
	}
	t.Cleanup(resume)
	stop, cancelStop := context.WithTimeout(t.Context(), 3*time.Second)
	code, _, err := harness.Container.Exec(stop, []string{"kill", "-STOP", strconv.FormatUint(uint64(peer.PID()), 10)})
	cancelStop()
	if err != nil || code != 0 {
		t.Fatal("could not stop exact owned fixture backend")
	}
	drop, cancelDrop := context.WithTimeout(t.Context(), 7*time.Second)
	failure := postgrescleanup.DropOwnedDatabase(drop, harness.AdminDSN(), database.Name)
	cancelDrop()
	resume()
	if !errors.Is(failure, context.DeadlineExceeded) {
		t.Fatalf("owned paused backend did not delay barrier: %v", failure)
	}
	var observed interface {
		PostgresCleanup() postgrescleanup.DropAttempt
	}
	if !errors.As(failure, &observed) {
		t.Fatal("drop lost structural observation")
	}
	attempt := observed.PostgresCleanup()
	if attempt.ServiceGeneration == nil || *attempt.ServiceGeneration != harness.ProvisioningGeneration() {
		t.Fatal("drop observation belongs to another PostgreSQL generation")
	}
	value := attempt.Wait
	if value.Status != "sampled" || value.Last.Type != "IPC" || value.Last.Event != "ProcSignalBarrier" {
		t.Fatalf("different wait observed: %s", value.String())
	}
	encoded, err := json.Marshal(value.Last)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot map[string]any
	if err := json.Unmarshal(encoded, &snapshot); err != nil {
		t.Fatal(err)
	}
	participant, ok := snapshot["barrier"].(map[string]any)
	if !ok || participant["status"] != "sampled" || participant["backend"] != "client backend" || participant["state"] != "active" || participant["type"] != "Timeout" || participant["operation"] != "select" || participant["database_scope"] != "admin" {
		t.Fatalf("missing redacted delayed participant: %s", encoded)
	}
	if len(participant) != 9 {
		t.Fatal("participant contains unexpected identity or diagnostic fields")
	}
}
