package graphprojection

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"testing"
)

func repeatedInputItemsV2(item any, count int) []any {
	items := make([]any, count)
	for i := range items {
		items[i] = item
	}
	return items
}

func expectInputFailureV2(t *testing.T, body []byte, code, reason string, details map[string]any) {
	t.Helper()
	result, err := ProjectV2(t.Context(), InvocationContextV2{GraphViewID: "admission", SourceOwnerID: "owner"}, body)
	var failure *ProjectionErrorV2
	if !errors.As(err, &failure) || failure.Code != code || failure.ReasonCode != reason ||
		failure.RetryAction != "do_not_retry" || !reflect.DeepEqual(failure.Details, details) || !reflect.DeepEqual(result, ProjectionResultV2{}) {
		t.Fatalf("want %s/%s %v and zero result, got %#v, result ID %q", code, reason, details, failure, result.ProjectionResultID)
	}
}

func expectCollectionOverflowV2(t *testing.T, body []byte, key string, limit int) {
	t.Helper()
	expectInputFailureV2(t, body, "projection_resource_limit_exceeded", key, map[string]any{"limit_key": key, "limit": limit, "observed": limit + 1})
}

func inputJSONV2(t *testing.T, input map[string]any) []byte {
	t.Helper()
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func filterItemV2() map[string]any {
	return map[string]any{"field_path": "properties.group", "operator": "exists", "include_if_missing": false}
}

func checkSourceIDPrecedenceV2(t *testing.T) {
	for _, collection := range []string{"source_entities", "source_relationships"} {
		for _, overflow := range []bool{false, true} {
			input := fieldPathInputV2()
			rows := input[collection].([]any)
			input[collection] = append(rows, rows[0])
			if overflow {
				limit := graphProjectionLimits.MaxEntityFilters
				input["filters"].(map[string]any)["entity_filters"] = repeatedInputItemsV2(filterItemV2(), limit+1)
			}
			expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "duplicate_identifier", map[string]any{"issue_count": 1})
		}
	}
	input := fieldPathInputV2()
	entity := input["source_entities"].([]any)[0].(map[string]any)
	relation := input["source_relationships"].([]any)[0].(map[string]any)
	relation["source_relationship_id"] = entity["source_entity_id"]
	if _, _, err := projectAggregationInputV2(t, input); err != nil {
		t.Fatalf("matching IDs in different source collections: %v", err)
	}
	entity["source_entity_kind"] = "/bad"
	input["source_relationships"] = []any{relation, relation}
	input["filters"].(map[string]any)["entity_filters"] = repeatedInputItemsV2(filterItemV2(), graphProjectionLimits.MaxEntityFilters+1)
	expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "invalid_input_shape", map[string]any{"issue_count": 1})
}

func checkInputCollectionLimitsV2(t *testing.T) {
	limits := graphProjectionLimits
	t.Run("source IDs share the issue budget", func(t *testing.T) {
		input := fieldPathInputV2()
		n := limits.MaxValidationIssues/2 + 1
		input["source_entities"] = repeatedInputItemsV2(input["source_entities"].([]any)[0], n)
		rows := repeatedInputItemsV2(input["source_relationships"].([]any)[0], n+1)
		input["source_relationships"] = rows[:n]
		expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "duplicate_identifier", map[string]any{"issue_count": limits.MaxValidationIssues})
		input["source_relationships"] = rows
		raw := inputJSONV2(t, input)
		expectCollectionOverflowV2(t, raw, "maximum_validation_issues", limits.MaxValidationIssues)
		calls := 0
		result, err := ProjectV2(t.Context(), InvocationContextV2{GraphViewID: "admission", SourceOwnerID: "owner", CancellationCheck: func(context.Context) error {
			calls++
			if calls == 2 {
				return context.Canceled
			}
			return nil
		}}, raw)
		var failure *ProjectionErrorV2
		if !errors.As(err, &failure) || failure.Code != "projection_cancelled" || failure.ReasonCode != "context_cancelled" ||
			failure.RetryAction != "do_not_retry" || calls != 2 || !reflect.DeepEqual(result, ProjectionResultV2{}) {
			t.Fatalf("source admission cancellation: %#v, %d callbacks", failure, calls)
		}
		// The same full budget is terminal on the first repeated metadata ID.
		input["source_relationships"] = rows[:n]
		for _, key := range []string{"aa", "bb"} {
			appendDefinitionV2(input, aggregationPropertyV2("observed", "edge", "edge", "source_relationship_kind", key, "string", "single_value"), true)
		}
		expectCollectionOverflowV2(t, inputJSONV2(t, input), "maximum_validation_issues", limits.MaxValidationIssues)

	})
	// The parent already projects both source collections at their real ceilings.
	expectCollectionOverflowV2(t, maximumProjectionInputV2(t, limits.MaxSourceEntities+1, 0), "maximum_source_entities", limits.MaxSourceEntities)
	expectCollectionOverflowV2(t, maximumProjectionInputV2(t, 2, limits.MaxSourceRelationships+1), "maximum_source_relationships", limits.MaxSourceRelationships)
	for _, tc := range []struct {
		prefix          string
		limit           int
		vertices, edges int
	}{
		{"entity-", limits.MaxSourceEntities, limits.MaxSourceEntities + 1, 0},
		{"relationship-", limits.MaxSourceRelationships, 2, limits.MaxSourceRelationships + 1},
	} {
		body := maximumProjectionInputV2(t, tc.vertices, tc.edges)
		last := []byte(`"` + tc.prefix + strconv.Itoa(tc.limit) + `"`)
		if bytes.Count(body, last) != 1 {
			t.Fatal("missing unique final source ID")
		}
		body = bytes.Replace(body, last, []byte(`"`+tc.prefix+`0"`), 1)
		expectInputFailureV2(t, body, "projection_validation_failed", "duplicate_identifier", map[string]any{"issue_count": 1})
	}
	for _, tc := range []struct {
		group, key string
		limit      int
	}{
		{"", "property_definitions", limits.MaxPropertyDefinitions},
		{"projection_config", "entity_mappings", limits.MaxEntityMappings},
		{"projection_config", "relationship_mappings", limits.MaxRelationshipMappings},
		{"projection_config", "metadata_mappings", limits.MaxMetadataMappings},
		{"projection_config", "aggregation_rules", limits.MaxAggregationRules},
		{"filters", "entity_filters", limits.MaxEntityFilters},
		{"filters", "relationship_filters", limits.MaxRelationshipFilters},
	} {
		t.Run(tc.key, func(t *testing.T) {
			input := fieldPathInputV2()
			config := input["projection_config"].(map[string]any)
			// A later, deterministic rule failure witnesses admission at capacity.
			rule := aggregationRuleInputV2("duplicate", "vertex", "source_entity", "kind-a", "group")
			config["aggregation_rules"] = []any{rule, rule}
			for _, metadata := range []bool{true, false} {
				appendDefinitionV2(input, aggregationPropertyV2("property", "vertex", "vertex", "properties.group", "group", "string", "single_value"), metadata)
			}
			filters := input["filters"].(map[string]any)
			filters["entity_filters"], filters["relationship_filters"] = []any{filterItemV2()}, []any{filterItemV2()}
			group := input
			if tc.group != "" {
				group = input[tc.group].(map[string]any)
			}
			rows := repeatedInputItemsV2(group[tc.key].([]any)[0], tc.limit+1)
			idKey := map[string]string{"property_definitions": "property_definition_id", "metadata_mappings": "metadata_mapping_id"}[tc.key]
			if idKey != "" {
				for i, raw := range rows {
					def := maps.Clone(raw.(map[string]any))
					def[idKey] = "field-" + strconv.Itoa(i)
					rows[i] = def
				}
			}
			group[tc.key] = rows[:tc.limit]
			expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "invalid_aggregation_rule", map[string]any{"issue_count": 1})
			group[tc.key] = rows
			expectCollectionOverflowV2(t, inputJSONV2(t, input), "maximum_"+tc.key, tc.limit)
		})
	}
	input := fieldPathInputV2()
	filters := input["filters"].(map[string]any)
	filters["entity_filters"] = repeatedInputItemsV2(filterItemV2(), limits.MaxEntityFilters+1)
	filters["relationship_filters"] = repeatedInputItemsV2(filterItemV2(), limits.MaxRelationshipFilters+1)
	expectCollectionOverflowV2(t, inputJSONV2(t, input), "maximum_entity_filters", limits.MaxEntityFilters)
	input["source_metadata"] = false
	expectInputFailureV2(t, inputJSONV2(t, input), "invalid_projection_request", "schema_type_mismatch", map[string]any{"field": "$.source_metadata"})
}

func checkSourceScalarPrecedenceV2(t *testing.T) {
	const code = "projection_validation_failed"
	for _, collection := range []string{"", "source_entities", "source_relationships"} {
		for _, overflow := range []bool{false, true} {
			input := fieldPathInputV2()
			rel := input["source_relationships"].([]any)[0].(map[string]any)
			if collection != "" {
				rows := input[collection].([]any)
				input[collection] = append(rows, deepCopyJSONMap(t, rows[0].(map[string]any)))
			}
			if overflow {
				input["filters"].(map[string]any)["entity_filters"] = repeatedInputItemsV2(filterItemV2(), graphProjectionLimits.MaxEntityFilters+1)
			}
			rel["direction"] = "sideways"
			for range 2 {
				expectInputFailureV2(t, inputJSONV2(t, input), code, "invalid_relationship_direction", map[string]any{"issue_count": 1})
				slices.Reverse(input["source_entities"].([]any))
				slices.Reverse(input["source_relationships"].([]any))
			}
		}
	}
	input := fieldPathInputV2()
	rel := input["source_relationships"].([]any)[0].(map[string]any)
	for _, direction := range []string{"forward", "reverse", "undirected", "bidirectional"} {
		rel["direction"] = direction
		if _, _, err := projectAggregationInputV2(t, input); err != nil {
			t.Fatal(direction, err)
		}
	}
	rel["direction"] = "sideways"
	rel["labels"] = []any{""}
	rel["dst_source_entity_id"] = ""
	expectInputFailureV2(t, inputJSONV2(t, input), code, "invalid_input_shape", map[string]any{"issue_count": 3})
	rel["labels"] = []any{}
	rel["direction"] = "forward"
	rel["dst_source_entity_id"] = "/bad"
	expectInputFailureV2(t, inputJSONV2(t, input), code, "invalid_input_shape", map[string]any{"issue_count": 1})
}

func checkSubmittedEndpointsV2(t *testing.T) {
	for _, mode := range []string{"mapped", "predicate excluded", "filtered"} {
		t.Run(mode, func(t *testing.T) {
			input := fieldPathInputV2()
			rel := input["source_relationships"].([]any)[0].(map[string]any)
			switch mode {
			case "predicate excluded":
				input["projection_config"].(map[string]any)["relationship_mappings"].([]any)[0].(map[string]any)["inclusion_predicate"] = map[string]any{"operator": "exists", "field_path": "properties.absent", "include_if_missing": false}
			case "filtered":
				input["filters"].(map[string]any)["relationship_filters"] = []any{map[string]any{"operator": "exists", "field_path": "properties.absent", "include_if_missing": false}}
			}
			if _, _, err := projectAggregationInputV2(t, input); err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"src_source_entity_id", "dst_source_entity_id"} {
				original := rel[field]
				rel[field] = "absent"
				expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "missing_relationship_endpoint", map[string]any{"issue_count": 1})
				rel[field] = original
			}
		})
	}
	input := fieldPathInputV2()
	rel := input["source_relationships"].([]any)[0].(map[string]any)
	rel["src_source_entity_id"] = "absent"
	input["projection_config"].(map[string]any)["relationship_mappings"] = []any{}
	expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "missing_relationship_mapping_rule", map[string]any{"issue_count": 1})
	input = fieldPathInputV2()
	input["filters"].(map[string]any)["entity_filters"] = []any{map[string]any{"operator": "equals", "field_path": "source_entity_id", "value": "entity-a", "include_if_missing": false}}
	expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "relationship_endpoint_not_projected", map[string]any{"issue_count": 1})
}

func checkDefinitionIDsV2(t *testing.T) {
	input := fieldPathInputV2()
	a := aggregationPropertyV2("one", "vertex", "vertex", "properties.group", "a", "string", "single_value")
	b := maps.Clone(a)
	b["projected_key"] = "b"
	defs := []any{a, b}
	input["property_definitions"] = defs
	for range 2 {
		expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "duplicate_identifier", map[string]any{"issue_count": 1})
		slices.Reverse(defs)
	}
	filters := input["filters"].(map[string]any)
	filters["entity_filters"] = repeatedInputItemsV2(filterItemV2(), graphProjectionLimits.MaxEntityFilters+1)
	expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "duplicate_identifier", map[string]any{"issue_count": 1})
	filters["entity_filters"] = []any{}
	b["property_definition_id"] = "two"
	before, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	slices.Reverse(defs)
	after, _, err := projectAggregationInputV2(t, input)
	if err != nil || before.ResultBindingV2() != after.ResultBindingV2() {
		t.Fatalf("permuted definitions: %v", err)
	}
	b["projected_key"], b["target_kind"] = "a", "*"
	expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "invalid_property_definition", map[string]any{"issue_count": 1})
}

func checkTerminalScalarBudgetV2(t *testing.T) {
	limit := graphProjectionLimits.MaxValidationIssues
	input := fieldPathInputV2()
	config := input["projection_config"].(map[string]any)
	labels := make([]string, limit+1)
	config["default_vertex_labels"] = labels[:limit]
	expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "invalid_input_shape", map[string]any{"issue_count": limit})
	input["source_entities"].([]any)[0].(map[string]any)["properties"] = map[string]any{"*": "value"}
	expectInputFailureV2(t, inputJSONV2(t, input), "invalid_projection_request", "invalid_property_key", map[string]any{"field": "$.source_entities[0].properties[\"*\"]"})
	config["default_vertex_labels"] = labels
	expectCollectionOverflowV2(t, inputJSONV2(t, input), "maximum_validation_issues", limit)
	config["default_vertex_labels"] = labels[:limit]
	config["default_edge_labels"] = []string{""}
	expectCollectionOverflowV2(t, inputJSONV2(t, input), "maximum_validation_issues", limit)
	config["projection_version"] = "/bad"
	expectInputFailureV2(t, inputJSONV2(t, input), "invalid_projection_request", "invalid_identifier", map[string]any{"field": "$.projection_config.projection_version"})
}

func checkScalarBudgetV2(t *testing.T) {
	input := fieldPathInputV2()
	rel := input["source_relationships"].([]any)[0].(map[string]any)
	rel["direction"] = "sideways"
	rel["src_source_entity_id"] = "/invalid"
	rows := repeatedInputItemsV2(rel, graphProjectionLimits.MaxValidationIssues/2)
	input["source_relationships"] = rows
	expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "invalid_input_shape", map[string]any{"issue_count": graphProjectionLimits.MaxValidationIssues})
	extra := deepCopyJSONMap(t, rel)
	extra["src_source_entity_id"] = "entity-a"
	input["source_relationships"] = append(rows, extra)
	expectCollectionOverflowV2(t, inputJSONV2(t, input), "maximum_validation_issues", graphProjectionLimits.MaxValidationIssues)
}

// Cancellation at the first uniqueness checkpoint must precede any sorting of
// rejected input. The public error alone cannot distinguish a late check.
func checkUniquenessCancellationV2(t *testing.T) {
	count := graphProjectionLimits.MaxSourceEntities + 1
	var input map[string]any
	if err := json.Unmarshal(maximumProjectionInputV2(t, count, 0), &input); err != nil {
		t.Fatal(err)
	}
	rows := input["source_entities"].([]any)
	slices.Reverse(rows)
	body := inputJSONV2(t, input)
	firstID := rows[0].(map[string]any)["source_entity_id"]
	// Each admitted row crosses one shape and one scalar checkpoint.
	config := input["projection_config"].(map[string]any)
	perPass := count + len(config["entity_mappings"].([]any)) + len(config["relationship_mappings"].([]any))
	afterScalars := 1 + 2*perPass/graphProjectionLimits.CancellationCheckIntervalItems
	calls := 0
	result, err := ProjectV2(t.Context(), InvocationContextV2{GraphViewID: "admission", SourceOwnerID: "owner", CancellationCheck: func(context.Context) error {
		calls++
		if calls > afterScalars {
			return context.Canceled
		}
		return nil
	}}, body)
	var failure *ProjectionErrorV2
	if !errors.As(err, &failure) || !errors.Is(err, context.Canceled) || failure.Code != "projection_cancelled" || failure.ReasonCode != "context_cancelled" || failure.RetryAction != "do_not_retry" || !reflect.DeepEqual(result, ProjectionResultV2{}) || calls != afterScalars+1 {
		t.Fatalf("public cancellation during uniqueness: %#v (%d checkpoints)", failure, calls)
	}

	checkpoint := newProjectionCheckpointContextV2(t.Context(), func(context.Context) error { return context.Canceled })
	defer checkpoint.cancel()
	if err := admitUniqueIdentifiersV2(checkpoint, input); !errors.Is(err, context.Canceled) {
		t.Fatalf("uniqueness cancellation: %v", err)
	}
	if rows[0].(map[string]any)["source_entity_id"] != firstID {
		t.Fatal("oversized collection was sorted before cancellation")
	}
}

func checkRequestAdmissionV2(t *testing.T) {
	t.Run("shape before configuration scalar", func(t *testing.T) {
		input := fieldPathInputV2()
		input["projection_config"].(map[string]any)["projection_version"] = "/bad"
		input["source_relationships"].([]any)[0].(map[string]any)["direction"] = 42
		expectInputFailureV2(t, inputJSONV2(t, input), "invalid_projection_request", "schema_type_mismatch", map[string]any{"field": "$.source_relationships[0].direction"})
	})
	t.Run("request scalar order", func(t *testing.T) {
		input := fieldPathInputV2()
		config := input["projection_config"].(map[string]any)
		config["default_vertex_labels"] = []any{""}
		input["source_entities"].([]any)[0].(map[string]any)["properties"] = map[string]any{"z/bad": true, "a/bad": false}
		definition := aggregationPropertyV2("p", "vertex", "*", "properties.group", "p", "string", "single_value")
		definition["source_field_path"] = "kind"
		appendDefinitionV2(input, definition, false)
		expectInputFailureV2(t, inputJSONV2(t, input), "invalid_projection_request", "invalid_property_key", map[string]any{"field": `$.source_entities[0].properties["a/bad"]`})
		config["projection_version"] = "/bad"
		expectInputFailureV2(t, inputJSONV2(t, input), "invalid_projection_request", "invalid_identifier", map[string]any{"field": "$.projection_config.projection_version"})
		definition["default_value"] = map[string]any{"nested": true}
		expectInputFailureV2(t, inputJSONV2(t, input), "invalid_projection_request", "nested_property_value", map[string]any{"field": "$.property_definitions[0].default_value"})
	})
	t.Run("inherited relationship fields precede reverse kind", func(t *testing.T) {
		input := reverseInputV2()
		m := input["projection_config"].(map[string]any)["relationship_mappings"].([]any)[0].(map[string]any)
		m["reverse_edge_kind"], m["required_property_keys"] = "/", []any{"*"}
		expectInputFailureV2(t, inputJSONV2(t, input), "invalid_projection_request", "invalid_property_key", map[string]any{"field": "$.projection_config.relationship_mappings[0].required_property_keys[0]"})
	})
	t.Run("counted scalar issues share the ceiling across the request", func(t *testing.T) {
		input := fieldPathInputV2()
		config := input["projection_config"].(map[string]any)
		labels := repeatedInputItemsV2("", graphProjectionLimits.MaxValidationIssues-1)
		config["default_edge_labels"] = labels
		input["source_relationships"].([]any)[0].(map[string]any)["direction"] = "invalid"
		expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "invalid_input_shape", map[string]any{"issue_count": graphProjectionLimits.MaxValidationIssues})
		config["default_edge_labels"] = append(labels, "")
		expectCollectionOverflowV2(t, inputJSONV2(t, input), "maximum_validation_issues", graphProjectionLimits.MaxValidationIssues)
	})
	for _, tc := range []struct {
		name, reason string
		change       func(map[string]any, map[string]any)
	}{
		{"default labels", "invalid_input_shape", func(input, config map[string]any) { config["default_vertex_labels"] = []any{""} }},
		{"mapping labels", "invalid_input_shape", func(input, config map[string]any) {
			config["entity_mappings"].([]any)[0].(map[string]any)["mapping_labels"] = []any{""}
		}},
		{"mapping direction", "invalid_direction_policy", func(input, config map[string]any) {
			config["relationship_mappings"].([]any)[0].(map[string]any)["direction_policy"] = "sideways"
		}},
		{"predicate operator", "invalid_filter", func(input, config map[string]any) {
			config["entity_mappings"].([]any)[0].(map[string]any)["inclusion_predicate"] = map[string]any{"field_path": "properties.group", "operator": "bad", "include_if_missing": false, "value": true}
		}},
		{"property missing policy", "invalid_property_definition", func(input, config map[string]any) {
			d := aggregationPropertyV2("p", "vertex", "*", "properties.group", "p", "string", "single_value")
			d["missing_behavior"] = "unknown"
			appendDefinitionV2(input, d, false)
		}},
		{"metadata projected type", "invalid_metadata_mapping", func(input, config map[string]any) {
			d := aggregationPropertyV2("p", "vertex", "*", "properties.group", "p", "wrong", "single_value")
			appendDefinitionV2(input, d, true)
		}},
		{"aggregate input scope", "invalid_aggregation_rule", func(input, config map[string]any) {
			config["aggregation_rules"] = []any{aggregationRuleInputV2("r", "vertex", "wrong", "kind", "output")}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := fieldPathInputV2()
			tc.change(input, input["projection_config"].(map[string]any))
			rows := input["source_entities"].([]any)
			input["source_entities"] = append(rows, rows[0])
			for _, overflow := range []bool{false, true} {
				if overflow {
					input["filters"].(map[string]any)["entity_filters"] = repeatedInputItemsV2(filterItemV2(), graphProjectionLimits.MaxEntityFilters+1)
				}
				slices.Reverse(input["source_entities"].([]any))
				expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", tc.reason, map[string]any{"issue_count": 1})
			}
		})
	}
}

func checkMetadataIDsV2(t *testing.T) {
	input := reverseInputV2()
	config := input["projection_config"].(map[string]any)
	for _, key := range []string{"aa", "bb"} {
		def := aggregationPropertyV2("probe", "edge", "edge", "source_relationship_kind", key, "string", "single_value")
		appendDefinitionV2(input, def, true)
		def["metadata_mapping_id"] = "probe"
	}
	defs := config["metadata_mappings"].([]any)
	want := map[string]any{"issue_count": 1}
	for range 2 {
		expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "duplicate_identifier", want)
		slices.Reverse(defs)
	}
	filters := input["filters"].(map[string]any)
	filters["entity_filters"] = repeatedInputItemsV2(filterItemV2(), graphProjectionLimits.MaxEntityFilters+1)
	expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "duplicate_identifier", want)
	filters["entity_filters"] = []any{}
	defs[1].(map[string]any)["metadata_mapping_id"] = "second"
	input["property_definitions"] = []any{aggregationPropertyV2("probe", "edge", "edge", "source_relationship_kind", "aa", "string", "single_value")}
	before, _, err := projectAggregationInputV2(t, input)
	if err != nil {
		t.Fatal(err)
	}
	slices.Reverse(defs)
	after, _, err := projectAggregationInputV2(t, input)
	if err != nil || before.ResultBindingV2() != after.ResultBindingV2() {
		t.Fatalf("distinct metadata ID permutation: %v", err)
	}
	defs[0].(map[string]any)["projected_metadata_key"] = "aa"
	expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "invalid_metadata_mapping", want)
}

func checkAggregationScalarOrderV2(t *testing.T) {
	input := fieldPathInputV2()
	config := input["projection_config"].(map[string]any)
	mapping := config["entity_mappings"].([]any)[0].(map[string]any)
	rule := aggregationRuleInputV2("group", "vertex", "source_entity", "/", "output")
	rule["target_scope"] = "unknown"
	config["aggregation_rules"] = []any{rule}
	limit := graphProjectionLimits.MaxValidationIssues
	labels := make([]string, limit)
	for _, count := range []int{limit - 1, limit} {
		mapping["mapping_labels"] = labels[:count]
		raw := inputJSONV2(t, input)
		// Swap two serialized members without changing the semantic request.
		reordered := bytes.Replace(raw, []byte(`"aggregation_rule_id":"group","grouping_keys":["properties.group"]`), []byte(`"grouping_keys":["properties.group"],"aggregation_rule_id":"group"`), 1)
		if bytes.Equal(raw, reordered) {
			t.Fatal("member permutation failed")
		}
		for _, body := range [][]byte{raw, reordered} {
			if count == limit {
				expectCollectionOverflowV2(t, body, "maximum_validation_issues", limit)
			} else {
				expectInputFailureV2(t, body, "invalid_projection_request", "invalid_identifier", map[string]any{"field": "$.projection_config.aggregation_rules[0].input_kind"})
			}
		}
	}
}

func checkMappingAdmissionV2(t *testing.T) {
	for _, family := range []string{"entity", "relationship"} {
		for _, scenario := range []string{"unused undeclared", "filtered undeclared", "other namespace", "one overlap", "many overlaps", "reversed overlaps", "valid"} {
			t.Run(family+"/"+scenario, func(t *testing.T) {
				input := reverseInputV2()
				c := input["projection_config"].(map[string]any)
				key := family + "_mappings"
				m := c[key].([]any)[0].(map[string]any)
				switch scenario {
				case "unused undeclared", "filtered undeclared", "other namespace":
					extra := maps.Clone(m)
					source := "absent"
					if scenario == "other namespace" {
						if family == "entity" {
							source = "relation"
						} else {
							source = "kind-a"
						}
					}
					extra["mapping_rule_id"], extra["source_"+family+"_kind"] = "extra", source
					c[key] = append(c[key].([]any), extra)
					if scenario == "filtered undeclared" {
						rows := "source_entities"
						if family == "relationship" {
							rows = "source_relationships"
						}
						input[rows].([]any)[0].(map[string]any)["source_"+family+"_kind"] = source
						extra["inclusion_predicate"] = map[string]any{"field_path": "properties.absent", "operator": "exists", "include_if_missing": false}
					}
					if scenario == "unused undeclared" {
						input["source_relationships"] = []any{}
						input["source_entities"] = []any{}
					}
				case "one overlap", "many overlaps", "reversed overlaps":
					m["required_property_keys"] = []string{"a", "b"}
					m["optional_property_keys"] = []string{"b", "c"}
					if scenario == "many overlaps" {
						m["optional_property_keys"] = []string{"a", "b"}
					}
					if scenario == "reversed overlaps" {
						m["optional_property_keys"] = []string{"c", "b"}
					}
				case "valid":
					m["required_property_keys"], m["optional_property_keys"] = []string{"a", "b"}, []string{"c", "d"}
					r, _, err := projectAggregationInputV2(t, input)
					if err != nil {
						t.Fatal(err)
					}
					m["required_property_keys"], m["optional_property_keys"] = []string{"b", "a"}, []string{"d", "c"}
					r2, _, err := projectAggregationInputV2(t, input)
					if err != nil || r2.ResultBindingV2() != r.ResultBindingV2() {
						t.Fatalf("permutation: %v", err)
					}
					return
				}
				expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "invalid_mapping_rule", map[string]any{"issue_count": 1})
			})
		}
	}
}

func checkLabelAdmissionV2(t *testing.T) {
	labels := make([]string, 257)
	for i := range labels {
		labels[i] = "label-" + strconv.Itoa(i)
	}
	for _, n := range []int{256, 257} {
		t.Run("dependency at "+strconv.Itoa(n), func(t *testing.T) {
			input := mixedAggregationInputV2(false)
			c := input["projection_config"].(map[string]any)
			c["default_vertex_labels"] = labels[:n]
			c["aggregation_rules"].([]any)[0].(map[string]any)["endpoint_grouping"].(map[string]any)["src_vertex_aggregation_rule_id"] = "absent"
			body, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			reason := "invalid_aggregation_rule"
			if n > 256 {
				reason = "resource_limit_exceeded"
			}
			expectInputFailureV2(t, body, "projection_validation_failed", reason, map[string]any{"issue_count": 1})
		})
	}
	for _, tc := range []struct{ name, reason string }{{"source before dependency", "source_item_resource_limit_exceeded"}, {"scalar before resource", "invalid_input_shape"}, {"duplicate before resource", "duplicate_identifier"}} {
		t.Run(tc.name, func(t *testing.T) {
			input := mixedAggregationInputV2(false)
			cfg := input["projection_config"].(map[string]any)
			cfg["aggregation_rules"].([]any)[0].(map[string]any)["endpoint_grouping"].(map[string]any)["src_vertex_aggregation_rule_id"] = "absent"
			entities := input["source_entities"].([]any)
			entities[0].(map[string]any)["labels"] = labels
			switch tc.name {
			case "scalar before resource":
				entities[1].(map[string]any)["labels"] = []string{""}
			case "duplicate before resource":
				input["source_entities"] = append(entities, entities[0])
			}
			expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", tc.reason, map[string]any{"issue_count": 1})
		})
	}
	for _, family := range []string{"vertex", "edge"} {
		for _, policy := range []string{"preserve_source", "mapping_then_source", "mapping_only"} {
			t.Run(family+"/"+policy, func(t *testing.T) {
				input := reverseInputV2()
				c := input["projection_config"].(map[string]any)
				c["default_"+family+"_labels"] = labels[:256]
				mapsKey, rows := "entity_mappings", "source_entities"
				if family == "edge" {
					mapsKey, rows = "relationship_mappings", "source_relationships"
				}
				m := c[mapsKey].([]any)[0].(map[string]any)
				m["label_policy"] = policy
				row := input[rows].([]any)[0].(map[string]any)
				for _, n := range []int{255, 256} {
					m["mapping_labels"] = []string{labels[n]}
					row["labels"] = []string{labels[n]}
					if n == 256 {
						expectInputFailureV2(t, inputJSONV2(t, input), "projection_resource_limit_exceeded", "maximum_labels_per_object", map[string]any{"limit_key": "maximum_labels_per_object", "limit": 256, "observed": 257})
						continue
					}
					r, _, err := projectAggregationInputV2(t, input)
					if err != nil {
						t.Fatal(err)
					}
					cr, err := r.CompletedResult()
					if err != nil || cr.Validate(t.Context()) != nil {
						t.Fatalf("completed at 256: %v", err)
					}
				}
			})
		}
	}
}

func checkShapeOrderV2(t *testing.T) {
	input := fieldPathInputV2()
	config := input["projection_config"].(map[string]any)
	config["projection_version"], config["allow_empty_kind_registry"] = 42, "no"
	body := inputJSONV2(t, input)
	want := map[string]any{"field": "$.projection_config.projection_version"}
	expectInputFailureV2(t, body, "invalid_projection_request", "schema_type_mismatch", want)
	body = bytes.Replace(body, []byte(`"allow_empty_kind_registry":"no",`), nil, 1)
	body = bytes.Replace(body, []byte(`"projection_version":42`), []byte(`"projection_version":42,"allow_empty_kind_registry":"no"`), 1)
	expectInputFailureV2(t, body, "invalid_projection_request", "schema_type_mismatch", want)
	delete(config, "projection_version")
	input["unexpected"] = false
	expectInputFailureV2(t, inputJSONV2(t, input), "invalid_projection_request", "missing_required_member", want)
	config["projection_version"] = "valid"
	mapping := config["entity_mappings"].([]any)[0].(map[string]any)
	mapping["inclusion_predicate"] = map[string]any{"field_path": "properties.group", "operator": "exists", "value": nil, "include_if_missing": "bad"}
	expectInputFailureV2(t, inputJSONV2(t, input), "invalid_projection_request", "filter_value_presence", map[string]any{"field": "$.projection_config.entity_mappings[0].inclusion_predicate.value"})
	config["allow_empty_kind_registry"] = false
	mapping["inclusion_predicate"] = "always"
	input["aaa"] = true
	expectInputFailureV2(t, inputJSONV2(t, input), "invalid_projection_request", "unknown_member", map[string]any{"field": "$.aaa"})
}

func expandedBudgetInputV2(kinds, definitions int) map[string]any {
	input := fieldPathInputV2()
	config := input["projection_config"].(map[string]any)
	template := config["entity_mappings"].([]any)[0].(map[string]any)
	mappings, sources := make([]any, kinds), make([]any, kinds)
	for i := range kinds {
		key := "kind-" + strconv.Itoa(i)
		m := maps.Clone(template)
		m["mapping_rule_id"], m["projected_vertex_kind"], m["source_entity_kind"] = key, key, key
		mappings[i], sources[i] = m, key
	}
	config["entity_mappings"], config["declared_source_entity_kinds"] = mappings, sources
	input["source_entities"], input["source_relationships"] = []any{}, []any{}
	for i := range definitions {
		key := "key-" + strconv.Itoa(i)
		def := aggregationPropertyV2(key, "vertex", "*", "properties.group", key, "string", "first")
		appendDefinitionV2(input, def, i%2 == 1)
		if i%2 == 1 {
			def["metadata_mapping_id"] = key
		}
	}
	return input
}

func checkInputBudgetsV2(t *testing.T) {
	t.Run("unused source property", func(t *testing.T) {
		input := fieldPathInputV2()
		input["source_entities"].([]any)[0].(map[string]any)["properties"].(map[string]any)["unused"] = repeatedInputItemsV2(false, 100001)
		expectCollectionOverflowV2(t, inputJSONV2(t, input), "maximum_array_items", 100000)
	})
	t.Run("expanded declarations", func(t *testing.T) {
		input := expandedBudgetInputV2(1000, 100)
		// Both declaration namespaces contribute to the same exact ceiling.
		result, _, err := projectAggregationInputV2(t, input)
		if err != nil || len(result.SchemaRegistry.PropertyKeys)+len(result.SchemaRegistry.MetadataKeys) != 100000 {
			t.Fatalf("at expansion ceiling: %v", err)
		}

		appendDefinitionV2(input, aggregationPropertyV2("extra", "vertex", "inert", "properties.group", "extra", "string", "first"), false)
		expectCollectionOverflowV2(t, inputJSONV2(t, input), "maximum_expanded_registry_declarations", 100000)
		// The input array verdict precedes expansion, but uniqueness still wins first.
		def := input["property_definitions"].([]any)[0].(map[string]any)
		def["default_value"] = repeatedInputItemsV2(false, 100001)
		expectCollectionOverflowV2(t, inputJSONV2(t, input), "maximum_array_items", 100000)
		input["property_definitions"] = append(input["property_definitions"].([]any), def)
		expectInputFailureV2(t, inputJSONV2(t, input), "projection_validation_failed", "duplicate_identifier", map[string]any{"issue_count": 1})
	})
}

func checkArrayAdmissionSitesV2(t *testing.T) {
	values := repeatedInputItemsV2(false, 100001)
	for _, site := range []string{"entity", "relationship", "filtered", "filter", "inclusion", "default", "metadata default", "mapped metadata"} {
		t.Run(site, func(t *testing.T) {
			for _, n := range []int{100000, 100001} {
				input := fieldPathInputV2()
				cfg := input["projection_config"].(map[string]any)
				field := "properties.unused"
				var predicate map[string]any
				switch site {
				case "entity", "filtered":
					for _, raw := range input["source_entities"].([]any) {
						raw.(map[string]any)["properties"].(map[string]any)["unused"] = values[:n]
					}
					if site == "filtered" {
						input["filters"].(map[string]any)["entity_filters"] = []any{map[string]any{"field_path": "properties.absent", "operator": "exists", "include_if_missing": false}}
						input["source_relationships"] = []any{}
					}
				case "relationship":
					input["source_relationships"].([]any)[0].(map[string]any)["properties"].(map[string]any)["unused"] = values[:n]
				case "filter", "inclusion":
					predicate = map[string]any{"field_path": "properties.group", "operator": "in", "value": values[:n], "include_if_missing": false}
					input["source_relationships"] = []any{}
					if site == "filter" {
						input["filters"].(map[string]any)["entity_filters"] = []any{predicate}
					} else {
						cfg["entity_mappings"].([]any)[0].(map[string]any)["inclusion_predicate"] = predicate
					}
				case "default", "metadata default":
					def := aggregationPropertyV2("default", "vertex", "inert", field, "value", "boolean_array", "first")
					def["default_value"], def["missing_behavior"] = values[:n], "default"
					appendDefinitionV2(input, def, site == "metadata default")
				case "mapped metadata":
					input["source_metadata"].(map[string]any)["unused"] = values[:n]
					appendDefinitionV2(input, aggregationPropertyV2("metadata", "graph_view", "*", "source_metadata.unused", "value", "boolean_array", "first"), true)
				}
				if site == "entity" || site == "relationship" {
					scope, kind := "vertex", "vertex"
					if site == "relationship" {
						scope, kind = "edge", "edge"
					}
					appendDefinitionV2(input, aggregationPropertyV2("array", scope, kind, field, "value", "boolean_array", "first"), false)
				}
				if n == 100001 {
					expectCollectionOverflowV2(t, inputJSONV2(t, input), "maximum_array_items", 100000)
				} else if _, _, err := projectAggregationInputV2(t, input); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func checkExpansionPreflightV2(t *testing.T) {
	input := expandedBudgetInputV2(10000, 10000)
	_, err := admitDeclarationTargetsV2(t.Context(), input)
	var e *ProjectionErrorV2
	if !errors.As(err, &e) || e.ReasonCode != "maximum_expanded_registry_declarations" || e.Details["observed"] != 100001 {
		t.Fatalf("unbounded expansion: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := admitDeclarationTargetsV2(ctx, input); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled preflight: %v", err)
	}
}

func checkStoredDeclarationBudgetV2(t *testing.T) {
	values := make([]any, graphProjectionLimits.MaxExpandedRegistryDeclarations)
	kind := map[string]any{"properties": values[:10]}
	reg := map[string]any{"vertex_kinds": []any{kind}, "edge_kinds": []any{}, "property_keys": values[:10], "metadata_keys": values[10:]}
	if !registryDeclarationsFitV2(reg) {
		t.Fatal("rejected combined ceiling")
	}
	reg["metadata_keys"] = values[9:]
	if registryDeclarationsFitV2(reg) {
		t.Fatal("accepted one excess top-level declaration")
	}
	reg["metadata_keys"] = []any{}
	kind["properties"] = values[:11]
	if registryDeclarationsFitV2(reg) {
		t.Fatal("accepted excess nested references")
	}
}
