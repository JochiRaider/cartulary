package networkflow

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	contractgraphprojection "github.com/JochiRaider/cartulary/internal/gen/contractgraphprojection"
	contractnetworkflow "github.com/JochiRaider/cartulary/internal/gen/contractnetworkflow"
	"github.com/JochiRaider/cartulary/internal/gen/networkflowroutes"
)

func TestNetworkFlowV7GraphContractProjection_Unit(t *testing.T) {
	t.Parallel()
	policy := decodeNetworkFlowContractArtifact(t, "contracts/network-flow/operation-policy.v1.json")
	if policy["contract_major"] != float64(7) || policy["query_completion_boundary"] != "audit_transaction_commit" || policy["durable_state_version"] != float64(4) || policy["semantic_query_version"] != float64(2) || policy["job_payload_version"] != float64(1) {
		t.Fatalf("operation policy cutover drift: %#v", policy)
	}
	policies := map[string]map[string]any{}
	for _, raw := range policy["operations"].([]any) {
		entry := raw.(map[string]any)
		policies[entry["route_id"].(string)] = entry
	}
	for _, route := range networkflowroutes.All() {
		entry, ok := policies[route.RouteID]
		if !ok {
			t.Fatalf("missing policy for %s", route.RouteID)
		}
		mutation := strings.HasSuffix(route.RouteID, ".create") || strings.HasSuffix(route.RouteID, ".patch") || strings.HasSuffix(route.RouteID, ".delete") || strings.HasSuffix(route.RouteID, ".refresh")
		if entry["csrf_required"] != mutation || (entry["body"] == "forbidden") != (route.Method == "GET") {
			t.Fatalf("admission projection drift for %s", route.RouteID)
		}
	}
	if len(policies) != len(networkflowroutes.All()) {
		t.Fatal("unbound operation policies")
	}

	index := decodeNetworkFlowContractArtifact(t, "contracts/network-flow/index.json")
	if index["contract_major"] != float64(7) || index["schema_id"] != "cartulary.network_flow_contract_index.v3" {
		t.Fatalf("Network Flow contract identity = %#v; want major 7/index v3", index)
	}
	graphSchemas := make([]string, 0)
	for _, rawSchemaID := range index["public_schema_ids"].([]any) {
		schemaID := rawSchemaID.(string)
		if strings.HasPrefix(schemaID, "cartulary.network_flow.graph_") {
			graphSchemas = append(graphSchemas, schemaID)
		}
	}
	wantGraphSchemas := []string{
		"cartulary.network_flow.graph_view_rename_request.v2",
		"cartulary.network_flow.graph_view_refresh_request.v1",
		"cartulary.network_flow.graph_view_retire_request.v1",
		"cartulary.network_flow.graph_query_request.v2",
		"cartulary.network_flow.graph_query_result.v2",
		"cartulary.network_flow.graph_semantic_query.v2",
		"cartulary.network_flow.graph_contributor_query_request.v2",
		"cartulary.network_flow.graph_contributor_query_continuation.v1",
		"cartulary.network_flow.graph_contributor_query_result.v2",
		"cartulary.network_flow.graph_view_create_request.v3",
		"cartulary.network_flow.graph_view_contributor_query_request.v2",
		"cartulary.network_flow.graph_view_contributor_query_result.v2",
		"cartulary.network_flow.graph_view.v4",
		"cartulary.network_flow.graph_view_list.v4",
		"cartulary.network_flow.graph_view_get.v4",
		"cartulary.network_flow.graph_view_accepted.v4",
		"cartulary.network_flow.graph_view_mutation_result.v4",
		"cartulary.network_flow.graph_view_result.v4",
	}
	if !slices.Equal(graphSchemas, wantGraphSchemas) {
		t.Fatalf("Network Flow v7 Graph schema allowlist = %#v; want %#v", graphSchemas, wantGraphSchemas)
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
	assertNetworkFlowGraphRoute(t, graphRoutes, "nf.graph_views.delete", "reviewer", "client_txn_id_required", 204)
	assertNetworkFlowGraphRoute(t, graphRoutes, "nf.graph_views.refresh", "editor", "client_txn_id_required", 202)
	for _, routeID := range []string{"nf.graph_views.list", "nf.graph_views.get", "nf.graph_views.result", "nf.graph_views.contributors.query"} {
		assertNetworkFlowGraphRoute(t, graphRoutes, routeID, "viewer", "read_route", 200)
	}

	schemas := decodeNetworkFlowContractArtifact(t, "contracts/network-flow/schemas.v3.json")
	definitions := schemas["$defs"].(map[string]any)
	declaration := definitions["GraphViewV4"].(map[string]any)
	wantDeclaration := []string{
		"schema_id", "graph_view_id", "incident_id", "display_name", "state",
		"semantic_query", "semantic_query_sha256", "desired_source_snapshot_id",
		"selected_result_binding", "graph_view_version", "materialization_generation",
		"created_by", "created_at", "updated_at", "latest_job_id", "last_failure_code", "last_failed_at",
	}
	declarationProperties := declaration["properties"].(map[string]any)
	if len(declarationProperties) != len(wantDeclaration) || declaration["additionalProperties"] != false {
		t.Fatalf("declaration must expose exactly the complete owner resource: %#v", declaration)
	}
	declarationRequired := anyStringBoolSet(declaration["required"].([]any))
	for _, member := range wantDeclaration {
		if _, present := declarationProperties[member]; !present || !declarationRequired[member] {
			t.Fatalf("declaration omits required owner fact %q", member)
		}
	}
	if len(declaration["oneOf"].([]any)) != 2 {
		t.Fatal("declaration must pair nullable failure code and time")
	}
	binding := definitions["GraphViewSelectedResult"].(map[string]any)
	if len(binding["required"].([]any)) != 7 || len(binding["properties"].(map[string]any)) != 7 {
		t.Fatal("immutable result binding must retain all seven identity members")
	}
	receipt := definitions["GraphViewAcceptedV4"].(map[string]any)
	if len(receipt["properties"].(map[string]any)) != 3 || !anyStringBoolSet(receipt["required"].([]any))["job"] {
		t.Fatal("accepted receipt must contain only schema, declaration and narrow job reference")
	}
	jobReference := definitions["CommonJobReference"].(map[string]any)
	jobMembers := anyStringBoolSet(jobReference["required"].([]any))
	if len(jobReference["properties"].(map[string]any)) != 2 || !jobMembers["job_id"] || !jobMembers["status_route"] {
		t.Fatal("initiating job reference must not copy execution metadata")
	}
	if graphRoutes["nf.graph_views.delete"]["success_schema_id"] != nil {
		t.Fatal("retirement must have no response schema or body")
	}
	graphResult := definitions["GraphProjectionResultV2"].(map[string]any)
	if graphResult["additionalProperties"] != false {
		t.Fatalf("nested Graph Projection v2 result must be closed: %#v", graphResult)
	}
	properties := graphResult["properties"].(map[string]any)
	if properties["projection_schema_id"].(map[string]any)["const"] != "graph_projection.v2" {
		t.Fatalf("ephemeral graph result did not cut directly to v2: %#v", properties["projection_schema_id"])
	}
	for _, definitionName := range []string{"GraphViewCreateRequestV3", "GraphViewRenameRequestV2", "GraphViewRefreshRequest", "GraphViewRetireRequest"} {
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
	storageArtifact := contractgraphprojection.Index["contracts/graph-projection/storage-maintenance.v1.json"]
	var storage map[string]any
	if err := json.Unmarshal([]byte(storageArtifact.JSON), &storage); err != nil {
		t.Fatalf("decode Graph storage-maintenance projection: %v", err)
	}
	sweep := storage["sweep_bounds"].(map[string]any)
	if sweep["maximum_results"] != float64(graphResultCleanupMaximumResults) ||
		sweep["maximum_duration_seconds"] != graphResultCleanupMaximumDuration.Seconds() ||
		sweep["continuation_delay_seconds"] != graphResultCleanupContinuationDelay.Seconds() ||
		sweep["base_cadence_seconds"] != graphResultCleanupBaseCadence.Seconds() {
		t.Fatalf("Network Flow cleanup runtime/Graph storage contract drifted: %#v", sweep)
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
