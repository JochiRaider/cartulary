package evidence_test

// Evidence upload route contracts.
import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"io"
	"net/http"
	"testing"
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
	putObject(t, harness.Server.HTTP.URL, uploadTarget["href"].(string), payload, "text/plain", login)

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
	if got := cells["evidence.lifecycle_state"].(map[string]any)["value"]; got != "received" {
		t.Fatalf("attachment must preserve evidence lifecycle_state: got %#v want received", got)
	}
	attachReplay := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/attach-blob", attachBody, authOptions(login)...)
	httptestx.RequireSuccessEnvelope(t, attachReplay, http.StatusOK)
	requireHTTPWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": int(row["row_version"].(float64)),
		"client_txn_id":    "txn-attach-blob-available",
		"changes": []map[string]any{{
			"field_key": "evidence.lifecycle_state",
			"value":     "available",
		}},
	})

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
