package extensions_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestExtensionJobCutoverMigration34FreshSchema_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, 34)
	db := migrationDB.SQL()
	var metadataColumns int
	if err := db.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = 'jobs'
   AND column_name IN (
       'extension_owner_profile_id',
       'extension_job_kind',
       'extension_idempotency_identity',
       'extension_idempotency_route_key',
       'extension_idempotency_scope_key',
       'extension_normalized_request_sha256'
   )
`).Scan(&metadataColumns); err != nil {
		t.Fatal(err)
	}
	if metadataColumns != 6 {
		t.Fatalf("extension job metadata columns = %d want 6", metadataColumns)
	}
}

func TestExtensionJobCutoverMigration34RejectsEveryRetiredHandlerBeforeMutation_Integration(t *testing.T) {
	retiredHandlers := []string{
		"imports.discovery",
		"imports.apply",
		"incident_bundles.execute",
		"reference_data.execute",
		"reporting.execute",
	}
	for _, handlerName := range retiredHandlers {
		t.Run(handlerName, func(t *testing.T) {
			harness := pgtest.Start(t)
			migrationDB := harness.MigrationDatabaseThroughT(t, 33)
			db := migrationDB.SQL()
			ctx := context.Background()
			if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, `
INSERT INTO jobs (
    scope_kind, incident_id, status, cancelable, submitted_by_user_id,
    submitted_at, updated_at, progress_completed, progress_total,
    auth_policy, handler_name
) VALUES (
    'deployment', NULL, 'queued', true,
    '00000000-0000-4000-8000-000000000001',
    $1, $1, 0, 1, 'deployment_admin', $2
)
`, time.Date(2026, 7, 24, 20, 0, 0, 0, time.UTC), handlerName); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, `SET session_replication_role = origin`); err != nil {
				t.Fatal(err)
			}
			err := migrationDB.ApplyThrough(ctx, 34)
			if err == nil {
				t.Fatal("expected clean-cutover migration rejection")
			}
			message := err.Error()
			if !strings.Contains(message, "extension profile job cutover requires database reset/reseed") {
				t.Fatalf("unexpected cutover diagnostic: %v", err)
			}
			var coordinationTable *string
			if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.extension_state_metadata')::text`).Scan(&coordinationTable); err != nil {
				t.Fatal(err)
			}
			if coordinationTable != nil {
				t.Fatalf("migration mutated extension coordination before rejection: %q", *coordinationTable)
			}
			var metadataColumns int
			if err := db.QueryRowContext(ctx, `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = 'jobs'
   AND column_name LIKE 'extension_%'
`).Scan(&metadataColumns); err != nil {
				t.Fatal(err)
			}
			if metadataColumns != 0 {
				t.Fatalf("migration added %d extension job columns before rejection", metadataColumns)
			}
		})
	}
}
