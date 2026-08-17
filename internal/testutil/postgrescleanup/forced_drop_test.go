package postgrescleanup

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

const testDatabaseName = "ct_deadbeef_cafe0001_000001_owned"

func TestDropOwnedDatabaseReleasesNormalAdmissionBeforeForcedFallback(t *testing.T) {
	var events []string
	operations := dropOperations{
		withNormalAdmission: func(
			ctx context.Context,
			adminDSN string,
			name string,
			operation func(databaseDropExecutor),
		) error {
			events = append(events, "shared-lock")
			if ctx.Err() != nil || adminDSN != "postgres://admin" || name != testDatabaseName {
				t.Fatalf("unexpected shared-lock input: dsn=%q name=%q err=%v", adminDSN, name, ctx.Err())
			}
			operation(func(ctx context.Context, name string, force bool) error {
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("normal drop did not receive a bounded context")
				}
				if name != testDatabaseName || force {
					t.Fatalf("unexpected normal drop input: name=%q force=%t", name, force)
				}
				events = append(events, "normal")
				return &pgconn.PgError{Code: "55006", Message: "database is in use"}
			})
			events = append(events, "shared-unlock")
			return nil
		},
		withForcedAdmission: func(
			ctx context.Context,
			adminDSN string,
			name string,
			operation func(databaseDropExecutor),
		) error {
			events = append(events, "exclusive-lock")
			if ctx.Err() != nil || adminDSN != "postgres://admin" || name != testDatabaseName {
				t.Fatalf("unexpected exclusive-lock input: dsn=%q name=%q err=%v", adminDSN, name, ctx.Err())
			}
			operation(func(ctx context.Context, name string, force bool) error {
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("forced drop did not receive a bounded context")
				}
				if name != testDatabaseName || !force {
					t.Fatalf("unexpected forced drop input: name=%q force=%t", name, force)
				}
				events = append(events, "forced")
				return nil
			})
			events = append(events, "exclusive-unlock")
			return nil
		},
	}

	forced, err := dropOwnedDatabase(
		context.Background(),
		"postgres://admin",
		testDatabaseName,
		time.Second,
		time.Second,
		operations,
	)
	if err != nil {
		t.Fatalf("drop owned database: %v", err)
	}
	if !forced {
		t.Fatal("active database did not report forced fallback")
	}
	want := []string{
		"shared-lock",
		"normal",
		"shared-unlock",
		"exclusive-lock",
		"forced",
		"exclusive-unlock",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("operation order = %v, want %v", events, want)
	}
}

func TestDropOwnedDatabaseUsesFreshForcedContextAfterNormalTimeout(t *testing.T) {
	dropCalls := 0
	operations := dropOperations{
		withNormalAdmission: func(_ context.Context, _ string, _ string, operation func(databaseDropExecutor)) error {
			operation(func(ctx context.Context, _ string, force bool) error {
				dropCalls++
				if force {
					t.Fatal("normal admission executed forced drop")
				}
				<-ctx.Done()
				return ctx.Err()
			})
			return nil
		},
		withForcedAdmission: func(_ context.Context, _ string, _ string, operation func(databaseDropExecutor)) error {
			operation(func(ctx context.Context, _ string, force bool) error {
				dropCalls++
				if !force {
					t.Fatal("exclusive lock executed ordinary drop")
				}
				if ctx.Err() != nil {
					t.Fatalf("forced fallback inherited expired normal context: %v", ctx.Err())
				}
				return nil
			})
			return nil
		},
	}

	forced, err := dropOwnedDatabase(
		context.Background(),
		"postgres://admin",
		testDatabaseName,
		time.Millisecond,
		time.Second,
		operations,
	)
	if err != nil {
		t.Fatalf("drop after normal timeout: %v", err)
	}
	if !forced || dropCalls != 2 {
		t.Fatalf("forced=%t calls=%d, want true and 2", forced, dropCalls)
	}
}

func TestDropOwnedDatabaseDoesNotForceUnrelatedOrParentCancellation(t *testing.T) {
	for _, test := range []struct {
		name      string
		parent    func() context.Context
		normalErr error
	}{
		{
			name:      "unrelated failure",
			parent:    context.Background,
			normalErr: errors.New("permission denied"),
		},
		{
			name: "parent cancellation",
			parent: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			normalErr: context.Canceled,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			dropCalls := 0
			exclusiveCalls := 0
			operations := dropOperations{
				withNormalAdmission: func(_ context.Context, _ string, _ string, operation func(databaseDropExecutor)) error {
					operation(func(_ context.Context, _ string, force bool) error {
						dropCalls++
						if force {
							t.Fatal("normal admission executed forced drop")
						}
						return test.normalErr
					})
					return nil
				},
				withForcedAdmission: func(_ context.Context, _ string, _ string, _ func(databaseDropExecutor)) error {
					exclusiveCalls++
					return nil
				},
			}
			forced, err := dropOwnedDatabase(
				test.parent(),
				"postgres://admin",
				testDatabaseName,
				time.Second,
				time.Second,
				operations,
			)
			if !errors.Is(err, test.normalErr) {
				t.Fatalf("drop error = %v, want %v", err, test.normalErr)
			}
			if forced || dropCalls != 1 || exclusiveCalls != 0 {
				t.Fatalf("forced=%t drop calls=%d exclusive calls=%d, want false, 1, 0", forced, dropCalls, exclusiveCalls)
			}
		})
	}
}

func TestDropOwnedDatabaseTreatsCoordinationFailureAsHardFailure(t *testing.T) {
	unlockErr := errors.New("shared unlock failed")
	exclusiveCalls := 0
	operations := dropOperations{
		withNormalAdmission: func(_ context.Context, _ string, _ string, operation func(databaseDropExecutor)) error {
			operation(func(_ context.Context, _ string, _ bool) error {
				return &pgconn.PgError{Code: "55006", Message: "database is in use"}
			})
			return unlockErr
		},
		withForcedAdmission: func(_ context.Context, _ string, _ string, _ func(databaseDropExecutor)) error {
			exclusiveCalls++
			return nil
		},
	}

	forced, err := dropOwnedDatabase(
		context.Background(),
		"postgres://admin",
		testDatabaseName,
		time.Second,
		time.Second,
		operations,
	)
	if !errors.Is(err, unlockErr) {
		t.Fatalf("drop error = %v, want shared coordination failure", err)
	}
	if forced || exclusiveCalls != 0 {
		t.Fatalf("forced=%t exclusive calls=%d, want false and 0", forced, exclusiveCalls)
	}
}

func TestDropOwnedDatabaseForcedFallbackWaitsForNormalAdmission(t *testing.T) {
	const readerDatabase = "ct_deadbeef_cafe0001_000002_owned"
	var catalog sync.RWMutex
	readerEntered := make(chan struct{})
	releaseReader := make(chan struct{})
	exclusiveWaiting := make(chan struct{})
	exclusiveEntered := make(chan struct{})
	operations := dropOperations{
		withNormalAdmission: func(_ context.Context, _ string, _ string, operation func(databaseDropExecutor)) error {
			catalog.RLock()
			defer catalog.RUnlock()
			operation(func(_ context.Context, name string, force bool) error {
				if force {
					t.Fatal("normal admission executed forced drop")
				}
				if name == readerDatabase {
					close(readerEntered)
					<-releaseReader
					return nil
				}
				return &pgconn.PgError{Code: "55006", Message: "database is in use"}
			})
			return nil
		},
		withForcedAdmission: func(_ context.Context, _ string, _ string, operation func(databaseDropExecutor)) error {
			close(exclusiveWaiting)
			catalog.Lock()
			defer catalog.Unlock()
			close(exclusiveEntered)
			operation(func(_ context.Context, _ string, force bool) error {
				if !force {
					t.Fatal("exclusive lock executed ordinary drop")
				}
				return nil
			})
			return nil
		},
	}

	readerDone := make(chan error, 1)
	go func() {
		_, err := dropOwnedDatabase(
			context.Background(),
			"postgres://admin",
			readerDatabase,
			time.Second,
			time.Second,
			operations,
		)
		readerDone <- err
	}()
	<-readerEntered

	fallbackDone := make(chan error, 1)
	go func() {
		_, err := dropOwnedDatabase(
			context.Background(),
			"postgres://admin",
			testDatabaseName,
			time.Second,
			time.Second,
			operations,
		)
		fallbackDone <- err
	}()
	<-exclusiveWaiting
	select {
	case <-exclusiveEntered:
		t.Fatal("forced fallback entered while an ordinary drop held admission")
	default:
	}
	close(releaseReader)
	if err := <-readerDone; err != nil {
		t.Fatalf("ordinary drop: %v", err)
	}
	if err := <-fallbackDone; err != nil {
		t.Fatalf("forced fallback: %v", err)
	}
	select {
	case <-exclusiveEntered:
	default:
		t.Fatal("forced fallback never acquired exclusive admission")
	}
}

func TestDropOwnedDatabaseRejectsIncompleteInputsBeforeLock(t *testing.T) {
	lockCalls := 0
	operations := dropOperations{
		withNormalAdmission: func(_ context.Context, _ string, _ string, _ func(databaseDropExecutor)) error {
			lockCalls++
			return nil
		},
		withForcedAdmission: func(_ context.Context, _ string, _ string, _ func(databaseDropExecutor)) error {
			lockCalls++
			return nil
		},
	}
	for _, test := range []struct {
		name        string
		adminDSN    string
		database    string
		normalLimit time.Duration
		forcedLimit time.Duration
	}{
		{name: "missing dsn", database: testDatabaseName, normalLimit: time.Second, forcedLimit: time.Second},
		{name: "unsafe database", adminDSN: "postgres://admin", database: `bad-name`, normalLimit: time.Second, forcedLimit: time.Second},
		{name: "missing normal limit", adminDSN: "postgres://admin", database: testDatabaseName, forcedLimit: time.Second},
		{name: "missing forced limit", adminDSN: "postgres://admin", database: testDatabaseName, normalLimit: time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := dropOwnedDatabase(
				context.Background(),
				test.adminDSN,
				test.database,
				test.normalLimit,
				test.forcedLimit,
				operations,
			); err == nil {
				t.Fatal("expected invalid input rejection")
			}
		})
	}
	if lockCalls != 0 {
		t.Fatalf("invalid inputs acquired catalog lock %d times", lockCalls)
	}
}
