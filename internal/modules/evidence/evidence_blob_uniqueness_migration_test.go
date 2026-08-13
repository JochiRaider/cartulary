package evidence_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestEvidenceBlobAssociationHeadSchemaContract_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "evidence-blob-association-head", postgres.PurposeRecovery)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
		t.Fatal(err)
	}
	const (
		blobID       = "53000000-0000-4000-8000-000000000001"
		firstRecord  = "53000000-0000-4000-8000-000000000011"
		secondRecord = "53000000-0000-4000-8000-000000000012"
		incidentID   = "53000000-0000-4000-8000-000000000021"
		actorID      = "53000000-0000-4000-8000-000000000031"
	)
	if _, err := db.ExecContext(ctx, `
INSERT INTO public.object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, observed_size, observed_content_type, observed_sha256_hex,
    target_expires_at, pending_expires_at, finalized_at
) VALUES (
	$1, $2, $3, 'incidents/fixture/objects/fixture', 'available',
	1, 1, 'text/plain', repeat('a', 64), now(), now(), now()
)
`, blobID, incidentID, actorID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO public.evidence (
	record_id, incident_id, title, lifecycle_state, upload_state, object_blob_id
) VALUES ($1, $2, 'first', 'received', 'available', $3)
`, firstRecord, incidentID, blobID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `SET session_replication_role = origin`); err != nil {
		t.Fatal(err)
	}

	requireEvidenceBlobIndexState(t, db, true, false)
	if _, err := db.ExecContext(ctx, `SET session_replication_role = replica`); err != nil {
		t.Fatal(err)
	}
	_, err := db.ExecContext(ctx, `
INSERT INTO public.evidence (
    record_id, incident_id, title, lifecycle_state, upload_state, object_blob_id
) VALUES ($1, $2, 'duplicate rejected', 'received', 'available', $3)
`, secondRecord, incidentID, blobID)
	if err == nil || !strings.Contains(err.Error(), "evidence_object_blob_unique_idx") {
		t.Fatalf("duplicate association error = %v, want evidence_object_blob_unique_idx", err)
	}
}

func requireEvidenceBlobIndexState(t testing.TB, db *sql.DB, uniquePresent bool, supersededPresent bool) {
	t.Helper()
	for indexName, want := range map[string]bool{
		"evidence_object_blob_unique_idx": uniquePresent,
		"evidence_object_blob_idx":        supersededPresent,
	} {
		var present bool
		if err := db.QueryRowContext(context.Background(), `
SELECT EXISTS (
    SELECT 1
      FROM pg_indexes
     WHERE schemaname = 'public'
       AND indexname = $1
)
`, indexName).Scan(&present); err != nil {
			t.Fatal(err)
		}
		if present != want {
			t.Fatalf("index %s present = %t, want %t", indexName, present, want)
		}
	}
}
