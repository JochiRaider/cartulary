package jobs_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

const supervisedJobKind = "test_platform.supervised_v1"
const supervisedHandlerName = "test.supervised"

func TestSupervisedRunnerContract_Integration(t *testing.T) {
	t.Run("registration is exact and immutable after activation", func(t *testing.T) {
		manager, _, _, _, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-registration")
		gate := &dequeueGate{}
		gate.open.Store(true)

		missing, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := missing.Activate(context.Background()); !errors.Is(err, jobs.ErrHandlerNotRegistered) {
			t.Fatalf("activation with a missing handler error = %v", err)
		}
		closeRunner(t, missing)

		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := runner.RegisterHandler("test.extra", func(context.Context, jobs.Execution) error { return nil }); !errors.Is(err, jobs.ErrHandlerNotInCatalog) {
			t.Fatalf("extra handler registration error = %v", err)
		}
		if err := runner.RegisterHandler(supervisedHandlerName, func(context.Context, jobs.Execution) error { return nil }); err != nil {
			t.Fatal(err)
		}
		if err := runner.RegisterHandler(supervisedHandlerName, func(context.Context, jobs.Execution) error { return nil }); !errors.Is(err, jobs.ErrHandlerAlreadyRegistered) {
			t.Fatalf("duplicate handler registration error = %v", err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := runner.RegisterHandler(supervisedHandlerName, func(context.Context, jobs.Execution) error { return nil }); !errors.Is(err, jobs.ErrRunnerActivated) {
			t.Fatalf("late handler registration error = %v", err)
		}
		closeRunner(t, runner)
	})

	t.Run("initial recovery failure is synchronous", func(t *testing.T) {
		manager, _, _, pool, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-initial-failure")
		gate := &dequeueGate{}
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := runner.RegisterHandler(supervisedHandlerName, func(context.Context, jobs.Execution) error { return nil }); err != nil {
			t.Fatal(err)
		}
		pool.Close()
		if err := runner.Activate(context.Background()); err == nil {
			t.Fatal("initial recovery unexpectedly succeeded with closed storage")
		}
		closeRunner(t, runner)
	})

	t.Run("initial recovery renews a long attempt", func(t *testing.T) {
		manager, actorID, incidentID, _, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-renewal")
		jobID := createSupervisedJob(t, manager, actorID, incidentID)
		gate := &dequeueGate{}
		gate.open.Store(true)
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, execution jobs.Execution) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(3 * policy.HandlerLease):
			}
			total := 1
			_, err := manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{
				Progress:      jobs.Progress{Completed: 1, Total: &total},
				ResultSummary: jobs.ResultSummary{Code: "supervised", Message: "Supervised."},
			})
			return err
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		if got := waitSupervisedJobStatus(t, manager, jobID, jobs.StatusSucceeded); got.Status != jobs.StatusSucceeded {
			t.Fatalf("long supervised job = %#v", got)
		}
		closeRunner(t, runner)
	})

	t.Run("renewal error without shutdown consumes one failure", func(t *testing.T) {
		manager, actorID, incidentID, pool, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-renewal-failure")
		jobID := createSupervisedJob(t, manager, actorID, incidentID)
		gate := &dequeueGate{}
		gate.open.Store(true)
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		renewalTicks := make(chan time.Time, 1)
		renewalAttempted := make(chan struct{})
		jobs.ConfigureRunnerSynchronizationForTest(runner, renewalTicks, nil, func(context.Context, jobs.Execution) error {
			close(renewalAttempted)
			return errors.New("forced operational renewal failure")
		})
		handlerEntered := make(chan struct{})
		handlerExited := make(chan struct{})
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, _ jobs.Execution) error {
			close(handlerEntered)
			<-ctx.Done()
			close(handlerExited)
			return ctx.Err()
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		select {
		case <-handlerEntered:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for renewal-failure handler")
		}
		gate.open.Store(false)
		renewalTicks <- time.Now().UTC()
		select {
		case <-renewalAttempted:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for forced renewal failure")
		}
		select {
		case <-handlerExited:
		case <-time.After(time.Second):
			t.Fatal("renewal failure did not cancel handler")
		}
		waitSupervisedAttemptState(t, pool, jobID, 1, true)
		closeRunner(t, runner)
	})

	t.Run("lost renewal remains a conflict without another failure", func(t *testing.T) {
		manager, actorID, incidentID, pool, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-renewal-conflict")
		jobID := createSupervisedJob(t, manager, actorID, incidentID)
		gate := &dequeueGate{}
		gate.open.Store(true)
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		renewalTicks := make(chan time.Time, 1)
		renewalAttempted := make(chan error, 1)
		jobs.ConfigureRunnerSynchronizationForTest(runner, renewalTicks, nil, func(ctx context.Context, execution jobs.Execution) error {
			err := manager.RenewExecution(ctx, execution)
			renewalAttempted <- err
			return err
		})
		handlerEntered := make(chan struct{})
		handlerExited := make(chan struct{})
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, _ jobs.Execution) error {
			close(handlerEntered)
			<-ctx.Done()
			close(handlerExited)
			return ctx.Err()
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		select {
		case <-handlerEntered:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for renewal-conflict handler")
		}
		gate.open.Store(false)
		if _, err := pool.Exec(context.Background(), `
UPDATE jobs
   SET handler_attempt_id = NULL,
       handler_lease_expires_at = NULL
 WHERE job_id = $1
`, jobID); err != nil {
			t.Fatal(err)
		}
		renewalTicks <- time.Now().UTC()
		select {
		case err := <-renewalAttempted:
			if !errors.Is(err, jobs.ErrExecutionLost) {
				t.Fatalf("lost renewal error = %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for lost renewal")
		}
		select {
		case <-handlerExited:
		case <-time.After(time.Second):
			t.Fatal("lost renewal did not cancel handler")
		}
		closeRunner(t, runner)
		waitSupervisedAttemptState(t, pool, jobID, 0, false)
	})

	t.Run("continuous scan retries without restart and recovers a dropped hint", func(t *testing.T) {
		manager, actorID, incidentID, pool, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-retry")
		gate := &dequeueGate{}
		gate.open.Store(true)
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		var attempts atomic.Int32
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, execution jobs.Execution) error {
			if attempts.Add(1) == 1 {
				return errors.New("first attempt fails")
			}
			total := 1
			_, err := manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{
				Progress:      jobs.Progress{Completed: 1, Total: &total},
				ResultSummary: jobs.ResultSummary{Code: "retried", Message: "Retried."},
			})
			return err
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		jobID := createSupervisedJob(t, manager, actorID, incidentID)
		// No notification is sent. The continuous durable scan is authoritative.
		if got := waitSupervisedJobStatus(t, manager, jobID, jobs.StatusSucceeded); got.Status != jobs.StatusSucceeded {
			t.Fatalf("retried supervised job = %#v", got)
		}
		if attempts.Load() != 2 {
			t.Fatalf("handler attempts = %d, want 2", attempts.Load())
		}
		var failureCount int
		if err := pool.QueryRow(context.Background(), `SELECT handler_failure_count FROM jobs WHERE job_id = $1`, jobID).Scan(&failureCount); err != nil {
			t.Fatal(err)
		}
		if failureCount != 1 {
			t.Fatalf("persisted failure count = %d, want 1", failureCount)
		}
		closeRunner(t, runner)
	})

	t.Run("concurrency is bounded and shutdown releases attempts neutrally", func(t *testing.T) {
		manager, actorID, incidentID, pool, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-concurrency")
		policy.HandlerConcurrency = 8
		gate := &dequeueGate{}
		gate.open.Store(true)
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		var current atomic.Int32
		var maximum atomic.Int32
		entered := make(chan struct{}, 16)
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, _ jobs.Execution) error {
			active := current.Add(1)
			defer current.Add(-1)
			for {
				observed := maximum.Load()
				if active <= observed || maximum.CompareAndSwap(observed, active) {
					break
				}
			}
			entered <- struct{}{}
			<-ctx.Done()
			return ctx.Err()
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		jobIDs := make([]uuid.UUID, 12)
		for index := range jobIDs {
			jobIDs[index] = createSupervisedJob(t, manager, actorID, incidentID)
			runner.Notify(jobIDs[index])
		}
		for range 8 {
			select {
			case <-entered:
			case <-time.After(2 * time.Second):
				t.Fatal("timed out waiting for eight concurrent attempts")
			}
		}
		select {
		case <-entered:
			t.Fatal("runner exceeded eight concurrent attempts")
		case <-time.After(3 * policy.RecoveryScan):
		}
		if maximum.Load() != 8 {
			t.Fatalf("maximum concurrency = %d, want 8", maximum.Load())
		}
		closeRunner(t, runner)
		for _, jobID := range jobIDs[:8] {
			var attemptID *uuid.UUID
			var failureCount int
			if err := pool.QueryRow(context.Background(), `
SELECT handler_attempt_id, handler_failure_count
  FROM jobs
 WHERE job_id = $1
`, jobID).Scan(&attemptID, &failureCount); err != nil {
				t.Fatal(err)
			}
			if attemptID != nil || failureCount != 0 {
				t.Fatalf("shutdown release for %s = attempt %v failures %d", jobID, attemptID, failureCount)
			}
		}
	})

	t.Run("shutdown wins simultaneous renewal and handler readiness", func(t *testing.T) {
		manager, actorID, incidentID, pool, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-shutdown-precedence")
		jobID := createSupervisedJob(t, manager, actorID, incidentID)
		gate := &dequeueGate{}
		gate.open.Store(true)
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}

		renewalTicks := make(chan time.Time, 1)
		atWait := make(chan struct{})
		releaseWait := make(chan struct{})
		var waitCalls atomic.Int32
		jobs.ConfigureRunnerSynchronizationForTest(runner, renewalTicks, func() {
			if waitCalls.Add(1) != 1 {
				return
			}
			close(atWait)
			<-releaseWait
		}, nil)

		handlerEntered := make(chan struct{})
		handlerExited := make(chan struct{})
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, _ jobs.Execution) error {
			close(handlerEntered)
			<-ctx.Done()
			close(handlerExited)
			return ctx.Err()
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		select {
		case <-handlerEntered:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for synchronized handler")
		}
		select {
		case <-atWait:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting at runner select boundary")
		}

		renewalTicks <- time.Now().UTC()
		closeResult := make(chan error, 1)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			closeResult <- runner.Close(ctx)
		}()
		select {
		case <-handlerExited:
		case <-time.After(time.Second):
			t.Fatal("shutdown did not cancel the synchronized handler")
		}
		close(releaseWait)
		if err := <-closeResult; err != nil {
			t.Fatalf("close synchronized runner: %v", err)
		}

		var attemptID *uuid.UUID
		var failureCount int
		var nextAttemptAt *time.Time
		if err := pool.QueryRow(context.Background(), `
SELECT handler_attempt_id, handler_failure_count, handler_next_attempt_at
  FROM jobs
 WHERE job_id = $1
`, jobID).Scan(&attemptID, &failureCount, &nextAttemptAt); err != nil {
			t.Fatal(err)
		}
		if attemptID != nil || failureCount != 0 || nextAttemptAt != nil {
			t.Fatalf("synchronized shutdown release = attempt %v failures %d next %v", attemptID, failureCount, nextAttemptAt)
		}
	})

	t.Run("successful close waits for handler drain", func(t *testing.T) {
		manager, actorID, incidentID, _, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-handler-drain")
		jobID := createSupervisedJob(t, manager, actorID, incidentID)
		gate := &dequeueGate{}
		gate.open.Store(true)
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		handlerEntered := make(chan struct{})
		cancellationObserved := make(chan struct{})
		allowHandlerExit := make(chan struct{})
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, _ jobs.Execution) error {
			close(handlerEntered)
			<-ctx.Done()
			close(cancellationObserved)
			<-allowHandlerExit
			return ctx.Err()
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		select {
		case <-handlerEntered:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for drain handler")
		}
		closeResult := make(chan error, 1)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			closeResult <- runner.Close(ctx)
		}()
		select {
		case <-cancellationObserved:
		case <-time.After(time.Second):
			t.Fatal("drain handler did not observe cancellation")
		}
		select {
		case err := <-closeResult:
			t.Fatalf("runner closed before handler drain: %v", err)
		default:
		}
		close(allowHandlerExit)
		if err := <-closeResult; err != nil {
			t.Fatalf("close drained runner: %v", err)
		}
		if resource, err := manager.Get(context.Background(), jobID); err != nil || resource.Status != jobs.StatusRunning {
			t.Fatalf("drained job resource = %#v/%v", resource, err)
		}
	})
}

func newSupervisedJobsHarness(
	t testing.TB,
	name string,
) (*jobs.Manager, uuid.UUID, uuid.UUID, *pgxpool.Pool, *jobs.Catalog, jobs.RuntimePolicy) {
	t.Helper()
	_, actorID, incidentID, pool := newJobsHarnessWithPool(t, name, func() time.Time { return time.Now().UTC() })
	definition := jobs.Definition{
		JobKind:        supervisedJobKind,
		ProgressUnitID: "test_platform.supervised.operation.v1",
		HandlerName:    supervisedHandlerName,
	}
	catalog, err := jobs.NewCatalog([]jobs.Definition{definition})
	if err != nil {
		t.Fatal(err)
	}
	transactions := collaborationsupport.NewJobTransactionsForCatalog(catalog)
	policy := jobs.ProductionRuntimePolicy()
	policy.HandlerLease = 120 * time.Millisecond
	policy.LeaseRenewal = 30 * time.Millisecond
	policy.RecoveryScan = 20 * time.Millisecond
	policy.RetryDelays = []time.Duration{20 * time.Millisecond, 40 * time.Millisecond}
	manager, err := jobs.NewManager(jobs.ManagerOptions{
		Postgres: pool, Transactions: transactions, Catalog: catalog,
		Policy: policy, Now: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		t.Fatal(err)
	}
	testJobCompositions.Store(manager, testJobComposition{
		catalog: catalog, transactions: transactions, pool: pool,
		now: func() time.Time { return time.Now().UTC() },
	})
	t.Cleanup(func() { testJobCompositions.Delete(manager) })
	return manager, actorID, incidentID, pool, catalog, policy
}

func createSupervisedJob(t testing.TB, manager *jobs.Manager, actorID uuid.UUID, incidentID uuid.UUID) uuid.UUID {
	t.Helper()
	resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind: supervisedJobKind, Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	return uuid.MustParse(resource.JobID)
}

func waitSupervisedJobStatus(t testing.TB, manager *jobs.Manager, jobID uuid.UUID, status string) jobs.Resource {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		resource, err := manager.Get(context.Background(), jobID)
		if err != nil {
			t.Fatal(err)
		}
		if resource.Status == status {
			return resource
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for job %s status %s: %#v", jobID, status, resource)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func waitSupervisedAttemptState(
	t testing.TB,
	pool *pgxpool.Pool,
	jobID uuid.UUID,
	wantFailureCount int,
	wantRetryDelay bool,
) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		var attemptID *uuid.UUID
		var failureCount int
		var nextAttemptAt *time.Time
		err := pool.QueryRow(context.Background(), `
SELECT handler_attempt_id, handler_failure_count, handler_next_attempt_at
  FROM jobs
 WHERE job_id = $1
`, jobID).Scan(&attemptID, &failureCount, &nextAttemptAt)
		if err != nil {
			t.Fatal(err)
		}
		if attemptID == nil && failureCount == wantFailureCount && (nextAttemptAt != nil) == wantRetryDelay {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf(
				"job attempt state = attempt %v failures %d next %v, want attempt nil failures %d retry_delay %t",
				attemptID,
				failureCount,
				nextAttemptAt,
				wantFailureCount,
				wantRetryDelay,
			)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func closeRunner(t testing.TB, runner *jobs.Runner) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Close(ctx); err != nil {
		t.Fatalf("close Jobs runner: %v", err)
	}
}
