package graphprojection

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func aggregationPropertyV2(id, scope, kind, source, key, valueType, merge string) map[string]any {
	return map[string]any{
		"property_definition_id": id, "target_scope": scope, "target_kind": kind,
		"source_field_path": source, "projected_key": key, "projected_type": valueType,
		"required": true, "missing_behavior": "error", "source_null_behavior": "error",
		"null_output_policy": "omit", "merge_behavior": merge,
	}
}

func mixedAggregationInputV2(sharedKind bool) map[string]any {
	input := twoEntityProjectionInputV2(false)
	for i, entry := range input["source_entities"].([]any) {
		entry.(map[string]any)["properties"] = map[string]any{"group": "all", "score": i + 2}
	}
	config := input["projection_config"].(map[string]any)
	for _, entry := range config["entity_mappings"].([]any) {
		entry.(map[string]any)["required_property_keys"] = []any{"group", "score"}
	}
	edgeKind := "bridge"
	if sharedKind {
		edgeKind = "endpoints"
	}
	config["aggregation_rules"] = []any{
		aggregationRuleInputV2("tail", "edge", "projected_vertex", "final", "tail"),
		aggregationRuleInputV2("final", "vertex", "projected_edge", edgeKind, "final"),
		aggregationRuleInputV2("bridge", "edge", "projected_vertex", "endpoints", edgeKind),
		aggregationRuleInputV2("root", "vertex", "projected_vertex", "vertex", "endpoints"),
	}
	definitions := []any{}
	for _, target := range []struct{ scope, kind, prefix string }{
		{"vertex", "vertex", "properties."}, {"vertex", "endpoints", "projected.properties."},
		{"edge", edgeKind, "projected.properties."}, {"vertex", "final", "projected.properties."}, {"edge", "tail", "projected.properties."},
	} {
		definitions = append(definitions,
			aggregationPropertyV2(target.scope+"-"+target.kind+"-group", target.scope, target.kind, target.prefix+"group", "group", "string", "single_value"),
			aggregationPropertyV2(target.scope+"-"+target.kind+"-score", target.scope, target.kind, target.prefix+"score", "score", "integer", "sum"))
	}
	input["property_definitions"] = definitions
	return input
}

func checkMixedAggregationExecutionV2(t *testing.T) {
	for _, shared := range []bool{false, true} {
		name := "distinct kinds"
		if shared {
			name = "shared vertex and edge kind"
		}
		t.Run(name, func(t *testing.T) {
			input := mixedAggregationInputV2(shared)
			first, _, err := projectAggregationInputV2(t, input)
			if err != nil {
				t.Fatal(err)
			}
			if len(first.Vertices) != 4 || len(first.Edges) != 2 {
				t.Fatalf("mixed chain produced %d vertices/%d edges, want 4/2", len(first.Vertices), len(first.Edges))
			}
			var root, final Vertex
			for _, vertex := range first.Vertices {
				if vertex.Metadata.AggregationRuleID == nil {
					continue
				}
				switch *vertex.Metadata.AggregationRuleID {
				case "root":
					root = vertex
				case "final":
					final = vertex
				}
			}
			var bridge, tail Edge
			for _, edge := range first.Edges {
				if edge.SrcVertexID != root.VertexID || edge.DstVertexID != root.VertexID {
					t.Fatal("mixed chain endpoint drift")
				}
				if *edge.Metadata.AggregationRuleID == "bridge" {
					bridge = edge
				} else {
					tail = edge
				}
			}
			for _, properties := range []map[string]any{root.Properties, bridge.Properties, final.Properties, tail.Properties} {
				if properties["score"] != int64(5) || properties["group"] != "all" {
					t.Fatalf("mixed chain lost merged values: %#v", properties)
				}
			}
			if len(root.Metadata.AggregationSourceRefs) != 2 || len(bridge.Metadata.AggregationSourceRefs) != 1 || len(final.Metadata.AggregationSourceRefs) != 1 || len(tail.Metadata.AggregationSourceRefs) != 1 {
				t.Fatal("mixed chain lost contributor references")
			}
			if bridge.Metadata.AggregationSourceRefs[0].RefID != root.VertexID || final.Metadata.AggregationSourceRefs[0].RefID != bridge.EdgeID || tail.Metadata.AggregationSourceRefs[0].RefID != final.VertexID {
				t.Fatal("mixed chain references the wrong predecessor")
			}
			want, err := canonicalJSON(normalizeCanonicalValueV2(first.outputResource()))
			if err != nil || sha256Hex(want) != first.CanonicalOutputSHA256 {
				t.Fatalf("canonical output mismatch: %v", err)
			}
			rules := input["projection_config"].(map[string]any)["aggregation_rules"].([]any)
			for shift := range len(rules) {
				input["projection_config"].(map[string]any)["aggregation_rules"] = append(append([]any{}, rules[shift:]...), rules[:shift]...)
				second, _, err := projectAggregationInputV2(t, input)
				got, canonicalErr := canonicalJSON(normalizeCanonicalValueV2(second.outputResource()))
				if err != nil || canonicalErr != nil || !bytes.Equal(want, got) || !reflect.DeepEqual(first, second) {
					t.Fatalf("permutation changed output: %v / %v", err, canonicalErr)
				}
			}
		})
	}
	for _, mode := range []string{"empty", "endpoint error", "endpoint exclude", "late grouping error", "cancelled"} {
		t.Run(mode, func(t *testing.T) {
			input := mixedAggregationInputV2(false)
			tail := input["projection_config"].(map[string]any)["aggregation_rules"].([]any)[0].(map[string]any)
			switch mode {
			case "empty":
				input["source_entities"] = []any{}
			case "endpoint error", "endpoint exclude":
				endpoint := tail["endpoint_grouping"].(map[string]any)
				endpoint["src_grouping_keys"] = []any{"projected.properties.absent"}
				if mode == "endpoint exclude" {
					endpoint["missing_endpoint_behavior"] = "exclude"
				}
			case "late grouping error":
				tail["grouping_keys"] = []any{"projected.properties.absent"}
			}
			result, body, err := projectAggregationInputV2(t, input)
			if mode == "cancelled" {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				result, err = ProjectV2(ctx, InvocationContextV2{GraphViewID: "mixed-graph", SourceOwnerID: "network_flow_activity"}, body)
			}
			switch mode {
			case "empty":
				if err != nil || len(result.Vertices) != 0 || len(result.Edges) != 0 {
					t.Fatalf("empty chain: %v", err)
				}
			case "endpoint exclude":
				if err != nil || len(result.Vertices) != 4 || len(result.Edges) != 1 {
					t.Fatalf("exclude returned %d/%d: %v", len(result.Vertices), len(result.Edges), err)
				}
			default:
				if err == nil || !reflect.DeepEqual(result, ProjectionResultV2{}) {
					t.Fatalf("chain failure returned partial result: %v", err)
				}
			}
		})
	}
}

func aggregationRuleInputV2(id, target, scope, inputKind, outputKind string) map[string]any {
	path := "properties.group"
	if strings.HasPrefix(scope, "projected_") {
		path = "projected.properties.group"
	}
	rule := map[string]any{
		"aggregation_rule_id": id, "target_scope": target, "input_scope": scope,
		"input_kind": inputKind, "projected_kind": outputKind, "grouping_keys": []any{path},
		"missing_grouping_key_behavior": "error", "property_merge_behavior": map[string]any{},
	}
	if target == "edge" {
		rule["edge_direction"] = "directed"
		rule["endpoint_grouping"] = map[string]any{
			"src_vertex_aggregation_rule_id": "root", "src_grouping_keys": []any{path},
			"dst_vertex_aggregation_rule_id": "root", "dst_grouping_keys": []any{path},
			"missing_endpoint_behavior": "error",
		}
	}
	return rule
}

func projectAggregationInputV2(t *testing.T, input map[string]any) (ProjectionResultV2, []byte, error) {
	t.Helper()
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	result, err := ProjectV2(context.Background(), InvocationContextV2{GraphViewID: "mixed-graph", SourceOwnerID: "network_flow_activity"}, body)
	return result, body, err
}

func checkAggregationDependencyScopesV2(t *testing.T) {
	rule := aggregationRuleInputV2
	root := func() map[string]any { return rule("root", "vertex", "source_entity", "kind-a", "endpoints") }
	for _, test := range []struct {
		name  string
		rules []any
		order []string
	}{
		{"vertex input ignores edge producer", []any{root(), rule("a", "vertex", "projected_vertex", "shared", "branch"), rule("b", "edge", "projected_vertex", "branch", "shared")}, []string{"a", "root", "b"}},
		{"edge input ignores vertex producer", []any{root(), rule("a", "edge", "projected_edge", "shared", "branch"), rule("b", "vertex", "projected_edge", "branch", "shared")}, []string{"root", "a", "b"}},
		{"shared names without input dependency", []any{root(), rule("a", "vertex", "source_entity", "kind-a", "shared"), rule("b", "edge", "source_entity", "kind-a", "shared")}, []string{"a", "root", "b"}},
		{"multiple producers and ready batches", []any{rule("z", "vertex", "source_entity", "kind-a", "seed"), rule("m", "vertex", "source_entity", "kind-b", "seed"), rule("a", "vertex", "projected_vertex", "seed", "leaf")}, []string{"m", "z", "a"}},
		{"input and repeated endpoints share one dependency", []any{root(), rule("a", "edge", "projected_vertex", "endpoints", "edge")}, []string{"root", "a"}},
		{"own output is not an input dependency", []any{rule("a", "vertex", "projected_vertex", "shared", "shared")}, []string{"a"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := twoEntityProjectionInputV2(false)
			// This admission test isolates dependency semantics from execution.
			input["source_entities"] = []any{}
			var first ProjectionResultV2
			for shift := range len(test.rules) {
				input["projection_config"].(map[string]any)["aggregation_rules"] = append(append([]any{}, test.rules[shift:]...), test.rules[:shift]...)
				result, body, err := projectAggregationInputV2(t, input)
				if err != nil {
					t.Fatal(err)
				}
				request, _, _, err := parseProjectionInputV2(body)
				if err != nil {
					t.Fatal(err)
				}
				var order []string
				for _, entry := range request.projectionConfig.AggregationRules {
					order = append(order, entry.AggregationRuleID)
				}
				if !reflect.DeepEqual(order, test.order) {
					t.Fatalf("order = %v, want %v", order, test.order)
				}
				if shift == 0 {
					first = result
				} else if !reflect.DeepEqual(first, result) {
					t.Fatal("permutation changed canonical result")
				}
			}
		})
	}
	for _, name := range []string{"mixed cycle", "missing endpoint", "edge endpoint"} {
		t.Run(name, func(t *testing.T) {
			input := twoEntityProjectionInputV2(false)
			edge := rule("edge", "edge", "source_entity", "kind-a", "edge-kind")
			rules := []any{root(), edge}
			switch name {
			case "mixed cycle":
				edge["input_scope"], edge["input_kind"], edge["grouping_keys"] = "projected_vertex", "vertex-kind", []any{"projected.properties.group"}
				rules = append(rules, rule("vertex", "vertex", "projected_edge", "edge-kind", "vertex-kind"))
			case "missing endpoint":
				edge["endpoint_grouping"].(map[string]any)["src_vertex_aggregation_rule_id"] = "absent"
			case "edge endpoint":
				edge["endpoint_grouping"].(map[string]any)["src_vertex_aggregation_rule_id"] = "edge"
			}
			input["projection_config"].(map[string]any)["aggregation_rules"] = rules
			result, _, err := projectAggregationInputV2(t, input)
			var projectionErr *ProjectionErrorV2
			if !errors.As(err, &projectionErr) || projectionErr.Code != "projection_validation_failed" || projectionErr.ReasonCode != "invalid_aggregation_rule" || projectionErr.RetryAction != "do_not_retry" || !reflect.DeepEqual(result, ProjectionResultV2{}) {
				t.Fatalf("invalid dependency returned result=%#v error=%v", result, err)
			}
		})
	}
}

func numericAggregationInputV2(values ...any) map[string]any {
	input := twoEntityProjectionInputV2(false)
	entities := input["source_entities"].([]any)
	for i, raw := range entities {
		entity := raw.(map[string]any)
		entity["source_entity_kind"] = "kind-a"
		entity["properties"] = map[string]any{"group": values[i]}
	}
	root := aggregationRuleInputV2("root", "vertex", "source_entity", "kind-a", "endpoint")
	edge := aggregationRuleInputV2("edge", "edge", "source_entity", "kind-a", "bridge")
	input["projection_config"].(map[string]any)["aggregation_rules"] = []any{edge, root}
	return input
}

func checkNumericAggregationTuplesV2(t *testing.T) {
	t.Run("null and missing", func(t *testing.T) {
		input := numericAggregationInputV2(nil, nil)
		result, _, err := projectAggregationInputV2(t, input)
		if err != nil || len(result.Vertices) != 3 || len(result.Edges) != 1 {
			t.Fatalf("null tuple failed: %v", err)
		}
		for _, raw := range input["source_entities"].([]any) {
			delete(raw.(map[string]any)["properties"].(map[string]any), "group")
		}
		result, _, err = projectAggregationInputV2(t, input)
		if err == nil || !reflect.DeepEqual(result, ProjectionResultV2{}) {
			t.Fatal("missing grouping key accepted")
		}
		rules := input["projection_config"].(map[string]any)["aggregation_rules"].([]any)
		for _, raw := range rules {
			raw.(map[string]any)["missing_grouping_key_behavior"] = "use_null"
		}
		rules[0].(map[string]any)["endpoint_grouping"].(map[string]any)["missing_endpoint_behavior"] = "exclude"
		result, _, err = projectAggregationInputV2(t, input)
		if err != nil || len(result.Vertices) != 3 || len(result.Edges) != 0 {
			t.Fatalf("missing/null policies conflated: %v", err)
		}
	})
	for _, values := range [][]any{
		{json.Number("1.5"), json.Number("2.5")},
		{json.Number("-9223372036854775808"), json.Number("9223372036854775807")},
		{json.Number("9223372036854775807"), json.Number("9223372036854775808.0")},
		{json.Number("1.7976931348623157e308"), json.Number("-1.7976931348623157e308")},
	} {
		t.Run("distinct/"+values[0].(json.Number).String(), func(t *testing.T) {
			input := numericAggregationInputV2(values...)
			result, _, err := projectAggregationInputV2(t, input)
			if err != nil || len(result.Vertices) != 4 || len(result.Edges) != 2 {
				t.Fatalf("distinct scalar groups collapsed: vertices=%d edges=%d err=%v", len(result.Vertices), len(result.Edges), err)
			}
			endpoints := map[string]string{}
			for _, vertex := range result.Vertices {
				if vertex.Metadata.AggregationRuleID != nil {
					refs := vertex.Metadata.AggregationSourceRefs
					if len(refs) != 1 {
						t.Fatal("distinct group contains multiple contributors")
					}
					endpoints[refs[0].RefID] = vertex.VertexID
				}
			}
			for _, edge := range result.Edges {
				refs := edge.Metadata.AggregationSourceRefs
				if len(refs) != 1 || edge.SrcVertexID != endpoints[refs[0].RefID] || edge.DstVertexID != endpoints[refs[0].RefID] {
					t.Fatal("numeric endpoint matched another group")
				}
			}
			entities := input["source_entities"].([]any)
			entities[0], entities[1] = entities[1], entities[0]
			second, _, err := projectAggregationInputV2(t, input)
			if err != nil || !reflect.DeepEqual(result, second) {
				t.Fatalf("numeric permutation changed result: %v", err)
			}
			completed, err := result.CompletedResult()
			if err != nil || completed.Validate(context.Background()) != nil {
				t.Fatalf("invalid numeric canonical output: %v", err)
			}
		})
	}
	for _, lexemes := range [][]string{{"1", "1.0", "1e0"}, {"1000000", "1000000.0", "1e6"}, {"9007199254740992", "9007199254740992.0", "9.007199254740992e15"}, {"-1000000", "-1000000.0", "-1e6"}, {"0", "-0", "-0.0", "0e0"}} {
		var baseline ProjectionResultV2
		for i, lexeme := range lexemes {
			result, _, err := projectAggregationInputV2(t, numericAggregationInputV2(json.Number(lexeme), json.Number(lexeme)))
			if err != nil || len(result.Vertices) != 3 || len(result.Edges) != 1 {
				t.Fatalf("equivalent numeric grouping failed: %s %v", lexeme, err)
			}
			if i == 0 {
				baseline = result
			} else if !reflect.DeepEqual(baseline.Vertices, result.Vertices) || !reflect.DeepEqual(baseline.Edges, result.Edges) {
				t.Errorf("equivalent lexeme %s changed groups, membership or endpoints", lexeme)
			}
		}
	}
	for _, endpoint := range []bool{false, true} {
		t.Run("array rejection/"+map[bool]string{false: "group", true: "endpoint"}[endpoint], func(t *testing.T) {
			input := numericAggregationInputV2(json.Number("1"), json.Number("2"))
			for _, raw := range input["source_entities"].([]any) {
				raw.(map[string]any)["properties"].(map[string]any)["invalid"] = []any{1, 2}
			}
			rule := input["projection_config"].(map[string]any)["aggregation_rules"].([]any)[0].(map[string]any)
			if endpoint {
				rule["endpoint_grouping"].(map[string]any)["src_grouping_keys"] = []any{"properties.invalid"}
				rule["endpoint_grouping"].(map[string]any)["missing_endpoint_behavior"] = "exclude"
			} else {
				rule["grouping_keys"] = []any{"properties.invalid"}
			}
			result, _, err := projectAggregationInputV2(t, input)
			var projectionErr *ProjectionErrorV2
			if !errors.As(err, &projectionErr) || projectionErr.Code != "projection_validation_failed" || projectionErr.ReasonCode != "invalid_aggregation_rule" || projectionErr.RetryAction != "do_not_retry" || !reflect.DeepEqual(result, ProjectionResultV2{}) {
				t.Fatalf("invalid scalar tuple did not fail closed: %v", err)
			}
		})
	}
}

func checkEmissionLimitsV2(t *testing.T) {
	t.Run("shared budget", checkSharedEmissionBudgetV2)
	entity := sourceEntity{SourceEntityID: "source", SourceEntityKind: "kind-a", Properties: map[string]any{"group": "all"}}
	groups := map[string][]contributor{"a": {{ID: "source", Entity: &entity}}, "b": {{ID: "source", Entity: &entity}}}
	run := &projectionWork{GraphViewID: "limits"}
	rule := aggregationRule{AggregationRuleID: "aggregate", TargetScope: "vertex", ProjectedKind: "aggregate"}
	t.Run("vertices", func(t *testing.T) {
		state := projectionState{vertices: make([]Vertex, graphProjectionLimits.MaxProjectedVertices-1), verticesByRuleAndDigest: map[string]map[string]Vertex{}}
		err := state.emitVertices(context.Background(), run, rule, groups)
		assertOutputLimitV2(t, err, "maximum_projected_vertices", graphProjectionLimits.MaxProjectedVertices)
		if len(state.vertices) != graphProjectionLimits.MaxProjectedVertices {
			t.Fatalf("retained %d vertices beyond budget", len(state.vertices))
		}
	})
	t.Run("edges and exclusions", func(t *testing.T) {
		digest, err := groupingTupleDigest("root", "vertex", "endpoint", []any{"all"})
		if err != nil {
			t.Fatal(err)
		}
		state := projectionState{edges: make([]Edge, graphProjectionLimits.MaxProjectedEdges-1), verticesByRuleAndDigest: map[string]map[string]Vertex{"root": {digest: {VertexID: "vx_endpoint"}}}, kinds: map[string]string{"root": "endpoint"}}
		edgeRule := rule
		edgeRule.TargetScope = "edge"
		edgeRule.endpointGrouping = &endpointGrouping{SourceVertexAggregationRuleID: "root", DestinationVertexAggregationRuleID: "root", SourceGroupingKeys: []string{"properties.group"}, DestinationGroupingKeys: []string{"properties.group"}, MissingEndpointBehavior: "error"}
		err = state.emitEdges(context.Background(), run, edgeRule, groups)
		assertOutputLimitV2(t, err, "maximum_projected_edges", graphProjectionLimits.MaxProjectedEdges)
		if len(state.edges) != graphProjectionLimits.MaxProjectedEdges {
			t.Fatalf("retained %d edges beyond budget", len(state.edges))
		}
		edgeRule.endpointGrouping.SourceGroupingKeys = []string{"properties.absent"}
		edgeRule.endpointGrouping.MissingEndpointBehavior = "exclude"
		if err := state.emitEdges(context.Background(), run, edgeRule, groups); err != nil || len(state.edges) != graphProjectionLimits.MaxProjectedEdges {
			t.Fatalf("excluded endpoint consumed capacity: %v", err)
		}
	})
	t.Run("cancel skipped contributors", func(t *testing.T) {
		run.Request.SourceEntities = make([]sourceEntity, 2048)
		run.Request.projectionConfig.AggregationRules = []aggregationRule{{AggregationRuleID: "root", TargetScope: "vertex", InputScope: "source_entity", InputKind: "absent", ProjectedKind: "endpoint"}}
		calls := 0
		ctx := newProjectionCheckpointContextV2(context.Background(), func(context.Context) error {
			calls++
			if calls == 2 {
				return context.Canceled
			}
			return nil
		})
		defer ctx.cancel()
		state := projectionState{}
		err := state.emitAggregations(ctx, run)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("skipped contributors escaped cancellation: %v", err)
		}
	})
}

func assertOutputLimitV2(t *testing.T, err error, key string, limit int) {
	t.Helper()
	var projectionErr *ProjectionErrorV2
	if !errors.As(err, &projectionErr) || projectionErr.Code != "projection_resource_limit_exceeded" || projectionErr.ReasonCode != key || projectionErr.RetryAction != "do_not_retry" || !reflect.DeepEqual(projectionErr.Details, map[string]any{"limit_key": key, "limit": limit, "observed": limit + 1}) {
		t.Fatalf("incorrect output resource failure: %v", err)
	}
}

func checkSharedEmissionBudgetV2(t *testing.T) {
	t.Run("builders do not run past limit", func(t *testing.T) {
		state := projectionState{vertices: make([]Vertex, graphProjectionLimits.MaxProjectedVertices), edges: make([]Edge, graphProjectionLimits.MaxProjectedEdges)}
		assertOutputLimitV2(t, state.emitVertex(t.Context(), &issueCollector{}, "", func() (Vertex, error) { t.Fatal("constructed over-limit vertex"); return Vertex{}, nil }), "maximum_projected_vertices", graphProjectionLimits.MaxProjectedVertices)
		assertOutputLimitV2(t, state.emitEdge(t.Context(), &issueCollector{}, func() (Edge, error) { t.Fatal("constructed over-limit edge"); return Edge{}, nil }), "maximum_projected_edges", graphProjectionLimits.MaxProjectedEdges)
	})
	t.Run("direct then multiple aggregate rules", func(t *testing.T) {
		input := numericAggregationInputV2(1, 1)
		config := input["projection_config"].(map[string]any)
		late := aggregationRuleInputV2("zz-late", "vertex", "source_entity", "kind-a", "late")
		late["grouping_keys"] = []any{"properties.absent"}
		config["aggregation_rules"] = []any{aggregationRuleInputV2("root", "vertex", "source_entity", "kind-a", "endpoint"), aggregationRuleInputV2("z-over", "vertex", "source_entity", "kind-a", "extra"), late}
		body, _ := json.Marshal(input)
		request, _, _, err := parseProjectionInputV2(body)
		if err != nil {
			t.Fatal(err)
		}
		run := &projectionWork{Request: request, GraphViewID: "limits"}
		state := projectionState{vertices: make([]Vertex, graphProjectionLimits.MaxProjectedVertices-3)}
		if err := state.emitDirectVertices(context.Background(), run); err != nil {
			t.Fatal(err)
		}
		err = state.emitAggregations(context.Background(), run)
		assertOutputLimitV2(t, err, "maximum_projected_vertices", graphProjectionLimits.MaxProjectedVertices)
		if len(state.vertices) != graphProjectionLimits.MaxProjectedVertices || len(state.verticesByRuleAndDigest["root"]) != 1 || len(state.verticesByRuleAndDigest["z-over"]) != 0 || len(run.issues.entries) != 0 {
			t.Fatal("budget reset or later rule executed after exhaustion")
		}
	})
	t.Run("direct and reverse share edge capacity", func(t *testing.T) {
		input := fieldPathInputV2()
		mapping := input["projection_config"].(map[string]any)["relationship_mappings"].([]any)[0].(map[string]any)
		mapping["emit_reverse_edge"], mapping["reverse_edge_kind"], mapping["direction_policy"] = true, "reverse", "normalize_forward"
		body, _ := json.Marshal(input)
		request, _, _, err := parseProjectionInputV2(body)
		if err != nil {
			t.Fatal(err)
		}
		run := &projectionWork{Request: request, GraphViewID: "limits"}
		state := projectionState{edges: make([]Edge, graphProjectionLimits.MaxProjectedEdges-1)}
		if err := state.emitDirectVertices(context.Background(), run); err != nil {
			t.Fatal(err)
		}
		err = state.emitDirectEdges(context.Background(), run)
		assertOutputLimitV2(t, err, "maximum_projected_edges", graphProjectionLimits.MaxProjectedEdges)
		if len(state.edges) != graphProjectionLimits.MaxProjectedEdges || state.edges[len(state.edges)-1].Metadata.IsReverseEdge {
			t.Fatal("reverse edge crossed shared capacity")
		}
	})
}

func checkPublicOutputLimitsV2(t *testing.T) {
	for _, edge := range []bool{false, true} {
		t.Run(map[bool]string{false: "direct plus aggregation", true: "direct plus reverse"}[edge], func(t *testing.T) {
			vertices, relationships := graphProjectionLimits.MaxProjectedVertices, 0
			if edge {
				vertices, relationships = 2, graphProjectionLimits.MaxProjectedEdges/2+1
			}
			decoder := json.NewDecoder(bytes.NewReader(maximumProjectionInputV2(t, vertices, relationships)))
			decoder.UseNumber()
			var input map[string]any
			if err := decoder.Decode(&input); err != nil {
				t.Fatal(err)
			}
			config := input["projection_config"].(map[string]any)
			key, limit := "maximum_projected_vertices", graphProjectionLimits.MaxProjectedVertices
			if edge {
				mapping := config["relationship_mappings"].([]any)[0].(map[string]any)
				mapping["emit_reverse_edge"], mapping["reverse_edge_kind"], mapping["direction_policy"] = true, "reverse", "normalize_forward"
				key, limit = "maximum_projected_edges", graphProjectionLimits.MaxProjectedEdges
			} else {
				rule := aggregationRuleInputV2("root", "vertex", "source_entity", "endpoint", "aggregate")
				rule["grouping_keys"] = []any{"source_entity_kind"}
				config["aggregation_rules"] = []any{rule}
			}
			result, _, err := projectAggregationInputV2(t, input)
			assertOutputLimitV2(t, err, key, limit)
			if !reflect.DeepEqual(result, ProjectionResultV2{}) {
				t.Fatal("resource failure exposed a partial result")
			}
		})
	}
}
