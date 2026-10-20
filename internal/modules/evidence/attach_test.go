package evidence

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestAttachBlobValidation_Unit(t *testing.T) {
	t.Run("request shape is exact", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			body string
		}{
			{name: "array", body: `[]`},
			{name: "missing object_blob_id", body: `{"base_row_version":1,"client_txn_id":"txn"}`},
			{name: "missing base_row_version", body: `{"object_blob_id":"` + uuid.NewString() + `","client_txn_id":"txn"}`},
			{name: "null base_row_version", body: `{"object_blob_id":"` + uuid.NewString() + `","base_row_version":null,"client_txn_id":"txn"}`},
			{name: "missing client_txn_id", body: `{"object_blob_id":"` + uuid.NewString() + `","base_row_version":1}`},
			{name: "null client_txn_id", body: `{"object_blob_id":"` + uuid.NewString() + `","base_row_version":1,"client_txn_id":null}`},
			{name: "unknown member", body: `{"object_blob_id":"` + uuid.NewString() + `","base_row_version":1,"client_txn_id":"txn","extra":true}`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				_, apiErr := decodeAttachBlobRequest(strings.NewReader(tc.body))
				if apiErr == nil || apiErr.Status != http.StatusBadRequest || apiErr.Code != "invalid_mutation_payload" {
					t.Fatalf("expected invalid_mutation_payload, got %#v", apiErr)
				}
			})
		}
	})

	harness := newTestStore(t, "evidence_lifecycle-attach-validation")
	store := newTestBlobLifecycleService(t, harness)
	actor := authstoretest.SeedActor(t, harness, authstoretest.ActorAttributes{Email: "evidence_lifecycle-attach@example.test", DisplayName: "EvidenceLifecycle Attach"})
	incident := createTestIncident(t, harness, actor, "txn-evidence_lifecycle-attach-incident", "IR-P5-ATTACH", "Evidence attach")
	otherIncident := createTestIncident(t, harness, actor, "txn-evidence_lifecycle-attach-other", "IR-P5-ATTACH-OTHER", "Evidence attach other")

	t.Run("success and replay are stable", func(t *testing.T) {
		recordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
		blobID := seedBlob(t, harness, incident.ID, actor.ID, "available", BlobOptions{
			ByteSize: 4, ObservedSize: ptrInt64(4), ObservedSHA: ptrString(strings.Repeat("a", 64)), ObservedContentType: ptrString("text/plain"),
		})
		request := attachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-attach-success"}
		result, err := store.AttachBlob(context.Background(), actor, recordID, request, attachBlobRequestHash(request), "req-success", time.Now().UTC())
		if err != nil {
			t.Fatalf("attach available blob: %v", err)
		}
		if result.StatusCode != http.StatusOK || result.Replayed {
			t.Fatalf("unexpected attach result: %#v", result)
		}
		requireEvidenceState(t, harness, recordID, "received", "available", blobID)
		changeSet := result.Payload["change_set_id"]

		replay, err := store.AttachBlob(context.Background(), actor, recordID, request, attachBlobRequestHash(request), "req-replay", time.Now().UTC())
		if err != nil {
			t.Fatalf("replay attach: %v", err)
		}
		if !replay.Replayed || replay.StatusCode != http.StatusOK || replay.Payload["change_set_id"] != changeSet {
			t.Fatalf("replay did not return original payload: %#v want change_set_id %#v", replay, changeSet)
		}
		requireChangeSetCount(t, harness, recordID, 1)
		// A committed receipt cannot bypass current blob eligibility after preflight.
		if _, err := harness.Exec(context.Background(), `UPDATE object_blobs SET upload_state = 'quarantined' WHERE object_blob_id = $1`, blobID); err != nil {
			t.Fatal(err)
		}
		_, err = store.AttachBlob(context.Background(), actor, recordID, request, attachBlobRequestHash(request), "req-replay-quarantined", time.Now().UTC())
		var rejected AttachRejectedError
		if !errors.As(err, &rejected) || rejected.ReasonCode != attachReasonBlobQuarantined {
			t.Fatalf("replay current eligibility = %v", err)
		}
		requireChangeSetCount(t, harness, recordID, 1)
	})

	t.Run("divergent replay is rejected without durable side effects", func(t *testing.T) {
		recordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
		blobID := seedBlob(t, harness, incident.ID, actor.ID, "available", BlobOptions{
			ByteSize: 4, ObservedSize: ptrInt64(4), ObservedSHA: ptrString(strings.Repeat("a", 64)), ObservedContentType: ptrString("text/plain"),
		})
		request := attachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-attach-divergent"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, request, attachBlobRequestHash(request), "req-divergent-first", time.Now().UTC()); err != nil {
			t.Fatalf("attach available blob: %v", err)
		}
		before := DurableAttachCounts(t, harness, recordID)

		otherBlobID := seedBlob(t, harness, incident.ID, actor.ID, "available", BlobOptions{
			ByteSize: 4, ObservedSize: ptrInt64(4), ObservedSHA: ptrString(strings.Repeat("b", 64)), ObservedContentType: ptrString("text/plain"),
		})
		divergent := attachBlobRequest{ObjectBlobID: otherBlobID, BaseRowVersion: 2, ClientTxnID: request.ClientTxnID}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, divergent, attachBlobRequestHash(divergent), "req-divergent-second", time.Now().UTC()); !errors.Is(err, ErrClientTxnConflict) {
			t.Fatalf("divergent replay got %v want ErrClientTxnConflict", err)
		}
		after := DurableAttachCounts(t, harness, recordID)
		if after != before {
			t.Fatalf("divergent replay changed durable counts: before=%+v after=%+v", before, after)
		}
		requireEvidenceState(t, harness, recordID, "received", "available", blobID)
	})

	t.Run("conflict and lifecycle failures leave evidence unchanged", func(t *testing.T) {
		recordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
		blobID := seedBlob(t, harness, incident.ID, actor.ID, "available", BlobOptions{ByteSize: 1, ObservedSize: ptrInt64(1)})
		request := attachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 2, ClientTxnID: "txn-row-conflict"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, request, attachBlobRequestHash(request), "req-conflict", time.Now().UTC()); err == nil || err.Error() != "evidence: row version conflict" {
			t.Fatalf("row-version conflict got %v", err)
		}
		requireEvidenceState(t, harness, recordID, "received", "pending", uuid.Nil)

		missing := attachBlobRequest{ObjectBlobID: uuid.New(), BaseRowVersion: 1, ClientTxnID: "txn-missing"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, missing, attachBlobRequestHash(missing), "req-missing", time.Now().UTC()); !errors.Is(err, ErrBlobNotFound) {
			t.Fatalf("missing blob got %v", err)
		}
		requireEvidenceState(t, harness, recordID, "received", "pending", uuid.Nil)

		foreignBlob := seedBlob(t, harness, otherIncident.ID, actor.ID, "available", BlobOptions{ByteSize: 1, ObservedSize: ptrInt64(1)})
		foreign := attachBlobRequest{ObjectBlobID: foreignBlob, BaseRowVersion: 1, ClientTxnID: "txn-foreign"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, foreign, attachBlobRequestHash(foreign), "req-foreign", time.Now().UTC()); !errors.Is(err, ErrIncidentMismatch) {
			t.Fatalf("incident mismatch got %v", err)
		} else {
			requireAttachRejectedReason(t, err, AttachReasonBlobNotVisible)
		}
		requireEvidenceState(t, harness, recordID, "received", "pending", uuid.Nil)

		for _, state := range []string{"failed", "quarantined"} {
			blockedBlob := seedBlob(t, harness, incident.ID, actor.ID, state, BlobOptions{ByteSize: 1})
			blocked := attachBlobRequest{ObjectBlobID: blockedBlob, BaseRowVersion: 1, ClientTxnID: "txn-" + state}
			if _, err := store.AttachBlob(context.Background(), actor, recordID, blocked, attachBlobRequestHash(blocked), "req-"+state, time.Now().UTC()); err == nil {
				t.Fatalf("%s blob got %v", state, err)
			} else {
				wantReason := "blob_failed"
				if state == "quarantined" {
					wantReason = "blob_quarantined"
				}
				requireAttachRejectedReason(t, err, wantReason)
			}
			requireEvidenceState(t, harness, recordID, "received", "pending", uuid.Nil)
		}
	})

	t.Run("terminal finalization failures persist on blob only", func(t *testing.T) {
		recordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
		expiredBlob := seedBlob(t, harness, incident.ID, actor.ID, "pending", BlobOptions{ByteSize: 1, PendingExpiresAt: time.Now().UTC().Add(-time.Minute)})
		expired := attachBlobRequest{ObjectBlobID: expiredBlob, BaseRowVersion: 1, ClientTxnID: "txn-expired"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, expired, attachBlobRequestHash(expired), "req-expired", time.Now().UTC()); err == nil {
			t.Fatalf("expired blob got %v", err)
		} else {
			requireAttachRejectedReason(t, err, "blob_failed")
		}
		requireBlobFailure(t, harness, expiredBlob, "pending_timeout")

		sizeBlob := seedBlob(t, harness, incident.ID, actor.ID, "pending", BlobOptions{ByteSize: 5})
		seedRetainedObservation(t, harness, sizeBlob, 4, strings.Repeat("a", 64))
		sizeMismatch := attachBlobRequest{ObjectBlobID: sizeBlob, BaseRowVersion: 1, ClientTxnID: "txn-size"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, sizeMismatch, attachBlobRequestHash(sizeMismatch), "req-size", time.Now().UTC()); err == nil {
			t.Fatalf("size mismatch got %v", err)
		} else {
			requireAttachRejectedReason(t, err, "accepted_contract_mismatch")
		}
		requireBlobFailure(t, harness, sizeBlob, "declared_size_mismatch")

		hashBlob := seedBlob(t, harness, incident.ID, actor.ID, "pending", BlobOptions{ByteSize: 4, ExpectedSHA: ptrString(strings.Repeat("e", 64))})
		seedRetainedObservation(t, harness, hashBlob, 4, strings.Repeat("a", 64))
		hashMismatch := attachBlobRequest{ObjectBlobID: hashBlob, BaseRowVersion: 1, ClientTxnID: "txn-hash"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, hashMismatch, attachBlobRequestHash(hashMismatch), "req-hash", time.Now().UTC()); err == nil {
			t.Fatalf("hash mismatch got %v", err)
		} else {
			requireAttachRejectedReason(t, err, "accepted_contract_mismatch")
		}
		requireBlobFailure(t, harness, hashBlob, "expected_sha256_mismatch")
		requireEvidenceState(t, harness, recordID, "received", "pending", uuid.Nil)
	})

	t.Run("transaction rejects an incomplete lease without effects", func(t *testing.T) {
		recordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
		blobID := seedBlob(t, harness, incident.ID, actor.ID, "pending", BlobOptions{ByteSize: 1})
		if _, err := harness.Exec(context.Background(), `UPDATE evidence_object_upload_leases SET lease_state='claimed', completed_at=NULL WHERE object_blob_id=$1`, blobID); err != nil {
			t.Fatal(err)
		}
		request := attachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-incomplete"}
		_, err := store.AttachBlob(context.Background(), actor, recordID, request, attachBlobRequestHash(request), "req-incomplete", time.Now().UTC())
		requireAttachRejectedReason(t, err, "blob_pending")
		requirePendingBlob(t, harness, blobID)
		requireEvidenceState(t, harness, recordID, "received", "pending", uuid.Nil)
	})

	t.Run("missing verification remains effect-free across explicit finalization calls", func(t *testing.T) {
		recordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
		blobID := seedBlob(t, harness, incident.ID, actor.ID, "pending", BlobOptions{ByteSize: 1})
		for attempt := 0; attempt < 4; attempt++ {
			request := attachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-retry"}
			_, err := store.AttachBlob(context.Background(), actor, recordID, request, attachBlobRequestHash(request), "req-retry", time.Now().UTC())
			requireAttachRejectedReason(t, err, "blob_pending")
			requirePendingBlob(t, harness, blobID)
		}
		requireEvidenceState(t, harness, recordID, "received", "pending", uuid.Nil)
	})
}

type BlobOptions struct {
	ByteSize            int64
	ExpectedSHA         *string
	ObservedSize        *int64
	ObservedContentType *string
	ObservedSHA         *string
	PendingExpiresAt    time.Time
}

func seedEvidenceAttachmentRecord(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorID uuid.UUID, lifecycleState string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	now := time.Now().UTC()
	if _, err := db.Exec(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, created_at, updated_by_user_id, updated_at, row_version)
VALUES ($1, $2, 'evidence', $3, $4, $3, $4, 1)
`, recordID, incidentID, actorID, now); err != nil {
		t.Fatalf("insert evidence record envelope: %v", err)
	}
	if _, err := db.Exec(context.Background(), `
INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state, requested_at, created_at, updated_at)
VALUES ($1, $2, 'Evidence evidence', $3, 'pending', $4, $4, $4)
`, recordID, incidentID, lifecycleState, now); err != nil {
		t.Fatalf("insert evidence row: %v", err)
	}
	return recordID
}

func seedBlob(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorID uuid.UUID, uploadState string, options BlobOptions) uuid.UUID {
	t.Helper()
	blobID := uuid.New()
	now := time.Now().UTC()
	pendingExpiresAt := options.PendingExpiresAt
	if pendingExpiresAt.IsZero() {
		pendingExpiresAt = now.Add(24 * time.Hour)
	}
	terminalReason := any(nil)
	failedAt := any(nil)
	cleanupDueAt := any(nil)
	finalizedAt := any(nil)
	if uploadState == "failed" {
		terminalReason = "pending_timeout"
		failedAt = now
		cleanupDueAt = now.Add(time.Hour)
	}
	if uploadState == "available" || uploadState == "quarantined" {
		finalizedAt = now
	}
	byteSize := options.ByteSize
	if byteSize == 0 && options.ObservedSize != nil {
		byteSize = *options.ObservedSize
	}
	if _, err := db.Exec(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, expected_sha256_hex, observed_size, observed_content_type, observed_sha256_hex,
    target_expires_at, pending_expires_at, finalized_at, terminal_reason, failed_at,
    cleanup_due_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13,
    $14, $15, $16, $17, $17
)
`, blobID, incidentID, actorID, "evidence_lifecycle/attach/"+blobID.String(), uploadState,
		byteSize, options.ExpectedSHA, options.ObservedSize, options.ObservedContentType, options.ObservedSHA,
		now.Add(time.Hour), pendingExpiresAt, finalizedAt, terminalReason, failedAt, cleanupDueAt, now); err != nil {
		t.Fatalf("insert object blob: %v", err)
	}
	if uploadState == "pending" {
		sessionID := uuid.New()
		fingerprint := sha256.Sum256([]byte(sessionID.String()))
		if _, err := db.Exec(context.Background(), `
INSERT INTO user_sessions (id, user_id, token_fingerprint, authenticated_at, last_qualifying_activity_at,
 idle_expires_at, absolute_expires_at, session_expires_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $4, $5, $5, $5, $4, $4)
`, sessionID, actorID, fingerprint[:], now, now.Add(time.Hour)); err != nil {
			t.Fatalf("seed finalization session: %v", err)
		}
		if _, err := db.Exec(context.Background(), `
INSERT INTO evidence_object_upload_leases (object_blob_id, lease_id, capability_hash, incident_id, issuing_user_id,
 issuing_session_id, issued_at, expires_at, required_method, required_headers, accepted_contract_sha256,
 lease_state, claimed_at, completed_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $2, $6, $7, 'PUT', '{}'::jsonb, $3, 'completed', $6, $6, $6, $6)
`, blobID, sessionID, fingerprint[:], incidentID, actorID, now, now.Add(time.Hour)); err != nil {
			t.Fatalf("seed completed finalization lease: %v", err)
		}
	}
	return blobID
}

func requireEvidenceState(t testing.TB, db postgres.DB, recordID uuid.UUID, wantLifecycle string, wantUpload string, wantBlobID uuid.UUID) {
	t.Helper()
	var lifecycle string
	var upload string
	var blobID *uuid.UUID
	if err := db.QueryRow(context.Background(), `
SELECT e.lifecycle_state, e.upload_state, e.object_blob_id
  FROM evidence e
 WHERE e.record_id = $1
`, recordID).Scan(&lifecycle, &upload, &blobID); err != nil {
		t.Fatalf("load evidence state: %v", err)
	}
	if lifecycle != wantLifecycle || upload != wantUpload {
		t.Fatalf("evidence state got lifecycle=%s upload=%s want lifecycle=%s upload=%s", lifecycle, upload, wantLifecycle, wantUpload)
	}
	if wantBlobID == uuid.Nil {
		if blobID != nil {
			t.Fatalf("evidence object_blob_id got %s want nil", *blobID)
		}
		return
	}
	if blobID == nil || *blobID != wantBlobID {
		t.Fatalf("evidence object_blob_id got %v want %s", blobID, wantBlobID)
	}
}

// Finalization must revalidate even a corrupted retained proof. These store-level
// adversarial fixtures are independent of the real-worker route tests.
func seedRetainedObservation(t testing.TB, db postgres.DB, blobID uuid.UUID, size int64, sha string) {
	t.Helper()
	_, err := db.Exec(context.Background(), `INSERT INTO evidence_blob_verifications(object_blob_id,upload_lease_id,job_id,verified_at,observed_size,observed_content_type,observed_sha256_hex)
    SELECT object_blob_id,lease_id,$2,now(),$3,'text/plain',$4 FROM evidence_object_upload_leases WHERE object_blob_id=$1`, blobID, uuid.New(), size, sha)
	if err != nil {
		t.Fatal(err)
	}
}
func requireBlobFailure(t testing.TB, db postgres.DB, blobID uuid.UUID, wantReason string) {
	t.Helper()
	var state, reason string
	var failed, cleanup *time.Time
	if err := db.QueryRow(context.Background(), `SELECT upload_state,terminal_reason,failed_at,cleanup_due_at FROM object_blobs WHERE object_blob_id=$1`, blobID).Scan(&state, &reason, &failed, &cleanup); err != nil {
		t.Fatal(err)
	}
	if state != "failed" || reason != wantReason || failed == nil || cleanup == nil || !cleanup.Equal(failed.Add(45*time.Minute)) {
		t.Fatalf("failure state=%s reason=%s failed=%v cleanup=%v", state, reason, failed, cleanup)
	}
}
func requirePendingBlob(t testing.TB, db postgres.DB, blobID uuid.UUID) {
	t.Helper()
	var state string
	var reason *string
	if err := db.QueryRow(context.Background(), `SELECT upload_state,terminal_reason FROM object_blobs WHERE object_blob_id=$1`, blobID).Scan(&state, &reason); err != nil {
		t.Fatal(err)
	}
	if state != "pending" || reason != nil {
		t.Fatalf("pending blob state=%s reason=%v", state, reason)
	}
}

func requireAttachRejectedReason(t testing.TB, err error, want string) {
	t.Helper()
	var rejected AttachRejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf("attach error %v does not expose AttachRejectedError", err)
	}
	if rejected.ReasonCode != want {
		t.Fatalf("attach rejected reason got %s want %s", rejected.ReasonCode, want)
	}
}

func requireChangeSetCount(t testing.TB, db postgres.DB, recordID uuid.UUID, want int) {
	t.Helper()
	var count int
	if err := db.QueryRow(context.Background(), `
SELECT COUNT(*)
  FROM record_revisions
 WHERE record_id = $1
`, recordID).Scan(&count); err != nil {
		t.Fatalf("count record revisions: %v", err)
	}
	if count != want {
		t.Fatalf("record revision count got %d want %d", count, want)
	}
}

func TestBlobAssociation_RejectsReuseWithConcealment(t *testing.T) {
	harness := newTestStore(t, "evidence-blob-association-contract")
	store := newTestBlobLifecycleService(t, harness)
	actor := authstoretest.SeedActor(t, harness, authstoretest.ActorAttributes{Email: "evidence-association@example.test", DisplayName: "Evidence Association"})
	incident := createTestIncident(t, harness, actor, "txn-evidence-association-incident", "IR-EVIDENCE-ASSOCIATION", "Evidence association")
	firstRecordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
	secondRecordID := seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received")
	blobID := seedBlob(t, harness, incident.ID, actor.ID, "available", BlobOptions{
		ByteSize:            4,
		ObservedSize:        ptrInt64(4),
		ObservedSHA:         ptrString(strings.Repeat("a", 64)),
		ObservedContentType: ptrString("text/plain"),
	})
	now := time.Date(2026, time.July, 30, 18, 0, 0, 0, time.UTC)

	firstRequest := attachBlobRequest{
		ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-association-first",
	}
	if _, err := store.AttachBlob(context.Background(), actor, firstRecordID, firstRequest, attachBlobRequestHash(firstRequest), "req-association-first", now); err != nil {
		t.Fatalf("attach shared blob to first record: %v", err)
	}
	secondRequest := attachBlobRequest{
		ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-association-second",
	}
	_, err := store.AttachBlob(context.Background(), actor, secondRecordID, secondRequest, attachBlobRequestHash(secondRequest), "req-association-second", now)
	requireAttachRejectedReason(t, err, AttachReasonBlobNotVisible)

	var associationCount int
	if err := harness.QueryRow(context.Background(), `
SELECT count(*)
  FROM evidence
 WHERE object_blob_id = $1
`, blobID).Scan(&associationCount); err != nil {
		t.Fatalf("count blob associations: %v", err)
	}
	if associationCount != 1 {
		t.Fatalf("association count = %d, want 1", associationCount)
	}
	requireChangeSetCount(t, harness, secondRecordID, 0)
}

func TestBlobAssociation_ConcurrentRaceHasOneWinner(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "evidence-blob-association-race")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatalf("open concurrent association pool: %v", err)
	}
	t.Cleanup(pool.Close)
	var harness postgres.DB = pool
	store := newTestBlobLifecycleService(t, harness)
	actor := authstoretest.SeedActor(t, harness, authstoretest.ActorAttributes{Email: "evidence-association-race@example.test", DisplayName: "Evidence Association Race"})
	incident := createTestIncident(t, harness, actor, "txn-evidence-association-race-incident", "IR-EVIDENCE-ASSOCIATION-RACE", "Evidence association race")
	recordIDs := []uuid.UUID{
		seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received"),
		seedEvidenceAttachmentRecord(t, harness, incident.ID, actor.ID, "received"),
	}
	blobID := seedBlob(t, harness, incident.ID, actor.ID, "available", BlobOptions{
		ByteSize:            4,
		ObservedSize:        ptrInt64(4),
		ObservedSHA:         ptrString(strings.Repeat("a", 64)),
		ObservedContentType: ptrString("text/plain"),
	})
	start := make(chan struct{})
	results := make(chan error, len(recordIDs))
	for index, recordID := range recordIDs {
		go func(index int, recordID uuid.UUID) {
			<-start
			request := attachBlobRequest{
				ObjectBlobID:   blobID,
				BaseRowVersion: 1,
				ClientTxnID:    "txn-association-race-" + string(rune('1'+index)),
			}
			_, err := store.AttachBlob(context.Background(), actor, recordID, request, attachBlobRequestHash(request), "req-association-race", time.Now().UTC())
			results <- err
		}(index, recordID)
	}
	close(start)

	successes := 0
	rejections := 0
	for range recordIDs {
		err := <-results
		if err == nil {
			successes++
			continue
		}
		var rejected AttachRejectedError
		if errors.As(err, &rejected) && rejected.ReasonCode == AttachReasonBlobNotVisible {
			rejections++
			continue
		}
		t.Fatalf("race result error = %v, want success or concealed rejection", err)
	}
	if successes != 1 || rejections != 1 {
		t.Fatalf("race results successes=%d rejections=%d, want 1 and 1", successes, rejections)
	}
	var associations int
	if err := harness.QueryRow(context.Background(), `SELECT count(*) FROM evidence WHERE object_blob_id = $1`, blobID).Scan(&associations); err != nil {
		t.Fatalf("count race associations: %v", err)
	}
	if associations != 1 {
		t.Fatalf("race associations = %d, want 1", associations)
	}
}

type DurableCounts struct {
	ChangeSets int
	Mutations  int
	Revisions  int
	BlobLinks  int
}

func DurableAttachCounts(t testing.TB, db postgres.DB, recordID uuid.UUID) DurableCounts {
	t.Helper()
	var counts DurableCounts
	if err := db.QueryRow(context.Background(), `
SELECT
    (SELECT COUNT(*) FROM change_sets cs JOIN record_revisions rr ON rr.change_set_id = cs.change_set_id WHERE rr.record_id = $1),
    (SELECT COUNT(*) FROM change_set_mutations m JOIN record_revisions rr ON rr.change_set_id = m.change_set_id WHERE rr.record_id = $1),
    (SELECT COUNT(*) FROM record_revisions WHERE record_id = $1),
    (SELECT COUNT(*) FROM evidence WHERE record_id = $1 AND object_blob_id IS NOT NULL)
`, recordID).Scan(&counts.ChangeSets, &counts.Mutations, &counts.Revisions, &counts.BlobLinks); err != nil {
		t.Fatalf("count durable attach state: %v", err)
	}
	return counts
}

func ptrString(value string) *string { return &value }

func ptrInt64(value int64) *int64 { return &value }
