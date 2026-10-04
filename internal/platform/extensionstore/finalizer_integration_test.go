package extensionstore

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestOwnerFinalizerAtomicSuccessAndFailure_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "extension-job-finalizer")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(context.Background(), `
CREATE TABLE extension_job_finalizer_test_effects (
    effect_id text PRIMARY KEY,
    created_at timestamptz NOT NULL
)`); err != nil {
		t.Fatal(err)
	}
	store, err := New(pool, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	definitions := collaborationsupport.TestJobDefinitions()
	definitions[1].Extension.ResourceRefs = []jobs.ExtensionResourceRefContract{{Kind: "thing", MaxRefs: 1}}
	catalog, err := jobs.NewCatalog(definitions)
	if err != nil {
		t.Fatal(err)
	}
	jobTransactions := collaborationsupport.NewJobTransactionsForCatalog(catalog)
	manager, err := jobs.NewManager(jobs.ManagerOptions{
		Postgres: pool, Transactions: jobTransactions, Catalog: catalog,
		Policy: jobs.ProductionRuntimePolicy(), Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	fatalCount := 0
	finalizer, err := NewOwnerFinalizer(store, jobTransactions, collaborationsupport.NewJobOwnerTransactionAdapters(), func() time.Time { return now }, func(error) { fatalCount++ })
	if err != nil {
		t.Fatal(err)
	}
	jobID := enqueueExtensionFinalizerTestJob(t, pool, now, "success")
	execution, claimed, err := manager.Claim(context.Background(), jobID)
	if err != nil || !claimed {
		t.Fatalf("claim success execution = %v/%v", claimed, err)
	}
	resource, err := finalizer.FinalizeSuccess(context.Background(), JobFinalizationRequest{
		Execution: execution,
		Completion: jobs.SuccessCompletion{
			Progress: jobs.Progress{Completed: 1, Total: intPointer(1)},
			ResultSummary: jobs.ResultSummary{
				Code: "done", Message: "Done.",
				ResourceRefs: []jobs.ResourceRef{{Kind: "thing", ID: "1", Route: "/things/1"}},
			},
		},
		FinalCommitID: "commit:success",
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO extension_job_finalizer_test_effects (effect_id, created_at) VALUES ('success', $1)`, now)
			return err
		},
	})
	if err != nil || resource.Status != jobs.StatusSucceeded {
		t.Fatalf("finalize success = %#v/%v", resource, err)
	}
	var effectCount, proofCount int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_job_finalizer_test_effects WHERE effect_id = 'success'`).Scan(&effectCount); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_job_commit_proofs WHERE job_id = $1`, jobID).Scan(&proofCount); err != nil {
		t.Fatal(err)
	}
	if effectCount != 1 || proofCount != 1 {
		t.Fatalf("atomic success effect=%d proof=%d", effectCount, proofCount)
	}
	var replayStatus string
	if err := pool.QueryRow(context.Background(), `
SELECT response_json->>'status'
  FROM route_idempotency
 WHERE client_txn_id = 'success'
`).Scan(&replayStatus); err != nil {
		t.Fatal(err)
	}
	if replayStatus != jobs.StatusSucceeded {
		t.Fatalf("final idempotency status = %q", replayStatus)
	}

	canceledJobID := enqueueExtensionFinalizerTestJob(t, pool, now, "cancel-race")
	canceledExecution, claimed, err := manager.Claim(context.Background(), canceledJobID)
	if err != nil || !claimed {
		t.Fatalf("claim canceled execution = %v/%v", claimed, err)
	}
	var canceledActorID uuid.UUID
	if err := pool.QueryRow(context.Background(), `SELECT submitted_by_user_id FROM jobs WHERE job_id = $1`, canceledJobID).Scan(&canceledActorID); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Cancel(context.Background(), jobs.CancelParams{
		JobID: canceledJobID, ActorUserID: canceledActorID, ClientTxnID: "cancel-before-final-commit",
		NormalizedRequest: []byte(`{"client_txn_id":"cancel-before-final-commit"}`),
	}); err != nil {
		t.Fatal(err)
	}
	_, err = finalizer.FinalizeSuccess(context.Background(), JobFinalizationRequest{
		Execution: canceledExecution,
		Completion: jobs.SuccessCompletion{
			Progress: jobs.Progress{Completed: 1, Total: intPointer(1)},
			ResultSummary: jobs.ResultSummary{
				Code: "done", Message: "Done.",
				ResourceRefs: []jobs.ResourceRef{{Kind: "thing", ID: "cancel-race", Route: "/things/cancel-race"}},
			},
		},
		FinalCommitID: "commit:cancel-race",
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO extension_job_finalizer_test_effects (effect_id, created_at) VALUES ('cancel-race', $1)`, now)
			return err
		},
	})
	if !errors.Is(err, jobs.ErrCancellationRequested) {
		t.Fatalf("cancel-wins finalization error = %v, want cancellation requested", err)
	}
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_job_finalizer_test_effects WHERE effect_id = 'cancel-race'`).Scan(&effectCount); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_job_commit_proofs WHERE job_id = $1`, canceledJobID).Scan(&proofCount); err != nil {
		t.Fatal(err)
	}
	var cancellationCount int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_job_cancellation_observations WHERE job_id = $1`, canceledJobID).Scan(&cancellationCount); err != nil {
		t.Fatal(err)
	}
	if effectCount != 0 || proofCount != 0 || cancellationCount != 1 {
		t.Fatalf("cancel-wins side effects=%d proofs=%d observations=%d", effectCount, proofCount, cancellationCount)
	}
	_, err = finalizer.FinalizeCancellation(context.Background(), JobCancellationFinalizationRequest{Execution: canceledExecution, Completion: jobs.CancellationCompletion{Progress: jobs.Progress{Completed: 0, Total: intPointer(1)}}, Mutate: func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO extension_job_finalizer_test_effects(effect_id,created_at) VALUES('cancel-evidence',$1)`, now); err != nil {
			return err
		}
		return errors.New("injected owner evidence failure")
	}})
	if err == nil {
		t.Fatal("cancellation committed despite failed owner evidence")
	}
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_job_finalizer_test_effects WHERE effect_id='cancel-evidence'`).Scan(&effectCount); err != nil || effectCount != 0 {
		t.Fatal("failed cancellation leaked owner state", err)
	}
	canceled, err := finalizer.FinalizeCancellation(context.Background(), JobCancellationFinalizationRequest{Execution: canceledExecution, Completion: jobs.CancellationCompletion{Progress: jobs.Progress{Completed: 0, Total: intPointer(1)}}, Mutate: func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO extension_job_finalizer_test_effects(effect_id,created_at) VALUES('cancel-evidence',$1)`, now)
		return err
	}})
	if err != nil || canceled.Status != jobs.StatusCanceled {
		t.Fatalf("atomic cancellation: %#v/%v", canceled, err)
	}
	if err := pool.QueryRow(context.Background(), `SELECT response_json->>'status' FROM route_idempotency WHERE client_txn_id='cancel-race'`).Scan(&replayStatus); err != nil || replayStatus != jobs.StatusCanceled {
		t.Fatalf("cancellation did not update original admission receipt: %s/%v", replayStatus, err)
	}

	failedJobID := enqueueExtensionFinalizerTestJob(t, pool, now, "proof-failure")
	// A timeout already classified by the shared deadline policy wins over a
	// cancellation observed at expiry. Ordinary failure may not bypass cancel.
	timeoutJobID := enqueueExtensionFinalizerTestJob(t, pool, now, "timeout-cancel-tie")
	timeoutExecution, timeoutClaimed, err := manager.Claim(context.Background(), timeoutJobID)
	if err != nil || !timeoutClaimed {
		t.Fatal("claim timeout", timeoutClaimed, err)
	}
	var timeoutActor uuid.UUID
	if err := pool.QueryRow(context.Background(), `SELECT submitted_by_user_id FROM jobs WHERE job_id=$1`, timeoutJobID).Scan(&timeoutActor); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Cancel(context.Background(), jobs.CancelParams{JobID: timeoutJobID, ActorUserID: timeoutActor, ClientTxnID: "cancel-at-expiry", NormalizedRequest: []byte(`{"client_txn_id":"cancel-at-expiry"}`)}); err != nil {
		t.Fatal(err)
	}
	timeoutRequest := JobFailureFinalizationRequest{Execution: timeoutExecution, Completion: jobs.FailureCompletion{Progress: jobs.Progress{Completed: 0, Total: intPointer(1)}, ErrorSummary: jobs.ErrorSummary{Code: "verification_timeout", Message: "Verification timed out."}}, Mutate: func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO extension_job_finalizer_test_effects(effect_id,created_at) VALUES('timeout-evidence',$1)`, now)
		return err
	}}
	if _, err := finalizer.FinalizeFailure(context.Background(), timeoutRequest); !errors.Is(err, jobs.ErrCancellationRequested) {
		t.Fatal("ordinary failure bypassed cancellation", err)
	}
	timedOut, err := finalizer.FinalizeTimeout(context.Background(), timeoutRequest)
	if err != nil || timedOut.Status != jobs.StatusFailed || timedOut.ErrorSummary == nil || timedOut.ErrorSummary.Code != "verification_timeout" {
		t.Fatal("classified timeout lost to cancellation", timedOut, err)
	}
	if err := pool.QueryRow(context.Background(), `SELECT response_json->>'status' FROM route_idempotency WHERE client_txn_id='timeout-cancel-tie'`).Scan(&replayStatus); err != nil || replayStatus != jobs.StatusFailed {
		t.Fatal("timeout receipt not atomic", replayStatus, err)
	}
	failedExecution, claimed, err := manager.Claim(context.Background(), failedJobID)
	if err != nil || !claimed {
		t.Fatalf("claim failed execution = %v/%v", claimed, err)
	}
	_, err = finalizer.FinalizeSuccess(context.Background(), JobFinalizationRequest{
		Execution: failedExecution,
		Completion: jobs.SuccessCompletion{
			Progress:      jobs.Progress{Completed: 1, Total: intPointer(1)},
			ResultSummary: jobs.ResultSummary{Code: "done", Message: "Done."},
		},
		FinalCommitID: "invalid commit id with spaces",
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO extension_job_finalizer_test_effects (effect_id, created_at) VALUES ('proof-failure', $1)`, now)
			return err
		},
	})
	if err == nil {
		t.Fatal("expected proof insertion failure")
	}
	failedResource, getErr := manager.Get(context.Background(), failedJobID)
	if getErr != nil || failedResource.Status != jobs.StatusRunning {
		t.Fatalf("failed finalization job = %#v/%v", failedResource, getErr)
	}
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_job_finalizer_test_effects WHERE effect_id = 'proof-failure'`).Scan(&effectCount); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_job_commit_proofs WHERE job_id = $1`, failedJobID).Scan(&proofCount); err != nil {
		t.Fatal(err)
	}
	if effectCount != 0 || proofCount != 0 {
		t.Fatalf("failed finalization leaked effect=%d proof=%d", effectCount, proofCount)
	}

	acknowledgementJobID := enqueueExtensionFinalizerTestJob(t, pool, now, "lost-acknowledgement")
	acknowledgementExecution, claimed, err := manager.Claim(context.Background(), acknowledgementJobID)
	if err != nil || !claimed {
		t.Fatalf("claim lost acknowledgement: %v/%v", claimed, err)
	}
	finalizer.commit = func(ctx context.Context, tx pgx.Tx) error {
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return errors.New("commit acknowledgement unavailable")
	}
	acknowledged, err := finalizer.FinalizeSuccess(context.Background(), JobFinalizationRequest{
		Execution:     acknowledgementExecution,
		Completion:    jobs.SuccessCompletion{Progress: jobs.Progress{Completed: 1, Total: intPointer(1)}, ResultSummary: jobs.ResultSummary{Code: "done", Message: "Done."}},
		FinalCommitID: "commit:lost-acknowledgement",
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO extension_job_finalizer_test_effects(effect_id,created_at) VALUES('lost-acknowledgement',$1)`, now)
			return err
		},
	})
	if err != nil || acknowledged.Status != jobs.StatusSucceeded || fatalCount != 0 {
		t.Fatalf("proven commit lost to acknowledgement: %#v / %v / fatal=%d", acknowledged, err, fatalCount)
	}
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_job_finalizer_test_effects WHERE effect_id='lost-acknowledgement'`).Scan(&effectCount); err != nil || effectCount != 1 {
		t.Fatal("committed effect missing", err)
	}
	// The synchronous operation path uses exact owner receipts, including when
	// the HTTP context is canceled after a committed mutation.
	baseFinalizer, err := NewOwnerMutationFinalizer(func(error) { fatalCount++ })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewOwnerMutationFinalizer(nil); err == nil {
		t.Fatal("missing fatal integrity sink admitted")
	}
	for _, cancelAfterCommit := range []bool{false, true} {
		proofCtx, cancel := context.WithCancel(context.Background())
		effect := "sync-commit"
		if cancelAfterCommit {
			effect = "sync-canceled-after-commit"
		}
		ownerTx, err := pool.Begin(proofCtx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ownerTx.Exec(proofCtx, `INSERT INTO extension_job_finalizer_test_effects(effect_id,created_at) VALUES($1,$2)`, effect, now); err != nil {
			t.Fatal(err)
		}
		baseFinalizer.commit = func(ctx context.Context, tx pgx.Tx) error {
			if err := tx.Commit(ctx); err != nil {
				return err
			}
			if cancelAfterCommit {
				cancel()
			}
			return errors.New("owner acknowledgement unavailable")
		}
		err = baseFinalizer.CommitOwnerMutation(proofCtx, ownerTx, func(ctx context.Context) (bool, error) {
			var found bool
			err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM extension_job_finalizer_test_effects WHERE effect_id=$1)`, effect).Scan(&found)
			return found, err
		})
		cancel()
		if err != nil || fatalCount != 0 {
			t.Fatalf("synchronous proven commit: %v, fatal=%d", err, fatalCount)
		}
	}
	finalizer.commit = func(ctx context.Context, tx pgx.Tx) error {
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return errors.New("commit acknowledgement unavailable")
	}
	failedAcknowledgementID := enqueueExtensionFinalizerTestJob(t, pool, now, "failed-lost-acknowledgement")
	failedAcknowledgementExecution, claimed, err := manager.Claim(context.Background(), failedAcknowledgementID)
	if err != nil || !claimed {
		t.Fatal("claim failed result", err)
	}
	failedAcknowledgement, err := finalizer.FinalizeFailure(context.Background(), JobFailureFinalizationRequest{Execution: failedAcknowledgementExecution, Completion: jobs.FailureCompletion{Progress: jobs.Progress{Completed: 1, Total: intPointer(1)}, ErrorSummary: jobs.ErrorSummary{Code: "reference_pack_verification_failed", Message: "Verification failed."}}, Mutate: func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO extension_job_finalizer_test_effects(effect_id,created_at) VALUES('failed-lost-acknowledgement',$1)`, now)
		return err
	}})
	if err != nil || failedAcknowledgement.Status != jobs.StatusFailed || fatalCount != 0 {
		t.Fatalf("committed failure was misclassified: %#v/%v", failedAcknowledgement, err)
	}

	indeterminateJobID := enqueueExtensionFinalizerTestJob(t, pool, now, "indeterminate")
	indeterminateExecution, claimed, err := manager.Claim(context.Background(), indeterminateJobID)
	if err != nil || !claimed {
		t.Fatalf("claim indeterminate execution = %v/%v", claimed, err)
	}
	finalizer.commit = func(context.Context, pgx.Tx) error { return errors.New("commit acknowledgement unavailable") }
	_, err = finalizer.FinalizeSuccess(context.Background(), JobFinalizationRequest{
		Execution: indeterminateExecution,
		Completion: jobs.SuccessCompletion{
			Progress:      jobs.Progress{Completed: 1, Total: intPointer(1)},
			ResultSummary: jobs.ResultSummary{Code: "done", Message: "Done."},
		},
		FinalCommitID: "commit:indeterminate",
	})
	if err == nil || errors.Is(err, ErrIndeterminateCommit) || fatalCount != 0 {
		t.Fatalf("proven rollback was classified as indeterminate: err=%v fatal_count=%d", err, fatalCount)
	}
	for _, status := range []string{jobs.StatusSucceeded, jobs.StatusFailed, "timeout", jobs.StatusCanceled} {
		t.Run("canceled context after committed "+status, func(t *testing.T) {
			id := enqueueExtensionFinalizerTestJob(t, pool, now, "postcommit-"+status)
			execution, claimed, err := manager.Claim(context.Background(), id)
			if err != nil || !claimed {
				t.Fatal("claim classification fixture", err)
			}
			if status == jobs.StatusCanceled {
				var actor uuid.UUID
				if err := pool.QueryRow(context.Background(), `SELECT submitted_by_user_id FROM jobs WHERE job_id=$1`, id).Scan(&actor); err != nil {
					t.Fatal(err)
				}
				if _, err := manager.Cancel(context.Background(), jobs.CancelParams{JobID: id, ActorUserID: actor, ClientTxnID: "postcommit-cancel", NormalizedRequest: []byte(`{"client_txn_id":"postcommit-cancel"}`)}); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			finalizer.commit = func(ctx context.Context, tx pgx.Tx) error {
				if err := tx.Commit(ctx); err != nil {
					return err
				}
				cancel()
				return errors.New("acknowledgement lost after caller cancellation")
			}
			var result jobs.Resource
			failure := JobFailureFinalizationRequest{Execution: execution, Completion: jobs.FailureCompletion{Progress: jobs.Progress{Completed: 0, Total: intPointer(1)}, ErrorSummary: jobs.ErrorSummary{Code: "reference_pack_verification_failed", Message: "Verification failed."}}}
			switch status {
			case jobs.StatusSucceeded:
				result, err = finalizer.FinalizeSuccess(ctx, JobFinalizationRequest{Execution: execution, Completion: jobs.SuccessCompletion{Progress: jobs.Progress{Completed: 1, Total: intPointer(1)}, ResultSummary: jobs.ResultSummary{Code: "done", Message: "Done."}}, FinalCommitID: "commit:postcommit-success"})
			case jobs.StatusFailed:
				result, err = finalizer.FinalizeFailure(ctx, failure)
			case "timeout":
				result, err = finalizer.FinalizeTimeout(ctx, failure)
			case jobs.StatusCanceled:
				result, err = finalizer.FinalizeCancellation(ctx, JobCancellationFinalizationRequest{Execution: execution, Completion: jobs.CancellationCompletion{Progress: jobs.Progress{Completed: 0, Total: intPointer(1)}}})
			}
			expectedStatus := status
			if status == "timeout" {
				expectedStatus = jobs.StatusFailed
			}
			if ctx.Err() == nil || err != nil || result.Status != expectedStatus || fatalCount != 0 {
				t.Fatalf("committed outcome lost to context cancellation: %s/%s/%v fatal=%d", status, result.Status, err, fatalCount)
			}
			retained, err := manager.Get(context.Background(), id)
			if err != nil || retained.Status != expectedStatus {
				t.Fatal("terminal outcome not retained", err)
			}
		})
	}
	for _, code := range []string{"40001", "40P01", "23505", "23503", "23514"} {
		tx, err := pool.Begin(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		baseFinalizer.commit = func(ctx context.Context, tx pgx.Tx) error {
			if err := tx.Rollback(ctx); err != nil {
				return err
			}
			return &pgconn.PgError{Code: code}
		}
		err = baseFinalizer.CommitOwnerMutation(context.Background(), tx, func(context.Context) (bool, error) {
			t.Fatal("proven server rollback attempted a proof read")
			return false, nil
		})
		var state *pgconn.PgError
		if !errors.As(err, &state) || state.Code != code || fatalCount != 0 {
			t.Fatal("server rollback not classified as absent", code, err)
		}
	}
	unreadableID := enqueueExtensionFinalizerTestJob(t, pool, now, "unreadable-commit")
	unreadableExecution, claimed, err := manager.Claim(context.Background(), unreadableID)
	if err != nil || !claimed {
		t.Fatal("claim unreadable result", err)
	}
	finalizer.commit = func(ctx context.Context, tx pgx.Tx) error {
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		pool.Close()
		return errors.New("connection and acknowledgement lost")
	}
	_, err = finalizer.FinalizeSuccess(context.Background(), JobFinalizationRequest{Execution: unreadableExecution, Completion: jobs.SuccessCompletion{Progress: jobs.Progress{Completed: 1, Total: intPointer(1)}, ResultSummary: jobs.ResultSummary{Code: "done", Message: "Done."}}, FinalCommitID: "commit:unreadable"})
	if !errors.Is(err, ErrIndeterminateCommit) || fatalCount != 1 {
		t.Fatalf("unknown mutation did not enter fatal path: %v/%d", err, fatalCount)
	}
}

func TestExtensionCancellationObservationIsAtomic_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "extension-job-cancellation")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	now := time.Date(2026, 7, 24, 20, 0, 0, 0, time.UTC)
	catalog := collaborationsupport.NewJobCatalog()
	jobTransactions := collaborationsupport.NewJobTransactionsForCatalog(catalog)
	manager, err := jobs.NewManager(jobs.ManagerOptions{
		Postgres: pool, Transactions: jobTransactions, Catalog: catalog,
		Policy: jobs.ProductionRuntimePolicy(), Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	jobID := enqueueExtensionFinalizerTestJob(t, pool, now, "cancel")
	var actorID uuid.UUID
	if err := pool.QueryRow(context.Background(), `SELECT submitted_by_user_id FROM jobs WHERE job_id = $1`, jobID).Scan(&actorID); err != nil {
		t.Fatal(err)
	}
	result, err := manager.Cancel(context.Background(), jobs.CancelParams{
		JobID: jobID, ActorUserID: actorID, ClientTxnID: "cancel-job",
		NormalizedRequest: []byte(`{"client_txn_id":"cancel-job"}`),
	})
	if err != nil || result.Resource.Status != jobs.StatusCancelRequested {
		t.Fatalf("cancel extension job = %#v/%v", result, err)
	}
	var observations int
	if err := pool.QueryRow(context.Background(), `
SELECT count(*)
  FROM extension_job_cancellation_observations
 WHERE job_id = $1
   AND observed_before_final_commit
`, jobID).Scan(&observations); err != nil {
		t.Fatal(err)
	}
	if observations != 1 {
		t.Fatalf("cancellation observations = %d", observations)
	}
}

func enqueueExtensionFinalizerTestJob(t *testing.T, pool *pgxpool.Pool, now time.Time, clientTxnID string) uuid.UUID {
	t.Helper()
	actorID := uuid.New()
	if _, err := pool.Exec(context.Background(), `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'Extension Job Finalizer', 'hash', false, true, true)
`, actorID, actorID.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	key := authn.RouteIdempotencyKey{
		RouteKey: "test.extension.run", ActorUserID: actorID,
		ScopeKey: "deployment", ClientTxnID: clientTxnID,
	}
	normalized := []byte(`{"client_txn_id":"` + clientTxnID + `"}`)
	admission, err := jobs.NewExtensionJobAdmission(
		"test_profile", jobs.NewRouteIdempotencyKey(key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID),
		jobs.Scope{Kind: jobs.ScopeKindDeployment}, normalized,
	)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	resource, err := collaborationsupport.NewJobTransactions().CreateQueuedTx(context.Background(), tx, jobs.EnqueueParams{
		JobKind: "test_profile.run_v1", Scope: jobs.Scope{Kind: jobs.ScopeKindDeployment}, SubmittedByUserID: actorID,
		Cancelable: true, Progress: jobs.Progress{Completed: 0, Total: intPointer(1)}, Extension: admission,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	requestHash, err := hex.DecodeString(admission.NormalizedRequestSHA256)
	if err != nil {
		t.Fatal(err)
	}
	if err := authn.InsertRouteIdempotencyPayload(context.Background(), tx, key, nil, requestHash, 202, resource); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}
	return uuid.MustParse(resource.JobID)
}

func intPointer(value int) *int {
	return &value
}
