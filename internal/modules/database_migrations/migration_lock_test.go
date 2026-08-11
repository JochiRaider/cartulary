package database_migrations

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

type recordingSessionLocker struct {
	lockCalls   int
	unlockCalls int
	lockErr     error
	unlockErr   error
	unlockWait  bool
	unlockCtx   context.Context
	lockConn    *sql.Conn
	unlockConn  *sql.Conn
}

func (locker *recordingSessionLocker) SessionLock(_ context.Context, conn *sql.Conn) error {
	locker.lockCalls++
	locker.lockConn = conn
	return locker.lockErr
}

func (locker *recordingSessionLocker) SessionUnlock(ctx context.Context, conn *sql.Conn) error {
	locker.unlockCalls++
	locker.unlockCtx = ctx
	locker.unlockConn = conn
	if locker.unlockWait {
		<-ctx.Done()
		return ctx.Err()
	}
	return locker.unlockErr
}

func TestValidatingSessionLockerReleasesAfterPreflightFailure(t *testing.T) {
	delegate := &recordingSessionLocker{}
	primary := newMigrationFailure(reasonHistoricalMigrationLineage, nil)
	conn := &sql.Conn{}
	var validationConn *sql.Conn
	locker := &validatingSessionLocker{
		delegate: delegate,
		validate: func(_ context.Context, conn *sql.Conn) error {
			validationConn = conn
			return primary
		},
		lockTimeout:   time.Second,
		unlockTimeout: time.Second,
	}

	err := locker.SessionLock(context.Background(), conn)
	requireMigrationFailureReason(t, err, reasonHistoricalMigrationLineage)
	if delegate.lockCalls != 1 || delegate.unlockCalls != 1 {
		t.Fatalf("unexpected lock lifecycle: lock=%d unlock=%d", delegate.lockCalls, delegate.unlockCalls)
	}
	if delegate.lockConn != conn || validationConn != conn || delegate.unlockConn != conn {
		t.Fatal("lock, validation, and unlock did not share one database session")
	}
}

func TestValidatingSessionLockerRecoversPanicAndReleases(t *testing.T) {
	delegate := &recordingSessionLocker{}
	locker := &validatingSessionLocker{
		delegate: delegate,
		validate: func(context.Context, *sql.Conn) error {
			panic("private provider detail")
		},
		lockTimeout:   time.Second,
		unlockTimeout: time.Second,
	}

	err := locker.SessionLock(context.Background(), &sql.Conn{})
	requireMigrationFailureReason(t, err, reasonSchemaMigrationExecutionFailed)
	if delegate.lockCalls != 1 || delegate.unlockCalls != 1 {
		t.Fatalf("unexpected panic lock lifecycle: lock=%d unlock=%d", delegate.lockCalls, delegate.unlockCalls)
	}
}

func TestValidatingSessionLockerBoundsDetachedUnlock(t *testing.T) {
	delegate := &recordingSessionLocker{unlockWait: true}
	locker := &validatingSessionLocker{
		delegate:      delegate,
		validate:      func(context.Context, *sql.Conn) error { return nil },
		lockTimeout:   time.Second,
		unlockTimeout: 20 * time.Millisecond,
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	started := time.Now()
	err := locker.SessionUnlock(canceled, &sql.Conn{})
	elapsed := time.Since(started)
	requireMigrationFailureReason(t, err, reasonSchemaMigrationCleanupFailed)
	if elapsed < 10*time.Millisecond || elapsed > time.Second {
		t.Fatalf("detached unlock duration = %s", elapsed)
	}
	if delegate.unlockCtx == nil || !errors.Is(delegate.unlockCtx.Err(), context.DeadlineExceeded) {
		t.Fatalf("unlock did not receive bounded detached context: %v", delegate.unlockCtx)
	}
}

func TestMigrationOperationTimeouts(t *testing.T) {
	if migrationLockTimeout != 5*time.Minute || migrationUnlockTimeout != 30*time.Second {
		t.Fatalf("migration lock ceilings = %s / %s", migrationLockTimeout, migrationUnlockTimeout)
	}
}

type multipleProviderErrors struct {
	errors []error
}

func (failure multipleProviderErrors) Error() string {
	return "private provider failure"
}

func (failure multipleProviderErrors) Unwrap() []error {
	return failure.errors
}

func TestProviderFailureNormalizationPreservesPrimaryError(t *testing.T) {
	primary := errors.New("SELECT secret FROM private_relation")
	cleanup := newMigrationFailure(reasonSchemaMigrationCleanupFailed, errors.New("private server detail"))
	err := normalizeProviderFailure(multipleProviderErrors{errors: []error{primary, cleanup}})
	requireMigrationFailureReason(t, err, reasonSchemaMigrationExecutionFailed)

	err = normalizeProviderFailure(multipleProviderErrors{errors: []error{cleanup}})
	requireMigrationFailureReason(t, err, reasonSchemaMigrationCleanupFailed)
}
