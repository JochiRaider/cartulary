// Package graphfixture constructs semantic input and invokes the current Graph
// engine. It never encodes output or derives result/object identities itself.
package graphfixture

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

type ResultOptions struct {
	EdgeStringProperties                                            map[string]string
	VertexLabels                                                    []string
	GraphViewID, SourceOwnerID, SourceSnapshotID, ProjectionVersion string
	VertexKind, EdgeKind                                            string
	VertexCount                                                     int
	PublishedAt                                                     time.Time
}

func Result(t testing.TB, options ResultOptions) graphprojection.CompletedResultV2 {
	t.Helper()
	if options.VertexLabels == nil {
		options.VertexLabels = []string{}
	}
	entities := make([]any, 0, options.VertexCount)
	for index := 0; index < options.VertexCount; index++ {
		entities = append(entities, map[string]any{
			"source_entity_id": fmt.Sprintf("nfe_%064x", index+1), "source_entity_kind": "entity",
			"properties": map[string]any{"endpoint_value": fmt.Sprintf("192.0.2.%d", index+1)}, "metadata": map[string]any{}, "labels": []string{},
		})
	}
	edgeProperties := map[string]any{"ip_protocol": 6, "dst_port": 443}
	edgeKeys := []string{"dst_port", "ip_protocol"}
	extraKeys := make([]string, 0, len(options.EdgeStringProperties))
	for key, value := range options.EdgeStringProperties {
		edgeProperties[key] = value
		extraKeys = append(extraKeys, key)
	}
	sort.Strings(extraKeys)
	edgeKeys = append(edgeKeys, extraKeys...)
	relationships := []any{}
	if options.VertexCount >= 2 {
		relationships = append(relationships, map[string]any{
			"source_relationship_id": fmt.Sprintf("nff_%064x", 1), "source_relationship_kind": "relationship",
			"src_source_entity_id": fmt.Sprintf("nfe_%064x", 1), "dst_source_entity_id": fmt.Sprintf("nfe_%064x", 2), "direction": "forward",
			"properties": edgeProperties, "metadata": map[string]any{}, "labels": []string{},
		})
	}
	definitions := []any{}
	for _, field := range []struct{ scope, kind, name, valueType string }{
		{"vertex", options.VertexKind, "endpoint_value", "string"},
		{"edge", options.EdgeKind, "ip_protocol", "integer"}, {"edge", options.EdgeKind, "dst_port", "integer"},
	} {
		definitions = append(definitions, map[string]any{
			"property_definition_id": "fixture." + field.name, "target_scope": field.scope, "target_kind": field.kind,
			"source_field_path": "properties." + field.name, "projected_key": field.name, "projected_type": field.valueType,
			"required": true, "missing_behavior": "error", "source_null_behavior": "error", "null_output_policy": "omit", "merge_behavior": "single_value",
		})
	}
	for _, key := range extraKeys {
		definitions = append(definitions, map[string]any{"property_definition_id": "fixture." + key, "target_scope": "edge", "target_kind": options.EdgeKind, "source_field_path": "properties." + key, "projected_key": key, "projected_type": "string", "required": true, "missing_behavior": "error", "source_null_behavior": "error", "null_output_policy": "omit", "merge_behavior": "single_value"})
	}
	input := map[string]any{
		"projection_schema_id": graphprojection.ProjectionSchemaIDV2, "source_snapshot_id": options.SourceSnapshotID,
		"projection_config": map[string]any{
			"projection_version": options.ProjectionVersion, "declared_source_entity_kinds": []string{"entity"}, "declared_source_relationship_kinds": []string{"relationship"},
			"entity_mappings": []any{map[string]any{
				"mapping_rule_id": "fixture.entity", "source_entity_kind": "entity", "projected_vertex_kind": options.VertexKind,
				"inclusion_predicate": "always", "label_policy": "mapping_only", "mapping_labels": options.VertexLabels, "required_property_keys": []string{"endpoint_value"}, "optional_property_keys": []string{},
			}},
			"relationship_mappings": []any{map[string]any{
				"mapping_rule_id": "fixture.relationship", "source_relationship_kind": "relationship", "projected_edge_kind": options.EdgeKind,
				"inclusion_predicate": "always", "direction_policy": "preserve", "emit_reverse_edge": false,
				"label_policy": "mapping_only", "mapping_labels": []string{}, "required_property_keys": edgeKeys, "optional_property_keys": []string{},
			}},
			"metadata_mappings": []any{}, "aggregation_rules": []any{}, "default_vertex_labels": []string{}, "default_edge_labels": []string{}, "allow_empty_kind_registry": false,
		},
		"source_entities": entities, "source_relationships": relationships, "source_metadata": map[string]any{},
		"filters": map[string]any{"entity_filters": []any{}, "relationship_filters": []any{}, "logic": "and"}, "property_definitions": definitions,
	}
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	projected, err := graphprojection.ProjectV2(context.Background(), graphprojection.InvocationContextV2{GraphViewID: options.GraphViewID, SourceOwnerID: options.SourceOwnerID}, raw)
	if err != nil {
		t.Fatalf("project current Graph fixture: %v", err)
	}
	result, err := projected.CompletedResult()
	if err != nil {
		t.Fatal(err)
	}
	result.PublishedAt = options.PublishedAt
	return result
}
