package database_migrations_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	postgres "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestMigrationLineagePreflightAllowsCurrentLine(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	migrationDB := postgresHarness.MigrationDatabaseThroughT(t, 1)
	db := migrationDB.SQL()

	if err := postgres.Apply(context.Background(), db, canonicalMigrationSource(t)); err != nil {
		t.Fatalf("current-line partial database should migrate to head: %v", err)
	}
}

func TestMigrationLineagePreflightRejectsHistoricalLine(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	migrationDB := postgresHarness.MigrationDatabaseThroughT(t, 1)
	db := migrationDB.SQL()
	if _, err := db.ExecContext(context.Background(), `DROP TABLE schema_migration_lineage`); err != nil {
		t.Fatalf("simulate historical migration line: %v", err)
	}

	err := postgres.Apply(context.Background(), db, canonicalMigrationSource(t))
	report := requireMigrationRemediation(t, err)
	if report.SchemaID != "cartulary.migration_remediation_report.v1" ||
		report.Boundary != expectedCanonicalLineageBoundary ||
		report.FromVersion != 1 ||
		report.ToVersion != 61 ||
		len(report.Findings) != 1 {
		t.Fatalf("unexpected remediation report: %#v", report)
	}
	finding := report.Findings[0]
	if finding.Field != "schema_migration_lineage" ||
		finding.ReasonCode != "historical_migration_lineage" ||
		finding.RawValue != nil ||
		finding.RawValuePair.ExpectedLineageID != expectedCanonicalLineageID ||
		finding.RawValuePair.LineageTablePresent {
		t.Fatalf("unexpected remediation finding: %#v", finding)
	}
	if got := finding.RawValuePair.ObservedLineageIDs; len(got) != 0 {
		t.Fatalf("unexpected observed lineages: %#v", got)
	}
}

func TestMigrationLineagePreflightReportsObservedWrongLineage(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	migrationDB := postgresHarness.MigrationDatabaseThroughT(t, 1)
	db := migrationDB.SQL()
	if _, err := db.ExecContext(context.Background(), `
DELETE FROM schema_migration_lineage;
INSERT INTO schema_migration_lineage (lineage_id, description)
VALUES ('cartulary.legacy_line.v1', 'legacy test line');
`); err != nil {
		t.Fatalf("simulate wrong migration line: %v", err)
	}

	err := postgres.Apply(context.Background(), db, canonicalMigrationSource(t))
	finding := requireMigrationRemediation(t, err).Findings[0]
	if finding.RawValue == nil || *finding.RawValue != "cartulary.legacy_line.v1" {
		t.Fatalf("unexpected raw observed lineage: %#v", finding)
	}
	if got := finding.RawValuePair.ObservedLineageIDs; len(got) != 1 || got[0] != "cartulary.legacy_line.v1" {
		t.Fatalf("unexpected observed lineages: %#v", got)
	}
	if finding.RawValuePair.ExpectedLineageID != expectedCanonicalLineageID ||
		!finding.RawValuePair.LineageTablePresent {
		t.Fatalf("unexpected remediation facts: %#v", finding.RawValuePair)
	}
}

func TestMigrationLineagePreflightReportsApplyHeadTarget(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	migrationDB := postgresHarness.MigrationDatabaseThroughT(t, 2)
	db := migrationDB.SQL()
	if _, err := db.ExecContext(context.Background(), `DROP TABLE schema_migration_lineage`); err != nil {
		t.Fatalf("simulate historical migration line: %v", err)
	}

	err := postgres.Apply(context.Background(), db, canonicalMigrationSource(t))
	report := requireMigrationRemediation(t, err)
	if report.FromVersion != 2 || report.ToVersion != 61 {
		t.Fatalf("unexpected apply-head remediation target: %#v", report)
	}
}

func TestProductionPreflightStateMatrix(t *testing.T) {
	tests := []struct {
		name       string
		through    int64
		mutate     string
		wantReason string
	}{
		{name: "ahead", through: 61, mutate: `INSERT INTO goose_db_version (version_id, is_applied) VALUES (62, true)`, wantReason: "schema_version_ahead"},
		{name: "duplicate", through: 2, mutate: `INSERT INTO goose_db_version (version_id, is_applied) VALUES (2, true)`, wantReason: "schema_migration_history_invalid"},
		{name: "false", through: 1, mutate: `UPDATE goose_db_version SET is_applied = false WHERE version_id = 1`, wantReason: "schema_migration_history_invalid"},
		{name: "gap", through: 3, mutate: `DELETE FROM goose_db_version WHERE version_id = 2`, wantReason: "schema_migration_history_invalid"},
		{name: "out of order", through: 2, mutate: `UPDATE goose_db_version SET version_id = CASE version_id WHEN 1 THEN 2 WHEN 2 THEN 1 ELSE version_id END WHERE version_id IN (1, 2)`, wantReason: "schema_migration_history_invalid"},
		{name: "corruption before wrong lineage", through: 3, mutate: `
DELETE FROM goose_db_version WHERE version_id = 2;
DELETE FROM schema_migration_lineage;
INSERT INTO schema_migration_lineage (lineage_id, description)
VALUES ('cartulary.legacy_line.v1', 'legacy test line');
`, wantReason: "schema_migration_history_invalid"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			postgresHarness := pgtest.Start(t)
			migrationDB := postgresHarness.MigrationDatabaseThroughT(t, test.through)
			db := migrationDB.SQL()
			if _, err := db.ExecContext(context.Background(), test.mutate); err != nil {
				t.Fatalf("seed production preflight state: %v", err)
			}

			var beforeCount int
			if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM goose_db_version`).Scan(&beforeCount); err != nil {
				t.Fatalf("count preflight ledger before apply: %v", err)
			}
			err := postgres.Apply(context.Background(), db, canonicalMigrationSource(t))
			requireExternalMigrationFailureReason(t, err, test.wantReason)
			var afterCount int
			if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM goose_db_version`).Scan(&afterCount); err != nil {
				t.Fatalf("count preflight ledger after apply: %v", err)
			}
			if afterCount != beforeCount {
				t.Fatalf("invalid preflight advanced ledger: before=%d after=%d", beforeCount, afterCount)
			}
		})
	}
}

type remediationReportView struct {
	SchemaID    string                   `json:"schema_id"`
	Boundary    string                   `json:"boundary"`
	FromVersion int64                    `json:"from_version"`
	ToVersion   int64                    `json:"to_version"`
	Findings    []remediationFindingView `json:"findings"`
}

type remediationFindingView struct {
	Field        string               `json:"field"`
	RawValue     *string              `json:"raw_value"`
	RawValuePair remediationFactsView `json:"raw_value_pair"`
	ReasonCode   string               `json:"reason_code"`
}

type remediationFactsView struct {
	CurrentVersion        int64    `json:"current_version"`
	ExpectedLineageID     string   `json:"expected_lineage_id"`
	LineageTablePresent   bool     `json:"lineage_table_present"`
	ObservedLineageIDs    []string `json:"observed_lineage_ids"`
	RepositoryHeadVersion int64    `json:"repository_head_version"`
	TargetVersion         int64    `json:"target_version"`
}

func requireMigrationRemediation(t testing.TB, err error) remediationReportView {
	t.Helper()
	var reporter postgres.RemediationReporter
	if !errors.As(err, &reporter) {
		t.Fatalf("expected migration remediation error, got %T %[1]v", err)
	}
	if reporter.ReasonCode() != "historical_migration_lineage" {
		t.Fatalf("unexpected remediation reason: %q", reporter.ReasonCode())
	}
	var report remediationReportView
	if decodeErr := json.Unmarshal([]byte(reporter.RemediationReportJSON()), &report); decodeErr != nil {
		t.Fatalf("decode remediation report: %v", decodeErr)
	}
	if len(report.Findings) != 1 {
		t.Fatalf("unexpected remediation report: %#v", report)
	}
	return report
}

func canonicalMigrationSource(t testing.TB) *postgres.Source {
	t.Helper()
	source, err := dbmigrations.Source()
	if err != nil {
		t.Fatalf("load canonical migration source: %v", err)
	}
	return source
}
