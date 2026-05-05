package evidence_test

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

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestPhase5_ObjectBlobCreate_U_5_01(t *testing.T) {
	incidentID := uuid.New()
	validSHA := fmt.Sprintf("%x", sha256.Sum256([]byte("phase5")))

	for _, tc := range []struct {
		name       string
		body       string
		field      string
		reasonCode string
	}{
		{
			name:       "array body",
			body:       `[]`,
			reasonCode: "request_not_object",
		},
		{
			name:       "null body",
			body:       `null`,
			reasonCode: "request_not_object",
		},
		{
			name:       "multiple JSON values",
			body:       fmt.Sprintf(`{"incident_id":%q,"client_txn_id":"txn","byte_size":1} {}`, incidentID.String()),
			reasonCode: "request_not_object",
		},
		{
			name:       "missing incident_id",
			body:       `{"client_txn_id":"txn","byte_size":1}`,
			field:      "incident_id",
			reasonCode: "missing_required_field",
		},
		{
			name:       "null client_txn_id",
			body:       fmt.Sprintf(`{"incident_id":%q,"client_txn_id":null,"byte_size":1}`, incidentID.String()),
			field:      "client_txn_id",
			reasonCode: "field_not_nullable",
		},
		{
			name:       "empty client_txn_id after normalization",
			body:       fmt.Sprintf(`{"incident_id":%q,"client_txn_id":" \t ","byte_size":1}`, incidentID.String()),
			field:      "client_txn_id",
			reasonCode: "field_empty_after_normalization",
		},
		{
			name:       "unknown field",
			body:       fmt.Sprintf(`{"incident_id":%q,"client_txn_id":"txn","byte_size":1,"unexpected":true}`, incidentID.String()),
			field:      "unexpected",
			reasonCode: "unknown_field",
		},
		{
			name:       "server managed field",
			body:       fmt.Sprintf(`{"incident_id":%q,"client_txn_id":"txn","byte_size":1,"object_blob_id":%q}`, incidentID.String(), uuid.NewString()),
			field:      "object_blob_id",
			reasonCode: "server_managed_field",
		},
		{
			name:       "non integer byte_size",
			body:       fmt.Sprintf(`{"incident_id":%q,"client_txn_id":"txn","byte_size":1.5}`, incidentID.String()),
			field:      "byte_size",
			reasonCode: "invalid_byte_size",
		},
		{
			name:       "negative byte_size",
			body:       fmt.Sprintf(`{"incident_id":%q,"client_txn_id":"txn","byte_size":-1}`, incidentID.String()),
			field:      "byte_size",
			reasonCode: "invalid_byte_size",
		},
		{
			name:       "invalid sha256_hex",
			body:       fmt.Sprintf(`{"incident_id":%q,"client_txn_id":"txn","byte_size":1,"sha256_hex":"%s"}`, incidentID.String(), strings.ToUpper(validSHA)),
			field:      "sha256_hex",
			reasonCode: "invalid_sha256_hex",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := evidence.DecodeBlobCreateRequest(strings.NewReader(tc.body), 10)
			requireBlobCreateReason(t, apiErr, tc.field, tc.reasonCode)
		})
	}

	request, apiErr := evidence.DecodeBlobCreateRequest(strings.NewReader(fmt.Sprintf(
		`{"incident_id":%q,"client_txn_id":" txn-normalized ","byte_size":42,"filename_hint":"  proof.txt  ","content_type_hint":" text/plain ","sha256_hex":%q}`,
		incidentID.String(),
		validSHA,
	)), 100)
	if apiErr != nil {
		t.Fatalf("decode valid blob create request: %v", apiErr)
	}
	if request.ClientTxnID != "txn-normalized" {
		t.Fatalf("client_txn_id got %q want txn-normalized", request.ClientTxnID)
	}
	requireAcceptedContract(t, request.AcceptedContract, map[string]any{
		"incident_id":       incidentID.String(),
		"byte_size":         int64(42),
		"filename_hint":     "proof.txt",
		"content_type_hint": "text/plain",
		"sha256_hex":        validSHA,
	})

	nullableRequest, apiErr := evidence.DecodeBlobCreateRequest(strings.NewReader(fmt.Sprintf(
		`{"incident_id":%q,"client_txn_id":"txn-nullable","byte_size":0,"filename_hint":null,"content_type_hint":"   ","sha256_hex":null}`,
		incidentID.String(),
	)), 100)
	if apiErr != nil {
		t.Fatalf("decode nullable blob create request: %v", apiErr)
	}
	requireAcceptedContract(t, nullableRequest.AcceptedContract, map[string]any{
		"incident_id":       incidentID.String(),
		"byte_size":         int64(0),
		"filename_hint":     nil,
		"content_type_hint": nil,
		"sha256_hex":        nil,
	})
}

func TestPhase5_BlobCreateIdempotency_U_5_02(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase5-blob-idempotency")
	store := evidence.NewStore(harness.DB)
	actorA := phase4storetest.SeedLocalUserFlags(t, harness.DB, "phase5-blob-actor-a@example.test", "Phase5 Blob Actor A", "Phase5BlobActorA1!", false, false, true)
	actorB := phase4storetest.SeedLocalUserFlags(t, harness.DB, "phase5-blob-actor-b@example.test", "Phase5 Blob Actor B", "Phase5BlobActorB1!", false, false, true)
	incidentA := phase4storetest.CreateIncidentInStore(t, harness.DB, actorA, "txn-phase5-blob-incident-a", "IR-P5-BLOB-A", "Phase 5 blob incident A")
	incidentB := phase4storetest.CreateIncidentInStore(t, harness.DB, actorA, "txn-phase5-blob-incident-b", "IR-P5-BLOB-B", "Phase 5 blob incident B")

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
	if _, err := store.CreateBlobSlot(context.Background(), blobSlotParams(divergent, actorA.ID, incidentA.ID, uuid.New(), "slot-a-divergent")); !errors.Is(err, authn.ErrClientTxnConflict) {
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

func TestPhase5_BlobCreateSizeCeiling_U_5_09(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-blob-size-ceiling")
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-size-incident",
		"incident_key":  "phase5-size",
		"title":         "Phase 5 size ceiling",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	maxCreate := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":   incidentID.String(),
		"client_txn_id": "txn-phase5-size-max",
		"byte_size":     int64(536870912),
	}, authOptions(login)...)
	maxData := httptestx.RequireSuccessEnvelope(t, maxCreate, http.StatusCreated)["data"].(map[string]any)
	if maxData["upload_state"] != "pending" || maxData["upload_target"] == nil || maxData["object_blob_id"] == nil {
		t.Fatalf("max-size create missing pending slot payload: %#v", maxData)
	}

	beforeRejected := countObjectBlobs(t, harness, incidentID)
	rejected := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":   incidentID.String(),
		"client_txn_id": "txn-phase5-size-too-large",
		"byte_size":     int64(536870913),
	}, authOptions(login)...)
	rejectedBody := httptestx.RequireErrorEnvelope(t, rejected, http.StatusRequestEntityTooLarge, "blob_create_rejected")
	details := rejectedBody["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "byte_size_exceeds_limit" {
		t.Fatalf("reason_code got %#v want byte_size_exceeds_limit", details["reason_code"])
	}
	if details["requested_byte_size"] != float64(536870913) || details["configured_limit_bytes"] != float64(536870912) {
		t.Fatalf("unexpected size details: %#v", details)
	}
	if _, ok := details["object_blob_id"]; ok {
		t.Fatalf("rejection exposed object_blob_id: %#v", details)
	}
	if got := countObjectBlobs(t, harness, incidentID); got != beforeRejected {
		t.Fatalf("oversize request created object_blobs: got %d want %d", got, beforeRejected)
	}
	if got := countBlobCreateIdempotency(t, harness, adminID, incidentID, "txn-phase5-size-too-large"); got != 0 {
		t.Fatalf("oversize request created idempotency success payload: got %d want 0", got)
	}
}

func TestPhase5_PreviewPayloadCeiling_U_5_09(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-preview-size")
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-preview-size-incident",
		"incident_key":  "phase5-preview-size",
		"title":         "Phase 5 preview size",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	atLimitRecordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, atLimitRecordID)
	atLimitBlob := attachUploadedBlobWithMetadata(t, harness, login, incidentID, atLimitRecordID, []byte("image at limit"), "limit.png", "image/png", "txn-phase5-preview-limit-blob", "txn-phase5-preview-limit-attach")
	updateBlobObservedMetadata(t, harness, phase4test.MustUUID(t, atLimitBlob["object_blob_id"].(string)), int64(33554432), "image/png", "")
	preview := issueEvidenceHandle(t, harness, login, atLimitRecordID, "preview-handle")
	if preview["handle_kind"] != "preview" || preview["preview_kind"] != "image_inline" {
		t.Fatalf("preview at limit failed to issue image_inline handle: %#v", preview)
	}

	oversizeRecordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, oversizeRecordID)
	oversizeBlob := attachUploadedBlobWithMetadata(t, harness, login, incidentID, oversizeRecordID, []byte("image too large"), "oversize.png", "image/png", "txn-phase5-preview-oversize-blob", "txn-phase5-preview-oversize-attach")
	updateBlobObservedMetadata(t, harness, phase4test.MustUUID(t, oversizeBlob["object_blob_id"].(string)), int64(33554433), "image/png", "")
	requireEvidenceAccessUnavailableReason(t,
		phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+oversizeRecordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...),
		"preview_payload_too_large",
	)
	download := issueEvidenceHandle(t, harness, login, oversizeRecordID, "download-handle")
	if download["handle_kind"] != "download" {
		t.Fatalf("oversize preview should not block download issuance: %#v", download)
	}
}

func requireBlobCreateReason(t testing.TB, apiErr *auth.APIError, field string, reasonCode string) {
	t.Helper()
	if apiErr == nil {
		t.Fatal("expected blob-create error, got nil")
		return
	}
	if apiErr.Code != "invalid_blob_create_request" || apiErr.Status != http.StatusBadRequest {
		t.Fatalf("unexpected blob-create error: status=%d code=%s details=%#v", apiErr.Status, apiErr.Code, apiErr.Details)
	}
	details := apiErr.Details
	if field != "" && details["field"] != field {
		t.Fatalf("field got %#v want %q in %#v", details["field"], field, details)
	}
	if reasonCode != "" && details["reason_code"] != reasonCode {
		t.Fatalf("reason_code got %#v want %q in %#v", details["reason_code"], reasonCode, details)
	}
}

func requireAcceptedContract(t testing.TB, got map[string]any, want map[string]any) {
	t.Helper()
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("accepted_contract[%s] got %#v want %#v in %#v", key, got[key], wantValue, got)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("accepted_contract got extra keys: %#v", got)
	}
}

func mustBlobCreateRequest(t testing.TB, incidentID uuid.UUID, clientTxnID string, byteSize int64, filenameHint string, contentTypeHint string, sha256Hex *string) evidence.BlobCreateRequest {
	t.Helper()
	body := fmt.Sprintf(`{"incident_id":%q,"client_txn_id":%q,"byte_size":%d,"filename_hint":%q,"content_type_hint":%q`,
		incidentID.String(), clientTxnID, byteSize, filenameHint, contentTypeHint)
	if sha256Hex != nil {
		body += fmt.Sprintf(`,"sha256_hex":%q`, *sha256Hex)
	}
	body += `}`
	request, apiErr := evidence.DecodeBlobCreateRequest(strings.NewReader(body), 536870912)
	if apiErr != nil {
		t.Fatalf("decode blob create request: %v details=%#v", apiErr, apiErr.Details)
	}
	return request
}

func createBlobSlot(t testing.TB, store *evidence.Store, request evidence.BlobCreateRequest, actorID uuid.UUID, incidentID uuid.UUID, objectBlobID uuid.UUID, label string) evidence.BlobSlotResult {
	t.Helper()
	result, err := store.CreateBlobSlot(context.Background(), blobSlotParams(request, actorID, incidentID, objectBlobID, label))
	if err != nil {
		t.Fatalf("create blob slot %s: %v", label, err)
	}
	return result
}

func blobSlotParams(request evidence.BlobCreateRequest, actorID uuid.UUID, incidentID uuid.UUID, objectBlobID uuid.UUID, label string) evidence.BlobSlotParams {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	return evidence.BlobSlotParams{
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
		RequestHash:      evidence.BlobCreateRequestHash(request),
		ClientTxnID:      request.ClientTxnID,
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

func countBlobCreateIdempotency(t testing.TB, harness *phase4test.ServerHarness, actorID uuid.UUID, incidentID uuid.UUID, clientTxnID string) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM route_idempotency
 WHERE route_key = 'object_blobs.create'
   AND actor_user_id = $1
   AND scope_key = $2
   AND client_txn_id = $3
`, actorID, incidentID.String(), clientTxnID).Scan(&count); err != nil {
		t.Fatalf("count blob-create idempotency: %v", err)
	}
	return count
}

func countObjectBlobsInStore(t testing.TB, harness *phase4storetest.StoreHarness, incidentID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRow(context.Background(), `SELECT count(*) FROM object_blobs WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		t.Fatalf("count object_blobs: %v", err)
	}
	return count
}
