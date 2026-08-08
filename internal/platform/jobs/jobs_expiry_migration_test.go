package jobs_test

import (
	"context"
	"strings"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	postgres "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestJobsExpiryMigration60FreshAndGuarded_Integration(t *testing.T) {
	t.Run("fresh shape and ordinary downgrade", func(t *testing.T) {
		harness := pgtest.Start(t)
		db := harness.MigrationDatabaseThroughT(t, "jobs-expiry-fresh", 60)
		ctx := context.Background()
		var columns, indexes int
		if err := db.QueryRowContext(ctx, `
SELECT (SELECT count(*) FROM information_schema.columns
         WHERE table_schema = 'public' AND table_name = 'jobs' AND column_name = 'expired_at'),
       (SELECT count(*) FROM pg_indexes
         WHERE schemaname = 'public' AND tablename = 'jobs' AND indexname = 'jobs_expiry_candidates_idx')
`).Scan(&columns, &indexes); err != nil {
			t.Fatal(err)
		}
		if columns != 1 || indexes != 1 {
			t.Fatalf("migration 60 shape = columns %d indexes %d; want 1/1", columns, indexes)
		}
		if _, err := postgres.RollbackThrough(ctx, db, dbmigrations.Source(), 59); err != nil {
			t.Fatalf("ordinary downgrade before compaction: %v", err)
		}
	})

	t.Run("tombstone blocks downgrade without partial mutation", func(t *testing.T) {
		harness := pgtest.Start(t)
		db := harness.MigrationDatabaseThroughT(t, "jobs-expiry-guarded", 60)
		ctx := context.Background()
		if _, err := db.ExecContext(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ('60000000-0000-4000-8000-000000000001', 'jobs-expiry@example.test', 'Jobs Expiry', 'hash', false, true, true);
INSERT INTO jobs (
    job_id, scope_kind, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, finished_at, retained_until,
    auth_policy, handler_name, job_kind, progress_unit_id, expired_at
) VALUES (
    '60000000-0000-4000-8000-000000000002', 'deployment', 'succeeded', false,
    '60000000-0000-4000-8000-000000000001', now() - interval '8 days',
    now() - interval '7 days', 1, now() - interval '7 days', now() - interval '1 hour',
    'deployment_admin', 'expiry.worker_v1', 'expiry.run_v1',
    'expiry.run.attempt.v1', now()
)
`); err != nil {
			t.Fatal(err)
		}
		_, err := postgres.RollbackThrough(ctx, db, dbmigrations.Source(), 59)
		if err == nil || !strings.Contains(err.Error(), "jobs expiry downgrade blocked") {
			t.Fatalf("guarded downgrade error = %v", err)
		}
		var columns, indexes int
		if err := db.QueryRowContext(ctx, `
SELECT (SELECT count(*) FROM information_schema.columns
         WHERE table_schema = 'public' AND table_name = 'jobs' AND column_name = 'expired_at'),
       (SELECT count(*) FROM pg_indexes
         WHERE schemaname = 'public' AND tablename = 'jobs' AND indexname = 'jobs_expiry_candidates_idx')
`).Scan(&columns, &indexes); err != nil {
			t.Fatal(err)
		}
		if columns != 1 || indexes != 1 {
			t.Fatalf("failed downgrade partially mutated shape: columns=%d indexes=%d", columns, indexes)
		}
	})
}
