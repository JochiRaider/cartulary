package graphprojection

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func checkIssueBudgetV2(t *testing.T) {
	input := twoEntityProjectionInputV2(false)
	cfg := input["projection_config"].(map[string]any)
	cfg["entity_mappings"].([]any)[0].(map[string]any)["inclusion_predicate"] = map[string]any{"field_path": "properties.hidden", "operator": "exists", "include_if_missing": false}
	var items, rules []any
	for i := 0; i < 1000; i++ {
		items = append(items, map[string]any{"source_entity_id": fmt.Sprintf("item-%04d", i), "source_entity_kind": "kind-a", "labels": []any{}, "metadata": map[string]any{}, "properties": map[string]any{}})
	}
	for i := 0; i < 100; i++ {
		rules = append(rules, aggregationRuleInputV2(fmt.Sprintf("rule-%03d", i), "vertex", "source_entity", "kind-a", "group"))
	}
	cfg["aggregation_rules"], input["source_entities"] = rules, items
	r, _, err := projectAggregationInputV2(t, input)
	var pe *ProjectionErrorV2
	if !errors.As(err, &pe) || pe.Code != "projection_validation_failed" || pe.ReasonCode != "aggregation_grouping_key_missing" || pe.Details["issue_count"] != 100000 || !reflect.DeepEqual(r, ProjectionResultV2{}) {
		t.Fatalf("exact issue budget: %v", err)
	}
	cfg["aggregation_rules"] = append(rules, aggregationRuleInputV2("z-over", "vertex", "source_entity", "kind-a", "over"), aggregationRuleInputV2("zz-late", "vertex", "source_entity", "kind-a", "late"))
	late := cfg["aggregation_rules"].([]any)[101].(map[string]any)
	late["grouping_keys"] = []any{"source_entity_kind"}
	r, raw, err := projectAggregationInputV2(t, input)
	assertOutputLimitV2(t, err, "maximum_validation_issues", 100000)
	if !reflect.DeepEqual(r, ProjectionResultV2{}) {
		t.Fatal("partial issue-budget result")
	}
	request, _, _, err := parseProjectionInputV2(raw)
	if err != nil {
		t.Fatal(err)
	}
	state := projectionState{}
	run := &projectionWork{GraphViewID: "limits", Request: request}
	err = state.emitAggregations(t.Context(), run)
	assertOutputLimitV2(t, err, "maximum_validation_issues", 100000)
	if len(run.issues.entries) != 100000 || len(state.edges) != 0 || len(state.vertices) != 0 {
		t.Fatal("issue retention exceeded or later rule executed")
	}
}

func checkArrayBudgetV2(t *testing.T) {
	for _, merge := range []string{"set", "ordered_list"} {
		for _, metadata := range []bool{false, true} {
			for _, mode := range []string{"exact", "overflow", "duplicates"} {
				if mode == "duplicates" && merge != "set" {
					continue
				}
				t.Run(fmt.Sprintf("%s/metadata=%t/%s", merge, metadata, mode), func(t *testing.T) {
					input := numericAggregationInputV2("same", "same")
					cfg := input["projection_config"].(map[string]any)
					cfg["aggregation_rules"] = cfg["aggregation_rules"].([]any)[1:]
					def := aggregationPropertyV2("values", "vertex", "endpoint", "properties.values", "values", "integer_array", merge)
					appendDefinitionV2(input, def, metadata)
					for i, raw := range input["source_entities"].([]any) {
						n := 50000
						if i == 1 && mode != "exact" {
							n++
						}
						arr := make([]int, n)
						for j := range arr {
							arr[j] = j + i*50000
						}
						if mode == "duplicates" && i == 1 {
							arr[n-1] = 1
						}
						raw.(map[string]any)["properties"].(map[string]any)["values"] = arr
					}
					r, _, err := projectAggregationInputV2(t, input)
					if mode == "overflow" {
						assertOutputLimitV2(t, err, "maximum_array_items", 100000)
						if !reflect.DeepEqual(r, ProjectionResultV2{}) {
							t.Fatal("partial result")
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					for _, v := range r.Vertices {
						if v.VertexFamily == "aggregated" {
							m := v.Properties
							if metadata {
								m = v.Metadata.MappedMetadata
							}
							if a, ok := m["values"].([]any); !ok || len(a) != 100000 {
								t.Fatal("incorrect cardinality")
							}
						}
					}
					c, err := r.CompletedResult()
					if err != nil || c.Validate(t.Context()) != nil {
						t.Fatalf("unpublishable bounded array: %v", err)
					}
				})
			}
		}
	}
}

func checkOneShotCheckpointV2(t *testing.T) {
	input := maximumProjectionInputV2(t, 2048, 0)
	for _, mode := range []string{"callback", "deadline", "success"} {
		t.Run(mode, func(t *testing.T) {
			ctx := t.Context()
			inv := InvocationContextV2{GraphViewID: "checkpoint", SourceOwnerID: "network_flow_activity"}
			cause := context.Canceled
			switch mode {
			case "callback":
				calls := 0
				inv.CancellationCheck = func(context.Context) error {
					calls++
					if calls == 2 {
						return fmt.Errorf("once: %w", context.Canceled)
					}
					return nil
				}
			case "deadline":
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, time.Unix(1, 0))
				defer cancel()
				cause = context.DeadlineExceeded
			}
			r, err := ProjectV2(ctx, inv, input)
			if mode == "success" {
				if err != nil || len(r.Vertices) != 2048 {
					t.Fatalf("uninterrupted projection: %v", err)
				}
				return
			}
			var pe *ProjectionErrorV2
			if !errors.As(err, &pe) || !errors.Is(err, cause) || pe.Code != "projection_cancelled" || pe.ReasonCode != "context_cancelled" || pe.RetryAction != "do_not_retry" || !reflect.DeepEqual(r, ProjectionResultV2{}) {
				t.Fatalf("cancellation misclassified: %v", err)
			}
		})
	}
}

// Retain an independent completed-envelope witness even when the producer refuses
// to construct an oversized mapped value.
func checkStoredArrayBoundsV2(t *testing.T) {
	for _, where := range []string{"vertex property", "vertex metadata", "edge metadata"} {
		t.Run(where, func(t *testing.T) {
			r, _, err := projectAggregationInputV2(t, fieldPathInputV2())
			if err != nil {
				t.Fatal(err)
			}
			array := make([]any, 100001)
			switch where {
			case "vertex property":
				r.Vertices[0].Properties["array"] = array
			case "vertex metadata":
				r.Vertices[0].Metadata.MappedMetadata["array"] = array
			case "edge metadata":
				r.Edges[0].Metadata.MappedMetadata["array"] = array
			}
			r.CanonicalOutputSHA256, err = canonicalOutputDigestV2(r.outputResource())
			if err != nil {
				t.Fatal(err)
			}
			r.ProjectionResultID, err = deriveProjectionResultIDV2(r.ResultBindingV2())
			if err != nil {
				t.Fatal(err)
			}
			c, err := r.CompletedResult()
			if err != nil {
				t.Fatal(err)
			}
			if !errors.Is(c.Validate(t.Context()), ErrResultV2Invalid) {
				t.Fatal("oversized stored array admitted")
			}
		})
	}
}
