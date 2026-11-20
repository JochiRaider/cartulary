package evidence

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

// serveAdmittedEvidenceRoutes binds every Evidence operation of s for central
// admission behind the registry's route resolution, as production does (Core
// 01 REQ-01-707). The authenticator reads sessions, so a request reaches its
// handler only after authentication and path-parameter admission.
func serveAdmittedEvidenceRoutes(t testing.TB, s *service, sessions httpauth.SessionStore, now time.Time) http.Handler {
	t.Helper()
	registry, err := httpapi.NewRouteRegistry(nil)
	if err != nil {
		t.Fatalf("create route registry: %v", err)
	}
	deps := httpapi.DependencySet{
		PublicRoutes:  registry,
		Authenticator: httpauth.NewAuthenticator(httpauth.Options{Store: sessions, Now: func() time.Time { return now }}),
	}
	mux := http.NewServeMux()
	if err := httpapi.BindAdmittedOwnerRoutes(mux, deps, "module.evidence", s.operationHandlers()); err != nil {
		t.Fatalf("bind Evidence routes: %v", err)
	}
	for operationID, mode := range registry.AdmissionModes() {
		if mode != httpapi.AdmissionCentral {
			t.Fatalf("Evidence operation %q admission mode = %q, want central", operationID, mode)
		}
	}
	return registry.Handler(mux)
}

// visibilityTargetOperations answers the target read that precedes each
// route's visibility decision with a target in incident. The visibility cases
// stop at that decision, so every other operation is unreached.
type visibilityTargetOperations struct {
	routeService
	incidentID uuid.UUID
	sessionID  uuid.UUID
}

func (o visibilityTargetOperations) LoadEvidenceAccess(_ context.Context, recordID uuid.UUID) (evidenceAccessRecord, error) {
	return evidenceAccessRecord{RecordID: recordID, IncidentID: o.incidentID}, nil
}

func (o visibilityTargetOperations) LoadHandle(_ context.Context, token string) (handleRecord, error) {
	return handleRecord{Token: token, IncidentID: o.incidentID, SessionID: o.sessionID}, nil
}

func (o visibilityTargetOperations) GetBlob(_ context.Context, objectBlobID uuid.UUID) (blobRecord, error) {
	return blobRecord{ObjectBlobID: objectBlobID, IncidentID: o.incidentID}, nil
}

func (o visibilityTargetOperations) LoadAttachTarget(context.Context, uuid.UUID) (evidenceAttachTarget, error) {
	return evidenceAttachTarget{IncidentID: o.incidentID, RecordType: "evidence"}, nil
}

// GetUploadLease is the first read after upload admission; the lease is gone.
func (visibilityTargetOperations) GetUploadLease(context.Context, uuid.UUID) (uploadLeaseRecord, error) {
	return uploadLeaseRecord{}, errUploadLeaseNotFound
}

// membershipLostAdmissionDB answers the first incident admission read with a
// visible editor and every later read as if the membership had been removed
// in between, counting the reads.
type membershipLostAdmissionDB struct {
	dbport.DB
	reads *int
}

func (db membershipLostAdmissionDB) QueryRow(context.Context, string, ...any) pgx.Row {
	*db.reads++
	if *db.reads > 1 {
		return failingAdmissionRow{err: pgx.ErrNoRows}
	}
	return visibleEditorRow{}
}

type visibleEditorRow struct{}

func (visibleEditorRow) Scan(dest ...any) error {
	if len(dest) != 2 {
		return errors.New("unexpected incident admission columns")
	}
	*dest[0].(*string) = "editor"
	*dest[1].(*string) = "active"
	return nil
}

// TestEvidenceRouteVisibilityFailuresAreInternal_Unit drives every Evidence
// route that decides incident visibility through the incident of a target it
// has just read: upload, preview and download handle issuance, handle
// redemption and Blob verification. Each request goes through the registry
// binding, so central admission authenticates and admits the path parameter
// first (REQ-01-707). A caller who cannot see the incident keeps the route's
// concealed response, identical in status, code and details (REQ-01-704). An
// injected visibility-read failure is 500 internal_error with empty details
// and no diagnostic, never the concealed response (REQ-01-706).
func TestEvidenceRouteVisibilityFailuresAreInternal_Unit(t *testing.T) {
	keys, err := authn.LoadMasterKeys(nil)
	if err != nil {
		t.Fatalf("load test master keys: %v", err)
	}
	now := time.Date(2026, 11, 18, 9, 30, 0, 0, time.UTC)
	sessions := activeBearerSessions{now: now}
	session, user, err := sessions.GetSessionByFingerprint(context.Background(), nil)
	if err != nil {
		t.Fatalf("read fake session: %v", err)
	}
	incidentID := uuid.MustParse("6b7c8d9e-0f1a-4b2c-9d3e-4f5a6b7c8d01")
	targetID := uuid.MustParse("6b7c8d9e-0f1a-4b2c-9d3e-4f5a6b7c8d02")
	binding, err := objectUploadStorageBinding(keys, "incidents/visibility/object-blobs/visibility")
	if err != nil {
		t.Fatalf("bind upload storage key: %v", err)
	}
	uploadToken, err := encodeObjectUploadToken(keys, objectUploadTokenClaims{
		Version: objectUploadTokenVersion, LeaseID: uuid.New(), ObjectBlobID: targetID, IncidentID: incidentID,
		IssuingUserID: user.ID, IssuingSessionID: session.ID, StorageKeyBinding: binding, ByteSize: 1,
		RequiredMethod: http.MethodPut, RequiredHeadersSHA256: strings.Repeat("a", 64), AcceptedContractSHA256: strings.Repeat("b", 64),
		IssuedAtUnixNano: now.UnixNano(), ExpiresAtUnixNano: now.Add(time.Hour).UnixNano(),
	})
	if err != nil {
		t.Fatalf("encode upload capability: %v", err)
	}
	routes := []struct {
		name, method, path, body string
		concealed                *httpapi.APIError
	}{
		{"uploadObjectBlobContent", http.MethodPut, "/api/v1/object-uploads/" + uploadToken, "x", objectUploadNotFoundOrRevoked()},
		{"issueEvidencePreviewHandle", http.MethodPost, "/api/v1/evidence-records/" + targetID.String() + "/preview-handle", `{}`, evidenceRecordNotFound()},
		{"issueEvidenceDownloadHandle", http.MethodPost, "/api/v1/evidence-records/" + targetID.String() + "/download-handle", `{}`, evidenceRecordNotFound()},
		{"redeemEvidenceHandle", http.MethodGet, "/api/v1/evidence-handles/hdl_visibility", "", handleNotFoundOrRevoked()},
		{"verifyObjectBlob", http.MethodPost, "/api/v1/object-blobs/" + targetID.String() + "/verify", `{"client_txn_id":"txn-visibility"}`, evidenceAttachRejected(AttachReasonBlobNotVisible)},
	}
	internal := httpapi.InternalAPIError(nil)
	cases := []struct {
		name string
		err  error
		want func(concealed *httpapi.APIError) *httpapi.APIError
	}{
		{"incident not visible", pgx.ErrNoRows, func(concealed *httpapi.APIError) *httpapi.APIError { return concealed }},
		{"visibility read failure", errors.New("injected incident admission database failure"), func(*httpapi.APIError) *httpapi.APIError { return internal }},
	}
	for _, route := range routes {
		for _, tc := range cases {
			t.Run(route.name+"/"+tc.name, func(t *testing.T) {
				s := &service{
					operations: visibilityTargetOperations{incidentID: incidentID, sessionID: session.ID},
					admission: routeAdmission{
						incidents: admission.NewChecker(failingAdmissionDB{err: tc.err}),
						auth:      sessions,
						now:       func() time.Time { return now },
					},
					keys: keys,
					now:  func() time.Time { return now },
				}
				request := httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
				request.Header.Set("Authorization", "Bearer evidence-route-visibility")
				request.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()
				serveAdmittedEvidenceRoutes(t, s, sessions, now).ServeHTTP(recorder, request)
				var envelope struct {
					Error struct {
						Code    string         `json:"code"`
						Details map[string]any `json:"details"`
					} `json:"error"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
					t.Fatalf("decode error envelope %q: %v", recorder.Body.String(), err)
				}
				want := tc.want(route.concealed)
				if recorder.Code != want.Status || envelope.Error.Code != want.Code || !reflect.DeepEqual(envelope.Error.Details, want.Details) {
					t.Fatalf("response = %d %s %#v, want %d %s %#v; body=%s",
						recorder.Code, envelope.Error.Code, envelope.Error.Details, want.Status, want.Code, want.Details, recorder.Body.String())
				}
				if strings.Contains(recorder.Body.String(), "injected") {
					t.Fatalf("response leaks the storage diagnostic: %s", recorder.Body.String())
				}
			})
		}
	}

	// Attach, upload and verification decide visibility and the editing role
	// from one admission read. A membership removed while the request runs
	// therefore either precedes that read, giving the route's concealed
	// response (REQ-01-704), or follows it, and the request proceeds to its
	// next step. It never surfaces as incident_not_found, which none of these
	// routes declares (REQ-01-702).
	admitted := []struct {
		name, method, path, body string
		want                     *httpapi.APIError
	}{
		{"attachBlobToEvidenceRecord", http.MethodPost, "/api/v1/evidence-records/" + targetID.String() + "/attach-blob", `{"unexpected":true}`, invalidMutationPayload("unexpected", "unknown_field")},
		{"uploadObjectBlobContent", http.MethodPut, "/api/v1/object-uploads/" + uploadToken, "x", objectUploadNotFoundOrRevoked()},
		{"verifyObjectBlob", http.MethodPost, "/api/v1/object-blobs/" + targetID.String() + "/verify", `{"client_txn_id":"txn-visibility","unexpected":true}`, invalidMutationPayload("unexpected", "unknown_field")},
	}
	for _, route := range admitted {
		t.Run(route.name+"/membership removed after the admission read", func(t *testing.T) {
			reads := 0
			s := &service{
				operations: visibilityTargetOperations{incidentID: incidentID, sessionID: session.ID},
				admission: routeAdmission{
					incidents: admission.NewChecker(membershipLostAdmissionDB{reads: &reads}),
					auth:      sessions,
					now:       func() time.Time { return now },
				},
				keys: keys,
				now:  func() time.Time { return now },
			}
			request := httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
			request.Header.Set("Authorization", "Bearer evidence-route-visibility")
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			serveAdmittedEvidenceRoutes(t, s, sessions, now).ServeHTTP(recorder, request)
			var envelope struct {
				Error struct {
					Code    string         `json:"code"`
					Details map[string]any `json:"details"`
				} `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode error envelope %q: %v", recorder.Body.String(), err)
			}
			if envelope.Error.Code == "incident_not_found" {
				t.Fatalf("membership removed between admission reads answered %d incident_not_found after %d admission reads; want the route's concealed response or its next step", recorder.Code, reads)
			}
			if reads != 1 {
				t.Fatalf("admission made %d incident reads, want one read deciding visibility and role", reads)
			}
			if recorder.Code != route.want.Status || envelope.Error.Code != route.want.Code || !reflect.DeepEqual(envelope.Error.Details, route.want.Details) {
				t.Fatalf("response = %d %s %#v, want %d %s %#v; body=%s",
					recorder.Code, envelope.Error.Code, envelope.Error.Details, route.want.Status, route.want.Code, route.want.Details, recorder.Body.String())
			}
		})
	}
}
