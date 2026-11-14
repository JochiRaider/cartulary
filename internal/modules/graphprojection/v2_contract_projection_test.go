package graphprojection

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dlclark/regexp2"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	contractgraphprojection "github.com/JochiRaider/cartulary/internal/gen/contractgraphprojection"
)

func TestGraphProjectionV2ContractProjection_Unit(t *testing.T) {
	t.Parallel()

	index := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/index.json")
	if index["contract_id"] != "cartulary.graph_projection_nlspec.v2.2.1" ||
		index["projection_schema_id"] != "graph_projection.v2" {
		t.Fatalf("Graph Projection v2 index identity drifted: %#v", index)
	}
	if _, present := index["limits"]; present {
		t.Fatal("Graph Projection index must not duplicate the semantic registry limits")
	}
	registry := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/semantic-registry.v1.json")
	limits := registry["semantic_limits"].(map[string]any)
	runtimeLimits := map[string]int{
		"cancellation_check_interval_items":      graphProjectionLimits.CancellationCheckIntervalItems,
		"maximum_aggregation_definitions":        graphProjectionLimits.MaxAggregationRules,
		"maximum_expanded_registry_declarations": graphProjectionLimits.MaxExpandedRegistryDeclarations,
		"maximum_array_items":                    graphProjectionLimits.MaxArrayItems,
		"maximum_entity_filters":                 graphProjectionLimits.MaxEntityFilters,
		"maximum_entity_mappings":                graphProjectionLimits.MaxEntityMappings,
		"maximum_result_bytes":                   graphProjectionLimits.MaxResultBytes,
		"maximum_input_json_depth":               graphProjectionLimits.MaxInputJSONDepth,
		"maximum_input_bytes":                    graphProjectionLimits.MaxInputBytes,
		"maximum_labels_per_object":              graphProjectionLimits.MaxLabelsPerObject,
		"maximum_metadata_mappings":              graphProjectionLimits.MaxMetadataMappings,
		"maximum_projected_edges":                graphProjectionLimits.MaxProjectedEdges,
		"maximum_projected_vertices":             graphProjectionLimits.MaxProjectedVertices,
		"maximum_property_definitions":           graphProjectionLimits.MaxPropertyDefinitions,
		"maximum_property_keys":                  graphProjectionLimits.MaxPropertyKeys,
		"maximum_relationship_filters":           graphProjectionLimits.MaxRelationshipFilters,
		"maximum_relationship_mappings":          graphProjectionLimits.MaxRelationshipMappings,
		"maximum_source_edges":                   graphProjectionLimits.MaxSourceRelationships,
		"maximum_source_vertices":                graphProjectionLimits.MaxSourceEntities,
		"maximum_traversal_depth":                graphProjectionLimits.MaxTraversalDepth,
		"maximum_traversal_kind_filters":         graphProjectionLimits.MaxTraversalKindFilters,
		"maximum_traversal_seeds":                graphProjectionLimits.MaxTraversalSeedVertices,
		"maximum_validation_issues":              graphProjectionLimits.MaxValidationIssues,
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
	checkProvenanceSchemaV2(t, schema)
	compiler := graphSchemaCompilerV2(t, schema)
	compile := func(name string) *jsonschema.Schema {
		t.Helper()
		s, err := compiler.Compile(graphSchemaURLV2 + "#/$defs/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	prop := compile("PropertyDefinition")
	def := aggregationPropertyV2("p", "vertex", "*", "properties.key", "key", "string", "first")
	if err := prop.Validate(def); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"metadata.key", `source_metadata["k.y"]`, `projected.metadata["k.y"]`} {
		def["source_field_path"] = path
		if prop.Validate(def) == nil {
			t.Fatalf("schema accepted property source %s", path)
		}
	}
	def["source_field_path"], def["target_scope"] = "properties.key", "graph_view"
	if prop.Validate(def) == nil {
		t.Fatal("schema admitted graph property")
	}
	decl := map[string]any{"target_scope": "vertex", "target_kind": "*", "projected_key": "key", "projected_type": "string", "required": false, "nullable_output": false, "missing_behavior": "omit", "source_null_behavior": "omit"}
	reg := compile("RegistryPropertyDeclaration")
	if err := reg.Validate(decl); err != nil {
		t.Fatal(err)
	}
	decl["target_scope"] = "graph_view"
	if reg.Validate(decl) == nil {
		t.Fatal("schema admitted graph property declaration")
	}
	r, _, err := projectAggregationInputV2(t, twoEntityProjectionInputV2(false))
	if err != nil {
		t.Fatal(err)
	}
	complete, err := r.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(complete.ResultJSON, &envelope); err != nil {
		t.Fatal(err)
	}
	current := compile("Result")
	if err := current.Validate(envelope); err != nil {
		t.Fatal(err)
	}
	envelope["properties"] = map[string]any{}
	if current.Validate(envelope) == nil {
		t.Fatal("schema admitted retired properties member")
	}
	paths, err := graphSchemaCompilerV2(t, schema).Compile(graphSchemaURLV2 + "#/$defs/FieldPath")
	if err != nil {
		t.Fatal(err)
	}
	for path, ok := range map[string]bool{"properties.kind": true, `projected.properties["a.b"]`: true, `metadata["a.\"b"]`: true, "kind": false, "properties.a.b": false, `properties["a"]`: false, `properties["a\u002eb"]`: false, "properties.*": false} {
		if (paths.Validate(path) == nil) != ok || validFieldPath(path) != ok {
			t.Fatalf("field path projection differs for %q", path)
		}
	}
	definitions := schema["$defs"].(map[string]any)
	for kind, want := range map[string][]any{"Vertex": {"direct", "aggregated"}, "Edge": {"direct", "reverse", "aggregated"}} {
		props := definitions[kind].(map[string]any)["properties"].(map[string]any)
		got := props[strings.ToLower(kind)+"_family"].(map[string]any)["enum"]
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s family contract = %v", kind, got)
		}
	}
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

	if _, present := contractgraphprojection.Index["contracts/graph-projection/storage-maintenance.v1.json"]; present {
		t.Fatal("retired Graph storage-maintenance contract remains packaged")
	}
	maintenance := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/storage-maintenance.v2.json")
	lockOrder, ok := maintenance["lock_order"].([]any)
	if maintenance["schema_id"] != "cartulary.graph_projection_storage_maintenance.v2" ||
		maintenance["owner_id"] != "module.graphprojection" ||
		maintenance["transaction_ownership"] != "borrowed" || !ok || len(lockOrder) != 2 ||
		lockOrder[0] != "projection_result" || lockOrder[1] != "source_declaration" {
		t.Fatalf("Graph Projection current storage-maintenance projection drifted: %#v", maintenance)
	}
	operations, ok := maintenance["operations"].([]any)
	if !ok || len(operations) != 2 {
		t.Fatalf("current Graph maintenance must expose only two cleanup operations: %#v", operations)
	}
	for i, expected := range []string{"lock_cleanup_candidate_tx_v1", "delete_locked_result_tx_v1"} {
		operation, ok := operations[i].(map[string]any)
		if !ok || operation["operation_id"] != expected || operation["requires_transaction"] != true || operation["maximum_rows"] != float64(1) {
			t.Fatalf("current Graph cleanup operation %d drifted: %#v", i, operations[i])
		}
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

const graphSchemaURLV2 = "https://graph.invalid/cartulary.graph_projection_schemas.v2"

func graphSchemaCompilerV2(t *testing.T, schema map[string]any) *jsonschema.Compiler {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	compiler.UseRegexpEngine(func(pattern string) (jsonschema.Regexp, error) {
		re, err := regexp2.Compile(pattern, regexp2.ECMAScript)
		if err != nil {
			return nil, err
		}
		re.MatchTimeout = time.Second
		return graphSchemaRegexp{re}, nil
	})

	if err := compiler.AddResource(graphSchemaURLV2, schema); err != nil {
		t.Fatal(err)
	}
	return compiler
}

func checkProvenanceSchemaV2(t *testing.T, schema map[string]any) {
	t.Helper()
	compiler := graphSchemaCompilerV2(t, schema)
	resultSchema, err := compiler.Compile(graphSchemaURLV2 + "#/$defs/Result")
	if err != nil {
		t.Fatal(err)
	}
	refSchema, err := compiler.Compile(graphSchemaURLV2 + "#/$defs/ContributorRef")
	if err != nil {
		t.Fatal(err)
	}
	ident, err := compiler.Compile(graphSchemaURLV2 + "#/$defs/Identifier")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		value               string
		schemaOK, runtimeOK bool
	}{
		{"a/b", false, false}, {"a\\b", false, false}, {"a\x00b", false, false}, {"a\u0085b", false, false}, {" left", false, false}, {"right\u3000", false, false}, {"a\nb", false, false}, {"a b", true, true}, {"a\u2028b", true, true}, {strings.Repeat("é", 127) + "x", true, true}, {strings.Repeat("é", 128), true, false},
	} {
		if (ident.Validate(tc.value) == nil) != tc.schemaOK || validIdentifierV2(tc.value) != tc.runtimeOK {
			t.Errorf("identifier schema/runtime disagree with contract: %q", tc.value)
		}
	}
	for _, tc := range []struct {
		scope, id string
		projected bool
	}{
		{"source_entity", "b", false}, {"source_relationship", "b", false},
		{"projected_vertex", "vx_" + strings.Repeat("0", 64), true}, {"projected_edge", "ed_" + strings.Repeat("0", 64), true},
	} {
		for key, valid := range map[string]bool{
			"csk_61.62..": !tc.projected, "csk_61.62.63.64": tc.projected,
			"csk_61.62..64": false, "csk_61.62.63.": false,
			"csk_6.62..": false, "csk_61.62.6.64": false,
			"csk_6A.62..": false, "csk_61.62.6A.64": false,
			"csk_61.62..\n": false, "csk_61.62.63.64\n": false,
		} {
			ref := map[string]any{"ref_kind": tc.scope, "ref_id": tc.id, "ref_kind_name": "a", "contributor_sort_key": key}
			if err := refSchema.Validate(ref); (err == nil) != valid {
				t.Errorf("%s key %q, want valid=%t: %v", tc.scope, key, valid, err)
			}
		}
	}
	var vectors map[string]CompletedResultV2
	if err := json.Unmarshal(resultVectorsV2, &vectors); err != nil {
		t.Fatal(err)
	}
	wantShapes := map[string][2]int{
		"empty": {0, 0}, "direct": {2, 1}, "reverse": {2, 2},
		"aggregate": {4, 2}, "null-entity": {2, 1}, "null-relationship": {2, 1},
	}
	if len(vectors) != len(wantShapes) {
		t.Fatalf("schema vectors = %d, want %d", len(vectors), len(wantShapes))
	}
	for name, shape := range wantShapes {
		vector, present := vectors[name]
		if !present {
			t.Fatalf("missing schema vector %s", name)
		}
		var result map[string]any
		if err := json.Unmarshal(vector.ResultJSON, &result); err != nil {
			t.Fatal(err)
		}
		vertices, edges := result["vertices"].([]any), result["edges"].([]any)
		if len(vertices) != shape[0] || len(edges) != shape[1] {
			t.Fatalf("%s schema witnesses = %d/%d, want %v", name, len(vertices), len(edges), shape)
		}
		if name == "aggregate" || name == "reverse" {
			families := map[string]int{}
			for _, raw := range vertices {
				families["vertex/"+raw.(map[string]any)["vertex_family"].(string)]++
			}
			for _, raw := range edges {
				families["edge/"+raw.(map[string]any)["edge_family"].(string)]++
			}
			want := map[string]int{"vertex/direct": 2, "edge/direct": 1, "edge/reverse": 1}
			if name == "aggregate" {
				want = map[string]int{"vertex/direct": 2, "vertex/aggregated": 2, "edge/aggregated": 2}
			}
			if !reflect.DeepEqual(families, want) {
				t.Fatalf("%s schema families = %v, want %v", name, families, want)
			}
		}
		wantValid := name != "null-entity" && name != "null-relationship"
		if err := resultSchema.Validate(result); (err == nil) != wantValid {
			t.Errorf("schema admission for %s: %v", name, err)
		}
		if name == "aggregate" {
			registry := result["schema_registry"].(map[string]any)
			for _, scope := range []string{"vertex", "edge"} {
				collection, field := "vertices", "source_entity_kinds"
				if scope == "edge" {
					collection, field = "edges", "source_relationship_kinds"
				}
				for _, raw := range result[collection].([]any) {
					row := raw.(map[string]any)
					if row[scope+"_family"] != "aggregated" {
						continue
					}
					saved := row["sort_key"]
					for _, key := range []string{"", "a", "sk_" + strings.Repeat("0", 63), "sk_" + strings.Repeat("A", 64), saved.(string) + "\n"} {
						row["sort_key"] = key
						if resultSchema.Validate(result) == nil {
							t.Fatalf("%s admits malformed aggregate sort key %q", scope, key)
						}
					}
					row["sort_key"] = saved
				}
				for _, raw := range registry[scope+"_kinds"].([]any) {
					row := raw.(map[string]any)
					if len(row[field].([]any)) != 0 {
						continue
					}
					row["source_labels_preserved"] = true
					if resultSchema.Validate(result) == nil {
						t.Fatalf("%s schema admits aggregate-only label preservation", scope)
					}
					row["source_labels_preserved"] = false
				}
			}
			reg := result["schema_registry"].(map[string]any)
			for _, field := range []string{"property_keys", "metadata_keys"} {
				saved := reg[field]
				reg[field] = []any{nil}
				if resultSchema.Validate(result) == nil {
					t.Errorf("schema admits null %s row", field)
				}
				reg[field] = saved
			}
			kind := reg["edge_kinds"].([]any)[0].(map[string]any)
			kind["extra"] = false
			if resultSchema.Validate(result) == nil {
				t.Error("schema admits unknown registry field")
			}
			delete(kind, "extra")
			saved := kind["directions"]
			kind["directions"] = []any{}
			if resultSchema.Validate(result) == nil {
				t.Error("schema admits empty directions for a configured edge kind")
			}
			kind["directions"] = saved
		}
		if name == "reverse" {
			for _, raw := range result["edges"].([]any) {
				edge := raw.(map[string]any)
				key := edge["sort_key"]
				for _, bad := range []string{"a", key.(string) + "\n"} {
					edge["sort_key"] = bad
					if resultSchema.Validate(result) == nil {
						t.Errorf("schema accepted %s sort key", edge["edge_family"])
					}
				}
				edge["sort_key"] = key
			}
			reg := result["schema_registry"].(map[string]any)
			for _, scope := range []string{"vertex", "edge"} {
				row := reg[scope+"_kinds"].([]any)[0].(map[string]any)
				field := "source_entity_kinds"
				if scope == "edge" {
					field = "source_relationship_kinds"
				}
				saved := row[field]
				row[field] = []any{}
				if resultSchema.Validate(result) == nil {
					t.Errorf("schema accepts producerless %s", scope)
				}
				row["aggregation_rule_ids"] = []any{"group"}
				if err := resultSchema.Validate(result); err != nil {
					t.Errorf("schema rejects aggregation-only kind: %v", err)
				}
				row[field], row["aggregation_rule_ids"] = saved, []any{}
			}
		}
		if name == "direct" {
			for _, collection := range []string{"vertices", "edges"} {
				row := result[collection].([]any)[0].(map[string]any)
				saved := row["labels"]
				oversized := make([]any, graphProjectionLimits.MaxLabelsPerObject+1)
				for i := range oversized {
					oversized[i] = strconv.Itoa(i)
				}
				for _, invalid := range [][]any{oversized, {""}, {"duplicate", "duplicate"}, {strings.Repeat("x", graphProjectionLimits.MaxLabelBytes+1)}} {
					row["labels"] = invalid
					if resultSchema.Validate(result) == nil {
						t.Errorf("%s schema admits invalid object labels", collection)
					}
				}
				row["labels"] = saved
			}
			row := result["schema_registry"].(map[string]any)["vertex_kinds"].([]any)[0].(map[string]any)
			delete(row, "aggregation_rule_ids")
			if resultSchema.Validate(result) == nil {
				t.Error("schema admits missing registry producer list")
			}
		}
		if name == "aggregate" {
			for _, raw := range result["vertices"].([]any) {
				vertex := raw.(map[string]any)
				if vertex["vertex_family"] == "aggregated" {
					refs := vertex["metadata"].(map[string]any)["aggregation_source_refs"].([]any)
					refs[0].(map[string]any)["ref_kind"] = "unknown"
					if resultSchema.Validate(result) == nil {
						t.Error("schema admits untyped aggregate contributor")
					}
					break
				}
			}
		}
	}
}

type graphSchemaRegexp struct{ *regexp2.Regexp }

func (re graphSchemaRegexp) MatchString(s string) bool {
	ok, err := re.Regexp.MatchString(s)
	return err == nil && ok
}
