package networkflow_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
)

func graphContributorRequest(graph map[string]any, selector any, limit int) map[string]any {
	return map[string]any{"schema_id": "cartulary.network_flow.graph_contributor_query_request.v2", "graph_query": graph["semantic_query"], "graph_query_digest": graph["graph_query_digest"], "selector": selector, "limit": limit}
}
func graphEndpointSelector(graph map[string]any, endpoint string) any {
	for _, value := range graph["vertex_selectors"].([]any) {
		selector := value.(map[string]any)["selector"].(map[string]any)
		if selector["endpoint_value"] == endpoint {
			return selector
		}
	}
	return nil
}

func TestNetworkFlowStaleGraphFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-009-soft-delete-stale-graph", 3, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return runGraphFixtureFamily(t, bundle, func(r *publicationFixtureRuntime, input graphFixtureCase, result map[string]any) error {
			root := "/api/v1/incidents/" + input.IncidentID + "/network-flow"
			tablePath := root + "/tables/nft_" + strings.ReplaceAll(input.TableUUID, "-", "")
			graph := result["graph_queries"].([]map[string]any)[0]["response"].(map[string]any)
			selector := graphEndpointSelector(graph, "192.0.2.10")
			if selector == nil {
				return fmt.Errorf("missing imported endpoint")
			}
			initial := graphContributorRequest(graph, selector, 1)
			page, token, payload := r.normalizedPage(r.request(http.MethodPost, root+"/graphs/contributors/query", initial, 200), "stale-contributor")
			rows, rowToken, rowPayload := r.normalizedPage(r.request(http.MethodPost, tablePath+"/query", map[string]any{"schema_id": "cartulary.network_flow.table_query_request.v1", "limit": 1}, 200), "stale-row")
			if token == "" || rowToken == "" {
				return fmt.Errorf("both stale-scope cursors must exist before deletion")
			}
			body := map[string]any{"client_txn_id": "fixture-name-delete", "base_table_version": 1}
			deletion := r.request(http.MethodDelete, tablePath, body, 200)
			state := r.durableState()
			if replay := r.request(http.MethodDelete, tablePath, body, 200); !reflect.DeepEqual(deletion, replay) || !reflect.DeepEqual(state, r.durableState()) {
				return fmt.Errorf("delete replay changed retained graph sources")
			}
			rejected := map[string]any{
				"initial_contributors":     r.observedError(http.MethodPost, root+"/graphs/contributors/query", initial),
				"contributor_continuation": r.observedError(http.MethodPost, root+"/graphs/contributors/query", map[string]any{"schema_id": "cartulary.network_flow.graph_contributor_query_continuation.v1", "cursor_token": token}),
				"row_continuation":         r.observedError(http.MethodPost, tablePath+"/query", map[string]any{"schema_id": "cartulary.network_flow.table_query_continuation.v1", "cursor_token": rowToken}),
				"explicit_deleted_scope":   r.observedError(http.MethodPost, root+"/graphs/query", input.Queries[0].Request),
			}
			if !reflect.DeepEqual(state, r.durableState()) {
				return fmt.Errorf("stale operations mutated retained state")
			}
			raw, err := json.Marshal(map[string]any{"schema_id": "cartulary.network_flow.graph_query_request.v2", "table_scope": map[string]any{"mode": "all_active_tables"}, "aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true}})
			if err != nil {
				return err
			}
			recomputed, err := r.compareGraphQuery(input.publicationFixtureCase, graphFixtureQuery{Name: "remaining-active-table", Request: raw})
			if err != nil {
				return err
			}
			result["stale_graph"] = map[string]any{"contributor_request": initial, "contributor_page": page, "contributor_cursor": payload, "row_page": rows, "row_cursor": rowPayload, "deletion": deletion, "durable_after_deletion": r.comparableState(state), "replay_unchanged": true, "rejected_operations": rejected, "recomputed": recomputed, "active_tables": r.request(http.MethodGet, root+"/tables", nil, 200)}
			return nil
		})
	})
}
