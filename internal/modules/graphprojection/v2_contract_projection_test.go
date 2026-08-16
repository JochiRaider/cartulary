package graphprojection

import (
	"encoding/json"
	"testing"

	contractgraphprojection "github.com/JochiRaider/cartulary/internal/gen/contractgraphprojection"
)

func TestGraphProjectionV2ContractProjection_Unit(t *testing.T) {
	t.Parallel()

	index := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/index.json")
	if index["contract_id"] != "cartulary.graph_projection_nlspec.v2.0.0" ||
		index["projection_schema_id"] != "graph_projection.v2" {
		t.Fatalf("Graph Projection v2 index identity drifted: %#v", index)
	}
	limits := index["limits"].(map[string]any)
	if limits["maximum_vertices"] != float64(100000) ||
		limits["maximum_edges"] != float64(250000) ||
		limits["cancellation_check_interval_items"] != float64(1024) {
		t.Fatalf("Graph Projection v2 limits drifted: %#v", limits)
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
