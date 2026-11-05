package graphprojection

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
)

func requireReboundAdmissionV2(t *testing.T, r ProjectionResultV2, valid bool) {
	t.Helper()
	err := reboundModelV2(t, r).Validate(t.Context())
	if (err == nil) != valid || err != nil && !errors.Is(err, ErrResultV2Invalid) {
		t.Fatalf("expected admissible=%t: %v", valid, err)
	}
}

func checkResultPresenceV2(t *testing.T) {
	input := reverseInputV2()
	for _, scope := range []string{"graph_view", "vertex", "edge"} {
		path := "metadata.field"
		if scope == "graph_view" {
			path = "source_metadata.field"
		}
		for _, metadata := range []bool{false, true} {
			if scope == "graph_view" && !metadata {
				continue
			}
			if !metadata {
				path = "properties.field"
			} else if scope != "graph_view" {
				path = "metadata.field"
			}
			for _, key := range []string{"always", "defaulted", "nullable", "optional"} {
				def := aggregationPropertyV2(scope+"-"+key, scope, "*", path, key, "string", "single_value")
				switch key {
				case "always":
					def["required"] = false
					def["null_output_policy"] = "emit_null"
				case "defaulted":
					def["source_field_path"] = strings.Replace(path, ".field", ".absent", 1)
					def["missing_behavior"], def["default_value"] = "default", "filled"
				case "nullable":
					def["source_field_path"] = strings.Replace(path, ".field", ".absent", 1)
					def["missing_behavior"], def["default_value"], def["null_output_policy"] = "default", nil, "emit_null"
				case "optional":
					def["source_null_behavior"] = "omit"
				}
				appendDefinitionV2(input, def, metadata)
				if metadata {
					def["metadata_mapping_id"] = scope + "-" + key
				}
			}
		}
	}
	input["source_metadata"].(map[string]any)["field"] = "value"
	for _, collection := range []string{"source_entities", "source_relationships"} {
		for _, raw := range input[collection].([]any) {
			raw.(map[string]any)["metadata"].(map[string]any)["field"] = "value"
			raw.(map[string]any)["properties"].(map[string]any)["field"] = "value"
		}
	}
	r, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	testMap := func(t *testing.T, values map[string]any) {
		t.Helper()
		original := values["always"]
		values["always"] = nil
		requireReboundAdmissionV2(t, r, false)
		values["always"] = original
		for _, key := range []string{"always", "defaulted", "nullable", "optional"} {
			val, present := values[key]
			if !present {
				t.Fatalf("producer omitted %s", key)
			}
			delete(values, key)
			requireReboundAdmissionV2(t, r, key == "optional")
			values[key] = val
		}
	}
	requireReboundAdmissionV2(t, r, true)
	for _, values := range []map[string]any{r.MappedMetadata, r.Vertices[0].Properties, r.Vertices[0].Metadata.MappedMetadata} {
		testMap(t, values)
	}
	for _, e := range r.Edges {
		testMap(t, e.Properties)
		testMap(t, e.Metadata.MappedMetadata)
	}

	input = mixedAggregationInputV2(false)
	for _, raw := range input["property_definitions"].([]any) {
		raw.(map[string]any)["null_output_policy"] = "emit_null"
		d := maps.Clone(raw.(map[string]any))
		id := d["property_definition_id"]
		appendDefinitionV2(input, d, true)
		d["metadata_mapping_id"] = id
	}
	r, _, err = projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range r.Vertices {
		if v.VertexFamily != "aggregated" {
			continue
		}
		for _, values := range []map[string]any{v.Properties, v.Metadata.MappedMetadata} {
			score := values["score"]
			values["score"] = nil
			requireReboundAdmissionV2(t, r, false)
			delete(values, "score")
			requireReboundAdmissionV2(t, r, false)
			values["score"] = score
		}
	}
	for _, e := range r.Edges {
		for _, values := range []map[string]any{e.Properties, e.Metadata.MappedMetadata} {
			score := values["score"]
			values["score"] = nil
			requireReboundAdmissionV2(t, r, false)
			delete(values, "score")
			requireReboundAdmissionV2(t, r, false)
			values["score"] = score
		}
	}
}

func checkObjectLabelsV2(t *testing.T) {
	r, _, err := projectAggregationInputV2(t, reverseInputV2())
	if err != nil {
		t.Fatal(err)
	}
	for i := range r.SchemaRegistry.EdgeKinds {
		r.SchemaRegistry.EdgeKinds[i].Labels = []string{"a", "b"}
	}
	for _, tc := range []struct {
		labels []string
		valid  bool
	}{
		{nil, true}, {[]string{"b"}, true}, {[]string{"a", "b"}, true}, {[]string{""}, false},
		{[]string{"b", "a"}, false}, {[]string{"a", "a"}, false}, {[]string{"foreign"}, false},
	} {
		for i := range r.Edges {
			r.Edges[i].Labels = tc.labels
		}
		requireReboundAdmissionV2(t, r, tc.valid)
	}
	for i := range r.Edges {
		r.Edges[i].Labels = nil
	}
	r.SchemaRegistry.VertexKinds[0].SourceLabelsPreserved = true
	v := &r.Vertices[0]
	v.Labels = []string{strings.Repeat("é", 128)}
	requireReboundAdmissionV2(t, r, true)
	v.Labels[0] += "x"
	requireReboundAdmissionV2(t, r, false)
	v.Labels = nil
	for i := 0; i <= graphProjectionLimits.MaxLabelsPerObject; i++ {
		v.Labels = append(v.Labels, fmt.Sprintf("%03d", i))
	}
	requireReboundAdmissionV2(t, r, false)
	v.Labels = v.Labels[:len(v.Labels)-1]
	requireReboundAdmissionV2(t, r, true)

	r, _, err = projectAggregationInputV2(t, mixedAggregationInputV2(false))
	if err != nil {
		t.Fatal(err)
	}
	for i := range r.SchemaRegistry.VertexKinds {
		r.SchemaRegistry.VertexKinds[i].SourceLabelsPreserved = true
	}
	for i := range r.SchemaRegistry.EdgeKinds {
		r.SchemaRegistry.EdgeKinds[i].SourceLabelsPreserved = true
	}
	for i := range r.Vertices {
		if r.Vertices[i].VertexFamily == "aggregated" {
			r.Vertices[i].Labels = []string{"foreign"}
			requireReboundAdmissionV2(t, r, false)
			r.Vertices[i].Labels = nil
		}
	}
	r.Edges[0].Labels = []string{"foreign"}
	requireReboundAdmissionV2(t, r, false)
}

func checkRelationshipClaimsV2(t *testing.T) {
	for _, scenario := range []string{"valid shared kind and self-loop", "duplicate source across kinds", "two mappings for one kind", "one mapping for two kinds", "two forward kinds", "two reverse kinds", "duplicate reverse", "missing sibling", "all reverse rows absent"} {
		t.Run(scenario, func(t *testing.T) {
			input := reverseInputV2()
			config := input["projection_config"].(map[string]any)
			mapping := config["relationship_mappings"].([]any)[0].(map[string]any)
			if scenario == "valid shared kind and self-loop" || scenario == "missing sibling" {
				mapping["reverse_edge_kind"] = "edge"
			}
			rel := maps.Clone(input["source_relationships"].([]any)[0].(map[string]any))
			rel["source_relationship_id"], rel["dst_source_entity_id"] = "second", rel["src_source_entity_id"]
			input["source_relationships"] = append(input["source_relationships"].([]any), rel)
			if scenario == "duplicate source across kinds" || scenario == "one mapping for two kinds" {
				second := maps.Clone(mapping)
				second["source_relationship_kind"], second["mapping_rule_id"] = "other", "second-map"
				config["declared_source_relationship_kinds"] = []any{"other", "relation"}
				config["relationship_mappings"] = []any{mapping, second}
				rel["source_relationship_kind"] = "other"
			}
			r, _, err := projectAggregationInputV2(t, input)
			if err != nil {
				t.Fatal(err)
			}
			requireReboundAdmissionV2(t, r, true)
			for i := range r.Edges {
				e := &r.Edges[i]
				if e.SourceRelationshipRef.SourceRelationshipID != "second" {
					continue
				}
				switch scenario {
				case "duplicate source across kinds":
					e.SourceRelationshipRef.SourceRelationshipID = "relation-1"
				case "two mappings for one kind", "one mapping for two kinds":
					name := "other-map"
					if scenario == "one mapping for two kinds" {
						name = "relation-map"
					}
					e.SourceRelationshipRef.MappingRuleID = name
					e.Metadata.MappingRuleID = &name
				case "two forward kinds":
					if e.EdgeFamily == "direct" {
						e.EdgeKind = "back"
					}
				case "two reverse kinds":
					if e.EdgeFamily == "reverse" {
						e.EdgeKind = "edge"
					}
				case "duplicate reverse":
					if e.EdgeFamily == "reverse" {
						copy := *e
						copy.EdgeID = "ed_" + strings.Repeat("f", 64)
						r.Edges = append(r.Edges, copy)
					}
				}
			}
			if scenario == "missing sibling" || scenario == "all reverse rows absent" {
				r.Edges = slices.DeleteFunc(r.Edges, func(e Edge) bool {
					return e.EdgeFamily == "reverse" && (scenario == "all reverse rows absent" || e.SourceRelationshipRef.SourceRelationshipID == "second")
				})
			}
			sortEdges(r.Edges)
			requireReboundAdmissionV2(t, r, scenario == "valid shared kind and self-loop")
		})
	}
}

// Change raw contributor tuples without changing projected-object identities.
func checkRawSourceClaimsV2(t *testing.T) {
	input := mixedAggregationInputV2(false)
	control := fieldPathInputV2()
	cfg := input["projection_config"].(map[string]any)
	for _, key := range []string{"declared_source_relationship_kinds", "relationship_mappings"} {
		cfg[key] = control["projection_config"].(map[string]any)[key]
	}
	input["source_relationships"] = control["source_relationships"]
	r, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	var root, final *Vertex
	for i := range r.Vertices {
		switch r.Vertices[i].VertexKind {
		case "endpoints":
			root = &r.Vertices[i]
		case "final":
			final = &r.Vertices[i]
		}
	}
	ref := func(scope, kind, id string) []SourceRef {
		return []SourceRef{{RefKind: scope, RefKindName: kind, RefID: id, ContributorSortKey: contributorOrder{kind, id, "", ""}.key()}}
	}
	for _, tc := range []struct{ scope, kind, id string }{{"source_entity", "kind-a", "entity-a"}, {"source_relationship", "relation", "relation-1"}} {
		root.Metadata.AggregationSourceRefs = ref(tc.scope, tc.kind, tc.id)
		final.Metadata.AggregationSourceRefs = ref(tc.scope, tc.kind, tc.id)
		requireReboundAdmissionV2(t, r, true)
		final.Metadata.AggregationSourceRefs = ref(tc.scope, "wrong", tc.id)
		requireReboundAdmissionV2(t, r, false)
		root.Metadata.AggregationSourceRefs = ref(tc.scope, "one", "raw")
		final.Metadata.AggregationSourceRefs = ref(tc.scope, "two", "raw")
		requireReboundAdmissionV2(t, r, false)
	}
	root.Metadata.AggregationSourceRefs = ref("source_entity", "relation", "raw")
	final.Metadata.AggregationSourceRefs = ref("source_relationship", "relation", "relation-1")
	requireReboundAdmissionV2(t, r, false)
	root.Metadata.AggregationSourceRefs = ref("source_entity", "one", "raw")
	final.Metadata.AggregationSourceRefs = ref("source_relationship", "two", "raw")
	requireReboundAdmissionV2(t, r, false)
	root.Metadata.AggregationSourceRefs = ref("source_entity", "kind-a", "raw")
	final.Metadata.AggregationSourceRefs = ref("source_relationship", "relation", "raw")
	requireReboundAdmissionV2(t, r, true)
}

func checkAggregateCapabilitiesV2(t *testing.T) {
	r, _, err := projectAggregationInputV2(t, mixedAggregationInputV2(false))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"", "a", "sk_" + strings.Repeat("a", 63), "sk_" + strings.Repeat("A", 64), "sk_" + strings.Repeat("0", 64) + "\n"} {
		altered := r
		altered.Edges = slices.Clone(r.Edges)
		altered.Edges[0].SortKey = key
		sortEdges(altered.Edges)
		requireReboundAdmissionV2(t, altered, false)
		altered = r
		altered.Vertices = slices.Clone(r.Vertices)
		for i := range altered.Vertices {
			if altered.Vertices[i].VertexFamily == "aggregated" {
				altered.Vertices[i].SortKey = key
				break
			}
		}
		sortVertices(altered.Vertices)
		requireReboundAdmissionV2(t, altered, false)
	}
	for _, empty := range []bool{false, true} {
		if empty {
			r.Vertices = nil
			r.Edges = nil
		}
		for i := range r.SchemaRegistry.EdgeKinds {
			kind := &r.SchemaRegistry.EdgeKinds[i]
			kind.SourceLabelsPreserved = true
			requireReboundAdmissionV2(t, r, false)
			kind.SourceRelationshipKinds = []string{"dormant-" + kind.EdgeKind}
			requireReboundAdmissionV2(t, r, true)
			kind.SourceLabelsPreserved = false
			kind.SourceRelationshipKinds = nil
			kind.Directions = []string{"directed", "undirected", "bidirectional"}
			requireReboundAdmissionV2(t, r, false)
			kind.Directions = []string{"directed"}
		}
		for i := range r.SchemaRegistry.VertexKinds {
			kind := &r.SchemaRegistry.VertexKinds[i]
			if len(kind.SourceEntityKinds) > 0 {
				continue
			}
			kind.SourceLabelsPreserved = true
			requireReboundAdmissionV2(t, r, false)
			kind.SourceLabelsPreserved = false
		}
	}
}
