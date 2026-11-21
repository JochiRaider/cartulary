package database_migrations

import (
	"context"
	"errors"
)

const (
	reasonMigrationContextInvalid        = "migration_context_invalid"
	reasonMigrationSourceInvalid         = "migration_source_invalid"
	reasonMigrationDatabaseUnavailable   = "migration_database_unavailable"
	reasonMigrationLockAcquisitionFailed = "migration_lock_acquisition_failed"
	reasonSchemaExtensionPrerequisite    = "schema_extension_prerequisite_invalid"
	reasonSchemaMigrationHistoryInvalid  = "schema_migration_history_invalid"
	reasonHistoricalMigrationLineage     = "historical_migration_lineage"
	reasonSchemaMigrationRequired        = "schema_migration_required"
	reasonSchemaVersionAhead             = "schema_version_ahead"
	reasonSchemaMigrationExecutionFailed = "schema_migration_execution_failed"
	reasonSchemaMigrationCleanupFailed   = "schema_migration_cleanup_failed"
	reasonSchemaMigrationPostcondition   = "schema_migration_postcondition_failed"
)

// MigrationFailure is the closed, safe diagnostic surface for migration
// failures. Implementations never return vendor or deployment data from
// Error or ReasonCode.
type MigrationFailure interface {
	error
	ReasonCode() string
}

// RemediationReporter is a migration failure with a stable operator report.
type RemediationReporter interface {
	MigrationFailure
	RemediationReportJSON() string
}

type migrationFailure struct {
	reasonCode string
	safeCause  error
}

func newMigrationFailure(reasonCode string, cause error) error {
	failure := &migrationFailure{reasonCode: reasonCode}
	switch {
	case errors.Is(cause, context.Canceled):
		failure.safeCause = context.Canceled
	case errors.Is(cause, context.DeadlineExceeded):
		failure.safeCause = context.DeadlineExceeded
	}
	return failure
}

func (failure *migrationFailure) Error() string {
	return failure.reasonCode
}

func (failure *migrationFailure) ReasonCode() string {
	return failure.reasonCode
}

func (failure *migrationFailure) Unwrap() error {
	return failure.safeCause
}

func normalizeMigrationFailure(err error, fallbackReason string) error {
	if err == nil {
		return nil
	}
	var failure MigrationFailure
	if errors.As(err, &failure) {
		return failure
	}
	return newMigrationFailure(fallbackReason, err)
}

func normalizeProviderFailure(err error) error {
	if err == nil {
		return nil
	}
	for candidate := err; candidate != nil; candidate = errors.Unwrap(candidate) {
		many, ok := candidate.(interface{ Unwrap() []error })
		if !ok {
			continue
		}
		parts := many.Unwrap()
		if len(parts) > 0 && parts[0] != nil {
			return normalizeMigrationFailure(parts[0], reasonSchemaMigrationExecutionFailed)
		}
		break
	}
	return normalizeMigrationFailure(err, reasonSchemaMigrationExecutionFailed)
}

var _ MigrationFailure = (*migrationFailure)(nil)
