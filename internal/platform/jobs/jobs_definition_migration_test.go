package jobs_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	postgres "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

var jobsV2Bindings = map[string]string{
	"import.discovery_v1":                       "import.discovery.session.v1",
	"import.apply_v1":                           "import.apply.import_unit.v1",
	"incident_portability.export_v1":            "incident_portability.export.request.v1",
	"incident_portability.import_v1":            "incident_portability.import.request.v1",
	"reference_pack.import_v1":                  "reference_pack.import.request.v1",
	"reference_pack.refresh_v1":                 "reference_pack.refresh.pack_key.v1",
	"reference_pack.reverify_v1":                "reference_pack.reverify.pack_version.v1",
	"snapshot_reporting.composition_preview_v1": "snapshot_reporting.composition_preview.render_attempt.v1",
	"snapshot_reporting.release_create_v1":      "snapshot_reporting.release_create.render_attempt.v1",
	"snapshot_reporting.snapshot_create_v1":     "snapshot_reporting.snapshot_create.materialization.v1",
}

func TestJobsDefinitionMigration58FreshAndExactUpgrade_Integration(t *testing.T) {
	t.Run("fresh", func(t *testing.T) {
		harness := pgtest.Start(t)
		db := harness.MigrationDatabaseThroughT(t, "jobs-definition-fresh", 58)
		assertJobsV2Shape(t, db)
	})

	t.Run("exact backfill and legacy terminal", func(t *testing.T) {
		harness := pgtest.Start(t)
		db := harness.MigrationDatabaseThroughT(t, "jobs-definition-upgrade", 57)
		ctx := context.Background()
		if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
			t.Fatal(err)
		}
		ordinal := 1
		for jobKind := range jobsV2Bindings {
			insertJobsV1DefinitionRow(t, db, jobKind, ownerForJobsV2Kind(jobKind), "queued", ordinal, nil)
			ordinal++
		}
		if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, `
INSERT INTO jobs (
    job_id, scope_kind, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, auth_policy,
    result_summary_json, finished_at, retained_until
) VALUES (
    '58000000-0000-4000-8000-000000000099', 'deployment', 'succeeded', false,
    '58000000-0000-4000-8000-000000000001', now(), now(), 0,
    'deployment_admin', '{"code":"legacy","message":"Legacy."}', now(), now() + interval '7 days'
)
`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, `SET session_replication_role = origin`); err != nil {
			t.Fatal(err)
		}
		if _, err := postgres.ApplyThrough(ctx, db, dbmigrations.Source(), 58); err != nil {
			t.Fatal(err)
		}
		assertJobsV2Shape(t, db)
		for kind, unit := range jobsV2Bindings {
			var got string
			if err := db.QueryRowContext(ctx, `SELECT progress_unit_id FROM jobs WHERE job_kind = $1`, kind).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != unit {
				t.Errorf("%s progress unit = %q want %q", kind, got, unit)
			}
		}
		var legacyKind, legacyUnit *string
		if err := db.QueryRowContext(ctx, `
SELECT job_kind, progress_unit_id FROM jobs WHERE job_id = '58000000-0000-4000-8000-000000000099'
`).Scan(&legacyKind, &legacyUnit); err != nil {
			t.Fatal(err)
		}
		if legacyKind != nil || legacyUnit != nil {
			t.Fatalf("legacy terminal identity changed: kind=%v unit=%v", legacyKind, legacyUnit)
		}
	})
}

func TestJobsDefinitionMigration58RejectsUnsafeCorporaBeforeMutation_Integration(t *testing.T) {
	tests := map[string]struct {
		seed       func(*testing.T, *sql.DB)
		wantMetric string
		secret     string
	}{
		"unknown mutable mapping": {
			seed: func(t *testing.T, db *sql.DB) {
				insertJobsV1DefinitionRow(t, db, "unknown.private_v1", "unknown_owner", "queued", 21, nil)
			},
			wantMetric: "unknown_mapping_count=1",
			secret:     "58000000-0000-4000-8000-000000000021",
		},
		"active lease": {
			seed: func(t *testing.T, db *sql.DB) {
				expires := time.Now().UTC().Add(time.Hour)
				insertJobsV1DefinitionRow(t, db, "import.discovery_v1", "import", "queued", 22, &expires)
			},
			wantMetric: "active_lease_count=1",
			secret:     "58000000-0000-4000-8000-000000000022",
		},
		"invalid progress": {
			seed: func(t *testing.T, db *sql.DB) {
				ctx := context.Background()
				if _, err := db.ExecContext(ctx, `ALTER TABLE jobs DROP CONSTRAINT jobs_progress_total_ck`); err != nil {
					t.Fatal(err)
				}
				insertJobsV1DefinitionRow(t, db, "import.discovery_v1", "import", "queued", 23, nil)
				if _, err := db.ExecContext(ctx, `UPDATE jobs SET progress_completed = 2, progress_total = 1`); err != nil {
					t.Fatal(err)
				}
			},
			wantMetric: "invalid_progress_count=1",
			secret:     "58000000-0000-4000-8000-000000000023",
		},
		"illegal replay transition": {
			seed: func(t *testing.T, db *sql.DB) {
				ctx := context.Background()
				if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
					t.Fatal(err)
				}
				if _, err := db.ExecContext(ctx, `
INSERT INTO collaboration_event_intents (
    intent_key, incident_id, event_family, canonical_payload,
    source_identity, mutation_ordinal, next_attempt_at, created_at, updated_at
) VALUES
    ('jobs-v2-replay-queued', '58000000-0000-4000-8000-000000000071', 'job_progress',
     '{"status":"queued"}', 'job:58000000-0000-4000-8000-000000000024', 0,
     now() - interval '1 minute', now() - interval '1 minute', now() - interval '1 minute'),
    ('jobs-v2-replay-failed', '58000000-0000-4000-8000-000000000071', 'job_progress',
     '{"status":"failed"}', 'job:58000000-0000-4000-8000-000000000024', 0,
     now(), now(), now())
`); err != nil {
					t.Fatal(err)
				}
				if _, err := db.ExecContext(ctx, `SET session_replication_role = origin`); err != nil {
					t.Fatal(err)
				}
			},
			wantMetric: "illegal_replay_count=1",
			secret:     "58000000-0000-4000-8000-000000000024",
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			harness := pgtest.Start(t)
			db := harness.MigrationDatabaseThroughT(t, "jobs-definition-reject", 57)
			testCase.seed(t, db)
			_, err := postgres.ApplyThrough(context.Background(), db, dbmigrations.Source(), 58)
			if err == nil || !strings.Contains(err.Error(), "jobs v2 compatibility preflight failed") ||
				!strings.Contains(err.Error(), testCase.wantMetric) {
				t.Fatalf("preflight error = %v; want %s", err, testCase.wantMetric)
			}
			if strings.Contains(err.Error(), testCase.secret) {
				t.Fatalf("preflight disclosed job identity: %v", err)
			}
			assertJobsV1ShapeUnchanged(t, db)
		})
	}
}

func insertJobsV1DefinitionRow(t testing.TB, db *sql.DB, jobKind string, owner string, status string, ordinal int, leaseExpires *time.Time) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.ExecContext(ctx, `SET session_replication_role = origin`); err != nil {
			t.Fatal(err)
		}
	}()
	jobID := fmt.Sprintf("58000000-0000-4000-8000-%012d", ordinal)
	_, err := db.ExecContext(ctx, `
INSERT INTO jobs (
    job_id, scope_kind, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, progress_total, auth_policy,
    handler_name, handler_lease_owner, handler_lease_expires_at,
    extension_owner_profile_id, extension_job_kind,
    extension_idempotency_identity, extension_idempotency_route_key,
    extension_idempotency_scope_key, extension_normalized_request_sha256
) VALUES (
    $1, 'deployment', $2, true,
    '58000000-0000-4000-8000-000000000001', now(), now(), 0, 1,
    'deployment_admin', 'test.worker_v1',
    CASE WHEN $5::timestamptz IS NULL THEN NULL ELSE 'attempt-safe' END, $5,
    $3, $4, '{"schema_id":"test"}', 'test.route', 'deployment:test',
    repeat('a', 64)
)
`, jobID, status, owner, jobKind, leaseExpires)
	if err != nil {
		t.Fatal(err)
	}
}

func ownerForJobsV2Kind(kind string) string {
	return strings.SplitN(kind, ".", 2)[0]
}

func assertJobsV2Shape(t testing.TB, db *sql.DB) {
	t.Helper()
	var newColumns int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*) FROM information_schema.columns
 WHERE table_schema = 'public' AND table_name = 'jobs'
   AND column_name IN ('job_kind', 'progress_unit_id')
`).Scan(&newColumns); err != nil {
		t.Fatal(err)
	}
	if newColumns != 2 {
		t.Fatalf("jobs v2 columns = %d want 2", newColumns)
	}
}

func assertJobsV1ShapeUnchanged(t testing.TB, db *sql.DB) {
	t.Helper()
	var oldColumns, newColumns int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*) FILTER (WHERE column_name = 'extension_job_kind'),
       count(*) FILTER (WHERE column_name IN ('job_kind', 'progress_unit_id'))
  FROM information_schema.columns
 WHERE table_schema = 'public' AND table_name = 'jobs'
`).Scan(&oldColumns, &newColumns); err != nil {
		t.Fatal(err)
	}
	if oldColumns != 1 || newColumns != 0 {
		t.Fatalf("failed preflight partially mutated jobs shape: old=%d new=%d", oldColumns, newColumns)
	}
}
