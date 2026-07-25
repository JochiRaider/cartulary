package postgres_test

import (
	"context"
	"strings"
	"testing"
	"time"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestIncidentBundleStorageReferenceMigration37FreshSchema_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.MigrationDatabaseT(t, "incident-bundle-storage-reference-fresh", "up-to", "37")
	ctx := context.Background()

	for table, columns := range map[string][]string{
		"incident_bundle_exports":      {"bundle_storage_ref"},
		"incident_bundle_job_payloads": {"bundle_staging_ref"},
	} {
		for _, column := range columns {
			var count int
			if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = $1
   AND column_name = $2
`, table, column).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("%s.%s count = %d want 1", table, column, count)
			}
		}
	}

	for table, oldColumn := range map[string]string{
		"incident_bundle_exports":      "bundle_storage_path",
		"incident_bundle_job_payloads": "bundle_staging_path",
	} {
		var count int
		if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = $1
   AND column_name = $2
`, table, oldColumn).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("legacy column %s.%s remains present", table, oldColumn)
		}
	}

	if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
	_, exportErr := db.ExecContext(ctx, `
INSERT INTO incident_bundle_exports (
    bundle_id, incident_id, export_job_id, exported_by_user_id, exported_at,
    manifest_sha256, reference_pack_mode, bundle_sha256, bundle_byte_size,
    bundle_storage_ref, created_at
) VALUES (
    '10000000-0000-4000-8000-000000000001',
    '10000000-0000-4000-8000-000000000002',
    '10000000-0000-4000-8000-000000000003',
    '10000000-0000-4000-8000-000000000004',
    $1, repeat('a', 64), 'refs_only', repeat('b', 64), 0, '/host/absolute.zip', $1
)
`, now)
	if exportErr == nil || !strings.Contains(exportErr.Error(), "incident_bundle_exports_storage_ref_ck") {
		t.Fatalf("absolute export reference must fail lexical check, got %v", exportErr)
	}
	_, stagingErr := db.ExecContext(ctx, `
INSERT INTO incident_bundle_job_payloads (
    job_id, job_kind, actor_user_id, incident_id, bundle_staging_ref,
    request_json, created_at, updated_at
) VALUES (
    '20000000-0000-4000-8000-000000000001',
    'export',
    '20000000-0000-4000-8000-000000000002',
    '20000000-0000-4000-8000-000000000003',
    'incident-bundles/../escape.bundle',
    '{}'::jsonb, $1, $1
)
`, now)
	if stagingErr == nil || !strings.Contains(stagingErr.Error(), "incident_bundle_job_payloads_staging_ref_ck") {
		t.Fatalf("traversing staging reference must fail lexical check, got %v", stagingErr)
	}
}

func TestIncidentBundleStorageReferenceMigration37RejectsPopulatedTablesBeforeMutation_Integration(t *testing.T) {
	cases := []struct {
		name   string
		insert string
	}{
		{
			name: "export",
			insert: `
INSERT INTO incident_bundle_exports (
    bundle_id, incident_id, export_job_id, exported_by_user_id, exported_at,
    manifest_sha256, reference_pack_mode, bundle_sha256, bundle_byte_size,
    bundle_storage_path, created_at
) VALUES (
    '30000000-0000-4000-8000-000000000001',
    '30000000-0000-4000-8000-000000000002',
    '30000000-0000-4000-8000-000000000003',
    '30000000-0000-4000-8000-000000000004',
    $1, repeat('a', 64), 'refs_only', repeat('b', 64), 0, '/host/export.zip', $1
)`,
		},
		{
			name: "job_payload",
			insert: `
INSERT INTO incident_bundle_job_payloads (
    job_id, job_kind, actor_user_id, incident_id, bundle_staging_path,
    request_json, created_at, updated_at
) VALUES (
    '40000000-0000-4000-8000-000000000001',
    'export',
    '40000000-0000-4000-8000-000000000002',
    '40000000-0000-4000-8000-000000000003',
    '/host/staged.bundle',
    '{}'::jsonb, $1, $1
)`,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			harness := pgtest.Start(t)
			db := harness.MigrationDatabaseT(t, "incident-bundle-storage-reference-reject", "up-to", "36")
			ctx := context.Background()
			if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, test.insert, time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, `SET session_replication_role = origin`); err != nil {
				t.Fatal(err)
			}

			_, err := postgres.Migrate(ctx, db, dbmigrations.Source(), "up-to", "37")
			if err == nil {
				t.Fatal("expected pre-release storage-reference cutover rejection")
			}
			message := err.Error()
			if !strings.Contains(message, "development database reset required") ||
				!strings.Contains(message, "CARTULARY_DESTRUCTIVE_CONFIRM=db-reset make db-reset") ||
				!strings.Contains(message, "reseed development data") {
				t.Fatalf("unexpected cutover diagnostic: %v", err)
			}

			var oldColumns int
			if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND (
       (table_name = 'incident_bundle_exports' AND column_name = 'bundle_storage_path')
       OR
       (table_name = 'incident_bundle_job_payloads' AND column_name = 'bundle_staging_path')
   )
`).Scan(&oldColumns); err != nil {
				t.Fatal(err)
			}
			if oldColumns != 2 {
				t.Fatalf("migration mutated schema before rejection: legacy column count %d want 2", oldColumns)
			}
		})
	}
}
