package graphprojection

import (
	"encoding/hex"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func checkContributorTupleOrderV2(t *testing.T) {
	t.Run("aggregate immediate provenance", checkAggregateImmediateOrderV2)
	t.Run("tuple encoding", checkContributorOrderEncodingV2)
	for _, mode := range []string{"source_entity", "source_relationship", "projected_vertex", "projected_edge", "reverse"} {
		t.Run(mode, func(t *testing.T) {
			input := fieldPathInputV2()
			config := input["projection_config"].(map[string]any)
			entity := mode == "source_entity" || mode == "projected_vertex"
			collection, kindMember, idMember, inputKind := "source_relationships", "source_relationship_kind", "source_relationship_id", "relation"
			if entity {
				collection, kindMember, idMember, inputKind = "source_entities", "source_entity_kind", "source_entity_id", "kind-a"
				input["source_relationships"] = []any{}
			}
			seed := input[collection].([]any)[0].(map[string]any)
			records := []any{}
			for _, id := range []string{"é", "b", "aa", "a"} {
				record := deepCopyJSONMap(t, seed)
				record[idMember], record[kindMember] = id, inputKind
				record["properties"] = map[string]any{"group": "all", "value": id, "tags": []any{id}}
				records = append(records, record)
			}
			input[collection] = records
			scope, prefix := mode, "properties."
			if strings.HasPrefix(mode, "projected_") || mode == "reverse" {
				scope, prefix = "projected_edge", "projected.properties."
				directScope, directKind := "edge", "edge"
				if entity {
					scope, directScope, directKind = "projected_vertex", "vertex", "vertex"
				}
				inputKind = directKind
				for _, key := range []string{"group", "value", "tags"} {
					valueType := "string"
					if key == "tags" {
						valueType = "string_array"
					}
					appendDefinitionV2(input, aggregationPropertyV2("direct-"+key, directScope, directKind, "properties."+key, key, valueType, "first"), false)
				}
				if mode == "reverse" {
					mapping := config["relationship_mappings"].([]any)[0].(map[string]any)
					mapping["emit_reverse_edge"], mapping["reverse_edge_kind"], mapping["direction_policy"] = true, "edge", "normalize_forward"
				}
			}
			config["aggregation_rules"] = []any{aggregationRuleInputV2("root", "vertex", scope, inputKind, "summary")}
			for _, merge := range []string{"first", "last", "ordered_list"} {
				key, valueType := "value", "string"
				if merge == "ordered_list" {
					key, valueType = "tags", "string_array"
				}
				appendDefinitionV2(input, aggregationPropertyV2(merge, "vertex", "summary", prefix+key, merge, valueType, merge), false)
			}
			appendDefinitionV2(input, aggregationPropertyV2("meta-first", "vertex", "summary", prefix+"value", "first", "string", "first"), true)
			result, _, err := projectAggregationInputV2(t, input)
			if err != nil {
				t.Fatal(err)
			}
			var aggregate Vertex
			for _, vertex := range result.Vertices {
				if vertex.VertexKind == "summary" {
					aggregate = vertex
				}
			}
			expectedList := []any{}
			for _, id := range []string{"a", "aa", "b", "é"} {
				expectedList = append(expectedList, id)
				if mode == "reverse" {
					expectedList = append(expectedList, id)
				}
			}
			if aggregate.Properties["first"] != "a" || aggregate.Properties["last"] != "é" || !reflect.DeepEqual(aggregate.Properties["ordered_list"], expectedList) || aggregate.Metadata.MappedMetadata["first"] != "a" {
				t.Fatalf("merge order differs from source tuple: properties=%v metadata=%v", aggregate.Properties, aggregate.Metadata.MappedMetadata)
			}
			type expectedRef struct{ source, id, kind, rule, prior string }
			expected := []expectedRef{}
			if prefix == "properties." {
				for _, id := range []string{"a", "aa", "b", "é"} {
					expected = append(expected, expectedRef{source: id, id: id, kind: inputKind})
				}
			} else if entity {
				for _, vertex := range result.Vertices {
					if ref := vertex.SourceEntityRef; ref != nil {
						expected = append(expected, expectedRef{source: ref.SourceEntityID, id: vertex.VertexID, kind: ref.SourceEntityKind, rule: ref.MappingRuleID, prior: vertex.VertexID})
					}
				}
			} else {
				for _, edge := range result.Edges {
					ref := edge.SourceRelationshipRef
					expected = append(expected, expectedRef{source: ref.SourceRelationshipID, id: edge.EdgeID, kind: ref.SourceRelationshipKind, rule: ref.MappingRuleID, prior: edge.EdgeID})
				}
			}
			sort.Slice(expected, func(i, j int) bool {
				if expected[i].source != expected[j].source {
					return expected[i].source < expected[j].source
				}
				return expected[i].id < expected[j].id
			})
			refs := aggregate.Metadata.AggregationSourceRefs
			if len(refs) != len(expected) {
				t.Fatal("contributor references lost")
			}
			for i, want := range expected {
				parts := []string{want.kind, want.source, want.rule, want.prior}
				for j := range parts {
					parts[j] = hex.EncodeToString([]byte(parts[j]))
				}
				key := "csk_" + strings.Join(parts, ".")
				if refs[i].RefID != want.id || refs[i].ContributorSortKey != key {
					t.Fatalf("reference order/key differs at %d: got %v want %s/%s", i, refs[i], want.id, key)
				}
				if i > 0 && refs[i-1].ContributorSortKey >= refs[i].ContributorSortKey {
					t.Fatal("encoded tuple is not order preserving")
				}
			}
			for left, right := 0, len(records)-1; left < right; left, right = left+1, right-1 {
				records[left], records[right] = records[right], records[left]
			}
			second, _, err := projectAggregationInputV2(t, input)
			if err != nil || !reflect.DeepEqual(result, second) {
				t.Fatalf("contributor permutation changed canonical result: %v", err)
			}
		})
	}
}

func checkAggregateImmediateOrderV2(t *testing.T) {
	input := fieldPathInputV2()
	for _, raw := range input["source_entities"].([]any) {
		entity := raw.(map[string]any)
		entity["properties"] = map[string]any{"group": "all", "value": entity["source_entity_id"]}
	}
	config := input["projection_config"].(map[string]any)
	config["aggregation_rules"] = []any{aggregationRuleInputV2("a", "vertex", "source_entity", "kind-a", "shared"), aggregationRuleInputV2("b", "vertex", "source_entity", "kind-b", "shared"), aggregationRuleInputV2("final", "vertex", "projected_vertex", "shared", "final")}
	for _, key := range []string{"group", "value"} {
		appendDefinitionV2(input, aggregationPropertyV2("shared-"+key, "vertex", "shared", "properties."+key, key, "string", "single_value"), false)
	}
	for _, merge := range []string{"first", "last"} {
		appendDefinitionV2(input, aggregationPropertyV2(merge, "vertex", "final", "projected.properties.value", merge, "string", merge), false)
	}
	result, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	var roots []Vertex
	var leaf Vertex
	for _, vertex := range result.Vertices {
		if vertex.VertexKind == "shared" {
			roots = append(roots, vertex)
		}
		if vertex.VertexKind == "final" {
			leaf = vertex
		}
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i].VertexID < roots[j].VertexID })
	if len(roots) != 2 || len(leaf.Metadata.AggregationSourceRefs) != 2 {
		t.Fatal("incomplete aggregate chain")
	}
	if leaf.Properties["first"] != roots[0].Properties["value"] || leaf.Properties["last"] != roots[1].Properties["value"] {
		t.Fatal("aggregate ancestry replaced immediate identity order")
	}
	for i, root := range roots {
		parts := []string{root.VertexKind, root.VertexID, *root.Metadata.AggregationRuleID, root.VertexID}
		for j := range parts {
			parts[j] = hex.EncodeToString([]byte(parts[j]))
		}
		if ref := leaf.Metadata.AggregationSourceRefs[i]; ref.RefID != root.VertexID || ref.ContributorSortKey != "csk_"+strings.Join(parts, ".") {
			t.Fatal("aggregate provenance/key drift")
		}
	}
}

func checkContributorOrderEncodingV2(t *testing.T) {
	ordered := []contributorOrder{{"", "a", "", ""}, {"a", "", "", ""}, {"a", "a", "", ""}, {"a", "a", "a", ""}, {"a", "a", "a", "a"}, {"a", "aa", "", ""}, {"aa", "", "", ""}, {"é", "", "", ""}}
	for i, left := range ordered {
		for j, right := range ordered {
			if left.less(right) != (i < j) || (left.key() < right.key()) != (i < j) {
				t.Fatal("tuple/key ordering disagrees with specified prefix order")
			}
		}
	}
	if got := (contributorOrder{"a", "é", "", ""}).key(); got != "csk_61.c3a9.." {
		t.Fatalf("wire tuple encoding drift: %s", got)
	}
	rule := "aggregate-edge"
	order, err := (contributor{Edge: &Edge{EdgeKind: "shared", EdgeID: "ed-a", Metadata: EdgeMetadata{AggregationRuleID: &rule}}}).orderTuple()
	if err != nil || order != (contributorOrder{"shared", "ed-a", "aggregate-edge", "ed-a"}) {
		t.Fatalf("aggregate edge order drift: %v %v", order, err)
	}
}
