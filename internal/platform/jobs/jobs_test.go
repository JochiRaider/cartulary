package jobs_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/intenttest"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

const testJobKind = collaborationsupport.TestJobKind

func TestManagerCreatesCancelsAndReplaysJobCancel(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID := newJobsHarness(t, "jobs-cancel-replay")

	resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           testJobKind,
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if resource.Status != jobs.StatusQueued || !resource.Cancelable || resource.Scope.IncidentID == nil || *resource.Scope.IncidentID != incidentID {
		t.Fatalf("unexpected created job resource: %#v", resource)
	}

	jobID := uuid.MustParse(resource.JobID)
	first, err := manager.Cancel(ctx, jobs.CancelParams{
		JobID:             jobID,
		ActorUserID:       actorID,
		ClientTxnID:       "txn-cancel",
		NormalizedRequest: []byte(`{"client_txn_id":"txn-cancel","reason":null}`),
	})
	if err != nil {
		t.Fatalf("cancel job: %v", err)
	}
	if first.Resource.Status != jobs.StatusCancelRequested || first.Resource.Cancelable {
		t.Fatalf("unexpected cancel resource: %#v", first.Resource)
	}

	replay, err := manager.Cancel(ctx, jobs.CancelParams{
		JobID:             jobID,
		ActorUserID:       actorID,
		ClientTxnID:       "txn-cancel",
		NormalizedRequest: []byte(`{"client_txn_id":"txn-cancel","reason":null}`),
	})
	if err != nil {
		t.Fatalf("replay cancel job: %v", err)
	}
	if !replay.Replayed || replay.Resource.Status != jobs.StatusCancelRequested {
		t.Fatalf("expected replayed cancel_requested resource, got %#v", replay)
	}

	_, err = manager.Cancel(ctx, jobs.CancelParams{
		JobID:             jobID,
		ActorUserID:       actorID,
		ClientTxnID:       "txn-cancel",
		NormalizedRequest: []byte(`{"client_txn_id":"txn-cancel","reason":"different"}`),
	})
	if !errors.Is(err, jobs.ErrClientTxnConflict) {
		t.Fatalf("divergent cancel replay error = %v, want ErrClientTxnConflict", err)
	}
}

func TestManagerCancellationUsesCanonicalTransitionLockOrder_Integration(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID, pool := newJobsHarnessWithPool(
		t,
		"jobs-cancel-lock-order",
		func() time.Time { return time.Now().UTC() },
	)
	compositionValue, present := testJobCompositions.Load(manager)
	if !present {
		t.Fatal("test Jobs composition is unavailable")
	}
	composition := compositionValue.(testJobComposition)

	enqueueCancelable := func(t *testing.T) (uuid.UUID, jobs.Execution) {
		t.Helper()
		resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
			JobKind:           testJobKind,
			Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
			SubmittedByUserID: actorID,
			Cancelable:        true,
			Progress:          jobs.Progress{Completed: 0},
		})
		if err != nil {
			t.Fatalf("create cancelable job: %v", err)
		}
		jobID := uuid.MustParse(resource.JobID)
		return jobID, claimTestExecution(t, manager, jobID)
	}
	cancel := func(jobID uuid.UUID, txnID string) <-chan error {
		result := make(chan error, 1)
		go func() {
			_, err := manager.Cancel(ctx, jobs.CancelParams{
				JobID:             jobID,
				ActorUserID:       actorID,
				ClientTxnID:       txnID,
				NormalizedRequest: []byte(`{"client_txn_id":"` + txnID + `","reason":null}`),
			})
			result <- err
		}()
		return result
	}

	t.Run("cancellation waits before locking the job row", func(t *testing.T) {
		jobID, _ := enqueueCancelable(t)
		lockConn, err := pool.Acquire(ctx)
		if err != nil {
			t.Fatalf("acquire transition blocker: %v", err)
		}
		defer lockConn.Release()
		var blockerPID int32
		if err := lockConn.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&blockerPID); err != nil {
			t.Fatalf("load blocker pid: %v", err)
		}
		const transitionLockSeed int64 = 49006006
		if _, err := lockConn.Exec(ctx, `SELECT pg_advisory_lock(hashtextextended($1::text, $2))`, jobID, transitionLockSeed); err != nil {
			t.Fatalf("hold transition lock: %v", err)
		}
		locked := true
		defer func() {
			if locked {
				_, _ = lockConn.Exec(context.Background(), `SELECT pg_advisory_unlock(hashtextextended($1::text, $2))`, jobID, transitionLockSeed)
			}
		}()

		cancelResult := cancel(jobID, "txn-cancel-lock-first")
		_ = waitForBlockedBackend(t, pool, blockerPID)

		probeTx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin row-lock probe: %v", err)
		}
		defer func() { _ = probeTx.Rollback(ctx) }()
		var present bool
		if err := probeTx.QueryRow(ctx, `SELECT true FROM jobs WHERE job_id = $1 FOR UPDATE NOWAIT`, jobID).Scan(&present); err != nil {
			t.Fatalf("cancellation locked the job row before the transition lock: %v", err)
		}
		if !present {
			t.Fatal("row-lock probe did not find job")
		}
		if err := probeTx.Rollback(ctx); err != nil {
			t.Fatalf("release row-lock probe: %v", err)
		}

		if _, err := lockConn.Exec(ctx, `SELECT pg_advisory_unlock(hashtextextended($1::text, $2))`, jobID, transitionLockSeed); err != nil {
			t.Fatalf("release transition lock: %v", err)
		}
		locked = false
		if err := <-cancelResult; err != nil {
			t.Fatalf("cancel after transition release: %v", err)
		}
	})

	t.Run("execution commit serializes before cancellation", func(t *testing.T) {
		jobID, execution := enqueueCancelable(t)
		ownerTx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin owner transaction: %v", err)
		}
		defer func() { _ = ownerTx.Rollback(ctx) }()
		if err := composition.transactions.ValidateExecutionTx(ctx, ownerTx, execution); err != nil {
			t.Fatalf("validate owner execution: %v", err)
		}
		var ownerPID int32
		if err := ownerTx.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&ownerPID); err != nil {
			t.Fatalf("load owner pid: %v", err)
		}
		cancelResult := cancel(jobID, "txn-owner-before-cancel")
		_ = waitForBlockedBackend(t, pool, ownerPID)
		if err := ownerTx.Commit(ctx); err != nil {
			t.Fatalf("commit owner transaction: %v", err)
		}
		if err := <-cancelResult; err != nil {
			t.Fatalf("cancel after owner commit: %v", err)
		}
	})

	t.Run("cancellation prevents later owner effects", func(t *testing.T) {
		jobID, execution := enqueueCancelable(t)
		if err := <-cancel(jobID, "txn-cancel-before-owner"); err != nil {
			t.Fatalf("cancel before owner validation: %v", err)
		}
		ownerTx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatalf("begin rejected owner transaction: %v", err)
		}
		defer func() { _ = ownerTx.Rollback(ctx) }()
		if err := composition.transactions.ValidateExecutionTx(ctx, ownerTx, execution); !errors.Is(err, jobs.ErrCancellationRequested) {
			t.Fatalf("owner validation error = %v, want ErrCancellationRequested", err)
		}
	})
}

func waitForBlockedBackend(t testing.TB, pool *pgxpool.Pool, blockerPID int32) int32 {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var waiterPID int32
		err := pool.QueryRow(context.Background(), `
SELECT pid
  FROM pg_stat_activity
 WHERE datname = current_database()
   AND pid <> pg_backend_pid()
   AND $1::integer = ANY(pg_blocking_pids(pid))
 ORDER BY pid
 LIMIT 1
`, blockerPID).Scan(&waiterPID)
		if err == nil {
			return waiterPID
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("inspect blocked backend: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for backend blocked by pid %d", blockerPID)
	return 0
}

func TestManagerTerminalSuccessRetainsJobResource(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
	manager, actorID, _ := newJobsHarnessWithClock(t, "jobs-terminal-success", func() time.Time { return now })

	resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           testJobKind,
		Scope:             jobs.Scope{Kind: jobs.ScopeKindDeployment},
		SubmittedByUserID: actorID,
		Cancelable:        false,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	jobID := uuid.MustParse(resource.JobID)
	execution := claimTestExecution(t, manager, jobID)
	total := 1
	if _, err := manager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 0, Total: &total}, nil); err != nil {
		t.Fatalf("start job: %v", err)
	}
	completed, err := manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{
		Progress: jobs.Progress{Completed: 1, Total: &total},
		ResultSummary: jobs.ResultSummary{
			Code:    "extension_profile_test_completed",
			Message: "Enterprise integration test completed.",
		},
	})
	if err != nil {
		t.Fatalf("complete job: %v", err)
	}
	if completed.Status != jobs.StatusSucceeded || completed.ResultSummary == nil || completed.ErrorSummary != nil {
		t.Fatalf("unexpected terminal resource: %#v", completed)
	}
	if completed.RetainedUntil == nil || completed.FinishedAt == nil || completed.RetainedUntil.Before(completed.FinishedAt.Add(7*24*time.Hour)) {
		t.Fatalf("terminal job retention too short: finished=%v retained=%v", completed.FinishedAt, completed.RetainedUntil)
	}
}

func TestManagerPersistsCanonicalIncidentProgressIntents_Integration(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID, pool := newJobsHarnessWithPool(
		t,
		"jobs-progress-intent",
		func() time.Time { return time.Date(2026, 7, 27, 21, 0, 0, 123000000, time.UTC) },
	)

	resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           testJobKind,
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create incident job: %v", err)
	}

	intentRecord := intenttest.LoadBySourceIdentity(t, pool, "job:"+resource.JobID)
	if intentRecord.EventFamily != "job_progress" || intentRecord.SourceIdentity != "job:"+resource.JobID {
		t.Fatalf("unexpected job progress identity: family=%q source=%q", intentRecord.EventFamily, intentRecord.SourceIdentity)
	}
	if wantPrefix := "job_progress:v2:" + resource.JobID + ":"; len(intentRecord.IntentKey) <= len(wantPrefix) || intentRecord.IntentKey[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("job progress intent key = %q, want prefix %q", intentRecord.IntentKey, wantPrefix)
	}
	var canonical map[string]any
	if err := json.Unmarshal(intentRecord.CanonicalPayload, &canonical); err != nil {
		t.Fatalf("decode job progress payload: %v", err)
	}
	scope, _ := canonical["scope"].(map[string]any)
	progress, _ := canonical["progress"].(map[string]any)
	if canonical["job_id"] != resource.JobID ||
		canonical["status"] != jobs.StatusQueued ||
		scope["kind"] != jobs.ScopeKindIncident ||
		scope["incident_id"] != incidentID.String() ||
		progress["completed"] != float64(0) {
		t.Fatalf("unexpected canonical job progress payload: %#v", canonical)
	}

	legacyKey := "job_progress:" + resource.JobID + ":legacy-v1"
	intenttest.InsertLegacyJobProgressV1(
		t, pool, legacyKey, incidentID, intentRecord.CanonicalPayload, "job:"+resource.JobID, resource.UpdatedAt,
	)
	execution := claimTestExecution(t, manager, uuid.MustParse(resource.JobID))
	if _, err := manager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 0}, nil); err != nil {
		t.Fatalf("publish v2 progress beside legacy v1: %v", err)
	}
	legacyCount, v2Count := intenttest.CountLegacyAndV2JobProgress(
		t, pool, legacyKey, "job_progress:v2:"+resource.JobID+":%", "job:"+resource.JobID,
	)
	if legacyCount != 1 || v2Count != 2 {
		t.Fatalf("job progress intent coexistence = v1 %d v2 %d want 1/2", legacyCount, v2Count)
	}
}

type rejectingProgressIntentAppender struct {
	err error
}

func (a rejectingProgressIntentAppender) AppendProgressIntentTx(context.Context, pgx.Tx, jobs.ProgressIntent) error {
	return a.err
}

func TestManagerProgressIntentFailureRollsBackJobMutation_Integration(t *testing.T) {
	_, actorID, incidentID, pool := newJobsHarnessWithPool(
		t,
		"jobs-progress-intent-rollback",
		func() time.Time { return time.Date(2026, 7, 28, 5, 15, 0, 0, time.UTC) },
	)
	intentFailure := errors.New("reject invalid job progress payload")
	ownerPorts := collaborationsupport.NewJobOwnerTransactionAdapters()
	definitions := collaborationsupport.TestJobDefinitions()
	catalog, err := jobs.NewCatalog([]jobs.Definition{definitions[0]})
	if err != nil {
		t.Fatal(err)
	}
	selection, err := jobs.FullRuntimeSelection(catalog, collaborationsupport.TestWorkerRuntimeContracts([]jobs.Definition{definitions[0]}))
	if err != nil {
		t.Fatal(err)
	}
	transactions, err := jobs.NewTransactionService(rejectingProgressIntentAppender{err: intentFailure}, jobs.OwnerTransactionPorts{
		RouteIdempotency: ownerPorts, ExtensionCancellation: ownerPorts,
	}, catalog, selection)
	if err != nil {
		t.Fatal(err)
	}
	manager, err := jobs.NewManager(jobs.ManagerOptions{
		Postgres: pool, Transactions: transactions, Catalog: catalog,
		Policy: jobs.ProductionRuntimePolicy(),
		Now:    func() time.Time { return time.Date(2026, 7, 28, 5, 15, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatal(err)
	}
	testJobCompositions.Store(manager, testJobComposition{catalog: catalog, transactions: transactions, pool: pool, now: func() time.Time { return time.Date(2026, 7, 28, 5, 15, 0, 0, time.UTC) }})
	_, err = enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           testJobKind,
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if !errors.Is(err, intentFailure) {
		t.Fatalf("create with rejected progress intent = %v want %v", err, intentFailure)
	}
	var jobCount int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM jobs`).Scan(&jobCount); err != nil {
		t.Fatalf("count rolled-back jobs: %v", err)
	}
	if jobCount != 0 {
		t.Fatalf("rejected progress intent left %d job rows", jobCount)
	}
}

func TestRunnerDispatchesDurableHandlerAndCompletesJob(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID := newJobsHarness(t, "jobs-durable-dispatch")

	resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           collaborationsupport.TestJobKindForHandler("test.complete"),
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
		HandlerPayload:    json.RawMessage(`{"mode":"dispatch"}`),
	})
	if err != nil {
		t.Fatalf("create durable job: %v", err)
	}
	jobID := uuid.MustParse(resource.JobID)

	gate := &dequeueGate{}
	gate.open.Store(true)
	runner := newTestRunner(t, manager, gate)
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runner.Close(closeCtx); err != nil {
			t.Fatalf("close runner: %v", err)
		}
	})

	handled := make(chan uuid.UUID, 1)
	registerAllTestHandlers(t, runner, map[string]jobs.HandlerFunc{"test.complete": func(ctx context.Context, execution jobs.Execution) error {
		got := execution.JobID()
		if got != jobID {
			return errors.New("handler received unexpected job id")
		}
		rawPayload, err := manager.HandlerPayload(ctx, execution)
		if err != nil {
			return err
		}
		var payload struct {
			Mode string `json:"mode"`
		}
		if err := json.Unmarshal(rawPayload, &payload); err != nil {
			return err
		}
		if payload.Mode != "dispatch" {
			return errors.New("handler received unexpected payload")
		}
		if _, err := manager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 0}, nil); err != nil {
			return err
		}
		total := 1
		if _, err := manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: jobs.ResultSummary{
				Code:    "durable_handler_completed",
				Message: "Durable handler completed.",
			},
		}); err != nil {
			return err
		}
		handled <- got
		return nil
	}})
	if err := runner.Activate(ctx); err != nil {
		t.Fatalf("activate durable runner: %v", err)
	}
	waitForDurableJobHandler(t, handled)
	completed, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatalf("load completed job: %v", err)
	}
	if completed.Status != jobs.StatusSucceeded || completed.ResultSummary == nil || completed.ResultSummary.Code != "durable_handler_completed" {
		t.Fatalf("unexpected completed durable job: %#v", completed)
	}
}

func TestRunnerRecoversQueuedDurableHandlerJob(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID := newJobsHarness(t, "jobs-durable-recover")

	resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           collaborationsupport.TestJobKindForHandler("test.recover"),
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create recoverable durable job: %v", err)
	}
	jobID := uuid.MustParse(resource.JobID)

	gate := &dequeueGate{}
	runner := newTestRunner(t, manager, gate)
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runner.Close(closeCtx); err != nil {
			t.Fatalf("close runner: %v", err)
		}
	})

	handled := make(chan uuid.UUID, 1)
	registerAllTestHandlers(t, runner, map[string]jobs.HandlerFunc{"test.recover": func(ctx context.Context, execution jobs.Execution) error {
		got := execution.JobID()
		total := 1
		if _, err := manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: jobs.ResultSummary{
				Code:    "durable_handler_recovered",
				Message: "Durable handler recovered.",
			},
		}); err != nil {
			return err
		}
		handled <- got
		return nil
	}})
	select {
	case <-handled:
		t.Fatal("recovered durable job before dequeue gate opened")
	case <-time.After(25 * time.Millisecond):
	}
	gate.open.Store(true)
	if err := runner.Activate(ctx); err != nil {
		t.Fatalf("activate durable handler recovery: %v", err)
	}
	waitForDurableJobHandler(t, handled)
	completed, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatalf("load recovered job: %v", err)
	}
	if completed.Status != jobs.StatusSucceeded || completed.ResultSummary == nil || completed.ResultSummary.Code != "durable_handler_recovered" {
		t.Fatalf("unexpected recovered durable job: %#v", completed)
	}
}

func TestRunnerNamedCompositionRejectsInvalidAndDuplicateRegistration(t *testing.T) {
	if _, err := jobs.NewRunner(jobs.RunnerOptions{}); !errors.Is(err, jobs.ErrNotConfigured) {
		t.Fatalf("unconfigured named composition error = %v; want ErrNotConfigured", err)
	}
	manager, _, _ := newJobsHarness(t, "jobs-runner-registration")
	gate := &dequeueGate{}
	gate.open.Store(true)
	runner := newTestRunner(t, manager, gate)
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runner.Close(closeCtx); err != nil {
			t.Fatalf("close runner: %v", err)
		}
	})
	handler := func(context.Context, jobs.Execution) error { return nil }
	if err := runner.RegisterHandler("test_platform.worker_v1", handler); err != nil {
		t.Fatalf("register named handler: %v", err)
	}
	if err := runner.RegisterHandler("test_platform.worker_v1", handler); !errors.Is(err, jobs.ErrHandlerAlreadyRegistered) {
		t.Fatalf("duplicate named handler error = %v; want ErrHandlerAlreadyRegistered", err)
	}
}

type dequeueGate struct {
	open atomic.Bool
}

func (gate *dequeueGate) AdmissionOpen() bool {
	return gate != nil && gate.open.Load()
}

func TestManagerFailsDurableHandlerClosedAfterMaxAttempts(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 8, 16, 0, 0, 0, time.UTC)
	manager, actorID, incidentID, pool := newJobsHarnessWithPool(
		t,
		"jobs-durable-exhausted",
		func() time.Time { return now },
	)

	resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           collaborationsupport.TestJobKindForHandler("test.exhausted"),
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create exhausted durable job: %v", err)
	}
	jobID := uuid.MustParse(resource.JobID)

	for attempt := range jobs.DefaultHandlerMaxAttempts {
		execution := claimTestExecution(t, manager, jobID)
		if err := manager.RecordExecutionFailure(ctx, execution, false); err != nil {
			t.Fatalf("record durable handler error: %v", err)
		}
		if attempt+1 < jobs.DefaultHandlerMaxAttempts {
			var nextAttemptAt time.Time
			if err := pool.QueryRow(ctx, `SELECT handler_next_attempt_at FROM jobs WHERE job_id = $1`, jobID).Scan(&nextAttemptAt); err != nil {
				t.Fatal(err)
			}
			delay := jobs.ProductionRuntimePolicy().RetryDelays[attempt]
			if !nextAttemptAt.Equal(now.Add(delay)) {
				t.Fatalf("attempt %d eligibility = %s, want %s", attempt+1, nextAttemptAt, now.Add(delay))
			}
			now = nextAttemptAt
		}
	}

	failed, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatalf("load exhausted job: %v", err)
	}
	if failed.Status != jobs.StatusFailed || failed.Cancelable || failed.ErrorSummary == nil || failed.ErrorSummary.Code != jobs.HandlerAttemptsExhausted {
		t.Fatalf("expected failed-closed durable job, got %#v", failed)
	}
	recoverable, err := manager.RecoverableJobs(ctx, 10)
	if err != nil {
		t.Fatalf("load recoverable durable jobs: %v", err)
	}
	if len(recoverable) != 0 {
		t.Fatalf("expected exhausted durable job to be excluded from recovery, got %v", recoverable)
	}
}

func waitForDurableJobHandler(t testing.TB, handled <-chan uuid.UUID) uuid.UUID {
	t.Helper()
	select {
	case jobID := <-handled:
		return jobID
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for durable job handler")
		return uuid.Nil
	}
}

func newJobsHarness(t testing.TB, prefix string) (*jobs.Manager, uuid.UUID, uuid.UUID) {
	t.Helper()
	return newJobsHarnessWithClock(t, prefix, func() time.Time { return time.Now().UTC() })
}

func newJobsHarnessWithClock(t testing.TB, prefix string, now func() time.Time) (*jobs.Manager, uuid.UUID, uuid.UUID) {
	manager, actorID, incidentID, _ := newJobsHarnessWithPool(t, prefix, now)
	return manager, actorID, incidentID
}

func newJobsHarnessWithPool(
	t testing.TB,
	prefix string,
	now func() time.Time,
) (*jobs.Manager, uuid.UUID, uuid.UUID, *pgxpool.Pool) {
	t.Helper()
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatalf("open pgx pool: %v", err)
	}
	t.Cleanup(pool.Close)

	actorID := uuid.New()
	incidentID := uuid.New()
	if _, err := pool.Exec(context.Background(), `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'Jobs Actor', 'hash', false, true, true)
`, actorID, actorID.String()+"@example.test"); err != nil {
		t.Fatalf("insert actor: %v", err)
	}
	if _, err := pool.Exec(context.Background(), `
INSERT INTO incidents (id, incident_key, incident_key_canonical, title, status, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $2, 'Jobs Incident', 'active', $3, $3)
`, incidentID, "jobs-"+incidentID.String(), actorID); err != nil {
		t.Fatalf("insert incident: %v", err)
	}
	if _, err := pool.Exec(context.Background(), `
INSERT INTO incident_memberships (incident_id, user_id, role, added_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'admin', $2, $2)
`, incidentID, actorID); err != nil {
		t.Fatalf("insert membership: %v", err)
	}

	catalog := collaborationsupport.NewJobCatalog()
	transactions := collaborationsupport.NewJobTransactionsForCatalog(catalog)
	manager, err := jobs.NewManager(jobs.ManagerOptions{
		Postgres: pool, Transactions: transactions, Catalog: catalog,
		Policy: jobs.ProductionRuntimePolicy(), Now: now,
		TelemetryServiceVersion: "0.0.0+unknown",
	})
	if err != nil {
		t.Fatalf("compose test Jobs manager: %v", err)
	}
	testJobCompositions.Store(manager, testJobComposition{
		catalog: catalog, transactions: transactions, pool: pool, now: now,
	})
	t.Cleanup(func() { testJobCompositions.Delete(manager) })
	return manager, actorID, incidentID, pool
}

type testJobComposition struct {
	catalog      *jobs.Catalog
	transactions *jobs.TransactionService
	pool         *pgxpool.Pool
	now          func() time.Time
}

var testJobCompositions sync.Map

func enqueueTestJob(t testing.TB, manager *jobs.Manager, params jobs.EnqueueParams) (jobs.Resource, error) {
	t.Helper()
	value, present := testJobCompositions.Load(manager)
	if !present {
		return jobs.Resource{}, errors.New("test Jobs transaction service is unavailable")
	}
	composition := value.(testJobComposition)
	tx, err := composition.pool.Begin(context.Background())
	if err != nil {
		return jobs.Resource{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	resource, err := composition.transactions.CreateQueuedTx(context.Background(), tx, params, composition.now().UTC())
	if err != nil {
		return jobs.Resource{}, err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return jobs.Resource{}, err
	}
	return resource, nil
}

func newTestRunner(t testing.TB, manager *jobs.Manager, gate jobs.DequeueGate) *jobs.Runner {
	t.Helper()
	value, present := testJobCompositions.Load(manager)
	if !present {
		t.Fatal("test Jobs composition is unavailable")
	}
	composition := value.(testJobComposition)
	runner, err := jobs.NewRunner(jobs.RunnerOptions{
		Manager: manager, Catalog: composition.catalog,
		Policy: jobs.ProductionRuntimePolicy(), DequeueGate: gate,
	})
	if err != nil {
		t.Fatalf("compose test Jobs runner: %v", err)
	}
	return runner
}

func registerAllTestHandlers(t testing.TB, runner *jobs.Runner, overrides map[string]jobs.HandlerFunc) {
	t.Helper()
	registered := map[string]struct{}{}
	for _, definition := range collaborationsupport.TestJobDefinitions() {
		if _, present := registered[definition.HandlerName]; present {
			continue
		}
		handler := overrides[definition.HandlerName]
		if handler == nil {
			handler = func(context.Context, jobs.Execution) error {
				return errors.New("unexpected test handler invocation")
			}
		}
		if err := runner.RegisterHandler(definition.HandlerName, handler); err != nil {
			t.Fatalf("register test handler %s: %v", definition.HandlerName, err)
		}
		registered[definition.HandlerName] = struct{}{}
	}
}

func claimTestExecution(t testing.TB, manager *jobs.Manager, jobID uuid.UUID) jobs.Execution {
	t.Helper()
	execution, claimed, err := manager.Claim(context.Background(), jobID)
	if err != nil || !claimed {
		t.Fatalf("claim test execution for %s: claimed=%t err=%v", jobID, claimed, err)
	}
	return execution
}
