package evidence_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	recordstoretest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestObjectBlobCreate_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence_lifecycle-blob-create-route")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-blob-create-incident",
		"incident_key":  "evidence_lifecycle-blob-create",
		"title":         "Evidence blob create",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	issuedAt := time.Now().UTC().Add(time.Minute).Truncate(time.Second)
	httptestx.SetClockFixed(t, harness.Server, issuedAt)
	validSHA := fmt.Sprintf("%x", sha256.Sum256([]byte("evidence_lifecycle")))

	createURL := harness.Server.HTTP.URL + "/api/v1/object-blobs"
	beforeInvalid := countObjectBlobs(t, harness, incidentID)
	for _, tc := range []struct {
		name       string
		body       string
		txn        string
		field      string
		reasonCode string
	}{
		{
			name:       "malformed body",
			body:       `{`,
			reasonCode: "request_not_object",
		},
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
			txn:        "txn",
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
			body:       fmt.Sprintf(`{"incident_id":%q,"client_txn_id":"txn-unknown","byte_size":1,"unexpected":true}`, incidentID.String()),
			txn:        "txn-unknown",
			field:      "unexpected",
			reasonCode: "unknown_field",
		},
		{
			name:       "server managed field",
			body:       fmt.Sprintf(`{"incident_id":%q,"client_txn_id":"txn-managed","byte_size":1,"object_blob_id":%q}`, incidentID.String(), uuid.NewString()),
			txn:        "txn-managed",
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
			resp := doRawJSON(t, http.MethodPost, createURL, tc.body, authOptions(login)...)
			requireBlobCreateHTTPReason(t, resp, tc.field, tc.reasonCode)
			if got := countObjectBlobs(t, harness, incidentID); got != beforeInvalid {
				t.Fatalf("invalid request wrote object_blobs: got %d want %d", got, beforeInvalid)
			}
			if tc.txn != "" {
				if got := countBlobCreateIdempotency(t, harness, adminID, incidentID, tc.txn); got != 0 {
					t.Fatalf("invalid request wrote idempotency state for %s: got %d want 0", tc.txn, got)
				}
			}
		})
	}

	unauthenticated := appsupport.DoJSON(t, http.MethodPost, createURL, map[string]any{
		"incident_id":   incidentID.String(),
		"client_txn_id": "txn-unauthenticated",
		"byte_size":     1,
	})
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")
	if got := countObjectBlobs(t, harness, incidentID); got != beforeInvalid {
		t.Fatalf("unauthenticated request wrote object_blobs: got %d want %d", got, beforeInvalid)
	}

	viewer := appsupport.SeedLocalUserFlags(t, harness.DB, "evidence_lifecycle-blob-viewer@example.test", "EvidenceLifecycle Blob Viewer", "EvidenceLifecycleBlobViewer1!", false, false, true)
	appsupport.SeedIncidentMembership(t, harness.DB, incidentID, viewer.ID, "EvidenceLifecycle Blob Viewer", "viewer", adminID)
	viewerLogin := loginLocalUserNoMFA(t, harness, "evidence_lifecycle-blob-viewer@example.test", "EvidenceLifecycleBlobViewer1!")
	denied := appsupport.DoJSON(t, http.MethodPost, createURL, map[string]any{
		"incident_id":   incidentID.String(),
		"client_txn_id": "txn-viewer-denied",
		"byte_size":     1,
	}, authOptions(viewerLogin)...)
	httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
	if got := countObjectBlobs(t, harness, incidentID); got != beforeInvalid {
		t.Fatalf("viewer-denied request wrote object_blobs: got %d want %d", got, beforeInvalid)
	}
	if got := countBlobCreateIdempotency(t, harness, viewer.ID, incidentID, "txn-viewer-denied"); got != 0 {
		t.Fatalf("viewer-denied request wrote idempotency state: got %d want 0", got)
	}

	createResp := appsupport.DoJSON(t, http.MethodPost, createURL, map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     " txn-normalized ",
		"byte_size":         42,
		"filename_hint":     "  proof.txt  ",
		"content_type_hint": " text/plain ",
		"sha256_hex":        validSHA,
	}, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	if createData["incident_id"] != incidentID.String() || createData["upload_state"] != "pending" || createData["object_blob_id"] == "" {
		t.Fatalf("unexpected blob create payload: %#v", createData)
	}
	requireCreateExpiry(t, createData, "target_expires_at", issuedAt.Add(60*time.Minute))
	requireCreateExpiry(t, createData, "pending_expires_at", issuedAt.Add(24*time.Hour))
	uploadTarget := createData["upload_target"].(map[string]any)
	href, _ := uploadTarget["href"].(string)
	if uploadTarget["method"] != "PUT" || !strings.HasPrefix(href, "/api/v1/object-uploads/upl_") {
		t.Fatalf("unexpected upload_target: %#v", uploadTarget)
	}
	if strings.Contains(href, "object-blobs/") || strings.Contains(href, "://") {
		t.Fatalf("upload_target href should be an opaque same-origin capability: %#v", uploadTarget)
	}
	requireAcceptedContract(t, createData["accepted_contract"].(map[string]any), map[string]any{
		"incident_id":       incidentID.String(),
		"byte_size":         float64(42),
		"filename_hint":     "proof.txt",
		"content_type_hint": "text/plain",
		"sha256_hex":        validSHA,
	})

	nullableResp := appsupport.DoJSON(t, http.MethodPost, createURL, map[string]any{
		"incident_id":   incidentID.String(),
		"client_txn_id": "txn-nullable",
		"byte_size":     0,
	}, authOptions(login)...)
	nullableData := httptestx.RequireSuccessEnvelope(t, nullableResp, http.StatusCreated)["data"].(map[string]any)
	requireAcceptedContract(t, nullableData["accepted_contract"].(map[string]any), map[string]any{
		"incident_id":       incidentID.String(),
		"byte_size":         float64(0),
		"filename_hint":     nil,
		"content_type_hint": nil,
		"sha256_hex":        nil,
	})
}

func TestBlobCreateIdempotency_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "evidence_lifecycle-blob-idempotency")
	revisionComposition := revisionsupport.MustComposition(t)
	store := evidence.NewStore(
		harness.DB,
		evidence.WithRevisionAppender(revisionComposition.Runtime.Appender()),
		evidence.WithCollaborationIntents(revisionComposition.Intents),
	)
	actorA := recordstoretest.SeedLocalUserFlags(t, harness.DB, "evidence_lifecycle-blob-actor-a@example.test", "EvidenceLifecycle Blob Actor A", "EvidenceLifecycleBlobActorA1!", false, false, true)
	actorB := recordstoretest.SeedLocalUserFlags(t, harness.DB, "evidence_lifecycle-blob-actor-b@example.test", "EvidenceLifecycle Blob Actor B", "EvidenceLifecycleBlobActorB1!", false, false, true)
	incidentA := recordstoretest.CreateIncidentInStore(t, harness.DB, actorA, "txn-evidence_lifecycle-blob-incident-a", "IR-P5-BLOB-A", "Evidence blob incident A")
	incidentB := recordstoretest.CreateIncidentInStore(t, harness.DB, actorA, "txn-evidence_lifecycle-blob-incident-b", "IR-P5-BLOB-B", "Evidence blob incident B")

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

func TestBlobCreateSizeCeiling_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence_lifecycle-blob-size-ceiling")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-size-incident",
		"incident_key":  "evidence_lifecycle-size",
		"title":         "Evidence size ceiling",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	maxCreate := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":   incidentID.String(),
		"client_txn_id": "txn-evidence_lifecycle-size-max",
		"byte_size":     int64(536870912),
	}, authOptions(login)...)
	maxData := httptestx.RequireSuccessEnvelope(t, maxCreate, http.StatusCreated)["data"].(map[string]any)
	if maxData["upload_state"] != "pending" || maxData["upload_target"] == nil || maxData["object_blob_id"] == nil {
		t.Fatalf("max-size create missing pending slot payload: %#v", maxData)
	}

	beforeRejected := countObjectBlobs(t, harness, incidentID)
	rejected := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":   incidentID.String(),
		"client_txn_id": "txn-evidence_lifecycle-size-too-large",
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
	if got := countBlobCreateIdempotency(t, harness, adminID, incidentID, "txn-evidence_lifecycle-size-too-large"); got != 0 {
		t.Fatalf("oversize request created idempotency success payload: got %d want 0", got)
	}
}

func TestPreviewPayloadCeiling_Unit(t *testing.T) {
	const (
		maxPreviewBytes = int64(64)
		maxTextBytes    = int64(32)
	)
	harness := appsupport.StartServerWithConfig(t, "evidence_lifecycle-preview-size", func(cfg *configassembly.Deployment) {
		cfg.Limits.Previews.MaxPreviewablePayloadBytes = maxPreviewBytes
		cfg.Limits.Previews.MaxTextInlineBytes = maxTextBytes
	})
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-preview-size-incident",
		"incident_key":  "evidence_lifecycle-preview-size",
		"title":         "Evidence preview size",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	atLimitRecordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, atLimitRecordID)
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, atLimitRecordID, []byte(strings.Repeat("i", int(maxPreviewBytes))), "limit.png", "image/png", "txn-evidence_lifecycle-preview-limit-blob", "txn-evidence_lifecycle-preview-limit-attach")
	preview := issueEvidenceHandle(t, harness, login, atLimitRecordID, "preview-handle")
	if preview["handle_kind"] != "preview" || preview["preview_kind"] != "image_inline" {
		t.Fatalf("preview at limit failed to issue image_inline handle: %#v", preview)
	}

	oversizeRecordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, oversizeRecordID)
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, oversizeRecordID, []byte(strings.Repeat("i", int(maxPreviewBytes+1))), "oversize.png", "image/png", "txn-evidence_lifecycle-preview-oversize-blob", "txn-evidence_lifecycle-preview-oversize-attach")
	requireEvidenceAccessUnavailableReason(t,
		appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+oversizeRecordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...),
		"preview_payload_too_large",
	)
	download := issueEvidenceHandle(t, harness, login, oversizeRecordID, "download-handle")
	if download["handle_kind"] != "download" {
		t.Fatalf("oversize preview should not block download issuance: %#v", download)
	}

	textAtLimitRecordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, textAtLimitRecordID)
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, textAtLimitRecordID, []byte(strings.Repeat("t", int(maxTextBytes))), "limit.txt", "text/plain", "txn-evidence_lifecycle-preview-text-limit-blob", "txn-evidence_lifecycle-preview-text-limit-attach")
	textPreview := issueEvidenceHandle(t, harness, login, textAtLimitRecordID, "preview-handle")
	if textPreview["handle_kind"] != "preview" || textPreview["preview_kind"] != "text_inline" || textPreview["size_bytes"] != float64(maxTextBytes) {
		t.Fatalf("text_inline preview at text limit failed: %#v", textPreview)
	}

	textOversizeRecordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, textOversizeRecordID)
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, textOversizeRecordID, []byte(strings.Repeat("t", int(maxTextBytes+1))), "oversize.txt", "text/plain", "txn-evidence_lifecycle-preview-text-oversize-blob", "txn-evidence_lifecycle-preview-text-oversize-attach")
	requireEvidenceAccessUnavailableReason(t,
		appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+textOversizeRecordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...),
		"preview_payload_too_large",
	)
	textDownload := issueEvidenceHandle(t, harness, login, textOversizeRecordID, "download-handle")
	if textDownload["handle_kind"] != "download" || textDownload["size_bytes"] != float64(maxTextBytes+1) {
		t.Fatalf("text_inline preview-size block should leave download issuance legal: %#v", textDownload)
	}
}

func requireBlobCreateHTTPReason(t testing.TB, resp *http.Response, field string, reasonCode string) {
	t.Helper()
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_blob_create_request")
	details := body["error"].(map[string]any)["details"].(map[string]any)
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

func countBlobCreateIdempotency(t testing.TB, harness *appsupport.ServerHarness, actorID uuid.UUID, incidentID uuid.UUID, clientTxnID string) int {
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

func countObjectBlobsInStore(t testing.TB, harness *recordstoretest.StoreHarness, incidentID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRow(context.Background(), `SELECT count(*) FROM object_blobs WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		t.Fatalf("count object_blobs: %v", err)
	}
	return count
}
