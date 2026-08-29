package evidence_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestEvidenceObjectBoundaryAndActiveContent_Integration(t *testing.T) {
	t.Run("AC-405 object bytes stay outside structured state and loss fails closed", func(t *testing.T) {
		harness := appsupport.StartServer(t, "evidence_lifecycle-i-04-object-boundary")
		login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-evidence_lifecycle-i-04-boundary-incident",
			"incident_key":  "evidence_lifecycle-i-04-boundary",
			"title":         "Evidence object boundary",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, recordID)

		marker := "evidence_lifecycle-ac405-marker-" + uuid.NewString() + "-payload"
		payload := []byte("prefix-" + marker + "-suffix")
		attachData := attachUploadedBlobWithMetadata(t, harness, login, incidentID, recordID, payload, "boundary.txt", "text/plain", "txn-evidence_lifecycle-i-04-boundary-blob", "txn-evidence_lifecycle-i-04-boundary-attach")
		objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))

		preview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
		if got := string(redeemHandle(t, harness.Server.HTTP.URL+preview["href"].(string), login)); got != string(payload) {
			t.Fatalf("preview before object loss got %q want %q", got, string(payload))
		}
		download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
		if got := string(redeemHandle(t, harness.Server.HTTP.URL+download["href"].(string), login)); got != string(payload) {
			t.Fatalf("download before object loss got %q want %q", got, string(payload))
		}

		structured := structuredTableText(t, harness,
			"records", "evidence", "object_blobs", "route_idempotency",
			"change_sets", "change_set_mutations", "record_revisions", "evidence_access_handles",
		)
		if strings.Contains(structured, marker) {
			t.Fatalf("structured Postgres state contains inline evidence payload marker %q", marker)
		}

		storageKey := blobStorageKey(t, harness, objectBlobID)
		if err := harness.ObjectStore.DeleteObject(context.Background(), storageKey); err != nil {
			t.Fatalf("delete object bytes: %v", err)
		}
		if got := countEvidenceBlobLinks(t, harness, recordID); got != 1 {
			t.Fatalf("object loss changed committed evidence blob link count: got %d want 1", got)
		}
		requireEvidenceStates(t, harness, recordID, "available", "available")
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			requireEvidenceAccessUnavailableReason(t,
				appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...),
				"blob_missing",
			)
		}
	})

	t.Run("active content uses observed media state for preview policy", func(t *testing.T) {
		for _, active := range []struct {
			name        string
			contentType string
			filename    string
		}{
			{name: "html", contentType: "text/html", filename: "pretend-image.png"},
			{name: "svg", contentType: "image/svg+xml", filename: "pretend-raster.png"},
		} {
			t.Run(active.name, func(t *testing.T) {
				harness := appsupport.StartServer(t, "evidence_lifecycle-i-04-active-"+active.name)
				login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
				incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
					"client_txn_id": "txn-evidence_lifecycle-i-04-active-" + active.name + "-incident",
					"incident_key":  "evidence_lifecycle-i-04-active-" + active.name,
					"title":         "Evidence active content " + active.name,
				})
				incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
				recordID := uuid.New()
				seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
				payload := []byte("<script>window.__cartulary_evidence_lifecycle_active_content = true</script>")
				attachData := attachUploadedBlobWithHints(t, harness, login, incidentID, recordID, payload, active.filename, active.contentType, active.contentType, "txn-evidence_lifecycle-i-04-active-"+active.name+"-blob", "txn-evidence_lifecycle-i-04-active-"+active.name+"-attach")
				objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))
				requireObservedContentType(t, harness, objectBlobID, active.contentType)

				requireEvidenceAccessUnavailableReason(t,
					appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...),
					"unsupported_preview",
				)
				download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
				if got := string(redeemHandle(t, harness.Server.HTTP.URL+download["href"].(string), login)); got != string(payload) {
					t.Fatalf("download active content got %q want %q", got, string(payload))
				}
			})
		}
	})
}
