package evidence_test

// Shared Evidence route fixtures and OpenAPI contract.
import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

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
	putObject(t, harness.Server.HTTP.URL, uploadTarget["href"].(string), payload, "text/plain", login)
	attachResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/attach-blob", map[string]any{
		"object_blob_id":   createData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    attachTxn,
	}, authOptions(login)...)
	attachData := httptestx.RequireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
	row := attachData["row"].(map[string]any)
	available := requireHTTPWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": int(row["row_version"].(float64)),
		"client_txn_id":    attachTxn + "-available",
		"changes": []map[string]any{{
			"field_key": "evidence.lifecycle_state",
			"value":     "available",
		}},
	})
	attachData["row"] = available["row"]
	return attachData
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
	if err := harness.ObjectStore.DeleteObject(context.Background(), storageKey); err != nil {
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

func putObject(t *testing.T, baseURL string, href string, payload []byte, contentType string, login appsupport.LoginResult) {
	t.Helper()
	if strings.HasPrefix(href, "/") {
		href = baseURL + href
	}
	req, err := http.NewRequest(http.MethodPut, href, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create object upload request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)
	for _, option := range authOptions(login) {
		option(req)
	}
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

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}
