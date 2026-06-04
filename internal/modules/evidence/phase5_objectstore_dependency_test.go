package evidence_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func requirePhaseDObjectStoreDependencyErrorsUseOwnerPublicMapping(t *testing.T) {
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

		runtime := phase4test.StartRuntime(t)
		harness := runtime.StartServerWithObjectStore(t, "phase-d-object-store-create-outage", failingStore)
		login, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
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

		runtime := phase4test.StartRuntime(t)
		harness := runtime.StartServerWithObjectStore(t, "phase-d-object-store-preview-capability", failingStore)
		login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-phase-d-object-store-preview-incident",
			"incident_key":  "phase-d-object-store-preview",
			"title":         "Phase D object-store preview outage",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
		linkSeededBlob(t, harness, incidentID, adminID, recordID, "available", "available", "phase-d/object-store-preview")

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
}

func (s *overrideObjectStore) UploadTarget(ctx context.Context, key string, expiresAt time.Time) (objectstore.UploadTarget, error) {
	if s.uploadTargetErr != nil {
		return objectstore.UploadTarget{}, s.uploadTargetErr
	}
	return s.Store.UploadTarget(ctx, key, expiresAt)
}

func (s *overrideObjectStore) StatObject(ctx context.Context, key string) (objectstore.ObjectInfo, error) {
	if s.statErr != nil {
		return objectstore.ObjectInfo{}, s.statErr
	}
	return s.Store.StatObject(ctx, key)
}

func (s *overrideObjectStore) ReadObject(ctx context.Context, key string, options objectstore.ReadOptions) (io.ReadCloser, objectstore.ObjectInfo, error) {
	if s.readErr != nil {
		return nil, objectstore.ObjectInfo{}, s.readErr
	}
	return s.Store.ReadObject(ctx, key, options)
}
