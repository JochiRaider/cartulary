package postgrescleanup

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

const testDatabaseName = "ct_deadbeef_cafe0001_000001_owned"

func TestDropOwnedDatabaseRunsOneBoundedOperation(t *testing.T) {
	for _, outcome := range []struct {
		name string
		err  error
	}{
		{"success", nil},
		{"deadline", context.DeadlineExceeded},
		{"cancellation", context.Canceled},
		{"permission", &pgconn.PgError{Code: "42501"}},
		{"prepared_transaction", &pgconn.PgError{Code: "55006"}},
	} {
		t.Run(outcome.name, func(t *testing.T) {
			admissions, statements := 0, 0
			observations := []DropAttempt{}
			operations := dropOperations{withAdmission: func(ctx context.Context, dsn, name string, operation func(databaseDropExecutor) error) error {
				admissions++
				if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > 35*time.Second {
					t.Fatal("unbounded cleanup owner")
				}
				if dsn != "postgres://admin" || name != testDatabaseName {
					t.Fatal("changed cleanup identity")
				}
				return operation(func(ctx context.Context, name string) error {
					statements++
					deadline, bounded := ctx.Deadline()
					if !bounded || time.Until(deadline) <= 19*time.Second || time.Until(deadline) > 20*time.Second {
						t.Fatal("drop must have a 20 second cap")
					}
					if name != testDatabaseName {
						t.Fatal("changed owned target")
					}
					return outcome.err
				})
			}, observe: func(attempt DropAttempt) { observations = append(observations, attempt) }}
			err := dropOwnedDatabase(context.Background(), "postgres://admin", testDatabaseName, operations)
			if !errors.Is(err, outcome.err) || admissions != 1 || statements != 1 {
				t.Fatalf("error=%v admissions=%d statements=%d", err, admissions, statements)
			}
			if len(observations) != 1 || observations[0].Validate() != nil || observations[0].Admission != "acquired" || (observations[0].Outcome == "completed") != (outcome.err == nil) {
				t.Fatal("attempt observation differs from operation", observations)
			}
		})
	}
}

func TestDropOwnedDatabaseClampsParentDeadline(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	deadline, _ := parent.Deadline()
	operations := dropOperations{withAdmission: func(ctx context.Context, _, _ string, operation func(databaseDropExecutor) error) error {
		if ctx != parent {
			t.Fatal("admission lost owning context")
		}
		return operation(func(ctx context.Context, _ string) error {
			actual, ok := ctx.Deadline()
			if !ok || actual != deadline {
				t.Fatal("operation escaped parent deadline")
			}
			cancel()
			return ctx.Err()
		})
	}}
	if err := dropOwnedDatabase(parent, "postgres://admin", testDatabaseName, operations); !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation: %v", err)
	}
}

func TestDropOwnedDatabasePreservesOperationAndSettlementFailures(t *testing.T) {
	operationErr := &pgconn.PgError{Code: "55006"}
	unlockErr, closeErr := errors.New("unlock"), errors.New("close")
	operations := dropOperations{withAdmission: func(_ context.Context, _, _ string, operation func(databaseDropExecutor) error) error {
		return errors.Join(operation(func(context.Context, string) error { return operationErr }), unlockErr, closeErr)
	}}
	err := dropOwnedDatabase(context.Background(), "postgres://admin", testDatabaseName, operations)
	var pgErr *pgconn.PgError
	if !errors.Is(err, operationErr) || !errors.Is(err, unlockErr) || !errors.Is(err, closeErr) || !errors.As(err, &pgErr) || pgErr != operationErr {
		t.Fatalf("lost independent cause: %v", err)
	}
}

func TestDropOwnedDatabaseRejectsAdmissionAndCancellation(t *testing.T) {
	admissions, statements := 0, 0
	rejected := errors.New("admission rejected")
	operations := dropOperations{withAdmission: func(context.Context, string, string, func(databaseDropExecutor) error) error {
		admissions++
		return rejected
	}}
	if err := dropOwnedDatabase(context.Background(), "postgres://admin", testDatabaseName, operations); !errors.Is(err, rejected) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := dropOwnedDatabase(ctx, "postgres://admin", testDatabaseName, operations); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if admissions != 1 {
		t.Fatal("pre-cancellation reached admission")
	}
	ctx, cancel = context.WithCancel(context.Background())
	operations.withAdmission = func(_ context.Context, _, _ string, operation func(databaseDropExecutor) error) error {
		cancel()
		return operation(func(context.Context, string) error { statements++; return nil })
	}
	if err := dropOwnedDatabase(ctx, "postgres://admin", testDatabaseName, operations); !errors.Is(err, context.Canceled) || statements != 0 {
		t.Fatalf("cancelled admission issued statement: %v count=%d", err, statements)
	}
}

func TestDropOwnedDatabaseRejectsIncompleteInputsBeforeLock(t *testing.T) {
	t.Run("barrier_notice_is_closed_and_transient", testBarrierNoticeDiagnostics)
	t.Run("checkpoint_timing_is_bounded_and_reset_aware", testCheckpointDiagnostics)
	lockCalls := 0
	operations := dropOperations{withAdmission: func(context.Context, string, string, func(databaseDropExecutor) error) error { lockCalls++; return nil }}
	for _, test := range []struct {
		name, dsn, database string
		ctx                 context.Context
	}{
		{"missing dsn", "", testDatabaseName, context.Background()},
		{"unsafe database", "postgres://admin", "bad-name", context.Background()},
		{"nil context", "postgres://admin", testDatabaseName, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := dropOwnedDatabase(test.ctx, test.dsn, test.database, operations); err == nil {
				t.Fatal("expected input rejection")
			}
		})
	}
	if lockCalls != 0 {
		t.Fatalf("invalid inputs acquired admission %d times", lockCalls)
	}
}

func testBarrierNoticeDiagnostics(t *testing.T) {
	for query, want := range map[string]string{"": "none", "CREATE DATABASE secret": "create_database", "DROP DATABASE secret WITH (FORCE)": "drop_database", " select secret from private ": "select", "/* private */ SELECT hidden": "other", "DO $$ private $$": "do", "none secret": "other", "create_database secret": "other"} {
		if got := barrierOperation(query); got != want {
			t.Fatalf("operation category: got %s want %s", got, want)
		}
	}
	valid := pgconn.Notice{SeverityUnlocalized: "LOG", Code: "00000", Routine: "WaitForProcSignalBarrier", Message: "still waiting for backend with PID 123 to accept ProcSignalBarrier"}
	if barrierNoticePID(&valid) != 123 || barrierNoticePID(nil) != 0 {
		t.Fatal("notice admission failed")
	}
	for _, mutate := range []func(*pgconn.Notice){
		func(n *pgconn.Notice) { n.Routine = "exec_stmt_raise" },
		func(n *pgconn.Notice) { n.Code = "P0001" },
		func(n *pgconn.Notice) { n.SeverityUnlocalized = "WARNING" },
		func(n *pgconn.Notice) { n.Message = "private server text" },
		func(n *pgconn.Notice) {
			n.Message = "still waiting for backend with PID -1 to accept ProcSignalBarrier"
		},
		func(n *pgconn.Notice) {
			n.Message = "still waiting for backend with PID 2147483648 to accept ProcSignalBarrier"
		},
		func(n *pgconn.Notice) { n.Message += " private suffix" },
	} {
		notice := valid
		mutate(&notice)
		if barrierNoticePID(&notice) != 0 {
			t.Fatal("unrecognized notice admitted")
		}
	}
	if got := sampleBarrierParticipant(context.Background(), nil, "target", nil); got.Status != "not_observed" || !got.valid() {
		t.Fatal("absent notice performed a query")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	db, err := sql.Open("pgx", "postgres://unused")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if got := sampleBarrierParticipant(ctx, db, "target", func() int64 { return 123 }); got.Status != "unavailable" || !got.valid() {
		t.Fatal("failed sample did not close")
	}
}
