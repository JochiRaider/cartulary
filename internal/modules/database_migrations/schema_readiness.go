package database_migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/config"
)

// LedgerReader is the complete read-only database capability required by
// migration readiness and evidence collection.
type LedgerReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// EnsureSchemaReady verifies that an already-open database matches the
// repository migration source before runtime subsystems touch schema objects.
func EnsureSchemaReady(ctx context.Context, pool *pgxpool.Pool, source MigrationSource) error {
	if ctx == nil {
		return errNilMigrateContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if pool == nil {
		return nil
	}

	source = normalizeMigrationSource(source)
	empty, err := migrationSourceEmpty(source)
	if err != nil {
		return fmt.Errorf("inspect migration directory: %w", err)
	}
	if empty {
		return nil
	}

	repositoryHeadVersion, err := migrationSourceHeadVersion(source)
	if err != nil {
		return fmt.Errorf("inspect migration source head: %w", err)
	}

	currentVersion, metadataPresent, err := currentGooseVersionPGX(ctx, pool)
	if err != nil {
		return fmt.Errorf("inspect migration version: %w", err)
	}
	if !metadataPresent || currentVersion == 0 {
		return schemaReadinessDiagnostics("schema_migration_required", fmt.Sprintf("database has no applied migration metadata; run migrate up before server startup so schema reaches repository migration head %d", repositoryHeadVersion))
	}

	if source.ExpectedLineageID != "" {
		lineageState, err := inspectMigrationLineagePGX(ctx, pool)
		if err != nil {
			return fmt.Errorf("inspect migration lineage: %w", err)
		}
		if !lineageState.HasExpected(source.ExpectedLineageID) {
			return &MigrationRemediationError{
				Report: migrationLineageRemediationReport(source, lineageState, currentVersion, repositoryHeadVersion, repositoryHeadVersion),
			}
		}
	}

	switch {
	case currentVersion < repositoryHeadVersion:
		return schemaReadinessDiagnostics("schema_migration_required", fmt.Sprintf("database schema version %d is behind repository migration head %d; run migrate up before server startup", currentVersion, repositoryHeadVersion))
	case currentVersion > repositoryHeadVersion:
		return schemaReadinessDiagnostics("schema_version_ahead", fmt.Sprintf("database schema version %d is ahead of repository migration head %d; start a server built from the matching repository version or restore a compatible database", currentVersion, repositoryHeadVersion))
	default:
		return nil
	}
}

func currentGooseVersionPGX(ctx context.Context, reader LedgerReader) (int64, bool, error) {
	var tableExists bool
	if err := reader.QueryRow(ctx, `SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&tableExists); err != nil {
		return 0, false, err
	}
	if !tableExists {
		return 0, false, nil
	}

	var version int64
	if err := reader.QueryRow(ctx, `SELECT COALESCE(MAX(version_id), 0)::bigint FROM goose_db_version WHERE is_applied = true`).Scan(&version); err != nil {
		return 0, true, err
	}
	return version, true, nil
}

func inspectMigrationLineagePGX(ctx context.Context, reader LedgerReader) (migrationLineageState, error) {
	var tableExists bool
	if err := reader.QueryRow(ctx, `SELECT to_regclass('public.schema_migration_lineage') IS NOT NULL`).Scan(&tableExists); err != nil {
		return migrationLineageState{}, err
	}
	if !tableExists {
		return migrationLineageState{TablePresent: false}, nil
	}

	rows, err := reader.Query(ctx, `SELECT lineage_id FROM schema_migration_lineage ORDER BY lineage_id ASC`)
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

func schemaReadinessDiagnostics(reasonCode string, message string) error {
	return config.NewDiagnosticsError(config.Diagnostic{
		Path:       "database.schema_version",
		ReasonCode: reasonCode,
		Message:    message,
	})
}
