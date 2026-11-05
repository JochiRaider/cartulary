package graphprojection

import (
	"bytes"
	"context"
	"errors"
	"maps"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func checkUnicodeCutoverV2(t *testing.T) {
	input := graphProjectionV2GoldenFixture(t)["input"].(map[string]any)
	input["source_metadata"] = map[string]any{}
	base := inputJSONV2(t, input)
	body := func(raw string) []byte {
		return bytes.Replace(base, []byte(`"source_metadata":{}`), []byte(`"source_metadata":`+raw), 1)
	}
	for _, raw := range []string{`{"\ud800":0}`, `{"a":"\udfff"}`, `{"\udc00":0}`, `{"a":"\ud800x"}`} {
		t.Run(raw, func(t *testing.T) {
			expectInputFailureV2(t, body(raw), "invalid_projection_request", "invalid_json_syntax", map[string]any{})
		})
	}
	project := func(raw string) ResultBindingV2 {
		r, err := ProjectV2(t.Context(), InvocationContextV2{GraphViewID: "v", SourceOwnerID: "o"}, body(raw))
		if err != nil {
			t.Fatal(err)
		}
		return r.ResultBindingV2()
	}
	if project(`{"😀":"�"}`) != project(`{"\ud83d\ude00":"\ufffd"}`) {
		t.Fatal("equivalent Unicode differs")
	}
	project(`{"a":"\\ud800"}`)
	deep := strings.Repeat("[", 128) + "null" + strings.Repeat("]", 128)
	expectInputFailureV2(t, []byte(`{"a":0,"a":"\ud800"}`), "invalid_projection_request", "duplicate_object_member", map[string]any{"field": "$.a"})
	expectCollectionOverflowV2(t, []byte(`{"a":`+deep+`,"b":"\ud800"}`), "maximum_input_json_depth", 128)
	for _, suffix := range []string{`,"a":0}`, `,"b":` + deep + "}"} {
		expectInputFailureV2(t, []byte(`{"a":"\ud800"`+suffix), "invalid_projection_request", "invalid_json_syntax", map[string]any{})
	}
}

func checkMetadataChannelsV2(t *testing.T) {
	for _, tc := range []struct{ path, kind, missing string }{
		{"metadata.field", "vertex", "error"}, {`metadata["a.b"]`, "*", "omit"}, {"projected.metadata.field", "endpoints", "default"}, {`projected.metadata["a.b"]`, "inert", "default"}, {"source_metadata.field", "inert", "omit"},
	} {
		input := mixedAggregationInputV2(false)
		input["source_entities"] = []any{}
		def := aggregationPropertyV2("decl", "vertex", tc.kind, tc.path, "probe", "string", "first")
		def["missing_behavior"], def["required"] = tc.missing, false
		if tc.missing == "default" {
			def["default_value"] = "masked"
		}
		input["property_definitions"] = []any{def}
		expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "invalid_field_path", map[string]any{"issue_count": 1})
	}
	input := fieldPathInputV2()
	input["property_definitions"] = []any{aggregationPropertyV2("decl", "graph_view", "*", "source_metadata.field", "graph", "string", "first")}
	expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "invalid_property_definition", map[string]any{"issue_count": 1})

	// Carry one literal dotted key through source predicates and each projected
	// grouping stage. Only the explicit metadata mappings may expose it.
	input = mixedAggregationInputV2(false)
	const raw = `metadata["group.key"]`
	const projected = `projected.metadata["group.key"]`
	filter := map[string]any{"field_path": raw, "operator": "equals", "value": "all", "include_if_missing": false}
	input["filters"].(map[string]any)["entity_filters"] = []any{filter}
	config := input["projection_config"].(map[string]any)
	for _, e := range input["source_entities"].([]any) {
		e.(map[string]any)["metadata"] = map[string]any{"group.key": "all"}
	}
	for _, m := range config["entity_mappings"].([]any) {
		m.(map[string]any)["inclusion_predicate"] = filter
	}
	for _, a := range config["aggregation_rules"].([]any) {
		r := a.(map[string]any)
		r["grouping_keys"] = []any{projected}
		if ep, ok := r["endpoint_grouping"].(map[string]any); ok {
			ep["src_grouping_keys"], ep["dst_grouping_keys"] = []any{projected}, []any{projected}
		}
	}
	for _, target := range []struct{ scope, kind, path string }{{"vertex", "vertex", raw}, {"vertex", "endpoints", projected}, {"edge", "bridge", projected}, {"vertex", "final", projected}, {"edge", "tail", projected}} {
		d := aggregationPropertyV2(target.kind, target.scope, target.kind, target.path, "group.key", "string", "single_value")
		appendDefinitionV2(input, d, true)
		d["metadata_mapping_id"] = target.kind
	}
	r, _, err := projectAggregationInputV2(t, input)
	if err != nil || len(r.Edges) != 2 || len(r.Vertices) != 4 {
		t.Fatalf("metadata selection and grouping: %v", err)
	}
	for _, edge := range r.Edges {
		if _, promoted := edge.Properties["group.key"]; promoted || edge.Metadata.MappedMetadata["group.key"] != "all" {
			t.Fatal("incorrect metadata emission")
		}
	}
	requireReboundAdmissionV2(t, r, true)
}

func checkPairedClaimsV2(t *testing.T) {
	for _, metadata := range []bool{false, true} {
		for _, tc := range []struct {
			name, typ     string
			first, second any
			omit          bool
		}{
			{"value", "string", "a", "b", false}, {"omitted or null", "string", nil, nil, true}, {"array order", "integer_array", []any{1, 2}, []any{2, 1}, false}, {"numeric spelling", "number", int64(1000000), float64(1000000), false},
		} {
			t.Run(tc.name+strconv.FormatBool(metadata), func(t *testing.T) {
				input := reverseInputV2()
				input["projection_config"].(map[string]any)["relationship_mappings"].([]any)[0].(map[string]any)["reverse_edge_kind"] = "edge"
				d := aggregationPropertyV2("paired", "edge", "edge", "properties.absent", "field", tc.typ, "single_value")
				d["required"], d["missing_behavior"], d["source_null_behavior"], d["null_output_policy"] = false, "omit", "emit_null", "emit_null"
				appendDefinitionV2(input, d, metadata)
				r, _, err := projectAggregationInputV2(t, input)
				if err != nil || len(r.Edges) != 2 {
					t.Fatalf("pair: %v", err)
				}
				fields := func(i int) map[string]any {
					if metadata {
						return r.Edges[i].Metadata.MappedMetadata
					}
					return r.Edges[i].Properties
				}
				fields(0)["field"], fields(1)["field"] = tc.first, tc.first
				requireReboundAdmissionV2(t, r, true)
				if tc.omit {
					delete(fields(1), "field")
				} else {
					fields(1)["field"] = tc.second
				}
				requireReboundAdmissionV2(t, r, false)
			})
		}
	}
	input := reverseInputV2()
	cfg := input["projection_config"].(map[string]any)
	mapping := cfg["relationship_mappings"].([]any)[0].(map[string]any)
	mapping["label_policy"] = "preserve_source"
	other := maps.Clone(mapping)
	other["mapping_rule_id"], other["source_relationship_kind"], other["projected_edge_kind"] = "extra", "other", "second"
	cfg["declared_source_relationship_kinds"] = []any{"relation", "other"}
	cfg["relationship_mappings"] = []any{mapping, other}
	input["source_relationships"].([]any)[0].(map[string]any)["labels"] = []any{"observed"}
	r, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	for i := range r.SchemaRegistry.EdgeKinds {
		k := &r.SchemaRegistry.EdgeKinds[i]
		if k.EdgeKind == "edge" {
			k.Labels = []string{"observed"}
			k.SourceLabelsPreserved = false
		}
	}
	requireReboundAdmissionV2(t, r, false)
	for i := range r.SchemaRegistry.EdgeKinds {
		if r.SchemaRegistry.EdgeKinds[i].EdgeKind == "back" {
			r.SchemaRegistry.EdgeKinds[i].Labels = []string{"observed"}
		}
	}
	requireReboundAdmissionV2(t, r, true)
	r.Edges = nil
	for i := range r.SchemaRegistry.EdgeKinds {
		if r.SchemaRegistry.EdgeKinds[i].EdgeKind == "second" {
			r.SchemaRegistry.EdgeKinds[i].SourceLabelsPreserved = false
		}
	}
	requireReboundAdmissionV2(t, r, false)
}

func checkSingleValueWorkV2(t *testing.T) {
	for _, firstDefault := range []bool{false, true} {
		for _, size := range []int{4, 1000} {
			// Every comparison encodes its candidate. The first actual comparison also encodes the retained operand.
			count := 0
			encode := func(v any) ([]byte, error) { count++; return canonicalJSON(v) }
			m := newProjectedMergeV2("single_value", nil)
			v := []any{"value"}
			if err := m.add(v, firstDefault); err != nil {
				t.Fatal(err)
			}
			for range size {
				if err := m.compareSingle(v, true, encode); err != nil {
					t.Fatal(err)
				}
			}
			want := 2
			if firstDefault {
				want = 0
			}
			if count != want {
				t.Fatalf("%d default contributors needed %d encodings, want %d", size, count, want)
			}
			if err := m.compareSingle([]any{"value"}, false, encode); err != nil {
				t.Fatal(err)
			}
			if value, ok, failed := m.result("string_array"); failed || !ok || !reflect.DeepEqual(value, v) {
				t.Fatal("equal explicit operand caused conflict")
			}
			if err := m.compareSingle([]any{"conflict"}, false, encode); err != nil {
				t.Fatal(err)
			}
			if _, _, failed := m.result("string_array"); !failed {
				t.Fatal("late conflict lost")
			}
		}
	}
	input := twoEntityProjectionInputV2(false)
	cfg := input["projection_config"].(map[string]any)
	cfg["aggregation_rules"] = []any{aggregationRuleInputV2("root", "vertex", "source_entity", "kind-a", "summary")}
	const count = 4096
	rows := make([]any, count)
	for i := range rows {
		rows[i] = map[string]any{"source_entity_id": "entity-" + strconv.Itoa(i), "source_entity_kind": "kind-a", "properties": map[string]any{"group": "one"}, "metadata": map[string]any{}, "labels": []any{}}
	}
	input["source_entities"] = rows
	rows[0].(map[string]any)["properties"].(map[string]any)["probe"] = "value"
	inv := InvocationContextV2{GraphViewID: "view", SourceOwnerID: "owner"}
	calls := 0
	inv.CancellationCheck = func(context.Context) error { calls++; return nil }
	if _, err := ProjectV2(t.Context(), inv, inputJSONV2(t, input)); err != nil {
		t.Fatal(err)
	}
	baseline := calls
	for _, metadata := range []bool{false, true} {
		input["property_definitions"] = []any{}
		cfg["metadata_mappings"] = []any{}
		d := aggregationPropertyV2("one", "vertex", "summary", "properties.probe", "probe", "string", "single_value")
		d["missing_behavior"], d["default_value"] = "default", "value"
		appendDefinitionV2(input, d, metadata)
		body := inputJSONV2(t, input)
		calls = 0
		stop := 0
		inv.CancellationCheck = func(context.Context) error {
			calls++
			if calls == stop {
				return context.Canceled
			}
			return nil
		}
		r, err := ProjectV2(t.Context(), inv, body)
		if err != nil {
			t.Fatal(err)
		}
		if calls-baseline < count/graphProjectionLimits.CancellationCheckIntervalItems {
			t.Fatalf("contributor loop skipped checkpoints: %d versus %d", calls, baseline)
		}
		for _, v := range r.Vertices {
			if v.VertexKind == "summary" {
				values := v.Properties
				if metadata {
					values = v.Metadata.MappedMetadata
				}
				if values["probe"] != "value" {
					t.Fatal("missing aggregate default")
				}
			}
		}
		// The last callback closes the projection. The preceding periodic callbacks
		// traverse this single group's contributors after collection has completed.
		stop = baseline + 2
		calls = 0
		r, err = ProjectV2(t.Context(), inv, body)
		var failure *ProjectionErrorV2
		if !errors.As(err, &failure) || failure.Code != "projection_cancelled" || failure.ReasonCode != "context_cancelled" || failure.RetryAction != "do_not_retry" || !reflect.DeepEqual(r, ProjectionResultV2{}) {
			t.Fatalf("cancelled merge returned %+v, %v", r.ProjectionResultID, err)
		}
		// The final canonical contributor is explicit and conflicts with every default.
		probe := rows[999].(map[string]any)["properties"].(map[string]any)
		probe["probe"] = "different"
		expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "aggregation_merge_conflict", map[string]any{"issue_count": 1})
		delete(probe, "probe")
	}
}

// Exercise the production ceiling without retaining two half-gigabyte byte slices.
// The seed envelope fixes punctuation, identity lengths and ordered child framing.
// Each permitted array member adds exactly its unescaped ASCII payload length.
func checkFullEnvelopeCeilingV2(t *testing.T) {
	const vertices, items = 512, 65
	input := twoEntityProjectionInputV2(false)
	rows := make([]any, vertices)
	for i := range rows {
		rows[i] = map[string]any{"source_entity_id": strconv.Itoa(i), "source_entity_kind": "kind-a", "labels": []any{}, "properties": map[string]any{}, "metadata": map[string]any{}}
	}
	input["source_entities"] = rows
	def := aggregationPropertyV2("payload", "vertex", "vertex", "properties.payload", "payload", "string_array", "first")
	def["missing_behavior"], def["default_value"] = "default", make([]string, items)
	input["property_definitions"] = []any{def}
	r, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	small, err := r.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	const limit = 512 << 20
	if graphProjectionLimits.MaxResultBytes != limit {
		t.Fatal("incorrect maximum_result_bytes")
	}
	extra := limit - len(small.ResultJSON)
	short := strings.Repeat("x", extra/(vertices*items))
	long := short + "x"
	if len(long)+1 > graphProjectionLimits.MaxStringBytes {
		t.Fatal("fixture exceeds the string limit")
	}
	for i := range r.Vertices {
		values := make([]any, items)
		for j := range values {
			values[j] = short
			if i*items+j < extra%(vertices*items) {
				values[j] = long
			}
		}
		r.Vertices[i].Properties["payload"] = values
	}
	b := newResultBudgetV2(limit)
	if err := b.charge(r.Resource()); err != nil || b.remaining != 0 {
		t.Fatalf("full envelope: %d bytes remaining, %v", b.remaining, err)
	}
	last := r.Vertices[vertices-1].Properties["payload"].([]any)
	last[items-1] = short + "x"
	assertOutputLimitV2(t, newResultBudgetV2(limit).charge(r.Resource()), "maximum_result_bytes", limit)
}
