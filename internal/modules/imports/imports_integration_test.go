package imports_test

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestExtensionImportUploadEarlyFailCreatesNoDurableRows(t *testing.T) {
	restore := claimImportProfileForTest()
	t.Cleanup(restore)

	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-import-early-fail")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase11-import-early-fail-incident",
		"incident_key":  "IR-PHASE11-EARLY",
		"title":         "Phase 11 import early fail",
	})

	metadata := `{"incident_id":"` + incident["incident_id"].(string) + `","client_txn_id":"txn-phase11-import-early-fail","extra":true}`
	resp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, "host,summary\nhost-1,alpha\n", "input.csv", false)
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_import_request")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "unknown_field" {
		t.Fatalf("unexpected import rejection details: %#v", details)
	}

	requireImportCounts(t, harness.DB, importCounts{})
}

func TestPhase11_I_11_IMPORT_01_UploadMetadataNonObjectCreatesNoDurableRows(t *testing.T) {
	restore := claimImportProfileForTest()
	t.Cleanup(restore)

	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-import-metadata-non-object")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)

	resp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, `[]`, "host,summary\nhost-1,alpha\n", "input.csv", false)
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_import_request")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "request_not_object" {
		t.Fatalf("unexpected import rejection details: %#v", details)
	}

	requireImportCounts(t, harness.DB, importCounts{})
}

func TestExtensionImportUploadExactReplayAndReadResources(t *testing.T) {
	restore := claimImportProfileForTest()
	t.Cleanup(restore)

	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-import-replay")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase11-import-replay-incident",
		"incident_key":  "IR-PHASE11-REPLAY",
		"title":         "Phase 11 import replay",
	})
	incidentID := incident["incident_id"].(string)
	metadata := `{"client_txn_id":"txn-phase11-import-replay","incident_id":"` + incidentID + `"}`
	csv := "host,summary\nhost-1,alpha\nhost-2,beta\n"

	firstResp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, csv, "first.csv", false)
	firstJob := httptestx.RequireSuccessEnvelope(t, firstResp, http.StatusAccepted)["data"].(map[string]any)
	firstJobID := firstJob["job_id"].(string)

	replayResp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, csv, "different-name.csv", true)
	replayJob := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusAccepted)["data"].(map[string]any)
	if replayJob["job_id"] != firstJobID {
		t.Fatalf("exact replay returned different job: first=%q replay=%q", firstJobID, replayJob["job_id"])
	}
	requireImportCounts(t, harness.DB, importCounts{Sessions: 1, Units: 1, Jobs: 1, RouteIdempotency: 1})

	divergentResp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, "host,summary\nhost-3,gamma\n", "first.csv", false)
	httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")
	requireImportCounts(t, harness.DB, importCounts{Sessions: 1, Units: 1, Jobs: 1, RouteIdempotency: 1})

	jobResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+firstJobID, nil, phase2test.WithCookies(adminLogin.SessionCookie))
	job := httptestx.RequireSuccessEnvelope(t, jobResp, http.StatusOK)["data"].(map[string]any)
	if job["status"] != "succeeded" {
		t.Fatalf("discovery job status = %#v, want succeeded", job["status"])
	}
	resultSummary := job["result_summary"].(map[string]any)
	refs := resultSummary["resource_refs"].([]any)
	sessionID := refs[0].(map[string]any)["id"].(string)

	sessionResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID, nil, phase2test.WithCookies(adminLogin.SessionCookie))
	session := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	if session["source_file_kind"] != imports.SourceFileKindCSV || session["original_filename"] != "first.csv" || session["session_status"] != "discovered" {
		t.Fatalf("unexpected import session resource: %#v", session)
	}
	if session["parser_profile_id"] != imports.ParserProfilePhase2WorkbookImport || session["parser_version"] != imports.ParserVersionPhase11 {
		t.Fatalf("unexpected parser provenance: %#v", session)
	}

	unitsResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units?limit=1", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	unitsBody := httptestx.RequireSuccessEnvelope(t, unitsResp, http.StatusOK)
	paging := unitsBody["meta"].(map[string]any)["paging"].(map[string]any)
	if paging["limit"] != float64(1) || paging["has_more"] != false || paging["next_cursor"] != nil {
		t.Fatalf("unexpected import unit paging meta: %#v", paging)
	}
	units := unitsBody["data"].(map[string]any)["import_units"].([]any)
	if len(units) != 1 {
		t.Fatalf("expected one import unit, got %#v", units)
	}
	unit := units[0].(map[string]any)
	unitID := unit["import_unit_id"].(string)
	if unit["unit_status"] != "discovered" || unit["locator_kind"] != "csv_file" {
		t.Fatalf("unexpected import unit resource: %#v", unit)
	}

	previewResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/preview", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	preview := httptestx.RequireSuccessEnvelope(t, previewResp, http.StatusOK)["data"].(map[string]any)
	if preview["truncated"] != false {
		t.Fatalf("unexpected preview truncation: %#v", preview)
	}
	rows := preview["preview_rows"].([]any)
	if len(rows) != 2 {
		t.Fatalf("expected two preview rows, got %#v", rows)
	}
}

func TestPhase11_I_11_IMPORT_03_XLSXDiscoveryUsesBoundedUsedRange(t *testing.T) {
	restore := claimImportProfileForTest()
	t.Cleanup(restore)

	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-import-xlsx-discovery")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase11-import-xlsx-incident",
		"incident_key":  "IR-PHASE11-XLSX",
		"title":         "Phase 11 XLSX import",
	})
	incidentID := incident["incident_id"].(string)
	metadata := `{"client_txn_id":"txn-phase11-import-xlsx-upload","incident_id":"` + incidentID + `"}`

	uploadResp := postImportUploadBytes(t, harness.Server.HTTP.URL, adminLogin, metadata, minimalXLSX(t), "input.xlsx", imports.MediaTypeXLSX, false)
	uploadJob := httptestx.RequireSuccessEnvelope(t, uploadResp, http.StatusAccepted)["data"].(map[string]any)
	jobResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+uploadJob["job_id"].(string), nil, phase2test.WithCookies(adminLogin.SessionCookie))
	job := httptestx.RequireSuccessEnvelope(t, jobResp, http.StatusOK)["data"].(map[string]any)
	if job["status"] != "succeeded" || job["result_summary"].(map[string]any)["code"] != "import_session_discovered" {
		t.Fatalf("unexpected discovery job: %#v", job)
	}
	sessionID := job["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)

	sessionResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID, nil, phase2test.WithCookies(adminLogin.SessionCookie))
	session := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	if session["source_file_kind"] != imports.SourceFileKindXLSX {
		t.Fatalf("unexpected XLSX session: %#v", session)
	}

	unitsResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	units := httptestx.RequireSuccessEnvelope(t, unitsResp, http.StatusOK)["data"].(map[string]any)["import_units"].([]any)
	unit := units[0].(map[string]any)
	if unit["locator_kind"] != "xlsx_used_range" || unit["inferred_row_count"] != float64(2) || unit["inferred_column_count"] != float64(2) {
		t.Fatalf("unexpected XLSX unit: %#v", unit)
	}

	previewResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units/"+unit["import_unit_id"].(string)+"/preview", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	preview := httptestx.RequireSuccessEnvelope(t, previewResp, http.StatusOK)["data"].(map[string]any)
	columns := preview["columns"].([]any)
	if columns[0].(map[string]any)["source_header_text"] != "host" || columns[1].(map[string]any)["source_header_text"] != "summary" {
		t.Fatalf("unexpected XLSX columns: %#v", columns)
	}
	rows := preview["preview_rows"].([]any)
	if len(rows) != 2 || rows[1].(map[string]any)["cells"].([]any)[1].(map[string]any)["display_text"] != "Beta summary" {
		t.Fatalf("unexpected XLSX preview rows: %#v", rows)
	}
}

func TestPhase11_I_11_IMPORT_02_MappingSelectApplyCreatesTimelineRows(t *testing.T) {
	restore := claimImportProfileForTest()
	t.Cleanup(restore)

	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-import-apply")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase11-import-apply-incident",
		"incident_key":  "IR-PHASE11-APPLY",
		"title":         "Phase 11 import apply",
	})
	incidentID := incident["incident_id"].(string)
	metadata := `{"client_txn_id":"txn-phase11-import-apply-upload","incident_id":"` + incidentID + `"}`
	csv := "host,summary\nhost-1, Alpha summary \nhost-2,Beta summary\n"

	uploadResp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, csv, "apply.csv", false)
	uploadJob := httptestx.RequireSuccessEnvelope(t, uploadResp, http.StatusAccepted)["data"].(map[string]any)
	jobResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+uploadJob["job_id"].(string), nil, phase2test.WithCookies(adminLogin.SessionCookie))
	job := httptestx.RequireSuccessEnvelope(t, jobResp, http.StatusOK)["data"].(map[string]any)
	sessionID := job["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)

	unitsResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	units := httptestx.RequireSuccessEnvelope(t, unitsResp, http.StatusOK)["data"].(map[string]any)["import_units"].([]any)
	unitID := units[0].(map[string]any)["import_unit_id"].(string)

	mapping := map[string]any{
		"client_txn_id":         "txn-phase11-import-apply-mapping",
		"target_view_schema_id": "cartulary.view.timeline.v2",
		"header_row_ref":        1,
		"data_start_row_ref":    2,
		"unknown_column_policy": "preserve_raw_capture",
		"source_columns": []map[string]any{
			{
				"source_column_ordinal": 1,
				"source_header_text":    "host",
				"field_key":             nil,
				"entity_binding_mode":   nil,
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
			{
				"source_column_ordinal": 2,
				"source_header_text":    "summary",
				"field_key":             "timeline.activity_synopsis_text",
				"entity_binding_mode":   nil,
				"transform_id":          "trim_v1",
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
		},
	}
	mappingResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPut, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping", mapping)
	mappedUnit := httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)["data"].(map[string]any)
	if mappedUnit["unit_status"] != "mapped" || mappedUnit["mapping_fingerprint"] == nil || mappedUnit["approved_mapping"] == nil {
		t.Fatalf("unexpected mapped unit: %#v", mappedUnit)
	}

	selectResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select", map[string]any{"client_txn_id": "txn-phase11-import-apply-select"})
	selected := httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)["data"].(map[string]any)
	if selected["session_status"] != "ready_to_apply" || selected["unit"].(map[string]any)["unit_status"] != "ready" {
		t.Fatalf("unexpected selected state: %#v", selected)
	}

	applyResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-phase11-import-apply-apply"})
	applyJob := httptestx.RequireSuccessEnvelope(t, applyResp, http.StatusAccepted)["data"].(map[string]any)
	applyJobResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+applyJob["job_id"].(string), nil, phase2test.WithCookies(adminLogin.SessionCookie))
	appliedJob := httptestx.RequireSuccessEnvelope(t, applyJobResp, http.StatusOK)["data"].(map[string]any)
	if appliedJob["status"] != "succeeded" || appliedJob["result_summary"].(map[string]any)["code"] != "import_session_applied" {
		t.Fatalf("unexpected apply job: %#v", appliedJob)
	}

	queryResp := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.timeline.v2/query", map[string]any{}, phase2test.WithCookies(adminLogin.SessionCookie))
	rows := httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)["data"].(map[string]any)["rows"].([]any)
	if len(rows) != 2 {
		t.Fatalf("expected two imported timeline rows, got %#v", rows)
	}
	summaries := map[string]bool{}
	for _, rowAny := range rows {
		row := rowAny.(map[string]any)
		cells := row["cells"].(map[string]any)
		summaries[cells["timeline.activity_synopsis_text"].(map[string]any)["value"].(string)] = true
	}
	if !summaries["Alpha summary"] || !summaries["Beta summary"] {
		t.Fatalf("imported summaries not found: %#v", summaries)
	}

	duplicateApply := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-phase11-import-apply-second"})
	errBody := httptestx.RequireErrorEnvelope(t, duplicateApply, http.StatusConflict, "import_apply_blocked")
	if errBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "duplicate_apply_blocked" {
		t.Fatalf("unexpected duplicate apply error: %#v", errBody)
	}
}

type importCounts struct {
	Sessions         int
	Units            int
	Jobs             int
	RouteIdempotency int
}

func claimImportProfileForTest() func() {
	profiles := httpapi.CurrentExtensionProfiles()
	for index := range profiles {
		if profiles[index].ProfileID == imports.ProfileID {
			profiles[index].Claimed = true
		}
	}
	return httpapi.SetCurrentExtensionProfilesForTesting(profiles)
}

func postImportUpload(t testing.TB, serverURL string, login phase2test.LoginResult, metadata string, file string, filename string, fileFirst bool) *http.Response {
	t.Helper()
	return postImportUploadBytes(t, serverURL, login, metadata, []byte(file), filename, imports.MediaTypeCSV, fileFirst)
}

func postImportUploadBytes(t testing.TB, serverURL string, login phase2test.LoginResult, metadata string, file []byte, filename string, contentType string, fileFirst bool) *http.Response {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if fileFirst {
		writeFilePart(t, writer, filename, contentType, file)
		writeMetadataPart(t, writer, metadata)
	} else {
		writeMetadataPart(t, writer, metadata)
		writeFilePart(t, writer, filename, contentType, file)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, serverURL+"/api/v1/import-sessions", &body)
	if err != nil {
		t.Fatalf("build import request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set(authn.CSRFHeaderName, login.CSRFCookie.Value)
	req.AddCookie(login.SessionCookie)
	req.AddCookie(login.CSRFCookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post import upload: %v", err)
	}
	return resp
}

func doImportJSON(t testing.TB, serverURL string, login phase2test.LoginResult, method string, path string, body any) *http.Response {
	t.Helper()
	return phase2test.DoJSON(
		t,
		method,
		serverURL+path,
		body,
		phase2test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func writeMetadataPart(t testing.TB, writer *multipart.Writer, metadata string) {
	t.Helper()
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{`form-data; name="metadata"`},
		"Content-Type":        []string{"application/json"},
	})
	if err != nil {
		t.Fatalf("create metadata part: %v", err)
	}
	if _, err := io.WriteString(part, metadata); err != nil {
		t.Fatalf("write metadata part: %v", err)
	}
}

func writeFilePart(t testing.TB, writer *multipart.Writer, filename string, contentType string, file []byte) {
	t.Helper()
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{`form-data; name="file"; filename="` + filename + `"`},
		"Content-Type":        []string{contentType},
	})
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := part.Write(file); err != nil {
		t.Fatalf("write file part: %v", err)
	}
}

func minimalXLSX(t testing.TB) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	writeZipText(t, writer, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`)
	writeZipText(t, writer, "xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets><sheet name="Sheet1" sheetId="1" r:id="rId1"/></sheets>
</workbook>`)
	writeZipText(t, writer, "xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`)
	writeZipText(t, writer, "xl/worksheets/sheet1.xml", `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="inlineStr"><is><t>host</t></is></c><c r="B1" t="inlineStr"><is><t>summary</t></is></c></row>
    <row r="2"><c r="A2" t="inlineStr"><is><t>host-1</t></is></c><c r="B2" t="inlineStr"><is><t>Alpha summary</t></is></c></row>
    <row r="3"><c r="A3" t="inlineStr"><is><t>host-2</t></is></c><c r="B3" t="inlineStr"><is><t>Beta summary</t></is></c></row>
  </sheetData>
</worksheet>`)
	if err := writer.Close(); err != nil {
		t.Fatalf("close XLSX zip: %v", err)
	}
	return buffer.Bytes()
}

func writeZipText(t testing.TB, writer *zip.Writer, name string, content string) {
	t.Helper()
	entry, err := writer.Create(name)
	if err != nil {
		t.Fatalf("create XLSX entry %s: %v", name, err)
	}
	if _, err := io.WriteString(entry, content); err != nil {
		t.Fatalf("write XLSX entry %s: %v", name, err)
	}
}

func requireImportCounts(t testing.TB, db *sql.DB, want importCounts) {
	t.Helper()
	got := importCounts{
		Sessions:         phase2test.QueryCount(t, db, `SELECT COUNT(*) FROM import_sessions`),
		Units:            phase2test.QueryCount(t, db, `SELECT COUNT(*) FROM import_units`),
		Jobs:             phase2test.QueryCount(t, db, `SELECT COUNT(*) FROM jobs`),
		RouteIdempotency: phase2test.QueryCount(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key LIKE 'imports.%'`),
	}
	if got != want {
		t.Fatalf("unexpected import durable counts: got %+v want %+v", got, want)
	}
}
