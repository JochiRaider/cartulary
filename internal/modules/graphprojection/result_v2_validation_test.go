package graphprojection

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func checkCompletedResultProofsV2(t *testing.T) {
	input := twoEntityProjectionInputV2(false)
	config := input["projection_config"].(map[string]any)
	config["declared_source_relationship_kinds"] = []string{"relationship"}
	config["relationship_mappings"] = []any{map[string]any{
		"mapping_rule_id": "link", "source_relationship_kind": "relationship", "projected_edge_kind": "edge",
		"inclusion_predicate": "always", "direction_policy": "preserve", "emit_reverse_edge": false, "label_policy": "mapping_only",
		"mapping_labels": []string{}, "required_property_keys": []string{}, "optional_property_keys": []string{},
	}}
	input["source_relationships"] = []any{map[string]any{
		"source_relationship_id": "relationship", "source_relationship_kind": "relationship", "src_source_entity_id": "entity-a", "dst_source_entity_id": "entity-b",
		"direction": "forward", "properties": map[string]any{}, "metadata": map[string]any{}, "labels": []string{},
	}}
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	producer := func() ProjectionResultV2 {
		t.Helper()
		projected, err := ProjectV2(context.Background(), InvocationContextV2{GraphViewID: "proof-view", SourceOwnerID: "proof-owner"}, raw)
		if err != nil {
			t.Fatal(err)
		}
		return projected
	}
	complete := func(projected ProjectionResultV2) CompletedResultV2 {
		t.Helper()
		result, err := projected.CompletedResult()
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	valid := complete(producer())
	if err := valid.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := valid.Validate(cancelled); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	for name, mutate := range map[string]func(*CompletedResultV2){
		"trailing LF": func(v *CompletedResultV2) { v.ResultJSON = append(v.ResultJSON, '\n') },
		"two values":  func(v *CompletedResultV2) { v.ResultJSON = append(v.ResultJSON, []byte("{}")...) },
		"duplicate member": func(v *CompletedResultV2) {
			v.ResultJSON = append([]byte(`{"mapped_metadata":{},`), v.ResultJSON[1:]...)
		},
		"unknown field":   func(v *CompletedResultV2) { v.ResultJSON = append([]byte(`{"unknown":true,`), v.ResultJSON[1:]...) },
		"invalid UTF8":    func(v *CompletedResultV2) { v.ResultJSON = append(v.ResultJSON, 0xff) },
		"null root":       func(v *CompletedResultV2) { v.ResultJSON = []byte("null") },
		"wrong tuple":     func(v *CompletedResultV2) { v.Binding.SourceOwnerID = "different" },
		"wrong digest":    func(v *CompletedResultV2) { v.Binding.CanonicalOutputSHA256 = strings.Repeat("a", 64) },
		"missing row":     func(v *CompletedResultV2) { v.Vertices = v.Vertices[:1] },
		"extra row":       func(v *CompletedResultV2) { v.Vertices = append(v.Vertices, v.Vertices[0]) },
		"vertex column":   func(v *CompletedResultV2) { v.Vertices[0].VertexKind = "different" },
		"vertex bytes":    func(v *CompletedResultV2) { v.Vertices[0].JSON = []byte("{}") },
		"edge column":     func(v *CompletedResultV2) { v.Edges[0].Direction = "undirected" },
		"edge bytes":      func(v *CompletedResultV2) { v.Edges[0].JSON = []byte("{}") },
		"vertex overflow": func(v *CompletedResultV2) { v.Vertices = make([]ResultVertexV2, MaximumResultVerticesV2+1) },
		"edge overflow":   func(v *CompletedResultV2) { v.Edges = make([]ResultEdgeV2, MaximumResultEdgesV2+1) },
	} {
		t.Run(name, func(t *testing.T) {
			value := complete(producer())
			mutate(&value)
			if err := value.Validate(context.Background()); !errors.Is(err, ErrResultV2Invalid) {
				t.Fatalf("invalid content accepted: %v", err)
			}
		})
	}
	// A matching content digest alone cannot admit structurally invalid output.
	for name, mutate := range map[string]func(*ProjectionResultV2){
		"duplicate vertex":   func(v *ProjectionResultV2) { v.Vertices[1] = v.Vertices[0] },
		"unordered vertices": func(v *ProjectionResultV2) { v.Vertices[0], v.Vertices[1] = v.Vertices[1], v.Vertices[0] },
		"missing endpoint":   func(v *ProjectionResultV2) { v.Edges[0].SrcVertexID = "vx_" + strings.Repeat("f", 64) },
		"unknown kind":       func(v *ProjectionResultV2) { v.Vertices[0].VertexKind = "unknown" },
		"invalid label": func(v *ProjectionResultV2) {
			v.Vertices[0].Labels = []string{strings.Repeat("x", graphProjectionLimits.MaxLabelBytes+1)}
		},
	} {
		t.Run(name, func(t *testing.T) {
			value := producer()
			mutate(&value)
			if err := reboundModelV2(t, value).Validate(t.Context()); !errors.Is(err, ErrResultV2Invalid) {
				t.Fatalf("invalid structure accepted: %v", err)
			}
		})
	}
}

func checkAggregateAdmissionV2(t *testing.T) {
	model, _, err := projectAggregationInputV2(t, mixedAggregationInputV2(false))
	if err != nil {
		t.Fatal(err)
	}
	var downstream SourceRef
	for _, v := range model.Vertices {
		if v.VertexKind == "final" {
			order := contributorOrder{v.VertexKind, v.VertexID, *v.Metadata.AggregationRuleID, v.VertexID}
			downstream = SourceRef{RefKind: "projected_vertex", RefID: v.VertexID, RefKindName: v.VertexKind, ContributorSortKey: order.key()}
		}
	}
	if downstream.RefID == "" {
		t.Fatal("missing downstream aggregate")
	}
	for name, mutate := range map[string]func(*Vertex){
		"empty contributors":   func(v *Vertex) { v.Metadata.AggregationSourceRefs = nil },
		"unknown scope":        func(v *Vertex) { v.Metadata.AggregationSourceRefs[0].RefKind = "unknown" },
		"unknown projected ID": func(v *Vertex) { v.Metadata.AggregationSourceRefs[0].RefID = "vx_" + strings.Repeat("f", 64) },
		"wrong kind":           func(v *Vertex) { v.Metadata.AggregationSourceRefs[0].RefKindName = "other" },
		"forged key":           func(v *Vertex) { v.Metadata.AggregationSourceRefs[0].ContributorSortKey += "00" },
		"another contributor's well-shaped key": func(v *Vertex) {
			refs := v.Metadata.AggregationSourceRefs
			refs[0].ContributorSortKey = refs[1].ContributorSortKey
			v.Metadata.AggregationSourceRefs = refs[:1]
		},
		"duplicate":        func(v *Vertex) { v.Metadata.AggregationSourceRefs[1] = v.Metadata.AggregationSourceRefs[0] },
		"reordered":        func(v *Vertex) { refs := v.Metadata.AggregationSourceRefs; refs[0], refs[1] = refs[1], refs[0] },
		"direct source":    func(v *Vertex) { v.SourceEntityRef = &SourceEntityRef{"source", "kind", "map"} },
		"missing producer": func(v *Vertex) { v.Metadata.AggregationRuleID = nil },
		"cyclic producer":  func(v *Vertex) { v.Metadata.AggregationSourceRefs = []SourceRef{downstream} },
		"self reference": func(v *Vertex) {
			order := contributorOrder{v.VertexKind, v.VertexID, *v.Metadata.AggregationRuleID, v.VertexID}
			v.Metadata.AggregationSourceRefs = []SourceRef{{RefKind: "projected_vertex", RefID: v.VertexID, RefKindName: v.VertexKind, ContributorSortKey: order.key()}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, _, err := projectAggregationInputV2(t, mixedAggregationInputV2(false))
			if err != nil {
				t.Fatal(err)
			}
			for i := range r.Vertices {
				if r.Vertices[i].VertexKind == "endpoints" {
					mutate(&r.Vertices[i])
				}
			}
			if err := reboundModelV2(t, r).Validate(t.Context()); !errors.Is(err, ErrResultV2Invalid) {
				t.Fatalf("accepted rebound aggregate: %v", err)
			}
		})
	}
}

func reboundModelV2(t *testing.T, r ProjectionResultV2) CompletedResultV2 {
	t.Helper()
	var err error
	r.CanonicalOutputSHA256, err = canonicalOutputDigestV2(r.outputResource())
	if err != nil {
		t.Fatal(err)
	}
	r.ProjectionResultID, err = deriveProjectionResultIDV2(r.ResultBindingV2())
	if err != nil {
		t.Fatal(err)
	}
	completed, err := r.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	return completed
}

func checkRegistryAdmissionV2(t *testing.T) {
	for _, tc := range []struct {
		name   string
		input  func() map[string]any
		mutate func(*ProjectionResultV2)
	}{
		{"invented producer", func() map[string]any { return mixedAggregationInputV2(false) }, func(r *ProjectionResultV2) {
			for i := range r.Edges {
				if r.Edges[i].EdgeKind == "tail" {
					id := "invented"
					r.Edges[i].Metadata.AggregationRuleID = &id
				}
			}
		}},
		{"producer from vertex scope", func() map[string]any { return mixedAggregationInputV2(false) }, func(r *ProjectionResultV2) {
			r.SchemaRegistry.VertexKinds[0].AggregationRuleIDs = append(r.SchemaRegistry.VertexKinds[0].AggregationRuleIDs, "unused")
			for i := range r.Edges {
				if r.Edges[i].EdgeKind == "tail" {
					id := "unused"
					r.Edges[i].Metadata.AggregationRuleID = &id
				}
			}
		}},
		{"unlisted aggregate direction", func() map[string]any { return mixedAggregationInputV2(false) }, func(r *ProjectionResultV2) {
			r.SchemaRegistry.EdgeKinds[0].Directions = []string{"undirected"}
		}},
		{"unlisted reverse direction", reverseInputV2, func(r *ProjectionResultV2) {
			for i, k := range r.SchemaRegistry.EdgeKinds {
				if k.EdgeKind == "back" {
					r.SchemaRegistry.EdgeKinds[i].Directions = []string{"undirected"}
				}
			}
		}},
		{"aggregate property type", func() map[string]any { return mixedAggregationInputV2(false) }, func(r *ProjectionResultV2) {
			r.Edges[0].Properties["score"] = "integer"
		}},
		{"direct source kind", fieldPathInputV2, func(r *ProjectionResultV2) { r.Vertices[0].SourceEntityRef.SourceEntityKind = "unlisted" }},
		{"reverse pair source kind", reverseInputV2, func(r *ProjectionResultV2) {
			for i := range r.Edges {
				r.Edges[i].SourceRelationshipRef.SourceRelationshipKind = "unlisted"
			}
		}},
		{"rule listed in two scopes", func() map[string]any { return mixedAggregationInputV2(true) }, func(r *ProjectionResultV2) {
			r.SchemaRegistry.EdgeKinds[0].AggregationRuleIDs = append(r.SchemaRegistry.EdgeKinds[0].AggregationRuleIDs, "root")
		}},
		{"unordered source kinds", fieldPathInputV2, func(r *ProjectionResultV2) {
			r.SchemaRegistry.VertexKinds[0].SourceEntityKinds = []string{"kind-b", "kind-a"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _, err := projectAggregationInputV2(t, tc.input())
			if err != nil {
				t.Fatal(err)
			}
			if err := reboundModelV2(t, r).Validate(t.Context()); err != nil {
				t.Fatal(err)
			}
			tc.mutate(&r)
			if err := reboundModelV2(t, r).Validate(t.Context()); !errors.Is(err, ErrResultV2Invalid) {
				t.Fatalf("registry mismatch admitted: %v", err)
			}
		})
	}
	r, _, err := projectAggregationInputV2(t, mixedAggregationInputV2(true))
	if err != nil {
		t.Fatal(err)
	}
	input := twoEntityProjectionInputV2(false)
	config := input["projection_config"].(map[string]any)
	kind := config["entity_mappings"].([]any)[0].(map[string]any)["projected_vertex_kind"].(string)
	rule := aggregationRuleInputV2("self-kind", "vertex", "projected_vertex", kind, kind)
	rule["grouping_keys"] = []any{"projected.vertex_id"}
	config["aggregation_rules"] = []any{rule}
	self, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := reboundModelV2(t, self).Validate(t.Context()); err != nil {
		t.Fatalf("own direct kind: %v", err)
	}
	// A declared rule may share an output kind without contributing any rows.
	r.SchemaRegistry.EdgeKinds[0].AggregationRuleIDs = append(r.SchemaRegistry.EdgeKinds[0].AggregationRuleIDs, "unused")
	if err := reboundModelV2(t, r).Validate(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func checkAggregateEndpointProofV2(t *testing.T) {
	for _, side := range []string{"source", "destination"} {
		for _, kind := range []string{"final", "vertex"} {
			t.Run(side+" "+kind, func(t *testing.T) {
				r, _, err := projectAggregationInputV2(t, mixedAggregationInputV2(false))
				if err != nil {
					t.Fatal(err)
				}
				for _, v := range r.Vertices {
					if v.VertexKind != kind {
						continue
					}
					for i := range r.Edges {
						e := &r.Edges[i]
						if e.EdgeKind != "bridge" {
							continue
						}
						if side == "source" {
							e.SrcVertexID = v.VertexID
						} else {
							e.DstVertexID = v.VertexID
						}
					}
					break
				}
				if err := reboundModelV2(t, r).Validate(t.Context()); !errors.Is(err, ErrResultV2Invalid) {
					t.Fatalf("impossible endpoint admitted: %v", err)
				}
			})
		}
	}
}

func checkEmittedFieldsV2(t *testing.T) {
	for _, scope := range []string{"graph_view", "vertex", "edge"} {
		for _, metadata := range []bool{false, true} {
			if scope == "graph_view" && !metadata {
				continue
			}
			name := scope + "/properties"
			if metadata {
				name = scope + "/mapped_metadata"
			}
			t.Run(name, func(t *testing.T) {
				input := reverseInputV2()
				kind, path := scope, "properties.absent"
				if scope == "graph_view" {
					kind, path = "*", "source_metadata.absent"
				}
				for _, field := range []struct{ key, typ string }{{"scalar", "integer"}, {"array", "string_array"}, {"nullable", "identifier"}} {
					def := aggregationPropertyV2(field.key, scope, kind, path, field.key, field.typ, "single_value")
					def["required"], def["missing_behavior"] = false, "omit"
					if field.key == "nullable" {
						def["null_output_policy"], def["source_null_behavior"] = "emit_null", "emit_null"
					}
					appendDefinitionV2(input, def, metadata)
					if metadata {
						def["metadata_mapping_id"] = field.key
					}
				}
				r, _, err := projectAggregationInputV2(t, input)
				if err != nil {
					t.Fatal(err)
				}
				var props map[string]any
				meta := r.MappedMetadata
				if scope == "vertex" {
					props, meta = r.Vertices[0].Properties, r.Vertices[0].Metadata.MappedMetadata
				}
				if scope == "edge" {
					for _, e := range r.Edges {
						if e.EdgeKind == "back" {
							values := e.Properties
							if metadata {
								values = e.Metadata.MappedMetadata
							}
							values["scalar"] = 1
							if !errors.Is(reboundModelV2(t, r).Validate(t.Context()), ErrResultV2Invalid) {
								t.Error("reverse kind borrowed forward declaration")
							}
							delete(values, "scalar")
						} else {
							props, meta = e.Properties, e.Metadata.MappedMetadata
						}
					}
				}
				fields, other := props, meta
				if metadata {
					fields, other = meta, props
				}
				for _, tc := range []struct {
					key   string
					value any
					valid bool
				}{
					{"scalar", int64(123), true}, {"array", []any{"x"}, true}, {"nullable", nil, true},
					{"other", 1, false}, {"scalar", "1", false}, {"scalar", nil, false},
					{"array", "x", false}, {"array", []any{1}, false}, {"array", []any{nil}, false},
				} {
					fields[tc.key] = tc.value
					err := reboundModelV2(t, r).Validate(t.Context())
					if (err == nil) != tc.valid || err != nil && !errors.Is(err, ErrResultV2Invalid) {
						t.Errorf("%s value %#v valid=%t: %v", tc.key, tc.value, tc.valid, err)
					}
					delete(fields, tc.key)
				}
				if err := reboundModelV2(t, r).Validate(t.Context()); err != nil {
					t.Fatalf("valid omission: %v", err)
				}
				if other == nil {
					return
				}
				other["scalar"] = 123
				if !errors.Is(reboundModelV2(t, r).Validate(t.Context()), ErrResultV2Invalid) {
					t.Error("opposite namespace borrowed declaration")
				}
			})
		}
	}
}

func checkDirectMappingClaimsV2(t *testing.T) {
	for _, tc := range []struct {
		name, kind, source, rule string
		valid                    bool
	}{
		{"another entity of same kind", "kind-a", "third", "map-a", true},
		{"same source across kinds", "kind-b", "entity-a", "map-b", false},
		{"second mapping for source kind", "kind-a", "third", "other", false},
		{"same mapping across source kinds", "third-kind", "third", "map-b", false},
		{"relationship mapping used by vertex", "third-kind", "third", "relation-map", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _, err := projectAggregationInputV2(t, fieldPathInputV2())
			if err != nil {
				t.Fatal(err)
			}
			// This kind has a declared producer but no other observed mapping.
			r.SchemaRegistry.VertexKinds[0].SourceEntityKinds = append(r.SchemaRegistry.VertexKinds[0].SourceEntityKinds, "third-kind")
			vertex := r.Vertices[0]
			vertex.Metadata.MappingRuleID = &tc.rule
			vertex.SourceEntityRef = &SourceEntityRef{SourceEntityID: tc.source, SourceEntityKind: tc.kind, MappingRuleID: tc.rule}
			digest := entityMappingDigestV2(tc.rule, tc.kind, vertex.VertexKind)
			vertex.VertexID, vertex.SortKey = directVertexIdentityV2(r.GraphViewID, vertex.VertexKind, tc.kind, tc.source, digest)
			r.Vertices = append(r.Vertices, vertex)
			sortVertices(r.Vertices)
			err = reboundModelV2(t, r).Validate(t.Context())
			if (err == nil) != tc.valid || err != nil && !errors.Is(err, ErrResultV2Invalid) {
				t.Errorf("valid=%t: %v", tc.valid, err)
			}
		})
	}
	r, _, err := projectAggregationInputV2(t, fieldPathInputV2())
	if err != nil {
		t.Fatal(err)
	}
	r.SchemaRegistry.VertexKinds = append(r.SchemaRegistry.VertexKinds, VertexKindSchema{VertexKind: "z-unused", SourceEntityKinds: []string{"kind-a"}})
	if !errors.Is(reboundModelV2(t, r).Validate(t.Context()), ErrResultV2Invalid) {
		t.Error("source kind registered for multiple output kinds")
	}
}

func checkRuleGroupsV2(t *testing.T) {
	for _, tc := range []struct {
		name, scope, kind, id string
		swap, valid           bool
	}{
		{"distinct groups", "source_entity", "kind-a", "b", false, true},
		{"overlapping groups", "source_entity", "kind-a", "a", false, false},
		{"different scope", "source_relationship", "kind-a", "b", false, false},
		{"different kind", "source_entity", "kind-b", "b", false, false},
		{"exchanged endpoint producers", "source_entity", "kind-a", "b", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _, err := projectAggregationInputV2(t, mixedAggregationInputV2(false))
			if err != nil {
				t.Fatal(err)
			}
			rawRef := func(scope, kind, id string) []SourceRef {
				return []SourceRef{{RefKind: scope, RefKindName: kind, RefID: id, ContributorSortKey: (contributorOrder{kind, id, "", ""}).key()}}
			}
			var extra Edge
			for i := range r.Edges {
				edge := &r.Edges[i]
				edge.Metadata.AggregationSourceRefs = rawRef("source_entity", "kind-a", "a") // Different rules may consume a common input.
				if edge.EdgeKind == "tail" {
					for _, v := range r.Vertices {
						if v.VertexKind == "final" {
							edge.DstVertexID = v.VertexID
						}
					}
					extra = *edge
				}
			}
			extra.EdgeID = "ed_" + strings.Repeat("f", 64)
			extra.Metadata.AggregationSourceRefs = rawRef(tc.scope, tc.kind, tc.id)
			if tc.swap {
				extra.SrcVertexID, extra.DstVertexID = extra.DstVertexID, extra.SrcVertexID
			}
			r.Edges = append(r.Edges, extra)
			sortEdges(r.Edges)
			err = reboundModelV2(t, r).Validate(t.Context())
			if (err == nil) != tc.valid || err != nil && !errors.Is(err, ErrResultV2Invalid) {
				t.Fatalf("group proof valid=%t: %v", tc.valid, err)
			}
			if tc.valid {
				for i := range r.SchemaRegistry.EdgeKinds {
					k := &r.SchemaRegistry.EdgeKinds[i]
					if k.EdgeKind == extra.EdgeKind {
						k.SourceRelationshipKinds = []string{"dormant"}
						k.Directions = []string{"directed", "undirected"}
					}
				}
				requireReboundAdmissionV2(t, r, true)
				for i := range r.Edges {
					if r.Edges[i].EdgeID == extra.EdgeID {
						r.Edges[i].Direction = "undirected"
					}
				}
				requireReboundAdmissionV2(t, r, false)
			}
		})
	}
}
