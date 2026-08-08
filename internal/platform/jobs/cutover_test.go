package jobs_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestDrainedV2CutoverAndRecovery_Integration(t *testing.T) {
	ctx := context.Background()
	harness := pgtest.Start(t)
	testDB := harness.NewMigrationDatabaseT(t, "jobs-drained-v2-cutover")
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := postgres.Migrate(ctx, db, dbmigrations.Source(), "up-to", "57"); err != nil {
		t.Fatal(err)
	}

	actorID := uuid.MustParse("58000000-0000-4000-8000-000000000081")
	jobID := uuid.MustParse("58000000-0000-4000-8000-000000000082")
	if _, err := db.ExecContext(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, 'jobs-cutover@example.test', 'Jobs Cutover', 'hash', false, true, true)
`, actorID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO jobs (
    job_id, scope_kind, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, progress_total, auth_policy,
    handler_name, extension_owner_profile_id, extension_job_kind,
    extension_idempotency_identity, extension_idempotency_route_key,
    extension_idempotency_scope_key, extension_normalized_request_sha256
) VALUES (
    $1, 'deployment', 'queued', true, $2,
    now(), now(), 0, 1, 'deployment_admin',
    'import.discovery.worker_v1', 'import', 'import.discovery_v1',
    '{"schema_id":"cartulary.route_scoped_idempotency_identity.v1"}',
    'imports.discovery', 'deployment', repeat('a', 64)
)
`, jobID, actorID); err != nil {
		t.Fatal(err)
	}

	var activeLeaseCount, illegalReplayCount int
	if err := db.QueryRowContext(ctx, `
SELECT (SELECT count(*)
          FROM jobs
         WHERE handler_lease_owner IS NOT NULL
           AND handler_lease_expires_at > now()),
       (SELECT count(*)
          FROM collaboration_event_intents
         WHERE event_family = 'job_progress'
           AND created_at >= now() - interval '24 hours')
`).Scan(&activeLeaseCount, &illegalReplayCount); err != nil {
		t.Fatal(err)
	}
	if activeLeaseCount != 0 || illegalReplayCount != 0 {
		t.Fatalf("cutover was not drained: active_leases=%d replay_candidates=%d", activeLeaseCount, illegalReplayCount)
	}
	if _, err := postgres.Migrate(ctx, db, dbmigrations.Source(), "up-to", "58"); err != nil {
		t.Fatal(err)
	}

	pool, err := pgxpool.New(ctx, testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	definition := jobs.ExtensionJobContract{
		OwnerProfileID: "import",
		JobKind:        "import.discovery_v1",
		ProgressUnitID: "import.discovery.session.v1",
		OperationKind:  "import.discovery",
		WorkerKind:     "import.discovery.worker_v1",
		ContractSHA256: strings.Repeat("a", 64),
		ProofRequired:  true,
		MaxProofBytes:  4096,
	}
	transactions := collaborationsupport.NewJobTransactionsWithDefinitions(definition)
	manager := jobs.NewManager()
	manager.Configure(pool, transactions, func() time.Time { return time.Now().UTC() })
	if err := manager.ConfigureExtensionContracts([]jobs.ExtensionJobContract{definition}); err != nil {
		t.Fatal(err)
	}
	if err := manager.ValidateStorageCatalog(ctx); err != nil {
		t.Fatalf("corrected writer startup validation: %v", err)
	}

	runner := jobs.NewRunner()
	runner.Configure(manager)
	gate := &dequeueGate{}
	runner.ConfigureDequeueGate(gate)
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runner.Close(closeCtx); err != nil {
			t.Fatalf("close cutover runner: %v", err)
		}
	})
	handled := make(chan error, 1)
	if err := runner.RegisterHandler(definition.WorkerKind, func(handlerCtx context.Context, got uuid.UUID) error {
		resource, loadErr := manager.Get(handlerCtx, got)
		if loadErr != nil {
			handled <- loadErr
			return loadErr
		}
		if resource.Status != jobs.StatusRunning || resource.StartedAt == nil {
			handlerErr := jobs.ErrInvalidTransition
			handled <- handlerErr
			return handlerErr
		}
		total := 1
		_, completeErr := manager.CompleteSucceeded(handlerCtx, jobs.TransitionParams{
			JobID:    got,
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: &jobs.ResultSummary{
				Code:    "cutover_recovered",
				Message: "Recovered after the v2 cutover.",
			},
		})
		handled <- completeErr
		return completeErr
	}); err != nil {
		t.Fatal(err)
	}
	if err := runner.RecoverHandler(ctx, definition.WorkerKind); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-handled:
		t.Fatalf("handler ran before dequeue admission reopened: %v", err)
	case <-time.After(25 * time.Millisecond):
	}
	gate.open.Store(true)
	if err := runner.Activate(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-handled:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for corrected recovery")
	}
	resource, err := manager.Get(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	if resource.Status != jobs.StatusSucceeded || resource.ResultSummary == nil ||
		resource.ResultSummary.Code != "cutover_recovered" {
		t.Fatalf("recovered cutover job = %#v", resource)
	}
	var storedKind, storedUnit string
	if err := pool.QueryRow(ctx, `SELECT job_kind, progress_unit_id FROM jobs WHERE job_id = $1`, jobID).Scan(&storedKind, &storedUnit); err != nil {
		t.Fatal(err)
	}
	if storedKind != definition.JobKind || storedUnit != definition.ProgressUnitID {
		t.Fatalf("cutover definition = %s/%s", storedKind, storedUnit)
	}
}
