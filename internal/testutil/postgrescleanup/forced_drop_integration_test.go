package postgrescleanup_test

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescatalog"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
)

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
			forced, err := postgrescleanup.DropOwnedDatabase(
				context.Background(),
				harness.AdminDSN(),
				database.Name,
				5*time.Second,
				15*time.Second,
			)
			if err != nil {
				errorsByDatabase <- fmt.Errorf("drop %s: %w", database.Name, err)
				return
			}
			if !forced {
				errorsByDatabase <- fmt.Errorf("drop %s did not use forced fallback", database.Name)
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
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- postgrescatalog.WithMutation(
			context.Background(),
			harness.AdminDSN(),
			databaseName,
			5*time.Second,
			func(*sql.DB) error {
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
			func(admin *sql.DB) error {
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
			func(admin *sql.DB) error {
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

	if err := postgrescleanup.ForceDropDatabase(context.Background(), harness.AdminDSN(), databaseName); err != nil {
		t.Fatalf("clean coordinated database: %v", err)
	}
	if err := postgrescleanup.ForceDropDatabase(context.Background(), harness.AdminDSN(), differentDatabaseName); err != nil {
		t.Fatalf("clean different-target database: %v", err)
	}
}
