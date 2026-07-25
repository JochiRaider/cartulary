package extensionassembly_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestInactiveExtensionJobReconciliation_ServiceBacked(t *testing.T) {
	ctx := context.Background()
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "extension-job-reconciliation")
	pool, err := pgxpool.New(ctx, testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	store, err := extensionstore.New(pool, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 24, 20, 30, 0, 0, time.UTC)
	manager := jobs.NewManager()
	manager.Configure(pool, func() time.Time { return now })

	provenJob, provenAdmission := enqueueInactiveJob(t, pool, now, "proven")
	canceledJob, _ := enqueueInactiveJob(t, pool, now.Add(time.Second), "canceled")
	unclaimedJob, _ := enqueueInactiveJob(t, pool, now.Add(2*time.Second), "unclaimed")

	terminalSummary, terminalJSON, refsJSON, terminalDigest, err :=
		jobs.CanonicalExtensionTerminalSuccess(reconciliationPlatformContract(), &jobs.ResultSummary{
			Code: "committed", Message: "Committed.",
		})
	if err != nil || terminalSummary.Code != "committed" {
		t.Fatalf("canonical terminal success = %#v/%v", terminalSummary, err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := extensionstore.InsertJobCommitProof(ctx, tx, extensionstore.JobCommitProof{
		JobID:                   provenJob,
		OwnerProfileID:          "test_profile",
		OperationKind:           "test_profile.run",
		FinalCommitID:           "commit:proven",
		IdempotencyIdentity:     provenAdmission.IdempotencyIdentity,
		NormalizedRequestSHA256: provenAdmission.NormalizedRequestSHA256,
		TerminalResult:          terminalJSON,
		TerminalResultSHA256:    terminalDigest,
		ResourceRefs:            refsJSON,
		CommittedAt:             now.Add(3 * time.Second),
	}); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	cancelInactiveJob(t, pool, manager, provenJob, "cancel-proven-after-commit")

	cancelInactiveJob(t, pool, manager, canceledJob, "cancel-reconciliation")

	adapter, err := extensionassembly.NewInactiveJobStore(store, func() time.Time { return now.Add(4 * time.Second) })
	if err != nil {
		t.Fatal(err)
	}
	if err := extensions.ReconcileInactiveExtensionJobs(
		ctx,
		adapter,
		[]string{"test_profile"},
		[]extensions.JobKindContract{reconciliationLogicalContract()},
		3,
		nil,
	); err != nil {
		t.Fatal(err)
	}

	assertJobTerminalState(t, pool, provenJob, jobs.StatusSucceeded, "committed")
	assertJobTerminalState(t, pool, canceledJob, jobs.StatusCanceled, "job_canceled")
	assertJobTerminalState(t, pool, unclaimedJob, jobs.StatusFailed, "extension_profile_unclaimed")

	raceFirst, _ := enqueueInactiveJob(t, pool, now.Add(5*time.Second), "race-first")
	raceSecond, _ := enqueueInactiveJob(t, pool, now.Add(6*time.Second), "race-second")
	loaded, err := store.LoadInactiveJobRecords(ctx, "test_profile", 3)
	if err != nil || len(loaded) != 2 {
		t.Fatalf("load race candidates = %d/%v", len(loaded), err)
	}
	cancelInactiveJob(t, pool, manager, raceSecond, "cancel-after-classification")
	commitOutcome, err := store.ApplyInactiveJobOutcomeRecords(ctx, "test_profile", []extensionstore.InactiveJobOutcomeRecord{
		{
			JobID: raceFirst, SubmittedAt: loaded[0].SubmittedAt, Status: jobs.StatusFailed,
			TerminalResult: []byte(`{"code":"extension_profile_unclaimed","message":"Extension profile is not claimed.","retryable":false,"details":{}}`),
			EvidenceKind:   "absence",
		},
		{
			JobID: raceSecond, SubmittedAt: loaded[1].SubmittedAt, Status: jobs.StatusFailed,
			TerminalResult: []byte(`{"code":"extension_profile_unclaimed","message":"Extension profile is not claimed.","retryable":false,"details":{}}`),
			EvidenceKind:   "absence",
		},
	}, now.Add(7*time.Second))
	if err == nil || commitOutcome != extensionstore.CommitAbsent {
		t.Fatalf("evidence race commit = %s/%v", commitOutcome, err)
	}
	assertJobStatus(t, pool, raceFirst, jobs.StatusQueued)
	assertJobStatus(t, pool, raceSecond, jobs.StatusCancelRequested)
}

func cancelInactiveJob(t *testing.T, pool *pgxpool.Pool, manager *jobs.Manager, jobID uuid.UUID, clientTxnID string) {
	t.Helper()
	var actorID uuid.UUID
	if err := pool.QueryRow(context.Background(), `SELECT submitted_by_user_id FROM jobs WHERE job_id = $1`, jobID).Scan(&actorID); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Cancel(context.Background(), jobs.CancelParams{
		JobID: jobID, ActorUserID: actorID, ClientTxnID: clientTxnID,
		NormalizedRequest: []byte(`{"client_txn_id":"` + clientTxnID + `"}`),
	}); err != nil {
		t.Fatal(err)
	}
}

func enqueueInactiveJob(t *testing.T, pool *pgxpool.Pool, now time.Time, clientTxnID string) (uuid.UUID, *jobs.ExtensionJobAdmission) {
	t.Helper()
	ctx := context.Background()
	actorID := uuid.New()
	if _, err := pool.Exec(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'Inactive Job Reconciliation', 'hash', false, true, true)
`, actorID, actorID.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	key := authn.RouteIdempotencyKey{
		RouteKey: "test.reconciliation.run", ActorUserID: actorID,
		ScopeKey: "deployment", ClientTxnID: clientTxnID,
	}
	normalized := []byte(`{"client_txn_id":"` + clientTxnID + `"}`)
	admission, err := jobs.NewExtensionJobAdmission(
		"test_profile",
		"test_profile.run_v1",
		key,
		jobs.Scope{Kind: jobs.ScopeKindDeployment},
		normalized,
	)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	resource, err := jobs.CreateQueuedTx(ctx, tx, jobs.CreateParams{
		Scope: jobs.Scope{Kind: jobs.ScopeKindDeployment}, SubmittedByUserID: actorID,
		Cancelable: true, Progress: jobs.Progress{Completed: 0},
		HandlerName: "test_profile.worker_v1", Extension: admission,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	return uuid.MustParse(resource.JobID), admission
}

func reconciliationPlatformContract() jobs.ExtensionJobContract {
	return jobs.ExtensionJobContract{
		OwnerProfileID: "test_profile",
		JobKind:        "test_profile.run_v1",
		OperationKind:  "test_profile.run",
		WorkerKind:     "test_profile.worker_v1",
		ContractSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ProofRequired:  true,
		MaxProofBytes:  4096,
	}
}

func reconciliationLogicalContract() extensions.JobKindContract {
	return extensions.JobKindContract{
		ProfileID:                   "test_profile",
		JobKind:                     "test_profile.run_v1",
		OperationKind:               "test_profile.run",
		ProofPolicy:                 "required_on_terminal_success",
		IdempotencyPolicy:           "required",
		IdempotencyIdentitySchemaID: "cartulary.route_scoped_idempotency_identity.v1",
		TerminalResultSchemaID:      "cartulary.common_job_terminal_success.v1",
		CancellationPolicy:          "precommit_observable",
		MaxProofBytes:               4096,
	}
}

func assertJobTerminalState(t *testing.T, pool *pgxpool.Pool, jobID uuid.UUID, status string, code string) {
	t.Helper()
	var gotStatus string
	var resultCode, errorCode *string
	if err := pool.QueryRow(context.Background(), `
SELECT status, result_summary_json->>'code', error_summary_json->>'code'
  FROM jobs
 WHERE job_id = $1
`, jobID).Scan(&gotStatus, &resultCode, &errorCode); err != nil {
		t.Fatal(err)
	}
	if gotStatus != status {
		t.Fatalf("job %s status = %q; want %q", jobID, gotStatus, status)
	}
	gotCode := ""
	if resultCode != nil {
		gotCode = *resultCode
	}
	if errorCode != nil {
		gotCode = *errorCode
	}
	if gotCode != code {
		t.Fatalf("job %s terminal code = %q; want %q", jobID, gotCode, code)
	}
}

func assertJobStatus(t *testing.T, pool *pgxpool.Pool, jobID uuid.UUID, status string) {
	t.Helper()
	var got string
	if err := pool.QueryRow(context.Background(), `SELECT status FROM jobs WHERE job_id = $1`, jobID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != status {
		t.Fatalf("job %s status = %q; want %q", jobID, got, status)
	}
}
