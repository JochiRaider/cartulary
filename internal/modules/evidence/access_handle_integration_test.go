package evidence_test

// Evidence access-handle route contracts.
import (
	"bytes"
	"context"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"net/http"
	"testing"
)

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

	if err := harness.ObjectStore.PutObject(context.Background(), originalStorageKey, bytes.NewReader(payload), int64(len(payload)), "text/plain"); err != nil {
		t.Fatalf("restore blob object: %v", err)
	}
	downloadBody := redeemHandle(t, downloadURL, login)
	if string(downloadBody) != string(payload) {
		t.Fatalf("download body mismatch after no-byte failure: got %q", string(downloadBody))
	}
	second := appsupport.DoJSON(t, http.MethodGet, downloadURL, nil, appsupport.WithCookies(login.SessionCookie))
	httptestx.RequireErrorEnvelope(t, second, http.StatusGone, "handle_consumed")
}
