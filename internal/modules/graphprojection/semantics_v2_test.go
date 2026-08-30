package graphprojection

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	contractgraphprojection "github.com/JochiRaider/cartulary/internal/gen/contractgraphprojection"
)

func TestProjectV2SemanticMatrixMatchesContractAndExecutesEveryPair_Unit(t *testing.T) {
	t.Parallel()

	registry := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/semantic-registry.v1.json")
	contractMatrix := registry["merge_matrix"].(map[string]any)
	if len(contractMatrix) != len(projectedMergeMatrixV2) {
		t.Fatalf("contract/runtime projected type counts = %d/%d", len(contractMatrix), len(projectedMergeMatrixV2))
	}
	allBehaviors := map[string]struct{}{"unsupported": {}}
	for projectedType, rawBehaviors := range contractMatrix {
		behaviors := rawBehaviors.([]any)
		runtimeBehaviors, ok := projectedMergeMatrixV2[projectedType]
		if !ok || len(runtimeBehaviors) != len(behaviors) {
			t.Fatalf("contract/runtime behaviors for %s = %#v/%#v", projectedType, behaviors, runtimeBehaviors)
		}
		for _, rawBehavior := range behaviors {
			behavior := rawBehavior.(string)
			allBehaviors[behavior] = struct{}{}
			if !validProjectedMergeV2(projectedType, behavior) {
				t.Fatalf("owner-valid pair rejected: %s/%s", projectedType, behavior)
			}
			values := semanticMergeCandidatesV2(projectedType, behavior)
			if _, ok, conflict := mergeProjectedValuesV2(projectedType, behavior, values); !ok || conflict {
				t.Fatalf("owner-valid pair failed execution: %s/%s ok=%v conflict=%v values=%#v", projectedType, behavior, ok, conflict, values)
			}
		}
	}
	for projectedType := range projectedMergeMatrixV2 {
		for behavior := range allBehaviors {
			_, expected := projectedMergeMatrixV2[projectedType][behavior]
			if validProjectedMergeV2(projectedType, behavior) != expected {
				t.Fatalf("pair validity drifted for %s/%s", projectedType, behavior)
			}
			if !expected {
				if _, ok, conflict := mergeProjectedValuesV2(projectedType, behavior, []any{int64(1)}); ok || conflict {
					t.Fatalf("invalid pair executed: %s/%s ok=%v conflict=%v", projectedType, behavior, ok, conflict)
				}
			}
		}
	}
	if validProjectedType("legacy_scalar") || validProjectedMergeV2("legacy_scalar", "single_value") {
		t.Fatal("unknown projected type was accepted")
	}
}

func TestProjectV2NonEmptySemanticGolden_Unit(t *testing.T) {
	t.Parallel()

	fixture := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/v2-fixtures/semantic-projection.json")
	trusted := fixture["trusted_context"].(map[string]any)
	input, err := json.Marshal(fixture["input"])
	if err != nil {
		t.Fatalf("marshal semantic golden input: %v", err)
	}
	result, err := ProjectV2(context.Background(), InvocationContextV2{
		GraphViewID: trusted["graph_view_id"].(string), SourceOwnerID: trusted["source_owner_id"].(string),
	}, input)
	if err != nil {
		t.Fatalf("project semantic golden input: %v", err)
	}
	expected := fixture["expected"].(map[string]any)
	if len(result.Vertices) != int(expected["vertex_count"].(float64)) || len(result.Edges) != int(expected["edge_count"].(float64)) {
		t.Fatalf("semantic golden counts = %d/%d, want %#v", len(result.Vertices), len(result.Edges), expected)
	}
	identities := map[string]string{
		"projection_result_id":            result.ProjectionResultID,
		"normalized_configuration_sha256": result.NormalizedConfigurationSHA256,
		"normalized_source_sha256":        result.NormalizedSourceSHA256,
		"canonical_output_sha256":         result.CanonicalOutputSHA256,
	}
	for key, got := range identities {
		if got != expected[key] {
			t.Fatalf("semantic golden %s = %q, want %q; all identities=%#v", key, got, expected[key], identities)
		}
	}
	for _, vertex := range result.Vertices {
		if vertex.VertexFamily == "aggregated" {
			if canonicalValueKey(vertex.Properties) != canonicalValueKey(expected["aggregated_properties"]) {
				t.Fatalf("semantic golden aggregate properties = %#v, want %#v", vertex.Properties, expected["aggregated_properties"])
			}
			return
		}
	}
	t.Fatal("semantic golden projection did not emit its aggregated vertex")
}

func TestProjectV2SemanticMergeRulesAreExact_Unit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		projectedType string
		behavior      string
		values        []any
		want          any
		wantOK        bool
		wantConflict  bool
	}{
		{name: "single equal canonical values", projectedType: "number", behavior: "single_value", values: []any{int64(1), float64(1)}, want: int64(1), wantOK: true},
		{name: "single conflict", projectedType: "string", behavior: "single_value", values: []any{"a", "b"}, wantConflict: true},
		{name: "first", projectedType: "boolean", behavior: "first", values: []any{true, false}, want: true, wantOK: true},
		{name: "last", projectedType: "boolean", behavior: "last", values: []any{true, false}, want: false, wantOK: true},
		{name: "integer minimum", projectedType: "integer", behavior: "min", values: []any{int64(9), int64(-2), int64(4)}, want: int64(-2), wantOK: true},
		{name: "integer maximum", projectedType: "integer", behavior: "max", values: []any{int64(9), int64(-2), int64(4)}, want: int64(9), wantOK: true},
		{name: "number exact ordering above safe integer", projectedType: "number", behavior: "min", values: []any{int64(math.MaxInt64), int64(math.MaxInt64 - 1)}, want: int64(math.MaxInt64 - 1), wantOK: true},
		{name: "timestamp order", projectedType: "timestamp", behavior: "max", values: []any{"2026-01-01T00:00:00Z", "2026-01-01T00:00:00.000001Z"}, want: "2026-01-01T00:00:00.000001Z", wantOK: true},
		{name: "identifier order", projectedType: "identifier", behavior: "min", values: []any{"z", "a"}, want: "a", wantOK: true},
		{name: "integer sum", projectedType: "integer", behavior: "sum", values: []any{int64(-2), int64(5)}, want: int64(3), wantOK: true},
		{name: "integer sum overflow", projectedType: "integer", behavior: "sum", values: []any{int64(math.MaxInt64), int64(1)}, wantConflict: true},
		{name: "integer sum underflow", projectedType: "integer", behavior: "sum", values: []any{int64(math.MinInt64), int64(-1)}, wantConflict: true},
		{name: "integer count", projectedType: "integer", behavior: "count", values: []any{int64(9), int64(9), int64(-2)}, want: int64(3), wantOK: true},
		{name: "number sum", projectedType: "number", behavior: "sum", values: []any{float64(1.25), int64(2)}, want: float64(3.25), wantOK: true},
		{name: "number sum non finite", projectedType: "number", behavior: "sum", values: []any{math.MaxFloat64, math.MaxFloat64}, wantConflict: true},
		{name: "set flattens deduplicates and canonical sorts", projectedType: "string_array", behavior: "set", values: []any{[]any{"z", "a"}, []any{"a", "m"}}, want: []any{"a", "m", "z"}, wantOK: true},
		{name: "ordered list preserves order and duplicates", projectedType: "integer_array", behavior: "ordered_list", values: []any{[]any{int64(2), int64(1)}, []any{int64(1)}}, want: []any{int64(2), int64(1), int64(1)}, wantOK: true},
		{name: "empty candidates", projectedType: "integer", behavior: "sum", values: nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok, conflict := mergeProjectedValuesV2(test.projectedType, test.behavior, test.values)
			if ok != test.wantOK || conflict != test.wantConflict || !reflect.DeepEqual(got, test.want) {
				t.Fatalf("merge = (%#v, %v, %v), want (%#v, %v, %v)", got, ok, conflict, test.want, test.wantOK, test.wantConflict)
			}
		})
	}
}

func TestProjectV2ProjectedValueAdmissionCoversEveryTypeAndBoundary_Unit(t *testing.T) {
	t.Parallel()

	valid := map[string]any{
		"boolean": true, "integer": json.Number("-9223372036854775808"), "number": json.Number("1.25e2"),
		"string": "value", "timestamp": "2026-08-29T12:34:56.123456Z", "identifier": "identifier",
		"boolean_array": []any{true, false}, "integer_array": []any{json.Number("9223372036854775807")},
		"number_array": []any{json.Number("1.25")}, "string_array": []any{"value"},
		"timestamp_array": []any{"2026-08-29T12:34:56Z"}, "identifier_array": []any{"identifier"},
	}
	for projectedType, input := range valid {
		if _, ok := normalizeProjectedValueV2(projectedType, input); !ok {
			t.Fatalf("valid %s value rejected: %#v", projectedType, input)
		}
	}
	invalid := []struct {
		projectedType string
		value         any
	}{
		{projectedType: "integer", value: json.Number("9223372036854775808")},
		{projectedType: "integer", value: json.Number("1.0")},
		{projectedType: "number", value: json.Number("1e309")},
		{projectedType: "number", value: math.Inf(1)},
		{projectedType: "timestamp", value: "2026-08-29T12:34:56+01:00"},
		{projectedType: "identifier", value: " spaced "},
		{projectedType: "string_array", value: []any{"value", nil}},
		{projectedType: "integer_array", value: []any{[]any{json.Number("1")}}},
		{projectedType: "unknown", value: "value"},
	}
	for _, test := range invalid {
		if _, ok := normalizeProjectedValueV2(test.projectedType, test.value); ok {
			t.Fatalf("invalid %s value accepted: %#v", test.projectedType, test.value)
		}
	}
	oneOver := make([]any, graphProjectionLimits.MaxArrayItems+1)
	for index := range oneOver {
		oneOver[index] = "value"
	}
	if _, ok := normalizeProjectedValueV2("string_array", oneOver); ok {
		t.Fatal("one-over flat array limit was accepted")
	}
}

func TestProjectV2NullDefaultWildcardOrderingAndFailureAtomicity_Unit(t *testing.T) {
	t.Parallel()

	fixture := decodeGraphProjectionContractArtifact(t, "contracts/graph-projection/v2-fixtures/semantic-projection.json")
	base := fixture["input"].(map[string]any)
	project := func(t *testing.T, value any, present bool, definition map[string]any) (ProjectionResultV2, error) {
		t.Helper()
		input := deepCopyJSONMap(t, base)
		config := input["projection_config"].(map[string]any)
		config["aggregation_rules"] = []any{}
		entities := input["source_entities"].([]any)[:1]
		properties := entities[0].(map[string]any)["properties"].(map[string]any)
		if present {
			properties["value"] = value
		} else {
			delete(properties, "value")
		}
		input["source_entities"] = entities
		input["property_definitions"] = []any{definition}
		encoded, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("marshal null/default input: %v", err)
		}
		return ProjectV2(context.Background(), InvocationContextV2{GraphViewID: "semantic-null-default", SourceOwnerID: "network_flow_activity"}, encoded)
	}
	definition := func(missing, sourceNull, nullOutput string, defaultValue any, hasDefault bool) map[string]any {
		value := map[string]any{
			"property_definition_id": "value-definition", "target_scope": "vertex", "target_kind": "*",
			"source_field_path": "properties.value", "projected_key": "projected_value", "projected_type": "string",
			"required": false, "missing_behavior": missing, "source_null_behavior": sourceNull,
			"null_output_policy": nullOutput, "merge_behavior": "single_value",
		}
		if hasDefault {
			value["default_value"] = defaultValue
		}
		return value
	}
	tests := []struct {
		name        string
		value       any
		present     bool
		definition  map[string]any
		wantPresent bool
		want        any
		wantError   bool
	}{
		{name: "wildcard present", value: "source", present: true, definition: definition("omit", "omit", "omit", nil, false), wantPresent: true, want: "source"},
		{name: "missing default", present: false, definition: definition("default", "omit", "omit", "fallback", true), wantPresent: true, want: "fallback"},
		{name: "null default", present: true, value: nil, definition: definition("omit", "default", "omit", "fallback", true), wantPresent: true, want: "fallback"},
		{name: "missing omit", present: false, definition: definition("omit", "omit", "omit", nil, false)},
		{name: "null emit", present: true, value: nil, definition: definition("omit", "emit_null", "emit_null", nil, false), wantPresent: true, want: nil},
		{name: "null error", present: true, value: nil, definition: definition("omit", "error", "omit", nil, false), wantError: true},
		{name: "invalid default type", present: false, definition: definition("default", "omit", "omit", int64(7), true), wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := project(t, test.value, test.present, test.definition)
			if test.wantError {
				var projectionErr *ProjectionErrorV2
				if !errors.As(err, &projectionErr) || len(result.Vertices) != 0 {
					t.Fatalf("failed projection = (%#v, %#v), want closed error and no partial result", result, err)
				}
				return
			}
			if err != nil || len(result.Vertices) != 1 {
				t.Fatalf("projection = (%#v, %#v), want one vertex", result, err)
			}
			got, present := result.Vertices[0].Properties["projected_value"]
			if present != test.wantPresent || !reflect.DeepEqual(got, test.want) {
				t.Fatalf("projected wildcard value = (%#v, %v), want (%#v, %v)", got, present, test.want, test.wantPresent)
			}
		})
	}

	invalidPair := definition("omit", "omit", "omit", nil, false)
	invalidPair["projected_type"] = "boolean"
	invalidPair["merge_behavior"] = "sum"
	if result, err := project(t, true, true, invalidPair); err == nil || len(result.Vertices) != 0 {
		t.Fatalf("invalid type/merge pair produced partial output: (%#v, %v)", result, err)
	}

	reordered := deepCopyJSONMap(t, base)
	entities := reordered["source_entities"].([]any)
	entities[0], entities[1] = entities[1], entities[0]
	reorderedBytes, err := json.Marshal(reordered)
	if err != nil {
		t.Fatalf("marshal reordered semantic input: %v", err)
	}
	trusted := fixture["trusted_context"].(map[string]any)
	result, err := ProjectV2(context.Background(), InvocationContextV2{GraphViewID: trusted["graph_view_id"].(string), SourceOwnerID: trusted["source_owner_id"].(string)}, reorderedBytes)
	if err != nil {
		t.Fatalf("project reordered contributors: %v", err)
	}
	expected := fixture["expected"].(map[string]any)
	if result.ProjectionResultID != expected["projection_result_id"] || result.CanonicalOutputSHA256 != expected["canonical_output_sha256"] {
		t.Fatalf("reordered contributors changed semantic identity: %#v", result.ResultBindingV2())
	}
}

func TestProjectV2UTF8ByteAndNumericBoundariesAreUniform_Unit(t *testing.T) {
	t.Parallel()

	base := graphProjectionV2GoldenFixture(t)["input"].(map[string]any)
	project := func(t *testing.T, mutate func(map[string]any)) error {
		t.Helper()
		input := deepCopyJSONMap(t, base)
		mutate(input)
		encoded, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("marshal boundary input: %v", err)
		}
		_, err = ProjectV2(context.Background(), InvocationContextV2{GraphViewID: "boundary-view", SourceOwnerID: "network_flow_activity"}, encoded)
		return err
	}
	assertAccepted := func(t *testing.T, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("boundary value rejected: %v", err)
		}
	}
	assertRejected := func(t *testing.T, err error) {
		t.Helper()
		var projectionErr *ProjectionErrorV2
		if !errors.As(err, &projectionErr) || projectionErr.RetryAction != "do_not_retry" {
			t.Fatalf("boundary error = %#v, want closed rejection", err)
		}
	}

	t.Run("identifier bytes", func(t *testing.T) {
		assertAccepted(t, projectV2IdentifierBoundaryV2(strings.Repeat("é", 127)+"a"))
		assertRejected(t, projectV2IdentifierBoundaryV2(strings.Repeat("é", 128)))
	})
	t.Run("label bytes", func(t *testing.T) {
		assertAccepted(t, project(t, func(input map[string]any) {
			input["projection_config"].(map[string]any)["default_vertex_labels"] = []any{strings.Repeat("é", 128)}
		}))
		assertRejected(t, project(t, func(input map[string]any) {
			input["projection_config"].(map[string]any)["default_vertex_labels"] = []any{strings.Repeat("é", 128) + "a"}
		}))
	})
	t.Run("string bytes", func(t *testing.T) {
		assertAccepted(t, project(t, func(input map[string]any) {
			input["source_metadata"] = map[string]any{"text": strings.Repeat("é", graphProjectionLimits.MaxStringBytes/2)}
		}))
		assertRejected(t, project(t, func(input map[string]any) {
			input["source_metadata"] = map[string]any{"text": strings.Repeat("é", graphProjectionLimits.MaxStringBytes/2) + "a"}
		}))
	})
	t.Run("property key bytes", func(t *testing.T) {
		assertAccepted(t, project(t, func(input map[string]any) {
			input["source_metadata"] = map[string]any{strings.Repeat("é", 127) + "a": true}
		}))
		assertRejected(t, project(t, func(input map[string]any) {
			input["source_metadata"] = map[string]any{strings.Repeat("é", 128): true}
		}))
	})
	t.Run("integer and number lexemes", func(t *testing.T) {
		for _, lexeme := range []string{"-9223372036854775808", "9223372036854775807", "1.7976931348623157e308"} {
			lexeme := lexeme
			assertAccepted(t, project(t, func(input map[string]any) {
				input["source_metadata"] = map[string]any{"value": json.Number(lexeme)}
			}))
		}
		for _, lexeme := range []string{"-9223372036854775809", "9223372036854775808", "1e309"} {
			lexeme := lexeme
			assertRejected(t, project(t, func(input map[string]any) {
				input["source_metadata"] = map[string]any{"value": json.Number(lexeme)}
			}))
		}
	})
	t.Run("invalid utf8", func(t *testing.T) {
		invalid := append([]byte(`{"projection_schema_id":"graph_projection.v2","source_snapshot_id":"`), 0xff)
		invalid = append(invalid, []byte(`"}`)...)
		assertRejected(t, func() error {
			_, err := ProjectV2(context.Background(), InvocationContextV2{GraphViewID: "boundary-view", SourceOwnerID: "network_flow_activity"}, invalid)
			return err
		}())
	})
}

func projectV2IdentifierBoundaryV2(graphViewID string) error {
	fixture := contractgraphprojection.Index["contracts/graph-projection/v2-fixtures/empty-projection.json"]
	var decoded map[string]any
	decoder := json.NewDecoder(strings.NewReader(fixture.JSON))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	input, err := json.Marshal(decoded["input"])
	if err != nil {
		return err
	}
	_, err = ProjectV2(context.Background(), InvocationContextV2{GraphViewID: graphViewID, SourceOwnerID: "network_flow_activity"}, input)
	return err
}

func semanticMergeCandidatesV2(projectedType, behavior string) []any {
	scalar := map[string][]any{
		"boolean": {true, false}, "integer": {int64(1), int64(2)}, "number": {float64(1.5), int64(2)},
		"string": {"a", "b"}, "timestamp": {"2026-01-01T00:00:00Z", "2026-01-02T00:00:00Z"}, "identifier": {"a", "b"},
	}
	if strings.HasSuffix(projectedType, "_array") {
		values := scalar[strings.TrimSuffix(projectedType, "_array")]
		if behavior == "single_value" {
			return []any{[]any{values[0]}, []any{values[0]}}
		}
		return []any{[]any{values[0], values[1]}, []any{values[0]}}
	}
	values := scalar[projectedType]
	if behavior == "single_value" {
		return []any{values[0], values[0]}
	}
	return values
}
