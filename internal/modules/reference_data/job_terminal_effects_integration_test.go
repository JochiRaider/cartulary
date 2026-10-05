package reference_data

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type failingTerminalEffects struct{ err error }

func (f failingTerminalEffects) ApplyJobTerminalEffectsTx(ctx context.Context, tx pgx.Tx, d jobs.TerminalDisposition) error {
	if err := (JobTerminalEffects{}).ApplyJobTerminalEffectsTx(ctx, tx, d); err != nil {
		return err
	}
	return f.err
}

func testJobTerminalOwnerRecovery(t *testing.T, c *Coordinator, pool *pgxpool.Pool, catalog *jobs.Catalog, transactions *jobs.TransactionService, definitions []jobs.Definition, actor uuid.UUID, container []byte, established bool) {
	t.Helper()
	ctx := context.Background()
	policy := jobs.ProductionRuntimePolicy()
	policy.MaximumFailures = 1
	policy.RetryDelays = nil
	manager, err := jobs.NewManager(jobs.ManagerOptions{Postgres: pool, Transactions: transactions, Catalog: catalog, Policy: policy, Now: func() time.Time { return time.Now().UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	cases := []string{"queued inactive cancellation", "exhausted before verification", "exhausted after private verification"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			var before []byte
			if established {
				if err := c.pool.QueryRow(ctx, `SELECT to_jsonb(c)::text FROM reference_pack_candidates c WHERE pack_key='type_registry.host' AND pack_version='signed-fixture.1'`).Scan(&before); err != nil {
					t.Fatal(err)
				}
			}
			if !established && name == "exhausted before verification" {
				// An earlier never-successful import can have a missing verdict.
				// Its history stays retained; the latest aborted import has none.
				if _, err := c.pool.Exec(ctx, `UPDATE reference_pack_candidates SET health='missing',last_failure_code='payload_missing',missing_reason='staging_loss' WHERE pack_key='type_registry.host' AND pack_version='signed-fixture.1' AND current_envelope_id IS NULL`); err != nil {
					t.Fatal(err)
				}
			}
			accepted, err := c.Import(ctx, actor, uuid.NewString(), bytes.NewReader(container))
			if err != nil {
				t.Fatal(err)
			}
			jobID := uuid.MustParse(accepted.Job.JobID)
			var operation uuid.UUID
			if err := c.pool.QueryRow(ctx, `SELECT operation_id FROM reference_pack_operations WHERE job_id=$1`, jobID).Scan(&operation); err != nil {
				t.Fatal(err)
			}
			expectedOutcome := "execution_failed"
			expectedAttempts := 0
			if name == "queued inactive cancellation" {
				expectedOutcome = "canceled"
				tx, err := c.pool.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback(ctx)
				grant, err := transactions.ValidateInactiveJobTx(ctx, tx, jobID, ProfileID, accepted.Job.SubmittedAt)
				if err != nil {
					t.Fatal(err)
				}
				if err := transactions.CompleteInactiveJobTx(ctx, tx, grant, jobs.NewInactiveCancellationOutcome(), time.Now().UTC()); err != nil {
					t.Fatal(err)
				}
				if err := tx.Commit(ctx); err != nil {
					t.Fatal(err)
				}
			} else {
				execution, claimed, err := manager.Claim(ctx, jobID)
				if err != nil || !claimed {
					t.Fatal(claimed, err)
				}
				if name == "exhausted after private verification" {
					a, err := c.beginAttempt(ctx, execution, operation, c.now())
					if err != nil {
						t.Fatal(err)
					}
					member, err := c.frozenMember(ctx, a, 1)
					if err != nil {
						t.Fatal(err)
					}
					if err := c.prepareMember(ctx, a, member); err != nil {
						t.Fatal(err)
					}
					expectedAttempts = 1
				}
				if _, err := manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{Progress: jobs.Progress{Completed: 1, Total: intPtr(1)}, ResultSummary: jobs.ResultSummary{Code: ResultReferencePackImported, Message: "Imported"}}); err == nil {
					t.Fatal("unpublished private preparation became a successful Job")
				}
				// Failure of an owner participant must roll back both sides, including
				// the durable exhaustion count. The same lease can then finish normally.
				sentinel := errors.New("terminal participant unavailable")
				faultyTransactions := collaborationsupport.NewJobTransactionsWithTerminalEffects(catalog, failingTerminalEffects{sentinel}, collaborationsupport.TestWorkerRuntimeContracts(definitions))
				faulty, err := jobs.NewManager(jobs.ManagerOptions{Postgres: pool, Transactions: faultyTransactions, Catalog: catalog, Policy: policy, Now: func() time.Time { return time.Now().UTC() }})
				if err != nil {
					t.Fatal(err)
				}
				if err := faulty.RecordExecutionFailure(ctx, execution, false); !errors.Is(err, sentinel) {
					t.Fatal("terminal failure did not propagate", err)
				}
				current, err := manager.Get(ctx, jobID)
				if err != nil || current.Status != jobs.StatusRunning {
					t.Fatal("Job transition escaped rollback", current.Status, err)
				}
				var terminal bool
				if err := c.pool.QueryRow(ctx, `SELECT terminal_at IS NOT NULL FROM reference_pack_operations WHERE operation_id=$1`, operation).Scan(&terminal); err != nil || terminal {
					t.Fatal("owner transition escaped rollback", terminal, err)
				}
				if err := manager.RecordExecutionFailure(ctx, execution, false); err != nil {
					t.Fatal(err)
				}
			}
			var resultOutcome string
			var count, unfinished, pending, events int
			if err := c.pool.QueryRow(ctx, `SELECT convert_from(final_outcome,'UTF8')::jsonb->>'outcome' FROM reference_pack_operations WHERE operation_id=$1 AND terminal_at IS NOT NULL`, operation).Scan(&resultOutcome); err != nil || resultOutcome != expectedOutcome {
				t.Fatal("owner outcome", resultOutcome, err)
			}
			if err := c.pool.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE completed_at IS NULL OR outcome<>$2) FROM reference_pack_attempts WHERE operation_id=$1`, operation, expectedOutcome).Scan(&count, &unfinished); err != nil || count != expectedAttempts || unfinished != 0 {
				t.Fatal("synthetic or unfinished attempts", count, unfinished, err)
			}
			if err := c.pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_operation_keys k JOIN reference_pack_operations o USING(operation_id) WHERE k.pack_key='type_registry.host' AND o.terminal_at IS NULL`).Scan(&pending); err != nil || pending != 0 {
				t.Fatal("orphan pending work", pending, err)
			}
			if err := c.pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_events WHERE operation_id=$1`, operation).Scan(&events); err != nil || events != 0 {
				t.Fatal("abort fabricated content attestation", events, err)
			}
			if established {
				var after []byte
				if err := c.pool.QueryRow(ctx, `SELECT to_jsonb(c)::text FROM reference_pack_candidates c WHERE pack_key='type_registry.host' AND pack_version='signed-fixture.1'`).Scan(&after); err != nil || !bytes.Equal(before, after) {
					t.Fatal("abort changed established health, envelope or disablement", err)
				}
			} else {
				var health string
				var envelope, code, missing *string
				if err := c.pool.QueryRow(ctx, `SELECT health,current_envelope_id,last_failure_code,missing_reason FROM reference_pack_candidates WHERE pack_key='type_registry.host' AND pack_version='signed-fixture.1'`).Scan(&health, &envelope, &code, &missing); err != nil || health != "failed" || envelope != nil || code != nil || missing != nil {
					t.Fatal("initial abort retained staged state or fabricated verdict", health, err)
				}
			}
			if err := validateHistoricalAttempts(ctx, c.pool); err != nil {
				t.Fatal("valid retained lifecycle evidence", err)
			}
			// Idempotent owner reconciliation cannot append another audit record.
			job, err := manager.Get(ctx, jobID)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := c.pool.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			if err := (JobTerminalEffects{}).ApplyJobTerminalEffectsTx(ctx, tx, jobs.TerminalDisposition{JobID: jobID, Status: job.Status, FinishedAt: *job.FinishedAt}); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			var audits int
			if err := c.pool.QueryRow(ctx, `SELECT count(*) FROM administrative_audit_projections WHERE target_id=$1 AND action_code='reference_pack_verification_completed'`, operation.String()).Scan(&audits); err != nil || audits != 1 {
				t.Fatal("repeated terminal audit", audits, err)
			}
		})
	}
}

// Simulate retained work written by a previous process without the terminal
// participant, then repair through the same composition used before readiness.
func testRetainedTerminalJobRepair(t *testing.T, c *Coordinator, pool *pgxpool.Pool, catalog *jobs.Catalog, transactions *jobs.TransactionService, definitions []jobs.Definition, actor uuid.UUID, container []byte) {
	t.Helper()
	ctx := context.Background()
	oldTransactions := collaborationsupport.NewJobTransactionsForCatalog(catalog, collaborationsupport.TestWorkerRuntimeContracts(definitions))
	policy := jobs.ProductionRuntimePolicy()
	policy.MaximumFailures = 1
	policy.RetryDelays = nil
	now := time.Now().UTC()
	old, err := jobs.NewManager(jobs.ManagerOptions{Postgres: pool, Transactions: oldTransactions, Catalog: catalog, Policy: policy, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := c.Import(ctx, actor, uuid.NewString(), bytes.NewReader(container))
	if err != nil {
		t.Fatal(err)
	}
	jobID := uuid.MustParse(accepted.Job.JobID)
	execution, claimed, err := old.Claim(ctx, jobID)
	if err != nil || !claimed {
		t.Fatal(claimed, err)
	}
	if err := old.RecordExecutionFailure(ctx, execution, false); err != nil {
		t.Fatal(err)
	}
	now = now.Add(8 * 24 * time.Hour)
	if _, err := old.Get(ctx, jobID); !errors.Is(err, jobs.ErrNotFound) {
		t.Fatal("public Job did not expire", err)
	}
	for range 2 {
		if err := ReconcileTerminalJobs(ctx, c.pool, transactions, c.actionFinalizer); err != nil {
			t.Fatal(err)
		}
	}
	var terminal bool
	var attempts int
	var health string
	if err := c.pool.QueryRow(ctx, `SELECT o.terminal_at IS NOT NULL,(SELECT count(*) FROM reference_pack_attempts a WHERE a.operation_id=o.operation_id),c.health FROM reference_pack_operations o JOIN reference_pack_operation_members m USING(operation_id) JOIN reference_pack_candidates c USING(pack_key,pack_version) WHERE o.job_id=$1`, jobID).Scan(&terminal, &attempts, &health); err != nil || !terminal || attempts != 0 || health != "failed" {
		t.Fatal("expired terminal repair", terminal, attempts, health, err)
	}
	pending, err := c.Import(ctx, actor, uuid.NewString(), bytes.NewReader(container))
	if err != nil {
		t.Fatal("repaired key remained blocked", err)
	}
	pendingID := uuid.MustParse(pending.Job.JobID)
	if err := ReconcileTerminalJobs(ctx, c.pool, transactions, c.actionFinalizer); err != nil {
		t.Fatal(err)
	}
	if err := c.pool.QueryRow(ctx, `SELECT terminal_at IS NOT NULL FROM reference_pack_operations WHERE job_id=$1`, pendingID).Scan(&terminal); err != nil || terminal {
		t.Fatal("startup condemned queued work", terminal, err)
	}
	tx, err := c.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	grant, err := transactions.ValidateInactiveJobTx(ctx, tx, pendingID, ProfileID, pending.Job.SubmittedAt)
	if err != nil {
		t.Fatal(err)
	}
	if err := transactions.CompleteInactiveJobTx(ctx, tx, grant, jobs.NewInactiveCancellationOutcome(), time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func testHistoricalAttemptCorruption(t *testing.T, c *Coordinator) {
	t.Helper()
	ctx := context.Background()
	cases := []string{
		`UPDATE reference_pack_attempts SET canonical_result=convert_to('{"schema_id":"cartulary.reference_pack_attempt_result.v1","outcome":"succeeded","member_count":1,"failed_count":0}','UTF8') WHERE outcome='execution_failed'`,
		`UPDATE reference_pack_attempts SET canonical_result=canonical_result||decode('0a','hex') WHERE outcome='execution_failed'`,
		`UPDATE reference_pack_operations SET final_outcome=convert_to('{"failed_count":0,"member_count":0,"outcome":"execution_failed","schema_id":"cartulary.reference_pack_attempt_result.v1"}','UTF8') WHERE kind='import' AND terminal_at IS NOT NULL`,
		`UPDATE reference_pack_operations SET frozen_input=convert_to('{}','UTF8') WHERE kind='import'`,
	}
	for _, statement := range cases {
		tx, err := c.pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `ALTER TABLE reference_pack_attempts DISABLE TRIGGER reference_pack_attempts_guard`); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `ALTER TABLE reference_pack_operations DISABLE TRIGGER reference_pack_operations_guard`); err != nil {
			t.Fatal(err)
		}
		changed, err := tx.Exec(ctx, statement)
		if err != nil || changed.RowsAffected() == 0 {
			t.Fatal("corruption fixture did not change retained evidence", err)
		}
		if err := validateHistoricalAttempts(ctx, tx); !errors.Is(err, errHistoricalIntegrity) {
			t.Fatal("corrupt lifecycle evidence admitted", err)
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if err := validateHistoricalAttempts(ctx, c.pool); err != nil {
		t.Fatal("corruption test escaped rollback", err)
	}
}
