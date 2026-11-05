package graphprojection

import (
	"context"
	"errors"
	"maps"
	"reflect"
	"testing"
)

func reverseInputV2() map[string]any {
	input := fieldPathInputV2()
	mapping := input["projection_config"].(map[string]any)["relationship_mappings"].([]any)[0].(map[string]any)
	mapping["emit_reverse_edge"], mapping["direction_policy"], mapping["reverse_edge_kind"] = true, "normalize_forward", "back"
	return input
}

func checkConfiguredDirectionsV2(t *testing.T) {
	for policy, want := range map[string][]string{
		"preserve":          {"directed", "undirected", "bidirectional"},
		"normalize_forward": {"directed"}, "normalize_reverse": {"directed"},
		"undirected": {"undirected"}, "bidirectional": {"bidirectional"},
	} {
		t.Run(policy, func(t *testing.T) {
			input := fieldPathInputV2()
			input["source_relationships"] = []any{}
			config := input["projection_config"].(map[string]any)
			mapping := config["relationship_mappings"].([]any)[0].(map[string]any)
			mapping["direction_policy"] = policy
			for _, shared := range []bool{false, true} {
				if shared {
					second := maps.Clone(mapping)
					second["direction_policy"], second["source_relationship_kind"], second["mapping_rule_id"] = "undirected", "other", "other"
					config["declared_source_relationship_kinds"] = []string{"relation", "other"}
					config["relationship_mappings"] = []any{mapping, second}
					if policy == "normalize_forward" || policy == "normalize_reverse" {
						want = []string{"directed", "undirected"}
					}
					if policy == "bidirectional" {
						want = []string{"undirected", "bidirectional"}
					}
				}
				r, _, err := projectAggregationInputV2(t, input)
				if err != nil {
					t.Fatal(err)
				}
				requireReboundAdmissionV2(t, r, true)
				if got := r.SchemaRegistry.EdgeKinds[0].Directions; !reflect.DeepEqual(got, want) {
					t.Errorf("shared=%t: directions %v, want %v", shared, got, want)
				}
			}
		})
	}
}

func checkReverseFieldsV2(t *testing.T) {
	input := reverseInputV2()
	for _, item := range []struct{ kind, key string }{{"edge", "forward"}, {"back", "reverse"}, {"*", "shared"}} {
		for _, metadata := range []bool{false, true} {
			def := aggregationPropertyV2(item.key, "edge", item.kind, "properties.group", item.key, "string", "single_value")
			appendDefinitionV2(input, def, metadata)
			if metadata {
				def["metadata_mapping_id"] = item.key
			}
		}
	}
	result, _, err := projectAggregationInputV2(t, input)
	if err != nil || len(result.Edges) != 2 {
		t.Fatalf("project reverse fields: %v", err)
	}
	for _, edge := range result.Edges {
		key := map[string]string{"edge": "forward", "back": "reverse"}[edge.EdgeKind]
		want := map[string]any{key: "relation", "shared": "relation"}
		if !reflect.DeepEqual(edge.Properties, want) || !reflect.DeepEqual(edge.Metadata.MappedMetadata, want) {
			t.Errorf("%s field sets: %v, %v", edge.EdgeKind, edge.Properties, edge.Metadata.MappedMetadata)
		}
		var schemaProps, schemaMeta = map[string]bool{}, map[string]bool{}
		for _, s := range result.SchemaRegistry.PropertyKeys {
			if s.TargetKind == edge.EdgeKind {
				schemaProps[s.ProjectedKey] = true
			}
		}
		for _, s := range result.SchemaRegistry.MetadataKeys {
			if s.TargetKind == edge.EdgeKind {
				schemaMeta[s.ProjectedMetadataKey] = true
			}
		}
		keys := map[string]bool{key: true, "shared": true}
		if !reflect.DeepEqual(schemaProps, keys) || !reflect.DeepEqual(schemaMeta, keys) {
			t.Errorf("%s registry fields: %v, %v", edge.EdgeKind, schemaProps, schemaMeta)
		}
	}
	input["property_definitions"].([]any)[1].(map[string]any)["source_field_path"] = "properties.absent"
	failed, _, err := projectAggregationInputV2(t, input)
	if err == nil || !reflect.DeepEqual(failed, ProjectionResultV2{}) {
		t.Fatal("missing required reverse-only field succeeded")
	}
}

func checkReverseFamilyV2(t *testing.T) {
	for _, name := range []string{"valid", "reverse first", "fabricated sort key", "unknown family", "reverse vertex", "bad reference", "missing reference", "wrong endpoints", "wrong mapping", "wrong source", "wrong flag", "undirected", "direct reverse flag"} {
		t.Run(name, func(t *testing.T) {
			input := reverseInputV2()
			if name == "reverse first" {
				input["source_relationships"].([]any)[0].(map[string]any)["source_relationship_id"] = "relation-4"
			}
			r, _, err := projectAggregationInputV2(t, input)
			if err != nil || len(r.Edges) != 2 {
				t.Fatalf("reverse fixture: %v", err)
			}
			a, b := &r.Edges[0], &r.Edges[1]
			if a.EdgeFamily == "reverse" {
				a, b = b, a
			}
			if b.Metadata.ReverseOfEdgeID == nil || *b.Metadata.ReverseOfEdgeID != a.EdgeID || a.SrcVertexID != b.DstVertexID || a.DstVertexID != b.SrcVertexID {
				t.Fatal("reverse link or endpoints")
			}
			switch name {
			case "reverse first":
				if a.EdgeID != "ed_cdbe7f66c26ba4b8afed354c1c13d944e6e233885a9927f30d6b981d8c5983b1" ||
					b.EdgeID != "ed_1c0b2410f3b917115949ae8e1b5594b45ec707997f6ac754aa0d449a8ca98176" ||
					r.Edges[0].EdgeFamily != "reverse" || b.SortKey != "sk_ce2796808e9c787ae496aa93b833fdd40dfcd2b1274f6ab1abfd5765125ea97e" || a.SortKey != "sk_dc16e85551a5ba52f5a920054f90ae8749cccfff96a4277d30ae8d024c955210" {
					t.Fatal("reverse-first identity vector")
				}
			case "fabricated sort key":
				b.SortKey = a.SortKey
				sortEdges(r.Edges)
			case "unknown family":
				b.EdgeFamily = "other"
			case "reverse vertex":
				r.Vertices[0].VertexFamily = "reverse"
			case "bad reference":
				b.Metadata.ReverseOfEdgeID = &b.EdgeID
			case "missing reference":
				b.Metadata.ReverseOfEdgeID = nil
			case "wrong endpoints":
				b.SrcVertexID = a.SrcVertexID
			case "wrong mapping":
				wrong := "other"
				b.Metadata.MappingRuleID = &wrong
			case "wrong source":
				ref := *b.SourceRelationshipRef
				ref.SourceRelationshipID = "other"
				b.SourceRelationshipRef = &ref
			case "wrong flag":
				b.Metadata.IsReverseEdge = false
			case "undirected":
				b.Direction = "undirected"
			case "direct reverse flag":
				a.Metadata.IsReverseEdge = true
			}
			r.CanonicalOutputSHA256, _ = canonicalOutputDigestV2(r.outputResource())
			r.ProjectionResultID, _ = deriveProjectionResultIDV2(r.ResultBindingV2())
			completed, err := r.CompletedResult()
			if err != nil {
				t.Fatal(err)
			}
			err = completed.Validate(context.Background())
			if name == "valid" || name == "reverse first" {
				if err != nil {
					t.Fatal(err)
				}
			} else if !errors.Is(err, ErrResultV2Invalid) {
				t.Fatalf("bad family/link accepted: %v", err)
			}
		})
	}
}

func checkEndpointAgreementV2(t *testing.T) {
	for _, policy := range []string{"exclude", "error"} {
		for _, side := range []string{"src", "dst"} {
			for _, mode := range []string{"equal", "null", "conflict", "missing", "unmatched", "array", "missing then array", "missing then conflict"} {
				for _, permuted := range []bool{false, true} {
					t.Run(policy+"/"+side+"/"+mode+map[bool]string{true: "/permuted"}[permuted], func(t *testing.T) {
						input := numericAggregationInputV2("a", "b")
						entities := input["source_entities"].([]any)
						props := func(i int) map[string]any { return entities[i].(map[string]any)["properties"].(map[string]any) }
						for i := range entities {
							props(i)["one"] = "group"
							props(i)["endpoint"] = "a"
						}
						edge := input["projection_config"].(map[string]any)["aggregation_rules"].([]any)[0].(map[string]any)
						edge["grouping_keys"] = []any{"properties.one"}
						ep := edge["endpoint_grouping"].(map[string]any)
						ep["src_grouping_keys"], ep["dst_grouping_keys"] = []any{"properties.endpoint"}, []any{"properties.endpoint"}
						ep[side+"_grouping_keys"] = []any{"properties.chosen"}
						ep["missing_endpoint_behavior"] = policy
						props(0)["chosen"], props(1)["chosen"] = "a", "a"
						switch mode {
						case "null":
							props(0)["group"], props(1)["group"], props(0)["chosen"], props(1)["chosen"], props(0)["endpoint"], props(1)["endpoint"] = nil, nil, nil, nil, nil, nil
						case "conflict":
							props(1)["chosen"] = "b"
						case "missing":
							delete(props(1), "chosen")
						case "array":
							props(1)["chosen"] = []any{"a"}
						case "unmatched":
							props(0)["chosen"], props(1)["chosen"] = "absent", "absent"
						case "missing then array":
							delete(props(0), "chosen")
							props(1)["chosen"] = []any{"a"}
						case "missing then conflict":
							delete(props(0), "chosen")
							entities = append(entities, map[string]any{"source_entity_id": "entity-c", "source_entity_kind": "kind-a", "properties": map[string]any{"group": "b", "one": "group", "endpoint": "a", "chosen": "b"}, "metadata": map[string]any{}, "labels": []any{}})
							input["source_entities"] = entities
						}
						if permuted {
							entities[0], entities[len(entities)-1] = entities[len(entities)-1], entities[0]
						}
						r, _, err := projectAggregationInputV2(t, input)
						if mode == "equal" || mode == "null" {
							if err != nil || len(r.Edges) != 1 {
								t.Fatalf("unanimous endpoints: %v", err)
							}
							refs := r.Edges[0].Metadata.AggregationSourceRefs
							if len(refs) != 2 || refs[0].RefID != "entity-a" || refs[1].RefID != "entity-b" {
								t.Fatal("incorrect provenance")
							}
							return
						}
						if policy == "exclude" && (mode == "missing" || mode == "unmatched") {
							if err != nil || len(r.Edges) != 0 {
								t.Fatalf("exclusion: %v", err)
							}
							return
						}
						reason := "aggregation_endpoint_missing"
						switch mode {
						case "array", "missing then array":
							reason = "invalid_aggregation_rule"
						case "conflict", "missing then conflict":
							reason = "aggregation_endpoint_conflict"
						}
						var pe *ProjectionErrorV2
						if !errors.As(err, &pe) || pe.Code != "projection_validation_failed" || pe.ReasonCode != reason || pe.RetryAction != "do_not_retry" || !reflect.DeepEqual(r, ProjectionResultV2{}) {
							t.Fatalf("want %s, got %v", reason, err)
						}
					})
				}
			}
		}
	}
}

func checkConfiguredLabelsV2(t *testing.T) {
	for _, scenario := range []string{"shared forward", "shared reverse", "coincident kinds"} {
		t.Run(scenario, func(t *testing.T) {
			input := reverseInputV2()
			config := input["projection_config"].(map[string]any)
			config["default_edge_labels"] = []any{"default"}
			input["source_relationships"] = []any{}
			a := config["relationship_mappings"].([]any)[0].(map[string]any)
			b := maps.Clone(a)
			a["mapping_rule_id"], a["mapping_labels"] = "a", []any{"from-a"}
			b["mapping_rule_id"], b["source_relationship_kind"], b["mapping_labels"] = "b", "other", []any{"from-b"}
			config["declared_source_relationship_kinds"] = []any{"relation", "other"}
			kind, labels, preserved := "back", []string{"default", "from-b"}, false
			switch scenario {
			case "shared forward":
				a["emit_reverse_edge"] = false
				delete(a, "reverse_edge_kind")
			case "shared reverse":
				a["label_policy"], b["projected_edge_kind"] = "preserve_source", "second"
				preserved = true
			case "coincident kinds":
				a["reverse_edge_kind"], b["reverse_edge_kind"] = "edge", "edge"
				kind, labels = "edge", []string{"default", "from-a", "from-b"}
			}
			for _, mappings := range [][]any{{a, b}, {b, a}} {
				config["relationship_mappings"] = mappings
				r, _, err := projectAggregationInputV2(t, input)
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for _, entry := range r.SchemaRegistry.EdgeKinds {
					if entry.EdgeKind != kind {
						continue
					}
					found = true
					if !reflect.DeepEqual(entry.Labels, labels) || entry.SourceLabelsPreserved != preserved {
						t.Errorf("%s labels %v, source preserved %v", kind, entry.Labels, entry.SourceLabelsPreserved)
					}
				}
				if !found {
					t.Fatalf("missing %s", kind)
				}
			}
		})
	}
}

// Start with producer-derived rows so a contradiction can be isolated to one
// shared mapping claim even when each row has a valid transcript on its own.
func checkEdgeIdentitiesV2(t *testing.T) {
	project := func(input map[string]any) ProjectionResultV2 {
		t.Helper()
		r, _, err := projectAggregationInputV2(t, input)
		if err != nil {
			t.Fatal(err)
		}
		requireReboundAdmissionV2(t, r, true)
		return r
	}
	input := reverseInputV2()
	m := input["projection_config"].(map[string]any)["relationship_mappings"].([]any)[0].(map[string]any)
	if digest := relationshipMappingDigestV2("relation-map", "relation", "edge", "normalize_forward", true, "back"); digest != "e9e114662b477a1ac64d9107ea53f6409d1a7d81e791e01c1bd7563f8639277a" {
		t.Fatal(digest)
	}
	m["mapping_rule_id"] = "映射"
	input["source_relationships"].([]any)[0].(map[string]any)["source_relationship_id"] = "雪"
	r := project(input)
	for _, e := range r.Edges {
		if e.EdgeFamily == "reverse" && (e.EdgeID != "ed_16c5abe16b3904238f7d765346bbaca3107485f9af82ab4af1f756987bdb4560" || e.SortKey != "sk_e1db3e3469842c4fecf158dd4795fdb6c8364578d80124c2febfb078f4c87157") {
			t.Fatal("multibyte reverse transcript")
		}
	}
	m["emit_reverse_edge"] = false
	delete(m, "reverse_edge_kind")
	for _, policy := range []string{"normalize_reverse", "normalize_forward", "undirected", "bidirectional", "preserve"} {
		m["direction_policy"] = policy
		r = project(input)
	}
	// A preserve mapping also declares undirected and bidirectional capability.
	r.SchemaRegistry.EdgeKinds[0].Directions = []string{"directed"}
	requireReboundAdmissionV2(t, r, false)
	m["direction_policy"] = "normalize_forward"
	first := project(input)
	m["direction_policy"] = "normalize_reverse"
	input["source_relationships"].([]any)[0].(map[string]any)["source_relationship_id"] = "second"
	second := project(input)
	first.Edges = append(first.Edges, second.Edges...)
	sortEdges(first.Edges)
	requireReboundAdmissionV2(t, first, false)
}
