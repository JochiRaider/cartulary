package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestPhase5_EvidenceUploadAttachProjection_E_5_SMOKE_01_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase5-e-smoke-01")
	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	incident := phase2CreateIncident(t, server, adminLogin.sessionCookie, adminLogin.csrfCookie, map[string]any{
		"client_txn_id": "txn-e-5-smoke-incident",
		"incident_key":  "IR-E5SMOKE",
		"title":         "Phase 5 process smoke",
	})
	incidentID := incident["incident_id"].(string)

	timeline := phase5CreateViewRow(t, server, adminLogin, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-e-5-smoke-timeline",
		"timeline.activity_synopsis_text": "Process evidence row",
	})
	timelineRow := timeline["row"].(map[string]any)
	timelineRecordID := timelineRow["record_id"].(string)
	timelineRowVersion := int(timelineRow["row_version"].(float64))

	evidence := phase5CreateViewRow(t, server, adminLogin, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-e-5-smoke-evidence",
		"evidence.title": "Process evidence",
	})
	evidenceRecordID := evidence["row"].(map[string]any)["record_id"].(string)
	payload := []byte("phase5 process evidence payload")
	blobCreate := phase1DoJSON(t, server, http.MethodPost, "/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID,
		"client_txn_id":     "txn-e-5-smoke-blob",
		"byte_size":         len(payload),
		"filename_hint":     "process.txt",
		"content_type_hint": "text/plain",
	}, withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie), withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value))
	blobData := httptestx.RequireSuccessEnvelope(t, blobCreate, http.StatusCreated)["data"].(map[string]any)
	uploadTarget := blobData["upload_target"].(map[string]any)
	phase5PutObject(t, server.BaseURL, uploadTarget["href"].(string), payload, "text/plain")

	attach := phase1DoJSON(t, server, http.MethodPost, "/api/v1/evidence-records/"+evidenceRecordID+"/attach-blob", map[string]any{
		"object_blob_id":   blobData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    "txn-e-5-smoke-attach",
	}, withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie), withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value))
	httptestx.RequireSuccessEnvelope(t, attach, http.StatusOK)

	phase5PatchRecord(t, server, adminLogin, timelineRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.timeline.v2",
		"base_row_version": timelineRowVersion,
		"client_txn_id":    "txn-e-5-smoke-link",
		"changes": []map[string]any{{
			"field_key": "timeline.attached_evidence_ids",
			"action_payload": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{{
					"op":               "add_record_ref",
					"linked_record_id": evidenceRecordID,
				}},
			},
		}},
	})
	phase5RequireTimelineEvidenceCount(t, server, adminLogin, incidentID, timelineRecordID, 1, true)

	requested := phase5CreateViewRow(t, server, adminLogin, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-e-5-smoke-requested",
		"evidence.title": "Requested no blob",
	})
	requestedRecordID := requested["row"].(map[string]any)["record_id"].(string)
	noBlobPreview := phase1DoJSON(t, server, http.MethodPost, "/api/v1/evidence-records/"+requestedRecordID+"/preview-handle", map[string]any{}, withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie), withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value))
	body := httptestx.RequireErrorEnvelope(t, noBlobPreview, http.StatusConflict, "evidence_access_unavailable")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "no_visible_blob" {
		t.Fatalf("preview no-blob reason got %v want no_visible_blob", details["reason_code"])
	}
}

func phase5CreateViewRow(t testing.TB, server *processtest.Server, login loginResult, incidentID string, viewSchemaID string, body map[string]any) map[string]any {
	t.Helper()
	resp := phase1DoJSON(t, server, http.MethodPost, "/api/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/rows", body, withCookies(login.sessionCookie, login.csrfCookie), withHeader(authn.CSRFHeaderName, login.csrfCookie.Value))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func phase5PatchRecord(t testing.TB, server *processtest.Server, login loginResult, recordID string, body map[string]any) map[string]any {
	t.Helper()
	resp := phase1DoJSON(t, server, http.MethodPatch, "/api/v1/records/"+recordID, body, withCookies(login.sessionCookie, login.csrfCookie), withHeader(authn.CSRFHeaderName, login.csrfCookie.Value))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func phase5RequireTimelineEvidenceCount(t testing.TB, server *processtest.Server, login loginResult, incidentID string, recordID string, wantCount int, wantHasEvidence bool) {
	t.Helper()
	resp := phase1DoJSON(t, server, http.MethodPost, "/api/v1/incidents/"+incidentID+"/views/cartulary.view.timeline.v2/query", map[string]any{}, withCookies(login.sessionCookie, login.csrfCookie), withHeader(authn.CSRFHeaderName, login.csrfCookie.Value))
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	for _, raw := range data["rows"].([]any) {
		row := raw.(map[string]any)
		if row["record_id"] != recordID {
			continue
		}
		cells := row["cells"].(map[string]any)
		gotCount := int(cells["timeline.evidence_count"].(map[string]any)["value"].(float64))
		gotHasEvidence := cells["timeline.has_evidence"].(map[string]any)["value"].(bool)
		if gotCount != wantCount || gotHasEvidence != wantHasEvidence {
			t.Fatalf("timeline evidence projection got count=%d has=%v want count=%d has=%v", gotCount, gotHasEvidence, wantCount, wantHasEvidence)
		}
		return
	}
	t.Fatalf("timeline row %s not found in query result", recordID)
}

func phase5PutObject(t testing.TB, baseURL string, href string, payload []byte, contentType string) {
	t.Helper()
	if strings.HasPrefix(href, "/") {
		href = baseURL + href
	}
	req, err := http.NewRequest(http.MethodPut, href, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create upload request: %v", err)
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
