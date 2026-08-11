package database_migrations_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	postgres "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestEnsureSchemaReadyAllowsCurrentHead(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "schema-ready-head")
	pool := openSchemaReadinessPool(t, testDB.DSN)

	if err := postgres.EnsureSchemaReady(context.Background(), pool, canonicalMigrationSource(t)); err != nil {
		t.Fatalf("current head schema should be ready: %v", err)
	}
}

func TestEnsureSchemaReadyRejectsEmptyDatabase(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	migrationDB := postgresHarness.MigrationDatabaseThroughT(t, 1)
	if err := migrationDB.RollbackThrough(context.Background(), 0); err != nil {
		t.Fatalf("rollback empty schema fixture: %v", err)
	}
	databaseName := migrationScratchDatabaseName(t, migrationDB.SQL())
	pool := openSchemaReadinessPoolForDatabase(t, postgresHarness.AdminDSN(), databaseName)

	err := postgres.EnsureSchemaReady(context.Background(), pool, canonicalMigrationSource(t))
	requireSchemaReadinessDiagnostic(t, err, "schema_migration_required")
}

func TestEnsureSchemaReadyRejectsBehindCurrentLine(t *testing.T) {
	_, pool := migratedSchemaReadinessDatabase(t, "schema-ready-behind", 1)

	err := postgres.EnsureSchemaReady(context.Background(), pool, canonicalMigrationSource(t))
	requireSchemaReadinessDiagnostic(t, err, "schema_migration_required")
}

func TestEnsureSchemaReadyRejectsAheadCurrentLine(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "schema-ready-ahead")
	db := openSchemaReadinessSQL(t, testDB.DSN)
	pool := openSchemaReadinessPool(t, testDB.DSN)

	if _, err := db.ExecContext(context.Background(), `INSERT INTO goose_db_version (version_id, is_applied) VALUES (30, true)`); err != nil {
		t.Fatalf("seed ahead migration version: %v", err)
	}

	err := postgres.EnsureSchemaReady(context.Background(), pool, canonicalMigrationSource(t))
	requireSchemaReadinessDiagnostic(t, err, "schema_version_ahead")
}

func TestEnsureSchemaReadyRejectsHistoricalLineAboveHead(t *testing.T) {
	db, pool := migratedSchemaReadinessDatabase(t, "schema-ready-historical-above-head", 0)
	if _, err := db.ExecContext(context.Background(), `
DROP TABLE schema_migration_lineage;
INSERT INTO goose_db_version (version_id, is_applied) VALUES (30, true);
`); err != nil {
		t.Fatalf("seed historical migration line above head: %v", err)
	}

	report := requireMigrationRemediation(t, postgres.EnsureSchemaReady(context.Background(), pool, canonicalMigrationSource(t)))
	if report.Boundary != expectedCanonicalLineageBoundary || report.FromVersion != 30 || report.ToVersion != 29 {
		t.Fatalf("unexpected remediation report: %#v", report)
	}
	finding := report.Findings[0]
	if finding.ReasonCode != "historical_migration_lineage" || finding.RawValue != nil || finding.RawValuePair.LineageTablePresent {
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

	report := requireMigrationRemediation(t, postgres.EnsureSchemaReady(context.Background(), pool, canonicalMigrationSource(t)))
	finding := report.Findings[0]
	if finding.RawValue == nil || *finding.RawValue != "cartulary.legacy_line.v1" {
		t.Fatalf("expected observed wrong lineage in raw_value: %#v", finding)
	}
	if finding.RawValuePair.ExpectedLineageID != expectedCanonicalLineageID || !finding.RawValuePair.LineageTablePresent {
		t.Fatalf("unexpected remediation facts: %#v", finding.RawValuePair)
	}
}

func TestEnsureSchemaReadyRejectsZeroOnlyHistory(t *testing.T) {
	db, pool := migratedSchemaReadinessDatabase(t, "schema-ready-zero-only", 1)
	if _, err := db.ExecContext(context.Background(), `
DELETE FROM goose_db_version WHERE version_id <> 0;
DROP TABLE schema_migration_lineage;
`); err != nil {
		t.Fatalf("seed zero-only history: %v", err)
	}
	requireSchemaReadinessDiagnostic(
		t,
		postgres.EnsureSchemaReady(context.Background(), pool, canonicalMigrationSource(t)),
		"schema_migration_required",
	)
}

func TestEnsureSchemaReadyRejectsInvalidHistoryMatrix(t *testing.T) {
	tests := []struct {
		name    string
		through int64
		mutate  string
	}{
		{name: "duplicate", through: 2, mutate: `INSERT INTO goose_db_version (version_id, is_applied) VALUES (2, true)`},
		{name: "false", through: 1, mutate: `UPDATE goose_db_version SET is_applied = false WHERE version_id = 1`},
		{name: "gap", through: 3, mutate: `DELETE FROM goose_db_version WHERE version_id = 2`},
		{name: "out of order", through: 2, mutate: `UPDATE goose_db_version SET version_id = CASE version_id WHEN 1 THEN 2 WHEN 2 THEN 1 ELSE version_id END WHERE version_id IN (1, 2)`},
		{name: "lineage without history", through: 1, mutate: `DELETE FROM goose_db_version WHERE version_id <> 0`},
		{name: "corruption before wrong lineage", through: 3, mutate: `
DELETE FROM goose_db_version WHERE version_id = 2;
DELETE FROM schema_migration_lineage;
INSERT INTO schema_migration_lineage (lineage_id, description)
VALUES ('cartulary.legacy_line.v1', 'legacy test line');
`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, pool := migratedSchemaReadinessDatabase(t, "schema-ready-invalid-"+test.name, test.through)
			if _, err := db.ExecContext(context.Background(), test.mutate); err != nil {
				t.Fatalf("seed invalid migration history: %v", err)
			}
			requireSchemaReadinessDiagnostic(
				t,
				postgres.EnsureSchemaReady(context.Background(), pool, canonicalMigrationSource(t)),
				"schema_migration_history_invalid",
			)
		})
	}
}

func TestEnsureSchemaReadyRejectsMixedLineage(t *testing.T) {
	db, pool := migratedSchemaReadinessDatabase(t, "schema-ready-mixed-lineage", 1)
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO schema_migration_lineage (lineage_id, description)
VALUES ('cartulary.legacy_line.v1', 'legacy test line');
`); err != nil {
		t.Fatalf("seed mixed migration lineage: %v", err)
	}
	report := requireMigrationRemediation(t, postgres.EnsureSchemaReady(context.Background(), pool, canonicalMigrationSource(t)))
	facts := report.Findings[0].RawValuePair
	if len(facts.ObservedLineageIDs) != 2 || facts.ObservedLineageIDs[0] != "cartulary.legacy_line.v1" || facts.ObservedLineageIDs[1] != expectedCanonicalLineageID {
		t.Fatalf("unexpected mixed lineage report: %#v", report)
	}
}

func migratedSchemaReadinessDatabase(t testing.TB, prefix string, throughVersion int64) (*sql.DB, *pgxpool.Pool) {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	var migrationDB *pgtest.MigrationDatabase
	if throughVersion == 0 {
		migrationDB = postgresHarness.MigrationDatabaseT(t)
	} else {
		migrationDB = postgresHarness.MigrationDatabaseThroughT(t, throughVersion)
	}
	db := migrationDB.SQL()
	databaseName := migrationScratchDatabaseName(t, db)
	return db, openSchemaReadinessPoolForDatabase(t, postgresHarness.AdminDSN(), databaseName)
}

func migrationScratchDatabaseName(t testing.TB, db *sql.DB) string {
	t.Helper()

	var databaseName string
	if err := db.QueryRowContext(context.Background(), `SELECT current_database()`).Scan(&databaseName); err != nil {
		t.Fatalf("resolve migration scratch database name: %v", err)
	}
	return databaseName
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

func openSchemaReadinessPoolForDatabase(t testing.TB, adminDSN string, databaseName string) *pgxpool.Pool {
	t.Helper()

	config, err := pgxpool.ParseConfig(adminDSN)
	if err != nil {
		t.Fatalf("parse postgres harness admin binding: %v", err)
	}
	config.ConnConfig.Database = databaseName
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatalf("open migration scratch pgx pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func requireSchemaReadinessDiagnostic(t testing.TB, err error, reasonCode string) {
	t.Helper()
	requireExternalMigrationFailureReason(t, err, reasonCode)
}
