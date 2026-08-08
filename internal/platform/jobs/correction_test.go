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
)

func TestTransitionAndProgressCorrection_Integration(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID, pool := newJobsHarnessWithPool(
		t,
		"jobs-transition-progress-correction",
		func() time.Time { return time.Now().UTC() },
	)

	terminalAllowed := map[string]map[string]bool{
		jobs.StatusSucceeded: {
			jobs.StatusRunning:         true,
			jobs.StatusCancelRequested: true,
		},
		jobs.StatusFailed: {
			jobs.StatusRunning:         true,
			jobs.StatusCancelRequested: true,
		},
		jobs.StatusCanceled: {
			jobs.StatusCancelRequested: true,
		},
	}
	states := []string{
		jobs.StatusQueued,
		jobs.StatusRunning,
		jobs.StatusCancelRequested,
		jobs.StatusSucceeded,
		jobs.StatusFailed,
		jobs.StatusCanceled,
	}
	for target, allowedSources := range terminalAllowed {
		for _, source := range states {
			t.Run("terminal_"+source+"_to_"+target, func(t *testing.T) {
				jobID := createCorrectionJob(t, manager, actorID, incidentID, "")
				transitionCorrectionJobTo(t, manager, actorID, jobID, source)
				before, err := manager.Get(ctx, jobID)
				if err != nil {
					t.Fatal(err)
				}
				got, err := completeCorrectionJob(manager, jobID, target, before.Progress)
				if allowedSources[source] {
					if err != nil {
						t.Errorf("legal %s -> %s rejected: %v", source, target, err)
					} else if got.Status != target {
						t.Errorf("legal %s -> %s produced %s", source, target, got.Status)
					}
					return
				}
				if !errors.Is(err, jobs.ErrInvalidTransition) {
					t.Errorf("illegal %s -> %s error = %v; want ErrInvalidTransition", source, target, err)
				}
				after, loadErr := manager.Get(ctx, jobID)
				if loadErr != nil {
					t.Fatal(loadErr)
				}
				if err != nil && !samePublicJobState(before, after) {
					t.Errorf("rejected %s -> %s mutated job: before=%#v after=%#v", source, target, before, after)
				}
			})
		}
	}

	t.Run("running progress monotonicity and no-op", func(t *testing.T) {
		jobID := createCorrectionJob(t, manager, actorID, incidentID, "")
		if _, err := manager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 2}, nil); err != nil {
			t.Fatal(err)
		}
		before, err := manager.Get(ctx, jobID)
		if err != nil {
			t.Fatal(err)
		}
		beforeIntents := correctionIntentCount(t, pool, jobID)
		repeated, err := manager.MarkRunning(ctx, jobID, before.Progress, before.Message)
		if err != nil {
			t.Errorf("exact repeated progress rejected: %v", err)
		} else if !samePublicJobState(before, repeated) {
			t.Errorf("exact repeated progress was not a true no-op: before=%#v after=%#v", before, repeated)
		}
		if got := correctionIntentCount(t, pool, jobID); got != beforeIntents {
			t.Errorf("exact repeated progress emitted intent: got %d want %d", got, beforeIntents)
		}

		total := 5
		discovered, err := manager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 2, Total: &total}, nil)
		if err != nil {
			t.Errorf("total discovery rejected: %v", err)
		} else if discovered.Progress.Total == nil || *discovered.Progress.Total != total {
			t.Errorf("total discovery result = %#v", discovered.Progress)
		}
		if _, err := manager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 3, Total: &total}, nil); err != nil {
			t.Errorf("monotonic progress increase rejected: %v", err)
		}

		assertRejectedProgressUnchanged(t, manager, pool, jobID, jobs.Progress{Completed: 1, Total: &total})
		lowerTotal := 4
		assertRejectedProgressUnchanged(t, manager, pool, jobID, jobs.Progress{Completed: 3, Total: &lowerTotal})
		assertRejectedProgressUnchanged(t, manager, pool, jobID, jobs.Progress{Completed: 3})
		overTotal := 2
		assertRejectedProgressUnchanged(t, manager, pool, jobID, jobs.Progress{Completed: 3, Total: &overTotal})
	})

	t.Run("known total success requires equality", func(t *testing.T) {
		jobID := createCorrectionJob(t, manager, actorID, incidentID, "")
		total := 5
		if _, err := manager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 4, Total: &total}, nil); err != nil {
			t.Fatal(err)
		}
		before, err := manager.Get(ctx, jobID)
		if err != nil {
			t.Fatal(err)
		}
		_, err = manager.CompleteSucceeded(ctx, jobs.TransitionParams{
			JobID: jobID,
			Progress: jobs.Progress{
				Completed: 4,
				Total:     &total,
			},
			ResultSummary: &jobs.ResultSummary{Code: "correction_complete", Message: "Complete."},
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
		total := 10
		if _, err := manager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 0, Total: &total}, nil); err != nil {
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
				_, updateErr := manager.MarkRunning(ctx, jobID, jobs.Progress{Completed: value, Total: &total}, nil)
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
	if claimed, err := manager.ClaimHandlerJob(ctx, jobID, "test.claim", "attempt-1", time.Minute); err != nil || !claimed {
		t.Fatalf("initial claim = %t, %v", claimed, err)
	}
	claimedResource, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	if claimedResource.Status != jobs.StatusRunning || claimedResource.StartedAt == nil {
		t.Errorf("handler claim did not publish running before execution: %#v", claimedResource)
	}
	assertLatestIntentMatchesResource(t, pool, claimedResource)
	if claimed, err := manager.ClaimHandlerJob(ctx, jobID, "test.claim", "attempt-1", time.Minute); err != nil {
		t.Fatal(err)
	} else if claimed {
		t.Error("one attempt identity claimed the same job twice")
	}
	var attempts int
	if err := pool.QueryRow(ctx, `SELECT handler_attempts FROM jobs WHERE job_id = $1`, jobID).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if attempts != 1 {
		t.Errorf("duplicate attempt identity incremented attempts to %d", attempts)
	}

	if _, err := pool.Exec(ctx, `UPDATE jobs SET handler_lease_expires_at = now() - interval '1 second' WHERE job_id = $1`, jobID); err != nil {
		t.Fatal(err)
	}
	beforeReclaim, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	beforeReclaimIntents := correctionIntentCount(t, pool, jobID)
	if claimed, err := manager.ClaimHandlerJob(ctx, jobID, "test.claim", "attempt-2", time.Minute); err != nil || !claimed {
		t.Errorf("expired running claim = %t, %v", claimed, err)
	}
	afterReclaim, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	if afterReclaim.Status != jobs.StatusRunning || !afterReclaim.UpdatedAt.Equal(beforeReclaim.UpdatedAt) {
		t.Errorf("running reclaim changed public state: before=%#v after=%#v", beforeReclaim, afterReclaim)
	}
	if got := correctionIntentCount(t, pool, jobID); got != beforeReclaimIntents {
		t.Errorf("lease-only running reclaim emitted intent: got %d want %d", got, beforeReclaimIntents)
	}

	if _, err := manager.Cancel(ctx, jobs.CancelParams{
		JobID: jobID, ActorUserID: actorID, ClientTxnID: "cancel-claim-recovery",
		NormalizedRequest: []byte(`{"client_txn_id":"cancel-claim-recovery"}`),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE jobs SET handler_lease_expires_at = now() - interval '1 second' WHERE job_id = $1`, jobID); err != nil {
		t.Fatal(err)
	}
	beforeCancelReclaim, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	beforeCancelIntents := correctionIntentCount(t, pool, jobID)
	if claimed, err := manager.ClaimHandlerJob(ctx, jobID, "test.claim", "attempt-3", time.Minute); err != nil || !claimed {
		t.Errorf("expired cancel-requested claim = %t, %v", claimed, err)
	}
	afterCancelReclaim, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	if afterCancelReclaim.Status != jobs.StatusCancelRequested || !afterCancelReclaim.UpdatedAt.Equal(beforeCancelReclaim.UpdatedAt) {
		t.Errorf("cancel-requested reclaim regressed public state: before=%#v after=%#v", beforeCancelReclaim, afterCancelReclaim)
	}
	if got := correctionIntentCount(t, pool, jobID); got != beforeCancelIntents {
		t.Errorf("lease-only cancellation reclaim emitted intent: got %d want %d", got, beforeCancelIntents)
	}

	t.Run("concurrent claim has one winner", func(t *testing.T) {
		concurrentID := createCorrectionJob(t, manager, actorID, incidentID, "test.concurrent")
		var winners atomic.Int32
		var wg sync.WaitGroup
		for _, attemptID := range []string{"concurrent-attempt-1", "concurrent-attempt-2"} {
			wg.Add(1)
			go func(identity string) {
				defer wg.Done()
				claimed, claimErr := manager.ClaimHandlerJob(ctx, concurrentID, "test.concurrent", identity, time.Minute)
				if claimErr != nil {
					t.Errorf("concurrent claim %s: %v", identity, claimErr)
					return
				}
				if claimed {
					winners.Add(1)
				}
			}(attemptID)
		}
		wg.Wait()
		if got := winners.Load(); got != 1 {
			t.Errorf("concurrent claim winners = %d; want 1", got)
		}
	})

	t.Run("one runner invokes one handler", func(t *testing.T) {
		duplicateID := createCorrectionJob(t, manager, actorID, incidentID, "test.duplicate")
		runner := jobs.NewRunner()
		runner.Configure(manager)
		entered := make(chan struct{}, 2)
		release := make(chan struct{})
		if err := runner.RegisterHandler("test.duplicate", func(context.Context, uuid.UUID) error {
			entered <- struct{}{}
			<-release
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if err := runner.DispatchJobID("test.duplicate", duplicateID); err != nil {
			t.Fatal(err)
		}
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			t.Fatal("first handler invocation timed out")
		}
		if err := runner.DispatchJobID("test.duplicate", duplicateID); err != nil {
			t.Fatal(err)
		}
		select {
		case <-entered:
			t.Error("duplicate dispatch invoked the handler twice within one runner")
		case <-time.After(200 * time.Millisecond):
		}
		close(release)
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runner.Close(closeCtx); err != nil {
			t.Fatal(err)
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
	runner := jobs.NewRunner()
	runner.Configure(manager)
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runner.Close(closeCtx); err != nil {
			t.Errorf("close runner: %v", err)
		}
	})

	registerAndDispatch := func(t *testing.T, handlerName string, handler jobs.HandlerFunc) uuid.UUID {
		t.Helper()
		jobID := createCorrectionJob(t, manager, actorID, incidentID, handlerName)
		if err := runner.RegisterHandler(handlerName, handler); err != nil {
			t.Fatal(err)
		}
		if err := runner.DispatchJobID(handlerName, jobID); err != nil {
			t.Fatal(err)
		}
		return jobID
	}

	t.Run("nil return is incomplete", func(t *testing.T) {
		handled := make(chan struct{}, 1)
		jobID := registerAndDispatch(t, "test.nil", func(context.Context, uuid.UUID) error {
			handled <- struct{}{}
			return nil
		})
		<-handled
		lastError := waitForHandlerAttemptResult(t, pool, jobID)
		if lastError != "job_handler_incomplete" {
			t.Errorf("nil mutable handler result = %q; want job_handler_incomplete", lastError)
		}
	})

	t.Run("raw error is secret", func(t *testing.T) {
		const sentinel = "SENTINEL_RAW_HANDLER_SECRET"
		handled := make(chan struct{}, 1)
		jobID := registerAndDispatch(t, "test.error", func(context.Context, uuid.UUID) error {
			handled <- struct{}{}
			return errors.New(sentinel)
		})
		<-handled
		lastError := waitForHandlerAttemptResult(t, pool, jobID)
		if strings.Contains(lastError, sentinel) || lastError != jobs.HandlerExecutionFailed {
			t.Errorf("unsafe handler error persisted as %q", lastError)
		}
	})

	t.Run("panic value is secret", func(t *testing.T) {
		const sentinel = "SENTINEL_PANIC_SECRET"
		handled := make(chan struct{}, 1)
		jobID := registerAndDispatch(t, "test.panic", func(context.Context, uuid.UUID) error {
			handled <- struct{}{}
			panic(sentinel)
		})
		<-handled
		lastError := waitForHandlerAttemptResult(t, pool, jobID)
		if strings.Contains(lastError, sentinel) || lastError != jobs.HandlerExecutionFailed {
			t.Errorf("unsafe panic value persisted as %q", lastError)
		}
	})

	t.Run("exhaustion publishes fixed safe summary", func(t *testing.T) {
		const sentinel = "SENTINEL_EXHAUSTED_HANDLER_SECRET"
		jobID := createCorrectionJob(t, manager, actorID, incidentID, "test.exhaustion")
		for attempt := 1; attempt <= jobs.DefaultHandlerMaxAttempts; attempt++ {
			attemptID := "exhaustion-attempt-" + string(rune('0'+attempt))
			if claimed, err := manager.ClaimHandlerJob(ctx, jobID, "test.exhaustion", attemptID, time.Minute); err != nil || !claimed {
				t.Fatalf("claim attempt %d = %t, %v", attempt, claimed, err)
			}
			if err := manager.RecordHandlerFailure(ctx, jobID, attemptID); err != nil {
				t.Fatal(err)
			}
			if attempt < jobs.DefaultHandlerMaxAttempts {
				if _, err := pool.Exec(ctx, `UPDATE jobs SET handler_lease_expires_at = now() - interval '1 second' WHERE job_id = $1`, jobID); err != nil {
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
		var leaseOwner *string
		if err := pool.QueryRow(ctx, `SELECT handler_last_error, handler_lease_owner FROM jobs WHERE job_id = $1`, jobID).Scan(&lastError, &leaseOwner); err != nil {
			t.Fatal(err)
		}
		if resource.Status != jobs.StatusFailed || resource.ErrorSummary == nil ||
			resource.ErrorSummary.Code != jobs.HandlerAttemptsExhausted || leaseOwner != nil {
			t.Errorf("unsafe exhaustion state: resource=%#v lease_owner=%v", resource, leaseOwner)
		}
		if strings.Contains(string(encoded), sentinel) || lastError == nil || strings.Contains(*lastError, sentinel) {
			t.Errorf("exhaustion leaked raw handler text: resource=%s last_error=%v", encoded, lastError)
		}
	})
}

func createCorrectionJob(t testing.TB, manager *jobs.Manager, actorID, incidentID uuid.UUID, handlerName string) uuid.UUID {
	t.Helper()
	resource, err := manager.Create(context.Background(), jobs.CreateParams{
		JobKind:           testJobKind,
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
		HandlerName:       handlerName,
	})
	if err != nil {
		t.Fatalf("create correction job: %v", err)
	}
	return uuid.MustParse(resource.JobID)
}

func transitionCorrectionJobTo(t testing.TB, manager *jobs.Manager, actorID, jobID uuid.UUID, status string) {
	t.Helper()
	ctx := context.Background()
	switch status {
	case jobs.StatusQueued:
		return
	case jobs.StatusRunning:
		if _, err := manager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 0}, nil); err != nil {
			t.Fatalf("prepare running: %v", err)
		}
	case jobs.StatusCancelRequested:
		if _, err := manager.Cancel(ctx, jobs.CancelParams{
			JobID: jobID, ActorUserID: actorID, ClientTxnID: "prepare-cancel-" + jobID.String(),
			NormalizedRequest: []byte(`{"reason":null}`),
		}); err != nil {
			t.Fatalf("prepare cancel_requested: %v", err)
		}
	case jobs.StatusSucceeded, jobs.StatusFailed:
		transitionCorrectionJobTo(t, manager, actorID, jobID, jobs.StatusRunning)
		resource, err := manager.Get(ctx, jobID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := completeCorrectionJob(manager, jobID, status, resource.Progress); err != nil {
			t.Fatalf("prepare %s: %v", status, err)
		}
	case jobs.StatusCanceled:
		transitionCorrectionJobTo(t, manager, actorID, jobID, jobs.StatusCancelRequested)
		resource, err := manager.Get(ctx, jobID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := completeCorrectionJob(manager, jobID, status, resource.Progress); err != nil {
			t.Fatalf("prepare canceled: %v", err)
		}
	default:
		t.Fatalf("unknown correction state %q", status)
	}
}

func completeCorrectionJob(manager *jobs.Manager, jobID uuid.UUID, status string, progress jobs.Progress) (jobs.Resource, error) {
	params := jobs.TransitionParams{JobID: jobID, Progress: progress}
	switch status {
	case jobs.StatusSucceeded:
		params.ResultSummary = &jobs.ResultSummary{Code: "correction_succeeded", Message: "Succeeded."}
		return manager.CompleteSucceeded(context.Background(), params)
	case jobs.StatusFailed:
		params.ErrorSummary = &jobs.ErrorSummary{Code: "correction_failed", Message: "Failed.", Retryable: false}
		return manager.CompleteFailed(context.Background(), params)
	case jobs.StatusCanceled:
		return manager.CompleteCanceled(context.Background(), params)
	default:
		return jobs.Resource{}, jobs.ErrInvalidTransition
	}
}

func assertRejectedProgressUnchanged(
	t testing.TB,
	manager *jobs.Manager,
	pool *pgxpool.Pool,
	jobID uuid.UUID,
	progress jobs.Progress,
) {
	t.Helper()
	before, err := manager.Get(context.Background(), jobID)
	if err != nil {
		t.Fatal(err)
	}
	beforeIntents := correctionIntentCount(t, pool, jobID)
	if _, err := manager.MarkRunning(context.Background(), jobID, progress, nil); !errors.Is(err, jobs.ErrInvalidTransition) {
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
	var count int
	if err := pool.QueryRow(context.Background(), `
SELECT count(*)
  FROM collaboration_event_intents
 WHERE source_identity = $1
`, "job:"+jobID.String()).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func assertLatestIntentMatchesResource(t testing.TB, pool *pgxpool.Pool, resource jobs.Resource) {
	t.Helper()
	var payload []byte
	if err := pool.QueryRow(context.Background(), `
SELECT canonical_payload
  FROM collaboration_event_intents
 WHERE source_identity = $1
 ORDER BY created_at DESC, intent_key DESC
 LIMIT 1
`, "job:"+resource.JobID).Scan(&payload); err != nil {
		t.Fatal(err)
	}
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
