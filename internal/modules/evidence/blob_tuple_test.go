package evidence

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

// TestLiveBlobTupleViolationFailsClosed_Unit proves REQ-02-189a on live Blob
// paths: a Blob whose dependent fields violate the lifecycle tuple fails the
// next load-for-update operation as an unclassified invariant error with no
// committed effect, while a coherent Blob still loads. Cleanup leaves violating
// Blobs unclaimed and scans past them, so they cannot starve the batch.
func TestLiveBlobTupleViolationFailsClosed_Unit(t *testing.T) {
	harness := newTestStore(t, "evidence-blob-tuple-violation")
	store := newTestBlobLifecycleService(t, harness)
	actor := authstoretest.SeedActor(t, harness, authstoretest.ActorAttributes{Email: "evidence-blob-tuple@example.test", DisplayName: "Evidence Blob Tuple"})
	incident := createTestIncident(t, harness, actor, "txn-evidence-blob-tuple-incident", "IR-BLOB-TUPLE", "Evidence Blob tuple")
	coherentAvailable := BlobOptions{ByteSize: 4, ObservedSize: ptrInt64(4), ObservedSHA: ptrString(strings.Repeat("a", 64)), ObservedContentType: ptrString("text/plain")}

	for _, tc := range []struct {
		name    string
		state   string
		options BlobOptions
		corrupt string
	}{
		{name: "available without finalized_at", state: "available", options: coherentAvailable, corrupt: `UPDATE object_blobs SET finalized_at = NULL WHERE object_blob_id = $1`},
		{name: "pending with a failure clock", state: "pending", options: BlobOptions{ByteSize: 1}, corrupt: `UPDATE object_blobs SET failed_at = created_at WHERE object_blob_id = $1`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
			blobID := seedBlob(t, harness, incident.ID, actor.ID, tc.state, tc.options)
			if _, err := harness.Exec(context.Background(), tc.corrupt, blobID); err != nil {
				t.Fatalf("corrupt Blob tuple: %v", err)
			}
			beforeBlob := blobRowSnapshot(t, harness, blobID)
			beforeCounts := DurableAttachCounts(t, harness, recordID)

			request := attachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-tuple-" + tc.state}
			_, err := store.AttachBlob(context.Background(), actor, recordID, request, attachBlobRequestHash(request), "req-tuple-"+tc.state, time.Now().UTC())
			if !errors.Is(err, errBlobTupleViolation) {
				t.Fatalf("attach error = %v, want the Blob tuple invariant violation", err)
			}
			var rejected AttachRejectedError
			var transition *blobTransitionError
			if errors.As(err, &rejected) || errors.As(err, &transition) {
				t.Fatalf("tuple violation %v was classified as a user-correctable rejection", err)
			}
			if apiErr := translateAttachError(err, request.ClientTxnID); apiErr == nil || apiErr.Status != http.StatusInternalServerError || apiErr.Code != "internal_error" || len(apiErr.Details) != 0 {
				t.Fatalf("tuple violation route error = %#v, want 500 internal_error with empty details", apiErr)
			}
			if after := blobRowSnapshot(t, harness, blobID); after != beforeBlob {
				t.Fatalf("Blob row changed after a failed load:\nbefore %s\nafter  %s", beforeBlob, after)
			}
			if after := DurableAttachCounts(t, harness, recordID); after != beforeCounts {
				t.Fatalf("durable attach state changed: before %#v after %#v", beforeCounts, after)
			}
			requireEvidenceState(t, harness, recordID, "received", "pending", uuid.Nil)
		})
	}

	t.Run("coherent Blob still loads", func(t *testing.T) {
		recordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
		blobID := seedBlob(t, harness, incident.ID, actor.ID, "available", coherentAvailable)
		request := attachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-tuple-coherent"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, request, attachBlobRequestHash(request), "req-tuple-coherent", time.Now().UTC()); err != nil {
			t.Fatalf("attach coherent Blob: %v", err)
		}
		requireEvidenceState(t, harness, recordID, "received", "available", blobID)
	})

	t.Run("cleanup scans past violating Blobs without claiming them", func(t *testing.T) {
		cleanup, err := newCleanupService(harness)
		if err != nil {
			t.Fatalf("compose cleanup service: %v", err)
		}
		base := time.Now().UTC().Truncate(time.Second)
		// A full batch of violating Blobs is due before the coherent one.
		violating := make([]uuid.UUID, 0, cleanupBatchSize)
		for range cleanupBatchSize {
			blobID := uuid.New()
			insertFailedCleanupBlob(t, harness, incident.ID, actor.ID, blobID, cleanupKey("tuple-violating", blobID), base)
			violating = append(violating, blobID)
		}
		if _, err := harness.Exec(context.Background(), `UPDATE object_blobs SET observed_size = byte_size WHERE object_blob_id = ANY($1)`, violating); err != nil {
			t.Fatalf("corrupt cleanup Blob tuples: %v", err)
		}
		coherent := uuid.New()
		insertFailedCleanupBlob(t, harness, incident.ID, actor.ID, coherent, cleanupKey("tuple-coherent", coherent), base.Add(time.Minute))

		claims, err := cleanup.claimCleanupBatch(context.Background(), base.Add(time.Minute))
		if err != nil {
			t.Fatalf("claim cleanup batch: %v", err)
		}
		if len(claims) != 1 || claims[0].ObjectBlobID != coherent {
			t.Fatalf("cleanup claims = %#v, want only the coherent Blob %s", claims, coherent)
		}
		var violatingClaims int
		if err := harness.QueryRow(context.Background(), `SELECT COUNT(*) FROM evidence_blob_cleanup_claims WHERE object_blob_id = ANY($1)`, violating).Scan(&violatingClaims); err != nil {
			t.Fatalf("count violating cleanup claims: %v", err)
		}
		if violatingClaims != 0 {
			t.Fatalf("cleanup claimed %d tuple-violating Blobs", violatingClaims)
		}
	})
}

func blobRowSnapshot(t testing.TB, db dbport.DB, blobID uuid.UUID) string {
	t.Helper()
	var snapshot string
	if err := db.QueryRow(context.Background(), `
SELECT jsonb_build_object(
           'blob', (SELECT to_jsonb(b) FROM object_blobs b WHERE b.object_blob_id = $1),
           'lease', (SELECT to_jsonb(l) FROM evidence_object_upload_leases l WHERE l.object_blob_id = $1)
       )::text
`, blobID).Scan(&snapshot); err != nil {
		t.Fatalf("snapshot Blob row: %v", err)
	}
	return snapshot
}
