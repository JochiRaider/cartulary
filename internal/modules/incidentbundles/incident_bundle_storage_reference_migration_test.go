package incidentbundles_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestIncidentBundleStorageReferenceHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "incident-bundle-storage-reference-head", postgres.PurposeRecovery)
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
