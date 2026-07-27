package evidence_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestObjectBlobCreate_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence-blob-routes")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence-incident",
		"incident_key":  "evidence-routes",
		"title":         "Evidence routes",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)

	beforeInvalidShape := countObjectBlobs(t, harness, incidentID)
	for name, body := range map[string]any{
		"unknown field": map[string]any{
			"incident_id":   incidentID.String(),
			"client_txn_id": "txn-blob-unknown-field",
			"byte_size":     1,
			"unexpected":    true,
		},
		"missing incident": map[string]any{
			"client_txn_id": "txn-blob-missing-incident",
			"byte_size":     1,
		},
		"invalid digest": map[string]any{
			"incident_id":   incidentID.String(),
			"client_txn_id": "txn-blob-invalid-digest",
			"byte_size":     1,
			"sha256_hex":    "not-a-digest",
		},
	} {
		t.Run("request shape "+name, func(t *testing.T) {
			resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", body, authOptions(login)...)
			httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_blob_create_request")
		})
	}
	if got := countObjectBlobs(t, harness, incidentID); got != beforeInvalidShape {
		t.Fatalf("invalid blob create request shape wrote object_blobs: got %d want %d", got, beforeInvalidShape)
	}

	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'viewer',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminID, adminID); err != nil {
		t.Fatalf("demote actor before blob auth re-derivation: %v", err)
	}
	deniedCreate := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":   incidentID.String(),
		"client_txn_id": "txn-blob-viewer-denied",
		"byte_size":     1,
	}, authOptions(login)...)
	httptestx.RequireErrorEnvelope(t, deniedCreate, http.StatusForbidden, "authorization_denied")
	if got := countObjectBlobs(t, harness, incidentID); got != beforeInvalidShape {
		t.Fatalf("unauthorized blob create wrote object_blobs: got %d want %d", got, beforeInvalidShape)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'admin',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminID, adminID); err != nil {
		t.Fatalf("restore actor membership before blob create: %v", err)
	}

	beforeRejectedBlob := countObjectBlobs(t, harness, incidentID)
	rejectedBlob := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":   incidentID.String(),
		"client_txn_id": "txn-blob-too-large",
		"byte_size":     int64(9223372036854775807),
	}, authOptions(login)...)
	rejectedBody := httptestx.RequireErrorEnvelope(t, rejectedBlob, http.StatusRequestEntityTooLarge, "blob_create_rejected")
	if details := rejectedBody["error"].(map[string]any)["details"].(map[string]any); details["reason_code"] != "byte_size_exceeds_limit" {
		t.Fatalf("expected byte_size_exceeds_limit, got %#v", details)
	}
	if got := countObjectBlobs(t, harness, incidentID); got != beforeRejectedBlob {
		t.Fatalf("oversize blob create wrote object_blobs: got %d want %d", got, beforeRejectedBlob)
	}

	payload := []byte("hello evidence")
	sum := sha256.Sum256(payload)
	createBody := map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-blob-create",
		"byte_size":         len(payload),
		"filename_hint":     " hello.txt ",
		"content_type_hint": "text/plain",
		"sha256_hex":        fmt.Sprintf("%x", sum[:]),
	}
	createResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", createBody, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	if got := createData["upload_state"]; got != "pending" {
		t.Fatalf("unexpected create upload_state: %#v", got)
	}
	accepted := createData["accepted_contract"].(map[string]any)
	if accepted["filename_hint"] != "hello.txt" {
		t.Fatalf("expected normalized filename_hint, got %#v", accepted["filename_hint"])
	}
	uploadTarget := createData["upload_target"].(map[string]any)
	putObject(t, harness.Server.HTTP.URL, uploadTarget["href"].(string), payload, "text/plain")

	replayResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", createBody, authOptions(login)...)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	if replayData["object_blob_id"] != createData["object_blob_id"] {
		t.Fatalf("blob replay should return original object_blob_id")
	}
	divergent := map[string]any{
		"incident_id":   incidentID.String(),
		"client_txn_id": "txn-blob-create",
		"byte_size":     len(payload) + 1,
	}
	httptestx.RequireErrorEnvelope(t, appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", divergent, authOptions(login)...), http.StatusConflict, "client_txn_conflict")

	attachBody := map[string]any{
		"object_blob_id":   createData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    "txn-attach-blob",
	}
	attachResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/attach-blob", attachBody, authOptions(login)...)
	if attachResp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(attachResp.Body)
		t.Fatalf("attach status %d: %s", attachResp.StatusCode, string(data))
	}
	attachData := httptestx.RequireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
	if attachData["object_blob_id"] != createData["object_blob_id"] {
		t.Fatalf("attach response object_blob_id mismatch: %#v", attachData)
	}
	row := attachData["row"].(map[string]any)
	cells := row["cells"].(map[string]any)
	if got := cells["evidence.upload_state"].(map[string]any)["value"]; got != "available" {
		t.Fatalf("expected attached evidence upload_state available, got %#v", got)
	}
	attachReplay := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/attach-blob", attachBody, authOptions(login)...)
	httptestx.RequireSuccessEnvelope(t, attachReplay, http.StatusOK)

	previewResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...)
	previewData := httptestx.RequireSuccessEnvelope(t, previewResp, http.StatusOK)["data"].(map[string]any)
	if previewData["handle_kind"] != "preview" || previewData["preview_kind"] != "text_inline" {
		t.Fatalf("unexpected preview handle payload: %#v", previewData)
	}
	previewBody := redeemHandle(t, harness.Server.HTTP.URL+previewData["href"].(string), login)
	if string(previewBody) != string(payload) {
		t.Fatalf("preview body mismatch: got %q", string(previewBody))
	}

	downloadResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/download-handle", map[string]any{}, authOptions(login)...)
	downloadData := httptestx.RequireSuccessEnvelope(t, downloadResp, http.StatusOK)["data"].(map[string]any)
	downloadURL := harness.Server.HTTP.URL + downloadData["href"].(string)
	downloadBody := redeemHandle(t, downloadURL, login)
	if string(downloadBody) != string(payload) {
		t.Fatalf("download body mismatch: got %q", string(downloadBody))
	}
	second := appsupport.DoJSON(t, http.MethodGet, downloadURL, nil, appsupport.WithCookies(login.SessionCookie))
	httptestx.RequireErrorEnvelope(t, second, http.StatusGone, "handle_consumed")
}

func TestEvidenceHandles_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "entity_linking-evidence-handles")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-entity_linking-handles-incident",
		"incident_key":  "entity_linking-handles",
		"title":         "Record relationships evidence handles",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
	payload := []byte("handle body")
	attachUploadedBlob(t, harness, login, incidentID, recordID, payload, "txn-entity_linking-handles-blob", "txn-entity_linking-handles-attach")

	beforeInvalidHandles := countAccessHandles(t, harness, incidentID)
	for _, endpoint := range []struct {
		name string
		path string
	}{
		{name: "preview", path: "preview-handle"},
		{name: "download", path: "download-handle"},
	} {
		for name, body := range map[string]string{
			"zero length": "",
			"null":        "null",
			"array":       "[]",
			"unknown":     `{"unexpected":true}`,
			"client txn":  `{"client_txn_id":"forbidden"}`,
		} {
			t.Run(endpoint.name+" "+name, func(t *testing.T) {
				resp := doRawJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint.path, body, authOptions(login)...)
				httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_evidence_handle_request")
			})
		}
	}
	if got := countAccessHandles(t, harness, incidentID); got != beforeInvalidHandles {
		t.Fatalf("invalid handle issuance requests wrote evidence_access_handles: got %d want %d", got, beforeInvalidHandles)
	}

	previewResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...)
	previewData := httptestx.RequireSuccessEnvelope(t, previewResp, http.StatusOK)["data"].(map[string]any)
	if previewData["handle_kind"] != "preview" || previewData["single_use"] != false || previewData["disposition"] != "inline" || previewData["preview_kind"] != "text_inline" {
		t.Fatalf("unexpected preview handle payload: %#v", previewData)
	}
	downloadResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/download-handle", map[string]any{}, authOptions(login)...)
	downloadData := httptestx.RequireSuccessEnvelope(t, downloadResp, http.StatusOK)["data"].(map[string]any)
	if downloadData["handle_kind"] != "download" || downloadData["single_use"] != true || downloadData["disposition"] != "attachment" {
		t.Fatalf("unexpected download handle payload: %#v", downloadData)
	}
	if _, ok := downloadData["preview_kind"]; ok {
		t.Fatalf("download handle must omit preview_kind, got %#v", downloadData)
	}
	afterIssuedHandles := countAccessHandles(t, harness, incidentID)
	if afterIssuedHandles != beforeInvalidHandles+2 {
		t.Fatalf("expected exactly two successful handle rows, got %d want %d", afterIssuedHandles, beforeInvalidHandles+2)
	}

	if _, err := harness.DB.ExecContext(context.Background(), `
DELETE FROM incident_memberships
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminID); err != nil {
		t.Fatalf("remove incident membership before handle auth re-derivation: %v", err)
	}
	for _, endpoint := range []string{"preview-handle", "download-handle"} {
		resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusNotFound, "evidence_record_not_found")
	}
	if got := countAccessHandles(t, harness, incidentID); got != afterIssuedHandles {
		t.Fatalf("authorization failures wrote evidence_access_handles: got %d want %d", got, afterIssuedHandles)
	}
}

func TestEvidenceHandleIssuanceReportsRegisteredUnavailableReasons(t *testing.T) {
	harness := appsupport.StartServer(t, "entity_linking-evidence-handle-issue-reasons")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-entity_linking-handle-issue-reasons-incident",
		"incident_key":  "entity_linking-handle-issue-reasons",
		"title":         "Record relationships handle issue reasons",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	scenarios := []struct {
		name       string
		reasonCode string
		arrange    func(t *testing.T, recordID uuid.UUID)
	}{
		{
			name:       "no linked blob",
			reasonCode: "no_visible_blob",
			arrange:    func(t *testing.T, recordID uuid.UUID) {},
		},
		{
			name:       "pending blob",
			reasonCode: "blob_pending",
			arrange: func(t *testing.T, recordID uuid.UUID) {
				linkSeededBlob(t, harness, incidentID, adminID, recordID, "pending", "available", "issue/pending")
			},
		},
		{
			name:       "failed blob",
			reasonCode: "blob_failed",
			arrange: func(t *testing.T, recordID uuid.UUID) {
				linkSeededBlob(t, harness, incidentID, adminID, recordID, "failed", "available", "issue/failed")
			},
		},
		{
			name:       "missing backing object",
			reasonCode: "blob_missing",
			arrange: func(t *testing.T, recordID uuid.UUID) {
				linkSeededBlobWithCanonicalStorageKey(t, harness, incidentID, adminID, recordID, "available", "available")
			},
		},
		{
			name:       "quarantined blob",
			reasonCode: "evidence_quarantined",
			arrange: func(t *testing.T, recordID uuid.UUID) {
				linkSeededBlob(t, harness, incidentID, adminID, recordID, "quarantined", "available", "issue/quarantined")
			},
		},
		{
			name:       "inconsistent lifecycle",
			reasonCode: "evidence_inconsistent",
			arrange: func(t *testing.T, recordID uuid.UUID) {
				linkSeededBlob(t, harness, incidentID, adminID, recordID, "available", "received", "issue/inconsistent")
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			recordID := uuid.New()
			seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
			scenario.arrange(t, recordID)

			beforeHandles := countAccessHandles(t, harness, incidentID)
			for _, endpoint := range []string{"preview-handle", "download-handle"} {
				resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...)
				requireEvidenceAccessUnavailableReason(t, resp, scenario.reasonCode)
			}
			if got := countAccessHandles(t, harness, incidentID); got != beforeHandles {
				t.Fatalf("blocked issuance wrote evidence_access_handles: got %d want %d", got, beforeHandles)
			}
		})
	}
}

func TestEvidenceHandleRedemptionReportsRegisteredUnavailableReasons(t *testing.T) {
	harness := appsupport.StartServer(t, "entity_linking-evidence-handle-redeem-reasons")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-entity_linking-handle-redeem-reasons-incident",
		"incident_key":  "entity_linking-handle-redeem-reasons",
		"title":         "Record relationships handle redeem reasons",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	scenarios := []struct {
		name       string
		reasonCode string
		mutate     func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID)
	}{
		{
			name:       "detached blob",
			reasonCode: "no_visible_blob",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				updateEvidenceBlobLink(t, harness, recordID, nil)
			},
		},
		{
			name:       "pending blob",
			reasonCode: "blob_pending",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				updateBlobState(t, harness, objectBlobID, "pending")
			},
		},
		{
			name:       "failed blob",
			reasonCode: "blob_failed",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				updateBlobState(t, harness, objectBlobID, "failed")
			},
		},
		{
			name:       "missing backing object",
			reasonCode: "blob_missing",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				deleteBlobObject(t, harness, blobStorageKey(t, harness, objectBlobID))
			},
		},
		{
			name:       "quarantined blob",
			reasonCode: "evidence_quarantined",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				updateBlobState(t, harness, objectBlobID, "quarantined")
			},
		},
		{
			name:       "inconsistent lifecycle",
			reasonCode: "evidence_inconsistent",
			mutate: func(t *testing.T, recordID uuid.UUID, objectBlobID uuid.UUID) {
				updateEvidenceLifecycle(t, harness, recordID, "received")
			},
		},
	}

	for _, scenario := range scenarios {
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			t.Run(endpoint+" "+scenario.name, func(t *testing.T) {
				recordID := uuid.New()
				seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
				attachData := attachUploadedBlob(t, harness, login, incidentID, recordID, []byte("redeem body"), "txn-"+recordID.String()+"-blob", "txn-"+recordID.String()+"-attach")
				objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))

				issueResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...)
				issueData := httptestx.RequireSuccessEnvelope(t, issueResp, http.StatusOK)["data"].(map[string]any)
				scenario.mutate(t, recordID, objectBlobID)

				redeemResp := appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+issueData["href"].(string), nil, appsupport.WithCookies(login.SessionCookie))
				requireEvidenceAccessUnavailableReason(t, redeemResp, scenario.reasonCode)
			})
		}
	}
}

func TestDownloadHandleBlobMissingDoesNotConsumeHandle_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "entity_linking-download-handle-no-byte-failure")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-entity_linking-download-no-byte-incident",
		"incident_key":  "entity_linking-download-no-byte",
		"title":         "Record relationships download no-byte failure",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	recordID := uuid.New()
	seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
	payload := []byte("download retry body")
	attachData := attachUploadedBlob(t, harness, login, incidentID, recordID, payload, "txn-entity_linking-download-no-byte-blob", "txn-entity_linking-download-no-byte-attach")
	objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))
	originalStorageKey := blobStorageKey(t, harness, objectBlobID)

	downloadResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/download-handle", map[string]any{}, authOptions(login)...)
	downloadData := httptestx.RequireSuccessEnvelope(t, downloadResp, http.StatusOK)["data"].(map[string]any)
	downloadURL := harness.Server.HTTP.URL + downloadData["href"].(string)

	deleteBlobObject(t, harness, originalStorageKey)
	missingResp := appsupport.DoJSON(t, http.MethodGet, downloadURL, nil, appsupport.WithCookies(login.SessionCookie))
	requireEvidenceAccessUnavailableReason(t, missingResp, "blob_missing")

	if err := harness.Server.Runtime.ObjectStore.PutObject(context.Background(), originalStorageKey, bytes.NewReader(payload), int64(len(payload)), "text/plain"); err != nil {
		t.Fatalf("restore blob object: %v", err)
	}
	downloadBody := redeemHandle(t, downloadURL, login)
	if string(downloadBody) != string(payload) {
		t.Fatalf("download body mismatch after no-byte failure: got %q", string(downloadBody))
	}
	second := appsupport.DoJSON(t, http.MethodGet, downloadURL, nil, appsupport.WithCookies(login.SessionCookie))
	httptestx.RequireErrorEnvelope(t, second, http.StatusGone, "handle_consumed")
}

func TestEvidenceHandleRequestRejectsUnknownMembers(t *testing.T) {
	body := strings.NewReader(`{"client_txn_id":"forbidden"}`)
	if apiErr := evidenceDecodeHandleForTest(body); apiErr == nil || apiErr.Code != "invalid_evidence_handle_request" {
		t.Fatalf("expected invalid_evidence_handle_request for handle issuance client_txn_id, got %#v", apiErr)
	}
}

func TestEvidenceOpenAPIIncludesRouteFamily(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(), "contracts", "openapi", "cartulary.openapi.yaml"))
	if err != nil {
		t.Fatalf("read openapi contract: %v", err)
	}
	var document struct {
		Paths map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode openapi contract: %v", err)
	}
	for _, path := range []string{
		"/api/v1/object-blobs",
		"/api/v1/evidence-records/{record_id}/attach-blob",
		"/api/v1/evidence-records/{record_id}/preview-handle",
		"/api/v1/evidence-records/{record_id}/download-handle",
		"/api/v1/evidence-handles/{handle_token}",
	} {
		if _, ok := document.Paths[path]; !ok {
			t.Fatalf("missing evidence OpenAPI path %s", path)
		}
	}
}

func authOptions(login appsupport.LoginResult) []func(*http.Request) {
	return []func(*http.Request){
		appsupport.WithCookies(login.SessionCookie, login.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	}
}

func attachUploadedBlob(t *testing.T, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, payload []byte, createTxn string, attachTxn string) map[string]any {
	t.Helper()
	sum := sha256.Sum256(payload)
	createBody := map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     createTxn,
		"byte_size":         len(payload),
		"filename_hint":     " handle.txt ",
		"content_type_hint": "text/plain",
		"sha256_hex":        fmt.Sprintf("%x", sum[:]),
	}
	createResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", createBody, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	uploadTarget := createData["upload_target"].(map[string]any)
	putObject(t, harness.Server.HTTP.URL, uploadTarget["href"].(string), payload, "text/plain")
	attachResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/attach-blob", map[string]any{
		"object_blob_id":   createData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    attachTxn,
	}, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
}

func countObjectBlobs(t *testing.T, harness *appsupport.ServerHarness, incidentID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM object_blobs WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		t.Fatalf("count object_blobs: %v", err)
	}
	return count
}

func countAccessHandles(t *testing.T, harness *appsupport.ServerHarness, incidentID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM evidence_access_handles WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		t.Fatalf("count evidence access handles: %v", err)
	}
	return count
}

func requireEvidenceAccessUnavailableReason(t *testing.T, resp *http.Response, wantReasonCode string) {
	t.Helper()
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "evidence_access_unavailable")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if got := details["reason_code"]; got != wantReasonCode {
		t.Fatalf("unexpected evidence_access_unavailable reason_code: got %v want %s", got, wantReasonCode)
	}
}

func linkSeededBlob(t *testing.T, harness *appsupport.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, uploadState string, evidenceLifecycle string, storageKey string) uuid.UUID {
	t.Helper()
	objectBlobID := uuid.New()
	return linkSeededBlobWithIDAndStorageKey(t, harness, incidentID, actorID, recordID, uploadState, evidenceLifecycle, objectBlobID, storageKey)
}

func linkSeededBlobWithCanonicalStorageKey(t *testing.T, harness *appsupport.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, uploadState string, evidenceLifecycle string) uuid.UUID {
	t.Helper()
	objectBlobID := uuid.New()
	storageKey, err := blobref.ObjectBlobStorageKey(incidentID, objectBlobID)
	if err != nil {
		t.Fatalf("canonical storage key: %v", err)
	}
	return linkSeededBlobWithIDAndStorageKey(t, harness, incidentID, actorID, recordID, uploadState, evidenceLifecycle, objectBlobID, storageKey)
}

func linkSeededBlobWithIDAndStorageKey(t *testing.T, harness *appsupport.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, uploadState string, evidenceLifecycle string, objectBlobID uuid.UUID, storageKey string) uuid.UUID {
	t.Helper()
	var terminalReason any
	var failedMarker any
	if uploadState == "failed" {
		terminalReason = "pending_timeout"
		failedMarker = "failed"
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint, observed_size, observed_content_type,
    observed_sha256_hex, target_expires_at, pending_expires_at, finalized_at,
    terminal_reason, failed_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    11, 'seed.txt', 'text/plain', 11, 'text/plain',
    '0000000000000000000000000000000000000000000000000000000000000000',
    now() + interval '1 hour', now() + interval '24 hours',
    CASE WHEN $5 IN ('available', 'quarantined') THEN now() ELSE NULL END,
    $6,
    CASE WHEN $7::text IS NULL THEN NULL ELSE now() END,
    now(), now()
)
`, objectBlobID, incidentID, actorID, storageKey, uploadState, terminalReason, failedMarker); err != nil {
		t.Fatalf("seed object blob: %v", err)
	}
	updateEvidenceBlobLink(t, harness, recordID, &objectBlobID)
	updateEvidenceLifecycle(t, harness, recordID, evidenceLifecycle)
	return objectBlobID
}

func deleteBlobObject(t *testing.T, harness *appsupport.ServerHarness, storageKey string) {
	t.Helper()
	if err := harness.Server.Runtime.ObjectStore.DeleteObject(context.Background(), storageKey); err != nil {
		t.Fatalf("delete blob object: %v", err)
	}
}

func updateEvidenceBlobLink(t *testing.T, harness *appsupport.ServerHarness, recordID uuid.UUID, objectBlobID *uuid.UUID) {
	t.Helper()
	var blobArg any
	if objectBlobID != nil {
		blobArg = *objectBlobID
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE evidence
   SET object_blob_id = $2,
       upload_state = CASE WHEN $2::uuid IS NULL THEN 'pending' ELSE upload_state END,
       updated_at = now()
 WHERE record_id = $1
`, recordID, blobArg); err != nil {
		t.Fatalf("update evidence object_blob_id: %v", err)
	}
}

func updateEvidenceLifecycle(t *testing.T, harness *appsupport.ServerHarness, recordID uuid.UUID, lifecycleState string) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE evidence
   SET lifecycle_state = $2,
       updated_at = now()
 WHERE record_id = $1
`, recordID, lifecycleState); err != nil {
		t.Fatalf("update evidence lifecycle_state: %v", err)
	}
}

func updateBlobState(t *testing.T, harness *appsupport.ServerHarness, objectBlobID uuid.UUID, uploadState string) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE object_blobs
   SET upload_state = $2,
       terminal_reason = CASE WHEN $2 = 'failed' THEN 'pending_timeout' ELSE NULL END,
       failed_at = CASE WHEN $2 = 'failed' THEN now() ELSE NULL END,
       updated_at = now()
 WHERE object_blob_id = $1
`, objectBlobID, uploadState); err != nil {
		t.Fatalf("update object blob upload_state: %v", err)
	}
}

func updateBlobStorageKey(t *testing.T, harness *appsupport.ServerHarness, objectBlobID uuid.UUID, storageKey string) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE object_blobs
   SET storage_key = $2,
       updated_at = now()
 WHERE object_blob_id = $1
`, objectBlobID, storageKey); err != nil {
		t.Fatalf("update object blob storage_key: %v", err)
	}
}

func blobStorageKey(t *testing.T, harness *appsupport.ServerHarness, objectBlobID uuid.UUID) string {
	t.Helper()
	var storageKey string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT storage_key
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&storageKey); err != nil {
		t.Fatalf("load object blob storage_key: %v", err)
	}
	return storageKey
}

func doRawJSON(t *testing.T, method string, url string, body string, options ...func(*http.Request)) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("create raw json request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func seedEvidenceRecord(t *testing.T, harness *appsupport.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id, row_version)
VALUES ($1, $2, 'evidence', $3, $3, 1)
`, recordID, incidentID, actorID); err != nil {
		t.Fatalf("seed evidence record envelope: %v", err)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state, requested_at)
VALUES ($1, $2, 'Evidence payload', 'received', 'pending', now())
`, recordID, incidentID); err != nil {
		t.Fatalf("seed evidence row: %v", err)
	}
}

func putObject(t *testing.T, baseURL string, href string, payload []byte, contentType string) {
	t.Helper()
	if strings.HasPrefix(href, "/") {
		href = baseURL + href
	}
	req, err := http.NewRequest(http.MethodPut, href, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create object upload request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload object: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload object status %d: %s", resp.StatusCode, string(data))
	}
}

func redeemHandle(t *testing.T, href string, login appsupport.LoginResult) []byte {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodGet, href, nil, appsupport.WithCookies(login.SessionCookie))
	httptestx.RequireStatus(t, resp, http.StatusOK)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read handle response: %v", err)
	}
	return data
}

func evidenceDecodeHandleForTest(reader io.Reader) *httpapi.APIError {
	return evidence.DecodeHandleIssueRequest(reader)
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}
