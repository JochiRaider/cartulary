package evidence_test

// Evidence attachment and authorization contracts.
import (
	"context"
	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"testing"
)

func TestEvidenceAuthorizationMatrixAC543_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence-authorization-matrix-ac543")
	bootstrapLogin, bootstrapAdminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, bootstrapLogin, map[string]any{
		"client_txn_id": "txn-evidence-ac543-incident",
		"incident_key":  "evidence-ac543",
		"title":         "Evidence authorization matrix AC-543",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	uploader := authflowtest.SeedLocalUserRecord(t, harness.DB, "evidence-ac543-uploader@example.test", "Evidence AC543 Uploader", "EvidenceAC543Uploader1!", false, true, true)
	incidentstoretest.SeedMembership(t, harness.DB, incidentID, uploader.ID, "Evidence AC543 Uploader", "editor", bootstrapAdminID)
	issuingLogin := loginLocalUserNoMFA(t, harness, "evidence-ac543-uploader@example.test", "EvidenceAC543Uploader1!")
	otherSessionLogin := loginLocalUserNoMFA(t, harness, "evidence-ac543-uploader@example.test", "EvidenceAC543Uploader1!")

	type slot struct {
		blobID     uuid.UUID
		href       string
		storageKey string
		headers    map[string]string
	}
	createSlot := func(t *testing.T, suffix string, login appsupport.LoginResult) slot {
		t.Helper()
		response := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
			"incident_id":       incidentID.String(),
			"client_txn_id":     "txn-evidence-ac543-" + suffix,
			"byte_size":         5,
			"filename_hint":     suffix + ".txt",
			"content_type_hint": "text/plain",
		}, authOptions(login)...)
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusCreated)["data"].(map[string]any)
		blobID := appsupport.MustUUID(t, data["object_blob_id"].(string))
		target := data["upload_target"].(map[string]any)
		headers := map[string]string{}
		for name, value := range target["headers"].(map[string]any) {
			headers[name] = value.(string)
		}
		var storageKey string
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT storage_key
  FROM object_blobs
 WHERE object_blob_id = $1
`, blobID).Scan(&storageKey); err != nil {
			t.Fatalf("load AC-543 storage key: %v", err)
		}
		return slot{blobID: blobID, href: target["href"].(string), storageKey: storageKey, headers: headers}
	}
	put := func(t *testing.T, candidate slot, login *appsupport.LoginResult, contentType string, includeCSRF bool) *http.Response {
		t.Helper()
		request, err := http.NewRequest(http.MethodPut, harness.Server.HTTP.URL+candidate.href, strings.NewReader("hello"))
		if err != nil {
			t.Fatalf("create AC-543 upload request: %v", err)
		}
		for name, value := range candidate.headers {
			request.Header.Set(name, value)
		}
		if contentType != "" {
			request.Header.Set("Content-Type", contentType)
		}
		if login != nil {
			if includeCSRF {
				for _, option := range authOptions(*login) {
					option(request)
				}
			} else {
				appsupport.WithCookies(login.SessionCookie, login.CSRFCookie)(request)
			}
		}
		return httptestx.Do(t, http.DefaultClient, request)
	}
	requireLeaseState := func(t *testing.T, blobID uuid.UUID, want string) {
		t.Helper()
		var got string
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT lease_state
  FROM evidence_object_upload_leases
 WHERE object_blob_id = $1
`, blobID).Scan(&got); err != nil {
			t.Fatalf("load AC-543 lease state: %v", err)
		}
		if got != want {
			t.Fatalf("AC-543 lease state = %q, want %q", got, want)
		}
	}
	requireObjectCount := func(t *testing.T, candidate slot, want int) {
		t.Helper()
		objects, err := harness.ObjectStore.ListObjects(context.Background(), candidate.storageKey)
		if err != nil {
			t.Fatalf("list AC-543 objects: %v", err)
		}
		if len(objects) != want {
			t.Fatalf("AC-543 object count for %q = %d, want %d", candidate.storageKey, len(objects), want)
		}
	}
	setRole := func(t *testing.T, role string) {
		t.Helper()
		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = $3,
       membership_version = membership_version + 1
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, uploader.ID, role); err != nil {
			t.Fatalf("set AC-543 role %q: %v", role, err)
		}
	}

	t.Run("authentication precedes opaque token diagnostics", func(t *testing.T) {
		unknown := slot{href: "/api/v1/object-uploads/not-a-token", headers: map[string]string{"Content-Type": "text/plain"}}
		httptestx.RequireErrorEnvelope(t, put(t, unknown, nil, "", false), http.StatusUnauthorized, "session_required")
	})
	t.Run("CSRF precedes capability diagnostics", func(t *testing.T) {
		unknown := slot{href: "/api/v1/object-uploads/not-a-token", headers: map[string]string{"Content-Type": "text/plain"}}
		httptestx.RequireErrorEnvelope(t, put(t, unknown, &issuingLogin, "", false), http.StatusForbidden, "csrf_verification_failed")
	})
	t.Run("current role denial leaves lease and object untouched", func(t *testing.T) {
		candidate := createSlot(t, "role-denied", issuingLogin)
		setRole(t, "viewer")
		httptestx.RequireErrorEnvelope(t, put(t, candidate, &issuingLogin, "", true), http.StatusForbidden, "authorization_denied")
		requireLeaseState(t, candidate.blobID, "issued")
		requireObjectCount(t, candidate, 0)
		setRole(t, "editor")
	})
	t.Run("cross-session capability is concealed without effects", func(t *testing.T) {
		candidate := createSlot(t, "cross-session", issuingLogin)
		httptestx.RequireErrorEnvelope(t, put(t, candidate, &otherSessionLogin, "", true), http.StatusNotFound, "object_upload_not_found_or_revoked")
		requireLeaseState(t, candidate.blobID, "issued")
		requireObjectCount(t, candidate, 0)
	})
	t.Run("required-header mismatch precedes claim and storage", func(t *testing.T) {
		candidate := createSlot(t, "header-mismatch", issuingLogin)
		body := httptestx.RequireErrorEnvelope(t, put(t, candidate, &issuingLogin, "application/json", true), http.StatusBadRequest, "object_upload_rejected")
		httptestx.RequireErrorDetail(t, body, "reason_code", "upload_contract_mismatch")
		requireLeaseState(t, candidate.blobID, "issued")
		requireObjectCount(t, candidate, 0)
	})
	t.Run("membership loss defeats deployment admin and conceals capability", func(t *testing.T) {
		candidate := createSlot(t, "membership-loss", issuingLogin)
		if _, err := harness.DB.ExecContext(context.Background(), `
DELETE FROM incident_memberships
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, uploader.ID); err != nil {
			t.Fatalf("remove AC-543 membership: %v", err)
		}
		httptestx.RequireErrorEnvelope(t, put(t, candidate, &issuingLogin, "", true), http.StatusNotFound, "object_upload_not_found_or_revoked")
		requireLeaseState(t, candidate.blobID, "issued")
		requireObjectCount(t, candidate, 0)
		incidentstoretest.SeedMembership(t, harness.DB, incidentID, uploader.ID, "Evidence AC543 Uploader", "editor", bootstrapAdminID)
	})
	t.Run("closed incident rejects before claim and storage", func(t *testing.T) {
		candidate := createSlot(t, "closed-incident", issuingLogin)
		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incidents
   SET status = 'closed',
       closed_at = now()
 WHERE id = $1
`, incidentID); err != nil {
			t.Fatalf("close AC-543 incident: %v", err)
		}
		httptestx.RequireErrorEnvelope(t, put(t, candidate, &issuingLogin, "", true), http.StatusConflict, "incident_closed")
		requireLeaseState(t, candidate.blobID, "issued")
		requireObjectCount(t, candidate, 0)
		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incidents
   SET status = 'active',
       closed_at = NULL
 WHERE id = $1
`, incidentID); err != nil {
			t.Fatalf("reopen AC-543 incident: %v", err)
		}
	})
	for _, role := range []string{"editor", "reviewer", "admin"} {
		t.Run(role+" may use one capability exactly once", func(t *testing.T) {
			setRole(t, role)
			candidate := createSlot(t, "allowed-"+role, issuingLogin)
			httptestx.RequireStatus(t, put(t, candidate, &issuingLogin, "", true), http.StatusNoContent)
			requireLeaseState(t, candidate.blobID, "completed")
			requireObjectCount(t, candidate, 1)
			httptestx.RequireErrorEnvelope(t, put(t, candidate, &issuingLogin, "", true), http.StatusNotFound, "object_upload_not_found_or_revoked")
			requireLeaseState(t, candidate.blobID, "completed")
			requireObjectCount(t, candidate, 1)
		})
	}
}

func TestAttachRouteContract_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	testDB := runtime.PrepareServerDatabase(t, "evidence_lifecycle-attach-route-contract")
	harness := runtime.StartServerWithDatabase(t, "evidence_lifecycle-attach-route-contract", testDB)
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	objectStoreAdmin := authflowtest.SeedLocalUserRecord(t, harness.DB, "evidence_lifecycle-object-store-admin@example.test", "EvidenceLifecycle Object Store Admin", "EvidenceLifecycleObjectStoreAdmin1!", false, true, true)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-attach-route-incident",
		"incident_key":  "evidence_lifecycle-attach-route",
		"title":         "Evidence attach route contract",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	otherIncident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-attach-route-other",
		"incident_key":  "evidence_lifecycle-attach-route-other",
		"title":         "Evidence attach route other",
	})
	otherIncidentID := appsupport.MustUUID(t, otherIncident["incident_id"].(string))

	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
	availableBlobID := insertRouteBlob(t, harness, incidentID, adminID, "available")
	attachURL := harness.Server.HTTP.URL + "/api/v1/evidence-records/" + recordID.String() + "/attach-blob"

	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "array", body: `[]`},
		{name: "missing object_blob_id", body: `{"base_row_version":1,"client_txn_id":"txn-route-shape"}`},
		{name: "missing base_row_version", body: `{"object_blob_id":"` + availableBlobID.String() + `","client_txn_id":"txn-route-shape"}`},
		{name: "missing client_txn_id", body: `{"object_blob_id":"` + availableBlobID.String() + `","base_row_version":1}`},
		{name: "null client_txn_id", body: `{"object_blob_id":"` + availableBlobID.String() + `","base_row_version":1,"client_txn_id":null}`},
		{name: "unknown member", body: `{"object_blob_id":"` + availableBlobID.String() + `","base_row_version":1,"client_txn_id":"txn-route-shape","extra":true}`},
	} {
		t.Run("shape "+tc.name, func(t *testing.T) {
			httptestx.RequireErrorEnvelope(t, doRawJSON(t, http.MethodPost, attachURL, tc.body, authOptions(login)...), http.StatusBadRequest, "invalid_mutation_payload")
		})
	}

	t.Run("viewer cannot attach", func(t *testing.T) {
		viewer := authflowtest.SeedLocalUserRecord(t, harness.DB, "evidence_lifecycle-attach-viewer@example.test", "EvidenceLifecycle Attach Viewer", "EvidenceLifecycleAttachViewer1!", false, false, true)
		incidentstoretest.SeedMembership(t, harness.DB, incidentID, viewer.ID, "EvidenceLifecycle Attach Viewer", "viewer", adminID)
		viewerLogin := loginLocalUserNoMFA(t, harness, "evidence_lifecycle-attach-viewer@example.test", "EvidenceLifecycleAttachViewer1!")
		resp := appsupport.DoJSON(t, http.MethodPost, attachURL, map[string]any{
			"object_blob_id":   availableBlobID.String(),
			"base_row_version": 1,
			"client_txn_id":    "txn-evidence_lifecycle-viewer-denied",
		}, authOptions(viewerLogin)...)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusForbidden, "authorization_denied")
		requireEvidenceStates(t, harness, recordID, "received", "pending")
	})

	t.Run("route-owned failures use canonical envelopes", func(t *testing.T) {
		cases := []struct {
			name       string
			blobID     uuid.UUID
			base       int
			wantStatus int
			wantCode   string
			wantReason string
		}{
			{name: "row version conflict", blobID: availableBlobID, base: 9, wantStatus: http.StatusConflict, wantCode: "row_version_conflict"},
			{name: "missing blob", blobID: uuid.New(), base: 1, wantStatus: http.StatusConflict, wantCode: "evidence_attach_rejected", wantReason: evidence.AttachReasonBlobNotVisible},
			{name: "foreign blob", blobID: insertRouteBlob(t, harness, otherIncidentID, adminID, "available"), base: 1, wantStatus: http.StatusConflict, wantCode: "evidence_attach_rejected", wantReason: evidence.AttachReasonBlobNotVisible},
			{name: "pending blob", blobID: insertRouteBlob(t, harness, incidentID, adminID, "pending"), base: 1, wantStatus: http.StatusConflict, wantCode: "evidence_attach_rejected", wantReason: evidence.AttachReasonBlobPending},
			{name: "failed blob", blobID: insertRouteBlob(t, harness, incidentID, adminID, "failed"), base: 1, wantStatus: http.StatusConflict, wantCode: "evidence_attach_rejected", wantReason: evidence.AttachReasonBlobFailed},
			{name: "quarantined blob", blobID: insertRouteBlob(t, harness, incidentID, adminID, "quarantined"), base: 1, wantStatus: http.StatusConflict, wantCode: "evidence_attach_rejected", wantReason: evidence.AttachReasonBlobQuarantined},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				resp := appsupport.DoJSON(t, http.MethodPost, attachURL, map[string]any{
					"object_blob_id":   tc.blobID.String(),
					"base_row_version": tc.base,
					"client_txn_id":    "txn-evidence_lifecycle-route-" + strings.ReplaceAll(tc.name, " ", "-"),
				}, authOptions(login)...)
				body := httptestx.RequireErrorEnvelope(t, resp, tc.wantStatus, tc.wantCode)
				if tc.wantReason != "" {
					details := body["error"].(map[string]any)["details"].(map[string]any)
					if got := details["reason_code"]; got != tc.wantReason {
						t.Fatalf("reason_code got %v want %s", got, tc.wantReason)
					}
				}
				requireEvidenceStates(t, harness, recordID, "received", "pending")
			})
		}
	})

	t.Run("divergent replay is rejected through HTTP", func(t *testing.T) {
		replayRecordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, replayRecordID)
		firstBlobID := insertRouteBlob(t, harness, incidentID, adminID, "available")
		replayURL := harness.Server.HTTP.URL + "/api/v1/evidence-records/" + replayRecordID.String() + "/attach-blob"
		body := map[string]any{"object_blob_id": firstBlobID.String(), "base_row_version": 1, "client_txn_id": "txn-evidence_lifecycle-route-divergent"}
		first := httptestx.RequireSuccessEnvelope(t, appsupport.DoJSON(t, http.MethodPost, replayURL, body, authOptions(login)...), http.StatusOK)["data"].(map[string]any)
		beforeRevisions := countEvidenceRevisions(t, harness, replayRecordID)
		secondBlobID := insertRouteBlob(t, harness, incidentID, adminID, "available")
		resp := appsupport.DoJSON(t, http.MethodPost, replayURL, map[string]any{
			"object_blob_id":   secondBlobID.String(),
			"base_row_version": 2,
			"client_txn_id":    "txn-evidence_lifecycle-route-divergent",
		}, authOptions(login)...)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "client_txn_conflict")
		if got := countEvidenceRevisions(t, harness, replayRecordID); got != beforeRevisions {
			t.Fatalf("divergent replay changed revision count: got %d want %d", got, beforeRevisions)
		}
		if got := countEvidenceBlobLinks(t, harness, replayRecordID); got != 1 {
			t.Fatalf("divergent replay changed blob link count: got %d want 1", got)
		}
		if first["object_blob_id"] != firstBlobID.String() {
			t.Fatalf("first attach object_blob_id got %#v want %s", first["object_blob_id"], firstBlobID)
		}
	})
	// The application-process lease intentionally forbids overlapping runtimes on
	// one database. The dependency-error cases below replace the server with
	// different object-store adapters, so release the contract fixture server first.
	harness.Server.Close()

	t.Run("object-store dependency errors use owner public mapping", func(t *testing.T) {
		requireObjectStoreDependencyErrorsUseOwnerPublicMapping(
			t,
			func(t testing.TB, prefix string, store objectstore.Store) *appsupport.ServerHarness {
				return runtime.StartServerWithDatabaseAndObjectStore(t, prefix, testDB, store)
			},
			ObjectStoreDependencyAdmin{
				email:    objectStoreAdmin.Email,
				password: "EvidenceLifecycleObjectStoreAdmin1!",
				userID:   objectStoreAdmin.ID,
			},
		)
	})
}
