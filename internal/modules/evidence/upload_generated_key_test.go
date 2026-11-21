package evidence

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
// Slot creation validates the key it generates (Core 01 REQ-01-246), so a
// malformed key fails with 500 object_store_invalid_request before any slot is
// persisted or capability minted. The production generator's key reaches the
// backend through one Put.
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
			s := generatedKeyService(keys, sessions, now, slots, objects, visibleEditorAdmissionDB{})
			s.storageKeys = tc.generate
			routes := serveAdmittedEvidenceRoutes(t, s, sessions, now)
			created := createGeneratedKeySlot(t, routes, incidentID)
			if tc.generate != nil {
				if created.Code != http.StatusInternalServerError || slots.params != nil {
					t.Fatalf("blob slot create = %d (slot persisted=%t) %s, want 500 before any slot", created.Code, slots.params != nil, created.Body.String())
				}
				requireObjectStoreInvalidRequest(t, created, tc.reason)
				return
			}
			href := createdUploadHref(t, created, slots)
			uploaded := putGeneratedKeyUpload(routes, href, func(*http.Request) {})
			if uploaded.Code != http.StatusNoContent || objects.puts != 1 || slots.claims != 1 {
				t.Fatalf("canonical upload = %d (put=%d claim=%d) %s, want 204 through one Put", uploaded.Code, objects.puts, slots.claims, uploaded.Body.String())
			}
		})
	}
}

// TestObjectUploadPersistedKeyValidatedAtObjectStoreBoundary_Unit covers a
// persisted slot whose physical key is malformed although its capability is
// valid, as slots issued before creation-time validation may be. Core 04 Table
// 2-B gates 8 and 9 come first, then the key fails with 500
// object_store_invalid_request immediately before the lease claim and the
// object-store Put (Core 01 REQ-01-246, Table 2-B gate 11).
func TestObjectUploadPersistedKeyValidatedAtObjectStoreBoundary_Unit(t *testing.T) {
	keys, err := authn.LoadMasterKeys(nil)
	if err != nil {
		t.Fatalf("load test master keys: %v", err)
	}
	now := time.Date(2026, 11, 19, 9, 30, 0, 0, time.UTC)
	sessions := activeBearerSessions{now: now}
	incidentID := uuid.MustParse("7c8d9e0f-1a2b-4c3d-8e4f-5a6b7c8d9e01")
	for _, tc := range []struct {
		name   string
		closed bool
		adjust func(*http.Request)
		status int
		code   string
	}{
		{name: "unknown content length", adjust: func(r *http.Request) { r.ContentLength = -1 }, status: http.StatusLengthRequired, code: "object_upload_length_required"},
		{name: "wrong required header", adjust: func(r *http.Request) { r.Header.Set("Content-Type", "application/octet-stream") }, status: http.StatusBadRequest, code: "object_upload_invalid"},
		{name: "closed incident", closed: true, adjust: func(*http.Request) {}, status: http.StatusConflict, code: "incident_closed"},
		{name: "otherwise valid upload", adjust: func(*http.Request) {}, status: http.StatusInternalServerError, code: "object_store_invalid_request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			slots := &generatedKeySlotOperations{}
			objects := &countingPutStore{}
			closed := false
			s := generatedKeyService(keys, sessions, now, slots, objects, closableEditorAdmissionDB{closed: &closed})
			routes := serveAdmittedEvidenceRoutes(t, s, sessions, now)
			href := createdUploadHref(t, createGeneratedKeySlot(t, routes, incidentID), slots)
			token, err := url.PathUnescape(strings.TrimPrefix(href, "/api/v1/object-uploads/"))
			if err != nil {
				t.Fatal(err)
			}
			claims, err := decodeObjectUploadToken(keys, token)
			if err != nil {
				t.Fatal(err)
			}
			const malformed = "objects/not-canonical-upload"
			rebound, err := s.uploads.createTarget(claims, malformed)
			if err != nil {
				t.Fatal(err)
			}
			slots.params.StorageKey = malformed
			slots.params.UploadLease.CapabilityHash = objectUploadTokenDigest(rebound.Token)
			closed = tc.closed
			uploaded := putGeneratedKeyUpload(routes, rebound.Target.Href, tc.adjust)
			if uploaded.Code != tc.status || objects.puts != 0 || slots.claims != 0 {
				t.Fatalf("persisted malformed key upload = %d (put=%d claim=%d) %s, want %d with no claim or Put", uploaded.Code, objects.puts, slots.claims, uploaded.Body.String(), tc.status)
			}
			var envelope struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(uploaded.Body.Bytes(), &envelope); err != nil || envelope.Error.Code != tc.code {
				t.Fatalf("upload error = %q (%v), want %s", uploaded.Body.String(), err, tc.code)
			}
			if tc.code == "object_store_invalid_request" {
				requireObjectStoreInvalidRequest(t, uploaded, "object_blob_storage_key_malformed")
			}
		})
	}
}

func generatedKeyService(keys authn.MasterKeys, sessions activeBearerSessions, now time.Time, slots *generatedKeySlotOperations, objects *countingPutStore, incidents dbport.DB) *service {
	return &service{
		operations: slots,
		admission: routeAdmission{
			incidents: admission.NewChecker(incidents),
			auth:      sessions,
			now:       func() time.Time { return now },
		},
		objects:      routeObjectStoreAdapter{store: objects},
		uploads:      uploadCapabilityService{keys: keys},
		keys:         keys,
		now:          func() time.Time { return now },
		maxBlobBytes: 1024,
	}
}

func createGeneratedKeySlot(t *testing.T, routes http.Handler, incidentID uuid.UUID) *httptest.ResponseRecorder {
	t.Helper()
	create := httptest.NewRequest(http.MethodPost, "/api/v1/object-blobs", strings.NewReader(
		`{"incident_id":"`+incidentID.String()+`","client_txn_id":"txn-generated-key","byte_size":4,"content_type_hint":"text/plain"}`))
	create.Header.Set("Authorization", "Bearer evidence-generated-key")
	create.Header.Set("Content-Type", "application/json")
	created := httptest.NewRecorder()
	routes.ServeHTTP(created, create)
	return created
}

func createdUploadHref(t *testing.T, created *httptest.ResponseRecorder, slots *generatedKeySlotOperations) string {
	t.Helper()
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
	return slot.Data.UploadTarget.Href
}

func putGeneratedKeyUpload(routes http.Handler, href string, adjust func(*http.Request)) *httptest.ResponseRecorder {
	upload := httptest.NewRequest(http.MethodPut, href, strings.NewReader("fail"))
	upload.Header.Set("Authorization", "Bearer evidence-generated-key")
	upload.Header.Set("Content-Type", "text/plain")
	adjust(upload)
	uploaded := httptest.NewRecorder()
	routes.ServeHTTP(uploaded, upload)
	return uploaded
}

func requireObjectStoreInvalidRequest(t *testing.T, response *httptest.ResponseRecorder, reason string) {
	t.Helper()
	var envelope struct {
		Error struct {
			Code      string         `json:"code"`
			Retryable bool           `json:"retryable"`
			Details   map[string]any `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error %q: %v", response.Body.String(), err)
	}
	if envelope.Error.Code != "object_store_invalid_request" || envelope.Error.Details["reason_code"] != reason || envelope.Error.Retryable {
		t.Fatalf("error = %s %#v retryable=%t, want object_store_invalid_request %s", envelope.Error.Code, envelope.Error.Details, envelope.Error.Retryable, reason)
	}
	if strings.Contains(response.Body.String(), "objects/") {
		t.Fatalf("error leaked a physical key: %s", response.Body.String())
	}
}

// closableEditorAdmissionDB answers incident admission reads for an incident
// the caller edits, active or closed as the test sets it.
type closableEditorAdmissionDB struct {
	dbport.DB
	closed *bool
}

func (db closableEditorAdmissionDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return closableEditorRow{closed: *db.closed}
}

type closableEditorRow struct{ closed bool }

func (row closableEditorRow) Scan(dest ...any) error {
	if err := (visibleEditorRow{}).Scan(dest...); err != nil {
		return err
	}
	if row.closed {
		*dest[1].(*string) = "closed"
	}
	return nil
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
