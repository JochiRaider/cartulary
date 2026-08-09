package database_migrations

import (
	"context"
	"database/sql"
	"fmt"
)

type sqlLedgerReader interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type migrationLedgerRow struct {
	Version   int64
	IsApplied bool
}

type migrationStateSnapshot struct {
	LedgerTablePresent  bool
	LedgerRows          []migrationLedgerRow
	LineageTablePresent bool
	LineageIDs          []string
}

type migrationState uint8

const (
	migrationStatePristine migrationState = iota
	migrationStateBehind
	migrationStateCurrent
	migrationStateAhead
)

type migrationClassification struct {
	State          migrationState
	CurrentVersion int64
}

func classifyMigrationState(source Source, snapshot migrationStateSnapshot) (migrationClassification, error) {
	if err := source.validate(); err != nil {
		return migrationClassification{}, newMigrationFailure(reasonMigrationSourceInvalid, err)
	}
	if (!snapshot.LedgerTablePresent && len(snapshot.LedgerRows) != 0) ||
		(!snapshot.LineageTablePresent && len(snapshot.LineageIDs) != 0) {
		return migrationClassification{}, invalidMigrationHistory()
	}

	seenVersions := make(map[int64]struct{}, len(snapshot.LedgerRows))
	expectedVersion := int64(1)
	seenZero := false
	currentVersion := int64(0)
	for index, row := range snapshot.LedgerRows {
		if !row.IsApplied {
			return migrationClassification{}, invalidMigrationHistory()
		}
		if row.Version == 0 {
			// Goose's single leading zero row is a control sentinel, not an
			// authored migration version.
			if seenZero || index != 0 || currentVersion != 0 {
				return migrationClassification{}, invalidMigrationHistory()
			}
			seenZero = true
			continue
		}
		if row.Version < 0 {
			return migrationClassification{}, invalidMigrationHistory()
		}
		if _, duplicate := seenVersions[row.Version]; duplicate {
			return migrationClassification{}, invalidMigrationHistory()
		}
		if row.Version != expectedVersion {
			return migrationClassification{}, invalidMigrationHistory()
		}
		if row.Version <= source.headVersion() && !source.hasVersion(row.Version) {
			return migrationClassification{}, invalidMigrationHistory()
		}
		seenVersions[row.Version] = struct{}{}
		currentVersion = row.Version
		expectedVersion++
	}

	if currentVersion == 0 {
		if snapshot.LineageTablePresent || len(snapshot.LineageIDs) != 0 {
			return migrationClassification{}, invalidMigrationHistory()
		}
		return migrationClassification{State: migrationStatePristine}, nil
	}

	lineageState := migrationLineageState{
		TablePresent: snapshot.LineageTablePresent,
		ObservedIDs:  append([]string{}, snapshot.LineageIDs...),
	}
	if !lineageState.HasExactExpected(source.lineageID) {
		return migrationClassification{}, newMigrationLineageRemediationError(
			source,
			lineageState,
			currentVersion,
			source.headVersion(),
			source.headVersion(),
		)
	}

	switch {
	case currentVersion < source.headVersion():
		return migrationClassification{State: migrationStateBehind, CurrentVersion: currentVersion}, nil
	case currentVersion == source.headVersion():
		return migrationClassification{State: migrationStateCurrent, CurrentVersion: currentVersion}, nil
	default:
		return migrationClassification{State: migrationStateAhead, CurrentVersion: currentVersion}, nil
	}
}

func invalidMigrationHistory() error {
	return newMigrationFailure(reasonSchemaMigrationHistoryInvalid, nil)
}

func readSQLMigrationState(ctx context.Context, reader sqlLedgerReader) (migrationStateSnapshot, error) {
	var snapshot migrationStateSnapshot
	if err := reader.QueryRowContext(ctx, `SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&snapshot.LedgerTablePresent); err != nil {
		return migrationStateSnapshot{}, fmt.Errorf("inspect migration ledger: %w", err)
	}
	if snapshot.LedgerTablePresent {
		rows, err := reader.QueryContext(ctx, `SELECT version_id, is_applied FROM goose_db_version ORDER BY id ASC`)
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
		if err := rows.Close(); err != nil {
			return migrationStateSnapshot{}, fmt.Errorf("close migration ledger: %w", err)
		}
	}

	if err := reader.QueryRowContext(ctx, `SELECT to_regclass('public.schema_migration_lineage') IS NOT NULL`).Scan(&snapshot.LineageTablePresent); err != nil {
		return migrationStateSnapshot{}, fmt.Errorf("inspect migration lineage: %w", err)
	}
	if !snapshot.LineageTablePresent {
		return snapshot, nil
	}
	rows, err := reader.QueryContext(ctx, `SELECT lineage_id FROM schema_migration_lineage ORDER BY lineage_id ASC`)
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
	if err := rows.Close(); err != nil {
		return migrationStateSnapshot{}, fmt.Errorf("close migration lineage: %w", err)
	}
	return snapshot, nil
}
