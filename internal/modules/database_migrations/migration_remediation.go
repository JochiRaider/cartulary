package database_migrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
)

const (
	migrationRemediationSchemaID     = "cartulary.migration_remediation_report.v1"
	defaultMigrationLineageBoundary  = "migration_lineage"
	historicalMigrationLineageReason = "historical_migration_lineage"
	historicalMigrationLineageHint   = "Reset this database, or move data through an explicit owner-approved export/import path before applying the production DDL rebaseline."
)

type MigrationRemediationReport struct {
	SchemaID    string                        `json:"schema_id"`
	Boundary    string                        `json:"boundary"`
	FromVersion int64                         `json:"from_version"`
	ToVersion   int64                         `json:"to_version"`
	Findings    []MigrationRemediationFinding `json:"findings"`
}

type MigrationRemediationFinding struct {
	IncidentID      string         `json:"incident_id,omitempty"`
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

func runMigrationPreflights(ctx context.Context, db *sql.DB, source MigrationSource, operation migrationOperation, targetVersion int64) error {
	if db == nil || source.ExpectedLineageID == "" {
		return nil
	}
	currentVersion, err := currentGooseVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("inspect migration version: %w", err)
	}
	if currentVersion == 0 {
		return nil
	}

	lineageState, err := inspectMigrationLineage(ctx, db)
	if err != nil {
		return fmt.Errorf("inspect migration lineage: %w", err)
	}
	if lineageState.HasExpected(source.ExpectedLineageID) {
		return nil
	}

	repositoryHeadVersion, err := migrationSourceHeadVersion(source)
	if err != nil {
		return fmt.Errorf("inspect migration source head: %w", err)
	}
	if operation == migrationOperationApply {
		targetVersion = repositoryHeadVersion
	}

	return &MigrationRemediationError{
		Report: migrationLineageRemediationReport(source, lineageState, currentVersion, targetVersion, repositoryHeadVersion),
	}
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

type migrationLineageState struct {
	TablePresent bool
	ObservedIDs  []string
}

func (state migrationLineageState) HasExpected(lineageID string) bool {
	for _, observed := range state.ObservedIDs {
		if observed == lineageID {
			return true
		}
	}
	return false
}

func inspectMigrationLineage(ctx context.Context, db *sql.DB) (migrationLineageState, error) {
	var tableExists bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.schema_migration_lineage') IS NOT NULL`).Scan(&tableExists); err != nil {
		return migrationLineageState{}, err
	}
	if !tableExists {
		return migrationLineageState{TablePresent: false}, nil
	}

	rows, err := db.QueryContext(ctx, `SELECT lineage_id FROM schema_migration_lineage ORDER BY lineage_id ASC`)
	if err != nil {
		return migrationLineageState{}, err
	}
	defer rows.Close()

	state := migrationLineageState{TablePresent: true}
	for rows.Next() {
		var lineageID string
		if err := rows.Scan(&lineageID); err != nil {
			return migrationLineageState{}, err
		}
		state.ObservedIDs = append(state.ObservedIDs, lineageID)
	}
	if err := rows.Err(); err != nil {
		return migrationLineageState{}, err
	}
	return state, nil
}

func migrationLineageRemediationReport(source MigrationSource, state migrationLineageState, currentVersion int64, targetVersion int64, repositoryHeadVersion int64) MigrationRemediationReport {
	boundary := source.ExpectedLineageBoundary
	if boundary == "" {
		boundary = defaultMigrationLineageBoundary
	}

	var rawValue *string
	if len(state.ObservedIDs) > 0 {
		raw := state.ObservedIDs[0]
		rawValue = &raw
	}
	observedIDs := append([]string{}, state.ObservedIDs...)

	return MigrationRemediationReport{
		SchemaID:    migrationRemediationSchemaID,
		Boundary:    boundary,
		FromVersion: currentVersion,
		ToVersion:   targetVersion,
		Findings: []MigrationRemediationFinding{
			{
				Field:    "schema_migration_lineage",
				RawValue: rawValue,
				RawValuePair: map[string]any{
					"current_version":         currentVersion,
					"expected_lineage_id":     source.ExpectedLineageID,
					"lineage_table_present":   state.TablePresent,
					"observed_lineage_ids":    observedIDs,
					"repository_head_version": repositoryHeadVersion,
					"target_version":          targetVersion,
				},
				ReasonCode:      historicalMigrationLineageReason,
				RemediationHint: historicalMigrationLineageHint,
			},
		},
	}
}

var migrationFilenamePattern = regexp.MustCompile(`^([0-9]+)_.+\.sql$`)

func migrationSourceHeadVersion(source MigrationSource) (int64, error) {
	var maxVersion int64
	visit := func(name string) error {
		match := migrationFilenamePattern.FindStringSubmatch(name)
		if match == nil {
			return nil
		}
		version, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return err
		}
		if version > maxVersion {
			maxVersion = version
		}
		return nil
	}

	if source.BaseFS != nil {
		err := fs.WalkDir(source.BaseFS, source.Path, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			return visit(entry.Name())
		})
		return maxVersion, err
	}

	err := filepath.WalkDir(source.Path, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		return visit(entry.Name())
	})
	if err != nil {
		return 0, err
	}
	return maxVersion, nil
}

var _ error = (*MigrationRemediationError)(nil)
