package evidence_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestEvidenceBlobUniquenessMigration53PreflightAndEnforcement_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	migrationDB := harness.MigrationDatabaseThroughT(t, 52)
	db := migrationDB.SQL()
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
INSERT INTO object_blobs (
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
INSERT INTO evidence (
    record_id, incident_id, title, lifecycle_state, upload_state, object_blob_id
) VALUES
    ($3, $2, 'first', 'received', 'available', $1),
    ($4, $2, 'second', 'received', 'available', $1)
`, blobID, incidentID, firstRecord, secondRecord); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `SET session_replication_role = origin`); err != nil {
		t.Fatal(err)
	}

	err := migrationDB.ApplyThrough(ctx, 53)
	if err == nil {
		t.Fatal("expected duplicate Evidence blob association preflight rejection")
	}
	message := err.Error()
	if !strings.Contains(message, "evidence blob association uniqueness preflight failed: duplicate_association_count=1") {
		t.Fatalf("preflight error = %v", err)
	}
	for _, privateID := range []string{blobID, firstRecord, secondRecord, incidentID} {
		if strings.Contains(message, privateID) {
			t.Fatalf("preflight disclosed row identity %s: %v", privateID, err)
		}
	}
	requireEvidenceBlobIndexState(t, db, false, true)

	if _, err := db.ExecContext(ctx, `DELETE FROM evidence WHERE record_id = $1`, secondRecord); err != nil {
		t.Fatal(err)
	}
	if err := migrationDB.ApplyThrough(ctx, 53); err != nil {
		t.Fatalf("apply migration after owner-reviewed correction: %v", err)
	}
	requireEvidenceBlobIndexState(t, db, true, false)
	if _, err := db.ExecContext(ctx, `
INSERT INTO evidence (
    record_id, incident_id, title, lifecycle_state, upload_state, object_blob_id
) VALUES ($1, $2, 'duplicate rejected', 'received', 'available', $3)
`, secondRecord, incidentID, blobID); err == nil || !strings.Contains(err.Error(), "evidence_object_blob_unique_idx") {
		t.Fatalf("duplicate association error = %v, want evidence_object_blob_unique_idx", err)
	}

	if err := migrationDB.RollbackThrough(ctx, 52); err != nil {
		t.Fatalf("down migration: %v", err)
	}
	requireEvidenceBlobIndexState(t, db, false, true)
	if err := migrationDB.ApplyThrough(ctx, 53); err != nil {
		t.Fatalf("reapply migration: %v", err)
	}
	requireEvidenceBlobIndexState(t, db, true, false)
}

func requireEvidenceBlobIndexState(t testing.TB, db *sql.DB, uniquePresent bool, legacyPresent bool) {
	t.Helper()
	for indexName, want := range map[string]bool{
		"evidence_object_blob_unique_idx": uniquePresent,
		"evidence_object_blob_idx":        legacyPresent,
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
