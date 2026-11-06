package networkflow_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

// Small owner fixtures isolate selected history from corpus-volume qualification.
// Every graph and catalog transition below goes through the production routes.
func TestNetworkFlowSelectedSourceHistory_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	h := claimedNetworkFlowServerForRouteTest(t, runtime, "selected-source-history")
	login, actor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, h.Server, login, map[string]any{
		"client_txn_id": "history-incident", "incident_key": "IR-SELECTED-HISTORY", "title": "Selected source history",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	store := newTestNetworkFlowStore(t, h.Pool, h.Revisions.Appender())
	first := createTestTable(t, h.Pool, store, actor, incidentID, "first.csv", nil, 1)
	base := h.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow"
	writes := []func(*http.Request){httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)}
	read := func(path string) map[string]any {
		return httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, path, nil, httptestx.WithCookies(login.Session)), http.StatusOK)["data"].(map[string]any)
	}
	create := func(scope, txn string) string {
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, base+"/graph-views", map[string]any{
			"schema_id": "cartulary.network_flow.graph_view_create_request.v5", "refresh_scope": scope, "client_txn_id": txn, "display_name": txn,
			"semantic_query": map[string]any{"schema_id": "cartulary.network_flow.graph_semantic_query.v3", "selected_table_ids": []string{first.TableID}, "filters": []any{},
				"time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}},
		}, writes...)
		created := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		waitForNetworkFlowJob(t, h.Server.HTTP.URL, login, created["job"].(map[string]any)["job_id"].(string), "succeeded")
		return base + "/graph-views/" + created["graph_view"].(map[string]any)["graph_view_id"].(string)
	}
	explicit := create("selected_tables", "history-explicit")
	all := create("all_active_tables", "history-all")
	initial := read(explicit + "/result")
	originalProjection := initial["result"].(map[string]any)["graph_projection_result"]
	originalBinding := initial["graph_view"].(map[string]any)["selected_source_binding"]
	allBefore := read(all + "/result")["graph_view"].(map[string]any)["selected_source_binding"]
	_ = createTestTable(t, h.Pool, store, actor, incidentID, "later.csv", nil, 2)
	renamed := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, base+"/tables/"+first.TableID, map[string]any{
		"client_txn_id": "history-rename", "base_table_version": 1, "display_name": "Renamed source",
	}, writes...), http.StatusOK)["data"].(map[string]any)
	_ = renamed
	for _, path := range []string{explicit, all} {
		selected := read(path + "/result")
		want := originalBinding
		if path == all {
			want = allBefore
		}
		if !reflect.DeepEqual(selected["graph_view"].(map[string]any)["selected_source_binding"], want) {
			t.Fatal("catalog change rewrote selected source")
		}
	}
	refreshed := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, all+"/refresh", map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_refresh_request.v1", "client_txn_id": "history-expand", "base_graph_view_version": 1,
	}, writes...), http.StatusAccepted)["data"].(map[string]any)
	waitForNetworkFlowJob(t, h.Server.HTTP.URL, login, refreshed["job"].(map[string]any)["job_id"].(string), "succeeded")
	expanded := read(all + "/result")["graph_view"].(map[string]any)["selected_source_binding"].(map[string]any)
	if len(expanded["tables"].([]any)) != 2 || reflect.DeepEqual(expanded, allBefore) {
		t.Fatal("explicit all-active refresh did not resolve expanded sources")
	}
	httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, base+"/tables/"+first.TableID, map[string]any{
		"client_txn_id": "history-delete", "base_table_version": 2,
	}, writes...), http.StatusOK)
	retained := read(explicit + "/result")
	if !reflect.DeepEqual(retained["result"].(map[string]any)["graph_projection_result"], originalProjection) ||
		!reflect.DeepEqual(retained["graph_view"].(map[string]any)["selected_source_binding"], originalBinding) {
		t.Fatal("source deletion changed selected meaning")
	}
	graph := retained["graph_view"].(map[string]any)
	denied := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, explicit+"/refresh", map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_refresh_request.v1", "client_txn_id": "history-deleted-refresh", "base_graph_view_version": graph["graph_view_version"],
	}, writes...)
	if denied.StatusCode < 400 || denied.StatusCode >= 500 {
		denied.Body.Close()
		t.Fatalf("deleted explicit source refresh status %d", denied.StatusCode)
	}
	denied.Body.Close()
	if !reflect.DeepEqual(read(explicit + "/result")["graph_view"].(map[string]any)["selected_source_binding"], originalBinding) {
		t.Fatal("failed refresh replaced selected history")
	}
	httptestx.RequireNoContent(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, explicit, map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_retire_request.v1", "client_txn_id": "history-retire", "base_graph_view_version": graph["graph_view_version"],
	}, writes...))
	historical := read(explicit + "/result")
	if historical["graph_view"].(map[string]any)["state"] != "retired" || !reflect.DeepEqual(historical["result"].(map[string]any)["graph_projection_result"], originalProjection) {
		t.Fatal("retirement erased selected history")
	}
	// Reconstruction consumes retained inputs, including deleted tables and
	// retired declarations, and verifies each complete immutable result binding.
	registration, err := NewGraphRestoreSourceRegistration(h.Pool)
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := registration.Enumerate(context.Background())
	if err != nil || len(candidates) != 2 {
		t.Fatalf("selected reconstruction inventory: %d %v", len(candidates), err)
	}
	for _, candidate := range candidates {
		projected, err := graphprojection.ProjectV2(t.Context(), graphprojection.InvocationContextV2{GraphViewID: candidate.GraphViewID, SourceOwnerID: ProfileID}, candidate.SemanticInput)
		if err != nil || projected.ResultBindingV2() != candidate.ExpectedBinding {
			t.Fatalf("selected reconstruction differs: %v", err)
		}
	}
}
