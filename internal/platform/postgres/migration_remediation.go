package postgres

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

func runMigrationPreflights(ctx context.Context, db *sql.DB, source MigrationSource, command string, args ...string) error {
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

	hasLineage, err := hasExpectedMigrationLineage(ctx, db, source.ExpectedLineageID)
	if err != nil {
		return fmt.Errorf("inspect migration lineage: %w", err)
	}
	if hasLineage {
		return nil
	}

	targetVersion, err := targetMigrationVersion(source, currentVersion, command, args)
	if err != nil {
		return fmt.Errorf("inspect migration target: %w", err)
	}
	boundary := source.ExpectedLineageBoundary
	if boundary == "" {
		boundary = defaultMigrationLineageBoundary
	}
	rawLineageID := source.ExpectedLineageID
	return &MigrationRemediationError{
		Report: MigrationRemediationReport{
			SchemaID:    migrationRemediationSchemaID,
			Boundary:    boundary,
			FromVersion: currentVersion,
			ToVersion:   targetVersion,
			Findings: []MigrationRemediationFinding{
				{
					Field:           "schema_migration_lineage",
					RawValue:        &rawLineageID,
					ReasonCode:      historicalMigrationLineageReason,
					RemediationHint: "Reset this database or move data through an explicit export/import path before applying the production DDL rebaseline.",
				},
			},
		},
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

func hasExpectedMigrationLineage(ctx context.Context, db *sql.DB, lineageID string) (bool, error) {
	var tableExists bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.schema_migration_lineage') IS NOT NULL`).Scan(&tableExists); err != nil {
		return false, err
	}
	if !tableExists {
		return false, nil
	}

	var hasLineage bool
	if err := db.QueryRowContext(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM schema_migration_lineage
     WHERE lineage_id = $1
)
`, lineageID).Scan(&hasLineage); err != nil {
		return false, err
	}
	return hasLineage, nil
}

func targetMigrationVersion(source MigrationSource, currentVersion int64, command string, args []string) (int64, error) {
	switch command {
	case "up":
		return migrationSourceHeadVersion(source)
	case "up-by-one":
		return currentVersion + 1, nil
	case "up-to":
		if len(args) == 0 {
			return 0, nil
		}
		target, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return 0, nil
		}
		return target, nil
	default:
		return 0, nil
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
