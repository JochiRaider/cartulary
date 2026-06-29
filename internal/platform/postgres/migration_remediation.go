package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	migrationRemediationSchemaID = "cartulary.migration_remediation_report.v1"
	incidentLifecycleBoundaryV36 = "incident_lifecycle_v36"
)

type MigrationRemediationReport struct {
	SchemaID    string                        `json:"schema_id"`
	Boundary    string                        `json:"boundary"`
	FromVersion int64                         `json:"from_version"`
	ToVersion   int64                         `json:"to_version"`
	Findings    []MigrationRemediationFinding `json:"findings"`
}

type MigrationRemediationFinding struct {
	IncidentID      string         `json:"incident_id"`
	Field           string         `json:"field"`
	RawValue        *string        `json:"raw_value,omitempty"`
	RawValuePair    map[string]any `json:"raw_value_pair,omitempty"`
	ReasonCode      string         `json:"reason_code"`
	RemediationHint string         `json:"remediation_hint"`
}

type MigrationRemediationError struct {
	Report MigrationRemediationReport
}

func (err *MigrationRemediationError) Error() string {
	if err == nil {
		return "migration remediation failed"
	}
	return err.ReportJSON()
}

func (err *MigrationRemediationError) ReportJSON() string {
	if err == nil {
		return ""
	}
	encoded, marshalErr := json.Marshal(err.Report)
	if marshalErr != nil {
		return fmt.Sprintf("%s: marshal report: %v", migrationRemediationSchemaID, marshalErr)
	}
	return string(encoded)
}

func runMigrationPreflights(ctx context.Context, db *sql.DB, command string, args ...string) error {
	if db == nil {
		return nil
	}
	currentVersion, err := currentGooseVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("inspect migration version: %w", err)
	}
	if !crossesMigrationBoundary(currentVersion, command, args, 36) {
		return nil
	}
	return preflightIncidentLifecycleV36(ctx, db, currentVersion)
}

func currentGooseVersion(ctx context.Context, db *sql.DB) (int64, error) {
	var tableExists bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&tableExists); err != nil {
		return 0, err
	}
	if !tableExists {
		return 0, nil
	}

	var version int64
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version_id), 0)::bigint FROM goose_db_version WHERE is_applied = true`).Scan(&version); err != nil {
		return 0, err
	}
	return version, nil
}

func crossesMigrationBoundary(currentVersion int64, command string, args []string, boundaryVersion int64) bool {
	switch command {
	case "up":
		return currentVersion < boundaryVersion
	case "up-by-one":
		return currentVersion == boundaryVersion-1
	case "up-to":
		if len(args) == 0 {
			return false
		}
		target, err := strconv.ParseInt(args[0], 10, 64)
		return err == nil && currentVersion < boundaryVersion && target >= boundaryVersion
	default:
		return false
	}
}

func preflightIncidentLifecycleV36(ctx context.Context, db *sql.DB, currentVersion int64) error {
	var tableExists bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.incidents') IS NOT NULL`).Scan(&tableExists); err != nil {
		return err
	}
	if !tableExists {
		return nil
	}

	rows, err := db.QueryContext(ctx, `
SELECT id::text,
       status,
       closed_at::text,
       CASE
           WHEN status NOT IN ('active', 'closed') THEN 'unknown_status'
           WHEN status = 'active' AND closed_at IS NOT NULL THEN 'active_with_closed_at'
           WHEN status = 'closed' AND closed_at IS NULL THEN 'closed_without_closed_at'
       END AS reason_code
  FROM incidents
 WHERE status NOT IN ('active', 'closed')
    OR (status = 'active' AND closed_at IS NOT NULL)
    OR (status = 'closed' AND closed_at IS NULL)
 ORDER BY id::text ASC, reason_code ASC
`)
	if err != nil {
		return err
	}
	defer rows.Close()

	findings := []MigrationRemediationFinding{}
	for rows.Next() {
		var incidentID string
		var status string
		var closedAt sql.NullString
		var reasonCode string
		if err := rows.Scan(&incidentID, &status, &closedAt, &reasonCode); err != nil {
			return err
		}
		findings = append(findings, lifecycleV36Finding(incidentID, status, closedAt, reasonCode))
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(findings) == 0 {
		return nil
	}

	return &MigrationRemediationError{
		Report: MigrationRemediationReport{
			SchemaID:    migrationRemediationSchemaID,
			Boundary:    incidentLifecycleBoundaryV36,
			FromVersion: currentVersion,
			ToVersion:   36,
			Findings:    findings,
		},
	}
}

func lifecycleV36Finding(incidentID string, status string, closedAt sql.NullString, reasonCode string) MigrationRemediationFinding {
	switch reasonCode {
	case "unknown_status":
		return MigrationRemediationFinding{
			IncidentID:      incidentID,
			Field:           "status",
			RawValue:        stringPointer(status),
			ReasonCode:      reasonCode,
			RemediationHint: "Set status to active with closed_at null, or closed with closed_at populated, before rerunning migration.",
		}
	case "active_with_closed_at":
		return MigrationRemediationFinding{
			IncidentID:      incidentID,
			Field:           "status_closed_at",
			RawValuePair:    map[string]any{"status": status, "closed_at": nullableStringValue(closedAt)},
			ReasonCode:      reasonCode,
			RemediationHint: "Clear closed_at for active incidents before rerunning migration.",
		}
	case "closed_without_closed_at":
		return MigrationRemediationFinding{
			IncidentID:      incidentID,
			Field:           "status_closed_at",
			RawValuePair:    map[string]any{"status": status, "closed_at": nullableStringValue(closedAt)},
			ReasonCode:      reasonCode,
			RemediationHint: "Populate closed_at for closed incidents before rerunning migration.",
		}
	default:
		return MigrationRemediationFinding{
			IncidentID:      incidentID,
			Field:           "status",
			RawValue:        stringPointer(status),
			ReasonCode:      reasonCode,
			RemediationHint: "Repair incident lifecycle state before rerunning migration.",
		}
	}
}

func nullableStringValue(value sql.NullString) any {
	if !value.Valid {
		return nil
	}
	return value.String
}

func stringPointer(value string) *string {
	return &value
}

var _ error = (*MigrationRemediationError)(nil)
