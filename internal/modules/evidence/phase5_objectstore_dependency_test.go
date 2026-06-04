package evidence_test

import (
	"context"
	"io"
	"net/http"
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
				Message:   "endpoint unreachable",
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
				Message:   "capability missing",
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
	})
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
