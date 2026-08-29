package evidence_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestQuarantineBoundaryPreservesTwoStepAttach_Integration(t *testing.T) {
	t.Run("quarantine bridges evidence and blocks attach preview and download", func(t *testing.T) {
		harness := appsupport.StartServer(t, "evidence_lifecycle-i-04-quarantine")
		login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-evidence_lifecycle-i-04-quarantine-incident",
			"incident_key":  "evidence_lifecycle-i-04-quarantine",
			"title":         "Evidence quarantine",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
		attachData := attachUploadedBlobWithMetadata(t, harness, login, incidentID, recordID, []byte("evidence_lifecycle quarantine body"), "quarantine.txt", "text/plain", "txn-evidence_lifecycle-i-04-quarantine-blob", "txn-evidence_lifecycle-i-04-quarantine-attach")
		objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))
		preview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
		download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE object_blobs
   SET upload_state = 'quarantined', updated_at = now()
 WHERE object_blob_id = $1
`, objectBlobID); err != nil {
			t.Fatalf("seed quarantined blob state: %v", err)
		}
		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE evidence
   SET lifecycle_state = 'quarantined', upload_state = 'quarantined', updated_at = now()
 WHERE record_id = $1
`, recordID); err != nil {
			t.Fatalf("seed quarantined evidence state: %v", err)
		}
		requireEvidenceStates(t, harness, recordID, "quarantined", "quarantined")
		for _, handle := range []map[string]any{preview, download} {
			requireEvidenceAccessUnavailableReason(t,
				appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, appsupport.WithCookies(login.SessionCookie)),
				"evidence_quarantined",
			)
		}
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			requireEvidenceAccessUnavailableReason(t,
				appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...),
				"evidence_quarantined",
			)
		}
		secondRecordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, secondRecordID)
		attachBlocked := httptestx.RequireErrorEnvelope(t,
			appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+secondRecordID.String()+"/attach-blob", map[string]any{
				"object_blob_id":   objectBlobID.String(),
				"base_row_version": 1,
				"client_txn_id":    "txn-evidence_lifecycle-i-04-quarantine-attach-blocked",
			}, authOptions(login)...),
			http.StatusConflict,
			"evidence_attach_rejected",
		)
		attachDetails := attachBlocked["error"].(map[string]any)["details"].(map[string]any)
		if got := attachDetails["reason_code"]; got != evidence.AttachReasonBlobNotVisible {
			t.Fatalf("associated quarantined blob reason got %v want %s", got, evidence.AttachReasonBlobNotVisible)
		}

	})

}
