package graphprojection

import (
	"context"
	"fmt"
	"slices"

	"strings"
)

func validateAdmittedRequest(run *projectionWork) {
	if run.issues.err != nil {
		return
	}
	request := run.Request
	if len(request.projectionConfig.DeclaredSourceEntityKinds) == 0 && len(request.projectionConfig.DeclaredSourceRelationshipKinds) == 0 && !request.projectionConfig.AllowEmptyKindRegistry {
		run.recordIssue("fatal", "invalid_projection_config", "projection_config", "projection_config", "$.projection_config", map[string]any{"field": "$.projection_config", "reason_code": "empty_kind_registry_not_allowed"})
	}
	checkDuplicateKinds := func(values []string) {
		for index := 1; index < len(values); index++ {
			if run.issues.err != nil {
				return
			}
			if values[index] == values[index-1] {
				run.recordIssue("fatal", "invalid_projection_config", "projection_config", "projection_config", nil, map[string]any{"field": "$.projection_config", "reason_code": "declared_kind_duplicate"})
			}
		}
	}
	checkDuplicateKinds(request.projectionConfig.DeclaredSourceEntityKinds)
	checkDuplicateKinds(request.projectionConfig.DeclaredSourceRelationshipKinds)
	validateFilters(run)
	validateMappingsAndDefinitions(run)
	for _, mapping := range request.projectionConfig.RelationshipMappings {
		if run.issues.err != nil {
			return
		}
		if mapping.EmitReverseEdge && !mappingDirectionPoliciesV2[mapping.DirectionPolicy].allowsReverse {
			run.recordIssue("fatal", "invalid_reverse_edge_policy", "mapping_rule", mapping.MappingRuleID, nil, map[string]any{"mapping_rule_id": mapping.MappingRuleID, "projected_direction": nil})
		}
	}
}

// Source entities are in ID order after admission. Resolution is independent
// of projection eligibility and needs no additional source-sized index.
func validateSubmittedEndpointsV2(ctx context.Context, request projectionRequest) error {
	missing := 0
	for _, relationship := range request.SourceRelationships {
		if err := itemCheckpointV2(ctx); err != nil {
			return err
		}
		for _, id := range []string{relationship.SrcSourceEntityID, relationship.DstSourceEntityID} {
			_, found := slices.BinarySearchFunc(request.SourceEntities, id, func(entity sourceEntity, id string) int {
				return strings.Compare(entity.SourceEntityID, id)
			})
			if !found {
				missing++
				if missing > graphProjectionLimits.MaxValidationIssues {
					return validationIssueLimitV2()
				}
			}
		}
	}
	if missing > 0 {
		return validationFailureV2("missing_relationship_endpoint", missing)
	}
	return nil
}

func validateFilters(run *projectionWork) {
	if run.issues.err != nil {
		return
	}
	validate := func(predicates []filterPredicate, collection, scope string) {
		for index, predicate := range predicates {
			if run.issues.err != nil {
				return
			}
			validateFilterPredicate(run, predicate, fmt.Sprintf("%s[%d]", collection, index), scope)
		}
	}
	validate(run.Request.filters.EntityFilters, "$.filters.entity_filters", "source_entity")
	validate(run.Request.filters.RelationshipFilters, "$.filters.relationship_filters", "source_relationship")
}

func validateFilterPredicate(run *projectionWork, predicate filterPredicate, base, scope string) {
	if run.issues.err != nil {
		return
	}
	if !validFilterFieldPath(predicate.FieldPath, scope) {
		run.recordIssue("fatal", "invalid_filter", "filter", base, base+".field_path", map[string]any{"field": base + ".field_path", "reason_code": "invalid_field_scope"})
	}
	if predicate.Operator == "in" {
		values, ok := predicate.Value.([]any)
		if !ok || len(values) == 0 {
			run.recordIssue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "invalid_value_shape"})
		}
	}
}

func validateMappingsAndDefinitions(run *projectionWork) {
	if run.issues.err != nil {
		return
	}
	request := run.Request
	declaredEntities := stringSet(request.projectionConfig.DeclaredSourceEntityKinds)
	declaredRelationships := stringSet(request.projectionConfig.DeclaredSourceRelationshipKinds)
	entityRuleIDs := map[string]bool{}
	entityKinds := map[string]string{}

	for _, mapping := range request.projectionConfig.EntityMappings {
		if run.issues.err != nil {
			return
		}
		if entityRuleIDs[mapping.MappingRuleID] {
			invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_mapping_rule_id")
		}
		validateMappingKeysV2(run, mapping.MappingRuleID, declaredEntities[mapping.SourceEntityKind], mapping.RequiredPropertyKeys, mapping.OptionalPropertyKeys)
		entityRuleIDs[mapping.MappingRuleID] = true
		if prior := entityKinds[mapping.SourceEntityKind]; prior != "" {
			invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_source_entity_kind_mapping")
		}
		entityKinds[mapping.SourceEntityKind] = mapping.MappingRuleID

		if mapping.InclusionFilter != nil {
			validateFilterPredicate(run, *mapping.InclusionFilter, "$.projection_config.entity_mappings."+mapping.MappingRuleID+".inclusion_predicate", "source_entity")
		}
	}
	relationshipRuleIDs := map[string]bool{}
	relationshipKinds := map[string]string{}

	for _, mapping := range request.projectionConfig.RelationshipMappings {
		if run.issues.err != nil {
			return
		}
		if entityRuleIDs[mapping.MappingRuleID] || relationshipRuleIDs[mapping.MappingRuleID] {
			invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_mapping_rule_id")
		}
		validateMappingKeysV2(run, mapping.MappingRuleID, declaredRelationships[mapping.SourceRelationshipKind], mapping.RequiredPropertyKeys, mapping.OptionalPropertyKeys)
		relationshipRuleIDs[mapping.MappingRuleID] = true
		if prior := relationshipKinds[mapping.SourceRelationshipKind]; prior != "" {
			invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_source_relationship_kind_mapping")
		}
		relationshipKinds[mapping.SourceRelationshipKind] = mapping.MappingRuleID
		if mapping.InclusionFilter != nil {
			validateFilterPredicate(run, *mapping.InclusionFilter, "$.projection_config.relationship_mappings."+mapping.MappingRuleID+".inclusion_predicate", "source_relationship")
		}
	}
	validateAggregationRules(run)
	for _, kind := range request.projectionConfig.DeclaredSourceEntityKinds {
		if run.issues.err != nil {
			return
		}
		if entityKinds[kind] == "" {
			run.recordIssue("fatal", "missing_entity_mapping_rule", "mapping_rule", kind, nil, map[string]any{"source_entity_kind": kind})
		}
	}
	for _, kind := range request.projectionConfig.DeclaredSourceRelationshipKinds {
		if run.issues.err != nil {
			return
		}
		if relationshipKinds[kind] == "" {
			run.recordIssue("error", "missing_relationship_mapping_rule", "mapping_rule", kind, nil, map[string]any{"source_relationship_kind": kind})
		}
	}
	propertyExpansions := map[string]string{}

	for _, definition := range request.PropertyDefinitions {
		if run.issues.err != nil {
			return
		}
		if definition.HasDefaultValue && (definition.DefaultValue == nil && definition.NullOutputPolicy != "emit_null" || definition.DefaultValue != nil && !valueMatchesType(definition.ProjectedType, definition.DefaultValue)) {
			invalidPropertyDefinitionIssue(run, definition, "invalid_default_value")
		}
		if (definition.MissingBehavior == "default" || definition.SourceNullBehavior == "default") && !definition.HasDefaultValue {
			invalidPropertyDefinitionIssue(run, definition, "invalid_default_value")
		}
		if definition.SourceNullBehavior == "emit_null" && definition.NullOutputPolicy != "emit_null" {
			invalidPropertyDefinitionIssue(run, definition, "invalid_null_policy")
		}
		if !validProjectedMergeV2(definition.ProjectedType, definition.MergeBehavior) {
			invalidPropertyDefinitionIssue(run, definition, "invalid_merge_behavior_type")
		}
		field, _ := classifyFieldPath(definition.SourceFieldPath)
		if field.member == "metadata" || field.member == "source_metadata" || !request.targets.allows(field, definition.TargetScope, definition.TargetKind) {
			run.recordIssue("fatal", "invalid_field_path", "projection_config", "projection_config", definition.SourceFieldPath, map[string]any{"field_path": definition.SourceFieldPath, "scope": definition.TargetScope})
		}
		for _, kind := range request.targets.expand(definition.TargetScope, definition.TargetKind) {
			if run.issues.err != nil {
				return
			}
			key := definition.TargetScope + "\n" + kind + "\n" + definition.ProjectedKey
			if propertyExpansions[key] != "" {
				run.recordIssue("fatal", "invalid_property_definition", "property_definition", definition.PropertyDefinitionID, nil, map[string]any{"property_definition_id": definition.PropertyDefinitionID, "reason_code": "duplicate_after_wildcard_expansion"})
			}
			propertyExpansions[key] = definition.PropertyDefinitionID
		}
	}
	metadataExpansions := map[string]string{}
	for _, mapping := range request.projectionConfig.MetadataMappings {
		if run.issues.err != nil {
			return
		}
		field, _ := classifyFieldPath(mapping.SourceFieldPath)
		if !request.targets.allows(field, mapping.TargetScope, mapping.TargetKind) {
			run.recordIssue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, mapping.SourceFieldPath, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "invalid_source_scope"})
		}
		if mapping.HasDefaultValue && (mapping.DefaultValue == nil && mapping.NullOutputPolicy != "emit_null" || mapping.DefaultValue != nil && !valueMatchesType(mapping.ProjectedType, mapping.DefaultValue)) {
			invalidMetadataMappingIssue(run, mapping, "invalid_default_value")
		}
		if (mapping.MissingBehavior == "default" || mapping.SourceNullBehavior == "default") && !mapping.HasDefaultValue {
			invalidMetadataMappingIssue(run, mapping, "invalid_default_value")
		}
		if mapping.SourceNullBehavior == "emit_null" && mapping.NullOutputPolicy != "emit_null" {
			invalidMetadataMappingIssue(run, mapping, "invalid_default_value")
		}
		if !validProjectedMergeV2(mapping.ProjectedType, mapping.MergeBehavior) {
			invalidMetadataMappingIssue(run, mapping, "invalid_merge_behavior_type")
		}
		for _, kind := range request.targets.expand(mapping.TargetScope, mapping.TargetKind) {
			if run.issues.err != nil {
				return
			}
			key := mapping.TargetScope + "\n" + kind + "\n" + mapping.ProjectedMetadataKey
			if metadataExpansions[key] != "" {
				run.recordIssue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, nil, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "duplicate_after_wildcard_expansion"})
			}
			metadataExpansions[key] = mapping.MetadataMappingID
		}
	}
}

func validateAggregationRules(run *projectionWork) {
	if run.issues.err != nil {
		return
	}
	rules := run.Request.projectionConfig.AggregationRules
	indexes := map[string]int{}
	byID := map[string]aggregationRule{}
	for index, rule := range rules {
		if run.issues.err != nil {
			return
		}
		if _, exists := indexes[rule.AggregationRuleID]; exists {
			invalidAggregationIssue(run, rule.AggregationRuleID, "aggregation_cycle")
		}
		indexes[rule.AggregationRuleID] = index
		byID[rule.AggregationRuleID] = rule
	}
	for index, rule := range rules {
		if run.issues.err != nil {
			return
		}
		if len(rule.GroupingKeys) == 0 || len(rule.GroupingKeys) > 32 || hasDuplicateStringsLocal(rule.GroupingKeys) {
			invalidAggregationIssue(run, rule.AggregationRuleID, "grouping_key_invalid")
		}
		for _, fieldPath := range rule.GroupingKeys {
			if run.issues.err != nil {
				return
			}
			if !validAggregationFieldPath(fieldPath, rule.InputScope) {
				invalidAggregationIssue(run, rule.AggregationRuleID, "grouping_key_invalid")
			}
		}
		for projectedKey, mergeBehavior := range rule.PropertyMergeBehavior {
			if run.issues.err != nil {
				return
			}
			matched := false
			for _, definition := range run.Request.PropertyDefinitions {
				if run.issues.err != nil {
					return
				}
				if definition.ProjectedKey == projectedKey && propertyApplies(definition, rule.TargetScope, rule.ProjectedKind) {
					matched = true
					if !validProjectedMergeV2(definition.ProjectedType, mergeBehavior) {
						invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_merge_behavior_type")
					}
				}
			}
			if !matched {
				invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_merge_behavior_type")
			}
		}
		if rule.TargetScope == "vertex" {
			continue
		}
		endpoint := rule.endpointGrouping
		for _, reference := range []struct {
			id   string
			keys []string
		}{{id: endpoint.SourceVertexAggregationRuleID, keys: endpoint.SourceGroupingKeys}, {id: endpoint.DestinationVertexAggregationRuleID, keys: endpoint.DestinationGroupingKeys}} {
			referenced, exists := byID[reference.id]
			if !exists || referenced.TargetScope != "vertex" {
				invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_rule_not_vertex_rule")
				continue
			}
			if indexes[reference.id] >= index {
				invalidAggregationIssue(run, rule.AggregationRuleID, "dependency_on_later_rule")
			}
			if len(reference.keys) != len(referenced.GroupingKeys) {
				invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_grouping_key_count_mismatch")
			}
			for _, fieldPath := range reference.keys {
				if run.issues.err != nil {
					return
				}
				if !validAggregationFieldPath(fieldPath, rule.InputScope) {
					invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_field_scope_invalid")
				}
			}
		}
	}
}

func invalidAggregationIssue(run *projectionWork, aggregationRuleID, reason string) {
	run.recordIssue("fatal", "invalid_aggregation_rule", "mapping_rule", aggregationRuleID, nil, map[string]any{"aggregation_rule_id": aggregationRuleID, "reason_code": reason})
}

func hasDuplicateStringsLocal(values []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}

func invalidPropertyDefinitionIssue(run *projectionWork, definition propertyDefinition, reason string) {
	run.recordIssue("fatal", "invalid_property_definition", "property_definition", definition.PropertyDefinitionID, nil, map[string]any{"property_definition_id": definition.PropertyDefinitionID, "reason_code": reason})
}

func invalidMetadataMappingIssue(run *projectionWork, mapping metadataMapping, reason string) {
	run.recordIssue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, nil, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": reason})
}

func validateMappingKeysV2(run *projectionWork, id string, declared bool, required, optional []string) {
	if !declared {
		invalidMappingIssue(run, id, "declared_source_kind_missing")
	}
	// Both lists are in canonical key order.
	for i, j := 0, 0; i < len(required) && j < len(optional); {
		switch strings.Compare(required[i], optional[j]) {
		case -1:
			i++
		case 1:
			j++
		default:
			invalidMappingIssue(run, id, "required_optional_overlap")
			return
		}
	}
}

func invalidMappingIssue(run *projectionWork, mappingRuleID, reason string) {
	run.recordIssue("fatal", "invalid_mapping_rule", "mapping_rule", mappingRuleID, nil, map[string]any{"mapping_rule_id": mappingRuleID, "reason_code": reason})
}

func projectedOutputLimitIssues(run *projectionWork, vertices []Vertex, edges []Edge) {
	if run.issues.err != nil {
		return
	}
	add := func(key string, observed, limit int) {
		if observed > limit {
			run.recordIssue("fatal", "projected_output_limit_exceeded", "graph_view", run.GraphViewID, nil, map[string]any{"limit_key": key, "limit": limit, "observed": observed})
		}
	}
	add("maximum_projected_vertices", len(vertices), graphProjectionLimits.MaxProjectedVertices)
	add("maximum_projected_edges", len(edges), graphProjectionLimits.MaxProjectedEdges)
	for _, vertex := range vertices {
		if run.issues.err != nil {
			return
		}
		add("maximum_property_keys", len(vertex.Properties), graphProjectionLimits.MaxPropertyKeys)
		add("maximum_property_keys", len(vertex.Metadata.MappedMetadata), graphProjectionLimits.MaxPropertyKeys)
	}
	for _, edge := range edges {
		if run.issues.err != nil {
			return
		}
		add("maximum_property_keys", len(edge.Properties), graphProjectionLimits.MaxPropertyKeys)
		add("maximum_property_keys", len(edge.Metadata.MappedMetadata), graphProjectionLimits.MaxPropertyKeys)
	}
}
