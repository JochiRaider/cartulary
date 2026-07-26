package imports_test

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestExtensionImportUploadEarlyFailCreatesNoDurableRows(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "extension_profile-import-early-fail")
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
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "extension_profile-import-metadata-non-object")
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
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "extension_profile-import-replay")
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

	jobResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+firstJobID, nil, httptestx.WithCookies(adminLogin.SessionCookie))
	job := httptestx.RequireSuccessEnvelope(t, jobResp, http.StatusOK)["data"].(map[string]any)
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
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "extension_profile-import-xlsx-discovery")
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
	jobResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+uploadJob["job_id"].(string), nil, httptestx.WithCookies(adminLogin.SessionCookie))
	job := httptestx.RequireSuccessEnvelope(t, jobResp, http.StatusOK)["data"].(map[string]any)
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

func TestMappingSelectApplyCreatesTimelineRows_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "extension_profile-import-apply")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-import-apply-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-APPLY",
		"title":         "Enterprise integration import apply",
	})
	incidentID := incident["incident_id"].(string)
	hostRecordID := createImportAutoResolutionCandidateHost(t, harness.Server.HTTP.URL, adminLogin, incidentID)
	metadata := `{"client_txn_id":"txn-extension_profile-import-apply-upload","incident_id":"` + incidentID + `"}`
	csv := "host,summary,source_note\nhost-1, Alpha summary ,raw-a\nhost-2,Beta summary,raw-b\n"

	uploadResp := postImportUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, csv, "apply.csv", false)
	uploadJob := httptestx.RequireSuccessEnvelope(t, uploadResp, http.StatusAccepted)["data"].(map[string]any)
	jobResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+uploadJob["job_id"].(string), nil, httptestx.WithCookies(adminLogin.SessionCookie))
	job := httptestx.RequireSuccessEnvelope(t, jobResp, http.StatusOK)["data"].(map[string]any)
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
				"source_header_text":    "summary",
				"field_key":             "timeline.activity_synopsis_text",
				"entity_binding_mode":   nil,
				"transform_id":          "trim_v1",
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			},
			{
				"source_column_ordinal": 3,
				"source_header_text":    "source_note",
				"field_key":             nil,
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
	applyJobResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+applyJob["job_id"].(string), nil, httptestx.WithCookies(adminLogin.SessionCookie))
	appliedJob := httptestx.RequireSuccessEnvelope(t, applyJobResp, http.StatusOK)["data"].(map[string]any)
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
	requireTimelineHostMentionUnresolved(t, rowsBySummary["Alpha summary"], "host-1")
	requireTimelineHostMentionUnresolved(t, rowsBySummary["Beta summary"], "host-2")
	requireImportedHostMentionNotAutoResolved(t, harness.DB, incidentID, recordIDsBySummary["Alpha summary"], hostRecordID, "host-1")
	requireTimelineImportProvenance(t, harness.DB, recordIDsBySummary["Alpha summary"], sessionID, unitID, "source_note", "raw-a", 2, 3, "A1:C3")
	requireTimelineImportProvenance(t, harness.DB, recordIDsBySummary["Beta summary"], sessionID, unitID, "source_note", "raw-b", 3, 3, "A1:C3")

	duplicateApply := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-extension_profile-import-apply-second"})
	errBody := httptestx.RequireErrorEnvelope(t, duplicateApply, http.StatusConflict, "import_apply_blocked")
	if errBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "duplicate_apply_blocked" {
		t.Fatalf("unexpected duplicate apply error: %#v", errBody)
	}
}

func TestTargetRegistryAndEntityOwnerFacade_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "extension_profile-import-target-registry-host")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-import-target-host-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-HOST",
		"title":         "Enterprise integration host import",
	})
	incidentID := incident["incident_id"].(string)

	sessionID, unitID := startCSVImportSession(t, harness.Server.HTTP.URL, adminLogin, incidentID, "txn-extension_profile-import-target-host-upload", "hostname\nimported-host-1\n", "hosts.csv")
	unknownTargetMapping := map[string]any{
		"client_txn_id":         "txn-extension_profile-import-target-unknown-mapping",
		"target_view_schema_id": "cartulary.view.unknown.v1",
		"header_row_ref":        1,
		"data_start_row_ref":    2,
		"unknown_column_policy": "reject_if_unmapped",
		"source_columns": []map[string]any{{
			"source_column_ordinal": 1,
			"source_header_text":    "hostname",
			"field_key":             "host.hostname",
			"entity_binding_mode":   "entity_origin",
			"transform_id":          nil,
			"transform_options":     map[string]any{},
			"empty_value_policy":    "omit_field",
		}},
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
		"source_columns": []map[string]any{{
			"source_column_ordinal": 1,
			"source_header_text":    "hostname",
			"field_key":             nil,
			"entity_binding_mode":   nil,
			"transform_id":          nil,
			"transform_options":     map[string]any{},
			"empty_value_policy":    "omit_field",
		}},
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
	applyJobResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+applyJob["job_id"].(string), nil, httptestx.WithCookies(adminLogin.SessionCookie))
	appliedJob := httptestx.RequireSuccessEnvelope(t, applyJobResp, http.StatusOK)["data"].(map[string]any)
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
}

func TestNetworkFlowImportMappingAndApplyCreatesOneAtomicTable(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "network-flow-import-apply")
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
   SET role = 'viewer', updated_at = now(), updated_by_user_id = $2
 WHERE incident_id::text = $1 AND user_id::text = $2
`, incidentID, adminUserID); err != nil {
		t.Fatalf("demote mapping preview actor: %v", err)
	}
	viewerPreviewResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping-preview", previewPayload)
	httptestx.RequireErrorEnvelope(t, viewerPreviewResp, http.StatusForbidden, "authorization_denied")
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'admin', updated_at = now(), updated_by_user_id = $2
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
	applyJobResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+applyJob["job_id"].(string), nil, httptestx.WithCookies(adminLogin.SessionCookie))
	appliedJob := httptestx.RequireSuccessEnvelope(t, applyJobResp, http.StatusOK)["data"].(map[string]any)
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
}

func TestEvidenceImportUsesOwnerFacadeAndJournal_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "extension_profile-import-evidence-owner-facade")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-import-evidence-owner-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-EVIDENCE",
		"title":         "Enterprise integration evidence owner facade",
	})
	incidentID := incident["incident_id"].(string)

	sessionID, unitID := startCSVImportSession(t, harness.Server.HTTP.URL, adminLogin, incidentID, "txn-extension_profile-import-evidence-owner-upload", "title\nEvidence from import\n", "evidence.csv")
	mapping := map[string]any{
		"client_txn_id":         "txn-extension_profile-import-evidence-owner-mapping",
		"target_view_schema_id": "cartulary.view.evidence.v1",
		"header_row_ref":        1,
		"data_start_row_ref":    2,
		"unknown_column_policy": "reject_if_unmapped",
		"source_columns": []map[string]any{{
			"source_column_ordinal": 1,
			"source_header_text":    "title",
			"field_key":             "evidence.title",
			"entity_binding_mode":   nil,
			"transform_id":          nil,
			"transform_options":     map[string]any{},
			"empty_value_policy":    "omit_field",
		}},
	}
	mappingResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPut, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/mapping", mapping)
	httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)
	selectResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/units/"+unitID+"/select", map[string]any{"client_txn_id": "txn-extension_profile-import-evidence-owner-select"})
	httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)

	applyResp := doImportJSON(t, harness.Server.HTTP.URL, adminLogin, http.MethodPost, "/api/v1/import-sessions/"+sessionID+"/apply", map[string]any{"client_txn_id": "txn-extension_profile-import-evidence-owner-apply"})
	applyJob := httptestx.RequireSuccessEnvelope(t, applyResp, http.StatusAccepted)["data"].(map[string]any)
	applyJobResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+applyJob["job_id"].(string), nil, httptestx.WithCookies(adminLogin.SessionCookie))
	appliedJob := httptestx.RequireSuccessEnvelope(t, applyJobResp, http.StatusOK)["data"].(map[string]any)
	if appliedJob["status"] != "succeeded" || appliedJob["result_summary"].(map[string]any)["code"] != "import_session_applied" {
		t.Fatalf("unexpected evidence apply job: %#v", appliedJob)
	}

	queryResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.evidence.v1/query", map[string]any{}, httptestx.WithCookies(adminLogin.SessionCookie))
	rows := httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)["data"].(map[string]any)["rows"].([]any)
	if len(rows) != 1 {
		t.Fatalf("expected one imported evidence row, got %#v", rows)
	}
	cells := rows[0].(map[string]any)["cells"].(map[string]any)
	if got := cells["evidence.title"].(map[string]any)["value"]; got != "Evidence from import" {
		t.Fatalf("unexpected imported evidence title: %#v row=%#v", got, rows[0])
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM import_apply_journal WHERE import_session_id::text = $1`, sessionID); got != 1 {
		t.Fatalf("expected one import apply journal row, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE source = 'imports.apply' AND client_txn_id = $1`, "import:"+sessionID+":"+unitID+":txn-extension_profile-import-evidence-owner-apply"); got != 1 {
		t.Fatalf("expected one unit-level import change set, got %d", got)
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

func startCSVImportSession(t testing.TB, serverURL string, login flowtest.LoginResult, incidentID string, clientTxnID string, csv string, filename string) (string, string) {
	t.Helper()

	metadata := `{"client_txn_id":"` + clientTxnID + `","incident_id":"` + incidentID + `"}`
	uploadResp := postImportUpload(t, serverURL, login, metadata, csv, filename, false)
	uploadJob := httptestx.RequireSuccessEnvelope(t, uploadResp, http.StatusAccepted)["data"].(map[string]any)
	jobResp := httptestx.DoJSON(t, http.MethodGet, serverURL+"/api/v1/jobs/"+uploadJob["job_id"].(string), nil, httptestx.WithCookies(login.SessionCookie))
	job := httptestx.RequireSuccessEnvelope(t, jobResp, http.StatusOK)["data"].(map[string]any)
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

func requireTimelineHostMentionUnresolved(t testing.TB, row map[string]any, rawText string) {
	t.Helper()

	cells := row["cells"].(map[string]any)
	hostRefs := cells["timeline.host_refs"].(map[string]any)["value"].(map[string]any)
	items := hostRefs["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected one imported host mention, got %#v", hostRefs)
	}
	item := items[0].(map[string]any)
	if item["item_kind"] != "unresolved_mention" || item["raw_text"] != rawText {
		t.Fatalf("expected unresolved imported host mention %q, got %#v", rawText, item)
	}
	if _, ok := item["resolved_record_id"]; ok {
		t.Fatalf("imported host mention must not be resolved: %#v", item)
	}
	if _, ok := item["resolution_method"]; ok {
		t.Fatalf("imported host mention must not expose resolution method: %#v", item)
	}
}

func requireImportedHostMentionNotAutoResolved(t testing.TB, db *sql.DB, incidentID string, timelineRecordID string, hostRecordID string, rawText string) {
	t.Helper()

	if got := dbassert.CountSQL(t, db, `
SELECT COUNT(*)
  FROM entity_mentions
 WHERE source_record_id::text = $1
   AND source_field_key = 'timeline.host_refs'
   AND raw_text = $2
   AND resolution_status = 'unresolved'
   AND resolved_record_id IS NULL
   AND resolution_method IS NULL
`, timelineRecordID, rawText); got != 1 {
		t.Fatalf("imported host token must remain one unresolved mention, got %d", got)
	}
	if got := dbassert.CountSQL(t, db, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND dst_record_id::text = $3
   AND link_type = 'observed_on_host'
   AND deleted_at IS NULL
`, incidentID, timelineRecordID, hostRecordID); got != 0 {
		t.Fatalf("imported exact alias must not create active host link, got %d", got)
	}
	if got := dbassert.CountSQL(t, db, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id::text = $1
   AND src_record_id::text = $2
   AND dst_record_id::text = $3
   AND provenance = 'auto_match'
   AND deleted_at IS NULL
`, incidentID, timelineRecordID, hostRecordID); got != 0 {
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
