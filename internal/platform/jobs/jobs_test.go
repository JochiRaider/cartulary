package jobs_test

import (
	"context"
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

func newJobsHarness(t testing.TB, prefix string) (*jobs.Manager, uuid.UUID, uuid.UUID) {
	t.Helper()
	return newJobsHarnessWithClock(t, prefix, func() time.Time { return time.Now().UTC() })
}

func newJobsHarnessWithClock(t testing.TB, prefix string, now func() time.Time) (*jobs.Manager, uuid.UUID, uuid.UUID) {
	t.Helper()
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PreparePackageDatabaseT(t, prefix)
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
