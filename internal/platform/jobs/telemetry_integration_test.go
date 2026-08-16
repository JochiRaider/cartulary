package jobs_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	telemetrytest "github.com/JochiRaider/cartulary/internal/platform/telemetry/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

func TestJobAttemptTelemetryUsesCatalogKindsAndClosedOutcomes_Integration(t *testing.T) {
	ctx := context.Background()
	capture := telemetrytest.StartCapture()
	t.Cleanup(func() { capture.Close(context.Background()) })

	manager, actorID, incidentID, pool := newJobsHarnessWithPool(t, "jobs-attempt-telemetry", func() time.Time { return time.Now().UTC() })
	activeKinds := []string{testJobKind, collaborationsupport.TestJobKindForHandler("test.complete")}
	for index, kind := range activeKinds {
		resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
			JobKind: kind, Scope: jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
			SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Completed: index},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, claimed, err := manager.Claim(ctx, uuid.MustParse(resource.JobID)); err != nil || !claimed {
			t.Fatalf("claim active gauge job = %t, %v", claimed, err)
		}
	}

	completeJob, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           collaborationsupport.TestJobKindForHandler("test.complete"),
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Completed: 0},
		HandlerPayload: []byte(`{"secret":"SENTINEL_JOB_PAYLOAD","path":"/private/jobs/input"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	lostJob, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           collaborationsupport.TestJobKindForHandler("test.error"),
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO jobs (
    scope_kind, status, cancelable, submitted_by_user_id, submitted_at, updated_at,
    progress_completed, finished_at, retained_until, result_summary_json,
    auth_policy, handler_name, job_kind, progress_unit_id
) VALUES (
    'deployment', 'succeeded', false, $1, now() - interval '8 days',
    now() - interval '7 days', 1, now() - interval '7 days', now() - interval '1 second',
    '{"code":"done","message":"Done."}', 'deployment_admin',
    'test_platform.worker_v1', $2, 'test_platform.generic.operation.v1'
)
`, actorID, testJobKind); err != nil {
		t.Fatal(err)
	}

	composition := testCompositionForManager(t, manager)
	policy := jobs.ProductionRuntimePolicy()
	policy.HandlerLease = 300 * time.Millisecond
	policy.LeaseRenewal = 50 * time.Millisecond
	policy.RecoveryScan = 30 * time.Second
	policy.ExpirySweep = 150 * time.Millisecond
	gate := &dequeueGate{}
	gate.open.Store(true)
	runner, err := jobs.NewRunner(jobs.RunnerOptions{
		Manager: manager, Catalog: composition.catalog, Policy: policy, DequeueGate: gate,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeRunner(t, runner) })
	completeStarted := make(chan jobs.Execution, 1)
	lostStarted := make(chan jobs.Execution, 1)
	releaseComplete := make(chan struct{})
	registerAllTestHandlers(t, runner, map[string]jobs.HandlerFunc{
		"test.complete": func(handlerCtx context.Context, execution jobs.Execution) error {
			if execution.JobID() != uuid.MustParse(completeJob.JobID) {
				return errors.New("SENTINEL_UNEXPECTED_JOB_ID")
			}
			completeStarted <- execution
			<-releaseComplete
			_, completeErr := manager.CompleteSucceeded(handlerCtx, execution, jobs.SuccessCompletion{
				Progress:      jobs.Progress{Completed: 1},
				ResultSummary: jobs.ResultSummary{Code: "telemetry_complete", Message: "Complete."},
			})
			return completeErr
		},
		"test.error": func(handlerCtx context.Context, execution jobs.Execution) error {
			if execution.JobID() != uuid.MustParse(lostJob.JobID) {
				return errors.New("SENTINEL_UNEXPECTED_JOB_ID")
			}
			lostStarted <- execution
			<-handlerCtx.Done()
			return handlerCtx.Err()
		},
		"test.nil": func(context.Context, jobs.Execution) error {
			return errors.New("SENTINEL_RAW_HANDLER_ERROR_SECRET")
		},
		"test.panic": func(context.Context, jobs.Execution) error {
			panic("SENTINEL_PANIC_VALUE_SECRET")
		},
	})
	if err := runner.Activate(ctx); err != nil {
		t.Fatal(err)
	}
	runner.Notify(uuid.MustParse(completeJob.JobID))
	runner.Notify(uuid.MustParse(lostJob.JobID))
	<-completeStarted
	<-lostStarted
	var completeAttemptID string
	var lostAttemptID string
	if err := pool.QueryRow(ctx, `SELECT handler_attempt_id::text FROM jobs WHERE job_id = $1`, uuid.MustParse(completeJob.JobID)).Scan(&completeAttemptID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT handler_attempt_id::text FROM jobs WHERE job_id = $1`, uuid.MustParse(lostJob.JobID)).Scan(&lostAttemptID); err != nil {
		t.Fatal(err)
	}
	if got := len(capture.EndedSpans()); got != 0 {
		t.Fatalf("attempt span ended before handler classification: %d", got)
	}
	time.Sleep(60 * time.Millisecond)
	if got := len(capture.EndedSpans()); got != 0 {
		t.Fatalf("attempt span did not cover blocked handler: %d", got)
	}
	errorJob, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           collaborationsupport.TestJobKindForHandler("test.nil"),
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Completed: 0},
		HandlerPayload: []byte(`{"secret":"SENTINEL_HANDLER_PAYLOAD_SECRET","path":"/private/jobs/error"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	panicJob, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           collaborationsupport.TestJobKindForHandler("test.panic"),
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	runner.Notify(uuid.MustParse(errorJob.JobID))
	runner.Notify(uuid.MustParse(panicJob.JobID))
	if _, err := pool.Exec(ctx, `UPDATE jobs SET handler_attempt_id = $2 WHERE job_id = $1`, uuid.MustParse(lostJob.JobID), uuid.New()); err != nil {
		t.Fatal(err)
	}
	close(releaseComplete)

	deadline := time.Now().Add(3 * time.Second)
	for (len(capture.EndedSpans()) < 4 || !jobExpired(t, pool, ctx)) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	spans := capture.EndedSpans()
	if len(spans) != 4 {
		t.Fatalf("ended attempt spans = %d; want 4", len(spans))
	}
	results := map[string]string{}
	terminalByKind := map[string]string{}
	for _, span := range spans {
		if span.Name != "cartulary.jobs.run" || span.FinishedAt.Before(span.StartedAt) {
			t.Fatalf("attempt span boundary = %s/%v", span.Name, span.FinishedAt.Sub(span.StartedAt))
		}
		attrs := span.Attributes
		kind := attrs["cartulary.job_kind"]
		if (kind == collaborationsupport.TestJobKindForHandler("test.complete") ||
			kind == collaborationsupport.TestJobKindForHandler("test.error")) &&
			span.FinishedAt.Sub(span.StartedAt) < 50*time.Millisecond {
			t.Fatalf("blocked attempt span ended too early for %s: %v", kind, span.FinishedAt.Sub(span.StartedAt))
		}
		results[kind] = attrs["cartulary.result"]
		terminalByKind[kind] = attrs["cartulary.job_terminal_status"]
		for _, value := range attrs {
			for _, forbidden := range []string{
				"SENTINEL", "/private/", completeJob.JobID, lostJob.JobID, errorJob.JobID, panicJob.JobID,
				completeAttemptID, lostAttemptID, "attempt", "operation.v1",
			} {
				if strings.Contains(value, forbidden) {
					t.Fatalf("attempt telemetry leaked %q in %#v", forbidden, attrs)
				}
			}
		}
	}
	if kind := collaborationsupport.TestJobKindForHandler("test.complete"); results[kind] != "success" || terminalByKind[kind] != jobs.StatusSucceeded {
		t.Fatalf("success attempt classification = result %q terminal %q", results[kind], terminalByKind[kind])
	}
	if kind := collaborationsupport.TestJobKindForHandler("test.error"); results[kind] != "conflict" || terminalByKind[kind] != "" {
		t.Fatalf("lost attempt classification = result %q terminal %q", results[kind], terminalByKind[kind])
	}
	for _, kind := range []string{
		collaborationsupport.TestJobKindForHandler("test.nil"),
		collaborationsupport.TestJobKindForHandler("test.panic"),
	} {
		if results[kind] != "failed" || terminalByKind[kind] != "" {
			t.Fatalf("failed attempt classification for %s = result %q terminal %q", kind, results[kind], terminalByKind[kind])
		}
	}

	metrics, err := capture.MetricPoints(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assertJobTelemetryMetrics(t, metrics, activeKinds)
}

func TestJobQueuedAndQueueWaitTelemetryUsesDurableEligibility_Integration(t *testing.T) {
	ctx := context.Background()
	base := time.Date(2026, 8, 16, 18, 0, 0, 0, time.UTC)
	clock := base
	capture := telemetrytest.StartCapture()
	t.Cleanup(func() { capture.Close(context.Background()) })
	manager, actorID, incidentID, pool := newJobsHarnessWithPool(t, "jobs-queue-telemetry", func() time.Time { return clock })

	claimable, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           collaborationsupport.TestJobKindForHandler("test.complete"),
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	retryDelayed, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           collaborationsupport.TestJobKindForHandler("test.error"),
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
UPDATE jobs
   SET status = 'running',
       handler_failure_count = 1,
       handler_next_attempt_at = $2
 WHERE job_id = $1
`, uuid.MustParse(retryDelayed.JobID), base.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	clock = base.Add(30 * time.Second)
	if _, claimed, err := manager.Claim(ctx, uuid.MustParse(claimable.JobID)); err != nil || !claimed {
		t.Fatalf("claim queue-wait fixture = %t, %v", claimed, err)
	}
	if _, claimed, err := manager.Claim(ctx, uuid.MustParse(claimable.JobID)); err != nil || claimed {
		t.Fatalf("duplicate claim emitted = %t, %v", claimed, err)
	}
	if _, claimed, err := manager.Claim(ctx, uuid.MustParse(retryDelayed.JobID)); err != nil || claimed {
		t.Fatalf("retry-delayed claim emitted = %t, %v", claimed, err)
	}

	points, err := capture.MetricPoints(ctx)
	if err != nil {
		t.Fatal(err)
	}
	queueWaitKind := collaborationsupport.TestJobKindForHandler("test.complete")
	queuedKind := collaborationsupport.TestJobKindForHandler("test.error")
	queueWaitCount := 0
	queuedCount := int64(-1)
	for _, point := range points {
		switch point.Name {
		case "cartulary.jobs.queue_wait.duration":
			if point.Attributes["cartulary.job_kind"] == queueWaitKind {
				queueWaitCount++
				if !point.IsFloat || point.FloatValue != 30 || len(point.Attributes) != 1 {
					t.Fatalf("queue-wait point = %#v", point)
				}
			}
		case "cartulary.jobs.queued":
			if point.Attributes["cartulary.job_kind"] == queuedKind {
				queuedCount = point.Value
			}
		}
	}
	if queueWaitCount != 1 {
		t.Fatalf("successful claim queue-wait points = %d want 1: %#v", queueWaitCount, points)
	}
	if queuedCount != 1 {
		t.Fatalf("retry-delayed queued gauge = %d want 1: %#v", queuedCount, points)
	}

	invalid, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           collaborationsupport.TestJobKindForHandler("test.nil"),
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID, Cancelable: true, Progress: jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE jobs SET submitted_at = $2 WHERE job_id = $1`, uuid.MustParse(invalid.JobID), clock.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, claimed, err := manager.Claim(ctx, uuid.MustParse(invalid.JobID)); !errors.Is(err, jobs.ErrInvalidTransition) || claimed {
		t.Fatalf("negative queue wait invariant = claimed %t err %v", claimed, err)
	}
}

func jobExpired(t testing.TB, pool *pgxpool.Pool, ctx context.Context) bool {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM jobs WHERE expired_at IS NOT NULL`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count == 1
}

func assertJobTelemetryMetrics(t testing.TB, points []telemetrytest.MetricPoint, activeKinds []string) {
	t.Helper()
	seen := map[string][]map[string]string{}
	for _, point := range points {
		if point.Value > 0 {
			seen[point.Name] = append(seen[point.Name], point.Attributes)
		}
	}
	for _, kind := range activeKinds {
		if !containsMetricKind(seen["cartulary.jobs.active"], kind) {
			t.Fatalf("active gauge omitted catalog kind %q: %#v", kind, seen["cartulary.jobs.active"])
		}
	}
	if len(seen["cartulary.jobs.attempts"]) != 4 ||
		!containsMetricResult(seen["cartulary.jobs.attempts"], "success") ||
		!containsMetricResult(seen["cartulary.jobs.attempts"], "failed") ||
		!containsMetricResult(seen["cartulary.jobs.attempts"], "conflict") {
		t.Fatalf("attempt counter = %#v", seen["cartulary.jobs.attempts"])
	}
	if !containsMetricResult(seen["cartulary.jobs.lease_renewal.failures"], "conflict") {
		t.Fatalf("lease-renewal counter = %#v", seen["cartulary.jobs.lease_renewal.failures"])
	}
	if !containsMetricKind(seen["cartulary.jobs.expired"], testJobKind) {
		t.Fatalf("expiry counter = %#v", seen["cartulary.jobs.expired"])
	}
}

func containsMetricKind(points []map[string]string, kind string) bool {
	for _, point := range points {
		if point["cartulary.job_kind"] == kind {
			return true
		}
	}
	return false
}

func containsMetricResult(points []map[string]string, result string) bool {
	for _, point := range points {
		if point["cartulary.result"] == result {
			return true
		}
	}
	return false
}
