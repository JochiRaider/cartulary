package evidence_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/phase4storetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestPhase5_AttachBlobValidation_U_5_03(t *testing.T) {
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
				_, apiErr := evidence.DecodeAttachBlobRequest(strings.NewReader(tc.body))
				if apiErr == nil || apiErr.Status != http.StatusBadRequest || apiErr.Code != "invalid_mutation_payload" {
					t.Fatalf("expected invalid_mutation_payload, got %#v", apiErr)
				}
			})
		}
	})

	harness := phase4storetest.StartStore(t, "phase5-attach-validation")
	store := evidence.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "phase5-attach@example.test", "Phase5 Attach", "Phase5Attach1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase5-attach-incident", "IR-P5-ATTACH", "Phase 5 attach")
	otherIncident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase5-attach-other", "IR-P5-ATTACH-OTHER", "Phase 5 attach other")

	t.Run("success and replay are stable", func(t *testing.T) {
		recordID := seedPhase5EvidenceRecord(t, harness.DB, incident.ID, actor.ID, "received")
		blobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "available", phase5BlobOptions{
			ByteSize: 4, ObservedSize: ptrInt64(4), ObservedSHA: ptrString(strings.Repeat("a", 64)), ObservedContentType: ptrString("text/plain"),
		})
		request := evidence.AttachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-attach-success"}
		result, err := store.AttachBlob(context.Background(), actor, recordID, request, evidence.AttachBlobRequestHash(request), nil, "req-success", time.Now().UTC())
		if err != nil {
			t.Fatalf("attach available blob: %v", err)
		}
		if result.StatusCode != http.StatusOK || result.Replayed {
			t.Fatalf("unexpected attach result: %#v", result)
		}
		requirePhase5EvidenceState(t, harness.DB, recordID, "available", "available", blobID)
		changeSet := result.Payload["change_set_id"]

		replay, err := store.AttachBlob(context.Background(), actor, recordID, request, evidence.AttachBlobRequestHash(request), nil, "req-replay", time.Now().UTC())
		if err != nil {
			t.Fatalf("replay attach: %v", err)
		}
		if !replay.Replayed || replay.StatusCode != http.StatusOK || replay.Payload["change_set_id"] != changeSet {
			t.Fatalf("replay did not return original payload: %#v want change_set_id %#v", replay, changeSet)
		}
		requirePhase5ChangeSetCount(t, harness.DB, recordID, 1)
	})

	t.Run("divergent replay is rejected without durable side effects", func(t *testing.T) {
		recordID := seedPhase5EvidenceRecord(t, harness.DB, incident.ID, actor.ID, "received")
		blobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "available", phase5BlobOptions{
			ByteSize: 4, ObservedSize: ptrInt64(4), ObservedSHA: ptrString(strings.Repeat("a", 64)), ObservedContentType: ptrString("text/plain"),
		})
		request := evidence.AttachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-attach-divergent"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, request, evidence.AttachBlobRequestHash(request), nil, "req-divergent-first", time.Now().UTC()); err != nil {
			t.Fatalf("attach available blob: %v", err)
		}
		before := phase5DurableAttachCounts(t, harness.DB, recordID)

		otherBlobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "available", phase5BlobOptions{
			ByteSize: 4, ObservedSize: ptrInt64(4), ObservedSHA: ptrString(strings.Repeat("b", 64)), ObservedContentType: ptrString("text/plain"),
		})
		divergent := evidence.AttachBlobRequest{ObjectBlobID: otherBlobID, BaseRowVersion: 2, ClientTxnID: request.ClientTxnID}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, divergent, evidence.AttachBlobRequestHash(divergent), nil, "req-divergent-second", time.Now().UTC()); !errors.Is(err, authn.ErrClientTxnConflict) {
			t.Fatalf("divergent replay got %v want authn.ErrClientTxnConflict", err)
		}
		after := phase5DurableAttachCounts(t, harness.DB, recordID)
		if after != before {
			t.Fatalf("divergent replay changed durable counts: before=%+v after=%+v", before, after)
		}
		requirePhase5EvidenceState(t, harness.DB, recordID, "available", "available", blobID)
	})

	t.Run("conflict and lifecycle failures leave evidence unchanged", func(t *testing.T) {
		recordID := seedPhase5EvidenceRecord(t, harness.DB, incident.ID, actor.ID, "received")
		blobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "available", phase5BlobOptions{ByteSize: 1, ObservedSize: ptrInt64(1)})
		request := evidence.AttachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 2, ClientTxnID: "txn-row-conflict"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, request, evidence.AttachBlobRequestHash(request), nil, "req-conflict", time.Now().UTC()); !errors.Is(err, evidence.ErrRowVersionConflict) {
			t.Fatalf("row-version conflict got %v", err)
		}
		requirePhase5EvidenceState(t, harness.DB, recordID, "received", "pending", uuid.Nil)

		missing := evidence.AttachBlobRequest{ObjectBlobID: uuid.New(), BaseRowVersion: 1, ClientTxnID: "txn-missing"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, missing, evidence.AttachBlobRequestHash(missing), nil, "req-missing", time.Now().UTC()); !errors.Is(err, evidence.ErrBlobNotFound) {
			t.Fatalf("missing blob got %v", err)
		}
		requirePhase5EvidenceState(t, harness.DB, recordID, "received", "pending", uuid.Nil)

		foreignBlob := seedPhase5Blob(t, harness.DB, otherIncident.ID, actor.ID, "available", phase5BlobOptions{ByteSize: 1, ObservedSize: ptrInt64(1)})
		foreign := evidence.AttachBlobRequest{ObjectBlobID: foreignBlob, BaseRowVersion: 1, ClientTxnID: "txn-foreign"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, foreign, evidence.AttachBlobRequestHash(foreign), nil, "req-foreign", time.Now().UTC()); !errors.Is(err, evidence.ErrIncidentMismatch) {
			t.Fatalf("incident mismatch got %v", err)
		} else {
			requirePhase5AttachRejectedReason(t, err, evidence.AttachReasonBlobNotVisible)
		}
		requirePhase5EvidenceState(t, harness.DB, recordID, "received", "pending", uuid.Nil)

		for _, state := range []string{"failed", "quarantined"} {
			blockedBlob := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, state, phase5BlobOptions{ByteSize: 1})
			blocked := evidence.AttachBlobRequest{ObjectBlobID: blockedBlob, BaseRowVersion: 1, ClientTxnID: "txn-" + state}
			if _, err := store.AttachBlob(context.Background(), actor, recordID, blocked, evidence.AttachBlobRequestHash(blocked), nil, "req-"+state, time.Now().UTC()); !errors.Is(err, evidence.ErrBlobNotAttachable) {
				t.Fatalf("%s blob got %v", state, err)
			} else {
				wantReason := evidence.AttachReasonBlobFailed
				if state == "quarantined" {
					wantReason = evidence.AttachReasonBlobQuarantined
				}
				requirePhase5AttachRejectedReason(t, err, wantReason)
			}
			requirePhase5EvidenceState(t, harness.DB, recordID, "received", "pending", uuid.Nil)
		}
	})

	t.Run("terminal finalization failures persist on blob only", func(t *testing.T) {
		recordID := seedPhase5EvidenceRecord(t, harness.DB, incident.ID, actor.ID, "received")
		expiredBlob := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "pending", phase5BlobOptions{ByteSize: 1, PendingExpiresAt: time.Now().UTC().Add(-time.Minute)})
		expired := evidence.AttachBlobRequest{ObjectBlobID: expiredBlob, BaseRowVersion: 1, ClientTxnID: "txn-expired"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, expired, evidence.AttachBlobRequestHash(expired), nil, "req-expired", time.Now().UTC()); !errors.Is(err, evidence.ErrBlobNotAttachable) {
			t.Fatalf("expired blob got %v", err)
		} else {
			requirePhase5AttachRejectedReason(t, err, evidence.AttachReasonBlobFailed)
		}
		requirePhase5BlobFailure(t, harness.DB, expiredBlob, "pending_timeout", 0)

		sizeBlob := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "pending", phase5BlobOptions{ByteSize: 5})
		sizeMismatch := evidence.AttachBlobRequest{ObjectBlobID: sizeBlob, BaseRowVersion: 1, ClientTxnID: "txn-size"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, sizeMismatch, evidence.AttachBlobRequestHash(sizeMismatch), &evidence.ObservedObject{Size: 4, ContentType: "text/plain", SHA256Hex: "size"}, "req-size", time.Now().UTC()); !errors.Is(err, evidence.ErrBlobNotAttachable) {
			t.Fatalf("size mismatch got %v", err)
		} else {
			requirePhase5AttachRejectedReason(t, err, evidence.AttachReasonAcceptedContractMismatch)
		}
		requirePhase5BlobFailure(t, harness.DB, sizeBlob, "declared_size_mismatch", 0)

		hashBlob := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "pending", phase5BlobOptions{ByteSize: 4, ExpectedSHA: ptrString(strings.Repeat("e", 64))})
		hashMismatch := evidence.AttachBlobRequest{ObjectBlobID: hashBlob, BaseRowVersion: 1, ClientTxnID: "txn-hash"}
		if _, err := store.AttachBlob(context.Background(), actor, recordID, hashMismatch, evidence.AttachBlobRequestHash(hashMismatch), &evidence.ObservedObject{Size: 4, ContentType: "text/plain", SHA256Hex: strings.Repeat("a", 64)}, "req-hash", time.Now().UTC()); !errors.Is(err, evidence.ErrBlobNotAttachable) {
			t.Fatalf("hash mismatch got %v", err)
		} else {
			requirePhase5AttachRejectedReason(t, err, evidence.AttachReasonAcceptedContractMismatch)
		}
		requirePhase5BlobFailure(t, harness.DB, hashBlob, "expected_sha256_mismatch", 0)
		requirePhase5EvidenceState(t, harness.DB, recordID, "received", "pending", uuid.Nil)
	})

	t.Run("non-terminal finalization retry budget fails on fourth attempt", func(t *testing.T) {
		recordID := seedPhase5EvidenceRecord(t, harness.DB, incident.ID, actor.ID, "received")
		blobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "pending", phase5BlobOptions{ByteSize: 1})
		for attempt := 1; attempt <= 4; attempt++ {
			request := evidence.AttachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-retry"}
			if _, err := store.AttachBlob(context.Background(), actor, recordID, request, evidence.AttachBlobRequestHash(request), nil, "req-retry", time.Now().UTC()); !errors.Is(err, evidence.ErrBlobNotAttachable) {
				t.Fatalf("attempt %d got %v", attempt, err)
			} else if attempt < 4 {
				requirePhase5AttachRejectedReason(t, err, evidence.AttachReasonBlobPending)
			} else {
				requirePhase5AttachRejectedReason(t, err, evidence.AttachReasonBlobFailed)
			}
			if attempt < 4 {
				requirePhase5PendingAttemptCount(t, harness.DB, blobID, attempt)
			}
		}
		requirePhase5BlobFailure(t, harness.DB, blobID, "finalize_retry_exhausted", 4)
		requirePhase5EvidenceState(t, harness.DB, recordID, "received", "pending", uuid.Nil)
	})
}

type phase5BlobOptions struct {
	ByteSize            int64
	ExpectedSHA         *string
	ObservedSize        *int64
	ObservedContentType *string
	ObservedSHA         *string
	PendingExpiresAt    time.Time
}

func seedPhase5EvidenceRecord(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorID uuid.UUID, lifecycleState string) uuid.UUID {
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
VALUES ($1, $2, 'Phase 5 evidence', $3, 'pending', $4, $4, $4)
`, recordID, incidentID, lifecycleState, now); err != nil {
		t.Fatalf("insert evidence row: %v", err)
	}
	return recordID
}

func seedPhase5Blob(t testing.TB, db postgres.DB, incidentID uuid.UUID, actorID uuid.UUID, uploadState string, options phase5BlobOptions) uuid.UUID {
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
`, blobID, incidentID, actorID, "phase5/attach/"+blobID.String(), uploadState,
		byteSize, options.ExpectedSHA, options.ObservedSize, options.ObservedContentType, options.ObservedSHA,
		now.Add(time.Hour), pendingExpiresAt, finalizedAt, terminalReason, failedAt, cleanupDueAt, now); err != nil {
		t.Fatalf("insert object blob: %v", err)
	}
	return blobID
}

func requirePhase5EvidenceState(t testing.TB, db postgres.DB, recordID uuid.UUID, wantLifecycle string, wantUpload string, wantBlobID uuid.UUID) {
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

func requirePhase5BlobFailure(t testing.TB, db postgres.DB, objectBlobID uuid.UUID, wantReason string, wantAttempts int) {
	t.Helper()
	var uploadState string
	var terminalReason string
	var failedAt *time.Time
	var cleanupDueAt *time.Time
	var attempts int
	if err := db.QueryRow(context.Background(), `
SELECT upload_state, terminal_reason, failed_at, cleanup_due_at, finalize_attempt_count
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&uploadState, &terminalReason, &failedAt, &cleanupDueAt, &attempts); err != nil {
		t.Fatalf("load blob failure: %v", err)
	}
	if uploadState != "failed" || terminalReason != wantReason || failedAt == nil || cleanupDueAt == nil || attempts != wantAttempts {
		t.Fatalf("blob failure got state=%s reason=%s failed_at=%v cleanup_due_at=%v attempts=%d want reason=%s attempts=%d", uploadState, terminalReason, failedAt, cleanupDueAt, attempts, wantReason, wantAttempts)
	}
}

func requirePhase5PendingAttemptCount(t testing.TB, db postgres.DB, objectBlobID uuid.UUID, wantAttempts int) {
	t.Helper()
	var uploadState string
	var terminalReason *string
	var attempts int
	if err := db.QueryRow(context.Background(), `
SELECT upload_state, terminal_reason, finalize_attempt_count
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&uploadState, &terminalReason, &attempts); err != nil {
		t.Fatalf("load blob retry count: %v", err)
	}
	if uploadState != "pending" || terminalReason != nil || attempts != wantAttempts {
		t.Fatalf("pending retry got state=%s reason=%v attempts=%d want attempts=%d", uploadState, terminalReason, attempts, wantAttempts)
	}
}

func requirePhase5AttachRejectedReason(t testing.TB, err error, want string) {
	t.Helper()
	var rejected evidence.AttachRejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf("attach error %v does not expose AttachRejectedError", err)
	}
	if rejected.ReasonCode != want {
		t.Fatalf("attach rejected reason got %s want %s", rejected.ReasonCode, want)
	}
}

func requirePhase5ChangeSetCount(t testing.TB, db postgres.DB, recordID uuid.UUID, want int) {
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

type phase5DurableCounts struct {
	ChangeSets int
	Mutations  int
	Revisions  int
	BlobLinks  int
}

func phase5DurableAttachCounts(t testing.TB, db postgres.DB, recordID uuid.UUID) phase5DurableCounts {
	t.Helper()
	var counts phase5DurableCounts
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
