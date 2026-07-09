package postgres_test

import (
	"context"
	"errors"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestMigrationLineagePreflightAllowsCurrentLine(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	db := postgresHarness.MigrationDatabaseT(t, "migration-lineage-current", "up-to", "1")

	if _, err := postgres.Migrate(context.Background(), db, dbmigrations.Source(), "up-to", "2"); err != nil {
		t.Fatalf("current-line partial database should continue migrating: %v", err)
	}
}

func TestMigrationLineagePreflightRejectsHistoricalLine(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	db := postgresHarness.MigrationDatabaseT(t, "migration-lineage-historical", "up-to", "1")
	if _, err := db.ExecContext(context.Background(), `DROP TABLE schema_migration_lineage`); err != nil {
		t.Fatalf("simulate historical migration line: %v", err)
	}

	_, err := postgres.Migrate(context.Background(), db, dbmigrations.Source(), "up-to", "2")
	var remediationErr *postgres.MigrationRemediationError
	if !errors.As(err, &remediationErr) {
		t.Fatalf("expected typed migration remediation error, got %T %[1]v", err)
	}
	report := remediationErr.Report
	if report.SchemaID != "cartulary.migration_remediation_report.v1" ||
		report.Boundary != dbmigrations.LineageBoundary ||
		report.FromVersion != 1 ||
		report.ToVersion != 2 ||
		len(report.Findings) != 1 {
		t.Fatalf("unexpected remediation report: %#v", report)
	}
	finding := report.Findings[0]
	if finding.Field != "schema_migration_lineage" ||
		finding.ReasonCode != "historical_migration_lineage" ||
		finding.RawValue == nil ||
		*finding.RawValue != dbmigrations.LineageID {
		t.Fatalf("unexpected remediation finding: %#v", finding)
	}
}
