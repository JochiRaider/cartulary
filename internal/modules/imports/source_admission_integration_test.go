package imports_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

const sourceAdmissionHeader = "Source IP Address,Destination IP Address,Source Port,Destination Port,Protocol,Bytes,Packets,Flow Start Time,Flow End Time,Input Interface,Output Interface\n"
const sourceAdmissionRow = "192.0.2.10,198.51.100.1,443,51515,TCP,1234,12,2026-07-10T12:00:00Z,2026-07-10T12:00:05Z,Gi0/1,Gi0/2\n"

func sourceAdmissionHarness(t *testing.T, overrides map[string]string) (*appsupport.ServerHarness, authcookietest.AuthCookies, string) {
	t.Helper()
	env := map[string]string{
		"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}
	for k, v := range overrides {
		env[k] = v
	}
	runtime := appsupport.StartRuntime(t)
	h := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "target-neutral-source", Env: env, TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "source-admission-incident", "incident_key": "SOURCE-ADMISSION", "title": "Source admission"})
	return h, login, incident["incident_id"].(string)
}

func TestNetworkFlowPublicSourceAdmission(t *testing.T) {
	h, login, incident := sourceAdmissionHarness(t, map[string]string{
		"CARTULARY__LIMITS__IMPORTS__MAX_ROWS":             "1",
		"CARTULARY__LIMITS__IMPORTS__MAX_COLUMNS":          "1",
		"CARTULARY__LIMITS__IMPORTS__MAX_CELLS":            "1",
		"CARTULARY__LIMITS__IMPORTS__MAX_CSV_SOURCE_BYTES": "32768",
	})
	for _, tc := range []struct {
		name, body                       string
		previewOK, applyOK, genericFails bool
		accepted                         int
	}{
		{"literal_quotes", strings.Replace(sourceAdmissionRow, "Gi0/1", "Gi\"0/1", 1), true, true, true, 1},
		{"generic_limits_do_not_gate_csv", strings.Repeat(sourceAdmissionRow, 2), true, true, false, 2},
		{"early_bad_quote", "\"unclosed\n", false, false, true, 0},
		{"early_bad_encoding", strings.Replace(sourceAdmissionRow, "Gi0/1", "\xff", 1), false, false, false, 0},
		{"late_bad_quote", strings.Repeat(sourceAdmissionRow, 50) + "\"unclosed\n", true, false, true, 0},
		{"late_bad_encoding", strings.Repeat(sourceAdmissionRow, 50) + strings.Replace(sourceAdmissionRow, "Gi0/1", "\xff", 1), true, false, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "upload-"+tc.name, sourceAdmissionHeader+tc.body, tc.name+".csv")
			base := "/api/v1/import-sessions/" + session + "/units/" + unit
			if tc.genericFails {
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, h.Server.HTTP.URL+base+"/preview", nil, httptestx.WithCookies(login.Session))
				httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "import_source_unsupported")
			} else if tc.name == "generic_limits_do_not_gate_csv" {
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, h.Server.HTTP.URL+base+"/preview", nil, httptestx.WithCookies(login.Session))
				httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
			}
			assertPublicAnalyticalSource(t, h, login, session, unit, tc.previewOK, tc.applyOK, tc.accepted)
		})
	}
	t.Run("owner_column_limit", func(t *testing.T) {
		session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "owner-column-limit", strings.Repeat("column,", 256)+"last\n", "columns.csv")
		resp := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+session+"/units/"+unit+"/source-inspection", map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"})
		httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_import_request")
	})
	t.Run("byte_overflow_has_no_durable_admission", func(t *testing.T) {
		counts := []int{}
		for _, table := range []string{"import_sessions", "jobs", "route_idempotency"} {
			counts = append(counts, dbassert.CountSQL(t, h.DB, "SELECT count(*) FROM "+table))
		}
		metadata := fmt.Sprintf(`{"client_txn_id":"byte-overflow","incident_id":%q}`, incident)
		resp := postImportUpload(t, h.Server.HTTP.URL, login, metadata, strings.Repeat("x", 32769), "overflow.csv", false)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusRequestEntityTooLarge, "import_source_rejected")
		for i, table := range []string{"import_sessions", "jobs", "route_idempotency"} {
			if got := dbassert.CountSQL(t, h.DB, "SELECT count(*) FROM "+table); got != counts[i] {
				t.Fatalf("overflow persisted %s", table)
			}
		}
	})
}

// The real default limit is intentional: a scaled fixture cannot demonstrate
// admission beyond Workbook's former 100000-row ceiling. Its catalog row reserves
// the backend_capacity_isolated profile; the apply observation is bounded below.
func TestNetworkFlowPublicDefaultLimit100001Rows(t *testing.T) {
	h, login, incident := sourceAdmissionHarness(t, nil)
	session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "large-source", sourceAdmissionHeader+strings.Repeat(sourceAdmissionRow, 100001), "large.csv")
	assertPublicAnalyticalSource(t, h, login, session, unit, true, true, 100001)
}

func assertPublicAnalyticalSource(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, session, unit string, previewOK, applyOK bool, accepted int) {
	t.Helper()
	base := "/api/v1/import-sessions/" + session + "/units/" + unit
	inspection := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/source-inspection", map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"})
	owner := httptestx.RequireSuccessEnvelope(t, inspection, http.StatusOK)["data"].(map[string]any)["owner_result"].(map[string]any)
	preview := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/mapping-preview", networkFlowMappingPreviewPayload())
	if !previewOK {
		httptestx.RequireErrorEnvelope(t, preview, http.StatusBadRequest, "invalid_import_request")
		return
	}
	previewResult := httptestx.RequireSuccessEnvelope(t, preview, http.StatusOK)["data"].(map[string]any)["owner_result"].(map[string]any)
	if previewResult["source_content_sha256"] != owner["source_content_sha256"] {
		t.Fatal("preview changed inspected source identity")
	}
	mapping := networkFlowMappingPayload("map-" + unit)
	echoes := mapping["source_columns"].([]map[string]any)
	for i, raw := range owner["source_columns"].([]any) {
		column := raw.(map[string]any)
		echoes[i]["source_header_text"] = column["raw_header_text"]
		echoes[i]["source_column_ordinal"] = column["source_column_ordinal"]
	}
	resp := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPut, base+"/mapping", mapping)
	approved := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	if approved["mapping_fingerprint"] != previewResult["mapping_fingerprint"] {
		t.Fatal("approval differs from preview")
	}
	resp = doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/select", map[string]any{"client_txn_id": "select-" + unit})
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	applyBody := map[string]any{"client_txn_id": "apply-" + unit}
	resp = doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+session+"/apply", applyBody)
	jobID := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
	deadline := time.Now().Add(2 * time.Minute)
	var job map[string]any
	for {
		resp = httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, h.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.Session))
		job = httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		if job["status"] == "succeeded" || job["status"] == "failed" || job["status"] == "canceled" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("apply observation exceeded declared large-input bound: %v", job["status"])
		}
		time.Sleep(50 * time.Millisecond)
	}
	wantStatus := "failed"
	if applyOK {
		wantStatus = "succeeded"
	}
	if job["status"] != wantStatus {
		t.Fatalf("apply status=%v want=%s result=%v", job["status"], wantStatus, job["result_summary"])
	}
	wantTables := 0
	if applyOK {
		wantTables = 1
	}
	if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM network_flow_tables WHERE source_import_session_id::text=$1`, session); got != wantTables {
		t.Fatalf("atomic table count=%d want=%d", got, wantTables)
	}
	if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM network_flow_rows r JOIN network_flow_tables t USING(network_flow_table_id) WHERE t.source_import_session_id::text=$1`, session); got != accepted {
		t.Fatalf("accepted rows=%d want=%d", got, accepted)
	}
	var materialized bool
	if err := h.DB.QueryRowContext(context.Background(), `SELECT source_rows_json IS NOT NULL FROM import_units WHERE import_unit_id::text=$1`, unit).Scan(&materialized); err != nil || materialized {
		t.Fatalf("analytical workflow materialized generic content: %v", err)
	}
	resp = doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+session+"/apply", applyBody)
	replayed := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	if replayed["job_id"] != jobID {
		t.Fatal("apply replay replaced job identity")
	}
}

func TestAnalyticalReadOnlyPOSTAuthentication(t *testing.T) {
	h, login, incident := sourceAdmissionHarness(t, nil)
	session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "read-only-auth", sourceAdmissionHeader+sourceAdmissionRow, "read-only.csv")
	before := importsSecurityState(t, h.DB)
	for _, suffix := range []string{"source-inspection", "mapping-preview"} {
		body := map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"}
		if suffix == "mapping-preview" {
			body = networkFlowMappingPreviewPayload()
		}
		path := h.Server.HTTP.URL + "/api/v1/import-sessions/" + session + "/units/" + unit + "/" + suffix
		unauthorized := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, path, body)
		httptestx.RequireErrorEnvelope(t, unauthorized, http.StatusUnauthorized, "session_required")
		cookie := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, path, body, httptestx.WithCookies(login.Session))
		httptestx.RequireSuccessEnvelope(t, cookie, http.StatusOK)
		bearer := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, path, body, httptestx.WithHeader("Authorization", "Bearer "+login.Session.Value))
		httptestx.RequireSuccessEnvelope(t, bearer, http.StatusOK)
	}
	after := importsSecurityState(t, h.DB)
	for key, want := range before {
		if after[key] != want {
			t.Fatalf("read-only POST changed %s", key)
		}
	}
}

func TestGenericCSVRaggedPublicLifecycle(t *testing.T) {
	h, login, incident := sourceAdmissionHarness(t, nil)
	session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, "ragged-upload", "summary,optional,unmapped\nshort\nexplicit,,\n", "ragged.csv")
	base := "/api/v1/import-sessions/" + session + "/units/" + unit
	preview := func() {
		t.Helper()
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, h.Server.HTTP.URL+base+"/preview", nil, httptestx.WithCookies(login.Session))
		result := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		rows := result["preview_rows"].([]any)
		if len(result["columns"].([]any)) != 3 || len(rows) != 2 {
			t.Fatal("wrong ragged preview geometry")
		}
		for i, want := range []int{1, 3} {
			if len(rows[i].(map[string]any)["cells"].([]any)) != want {
				t.Fatalf("row %d was padded", i)
			}
		}
	}
	preview()
	if dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM import_units WHERE import_unit_id::text=$1 AND source_rows_json IS NULL AND approved_mapping_json IS NULL`, unit) != 1 {
		t.Fatal("preview wrote projection")
	}
	for range 2 {
		approveTimelineImportMapping(t, h.Server.HTTP.URL, login, session, unit, []string{"summary", "optional", "unmapped"}, "ragged")
	}
	preview()
	var count int
	if err := h.DB.QueryRowContext(t.Context(), `SELECT sum(jsonb_array_length(r->'cells')) FROM import_units u, LATERAL jsonb_array_elements(u.source_rows_json) r WHERE import_unit_id::text=$1`, unit).Scan(&count); err != nil || count != 4 {
		t.Fatalf("materialized cell count %d: %v", count, err)
	}
	selected := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, base+"/select", map[string]any{"client_txn_id": "ragged-select"})
	httptestx.RequireSuccessEnvelope(t, selected, http.StatusOK)
	var jobID string
	for range 2 {
		response := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+session+"/apply", map[string]any{"client_txn_id": "ragged-apply"})
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		id := job["job_id"].(string)
		if jobID != "" && id != jobID {
			t.Fatal("apply replay created another job")
		}
		jobID = id
		terminal := waitImportJobTerminal(t, h.Server.HTTP.URL, login, id)
		if terminal["status"] != "succeeded" {
			t.Fatalf("ragged apply: %#v", terminal)
		}
	}
	if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM timeline_events WHERE incident_id::text=$1`, incident); got != 2 {
		t.Fatalf("owner rows %d", got)
	}
	preview()
}
