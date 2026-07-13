package graphprojection

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/fixturetest"
)

// TestGraphProjectionFixtureCandidate is intentionally excluded from normal
// backend targets. The Make wrapper supplies one fixture ID and writes only a
// disposable candidate under the test-results root.
func TestGraphProjectionFixtureCandidate(t *testing.T) {
	fixtureID := os.Getenv("GRAPH_PROJECTION_FIXTURE")
	if fixtureID == "" {
		t.Skip("candidate mode is explicit")
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := fixturetest.RepoRoot(workingDirectory)
	if err != nil {
		t.Fatal(err)
	}
	manifest, directory, err := fixturetest.Load(root, fixtureID)
	if err != nil {
		t.Fatal(err)
	}
	input, err := fixturetest.ReadArtifact(directory, manifest.Steps[0].InputArtifact)
	if err != nil {
		t.Fatal(err)
	}
	if candidate, ok := candidateProjectionInput(t, fixtureID); ok {
		input = mustJSON(t, candidate)
	}
	execution, err := unitFixtureExecutor{}.ExecuteFixtureStep(manifest, manifest.Steps[0], input)
	body := execution.Artifact
	if err != nil {
		t.Fatal(err)
	}
	resultsRoot := os.Getenv("CARTULARY_TEST_RESULTS_DIR")
	if resultsRoot == "" {
		t.Fatal("CARTULARY_TEST_RESULTS_DIR is required for candidate mode")
	}
	directory = filepath.Join(resultsRoot, "graph-projection-fixture-candidate", fixtureID)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "candidate-input.json"), append(input, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "candidate-response.json"), append(body, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

func candidateProjectionInput(t *testing.T, fixtureID string) (map[string]any, bool) {
	t.Helper()
	switch fixtureID {
	case "GP-FIX-008", "GP-FIX-009", "GP-FIX-010", "GP-FIX-011", "GP-FIX-012", "GP-FIX-013", "GP-FIX-024", "GP-FIX-025", "GP-FIX-026", "GP-FIX-027":
	default:
		return nil, false
	}
	input := incidentGraphInput(t)
	switch fixtureID {
	case "GP-FIX-008":
		input["property_definitions"].([]any)[1].(map[string]any)["source_field_path"] = "properties.unused_missing"
	case "GP-FIX-009":
		delete(input["source_relationships"].([]any)[0].(map[string]any)["properties"].(map[string]any), "src_site")
	case "GP-FIX-010":
		delete(input["source_relationships"].([]any)[0].(map[string]any)["properties"].(map[string]any), "src_site")
		input["projection_config"].(map[string]any)["aggregation_rules"].([]any)[1].(map[string]any)["endpoint_grouping"].(map[string]any)["missing_endpoint_behavior"] = "exclude"
	case "GP-FIX-011":
		input["source_relationships"].([]any)[0].(map[string]any)["properties"].(map[string]any)["src_site"] = "remote"
	case "GP-FIX-012":
		input["projection_config"].(map[string]any)["entity_mappings"].([]any)[0].(map[string]any)["label_policy"] = "mapping_then_source"
	case "GP-FIX-013":
		input["property_definitions"] = append(input["property_definitions"].([]any), map[string]any{"property_definition_id": "pd_wild", "target_scope": "vertex", "target_kind": "*", "source_field_path": "properties.hostname", "projected_key": "wild_name", "projected_type": "string"})
	case "GP-FIX-024":
		input = minimalInput(t, "fix024")
		input["projection_config"].(map[string]any)["retention_policy"] = map[string]any{"retention_count": 1, "unexpected": true}
	case "GP-FIX-025":
		input = minimalInput(t, "fix025")
		input["source_snapshot_id"] = " invalid"
	case "GP-FIX-026":
		input["source_entities"] = append(input["source_entities"].([]any), map[string]any{"source_entity_id": "host1", "source_entity_kind": "host"})
	case "GP-FIX-027":
		config := input["projection_config"].(map[string]any)
		config["metadata_mappings"] = []any{map[string]any{"metadata_mapping_id": "mm_site", "target_scope": "vertex", "target_kind": "host_vertex", "source_field_path": "properties.site", "projected_metadata_key": "site_meta", "projected_type": "string"}}
		config["aggregation_rules"] = append(config["aggregation_rules"].([]any), map[string]any{"aggregation_rule_id": "agg_projected_by_site", "target_scope": "vertex", "input_scope": "projected_vertex", "input_kind": "host_vertex", "projected_kind": "site_from_projected", "grouping_keys": []any{"projected.metadata.site_meta"}, "missing_grouping_key_behavior": "error"})
	}
	return input, true
}
