package reference_data_test

import (
	"context"
	"strings"
	"testing"
	"time"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	postgres "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestReferencePackStorageReferenceMigration38FreshSchema_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.MigrationDatabaseThroughT(t, "reference-pack-storage-reference-fresh", 38)
	ctx := context.Background()

	for table, column := range map[string]string{
		"reference_packs":             "bundle_storage_ref",
		"reference_pack_job_payloads": "bundle_staging_ref",
	} {
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

	for table, oldColumn := range map[string]string{
		"reference_packs":             "bundle_storage_path",
		"reference_pack_job_payloads": "bundle_staging_path",
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
	_, packErr := db.ExecContext(ctx, `
INSERT INTO reference_packs (
    pack_key, version, pack_kind, manifest_sha256, payload_sha256,
    pack_contract_version, verification_method, status, imported_at,
    verification_result, bundle_sha256, bundle_storage_ref, metadata
) VALUES (
    'type_registry.invalid', '1', 'type_registry', repeat('a', 64), repeat('b', 64),
    'reference_pack.v1', 'manifest_sha256_v1', 'available', $1,
    'passed', repeat('c', 64), '/host/reference.bundle', '{}'::jsonb
)
`, now)
	if packErr == nil || !strings.Contains(packErr.Error(), "reference_packs_bundle_storage_ref_relative_check") {
		t.Fatalf("absolute pack reference must fail lexical check, got %v", packErr)
	}
	_, stagingErr := db.ExecContext(ctx, `
INSERT INTO reference_pack_job_payloads (
    job_id, job_kind, actor_user_id, resolved_pack_keys, bundle_staging_ref,
    request_json, created_at
) VALUES (
    '50000000-0000-4000-8000-000000000001',
    'import',
    '50000000-0000-4000-8000-000000000002',
    '{}'::text[],
    'reference-packs/../escape.bundle',
    '{}'::jsonb, $1
)
`, now)
	if stagingErr == nil || !strings.Contains(stagingErr.Error(), "reference_pack_job_payloads_bundle_staging_ref_relative_check") {
		t.Fatalf("traversing staging reference must fail lexical check, got %v", stagingErr)
	}
}

func TestReferencePackStorageReferenceMigration38RejectsPopulatedTablesBeforeMutation_Integration(t *testing.T) {
	cases := []struct {
		name   string
		insert string
	}{
		{
			name: "pack",
			insert: `
INSERT INTO reference_packs (
    pack_key, version, pack_kind, manifest_sha256, payload_sha256,
    pack_contract_version, verification_method, status, imported_at,
    verification_result, bundle_sha256, bundle_storage_path, metadata
) VALUES (
    'type_registry.legacy', '1', 'type_registry', repeat('a', 64), repeat('b', 64),
    'reference_pack.v1', 'manifest_sha256_v1', 'available', $1,
    'passed', repeat('c', 64), '/host/reference.bundle', '{}'::jsonb
)`,
		},
		{
			name: "job_payload",
			insert: `
INSERT INTO reference_pack_job_payloads (
    job_id, job_kind, actor_user_id, resolved_pack_keys, bundle_staging_path,
    request_json, created_at
) VALUES (
    '60000000-0000-4000-8000-000000000001',
    'import',
    '60000000-0000-4000-8000-000000000002',
    '{}'::text[],
    '/host/staged.bundle',
    '{}'::jsonb, $1
)`,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			harness := pgtest.Start(t)
			db := harness.MigrationDatabaseThroughT(t, "reference-pack-storage-reference-reject", 37)
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

			_, err := postgres.ApplyThrough(ctx, db, dbmigrations.Source(), 38)
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
       (table_name = 'reference_packs' AND column_name = 'bundle_storage_path')
       OR
       (table_name = 'reference_pack_job_payloads' AND column_name = 'bundle_staging_path')
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
