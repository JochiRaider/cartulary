package networkflow

import (
	"encoding/json"
	"testing"

	contractnetworkflow "github.com/JochiRaider/cartulary/internal/gen/contractnetworkflow"
)

func TestNetworkFlowV4GraphContractProjection_Unit(t *testing.T) {
	t.Parallel()

	index := decodeNetworkFlowContractArtifact(t, "contracts/network-flow/index.json")
	if index["contract_major"] != float64(4) || index["schema_id"] != "cartulary.network_flow_contract_index.v3" {
		t.Fatalf("Network Flow contract identity = %#v; want major 4/index v3", index)
	}
	publicSchemas := anyStringBoolSet(index["public_schema_ids"].([]any))
	for _, schemaID := range []string{
		"cartulary.network_flow.graph_semantic_query.v1",
		"cartulary.network_flow.graph_semantic_query.v2",
		"cartulary.network_flow.graph_view_list.v2",
		"cartulary.network_flow.graph_view_create_request.v2",
		"cartulary.network_flow.graph_view_accepted.v2",
		"cartulary.network_flow.graph_view_result.v2",
		"cartulary.network_flow.graph_view_contributor_query_result.v2",
	} {
		if !publicSchemas[schemaID] {
			t.Fatalf("Network Flow v4 public schema registry omits %q", schemaID)
		}
	}

	routesDocument := decodeNetworkFlowContractArtifact(t, "contracts/network-flow/routes.v1.json")
	routes := routesDocument["routes"].([]any)
	graphRoutes := map[string]map[string]any{}
	for _, raw := range routes {
		route := raw.(map[string]any)
		path := route["path"].(string)
		if len(path) >= len("/api/v1/incidents/{incident_id}/network-flow/graph-views") &&
			path[:len("/api/v1/incidents/{incident_id}/network-flow/graph-views")] == "/api/v1/incidents/{incident_id}/network-flow/graph-views" {
			graphRoutes[route["route_id"].(string)] = route
		}
	}
	if len(graphRoutes) != 8 {
		t.Fatalf("saved-graph route count = %d; want 8", len(graphRoutes))
	}
	assertNetworkFlowGraphRoute(t, graphRoutes, "nf.graph_views.create", "editor", "client_txn_id_required", 202)
	assertNetworkFlowGraphRoute(t, graphRoutes, "nf.graph_views.patch", "editor", "client_txn_id_required", 200)
	assertNetworkFlowGraphRoute(t, graphRoutes, "nf.graph_views.delete", "reviewer", "client_txn_id_required", 200)
	assertNetworkFlowGraphRoute(t, graphRoutes, "nf.graph_views.refresh", "editor", "client_txn_id_required", 202)
	for _, routeID := range []string{"nf.graph_views.list", "nf.graph_views.get", "nf.graph_views.result", "nf.graph_views.contributors.query"} {
		assertNetworkFlowGraphRoute(t, graphRoutes, routeID, "viewer", "read_route", 200)
	}

	schemas := decodeNetworkFlowContractArtifact(t, "contracts/network-flow/schemas.v2.json")
	definitions := schemas["$defs"].(map[string]any)
	graphResult := definitions["GraphProjectionResultV2"].(map[string]any)
	if graphResult["additionalProperties"] != false {
		t.Fatalf("nested Graph Projection v2 result must be closed: %#v", graphResult)
	}
	properties := graphResult["properties"].(map[string]any)
	if properties["projection_schema_id"].(map[string]any)["const"] != "graph_projection.v2" {
		t.Fatalf("ephemeral graph result did not cut directly to v2: %#v", properties["projection_schema_id"])
	}
	for _, definitionName := range []string{"GraphViewCreateRequestV2", "GraphViewRenameRequest", "GraphViewRefreshRequest", "GraphViewRetireRequest"} {
		definition := definitions[definitionName].(map[string]any)
		required := anyStringBoolSet(definition["required"].([]any))
		if !required["client_txn_id"] {
			t.Fatalf("%s omits owner-required client_txn_id", definitionName)
		}
	}
	limits := definitions["EffectiveLimitsV2"].(map[string]any)["properties"].(map[string]any)
	for _, limitName := range []string{
		"network_flow.max_nonterminal_graph_jobs_per_incident",
		"network_flow.max_contributing_rows_per_graph",
		"network_flow.max_time_buckets_per_graph",
		"network_flow.graph_materialization_timeout_seconds",
	} {
		if _, present := limits[limitName]; !present {
			t.Fatalf("effective-limit projection omits %q: %#v", limitName, limits)
		}
	}
	if _, stale := limits["network_flow.max_nonterminal_graph_view_jobs_per_incident"]; stale {
		t.Fatal("legacy saved-graph job limit name remains projected")
	}

	semantics := decodeNetworkFlowContractArtifact(t, "contracts/network-flow/graph-semantics.v2.json")
	if semantics["graph_projection_schema_id"] != "graph_projection.v2" {
		t.Fatalf("Network Flow graph semantics changed Graph Projection ownership: %#v", semantics)
	}
	timeBucket := semantics["time_bucket"].(map[string]any)
	if timeBucket["mode"] != "time_bucket_v1" ||
		timeBucket["projection_version"] != "network_flow_activity.time_bucket.v1" ||
		timeBucket["relationship_kind"] != "network_flow.bucketed_flow_edge.v1" {
		t.Fatalf("Network Flow time-bucket semantics drifted: %#v", timeBucket)
	}
	resourceLimits := decodeNetworkFlowContractArtifact(t, "contracts/network-flow/resource-limits.v2.json")
	if len(resourceLimits["limits"].([]any)) != 23 {
		t.Fatalf("effective-limit registry count = %d; want 23", len(resourceLimits["limits"].([]any)))
	}
}

func decodeNetworkFlowContractArtifact(t testing.TB, path string) map[string]any {
	t.Helper()
	artifact, ok := contractnetworkflow.Index[path]
	if !ok {
		t.Fatalf("generated Network Flow artifact %q is missing", path)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &decoded); err != nil {
		t.Fatalf("decode generated Network Flow artifact %q: %v", path, err)
	}
	return decoded
}

func anyStringBoolSet(values []any) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value.(string)] = true
	}
	return set
}

func assertNetworkFlowGraphRoute(t testing.TB, routes map[string]map[string]any, routeID, auth, idempotency string, status float64) {
	t.Helper()
	route := routes[routeID]
	if route == nil || route["auth_context"] != auth || route["idempotency"] != idempotency {
		t.Fatalf("saved-graph route %q contract drifted: %#v", routeID, route)
	}
	statuses := route["success_http_statuses"].([]any)
	if len(statuses) != 1 || statuses[0] != status {
		t.Fatalf("saved-graph route %q statuses = %#v; want [%v]", routeID, statuses, status)
	}
}
