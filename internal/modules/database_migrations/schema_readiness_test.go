package database_migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	postgres "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestEnsureSchemaReadyAllowsCurrentHead(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "schema-ready-head")
	pool := openSchemaReadinessPool(t, testDB.DSN)

	if err := postgres.EnsureSchemaReady(context.Background(), pool, dbmigrations.Source()); err != nil {
		t.Fatalf("current head schema should be ready: %v", err)
	}
}

func TestEnsureSchemaReadyRejectsEmptyDatabase(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.NewMigrationDatabaseT(t, "schema-ready-empty")
	pool := openSchemaReadinessPool(t, testDB.DSN)

	err := postgres.EnsureSchemaReady(context.Background(), pool, dbmigrations.Source())
	requireSchemaReadinessDiagnostic(t, err, "schema_migration_required")
}

func TestEnsureSchemaReadyRejectsBehindCurrentLine(t *testing.T) {
	_, pool := migratedSchemaReadinessDatabase(t, "schema-ready-behind", 1)

	err := postgres.EnsureSchemaReady(context.Background(), pool, dbmigrations.Source())
	requireSchemaReadinessDiagnostic(t, err, "schema_migration_required")
}

func TestEnsureSchemaReadyRejectsAheadCurrentLine(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "schema-ready-ahead")
	db := openSchemaReadinessSQL(t, testDB.DSN)
	pool := openSchemaReadinessPool(t, testDB.DSN)

	if _, err := db.ExecContext(context.Background(), `INSERT INTO goose_db_version (version_id, is_applied) VALUES (61, true)`); err != nil {
		t.Fatalf("seed ahead migration version: %v", err)
	}

	err := postgres.EnsureSchemaReady(context.Background(), pool, dbmigrations.Source())
	requireSchemaReadinessDiagnostic(t, err, "schema_version_ahead")
}

func TestEnsureSchemaReadyRejectsHistoricalLineAboveHead(t *testing.T) {
	db, pool := migratedSchemaReadinessDatabase(t, "schema-ready-historical-above-head", 0)
	if _, err := db.ExecContext(context.Background(), `
DROP TABLE schema_migration_lineage;
INSERT INTO goose_db_version (version_id, is_applied) VALUES (61, true);
`); err != nil {
		t.Fatalf("seed historical migration line above head: %v", err)
	}

	report := requireMigrationRemediation(t, postgres.EnsureSchemaReady(context.Background(), pool, dbmigrations.Source()))
	if report.Boundary != dbmigrations.LineageBoundary || report.FromVersion != 61 || report.ToVersion != 60 {
		t.Fatalf("unexpected remediation report: %#v", report)
	}
	finding := report.Findings[0]
	if finding.ReasonCode != "historical_migration_lineage" || finding.RawValue != nil || finding.RawValuePair["lineage_table_present"] != false {
		t.Fatalf("unexpected remediation finding: %#v", finding)
	}
}

func TestEnsureSchemaReadyReportsWrongLineage(t *testing.T) {
	db, pool := migratedSchemaReadinessDatabase(t, "schema-ready-wrong-lineage", 1)
	if _, err := db.ExecContext(context.Background(), `
DELETE FROM schema_migration_lineage;
INSERT INTO schema_migration_lineage (lineage_id, description)
VALUES ('cartulary.legacy_line.v1', 'legacy test line');
`); err != nil {
		t.Fatalf("seed wrong lineage: %v", err)
	}

	report := requireMigrationRemediation(t, postgres.EnsureSchemaReady(context.Background(), pool, dbmigrations.Source()))
	finding := report.Findings[0]
	if finding.RawValue == nil || *finding.RawValue != "cartulary.legacy_line.v1" {
		t.Fatalf("expected observed wrong lineage in raw_value: %#v", finding)
	}
	if finding.RawValuePair["expected_lineage_id"] != dbmigrations.LineageID || finding.RawValuePair["lineage_table_present"] != true {
		t.Fatalf("unexpected remediation facts: %#v", finding.RawValuePair)
	}
}

func migratedSchemaReadinessDatabase(t testing.TB, prefix string, throughVersion int64) (*sql.DB, *pgxpool.Pool) {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.NewMigrationDatabaseT(t, prefix)
	db := openSchemaReadinessSQL(t, testDB.DSN)
	var err error
	if throughVersion == 0 {
		_, err = postgres.Apply(context.Background(), db, dbmigrations.Source())
	} else {
		_, err = postgres.ApplyThrough(context.Background(), db, dbmigrations.Source(), throughVersion)
	}
	if err != nil {
		t.Fatalf("migrate schema readiness database: %v", err)
	}
	return db, openSchemaReadinessPool(t, testDB.DSN)
}

func openSchemaReadinessSQL(t testing.TB, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open sql database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func openSchemaReadinessPool(t testing.TB, dsn string) *pgxpool.Pool {
	t.Helper()

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open pgx pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func requireSchemaReadinessDiagnostic(t testing.TB, err error, reasonCode string) {
	t.Helper()

	var diagnosticsErr *config.DiagnosticsError
	if !errors.As(err, &diagnosticsErr) {
		t.Fatalf("expected diagnostics error, got %T %[1]v", err)
	}
	for _, diagnostic := range diagnosticsErr.Diagnostics {
		if diagnostic.ReasonCode == reasonCode {
			return
		}
	}
	t.Fatalf("missing reason_code=%q in %#v", reasonCode, diagnosticsErr.Diagnostics)
}

func requireMigrationRemediation(t testing.TB, err error) postgres.MigrationRemediationReport {
	t.Helper()

	var remediationErr *postgres.MigrationRemediationError
	if !errors.As(err, &remediationErr) {
		t.Fatalf("expected migration remediation error, got %T %[1]v", err)
	}
	if len(remediationErr.Report.Findings) != 1 {
		t.Fatalf("unexpected remediation report: %#v", remediationErr.Report)
	}
	return remediationErr.Report
}
