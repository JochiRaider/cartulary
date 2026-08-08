package jobs_test

import (
	"context"
	"strings"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	postgres "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestJobsExecutionMigration59FreshAndGuarded_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.MigrationDatabaseThroughT(t, "jobs-execution-fresh", 59)
	var columns int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = 'jobs'
   AND column_name IN ('handler_attempt_id', 'handler_failure_count', 'handler_next_attempt_at')
`).Scan(&columns); err != nil {
		t.Fatal(err)
	}
	if columns != 3 {
		t.Fatalf("migration 59 execution columns = %d want 3", columns)
	}
}

func TestJobsExecutionMigration59RejectsUnsafeIdentityBeforeMutation_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.MigrationDatabaseThroughT(t, "jobs-execution-reject", 58)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO jobs (
    job_id, scope_kind, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, auth_policy,
    result_summary_json, finished_at, retained_until
) VALUES (
    '59000000-0000-4000-8000-000000000099', 'deployment', 'succeeded', false,
    '58000000-0000-4000-8000-000000000001', now(), now(), 0,
    'deployment_admin', '{"code":"legacy","message":"Legacy."}', now(), now() + interval '7 days'
)
`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `SET session_replication_role = origin`); err != nil {
		t.Fatal(err)
	}
	_, err := postgres.ApplyThrough(ctx, db, dbmigrations.Source(), 59)
	if err == nil || !strings.Contains(err.Error(), "jobs execution preflight failed") ||
		strings.Contains(err.Error(), "59000000-0000-4000-8000-000000000099") {
		t.Fatalf("migration 59 preflight error = %v", err)
	}
	var oldColumns, newColumns int
	if err := db.QueryRowContext(ctx, `
SELECT count(*) FILTER (WHERE column_name IN ('handler_attempts', 'handler_max_attempts', 'handler_lease_owner')),
       count(*) FILTER (WHERE column_name IN ('handler_attempt_id', 'handler_failure_count', 'handler_next_attempt_at'))
  FROM information_schema.columns
 WHERE table_schema = 'public' AND table_name = 'jobs'
`).Scan(&oldColumns, &newColumns); err != nil {
		t.Fatal(err)
	}
	if oldColumns != 3 || newColumns != 0 {
		t.Fatalf("failed preflight partially mutated jobs shape: old=%d new=%d", oldColumns, newColumns)
	}
}
