package jobs_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
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
		var releasing atomic.Int32
		var maximumReleasing atomic.Int32
		releaseWaveReady := make(chan struct{})
		var releaseWaveOnce sync.Once
		entered := make(chan uuid.UUID, 16)
		jobs.ConfigureRunnerReleaseForTest(runner, func(ctx context.Context, execution jobs.Execution) error {
			active := releasing.Add(1)
			defer releasing.Add(-1)
			for {
				observed := maximumReleasing.Load()
				if active <= observed || maximumReleasing.CompareAndSwap(observed, active) {
					break
				}
			}
			if active == int32(policy.HandlerConcurrency) {
				releaseWaveOnce.Do(func() { close(releaseWaveReady) })
			}
			select {
			case <-releaseWaveReady:
			case <-ctx.Done():
				return ctx.Err()
			}
			return manager.ReleaseExecution(ctx, execution)
		})
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, execution jobs.Execution) error {
			active := current.Add(1)
			defer current.Add(-1)
			for {
				observed := maximum.Load()
				if active <= observed || maximum.CompareAndSwap(observed, active) {
					break
				}
			}
			entered <- execution.JobID()
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
		enteredJobIDs := make([]uuid.UUID, 0, policy.HandlerConcurrency)
		enteredSet := make(map[uuid.UUID]struct{}, policy.HandlerConcurrency)
		for range policy.HandlerConcurrency {
			select {
			case jobID := <-entered:
				if _, duplicate := enteredSet[jobID]; duplicate {
					t.Fatalf("handler entered twice for %s", jobID)
				}
				enteredSet[jobID] = struct{}{}
				enteredJobIDs = append(enteredJobIDs, jobID)
			case <-time.After(2 * time.Second):
				t.Fatal("timed out waiting for eight concurrent attempts")
			}
		}
		if maximum.Load() != 8 {
			t.Fatalf("maximum concurrency = %d, want 8", maximum.Load())
		}
		closeRunnerWithBudget(t, runner, 2*policy.AttemptOperationTimeout+time.Second)
		if maximumReleasing.Load() > int32(policy.HandlerConcurrency) {
			t.Fatalf("maximum release concurrency = %d, want <= %d", maximumReleasing.Load(), policy.HandlerConcurrency)
		}
		if maximumReleasing.Load() < 2 {
			t.Fatalf("shutdown releases were unexpectedly serialized: maximum concurrency %d", maximumReleasing.Load())
		}
		for _, jobID := range enteredJobIDs {
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
				t.Fatalf("shutdown release for %s = attempt %v failures %d next %v", jobID, attemptID, failureCount, nextAttemptAt)
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

	t.Run("shutdown release timeout is neutral and safely diagnosed", func(t *testing.T) {
		manager, actorID, incidentID, pool, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-release-timeout")
		policy.LeaseRenewal = 2 * time.Millisecond
		policy.AttemptOperationTimeout = 50 * time.Millisecond
		gate := &dequeueGate{}
		gate.open.Store(true)
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		releaseBudget := make(chan time.Duration, 1)
		jobs.ConfigureRunnerReleaseForTest(runner, func(ctx context.Context, _ jobs.Execution) error {
			deadline, ok := ctx.Deadline()
			if !ok {
				return errors.New("release context has no deadline")
			}
			releaseBudget <- time.Until(deadline)
			<-ctx.Done()
			return ctx.Err()
		})
		entered := make(chan struct{})
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, _ jobs.Execution) error {
			close(entered)
			<-ctx.Done()
			return ctx.Err()
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		jobID := createSupervisedJob(t, manager, actorID, incidentID)
		runner.Notify(jobID)
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for release-timeout handler")
		}
		closeCtx, cancelClose := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancelClose()
		closeErr := runner.Close(closeCtx)
		if closeErr == nil {
			t.Fatal("expected shutdown release timeout")
		}
		budget := <-releaseBudget
		if budget < 35*time.Millisecond {
			t.Fatalf("attempt operation budget %v was coupled to %v renewal cadence", budget, policy.LeaseRenewal)
		}
		diagnostic := closeErr.Error()
		for _, want := range []string{"stage=release", "job_kind=" + supervisedJobKind, "attempt_slot=1", "reason=deadline_exceeded"} {
			if !strings.Contains(diagnostic, want) {
				t.Fatalf("shutdown diagnostic %q missing %q", diagnostic, want)
			}
		}
		if strings.Contains(diagnostic, jobID.String()) {
			t.Fatalf("shutdown diagnostic exposed job identity: %q", diagnostic)
		}
		assertNeutralAttemptState(t, pool, jobID, true)
	})

	t.Run("multiple shutdown release failures are slot ordered and redacted", func(t *testing.T) {
		manager, actorID, incidentID, pool, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-release-aggregate")
		policy.HandlerConcurrency = 3
		gate := &dequeueGate{}
		gate.open.Store(true)
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		jobs.ConfigureRunnerReleaseForTest(runner, func(_ context.Context, execution jobs.Execution) error {
			return fmt.Errorf("raw database release failure for %s", execution.JobID())
		})
		entered := make(chan uuid.UUID, policy.HandlerConcurrency)
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, execution jobs.Execution) error {
			entered <- execution.JobID()
			<-ctx.Done()
			return ctx.Err()
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		jobIDs := make([]uuid.UUID, policy.HandlerConcurrency)
		for index := range jobIDs {
			jobIDs[index] = createSupervisedJob(t, manager, actorID, incidentID)
			runner.Notify(jobIDs[index])
		}
		for range policy.HandlerConcurrency {
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for aggregate release handlers")
			}
		}
		closeCtx, cancelClose := context.WithTimeout(context.Background(), time.Second)
		defer cancelClose()
		closeErr := runner.Close(closeCtx)
		if closeErr == nil {
			t.Fatal("expected aggregate shutdown release failure")
		}
		diagnostic := closeErr.Error()
		first := strings.Index(diagnostic, "attempt_slot=1")
		second := strings.Index(diagnostic, "attempt_slot=2")
		third := strings.Index(diagnostic, "attempt_slot=3")
		if first < 0 || second <= first || third <= second {
			t.Fatalf("shutdown failures are not slot ordered: %q", diagnostic)
		}
		if strings.Contains(diagnostic, "raw database release failure") {
			t.Fatalf("shutdown diagnostic exposed raw database error: %q", diagnostic)
		}
		for _, jobID := range jobIDs {
			if strings.Contains(diagnostic, jobID.String()) {
				t.Fatalf("shutdown diagnostic exposed job identity: %q", diagnostic)
			}
			assertNeutralAttemptState(t, pool, jobID, true)
		}
	})

	t.Run("execution loss during shutdown remains neutral", func(t *testing.T) {
		manager, actorID, incidentID, pool, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-release-loss")
		gate := &dequeueGate{}
		gate.open.Store(true)
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		jobs.ConfigureRunnerReleaseForTest(runner, func(context.Context, jobs.Execution) error {
			return jobs.ErrExecutionLost
		})
		entered := make(chan struct{})
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, _ jobs.Execution) error {
			close(entered)
			<-ctx.Done()
			return ctx.Err()
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		jobID := createSupervisedJob(t, manager, actorID, incidentID)
		runner.Notify(jobID)
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for execution-loss handler")
		}
		closeRunnerWithBudget(t, runner, 2*policy.AttemptOperationTimeout+time.Second)
		assertNeutralAttemptState(t, pool, jobID, true)
	})

	t.Run("caller deadline bounds handler drain", func(t *testing.T) {
		manager, actorID, incidentID, _, catalog, policy := newSupervisedJobsHarness(t, "jobs-supervisor-close-deadline")
		gate := &dequeueGate{}
		gate.open.Store(true)
		runner, err := jobs.NewRunner(jobs.RunnerOptions{
			Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
		})
		if err != nil {
			t.Fatal(err)
		}
		jobs.ConfigureRunnerReleaseForTest(runner, func(context.Context, jobs.Execution) error {
			return jobs.ErrExecutionLost
		})
		entered := make(chan struct{})
		allowDrain := make(chan struct{})
		if err := runner.RegisterHandler(supervisedHandlerName, func(ctx context.Context, _ jobs.Execution) error {
			close(entered)
			<-ctx.Done()
			<-allowDrain
			return ctx.Err()
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.Activate(context.Background()); err != nil {
			t.Fatal(err)
		}
		jobID := createSupervisedJob(t, manager, actorID, incidentID)
		runner.Notify(jobID)
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for deadline handler")
		}
		closeCtx, cancelClose := context.WithTimeout(context.Background(), 20*time.Millisecond)
		err = runner.Close(closeCtx)
		cancelClose()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("close deadline error = %v", err)
		}
		close(allowDrain)
		closeRunnerWithBudget(t, runner, time.Second)
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

func TestRunnerWorkerCapacityAndMixedKindFairness_Integration(t *testing.T) {
	base := time.Date(2026, 8, 16, 15, 0, 0, 0, time.UTC)
	_, actorID, incidentID, pool := newJobsHarnessWithPool(t, "jobs-worker-capacity", func() time.Time { return base })
	graphDefinition := jobs.Definition{
		JobKind: "test_platform.graph_materialize_v1", ProgressUnitID: "test_platform.graph.attempt.v1",
		HandlerName: "test_platform.graph_worker_v1",
	}
	regularDefinition := jobs.Definition{
		JobKind: "test_platform.regular_v1", ProgressUnitID: "test_platform.regular.attempt.v1",
		HandlerName: "test_platform.regular_worker_v1",
	}
	definitions := []jobs.Definition{graphDefinition, regularDefinition}
	workerContracts := []jobs.WorkerRuntimeContract{
		{ProfileID: "base", WorkerKind: graphDefinition.HandlerName, JobKinds: []string{graphDefinition.JobKind}, MaxActiveAttemptsPerProcess: 1},
		{ProfileID: "base", WorkerKind: regularDefinition.HandlerName, JobKinds: []string{regularDefinition.JobKind}, MaxActiveAttemptsPerProcess: 8},
	}
	catalog, err := jobs.NewCatalog(definitions)
	if err != nil {
		t.Fatal(err)
	}
	transactions := collaborationsupport.NewJobTransactionsForCatalog(catalog, workerContracts)
	clock := base
	policy := jobs.ProductionRuntimePolicy()
	policy.HandlerConcurrency = 4
	policy.RecoveryScan = time.Hour
	manager, err := jobs.NewManager(jobs.ManagerOptions{
		Postgres: pool, Transactions: transactions, Catalog: catalog, Policy: policy,
		Now: func() time.Time { return clock },
	})
	if err != nil {
		t.Fatal(err)
	}
	testJobCompositions.Store(manager, testJobComposition{
		catalog: catalog, transactions: transactions, pool: pool, now: func() time.Time { return clock },
	})
	t.Cleanup(func() { testJobCompositions.Delete(manager) })
	graphJobIDs := make([]uuid.UUID, 0, policy.RecoveryBatch+1)
	for range policy.RecoveryBatch + 1 {
		resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
			JobKind: graphDefinition.JobKind, Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
			SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Completed: 0},
		})
		if err != nil {
			t.Fatal(err)
		}
		graphJobIDs = append(graphJobIDs, uuid.MustParse(resource.JobID))
	}
	clock = base.Add(time.Minute)
	regularResource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind: regularDefinition.JobKind, Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatal(err)
	}

	gate := &dequeueGate{}
	gate.open.Store(true)
	runner, err := jobs.NewRunner(jobs.RunnerOptions{Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeRunner(t, runner) })
	graphEntered := make(chan uuid.UUID, 3)
	allowGraphCompletion := make(chan struct{})
	var graphActive atomic.Int32
	var graphMaximum atomic.Int32
	if err := runner.RegisterHandler(graphDefinition.HandlerName, func(ctx context.Context, execution jobs.Execution) error {
		active := graphActive.Add(1)
		defer graphActive.Add(-1)
		for {
			observed := graphMaximum.Load()
			if active <= observed || graphMaximum.CompareAndSwap(observed, active) {
				break
			}
		}
		graphEntered <- execution.JobID()
		select {
		case <-allowGraphCompletion:
			total := 1
			_, err := manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{
				Progress:      jobs.Progress{Completed: 1, Total: &total},
				ResultSummary: jobs.ResultSummary{Code: "worker_capacity_complete", Message: "Worker capacity complete."},
			})
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}); err != nil {
		t.Fatal(err)
	}
	regularCompleted := make(chan struct{}, 1)
	if err := runner.RegisterHandler(regularDefinition.HandlerName, func(ctx context.Context, execution jobs.Execution) error {
		total := 1
		if _, err := manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{
			Progress:      jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: jobs.ResultSummary{Code: "worker_fairness_complete", Message: "Worker fairness complete."},
		}); err != nil {
			return err
		}
		regularCompleted <- struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := runner.Activate(context.Background()); err != nil {
		t.Fatal(err)
	}
	var firstGraphJobID uuid.UUID
	select {
	case firstGraphJobID = <-graphEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("graph worker did not claim its one process slot")
	}
	select {
	case <-regularCompleted:
	case <-time.After(2 * time.Second):
		t.Fatal("regular worker was starved behind a full graph recovery batch")
	}
	select {
	case <-graphEntered:
		t.Fatal("graph worker exceeded its generated process capacity")
	case <-time.After(100 * time.Millisecond):
	}
	if graphMaximum.Load() != 1 {
		t.Fatalf("graph maximum active attempts = %d; want 1", graphMaximum.Load())
	}
	close(allowGraphCompletion)
	deadline := time.Now().Add(2 * time.Second)
	for graphActive.Load() != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if graphActive.Load() != 0 {
		t.Fatal("completed graph attempt did not release its worker reservation")
	}
	successorJobID := graphJobIDs[0]
	if successorJobID == firstGraphJobID {
		successorJobID = graphJobIDs[1]
	}
	successorDeadline := time.Now().Add(2 * time.Second)
	successorAdmitted := false
	for !successorAdmitted {
		runner.Notify(successorJobID)
		select {
		case <-graphEntered:
			successorAdmitted = true
		case <-time.After(10 * time.Millisecond):
		}
		if !successorAdmitted && time.Now().After(successorDeadline) {
			t.Fatal("released graph worker reservation did not admit a notified successor")
		}
	}
	regular, err := manager.Get(context.Background(), uuid.MustParse(regularResource.JobID))
	if err != nil || regular.Status != jobs.StatusSucceeded {
		t.Fatalf("regular worker result = %#v/%v", regular, err)
	}
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
	transactions := collaborationsupport.NewJobTransactionsForCatalog(catalog, collaborationsupport.TestWorkerRuntimeContracts([]jobs.Definition{definition}))
	policy := jobs.ProductionRuntimePolicy()
	policy.HandlerLease = 120 * time.Millisecond
	policy.LeaseRenewal = 30 * time.Millisecond
	policy.AttemptOperationTimeout = 250 * time.Millisecond
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
	closeRunnerWithBudget(t, runner, time.Second)
}

func closeRunnerWithBudget(t testing.TB, runner *jobs.Runner, budget time.Duration) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()
	if err := runner.Close(ctx); err != nil {
		t.Fatalf("close Jobs runner: %v", err)
	}
}

func assertNeutralAttemptState(t testing.TB, pool *pgxpool.Pool, jobID uuid.UUID, expectAttempt bool) {
	t.Helper()
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
	if (attemptID != nil) != expectAttempt || failureCount != 0 || nextAttemptAt != nil {
		t.Fatalf(
			"neutral shutdown state for %s = attempt %v failures %d next %v, want attempt_present %t failures 0 next nil",
			jobID,
			attemptID,
			failureCount,
			nextAttemptAt,
			expectAttempt,
		)
	}
}
