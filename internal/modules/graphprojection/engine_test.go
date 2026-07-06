package graphprojection

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestCanonicalDigestAndIDs(t *testing.T) {
	t.Parallel()

	emptyID, err := DeriveGraphViewID("empty")
	if err != nil {
		t.Fatalf("derive empty graph view id: %v", err)
	}
	if emptyID != "gv_0bfa120793d470c3cf37aa2c6ac0f69c067fa2e598da3f5116512b92f3bc3752" {
		t.Fatalf("empty graph view id = %s", emptyID)
	}

	incidentID, err := DeriveGraphViewID("incident_graph")
	if err != nil {
		t.Fatalf("derive incident graph view id: %v", err)
	}
	if incidentID != "gv_7b34489c234cb6caa432f92afc6fb122788e525f8bb057d214c96be9289d8893" {
		t.Fatalf("incident graph view id = %s", incidentID)
	}
}

func TestDuplicateObjectMembersRejected(t *testing.T) {
	t.Parallel()

	_, err := AdmitProjectionInput([]byte(`{"projection_schema_id":"graph_projection.v1","projection_schema_id":"graph_projection.v1"}`), AdmitOptions{})
	var opErr *OperationError
	if !errors.As(err, &opErr) {
		t.Fatalf("expected operation error, got %T %v", err, err)
	}
	if opErr.ReasonCode != "duplicate_object_member" {
		t.Fatalf("reason = %s", opErr.ReasonCode)
	}
}

func TestAdmissionRejectsInvalidGraphViewID(t *testing.T) {
	t.Parallel()

	input := minimalInput(t, "empty")
	input["graph_view_id"] = "gv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	_, err := AdmitProjectionInput(mustJSON(t, input), AdmitOptions{})
	var opErr *OperationError
	if !errors.As(err, &opErr) {
		t.Fatalf("expected operation error, got %T %v", err, err)
	}
	if opErr.ReasonCode != "invalid_graph_view_id" {
		t.Fatalf("reason = %s", opErr.ReasonCode)
	}
}

func TestAdmissionAndNormalization(t *testing.T) {
	t.Parallel()

	input := minimalInput(t, "empty")
	run, err := AdmitProjectionInput(mustJSON(t, input), AdmitOptions{
		ProjectionRunNonce: "nonce-1",
		AcceptedAt:         fixedTime(),
	})
	if err != nil {
		t.Fatalf("admit: %v", err)
	}
	if run.GraphViewID != input["graph_view_id"] {
		t.Fatalf("graph view id drift: %s", run.GraphViewID)
	}
	if run.ProjectionRunID == "" || run.ProjectionConfigDigest == "" || run.ProjectionSourceDigest == "" {
		t.Fatalf("missing derived run identifiers: %#v", run)
	}
	if got := run.Request.ProjectionConfig.RetentionPolicy.RetentionCount; got != defaultRetentionCount {
		t.Fatalf("retention count default = %d", got)
	}
}

func TestProjectDirectReverseAndAggregation(t *testing.T) {
	t.Parallel()

	input := incidentGraphInput(t)
	run, err := Project(mustJSON(t, input), ProjectOptions{
		ProjectionRunNonce: "nonce-direct",
		AcceptedAt:         fixedTime(),
		GeneratedAt:        fixedTime().Add(time.Second),
	})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if run.State != RunStateAvailable {
		t.Fatalf("state = %s issues=%v", run.State, run.ValidationSummary.Issues)
	}
	if run.GraphView == nil {
		t.Fatal("missing graph view")
	}
	graphBytes, err := canonicalJSON(run.GraphView)
	if err != nil {
		t.Fatalf("canonical graph view: %v", err)
	}
	if want := sha256Hex(graphBytes); run.ProjectionOutputDigest != want {
		t.Fatalf("projection output digest = %s want %s", run.ProjectionOutputDigest, want)
	}
	if len(run.GraphView.Vertices) != 3 {
		t.Fatalf("vertices = %d; %#v", len(run.GraphView.Vertices), run.GraphView.Vertices)
	}
	if len(run.GraphView.Edges) != 3 {
		t.Fatalf("edges = %d; %#v", len(run.GraphView.Edges), run.GraphView.Edges)
	}
	var reverseFound bool
	var aggregateFound bool
	for _, edge := range run.GraphView.Edges {
		if edge.EdgeFamily == "reverse" && edge.Metadata.ReverseOfEdgeID != nil && edge.Direction == "directed" {
			reverseFound = true
		}
		if edge.EdgeFamily == "aggregated" {
			aggregateFound = true
		}
	}
	if !reverseFound {
		t.Fatal("expected reverse edge")
	}
	if !aggregateFound {
		t.Fatal("expected aggregated edge")
	}
}

func TestFilterTruthTable(t *testing.T) {
	t.Parallel()

	input := incidentGraphInput(t)
	input["filters"] = map[string]any{
		"entity_filters": []any{
			map[string]any{"field_path": "properties.hostname", "operator": "eq", "value": "WS-023"},
		},
		"relationship_filters": []any{},
		"logic":                "and",
	}
	run, err := Project(mustJSON(t, input), ProjectOptions{ProjectionRunNonce: "nonce-filter", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	var directVertices int
	for _, vertex := range run.GraphView.Vertices {
		if vertex.VertexFamily == "direct" {
			directVertices++
		}
	}
	if directVertices != 1 {
		t.Fatalf("filtered direct vertices = %d", directVertices)
	}
}

func TestCanonicalJSON(t *testing.T) {
	t.Parallel()

	encoded, err := canonicalJSON(map[string]any{
		"b": "line\nslash/sep\u2028",
		"a": []any{json.Number("1"), true, nil},
	})
	if err != nil {
		t.Fatalf("canonical json: %v", err)
	}
	want := `{"a":[1,true,null],"b":"line\nslash/sep` + "\u2028" + `"}`
	if string(encoded) != want {
		t.Fatalf("canonical json = %q want %q", string(encoded), want)
	}
}

func TestScalarAndFieldPathValidation(t *testing.T) {
	t.Parallel()

	if !validIdentifier("host_1") || validIdentifier(" host_1") || validIdentifier("host/1") {
		t.Fatal("identifier validation mismatch")
	}
	if !validPropertyKey("hostname") || validPropertyKey("properties.bad") || validPropertyKey("metadata") {
		t.Fatal("property key validation mismatch")
	}
	if !validFieldPath("properties.hostname") || validFieldPath("properties.bad.extra") {
		t.Fatal("field path validation mismatch")
	}
}

func TestFailedRunInspectionShape(t *testing.T) {
	t.Parallel()

	input := incidentGraphInput(t)
	config := input["projection_config"].(map[string]any)
	relMappings := config["relationship_mappings"].([]any)
	relMappings[0].(map[string]any)["emit_reverse_edge"] = true
	relMappings[0].(map[string]any)["direction_policy"] = "preserve"
	run, err := Project(mustJSON(t, input), ProjectOptions{ProjectionRunNonce: "nonce-failed", AcceptedAt: fixedTime(), GeneratedAt: fixedTime()})
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if run.State != RunStateFailed {
		t.Fatalf("state = %s", run.State)
	}
	if run.GraphView != nil {
		t.Fatal("failed run should not expose consumable graph")
	}
	if run.ValidationSummary.Status != "failed" || len(run.ValidationSummary.Issues) == 0 {
		t.Fatalf("validation summary = %#v", run.ValidationSummary)
	}
}

func minimalInput(t *testing.T, key string) map[string]any {
	t.Helper()
	graphViewID, err := DeriveGraphViewID(key)
	if err != nil {
		t.Fatalf("derive graph view id: %v", err)
	}
	return map[string]any{
		"projection_schema_id": "graph_projection.v1",
		"graph_view_id":        graphViewID,
		"source_snapshot_id":   "snap_empty_1",
		"projection_config": map[string]any{
			"graph_view_key":               key,
			"declared_source_entity_kinds": []any{},
			"entity_mappings":              []any{},
			"allow_empty_kind_registry":    true,
		},
		"source_entities":      []any{},
		"source_relationships": []any{},
		"requested_at":         "2026-05-30T00:00:00Z",
		"requested_by":         "fixture",
	}
}

func incidentGraphInput(t *testing.T) map[string]any {
	t.Helper()
	graphViewID, err := DeriveGraphViewID("incident_graph")
	if err != nil {
		t.Fatalf("derive graph view id: %v", err)
	}
	return map[string]any{
		"projection_schema_id": "graph_projection.v1",
		"graph_view_id":        graphViewID,
		"source_snapshot_id":   "snap_incident_1",
		"projection_config": map[string]any{
			"graph_view_key":                     "incident_graph",
			"declared_source_entity_kinds":       []any{"host"},
			"declared_source_relationship_kinds": []any{"logon"},
			"default_vertex_labels":              []any{"asset"},
			"default_edge_labels":                []any{"activity"},
			"entity_mappings": []any{
				map[string]any{
					"mapping_rule_id":        "map_host",
					"source_entity_kind":     "host",
					"projected_vertex_kind":  "host_vertex",
					"required_property_keys": []any{"hostname"},
					"mapping_labels":         []any{"host"},
				},
			},
			"relationship_mappings": []any{
				map[string]any{
					"mapping_rule_id":          "map_logon",
					"source_relationship_kind": "logon",
					"projected_edge_kind":      "logon_edge",
					"direction_policy":         "normalize_forward",
					"emit_reverse_edge":        true,
					"reverse_edge_kind":        "logged_on_by_edge",
				},
			},
			"aggregation_rules": []any{
				map[string]any{
					"aggregation_rule_id":           "agg_host_by_site",
					"target_scope":                  "vertex",
					"input_scope":                   "source_entity",
					"input_kind":                    "host",
					"projected_kind":                "site_vertex",
					"grouping_keys":                 []any{"properties.site"},
					"missing_grouping_key_behavior": "error",
					"property_merge_behavior":       map[string]any{"host_count": "count"},
				},
				map[string]any{
					"aggregation_rule_id":           "agg_logon_by_site",
					"target_scope":                  "edge",
					"input_scope":                   "source_relationship",
					"input_kind":                    "logon",
					"projected_kind":                "site_logon_edge",
					"grouping_keys":                 []any{"properties.site"},
					"missing_grouping_key_behavior": "error",
					"edge_direction":                "directed",
					"property_merge_behavior":       map[string]any{"logon_count": "count"},
					"endpoint_grouping": map[string]any{
						"src_vertex_aggregation_rule_id": "agg_host_by_site",
						"src_grouping_keys":              []any{"properties.src_site"},
						"dst_vertex_aggregation_rule_id": "agg_host_by_site",
						"dst_grouping_keys":              []any{"properties.dst_site"},
						"missing_endpoint_behavior":      "error",
					},
				},
			},
		},
		"source_entities": []any{
			map[string]any{"source_entity_id": "host1", "source_entity_kind": "host", "properties": map[string]any{"hostname": "WS-023", "site": "hq"}, "labels": []any{"workstation"}},
			map[string]any{"source_entity_id": "host2", "source_entity_kind": "host", "properties": map[string]any{"hostname": "SRV-01", "site": "hq"}},
		},
		"source_relationships": []any{
			map[string]any{"source_relationship_id": "logon1", "source_relationship_kind": "logon", "src_source_entity_id": "host1", "dst_source_entity_id": "host2", "direction": "forward", "properties": map[string]any{"site": "hq", "src_site": "hq", "dst_site": "hq"}},
		},
		"property_definitions": []any{
			map[string]any{"property_definition_id": "pd_hostname", "target_scope": "vertex", "target_kind": "host_vertex", "source_field_path": "properties.hostname", "projected_key": "hostname", "projected_type": "string", "required": true},
			map[string]any{"property_definition_id": "pd_host_count", "target_scope": "vertex", "target_kind": "site_vertex", "source_field_path": "properties.hostname", "projected_key": "host_count", "projected_type": "integer", "merge_behavior": "count"},
			map[string]any{"property_definition_id": "pd_logon_count", "target_scope": "edge", "target_kind": "site_logon_edge", "source_field_path": "source_relationship_id", "projected_key": "logon_count", "projected_type": "integer", "merge_behavior": "count"},
		},
		"source_metadata": map[string]any{"case": "alpha"},
		"requested_at":    "2026-05-30T00:00:00Z",
		"requested_by":    "fixture",
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
