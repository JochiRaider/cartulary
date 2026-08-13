package evidence_test

import (
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
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
)

type ObjectStoreDependencyAdmin struct {
	email    string
	password string
	userID   uuid.UUID
}

func requireObjectStoreDependencyErrorsUseOwnerPublicMapping(
	t *testing.T,
	startServer func(testing.TB, string, objectstore.Store) *appsupport.ServerHarness,
	admin ObjectStoreDependencyAdmin,
) {
	t.Run("object upload maps endpoint outage without exposing storage target", func(t *testing.T) {
		baseStore, err := objectstore.NewFilesystemStore(t.TempDir())
		if err != nil {
			t.Fatalf("create filesystem object store: %v", err)
		}
		failingStore := &overrideObjectStore{
			Store: baseStore,
			putErr: &objectstore.AdapterError{
				Code:      objectstore.ErrorCodeUnavailable,
				Reason:    objectstore.ReasonEndpointUnreachable,
				Operation: objectstore.OperationPutObject,
				Retryable: true,
				Message:   "endpoint unreachable source.internal private-bucket incidents/raw/key CARTULARY_S3_SECRET_KEY=secret",
			},
		}

		harness := startServer(t, "object-store-upload-outage", failingStore)
		login := loginLocalUserNoMFA(t, harness, admin.email, admin.password)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-object-store-upload-incident",
			"incident_key":  "object-store-upload",
			"title":         "Object-store upload outage",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

		createResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
			"incident_id":       incidentID.String(),
			"client_txn_id":     "txn-object-store-upload-blob",
			"byte_size":         4,
			"filename_hint":     "outage.txt",
			"content_type_hint": "text/plain",
		}, authOptions(login)...)
		createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
		uploadTarget := createData["upload_target"].(map[string]any)
		href, ok := uploadTarget["href"].(string)
		if !ok || !strings.HasPrefix(href, "/api/v1/object-uploads/upl_") || strings.Contains(href, "://") {
			t.Fatalf("blob create returned non-opaque upload target: %#v", uploadTarget)
		}

		req, err := http.NewRequest(http.MethodPut, harness.Server.HTTP.URL+href, strings.NewReader("fail"))
		if err != nil {
			t.Fatalf("create object upload request: %v", err)
		}
		req.Header.Set("Content-Type", "text/plain")
		for _, option := range authOptions(login) {
			option(req)
		}
		resp := httptestx.Do(t, http.DefaultClient, req)
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusServiceUnavailable, "object_store_unavailable")
		errorValue := body["error"].(map[string]any)
		if retryable, ok := errorValue["retryable"].(bool); !ok || !retryable {
			t.Fatalf("object_store_unavailable retryable got %#v want true", errorValue["retryable"])
		}
		httptestx.RequireErrorDetail(t, body, "reason_code", "endpoint_unreachable")
		requireNoPublicEvidenceLeak(t, "object upload dependency error", body, []string{
			"source.internal",
			"private-bucket",
			"incidents/raw/key",
			"CARTULARY_S3_SECRET_KEY",
			"secret",
		})
		if got := countObjectBlobs(t, harness, incidentID); got != 1 {
			t.Fatalf("app-mediated upload outage should retain the pending blob slot: got %d want 1", got)
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

		harness := startServer(t, "object-store-preview-capability", failingStore)
		login := loginLocalUserNoMFA(t, harness, admin.email, admin.password)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-object-store-preview-incident",
			"incident_key":  "object-store-preview",
			"title":         "Object-store preview outage",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, admin.userID, recordID)
		linkSeededBlobWithCanonicalStorageKey(t, harness, incidentID, admin.userID, recordID, "available", "available")

		resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...)
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
		harness := startServer(t, "object-key-issue-invalid", recordingStore)
		login := loginLocalUserNoMFA(t, harness, admin.email, admin.password)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-object-key-issue-incident",
			"incident_key":  "object-key-issue",
			"title":         "Object key issue",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, admin.userID, recordID)
		attachData := attachUploadedBlob(t, harness, login, incidentID, recordID, []byte("invalid key issue"), "txn-object-key-issue-blob", "txn-object-key-issue-attach")
		objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))
		malformedKey := "objects/not-canonical"
		updateBlobStorageKey(t, harness, objectBlobID, malformedKey)
		recordingStore.resetCounts()

		resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...)
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
		harness := startServer(t, "object-key-redeem-invalid", recordingStore)
		login := loginLocalUserNoMFA(t, harness, admin.email, admin.password)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-object-key-redeem-incident",
			"incident_key":  "object-key-redeem",
			"title":         "Object key redeem",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, admin.userID, recordID)
		attachData := attachUploadedBlob(t, harness, login, incidentID, recordID, []byte("invalid key redeem"), "txn-object-key-redeem-blob", "txn-object-key-redeem-attach")
		objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))
		handle := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
		mismatchedKey, err := blobref.ObjectBlobStorageKey(incidentID, uuid.New())
		if err != nil {
			t.Fatalf("mismatched storage key: %v", err)
		}
		updateBlobStorageKey(t, harness, objectBlobID, mismatchedKey)
		recordingStore.resetCounts()

		resp := appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, appsupport.WithCookies(login.SessionCookie))
		body := requireObjectStoreInvalidRequestReason(t, resp, "object_blob_storage_key_identity_mismatch")
		requireNoPublicEvidenceLeak(t, "identity-mismatched persisted key error", body, []string{mismatchedKey})
		if statCalls, readCalls := recordingStore.counts(); statCalls != 0 || readCalls != 0 {
			t.Fatalf("identity-mismatched storage key reached backend: stat=%d read=%d", statCalls, readCalls)
		}
	})
}

func TestPublicEvidenceResponsesDoNotLeakObjectStoreIdentifiers(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence-redaction")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence-redaction-incident",
		"incident_key":  "evidence-redaction",
		"title":         "Evidence redaction",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)

	payload := []byte("public evidence response body")
	attachData := attachUploadedBlobWithMetadata(
		t,
		harness,
		login,
		incidentID,
		recordID,
		payload,
		"evidence-redaction.txt",
		"text/plain",
		"txn-evidence-redaction-blob",
		"txn-evidence-redaction-attach",
	)
	objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))
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

	createResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-evidence-redaction-extra-blob",
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
	putErr          error
	statErr         error
	readErr         error
	mu              sync.Mutex
	statCalls       int
	readCalls       int
}

var _ objectstore.Store = (*overrideObjectStore)(nil)
var _ objectstore.TypedStore = (*overrideObjectStore)(nil)

func (s *overrideObjectStore) UploadTarget(ctx context.Context, key string, expiresAt time.Time) (objectstore.UploadTarget, error) {
	if s.uploadTargetErr != nil {
		return objectstore.UploadTarget{}, s.uploadTargetErr
	}
	return s.Store.UploadTarget(ctx, key, expiresAt)
}

func (s *overrideObjectStore) PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if s.putErr != nil {
		return s.putErr
	}
	return s.Store.PutObject(ctx, key, body, size, contentType)
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

func (s *overrideObjectStore) CreateUploadTarget(ctx context.Context, request objectstore.UploadTargetRequest) (objectstore.UploadTarget, error) {
	if s.uploadTargetErr != nil {
		return objectstore.UploadTarget{}, s.uploadTargetErr
	}
	if typed, ok := s.Store.(objectstore.TypedStore); ok {
		return typed.CreateUploadTarget(ctx, request)
	}
	return s.Store.UploadTarget(ctx, request.Key, request.ExpiresAt)
}

func (s *overrideObjectStore) Put(ctx context.Context, request objectstore.PutObjectRequest) (objectstore.PutObjectResult, error) {
	if s.putErr != nil {
		return objectstore.PutObjectResult{}, s.putErr
	}
	if typed, ok := s.Store.(objectstore.TypedStore); ok {
		return typed.Put(ctx, request)
	}
	if err := s.Store.PutObject(ctx, request.Key, request.Body, request.Size, request.ContentType); err != nil {
		return objectstore.PutObjectResult{}, err
	}
	return objectstore.PutObjectResult{SizeBytes: request.Size, ContentType: request.ContentType, Metadata: request.Metadata}, nil
}

func (s *overrideObjectStore) Head(ctx context.Context, request objectstore.HeadObjectRequest) (objectstore.ObjectInfo, error) {
	s.mu.Lock()
	s.statCalls++
	s.mu.Unlock()
	if s.statErr != nil {
		return objectstore.ObjectInfo{}, s.statErr
	}
	if typed, ok := s.Store.(objectstore.TypedStore); ok {
		return typed.Head(ctx, request)
	}
	return s.Store.StatObject(ctx, request.Key)
}

func (s *overrideObjectStore) Get(ctx context.Context, request objectstore.GetObjectRequest) (io.ReadCloser, objectstore.ObjectInfo, error) {
	s.mu.Lock()
	s.readCalls++
	s.mu.Unlock()
	if s.readErr != nil {
		return nil, objectstore.ObjectInfo{}, s.readErr
	}
	if typed, ok := s.Store.(objectstore.TypedStore); ok {
		return typed.Get(ctx, request)
	}
	return s.Store.ReadObject(ctx, request.Key, objectstore.ReadOptions{RangeStart: request.RangeStart, RangeEnd: request.RangeEnd})
}

func (s *overrideObjectStore) ListPrefix(ctx context.Context, request objectstore.ListPrefixRequest) (objectstore.ListPrefixResult, error) {
	if typed, ok := s.Store.(objectstore.TypedStore); ok {
		return typed.ListPrefix(ctx, request)
	}
	objects, err := s.Store.ListObjects(ctx, request.Prefix)
	if err != nil {
		return objectstore.ListPrefixResult{}, err
	}
	return objectstore.ListPrefixResult{Objects: objects}, nil
}

func (s *overrideObjectStore) Delete(ctx context.Context, request objectstore.DeleteObjectRequest) error {
	if typed, ok := s.Store.(objectstore.TypedStore); ok {
		return typed.Delete(ctx, request)
	}
	return s.Store.DeleteObject(ctx, request.Key)
}

func (s *overrideObjectStore) EnsureBucketForDevTest(ctx context.Context, request objectstore.EnsureBucketRequest) (objectstore.EnsureBucketResult, error) {
	if typed, ok := s.Store.(objectstore.TypedStore); ok {
		return typed.EnsureBucketForDevTest(ctx, request)
	}
	return objectstore.EnsureBucketResult{}, &objectstore.AdapterError{
		Code:      objectstore.ErrorCodeInvalidRequest,
		Reason:    objectstore.ReasonInvalidRequest,
		Operation: objectstore.OperationEnsureBucketForDevTest,
		Message:   "typed object-store operation unavailable on override store",
	}
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
