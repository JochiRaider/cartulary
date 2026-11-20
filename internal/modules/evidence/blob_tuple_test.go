package evidence

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	evidenceblob "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/blob"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/testsupport/blobfixture"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

// TestLiveBlobTupleViolationFailsClosed_Unit proves REQ-02-189a on live Blob
// paths: a Blob whose dependent fields violate the lifecycle tuple fails the
// next load-for-update operation as an unclassified invariant error with no
// committed effect, while a coherent Blob still loads. Cleanup leaves violating
// Blobs unclaimed, counts them, and scans past them, so they cannot starve the
// batch.
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

		claims, violations, err := cleanup.claimCleanupBatch(context.Background(), base.Add(time.Minute))
		if err != nil {
			t.Fatalf("claim cleanup batch: %v", err)
		}
		if len(claims) != 1 || claims[0].ObjectBlobID != coherent {
			t.Fatalf("cleanup claims = %#v, want only the coherent Blob %s", claims, coherent)
		}
		if violations != cleanupBatchSize {
			t.Fatalf("cleanup claim invariant violations = %d, want every skipped Blob counted (%d)", violations, cleanupBatchSize)
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

// TestBlobLinkAdmissionDecidesAgainstLockedBlob_Unit proves the single Blob
// authority's link admission (REQ-02-190, REQ-02-196a): it classifies every
// association-rule and bridge violation as a rejection, admits only a locked,
// coherent, same-incident, finalized, unassociated and unclaimed Blob, and
// fails an out-of-vocabulary Evidence lifecycle or a violating Blob tuple as a
// storage invariant error.
func TestBlobLinkAdmissionDecidesAgainstLockedBlob_Unit(t *testing.T) {
	harness := newTestStore(t, "evidence-blob-link-admission")
	actor := authstoretest.SeedActor(t, harness, authstoretest.ActorAttributes{Email: "evidence-blob-link@example.test", DisplayName: "Evidence Blob Link"})
	incident := createTestIncident(t, harness, actor, "txn-evidence-blob-link-incident", "IR-BLOB-LINK", "Evidence Blob link admission")
	foreign := createTestIncident(t, harness, actor, "txn-evidence-blob-link-foreign", "IR-BLOB-LINK-FOREIGN", "Evidence Blob link foreign")
	finalized := func(state string, incidentID uuid.UUID) uuid.UUID {
		return blobfixture.Seed(t, harness, blobfixture.Spec{IncidentID: incidentID, CreatedByUserID: actor.ID, State: state, ByteSize: 3})
	}
	evidenceID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
	other := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
	associated := finalized(blobfixture.Available, incident.ID)
	linkEvidenceBlobInStore(t, harness, other, associated)
	own := finalized(blobfixture.Available, incident.ID)
	linkEvidenceBlobInStore(t, harness, evidenceID, own)
	claimed := finalized(blobfixture.Available, incident.ID)
	blobfixture.ClaimForCleanup(t, harness, claimed)
	breached := finalized(blobfixture.Available, incident.ID)
	blobfixture.BreachTupleInvariant(t, harness, breached)
	missing := uuid.New()

	admit := func(t *testing.T, lifecycle string, blobID *uuid.UUID) (evidenceblob.Admission, error) {
		t.Helper()
		tx, err := harness.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin admission transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		return evidenceblob.AdmitLinkTx(context.Background(), tx, evidenceblob.LinkRequest{
			EvidenceID: evidenceID, IncidentID: incident.ID, Lifecycle: lifecycle, BlobID: blobID, Mode: evidenceblob.ShareNoWait,
		})
	}
	for _, tc := range []struct {
		name      string
		lifecycle string
		blobID    uuid.UUID
		outcome   evidenceblob.Outcome
		rejection evidenceblob.Rejection
	}{
		{name: "unlinked received", lifecycle: "received", outcome: evidenceblob.Admitted},
		{name: "unlinked available", lifecycle: "available", outcome: evidenceblob.Rejected, rejection: evidenceblob.RejectBridge},
		{name: "available", lifecycle: "available", blobID: finalized(blobfixture.Available, incident.ID), outcome: evidenceblob.Admitted},
		{name: "own association", lifecycle: "available", blobID: own, outcome: evidenceblob.Admitted},
		{name: "quarantined under available", lifecycle: "available", blobID: finalized(blobfixture.Quarantined, incident.ID), outcome: evidenceblob.Rejected, rejection: evidenceblob.RejectBridge},
		{name: "pending", lifecycle: "received", blobID: finalized(blobfixture.Pending, incident.ID), outcome: evidenceblob.Rejected, rejection: evidenceblob.RejectNotFinalized},
		{name: "failed", lifecycle: "received", blobID: finalized(blobfixture.Failed, incident.ID), outcome: evidenceblob.Rejected, rejection: evidenceblob.RejectNotFinalized},
		{name: "foreign incident", lifecycle: "received", blobID: finalized(blobfixture.Available, foreign.ID), outcome: evidenceblob.Rejected, rejection: evidenceblob.RejectForeignIncident},
		{name: "associated elsewhere", lifecycle: "received", blobID: associated, outcome: evidenceblob.Rejected, rejection: evidenceblob.RejectAssociated},
		{name: "cleanup claimed", lifecycle: "received", blobID: claimed, outcome: evidenceblob.Rejected, rejection: evidenceblob.RejectCleanupClaimed},
		{name: "missing", lifecycle: "received", blobID: missing, outcome: evidenceblob.Rejected, rejection: evidenceblob.RejectMissing},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var blobID *uuid.UUID
			if tc.blobID != uuid.Nil {
				blobID = &tc.blobID
			}
			got, err := admit(t, tc.lifecycle, blobID)
			if err != nil {
				t.Fatalf("admission error = %v", err)
			}
			if got.Outcome != tc.outcome || got.Rejection != tc.rejection {
				t.Fatalf("admission = outcome %d rejection %q, want outcome %d rejection %q", got.Outcome, got.Rejection, tc.outcome, tc.rejection)
			}
			if tc.outcome == evidenceblob.Admitted && blobID != nil && (got.Blob == nil || got.Blob.ObjectBlobID != *blobID) {
				t.Fatalf("admitted Blob = %#v, want the locked linked Blob %s", got.Blob, *blobID)
			}
		})
	}
	t.Run("lifecycle outside the vocabulary is an invariant failure", func(t *testing.T) {
		got, err := admit(t, "archived", &own)
		if !errors.Is(err, evidenceblob.ErrEvidenceLifecycleViolation) || got.Outcome != 0 {
			t.Fatalf("admission = (%#v, %v), want the Evidence lifecycle invariant failure", got, err)
		}
	})
	t.Run("violating Blob tuple is an invariant failure", func(t *testing.T) {
		got, err := admit(t, "received", &breached)
		if !errors.Is(err, errBlobTupleViolation) || got.Outcome != 0 {
			t.Fatalf("admission = (%#v, %v), want the Blob tuple invariant failure", got, err)
		}
	})
}

// TestBlobTimeoutAndCleanupCompletionFailClosed_Unit proves REQ-02-189a on
// the background Blob sweep: the pending timeout, the cleanup claim and the
// cleanup completion verify the tuple under the Blob lock, leave a violating
// Blob (and its claim) unchanged, count it, and continue, so a coherent
// sibling in the same sweep is still processed and committed.
func TestBlobTimeoutAndCleanupCompletionFailClosed_Unit(t *testing.T) {
	harness := newTestStore(t, "evidence-blob-background-tuple")
	actor := authstoretest.SeedActor(t, harness, authstoretest.ActorAttributes{Email: "evidence-blob-background@example.test", DisplayName: "Evidence Blob Background"})
	incident := createTestIncident(t, harness, actor, "txn-evidence-blob-background-incident", "IR-BLOB-BACKGROUND", "Evidence Blob background tuple")
	now := time.Now().UTC().Truncate(time.Second)
	cleanup, err := newCleanupService(harness)
	if err != nil {
		t.Fatalf("compose cleanup service: %v", err)
	}
	claimSnapshot := func(t *testing.T, blobID uuid.UUID) string {
		t.Helper()
		var snapshot string
		if err := harness.QueryRow(context.Background(), `SELECT COALESCE((SELECT to_jsonb(c)::text FROM evidence_blob_cleanup_claims c WHERE c.object_blob_id = $1), 'none')`, blobID).Scan(&snapshot); err != nil {
			t.Fatalf("snapshot cleanup claim: %v", err)
		}
		return snapshot
	}

	t.Run("pending timeout counts a violating Blob and times out its sibling", func(t *testing.T) {
		expired := func() uuid.UUID {
			return blobfixture.Seed(t, harness, blobfixture.Spec{IncidentID: incident.ID, CreatedByUserID: actor.ID, State: blobfixture.Pending, ByteSize: 2, CreatedAt: now.Add(-25 * time.Hour)})
		}
		coherent, violating := expired(), expired()
		blobfixture.BreachTupleInvariant(t, harness, violating)
		before := blobRowSnapshot(t, harness, violating)
		count, violations, err := (blobLifecycleRepository{db: harness}).markExpiredPending(context.Background(), now)
		if err != nil {
			t.Fatalf("mark expired pending Blobs: %v", err)
		}
		if count != 1 || violations != 1 {
			t.Fatalf("pending timeout = %d expired, %d invariant violations; want the coherent Blob expired and the violating one counted", count, violations)
		}
		if after := blobRowSnapshot(t, harness, violating); after != before {
			t.Fatalf("pending timeout changed a tuple-violating Blob:\nbefore %s\nafter  %s", before, after)
		}
		var state, reason string
		if err := harness.QueryRow(context.Background(), `SELECT upload_state, terminal_reason FROM object_blobs WHERE object_blob_id = $1`, coherent).Scan(&state, &reason); err != nil {
			t.Fatalf("load timed-out Blob: %v", err)
		}
		if state != "failed" || reason != "pending_timeout" {
			t.Fatalf("coherent expired Blob = %s/%s, want failed/pending_timeout", state, reason)
		}
		blobfixture.RequireCoherent(t, harness, coherent)
		deleteCleanupFixture(t, harness, violating)
	})

	t.Run("cleanup claim counts a violating Blob and claims its sibling", func(t *testing.T) {
		coherent, violating := uuid.New(), uuid.New()
		insertFailedCleanupBlob(t, harness, incident.ID, actor.ID, coherent, cleanupKey("claim-coherent", coherent), now)
		insertFailedCleanupBlob(t, harness, incident.ID, actor.ID, violating, cleanupKey("claim-violating", violating), now)
		blobfixture.BreachTupleInvariant(t, harness, violating)
		before := blobRowSnapshot(t, harness, violating)
		claims, violations, err := cleanup.claimCleanupBatch(context.Background(), now)
		if err != nil {
			t.Fatalf("claim cleanup batch: %v", err)
		}
		if len(claims) != 1 || claims[0].ObjectBlobID != coherent || violations != 1 {
			t.Fatalf("cleanup claim = %#v with %d invariant violations; want only %s claimed and one violation counted", claims, violations, coherent)
		}
		if after := blobRowSnapshot(t, harness, violating); after != before {
			t.Fatalf("cleanup claim changed a tuple-violating Blob:\nbefore %s\nafter  %s", before, after)
		}
		if claim := claimSnapshot(t, violating); claim != "none" {
			t.Fatalf("cleanup claimed a tuple-violating Blob: %s", claim)
		}
		deleteCleanupFixture(t, harness, violating)
	})

	t.Run("cleanup completion counts a violating Blob and completes its sibling", func(t *testing.T) {
		coherent, violating := uuid.New(), uuid.New()
		insertFailedCleanupBlob(t, harness, incident.ID, actor.ID, coherent, cleanupKey("completion-coherent", coherent), now)
		insertFailedCleanupBlob(t, harness, incident.ID, actor.ID, violating, cleanupKey("completion-violating", violating), now)
		// The Blob is coherent when claimed; the breach lands between the claim
		// and its completion, while the sweep deletes the object bytes.
		deleter := &tupleBreachingCleanupDeleter{t: t, db: harness, key: cleanupKey("completion-violating", violating), blobID: violating, claim: claimSnapshot}
		result, err := cleanup.SweepFailedUnattachedBlobs(context.Background(), deleter, now)
		if err != nil {
			t.Fatalf("sweep stopped on a tuple-violating Blob: %v", err)
		}
		if result.ClaimedBlobCount != 2 || result.CleanedBlobCount != 1 || result.RetryScheduledCount != 0 || result.InvariantViolationCount != 1 || !result.HealthSnapshotValid {
			t.Fatalf("sweep result = %#v, want both claimed, the coherent Blob cleaned and one invariant violation counted", result)
		}
		if deleter.blobAfterBreach == "" {
			t.Fatal("sweep never deleted the violating Blob's object")
		}
		if after := blobRowSnapshot(t, harness, violating); after != deleter.blobAfterBreach {
			t.Fatalf("cleanup completion changed a tuple-violating Blob:\nbefore %s\nafter  %s", deleter.blobAfterBreach, after)
		}
		if after := claimSnapshot(t, violating); after != deleter.claimAfterBreach {
			t.Fatalf("cleanup completion changed the claim of a tuple-violating Blob:\nbefore %s\nafter  %s", deleter.claimAfterBreach, after)
		}
		var cleaned bool
		var claimState string
		if err := harness.QueryRow(context.Background(), `
SELECT b.cleaned_up_at IS NOT NULL, c.claim_state
  FROM object_blobs b
  JOIN evidence_blob_cleanup_claims c ON c.object_blob_id = b.object_blob_id
 WHERE b.object_blob_id = $1
`, coherent).Scan(&cleaned, &claimState); err != nil {
			t.Fatalf("load completed cleanup Blob: %v", err)
		}
		if !cleaned || claimState != "completed" {
			t.Fatalf("coherent sibling cleaned=%t claim=%s, want its completion committed", cleaned, claimState)
		}
	})
}

// tupleBreachingCleanupDeleter deletes object bytes successfully and breaches
// one Blob's tuple at its deletion, after the claim and before completion.
type tupleBreachingCleanupDeleter struct {
	t                *testing.T
	db               dbport.DB
	key              string
	blobID           uuid.UUID
	claim            func(*testing.T, uuid.UUID) string
	blobAfterBreach  string
	claimAfterBreach string
}

func (deleter *tupleBreachingCleanupDeleter) DeleteObject(_ context.Context, key string) error {
	if key == deleter.key {
		blobfixture.BreachTupleInvariant(deleter.t, deleter.db, deleter.blobID)
		deleter.blobAfterBreach = blobRowSnapshot(deleter.t, deleter.db, deleter.blobID)
		deleter.claimAfterBreach = deleter.claim(deleter.t, deleter.blobID)
	}
	return nil
}
