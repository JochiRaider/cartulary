package imports_test

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestExtensionImportUploadEarlyFailCreatesNoDurableRows(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "extension_profile-import-early-fail")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-import-early-fail-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-EARLY",
		"title":         "Enterprise integration import early fail",
	})

	metadata := `{"incident_id":"` + incident["incident_id"].(string) + `","client_txn_id":"txn-extension_profile-import-early-fail","extra":true}`
	resp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, "host,summary\nhost-1,alpha\n", "input.csv", false)
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_import_request")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "unknown_field" {
		t.Fatalf("unexpected import rejection details: %#v", details)
	}

	requireImportCounts(t, harness.DB, importCounts{})
}

func TestUploadMetadataNonObjectCreatesNoDurableRows_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "extension_profile-import-metadata-non-object")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	resp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, `[]`, "host,summary\nhost-1,alpha\n", "input.csv", false)
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_import_request")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "request_not_object" {
		t.Fatalf("unexpected import rejection details: %#v", details)
	}

	requireImportCounts(t, harness.DB, importCounts{})
}

func TestExtensionImportUploadExactReplayAndReadResources(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "extension_profile-import-replay")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-import-replay-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-REPLAY",
		"title":         "Enterprise integration import replay",
	})
	incidentID := incident["incident_id"].(string)
	metadata := `{"client_txn_id":"txn-extension_profile-import-replay","incident_id":"` + incidentID + `"}`
	csv := "host,summary\nhost-1,alpha\nhost-2,beta\n"

	firstResp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, csv, "first.csv", false)
	firstJob := httptestx.RequireSuccessEnvelope(t, firstResp, http.StatusAccepted)["data"].(map[string]any)
	firstJobID := firstJob["job_id"].(string)

	replayResp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, csv, "different-name.csv", true)
	replayJob := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusAccepted)["data"].(map[string]any)
	if replayJob["job_id"] != firstJobID {
		t.Fatalf("exact replay returned different job: first=%q replay=%q", firstJobID, replayJob["job_id"])
	}
	requireImportCounts(t, harness.DB, importCounts{Sessions: 1, Units: 1, SourceStreams: 1, Jobs: 1, RouteIdempotency: 1})

	divergentResp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, "host,summary\nhost-3,gamma\n", "first.csv", false)
	httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")
	requireImportCounts(t, harness.DB, importCounts{Sessions: 1, Units: 1, SourceStreams: 1, Jobs: 1, RouteIdempotency: 1})

	job := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, firstJobID)
	if job["status"] != "succeeded" {
		t.Fatalf("discovery job status = %#v, want succeeded", job["status"])
	}
	requireImportProof(t, harness.DB, firstJobID, "import.discovery")
	resultSummary := job["result_summary"].(map[string]any)
	refs := resultSummary["resource_refs"].([]any)
	sessionID := refs[0].(map[string]any)["id"].(string)

	sessionResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID, nil, httptestx.WithCookies(adminLogin.SessionCookie))
	session := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	if session["source_file_kind"] != imports.SourceFileKindCSV || session["original_filename"] != "first.csv" || session["session_status"] != "discovered" {
		t.Fatalf("unexpected import session resource: %#v", session)
	}
	if session["parser_profile_id"] != imports.ParserProfileWorkbookImport || session["parser_version"] != imports.ParserVersionWorkbookImport {
		t.Fatalf("unexpected parser provenance: %#v", session)
	}

	unitsResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units?limit=1", nil, httptestx.WithCookies(adminLogin.SessionCookie))
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
	if _, ok := unit["source_stream_ref"]; ok {
		t.Fatalf("import unit resource must not expose source stream ref: %#v", unit)
	}
	var sourceStreamRef string
	var sourceContentSHA256 string
	var sourceBytes []byte
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT source_stream_ref, source_content_sha256, source_bytes
  FROM import_source_streams
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
`, sessionID, unitID).Scan(&sourceStreamRef, &sourceContentSHA256, &sourceBytes); err != nil {
		t.Fatalf("query private import source stream: %v", err)
	}
	if !strings.HasPrefix(sourceStreamRef, "impsrc_") || sourceContentSHA256 != session["source_content_sha256"] || !bytes.Equal(sourceBytes, []byte(csv)) {
		t.Fatalf("unexpected private source stream: ref=%q sha=%q bytes=%q session=%#v", sourceStreamRef, sourceContentSHA256, string(sourceBytes), session)
	}

	previewResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/preview", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	preview := httptestx.RequireSuccessEnvelope(t, previewResp, http.StatusOK)["data"].(map[string]any)
	if preview["truncated"] != false {
		t.Fatalf("unexpected preview truncation: %#v", preview)
	}
	rows := preview["preview_rows"].([]any)
	if len(rows) != 2 {
		t.Fatalf("expected two preview rows, got %#v", rows)
	}
}

func TestXLSXDiscoveryUsesBoundedUsedRange_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "extension_profile-import-xlsx-discovery")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-import-xlsx-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-XLSX",
		"title":         "Enterprise integration XLSX import",
	})
	incidentID := incident["incident_id"].(string)
	metadata := `{"client_txn_id":"txn-extension_profile-import-xlsx-upload","incident_id":"` + incidentID + `"}`

	uploadResp := postImportUploadBytes(t, harness.Server.HTTP.URL, adminLogin, metadata, multipleSheetXLSX(t), "input.xlsx", imports.MediaTypeXLSX, false)
	uploadJob := httptestx.RequireSuccessEnvelope(t, uploadResp, http.StatusAccepted)["data"].(map[string]any)
	job := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, uploadJob["job_id"].(string))
	if job["status"] != "succeeded" || job["result_summary"].(map[string]any)["code"] != "import_session_discovered" {
		t.Fatalf("unexpected discovery job: %#v", job)
	}
	sessionID := job["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)

	sessionResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID, nil, httptestx.WithCookies(adminLogin.SessionCookie))
	session := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	if session["source_file_kind"] != imports.SourceFileKindXLSX {
		t.Fatalf("unexpected XLSX session: %#v", session)
	}

	unitsResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	units := httptestx.RequireSuccessEnvelope(t, unitsResp, http.StatusOK)["data"].(map[string]any)["import_units"].([]any)
	if len(units) != 2 {
		t.Fatalf("expected two workbook units, got %#v", units)
	}
	unit := units[0].(map[string]any)
	if unit["locator_kind"] != "xlsx_used_range" || unit["inferred_row_count"] != float64(2) || unit["inferred_column_count"] != float64(2) {
		t.Fatalf("unexpected XLSX unit: %#v", unit)
	}
	if unit["locator"].(map[string]any)["sheet_name"] != "Sheet1" ||
		units[1].(map[string]any)["locator"].(map[string]any)["sheet_name"] != "Sheet2" {
		t.Fatalf("workbook units are not in sheet order: %#v", units)
	}

	previewResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units/"+unit["import_unit_id"].(string)+"/preview", nil, httptestx.WithCookies(adminLogin.SessionCookie))
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

func TestXLSXOperatorRegionCreatesDurableExactReplay_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "imports-xlsx-operator-region")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-imports-region-incident",
		"incident_key":  "IR-IMPORT-REGION",
		"title":         "Operator region import",
	})
	incidentID := incident["incident_id"].(string)
	metadata := `{"client_txn_id":"txn-imports-region-upload","incident_id":"` + incidentID + `"}`
	uploadResp := postImportUploadBytes(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		metadata,
		multipleSheetXLSX(t),
		"regions.xlsx",
		imports.MediaTypeXLSX,
		false,
	)
	uploadJob := httptestx.RequireSuccessEnvelope(t, uploadResp, http.StatusAccepted)["data"].(map[string]any)
	job := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, uploadJob["job_id"].(string))
	sessionID := job["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)
	unitsResp := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	units := httptestx.RequireSuccessEnvelope(t, unitsResp, http.StatusOK)["data"].(map[string]any)["import_units"].([]any)
	baseUnitID := units[0].(map[string]any)["import_unit_id"].(string)
	regionPath := "/api/v1/import-sessions/" + sessionID + "/units/" + baseUnitID + "/regions"
	request := map[string]any{
		"client_txn_id": "txn-imports-region-create",
		"source_rect": map[string]any{
			"start_row": 1, "start_column": 1, "end_row": 2, "end_column": 2,
		},
	}
	createResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, regionPath, request)
	created := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	if created["locator_kind"] != "operator_region" ||
		created["source_rect_a1"] != "A1:B2" ||
		created["unit_status"] != "discovered" {
		t.Fatalf("unexpected operator region: %#v", created)
	}
	regionUnitID := created["import_unit_id"].(string)

	replayResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, regionPath, request)
	replayed := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusCreated)["data"].(map[string]any)
	if replayed["import_unit_id"] != regionUnitID {
		t.Fatalf("exact replay created a different unit: first=%#v replay=%#v", created, replayed)
	}
	semanticReplay := map[string]any{
		"client_txn_id": "txn-imports-region-semantic-replay",
		"source_rect":   request["source_rect"],
	}
	semanticResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, regionPath, semanticReplay)
	semantic := httptestx.RequireSuccessEnvelope(t, semanticResp, http.StatusCreated)["data"].(map[string]any)
	if semantic["import_unit_id"] != regionUnitID {
		t.Fatalf("semantic replay created a duplicate unit: %#v", semantic)
	}
	divergent := map[string]any{
		"client_txn_id": "txn-imports-region-create",
		"source_rect": map[string]any{
			"start_row": 1, "start_column": 1, "end_row": 3, "end_column": 2,
		},
	}
	conflictResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, regionPath, divergent)
	httptestx.RequireErrorEnvelope(t, conflictResp, http.StatusConflict, "client_txn_conflict")
	outside := map[string]any{
		"client_txn_id": "txn-imports-region-outside",
		"source_rect": map[string]any{
			"start_row": 1, "start_column": 1, "end_row": 4, "end_column": 2,
		},
	}
	outsideResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, regionPath, outside)
	outsideError := httptestx.RequireErrorEnvelope(t, outsideResp, http.StatusBadRequest, "invalid_import_request")
	if outsideError["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "invalid_source_rect" {
		t.Fatalf("unexpected invalid rectangle error: %#v", outsideError)
	}

	var baseID string
	var regionSequence int
	var selected bool
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT base_import_unit_id::text,
       operator_region_sequence,
       import_unit_id = ANY(
           SELECT unnest(selected_unit_ids)
             FROM import_sessions
            WHERE import_session_id::text = $1
       )
  FROM import_units
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
`, sessionID, regionUnitID).Scan(&baseID, &regionSequence, &selected); err != nil {
		t.Fatalf("read durable operator region: %v", err)
	}
	if baseID != baseUnitID || regionSequence != 1 || selected {
		t.Fatalf("unexpected durable region binding: base=%q sequence=%d selected=%t", baseID, regionSequence, selected)
	}
	if count := dbassert.CountSQL(
		t,
		harness.DB,
		`SELECT COUNT(*) FROM import_source_streams WHERE import_unit_id::text = $1`,
		regionUnitID,
	); count != 1 {
		t.Fatalf("operator region must retain exactly one private source stream, got %d", count)
	}
}

func TestSelectionLifecycleEnforcesOverlapAndRetainsSkippedMapping_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "imports-selection-lifecycle")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-imports-selection-lifecycle-incident",
		"incident_key":  "IR-IMPORT-SELECTION",
		"title":         "Import selection lifecycle",
	})
	incidentID := incident["incident_id"].(string)
	metadata := `{"client_txn_id":"txn-imports-selection-lifecycle-upload","incident_id":"` + incidentID + `"}`
	uploadResp := postImportUploadBytes(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		metadata,
		multipleSheetXLSX(t),
		"selection.xlsx",
		imports.MediaTypeXLSX,
		false,
	)
	uploadJob := httptestx.RequireSuccessEnvelope(t, uploadResp, http.StatusAccepted)["data"].(map[string]any)
	discoveryJob := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, uploadJob["job_id"].(string))
	sessionID := discoveryJob["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)
	unitsResp := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	units := httptestx.RequireSuccessEnvelope(t, unitsResp, http.StatusOK)["data"].(map[string]any)["import_units"].([]any)
	if len(units) != 2 {
		t.Fatalf("selection workbook discovered %d units, want 2", len(units))
	}
	firstUnitID := units[0].(map[string]any)["import_unit_id"].(string)
	secondUnitID := units[1].(map[string]any)["import_unit_id"].(string)
	approveTimelineImportMapping(t, harness.Server.HTTP.URL, adminLogin, sessionID, firstUnitID, []string{"host", "summary"}, "selection-first")
	approveTimelineImportMapping(t, harness.Server.HTTP.URL, adminLogin, sessionID, secondUnitID, []string{"indicator", "type"}, "selection-second")

	var originalFingerprint string
	var originalMapping string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT mapping_fingerprint, approved_mapping_json::text
  FROM import_units
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
`, sessionID, firstUnitID).Scan(&originalFingerprint, &originalMapping); err != nil {
		t.Fatalf("read approved mapping before skip: %v", err)
	}
	firstSelect := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/units/"+firstUnitID+"/select",
		map[string]any{"client_txn_id": "txn-imports-selection-first-select"},
	)
	firstSelectBody := httptestx.RequireSuccessEnvelope(t, firstSelect, http.StatusOK)
	firstSelected := firstSelectBody["data"].(map[string]any)
	if firstSelected["session_status"] != "ready_to_apply" ||
		firstSelected["unit"].(map[string]any)["unit_status"] != "ready" {
		t.Fatalf("mapped unit was not selected as ready: %#v", firstSelected)
	}
	firstSelectReplay := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/units/"+firstUnitID+"/select",
		map[string]any{"client_txn_id": "txn-imports-selection-first-select"},
	)
	firstReplayBody := httptestx.RequireSuccessEnvelope(t, firstSelectReplay, http.StatusOK)
	if !reflect.DeepEqual(firstSelectBody["data"], firstReplayBody["data"]) {
		t.Fatalf("exact selection replay changed response: first=%#v replay=%#v", firstSelectBody, firstReplayBody)
	}
	skipResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/units/"+firstUnitID+"/skip",
		map[string]any{"client_txn_id": "txn-imports-selection-first-skip"},
	)
	skipped := httptestx.RequireSuccessEnvelope(t, skipResp, http.StatusOK)["data"].(map[string]any)
	if skipped["unit"].(map[string]any)["unit_status"] != "skipped" {
		t.Fatalf("unit was not skipped: %#v", skipped)
	}
	reselectResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/units/"+firstUnitID+"/select",
		map[string]any{"client_txn_id": "txn-imports-selection-first-reselect"},
	)
	reselected := httptestx.RequireSuccessEnvelope(t, reselectResp, http.StatusOK)["data"].(map[string]any)
	if reselected["unit"].(map[string]any)["unit_status"] != "ready" {
		t.Fatalf("skipped mapped unit was not reselected as ready: %#v", reselected)
	}
	var retainedFingerprint string
	var retainedMapping string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT mapping_fingerprint, approved_mapping_json::text
  FROM import_units
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
`, sessionID, firstUnitID).Scan(&retainedFingerprint, &retainedMapping); err != nil {
		t.Fatalf("read approved mapping after reselection: %v", err)
	}
	if retainedFingerprint != originalFingerprint || retainedMapping != originalMapping {
		t.Fatalf("reselection changed approved mapping: fingerprint %q/%q mapping %q/%q", originalFingerprint, retainedFingerprint, originalMapping, retainedMapping)
	}
	clearFirstResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/units/"+firstUnitID+"/skip",
		map[string]any{"client_txn_id": "txn-imports-selection-first-clear"},
	)
	httptestx.RequireSuccessEnvelope(t, clearFirstResp, http.StatusOK)

	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE import_units
   SET locator = '{"sheet_name":"Sheet1","rect_a1":"B2:C4"}',
       source_rect_a1 = 'B2:C4'
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
`, sessionID, secondUnitID); err != nil {
		t.Fatalf("make selection rectangles overlap: %v", err)
	}

	type concurrentSelectResult struct {
		unitID string
		status int
		body   []byte
		err    error
	}
	start := make(chan struct{})
	results := make(chan concurrentSelectResult, 2)
	for index, unitID := range []string{firstUnitID, secondUnitID} {
		requestBody, err := json.Marshal(map[string]any{
			"client_txn_id": "txn-imports-selection-concurrent-" + strconv.Itoa(index+1),
		})
		if err != nil {
			t.Fatalf("marshal concurrent select request: %v", err)
		}
		request, err := http.NewRequest(
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select",
			bytes.NewReader(requestBody),
		)
		if err != nil {
			t.Fatalf("create concurrent select request: %v", err)
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)
		request.AddCookie(adminLogin.SessionCookie)
		request.AddCookie(adminLogin.CSRFCookie)
		go func(selectedUnitID string, selectRequest *http.Request) {
			<-start
			response, requestErr := http.DefaultClient.Do(selectRequest)
			if requestErr != nil {
				results <- concurrentSelectResult{unitID: selectedUnitID, err: requestErr}
				return
			}
			defer response.Body.Close()
			responseBody, readErr := io.ReadAll(response.Body)
			results <- concurrentSelectResult{
				unitID: selectedUnitID,
				status: response.StatusCode,
				body:   responseBody,
				err:    readErr,
			}
		}(unitID, request)
	}
	close(start)

	var selectedUnitID string
	successCount := 0
	conflictCount := 0
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatalf("concurrent select %s failed: %v", result.unitID, result.err)
		}
		switch result.status {
		case http.StatusOK:
			successCount++
			selectedUnitID = result.unitID
		case http.StatusConflict:
			conflictCount++
			var envelope map[string]any
			if err := json.Unmarshal(result.body, &envelope); err != nil {
				t.Fatalf("decode overlap conflict: %v body=%s", err, string(result.body))
			}
			apiError := envelope["error"].(map[string]any)
			if apiError["code"] != "import_apply_blocked" ||
				apiError["details"].(map[string]any)["reason_code"] != "overlapping_units" {
				t.Fatalf("unexpected overlap conflict: %#v", envelope)
			}
		default:
			t.Fatalf("concurrent select %s status=%d body=%s", result.unitID, result.status, string(result.body))
		}
	}
	if successCount != 1 || conflictCount != 1 {
		t.Fatalf("concurrent overlapping selects success=%d conflict=%d, want one each", successCount, conflictCount)
	}
	var selectedCount int
	if err := harness.DB.QueryRowContext(
		context.Background(),
		`SELECT cardinality(selected_unit_ids) FROM import_sessions WHERE import_session_id::text = $1`,
		sessionID,
	).Scan(&selectedCount); err != nil {
		t.Fatalf("count persisted selection: %v", err)
	}
	if selectedCount != 1 {
		t.Fatalf("concurrent selection persisted %d units, want 1", selectedCount)
	}
	clearWinnerResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/units/"+selectedUnitID+"/skip",
		map[string]any{"client_txn_id": "txn-imports-selection-concurrent-clear"},
	)
	httptestx.RequireSuccessEnvelope(t, clearWinnerResp, http.StatusOK)

	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE import_units
   SET locator = '{"sheet_name":"Sheet2","rect_a1":"A1:B3"}',
       source_rect_a1 = 'A1:B3'
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
`, sessionID, secondUnitID); err != nil {
		t.Fatalf("restore non-overlapping rectangle: %v", err)
	}
	for index, unitID := range []string{firstUnitID, secondUnitID} {
		selectResp := doImportJSON(
			t,
			harness.Server.HTTP.URL,
			adminLogin,
			http.MethodPost,
			"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select",
			map[string]any{"client_txn_id": "txn-imports-selection-apply-setup-" + strconv.Itoa(index+1)},
		)
		httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE import_units
   SET locator = '{"sheet_name":"Sheet1","rect_a1":"B2:C4"}',
       source_rect_a1 = 'B2:C4'
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
`, sessionID, secondUnitID); err != nil {
		t.Fatalf("inject apply-time overlap: %v", err)
	}
	applyResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/apply",
		map[string]any{"client_txn_id": "txn-imports-selection-overlap-apply"},
	)
	applyError := httptestx.RequireErrorEnvelope(t, applyResp, http.StatusConflict, "import_apply_blocked")
	if applyError["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "overlapping_units" {
		t.Fatalf("unexpected apply-time overlap error: %#v", applyError)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM import_apply_unit_plans WHERE import_session_id::text = $1`, sessionID); got != 0 {
		t.Fatalf("apply-time overlap created %d durable plans", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM import_unit_apply_outcomes WHERE import_session_id::text = $1`, sessionID); got != 0 {
		t.Fatalf("apply-time overlap created %d durable outcomes", got)
	}
}

func TestFreshSessionIsTheOnlyExplicitReimportWorkflow_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "imports-fresh-session-reimport")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-imports-fresh-session-incident",
		"incident_key":  "IR-IMPORT-FRESH-SESSION",
		"title":         "Fresh-session import",
	})
	incidentID := incident["incident_id"].(string)
	const source = "summary\nsame source content\n"

	sessionIDs := make([]string, 0, 2)
	for index := 1; index <= 2; index++ {
		prefix := "fresh-session-" + strconv.Itoa(index)
		sessionID, unitID := startCSVImportSession(
			t,
			harness.Server.HTTP.URL,
			adminLogin,
			incidentID,
			"txn-"+prefix+"-upload",
			source,
			"same-source.csv",
		)
		sessionIDs = append(sessionIDs, sessionID)
		approveTimelineImportMapping(t, harness.Server.HTTP.URL, adminLogin, sessionID, unitID, []string{"summary"}, prefix)
		selectResp := doImportJSON(
			t,
			harness.Server.HTTP.URL,
			adminLogin,
			http.MethodPost,
			"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select",
			map[string]any{"client_txn_id": "txn-" + prefix + "-select"},
		)
		httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)
		applyResp := doImportJSON(
			t,
			harness.Server.HTTP.URL,
			adminLogin,
			http.MethodPost,
			"/api/v1/import-sessions/"+sessionID+"/apply",
			map[string]any{"client_txn_id": "txn-" + prefix + "-apply"},
		)
		applyJob := httptestx.RequireSuccessEnvelope(t, applyResp, http.StatusAccepted)["data"].(map[string]any)
		job := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, applyJob["job_id"].(string))
		if job["status"] != "succeeded" {
			t.Fatalf("fresh-session apply %d did not succeed: %#v", index, job)
		}
	}
	if sessionIDs[0] == sessionIDs[1] {
		t.Fatalf("different upload transactions reused one import session: %#v", sessionIDs)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM timeline_events WHERE incident_id::text = $1`, incidentID); got != 2 {
		t.Fatalf("fresh-session re-import created %d Timeline rows, want 2", got)
	}

	duplicateApply := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionIDs[0]+"/apply",
		map[string]any{"client_txn_id": "txn-fresh-session-same-session-duplicate"},
	)
	duplicateError := httptestx.RequireErrorEnvelope(t, duplicateApply, http.StatusConflict, "import_apply_blocked")
	if duplicateError["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "duplicate_apply_blocked" {
		t.Fatalf("unexpected same-session duplicate error: %#v", duplicateError)
	}
	unsupportedMode := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionIDs[1]+"/apply",
		map[string]any{
			"client_txn_id": "txn-fresh-session-unsupported-mode",
			"reimport":      true,
		},
	)
	modeError := httptestx.RequireErrorEnvelope(t, unsupportedMode, http.StatusBadRequest, "invalid_import_request")
	if modeError["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "unknown_field" {
		t.Fatalf("unexpected unsupported re-import mode error: %#v", modeError)
	}
}

func TestMappingSelectApplyCreatesTimelineRows_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "extension_profile-import-apply")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-import-apply-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-APPLY",
		"title":         "Enterprise integration import apply",
	})
	incidentID := incident["incident_id"].(string)
	hostRecordID := createImportAutoResolutionCandidateHost(t, harness.Server.HTTP.URL, adminLogin, incidentID)
	identityRecordID := createImportAutoResolutionCandidateIdentity(t, harness.Server.HTTP.URL, adminLogin, incidentID)
	metadata := `{"client_txn_id":"txn-extension_profile-import-apply-upload","incident_id":"` + incidentID + `"}`
	csv := "host,identity,summary,source_note,tag\nhost-1,identity-1, Alpha summary ,raw-a,Urgent\nhost-2,identity-2,Beta summary,raw-b,Review\n"

	uploadResp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, csv, "apply.csv", false)
	uploadJob := httptestx.RequireSuccessEnvelope(t, uploadResp, http.StatusAccepted)["data"].(map[string]any)
	job := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, uploadJob["job_id"].(string))
	sessionID := job["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)

	unitsResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	units := httptestx.RequireSuccessEnvelope(t, unitsResp, http.StatusOK)["data"].(map[string]any)["import_units"].([]any)
	unitID := units[0].(map[string]any)["import_unit_id"].(string)

	mapping := map[string]any{
		"client_txn_id":         "txn-extension_profile-import-apply-mapping",
		"target_view_schema_id": "cartulary.view.timeline.v2",
		"header_row_ref":        1,
		"data_start_row_ref":    2,
		"unknown_column_policy": "preserve_raw_capture",
		"source_columns": []map[string]any{
			{
				"source_column_ordinal": 1,
				"source_header_text":    "host",
				"field_key":             "timeline.host_refs",
				"entity_binding_mode":   "mention_origin",
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
			{
				"source_column_ordinal": 2,
				"source_header_text":    "identity",
				"field_key":             "timeline.identity_refs",
				"entity_binding_mode":   "mention_origin",
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
			{
				"source_column_ordinal": 3,
				"source_header_text":    "summary",
				"field_key":             "timeline.activity_synopsis_text",
				"entity_binding_mode":   nil,
				"transform_id":          "trim_v1",
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
			{
				"source_column_ordinal": 4,
				"source_header_text":    "source_note",
				"field_key":             nil,
				"entity_binding_mode":   nil,
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
			{
				"source_column_ordinal": 5,
				"source_header_text":    "tag",
				"field_key":             "timeline.tags",
				"entity_binding_mode":   nil,
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
		},
	}
	customAttrsMapping := cloneImportMappingPayload(t, mapping)
	customAttrsMapping["client_txn_id"] = "txn-extension_profile-import-apply-custom-attrs"
	customAttrsMapping["unknown_column_policy"] = "preserve_custom_attrs"
	customAttrsResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPut, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping", customAttrsMapping)
	customAttrsErr := httptestx.RequireErrorEnvelope(t, customAttrsResp, http.StatusBadRequest, "invalid_import_request")
	if customAttrsErr["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "unknown_column_policy_not_supported_for_target" {
		t.Fatalf("unexpected preserve_custom_attrs rejection: %#v", customAttrsErr)
	}

	rejectUnmappedMapping := cloneImportMappingPayload(t, mapping)
	rejectUnmappedMapping["client_txn_id"] = "txn-extension_profile-import-apply-reject-unmapped"
	rejectUnmappedMapping["unknown_column_policy"] = "reject_if_unmapped"
	rejectUnmappedResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPut, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping", rejectUnmappedMapping)
	rejectUnmappedErr := httptestx.RequireErrorEnvelope(t, rejectUnmappedResp, http.StatusBadRequest, "invalid_import_request")
	if rejectUnmappedErr["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "invalid_source_columns" {
		t.Fatalf("unexpected reject_if_unmapped rejection: %#v", rejectUnmappedErr)
	}

	mappingResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPut, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping", mapping)
	mappedUnit := httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)["data"].(map[string]any)
	if mappedUnit["unit_status"] != "mapped" || mappedUnit["mapping_fingerprint"] == nil || mappedUnit["approved_mapping"] == nil {
		t.Fatalf("unexpected mapped unit: %#v", mappedUnit)
	}

	selectResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select", map[string]any{"client_txn_id": "txn-extension_profile-import-apply-select"})
	selected := httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)["data"].(map[string]any)
	if selected["session_status"] != "ready_to_apply" || selected["unit"].(map[string]any)["unit_status"] != "ready" {
		t.Fatalf("unexpected selected state: %#v", selected)
	}

	applyResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-extension_profile-import-apply-apply"})
	applyJob := httptestx.RequireSuccessEnvelope(t, applyResp, http.StatusAccepted)["data"].(map[string]any)
	appliedJob := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, applyJob["job_id"].(string))
	if appliedJob["status"] != "succeeded" || appliedJob["result_summary"].(map[string]any)["code"] != "import_session_applied" {
		t.Fatalf("unexpected apply job: %#v", appliedJob)
	}
	requireImportProof(t, harness.DB, applyJob["job_id"].(string), "import.apply")

	queryResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.timeline.v2/query", map[string]any{}, httptestx.WithCookies(adminLogin.SessionCookie))
	rows := httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)["data"].(map[string]any)["rows"].([]any)
	if len(rows) != 2 {
		t.Fatalf("expected two imported timeline rows, got %#v", rows)
	}
	summaries := map[string]bool{}
	recordIDsBySummary := map[string]string{}
	rowsBySummary := map[string]map[string]any{}
	for _, rowAny := range rows {
		row := rowAny.(map[string]any)
		cells := row["cells"].(map[string]any)
		summary := cells["timeline.activity_synopsis_text"].(map[string]any)["value"].(string)
		summaries[summary] = true
		recordIDsBySummary[summary] = row["record_id"].(string)
		rowsBySummary[summary] = row
	}
	if !summaries["Alpha summary"] || !summaries["Beta summary"] {
		t.Fatalf("imported summaries not found: %#v", summaries)
	}
	requireTimelineMentionUnresolved(t, rowsBySummary["Alpha summary"], "timeline.host_refs", "host-1")
	requireTimelineMentionUnresolved(t, rowsBySummary["Alpha summary"], "timeline.identity_refs", "identity-1")
	requireTimelineMentionUnresolved(t, rowsBySummary["Beta summary"], "timeline.host_refs", "host-2")
	requireTimelineMentionUnresolved(t, rowsBySummary["Beta summary"], "timeline.identity_refs", "identity-2")
	requireImportedMentionNotAutoResolved(t, harness.DB, incidentID, recordIDsBySummary["Alpha summary"], hostRecordID, "timeline.host_refs", "observed_on_host", "host-1")
	requireImportedMentionNotAutoResolved(t, harness.DB, incidentID, recordIDsBySummary["Alpha summary"], identityRecordID, "timeline.identity_refs", "observed_as_identity", "identity-1")
	requireTimelineImportProvenance(t, harness.DB, recordIDsBySummary["Alpha summary"], sessionID, unitID, "source_note", "raw-a", 2, 4, "A1:E3")
	requireTimelineImportProvenance(t, harness.DB, recordIDsBySummary["Beta summary"], sessionID, unitID, "source_note", "raw-b", 3, 4, "A1:E3")

	unitClientTxnID := "import:" + sessionID + ":" + unitID + ":txn-extension_profile-import-apply-apply"
	var changeSetID string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT change_set_id::text
  FROM change_sets
 WHERE source = 'imports.apply'
   AND client_txn_id = $1
`, unitClientTxnID).Scan(&changeSetID); err != nil {
		t.Fatalf("query Timeline unit change set: %v", err)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM import_apply_unit_plans p
  JOIN import_unit_apply_outcomes o
    ON o.import_session_id = p.import_session_id
   AND o.import_unit_id = p.import_unit_id
   AND o.apply_job_id = p.apply_job_id
 WHERE p.import_session_id::text = $1
   AND p.import_unit_id::text = $2
   AND p.apply_job_id::text = $3
   AND p.source_content_sha256 = o.source_content_sha256
   AND p.mapping_fingerprint = o.mapping_fingerprint
   AND o.outcome_status = 'applied'
   AND o.target_kind = 'view_schema'
   AND o.target_view_schema_id = 'cartulary.view.timeline.v2'
   AND o.owner_binding_id = 'timeline.import_create'
   AND o.change_set_id::text = $4
   AND jsonb_typeof(o.owner_result_json) = 'array'
`, sessionID, unitID, applyJob["job_id"].(string), changeSetID); got != 1 {
		t.Fatalf("expected one immutable Timeline apply plan/outcome pair, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations
 WHERE change_set_id::text = $1
`, changeSetID); got != 4 {
		t.Fatalf("expected two Timeline row and two tag mutations, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations
 WHERE change_set_id::text = $1
   AND sequence_no BETWEEN 1 AND 4
`, changeSetID); got != 4 {
		t.Fatalf("expected contiguous Timeline mutation order, got %d rows", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM record_revisions
 WHERE change_set_id::text = $1
`, changeSetID); got != 2 {
		t.Fatalf("expected two Timeline revisions in the unit change set, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM timeline_grid_projection
 WHERE incident_id::text = $1
`, incidentID); got != 2 {
		t.Fatalf("expected two Timeline projection rows, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM import_apply_journal
 WHERE import_session_id::text = $1
   AND change_set_id::text = $2
`, sessionID, changeSetID); got != 2 {
		t.Fatalf("expected two Timeline journal rows in the unit change set, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM collaboration_event_intents
 WHERE source_change_set_id::text = $1
   AND event_family = 'record_changed'
`, changeSetID); got != 2 {
		t.Fatalf("expected two Timeline record-change intents, got %d", got)
	}

	replayApply := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-extension_profile-import-apply-apply"})
	replayedJob := httptestx.RequireSuccessEnvelope(t, replayApply, http.StatusAccepted)["data"].(map[string]any)
	if replayedJob["job_id"] != applyJob["job_id"] {
		t.Fatalf("exact Timeline apply replay returned a different job: first=%#v replay=%#v", applyJob, replayedJob)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM timeline_events
 WHERE incident_id::text = $1
`, incidentID); got != 2 {
		t.Fatalf("exact Timeline replay duplicated owner effects: %d rows", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM import_unit_apply_outcomes
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
`, sessionID, unitID); got != 1 {
		t.Fatalf("exact Timeline replay changed immutable unit outcomes: %d rows", got)
	}

	duplicateApply := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-extension_profile-import-apply-second"})
	errBody := httptestx.RequireErrorEnvelope(t, duplicateApply, http.StatusConflict, "import_apply_blocked")
	if errBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "duplicate_apply_blocked" {
		t.Fatalf("unexpected duplicate apply error: %#v", errBody)
	}
}

func TestTimelineOwnerUnitRollsBackWhenJournalFails_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "extension_profile-import-timeline-owner-rollback")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-import-timeline-rollback-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-TIMELINE-ROLLBACK",
		"title":         "Timeline import owner rollback",
	})
	incidentID := incident["incident_id"].(string)
	createImportAutoResolutionCandidateHost(t, harness.Server.HTTP.URL, adminLogin, incidentID)
	createImportAutoResolutionCandidateIdentity(t, harness.Server.HTTP.URL, adminLogin, incidentID)
	sessionID, unitID := startCSVImportSession(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		incidentID,
		"txn-extension_profile-import-timeline-rollback-upload",
		"host,identity,summary\nhost-1,identity-1,must roll back\n",
		"timeline-rollback.csv",
	)
	mapping := map[string]any{
		"client_txn_id":         "txn-extension_profile-import-timeline-rollback-mapping",
		"target_view_schema_id": "cartulary.view.timeline.v2",
		"header_row_ref":        1,
		"data_start_row_ref":    2,
		"unknown_column_policy": "preserve_raw_capture",
		"source_columns": []map[string]any{
			{
				"source_column_ordinal": 1,
				"source_header_text":    "host",
				"field_key":             "timeline.host_refs",
				"entity_binding_mode":   "mention_origin",
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
			{
				"source_column_ordinal": 2,
				"source_header_text":    "identity",
				"field_key":             "timeline.identity_refs",
				"entity_binding_mode":   "mention_origin",
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
			{
				"source_column_ordinal": 3,
				"source_header_text":    "summary",
				"field_key":             "timeline.activity_synopsis_text",
				"entity_binding_mode":   nil,
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
		},
	}
	mappingResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPut,
		"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping",
		mapping,
	)
	httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)
	selectResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select",
		map[string]any{"client_txn_id": "txn-extension_profile-import-timeline-rollback-select"},
	)
	httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)

	if _, err := harness.DB.ExecContext(
		context.Background(),
		`
CREATE FUNCTION public.fail_import_journal_timeline_rs04()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'injected Timeline journal failure';
END;
$$
`,
	); err != nil {
		t.Fatalf("create Timeline journal failure function: %v", err)
	}
	if _, err := harness.DB.ExecContext(
		context.Background(),
		`
CREATE TRIGGER fail_import_journal_timeline_rs04
BEFORE INSERT ON import_apply_journal
FOR EACH ROW
EXECUTE FUNCTION public.fail_import_journal_timeline_rs04()
`,
	); err != nil {
		t.Fatalf("create Timeline journal failure trigger: %v", err)
	}
	t.Cleanup(func() {
		_, _ = harness.DB.ExecContext(
			context.Background(),
			"DROP TRIGGER IF EXISTS fail_import_journal_timeline_rs04 ON import_apply_journal",
		)
		_, _ = harness.DB.ExecContext(
			context.Background(),
			"DROP FUNCTION IF EXISTS public.fail_import_journal_timeline_rs04()",
		)
	})
	applyResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/apply",
		map[string]any{"client_txn_id": "txn-extension_profile-import-timeline-rollback-apply"},
	)
	applyJob := httptestx.RequireSuccessEnvelope(t, applyResp, http.StatusAccepted)["data"].(map[string]any)
	failedJob := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, applyJob["job_id"].(string))
	if failedJob["status"] != "failed" {
		t.Fatalf("expected failed Timeline apply job, got %#v", failedJob)
	}
	for table, query := range map[string]string{
		"Timeline source":     `SELECT COUNT(*) FROM timeline_events WHERE incident_id::text = $1`,
		"record envelope":     `SELECT COUNT(*) FROM records WHERE incident_id::text = $1 AND record_type = 'timeline_event'`,
		"entity mention":      `SELECT COUNT(*) FROM entity_mentions mention JOIN records record ON record.record_id = mention.source_record_id WHERE record.incident_id::text = $1 AND record.record_type = 'timeline_event'`,
		"record link":         `SELECT COUNT(*) FROM record_links link JOIN records record ON record.record_id = link.src_record_id WHERE record.incident_id::text = $1 AND record.record_type = 'timeline_event'`,
		"Timeline projection": `SELECT COUNT(*) FROM timeline_grid_projection WHERE incident_id::text = $1`,
		"change-set mutation": `SELECT COUNT(*) FROM change_set_mutations mutation JOIN change_sets change_set USING (change_set_id) WHERE change_set.incident_id::text = $1 AND change_set.source = 'imports.apply'`,
		"record revision":     `SELECT COUNT(*) FROM record_revisions revision JOIN change_sets change_set USING (change_set_id) WHERE change_set.incident_id::text = $1 AND change_set.source = 'imports.apply'`,
		"apply journal":       `SELECT COUNT(*) FROM import_apply_journal WHERE import_session_id::text = $1`,
		"unit change set":     `SELECT COUNT(*) FROM change_sets WHERE incident_id::text = $1 AND source = 'imports.apply' AND client_txn_id = $2`,
	} {
		args := []any{incidentID}
		if table == "apply journal" {
			args = []any{sessionID}
		}
		if table == "unit change set" {
			args = []any{
				incidentID,
				"import:" + sessionID + ":" + unitID + ":txn-extension_profile-import-timeline-rollback-apply",
			}
		}
		if got := dbassert.CountSQL(t, harness.DB, query, args...); got != 0 {
			t.Fatalf("%s survived failed Timeline unit transaction: %d", table, got)
		}
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM collaboration_event_intents intent
  JOIN change_sets change_set
    ON change_set.change_set_id = intent.source_change_set_id
 WHERE change_set.incident_id::text = $1
   AND change_set.source = 'imports.apply'
   AND change_set.client_txn_id = $2
   AND intent.event_family = 'record_changed'
`, incidentID, "import:"+sessionID+":"+unitID+":txn-extension_profile-import-timeline-rollback-apply"); got != 0 {
		t.Fatalf("failed Timeline owner transaction retained %d record-change intents", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM import_unit_apply_outcomes
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
   AND apply_job_id::text = $3
   AND outcome_status = 'failed'
   AND change_set_id IS NULL
   AND error_code = 'import_apply_blocked'
   AND reason_code = 'owner_apply_validation_failed'
   AND error_retryable = false
   AND error_details_json = '{"reason_code":"owner_apply_validation_failed"}'::jsonb
`, sessionID, unitID, applyJob["job_id"].(string)); got != 1 {
		t.Fatalf("failed Timeline unit did not persist one truthful terminal outcome: %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM import_sessions s
  JOIN import_units u USING (import_session_id)
 WHERE s.import_session_id::text = $1
   AND s.session_status = 'failed'
   AND u.import_unit_id::text = $2
   AND u.unit_status = 'failed'
`, sessionID, unitID); got != 1 {
		t.Fatalf("failed Timeline finalization did not derive terminal state from its outcome")
	}
}

func TestApplyRevalidatesTransactionCurrentState_Integration(t *testing.T) {
	testCases := []struct {
		name       string
		mutateSQL  string
		mutateArgs string
		errorCode  string
		reasonCode string
	}{
		{
			name: "role revoked",
			mutateSQL: `
UPDATE incident_memberships
   SET role = 'viewer',
       membership_version = membership_version + 1,
       updated_at = now()
 WHERE incident_id::text = $1
   AND user_id::text = $2
`,
			errorCode:  "authorization_denied",
			reasonCode: "authorization_changed",
			mutateArgs: "incident_actor",
		},
		{
			name: "membership removed",
			mutateSQL: `
DELETE FROM incident_memberships
 WHERE incident_id::text = $1
   AND user_id::text = $2
`,
			errorCode:  "authorization_denied",
			reasonCode: "authorization_changed",
			mutateArgs: "incident_actor",
		},
		{
			name: "incident closed",
			mutateSQL: `
UPDATE incidents
   SET status = 'closed',
       closed_at = now(),
       incident_version = incident_version + 1,
       updated_at = now()
 WHERE id::text = $1
`,
			errorCode:  "incident_closed",
			reasonCode: "incident_closed",
			mutateArgs: "incident",
		},
		{
			name: "actor deactivated",
			mutateSQL: `
UPDATE users
   SET is_active = false,
       user_version = user_version + 1,
       updated_at = now()
 WHERE id::text = $1
`,
			errorCode:  "authorization_denied",
			reasonCode: "authorization_changed",
			mutateArgs: "actor",
		},
		{
			name: "source rows changed",
			mutateSQL: `
UPDATE import_units
   SET source_rows_json = '[]'::jsonb,
       updated_at = now()
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
`,
			errorCode:  "import_apply_blocked",
			reasonCode: "source_changed",
			mutateArgs: "unit",
		},
		{
			name: "approved mapping changed",
			mutateSQL: `
UPDATE import_units
   SET approved_mapping_json =
           jsonb_set(approved_mapping_json, '{unknown_column_policy}', '"reject_if_unmapped"'),
       updated_at = now()
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
`,
			errorCode:  "import_apply_blocked",
			reasonCode: "source_changed",
			mutateArgs: "unit",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			runtime := appsupport.StartRuntime(t)
			harness := runtime.StartDefaultServer(t, "imports-transaction-current-"+strings.ReplaceAll(testCase.name, " ", "-"))
			adminLogin, adminUserID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
			incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
				"client_txn_id": "txn-imports-current-incident-" + strings.ReplaceAll(testCase.name, " ", "-"),
				"incident_key":  "IR-IMPORT-CURRENT-" + strings.ToUpper(strings.ReplaceAll(testCase.name, " ", "-")),
				"title":         "Import transaction-current " + testCase.name,
			})
			incidentID := incident["incident_id"].(string)
			sessionID, unitID := startCSVImportSession(
				t,
				harness.Server.HTTP.URL,
				adminLogin,
				incidentID,
				"txn-imports-current-upload-"+strings.ReplaceAll(testCase.name, " ", "-"),
				"summary\nmust not apply\n",
				"current.csv",
			)
			mappingResp := doImportJSON(
				t,
				harness.Server.HTTP.URL,
				adminLogin,
				http.MethodPut,
				"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping",
				map[string]any{
					"client_txn_id":         "txn-imports-current-mapping-" + strings.ReplaceAll(testCase.name, " ", "-"),
					"target_view_schema_id": "cartulary.view.timeline.v2",
					"header_row_ref":        1,
					"data_start_row_ref":    2,
					"unknown_column_policy": "preserve_raw_capture",
					"source_columns": []map[string]any{{
						"source_column_ordinal": 1,
						"source_header_text":    "summary",
						"field_key":             "timeline.activity_synopsis_text",
						"entity_binding_mode":   nil,
						"transform_id":          nil,
						"transform_options":     map[string]any{},
						"empty_value_policy":    "omit_field",
					}},
				},
			)
			httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)
			selectResp := doImportJSON(
				t,
				harness.Server.HTTP.URL,
				adminLogin,
				http.MethodPost,
				"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select",
				map[string]any{
					"client_txn_id": "txn-imports-current-select-" + strings.ReplaceAll(testCase.name, " ", "-"),
				},
			)
			httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)

			const advisoryKey int64 = 49006001
			blocker, err := harness.DB.Conn(context.Background())
			if err != nil {
				t.Fatalf("acquire apply-start blocker connection: %v", err)
			}
			defer blocker.Close()
			if _, err := blocker.ExecContext(
				context.Background(),
				"SELECT pg_advisory_lock($1)",
				advisoryKey,
			); err != nil {
				t.Fatalf("acquire apply-start advisory lock: %v", err)
			}
			defer func() {
				_, _ = blocker.ExecContext(
					context.Background(),
					"SELECT pg_advisory_unlock($1)",
					advisoryKey,
				)
			}()
			if _, err := harness.DB.ExecContext(context.Background(), `
CREATE FUNCTION public.block_import_apply_start_rs06()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.status = 'queued' AND NEW.status = 'running' THEN
        PERFORM pg_advisory_xact_lock(49006001);
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER block_import_apply_start_rs06
BEFORE UPDATE ON jobs
FOR EACH ROW
EXECUTE FUNCTION public.block_import_apply_start_rs06()
`); err != nil {
				t.Fatalf("install apply-start serialization fixture: %v", err)
			}

			type applyResponse struct {
				response *http.Response
				err      error
			}
			responseChannel := make(chan applyResponse, 1)
			go func() {
				request, err := http.NewRequest(
					http.MethodPost,
					harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/apply",
					bytes.NewBufferString(`{"client_txn_id":"txn-imports-current-apply-`+
						strings.ReplaceAll(testCase.name, " ", "-")+`"}`),
				)
				if err != nil {
					responseChannel <- applyResponse{err: err}
					return
				}
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)
				request.AddCookie(adminLogin.SessionCookie)
				request.AddCookie(adminLogin.CSRFCookie)
				response, err := http.DefaultClient.Do(request)
				responseChannel <- applyResponse{response: response, err: err}
			}()

			var applyJobID string
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				err = harness.DB.QueryRowContext(context.Background(), `
SELECT apply_job_id::text
  FROM import_sessions
 WHERE import_session_id::text = $1
   AND session_status = 'applying'
   AND apply_job_id IS NOT NULL
`, sessionID).Scan(&applyJobID)
				if err == nil {
					break
				}
				if err != sql.ErrNoRows {
					t.Fatalf("observe admitted apply job: %v", err)
				}
				time.Sleep(10 * time.Millisecond)
			}
			if applyJobID == "" {
				t.Fatal("apply job did not reach the serialized post-admission boundary")
			}
			var mutationArgs []any
			switch testCase.mutateArgs {
			case "incident_actor":
				mutationArgs = []any{incidentID, adminUserID}
			case "incident":
				mutationArgs = []any{incidentID}
			case "actor":
				mutationArgs = []any{adminUserID}
			case "unit":
				mutationArgs = []any{sessionID, unitID}
			default:
				t.Fatalf("unsupported transaction-current mutation argument kind %q", testCase.mutateArgs)
			}
			if _, err := harness.DB.ExecContext(
				context.Background(),
				testCase.mutateSQL,
				mutationArgs...,
			); err != nil {
				t.Fatalf("apply transaction-current mutation: %v", err)
			}
			if _, err := blocker.ExecContext(
				context.Background(),
				"SELECT pg_advisory_unlock($1)",
				advisoryKey,
			); err != nil {
				t.Fatalf("release apply-start advisory lock: %v", err)
			}

			select {
			case result := <-responseChannel:
				if result.err != nil {
					t.Fatalf("submit serialized apply: %v", result.err)
				}
				httptestx.RequireSuccessEnvelope(t, result.response, http.StatusAccepted)
			case <-time.After(10 * time.Second):
				t.Fatal("timed out waiting for serialized apply")
			}
			var jobStatus string
			terminalDeadline := time.Now().Add(5 * time.Second)
			for {
				if err := harness.DB.QueryRowContext(
					context.Background(),
					"SELECT status FROM jobs WHERE job_id::text = $1",
					applyJobID,
				).Scan(&jobStatus); err != nil {
					t.Fatalf("read terminal apply job after %s: %v", testCase.name, err)
				}
				if jobStatus == "failed" || time.Now().After(terminalDeadline) {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			if jobStatus != "failed" {
				t.Fatalf("expected failed apply after %s, got %q", testCase.name, jobStatus)
			}
			if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM import_unit_apply_outcomes
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
   AND apply_job_id::text = $3
   AND outcome_status = 'failed'
   AND error_code = $4
   AND reason_code = $5
`, sessionID, unitID, applyJobID, testCase.errorCode, testCase.reasonCode); got != 1 {
				t.Fatalf("expected one safe %s outcome, got %d", testCase.name, got)
			}
			if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM timeline_events
 WHERE incident_id::text = $1
`, incidentID); got != 0 {
				t.Fatalf("%s allowed %d owner effects", testCase.name, got)
			}
		})
	}
}

func TestUnitCommitCrashRecoveryDoesNotReplayOwnerEffects_Integration(t *testing.T) {
	testCases := []struct {
		name                  string
		triggerTable          string
		triggerFunctionBody   string
		preRecoveryOwnerCount int
		preRecoveryOutcome    int
		preRecoveryUnitStatus string
	}{
		{
			name:         "crash before unit commit",
			triggerTable: "import_unit_apply_outcomes",
			triggerFunctionBody: `
BEGIN
    RAISE EXCEPTION 'injected unit-outcome commit failure';
END;
`,
			preRecoveryOwnerCount: 0,
			preRecoveryOutcome:    0,
			preRecoveryUnitStatus: "applying",
		},
		{
			name:         "crash after unit commit",
			triggerTable: "jobs",
			triggerFunctionBody: `
BEGIN
    IF NEW.status IN ('succeeded', 'failed', 'canceled') THEN
        RAISE EXCEPTION 'injected finalizer commit failure';
    END IF;
    RETURN NEW;
END;
`,
			preRecoveryOwnerCount: 1,
			preRecoveryOutcome:    1,
			preRecoveryUnitStatus: "applied",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			slug := strings.ReplaceAll(testCase.name, " ", "-")
			runtime := appsupport.StartRuntime(t)
			database := runtime.PrepareIsolatedDatabase(t, "imports-recovery-"+slug)
			durableObjects, err := objectstore.NewFilesystemStore(t.TempDir())
			if err != nil {
				t.Fatalf("create durable import recovery object store: %v", err)
			}
			first := runtime.StartServerWithDatabaseAndObjectStore(
				t,
				"imports-recovery-first-"+slug,
				database,
				durableObjects,
			)
			adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, first.Server.HTTP.URL)
			incident := scenariotest.CreateIncident(t, first.Server, adminLogin, map[string]any{
				"client_txn_id": "txn-imports-recovery-incident-" + slug,
				"incident_key":  "IR-IMPORT-RECOVERY-" + strings.ToUpper(slug),
				"title":         "Import recovery " + testCase.name,
			})
			incidentID := incident["incident_id"].(string)
			sessionID, unitID := prepareReadyTimelineImport(
				t,
				first,
				adminLogin,
				incidentID,
				"imports-recovery-"+slug,
			)

			if _, err := first.DB.ExecContext(context.Background(), `
CREATE FUNCTION public.fail_import_apply_boundary_rs06()
RETURNS trigger
LANGUAGE plpgsql
AS $$
`+testCase.triggerFunctionBody+`
$$
`); err != nil {
				t.Fatalf("create import recovery failure function: %v", err)
			}
			if _, err := first.DB.ExecContext(
				context.Background(),
				`
CREATE TRIGGER fail_import_apply_boundary_rs06
BEFORE INSERT OR UPDATE ON `+testCase.triggerTable+`
FOR EACH ROW
EXECUTE FUNCTION public.fail_import_apply_boundary_rs06()
`,
			); err != nil {
				t.Fatalf("create import recovery failure trigger: %v", err)
			}

			applyResp := doImportJSON(
				t,
				first.Server.HTTP.URL,
				adminLogin,
				http.MethodPost,
				"/api/v1/import-sessions/"+sessionID+"/apply",
				map[string]any{"client_txn_id": "txn-imports-recovery-apply-" + slug},
			)
			applyJob := httptestx.RequireSuccessEnvelope(t, applyResp, http.StatusAccepted)["data"].(map[string]any)
			applyJobID := applyJob["job_id"].(string)
			var nextAttemptAt *time.Time
			failureDeadline := time.Now().Add(5 * time.Second)
			for {
				var failureCount int
				var attemptID *string
				err := first.DB.QueryRowContext(context.Background(), `
SELECT handler_failure_count, handler_attempt_id::text, handler_next_attempt_at
  FROM jobs
 WHERE job_id::text = $1
`, applyJobID).Scan(&failureCount, &attemptID, &nextAttemptAt)
				if err != nil {
					t.Fatalf("read interrupted apply attempt: %v", err)
				}
				if failureCount == 1 && attemptID == nil && nextAttemptAt != nil {
					break
				}
				if time.Now().After(failureDeadline) {
					t.Fatalf("timed out waiting for interrupted attempt: failure_count=%d attempt_id=%v", failureCount, attemptID)
				}
				time.Sleep(10 * time.Millisecond)
			}
			if got := dbassert.CountSQL(t, first.DB, `
SELECT COUNT(*)
  FROM timeline_events
 WHERE incident_id::text = $1
`, incidentID); got != testCase.preRecoveryOwnerCount {
				t.Fatalf("pre-recovery owner count = %d, want %d", got, testCase.preRecoveryOwnerCount)
			}
			if got := dbassert.CountSQL(t, first.DB, `
SELECT COUNT(*)
  FROM import_unit_apply_outcomes
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
`, sessionID, unitID); got != testCase.preRecoveryOutcome {
				t.Fatalf("pre-recovery outcome count = %d, want %d", got, testCase.preRecoveryOutcome)
			}
			var interruptedSessionStatus string
			var interruptedUnitStatus string
			var interruptedJobStatus string
			if err := first.DB.QueryRowContext(context.Background(), `
SELECT s.session_status, u.unit_status, j.status
  FROM import_sessions s
  JOIN import_units u USING (import_session_id)
  JOIN jobs j ON j.job_id = s.apply_job_id
 WHERE s.import_session_id::text = $1
   AND u.import_unit_id::text = $2
`, sessionID, unitID).Scan(
				&interruptedSessionStatus,
				&interruptedUnitStatus,
				&interruptedJobStatus,
			); err != nil {
				t.Fatalf("read interrupted apply state: %v", err)
			}
			if interruptedSessionStatus != "applying" ||
				interruptedUnitStatus != testCase.preRecoveryUnitStatus ||
				interruptedJobStatus != "running" {
				t.Fatalf(
					"unexpected interrupted state: session=%q unit=%q job=%q",
					interruptedSessionStatus,
					interruptedUnitStatus,
					interruptedJobStatus,
				)
			}

			if _, err := first.DB.ExecContext(
				context.Background(),
				"DROP TRIGGER fail_import_apply_boundary_rs06 ON "+testCase.triggerTable,
			); err != nil {
				t.Fatalf("drop import recovery failure trigger: %v", err)
			}
			if _, err := first.DB.ExecContext(
				context.Background(),
				"DROP FUNCTION public.fail_import_apply_boundary_rs06()",
			); err != nil {
				t.Fatalf("drop import recovery failure function: %v", err)
			}
			first.Server.Close()
			if delay := time.Until(*nextAttemptAt); delay > 0 {
				time.Sleep(delay)
			}

			second := runtime.StartServerWithDatabaseAndObjectStore(
				t,
				"imports-recovery-second-"+slug,
				database,
				durableObjects,
			)
			deadline := time.Now().Add(10 * time.Second)
			var recoveredJobStatus string
			for time.Now().Before(deadline) {
				err := second.DB.QueryRowContext(
					context.Background(),
					"SELECT status FROM jobs WHERE job_id::text = $1",
					applyJobID,
				).Scan(&recoveredJobStatus)
				if err != nil {
					t.Fatalf("read recovered apply job: %v", err)
				}
				if recoveredJobStatus == "succeeded" {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			if recoveredJobStatus != "succeeded" {
				t.Fatalf("recovered apply job status = %q, want succeeded", recoveredJobStatus)
			}
			if got := dbassert.CountSQL(t, second.DB, `
SELECT COUNT(*)
  FROM timeline_events
 WHERE incident_id::text = $1
`, incidentID); got != 1 {
				t.Fatalf("recovery created %d Timeline owner effects, want exactly one", got)
			}
			if got := dbassert.CountSQL(t, second.DB, `
SELECT COUNT(*)
  FROM import_unit_apply_outcomes
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
   AND apply_job_id::text = $3
   AND outcome_status = 'applied'
`, sessionID, unitID, applyJobID); got != 1 {
				t.Fatalf("recovery created %d applied unit outcomes, want exactly one", got)
			}
			if got := dbassert.CountSQL(t, second.DB, `
SELECT COUNT(*)
  FROM import_sessions s
  JOIN import_units u USING (import_session_id)
 WHERE s.import_session_id::text = $1
   AND s.session_status = 'applied'
   AND u.import_unit_id::text = $2
   AND u.unit_status = 'applied'
`, sessionID, unitID); got != 1 {
				t.Fatalf("recovery did not derive applied session/unit state")
			}
			requireImportProof(t, second.DB, applyJobID, "import.apply")
		})
	}
}

func TestTargetRegistryAndEntityOwnerFacade_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "extension_profile-import-target-registry-host")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-import-target-host-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-HOST",
		"title":         "Enterprise integration host import",
	})
	incidentID := incident["incident_id"].(string)

	sessionID, unitID := startCSVImportSession(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		incidentID,
		"txn-extension_profile-import-target-host-upload",
		"hostname,location\nimported-host-1,\n",
		"hosts.csv",
	)
	unknownTargetMapping := map[string]any{
		"client_txn_id":         "txn-extension_profile-import-target-unknown-mapping",
		"target_view_schema_id": "cartulary.view.unknown.v1",
		"header_row_ref":        1,
		"data_start_row_ref":    2,
		"unknown_column_policy": "reject_if_unmapped",
		"source_columns": []map[string]any{
			{
				"source_column_ordinal": 1,
				"source_header_text":    "hostname",
				"field_key":             "host.hostname",
				"entity_binding_mode":   "entity_origin",
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
			{
				"source_column_ordinal": 2,
				"source_header_text":    "location",
				"field_key":             "host.location",
				"entity_binding_mode":   "entity_origin",
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "write_null",
			},
		},
	}
	unknownResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPut, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping", unknownTargetMapping)
	unknownErr := httptestx.RequireErrorEnvelope(t, unknownResp, http.StatusBadRequest, "invalid_import_request")
	if unknownErr["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "target_view_schema_not_importable" {
		t.Fatalf("unexpected unknown target rejection: %#v", unknownErr)
	}

	networkFlowMapping := map[string]any{
		"client_txn_id":           "txn-extension_profile-import-target-network-flow-mapping",
		"target_kind":             imports.ImportTargetKindNetworkFlowTable,
		"extension_profile_id":    imports.NetworkFlowExtensionProfileID,
		"owner_mapping_schema_id": "cartulary.network_flow.mapping_candidate.v1",
		"owner_mapping": map[string]any{
			"source_profile": "cisco_sna_csv_v1",
		},
		"header_row_ref":     1,
		"data_start_row_ref": 2,
		"source_columns": []map[string]any{
			{
				"source_column_ordinal": 1,
				"source_header_text":    "hostname",
				"field_key":             nil,
				"entity_binding_mode":   nil,
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
			{
				"source_column_ordinal": 2,
				"source_header_text":    "location",
				"field_key":             nil,
				"entity_binding_mode":   nil,
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
		},
	}
	networkFlowResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPut, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping", networkFlowMapping)
	networkFlowErr := httptestx.RequireErrorEnvelope(t, networkFlowResp, http.StatusBadRequest, "invalid_import_request")
	if networkFlowErr["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "target_kind_not_importable" {
		t.Fatalf("unexpected network flow target rejection: %#v", networkFlowErr)
	}

	hostMapping := cloneImportMappingPayload(t, unknownTargetMapping)
	hostMapping["client_txn_id"] = "txn-extension_profile-import-target-host-mapping"
	hostMapping["target_view_schema_id"] = "cartulary.view.hosts.v1"
	mappingResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPut, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping", hostMapping)
	mappedUnit := httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)["data"].(map[string]any)
	if mappedUnit["unit_status"] != "mapped" || mappedUnit["mapping_fingerprint"] == nil {
		t.Fatalf("unexpected host mapped unit: %#v", mappedUnit)
	}

	selectResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select", map[string]any{"client_txn_id": "txn-extension_profile-import-target-host-select"})
	httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)
	applyResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-extension_profile-import-target-host-apply"})
	applyJob := httptestx.RequireSuccessEnvelope(t, applyResp, http.StatusAccepted)["data"].(map[string]any)
	appliedJob := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, applyJob["job_id"].(string))
	if appliedJob["status"] != "succeeded" || appliedJob["result_summary"].(map[string]any)["code"] != "import_session_applied" {
		t.Fatalf("unexpected host apply job: %#v", appliedJob)
	}

	queryResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.hosts.v1/query", map[string]any{}, httptestx.WithCookies(adminLogin.SessionCookie))
	rows := httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)["data"].(map[string]any)["rows"].([]any)
	if len(rows) != 1 {
		t.Fatalf("expected one imported host row, got %#v", rows)
	}
	hostCells := rows[0].(map[string]any)["cells"].(map[string]any)
	if got := hostCells["host.hostname"].(map[string]any)["value"]; got != "imported-host-1" {
		t.Fatalf("unexpected imported host hostname: %#v row=%#v", got, rows[0])
	}
	if got := hostCells["host.location"].(map[string]any)["value"]; got != nil {
		t.Fatalf("write_null host location = %#v, want null", got)
	}
}

func TestNetworkFlowImportMappingAndApplyCreatesOneAtomicTable(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix: "network-flow-import-apply",
		Env: map[string]string{
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
		},
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
	adminLogin, adminUserID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-import-incident",
		"incident_key":  "IR-NF-IMPORT",
		"title":         "Network Flow import",
	})
	incidentID := incident["incident_id"].(string)
	csv := strings.Join([]string{
		"Source IP Address,Destination IP Address,Source Port,Destination Port,Protocol,Bytes,Packets,Flow Start Time,Flow End Time,Input Interface,Output Interface",
		"192.0.2.10,2001:db8::1,443,51515,TCP,1234,12,2026-07-10T12:00:00Z,2026-07-10T12:00:05Z, Gi0/1 , Gi0/2 ",
		"192.0.2.11,192.0.2.12,53,53000,UDP,64,1,2026-07-10T12:01:00Z,2026-07-10T12:01:00Z,,",
	}, "\n") + "\n"
	sessionID, unitID := startCSVImportSession(t, harness.Server.HTTP.URL, adminLogin, incidentID, "txn-network-flow-import-upload", csv, "C:\\tmp\\flows.csv")
	previewPayload := networkFlowMappingPreviewPayload()
	beforePreviewIdempotency := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM route_idempotency`)
	invalidPreviewPayload := cloneImportMappingPayload(t, previewPayload)
	invalidPreviewPayload["client_txn_id"] = "preview-must-not-have-a-transaction"
	invalidPreviewResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping-preview", invalidPreviewPayload)
	invalidPreviewError := httptestx.RequireErrorEnvelope(t, invalidPreviewResp, http.StatusBadRequest, "invalid_import_request")
	if invalidPreviewError["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "unknown_field" {
		t.Fatalf("unexpected mapping preview closed-request rejection: %#v", invalidPreviewError)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'viewer', updated_at = now(), updated_by_user_id = $2::uuid
 WHERE incident_id::text = $1 AND user_id::text = $2
`, incidentID, adminUserID); err != nil {
		t.Fatalf("demote mapping preview actor: %v", err)
	}
	viewerPreviewResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping-preview", previewPayload)
	httptestx.RequireErrorEnvelope(t, viewerPreviewResp, http.StatusForbidden, "authorization_denied")
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'admin', updated_at = now(), updated_by_user_id = $2::uuid
 WHERE incident_id::text = $1 AND user_id::text = $2
`, incidentID, adminUserID); err != nil {
		t.Fatalf("restore mapping preview actor: %v", err)
	}
	previewResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping-preview", previewPayload)
	previewResource := httptestx.RequireSuccessEnvelope(t, previewResp, http.StatusOK)["data"].(map[string]any)
	if previewResource["schema_id"] != imports.ExtensionMappingPreviewResultSchemaID ||
		previewResource["import_session_id"] != sessionID ||
		previewResource["import_unit_id"] != unitID ||
		previewResource["target_kind"] != imports.ImportTargetKindNetworkFlowTable ||
		previewResource["extension_profile_id"] != imports.NetworkFlowExtensionProfileID ||
		previewResource["owner_result_schema_id"] != "cartulary.network_flow.import_preview_result.v1" {
		t.Fatalf("unexpected Network Flow mapping preview wrapper: %#v", previewResource)
	}
	ownerPreview := previewResource["owner_result"].(map[string]any)
	previewFingerprint, _ := ownerPreview["mapping_fingerprint"].(string)
	if len(previewFingerprint) != 64 || ownerPreview["preview_record_count"] != float64(2) || ownerPreview["preview_accepted_count"] != float64(2) || ownerPreview["preview_rejected_count"] != float64(0) {
		t.Fatalf("unexpected Network Flow owner preview: %#v", ownerPreview)
	}
	replayedPreviewResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping-preview", previewPayload)
	replayedPreview := httptestx.RequireSuccessEnvelope(t, replayedPreviewResp, http.StatusOK)["data"].(map[string]any)["owner_result"].(map[string]any)
	if replayedPreview["mapping_fingerprint"] != previewFingerprint {
		t.Fatalf("repeat preview changed fingerprint: first=%q second=%#v", previewFingerprint, replayedPreview)
	}
	unitResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units/"+unitID, nil, httptestx.WithCookies(adminLogin.SessionCookie))
	unitBeforeApproval := httptestx.RequireSuccessEnvelope(t, unitResp, http.StatusOK)["data"].(map[string]any)
	if unitBeforeApproval["unit_status"] != "discovered" || unitBeforeApproval["mapping_fingerprint"] != nil || unitBeforeApproval["approved_mapping"] != nil {
		t.Fatalf("mapping preview persisted durable unit state: %#v", unitBeforeApproval)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM route_idempotency`); got != beforePreviewIdempotency {
		t.Fatalf("mapping preview created an idempotency record: before=%d after=%d", beforePreviewIdempotency, got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM network_flow_tables WHERE incident_id::text = $1`, incidentID); got != 0 {
		t.Fatalf("mapping preview allocated Network Flow tables: %d", got)
	}

	mappingResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPut, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping", networkFlowMappingPayload("txn-network-flow-import-mapping"))
	mappedUnit := httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)["data"].(map[string]any)
	mappingFingerprint, _ := mappedUnit["mapping_fingerprint"].(string)
	if len(mappingFingerprint) != 64 {
		t.Fatalf("expected Network Flow mapping fingerprint, got %#v", mappedUnit)
	}
	if mappingFingerprint != previewFingerprint {
		t.Fatalf("durable approval fingerprint %q does not match preview %q", mappingFingerprint, previewFingerprint)
	}
	approved := mappedUnit["approved_mapping"].(map[string]any)
	if approved["target_kind"] != imports.ImportTargetKindNetworkFlowTable || approved["source_columns"] == nil {
		t.Fatalf("expected materialized Network Flow owner mapping, got %#v", approved)
	}

	selectResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select", map[string]any{"client_txn_id": "txn-network-flow-import-select"})
	httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)
	applyResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-network-flow-import-apply"})
	applyJob := httptestx.RequireSuccessEnvelope(t, applyResp, http.StatusAccepted)["data"].(map[string]any)
	appliedJob := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, applyJob["job_id"].(string))
	if appliedJob["status"] != "succeeded" {
		t.Fatalf("unexpected Network Flow apply job: %#v", appliedJob)
	}
	refs := appliedJob["result_summary"].(map[string]any)["resource_refs"].([]any)
	if len(refs) != 2 || refs[1].(map[string]any)["kind"] != imports.ImportTargetKindNetworkFlowTable {
		t.Fatalf("expected import session and Network Flow table refs, got %#v", refs)
	}
	tableID := refs[1].(map[string]any)["id"].(string)
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM network_flow_tables WHERE network_flow_table_id = $1 AND incident_id::text = $2 AND display_name = 'flows' AND row_count_accepted = 2`, tableID, incidentID); got != 1 {
		t.Fatalf("expected one created Network Flow table, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM network_flow_rows WHERE network_flow_table_id = $1 AND src_ip IN ('192.0.2.10', '192.0.2.11')`, tableID); got != 2 {
		t.Fatalf("expected two accepted Network Flow rows, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM network_flow_rejected_row_diagnostics WHERE network_flow_table_id = $1`, tableID); got != 0 {
		t.Fatalf("expected no rejected-row diagnostics, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM import_apply_unit_plans p
  JOIN import_unit_apply_outcomes o
    ON o.import_session_id = p.import_session_id
   AND o.import_unit_id = p.import_unit_id
 WHERE p.import_session_id::text = $1
   AND p.import_unit_id::text = $2
   AND p.apply_job_id::text = $3
   AND p.target_kind = 'network_flow_table'
   AND p.extension_profile_id = 'network_flow_activity'
   AND p.owner_binding_id = 'network_flow_activity.import_facade.v1'
   AND o.outcome_status = 'applied'
   AND o.change_set_id IS NULL
   AND jsonb_array_length(o.resource_refs_json) = 1
   AND o.resource_refs_json->0->>'kind' = 'network_flow_table'
   AND o.resource_refs_json->0->>'id' = $4
   AND o.owner_result_json->'table_ref'->>'id' = $4
`, sessionID, unitID, applyJob["job_id"].(string), tableID); got != 1 {
		t.Fatalf("Network Flow owner effects and durable outcome did not commit as one unit: %d", got)
	}
}

func TestNetworkFlowOwnerErrorsTranslateToSafeImportsFailures_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix: "network-flow-import-owner-errors",
		Env: map[string]string{
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
		},
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-import-errors-incident",
		"incident_key":  "IR-NF-IMPORT-ERRORS",
		"title":         "Network Flow owner-error translation",
	})
	incidentID := incident["incident_id"].(string)
	header := "Source IP Address,Destination IP Address,Source Port,Destination Port,Protocol,Bytes,Packets,Flow Start Time,Flow End Time,Input Interface,Output Interface"

	noDataSessionID, noDataUnitID := startCSVImportSession(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		incidentID,
		"txn-network-flow-import-errors-no-data-upload",
		header+"\n",
		"header-only.csv",
	)
	noDataResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+noDataSessionID+"/units/"+noDataUnitID+
			"/mapping-preview",
		networkFlowMappingPreviewPayload(),
	)
	noDataEnvelope := httptestx.RequireErrorEnvelope(
		t,
		noDataResp,
		http.StatusBadRequest,
		"invalid_import_request",
	)
	noDataDetails := noDataEnvelope["error"].(map[string]any)["details"].(map[string]any)
	noDataOwner := noDataDetails["owner_error"].(map[string]any)
	if noDataDetails["reason_code"] != "owner_preview_validation_failed" ||
		noDataOwner["schema_id"] != "cartulary.network_flow.import_owner_error.v1" ||
		noDataOwner["owner_code"] != "network_flow_no_data_rows" ||
		noDataOwner["retryable"] != false ||
		len(noDataOwner["safe_details"].(map[string]any)) != 0 {
		t.Fatalf("unsafe or incorrect no-data translation: %#v", noDataEnvelope)
	}

	allRejectedCSV := strings.Join([]string{
		header,
		"raw-secret-invalid-source,raw-secret-invalid-destination,443,51515,TCP,1234,12,2026-07-10T12:00:00Z,2026-07-10T12:00:05Z,Gi0/1,Gi0/2",
	}, "\n") + "\n"
	sessionID, unitID := startCSVImportSession(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		incidentID,
		"txn-network-flow-import-errors-all-rejected-upload",
		allRejectedCSV,
		"all-rejected.csv",
	)
	mappingResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPut,
		"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping",
		networkFlowMappingPayload("txn-network-flow-import-errors-mapping"),
	)
	httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)
	selectResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select",
		map[string]any{"client_txn_id": "txn-network-flow-import-errors-select"},
	)
	httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)
	applyResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/apply",
		map[string]any{"client_txn_id": "txn-network-flow-import-errors-apply"},
	)
	applyJob := httptestx.RequireSuccessEnvelope(
		t,
		applyResp,
		http.StatusAccepted,
	)["data"].(map[string]any)
	failedJob := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, applyJob["job_id"].(string))
	errorSummary := failedJob["error_summary"].(map[string]any)
	errorDetails := errorSummary["details"].(map[string]any)
	ownerError := errorDetails["owner_error"].(map[string]any)
	safeDetails := ownerError["safe_details"].(map[string]any)
	diagnostics := safeDetails["diagnostics_sample"].([]any)
	if failedJob["status"] != "failed" ||
		errorSummary["code"] != "import_apply_blocked" ||
		errorSummary["retryable"] != false ||
		errorDetails["reason_code"] != "owner_apply_validation_failed" ||
		ownerError["owner_code"] != "network_flow_all_rows_rejected" ||
		len(diagnostics) == 0 ||
		strings.Contains(fmt.Sprint(errorSummary), "raw-secret") ||
		strings.Contains(fmt.Sprint(errorSummary), "all-rejected.csv") {
		t.Fatalf("unsafe or incorrect all-rejected translation: %#v", failedJob)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM import_unit_apply_outcomes
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
   AND outcome_status = 'failed'
   AND error_code = 'import_apply_blocked'
   AND reason_code = 'owner_apply_validation_failed'
   AND error_retryable = false
   AND error_details_json->'owner_error'->>'owner_code' =
       'network_flow_all_rows_rejected'
`, sessionID, unitID); got != 1 {
		t.Fatalf("durable translated owner-error outcome count = %d, want 1", got)
	}
	if got := dbassert.CountSQL(
		t,
		harness.DB,
		`SELECT COUNT(*) FROM network_flow_tables WHERE incident_id::text = $1`,
		incidentID,
	); got != 0 {
		t.Fatalf("owner validation failure created %d Network Flow tables", got)
	}
}

func TestCancellationAfterCommittedUnitDerivesPartialApplication_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "imports-partial-cancellation")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-imports-partial-cancel-incident",
		"incident_key":  "IR-IMPORT-PARTIAL-CANCEL",
		"title":         "Import partial cancellation",
	})
	incidentID := incident["incident_id"].(string)
	metadata := `{"client_txn_id":"txn-imports-partial-cancel-upload","incident_id":"` + incidentID + `"}`
	uploadResp := postImportUploadBytes(
		t,
		harness.Server.HTTP.URL,
		adminLogin,
		metadata,
		multipleSheetXLSX(t),
		"partial-cancel.xlsx",
		imports.MediaTypeXLSX,
		false,
	)
	uploadJob := httptestx.RequireSuccessEnvelope(t, uploadResp, http.StatusAccepted)["data"].(map[string]any)
	discoveryJob := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, uploadJob["job_id"].(string))
	sessionID := discoveryJob["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)
	unitsResp := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/units",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	units := httptestx.RequireSuccessEnvelope(t, unitsResp, http.StatusOK)["data"].(map[string]any)["import_units"].([]any)
	if len(units) != 2 {
		t.Fatalf("partial-cancellation workbook discovered %d units, want 2", len(units))
	}
	headers := [][2]string{{"host", "summary"}, {"indicator", "type"}}
	unitIDs := make([]string, len(units))
	for index, rawUnit := range units {
		unitID := rawUnit.(map[string]any)["import_unit_id"].(string)
		unitIDs[index] = unitID
		mappingResp := doImportJSON(
			t,
			harness.Server.HTTP.URL,
			adminLogin,
			http.MethodPut,
			"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping",
			map[string]any{
				"client_txn_id":         "txn-imports-partial-cancel-mapping-" + string(rune('1'+index)),
				"target_view_schema_id": "cartulary.view.timeline.v2",
				"header_row_ref":        1,
				"data_start_row_ref":    2,
				"unknown_column_policy": "preserve_raw_capture",
				"source_columns": []map[string]any{
					{
						"source_column_ordinal": 1,
						"source_header_text":    headers[index][0],
						"field_key":             "timeline.activity_synopsis_text",
						"entity_binding_mode":   nil,
						"transform_id":          nil,
						"transform_options":     map[string]any{},
						"empty_value_policy":    "omit_field",
					},
					{
						"source_column_ordinal": 2,
						"source_header_text":    headers[index][1],
						"field_key":             nil,
						"entity_binding_mode":   nil,
						"transform_id":          nil,
						"transform_options":     map[string]any{},
						"empty_value_policy":    "omit_field",
					},
				},
			},
		)
		httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)
		selectResp := doImportJSON(
			t,
			harness.Server.HTTP.URL,
			adminLogin,
			http.MethodPost,
			"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select",
			map[string]any{
				"client_txn_id": "txn-imports-partial-cancel-select-" + string(rune('1'+index)),
			},
		)
		httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)
	}

	var advisoryNamespace int32
	var advisoryKey int32
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT hashtext(current_database()), hashtext($1)
`, t.Name()).Scan(&advisoryNamespace, &advisoryKey); err != nil {
		t.Fatalf("derive partial-cancellation advisory key: %v", err)
	}
	blocker, err := harness.DB.Conn(context.Background())
	if err != nil {
		t.Fatalf("acquire partial-cancellation blocker connection: %v", err)
	}
	defer blocker.Close()
	var blockerPID int32
	if err := blocker.QueryRowContext(context.Background(), `SELECT pg_backend_pid()`).Scan(&blockerPID); err != nil {
		t.Fatalf("load partial-cancellation blocker pid: %v", err)
	}
	if _, err := blocker.ExecContext(context.Background(), "SELECT pg_advisory_lock($1, $2)", advisoryNamespace, advisoryKey); err != nil {
		t.Fatalf("acquire partial-cancellation advisory lock: %v", err)
	}
	blockerLocked := true
	defer func() {
		if blockerLocked {
			_, _ = blocker.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1, $2)", advisoryNamespace, advisoryKey)
		}
	}()
	fixtureSQL := fmt.Sprintf(`
CREATE FUNCTION public.block_first_import_outcome_rs06()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.discovery_sequence = 1 THEN
        PERFORM pg_advisory_xact_lock(%d, %d);
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER block_first_import_outcome_rs06
BEFORE INSERT ON import_unit_apply_outcomes
FOR EACH ROW
EXECUTE FUNCTION public.block_first_import_outcome_rs06()
`, advisoryNamespace, advisoryKey)
	if _, err := harness.DB.ExecContext(context.Background(), fixtureSQL); err != nil {
		t.Fatalf("install partial-cancellation serialization fixture: %v", err)
	}

	type asyncResponse struct {
		response *http.Response
		err      error
	}
	applyResponse := make(chan asyncResponse, 1)
	go func() {
		request, requestErr := http.NewRequest(
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/import-sessions/"+sessionID+"/apply",
			bytes.NewBufferString(`{"client_txn_id":"txn-imports-partial-cancel-apply"}`),
		)
		if requestErr != nil {
			applyResponse <- asyncResponse{err: requestErr}
			return
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)
		request.AddCookie(adminLogin.SessionCookie)
		request.AddCookie(adminLogin.CSRFCookie)
		response, requestErr := http.DefaultClient.Do(request)
		applyResponse <- asyncResponse{response: response, err: requestErr}
	}()

	var applyJobID string
	var applyBackendPID int32
	deadline := time.Now().Add(5 * time.Second)
	outcomeBlocked := false
	for time.Now().Before(deadline) {
		err = harness.DB.QueryRowContext(context.Background(), `
SELECT
    COALESCE((
        SELECT activity.pid
          FROM pg_stat_activity AS activity
         WHERE activity.datname = current_database()
           AND $2::integer = ANY(pg_blocking_pids(activity.pid))
           AND EXISTS (
               SELECT 1
                 FROM pg_locks AS waiting_lock
                WHERE waiting_lock.pid = activity.pid
                  AND waiting_lock.locktype = 'advisory'
                  AND NOT waiting_lock.granted
           )
         ORDER BY activity.pid
         LIMIT 1
    ), 0)::integer,
    COALESCE((
        SELECT apply_job_id::text
          FROM import_sessions
         WHERE import_session_id::text = $1
           AND apply_job_id IS NOT NULL
    ), '')
`, sessionID, blockerPID).Scan(&applyBackendPID, &applyJobID)
		if err != nil {
			t.Fatalf("observe first unit commit boundary: %v", err)
		}
		if applyBackendPID != 0 && applyJobID != "" {
			outcomeBlocked = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !outcomeBlocked || applyJobID == "" {
		t.Fatalf("apply did not reach the first unit commit boundary; locks=%s", databaseLockGraph(harness.DB))
	}

	cancelResponse := make(chan asyncResponse, 1)
	go func() {
		request, requestErr := http.NewRequest(
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/jobs/"+applyJobID+"/cancel",
			bytes.NewBufferString(`{"client_txn_id":"txn-imports-partial-cancel-request"}`),
		)
		if requestErr != nil {
			cancelResponse <- asyncResponse{err: requestErr}
			return
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)
		request.AddCookie(adminLogin.SessionCookie)
		request.AddCookie(adminLogin.CSRFCookie)
		response, requestErr := http.DefaultClient.Do(request)
		cancelResponse <- asyncResponse{response: response, err: requestErr}
	}()

	deadline = time.Now().Add(5 * time.Second)
	cancelWaiting := false
	var cancelBackendPID int32
	for time.Now().Before(deadline) {
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT COALESCE((
    SELECT activity.pid
      FROM pg_stat_activity AS activity
     WHERE activity.datname = current_database()
       AND $1::integer = ANY(pg_blocking_pids(activity.pid))
       AND EXISTS (
           SELECT 1
             FROM pg_locks AS waiting_lock
            WHERE waiting_lock.pid = activity.pid
              AND waiting_lock.locktype = 'advisory'
              AND NOT waiting_lock.granted
       )
     ORDER BY activity.pid
     LIMIT 1
), 0)::integer
`, applyBackendPID).Scan(&cancelBackendPID); err != nil {
			t.Fatalf("observe waiting cancellation: %v", err)
		}
		if cancelBackendPID != 0 {
			cancelWaiting = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !cancelWaiting {
		t.Fatalf(
			"cancel request did not wait on the apply transition lock held by pid %d; locks=%s",
			applyBackendPID,
			databaseLockGraph(harness.DB),
		)
	}
	if _, err := blocker.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1, $2)", advisoryNamespace, advisoryKey); err != nil {
		t.Fatalf("release partial-cancellation unit commit: %v", err)
	}
	blockerLocked = false

	select {
	case result := <-cancelResponse:
		if result.err != nil {
			t.Fatalf("cancel partially applied import: %v", result.err)
		}
		httptestx.RequireSuccessEnvelope(t, result.response, http.StatusOK)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for partial-cancellation request")
	}
	select {
	case result := <-applyResponse:
		if result.err != nil {
			t.Fatalf("complete partially canceled apply: %v", result.err)
		}
		httptestx.RequireSuccessEnvelope(t, result.response, http.StatusAccepted)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for partially canceled apply")
	}

	if got := waitForSQLCount(t, harness.DB, 60*time.Second, 1, `
SELECT COUNT(*)
  FROM import_sessions s
  JOIN jobs j ON j.job_id = s.apply_job_id
 WHERE s.import_session_id::text = $1
   AND s.session_status = 'partially_applied'
   AND j.status = 'canceled'
   AND j.result_summary_json->>'code' = 'import_session_partially_applied'
`, sessionID); got != 1 {
		t.Fatalf(
			"cancellation after a committed unit did not derive truthful partial application: %d state=%#v",
			got,
			importApplyState(t, harness.DB, sessionID),
		)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM import_unit_apply_outcomes
 WHERE import_session_id::text = $1
   AND apply_job_id::text = $2
   AND outcome_status IN ('applied', 'canceled')
`, sessionID, applyJobID); got != 2 {
		t.Fatalf("partial cancellation did not persist one outcome per unit: %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM import_unit_apply_outcomes
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
   AND outcome_status = 'applied'
`, sessionID, unitIDs[0]); got != 1 {
		t.Fatalf("first unit was not durably applied before cancellation: %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM import_unit_apply_outcomes
 WHERE import_session_id::text = $1
   AND import_unit_id::text = $2
   AND outcome_status = 'canceled'
   AND error_code = 'import_apply_canceled'
   AND reason_code = 'cancel_requested'
`, sessionID, unitIDs[1]); got != 1 {
		t.Fatalf("remaining unit did not receive a durable canceled outcome: %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM timeline_events
 WHERE incident_id::text = $1
`, incidentID); got != 2 {
		t.Fatalf("partial cancellation committed %d Timeline effects, want only the first unit's 2", got)
	}
}

func databaseLockGraph(db *sql.DB) string {
	rows, err := db.QueryContext(context.Background(), `
SELECT activity.pid,
       lock.locktype,
       lock.mode,
       lock.granted,
       pg_blocking_pids(activity.pid)::text
  FROM pg_stat_activity AS activity
  JOIN pg_locks AS lock ON lock.pid = activity.pid
 WHERE activity.datname = current_database()
 ORDER BY activity.pid, lock.locktype, lock.mode, lock.granted
 LIMIT 32
`)
	if err != nil {
		return "unavailable:" + err.Error()
	}
	defer rows.Close()
	observations := make([]string, 0, 32)
	for rows.Next() {
		var pid int32
		var lockType string
		var mode string
		var granted bool
		var blockers string
		if err := rows.Scan(&pid, &lockType, &mode, &granted, &blockers); err != nil {
			return "unavailable:" + err.Error()
		}
		observations = append(observations, fmt.Sprintf(
			"pid=%d type=%s mode=%s granted=%t blockers=%s",
			pid,
			lockType,
			mode,
			granted,
			blockers,
		))
	}
	if err := rows.Err(); err != nil {
		return "unavailable:" + err.Error()
	}
	return strings.Join(observations, "; ")
}

func TestImportsEvidenceCreateParity_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "extension_profile-import-evidence-owner-facade")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-import-evidence-owner-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-EVIDENCE",
		"title":         "Enterprise integration evidence owner facade",
	})
	incidentID := incident["incident_id"].(string)
	requestedAt := "2026-08-11T20:00:00Z"
	createBody := map[string]any{
		"client_txn_id":         "txn-extension_profile-evidence-ordinary-create",
		"evidence.title":        "Evidence create parity",
		"evidence.requested_at": requestedAt,
	}
	createResp := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.evidence.v1/rows",
		createBody,
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	ordinaryRow := createData["row"].(map[string]any)
	ordinaryRecordID := ordinaryRow["record_id"].(string)
	createReplay := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.evidence.v1/rows",
		createBody,
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	replayedCreate := httptestx.RequireSuccessEnvelope(t, createReplay, http.StatusOK)["data"].(map[string]any)
	if replayedCreate["row"].(map[string]any)["record_id"] != ordinaryRecordID {
		t.Fatalf("ordinary Evidence replay changed record: first=%s replay=%#v", ordinaryRecordID, replayedCreate)
	}

	sessionID, unitID := startCSVImportSession(t, harness.Server.HTTP.URL, adminLogin, incidentID, "txn-extension_profile-import-evidence-owner-upload", "title,requested_at\nEvidence create parity,"+requestedAt+"\n", "evidence.csv")
	mapping := map[string]any{
		"client_txn_id":         "txn-extension_profile-import-evidence-owner-mapping",
		"target_view_schema_id": "cartulary.view.evidence.v1",
		"header_row_ref":        1,
		"data_start_row_ref":    2,
		"unknown_column_policy": "reject_if_unmapped",
		"source_columns": []map[string]any{
			{
				"source_column_ordinal": 1,
				"source_header_text":    "title",
				"field_key":             "evidence.title",
				"entity_binding_mode":   nil,
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
			{
				"source_column_ordinal": 2,
				"source_header_text":    "requested_at",
				"field_key":             "evidence.requested_at",
				"entity_binding_mode":   nil,
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
		},
	}
	mappingResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPut, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping", mapping)
	httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)
	selectResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select", map[string]any{"client_txn_id": "txn-extension_profile-import-evidence-owner-select"})
	httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)

	applyResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-extension_profile-import-evidence-owner-apply"})
	applyJob := httptestx.RequireSuccessEnvelope(t, applyResp, http.StatusAccepted)["data"].(map[string]any)
	appliedJob := waitImportJobTerminal(t, harness.Server.HTTP.URL, adminLogin, applyJob["job_id"].(string))
	if appliedJob["status"] != "succeeded" || appliedJob["result_summary"].(map[string]any)["code"] != "import_session_applied" {
		t.Fatalf("unexpected evidence apply job: %#v", appliedJob)
	}
	replayApplyResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-extension_profile-import-evidence-owner-apply"})
	replayedApplyJob := httptestx.RequireSuccessEnvelope(t, replayApplyResp, http.StatusAccepted)["data"].(map[string]any)
	if replayedApplyJob["job_id"] != applyJob["job_id"] {
		t.Fatalf("Evidence import replay returned a different job: first=%#v replay=%#v", applyJob, replayedApplyJob)
	}

	var importedRecordID string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT record_id::text
  FROM import_apply_journal
 WHERE import_session_id::text = $1
`, sessionID).Scan(&importedRecordID); err != nil {
		t.Fatalf("load imported Evidence record: %v", err)
	}

	queryResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.evidence.v1/query", map[string]any{}, httptestx.WithCookies(adminLogin.SessionCookie))
	rows := httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)["data"].(map[string]any)["rows"].([]any)
	if len(rows) != 2 {
		t.Fatalf("expected ordinary and imported Evidence rows, got %#v", rows)
	}
	rowsByID := make(map[string]map[string]any, len(rows))
	for _, value := range rows {
		row := value.(map[string]any)
		rowsByID[row["record_id"].(string)] = row
	}
	importedRow := rowsByID[importedRecordID]
	if importedRow == nil {
		t.Fatalf("query omitted imported Evidence record %s: %#v", importedRecordID, rows)
	}
	ordinaryQueriedRow := rowsByID[ordinaryRecordID]
	if ordinaryQueriedRow == nil {
		t.Fatalf("query omitted ordinary Evidence record %s: %#v", ordinaryRecordID, rows)
	}
	normalizedCells := func(row map[string]any) map[string]any {
		cloned := cloneImportMappingPayload(t, row["cells"].(map[string]any))
		delete(cloned, "evidence.edited_at")
		return cloned
	}
	if !reflect.DeepEqual(normalizedCells(ordinaryQueriedRow), normalizedCells(importedRow)) {
		t.Fatalf("ordinary/import returned projection cells diverged: ordinary=%#v imported=%#v", ordinaryQueriedRow, importedRow)
	}
	if ordinaryQueriedRow["row_version"] != importedRow["row_version"] {
		t.Fatalf("ordinary/import row versions diverged: ordinary=%#v imported=%#v", ordinaryQueriedRow["row_version"], importedRow["row_version"])
	}

	var ordinarySource, importedSource string
	for recordID, destination := range map[string]*string{
		ordinaryRecordID: &ordinarySource,
		importedRecordID: &importedSource,
	} {
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT (to_jsonb(e) - 'record_id' - 'incident_id' - 'created_at' - 'updated_at')::text
  FROM evidence e
 WHERE record_id::text = $1
`, recordID).Scan(destination); err != nil {
			t.Fatalf("load Evidence source state for %s: %v", recordID, err)
		}
	}
	if ordinarySource != importedSource {
		t.Fatalf("ordinary/import source rows diverged: ordinary=%s imported=%s", ordinarySource, importedSource)
	}
	for _, recordID := range []string{ordinaryRecordID, importedRecordID} {
		for effect, query := range map[string]string{
			"record revision":      `SELECT COUNT(*) FROM record_revisions WHERE record_id::text = $1 AND row_version = 1`,
			"change-set mutation":  `SELECT COUNT(*) FROM change_set_mutations WHERE target_kind = 'record' AND target_id = $1 AND operation_kind = 'create'`,
			"Collaboration intent": `SELECT COUNT(*) FROM collaboration_event_intents WHERE source_record_id::text = $1 AND event_family = 'record_changed'`,
			"projection row":       `SELECT COUNT(*) FROM evidence_grid_projection WHERE record_id::text = $1 AND row_version = 1`,
		} {
			if got := dbassert.CountSQL(t, harness.DB, query, recordID); got != 1 {
				t.Fatalf("%s count for Evidence record %s = %d, want 1", effect, recordID, got)
			}
		}
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM import_apply_journal WHERE import_session_id::text = $1`, sessionID); got != 1 {
		t.Fatalf("expected one import apply journal row, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE source = 'imports.apply' AND client_txn_id = $1`, "import:"+sessionID+":"+unitID+":txn-extension_profile-import-evidence-owner-apply"); got != 1 {
		t.Fatalf("expected one unit-level import change set, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM evidence WHERE incident_id::text = $1`, incidentID); got != 2 {
		t.Fatalf("ordinary/import replays created duplicate Evidence rows: %d", got)
	}
}

func createImportAutoResolutionCandidateHost(t testing.TB, serverURL string, login flowtest.LoginResult, incidentID string) string {
	t.Helper()

	resp := httptestx.DoJSON(
		t,
		http.MethodPost,
		serverURL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.hosts.v1/rows",
		map[string]any{
			"client_txn_id":     "txn-extension_profile-import-apply-host-alias-candidate",
			"host.display_name": "Import existing host",
			"host.hostname":     "import-existing-host",
			"host.aliases": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_alias", "alias_text": "host-1"},
				},
			},
		},
		httptestx.WithCookies(login.SessionCookie, login.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
	return data["row"].(map[string]any)["record_id"].(string)
}

func createImportAutoResolutionCandidateIdentity(t testing.TB, serverURL string, login flowtest.LoginResult, incidentID string) string {
	t.Helper()

	resp := httptestx.DoJSON(
		t,
		http.MethodPost,
		serverURL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.identities.v1/rows",
		map[string]any{
			"client_txn_id":         "txn-extension_profile-import-apply-identity-alias-candidate",
			"identity.display_name": "Import existing identity",
			"identity.aliases": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_alias", "alias_text": "identity-1"},
				},
			},
		},
		httptestx.WithCookies(login.SessionCookie, login.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
	return data["row"].(map[string]any)["record_id"].(string)
}

func cloneImportMappingPayload(t testing.TB, source map[string]any) map[string]any {
	t.Helper()

	data, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("marshal import mapping payload: %v", err)
	}
	var cloned map[string]any
	if err := json.Unmarshal(data, &cloned); err != nil {
		t.Fatalf("clone import mapping payload: %v", err)
	}
	return cloned
}

func prepareReadyTimelineImport(
	t testing.TB,
	harness *appsupport.ServerHarness,
	login flowtest.LoginResult,
	incidentID string,
	clientPrefix string,
) (string, string) {
	t.Helper()
	sessionID, unitID := startCSVImportSession(
		t,
		harness.Server.HTTP.URL,
		login,
		incidentID,
		"txn-"+clientPrefix+"-upload",
		"summary\nrecover exactly once\n",
		"recovery.csv",
	)
	mappingResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		login,
		http.MethodPut,
		"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping",
		map[string]any{
			"client_txn_id":         "txn-" + clientPrefix + "-mapping",
			"target_view_schema_id": "cartulary.view.timeline.v2",
			"header_row_ref":        1,
			"data_start_row_ref":    2,
			"unknown_column_policy": "preserve_raw_capture",
			"source_columns": []map[string]any{{
				"source_column_ordinal": 1,
				"source_header_text":    "summary",
				"field_key":             "timeline.activity_synopsis_text",
				"entity_binding_mode":   nil,
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			}},
		},
	)
	httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)
	selectResp := doImportJSON(
		t,
		harness.Server.HTTP.URL,
		login,
		http.MethodPost,
		"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select",
		map[string]any{"client_txn_id": "txn-" + clientPrefix + "-select"},
	)
	httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)
	return sessionID, unitID
}

func approveTimelineImportMapping(
	t testing.TB,
	serverURL string,
	login flowtest.LoginResult,
	sessionID string,
	unitID string,
	headers []string,
	clientPrefix string,
) {
	t.Helper()
	sourceColumns := make([]map[string]any, 0, len(headers))
	for index, header := range headers {
		var fieldKey any
		if index == 0 {
			fieldKey = "timeline.activity_synopsis_text"
		}
		sourceColumns = append(sourceColumns, map[string]any{
			"source_column_ordinal": index + 1,
			"source_header_text":    header,
			"field_key":             fieldKey,
			"entity_binding_mode":   nil,
			"transform_id":          nil,
			"transform_options":     map[string]any{},
			"empty_value_policy":    "omit_field",
		})
	}
	mappingResp := doImportJSON(
		t,
		serverURL,
		login,
		http.MethodPut,
		"/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping",
		map[string]any{
			"client_txn_id":         "txn-" + clientPrefix + "-mapping",
			"target_view_schema_id": "cartulary.view.timeline.v2",
			"header_row_ref":        1,
			"data_start_row_ref":    2,
			"unknown_column_policy": "preserve_raw_capture",
			"source_columns":        sourceColumns,
		},
	)
	httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)
}

func startCSVImportSession(t testing.TB, serverURL string, login flowtest.LoginResult, incidentID string, clientTxnID string, csv string, filename string) (string, string) {
	t.Helper()

	metadata := `{"client_txn_id":"` + clientTxnID + `","incident_id":"` + incidentID + `"}`
	uploadResp := postImportUpload(t, serverURL, login, metadata, csv, filename, false)
	uploadJob := httptestx.RequireSuccessEnvelope(t, uploadResp, http.StatusAccepted)["data"].(map[string]any)
	job := waitImportJobTerminal(t, serverURL, login, uploadJob["job_id"].(string))
	if job["status"] != "succeeded" {
		t.Fatalf("unexpected discovery job: %#v", job)
	}
	sessionID := job["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)
	unitsResp := httptestx.DoJSON(t, http.MethodGet, serverURL+"/api/v1/import-sessions/"+sessionID+"/units", nil, httptestx.WithCookies(login.SessionCookie))
	units := httptestx.RequireSuccessEnvelope(t, unitsResp, http.StatusOK)["data"].(map[string]any)["import_units"].([]any)
	if len(units) != 1 {
		t.Fatalf("expected one import unit, got %#v", units)
	}
	return sessionID, units[0].(map[string]any)["import_unit_id"].(string)
}

func waitImportJobTerminal(t testing.TB, serverURL string, login flowtest.LoginResult, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		response := httptestx.DoJSON(t, http.MethodGet, serverURL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.SessionCookie))
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		switch job["status"] {
		case "succeeded", "failed", "canceled":
			return job
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for import job %s: %#v", jobID, job)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func requireTimelineMentionUnresolved(t testing.TB, row map[string]any, fieldKey string, rawText string) {
	t.Helper()

	cells := row["cells"].(map[string]any)
	mentions := cells[fieldKey].(map[string]any)["value"].(map[string]any)
	items := mentions["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected one imported %s mention, got %#v", fieldKey, mentions)
	}
	item := items[0].(map[string]any)
	if item["item_kind"] != "unresolved_mention" || item["raw_text"] != rawText {
		t.Fatalf("expected unresolved imported %s mention %q, got %#v", fieldKey, rawText, item)
	}
	if _, ok := item["resolved_record_id"]; ok {
		t.Fatalf("imported %s mention must not be resolved: %#v", fieldKey, item)
	}
	if _, ok := item["resolution_method"]; ok {
		t.Fatalf("imported %s mention must not expose resolution method: %#v", fieldKey, item)
	}
}

func requireImportedMentionNotAutoResolved(t testing.TB, db *sql.DB, incidentID string, timelineRecordID string, entityRecordID string, sourceFieldKey string, linkType string, rawText string) {
	t.Helper()

	if got := dbassert.CountSQL(t, db, `
SELECT COUNT(*)
  FROM entity_mentions
 WHERE source_record_id::text = $1
   AND source_field_key = $2
   AND raw_text = $3
   AND resolution_status = 'unresolved'
   AND resolved_record_id IS NULL
   AND resolution_method IS NULL
`, timelineRecordID, sourceFieldKey, rawText); got != 1 {
		t.Fatalf("imported %s token must remain one unresolved mention, got %d", sourceFieldKey, got)
	}
	if got := dbassert.CountSQL(t, db, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND dst_record_id::text = $3
   AND link_type = $4
   AND deleted_at IS NULL
`, incidentID, timelineRecordID, entityRecordID, linkType); got != 0 {
		t.Fatalf("imported exact alias must not create active %s link, got %d", linkType, got)
	}
	if got := dbassert.CountSQL(t, db, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND dst_record_id::text = $3
   AND provenance = 'auto_match'
   AND deleted_at IS NULL
`, incidentID, timelineRecordID, entityRecordID); got != 0 {
		t.Fatalf("imported exact alias must not create auto_match link, got %d", got)
	}
}

func requireTimelineImportProvenance(t testing.TB, db *sql.DB, recordID string, sessionID string, unitID string, header string, rawValue string, rowOrdinal int, columnOrdinal int, sourceRect string) {
	t.Helper()

	var (
		metadataJSON []byte
		headerJSON   []byte
		storedRaw    string
		cellKind     *string
		storedRow    int
		storedColumn int
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT source_metadata, source_header_json, raw_value, cell_kind,
       source_row_ordinal, source_column_ordinal
  FROM timeline_source_provenance
 WHERE record_id::text = $1
`, recordID).Scan(&metadataJSON, &headerJSON, &storedRaw, &cellKind, &storedRow, &storedColumn); err != nil {
		t.Fatalf("query timeline source provenance: %v", err)
	}
	var column map[string]any
	if err := json.Unmarshal(metadataJSON, &column); err != nil {
		t.Fatalf("decode timeline source metadata: %v", err)
	}
	var storedHeader any
	if err := json.Unmarshal(headerJSON, &storedHeader); err != nil {
		t.Fatalf("decode timeline source header: %v", err)
	}
	column["source_header_text"] = storedHeader
	column["raw_value"] = storedRaw
	column["source_row_ordinal"] = storedRow
	column["source_column_ordinal"] = storedColumn
	if cellKind != nil {
		column["cell_kind"] = *cellKind
	}
	want := map[string]any{
		"source_kind":           "file_import",
		"import_session_id":     sessionID,
		"import_unit_id":        unitID,
		"source_file_kind":      imports.SourceFileKindCSV,
		"parser_profile_id":     imports.ParserProfileWorkbookImport,
		"parser_version":        imports.ParserVersionWorkbookImport,
		"locator_kind":          "csv_file",
		"locator":               "file",
		"source_rect_a1":        sourceRect,
		"source_header_text":    header,
		"raw_value":             rawValue,
		"cell_kind":             "string",
		"source_row_ordinal":    rowOrdinal,
		"source_column_ordinal": columnOrdinal,
	}
	for key, wantValue := range want {
		if column[key] != wantValue {
			t.Fatalf("unexpected source provenance %s: got %#v want %#v column=%#v", key, column[key], wantValue, column)
		}
	}
	if column["mapping_fingerprint"] == "" || column["source_content_sha256"] == "" {
		t.Fatalf("expected mapping fingerprint and source content hash in source provenance, got %#v", column)
	}
}

type importCounts struct {
	Sessions         int
	Units            int
	SourceStreams    int
	Jobs             int
	RouteIdempotency int
}

func networkFlowMappingPayload(clientTxnID string) map[string]any {
	headers := []string{
		"Source IP Address",
		"Destination IP Address",
		"Source Port",
		"Destination Port",
		"Protocol",
		"Bytes",
		"Packets",
		"Flow Start Time",
		"Flow End Time",
		"Input Interface",
		"Output Interface",
	}
	fieldKeys := []string{
		"network_flow.src_ip",
		"network_flow.dst_ip",
		"network_flow.src_port",
		"network_flow.dst_port",
		"network_flow.ip_protocol",
		"network_flow.bytes_count",
		"network_flow.packets_count",
		"network_flow.flow_start_utc",
		"network_flow.flow_end_utc",
		"network_flow.input_interface",
		"network_flow.output_interface",
	}
	transforms := map[string]string{
		"network_flow.src_ip":           "ip_literal_v1",
		"network_flow.dst_ip":           "ip_literal_v1",
		"network_flow.src_port":         "port_number_v1",
		"network_flow.dst_port":         "port_number_v1",
		"network_flow.ip_protocol":      "protocol_number_or_token_v1",
		"network_flow.bytes_count":      "uint64_decimal_string_v1",
		"network_flow.packets_count":    "uint64_decimal_string_v1",
		"network_flow.flow_start_utc":   "timestamp_profile_v1",
		"network_flow.flow_end_utc":     "timestamp_profile_v1",
		"network_flow.input_interface":  "trim_ascii_space_v1",
		"network_flow.output_interface": "trim_ascii_space_v1",
	}
	sourceColumns := make([]map[string]any, 0, len(headers))
	fieldMappings := make([]map[string]any, 0, len(headers))
	for index, header := range headers {
		sourceColumns = append(sourceColumns, map[string]any{
			"source_column_ordinal": index + 1,
			"source_header_text":    header,
			"field_key":             nil,
			"entity_binding_mode":   nil,
			"transform_id":          nil,
			"transform_options":     map[string]any{},
			"empty_value_policy":    "omit_field",
		})
		fieldKey := fieldKeys[index]
		emptyPolicy := "empty_string_is_invalid"
		if fieldKey == "network_flow.input_interface" || fieldKey == "network_flow.output_interface" {
			emptyPolicy = "empty_string_is_null"
		}
		fieldMappings = append(fieldMappings, map[string]any{
			"mapping_kind":          "source_column",
			"field_key":             fieldKey,
			"source_column_ordinal": index + 1,
			"transform_id":          transforms[fieldKey],
			"empty_value_policy":    emptyPolicy,
			"combinability":         "single_source_only",
		})
	}
	return map[string]any{
		"client_txn_id":           clientTxnID,
		"target_kind":             imports.ImportTargetKindNetworkFlowTable,
		"extension_profile_id":    imports.NetworkFlowExtensionProfileID,
		"owner_mapping_schema_id": "cartulary.network_flow.mapping_candidate.v1",
		"owner_mapping": map[string]any{
			"target_kind":            imports.ImportTargetKindNetworkFlowTable,
			"target_table_schema_id": "cartulary.network_flow_table.v1",
			"source_profile_id":      "cisco_sna_netflow_csv_v1",
			"parser_profile_id":      "rfc4180_headered_csv_v1",
			"unknown_column_policy":  "preserve_unmapped_raw",
			"timestamp_profile": map[string]any{
				"schema_id":                   "cartulary.network_flow.timestamp_profile.v1",
				"mode":                        "rfc3339",
				"precision":                   "seconds",
				"timezone":                    nil,
				"timezone_ruleset_id":         nil,
				"ambiguous_local_time_policy": "reject",
				"local_time_gap_policy":       "reject",
			},
			"field_mappings": fieldMappings,
		},
		"header_row_ref":     1,
		"data_start_row_ref": 2,
		"source_columns":     sourceColumns,
	}
}

func networkFlowMappingPreviewPayload() map[string]any {
	approval := networkFlowMappingPayload("preview-does-not-use-client-transaction")
	return map[string]any{
		"target_kind":             approval["target_kind"],
		"extension_profile_id":    approval["extension_profile_id"],
		"owner_mapping_schema_id": approval["owner_mapping_schema_id"],
		"owner_mapping":           approval["owner_mapping"],
	}
}

func postImportUpload(t testing.TB, serverURL string, login flowtest.LoginResult, metadata string, file string, filename string, fileFirst bool) *http.Response {
	t.Helper()
	return postImportUploadBytes(t, serverURL, login, metadata, []byte(file), filename, imports.MediaTypeCSV, fileFirst)
}

func postImportUploadBytes(t testing.TB, serverURL string, login flowtest.LoginResult, metadata string, file []byte, filename string, contentType string, fileFirst bool) *http.Response {
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

func doImportJSON(t testing.TB, serverURL string, login flowtest.LoginResult, method string, path string, body any) *http.Response {
	t.Helper()
	return httptestx.DoJSON(
		t,
		method,
		serverURL+path,
		body,
		httptestx.WithCookies(login.SessionCookie, login.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
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

func multipleSheetXLSX(t testing.TB) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	writeZipText(t, writer, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/worksheets/sheet2.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`)
	writeZipText(t, writer, "xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Sheet1" sheetId="1" r:id="rId1"/>
    <sheet name="Sheet2" sheetId="2" r:id="rId2"/>
  </sheets>
</workbook>`)
	writeZipText(t, writer, "xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
</Relationships>`)
	writeZipText(t, writer, "xl/worksheets/sheet1.xml", `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="inlineStr"><is><t>host</t></is></c><c r="B1" t="inlineStr"><is><t>summary</t></is></c></row>
    <row r="2"><c r="A2" t="inlineStr"><is><t>host-1</t></is></c><c r="B2" t="inlineStr"><is><t>Alpha summary</t></is></c></row>
    <row r="3"><c r="A3" t="inlineStr"><is><t>host-2</t></is></c><c r="B3" t="inlineStr"><is><t>Beta summary</t></is></c></row>
  </sheetData>
</worksheet>`)
	writeZipText(t, writer, "xl/worksheets/sheet2.xml", `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="inlineStr"><is><t>indicator</t></is></c><c r="B1" t="inlineStr"><is><t>type</t></is></c></row>
    <row r="2"><c r="A2" t="inlineStr"><is><t>203.0.113.42</t></is></c><c r="B2" t="inlineStr"><is><t>ipv4_addr</t></is></c></row>
  </sheetData>
</worksheet>`)
	if err := writer.Close(); err != nil {
		t.Fatalf("close XLSX zip: %v", err)
	}
	return buffer.Bytes()
}

func requireImportProof(t testing.TB, db *sql.DB, jobID string, operationKind string) {
	t.Helper()
	var ownerProfileID string
	var actualOperationKind string
	var finalCommitID string
	if err := db.QueryRow(`
SELECT owner_profile_id, operation_kind, final_commit_id
  FROM extension_job_commit_proofs
 WHERE job_id::text = $1
`, jobID).Scan(&ownerProfileID, &actualOperationKind, &finalCommitID); err != nil {
		t.Fatalf("read import proof for job %s: %v", jobID, err)
	}
	if ownerProfileID != imports.ProfileID ||
		actualOperationKind != operationKind ||
		finalCommitID == "" {
		t.Fatalf(
			"unexpected import proof: owner=%q operation=%q commit=%q",
			ownerProfileID,
			actualOperationKind,
			finalCommitID,
		)
	}
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
		Sessions:         dbassert.CountSQL(t, db, `SELECT COUNT(*) FROM import_sessions`),
		Units:            dbassert.CountSQL(t, db, `SELECT COUNT(*) FROM import_units`),
		SourceStreams:    dbassert.CountSQL(t, db, `SELECT COUNT(*) FROM import_source_streams`),
		Jobs:             dbassert.CountSQL(t, db, `SELECT COUNT(*) FROM jobs`),
		RouteIdempotency: dbassert.CountSQL(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key LIKE 'imports.%'`),
	}
	if got != want {
		t.Fatalf("unexpected import durable counts: got %+v want %+v", got, want)
	}
}

func waitForSQLCount(t testing.TB, db *sql.DB, timeout time.Duration, want int, query string, args ...any) int {
	t.Helper()
	deadline := time.Now().Add(timeout)
	got := 0
	for {
		got = dbassert.CountSQL(t, db, query, args...)
		if got == want || time.Now().After(deadline) {
			return got
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func importApplyState(t testing.TB, db *sql.DB, sessionID string) map[string]string {
	t.Helper()
	var sessionStatus string
	var applyJobID string
	var jobStatus string
	var resultSummary string
	var errorSummary string
	if err := db.QueryRowContext(context.Background(), `
SELECT s.session_status,
       COALESCE(s.apply_job_id::text, ''),
       COALESCE(j.status, ''),
       COALESCE(j.result_summary_json::text, ''),
       COALESCE(j.error_summary_json::text, '')
  FROM import_sessions s
  LEFT JOIN jobs j ON j.job_id = s.apply_job_id
 WHERE s.import_session_id::text = $1
`, sessionID).Scan(
		&sessionStatus,
		&applyJobID,
		&jobStatus,
		&resultSummary,
		&errorSummary,
	); err != nil {
		t.Fatalf("read import apply session state: %v", err)
	}
	var unitStatuses string
	if err := db.QueryRowContext(context.Background(), `
SELECT COALESCE(string_agg(unit_status, ',' ORDER BY discovery_sequence), '')
  FROM import_units
 WHERE import_session_id::text = $1
`, sessionID).Scan(&unitStatuses); err != nil {
		t.Fatalf("read import apply unit state: %v", err)
	}
	var outcomeStatuses string
	if err := db.QueryRowContext(context.Background(), `
SELECT COALESCE(string_agg(outcome_status, ',' ORDER BY discovery_sequence), '')
  FROM import_unit_apply_outcomes
 WHERE import_session_id::text = $1
`, sessionID).Scan(&outcomeStatuses); err != nil {
		t.Fatalf("read import apply outcome state: %v", err)
	}
	return map[string]string{
		"session_status":   sessionStatus,
		"apply_job_id":     applyJobID,
		"job_status":       jobStatus,
		"result_summary":   resultSummary,
		"error_summary":    errorSummary,
		"unit_statuses":    unitStatuses,
		"outcome_statuses": outcomeStatuses,
	}
}
