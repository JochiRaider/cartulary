package graphprojection

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func checkInputDepthAndPathsV2(t *testing.T) {
	inv := InvocationContextV2{GraphViewID: "depth-test", SourceOwnerID: "owner"}
	input := fieldPathInputV2()
	input["source_metadata"] = map[string]any{}
	body := inputJSONV2(t, input)
	nested := func(depth int) []byte {
		value := "null"
		for i := 2; i < depth; i++ {
			if i%2 == 0 {
				value = "[" + value + "]"
			} else {
				value = `{"x":` + value + "}"
			}
		}
		return bytes.Replace(body, []byte(`"source_metadata":{}`), []byte(`"source_metadata":{"root":`+value+"}"), 1)
	}
	for _, depth := range []int{127, 128} {
		if _, err := ProjectV2(t.Context(), inv, nested(depth)); err != nil {
			t.Fatalf("depth %d: %v", depth, err)
		}
	}
	for _, body := range [][]byte{nested(129), append(nested(129), '!'), []byte(strings.Repeat(`{"`+strings.Repeat("x", 255)+`":`, 4000) + "null" + strings.Repeat("}", 4000))} {
		expectCollectionOverflowV2(t, body, "maximum_input_json_depth", 128)
	}
	expectInputFailureV2(t, append([]byte(`{"x":0,"x":`), nested(129)...), "invalid_projection_request", "duplicate_object_member", map[string]any{"field": "$.x"})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := ProjectV2(ctx, inv, body); err == nil {
		t.Fatal("cancelled decoder returned a result")
	}
	key := strings.Repeat("k", 255)
	bad := []byte(strings.Repeat(`{"`+key+`":`, 10) + `{"x":1,"x":2}` + strings.Repeat("}", 10))
	expectInputFailureV2(t, bad, "invalid_projection_request", "duplicate_object_member", map[string]any{"field": "$" + strings.Repeat("."+key, 3)})
	// Path components are references. Zero prefixes are rendered on success.
	var shallow, deep inputPathV2
	shallow = append(shallow, key)
	for range 120 {
		deep = append(deep, key)
	}
	if testing.AllocsPerRun(10, func() { _ = deep.render("$") }) > testing.AllocsPerRun(10, func() { _ = shallow.render("$") })+5 {
		t.Fatal("diagnostic allocation grew beyond the bounded prefix")
	}

	for _, prefix := range []string{"properties", "metadata", "source_metadata", "projected.properties", "projected.metadata"} {
		for _, key := range []string{"customer.name", `a."é`, "kind"} {
			path := prefix + "." + key
			if strings.Contains(key, ".") {
				path = prefix + "[" + canonicalJSONString(key) + "]"
			}
			f, ok := classifyFieldPath(path)
			if !ok || f.key != key {
				t.Fatalf("invalid literal resolution %q: %#v", path, f)
			}
		}
	}
	for _, bad := range []string{`properties.a.b`, `properties["a"]`, `properties["a\u002eb"]`, `properties["a.b"].more`, "kind", "projected.metadata.mapping_rule_id"} {
		if validFieldPath(bad) {
			t.Fatalf("accepted %q", bad)
		}
	}
	for _, raw := range input["source_entities"].([]any) {
		row := raw.(map[string]any)
		row["metadata"] = map[string]any{"customer.name": "literal", "customer": map[string]any{"name": "nested", "a/b": true, "*": true}}
		props := row["properties"].(map[string]any)
		props["kind"] = "qualified"
		props["customer.name"] = "property"
	}
	def := aggregationPropertyV2("dot", "vertex", "vertex", `properties["customer.name"]`, "out.dot", "string", "first")
	input["property_definitions"] = []any{def, aggregationPropertyV2("key", "vertex", "vertex", "properties.kind", "kind", "string", "first")}
	appendDefinitionV2(input, aggregationPropertyV2("meta", "vertex", "vertex", `metadata["customer.name"]`, "out.dot", "string", "first"), true)
	first, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Vertices) != 2 {
		t.Fatalf("literal-field witnesses = %d vertices, want 2", len(first.Vertices))
	}
	for _, v := range first.Vertices {
		if v.Properties["out.dot"] != "property" || v.Metadata.MappedMetadata["out.dot"] != "literal" || v.Properties["kind"] != "qualified" {
			t.Fatalf("incorrect key selection: %#v", v.Properties)
		}
	}
	requireReboundAdmissionV2(t, first, true)
	defs := input["property_definitions"].([]any)
	defs[0], defs[1] = defs[1], defs[0]
	second, _, err := projectAggregationInputV2(t, input)
	if err != nil || second.ResultBindingV2() != first.ResultBindingV2() {
		t.Fatal("permuted dotted-key definitions changed identity", err)
	}
}

func checkEnvelopeCapacityV2(t *testing.T) {
	inv := InvocationContextV2{GraphViewID: "wire-budget", SourceOwnerID: "owner"}
	cases := map[string]map[string]any{"empty": graphProjectionV2GoldenFixture(t)["input"].(map[string]any), "reverse": reverseInputV2(), "aggregate": mixedAggregationInputV2(false)}
	input := twoEntityProjectionInputV2(false)
	def := aggregationPropertyV2("default", "vertex", "vertex", "properties.absent", "flags", "boolean_array", "first")
	def["default_value"], def["missing_behavior"] = make([]bool, 128), "default"
	input["property_definitions"] = []any{def}
	rows := []any{}
	for i := range 64 {
		rows = append(rows, map[string]any{"source_entity_id": fmt.Sprintf("entity-%d", i), "source_entity_kind": "kind-a", "properties": map[string]any{}, "labels": []any{}, "metadata": map[string]any{}})
	}
	input["source_entities"] = rows
	cases["repeated default"] = input
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			body := inputJSONV2(t, input)
			result, err := ProjectV2(t.Context(), inv, body)
			if err != nil {
				t.Fatal(err)
			}
			completed, err := result.CompletedResult()
			if err != nil {
				t.Fatal(err)
			}
			n := len(completed.ResultJSON)
			same, err := projectWithResultLimitV2(t.Context(), inv, body, n)
			if err != nil || same.ResultBindingV2() != result.ResultBindingV2() {
				t.Fatalf("capacity %d: %v", n, err)
			}
			failed, err := projectWithResultLimitV2(t.Context(), inv, body, n-1)
			assertOutputLimitV2(t, err, "maximum_result_bytes", n-1)
			if !reflect.DeepEqual(failed, ProjectionResultV2{}) {
				t.Fatal("partial result on overflow")
			}
		})
	}
	for _, v := range []any{nil, "é\n\\\"", []any{int64(120), float64(1e6)}, map[string]any{"escaped\"": true}} {
		b, err := canonicalJSON(v)
		if err != nil {
			t.Fatal(err)
		}
		budget := newResultBudgetV2(len(b))
		if err = budget.charge(v); err != nil || budget.remaining != 0 {
			t.Fatal("count differs from canonical encoding", err)
		}
	}
}
func checkObservedScopeDefaultsV2(t *testing.T) {
	input := mixedAggregationInputV2(false)
	config := input["projection_config"].(map[string]any)
	config["default_vertex_labels"], config["default_edge_labels"] = []any{"vertex-default"}, []any{"edge-default"}
	r, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Vertices) != 4 || len(r.Edges) != 2 {
		t.Fatalf("scope-default witnesses = %d vertices/%d edges, want 4/2", len(r.Vertices), len(r.Edges))
	}
	requireReboundAdmissionV2(t, r, true)
	for i := range r.Vertices {
		old := r.Vertices[i].Labels
		r.Vertices[i].Labels = nil
		requireReboundAdmissionV2(t, r, false)
		r.Vertices[i].Labels = old
	}
	r.Vertices = nil
	r.Edges = nil
	requireReboundAdmissionV2(t, r, true)
	r.SchemaRegistry.VertexKinds[0].Labels = []string{"wrong"}
	requireReboundAdmissionV2(t, r, false)
}

// Use a dotted key across both source and projected scopes, including the
// dependency between an edge group and each endpoint producer.
func checkDottedAggregationV2(t *testing.T) {
	input := mixedAggregationInputV2(false)
	const key = "group.key"
	const source = `properties["group.key"]`
	const projected = `projected.properties["group.key"]`
	for _, raw := range input["source_entities"].([]any) {
		p := raw.(map[string]any)["properties"].(map[string]any)
		p[key] = p["group"]
		delete(p, "group")
	}
	filter := map[string]any{"field_path": source, "operator": "equals", "value": "all", "include_if_missing": false}
	input["filters"].(map[string]any)["entity_filters"] = []any{filter}
	cfg := input["projection_config"].(map[string]any)
	for _, raw := range cfg["entity_mappings"].([]any) {
		m := raw.(map[string]any)
		m["required_property_keys"] = []any{key, "score"}
		m["inclusion_predicate"] = filter
	}
	for _, raw := range cfg["aggregation_rules"].([]any) {
		r := raw.(map[string]any)
		r["grouping_keys"] = []any{projected}
		if ep, ok := r["endpoint_grouping"].(map[string]any); ok {
			ep["src_grouping_keys"], ep["dst_grouping_keys"] = []any{projected}, []any{projected}
		}
	}
	for _, raw := range input["property_definitions"].([]any) {
		def := raw.(map[string]any)
		if def["projected_key"] == "group" {
			def["projected_key"] = key
			if def["source_field_path"] == "properties.group" {
				def["source_field_path"] = source
			} else {
				def["source_field_path"] = projected
			}
		}
	}
	r, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Edges) != 2 {
		t.Fatalf("missing grouped edges: %d", len(r.Edges))
	}
	for _, edge := range r.Edges {
		if edge.Properties[key] != "all" {
			t.Fatal("incorrect grouped value", edge.Properties)
		}
	}
	requireReboundAdmissionV2(t, r, true)
}

func checkMergeCapacityV2(t *testing.T) {
	short := []any{"x"}
	long := []any{strings.Repeat("a", 80)}
	for _, behavior := range []string{"first", "last", "set", "ordered_list"} {
		b := newResultBudgetV2(4)
		m := newProjectedMergeV2(behavior, b)
		a, z := short, long
		if behavior == "last" {
			a, z = long, short
		}
		if behavior == "set" {
			z = short
		}
		if err := m.add(a, false); err != nil {
			t.Fatal(err)
		}
		err := m.add(z, false)
		if behavior == "ordered_list" {
			assertOutputLimitV2(t, err, "maximum_result_bytes", 4)
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		v, ok, conflict := m.result("string_array")
		if !ok || conflict || !reflect.DeepEqual(v, short) || b.remaining != 4 {
			t.Fatalf("incorrect %s reduction", behavior)
		}
	}
}
