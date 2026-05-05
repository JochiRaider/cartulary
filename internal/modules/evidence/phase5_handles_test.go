package evidence_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestPhase5_HandleIssueEmptyBodyNonIdempotent_U_5_05(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-handle-issue")
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-handle-issue-incident",
		"incident_key":  "phase5-handle-issue",
		"title":         "Phase 5 handle issue",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, recordID, []byte("phase5 handle issue"), "issue.txt", "text/plain", "txn-phase5-handle-issue-blob", "txn-phase5-handle-issue-attach")

	for _, endpoint := range []string{"preview-handle", "download-handle"} {
		for name, body := range map[string]string{
			"empty body":    "",
			"null":          "null",
			"array":         "[]",
			"client txn":    `{"client_txn_id":"forbidden"}`,
			"unknown field": `{"unexpected":true}`,
		} {
			t.Run(endpoint+" rejects "+name, func(t *testing.T) {
				resp := doRawJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, body, authOptions(login)...)
				httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_evidence_handle_request")
			})
		}
	}

	before := countAccessHandles(t, harness, incidentID)
	firstPreview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
	secondPreview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
	firstDownload := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
	secondDownload := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
	for name, pair := range map[string][2]map[string]any{
		"preview":  {firstPreview, secondPreview},
		"download": {firstDownload, secondDownload},
	} {
		if pair[0]["href"] == pair[1]["href"] {
			t.Fatalf("%s issuance reused href: first=%#v second=%#v", name, pair[0], pair[1])
		}
	}
	if got := countAccessHandles(t, harness, incidentID); got != before+4 {
		t.Fatalf("successful issuances wrote %d handles, want %d", got, before+4)
	}
}

func TestPhase5_HandleRedemptionRechecksCurrentState_U_5_06(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase5-handle-current-state")
	store := evidence.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "phase5-handle-current@example.test", "Phase5 Handle Current", "Phase5HandleCurrent1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase5-handle-current-incident", "IR-P5-HANDLE-CURRENT", "Phase 5 handle current state")

	unsupportedRecordID := seedPhase5EvidenceRecord(t, harness.DB, incident.ID, actor.ID, "available")
	unsupportedBlobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "available", phase5BlobOptions{
		ByteSize:            11,
		ObservedSize:        ptrInt64(11),
		ObservedContentType: ptrString("audio/wav"),
		ObservedSHA:         ptrString(strings.Repeat("a", 64)),
	})
	linkPhase5EvidenceBlobInStore(t, harness.DB, unsupportedRecordID, unsupportedBlobID)
	unsupportedAccess, err := store.LoadEvidenceAccess(context.Background(), unsupportedRecordID)
	if err != nil {
		t.Fatalf("load unsupported preview access: %v", err)
	}
	if unsupportedAccess.PreviewKind != nil {
		t.Fatalf("audio evidence unexpectedly previewable: %#v", unsupportedAccess)
	}
	if reason, err := store.CheckHandleAccess(context.Background(), handleFromAccess(unsupportedAccess, "download")); err != nil || reason != "" {
		t.Fatalf("unsupported preview evidence should remain downloadable: reason=%q err=%v", reason, err)
	}

	scenarios := []struct {
		name       string
		reasonCode string
		arrange    func(recordID uuid.UUID) evidence.HandleRecord
	}{
		{name: "pending blob", reasonCode: "blob_pending", arrange: func(recordID uuid.UUID) evidence.HandleRecord {
			blobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "pending", phase5BlobOptions{ByteSize: 11})
			linkPhase5EvidenceBlobInStore(t, harness.DB, recordID, blobID)
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "failed blob", reasonCode: "blob_failed", arrange: func(recordID uuid.UUID) evidence.HandleRecord {
			blobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "failed", phase5BlobOptions{ByteSize: 11})
			linkPhase5EvidenceBlobInStore(t, harness.DB, recordID, blobID)
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "quarantined", reasonCode: "evidence_quarantined", arrange: func(recordID uuid.UUID) evidence.HandleRecord {
			blobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "quarantined", phase5BlobOptions{ByteSize: 11})
			linkPhase5EvidenceBlobInStore(t, harness.DB, recordID, blobID)
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "inconsistent", reasonCode: "evidence_inconsistent", arrange: func(recordID uuid.UUID) evidence.HandleRecord {
			blobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "available", phase5BlobOptions{
				ByteSize:            11,
				ObservedSize:        ptrInt64(11),
				ObservedContentType: ptrString("text/plain"),
				ObservedSHA:         ptrString(strings.Repeat("b", 64)),
			})
			linkPhase5EvidenceBlobInStore(t, harness.DB, recordID, blobID)
			updatePhase5EvidenceLifecycleInStore(t, harness.DB, recordID, "received")
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "row version drift", reasonCode: "evidence_inconsistent", arrange: func(recordID uuid.UUID) evidence.HandleRecord {
			blobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "available", phase5BlobOptions{
				ByteSize:            11,
				ObservedSize:        ptrInt64(11),
				ObservedContentType: ptrString("text/plain"),
				ObservedSHA:         ptrString(strings.Repeat("c", 64)),
			})
			linkPhase5EvidenceBlobInStore(t, harness.DB, recordID, blobID)
			handle := currentStoreHandle(t, store, recordID, "preview")
			incrementPhase5RecordVersionInStore(t, harness.DB, recordID)
			return handle
		}},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			recordID := seedPhase5EvidenceRecord(t, harness.DB, incident.ID, actor.ID, "available")
			handle := scenario.arrange(recordID)
			reason, err := store.CheckHandleAccess(context.Background(), handle)
			if err != nil {
				t.Fatalf("CheckHandleAccess: %v", err)
			}
			if reason != scenario.reasonCode {
				t.Fatalf("reason got %q want %q", reason, scenario.reasonCode)
			}
		})
	}
}

func TestPhase5_DownloadDispositionFallback_U_5_07(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-disposition")
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-disposition-incident",
		"incident_key":  "phase5-disposition",
		"title":         "Phase 5 disposition",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	sanitizedRecordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, sanitizedRecordID)
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, sanitizedRecordID, []byte("filename body"), "dir/evil\\bad\r\nname.txt", "text/plain", "txn-phase5-disposition-sanitize-blob", "txn-phase5-disposition-sanitize-attach")
	download := issueEvidenceHandle(t, harness, login, sanitizedRecordID, "download-handle")
	if got := download["filename"]; got != "direvilbadname.txt" {
		t.Fatalf("sanitized filename got %#v want direvilbadname.txt", got)
	}
	resp := phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+download["href"].(string), nil, phase4test.WithCookies(login.SessionCookie))
	httptestx.RequireStatus(t, resp, http.StatusOK)
	_ = resp.Body.Close()
	disposition := resp.Header.Get("Content-Disposition")
	if !strings.HasPrefix(disposition, "attachment;") || !strings.Contains(disposition, `filename=direvilbadname.txt`) || !strings.Contains(disposition, "filename*=UTF-8''direvilbadname.txt") {
		t.Fatalf("unexpected sanitized Content-Disposition: %q", disposition)
	}

	fallbackRecordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, fallbackRecordID)
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, fallbackRecordID, []byte("pngish"), "../..", "image/png", "txn-phase5-disposition-fallback-blob", "txn-phase5-disposition-fallback-attach")
	preview := issueEvidenceHandle(t, harness, login, fallbackRecordID, "preview-handle")
	wantFilename := "evidence-" + fallbackRecordID.String() + ".png"
	if got := preview["filename"]; got != wantFilename {
		t.Fatalf("fallback filename got %#v want %q", got, wantFilename)
	}
	resp = phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+preview["href"].(string), nil, phase4test.WithCookies(login.SessionCookie))
	httptestx.RequireStatus(t, resp, http.StatusOK)
	_ = resp.Body.Close()
	disposition = resp.Header.Get("Content-Disposition")
	if !strings.HasPrefix(disposition, "inline;") || !strings.Contains(disposition, "filename="+wantFilename) || !strings.Contains(disposition, "filename*=UTF-8''"+wantFilename) {
		t.Fatalf("unexpected fallback Content-Disposition: %q", disposition)
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

func issueEvidenceHandle(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, recordID uuid.UUID, endpoint string) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func attachUploadedBlobWithMetadata(t *testing.T, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, payload []byte, filename string, contentType string, createTxn string, attachTxn string) map[string]any {
	t.Helper()
	createResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     createTxn,
		"byte_size":         len(payload),
		"filename_hint":     filename,
		"content_type_hint": contentType,
		"sha256_hex":        fmt.Sprintf("%x", sha256Sum(payload)),
	}, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	putObject(t, createData["upload_target"].(map[string]any)["href"].(string), payload, contentType)
	attachResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/attach-blob", map[string]any{
		"object_blob_id":   createData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    attachTxn,
	}, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
}

func updateBlobObservedMetadata(t testing.TB, harness *phase4test.ServerHarness, objectBlobID uuid.UUID, size int64, contentType string, sha string) {
	t.Helper()
	shaValue := any(nil)
	if sha != "" {
		shaValue = sha
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE object_blobs
   SET observed_size = $2,
       observed_content_type = $3,
       observed_sha256_hex = COALESCE($4, observed_sha256_hex),
       updated_at = now()
 WHERE object_blob_id = $1
`, objectBlobID, size, contentType, shaValue); err != nil {
		t.Fatalf("update blob observed metadata: %v", err)
	}
}

func linkPhase5EvidenceBlobInStore(t testing.TB, db postgres.DB, recordID uuid.UUID, blobID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
UPDATE evidence
   SET object_blob_id = $2,
       lifecycle_state = 'available',
       upload_state = (SELECT upload_state FROM object_blobs WHERE object_blob_id = $2),
       updated_at = now()
 WHERE record_id = $1
`, recordID, blobID); err != nil {
		t.Fatalf("link evidence blob in store: %v", err)
	}
}

func currentStoreHandle(t testing.TB, store *evidence.Store, recordID uuid.UUID, kind string) evidence.HandleRecord {
	t.Helper()
	access, err := store.LoadEvidenceAccess(context.Background(), recordID)
	if err != nil {
		t.Fatalf("load current handle access: %v", err)
	}
	return handleFromAccess(access, kind)
}

func handleFromAccess(access evidence.EvidenceAccessRecord, kind string) evidence.HandleRecord {
	disposition := "attachment"
	previewKind := access.PreviewKind
	if kind == "preview" {
		disposition = "inline"
	} else {
		previewKind = nil
	}
	var objectBlobID uuid.UUID
	if access.ObjectBlobID != nil {
		objectBlobID = *access.ObjectBlobID
	}
	var storageKey string
	if access.StorageKey != nil {
		storageKey = *access.StorageKey
	}
	return evidence.HandleRecord{
		Token:                  "test-handle-" + uuid.NewString(),
		IncidentID:             access.IncidentID,
		RecordID:               access.RecordID,
		RecordRowVersion:       access.RecordRowVersion,
		ObjectBlobID:           objectBlobID,
		StorageKey:             storageKey,
		SessionID:              uuid.New(),
		HandleKind:             kind,
		MediaClass:             access.MediaClass,
		PreviewKind:            previewKind,
		Disposition:            disposition,
		Filename:               access.FilenameSource,
		ContentType:            access.ContentType,
		SizeBytes:              access.SizeBytes,
		SHA256:                 access.SHA256,
		EvidenceLifecycleState: access.EvidenceLifecycleState,
		UploadState:            access.UploadState,
	}
}

func updatePhase5EvidenceLifecycleInStore(t testing.TB, db postgres.DB, recordID uuid.UUID, lifecycleState string) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
UPDATE evidence
   SET lifecycle_state = $2,
       updated_at = now()
 WHERE record_id = $1
`, recordID, lifecycleState); err != nil {
		t.Fatalf("update evidence lifecycle in store: %v", err)
	}
}

func incrementPhase5RecordVersionInStore(t testing.TB, db postgres.DB, recordID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
UPDATE records
   SET row_version = row_version + 1,
       updated_at = now()
 WHERE record_id = $1
`, recordID); err != nil {
		t.Fatalf("increment record row version in store: %v", err)
	}
}

func sha256Sum(payload []byte) []byte {
	sum := sha256.Sum256(payload)
	return sum[:]
}

func readHandleBody(t testing.TB, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read handle body: %v", err)
	}
	return data
}
