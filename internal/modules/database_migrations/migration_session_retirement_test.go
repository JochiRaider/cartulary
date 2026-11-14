package database_migrations

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func assertMigrationSessionRetirement(t *testing.T) {
	t.Helper()
	for _, phase := range []string{"acquire", "unlock"} {
		for _, mode := range []string{"error", "panic", "timeout", "success"} {
			for _, primaryFailure := range []bool{false, true} {
				if primaryFailure && phase == "acquire" {
					continue
				}
				t.Run(phase+" "+mode+" primary="+map[bool]string{false: "false", true: "true"}[primaryFailure], func(t *testing.T) {
					db, conn, physical := migrationSessionFixture(t)
					primary := newMigrationFailure(reasonHistoricalMigrationLineage, nil)
					delegate := &retirementSessionLocker{phase: phase, mode: mode}
					locker := &validatingSessionLocker{delegate: delegate, lockTimeout: 10 * time.Millisecond, unlockTimeout: 10 * time.Millisecond, validate: func(context.Context, *sql.Conn) error {
						if primaryFailure {
							return primary
						}
						return nil
					}}
					err := locker.SessionLock(t.Context(), conn)
					if err == nil {
						err = locker.SessionUnlock(t.Context(), conn)
					}
					if primaryFailure {
						if err != primary {
							t.Fatalf("cleanup replaced primary: %v", err)
						}
					} else if mode == "success" {
						if err != nil {
							t.Fatal(err)
						}
					} else {
						want := reasonSchemaMigrationCleanupFailed
						if phase == "acquire" {
							want = reasonMigrationLockAcquisitionFailed
						}
						requireMigrationFailureReason(t, err, want)
						if mode == "timeout" && !errors.Is(err, context.DeadlineExceeded) {
							t.Fatal("lost deadline identity")
						}
					}
					if err != nil && strings.Contains(err.Error(), "private") {
						t.Fatal("private lock failure escaped")
					}
					retired := mode != "success"
					wantCloses := int32(0)
					if retired {
						wantCloses = 1
					}
					if physical.closes.Load() != wantCloses || physical.locked.Load() {
						t.Fatalf("physical retirement closes=%d locked=%v", physical.closes.Load(), physical.locked.Load())
					}
					if err := conn.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
						t.Fatal(err)
					}
					replacement, err := db.Conn(t.Context())
					if err != nil {
						t.Fatalf("borrowed pool unusable: %v", err)
					}
					defer replacement.Close()
					if err := replacement.Raw(func(value any) error {
						same := value == physical
						if same == retired {
							t.Errorf("physical reuse=%v retired=%v", same, retired)
						}
						return nil
					}); err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

type retirementSessionLocker struct{ phase, mode string }

func (l *retirementSessionLocker) SessionLock(ctx context.Context, conn *sql.Conn) error {
	if err := conn.Raw(func(value any) error { value.(*migrationPhysicalConn).locked.Store(true); return nil }); err != nil {
		return err
	}
	if l.phase == "acquire" {
		return l.fail(ctx)
	}
	return nil
}
func (l *retirementSessionLocker) SessionUnlock(ctx context.Context, conn *sql.Conn) error {
	if l.phase == "unlock" {
		if err := l.fail(ctx); err != nil {
			return err
		}
	}
	return conn.Raw(func(value any) error { value.(*migrationPhysicalConn).locked.Store(false); return nil })
}
func (l *retirementSessionLocker) fail(ctx context.Context) error {
	switch l.mode {
	case "panic":
		panic("private lock failure")
	case "timeout":
		<-ctx.Done()
		return ctx.Err()
	case "error":
		return errors.New("private lock failure")
	}
	return nil
}

type migrationPhysicalConn struct {
	closes atomic.Int32
	locked atomic.Bool
}

func (c *migrationPhysicalConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected statement")
}
func (c *migrationPhysicalConn) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected transaction")
}
func (c *migrationPhysicalConn) Close() error { c.closes.Add(1); c.locked.Store(false); return nil }

type migrationConnector struct{}

func (migrationConnector) Connect(context.Context) (driver.Conn, error) {
	return &migrationPhysicalConn{}, nil
}
func (migrationConnector) Driver() driver.Driver { return migrationDriver{} }

type migrationDriver struct{}

func (migrationDriver) Open(string) (driver.Conn, error) { return &migrationPhysicalConn{}, nil }

func migrationSessionFixture(t *testing.T) (*sql.DB, *sql.Conn, *migrationPhysicalConn) {
	t.Helper()
	db := sql.OpenDB(migrationConnector{})
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	conn, err := db.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
			t.Error(err)
		}
	})
	var physical *migrationPhysicalConn
	if err := conn.Raw(func(value any) error { physical = value.(*migrationPhysicalConn); return nil }); err != nil {
		t.Fatal(err)
	}
	return db, conn, physical
}
