package postgres_test

import (
	"context"
	"encoding/json"
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
		finding.RawValue != nil ||
		finding.RawValuePair["expected_lineage_id"] != dbmigrations.LineageID ||
		finding.RawValuePair["lineage_table_present"] != false {
		t.Fatalf("unexpected remediation finding: %#v", finding)
	}
	if got := jsonStringSlice(t, finding.RawValuePair["observed_lineage_ids"]); len(got) != 0 {
		t.Fatalf("unexpected observed lineages: %#v", got)
	}
}

func TestMigrationLineagePreflightReportsObservedWrongLineage(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	db := postgresHarness.MigrationDatabaseT(t, "migration-lineage-wrong", "up-to", "1")
	if _, err := db.ExecContext(context.Background(), `
DELETE FROM schema_migration_lineage;
INSERT INTO schema_migration_lineage (lineage_id, description)
VALUES ('cartulary.legacy_line.v1', 'legacy test line');
`); err != nil {
		t.Fatalf("simulate wrong migration line: %v", err)
	}

	_, err := postgres.Migrate(context.Background(), db, dbmigrations.Source(), "up-to", "2")
	var remediationErr *postgres.MigrationRemediationError
	if !errors.As(err, &remediationErr) {
		t.Fatalf("expected typed migration remediation error, got %T %[1]v", err)
	}
	finding := remediationErr.Report.Findings[0]
	if finding.RawValue == nil || *finding.RawValue != "cartulary.legacy_line.v1" {
		t.Fatalf("unexpected raw observed lineage: %#v", finding)
	}
	if got := jsonStringSlice(t, finding.RawValuePair["observed_lineage_ids"]); len(got) != 1 || got[0] != "cartulary.legacy_line.v1" {
		t.Fatalf("unexpected observed lineages: %#v", got)
	}
	if finding.RawValuePair["expected_lineage_id"] != dbmigrations.LineageID ||
		finding.RawValuePair["lineage_table_present"] != true {
		t.Fatalf("unexpected remediation facts: %#v", finding.RawValuePair)
	}
}

func jsonStringSlice(t testing.TB, value any) []string {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}
	var items []string
	if err := json.Unmarshal(encoded, &items); err != nil {
		t.Fatalf("unmarshal string slice from %#v: %v", value, err)
	}
	return items
}
