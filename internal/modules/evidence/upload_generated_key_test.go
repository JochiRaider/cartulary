package evidence

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

// TestObjectUploadMalformedGeneratedKeyFailsBeforeBackend_Unit covers a
// malformed server-generated physical key (Core 01 object_blob_storage_key_v2).
// The production create route generates the key, persists it with the slot
// and mints the capability bound to it, so nothing is forged: the upload
// passes every capability binding (Core 04 Table 2-B gate 7) and then fails
// locally with 500 object_store_invalid_request before any object-store Put.
func TestObjectUploadMalformedGeneratedKeyFailsBeforeBackend_Unit(t *testing.T) {
	keys, err := authn.LoadMasterKeys(nil)
	if err != nil {
		t.Fatalf("load test master keys: %v", err)
	}
	now := time.Date(2026, 11, 19, 9, 30, 0, 0, time.UTC)
	sessions := activeBearerSessions{now: now}
	incidentID := uuid.MustParse("7c8d9e0f-1a2b-4c3d-8e4f-5a6b7c8d9e01")
	cases := []struct {
		name     string
		generate func(incidentID, objectBlobID uuid.UUID) (string, error)
		reason   string
	}{
		{name: "malformed generated key", reason: "object_blob_storage_key_malformed", generate: func(uuid.UUID, uuid.UUID) (string, error) {
			return "objects/not-canonical-upload", nil
		}},
		{name: "identity-mismatched generated key", reason: "object_blob_storage_key_identity_mismatch", generate: func(incidentID, _ uuid.UUID) (string, error) {
			return blobref.ObjectBlobStorageKey(incidentID, uuid.New(), uuid.New())
		}},
		// The production generator shows the same issuance reaches the backend.
		{name: "production generator"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			slots := &generatedKeySlotOperations{}
			objects := &countingPutStore{}
			s := &service{
				operations: slots,
				admission: routeAdmission{
					incidents: admission.NewChecker(visibleEditorAdmissionDB{}),
					auth:      sessions,
					now:       func() time.Time { return now },
				},
				objects:      routeObjectStoreAdapter{store: objects},
				uploads:      uploadCapabilityService{keys: keys},
				keys:         keys,
				now:          func() time.Time { return now },
				maxBlobBytes: 1024,
				storageKeys:  tc.generate,
			}
			routes := serveAdmittedEvidenceRoutes(t, s, sessions, now)

			create := httptest.NewRequest(http.MethodPost, "/api/v1/object-blobs", strings.NewReader(
				`{"incident_id":"`+incidentID.String()+`","client_txn_id":"txn-generated-key","byte_size":4,"content_type_hint":"text/plain"}`))
			create.Header.Set("Authorization", "Bearer evidence-generated-key")
			create.Header.Set("Content-Type", "application/json")
			created := httptest.NewRecorder()
			routes.ServeHTTP(created, create)
			var slot struct {
				Data struct {
					UploadTarget struct {
						Href string `json:"href"`
					} `json:"upload_target"`
				} `json:"data"`
			}
			if created.Code != http.StatusCreated || json.Unmarshal(created.Body.Bytes(), &slot) != nil || slots.params == nil {
				t.Fatalf("blob slot create = %d %s", created.Code, created.Body.String())
			}
			storageKey := slots.params.StorageKey

			upload := httptest.NewRequest(http.MethodPut, slot.Data.UploadTarget.Href, strings.NewReader("fail"))
			upload.Header.Set("Authorization", "Bearer evidence-generated-key")
			upload.Header.Set("Content-Type", "text/plain")
			uploaded := httptest.NewRecorder()
			routes.ServeHTTP(uploaded, upload)

			if tc.generate == nil {
				if uploaded.Code != http.StatusNoContent || objects.puts != 1 || slots.claims != 1 {
					t.Fatalf("canonical upload = %d (put=%d claim=%d) %s, want 204 through one Put", uploaded.Code, objects.puts, slots.claims, uploaded.Body.String())
				}
				return
			}
			if uploaded.Code != http.StatusInternalServerError || objects.puts != 0 || slots.claims != 0 {
				t.Fatalf("generated-key upload = %d (put=%d claim=%d) %s, want 500 before the lease claim and any Put", uploaded.Code, objects.puts, slots.claims, uploaded.Body.String())
			}
			var envelope struct {
				Error struct {
					Code      string         `json:"code"`
					Retryable bool           `json:"retryable"`
					Details   map[string]any `json:"details"`
				} `json:"error"`
			}
			if err := json.Unmarshal(uploaded.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode upload error %q: %v", uploaded.Body.String(), err)
			}
			if envelope.Error.Code != "object_store_invalid_request" || envelope.Error.Details["reason_code"] != tc.reason || envelope.Error.Retryable {
				t.Fatalf("upload error = %s %#v retryable=%t, want object_store_invalid_request %s", envelope.Error.Code, envelope.Error.Details, envelope.Error.Retryable, tc.reason)
			}
			if strings.Contains(uploaded.Body.String(), storageKey) {
				t.Fatalf("upload error leaked the physical key: %s", uploaded.Body.String())
			}
		})
	}
}

// generatedKeySlotOperations keeps the slot and lease the create route
// writes and serves them back to the upload route unchanged, as persistence
// would. The capability is the one the create route minted.
type generatedKeySlotOperations struct {
	routeService
	params *blobSlotParams
	claims int
}

func (o *generatedKeySlotOperations) CreateBlobSlot(_ context.Context, params blobSlotParams) (blobSlotResult, error) {
	o.params = &params
	return blobSlotResult{StatusCode: http.StatusCreated, Payload: map[string]any{
		"incident_id":        params.IncidentID.String(),
		"object_blob_id":     params.ObjectBlobID.String(),
		"upload_state":       "pending",
		"target_expires_at":  formatHTTPTime(params.TargetExpiresAt),
		"pending_expires_at": formatHTTPTime(params.PendingExpiresAt),
		"upload_target":      params.UploadTarget,
		"accepted_contract":  params.AcceptedContract,
	}}, nil
}

func (o *generatedKeySlotOperations) GetUploadLease(_ context.Context, leaseID uuid.UUID) (uploadLeaseRecord, error) {
	if o.params == nil || o.params.UploadLease.LeaseID != leaseID {
		return uploadLeaseRecord{}, errUploadLeaseNotFound
	}
	slot, lease := *o.params, o.params.UploadLease
	return uploadLeaseRecord{
		LeaseID: lease.LeaseID, ObjectBlobID: slot.ObjectBlobID, IncidentID: slot.IncidentID,
		CapabilityHash: lease.CapabilityHash, IssuingUserID: lease.IssuingUserID, IssuingSessionID: lease.IssuingSessionID,
		IssuedAt: lease.IssuedAt, ExpiresAt: lease.ExpiresAt, RequiredMethod: lease.RequiredMethod,
		RequiredHeaders: lease.RequiredHeaders, AcceptedContractSHA256: lease.AcceptedContractSHA256, LeaseState: "issued",
		Blob: blobRecord{
			ObjectBlobID: slot.ObjectBlobID, IncidentID: slot.IncidentID, StorageKey: slot.StorageKey, UploadState: "pending",
			ByteSize: slot.ByteSize, FilenameHint: slot.FilenameHint, ContentTypeHint: slot.ContentTypeHint,
			ExpectedSHA256Hex: slot.ExpectedSHA256Hex, TargetExpiresAt: slot.TargetExpiresAt,
			PendingExpiresAt: slot.PendingExpiresAt, UploadLeaseState: "issued",
		},
	}, nil
}

func (o *generatedKeySlotOperations) ClaimUploadLease(context.Context, uuid.UUID, []byte, time.Time) error {
	o.claims++
	return nil
}

func (o *generatedKeySlotOperations) CompleteUploadLease(context.Context, uuid.UUID, time.Time) error {
	return nil
}

// countingPutStore counts backend writes; no other object-store operation is
// reachable from the upload route.
type countingPutStore struct {
	objectstore.TypedStore
	puts int
}

func (s *countingPutStore) Put(_ context.Context, request objectstore.PutObjectRequest) (objectstore.PutObjectResult, error) {
	s.puts++
	if _, err := io.Copy(io.Discard, request.Body); err != nil {
		return objectstore.PutObjectResult{}, err
	}
	return objectstore.PutObjectResult{SizeBytes: request.Size, ContentType: request.ContentType}, nil
}

// visibleEditorAdmissionDB answers every incident admission read with an
// active incident the caller edits.
type visibleEditorAdmissionDB struct{ dbport.DB }

func (visibleEditorAdmissionDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return visibleEditorRow{}
}
