package graphprojection

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	contractgraphprojection "github.com/JochiRaider/cartulary/internal/gen/contractgraphprojection"
)

func TestProjectV2GoldenDeterminismAndCompletedResult_Unit(t *testing.T) {
	fixture := graphProjectionV2GoldenFixture(t)
	trusted := fixture["trusted_context"].(map[string]any)
	input := fixture["input"].(map[string]any)
	semanticInput, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal golden semantic input: %v", err)
	}
	invocation := InvocationContextV2{
		GraphViewID:   trusted["graph_view_id"].(string),
		SourceOwnerID: trusted["source_owner_id"].(string),
	}
	first, err := ProjectV2(context.Background(), invocation, semanticInput)
	if err != nil {
		t.Fatalf("project golden input: %v", err)
	}
	second, err := ProjectV2(context.Background(), invocation, semanticInput)
	if err != nil {
		t.Fatalf("repeat golden input: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated projection differs:\nfirst=%#v\nsecond=%#v", first, second)
	}
	expected := fixture["expected"].(map[string]any)
	wantResultID, resultIDPresent := expected["projection_result_id"].(string)
	wantConfigDigest, configDigestPresent := expected["normalized_configuration_sha256"].(string)
	wantSourceDigest, sourceDigestPresent := expected["normalized_source_sha256"].(string)
	wantOutputDigest, outputDigestPresent := expected["canonical_output_sha256"].(string)
	if !resultIDPresent || !configDigestPresent || !sourceDigestPresent || !outputDigestPresent {
		t.Fatalf("golden identity values missing; projection_result_id=%s normalized_configuration_sha256=%s normalized_source_sha256=%s canonical_output_sha256=%s", first.ProjectionResultID, first.NormalizedConfigurationSHA256, first.NormalizedSourceSHA256, first.CanonicalOutputSHA256)
	}
	if first.ProjectionResultID != wantResultID ||
		first.NormalizedConfigurationSHA256 != wantConfigDigest ||
		first.NormalizedSourceSHA256 != wantSourceDigest ||
		first.CanonicalOutputSHA256 != wantOutputDigest {
		t.Fatalf("golden identity drifted: result=%s config=%s source=%s output=%s", first.ProjectionResultID, first.NormalizedConfigurationSHA256, first.NormalizedSourceSHA256, first.CanonicalOutputSHA256)
	}
	if first.ProjectionSchemaID != ProjectionSchemaIDV2 || len(first.Vertices) != 0 || len(first.Edges) != 0 || first.ValidationSummary.IssueCount != 0 {
		t.Fatalf("unexpected empty projection result: %#v", first)
	}
	resource := first.Resource()
	for _, forbidden := range []string{"projection_run_id", "ephemeral_projection_id", "state", "generated_at", "requested_at", "requested_by", "previous_projection_result_id"} {
		if _, present := resource[forbidden]; present {
			t.Fatalf("v2 result contains operational member %q", forbidden)
		}
	}
	completed, err := first.CompletedResult()
	if err != nil {
		t.Fatalf("create completed result: %v", err)
	}
	if completed.Binding != first.ResultBindingV2() || len(completed.ResultJSON) == 0 || len(completed.Vertices) != 0 || len(completed.Edges) != 0 || !completed.PublishedAt.IsZero() {
		t.Fatalf("completed result does not preserve the pure semantic boundary: %#v", completed)
	}
}

func TestProjectV2RejectsRemovedUnknownAndUnsafeMembers_Unit(t *testing.T) {
	fixture := graphProjectionV2GoldenFixture(t)
	trusted := fixture["trusted_context"].(map[string]any)
	invocation := InvocationContextV2{GraphViewID: trusted["graph_view_id"].(string), SourceOwnerID: trusted["source_owner_id"].(string)}
	base := fixture["input"].(map[string]any)
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "requested_at", mutate: func(value map[string]any) { value["requested_at"] = "2026-08-15T00:00:00Z" }},
		{name: "requested_by", mutate: func(value map[string]any) { value["requested_by"] = "spoofed-actor" }},
		{name: "graph_view_id", mutate: func(value map[string]any) { value["graph_view_id"] = "spoofed-view" }},
		{name: "source_owner_id", mutate: func(value map[string]any) { value["source_owner_id"] = "spoofed-owner" }},
		{name: "graph_view_key", mutate: func(value map[string]any) {
			value["projection_config"].(map[string]any)["graph_view_key"] = "spoofed-view"
		}},
		{name: "op", mutate: func(value map[string]any) {
			value["filters"].(map[string]any)["entity_filters"] = []any{map[string]any{
				"field_path": "source_entity_id", "op": "exists", "include_if_missing": false,
			}}
		}},
		{name: "retention_policy", mutate: func(value map[string]any) {
			value["projection_config"].(map[string]any)["retention_policy"] = map[string]any{}
		}},
		{name: "custom_config", mutate: func(value map[string]any) {
			value["projection_config"].(map[string]any)["custom_config"] = map[string]any{}
		}},
		{name: "nested property", mutate: func(value map[string]any) {
			value["source_entities"] = []any{map[string]any{
				"source_entity_id": "entity", "source_entity_kind": "kind", "properties": map[string]any{"secret": map[string]any{"raw": "DO_NOT_ECHO"}}, "metadata": map[string]any{}, "labels": []any{},
			}}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := deepCopyJSONMap(t, base)
			test.mutate(input)
			encoded, err := json.Marshal(input)
			if err != nil {
				t.Fatalf("marshal mutated input: %v", err)
			}
			_, err = ProjectV2(context.Background(), invocation, encoded)
			var projectionErr *ProjectionErrorV2
			if !errors.As(err, &projectionErr) || projectionErr.Code != "invalid_projection_request" || projectionErr.RetryAction != "do_not_retry" {
				t.Fatalf("error = %#v, want closed invalid_projection_request", err)
			}
			errorJSON, marshalErr := json.Marshal(projectionErr)
			if marshalErr != nil {
				t.Fatalf("marshal safe error: %v", marshalErr)
			}
			if strings.Contains(string(errorJSON), "DO_NOT_ECHO") || strings.Contains(err.Error(), "DO_NOT_ECHO") {
				t.Fatalf("projection error disclosed source value: %s", errorJSON)
			}
		})
	}
}

func TestProjectV2TrustedInvocationDoesNotEnterConfigurationDigest_Unit(t *testing.T) {
	fixture := graphProjectionV2GoldenFixture(t)
	semanticInput, err := json.Marshal(fixture["input"])
	if err != nil {
		t.Fatalf("marshal semantic input: %v", err)
	}
	first, err := ProjectV2(context.Background(), InvocationContextV2{GraphViewID: "graph-view-a", SourceOwnerID: "network_flow_activity"}, semanticInput)
	if err != nil {
		t.Fatalf("project first trusted context: %v", err)
	}
	second, err := ProjectV2(context.Background(), InvocationContextV2{GraphViewID: "graph-view-b", SourceOwnerID: "other_owner"}, semanticInput)
	if err != nil {
		t.Fatalf("project second trusted context: %v", err)
	}
	if first.NormalizedConfigurationSHA256 != second.NormalizedConfigurationSHA256 ||
		first.NormalizedSourceSHA256 != second.NormalizedSourceSHA256 {
		t.Fatalf("trusted invocation changed semantic digests: first=%#v second=%#v", first.ResultBindingV2(), second.ResultBindingV2())
	}
	if first.ProjectionResultID == second.ProjectionResultID {
		t.Fatal("trusted result identity fields did not distinguish separate invocations")
	}
}

func TestProjectV2SemanticOrderingAndNumericNormalization_Unit(t *testing.T) {
	left := twoEntityProjectionInputV2(false)
	right := twoEntityProjectionInputV2(true)
	leftBytes, _ := json.Marshal(left)
	rightBytes, _ := json.Marshal(right)
	invocation := InvocationContextV2{GraphViewID: "graph-view-ordering", SourceOwnerID: "network_flow_activity"}
	first, err := ProjectV2(context.Background(), invocation, leftBytes)
	if err != nil {
		t.Fatalf("project first ordering: %v", err)
	}
	second, err := ProjectV2(context.Background(), invocation, rightBytes)
	if err != nil {
		t.Fatalf("project second ordering: %v", err)
	}
	if first.ProjectionResultID != second.ProjectionResultID || first.NormalizedConfigurationSHA256 != second.NormalizedConfigurationSHA256 || first.NormalizedSourceSHA256 != second.NormalizedSourceSHA256 || first.CanonicalOutputSHA256 != second.CanonicalOutputSHA256 {
		t.Fatalf("semantic reordering changed identity:\nfirst=%#v\nsecond=%#v", first.ResultBindingV2(), second.ResultBindingV2())
	}
	if len(first.Vertices) != 2 || !sort.SliceIsSorted(first.Vertices, func(i, j int) bool {
		return first.Vertices[i].SortKey < first.Vertices[j].SortKey || first.Vertices[i].SortKey == first.Vertices[j].SortKey && first.Vertices[i].VertexID < first.Vertices[j].VertexID
	}) {
		t.Fatalf("vertices are not deterministically ordered: %#v", first.Vertices)
	}
}

func TestProjectV2CancellationAndIdentityFraming_Unit(t *testing.T) {
	fixture := graphProjectionV2GoldenFixture(t)
	input, _ := json.Marshal(fixture["input"])
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ProjectV2(ctx, InvocationContextV2{GraphViewID: "graph-view-cancel", SourceOwnerID: "network_flow_activity"}, input)
	var projectionErr *ProjectionErrorV2
	if !errors.As(err, &projectionErr) || projectionErr.Code != "projection_cancelled" || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %#v", err)
	}

	digests := strings.Repeat("a", 64)
	first, err := deriveProjectionResultIDV2(ResultBindingV2{GraphViewID: "ab", SourceOwnerID: "c", SourceSnapshotID: "snapshot", ProjectionVersion: "v", NormalizedConfigurationSHA256: digests, NormalizedSourceSHA256: digests, CanonicalOutputSHA256: digests})
	if err != nil {
		t.Fatalf("derive first identity: %v", err)
	}
	second, err := deriveProjectionResultIDV2(ResultBindingV2{GraphViewID: "a", SourceOwnerID: "bc", SourceSnapshotID: "snapshot", ProjectionVersion: "v", NormalizedConfigurationSHA256: digests, NormalizedSourceSHA256: digests, CanonicalOutputSHA256: digests})
	if err != nil {
		t.Fatalf("derive second identity: %v", err)
	}
	if first == second {
		t.Fatal("length-framed identity transcript was ambiguous")
	}

	checkpointCalls := 0
	checkpoint := newProjectionCheckpointContextV2(context.Background(), func(context.Context) error {
		checkpointCalls++
		if checkpointCalls == 2 {
			return errors.New("cancelled by owner")
		}
		return nil
	})
	if err := checkpoint.Err(); err != nil || checkpointCalls != 1 {
		t.Fatalf("initial cancellation checkpoint = (%v, %d), want (nil, 1)", err, checkpointCalls)
	}
	for range 1022 {
		if err := checkpoint.Err(); err != nil {
			t.Fatalf("unexpected cancellation before item 1024: %v", err)
		}
	}
	if err := checkpoint.Err(); !errors.Is(err, context.Canceled) || checkpointCalls != 2 {
		t.Fatalf("item 1024 cancellation checkpoint = (%v, %d), want (context canceled, 2)", err, checkpointCalls)
	}
}

func TestProjectV2MaximumSemanticBounds_Unit(t *testing.T) {
	semanticInput := maximumProjectionInputV2(t, MaximumResultVerticesV2, MaximumResultEdgesV2)
	if len(semanticInput) > graphProjectionLimits.MaxInputBytes {
		t.Fatalf("maximum semantic fixture uses %d bytes; limit is %d", len(semanticInput), graphProjectionLimits.MaxInputBytes)
	}
	result, err := ProjectV2(context.Background(), InvocationContextV2{
		GraphViewID: "graph-view-maximum-bounds", SourceOwnerID: "network_flow_activity",
	}, semanticInput)
	if err != nil {
		t.Fatalf("project maximum semantic bounds: %v", err)
	}
	if len(result.Vertices) != MaximumResultVerticesV2 || len(result.Edges) != MaximumResultEdgesV2 {
		t.Fatalf("maximum projection counts = %d/%d; want %d/%d", len(result.Vertices), len(result.Edges), MaximumResultVerticesV2, MaximumResultEdgesV2)
	}
	if !strings.HasPrefix(result.ProjectionResultID, "gpres_") || len(result.CanonicalOutputSHA256) != 64 {
		t.Fatalf("maximum projection did not produce immutable identity: %#v", result.ResultBindingV2())
	}
}

func maximumProjectionInputV2(t testing.TB, vertexCount, edgeCount int) []byte {
	t.Helper()
	var body bytes.Buffer
	body.Grow(80 << 20)
	body.WriteString(`{"projection_schema_id":"graph_projection.v2","source_snapshot_id":"snapshot-maximum-bounds","projection_config":{"projection_version":"maximum-v1","declared_source_entity_kinds":["endpoint"],"declared_source_relationship_kinds":["flow"],"entity_mappings":[{"mapping_rule_id":"map-endpoint","source_entity_kind":"endpoint","projected_vertex_kind":"endpoint","inclusion_predicate":"always","label_policy":"mapping_only","mapping_labels":[],"required_property_keys":[],"optional_property_keys":[]}],"relationship_mappings":[{"mapping_rule_id":"map-flow","source_relationship_kind":"flow","projected_edge_kind":"flow","inclusion_predicate":"always","direction_policy":"preserve","emit_reverse_edge":false,"label_policy":"mapping_only","mapping_labels":[],"required_property_keys":[],"optional_property_keys":[]}],"metadata_mappings":[],"aggregation_rules":[],"default_vertex_labels":[],"default_edge_labels":[],"allow_empty_kind_registry":false},"source_entities":[`)
	for index := range vertexCount {
		if index > 0 {
			body.WriteByte(',')
		}
		body.WriteString(`{"source_entity_id":"entity-`)
		body.WriteString(strconv.Itoa(index))
		body.WriteString(`","source_entity_kind":"endpoint","properties":{},"metadata":{},"labels":[]}`)
	}
	body.WriteString(`],"source_relationships":[`)
	for index := range edgeCount {
		if index > 0 {
			body.WriteByte(',')
		}
		body.WriteString(`{"source_relationship_id":"relationship-`)
		body.WriteString(strconv.Itoa(index))
		body.WriteString(`","source_relationship_kind":"flow","src_source_entity_id":"entity-`)
		body.WriteString(strconv.Itoa(index % vertexCount))
		body.WriteString(`","dst_source_entity_id":"entity-`)
		body.WriteString(strconv.Itoa((index + 1) % vertexCount))
		body.WriteString(`","direction":"forward","properties":{},"metadata":{},"labels":[]}`)
	}
	body.WriteString(`],"source_metadata":{},"filters":{"entity_filters":[],"relationship_filters":[],"logic":"and"},"property_definitions":[]}`)
	return body.Bytes()
}

func graphProjectionV2GoldenFixture(t testing.TB) map[string]any {
	t.Helper()
	artifact := contractgraphprojection.Index["contracts/graph-projection/v2-fixtures/empty-projection.json"]
	var fixture map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &fixture); err != nil {
		t.Fatalf("decode Graph Projection v2 fixture: %v", err)
	}
	return fixture
}

func deepCopyJSONMap(t testing.TB, input map[string]any) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal JSON copy: %v", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(encoded)))
	decoder.UseNumber()
	var output map[string]any
	if err := decoder.Decode(&output); err != nil {
		t.Fatalf("decode JSON copy: %v", err)
	}
	return output
}

func twoEntityProjectionInputV2(reverse bool) map[string]any {
	kinds := []any{"kind-a", "kind-b"}
	mappings := []any{
		map[string]any{"mapping_rule_id": "map-a", "source_entity_kind": "kind-a", "projected_vertex_kind": "vertex", "inclusion_predicate": "always", "label_policy": "mapping_only", "mapping_labels": []any{}, "required_property_keys": []any{}, "optional_property_keys": []any{}},
		map[string]any{"mapping_rule_id": "map-b", "source_entity_kind": "kind-b", "projected_vertex_kind": "vertex", "inclusion_predicate": "always", "label_policy": "mapping_only", "mapping_labels": []any{}, "required_property_keys": []any{}, "optional_property_keys": []any{}},
	}
	entities := []any{
		map[string]any{"source_entity_id": "entity-a", "source_entity_kind": "kind-a", "properties": map[string]any{}, "metadata": map[string]any{"number": json.Number("1.0")}, "labels": []any{}},
		map[string]any{"source_entity_id": "entity-b", "source_entity_kind": "kind-b", "properties": map[string]any{}, "metadata": map[string]any{"number": json.Number("1e0")}, "labels": []any{}},
	}
	if reverse {
		kinds[0], kinds[1] = kinds[1], kinds[0]
		mappings[0], mappings[1] = mappings[1], mappings[0]
		entities[0], entities[1] = entities[1], entities[0]
	}
	return map[string]any{
		"projection_schema_id": ProjectionSchemaIDV2,
		"source_snapshot_id":   "snapshot-ordering",
		"projection_config": map[string]any{
			"projection_version": "test-v1", "declared_source_entity_kinds": kinds, "declared_source_relationship_kinds": []any{},
			"entity_mappings": mappings, "relationship_mappings": []any{}, "metadata_mappings": []any{}, "aggregation_rules": []any{},
			"default_vertex_labels": []any{}, "default_edge_labels": []any{}, "allow_empty_kind_registry": false,
		},
		"source_entities": entities, "source_relationships": []any{}, "source_metadata": map[string]any{},
		"filters": map[string]any{"entity_filters": []any{}, "relationship_filters": []any{}, "logic": "and"}, "property_definitions": []any{},
	}
}
