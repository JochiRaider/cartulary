package jobs_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestDrainedV2CutoverAndRecovery_Integration(t *testing.T) {
	ctx := context.Background()
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, 57)
	db := migrationDB.SQL()
	migrationDatabaseName := jobsMigrationScratchDatabaseName(t, ctx, db)

	actorID := uuid.MustParse("58000000-0000-4000-8000-000000000081")
	jobID := uuid.MustParse("58000000-0000-4000-8000-000000000082")
	expiredJobID := uuid.MustParse("58000000-0000-4000-8000-000000000083")
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
	if _, err := db.ExecContext(ctx, `
INSERT INTO jobs (
    job_id, scope_kind, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, progress_total, auth_policy,
    handler_name, extension_owner_profile_id, extension_job_kind,
    extension_idempotency_identity, extension_idempotency_route_key,
    extension_idempotency_scope_key, extension_normalized_request_sha256,
    finished_at, retained_until, result_summary_json
) VALUES (
    $1, 'deployment', 'succeeded', false, $2,
    now() - interval '8 days', now() - interval '7 days', 1, 1, 'deployment_admin',
    'import.discovery.worker_v1', 'import', 'import.discovery_v1',
    '{"schema_id":"cartulary.route_scoped_idempotency_identity.v1"}',
    'imports.discovery.expired', 'deployment', repeat('b', 64),
    now() - interval '7 days', now() - interval '1 second',
    '{"code":"pre_cutover_complete","message":"Completed before cutover."}'
)
`, expiredJobID, actorID); err != nil {
		t.Fatal(err)
	}

	poolConfig, err := pgxpool.ParseConfig(harness.AdminDSN())
	if err != nil {
		t.Fatal(err)
	}
	poolConfig.ConnConfig.Database = migrationDatabaseName
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	oldWriterLease, err := processlease.AcquireApplicationProcess(ctx, pool, 5*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire old-writer process lease: %v", err)
	}
	t.Cleanup(oldWriterLease.Close)
	if _, err := processlease.AcquireApplicationProcess(ctx, pool, 20*time.Millisecond, 100*time.Millisecond); !errors.Is(err, processlease.ErrApplicationProcessActive) {
		t.Fatalf("second old writer was not excluded: %v", err)
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
	backupPath := createJobsCutoverBackup(t, ctx, harness.AdminDSN(), migrationDatabaseName)
	if err := oldWriterLease.Release(ctx); err != nil {
		t.Fatalf("stop old writer under process lease: %v", err)
	}
	if err := migrationDB.ApplyThrough(ctx, 60); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(backupPath); err != nil || info.Size() == 0 {
		t.Fatalf("pre-cutover backup is unavailable: info=%v err=%v", info, err)
	}
	correctedWriterLease, err := processlease.AcquireApplicationProcess(ctx, pool, 5*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire corrected-writer process lease: %v", err)
	}
	defer correctedWriterLease.Close()
	if _, err := processlease.AcquireApplicationProcess(ctx, pool, 20*time.Millisecond, 100*time.Millisecond); !errors.Is(err, processlease.ErrApplicationProcessActive) {
		t.Fatalf("second corrected writer was not excluded: %v", err)
	}
	definition := jobs.Definition{
		JobKind:        "import.discovery_v1",
		ProgressUnitID: "import.discovery.session.v1",
		HandlerName:    "import.discovery.worker_v1",
		Extension: &jobs.ExtensionPolicy{
			OwnerProfileID: "import", OperationKind: "import.discovery",
			ContractSHA256: strings.Repeat("a", 64), ProofRequired: true, MaxProofBytes: 4096,
		},
	}
	catalog, err := jobs.NewCatalog([]jobs.Definition{definition})
	if err != nil {
		t.Fatal(err)
	}
	transactions := collaborationsupport.NewJobTransactionsForCatalog(catalog)
	policy := jobs.ProductionRuntimePolicy()
	manager, err := jobs.NewManager(jobs.ManagerOptions{
		Postgres: pool, Transactions: transactions, Catalog: catalog, Policy: policy,
		Now: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.ValidateStorageCatalog(ctx); err != nil {
		t.Fatalf("corrected writer startup validation: %v", err)
	}

	gate := &dequeueGate{}
	runner, err := jobs.NewRunner(jobs.RunnerOptions{
		Manager: manager, Catalog: catalog, Policy: policy, DequeueGate: gate,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := runner.Close(closeCtx); err != nil {
			t.Fatalf("close cutover runner: %v", err)
		}
	})
	handled := make(chan error, 1)
	if err := runner.RegisterHandler(definition.HandlerName, func(handlerCtx context.Context, execution jobs.Execution) error {
		got := execution.JobID()
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
		_, completeErr := manager.CompleteSucceeded(handlerCtx, execution, jobs.SuccessCompletion{
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: jobs.ResultSummary{
				Code:    "cutover_recovered",
				Message: "Recovered after the v2 cutover.",
			},
		})
		handled <- completeErr
		return completeErr
	}); err != nil {
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
	var expiredAt *time.Time
	if err := pool.QueryRow(ctx, `SELECT expired_at FROM jobs WHERE job_id = $1`, expiredJobID).Scan(&expiredAt); err != nil {
		t.Fatal(err)
	}
	if expiredAt != nil {
		t.Fatal("first compaction ran before the full production sweep interval")
	}
	if err := runner.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if err := correctedWriterLease.Release(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestDrainedJobsCutoverRollbackBeforeFirstCompaction_Integration(t *testing.T) {
	ctx := context.Background()
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, 58)
	db := migrationDB.SQL()
	actorID := uuid.MustParse("58000000-0000-4000-8000-000000000091")
	jobID := uuid.MustParse("58000000-0000-4000-8000-000000000092")
	if _, err := db.ExecContext(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, 'jobs-rollback@example.test', 'Jobs Rollback', 'hash', false, true, true)
`, actorID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO jobs (
    job_id, scope_kind, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, auth_policy,
    handler_name, job_kind, progress_unit_id
) VALUES (
    $2, 'deployment', 'queued', true, $1, now(), now(), 0,
    'deployment_admin', 'rollback.worker_v1', 'rollback.run_v1', 'rollback.run.operation.v1'
)
`, actorID, jobID); err != nil {
		t.Fatal(err)
	}
	var beforeKind, beforeUnit, beforeHandler string
	if err := db.QueryRowContext(ctx, `SELECT job_kind, progress_unit_id, handler_name FROM jobs WHERE job_id = $1`, jobID).Scan(&beforeKind, &beforeUnit, &beforeHandler); err != nil {
		t.Fatal(err)
	}
	if err := migrationDB.ApplyThrough(ctx, 60); err != nil {
		t.Fatal(err)
	}
	if err := migrationDB.RollbackThrough(ctx, 58); err != nil {
		t.Fatalf("guarded rollback before corrected writes or compaction: %v", err)
	}
	var afterKind, afterUnit, afterHandler string
	if err := db.QueryRowContext(ctx, `SELECT job_kind, progress_unit_id, handler_name FROM jobs WHERE job_id = $1`, jobID).Scan(&afterKind, &afterUnit, &afterHandler); err != nil {
		t.Fatal(err)
	}
	if beforeKind != afterKind || beforeUnit != afterUnit || beforeHandler != afterHandler {
		t.Fatalf("rollback changed retained identity: before=%s/%s/%s after=%s/%s/%s", beforeKind, beforeUnit, beforeHandler, afterKind, afterUnit, afterHandler)
	}
	var oldColumns, correctedColumns int
	if err := db.QueryRowContext(ctx, `
SELECT count(*) FILTER (WHERE column_name IN ('handler_attempts', 'handler_max_attempts', 'handler_lease_owner')),
       count(*) FILTER (WHERE column_name IN ('handler_attempt_id', 'handler_failure_count', 'handler_next_attempt_at', 'expired_at'))
  FROM information_schema.columns
 WHERE table_schema = 'public' AND table_name = 'jobs'
`).Scan(&oldColumns, &correctedColumns); err != nil {
		t.Fatal(err)
	}
	if oldColumns != 3 || correctedColumns != 0 {
		t.Fatalf("rollback shape = old %d corrected %d; want 3/0", oldColumns, correctedColumns)
	}
}

func jobsMigrationScratchDatabaseName(t testing.TB, ctx context.Context, db *sql.DB) string {
	t.Helper()
	var databaseName string
	if err := db.QueryRowContext(ctx, `SELECT current_database()`).Scan(&databaseName); err != nil {
		t.Fatalf("resolve migration scratch database name: %v", err)
	}
	return databaseName
}

func createJobsCutoverBackup(t testing.TB, ctx context.Context, adminDSN string, databaseName string) string {
	t.Helper()
	config, err := pgx.ParseConfig(adminDSN)
	if err != nil {
		t.Fatalf("parse cutover backup database binding: %v", err)
	}
	config.Database = databaseName
	backupPath := filepath.Join(t.TempDir(), "jobs-pre-cutover.dump")
	command := exec.CommandContext(ctx, "pg_dump", "--format=custom", "--no-owner", "--file", backupPath)
	command.Env = append(os.Environ(),
		"PGHOST="+config.Host,
		"PGPORT="+strconv.FormatUint(uint64(config.Port), 10),
		"PGUSER="+config.User,
		"PGPASSWORD="+config.Password,
		"PGDATABASE="+config.Database,
		"PGSSLMODE=disable",
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("create pre-cutover backup: %v: %s", err, strings.TrimSpace(string(output)))
	}
	return backupPath
}
