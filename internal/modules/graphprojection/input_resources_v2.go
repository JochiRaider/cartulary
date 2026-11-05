package graphprojection

import (
	"context"
	"slices"
)

// Resource admission follows scalar and unique-ID admission. Only counts and
// the winning reason are public. No normalized objects or issue IDs are needed.
func admitInputResourceCountsV2(ctx context.Context, root map[string]any) error {
	limits := graphProjectionLimits
	config := root["projection_config"].(map[string]any)
	count := 0
	add := func(excess bool) bool {
		if excess {
			count++
		}
		return count > limits.MaxValidationIssues
	}
	add(len(config["declared_source_entity_kinds"].([]any)) > limits.MaxEntityMappings)
	add(len(config["declared_source_relationship_kinds"].([]any)) > limits.MaxRelationshipMappings)
	add(len(root["source_metadata"].(map[string]any)) > limits.MaxPropertyKeys)
	for _, key := range []string{"default_vertex_labels", "default_edge_labels"} {
		add(exceedsStringSetLimitV2(config[key].([]any), limits.MaxLabelsPerObject))
	}
	for _, key := range []string{"entity_mappings", "relationship_mappings"} {
		for _, raw := range config[key].([]any) {
			mapping := raw.(map[string]any)
			add(exceedsStringSetLimitV2(mapping["mapping_labels"].([]any), limits.MaxLabelsPerObject))
			add(exceedsStringSetLimitV2(mapping["required_property_keys"].([]any), limits.MaxPropertyKeys))
			add(exceedsStringSetLimitV2(mapping["optional_property_keys"].([]any), limits.MaxPropertyKeys))
		}
	}
	// The nine collection ceilings bound the configuration checks above to
	// fewer than MaxValidationIssues. Their fatal reason precedes source issues.
	reason := "source_item_resource_limit_exceeded"
	if count > 0 {
		reason = "resource_limit_exceeded"
	}
	for _, key := range []string{"source_entities", "source_relationships"} {
		for _, raw := range root[key].([]any) {
			if err := itemCheckpointV2(ctx); err != nil {
				return cancelledProjectionErrorV2(err)
			}
			row := raw.(map[string]any)
			if add(len(row["properties"].(map[string]any)) > limits.MaxPropertyKeys) ||
				add(len(row["metadata"].(map[string]any)) > limits.MaxPropertyKeys) ||
				add(exceedsStringSetLimitV2(row["labels"].([]any), limits.MaxLabelsPerObject)) {
				return validationIssueLimitV2()
			}
		}
	}
	if count > 0 {
		return validationFailureV2(reason, count)
	}
	return nil
}

// The v2 representation deduplicates mapping keys and labels before counting.
func exceedsStringSetLimitV2(values []any, limit int) bool {
	if len(values) <= limit {
		return false
	}
	seen := make(map[string]bool, limit)
	for _, raw := range values {
		value := raw.(string)
		if !seen[value] && len(seen) == limit {
			return true
		}
		seen[value] = true
	}
	return false
}

func admitDeclarationTargetsV2(ctx context.Context, root map[string]any) (configuredTargetsV2, error) {
	inputScopes := map[string]fieldScope{"source_entity": fieldSourceEntity, "source_relationship": fieldSourceRelationship, "projected_vertex": fieldProjectedVertex, "projected_edge": fieldProjectedEdge}
	targets := configuredTargetsV2{map[definitionTarget]fieldScope{}, map[string][]string{}, map[string]fieldScope{}}
	add := func(scope, kind string, input fieldScope) {
		key := definitionTarget{scope, kind}
		if _, exists := targets.contexts[key]; !exists {
			targets.byScope[scope] = append(targets.byScope[scope], kind)
		}
		targets.contexts[key] |= input
		targets.wildcardContexts[scope] |= input
	}
	config := root["projection_config"].(map[string]any)
	for _, collection := range []string{"entity_mappings", "relationship_mappings", "aggregation_rules"} {
		for _, raw := range config[collection].([]any) {
			if err := itemCheckpointV2(ctx); err != nil {
				return targets, cancelledProjectionErrorV2(err)
			}
			row := raw.(map[string]any)
			switch collection {
			case "entity_mappings":
				add("vertex", row["projected_vertex_kind"].(string), fieldSourceEntity)
			case "relationship_mappings":
				add("edge", row["projected_edge_kind"].(string), fieldSourceRelationship)
				if row["emit_reverse_edge"].(bool) {
					add("edge", row["reverse_edge_kind"].(string), fieldSourceRelationship)
				}
			case "aggregation_rules":

				add(row["target_scope"].(string), row["projected_kind"].(string), inputScopes[row["input_scope"].(string)])
			}
		}
	}
	limit := graphProjectionLimits.MaxExpandedRegistryDeclarations
	remaining := limit
	for _, definitions := range [][]any{root["property_definitions"].([]any), config["metadata_mappings"].([]any)} {
		for _, raw := range definitions {
			if err := itemCheckpointV2(ctx); err != nil {
				return targets, cancelledProjectionErrorV2(err)
			}
			row := raw.(map[string]any)
			remaining -= len(targets.expand(row["target_scope"].(string), row["target_kind"].(string)))
			if remaining < 0 {
				return targets, resourceProjectionErrorV2("maximum_expanded_registry_declarations", limit, limit+1)
			}
		}
	}
	for _, kinds := range targets.byScope {
		slices.Sort(kinds)
	}
	return targets, nil
}
