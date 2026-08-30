package graphprojection

import (
	"encoding/json"
	"testing"

	contractgraphprojection "github.com/JochiRaider/cartulary/internal/gen/contractgraphprojection"
)

func TestGraphProjectionV2ContractProjection_Unit(t *testing.T) {
	t.Parallel()

	index := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/index.json")
	if index["contract_id"] != "cartulary.graph_projection_nlspec.v2.2.0" ||
		index["projection_schema_id"] != "graph_projection.v2" {
		t.Fatalf("Graph Projection v2 index identity drifted: %#v", index)
	}
	if _, present := index["limits"]; present {
		t.Fatal("Graph Projection index must not duplicate the semantic registry limits")
	}
	registry := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/semantic-registry.v1.json")
	limits := registry["semantic_limits"].(map[string]any)
	runtimeLimits := map[string]int{
		"cancellation_check_interval_items": graphProjectionLimits.CancellationCheckIntervalItems,
		"maximum_aggregation_definitions":   graphProjectionLimits.MaxAggregationRules,
		"maximum_array_items":               graphProjectionLimits.MaxArrayItems,
		"maximum_entity_filters":            graphProjectionLimits.MaxEntityFilters,
		"maximum_entity_mappings":           graphProjectionLimits.MaxEntityMappings,
		"maximum_input_bytes":               graphProjectionLimits.MaxInputBytes,
		"maximum_labels_per_object":         graphProjectionLimits.MaxLabelsPerObject,
		"maximum_metadata_mappings":         graphProjectionLimits.MaxMetadataMappings,
		"maximum_projected_edges":           graphProjectionLimits.MaxProjectedEdges,
		"maximum_projected_vertices":        graphProjectionLimits.MaxProjectedVertices,
		"maximum_property_definitions":      graphProjectionLimits.MaxPropertyDefinitions,
		"maximum_property_keys":             graphProjectionLimits.MaxPropertyKeys,
		"maximum_relationship_filters":      graphProjectionLimits.MaxRelationshipFilters,
		"maximum_relationship_mappings":     graphProjectionLimits.MaxRelationshipMappings,
		"maximum_source_edges":              graphProjectionLimits.MaxSourceRelationships,
		"maximum_source_vertices":           graphProjectionLimits.MaxSourceEntities,
		"maximum_traversal_depth":           graphProjectionLimits.MaxTraversalDepth,
		"maximum_traversal_kind_filters":    graphProjectionLimits.MaxTraversalKindFilters,
		"maximum_traversal_seeds":           graphProjectionLimits.MaxTraversalSeedVertices,
		"maximum_validation_issues":         graphProjectionLimits.MaxValidationIssues,
	}
	if len(limits) != len(runtimeLimits) {
		t.Fatalf("Graph Projection projected/runtime semantic limit counts = %d/%d", len(limits), len(runtimeLimits))
	}
	for key, runtimeValue := range runtimeLimits {
		if limits[key] != float64(runtimeValue) {
			t.Fatalf("Graph Projection semantic limit %s = %#v, runtime=%d", key, limits[key], runtimeValue)
		}
	}
	textLimits := registry["text_limits_bytes"].(map[string]any)
	runtimeTextLimits := map[string]int{
		"field_path":   graphProjectionLimits.MaxFieldPathBytes,
		"identifier":   graphProjectionLimits.MaxIdentifierBytes,
		"label":        graphProjectionLimits.MaxLabelBytes,
		"property_key": graphProjectionLimits.MaxPropertyKeyBytes,
		"string":       graphProjectionLimits.MaxStringBytes,
	}
	if len(textLimits) != len(runtimeTextLimits) {
		t.Fatalf("Graph Projection projected/runtime text limit counts = %d/%d", len(textLimits), len(runtimeTextLimits))
	}
	for key, runtimeValue := range runtimeTextLimits {
		if textLimits[key] != float64(runtimeValue) {
			t.Fatalf("Graph Projection text limit %s = %#v, runtime=%d", key, textLimits[key], runtimeValue)
		}
	}
	mergeMatrix := registry["merge_matrix"].(map[string]any)
	if len(mergeMatrix) != 12 {
		t.Fatalf("Graph Projection v2 merge matrix must cover all twelve projected types: %#v", mergeMatrix)
	}

	schema := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/schemas.v2.json")
	definitions := schema["$defs"].(map[string]any)
	input := definitions["SemanticInput"].(map[string]any)
	if input["additionalProperties"] != false {
		t.Fatalf("Graph Projection v2 input must be closed: %#v", input)
	}
	properties := input["properties"].(map[string]any)
	for _, removed := range []string{
		"requested_at", "requested_by", "retention_policy", "custom_configuration",
		"run_state", "previous_run_id", "source_owner_id", "graph_view_id",
	} {
		if _, present := properties[removed]; present {
			t.Fatalf("Graph Projection v2 semantic input retains removed field %q", removed)
		}
	}
	result := definitions["Result"].(map[string]any)
	resultProperties := result["properties"].(map[string]any)
	for _, required := range []string{
		"projection_result_id", "normalized_configuration_sha256", "normalized_source_sha256", "canonical_output_sha256",
	} {
		if _, present := resultProperties[required]; !present {
			t.Fatalf("Graph Projection v2 result omits identity field %q", required)
		}
	}

	fixture := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/v2-fixtures/empty-projection.json")
	trusted := fixture["trusted_context"].(map[string]any)
	semanticInput := fixture["input"].(map[string]any)
	if trusted["source_owner_id"] != "network_flow_activity" || semanticInput["projection_schema_id"] != "graph_projection.v2" {
		t.Fatalf("Graph Projection v2 fixture does not separate trusted context: %#v", fixture)
	}
	semanticFixture := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/v2-fixtures/semantic-projection.json")
	if semanticFixture["fixture_id"] != "GP2-FIX-002" || len(semanticFixture["input"].(map[string]any)["source_entities"].([]any)) == 0 {
		t.Fatalf("Graph Projection semantic golden must provide non-empty evidence: %#v", semanticFixture)
	}

	maintenance := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/storage-maintenance.v1.json")
	lockOrder := maintenance["lock_order"].([]any)
	if maintenance["transaction_ownership"] != "borrowed" || len(lockOrder) != 2 ||
		lockOrder[0] != "projection_result" || lockOrder[1] != "source_declaration" {
		t.Fatalf("Graph Projection storage-maintenance projection drifted: %#v", maintenance)
	}
}

func decodeGraphProjectionContractArtifact(t testing.TB, path string) map[string]any {
	t.Helper()
	artifact, ok := contractgraphprojection.Index[path]
	if !ok {
		t.Fatalf("generated Graph Projection artifact %q is missing", path)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &decoded); err != nil {
		t.Fatalf("decode generated Graph Projection artifact %q: %v", path, err)
	}
	return decoded
}
