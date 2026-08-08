package jobs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func TestLogicalExpiryMasksReadsAndCancelReplay_Integration(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 8, 16, 0, 0, 0, time.UTC)
	manager, actorID, incidentID := newJobsHarnessWithClock(t, "jobs-logical-expiry", func() time.Time { return now })
	resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           testJobKind,
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	jobID := uuid.MustParse(resource.JobID)
	execution := claimTestExecution(t, manager, jobID)
	cancel := jobs.CancelParams{
		JobID: jobID, ActorUserID: actorID, ClientTxnID: "expiry-replay",
		NormalizedRequest: []byte(`{"job_id":"` + jobID.String() + `"}`),
	}
	if _, err := manager.Cancel(ctx, cancel); err != nil {
		t.Fatal(err)
	}
	terminal, err := manager.CompleteCanceled(ctx, execution, jobs.CancellationCompletion{
		Progress: jobs.Progress{Completed: 0},
	})
	if err != nil || terminal.RetainedUntil == nil {
		t.Fatalf("terminal cancellation = %#v, %v", terminal, err)
	}
	cutoff := *terminal.RetainedUntil
	now = cutoff.Add(-time.Nanosecond)
	if _, err := manager.Get(ctx, jobID); err != nil {
		t.Fatalf("read immediately before cutoff: %v", err)
	}
	if replay, err := manager.Cancel(ctx, cancel); err != nil || !replay.Replayed {
		t.Fatalf("cancel replay immediately before cutoff = %#v, %v", replay, err)
	}

	now = cutoff
	if _, err := manager.Get(ctx, jobID); !errors.Is(err, jobs.ErrNotFound) {
		t.Fatalf("read at cutoff error = %v; want ErrNotFound", err)
	}
	if _, err := manager.Cancel(ctx, cancel); !errors.Is(err, jobs.ErrNotFound) {
		t.Fatalf("cancel replay at cutoff error = %v; want ErrNotFound", err)
	}
}

func TestRunnerDelaysFirstExpirySweep_Integration(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID, pool := newJobsHarnessWithPool(t, "jobs-expiry-supervisor", func() time.Time { return time.Now().UTC() })
	resource, err := enqueueTestJob(t, manager, jobs.EnqueueParams{
		JobKind:           testJobKind,
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	jobID := uuid.MustParse(resource.JobID)
	execution := claimTestExecution(t, manager, jobID)
	if _, err := manager.CompleteSucceeded(ctx, execution, jobs.SuccessCompletion{
		Progress:      jobs.Progress{Completed: 1},
		ResultSummary: jobs.ResultSummary{Code: "expiry_test_complete", Message: "Complete."},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE jobs SET retained_until = now() - interval '1 second' WHERE job_id = $1`, jobID); err != nil {
		t.Fatal(err)
	}
	composition := testCompositionForManager(t, manager)
	policy := jobs.ProductionRuntimePolicy()
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
	registerAllTestHandlers(t, runner, nil)
	if err := runner.Activate(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	var expiredAt *time.Time
	if err := pool.QueryRow(ctx, `SELECT expired_at FROM jobs WHERE job_id = $1`, jobID).Scan(&expiredAt); err != nil {
		t.Fatal(err)
	}
	if expiredAt != nil {
		t.Fatalf("first compaction ran before a full sweep interval: %v", expiredAt)
	}
	deadline := time.Now().Add(2 * time.Second)
	for expiredAt == nil && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		if err := pool.QueryRow(ctx, `SELECT expired_at FROM jobs WHERE job_id = $1`, jobID).Scan(&expiredAt); err != nil {
			t.Fatal(err)
		}
	}
	if expiredAt == nil {
		t.Fatal("expiry supervisor did not compact after the first full interval")
	}
}

func testCompositionForManager(t testing.TB, manager *jobs.Manager) testJobComposition {
	t.Helper()
	value, present := testJobCompositions.Load(manager)
	if !present {
		t.Fatal("test Jobs composition is unavailable")
	}
	return value.(testJobComposition)
}
