package graphprojection

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func reverseInputV2() map[string]any {
	input := fieldPathInputV2()
	mapping := input["projection_config"].(map[string]any)["relationship_mappings"].([]any)[0].(map[string]any)
	mapping["emit_reverse_edge"], mapping["direction_policy"], mapping["reverse_edge_kind"] = true, "normalize_forward", "back"
	return input
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
	project := func() ProjectionResultV2 {
		r, _, err := projectAggregationInputV2(t, reverseInputV2())
		if err != nil || len(r.Edges) != 2 {
			t.Fatalf("reverse fixture: %v", err)
		}
		return r
	}
	for _, name := range []string{"valid", "reverse first", "unknown family", "reverse vertex", "bad reference", "missing reference", "wrong endpoints", "wrong mapping", "wrong source", "wrong flag", "undirected", "direct reverse flag"} {
		t.Run(name, func(t *testing.T) {
			r := project()
			a, b := &r.Edges[0], &r.Edges[1]
			if a.EdgeFamily == "reverse" {
				a, b = b, a
			}
			if b.Metadata.ReverseOfEdgeID == nil || *b.Metadata.ReverseOfEdgeID != a.EdgeID || a.SrcVertexID != b.DstVertexID || a.DstVertexID != b.SrcVertexID {
				t.Fatal("reverse link or endpoints")
			}
			switch name {
			case "reverse first":
				b.SortKey = "a"
				a.SortKey = "b"
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
