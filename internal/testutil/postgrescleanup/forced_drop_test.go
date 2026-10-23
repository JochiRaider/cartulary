package postgrescleanup

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"testing/synctest"
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
			synctest.Test(t, func(t *testing.T) {
				admissions, statements := 0, 0
				observations := []DropAttempt{}
				operations := dropOperations{withAdmission: func(ctx context.Context, dsn, name string, operation func(databaseDropExecutor) error) error {
					admissions++
					ownerDeadline, bounded := ctx.Deadline()
					if !bounded || time.Until(ownerDeadline) > 35*time.Second {
						t.Fatal("unbounded cleanup owner")
					}
					if dsn != "postgres://admin" || name != testDatabaseName {
						t.Fatal("changed cleanup identity")
					}
					return operation(func(ctx context.Context, name string) error {
						statements++
						deadline, bounded := ctx.Deadline()
						if !bounded || deadline != ownerDeadline || time.Until(deadline) != 35*time.Second {
							t.Fatal("drop lost the shared owner deadline or unused admission budget")
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
		})
	}
	for _, expires := range []bool{false, true} {
		name := "unused_admission_time_remains_available"
		if expires {
			name = "admission_time_cannot_be_replenished"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				statements := 0
				var observation DropAttempt
				operations := dropOperations{withAdmission: func(ctx context.Context, _, _ string, operation func(databaseDropExecutor) error) error {
					deadline, _ := ctx.Deadline()
					time.Sleep(10 * time.Second)
					return operation(func(ctx context.Context, _ string) error {
						statements++
						actual, bounded := ctx.Deadline()
						if !bounded || actual != deadline || time.Until(actual) != 25*time.Second {
							t.Fatal("phase transition changed the owner deadline")
						}
						work := 21 * time.Second
						if expires {
							work = 26 * time.Second
						}
						select {
						case <-time.After(work):
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					})
				}, observe: func(value DropAttempt) { observation = value }}
				err := dropOwnedDatabase(context.Background(), "postgres://admin", testDatabaseName, operations)
				wantDuration, wantOperation := int64(31000), int64(21000)
				if expires {
					wantDuration, wantOperation = 35000, 25000
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Fatalf("owner expiry lost: %v", err)
					}
				} else if err != nil {
					t.Fatalf("operation within owner budget failed: %v", err)
				}
				if statements != 1 || observation.DurationMS != wantDuration || observation.AdmissionMS != 10000 || observation.AdmissionBudgetMS != 15000 || observation.OperationMS != wantOperation || observation.OperationBudgetMS != 25000 || observation.Validate() != nil {
					t.Fatalf("shared budget not represented faithfully: statements=%d observation=%+v", statements, observation)
				}
			})
		})
	}
}

func TestDropOwnedDatabaseClampsParentDeadline(t *testing.T) {
	for _, allowance := range []time.Duration{time.Second, time.Minute} {
		t.Run(allowance.String(), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				parent, cancel := context.WithTimeout(context.Background(), allowance)
				defer cancel()
				deadline := time.Now().Add(min(allowance, 35*time.Second))
				operations := dropOperations{withAdmission: func(ctx context.Context, _, _ string, operation func(databaseDropExecutor) error) error {
					if actual, bounded := ctx.Deadline(); !bounded || actual != deadline {
						t.Fatal("admission escaped the shorter parent or default owner deadline")
					}
					return operation(func(ctx context.Context, _ string) error {
						actual, bounded := ctx.Deadline()
						if !bounded || actual != deadline {
							t.Fatal("operation escaped the owner deadline")
						}
						cancel()
						return ctx.Err()
					})
				}}
				if err := dropOwnedDatabase(parent, "postgres://admin", testDatabaseName, operations); !errors.Is(err, context.Canceled) {
					t.Fatalf("lost parent cancellation: %v", err)
				}
			})
		})
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
	expired, stopExpired := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer stopExpired()
	if err := dropOwnedDatabase(expired, "postgres://admin", testDatabaseName, operations); !errors.Is(err, context.DeadlineExceeded) || admissions != 1 {
		t.Fatalf("expired parent reached admission: %v admissions=%d", err, admissions)
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
	t.Run("shared_cleanup_budget_projection", testCleanupBudgetProjection)
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
