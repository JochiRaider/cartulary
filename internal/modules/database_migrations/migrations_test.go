package database_migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	postgres "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestApplyCancelsLongRunningMigration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB, err := postgresHarness.NewMigrationDatabase(context.Background(), "migration-cancel")
	if err != nil {
		t.Fatalf("create migration database: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresHarness.DropMigrationDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop migration database: %v", err)
		}
	})

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open migration database: %v", err)
	}
	defer db.Close()

	migrationDir := t.TempDir()
	migrationPath := filepath.Join(migrationDir, "00001_sleep.sql")
	if err := os.WriteFile(migrationPath, []byte(`-- +goose Up
SELECT pg_sleep(10);
-- +goose Down
SELECT 1;
`), 0o600); err != nil {
		t.Fatalf("write sleep migration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	source, err := postgres.NewSource(os.DirFS(migrationDir), ".", "cancellation.test.v1", "cancellation_test_v1")
	if err != nil {
		t.Fatalf("construct cancellation migration source: %v", err)
	}
	err = postgres.Apply(ctx, db, source)
	if err == nil {
		t.Fatal("expected migration cancellation error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context cancellation migration error, got %v", err)
	}
	requireExternalMigrationFailureReason(t, err, "schema_migration_execution_failed")
}

func TestConcurrentApplyLocking(t *testing.T) {
	for iteration := 0; iteration < 3; iteration++ {
		t.Run(fmt.Sprintf("iteration-%d", iteration+1), func(t *testing.T) {
			postgresHarness := pgtest.Start(t)
			testDB, err := postgresHarness.NewMigrationDatabase(context.Background(), "migration-lock")
			if err != nil {
				t.Fatalf("create migration database: %v", err)
			}
			t.Cleanup(func() {
				if err := postgresHarness.DropMigrationDatabase(context.Background(), testDB.Name); err != nil {
					t.Fatalf("drop migration database: %v", err)
				}
			})

			source := testMigrationSource(t, `-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT pg_advisory_unlock(4097083626::bigint) THEN
        RAISE EXCEPTION 'execution lock missing';
    END IF;
    PERFORM pg_advisory_lock(4097083626::bigint);
END
$$;
-- +goose StatementEnd
SELECT pg_sleep(0.15);
CREATE TABLE migration_lock_probe (singleton boolean PRIMARY KEY);
INSERT INTO migration_lock_probe (singleton) VALUES (true);
-- +goose Down
DROP TABLE migration_lock_probe;
`)
			dbOne := openMigrationDatabase(t, testDB.DSN)
			dbTwo := openMigrationDatabase(t, testDB.DSN)
			start := make(chan struct{})
			errorsByApply := make([]error, 2)
			var wait sync.WaitGroup
			for index, db := range []*sql.DB{dbOne, dbTwo} {
				index := index
				db := db
				wait.Add(1)
				go func() {
					defer wait.Done()
					<-start
					errorsByApply[index] = postgres.Apply(context.Background(), db, source)
				}()
			}
			close(start)
			wait.Wait()
			for index, applyErr := range errorsByApply {
				if applyErr != nil {
					var version int64
					_ = dbOne.QueryRowContext(context.Background(), `SELECT COALESCE(MAX(version_id), -1) FROM goose_db_version WHERE is_applied`).Scan(&version)
					t.Fatalf("concurrent apply %d: %v (observed version %d)", index+1, applyErr, version)
				}
			}

			var probeRows int
			if err := dbOne.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM migration_lock_probe`).Scan(&probeRows); err != nil {
				t.Fatalf("inspect lock probe: %v", err)
			}
			if probeRows != 1 {
				t.Fatalf("lock probe row count = %d", probeRows)
			}
			if err := dbOne.PingContext(context.Background()); err != nil {
				t.Fatalf("caller-owned database was closed: %v", err)
			}
		})
	}
}

func TestApplyCancellationWhileWaitingForLock(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB, err := postgresHarness.NewMigrationDatabase(context.Background(), "migration-lock-cancel")
	if err != nil {
		t.Fatalf("create migration database: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresHarness.DropMigrationDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop migration database: %v", err)
		}
	})

	db := openMigrationDatabase(t, testDB.DSN)
	holder, err := db.Conn(context.Background())
	if err != nil {
		t.Fatalf("open lock holder: %v", err)
	}
	defer holder.Close()
	if _, err := holder.ExecContext(context.Background(), `SELECT pg_advisory_lock($1)`, int64(4097083626)); err != nil {
		t.Fatalf("hold migration lock: %v", err)
	}
	locked := true
	defer func() {
		if locked {
			_, _ = holder.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, int64(4097083626))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	err = postgres.Apply(ctx, db, testMigrationSource(t, validExternalMigrationBody))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lock wait did not retain deadline identity: %v", err)
	}
	requireExternalMigrationFailureReason(t, err, "migration_lock_acquisition_failed")

	if _, err := holder.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, int64(4097083626)); err != nil {
		t.Fatalf("release migration lock: %v", err)
	}
	locked = false
	if err := postgres.Apply(context.Background(), db, testMigrationSource(t, validExternalMigrationBody)); err != nil {
		t.Fatalf("apply after lock cancellation: %v", err)
	}
}

func TestApplyExecutionFailureIsSafeAndReleasesLock(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB, err := postgresHarness.NewMigrationDatabase(context.Background(), "migration-execution-failure")
	if err != nil {
		t.Fatalf("create migration database: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresHarness.DropMigrationDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop migration database: %v", err)
		}
	})
	db := openMigrationDatabase(t, testDB.DSN)
	source := testMigrationSource(t, `-- +goose Up
SELECT secret_bind_payload FROM deliberately_missing_sensitive_relation;
-- +goose Down
SELECT 1;
`)

	err = postgres.Apply(context.Background(), db, source)
	requireExternalMigrationFailureReason(t, err, "schema_migration_execution_failed")
	for _, forbidden := range []string{"secret_bind_payload", "deliberately_missing_sensitive_relation", testDB.DSN, "00001_test.sql"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("safe migration error disclosed %q: %q", forbidden, err.Error())
		}
	}
	requireNoMigrationAdvisoryLocks(t, db)
}

func TestApplyPreflightFailureReleasesLock(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	migrationDB := postgresHarness.MigrationDatabaseThroughT(t, "migration-preflight-failure", 1)
	db := migrationDB.SQL()
	if _, err := db.ExecContext(context.Background(), `UPDATE schema_migration_lineage SET lineage_id = 'historical.private.v1'`); err != nil {
		t.Fatalf("seed historical lineage: %v", err)
	}
	source, err := dbmigrations.Source()
	if err != nil {
		t.Fatalf("load canonical migration source: %v", err)
	}

	err = postgres.Apply(context.Background(), db, source)
	var remediation postgres.RemediationReporter
	if !errors.As(err, &remediation) {
		t.Fatalf("expected remediation reporter, got %T: %v", err, err)
	}
	if remediation.ReasonCode() != "historical_migration_lineage" {
		t.Fatalf("unexpected remediation reason: %q", remediation.ReasonCode())
	}
	requireNoMigrationAdvisoryLocks(t, db)
}

func TestApplyFailedPostconditionReleasesLock(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB, err := postgresHarness.NewMigrationDatabase(context.Background(), "migration-postcondition")
	if err != nil {
		t.Fatalf("create migration database: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresHarness.DropMigrationDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop migration database: %v", err)
		}
	})
	db := openMigrationDatabase(t, testDB.DSN)
	source, err := postgres.NewSource(fstest.MapFS{
		"00001_lineage.sql": &fstest.MapFile{Data: []byte(`-- +goose Up
CREATE TABLE schema_migration_lineage (
    lineage_id text PRIMARY KEY,
    applied_at timestamp with time zone DEFAULT now() NOT NULL,
    description text NOT NULL
);
INSERT INTO schema_migration_lineage (lineage_id, description)
VALUES ('locking.test.v1', 'Migration locking integration test.');
-- +goose Down
DROP TABLE schema_migration_lineage;
`)},
		"00002_suppress.sql": &fstest.MapFile{Data: []byte(`-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION suppress_migration_version() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RETURN NULL;
END
$$;
-- +goose StatementEnd
CREATE TRIGGER suppress_migration_version
BEFORE INSERT ON goose_db_version
FOR EACH ROW
WHEN (NEW.version_id = 2)
EXECUTE FUNCTION suppress_migration_version();
-- +goose Down
DROP FUNCTION suppress_migration_version();
`)},
	}, ".", "locking.test.v1", "locking_test_v1")
	if err != nil {
		t.Fatalf("construct failed-postcondition source: %v", err)
	}

	err = postgres.Apply(context.Background(), db, source)
	requireExternalMigrationFailureReason(t, err, "schema_migration_postcondition_failed")
	requireNoMigrationAdvisoryLocks(t, db)
}

func testMigrationSource(t testing.TB, body string) postgres.Source {
	t.Helper()
	body = strings.Replace(body, "-- +goose Up\n", `-- +goose Up
CREATE TABLE IF NOT EXISTS schema_migration_lineage (
    lineage_id text PRIMARY KEY,
    applied_at timestamp with time zone DEFAULT now() NOT NULL,
    description text NOT NULL
);
INSERT INTO schema_migration_lineage (lineage_id, description)
VALUES ('locking.test.v1', 'Migration locking integration test.')
ON CONFLICT (lineage_id) DO NOTHING;
`, 1)
	source, err := postgres.NewSource(
		fstest.MapFS{"00001_test.sql": &fstest.MapFile{Data: []byte(body)}},
		".",
		"locking.test.v1",
		"locking_test_v1",
	)
	if err != nil {
		t.Fatalf("construct test source: %v", err)
	}
	return source
}

const validExternalMigrationBody = "-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 1;\n"

func openMigrationDatabase(t testing.TB, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open migration database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func requireExternalMigrationFailureReason(t testing.TB, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected migration failure %q", want)
	}
	var failure postgres.MigrationFailure
	if !errors.As(err, &failure) {
		t.Fatalf("expected MigrationFailure, got %T: %v", err, err)
	}
	if failure.ReasonCode() != want || failure.Error() != want {
		t.Fatalf("migration failure = %q / %q, want %q", failure.ReasonCode(), failure.Error(), want)
	}
}

func requireNoMigrationAdvisoryLocks(t testing.TB, db *sql.DB) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), `
SELECT COUNT(*)
FROM pg_locks
WHERE locktype = 'advisory'
  AND database = (SELECT oid FROM pg_database WHERE datname = current_database())
  AND granted
`).Scan(&count); err != nil {
		t.Fatalf("inspect migration advisory locks: %v", err)
	}
	if count != 0 {
		t.Fatalf("migration advisory locks remain: %d", count)
	}
}
