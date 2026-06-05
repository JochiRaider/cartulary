package evidence_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

type phase5ObjectStoreDependencyAdmin struct {
	email    string
	password string
	userID   uuid.UUID
}

func requirePhaseDObjectStoreDependencyErrorsUseOwnerPublicMapping(
	t *testing.T,
	startServer func(testing.TB, string, objectstore.Store) *phase4test.ServerHarness,
	admin phase5ObjectStoreDependencyAdmin,
) {
	t.Run("blob slot create maps endpoint outage before writing slot", func(t *testing.T) {
		baseStore, err := objectstore.NewFilesystemStore(t.TempDir())
		if err != nil {
			t.Fatalf("create filesystem object store: %v", err)
		}
		failingStore := &overrideObjectStore{
			Store: baseStore,
			uploadTargetErr: &objectstore.AdapterError{
				Code:      objectstore.ErrorCodeUnavailable,
				Reason:    objectstore.ReasonEndpointUnreachable,
				Operation: objectstore.OperationCreateUploadTarget,
				Retryable: true,
				Message:   "endpoint unreachable source.internal private-bucket incidents/raw/key CARTULARY_S3_SECRET_KEY=secret",
			},
		}

		harness := startServer(t, "phase-d-object-store-create-outage", failingStore)
		login := loginLocalUserNoMFA(t, harness, admin.email, admin.password)
		incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-phase-d-object-store-create-incident",
			"incident_key":  "phase-d-object-store-create",
			"title":         "Phase D object-store create outage",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

		resp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
			"incident_id":       incidentID.String(),
			"client_txn_id":     "txn-phase-d-object-store-create-blob",
			"byte_size":         4,
			"filename_hint":     "outage.txt",
			"content_type_hint": "text/plain",
		}, authOptions(login)...)
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusServiceUnavailable, "object_store_unavailable")
		errorValue := body["error"].(map[string]any)
		if retryable, ok := errorValue["retryable"].(bool); !ok || !retryable {
			t.Fatalf("object_store_unavailable retryable got %#v want true", errorValue["retryable"])
		}
		httptestx.RequireErrorDetail(t, body, "reason_code", "endpoint_unreachable")
		requireNoPublicEvidenceLeak(t, "blob slot dependency error", body, []string{
			"source.internal",
			"private-bucket",
			"incidents/raw/key",
			"CARTULARY_S3_SECRET_KEY",
			"secret",
		})
		if got := countObjectBlobs(t, harness, incidentID); got != 0 {
			t.Fatalf("object-store outage wrote blob slots: got %d want 0", got)
		}
	})

	t.Run("preview issuance maps capability rejection after visible state checks", func(t *testing.T) {
		baseStore, err := objectstore.NewFilesystemStore(t.TempDir())
		if err != nil {
			t.Fatalf("create filesystem object store: %v", err)
		}
		failingStore := &overrideObjectStore{
			Store: baseStore,
			statErr: &objectstore.AdapterError{
				Code:      objectstore.ErrorCodeAccessRejected,
				Reason:    objectstore.ReasonCapabilityMissing,
				Operation: objectstore.OperationHeadObject,
				Retryable: false,
				Message:   "capability missing target.internal private-bucket incidents/raw/key AWS_SECRET_ACCESS_KEY=secret",
			},
		}

		harness := startServer(t, "phase-d-object-store-preview-capability", failingStore)
		login := loginLocalUserNoMFA(t, harness, admin.email, admin.password)
		incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-phase-d-object-store-preview-incident",
			"incident_key":  "phase-d-object-store-preview",
			"title":         "Phase D object-store preview outage",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, admin.userID, recordID)
		linkSeededBlobWithCanonicalStorageKey(t, harness, incidentID, admin.userID, recordID, "available", "available")

		resp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...)
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusServiceUnavailable, "object_store_access_rejected")
		httptestx.RequireErrorDetail(t, body, "reason_code", "capability_missing")
		errorValue := body["error"].(map[string]any)
		if retryable, ok := errorValue["retryable"].(bool); !ok || retryable {
			t.Fatalf("object_store_access_rejected retryable got %#v want false", errorValue["retryable"])
		}
		requireNoPublicEvidenceLeak(t, "preview dependency error", body, []string{
			"target.internal",
			"private-bucket",
			"incidents/raw/key",
			"AWS_SECRET_ACCESS_KEY",
			"secret",
		})
	})

	t.Run("malformed persisted key fails before backend on handle issuance", func(t *testing.T) {
		baseStore, err := objectstore.NewFilesystemStore(t.TempDir())
		if err != nil {
			t.Fatalf("create filesystem object store: %v", err)
		}
		recordingStore := &overrideObjectStore{Store: baseStore}
		harness := startServer(t, "phase-g-object-key-issue-invalid", recordingStore)
		login := loginLocalUserNoMFA(t, harness, admin.email, admin.password)
		incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-phase-g-object-key-issue-incident",
			"incident_key":  "phase-g-object-key-issue",
			"title":         "Phase G object key issue",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, admin.userID, recordID)
		attachData := attachUploadedBlob(t, harness, login, incidentID, recordID, []byte("invalid key issue"), "txn-phase-g-object-key-issue-blob", "txn-phase-g-object-key-issue-attach")
		objectBlobID := phase4test.MustUUID(t, attachData["object_blob_id"].(string))
		malformedKey := "objects/not-canonical"
		updateBlobStorageKey(t, harness, objectBlobID, malformedKey)
		recordingStore.resetCounts()

		resp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...)
		body := requireObjectStoreInvalidRequestReason(t, resp, "object_blob_storage_key_malformed")
		requireNoPublicEvidenceLeak(t, "malformed persisted key error", body, []string{malformedKey})
		if statCalls, readCalls := recordingStore.counts(); statCalls != 0 || readCalls != 0 {
			t.Fatalf("malformed storage key reached backend: stat=%d read=%d", statCalls, readCalls)
		}
	})

	t.Run("identity-mismatched persisted key fails before backend on handle redemption", func(t *testing.T) {
		baseStore, err := objectstore.NewFilesystemStore(t.TempDir())
		if err != nil {
			t.Fatalf("create filesystem object store: %v", err)
		}
		recordingStore := &overrideObjectStore{Store: baseStore}
		harness := startServer(t, "phase-g-object-key-redeem-invalid", recordingStore)
		login := loginLocalUserNoMFA(t, harness, admin.email, admin.password)
		incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-phase-g-object-key-redeem-incident",
			"incident_key":  "phase-g-object-key-redeem",
			"title":         "Phase G object key redeem",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, admin.userID, recordID)
		attachData := attachUploadedBlob(t, harness, login, incidentID, recordID, []byte("invalid key redeem"), "txn-phase-g-object-key-redeem-blob", "txn-phase-g-object-key-redeem-attach")
		objectBlobID := phase4test.MustUUID(t, attachData["object_blob_id"].(string))
		handle := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
		mismatchedKey, err := blobref.ObjectBlobStorageKey(incidentID, uuid.New())
		if err != nil {
			t.Fatalf("mismatched storage key: %v", err)
		}
		updateBlobStorageKey(t, harness, objectBlobID, mismatchedKey)
		recordingStore.resetCounts()

		resp := phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, phase4test.WithCookies(login.SessionCookie))
		body := requireObjectStoreInvalidRequestReason(t, resp, "object_blob_storage_key_identity_mismatch")
		requireNoPublicEvidenceLeak(t, "identity-mismatched persisted key error", body, []string{mismatchedKey})
		if statCalls, readCalls := recordingStore.counts(); statCalls != 0 || readCalls != 0 {
			t.Fatalf("identity-mismatched storage key reached backend: stat=%d read=%d", statCalls, readCalls)
		}
	})
}

func TestPhaseG_PublicEvidenceResponsesDoNotLeakObjectStoreIdentifiers(t *testing.T) {
	harness := phase4test.StartServer(t, "phase-g-evidence-redaction")
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase-g-redaction-incident",
		"incident_key":  "phase-g-redaction",
		"title":         "Phase G evidence redaction",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)

	payload := []byte("phase g public evidence response body")
	attachData := attachUploadedBlobWithMetadata(
		t,
		harness,
		login,
		incidentID,
		recordID,
		payload,
		"phase-g.txt",
		"text/plain",
		"txn-phase-g-redaction-blob",
		"txn-phase-g-redaction-attach",
	)
	objectBlobID := phase4test.MustUUID(t, attachData["object_blob_id"].(string))
	storageKey := "incidents/" + incidentID.String() + "/object-blobs/" + objectBlobID.String()
	forbidden := []string{
		storageKey,
		"storage_ref",
		"s3://",
		"private-bucket",
		"source.internal",
		"target.internal",
		"CARTULARY_S3_ACCESS_KEY",
		"CARTULARY_S3_SECRET_KEY",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"X-Amz-Credential",
		"X-Amz-Signature",
	}

	createResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-phase-g-redaction-extra-blob",
		"byte_size":         1,
		"filename_hint":     "extra.txt",
		"content_type_hint": "text/plain",
	}, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	requireNoPublicEvidenceLeak(t, "blob create response", createData, forbidden)

	for _, endpoint := range []string{"preview-handle", "download-handle"} {
		handle := issueEvidenceHandle(t, harness, login, recordID, endpoint)
		requireNoPublicEvidenceLeak(t, endpoint+" response", handle, forbidden)
		href, ok := handle["href"].(string)
		if !ok || !strings.HasPrefix(href, "/api/v1/evidence-handles/hdl_") {
			t.Fatalf("%s returned non-opaque same-origin href: %#v", endpoint, handle["href"])
		}
	}
}

func requireNoPublicEvidenceLeak(t testing.TB, label string, value any, forbidden []string) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %s: %v", label, err)
	}
	text := string(data)
	for _, marker := range forbidden {
		if marker != "" && strings.Contains(text, marker) {
			t.Fatalf("%s leaked forbidden marker %q in %s", label, marker, text)
		}
	}
}

type overrideObjectStore struct {
	objectstore.Store
	uploadTargetErr error
	statErr         error
	readErr         error
	mu              sync.Mutex
	statCalls       int
	readCalls       int
}

func (s *overrideObjectStore) UploadTarget(ctx context.Context, key string, expiresAt time.Time) (objectstore.UploadTarget, error) {
	if s.uploadTargetErr != nil {
		return objectstore.UploadTarget{}, s.uploadTargetErr
	}
	return s.Store.UploadTarget(ctx, key, expiresAt)
}

func (s *overrideObjectStore) StatObject(ctx context.Context, key string) (objectstore.ObjectInfo, error) {
	s.mu.Lock()
	s.statCalls++
	s.mu.Unlock()
	if s.statErr != nil {
		return objectstore.ObjectInfo{}, s.statErr
	}
	return s.Store.StatObject(ctx, key)
}

func (s *overrideObjectStore) ReadObject(ctx context.Context, key string, options objectstore.ReadOptions) (io.ReadCloser, objectstore.ObjectInfo, error) {
	s.mu.Lock()
	s.readCalls++
	s.mu.Unlock()
	if s.readErr != nil {
		return nil, objectstore.ObjectInfo{}, s.readErr
	}
	return s.Store.ReadObject(ctx, key, options)
}

func (s *overrideObjectStore) resetCounts() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statCalls = 0
	s.readCalls = 0
}

func (s *overrideObjectStore) counts() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statCalls, s.readCalls
}

func requireObjectStoreInvalidRequestReason(t *testing.T, resp *http.Response, wantReasonCode string) map[string]any {
	t.Helper()
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusInternalServerError, "object_store_invalid_request")
	httptestx.RequireErrorDetail(t, body, "reason_code", wantReasonCode)
	errorValue := body["error"].(map[string]any)
	if retryable, ok := errorValue["retryable"].(bool); !ok || retryable {
		t.Fatalf("object_store_invalid_request retryable got %#v want false", errorValue["retryable"])
	}
	return body
}
