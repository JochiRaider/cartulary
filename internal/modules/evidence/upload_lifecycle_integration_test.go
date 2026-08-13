package evidence_test

// Evidence upload lifecycle and atomic-create contracts.
import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestObjectUploadAtomicEvidenceCreate_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence-atomic-initial-blob-create")
	login, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence-atomic-create-incident",
		"incident_key":  "evidence-atomic-create",
		"title":         "Evidence atomic initial blob create",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	payload := []byte("atomic evidence object")
	sum := sha256.Sum256(payload)

	createSlot := func(clientTxnID string) map[string]any {
		response := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
			"incident_id":       incidentID.String(),
			"client_txn_id":     clientTxnID,
			"byte_size":         len(payload),
			"filename_hint":     "atomic.txt",
			"content_type_hint": "text/plain",
			"sha256_hex":        fmt.Sprintf("%x", sum[:]),
		}, authOptions(login)...)
		return httptestx.RequireSuccessEnvelope(t, response, http.StatusCreated)["data"].(map[string]any)
	}

	pending := createSlot("txn-evidence-atomic-create-pending-slot")
	pendingCreate := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.evidence.v1/rows", map[string]any{
		"client_txn_id":                   "txn-evidence-atomic-create-pending-row",
		"evidence.initial_object_blob_id": pending["object_blob_id"],
	}, authOptions(login)...)
	pendingError := httptestx.RequireErrorEnvelope(t, pendingCreate, http.StatusConflict, "evidence_attach_rejected")
	httptestx.RequireErrorDetail(t, pendingError, "reason_code", evidence.AttachReasonBlobPending)

	slot := createSlot("txn-evidence-atomic-create-slot")
	target := slot["upload_target"].(map[string]any)
	putObject(t, harness.Server.HTTP.URL, target["href"].(string), payload, "text/plain", login)
	createBody := map[string]any{
		"client_txn_id":                   "txn-evidence-atomic-create-row",
		"evidence.initial_object_blob_id": slot["object_blob_id"],
	}
	createResponse := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.evidence.v1/rows", createBody, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResponse, http.StatusCreated)["data"].(map[string]any)
	if _, exposed := createData["object_blob_id"]; exposed {
		t.Fatalf("generic create response exposed sibling object_blob_id: %#v", createData)
	}
	row := createData["row"].(map[string]any)
	recordID := appsupport.MustUUID(t, row["record_id"].(string))
	cells := row["cells"].(map[string]any)
	if lifecycle := cells["evidence.lifecycle_state"].(map[string]any)["value"]; lifecycle != "requested" {
		t.Fatalf("blob-backed create lifecycle = %#v, want requested", lifecycle)
	}
	if got := countEvidenceBlobLinks(t, harness, recordID); got != 1 {
		t.Fatalf("blob-backed create links = %d, want 1", got)
	}
	if got := countEvidenceRevisions(t, harness, recordID); got != 1 {
		t.Fatalf("blob-backed create revisions = %d, want 1", got)
	}

	replayResponse := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.evidence.v1/rows", createBody, authOptions(login)...)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResponse, http.StatusOK)["data"].(map[string]any)
	if replayData["change_set_id"] != createData["change_set_id"] {
		t.Fatalf("replay change_set_id = %#v, want %#v", replayData["change_set_id"], createData["change_set_id"])
	}

	reuseResponse := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.evidence.v1/rows", map[string]any{
		"client_txn_id":                   "txn-evidence-atomic-create-reuse",
		"evidence.title":                  "must not commit",
		"evidence.initial_object_blob_id": slot["object_blob_id"],
	}, authOptions(login)...)
	reuseError := httptestx.RequireErrorEnvelope(t, reuseResponse, http.StatusConflict, "evidence_attach_rejected")
	httptestx.RequireErrorDetail(t, reuseError, "reason_code", evidence.AttachReasonBlobNotVisible)
	var rowCount int
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM evidence
 WHERE incident_id = $1
`, incidentID).Scan(&rowCount); err != nil {
		t.Fatalf("count Evidence rows: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("Evidence row count after losing create = %d, want 1", rowCount)
	}
}

func TestObjectUploadCapabilityRoute_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence_lifecycle-object-upload-capability")
	login, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-upload-capability-incident",
		"incident_key":  "evidence_lifecycle-upload-capability",
		"title":         "Evidence upload capability",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	createSlot := func(t *testing.T, txn string) map[string]any {
		t.Helper()
		resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
			"incident_id":       incidentID.String(),
			"client_txn_id":     txn,
			"byte_size":         5,
			"filename_hint":     "capability.txt",
			"content_type_hint": "text/plain",
		}, authOptions(login)...)
		data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
		target := data["upload_target"].(map[string]any)
		href, _ := target["href"].(string)
		if !strings.HasPrefix(href, "/api/v1/object-uploads/upl_") || strings.Contains(href, "://") {
			t.Fatalf("upload target must be an opaque same-origin capability: %#v", target)
		}
		return data
	}
	putUpload := func(t *testing.T, href string, payload string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(http.MethodPut, harness.Server.HTTP.URL+href, strings.NewReader(payload))
		if err != nil {
			t.Fatalf("create upload request: %v", err)
		}
		req.Header.Set("Content-Type", "text/plain")
		for _, option := range authOptions(login) {
			option(req)
		}
		return httptestx.Do(t, http.DefaultClient, req)
	}

	malformed := putUpload(t, "/api/v1/object-uploads/not-a-token", "hello")
	httptestx.RequireErrorEnvelope(t, malformed, http.StatusNotFound, "object_upload_not_found_or_revoked")

	undersizeData := createSlot(t, "txn-evidence_lifecycle-upload-capability-undersize")
	undersizeTarget := undersizeData["upload_target"].(map[string]any)
	undersizeResp := putUpload(t, undersizeTarget["href"].(string), "hell")
	undersizeBody := httptestx.RequireErrorEnvelope(t, undersizeResp, http.StatusBadRequest, "object_upload_rejected")
	httptestx.RequireErrorDetail(t, undersizeBody, "reason_code", "byte_size_mismatch")

	oversizeData := createSlot(t, "txn-evidence_lifecycle-upload-capability-oversize")
	oversizeTarget := oversizeData["upload_target"].(map[string]any)
	oversizeResp := putUpload(t, oversizeTarget["href"].(string), "helloo")
	oversizeBody := httptestx.RequireErrorEnvelope(t, oversizeResp, http.StatusRequestEntityTooLarge, "object_upload_rejected")
	httptestx.RequireErrorDetail(t, oversizeBody, "reason_code", "byte_size_exceeds_contract")

	wrongStateData := createSlot(t, "txn-evidence_lifecycle-upload-capability-wrong-state")
	wrongStateBlobID := appsupport.MustUUID(t, wrongStateData["object_blob_id"].(string))
	updateBlobState(t, harness, wrongStateBlobID, "available")
	wrongStateTarget := wrongStateData["upload_target"].(map[string]any)
	wrongStateResp := putUpload(t, wrongStateTarget["href"].(string), "hello")
	wrongStateBody := httptestx.RequireErrorEnvelope(t, wrongStateResp, http.StatusConflict, "object_upload_rejected")
	httptestx.RequireErrorDetail(t, wrongStateBody, "reason_code", "blob_not_pending")

	successData := createSlot(t, "txn-evidence_lifecycle-upload-capability-success")
	successTarget := successData["upload_target"].(map[string]any)
	successResp := putUpload(t, successTarget["href"].(string), "hello")
	httptestx.RequireStatus(t, successResp, http.StatusNoContent)
}

func TestExpiredSlotReplay_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence_lifecycle-expired-slot-replay")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-expired-slot-incident",
		"incident_key":  "evidence_lifecycle-expired-slot",
		"title":         "Evidence expired slot replay",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	issuedAt := time.Now().UTC().Add(time.Minute).Truncate(time.Second)
	httptestx.SetClockFixed(t, harness.Server, issuedAt)
	createBody := map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-evidence_lifecycle-expired-slot",
		"byte_size":         17,
		"filename_hint":     " expired.txt ",
		"content_type_hint": "text/plain",
	}
	createResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", createBody, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	requireCreateExpiry(t, createData, "target_expires_at", issuedAt.Add(60*time.Minute))
	requireCreateExpiry(t, createData, "pending_expires_at", issuedAt.Add(24*time.Hour))

	replayAt := issuedAt.Add(61 * time.Minute)
	extendSessionForClockJump(t, harness, adminID, replayAt.Add(30*time.Minute))
	httptestx.SetClockFixed(t, harness.Server, replayAt)
	replayResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", createBody, authOptions(login)...)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	for _, key := range []string{"object_blob_id", "target_expires_at", "pending_expires_at"} {
		if replayData[key] != createData[key] {
			t.Fatalf("expired replay refreshed %s: replay=%#v create=%#v", key, replayData, createData)
		}
	}
	replayedTarget := mustParseTime(t, replayData["target_expires_at"].(string))
	if !replayedTarget.Before(replayAt) {
		t.Fatalf("replayed target should remain expired: got %s replay_at %s", replayedTarget, replayAt)
	}
	replayedUploadTarget := replayData["upload_target"].(map[string]any)
	replayedHref, _ := replayedUploadTarget["href"].(string)
	expiredReq, err := http.NewRequest(http.MethodPut, harness.Server.HTTP.URL+replayedHref, strings.NewReader(strings.Repeat("x", 17)))
	if err != nil {
		t.Fatalf("create expired upload request: %v", err)
	}
	expiredReq.Header.Set("Content-Type", "text/plain")
	for _, option := range authOptions(login) {
		option(expiredReq)
	}
	expiredUpload := httptestx.Do(t, http.DefaultClient, expiredReq)
	httptestx.RequireErrorEnvelope(t, expiredUpload, http.StatusGone, "object_upload_expired")

	freshBody := map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-evidence_lifecycle-expired-slot-fresh",
		"byte_size":         17,
		"filename_hint":     " expired.txt ",
		"content_type_hint": "text/plain",
	}
	freshResp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", freshBody, authOptions(login)...)
	freshData := httptestx.RequireSuccessEnvelope(t, freshResp, http.StatusCreated)["data"].(map[string]any)
	if freshData["object_blob_id"] == createData["object_blob_id"] {
		t.Fatalf("fresh client_txn_id reused expired object_blob_id: %#v", freshData)
	}
	requireCreateExpiry(t, freshData, "target_expires_at", replayAt.Add(60*time.Minute))
	requireCreateExpiry(t, freshData, "pending_expires_at", replayAt.Add(24*time.Hour))
	if got := countObjectBlobs(t, harness, incidentID); got != 2 {
		t.Fatalf("fresh target should create exactly one additional blob slot: got %d want 2", got)
	}
}
