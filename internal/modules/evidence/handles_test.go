package evidence_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	recordstoretest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestHandleIssueEmptyBodyNonIdempotent_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence_lifecycle-handle-issue")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-handle-issue-incident",
		"incident_key":  "evidence_lifecycle-handle-issue",
		"title":         "Evidence handle issue",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
	payload := []byte("evidence_lifecycle handle issue")
	expectedSHA := fmt.Sprintf("%x", sha256Sum(payload))
	attachData := attachUploadedBlobWithMetadata(t, harness, login, incidentID, recordID, payload, "issue.txt", "text/plain", "txn-evidence_lifecycle-handle-issue-blob", "txn-evidence_lifecycle-handle-issue-attach")
	objectBlobID := attachData["object_blob_id"].(string)

	beforeInvalid := countAccessHandles(t, harness, incidentID)
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
	if got := countAccessHandles(t, harness, incidentID); got != beforeInvalid {
		t.Fatalf("invalid handle issuance requests wrote %d handles, want %d", got, beforeInvalid)
	}

	issuedAt := time.Now().UTC().Add(time.Minute).Truncate(time.Second)
	httptestx.SetClockFixed(t, harness.Server, issuedAt)
	firstPreview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
	secondPreview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
	firstDownload := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
	secondDownload := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
	requireEvidenceHandleContract(t, firstPreview, evidenceHandleContract{
		IncidentID: incidentID.String(), RecordID: recordID.String(), ObjectBlobID: objectBlobID,
		Kind: "preview", ExpiresAt: issuedAt.Add(5 * time.Minute), SingleUse: false,
		MediaClass: "text", Disposition: "inline", PreviewKind: "text_inline", Filename: "issue.txt",
		ContentType: "text/plain", SizeBytes: int64(len(payload)), SHA256: expectedSHA,
	})
	requireEvidenceHandleContract(t, firstDownload, evidenceHandleContract{
		IncidentID: incidentID.String(), RecordID: recordID.String(), ObjectBlobID: objectBlobID,
		Kind: "download", ExpiresAt: issuedAt.Add(2 * time.Minute), SingleUse: true,
		MediaClass: "text", Disposition: "attachment", PreviewKind: "", Filename: "issue.txt",
		ContentType: "text/plain", SizeBytes: int64(len(payload)), SHA256: expectedSHA,
	})
	for name, pair := range map[string][2]map[string]any{
		"preview":  {firstPreview, secondPreview},
		"download": {firstDownload, secondDownload},
	} {
		if pair[0]["href"] == pair[1]["href"] {
			t.Fatalf("%s issuance reused href: first=%#v second=%#v", name, pair[0], pair[1])
		}
	}
	if got := countAccessHandles(t, harness, incidentID); got != beforeInvalid+4 {
		t.Fatalf("successful issuances wrote %d handles, want %d", got, beforeInvalid+4)
	}
}

func TestHandleRedemptionRechecksCurrentState_Unit(t *testing.T) {
	harness := recordstoretest.StartStore(t, "evidence_lifecycle-handle-current-state")
	revisionComposition := revisionsupport.MustComposition(t)
	store := evidence.NewStore(
		harness.DB,
		evidence.WithRevisionAppender(revisionComposition.Runtime.Appender()),
		evidence.WithCollaborationIntents(revisionComposition.Intents),
	)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "evidence_lifecycle-handle-current@example.test", "EvidenceLifecycle Handle Current", "EvidenceLifecycleHandleCurrent1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-evidence_lifecycle-handle-current-incident", "IR-P5-HANDLE-CURRENT", "Evidence handle current state")

	unsupportedRecordID := seedEvidenceAttachmentRecord(t, harness.DB, incident.ID, actor.ID, "available")
	unsupportedBlobID := seedBlob(t, harness.DB, incident.ID, actor.ID, "available", BlobOptions{
		ByteSize:            11,
		ObservedSize:        ptrInt64(11),
		ObservedContentType: ptrString("audio/wav"),
		ObservedSHA:         ptrString(strings.Repeat("a", 64)),
	})
	linkEvidenceBlobInStore(t, harness.DB, unsupportedRecordID, unsupportedBlobID)
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
			blobID := seedBlob(t, harness.DB, incident.ID, actor.ID, "pending", BlobOptions{ByteSize: 11})
			linkEvidenceBlobInStore(t, harness.DB, recordID, blobID)
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "failed blob", reasonCode: "blob_failed", arrange: func(recordID uuid.UUID) evidence.HandleRecord {
			blobID := seedBlob(t, harness.DB, incident.ID, actor.ID, "failed", BlobOptions{ByteSize: 11})
			linkEvidenceBlobInStore(t, harness.DB, recordID, blobID)
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "quarantined", reasonCode: "evidence_quarantined", arrange: func(recordID uuid.UUID) evidence.HandleRecord {
			blobID := seedBlob(t, harness.DB, incident.ID, actor.ID, "quarantined", BlobOptions{ByteSize: 11})
			linkEvidenceBlobInStore(t, harness.DB, recordID, blobID)
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "inconsistent", reasonCode: "evidence_inconsistent", arrange: func(recordID uuid.UUID) evidence.HandleRecord {
			blobID := seedBlob(t, harness.DB, incident.ID, actor.ID, "available", BlobOptions{
				ByteSize:            11,
				ObservedSize:        ptrInt64(11),
				ObservedContentType: ptrString("text/plain"),
				ObservedSHA:         ptrString(strings.Repeat("b", 64)),
			})
			linkEvidenceBlobInStore(t, harness.DB, recordID, blobID)
			updateEvidenceLifecycleInStore(t, harness.DB, recordID, "received")
			return currentStoreHandle(t, store, recordID, "preview")
		}},
		{name: "row version drift", reasonCode: "evidence_inconsistent", arrange: func(recordID uuid.UUID) evidence.HandleRecord {
			blobID := seedBlob(t, harness.DB, incident.ID, actor.ID, "available", BlobOptions{
				ByteSize:            11,
				ObservedSize:        ptrInt64(11),
				ObservedContentType: ptrString("text/plain"),
				ObservedSHA:         ptrString(strings.Repeat("c", 64)),
			})
			linkEvidenceBlobInStore(t, harness.DB, recordID, blobID)
			handle := currentStoreHandle(t, store, recordID, "preview")
			incrementRecordVersionInStore(t, harness.DB, recordID)
			return handle
		}},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			recordID := seedEvidenceAttachmentRecord(t, harness.DB, incident.ID, actor.ID, "available")
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

func TestDownloadDispositionFallback_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence_lifecycle-disposition")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-disposition-incident",
		"incident_key":  "evidence_lifecycle-disposition",
		"title":         "Evidence disposition",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	sanitizedRecordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, sanitizedRecordID)
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, sanitizedRecordID, []byte("filename body"), "dir/evil\\bad\r\nname.txt", "text/plain", "txn-evidence_lifecycle-disposition-sanitize-blob", "txn-evidence_lifecycle-disposition-sanitize-attach")
	download := issueEvidenceHandle(t, harness, login, sanitizedRecordID, "download-handle")
	if got := download["filename"]; got != "direvilbadname.txt" {
		t.Fatalf("sanitized filename got %#v want direvilbadname.txt", got)
	}
	resp := appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+download["href"].(string), nil, appsupport.WithCookies(login.SessionCookie))
	httptestx.RequireStatus(t, resp, http.StatusOK)
	_ = resp.Body.Close()
	disposition := resp.Header.Get("Content-Disposition")
	if !strings.HasPrefix(disposition, "attachment;") || !strings.Contains(disposition, `filename=direvilbadname.txt`) || !strings.Contains(disposition, "filename*=UTF-8''direvilbadname.txt") {
		t.Fatalf("unexpected sanitized Content-Disposition: %q", disposition)
	}

	fallbackRecordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, fallbackRecordID)
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, fallbackRecordID, []byte("pngish"), "../..", "image/png", "txn-evidence_lifecycle-disposition-fallback-blob", "txn-evidence_lifecycle-disposition-fallback-attach")
	preview := issueEvidenceHandle(t, harness, login, fallbackRecordID, "preview-handle")
	wantFilename := "evidence-" + fallbackRecordID.String() + ".png"
	if got := preview["filename"]; got != wantFilename {
		t.Fatalf("fallback filename got %#v want %q", got, wantFilename)
	}
	resp = appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+preview["href"].(string), nil, appsupport.WithCookies(login.SessionCookie))
	httptestx.RequireStatus(t, resp, http.StatusOK)
	_ = resp.Body.Close()
	disposition = resp.Header.Get("Content-Disposition")
	if !strings.HasPrefix(disposition, "inline;") || !strings.Contains(disposition, "filename="+wantFilename) || !strings.Contains(disposition, "filename*=UTF-8''"+wantFilename) {
		t.Fatalf("unexpected fallback Content-Disposition: %q", disposition)
	}
}

func TestSanitizeFilenameRemovesNUL_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence_lifecycle-disposition-nul")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-disposition-nul-incident",
		"incident_key":  "evidence_lifecycle-disposition-nul",
		"title":         "Evidence disposition NUL",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
	attachUploadedBlobWithMetadata(t, harness, login, incidentID, recordID, []byte("filename nul body"), "evil\x00name.txt", "text/plain", "txn-evidence_lifecycle-disposition-nul-blob", "txn-evidence_lifecycle-disposition-nul-attach")
	download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
	if got := download["filename"]; got != "evilname.txt" {
		t.Fatalf("NUL-sanitized filename got %#v want evilname.txt", got)
	}
	resp := appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+download["href"].(string), nil, appsupport.WithCookies(login.SessionCookie))
	httptestx.RequireStatus(t, resp, http.StatusOK)
	_ = resp.Body.Close()
	disposition := resp.Header.Get("Content-Disposition")
	if strings.Contains(disposition, "\x00") || !strings.Contains(disposition, `filename=evilname.txt`) || !strings.Contains(disposition, "filename*=UTF-8''evilname.txt") {
		t.Fatalf("unexpected NUL-sanitized Content-Disposition: %q", disposition)
	}
}

func issueEvidenceHandle(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, endpoint string) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

type evidenceHandleContract struct {
	IncidentID   string
	RecordID     string
	ObjectBlobID string
	Kind         string
	ExpiresAt    time.Time
	SingleUse    bool
	MediaClass   string
	Disposition  string
	PreviewKind  string
	Filename     string
	ContentType  string
	SizeBytes    int64
	SHA256       string
}

func requireEvidenceHandleContract(t testing.TB, data map[string]any, want evidenceHandleContract) {
	t.Helper()
	for key, wantValue := range map[string]any{
		"incident_id":              want.IncidentID,
		"record_id":                want.RecordID,
		"object_blob_id":           want.ObjectBlobID,
		"handle_kind":              want.Kind,
		"method":                   "GET",
		"single_use":               want.SingleUse,
		"media_class":              want.MediaClass,
		"disposition":              want.Disposition,
		"filename":                 want.Filename,
		"content_type":             want.ContentType,
		"size_bytes":               float64(want.SizeBytes),
		"sha256":                   want.SHA256,
		"evidence_lifecycle_state": "available",
		"upload_state":             "available",
	} {
		if got := data[key]; got != wantValue {
			t.Fatalf("handle[%s] got %#v want %#v in %#v", key, got, wantValue, data)
		}
	}
	href, ok := data["href"].(string)
	if !ok || !strings.HasPrefix(href, "/api/v1/evidence-handles/hdl_") {
		t.Fatalf("handle href is not an opaque same-origin handle: %#v in %#v", data["href"], data)
	}
	if strings.Contains(href, want.RecordID) || strings.Contains(href, want.ObjectBlobID) {
		t.Fatalf("handle href leaks stable record/blob identity: %q", href)
	}
	expiresAt, ok := data["expires_at"].(string)
	if !ok {
		t.Fatalf("handle expires_at got %T in %#v", data["expires_at"], data)
	}
	if got := mustParseTime(t, expiresAt); !got.Equal(want.ExpiresAt.UTC()) {
		t.Fatalf("handle expires_at got %s want %s", got, want.ExpiresAt.UTC())
	}
	if want.PreviewKind == "" {
		if _, exists := data["preview_kind"]; exists {
			t.Fatalf("download handle must omit preview_kind: %#v", data)
		}
		return
	}
	if got := data["preview_kind"]; got != want.PreviewKind {
		t.Fatalf("handle preview_kind got %#v want %q in %#v", got, want.PreviewKind, data)
	}
}

func attachUploadedBlobWithMetadata(t *testing.T, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, payload []byte, filename string, contentType string, createTxn string, attachTxn string) map[string]any {
	t.Helper()
	createResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     createTxn,
		"byte_size":         len(payload),
		"filename_hint":     filename,
		"content_type_hint": contentType,
		"sha256_hex":        fmt.Sprintf("%x", sha256Sum(payload)),
	}, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	putObject(t, harness.Server.HTTP.URL, createData["upload_target"].(map[string]any)["href"].(string), payload, contentType)
	attachResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/attach-blob", map[string]any{
		"object_blob_id":   createData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    attachTxn,
	}, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
}

func updateBlobObservedMetadata(t testing.TB, harness *appsupport.ServerHarness, objectBlobID uuid.UUID, size int64, contentType string, sha string) {
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

func linkEvidenceBlobInStore(t testing.TB, db postgres.DB, recordID uuid.UUID, blobID uuid.UUID) {
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

func updateEvidenceLifecycleInStore(t testing.TB, db postgres.DB, recordID uuid.UUID, lifecycleState string) {
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

func incrementRecordVersionInStore(t testing.TB, db postgres.DB, recordID uuid.UUID) {
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
