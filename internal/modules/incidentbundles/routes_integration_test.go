package incidentbundles_test

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestPhase11_I_11_INCIDENT_BUNDLES_01_ExportJobIdempotencyAndDescriptor(t *testing.T) {
	withIncidentPortabilityClaimed(t)
	harness := phase2test.StartRuntime(t).StartServer(t, "phase11-incident-bundle-export")
	admin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-incident-bundle-source",
		"incident_key":  "BUNDLE-EXPORT",
		"title":         "Incident bundle export",
	})
	incidentID := incident["incident_id"].(string)
	phase2test.CreateTimelineRow(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id":    "txn-incident-bundle-row",
		"timeline.summary": "Portable event",
	})

	first := postExport(t, harness.Server, admin, map[string]any{
		"incident_id":           incidentID,
		"client_txn_id":         "txn-export-bundle",
		"optional_sections":     []string{"snapshots", "reference_packs", "snapshots"},
		"reference_pack_mode":   "refs_only",
		"required_capabilities": []string{},
	})
	job := httptestx.RequireSuccessEnvelope(t, first, http.StatusAccepted)["data"].(map[string]any)
	replay := postExport(t, harness.Server, admin, map[string]any{
		"incident_id":           incidentID,
		"client_txn_id":         "txn-export-bundle",
		"optional_sections":     []string{"reference_packs", "snapshots"},
		"reference_pack_mode":   "refs_only",
		"required_capabilities": []string{},
	})
	replayedJob := httptestx.RequireSuccessEnvelope(t, replay, http.StatusAccepted)["data"].(map[string]any)
	if replayedJob["job_id"] != job["job_id"] {
		t.Fatalf("exact replay returned different job: first=%v replay=%v", job["job_id"], replayedJob["job_id"])
	}
	divergent := postExport(t, harness.Server, admin, map[string]any{
		"incident_id":         incidentID,
		"client_txn_id":       "txn-export-bundle",
		"reference_pack_mode": "embedded",
	})
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")

	terminal := waitJob(t, harness.Server, admin, job["job_id"].(string))
	summary := terminal["result_summary"].(map[string]any)
	if summary["code"] != incidentbundles.ResultIncidentBundleExported {
		t.Fatalf("unexpected export summary: %#v", summary)
	}
	refs := summary["resource_refs"].([]any)
	if len(refs) != 1 || refs[0].(map[string]any)["kind"] != "incident_bundle" {
		t.Fatalf("export summary must contain one incident_bundle ref: %#v", refs)
	}
	descriptorRoute := refs[0].(map[string]any)["route"].(string)
	resp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+descriptorRoute, nil, phase2test.WithCookies(admin.SessionCookie))
	descriptor := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	if descriptor["history_mode"] != "full" || descriptor["blob_mode"] != "full" || descriptor["manifest_sha256"] == "" {
		t.Fatalf("descriptor missing fixed modes or manifest hash: %#v", descriptor)
	}
}

func TestPhase11_I_11_INCIDENT_BUNDLES_02_ImportEnvelopeIdempotencyAndImportedIncidentOpen(t *testing.T) {
	withIncidentPortabilityClaimed(t)
	sourceHarness := phase2test.StartRuntime(t).StartServer(t, "phase11-incident-bundle-source")
	sourceAdmin, _ := phase2test.ProvisionBootstrapAdmin(t, sourceHarness.Server)
	incident := phase2test.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-incident-bundle-import-source",
		"incident_key":  "BUNDLE-IMPORT",
		"title":         "Incident bundle import",
	})
	incidentID := incident["incident_id"].(string)
	row := phase2test.CreateTimelineRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":    "txn-incident-bundle-import-row",
		"timeline.summary": "Imported portable event",
	})
	recordID := row["row"].(map[string]any)["record_id"].(string)
	sourceViewerID := phase2test.SeedLocalUserFlags(t, sourceHarness.DB, "phase11-import-source-viewer@example.test", "Phase 11 Import Source Viewer", "Phase11ImportViewer1!", false, false, true)
	phase2test.CreateMembership(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id": "txn-incident-bundle-source-viewer",
		"user_id":       sourceViewerID,
		"role":          "viewer",
	})
	exportJob := httptestx.RequireSuccessEnvelope(t, postExport(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"incident_id":   incidentID,
		"client_txn_id": "txn-export-for-import",
	}), http.StatusAccepted)["data"].(map[string]any)
	exportTerminal := waitJob(t, sourceHarness.Server, sourceAdmin, exportJob["job_id"].(string))
	exportRef := exportTerminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)
	bundleID := exportRef["id"].(string)
	var bundlePath string
	if err := sourceHarness.DB.QueryRow(`SELECT bundle_storage_path FROM incident_bundle_exports WHERE bundle_id = $1`, bundleID).Scan(&bundlePath); err != nil {
		t.Fatalf("query exported bundle path: %v", err)
	}
	bundleBytes, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatalf("read exported bundle: %v", err)
	}

	if _, err := sourceHarness.DB.Exec(`DELETE FROM deployment_admin_audit_events WHERE incident_id = $1`, incidentID); err != nil {
		t.Fatalf("remove source incident audit events before import: %v", err)
	}
	if _, err := sourceHarness.DB.Exec(`DELETE FROM incidents WHERE id = $1`, incidentID); err != nil {
		t.Fatalf("remove source incident before import: %v", err)
	}

	corruptBytes := corruptZipMember(t, bundleBytes, "records.ndjson")
	failed := postImport(t, sourceHarness.Server, sourceAdmin, `{"client_txn_id":"txn-import-bundle-corrupt"}`, corruptBytes, "bundle-corrupt.zip")
	failedJob := httptestx.RequireSuccessEnvelope(t, failed, http.StatusAccepted)["data"].(map[string]any)
	failedTerminal := waitFailedJob(t, sourceHarness.Server, sourceAdmin, failedJob["job_id"].(string))
	errorSummary := failedTerminal["error_summary"].(map[string]any)
	if errorSummary["code"] != "incident_bundle_import_rejected" {
		t.Fatalf("corrupt import must fail closed with incident_bundle_import_rejected: %#v", errorSummary)
	}
	details := errorSummary["details"].(map[string]any)
	if details["reason_code"] != "checksum_mismatch" {
		t.Fatalf("corrupt import reason mismatch: %#v", details)
	}
	if countRows(t, sourceHarness.DB, `SELECT count(*) FROM incidents WHERE id = $1`, incidentID) != 0 {
		t.Fatalf("failed import must not make incident visible")
	}
	var failedStagingPath string
	if err := sourceHarness.DB.QueryRow(`SELECT bundle_staging_path FROM incident_bundle_job_payloads WHERE job_id = $1`, failedJob["job_id"].(string)).Scan(&failedStagingPath); err != nil {
		t.Fatalf("query failed import staging path: %v", err)
	}
	if _, err := os.Stat(failedStagingPath); !os.IsNotExist(err) {
		t.Fatalf("failed import staging path must be cleaned up, stat err=%v path=%s", err, failedStagingPath)
	}

	first := postImport(t, sourceHarness.Server, sourceAdmin, `{"client_txn_id":"txn-import-bundle"}`, bundleBytes, "bundle.zip")
	importJob := httptestx.RequireSuccessEnvelope(t, first, http.StatusAccepted)["data"].(map[string]any)
	replay := postImport(t, sourceHarness.Server, sourceAdmin, `{"client_txn_id":"txn-import-bundle"}`, bundleBytes, "bundle-renamed.zip")
	replayedJob := httptestx.RequireSuccessEnvelope(t, replay, http.StatusAccepted)["data"].(map[string]any)
	if replayedJob["job_id"] != importJob["job_id"] {
		t.Fatalf("exact import replay returned different job: first=%v replay=%v", importJob["job_id"], replayedJob["job_id"])
	}
	divergent := postImport(t, sourceHarness.Server, sourceAdmin, `{"client_txn_id":"txn-import-bundle"}`, append(bundleBytes, 0), "bundle.zip")
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")

	terminal := waitJob(t, sourceHarness.Server, sourceAdmin, importJob["job_id"].(string))
	summary := terminal["result_summary"].(map[string]any)
	if summary["code"] != incidentbundles.ResultIncidentBundleImported {
		t.Fatalf("unexpected import summary: %#v", summary)
	}
	ref := summary["resource_refs"].([]any)[0].(map[string]any)
	if ref["kind"] != "incident" || ref["id"] != incidentID {
		t.Fatalf("import summary must reference imported incident: %#v", ref)
	}
	openResp := phase2test.DoJSON(t, http.MethodGet, sourceHarness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-startup", nil, phase2test.WithCookies(sourceAdmin.SessionCookie))
	httptestx.RequireSuccessEnvelope(t, openResp, http.StatusOK)
	var projectionCount int
	if err := sourceHarness.DB.QueryRow(`SELECT count(*) FROM timeline_grid_projection WHERE record_id = $1`, recordID).Scan(&projectionCount); err != nil {
		t.Fatalf("query imported projection: %v", err)
	}
	if projectionCount != 1 {
		t.Fatalf("imported timeline projection missing, count=%d", projectionCount)
	}
	if countRows(t, sourceHarness.DB, `SELECT count(*) FROM incident_memberships WHERE incident_id = $1`, incidentID) != 1 {
		t.Fatalf("import should create only local importer membership")
	}
	if countRows(t, sourceHarness.DB, `SELECT count(*) FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`, incidentID, sourceViewerID) != 0 {
		t.Fatalf("import must not recreate source deployment-local membership")
	}
	if countRows(t, sourceHarness.DB, `SELECT count(*) FROM incident_bundle_imported_actors WHERE incident_id = $1`, incidentID) == 0 {
		t.Fatalf("import must preserve historical attribution as inert actor descriptors")
	}
}

func postExport(t testing.TB, server *httptestx.Server, login phase2test.LoginResult, body map[string]any) *http.Response {
	t.Helper()
	return phase2test.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/incident-bundles/export", body,
		phase2test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func postImport(t testing.TB, server *httptestx.Server, login phase2test.LoginResult, metadata string, file []byte, filename string) *http.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	metadataPart, err := writer.CreatePart(textprotoMIMEHeader(map[string]string{
		"Content-Disposition": `form-data; name="metadata"`,
		"Content-Type":        "application/json; charset=utf-8",
	}))
	if err != nil {
		t.Fatalf("create metadata part: %v", err)
	}
	if _, err := io.WriteString(metadataPart, metadata); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	filePart, err := writer.CreatePart(textprotoMIMEHeader(map[string]string{
		"Content-Disposition": `form-data; name="file"; filename="` + filename + `"`,
		"Content-Type":        incidentbundles.MediaTypeZip,
	}))
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := filePart.Write(file); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, server.HTTP.URL+"/api/v1/incident-bundles/import", &body)
	if err != nil {
		t.Fatalf("create import request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(login.SessionCookie)
	req.AddCookie(login.CSRFCookie)
	req.Header.Set(authn.CSRFHeaderName, login.CSRFCookie.Value)
	return httptestx.Do(t, http.DefaultClient, req)
}

func waitJob(t testing.TB, server *httptestx.Server, login phase2test.LoginResult, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp := phase2test.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, phase2test.WithCookies(login.SessionCookie))
		job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		status := job["status"].(string)
		switch status {
		case "succeeded":
			return job
		case "failed", "canceled":
			encoded, _ := json.MarshalIndent(job, "", "  ")
			t.Fatalf("job reached terminal non-success state: %s", encoded)
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("job %s did not finish", jobID)
	return nil
}

func waitFailedJob(t testing.TB, server *httptestx.Server, login phase2test.LoginResult, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp := phase2test.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, phase2test.WithCookies(login.SessionCookie))
		job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		status := job["status"].(string)
		switch status {
		case "failed":
			return job
		case "succeeded", "canceled":
			encoded, _ := json.MarshalIndent(job, "", "  ")
			t.Fatalf("job reached unexpected terminal state: %s", encoded)
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("job %s did not fail", jobID)
	return nil
}

func corruptZipMember(t testing.TB, bundle []byte, memberPath string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip for corruption: %v", err)
	}
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	found := false
	for _, member := range reader.File {
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", member.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", member.Name, err)
		}
		if member.Name == memberPath {
			data = append(data, []byte("corrupt\n")...)
			found = true
		}
		w, err := writer.Create(member.Name)
		if err != nil {
			t.Fatalf("create corrupt member %s: %v", member.Name, err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write corrupt member %s: %v", member.Name, err)
		}
	}
	if !found {
		t.Fatalf("zip member %s not found", memberPath)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close corrupt zip: %v", err)
	}
	return buf.Bytes()
}

func textprotoMIMEHeader(values map[string]string) textproto.MIMEHeader {
	header := textproto.MIMEHeader{}
	for key, value := range values {
		header.Set(key, value)
	}
	return header
}

func countRows(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func withIncidentPortabilityClaimed(t testing.TB) {
	t.Helper()
	restore := httpapi.SetCurrentExtensionProfilesForTesting([]httpapi.ExtensionProfile{
		{ProfileID: "enterprise_authentication", Claimed: false, RouteFamilies: []string{"/api/v1/auth/oidc", "/api/v1/auth/providers", "/api/v1/auth/saml", "/api/v1/users/{user_id}/auth-bindings"}},
		{ProfileID: "import", Claimed: true, RouteFamilies: []string{"/api/v1/import-sessions"}},
		{ProfileID: incidentbundles.ProfileID, Claimed: true, RouteFamilies: []string{"/api/v1/incident-bundles"}},
		{ProfileID: "reference_pack", Claimed: true, RouteFamilies: []string{"/api/v1/reference-packs"}},
		{ProfileID: "snapshot_reporting", Claimed: true, RouteFamilies: []string{"/api/v1/releases", "/api/v1/snapshots"}},
	})
	t.Cleanup(restore)
}
