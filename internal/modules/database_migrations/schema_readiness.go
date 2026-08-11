package database_migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LedgerReader is the complete read-only database capability required by
// migration readiness and evidence collection.
type LedgerReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// EnsureSchemaReady verifies that an already-open database matches the
// repository migration source before runtime subsystems touch schema objects.
func EnsureSchemaReady(ctx context.Context, pool *pgxpool.Pool, source *Source) error {
	if ctx == nil {
		return newMigrationFailure(reasonMigrationContextInvalid, errNilMigrateContext)
	}
	if err := ctx.Err(); err != nil {
		return newMigrationFailure(reasonMigrationDatabaseUnavailable, err)
	}
	if err := validateSource(source); err != nil {
		return newMigrationFailure(reasonMigrationSourceInvalid, err)
	}
	if pool == nil {
		return newMigrationFailure(reasonMigrationDatabaseUnavailable, nil)
	}

	snapshot, err := readPGXMigrationState(ctx, pool)
	if err != nil {
		return newMigrationFailure(reasonMigrationDatabaseUnavailable, err)
	}
	classification, err := classifyMigrationState(source, snapshot)
	if err != nil {
		return err
	}
	switch classification.State {
	case migrationStatePristine, migrationStateBehind:
		return newMigrationFailure(reasonSchemaMigrationRequired, nil)
	case migrationStateCurrent:
		return nil
	case migrationStateAhead:
		return newMigrationFailure(reasonSchemaVersionAhead, nil)
	default:
		return newMigrationFailure(reasonSchemaMigrationHistoryInvalid, nil)
	}
}

func readPGXMigrationState(ctx context.Context, reader LedgerReader) (migrationStateSnapshot, error) {
	var snapshot migrationStateSnapshot
	if err := reader.QueryRow(ctx, `SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&snapshot.LedgerTablePresent); err != nil {
		return migrationStateSnapshot{}, fmt.Errorf("inspect migration ledger: %w", err)
	}
	if snapshot.LedgerTablePresent {
		rows, err := reader.Query(ctx, `SELECT version_id, is_applied FROM public.goose_db_version ORDER BY id ASC`)
		if err != nil {
			return migrationStateSnapshot{}, fmt.Errorf("read migration ledger: %w", err)
		}
		for rows.Next() {
			var row migrationLedgerRow
			if err := rows.Scan(&row.Version, &row.IsApplied); err != nil {
				rows.Close()
				return migrationStateSnapshot{}, fmt.Errorf("scan migration ledger: %w", err)
			}
			snapshot.LedgerRows = append(snapshot.LedgerRows, row)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return migrationStateSnapshot{}, fmt.Errorf("iterate migration ledger: %w", err)
		}
		rows.Close()
	}

	if err := reader.QueryRow(ctx, `SELECT to_regclass('public.schema_migration_lineage') IS NOT NULL`).Scan(&snapshot.LineageTablePresent); err != nil {
		return migrationStateSnapshot{}, fmt.Errorf("inspect migration lineage: %w", err)
	}
	if !snapshot.LineageTablePresent {
		return snapshot, nil
	}
	rows, err := reader.Query(ctx, `SELECT lineage_id FROM public.schema_migration_lineage ORDER BY lineage_id ASC`)
	if err != nil {
		return migrationStateSnapshot{}, fmt.Errorf("read migration lineage: %w", err)
	}
	for rows.Next() {
		var lineageID string
		if err := rows.Scan(&lineageID); err != nil {
			rows.Close()
			return migrationStateSnapshot{}, fmt.Errorf("scan migration lineage: %w", err)
		}
		snapshot.LineageIDs = append(snapshot.LineageIDs, lineageID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return migrationStateSnapshot{}, fmt.Errorf("iterate migration lineage: %w", err)
	}
	rows.Close()
	return snapshot, nil
}
