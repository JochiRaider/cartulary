package imports_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
)

// The small rejection fixture isolates historical transport from corpus totals.
// Import, catalog changes, graph selection and transfer use public operations.
func TestNetworkFlowHistoricalTransferPreservesDiagnosticsAndOtherIncidents_Integration(t *testing.T) {
	runRoot, err := networkflowfixtures.SelectedRunRoot()
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../../../fixtures/network-flow")
	if err != nil {
		t.Fatal(err)
	}
	summary, err := networkflowfixtures.Run(networkflowfixtures.Config{CorpusRoot: root, RunRoot: runRoot, FixtureID: "NF-FIX-004-rejected-rows", Revision: 9}, func(fixture *networkflowfixtures.Bundle) error {
		source, err := fixture.Read("source/partial-rejection.csv")
		if err != nil {
			return err
		}
		raw, err := fixture.Read("source/cases.json")
		if err != nil {
			return err
		}
		var cases []struct {
			Name    string          `json:"name"`
			Mapping json.RawMessage `json:"mapping"`
		}
		if err := json.Unmarshal(raw, &cases); err != nil {
			return err
		}
		var mapping json.RawMessage
		for _, candidate := range cases {
			if candidate.Name == "partial-rejection" {
				mapping = candidate.Mapping
			}
		}
		if len(mapping) == 0 {
			t.Fatal("named partial-rejection mapping absent")
		}
		runtime := appsupport.StartRuntime(t)
		start := func(prefix string) *appsupport.ServerHarness {
			return runtime.StartServer(t, appsupport.ServerOptions{Prefix: prefix, TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{
				"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
				"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
				"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
				"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
			}})
		}
		a, b := start("history-transfer-a"), start("history-transfer-b")
		loginA, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, a.Server.HTTP.URL)
		loginB, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, b.Server.HTTP.URL)
		incident := scenariotest.CreateIncident(t, a.Server, loginA, map[string]any{"client_txn_id": "history-source", "incident_key": "HISTORY-SOURCE", "title": "Historical flow transfer"})["incident_id"].(string)
		other := scenariotest.CreateIncident(t, b.Server, loginB, map[string]any{"client_txn_id": "history-other", "incident_key": "HISTORY-OTHER", "title": "Unrelated destination investigation"})["incident_id"].(string)
		left := irPublicClient{t: t, origin: a.Server.HTTP.URL, login: loginA, incident: incident}
		right := irPublicClient{t: t, origin: b.Server.HTTP.URL, login: loginB, incident: incident}
		otherClient := irPublicClient{t: t, origin: b.Server.HTTP.URL, login: loginB, incident: other}
		table := networkflowsupport.ImportCSV(t, http.DefaultClient, left.origin, loginA, incident, "history-source", "partial-rejection.csv", source, mapping)
		otherTable := networkflowsupport.SeedGraphSource(t, http.DefaultClient, right.origin, loginB, other)
		graph := func(c irPublicClient, table, txn string) string {
			created := c.request(http.MethodPost, "/api/v1/incidents/"+c.incident+"/network-flow/graph-views", map[string]any{
				"schema_id": "cartulary.network_flow.graph_view_create_request.v5", "refresh_scope": "selected_tables", "client_txn_id": txn, "display_name": "Retained analysis",
				"semantic_query": map[string]any{"schema_id": "cartulary.network_flow.graph_semantic_query.v3", "selected_table_ids": []string{table}, "filters": []any{}, "time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}},
			}, http.StatusAccepted)
			if waitIRApply(t, c.origin, c.login, created["job"].(map[string]any)["job_id"].(string))["status"] != "succeeded" {
				t.Fatal("historical graph failed materialization")
			}
			return "/api/v1/incidents/" + c.incident + "/network-flow/graph-views/" + created["graph_view"].(map[string]any)["graph_view_id"].(string)
		}
		active, retired, unrelated := graph(left, table, "retained-active"), graph(left, table, "retained-retired"), graph(otherClient, otherTable, "unrelated-selected")
		duplicateName := graph(left, table, "retained-active-duplicate-name")
		before := left.request(http.MethodGet, active+"/result", nil, http.StatusOK)
		otherBefore := otherClient.request(http.MethodGet, unrelated+"/result", nil, http.StatusOK)
		base := "/api/v1/incidents/" + incident + "/network-flow"
		left.request(http.MethodPatch, base+"/tables/"+table, map[string]any{"client_txn_id": "history-rename", "base_table_version": 1, "display_name": "Renamed then deleted"}, http.StatusOK)
		left.request(http.MethodDelete, base+"/tables/"+table, map[string]any{"client_txn_id": "history-delete", "base_table_version": 2}, http.StatusOK)
		retiredBefore := left.request(http.MethodGet, retired+"/result", nil, http.StatusOK)
		response := doImportJSON(t, left.origin, loginA, http.MethodDelete, retired, map[string]any{"schema_id": "cartulary.network_flow.graph_view_retire_request.v1", "client_txn_id": "history-retire", "base_graph_view_version": retiredBefore["graph_view"].(map[string]any)["graph_view_version"]})
		httptestx.RequireNoContent(t, response)
		bundle := readPublicBundle(t, left.origin, loginA, incident, "history-export")
		job := uploadPublicSourceBundle(t, right.origin, loginB, bundle, "history-import")
		if waitIRApply(t, right.origin, loginB, job["job_id"].(string))["status"] != "succeeded" {
			t.Fatal("historical Network Flow transfer failed")
		}
		for _, path := range []string{active, duplicateName, retired} {
			result := right.request(http.MethodGet, path+"/result", nil, http.StatusOK)
			if !reflect.DeepEqual(before["result"].(map[string]any)["graph_projection_result"], result["result"].(map[string]any)["graph_projection_result"]) && path == active {
				t.Fatal("historical result changed")
			}
			original := left.request(http.MethodGet, path+"/result", nil, http.StatusOK)
			if result["graph_view"].(map[string]any)["display_name"] != "Retained analysis" {
				t.Fatal("duplicate display name changed during transfer")
			}
			if !reflect.DeepEqual(original["result"], result["result"]) || !reflect.DeepEqual(original["graph_view"].(map[string]any)["selected_source_binding"], result["graph_view"].(map[string]any)["selected_source_binding"]) {
				t.Fatal("retained selection was refreshed or lost")
			}
			if path == retired && result["graph_view"].(map[string]any)["state"] != "retired" {
				t.Fatal("retirement marker lost")
			}
		}
		if !reflect.DeepEqual(otherBefore, otherClient.request(http.MethodGet, unrelated+"/result", nil, http.StatusOK)) {
			t.Fatal("incident transfer changed unrelated Graph state")
		}
		again := readPublicBundle(t, right.origin, loginB, incident, "history-reexport")
		prefix := "ext/extensions/network_flow_activity/members/"
		for _, family := range []string{"tables.ndjson", "observations.ndjson", "diagnostics.ndjson", "indicator_bindings.ndjson", "graph_views.ndjson"} {
			if !bytes.Equal(sourceBundleMember(t, bundle, prefix+family), sourceBundleMember(t, again, prefix+family)) {
				t.Fatalf("historical family changed: %s", family)
			}
		}
		var tableRow map[string]any
		if err := json.Unmarshal(bytes.TrimSpace(sourceBundleMember(t, again, prefix+"tables.ndjson")), &tableRow); err != nil {
			return err
		}
		if tableRow["row_count_accepted"] != float64(1) || tableRow["row_count_rejected"] != float64(1) || tableRow["table_status"] != "soft_deleted" {
			t.Fatalf("historical source accounting changed: %v", tableRow["table_status"])
		}
		if len(bytes.TrimSpace(sourceBundleMember(t, again, prefix+"diagnostics.ndjson"))) == 0 {
			t.Fatal("rejection diagnostics disappeared")
		}
		return nil
	})
	t.Logf("historical fixture comparison: %s", summary)
	if err != nil {
		t.Fatal(err)
	}
}
