package jobs_test

import (
	"context"
	"encoding/json"
	"errors"
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

func TestTransitionAndProgressCorrection_Integration(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID, pool := newJobsHarnessWithPool(
		t,
		"jobs-transition-progress-correction",
		func() time.Time { return time.Now().UTC() },
	)

	t.Run("running progress monotonicity and no-op", func(t *testing.T) {
		jobID := createCorrectionJob(t, manager, actorID, incidentID, "")
		execution := claimTestExecution(t, manager, jobID)
		if _, err := manager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 2}, nil); err != nil {
			t.Fatal(err)
		}
		before, err := manager.Get(ctx, jobID)
		if err != nil {
			t.Fatal(err)
		}
		beforeIntents := correctionIntentCount(t, pool, jobID)
		repeated, err := manager.UpdateProgress(ctx, execution, before.Progress, before.Message)
		if err != nil {
			t.Errorf("exact repeated progress rejected: %v", err)
		} else if !samePublicJobState(before, repeated) {
			t.Errorf("exact repeated progress was not a true no-op: before=%#v after=%#v", before, repeated)
		}
		if got := correctionIntentCount(t, pool, jobID); got != beforeIntents {
			t.Errorf("exact repeated progress emitted intent: got %d want %d", got, beforeIntents)
		}

		total := 5
		discovered, err := manager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 2, Total: &total}, nil)
		if err != nil {
			t.Errorf("total discovery rejected: %v", err)
		} else if discovered.Progress.Total == nil || *discovered.Progress.Total != total {
			t.Errorf("total discovery result = %#v", discovered.Progress)
		}
		if _, err := manager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 3, Total: &total}, nil); err != nil {
			t.Errorf("monotonic progress increase rejected: %v", err)
		}

		assertRejectedProgressUnchanged(t, manager, pool, execution, jobs.Progress{Completed: 1, Total: &total})
		lowerTotal := 4
		assertRejectedProgressUnchanged(t, manager, pool, execution, jobs.Progress{Completed: 3, Total: &lowerTotal})
		assertRejectedProgressUnchanged(t, manager, pool, execution, jobs.Progress{Completed: 3})
		overTotal := 2
		assertRejectedProgressUnchanged(t, manager, pool, execution, jobs.Progress{Completed: 3, Total: &overTotal})
	})

	t.Run("known total success requires equality", func(t *testing.T) {
		jobID := createCorrectionJob(t, manager, actorID, incidentID, "")
		execution := claimTestExecution(t, manager, jobID)
		total := 5
		if _, err := manager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 4, Total: &total}, nil); err != nil {
			t.Fatal(err)
		}
		before, err := manager.Get(ctx, jobID)
		if err != nil {
			t.Fatal(err)
		}
		_, err = manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{
			Progress: jobs.Progress{
				Completed: 4,
				Total:     &total,
			},
			ResultSummary: jobs.ResultSummary{Code: "correction_complete", Message: "Complete."},
		})
		if !errors.Is(err, jobs.ErrInvalidTransition) {
			t.Errorf("incomplete known-total success error = %v; want ErrInvalidTransition", err)
		}
		after, loadErr := manager.Get(ctx, jobID)
		if loadErr != nil {
			t.Fatal(loadErr)
		}
		if err != nil && !samePublicJobState(before, after) {
			t.Errorf("rejected incomplete success mutated job: before=%#v after=%#v", before, after)
		}
	})

	t.Run("concurrent progress preserves the maximum", func(t *testing.T) {
		jobID := createCorrectionJob(t, manager, actorID, incidentID, "")
		execution := claimTestExecution(t, manager, jobID)
		total := 10
		if _, err := manager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 0, Total: &total}, nil); err != nil {
			t.Fatal(err)
		}
		start := make(chan struct{})
		var wg sync.WaitGroup
		errorsByCompleted := make(chan struct {
			completed int
			err       error
		}, 2)
		for _, completed := range []int{3, 4} {
			wg.Add(1)
			go func(value int) {
				defer wg.Done()
				<-start
				_, updateErr := manager.UpdateProgress(ctx, execution, jobs.Progress{Completed: value, Total: &total}, nil)
				errorsByCompleted <- struct {
					completed int
					err       error
				}{completed: value, err: updateErr}
			}(completed)
		}
		close(start)
		wg.Wait()
		close(errorsByCompleted)
		for result := range errorsByCompleted {
			if result.completed == 4 && result.err != nil {
				t.Errorf("highest concurrent progress rejected: %v", result.err)
			}
			if result.err != nil && !errors.Is(result.err, jobs.ErrInvalidTransition) {
				t.Errorf("concurrent progress %d returned unsafe error: %v", result.completed, result.err)
			}
		}
		resource, err := manager.Get(ctx, jobID)
		if err != nil {
			t.Fatal(err)
		}
		if resource.Progress.Completed != 4 || resource.Progress.Total == nil || *resource.Progress.Total != total {
			t.Errorf("concurrent progress final value = %#v; want 4/10", resource.Progress)
		}
		assertLatestIntentMatchesResource(t, pool, resource)
	})

	t.Run("terminal completion fences stale execution", func(t *testing.T) {
		jobID := createCorrectionJob(t, manager, actorID, incidentID, "")
		execution := claimTestExecution(t, manager, jobID)
		completed, err := manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{
			Progress:      jobs.Progress{Completed: 1},
			ResultSummary: jobs.ResultSummary{Code: "correction_succeeded", Message: "Succeeded."},
		})
		if err != nil || completed.Status != jobs.StatusSucceeded {
			t.Fatalf("typed completion = %#v, %v", completed, err)
		}
		if _, err := manager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 2}, nil); !errors.Is(err, jobs.ErrExecutionLost) {
			t.Fatalf("stale execution update error = %v; want ErrExecutionLost", err)
		}
	})
}

func TestStorageCatalogStartupValidation_Integration(t *testing.T) {
	manager, actorID, incidentID, pool := newJobsHarnessWithPool(
		t,
		"jobs-storage-catalog-validation",
		func() time.Time { return time.Now().UTC() },
	)
	jobID := createCorrectionJob(t, manager, actorID, incidentID, "")
	if err := manager.ValidateStorageCatalog(context.Background()); err != nil {
		t.Fatalf("valid mutable catalog rejected: %v", err)
	}
	privateUnit := "private.secret.unit.v1"
	if _, err := pool.Exec(context.Background(), `UPDATE jobs SET progress_unit_id = $2 WHERE job_id = $1`, jobID, privateUnit); err != nil {
		t.Fatal(err)
	}
	err := manager.ValidateStorageCatalog(context.Background())
	if !errors.Is(err, jobs.ErrStorageIncompatible) {
		t.Fatalf("invalid mutable catalog error = %v; want ErrStorageIncompatible", err)
	}
	diagnostic := err.Error()
	if !strings.Contains(diagnostic, "invalid_mutable_count=1") ||
		!strings.Contains(diagnostic, testJobKind+":"+jobs.StatusQueued) {
		t.Fatalf("bounded compatibility diagnostic = %q", diagnostic)
	}
	if strings.Contains(diagnostic, jobID.String()) || strings.Contains(diagnostic, privateUnit) {
		t.Fatalf("compatibility diagnostic leaked private identity: %q", diagnostic)
	}
}

func TestDurableClaimsRecoveryAndDuplicateDispatch_Integration(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID, pool := newJobsHarnessWithPool(
		t,
		"jobs-claim-recovery-correction",
		func() time.Time { return time.Now().UTC() },
	)
	jobID := createCorrectionJob(t, manager, actorID, incidentID, "test.claim")
	execution := claimTestExecution(t, manager, jobID)
	claimedResource, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	if claimedResource.Status != jobs.StatusRunning || claimedResource.StartedAt == nil {
		t.Errorf("handler claim did not publish running before execution: %#v", claimedResource)
	}
	assertLatestIntentMatchesResource(t, pool, claimedResource)
	if _, claimed, err := manager.Claim(ctx, jobID); err != nil {
		t.Fatal(err)
	} else if claimed {
		t.Error("a live execution allowed a second claim")
	}
	var failures int
	if err := pool.QueryRow(ctx, `SELECT handler_failure_count FROM jobs WHERE job_id = $1`, jobID).Scan(&failures); err != nil {
		t.Fatal(err)
	}
	if failures != 0 {
		t.Errorf("duplicate claim changed failure count to %d", failures)
	}

	if _, err := pool.Exec(ctx, `UPDATE jobs SET handler_lease_expires_at = now() - interval '1 second' WHERE job_id = $1`, jobID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.Claim(ctx, jobID); err != nil {
		t.Fatal(err)
	}
	var nextAttemptAt *time.Time
	var attemptID *uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT handler_failure_count, handler_next_attempt_at, handler_attempt_id FROM jobs WHERE job_id = $1`, jobID).Scan(&failures, &nextAttemptAt, &attemptID); err != nil {
		t.Fatal(err)
	}
	if failures != 1 || nextAttemptAt == nil || attemptID != nil {
		t.Fatalf("expired execution accounting = failures=%d next=%v attempt=%v", failures, nextAttemptAt, attemptID)
	}
	if _, err := manager.HandlerPayload(ctx, execution); !errors.Is(err, jobs.ErrExecutionLost) {
		t.Fatalf("expired execution payload error = %v; want ErrExecutionLost", err)
	}
	beforeLostOperations, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	beforeLostIntents := correctionIntentCount(t, pool, jobID)
	total := 1
	lostOperations := []struct {
		name string
		call func() error
	}{
		{name: "observe", call: func() error {
			_, err := manager.ObserveExecution(ctx, execution)
			return err
		}},
		{name: "progress", call: func() error {
			_, err := manager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 1, Total: &total}, nil)
			return err
		}},
		{name: "renew", call: func() error { return manager.RenewExecution(ctx, execution) }},
		{name: "release", call: func() error { return manager.ReleaseExecution(ctx, execution) }},
		{name: "failure", call: func() error { return manager.RecordExecutionFailure(ctx, execution, false) }},
		{name: "success completion", call: func() error {
			_, err := manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{
				Progress:      jobs.Progress{Completed: 1, Total: &total},
				ResultSummary: jobs.ResultSummary{Code: "done", Message: "Done."},
			})
			return err
		}},
		{name: "failure completion", call: func() error {
			_, err := manager.CompleteFailed(ctx, execution, jobs.FailureCompletion{
				Progress:     jobs.Progress{Completed: 0, Total: &total},
				ErrorSummary: jobs.ErrorSummary{Code: "failed", Message: "Failed."},
			})
			return err
		}},
		{name: "cancellation completion", call: func() error {
			_, err := manager.CompleteCanceled(ctx, execution, jobs.CancellationCompletion{
				Progress: jobs.Progress{Completed: 0, Total: &total},
			})
			return err
		}},
	}
	for _, operation := range lostOperations {
		if err := operation.call(); !errors.Is(err, jobs.ErrExecutionLost) {
			t.Errorf("expired execution %s error = %v; want ErrExecutionLost", operation.name, err)
		}
	}
	afterLostOperations, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	if !samePublicJobState(beforeLostOperations, afterLostOperations) {
		t.Fatalf("expired execution mutated job: before=%#v after=%#v", beforeLostOperations, afterLostOperations)
	}
	if got := correctionIntentCount(t, pool, jobID); got != beforeLostIntents {
		t.Fatalf("expired execution emitted intent: got %d want %d", got, beforeLostIntents)
	}
	if _, err := pool.Exec(ctx, `UPDATE jobs SET handler_next_attempt_at = now() - interval '1 second' WHERE job_id = $1`, jobID); err != nil {
		t.Fatal(err)
	}
	successor, claimed, err := manager.Claim(ctx, jobID)
	if err != nil || !claimed {
		t.Fatalf("claim successor execution = %t/%v", claimed, err)
	}
	if _, err := manager.ObserveExecution(ctx, execution); !errors.Is(err, jobs.ErrExecutionLost) {
		t.Fatalf("superseded execution observation error = %v; want ErrExecutionLost", err)
	}
	if _, err := manager.ObserveExecution(ctx, successor); err != nil {
		t.Fatalf("successor execution observation: %v", err)
	}

	t.Run("concurrent claim has one winner", func(t *testing.T) {
		concurrentID := createCorrectionJob(t, manager, actorID, incidentID, "test.concurrent")
		var winners atomic.Int32
		var wg sync.WaitGroup
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, claimed, claimErr := manager.Claim(ctx, concurrentID)
				if claimErr != nil {
					t.Errorf("concurrent claim: %v", claimErr)
					return
				}
				if claimed {
					winners.Add(1)
				}
			}()
		}
		wg.Wait()
		if got := winners.Load(); got != 1 {
			t.Errorf("concurrent claim winners = %d; want 1", got)
		}
	})

}

func TestRunnerFailureSecurity_Integration(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID, pool := newJobsHarnessWithPool(
		t,
		"jobs-runner-failure-security",
		func() time.Time { return time.Now().UTC() },
	)
	t.Run("exhaustion publishes fixed safe summary", func(t *testing.T) {
		const sentinel = "SENTINEL_EXHAUSTED_HANDLER_SECRET"
		jobID := createCorrectionJob(t, manager, actorID, incidentID, "test.exhaustion")
		for attempt := 1; attempt <= jobs.DefaultHandlerMaxAttempts; attempt++ {
			execution := claimTestExecution(t, manager, jobID)
			if err := manager.RecordExecutionFailure(ctx, execution, false); err != nil {
				t.Fatal(err)
			}
			if attempt < jobs.DefaultHandlerMaxAttempts {
				if _, err := pool.Exec(ctx, `UPDATE jobs SET handler_next_attempt_at = now() - interval '1 second' WHERE job_id = $1`, jobID); err != nil {
					t.Fatal(err)
				}
			}
		}
		resource, err := manager.Get(ctx, jobID)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(resource)
		if err != nil {
			t.Fatal(err)
		}
		var lastError *string
		var attemptID *uuid.UUID
		if err := pool.QueryRow(ctx, `SELECT handler_last_error, handler_attempt_id FROM jobs WHERE job_id = $1`, jobID).Scan(&lastError, &attemptID); err != nil {
			t.Fatal(err)
		}
		if resource.Status != jobs.StatusFailed || resource.ErrorSummary == nil ||
			resource.ErrorSummary.Code != jobs.HandlerAttemptsExhausted || attemptID != nil {
			t.Errorf("unsafe exhaustion state: resource=%#v attempt_id=%v", resource, attemptID)
		}
		if strings.Contains(string(encoded), sentinel) || lastError == nil || strings.Contains(*lastError, sentinel) {
			t.Errorf("exhaustion leaked raw handler text: resource=%s last_error=%v", encoded, lastError)
		}
	})

	gate := &dequeueGate{}
	gate.open.Store(true)
	runner := newTestRunner(t, manager, gate)
	t.Cleanup(func() { closeRunner(t, runner) })
	nilHandled := make(chan struct{}, 1)
	errorHandled := make(chan struct{}, 1)
	panicHandled := make(chan struct{}, 1)
	nilJobID := createCorrectionJob(t, manager, actorID, incidentID, "test.nil")
	errorJobID := createCorrectionJob(t, manager, actorID, incidentID, "test.error")
	panicJobID := createCorrectionJob(t, manager, actorID, incidentID, "test.panic")
	registerAllTestHandlers(t, runner, map[string]jobs.HandlerFunc{
		"test.nil": func(context.Context, jobs.Execution) error {
			nilHandled <- struct{}{}
			return nil
		},
		"test.error": func(context.Context, jobs.Execution) error {
			errorHandled <- struct{}{}
			return errors.New("SENTINEL_RAW_HANDLER_SECRET")
		},
		"test.panic": func(context.Context, jobs.Execution) error {
			panicHandled <- struct{}{}
			panic("SENTINEL_PANIC_SECRET")
		},
	})
	if err := runner.Activate(ctx); err != nil {
		t.Fatal(err)
	}

	t.Run("nil return is incomplete", func(t *testing.T) {
		<-nilHandled
		lastError := waitForHandlerAttemptResult(t, pool, nilJobID)
		if lastError != jobs.HandlerIncomplete {
			t.Errorf("nil mutable handler result = %q; want %s", lastError, jobs.HandlerIncomplete)
		}
	})

	t.Run("raw error is secret", func(t *testing.T) {
		<-errorHandled
		lastError := waitForHandlerAttemptResult(t, pool, errorJobID)
		if strings.Contains(lastError, "SENTINEL_RAW_HANDLER_SECRET") || lastError != jobs.HandlerExecutionFailed {
			t.Errorf("unsafe handler error persisted as %q", lastError)
		}
	})

	t.Run("panic value is secret", func(t *testing.T) {
		<-panicHandled
		lastError := waitForHandlerAttemptResult(t, pool, panicJobID)
		if strings.Contains(lastError, "SENTINEL_PANIC_SECRET") || lastError != jobs.HandlerExecutionFailed {
			t.Errorf("unsafe panic value persisted as %q", lastError)
		}
	})
}

func createCorrectionJob(t testing.TB, manager *jobs.Manager, actorID, incidentID uuid.UUID, handlerName string) uuid.UUID {
	t.Helper()
	jobKind := testJobKind
	if handlerName != "" {
		jobKind = collaborationsupport.TestJobKindForHandler(handlerName)
	}
	resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           jobKind,
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create correction job: %v", err)
	}
	return uuid.MustParse(resource.JobID)
}

func assertRejectedProgressUnchanged(
	t testing.TB,
	manager *jobs.Manager,
	pool *pgxpool.Pool,
	execution jobs.Execution,
	progress jobs.Progress,
) {
	t.Helper()
	jobID := execution.JobID()
	before, err := manager.Get(context.Background(), jobID)
	if err != nil {
		t.Fatal(err)
	}
	beforeIntents := correctionIntentCount(t, pool, jobID)
	if _, err := manager.UpdateProgress(context.Background(), execution, progress, nil); !errors.Is(err, jobs.ErrInvalidTransition) {
		t.Errorf("rejected progress error = %v; want ErrInvalidTransition", err)
	}
	after, err := manager.Get(context.Background(), jobID)
	if err != nil {
		t.Fatal(err)
	}
	if !samePublicJobState(before, after) {
		t.Errorf("rejected progress mutated job: before=%#v after=%#v", before, after)
	}
	if got := correctionIntentCount(t, pool, jobID); got != beforeIntents {
		t.Errorf("rejected progress emitted intent: got %d want %d", got, beforeIntents)
	}
}

func samePublicJobState(left, right jobs.Resource) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}

func correctionIntentCount(t testing.TB, pool *pgxpool.Pool, jobID uuid.UUID) int {
	t.Helper()
	return collaborationsupport.CountIntents(t, pool, collaborationsupport.IntentSelector{
		SourceIdentity: "job:" + jobID.String(),
	})
}

func assertLatestIntentMatchesResource(t testing.TB, pool *pgxpool.Pool, resource jobs.Resource) {
	t.Helper()
	payload := collaborationsupport.LoadLatestIntent(t, pool, collaborationsupport.IntentSelector{
		SourceIdentity: "job:" + resource.JobID,
	}).CanonicalPayload
	var projected struct {
		Status    string        `json:"status"`
		Progress  jobs.Progress `json:"progress"`
		UpdatedAt time.Time     `json:"updated_at"`
	}
	if err := json.Unmarshal(payload, &projected); err != nil {
		t.Fatal(err)
	}
	if projected.Status != resource.Status || projected.Progress.Completed != resource.Progress.Completed ||
		!projected.UpdatedAt.Equal(resource.UpdatedAt) ||
		(projected.Progress.Total == nil) != (resource.Progress.Total == nil) ||
		(projected.Progress.Total != nil && *projected.Progress.Total != *resource.Progress.Total) {
		t.Errorf("job/intent projection mismatch: resource=%#v payload=%s", resource, payload)
	}
}

func waitForHandlerAttemptResult(t testing.TB, pool *pgxpool.Pool, jobID uuid.UUID) string {
	t.Helper()
	deadline := time.Now().Add(750 * time.Millisecond)
	for {
		var value *string
		if err := pool.QueryRow(context.Background(), `SELECT handler_last_error FROM jobs WHERE job_id = $1`, jobID).Scan(&value); err != nil {
			t.Fatal(err)
		}
		if value != nil {
			return *value
		}
		if time.Now().After(deadline) {
			return ""
		}
		time.Sleep(10 * time.Millisecond)
	}
}
