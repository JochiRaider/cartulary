package reporting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestReportingPayloadReadsRequireLiveExecution_Integration(t *testing.T) {
	ctx := context.Background()
	database := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "reporting-execution-input")
	pool, err := pgxpool.New(ctx, database.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	actor, incident := uuid.New(), uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash) VALUES($1,$2,'Execution fixture','unused')`, actor, actor.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,$2,$2,'Execution fixture','active',$3,$3)`, incident, incident.String(), actor); err != nil {
		t.Fatal(err)
	}
	definitions := []jobs.Definition{{JobKind: "test.reporting_input", ProgressUnitID: "test.reporting_input.operation.v1", HandlerName: "test.reporting_input"}}
	catalog, err := jobs.NewCatalog(definitions)
	if err != nil {
		t.Fatal(err)
	}
	transactions := collaborationsupport.NewJobTransactionsForCatalog(catalog, collaborationsupport.TestWorkerRuntimeContracts(definitions))
	manager, err := jobs.NewManager(jobs.ManagerOptions{Postgres: pool, Transactions: transactions, Catalog: catalog, Policy: jobs.ProductionRuntimePolicy(), Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	store := &Store{pool: pool, jobTransactions: transactions}
	admitted := func(t *testing.T) jobs.Execution {
		t.Helper()
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		resource, err := transactions.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{JobKind: definitions[0].JobKind, Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incident}, SubmittedByUserID: actor, Cancelable: true, Progress: jobs.Progress{Completed: 0}}, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		id := uuid.MustParse(resource.JobID)
		// A minimal payload is sufficient for the kind read. Full model decoding
		// must never run for a fenced execution, including when bytes are invalid.
		if _, err := tx.Exec(ctx, `INSERT INTO reporting_job_payloads(job_id,job_kind,incident_id,actor_user_id,request_json,created_at,snapshot_admission_json,snapshot_admission_sha256) VALUES($1,'snapshot_create',$2,$3,NULL,now(),'{"schema_id":"cartulary.reporting_snapshot_admission.v1"}',repeat('0',64))`, id, incident, actor); err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		execution, claimed, err := manager.Claim(ctx, id)
		if err != nil || !claimed {
			t.Fatal("claim", claimed, err)
		}
		if kind, err := store.ReportingJobKind(ctx, execution); err != nil || kind != reportingJobKindSnapshotCreate {
			t.Fatal("live payload", kind, err)
		}
		return execution
	}
	for _, state := range []string{"expired", "replaced", "cancel_requested", "terminal"} {
		t.Run(state, func(t *testing.T) {
			execution := admitted(t)
			expected := jobs.ErrExecutionLost
			switch state {
			case "expired":
				_, err = pool.Exec(ctx, `UPDATE jobs SET handler_lease_expires_at=clock_timestamp()-interval '1 second' WHERE job_id=$1`, execution.JobID())
			case "replaced":
				if err = manager.ReleaseExecution(ctx, execution); err == nil {
					var replacement jobs.Execution
					var claimed bool
					replacement, claimed, err = manager.Claim(ctx, execution.JobID())
					if err != nil || !claimed {
						t.Fatal("replacement claim", claimed, err)
					}
					if _, err = store.ReportingJobKind(ctx, replacement); err != nil {
						t.Fatal("replacement cannot read", err)
					}
				}
			case "cancel_requested":
				_, err = manager.Cancel(ctx, jobs.CancelParams{JobID: execution.JobID(), ActorUserID: actor, ClientTxnID: "cancel-fenced-input", NormalizedRequest: []byte(`{"client_txn_id":"cancel-fenced-input","reason":null}`)})
				expected = jobs.ErrCancellationRequested
			case "terminal":
				_, err = manager.CompleteFailed(ctx, execution, jobs.FailureCompletion{Progress: jobs.Progress{Completed: 0}, ErrorSummary: jobs.ErrorSummary{Code: "internal_error", Message: "Fixture failure."}})
			}
			if err != nil {
				t.Fatal(err)
			}
			locked, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer locked.Rollback(ctx)
			if _, err = locked.Exec(ctx, `LOCK TABLE reporting_job_payloads IN ACCESS EXCLUSIVE MODE`); err != nil {
				t.Fatal(err)
			}
			readers := []struct {
				name string
				read func(context.Context) error
			}{
				{"kind", func(ctx context.Context) error { _, err := store.ReportingJobKind(ctx, execution); return err }},
				{"release", func(ctx context.Context) error { _, err := store.ReleasePayloadForJob(ctx, execution); return err }},
				{"preview", func(ctx context.Context) error {
					_, err := store.CompositionPreviewPayloadForJob(ctx, execution)
					return err
				}},
			}
			for _, reader := range readers {
				// The lock makes any premature owner-payload access observable as a
				// deadline, instead of permitting an unsafe read to pass unnoticed.
				readCtx, cancel := context.WithTimeout(ctx, time.Second)
				readErr := reader.read(readCtx)
				cancel()
				if !errors.Is(readErr, expected) {
					t.Fatalf("%s: got %v want %v", reader.name, readErr, expected)
				}
			}
		})
	}
	t.Run("expiry is evaluated after transaction start", func(t *testing.T) {
		execution := admitted(t)
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		var pastStart, beforeNow bool
		err = tx.QueryRow(ctx, `UPDATE jobs SET handler_lease_expires_at=transaction_timestamp()+(clock_timestamp()-transaction_timestamp())/2 WHERE job_id=$1 RETURNING handler_lease_expires_at>transaction_timestamp(),handler_lease_expires_at<clock_timestamp()`, execution.JobID()).Scan(&pastStart, &beforeNow)
		if err != nil || !pastStart || !beforeNow {
			t.Fatal("expiry fixture", pastStart, beforeNow, err)
		}
		if err = transactions.ValidateExecutionTx(ctx, tx, execution); !errors.Is(err, jobs.ErrExecutionLost) {
			t.Fatal("expired lease admitted using transaction start time", err)
		}
	})
}
