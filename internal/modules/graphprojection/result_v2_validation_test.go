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
		"trailing LF":      func(v *CompletedResultV2) { v.ResultJSON = append(v.ResultJSON, '\n') },
		"two values":       func(v *CompletedResultV2) { v.ResultJSON = append(v.ResultJSON, []byte("{}")...) },
		"duplicate member": func(v *CompletedResultV2) { v.ResultJSON = append([]byte(`{"properties":{},`), v.ResultJSON[1:]...) },
		"unknown field":    func(v *CompletedResultV2) { v.ResultJSON = append([]byte(`{"unknown":true,`), v.ResultJSON[1:]...) },
		"invalid UTF8":     func(v *CompletedResultV2) { v.ResultJSON = append(v.ResultJSON, 0xff) },
		"null root":        func(v *CompletedResultV2) { v.ResultJSON = []byte("null") },
		"wrong tuple":      func(v *CompletedResultV2) { v.Binding.SourceOwnerID = "different" },
		"wrong digest":     func(v *CompletedResultV2) { v.Binding.CanonicalOutputSHA256 = strings.Repeat("a", 64) },
		"missing row":      func(v *CompletedResultV2) { v.Vertices = v.Vertices[:1] },
		"extra row":        func(v *CompletedResultV2) { v.Vertices = append(v.Vertices, v.Vertices[0]) },
		"vertex column":    func(v *CompletedResultV2) { v.Vertices[0].VertexKind = "different" },
		"vertex bytes":     func(v *CompletedResultV2) { v.Vertices[0].JSON = []byte("{}") },
		"edge column":      func(v *CompletedResultV2) { v.Edges[0].Direction = "undirected" },
		"edge bytes":       func(v *CompletedResultV2) { v.Edges[0].JSON = []byte("{}") },
		"vertex overflow":  func(v *CompletedResultV2) { v.Vertices = make([]ResultVertexV2, MaximumResultVerticesV2+1) },
		"edge overflow":    func(v *CompletedResultV2) { v.Edges = make([]ResultEdgeV2, MaximumResultEdgesV2+1) },
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
			value.CanonicalOutputSHA256, err = canonicalOutputDigestV2(value.outputResource())
			if err != nil {
				t.Fatal(err)
			}
			value.ProjectionResultID, err = deriveProjectionResultIDV2(value.ResultBindingV2())
			if err != nil {
				t.Fatal(err)
			}
			if err := complete(value).Validate(context.Background()); !errors.Is(err, ErrResultV2Invalid) {
				t.Fatalf("invalid structure accepted: %v", err)
			}
		})
	}
}
