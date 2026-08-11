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
