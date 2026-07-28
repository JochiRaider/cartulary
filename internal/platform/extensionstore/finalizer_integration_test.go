package extensionstore

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	manager := jobs.NewManager()
	now := time.Date(2026, 7, 24, 20, 0, 0, 0, time.UTC)
	jobTransactions := collaborationsupport.NewJobTransactions()
	manager.Configure(pool, jobTransactions, func() time.Time { return now })
	if err := manager.ConfigureExtensionContracts([]jobs.ExtensionJobContract{{
		OwnerProfileID: "test_profile",
		JobKind:        "test_profile.run_v1",
		OperationKind:  "test_profile.run",
		WorkerKind:     "test_profile.worker_v1",
		ContractSHA256: strings.Repeat("a", 64),
		ProofRequired:  true,
		MaxProofBytes:  4096,
		ResourceRefs: []jobs.ExtensionResourceRefContract{{
			Kind: "thing", MaxRefs: 1,
		}},
	}}); err != nil {
		t.Fatal(err)
	}
	fatalCount := 0
	finalizer, err := NewOwnerFinalizer(store, manager, jobTransactions, func() time.Time { return now }, func(error) { fatalCount++ })
	if err != nil {
		t.Fatal(err)
	}
	jobID := enqueueExtensionFinalizerTestJob(t, pool, now, "success")
	if _, err := manager.MarkRunning(context.Background(), jobID, jobs.Progress{Completed: 0, Total: intPointer(1)}, nil); err != nil {
		t.Fatal(err)
	}
	var successActorID uuid.UUID
	if err := pool.QueryRow(context.Background(), `SELECT submitted_by_user_id FROM jobs WHERE job_id = $1`, jobID).Scan(&successActorID); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Cancel(context.Background(), jobs.CancelParams{
		JobID: jobID, ActorUserID: successActorID, ClientTxnID: "cancel-before-final-commit",
		NormalizedRequest: []byte(`{"client_txn_id":"cancel-before-final-commit"}`),
	}); err != nil {
		t.Fatal(err)
	}
	resource, err := finalizer.FinalizeSuccess(context.Background(), JobFinalizationRequest{
		Transition: jobs.TransitionParams{
			JobID: jobID, Progress: jobs.Progress{Completed: 1, Total: intPointer(1)},
			ResultSummary: &jobs.ResultSummary{
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
	var cancellationCount int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_job_cancellation_observations WHERE job_id = $1`, jobID).Scan(&cancellationCount); err != nil {
		t.Fatal(err)
	}
	if cancellationCount != 1 {
		t.Fatalf("success/cancel race observations = %d", cancellationCount)
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

	failedJobID := enqueueExtensionFinalizerTestJob(t, pool, now.Add(time.Second), "proof-failure")
	if _, err := manager.MarkRunning(context.Background(), failedJobID, jobs.Progress{Completed: 0, Total: intPointer(1)}, nil); err != nil {
		t.Fatal(err)
	}
	_, err = finalizer.FinalizeSuccess(context.Background(), JobFinalizationRequest{
		Transition: jobs.TransitionParams{
			JobID: failedJobID, Progress: jobs.Progress{Completed: 1, Total: intPointer(1)},
			ResultSummary: &jobs.ResultSummary{Code: "done", Message: "Done."},
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

	indeterminateJobID := enqueueExtensionFinalizerTestJob(t, pool, now.Add(2*time.Second), "indeterminate")
	if _, err := manager.MarkRunning(context.Background(), indeterminateJobID, jobs.Progress{Completed: 0, Total: intPointer(1)}, nil); err != nil {
		t.Fatal(err)
	}
	finalizer.commit = func(context.Context, pgx.Tx) error { return errors.New("commit acknowledgement unavailable") }
	_, err = finalizer.FinalizeSuccess(context.Background(), JobFinalizationRequest{
		Transition: jobs.TransitionParams{
			JobID: indeterminateJobID, Progress: jobs.Progress{Completed: 1, Total: intPointer(1)},
			ResultSummary: &jobs.ResultSummary{Code: "done", Message: "Done."},
		},
		FinalCommitID: "commit:indeterminate",
	})
	if !errors.Is(err, ErrIndeterminateCommit) || fatalCount != 1 {
		t.Fatalf("indeterminate commit err=%v fatal_count=%d", err, fatalCount)
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
	manager := jobs.NewManager()
	now := time.Date(2026, 7, 24, 20, 0, 0, 0, time.UTC)
	manager.Configure(pool, collaborationsupport.NewJobTransactions(), func() time.Time { return now })
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
		"test_profile", "test_profile.run_v1", key,
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
	resource, err := collaborationsupport.NewJobTransactions().CreateQueuedTx(context.Background(), tx, jobs.CreateParams{
		Scope: jobs.Scope{Kind: jobs.ScopeKindDeployment}, SubmittedByUserID: actorID,
		Cancelable: true, Progress: jobs.Progress{Completed: 0, Total: intPointer(1)},
		HandlerName: "test_profile.worker_v1", Extension: admission,
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
