package imports_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/server"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/irfixtures"
)

// The bundle owner qualifies the complete Network Flow participant through real
// uploads and public transfer operations, with independent destination key rings.
func TestNetworkFlowIncidentTransferThroughPublicOperations_Integration(t *testing.T) {
	fixtureRoot, err := filepath.Abs("../../../fixtures")
	if err != nil {
		t.Fatal(err)
	}
	corpus := irfixtures.ForTest(t, fixtureRoot)
	runtime := appsupport.StartRuntime(t)
	flowEnvironment := func(key byte, claimed bool) map[string]string {
		env := map[string]string{"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED": fmt.Sprint(claimed), "CARTULARY_AUTH_MASTER_KEY": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{key + 2}, 32))}
		if claimed {
			for name, value := range map[string]string{
				"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
				"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{key}, 32)),
				"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{key + 1}, 32)),
			} {
				env[name] = value
			}
		}
		return env
	}
	start := func(prefix string, key byte, claimed bool) *appsupport.ServerHarness {
		return runtime.StartServer(t, appsupport.ServerOptions{Prefix: prefix, TestRouteMode: httptestx.TestRouteModeDisabled, Env: flowEnvironment(key, claimed), ConfigureRuntime: func(options *server.Options) { options.ReportingTemplatePacks = irReportTemplates(t) }})
	}
	a, b := start("nf-portable-a", 11, true), start("nf-portable-b", 31, true)
	loginA, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, a.Server.HTTP.URL)
	loginB, secretB := irBootstrapRecoveryAdmin(t, b.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, a.Server, loginA, map[string]any{"client_txn_id": "nf-transfer-incident", "incident_key": "IR-NF-TRANSFER", "title": "Portable Network Flow investigation"})
	id := incident["incident_id"].(string)
	irImportCoreCorpus(t, corpus, a, loginA, id)
	if t.Failed() {
		t.FailNow()
	}
	root := "/api/v1/incidents/" + id + "/network-flow"
	call := func(base string, login authcookietest.AuthCookies, method, path string, body any, status int) map[string]any {
		t.Helper()
		response := doImportJSON(t, base, login, method, path, body)
		if response.StatusCode != status {
			defer response.Body.Close()
			t.Fatalf("%s %s: status %d, want %d", method, path, response.StatusCode, status)
		}
		return httptestx.RequireSuccessEnvelope(t, response, status)["data"].(map[string]any)
	}
	var cases []struct {
		Mapping map[string]any `json:"mapping"`
	}
	mappingBytes, err := os.ReadFile(filepath.Join(fixtureRoot, "network-flow", "NF-FIX-029-cisco-sna-flow-search", "source", "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(mappingBytes, &cases); err != nil {
		t.Fatal(err)
	}
	var tableIDs []string
	var selected map[string]any
	var graphPath string
	for index, name := range []string{"day1", "day2", "overlap"} {
		source, err := corpus.Read("network-flow-orientation/v1/" + name + ".csv")
		if err != nil {
			t.Fatal(err)
		}
		session, unit := startCSVImportSession(t, a.Server.HTTP.URL, loginA, id, name+"-upload", string(source), name+".csv")
		unitPath := "/api/v1/import-sessions/" + session + "/units/" + unit
		inspection := call(a.Server.HTTP.URL, loginA, http.MethodPost, unitPath+"/source-inspection", map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"}, http.StatusOK)
		columns := []map[string]any{}
		for _, raw := range inspection["owner_result"].(map[string]any)["source_columns"].([]any) {
			column := raw.(map[string]any)
			columns = append(columns, map[string]any{"source_column_ordinal": column["source_column_ordinal"], "source_header_text": column["raw_header_text"], "field_key": nil, "entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"})
		}
		mapping := map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity", "owner_mapping_schema_id": "cartulary.network_flow.mapping_candidate.v2", "owner_mapping": cases[0].Mapping}
		call(a.Server.HTTP.URL, loginA, http.MethodPost, unitPath+"/mapping-preview", mapping, http.StatusOK)
		mapping["client_txn_id"] = name + "-approve"
		mapping["header_row_ref"] = 1
		mapping["data_start_row_ref"] = 2
		mapping["source_columns"] = columns
		call(a.Server.HTTP.URL, loginA, http.MethodPut, unitPath+"/mapping", mapping, http.StatusOK)
		call(a.Server.HTTP.URL, loginA, http.MethodPost, unitPath+"/select", map[string]any{"client_txn_id": name + "-select"}, http.StatusOK)
		job := call(a.Server.HTTP.URL, loginA, http.MethodPost, "/api/v1/import-sessions/"+session+"/apply", map[string]any{"client_txn_id": name + "-apply"}, http.StatusAccepted)
		terminal := waitIRApply(t, a.Server.HTTP.URL, loginA, job["job_id"].(string))
		if terminal["status"] != "succeeded" {
			t.Fatalf("flow import %s failed: %#v", name, terminal)
		}
		tables := call(a.Server.HTTP.URL, loginA, http.MethodGet, root+"/tables", nil, http.StatusOK)["tables"].([]any)
		if len(tables) != index+1 {
			t.Fatal("flow import lost table multiplicity")
		}
		for _, raw := range tables {
			table := raw.(map[string]any)
			if table["source_filename_display"] == name+".csv" {
				tableIDs = append(tableIDs, table["network_flow_table_id"].(string))
				want := 2000.
				if name == "overlap" {
					want = 500
				}
				if table["row_count_accepted"] != want {
					t.Fatal("flow count changed")
				}
			}
		}
		if len(tableIDs) != index+1 {
			t.Fatal("missing imported table")
		}
		if index == 0 {
			created := call(a.Server.HTTP.URL, loginA, http.MethodPost, root+"/graph-views", map[string]any{"schema_id": "cartulary.network_flow.graph_view_create_request.v5", "refresh_scope": "all_active_tables", "client_txn_id": "selected-day1", "display_name": "Day 1 selected history", "semantic_query": map[string]any{"schema_id": "cartulary.network_flow.graph_semantic_query.v3", "selected_table_ids": tableIDs, "filters": []any{}, "time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}}}, http.StatusAccepted)
			terminal := waitIRApply(t, a.Server.HTTP.URL, loginA, created["job"].(map[string]any)["job_id"].(string))
			if terminal["status"] != "succeeded" {
				t.Fatalf("graph materialization failed: %#v", terminal)
			}
			graphPath = root + "/graph-views/" + created["graph_view"].(map[string]any)["graph_view_id"].(string)
			selected = call(a.Server.HTTP.URL, loginA, http.MethodGet, graphPath+"/result", nil, http.StatusOK)
		}
	}
	clientA := irPublicClient{t: t, origin: a.Server.HTTP.URL, incident: id, login: loginA}
	clientB := irPublicClient{t: t, origin: b.Server.HTTP.URL, incident: id, login: loginB}
	investigation := irCreateInvestigation(clientA, corpus)
	irBindFlowIndicator(clientA, tableIDs[0], investigation.indicatorID)
	report := irCreateReport(clientA, selected)
	clientA.request(http.MethodPatch, root+"/tables/"+tableIDs[0], map[string]any{"client_txn_id": "ir-post-report-rename", "base_table_version": 1, "display_name": "Day 1 renamed after report"}, http.StatusOK)
	for _, note := range irQueryAll(t, clientA.origin, loginA, id, "cartulary.view.notes.v1") {
		if note["record_id"] == investigation.noteID {
			clientA.patch("cartulary.view.notes.v1", "ir-post-report-note", note, irValue("note.title", "SYNTHETIC IR FIXTURE: live note changed after report"))
		}
	}
	bundle := readPublicBundle(t, a.Server.HTTP.URL, loginA, id, "export-network-flow", "snapshots")
	irVerifyReport(clientA, report, bundle, true)
	job := uploadPublicSourceBundle(t, b.Server.HTTP.URL, loginB, bundle, "transfer-network-flow")
	terminal := waitIRApply(t, b.Server.HTTP.URL, loginB, job["job_id"].(string))
	if terminal["status"] != "succeeded" {
		t.Fatalf("Network Flow transfer failed: %#v", terminal)
	}
	replay := uploadPublicSourceBundle(t, b.Server.HTTP.URL, loginB, bundle, "transfer-network-flow")
	if replay["job_id"] != job["job_id"] {
		t.Fatal("transfer replay duplicated effects")
	}
	verifyDestination := func(destination irPublicClient, label string) {
		irVerifyInvestigationTransfer(clientA, destination, investigation)
		for _, source := range corpus.Manifest.Sources {
			view := map[string]string{"hosts.csv": "cartulary.view.hosts.v1", "identities.csv": "cartulary.view.identities.v1", "timeline.csv": "cartulary.view.timeline.v2"}[source.Path]
			if view == "" {
				continue
			}
			left, right := irQueryAll(t, a.Server.HTTP.URL, loginA, id, view), irQueryAll(t, destination.origin, destination.login, id, view)
			if len(left) != source.Records || len(right) != source.Records {
				t.Fatalf("transferred %s count: %d/%d", source.Path, len(left), len(right))
			}
			byID := map[string]map[string]any{}
			for _, row := range right {
				byID[row["record_id"].(string)] = row
			}
			for _, row := range left {
				other := byID[row["record_id"].(string)]
				if other == nil || row["row_version"] != other["row_version"] || !reflect.DeepEqual(row["cells"], other["cells"]) {
					t.Fatalf("transferred %s row content or version changed", source.Path)
				}
			}
		}
		if sources := irCommittedSources(t, destination.origin, destination.login, id); len(sources) != 6 {
			t.Fatalf("complete incident retained %d sources, want 6", len(sources))
		}
		imported := call(destination.origin, destination.login, http.MethodGet, graphPath+"/result", nil, http.StatusOK)
		leftGraph, rightGraph := selected["graph_view"].(map[string]any), imported["graph_view"].(map[string]any)
		if len(leftGraph) != len(rightGraph) {
			t.Fatal("graph resource shape changed")
		}
		for field, value := range leftGraph {
			switch field {
			case "latest_job_id", "last_failure_code", "last_failed_at":
				if rightGraph[field] != nil {
					t.Fatalf("operational graph state %s was transferred", field)
				}
			default:
				if !reflect.DeepEqual(value, rightGraph[field]) {
					t.Fatalf("selected graph field %s changed", field)
				}
			}
		}
		if !reflect.DeepEqual(selected["result"].(map[string]any)["graph_projection_result"], imported["result"].(map[string]any)["graph_projection_result"]) {
			t.Fatal("selected day 1 projection changed")
		}
		again := readPublicBundle(t, destination.origin, destination.login, id, "reexport-network-flow-"+label, "snapshots")
		irVerifyReport(destination, report, again, false)
		for _, path := range []string{"tables", "observations", "diagnostics", "indicator_bindings", "graph_views"} {
			member := "ext/extensions/network_flow_activity/members/" + path + ".ndjson"
			if !bytes.Equal(sourceBundleMember(t, bundle, member), sourceBundleMember(t, again, member)) {
				t.Fatalf("portable %s changed during transfer", path)
			}
		}
		irCompareCorePortableInventory(t, bundle, again)
		var total int
		for _, tableID := range tableIDs {
			path := root + "/tables/" + tableID
			leftTable := call(a.Server.HTTP.URL, loginA, http.MethodGet, path, nil, http.StatusOK)["table"].(map[string]any)
			rightTable := call(destination.origin, destination.login, http.MethodGet, path, nil, http.StatusOK)["table"].(map[string]any)
			if len(leftTable) != len(rightTable) {
				t.Fatal("table resource shape changed")
			}
			if leftTable["source_filename_digest"] == rightTable["source_filename_digest"] {
				t.Fatal("deployment-keyed correlation was not regenerated")
			}
			for field, value := range leftTable {
				switch field {
				case "source_filename_digest", "source_filename_digest_key_id":
					continue
				}
				if !reflect.DeepEqual(value, rightTable[field]) {
					t.Fatalf("public table field %s changed", field)
				}
			}
			leftQuery := map[string]any{"schema_id": "cartulary.network_flow.table_query_request.v2", "limit": 200}
			rightQuery := leftQuery
			for {
				left := call(a.Server.HTTP.URL, loginA, http.MethodPost, path+"/query", leftQuery, http.StatusOK)
				right := call(destination.origin, destination.login, http.MethodPost, path+"/query", rightQuery, http.StatusOK)
				if !reflect.DeepEqual(left["rows"], right["rows"]) {
					t.Fatal("public retained observation or historical attribution changed")
				}
				total += len(left["rows"].([]any))
				leftCursor := left["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"]
				rightCursor := right["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"]
				if (leftCursor == nil) != (rightCursor == nil) {
					t.Fatal("transferred pagination lost records")
				}
				if leftCursor == nil {
					break
				}
				leftQuery = map[string]any{"schema_id": "cartulary.network_flow.table_query_continuation.v2", "cursor_token": leftCursor}
				rightQuery = map[string]any{"schema_id": "cartulary.network_flow.table_query_continuation.v2", "cursor_token": rightCursor}
			}
		}
		if total != 4500 {
			t.Fatalf("public retained observation count = %d", total)
		}
		response := doImportJSON(t, destination.origin, loginA, http.MethodGet, root+"/tables", nil)
		defer response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatal("source credentials authenticated to destination")
		}
	}
	verifyDestination(clientB, "b")
	// An unclaimed destination must reject the complete incident, leaving no public state.
	unsupported := start("nf-portable-disabled", 51, false)
	loginDisabled, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, unsupported.Server.HTTP.URL)
	denied := uploadPublicSourceBundle(t, unsupported.Server.HTTP.URL, loginDisabled, bundle, "reject-unavailable-profile")
	failed := waitIRApply(t, unsupported.Server.HTTP.URL, loginDisabled, denied["job_id"].(string))
	if failed["status"] != "failed" {
		t.Fatalf("unavailable profile did not reject: %#v", failed)
	}
	errorSummary := failed["error_summary"].(map[string]any)
	if errorSummary["code"] != "incident_bundle_import_rejected" || errorSummary["retryable"] != false ||
		!reflect.DeepEqual(errorSummary["details"], map[string]any{
			"reason_code":      "source_family_invalid",
			"source_family_id": "extension_payload",
			"invariant_id":     "extension_payload.participant_admitted",
		}) {
		t.Fatalf("unavailable profile failure differs: %#v", errorSummary)
	}
	response := doImportJSON(t, unsupported.Server.HTTP.URL, loginDisabled, http.MethodGet, "/api/v1/incidents/"+id, nil)
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatal("rejected transfer exposed an incident")
	}
	c := runtime.BackupRestoreServerWithOperator(t, b, appsupport.ServerOptions{Prefix: "ir-restored-c", TestRouteMode: httptestx.TestRouteModeDisabled, Env: flowEnvironment(31, true), ConfigureRuntime: func(options *server.Options) { options.ReportingTemplatePacks = irReportTemplates(t) }})
	for _, prior := range []authcookietest.AuthCookies{loginA, loginB} {
		response := doImportJSON(t, c.Server.HTTP.URL, prior, http.MethodGet, "/api/v1/auth/session", nil)
		response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatal("pre-transfer or pre-backup session authenticated after restore")
		}
	}
	loginC := flowtest.LoginLocalUserWithSecondFactor(t, http.DefaultClient, c.Server.HTTP.URL, "bootstrap-admin@example.test", "BootstrapPass1!", flowtest.GenerateTOTPCode(t, secretB))
	// Recovery admits the deployment first; the claimed Graph owner then
	// reconstructs derived results behind its separate readiness gate. Observe
	// only that documented pending response, without refreshing the declaration.
	deadline, ok := t.Deadline()
	if !ok {
		t.Fatal("restored Graph readiness requires a harness deadline")
	}
	graphContext, cancelGraphWait := context.WithDeadline(t.Context(), deadline)
	defer cancelGraphWait()
	for {
		response := pollImportJSON(t, graphContext, c.Server.HTTP.URL, loginC, graphPath+"/result")
		if response.StatusCode == http.StatusOK {
			response.Body.Close()
			break
		}
		var failure struct {
			Error struct {
				Code    string `json:"code"`
				Details struct {
					Reason string `json:"reason_code"`
				} `json:"details"`
			} `json:"error"`
		}
		err := json.NewDecoder(response.Body).Decode(&failure)
		response.Body.Close()
		if err != nil || response.StatusCode != http.StatusBadGateway || failure.Error.Code != "network_flow_graph_projection_failed" || failure.Error.Details.Reason != "projection_unavailable" {
			t.Fatalf("restored graph readiness: status=%d code=%s reason=%s decode=%v", response.StatusCode, failure.Error.Code, failure.Error.Details.Reason, err)
		}
		select {
		case <-graphContext.Done():
			t.Fatalf("restored Graph never became available: %v", graphContext.Err())
		case <-time.After(100 * time.Millisecond):
		}
	}
	verifyDestination(irPublicClient{t: t, origin: c.Server.HTTP.URL, incident: id, login: loginC}, "c")

}

// Compare every Core source member, including aliases, revisions and saved views.
// The only excluded source field is Evidence's deployment-local storage key;
// exact attachment bytes and hashes are verified separately through public APIs.
func irCompareCorePortableInventory(t *testing.T, before, after []byte) {
	t.Helper()
	paths := func(bundle []byte) []string {
		reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
		if err != nil {
			t.Fatal(err)
		}
		var result []string
		for _, file := range reader.File {
			if strings.HasPrefix(file.Name, "data/") {
				result = append(result, file.Name)
			}
		}
		return result
	}
	if !reflect.DeepEqual(paths(before), paths(after)) {
		t.Fatal("portable Core member inventory changed")
	}
	for _, member := range paths(before) {
		left, right := sourceBundleMember(t, before, member), sourceBundleMember(t, after, member)
		if member == "data/object_blobs.ndjson" {
			decode := func(content []byte) []map[string]any {
				var rows []map[string]any
				for _, line := range bytes.Split(bytes.TrimSpace(content), []byte{'\n'}) {
					if len(line) == 0 {
						continue
					}
					var row map[string]any
					decoder := json.NewDecoder(bytes.NewReader(line))
					decoder.UseNumber()
					if err := decoder.Decode(&row); err != nil {
						t.Fatal(err)
					}
					if _, ok := row["storage_key"].(string); !ok {
						t.Fatal("Evidence storage key is absent")
					}
					delete(row, "storage_key")
					rows = append(rows, row)
				}
				return rows
			}
			if !reflect.DeepEqual(decode(left), decode(right)) {
				t.Fatal("portable Evidence blob metadata changed")
			}
		} else if !bytes.Equal(left, right) {
			t.Fatalf("portable Core state changed in %s", member)
		}
	}
}

// Retain the enrolled credential only in this test's memory so recovery can
// prove fresh authentication without seeding an account into the restored DB.
func irBootstrapRecoveryAdmin(t *testing.T, origin string) (authcookietest.AuthCookies, string) {
	t.Helper()
	token := flowtest.RequireBootstrapLogin(t, http.DefaultClient, origin, "bootstrap-admin@example.test", "BootstrapPass1!")
	begin := flowtest.BeginTOTPEnrollment(t, http.DefaultClient, origin, token, map[string]any{"client_txn_id": "ir-bootstrap-begin"})
	secret := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	flowtest.CompleteInitialEnrollment(t, http.DefaultClient, origin, token, begin["enrollment_id"].(string), secret, "ir-bootstrap-complete")
	return flowtest.LoginLocalUserWithSecondFactor(t, http.DefaultClient, origin, "bootstrap-admin@example.test", "BootstrapPass1!", flowtest.GenerateTOTPCode(t, secret)), secret
}
