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

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
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
		login := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, admin.email, admin.password, nil)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-object-store-upload-incident",
			"incident_key":  "object-store-upload",
			"title":         "Object-store upload outage",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

		resp := uploadThroughFreshBlobSlot(t, harness, login, incidentID, "txn-object-store-upload-blob")
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

	t.Run("object upload maps access rejection without exposing storage target", func(t *testing.T) {
		baseStore, err := objectstore.NewFilesystemStore(t.TempDir())
		if err != nil {
			t.Fatalf("create filesystem object store: %v", err)
		}
		failingStore := &overrideObjectStore{
			Store: baseStore,
			putErr: &objectstore.AdapterError{
				Code:      objectstore.ErrorCodeAccessRejected,
				Reason:    objectstore.ReasonCredentialDenied,
				Operation: objectstore.OperationPutObject,
				Retryable: false,
				Message:   "credential denied source.internal private-bucket incidents/raw/key CARTULARY_S3_SECRET_KEY=secret",
			},
		}

		harness := startServer(t, "object-store-upload-access", failingStore)
		login := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, admin.email, admin.password, nil)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-object-store-upload-access-incident",
			"incident_key":  "object-store-upload-access",
			"title":         "Object-store upload access rejection",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

		resp := uploadThroughFreshBlobSlot(t, harness, login, incidentID, "txn-object-store-upload-access-blob")
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusServiceUnavailable, "object_store_access_rejected")
		httptestx.RequireErrorDetail(t, body, "reason_code", "credential_denied")
		errorValue := body["error"].(map[string]any)
		if retryable, ok := errorValue["retryable"].(bool); !ok || retryable {
			t.Fatalf("object_store_access_rejected retryable got %#v want false", errorValue["retryable"])
		}
		requireNoPublicEvidenceLeak(t, "object upload access rejection", body, []string{
			"source.internal",
			"private-bucket",
			"incidents/raw/key",
			"CARTULARY_S3_SECRET_KEY",
			"secret",
		})
		if got := countObjectBlobs(t, harness, incidentID); got != 1 {
			t.Fatalf("app-mediated upload access rejection should retain the pending blob slot: got %d want 1", got)
		}
	})

	t.Run("malformed persisted key after issuance fails the capability binding before backend", func(t *testing.T) {
		baseStore, err := objectstore.NewFilesystemStore(t.TempDir())
		if err != nil {
			t.Fatalf("create filesystem object store: %v", err)
		}
		recordingStore := &overrideObjectStore{Store: baseStore}
		harness := startServer(t, "object-key-upload-binding", recordingStore)
		login := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, admin.email, admin.password, nil)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-object-key-upload-binding-incident",
			"incident_key":  "object-key-upload-binding",
			"title":         "Object key upload binding",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		// The server issues a genuine capability bound to the key it generated.
		// A key that changed afterwards no longer matches that binding, so Core 04
		// Table 2-B gate 7 rejects the upload before any object-store work.
		objectBlobID, href := createFreshBlobSlot(t, harness, login, incidentID, "txn-object-key-upload-binding-blob")
		malformedKey := "objects/not-canonical-upload"
		updateBlobStorageKey(t, harness, objectBlobID, malformedKey)
		recordingStore.resetCounts()

		resp := putThroughUploadTarget(t, harness, login, href)
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "object_upload_rejected")
		httptestx.RequireErrorDetail(t, body, "reason_code", "upload_contract_mismatch")
		requireNoPublicEvidenceLeak(t, "upload with a key changed after issuance", body, []string{malformedKey})
		if puts := recordingStore.putCount(); puts != 0 {
			t.Fatalf("upload with a key changed after issuance reached the backend: put=%d", puts)
		}
		var uploadState, leaseState string
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT b.upload_state, l.lease_state
  FROM object_blobs b
  JOIN evidence_object_upload_leases l ON l.object_blob_id = b.object_blob_id
 WHERE b.object_blob_id = $1
`, objectBlobID).Scan(&uploadState, &leaseState); err != nil {
			t.Fatalf("load rejected upload slot: %v", err)
		}
		if uploadState != "pending" || leaseState != "issued" {
			t.Fatalf("rejected upload changed the slot: upload_state=%s lease_state=%s, want pending/issued", uploadState, leaseState)
		}
	})

	t.Run("handle issuance maps every object-store failure class for both handle kinds", func(t *testing.T) {
		baseStore, err := objectstore.NewFilesystemStore(t.TempDir())
		if err != nil {
			t.Fatalf("create filesystem object store: %v", err)
		}
		faultStore := &overrideObjectStore{Store: baseStore}
		harness := startServer(t, "object-store-issue-failures", faultStore)
		login := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, admin.email, admin.password, nil)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-object-store-issue-failures-incident",
			"incident_key":  "object-store-issue-failures",
			"title":         "Object-store handle issuance failures",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, admin.userID, recordID)
		objectBlobID := linkSeededBlobWithCanonicalStorageKey(t, harness, incidentID, admin.userID, recordID, "available", "available")
		leaked := []string{"target.internal", "private-bucket", "incidents/raw/key", "AWS_SECRET_ACCESS_KEY", "secret"}
		beforeHandles := countAccessHandles(t, harness, incidentID)

		failures := []struct {
			name      string
			fault     *objectstore.AdapterError
			code      string
			reason    string
			retryable bool
		}{
			{
				name: "endpoint outage",
				fault: &objectstore.AdapterError{
					Code:      objectstore.ErrorCodeUnavailable,
					Reason:    objectstore.ReasonEndpointUnreachable,
					Operation: objectstore.OperationHeadObject,
					Retryable: true,
					Message:   "endpoint unreachable target.internal private-bucket incidents/raw/key AWS_SECRET_ACCESS_KEY=secret",
				},
				code: "object_store_unavailable", reason: "endpoint_unreachable", retryable: true,
			},
			{
				name: "credential denial",
				fault: &objectstore.AdapterError{
					Code:      objectstore.ErrorCodeAccessRejected,
					Reason:    objectstore.ReasonCredentialDenied,
					Operation: objectstore.OperationHeadObject,
					Message:   "credential denied target.internal private-bucket incidents/raw/key AWS_SECRET_ACCESS_KEY=secret",
				},
				code: "object_store_access_rejected", reason: "credential_denied", retryable: false,
			},
		}
		for _, failure := range failures {
			faultStore.setFaults(failure.fault, nil)
			for _, endpoint := range []string{"preview-handle", "download-handle"} {
				t.Run(failure.name+" "+endpoint, func(t *testing.T) {
					resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...)
					body := httptestx.RequireErrorEnvelope(t, resp, http.StatusServiceUnavailable, failure.code)
					httptestx.RequireErrorDetail(t, body, "reason_code", failure.reason)
					errorValue := body["error"].(map[string]any)
					if retryable, ok := errorValue["retryable"].(bool); !ok || retryable != failure.retryable {
						t.Fatalf("%s retryable got %#v want %t", failure.code, errorValue["retryable"], failure.retryable)
					}
					requireNoPublicEvidenceLeak(t, endpoint+" "+failure.name, body, leaked)
				})
			}
		}

		t.Run("malformed persisted key download-handle", func(t *testing.T) {
			faultStore.setFaults(nil, nil)
			// object_blobs.storage_key is unique across the shared database.
			malformedKey := "objects/not-canonical-download-issuance"
			updateBlobStorageKey(t, harness, objectBlobID, malformedKey)
			faultStore.resetCounts()
			resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/download-handle", map[string]any{}, authOptions(login)...)
			body := requireObjectStoreInvalidRequestReason(t, resp, "object_blob_storage_key_malformed")
			requireNoPublicEvidenceLeak(t, "download issuance malformed persisted key", body, []string{malformedKey})
			if statCalls, readCalls := faultStore.counts(); statCalls != 0 || readCalls != 0 {
				t.Fatalf("malformed storage key reached backend on download issuance: stat=%d read=%d", statCalls, readCalls)
			}
		})
		if got := countAccessHandles(t, harness, incidentID); got != beforeHandles {
			t.Fatalf("failed issuance wrote evidence_access_handles: got %d want %d", got, beforeHandles)
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
		login := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, admin.email, admin.password, nil)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-object-store-preview-incident",
			"incident_key":  "object-store-preview",
			"title":         "Object-store preview outage",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, admin.userID, recordID)
		linkSeededBlobWithCanonicalStorageKey(t, harness, incidentID, admin.userID, recordID, "available", "available")

		resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...)
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
		login := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, admin.email, admin.password, nil)
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

		resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...)
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
		login := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, admin.email, admin.password, nil)
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
		mismatchedKey, err := blobref.ObjectBlobStorageKey(incidentID, uuid.New(), uuid.New())
		if err != nil {
			t.Fatalf("mismatched storage key: %v", err)
		}
		updateBlobStorageKey(t, harness, objectBlobID, mismatchedKey)
		recordingStore.resetCounts()

		resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, httptestx.WithCookies(login.Session))
		body := requireObjectStoreInvalidRequestReason(t, resp, "object_blob_storage_key_identity_mismatch")
		requireNoPublicEvidenceLeak(t, "identity-mismatched persisted key error", body, []string{mismatchedKey})
		if statCalls, readCalls := recordingStore.counts(); statCalls != 0 || readCalls != 0 {
			t.Fatalf("identity-mismatched storage key reached backend: stat=%d read=%d", statCalls, readCalls)
		}
	})
}

func TestPublicEvidenceResponsesDoNotLeakObjectStoreIdentifiers(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "evidence-redaction", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
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
	storageKey := blobStorageKey(t, harness, objectBlobID)
	forbidden := []string{
		storageKey,
		"evidence.storage_ref",
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

	createResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-evidence-redaction-extra-blob",
		"byte_size":         1,
		"filename_hint":     "extra.txt",
		"content_type_hint": "text/plain",
	}, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	extraBlobID := appsupport.MustUUID(t, createData["object_blob_id"].(string))
	forbidden = append(forbidden, blobStorageKey(t, harness, extraBlobID))
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

// uploadThroughFreshBlobSlot creates a four-byte pending blob slot, checks
// that its upload target stays opaque, and PUTs the bytes through it.
func uploadThroughFreshBlobSlot(t *testing.T, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, incidentID uuid.UUID, clientTxnID string) *http.Response {
	t.Helper()
	_, href := createFreshBlobSlot(t, harness, login, incidentID, clientTxnID)
	return putThroughUploadTarget(t, harness, login, href)
}

// createFreshBlobSlot creates a four-byte pending blob slot through the public
// route and returns its identity and opaque upload target.
func createFreshBlobSlot(t *testing.T, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, incidentID uuid.UUID, clientTxnID string) (uuid.UUID, string) {
	t.Helper()
	createResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     clientTxnID,
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
	return appsupport.MustUUID(t, createData["object_blob_id"].(string)), href
}

// putThroughUploadTarget PUTs four text bytes through an issued upload target.
func putThroughUploadTarget(t *testing.T, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, href string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, harness.Server.HTTP.URL+href, strings.NewReader("fail"))
	if err != nil {
		t.Fatalf("create object upload request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")
	for _, option := range authOptions(login) {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
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
	putCalls        int
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
	if putErr := s.putFault(); putErr != nil {
		return putErr
	}
	return s.Store.PutObject(ctx, key, body, size, contentType)
}

func (s *overrideObjectStore) StatObject(ctx context.Context, key string) (objectstore.ObjectInfo, error) {
	s.mu.Lock()
	s.statCalls++
	statErr := s.statErr
	s.mu.Unlock()
	if statErr != nil {
		return objectstore.ObjectInfo{}, statErr
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
	if putErr := s.putFault(); putErr != nil {
		return objectstore.PutObjectResult{}, putErr
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
	statErr := s.statErr
	s.mu.Unlock()
	if statErr != nil {
		return objectstore.ObjectInfo{}, statErr
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

// setFaults replaces the injected Head/Stat and Put failures while a server
// holds the store; nil restores the wrapped behavior.
func (s *overrideObjectStore) setFaults(statErr, putErr error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statErr = statErr
	s.putErr = putErr
}

// putFault counts one backend write and returns the injected Put failure.
func (s *overrideObjectStore) putFault() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.putCalls++
	return s.putErr
}

func (s *overrideObjectStore) resetCounts() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statCalls = 0
	s.readCalls = 0
	s.putCalls = 0
}

func (s *overrideObjectStore) putCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.putCalls
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
