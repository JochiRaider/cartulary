package graphprojection_test

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

func minimalInput(t *testing.T, key string) map[string]any {
	t.Helper()
	graphViewID, err := DeriveGraphViewID(key)
	if err != nil {
		t.Fatalf("derive graph view id: %v", err)
	}
	return map[string]any{
		"projection_schema_id": "graph_projection.v1", "graph_view_id": graphViewID,
		"source_snapshot_id": "snap_empty_1",
		"projection_config":  map[string]any{"graph_view_key": key, "declared_source_entity_kinds": []any{}, "entity_mappings": []any{}, "allow_empty_kind_registry": true},
		"source_entities":    []any{}, "source_relationships": []any{},
		"requested_at": "2026-05-30T00:00:00Z", "requested_by": "fixture",
	}
}

func incidentGraphInput(t *testing.T) map[string]any {
	t.Helper()
	graphViewID, err := DeriveGraphViewID("incident_graph")
	if err != nil {
		t.Fatalf("derive graph view id: %v", err)
	}
	return map[string]any{
		"projection_schema_id": "graph_projection.v1", "graph_view_id": graphViewID, "source_snapshot_id": "snap_incident_1",
		"projection_config": map[string]any{
			"graph_view_key": "incident_graph", "declared_source_entity_kinds": []any{"host"}, "declared_source_relationship_kinds": []any{"logon"},
			"default_vertex_labels": []any{"asset"}, "default_edge_labels": []any{"activity"},
			"entity_mappings":       []any{map[string]any{"mapping_rule_id": "map_host", "source_entity_kind": "host", "projected_vertex_kind": "host_vertex", "required_property_keys": []any{"hostname"}, "mapping_labels": []any{"host"}}},
			"relationship_mappings": []any{map[string]any{"mapping_rule_id": "map_logon", "source_relationship_kind": "logon", "projected_edge_kind": "logon_edge", "direction_policy": "normalize_forward", "emit_reverse_edge": true, "reverse_edge_kind": "logged_on_by_edge"}},
			"aggregation_rules": []any{
				map[string]any{"aggregation_rule_id": "agg_host_by_site", "target_scope": "vertex", "input_scope": "source_entity", "input_kind": "host", "projected_kind": "site_vertex", "grouping_keys": []any{"properties.site"}, "missing_grouping_key_behavior": "error", "property_merge_behavior": map[string]any{"host_count": "count"}},
				map[string]any{"aggregation_rule_id": "agg_logon_by_site", "target_scope": "edge", "input_scope": "source_relationship", "input_kind": "logon", "projected_kind": "site_logon_edge", "grouping_keys": []any{"properties.site"}, "missing_grouping_key_behavior": "error", "edge_direction": "directed", "property_merge_behavior": map[string]any{"logon_count": "count"}, "endpoint_grouping": map[string]any{"src_vertex_aggregation_rule_id": "agg_host_by_site", "src_grouping_keys": []any{"properties.src_site"}, "dst_vertex_aggregation_rule_id": "agg_host_by_site", "dst_grouping_keys": []any{"properties.dst_site"}, "missing_endpoint_behavior": "error"}},
			},
		},
		"source_entities": []any{
			map[string]any{"source_entity_id": "host1", "source_entity_kind": "host", "properties": map[string]any{"hostname": "WS-023", "site": "hq"}, "labels": []any{"workstation"}},
			map[string]any{"source_entity_id": "host2", "source_entity_kind": "host", "properties": map[string]any{"hostname": "SRV-01", "site": "hq"}},
		},
		"source_relationships": []any{map[string]any{"source_relationship_id": "logon1", "source_relationship_kind": "logon", "src_source_entity_id": "host1", "dst_source_entity_id": "host2", "direction": "forward", "properties": map[string]any{"site": "hq", "src_site": "hq", "dst_site": "hq"}}},
		"property_definitions": []any{
			map[string]any{"property_definition_id": "pd_hostname", "target_scope": "vertex", "target_kind": "host_vertex", "source_field_path": "properties.hostname", "projected_key": "hostname", "projected_type": "string", "required": true},
			map[string]any{"property_definition_id": "pd_host_count", "target_scope": "vertex", "target_kind": "site_vertex", "source_field_path": "properties.hostname", "projected_key": "host_count", "projected_type": "integer", "merge_behavior": "count"},
			map[string]any{"property_definition_id": "pd_logon_count", "target_scope": "edge", "target_kind": "site_logon_edge", "source_field_path": "source_relationship_id", "projected_key": "logon_count", "projected_type": "integer", "merge_behavior": "count"},
		},
		"source_metadata": map[string]any{"case": "alpha"}, "requested_at": "2026-05-30T00:00:00Z", "requested_by": "fixture",
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	return data
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC)
}
