package extensions_test

import (
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestExtensionJobHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.BeginRollbackDBT(t, "extension-job-head-contract")
	if _, err := db.Exec(context.Background(), `SET LOCAL ROLE cartulary_runtime`); err != nil {
		t.Fatal(err)
	}
	var metadataColumns int
	if err := db.QueryRow(context.Background(), `
SELECT count(*)
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = 'jobs'
   AND column_name IN (
       'extension_owner_profile_id',
	       'job_kind',
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
