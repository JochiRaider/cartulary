package serverprocess

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestEvidenceUploadAttachProjection_Process(t *testing.T) {
	server, db := startServerProcessWithDB(t, "evidence_lifecycle-e-smoke-01")
	adminLogin, _ := provisionBootstrapAdmin(t, server)
	incident := createIncident(t, server, adminLogin.SessionCookie, adminLogin.CSRFCookie, map[string]any{
		"client_txn_id": "txn-e-5-smoke-incident",
		"incident_key":  "IR-E5SMOKE",
		"title":         "Evidence process smoke",
	})
	incidentID := incident["incident_id"].(string)

	timeline := createViewRow(t, server, adminLogin, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-e-5-smoke-timeline",
		"timeline.activity_synopsis_text": "Process evidence row",
	})
	timelineRow := timeline["row"].(map[string]any)
	timelineRecordID := timelineRow["record_id"].(string)
	timelineRowVersion := int(timelineRow["row_version"].(float64))

	evidence := createViewRow(t, server, adminLogin, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-e-5-smoke-evidence",
		"evidence.title": "Process evidence",
	})
	evidenceRecordID := evidence["row"].(map[string]any)["record_id"].(string)
	payload := []byte("evidence_lifecycle process evidence payload")
	blobCreate := doJSON(t, server, http.MethodPost, "/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID,
		"client_txn_id":     "txn-e-5-smoke-blob",
		"byte_size":         len(payload),
		"filename_hint":     "process.txt",
		"content_type_hint": "text/plain",
	}, withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	blobData := httptestx.RequireSuccessEnvelope(t, blobCreate, http.StatusCreated)["data"].(map[string]any)
	uploadTarget := blobData["upload_target"].(map[string]any)
	putObject(t, server.BaseURL, uploadTarget, payload, adminLogin)

	attach := doJSON(t, server, http.MethodPost, "/api/v1/evidence-records/"+evidenceRecordID+"/attach-blob", map[string]any{
		"object_blob_id":   blobData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    "txn-e-5-smoke-attach",
	}, withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	attachData := httptestx.RequireSuccessEnvelope(t, attach, http.StatusOK)["data"].(map[string]any)
	attachedRow := attachData["row"].(map[string]any)
	patchRecord(t, server, adminLogin, evidenceRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": attachedRow["row_version"],
		"client_txn_id":    "txn-e-5-smoke-available",
		"changes": []map[string]any{{
			"field_key": "evidence.lifecycle_state",
			"value":     "available",
		}},
	})

	patched := patchRecord(t, server, adminLogin, timelineRecordID, map[string]any{
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
	var activeLinks int
	if err := db.QueryRow(`
SELECT COUNT(*)
  FROM active_record_links_v1
 WHERE src_record_id::text = $1
   AND dst_record_id::text = $2
   AND link_type = 'attached_evidence'
`, timelineRecordID, evidenceRecordID).Scan(&activeLinks); err != nil {
		t.Fatalf("query active Timeline Evidence links: %v", err)
	}
	if activeLinks != 1 {
		t.Fatalf("active Timeline Evidence links got %d want 1", activeLinks)
	}
	patchedCells := patched["row"].(map[string]any)["cells"].(map[string]any)
	if got := int(patchedCells["timeline.evidence_count"].(map[string]any)["value"].(float64)); got != 1 {
		t.Fatalf("timeline patch response evidence_count got %d want 1", got)
	}
	requireTimelineEvidenceCount(t, server, adminLogin, incidentID, timelineRecordID, 1, true)

	requested := createViewRow(t, server, adminLogin, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-e-5-smoke-requested",
		"evidence.title": "Requested no blob",
	})
	requestedRecordID := requested["row"].(map[string]any)["record_id"].(string)
	noBlobPreview := doJSON(t, server, http.MethodPost, "/api/v1/evidence-records/"+requestedRecordID+"/preview-handle", map[string]any{}, withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	body := httptestx.RequireErrorEnvelope(t, noBlobPreview, http.StatusConflict, "evidence_access_unavailable")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "no_visible_blob" {
		t.Fatalf("preview no-blob reason got %v want no_visible_blob", details["reason_code"])
	}
}

func createViewRow(t testing.TB, server *processtest.Server, login flowtest.LoginResult, incidentID string, viewSchemaID string, body map[string]any) map[string]any {
	t.Helper()
	resp := doJSON(t, server, http.MethodPost, "/api/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/rows", body, withCookies(login.SessionCookie, login.CSRFCookie), withHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func patchRecord(t testing.TB, server *processtest.Server, login flowtest.LoginResult, recordID string, body map[string]any) map[string]any {
	t.Helper()
	resp := doJSON(t, server, http.MethodPatch, "/api/v1/records/"+recordID, body, withCookies(login.SessionCookie, login.CSRFCookie), withHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func requireTimelineEvidenceCount(t testing.TB, server *processtest.Server, login flowtest.LoginResult, incidentID string, recordID string, wantCount int, wantHasEvidence bool) map[string]any {
	t.Helper()
	resp := doJSON(t, server, http.MethodPost, "/api/v1/incidents/"+incidentID+"/views/cartulary.view.timeline.v2/query", map[string]any{}, withCookies(login.SessionCookie, login.CSRFCookie), withHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
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
		return row
	}
	t.Fatalf("timeline row %s not found in query result", recordID)
	return nil
}

func putObject(t testing.TB, baseURL string, target map[string]any, payload []byte, login flowtest.LoginResult) {
	t.Helper()
	href := target["href"].(string)
	if strings.HasPrefix(href, "/") {
		href = baseURL + href
	}
	req, err := http.NewRequest(target["method"].(string), href, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create upload request: %v", err)
	}
	for name, rawValue := range target["headers"].(map[string]any) {
		req.Header.Set(name, rawValue.(string))
	}
	req.Header.Set(authn.CSRFHeaderName, login.CSRFCookie.Value)
	req.AddCookie(login.SessionCookie)
	req.AddCookie(login.CSRFCookie)
	resp, err := newProcessHTTPClient().Do(req)
	if err != nil {
		t.Fatalf("upload object: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload object status %d: %s", resp.StatusCode, string(data))
	}
}
