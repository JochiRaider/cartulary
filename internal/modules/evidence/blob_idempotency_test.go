package evidence

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestBlobCreateIdempotency_Unit(t *testing.T) {
	harness := newTestStore(t, "evidence_lifecycle-blob-idempotency")
	store := newTestBlobLifecycleService(t, harness)
	actorA := authstoretest.SeedLocalUserRecord(t, harness, "evidence_lifecycle-blob-actor-a@example.test", "EvidenceLifecycle Blob Actor A", "EvidenceLifecycleBlobActorA1!", false, false, true)
	actorB := authstoretest.SeedLocalUserRecord(t, harness, "evidence_lifecycle-blob-actor-b@example.test", "EvidenceLifecycle Blob Actor B", "EvidenceLifecycleBlobActorB1!", false, false, true)
	seedBlobCreateTestSession(t, harness, actorA.ID)
	seedBlobCreateTestSession(t, harness, actorB.ID)
	incidentA := createTestIncident(t, harness, actorA, "txn-evidence_lifecycle-blob-incident-a", "IR-P5-BLOB-A", "Evidence blob incident A")
	incidentB := createTestIncident(t, harness, actorA, "txn-evidence_lifecycle-blob-incident-b", "IR-P5-BLOB-B", "Evidence blob incident B")

	baseRequest := mustBlobCreateRequest(t, incidentA.ID, "txn-shared-blob", 12, " proof.bin ", " application/octet-stream ", nil)
	first := createBlobSlot(t, store, baseRequest, actorA.ID, incidentA.ID, uuid.New(), "slot-a-first")
	firstData := first.Payload
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first create status got %d want %d", first.StatusCode, http.StatusCreated)
	}

	replay := createBlobSlot(t, store, baseRequest, actorA.ID, incidentA.ID, uuid.New(), "slot-a-replay")
	if replay.StatusCode != http.StatusOK || !replay.Replayed {
		t.Fatalf("replay got status=%d replayed=%v", replay.StatusCode, replay.Replayed)
	}
	requireStableBlobPayload(t, replay.Payload, firstData)
	if got := countObjectBlobsInStore(t, harness, incidentA.ID); got != 1 {
		t.Fatalf("exact replay created extra object_blobs: got %d want 1", got)
	}

	divergent := mustBlobCreateRequest(t, incidentA.ID, "txn-shared-blob", 13, "proof.bin", "application/octet-stream", nil)
	if _, err := store.CreateBlobSlot(context.Background(), testBlobSlotParams(divergent, actorA.ID, incidentA.ID, uuid.New(), "slot-a-divergent")); !errors.Is(err, ErrClientTxnConflict) {
		t.Fatalf("divergent replay error got %v want client txn conflict", err)
	}
	if got := countObjectBlobsInStore(t, harness, incidentA.ID); got != 1 {
		t.Fatalf("divergent replay created object_blobs: got %d want 1", got)
	}

	otherIncidentRequest := mustBlobCreateRequest(t, incidentB.ID, "txn-shared-blob", 12, "proof.bin", "application/octet-stream", nil)
	otherIncident := createBlobSlot(t, store, otherIncidentRequest, actorA.ID, incidentB.ID, uuid.New(), "slot-b-first")
	if otherIncident.StatusCode != http.StatusCreated {
		t.Fatalf("same actor/client_txn in different incident status got %d want %d", otherIncident.StatusCode, http.StatusCreated)
	}
	if otherIncident.Payload["object_blob_id"] == firstData["object_blob_id"] {
		t.Fatalf("different incident reused object_blob_id: %#v", otherIncident.Payload)
	}

	otherActor := createBlobSlot(t, store, baseRequest, actorB.ID, incidentA.ID, uuid.New(), "slot-a-actor-b")
	if otherActor.StatusCode != http.StatusCreated {
		t.Fatalf("different actor same incident/client_txn status got %d want %d", otherActor.StatusCode, http.StatusCreated)
	}
	if otherActor.Payload["object_blob_id"] == firstData["object_blob_id"] {
		t.Fatalf("different actor reused object_blob_id: %#v", otherActor.Payload)
	}
	if got := countObjectBlobsInStore(t, harness, incidentA.ID); got != 2 {
		t.Fatalf("actor-scoped idempotency object_blobs in incident A got %d want 2", got)
	}
}

func mustBlobCreateRequest(t testing.TB, incidentID uuid.UUID, clientTxnID string, byteSize int64, filenameHint string, contentTypeHint string, sha256Hex *string) blobCreateRequest {
	t.Helper()
	body := fmt.Sprintf(`{"incident_id":%q,"client_txn_id":%q,"byte_size":%d,"filename_hint":%q,"content_type_hint":%q`,
		incidentID.String(), clientTxnID, byteSize, filenameHint, contentTypeHint)
	if sha256Hex != nil {
		body += fmt.Sprintf(`,"sha256_hex":%q`, *sha256Hex)
	}
	body += `}`
	request, apiErr := decodeBlobCreateRequest(strings.NewReader(body), 536870912)
	if apiErr != nil {
		t.Fatalf("decode blob create request: %v details=%#v", apiErr, apiErr.Details)
	}
	return request
}

func createBlobSlot(t testing.TB, store *blobLifecycleService, request blobCreateRequest, actorID uuid.UUID, incidentID uuid.UUID, objectBlobID uuid.UUID, label string) blobSlotResult {
	t.Helper()
	result, err := store.CreateBlobSlot(context.Background(), testBlobSlotParams(request, actorID, incidentID, objectBlobID, label))
	if err != nil {
		t.Fatalf("create blob slot %s: %v", label, err)
	}
	return result
}

func testBlobSlotParams(request blobCreateRequest, actorID uuid.UUID, incidentID uuid.UUID, objectBlobID uuid.UUID, label string) blobSlotParams {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	capabilityHash := sha256.Sum256([]byte("capability:" + label))
	contractHash := sha256.Sum256([]byte("contract:" + label))
	return blobSlotParams{
		ObjectBlobID:      objectBlobID,
		IncidentID:        incidentID,
		ActorUserID:       actorID,
		StorageKey:        "incidents/" + incidentID.String() + "/object-blobs/" + objectBlobID.String(),
		ByteSize:          request.ByteSize,
		FilenameHint:      request.FilenameHint,
		ContentTypeHint:   request.ContentTypeHint,
		ExpectedSHA256Hex: request.SHA256Hex,
		TargetExpiresAt:   now.Add(60 * time.Minute),
		PendingExpiresAt:  now.Add(24 * time.Hour),
		UploadTarget: map[string]any{
			"href":       "https://uploads.example.test/" + label,
			"method":     "PUT",
			"expires_at": now.Add(60 * time.Minute).Format(time.RFC3339Nano),
			"headers":    map[string]string{},
		},
		AcceptedContract: request.AcceptedContract,
		RequestHash:      blobCreateRequestHash(request),
		ClientTxnID:      request.ClientTxnID,
		UploadLease: uploadLeaseCreateParams{
			LeaseID: uuid.New(), CapabilityHash: capabilityHash[:],
			IssuingUserID: actorID, IssuingSessionID: actorID,
			IssuedAt: now, ExpiresAt: now.Add(60 * time.Minute), RequiredMethod: http.MethodPut,
			RequiredHeaders: map[string]string{}, AcceptedContractSHA256: contractHash[:],
		},
	}
}

func seedBlobCreateTestSession(t testing.TB, harness postgres.DB, userID uuid.UUID) {
	t.Helper()
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	fingerprint := sha256.Sum256([]byte("session:" + userID.String()))
	if _, err := harness.Exec(context.Background(), `
INSERT INTO user_sessions (
    id, user_id, token_fingerprint, authenticated_at,
    last_qualifying_activity_at, idle_expires_at, absolute_expires_at,
    session_expires_at, created_at, updated_at
) VALUES ($1, $1, $2, $3, $3, $4, $4, $4, $3, $3)
`, userID, fingerprint[:], now, now.Add(time.Hour)); err != nil {
		t.Fatalf("seed blob create test session: %v", err)
	}
}

func requireStableBlobPayload(t testing.TB, got map[string]any, want map[string]any) {
	t.Helper()
	for _, key := range []string{"incident_id", "object_blob_id", "upload_state", "target_expires_at", "pending_expires_at"} {
		if got[key] != want[key] {
			t.Fatalf("payload[%s] got %#v want %#v; got=%#v want=%#v", key, got[key], want[key], got, want)
		}
	}
	gotTarget := got["upload_target"].(map[string]any)
	wantTarget := want["upload_target"].(map[string]any)
	if gotTarget["href"] != wantTarget["href"] || gotTarget["expires_at"] != wantTarget["expires_at"] {
		t.Fatalf("upload target changed on replay: got=%#v want=%#v", gotTarget, wantTarget)
	}
	gotAccepted := got["accepted_contract"].(map[string]any)
	wantAccepted := want["accepted_contract"].(map[string]any)
	for _, key := range []string{"incident_id", "filename_hint", "content_type_hint", "sha256_hex"} {
		if gotAccepted[key] != wantAccepted[key] {
			t.Fatalf("accepted_contract[%s] changed: got=%#v want=%#v", key, gotAccepted[key], wantAccepted[key])
		}
	}
}

func countObjectBlobsInStore(t testing.TB, harness postgres.DB, incidentID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.QueryRow(context.Background(), `SELECT count(*) FROM object_blobs WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		t.Fatalf("count object_blobs: %v", err)
	}
	return count
}
