package database_migrations

import (
	"context"
	"database/sql"
	"errors"
	"time"

	gooselock "github.com/pressly/goose/v3/lock"
)

const (
	migrationLockTimeout   = 5 * time.Minute
	migrationUnlockTimeout = 30 * time.Second
)

type validatingSessionLocker struct {
	delegate      gooselock.SessionLocker
	validate      func(context.Context, *sql.Conn) error
	lockTimeout   time.Duration
	unlockTimeout time.Duration
}

func (locker *validatingSessionLocker) SessionLock(ctx context.Context, conn *sql.Conn) (retErr error) {
	if locker == nil || locker.delegate == nil || locker.validate == nil {
		return newMigrationFailure(reasonMigrationLockAcquisitionFailed, nil)
	}

	lockCtx, cancel := context.WithTimeout(ctx, locker.effectiveLockTimeout())
	defer cancel()
	if err := locker.delegate.SessionLock(lockCtx, conn); err != nil {
		return newMigrationFailure(reasonMigrationLockAcquisitionFailed, err)
	}

	release := true
	defer func() {
		if recovered := recover(); recovered != nil {
			retErr = newMigrationFailure(reasonSchemaMigrationExecutionFailed, nil)
		}
		if !release {
			return
		}
		if unlockErr := locker.unlock(ctx, conn); unlockErr != nil && retErr == nil {
			retErr = unlockErr
		}
	}()

	if err := locker.validate(ctx, conn); err != nil {
		return err
	}
	release = false
	return nil
}

func (locker *validatingSessionLocker) SessionUnlock(ctx context.Context, conn *sql.Conn) error {
	if locker == nil || locker.delegate == nil {
		return newMigrationFailure(reasonSchemaMigrationCleanupFailed, nil)
	}
	return locker.unlock(ctx, conn)
}

func (locker *validatingSessionLocker) unlock(ctx context.Context, conn *sql.Conn) error {
	detached := context.WithoutCancel(ctx)
	unlockCtx, cancel := context.WithTimeout(detached, locker.effectiveUnlockTimeout())
	defer cancel()
	if err := locker.delegate.SessionUnlock(unlockCtx, conn); err != nil {
		return newMigrationFailure(reasonSchemaMigrationCleanupFailed, err)
	}
	return nil
}

func (locker *validatingSessionLocker) effectiveLockTimeout() time.Duration {
	if locker.lockTimeout > 0 {
		return locker.lockTimeout
	}
	return migrationLockTimeout
}

func (locker *validatingSessionLocker) effectiveUnlockTimeout() time.Duration {
	if locker.unlockTimeout > 0 {
		return locker.unlockTimeout
	}
	return migrationUnlockTimeout
}

func withLockedMigrationSession(
	ctx context.Context,
	db *sql.DB,
	locker gooselock.SessionLocker,
	fn func(context.Context, *sql.Conn) error,
) (retErr error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return newMigrationFailure(reasonMigrationDatabaseUnavailable, err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil && retErr == nil {
			retErr = newMigrationFailure(reasonSchemaMigrationCleanupFailed, closeErr)
		}
	}()

	validating := &validatingSessionLocker{
		delegate:      locker,
		validate:      fn,
		lockTimeout:   migrationLockTimeout,
		unlockTimeout: migrationUnlockTimeout,
	}
	if err := validating.SessionLock(ctx, conn); err != nil {
		return err
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			retErr = newMigrationFailure(reasonSchemaMigrationExecutionFailed, nil)
		}
		if unlockErr := validating.SessionUnlock(ctx, conn); unlockErr != nil && retErr == nil {
			retErr = unlockErr
		}
	}()

	if fn == nil {
		return newMigrationFailure(reasonSchemaMigrationExecutionFailed, errors.New("missing locked migration operation"))
	}
	return nil
}

func migrationWorkNeeded(ctx context.Context, reader sqlLedgerReader, source *Source) (bool, error) {
	snapshot, err := readSQLMigrationState(ctx, reader)
	if err != nil {
		return false, newMigrationFailure(reasonMigrationDatabaseUnavailable, err)
	}
	classification, err := classifyMigrationState(source, snapshot)
	if err != nil {
		return false, err
	}
	if err := validateMigrationPrerequisites(ctx, reader); err != nil {
		return false, err
	}
	if classification.State == migrationStatePristine {
		contaminated, contaminationErr := databaseHasPreexistingObjects(ctx, reader)
		if contaminationErr != nil {
			return false, contaminationErr
		}
		if contaminated {
			return false, newMigrationLineageRemediationError(
				source,
				migrationLineageState{},
				0,
				sourceHeadVersion(source),
				sourceHeadVersion(source),
			)
		}
	}
	switch classification.State {
	case migrationStatePristine, migrationStateBehind:
		return true, nil
	case migrationStateCurrent:
		return false, nil
	case migrationStateAhead:
		return false, newMigrationFailure(reasonSchemaVersionAhead, nil)
	}
	return false, newMigrationFailure(reasonSchemaMigrationHistoryInvalid, nil)
}

func validateMigrationPrerequisites(ctx context.Context, reader sqlLedgerReader) error {
	var valid bool
	if err := reader.QueryRowContext(ctx, `
SELECT COUNT(*) = 2
   AND BOOL_AND(
       (extension.extname = 'pgcrypto' AND extension.extversion = '1.3')
       OR (extension.extname = 'citext' AND extension.extversion = '1.6')
   )
   AND BOOL_AND(namespace.nspname = 'public')
FROM pg_catalog.pg_extension AS extension
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid = extension.extnamespace
WHERE extension.extname IN ('pgcrypto', 'citext')
`).Scan(&valid); err != nil {
		return newMigrationFailure(reasonMigrationDatabaseUnavailable, err)
	}
	if !valid {
		return newMigrationFailure(reasonSchemaExtensionPrerequisite, nil)
	}
	return nil
}

func databaseHasPreexistingObjects(ctx context.Context, reader sqlLedgerReader) (bool, error) {
	var contaminated bool
	if err := reader.QueryRowContext(ctx, `
WITH extension_objects AS (
    SELECT dependency.classid, dependency.objid
    FROM pg_catalog.pg_depend AS dependency
    JOIN pg_catalog.pg_extension AS extension
      ON extension.oid = dependency.refobjid
    WHERE dependency.deptype = 'e'
      AND extension.extname IN ('pgcrypto', 'citext')
), contamination AS (
    SELECT relation.oid
    FROM pg_catalog.pg_class AS relation
    JOIN pg_catalog.pg_namespace AS namespace
      ON namespace.oid = relation.relnamespace
    WHERE namespace.nspname = 'public'
      AND relation.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')
      AND relation.relname NOT IN ('goose_db_version', 'goose_db_version_id_seq')
      AND NOT EXISTS (
          SELECT 1
          FROM extension_objects
          WHERE extension_objects.classid = 'pg_catalog.pg_class'::pg_catalog.regclass
            AND extension_objects.objid = relation.oid
      )
    UNION ALL
    SELECT procedure.oid
    FROM pg_catalog.pg_proc AS procedure
    JOIN pg_catalog.pg_namespace AS namespace
      ON namespace.oid = procedure.pronamespace
    WHERE namespace.nspname = 'public'
      AND NOT EXISTS (
          SELECT 1
          FROM extension_objects
          WHERE extension_objects.classid = 'pg_catalog.pg_proc'::pg_catalog.regclass
            AND extension_objects.objid = procedure.oid
      )
    UNION ALL
    SELECT namespace.oid
    FROM pg_catalog.pg_namespace AS namespace
    WHERE namespace.nspname NOT IN ('public', 'information_schema')
      AND namespace.nspname NOT LIKE 'pg\_%' ESCAPE E'\\'
)
SELECT EXISTS (SELECT 1 FROM contamination)
`).Scan(&contaminated); err != nil {
		return false, newMigrationFailure(reasonMigrationDatabaseUnavailable, err)
	}
	return contaminated, nil
}

func verifyMigrationPostcondition(ctx context.Context, reader sqlLedgerReader, source *Source) error {
	needed, err := migrationWorkNeeded(ctx, reader, source)
	if err != nil {
		return err
	}
	if needed {
		return newMigrationFailure(reasonSchemaMigrationPostcondition, nil)
	}
	return nil
}

func recoverProviderPanic(retErr *error) {
	if recovered := recover(); recovered != nil {
		*retErr = newMigrationFailure(reasonSchemaMigrationExecutionFailed, nil)
	}
}

var _ gooselock.SessionLocker = (*validatingSessionLocker)(nil)
