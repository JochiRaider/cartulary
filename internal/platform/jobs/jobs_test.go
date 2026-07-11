package jobs_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestManagerCreatesCancelsAndReplaysJobCancel(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID := newJobsHarness(t, "jobs-cancel-replay")

	resource, err := manager.Create(ctx, jobs.CreateParams{
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

func TestManagerTerminalSuccessRetainsJobResource(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
	manager, actorID, _ := newJobsHarnessWithClock(t, "jobs-terminal-success", func() time.Time { return now })

	resource, err := manager.Create(ctx, jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindDeployment},
		SubmittedByUserID: actorID,
		Cancelable:        false,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	jobID := uuid.MustParse(resource.JobID)
	total := 1
	completed, err := manager.CompleteSucceeded(ctx, jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: 1, Total: &total},
		ResultSummary: &jobs.ResultSummary{
			Code:    "phase11_test_completed",
			Message: "Phase 11 test completed.",
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

func TestRunnerDispatchesDurableHandlerAndCompletesJob(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID := newJobsHarness(t, "jobs-durable-dispatch")

	resource, err := manager.Create(ctx, jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
		HandlerName:       "test.complete",
		HandlerPayload:    json.RawMessage(`{"mode":"dispatch"}`),
	})
	if err != nil {
		t.Fatalf("create durable job: %v", err)
	}
	jobID := uuid.MustParse(resource.JobID)

	runner := jobs.NewRunner()
	runner.Configure(manager)
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runner.Close(closeCtx); err != nil {
			t.Fatalf("close runner: %v", err)
		}
	})

	handled := make(chan uuid.UUID, 1)
	if err := runner.RegisterHandler("test.complete", func(ctx context.Context, got uuid.UUID) error {
		if got != jobID {
			return errors.New("handler received unexpected job id")
		}
		rawPayload, err := manager.HandlerPayload(ctx, got)
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
		if _, err := manager.MarkRunning(ctx, got, jobs.Progress{Completed: 0}, nil); err != nil {
			return err
		}
		total := 1
		if _, err := manager.CompleteSucceeded(ctx, jobs.TransitionParams{
			JobID:    got,
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: &jobs.ResultSummary{
				Code:    "durable_handler_completed",
				Message: "Durable handler completed.",
			},
		}); err != nil {
			return err
		}
		handled <- got
		return nil
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	if err := runner.DispatchJobID("test.complete", jobID); err != nil {
		t.Fatalf("dispatch durable job: %v", err)
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

	resource, err := manager.Create(ctx, jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
		HandlerName:       "test.recover",
	})
	if err != nil {
		t.Fatalf("create recoverable durable job: %v", err)
	}
	jobID := uuid.MustParse(resource.JobID)

	runner := jobs.NewRunner()
	runner.Configure(manager)
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runner.Close(closeCtx); err != nil {
			t.Fatalf("close runner: %v", err)
		}
	})

	handled := make(chan uuid.UUID, 1)
	if err := runner.RegisterHandler("test.recover", func(ctx context.Context, got uuid.UUID) error {
		total := 1
		if _, err := manager.CompleteSucceeded(ctx, jobs.TransitionParams{
			JobID:    got,
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: &jobs.ResultSummary{
				Code:    "durable_handler_recovered",
				Message: "Durable handler recovered.",
			},
		}); err != nil {
			return err
		}
		handled <- got
		return nil
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	if err := runner.RecoverHandler(ctx, "test.recover"); err != nil {
		t.Fatalf("recover durable handler jobs: %v", err)
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

func TestManagerFailsDurableHandlerClosedAfterMaxAttempts(t *testing.T) {
	ctx := context.Background()
	manager, actorID, incidentID := newJobsHarness(t, "jobs-durable-exhausted")

	resource, err := manager.Create(ctx, jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: actorID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
		HandlerName:       "test.exhausted",
	})
	if err != nil {
		t.Fatalf("create exhausted durable job: %v", err)
	}
	jobID := uuid.MustParse(resource.JobID)

	for range jobs.DefaultHandlerMaxAttempts {
		claimed, err := manager.ClaimHandlerJob(ctx, jobID, "test.exhausted", "worker-1", time.Minute)
		if err != nil {
			t.Fatalf("claim durable job: %v", err)
		}
		if !claimed {
			t.Fatal("expected durable job claim")
		}
		if err := manager.RecordHandlerError(ctx, jobID, "worker-1", errors.New("handler failed")); err != nil {
			t.Fatalf("record durable handler error: %v", err)
		}
	}

	failed, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatalf("load exhausted job: %v", err)
	}
	if failed.Status != jobs.StatusFailed || failed.Cancelable || failed.ErrorSummary == nil || failed.ErrorSummary.Code != jobs.HandlerAttemptsExhausted {
		t.Fatalf("expected failed-closed durable job, got %#v", failed)
	}
	recoverable, err := manager.RecoverableHandlerJobs(ctx, "test.exhausted", 10)
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

	manager := jobs.NewManager()
	manager.Configure(pool, now)
	return manager, actorID, incidentID
}
