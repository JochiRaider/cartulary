package graphprojection

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func fieldPathInputV2() map[string]any {
	input := twoEntityProjectionInputV2(false)
	config := input["projection_config"].(map[string]any)
	config["declared_source_relationship_kinds"] = []any{"relation"}
	config["relationship_mappings"] = []any{map[string]any{
		"mapping_rule_id": "relation-map", "source_relationship_kind": "relation", "projected_edge_kind": "edge",
		"inclusion_predicate": "always", "direction_policy": "preserve", "emit_reverse_edge": false,
		"label_policy": "mapping_only", "mapping_labels": []any{}, "required_property_keys": []any{}, "optional_property_keys": []any{},
	}}
	input["source_relationships"] = []any{map[string]any{
		"source_relationship_id": "relation-1", "source_relationship_kind": "relation", "src_source_entity_id": "entity-a", "dst_source_entity_id": "entity-b",
		"direction": "forward", "properties": map[string]any{"group": "relation"}, "metadata": map[string]any{}, "labels": []any{},
	}}
	for _, raw := range input["source_entities"].([]any) {
		entity := raw.(map[string]any)
		entity["properties"] = map[string]any{"group": entity["source_entity_kind"]}
	}
	return input
}

func checkCanonicalFieldPathsV2(t *testing.T) {
	for _, scope := range []string{"projected_vertex", "projected_edge"} {
		for _, use := range []string{"grouping", "source endpoint", "destination endpoint"} {
			for _, path := range []string{"kind", "source_entity_kind", "source_relationship_kind", "projected.vertex_kind", "projected.edge_kind"} {
				t.Run(scope+"/"+use+"/"+path, func(t *testing.T) {
					input := mixedAggregationInputV2(false)
					rules := input["projection_config"].(map[string]any)["aggregation_rules"].([]any)
					rule := rules[0].(map[string]any)
					if scope == "projected_edge" {
						rule["input_scope"], rule["input_kind"] = scope, "bridge"
					}
					if use == "grouping" {
						rule["grouping_keys"] = []any{path}
					} else {
						key := "src_grouping_keys"
						if use == "destination endpoint" {
							key = "dst_grouping_keys"
						}
						rule["endpoint_grouping"].(map[string]any)[key] = []any{path}
					}
					result, _, err := projectAggregationInputV2(t, input)
					if err == nil || !reflect.DeepEqual(result, ProjectionResultV2{}) {
						t.Fatal("projected input admitted a source-kind or unsupported kind path")
					}
				})
			}
		}
	}
	for _, use := range []string{"property", "metadata"} {
		for _, path := range []string{"kind", "source_entity_kind", "source_relationship_kind", "projected.vertex_kind", "projected.edge_kind"} {
			t.Run("graph_view/"+use+"/"+path, func(t *testing.T) {
				input := fieldPathInputV2()
				definition := aggregationPropertyV2("invalid", "graph_view", "*", path, "source_type", "string", "single_value")
				definition["missing_behavior"], definition["default_value"] = "default", "masked"
				if use == "property" {
					input["property_definitions"] = []any{definition}
				} else {
					delete(definition, "property_definition_id")
					delete(definition, "projected_key")
					definition["metadata_mapping_id"], definition["projected_metadata_key"] = "invalid", "source_type"
					input["projection_config"].(map[string]any)["metadata_mappings"] = []any{definition}
				}
				result, _, err := projectAggregationInputV2(t, input)
				if err == nil || !reflect.DeepEqual(result, ProjectionResultV2{}) {
					t.Fatal("graph metadata admitted an object kind path")
				}
			})
		}
	}
	for _, source := range []string{"entity", "relationship"} {
		canonical, wrong, target, kind, inputKind := "source_entity_kind", "source_relationship_kind", "vertex", "vertex", "kind-a"
		if source == "relationship" {
			canonical, wrong, target, kind, inputKind = "source_relationship_kind", "source_entity_kind", "edge", "edge", "relation"
		}
		for _, use := range []string{"filter", "inclusion", "property", "metadata", "grouping", "source endpoint", "destination endpoint"} {
			for _, path := range []string{canonical, "kind", wrong, "projected.vertex_kind", "projected.edge_kind", "properties.", "properties.a.b", strings.Repeat("x", 512)} {
				t.Run(source+"/"+use+"/"+path, func(t *testing.T) {
					input := fieldPathInputV2()
					config := input["projection_config"].(map[string]any)
					predicate := map[string]any{"field_path": path, "operator": "exists", "include_if_missing": false}
					switch use {
					case "filter":
						input["filters"].(map[string]any)[source+"_filters"] = []any{predicate}
					case "inclusion":
						for _, raw := range config[source+"_mappings"].([]any) {
							raw.(map[string]any)["inclusion_predicate"] = predicate
						}
					case "property", "metadata":
						definition := aggregationPropertyV2("source-type", target, kind, path, "source_type", "string", "single_value")
						// Defaults cannot turn an invalid path into an admitted path.
						definition["missing_behavior"], definition["default_value"] = "default", "masked"
						if use == "property" {
							input["property_definitions"] = []any{definition}
						} else {
							delete(definition, "property_definition_id")
							delete(definition, "projected_key")
							definition["metadata_mapping_id"], definition["projected_metadata_key"] = "source-type", "source_type"
							config["metadata_mappings"] = []any{definition}
						}
					case "grouping":
						rule := aggregationRuleInputV2("root", "vertex", "source_"+source, inputKind, "endpoint")
						rule["grouping_keys"] = []any{path}
						config["aggregation_rules"] = []any{rule}
					case "source endpoint", "destination endpoint":
						root := aggregationRuleInputV2("root", "vertex", "source_"+source, inputKind, "endpoint")
						edge := aggregationRuleInputV2("aggregate-edge", "edge", "source_"+source, inputKind, "aggregate-edge")
						key := "src_grouping_keys"
						if use == "destination endpoint" {
							key = "dst_grouping_keys"
						}
						edge["endpoint_grouping"].(map[string]any)[key] = []any{path}
						config["aggregation_rules"] = []any{root, edge}
					}
					result, _, err := projectAggregationInputV2(t, input)
					if path != canonical {
						var projectionErr *ProjectionErrorV2
						if !errors.As(err, &projectionErr) || projectionErr.RetryAction != "do_not_retry" || !reflect.DeepEqual(result, ProjectionResultV2{}) {
							t.Fatalf("invalid path admitted or partial result: %v", err)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					wantVertices, wantEdges := 2, 1
					if use == "grouping" {
						wantVertices++
					}
					if strings.HasSuffix(use, "endpoint") {
						wantVertices++
						wantEdges++
					}
					if len(result.Vertices) != wantVertices || len(result.Edges) != wantEdges {
						t.Fatalf("canonical field resolution lost output: %d/%d", len(result.Vertices), len(result.Edges))
					}
					if use == "property" || use == "metadata" {
						if target == "vertex" {
							for _, vertex := range result.Vertices {
								values := vertex.Properties
								if use == "metadata" {
									values = vertex.Metadata.MappedMetadata
								}
								if values["source_type"] != vertex.SourceEntityRef.SourceEntityKind {
									t.Fatalf("source kind was masked or misread: %#v", values)
								}
							}
						} else {
							values := result.Edges[0].Properties
							if use == "metadata" {
								values = result.Edges[0].Metadata.MappedMetadata
							}
							if values["source_type"] != "relation" {
								t.Fatalf("relationship kind was masked or misread: %#v", values)
							}
						}
					}
				})
			}
		}
	}
	// A path's length is measured in decoded UTF-8 bytes, independently of
	// JSON escapes. Keep lexical rejection before missing/default handling.
	input := fieldPathInputV2()
	input["filters"].(map[string]any)["entity_filters"] = []any{map[string]any{"field_path": "properties." + strings.Repeat("é", 256), "operator": "exists", "include_if_missing": true}}
	result, _, err := projectAggregationInputV2(t, input)
	if err == nil || !reflect.DeepEqual(result, ProjectionResultV2{}) {
		t.Fatal("oversized decoded property path admitted")
	}
}

func appendDefinitionV2(input map[string]any, definition map[string]any, metadata bool) {
	if metadata {
		delete(definition, "property_definition_id")
		definition["metadata_mapping_id"] = "observed"
		definition["projected_metadata_key"] = definition["projected_key"]
		delete(definition, "projected_key")
		config := input["projection_config"].(map[string]any)
		config["metadata_mappings"] = append(config["metadata_mappings"].([]any), definition)
	} else {
		input["property_definitions"] = append(input["property_definitions"].([]any), definition)
	}
}

func checkDefinitionProducerContextsV2(t *testing.T) {
	t.Run("producer sets", checkDefinitionProducerSetsV2)
	for _, test := range []struct{ name, scope, kind, path, wrong, want string }{
		{"entity to edge", "edge", "observed", "source_entity_kind", "source_relationship_kind", "kind-a"},
		{"relationship to vertex", "vertex", "observed", "source_relationship_kind", "source_entity_kind", "relation"},
		{"edge to vertex", "vertex", "final", "projected.edge_id", "projected.vertex_id", "bridge"},
		{"vertex to edge", "edge", "tail", "projected.vertex_id", "projected.edge_id", "final"},
	} {
		for _, metadata := range []bool{false, true} {
			for _, wrong := range []bool{false, true} {
				t.Run(test.name+map[bool]string{false: "/property", true: "/metadata"}[metadata]+map[bool]string{false: "/valid", true: "/wrong"}[wrong], func(t *testing.T) {
					input := mixedAggregationInputV2(false)
					if test.kind == "observed" {
						input = fieldPathInputV2()
						scope, kind := "source_entity", "kind-a"
						if test.scope == "vertex" {
							scope, kind = "source_relationship", "relation"
						}
						root := aggregationRuleInputV2("root", "vertex", scope, kind, "endpoint")
						rule := aggregationRuleInputV2("observed", test.scope, scope, kind, "observed")
						input["projection_config"].(map[string]any)["aggregation_rules"] = []any{root, rule}
					}
					path := test.path
					if wrong {
						path = test.wrong
					}
					definition := aggregationPropertyV2("observed", test.scope, test.kind, path, "observed", "string", "single_value")
					definition["missing_behavior"], definition["default_value"] = "default", "masked"
					appendDefinitionV2(input, definition, metadata)
					result, _, err := projectAggregationInputV2(t, input)
					if wrong {
						assertDefinitionContextRejectedV2(t, result, err, metadata)
						return
					}
					if err != nil {
						t.Fatalf("actual producer context rejected: %v", err)
					}
					want := test.want
					for _, vertex := range result.Vertices {
						if test.want == "final" && vertex.VertexKind == "final" {
							want = vertex.VertexID
						}
					}
					for _, edge := range result.Edges {
						if test.want == "bridge" && edge.EdgeKind == "bridge" {
							want = edge.EdgeID
						}
					}
					count := 0
					check := func(properties, mapped map[string]any) {
						t.Helper()
						values := properties
						if metadata {
							values = mapped
						}
						if values["observed"] != want {
							t.Fatalf("definition read wrong context: got %v want %s", values["observed"], want)
						}
						count++
					}
					if test.scope == "vertex" {
						for _, vertex := range result.Vertices {
							if vertex.VertexKind == test.kind {
								check(vertex.Properties, vertex.Metadata.MappedMetadata)
							}
						}
					} else {
						for _, edge := range result.Edges {
							if edge.EdgeKind == test.kind {
								check(edge.Properties, edge.Metadata.MappedMetadata)
							}
						}
					}
					if count == 0 {
						t.Fatal("definition was not exercised")
					}
				})
			}
		}
	}
	for _, metadata := range []bool{false, true} {
		for _, target := range []string{"vertex", "*"} {
			for _, empty := range []bool{false, true} {
				t.Run("all producers/"+target+map[bool]string{false: "/property", true: "/metadata"}[metadata]+map[bool]string{false: "/nonempty", true: "/empty"}[empty], func(t *testing.T) {
					input := fieldPathInputV2()
					input["projection_config"].(map[string]any)["aggregation_rules"] = []any{aggregationRuleInputV2("root", "vertex", "source_relationship", "relation", "vertex")}
					if empty {
						input["source_entities"], input["source_relationships"] = []any{}, []any{}
					}
					definition := aggregationPropertyV2("observed", "vertex", target, "source_entity_kind", "observed", "string", "single_value")
					definition["missing_behavior"], definition["default_value"] = "default", "masked"
					appendDefinitionV2(input, definition, metadata)
					result, _, err := projectAggregationInputV2(t, input)
					assertDefinitionContextRejectedV2(t, result, err, metadata)
				})
			}
		}
	}
}

func checkDefinitionProducerSetsV2(t *testing.T) {
	for _, metadata := range []bool{false, true} {
		for _, mode := range []string{"shared source property", "filtered incompatible", "projected incompatible", "reverse compatible", "reverse incompatible", "inert", "inert invalid"} {
			t.Run(mode+map[bool]string{false: "/property", true: "/metadata"}[metadata], func(t *testing.T) {
				input := fieldPathInputV2()
				config := input["projection_config"].(map[string]any)
				config["aggregation_rules"] = []any{aggregationRuleInputV2("root", "vertex", "source_relationship", "relation", "vertex")}
				scope, target, path, reject := "vertex", "*", "properties.group", false
				switch mode {
				case "filtered incompatible":
					path, reject = "source_entity_kind", true
					input["filters"].(map[string]any)["entity_filters"] = []any{map[string]any{"field_path": "source_entity_kind", "operator": "equals", "value": "absent", "include_if_missing": false}}
				case "projected incompatible":
					reject = true
					config["aggregation_rules"] = []any{aggregationRuleInputV2("root", "vertex", "projected_vertex", "vertex", "vertex")}
				case "reverse compatible", "reverse incompatible":
					scope, target = "edge", "reverse"
					mapping := config["relationship_mappings"].([]any)[0].(map[string]any)
					mapping["emit_reverse_edge"], mapping["reverse_edge_kind"], mapping["direction_policy"] = true, "reverse", "normalize_forward"
					config["aggregation_rules"] = []any{aggregationRuleInputV2("root", "vertex", "source_entity", "kind-a", "endpoint"), aggregationRuleInputV2("edge", "edge", "source_entity", "kind-a", "reverse")}
					if mode == "reverse incompatible" {
						path, reject = "source_relationship_kind", true
					}
				case "inert", "inert invalid":
					target, path = "unused", "projected.edge_id"
					if mode == "inert invalid" {
						path, reject = "kind", true
					}
				}
				definition := aggregationPropertyV2("observed", scope, target, path, "observed", "string", "single_value")
				definition["missing_behavior"], definition["default_value"] = "default", "masked"
				appendDefinitionV2(input, definition, metadata)
				result, _, err := projectAggregationInputV2(t, input)
				if mode == "inert invalid" {
					var projectionErr *ProjectionErrorV2
					if !errors.As(err, &projectionErr) || projectionErr.Code != "invalid_projection_request" || projectionErr.ReasonCode != "invalid_field_path" || !reflect.DeepEqual(result, ProjectionResultV2{}) {
						t.Fatalf("invalid inert syntax accepted: %v", err)
					}
					return
				}
				if reject {
					assertDefinitionContextRejectedV2(t, result, err, metadata)
					return
				}
				if err != nil {
					t.Fatalf("compatible configured contexts rejected: %v", err)
				}
			})
		}
	}
}

func assertDefinitionContextRejectedV2(t *testing.T, result ProjectionResultV2, err error, metadata bool) {
	t.Helper()
	reason := "invalid_field_path"
	if metadata {
		reason = "invalid_metadata_mapping"
	}
	var projectionErr *ProjectionErrorV2
	if !errors.As(err, &projectionErr) || projectionErr.Code != "projection_validation_failed" || projectionErr.ReasonCode != reason || projectionErr.RetryAction != "do_not_retry" || !reflect.DeepEqual(result, ProjectionResultV2{}) {
		t.Fatalf("wrong context did not reject during admission: %v", err)
	}
}
