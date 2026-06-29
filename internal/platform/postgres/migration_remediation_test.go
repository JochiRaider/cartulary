package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestMigrationPreflightIncidentLifecycleV36PassesValidRows(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	db := postgresHarness.MigrationDatabaseT(t, "migration-preflight-v36-valid", "up-to", "35")
	ctx := context.Background()
	actorID := seedMigrationRemediationUser(t, db, "v36-valid")
	dropLegacyIncidentStatusConstraint(t, db)

	insertMigrationRemediationIncident(t, db, "00000000-0000-0000-0000-000000036001", "IR-V36-ACTIVE", "active", nil, actorID, nil)
	closedAt := "2026-04-17T12:00:00Z"
	insertMigrationRemediationIncident(t, db, "00000000-0000-0000-0000-000000036002", "IR-V36-CLOSED", "closed", &closedAt, actorID, nil)

	if _, err := postgres.Migrate(ctx, db, dbmigrations.Source(), "up-to", "36"); err != nil {
		t.Fatalf("valid lifecycle rows should cross v36: %v", err)
	}
}

func TestMigrationPreflightIncidentLifecycleV36ReportsFailures(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	db := postgresHarness.MigrationDatabaseT(t, "migration-preflight-v36-invalid", "up-to", "35")
	ctx := context.Background()
	actorID := seedMigrationRemediationUser(t, db, "v36-invalid")
	dropLegacyIncidentStatusConstraint(t, db)

	insertMigrationRemediationIncident(t, db, "00000000-0000-0000-0000-000000036101", "IR-V36-UNKNOWN", "triaged", nil, actorID, nil)
	activeClosedAt := "2026-04-17T12:00:00Z"
	insertMigrationRemediationIncident(t, db, "00000000-0000-0000-0000-000000036102", "IR-V36-ACTIVE-CLOSED", "active", &activeClosedAt, actorID, nil)
	insertMigrationRemediationIncident(t, db, "00000000-0000-0000-0000-000000036103", "IR-V36-CLOSED-OPEN", "closed", nil, actorID, nil)

	_, err := postgres.Migrate(ctx, db, dbmigrations.Source(), "up-to", "36")
	var remediationErr *postgres.MigrationRemediationError
	if !errors.As(err, &remediationErr) {
		t.Fatalf("expected typed migration remediation error, got %T %[1]v", err)
	}
	report := remediationErr.Report
	if report.SchemaID != "cartulary.migration_remediation_report.v1" ||
		report.Boundary != "incident_lifecycle_v36" ||
		report.FromVersion != 35 ||
		report.ToVersion != 36 ||
		len(report.Findings) != 3 {
		t.Fatalf("unexpected remediation report: %#v", report)
	}
	reasonCodes := []string{
		report.Findings[0].ReasonCode,
		report.Findings[1].ReasonCode,
		report.Findings[2].ReasonCode,
	}
	wantReasonCodes := []string{"unknown_status", "active_with_closed_at", "closed_without_closed_at"}
	for index, want := range wantReasonCodes {
		if reasonCodes[index] != want {
			t.Fatalf("unexpected finding order/reason codes: got %#v want %#v", reasonCodes, wantReasonCodes)
		}
	}
}

func TestMigrationIncidentMetadataV40NormalizesLegacyValues(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	db := postgresHarness.MigrationDatabaseT(t, "migration-v40-normalize", "up-to", "39")
	ctx := context.Background()
	actorID := seedMigrationRemediationUser(t, db, "v40-normalize")

	insertMigrationRemediationIncident(t, db, "00000000-0000-0000-0000-000000040001", "IR-V40-NORMALIZE", "active", nil, actorID, map[string]any{
		"description":               "  First line\r\nSecond line  ",
		"severity":                  "   ",
		"tlp":                       " amber ",
		"current_phase":             " containment ",
		"primary_external_case_ref": " CASE-40 ",
	})

	if _, err := postgres.Migrate(ctx, db, dbmigrations.Source(), "up-to", "40"); err != nil {
		t.Fatalf("legacy metadata should normalize through v40: %v", err)
	}

	row := db.QueryRowContext(ctx, `
SELECT description, severity, tlp, current_phase, primary_external_case_ref
  FROM incidents
 WHERE id = '00000000-0000-0000-0000-000000040001'
`)
	var description, severity, tlp, currentPhase, primaryExternalCaseRef sql.NullString
	if err := row.Scan(&description, &severity, &tlp, &currentPhase, &primaryExternalCaseRef); err != nil {
		t.Fatalf("query normalized incident metadata: %v", err)
	}
	if !description.Valid || description.String != "First line\nSecond line" {
		t.Fatalf("unexpected normalized description: %#v", description)
	}
	if severity.Valid {
		t.Fatalf("whitespace severity should clear to null: %#v", severity)
	}
	if !tlp.Valid || tlp.String != "TLP:AMBER" {
		t.Fatalf("unexpected normalized tlp: %#v", tlp)
	}
	if !currentPhase.Valid || currentPhase.String != "containment" {
		t.Fatalf("unexpected normalized current_phase: %#v", currentPhase)
	}
	if !primaryExternalCaseRef.Valid || primaryExternalCaseRef.String != "CASE-40" {
		t.Fatalf("unexpected normalized primary_external_case_ref: %#v", primaryExternalCaseRef)
	}
}

func TestMigrationIncidentMetadataV40ReportsInvalidRows(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	db := postgresHarness.MigrationDatabaseT(t, "migration-v40-invalid", "up-to", "39")
	ctx := context.Background()
	actorID := seedMigrationRemediationUser(t, db, "v40-invalid")

	insertMigrationRemediationIncident(t, db, "00000000-0000-0000-0000-000000040101", "IR-V40-UNKNOWN-TLP", "active", nil, actorID, map[string]any{
		"tlp": "purple",
	})
	insertMigrationRemediationIncident(t, db, "00000000-0000-0000-0000-000000040102", "IR-V40-BAD-SEVERITY", "active", nil, actorID, map[string]any{
		"severity": strings.Repeat("a", 129),
	})
	insertMigrationRemediationIncident(t, db, "00000000-0000-0000-0000-000000040103", "IR-V40-BAD-DESCRIPTION", "active", nil, actorID, map[string]any{
		"description": "bad" + string(rune(1)),
	})

	_, err := postgres.Migrate(ctx, db, dbmigrations.Source(), "up-to", "40")
	if err == nil {
		t.Fatal("expected invalid v40 metadata to fail migration")
	}
	errText := err.Error()
	required := []string{
		"cartulary.migration_remediation_report.v1",
		"incident_metadata_canonicalization_v40",
		"unknown_tlp",
		"invalid_severity",
		"invalid_description",
	}
	for _, needle := range required {
		if !strings.Contains(errText, needle) {
			t.Fatalf("expected migration error to include %q, got %v", needle, err)
		}
	}
}

func seedMigrationRemediationUser(t testing.TB, db *sql.DB, suffix string) string {
	t.Helper()

	userID := "00000000-0000-0000-0000-000000000401"
	_, err := db.ExecContext(context.Background(), `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'Migration Remediation User', 'test-hash', false, true, true)
`, userID, "migration-remediation-"+suffix+"@example.test")
	if err != nil {
		t.Fatalf("seed migration remediation user: %v", err)
	}
	return userID
}

func dropLegacyIncidentStatusConstraint(t testing.TB, db *sql.DB) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `ALTER TABLE incidents DROP CONSTRAINT IF EXISTS incidents_status_check`); err != nil {
		t.Fatalf("drop legacy incident status constraint: %v", err)
	}
}

func insertMigrationRemediationIncident(t testing.TB, db *sql.DB, incidentID string, incidentKey string, status string, closedAt *string, actorID string, metadata map[string]any) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), `
INSERT INTO incidents (
    id,
    incident_key,
    incident_key_canonical,
    title,
    description,
    status,
    severity,
    tlp,
    current_phase,
    primary_external_case_ref,
    created_by_user_id,
    updated_by_user_id,
    closed_at
)
VALUES ($1, $2, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10, $11)
`, incidentID,
		incidentKey,
		"Migration Remediation "+incidentKey,
		metadataValue(metadata, "description"),
		status,
		metadataValue(metadata, "severity"),
		metadataValue(metadata, "tlp"),
		metadataValue(metadata, "current_phase"),
		metadataValue(metadata, "primary_external_case_ref"),
		actorID,
		closedAt,
	)
	if err != nil {
		t.Fatalf("insert migration remediation incident %s: %v", incidentKey, err)
	}
}

func metadataValue(metadata map[string]any, key string) any {
	if metadata == nil {
		return nil
	}
	return metadata[key]
}
