package graphprojection

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func validateAdmittedRequest(run projectionWork) []ValidationIssue {
	request := run.Request
	issues := []ValidationIssue{}
	identifierValid := validIdentifierV2
	scalar := func(valid bool, field string) {
		if !valid {
			issues = append(issues, run.issue("fatal", "invalid_input_shape", "projection_input", "projection_input", field, map[string]any{"field": field, "reason_code": "scalar_contract_violation"}))
		}
	}
	scalar(identifierValid(request.SourceSnapshotID), "$.source_snapshot_id")
	for index, entity := range request.SourceEntities {
		base := fmt.Sprintf("$.source_entities[%d]", index)
		scalar(identifierValid(entity.SourceEntityID), base+".source_entity_id")
		scalar(identifierValid(entity.SourceEntityKind), base+".source_entity_kind")
	}
	for index, relationship := range request.SourceRelationships {
		base := fmt.Sprintf("$.source_relationships[%d]", index)
		scalar(identifierValid(relationship.SourceRelationshipID), base+".source_relationship_id")
		scalar(identifierValid(relationship.SourceRelationshipKind), base+".source_relationship_kind")
		if relationship.SrcSourceEntityID != "" {
			scalar(identifierValid(relationship.SrcSourceEntityID), base+".src_source_entity_id")
		}
		if relationship.DstSourceEntityID != "" {
			scalar(identifierValid(relationship.DstSourceEntityID), base+".dst_source_entity_id")
		}
	}
	if len(request.projectionConfig.DeclaredSourceEntityKinds) == 0 && len(request.projectionConfig.DeclaredSourceRelationshipKinds) == 0 && !request.projectionConfig.AllowEmptyKindRegistry {
		issues = append(issues, run.issue("fatal", "invalid_projection_config", "projection_config", "projection_config", "$.projection_config", map[string]any{"field": "$.projection_config", "reason_code": "empty_kind_registry_not_allowed"}))
	}
	checkDuplicates := func(values []string, collection string) {
		seen := map[string]struct{}{}
		for _, value := range values {
			if _, ok := seen[value]; ok {
				issues = append(issues, run.issue("fatal", "duplicate_identifier", "projection_input", "projection_input", nil, map[string]any{"identifier_value": value, "collection": collection}))
			}
			seen[value] = struct{}{}
		}
	}
	checkDuplicates(entityIDs(request.SourceEntities), "$.source_entities")
	checkDuplicates(relationshipIDs(request.SourceRelationships), "$.source_relationships")
	checkDuplicateKinds := func(values []string) {
		for index := 1; index < len(values); index++ {
			if values[index] == values[index-1] {
				issues = append(issues, run.issue("fatal", "invalid_projection_config", "projection_config", "projection_config", nil, map[string]any{"field": "$.projection_config", "reason_code": "declared_kind_duplicate"}))
			}
		}
	}
	checkDuplicateKinds(request.projectionConfig.DeclaredSourceEntityKinds)
	checkDuplicateKinds(request.projectionConfig.DeclaredSourceRelationshipKinds)
	issues = append(issues, admittedResourceLimitIssues(run)...)
	issues = append(issues, validateLabels(run)...)
	issues = append(issues, validateFilters(run)...)
	issues = append(issues, validateMappingsAndDefinitions(run)...)
	for _, mapping := range request.projectionConfig.RelationshipMappings {
		if mapping.EmitReverseEdge && mapping.DirectionPolicy != "normalize_forward" && mapping.DirectionPolicy != "normalize_reverse" {
			issues = append(issues, run.issue("fatal", "invalid_reverse_edge_policy", "mapping_rule", mapping.MappingRuleID, nil, map[string]any{"mapping_rule_id": mapping.MappingRuleID, "projected_direction": nil}))
		}
	}
	return issues
}

func validateFilters(run projectionWork) []ValidationIssue {
	issues := []ValidationIssue{}
	if run.Request.filters.Logic != "and" {
		issues = append(issues, run.issue("fatal", "invalid_filter", "filter", "$.filters.logic", "$.filters.logic", map[string]any{"field": "$.filters.logic", "reason_code": "unsupported_logic"}))
	}
	validate := func(predicates []filterPredicate, collection, scope string) {
		for index, predicate := range predicates {
			base := fmt.Sprintf("%s[%d]", collection, index)
			if !validFilterFieldPath(predicate.FieldPath, scope) {
				issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".field_path", map[string]any{"field": base + ".field_path", "reason_code": "invalid_field_scope"}))
			}
			switch predicate.Operator {
			case "exists":
				if predicate.HasValue {
					issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "value_forbidden"}))
				}
			case "equals", "not_equals":
				if !predicate.HasValue {
					issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "value_required"}))
				}
			case "in":
				values, ok := predicate.Value.([]any)
				if !predicate.HasValue {
					issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "value_required"}))
				} else if !ok || len(values) == 0 || !scalarArray(values) {
					issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "invalid_value_shape"}))
				}
			default:
				issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".operator", map[string]any{"field": base + ".operator", "reason_code": "invalid_operator"}))
			}
		}
	}
	validate(run.Request.filters.EntityFilters, "$.filters.entity_filters", "source_entity")
	validate(run.Request.filters.RelationshipFilters, "$.filters.relationship_filters", "source_relationship")
	return issues
}

func scalarArray(values []any) bool {
	for _, value := range values {
		switch value.(type) {
		case nil, string, bool, json.Number:
		default:
			return false
		}
	}
	return true
}

func validFilterFieldPath(path, scope string) bool {
	if strings.HasPrefix(path, "projected.") || strings.HasPrefix(path, "source_metadata.") {
		return false
	}
	parts := strings.Split(path, ".")
	if scope == "source_entity" {
		if len(parts) == 1 && (parts[0] == "source_entity_id" || parts[0] == "kind") {
			return true
		}
	} else if len(parts) == 1 && (parts[0] == "source_relationship_id" || parts[0] == "src_source_entity_id" || parts[0] == "dst_source_entity_id" || parts[0] == "kind" || parts[0] == "direction") {
		return true
	}
	return len(parts) == 2 && (parts[0] == "properties" || parts[0] == "metadata") && validPropertyKey(parts[1])
}

func validateMappingsAndDefinitions(run projectionWork) []ValidationIssue {
	request := run.Request
	issues := []ValidationIssue{}
	entityRuleIDs := map[string]bool{}
	entityKinds := map[string]string{}
	vertexKinds := map[string]bool{}
	for _, mapping := range request.projectionConfig.EntityMappings {
		if entityRuleIDs[mapping.MappingRuleID] {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_mapping_rule_id"))
		}
		entityRuleIDs[mapping.MappingRuleID] = true
		if prior := entityKinds[mapping.SourceEntityKind]; prior != "" {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_source_entity_kind_mapping"))
		}
		entityKinds[mapping.SourceEntityKind] = mapping.MappingRuleID
		vertexKinds[mapping.ProjectedVertexKind] = true
		if mapping.LabelPolicy != "mapping_only" && mapping.LabelPolicy != "preserve_source" && mapping.LabelPolicy != "mapping_then_source" {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "label_invalid"))
		}
		if mapping.InclusionFilter == nil && mapping.InclusionPredicate != "always" {
			issues = append(issues, run.issue("fatal", "invalid_filter", "filter", mapping.MappingRuleID, nil, map[string]any{"field": "$.projection_config.entity_mappings.inclusion_predicate", "reason_code": "invalid_operator"}))
		}
	}
	relationshipRuleIDs := map[string]bool{}
	relationshipKinds := map[string]string{}
	edgeKinds := map[string]bool{}
	for _, mapping := range request.projectionConfig.RelationshipMappings {
		if entityRuleIDs[mapping.MappingRuleID] || relationshipRuleIDs[mapping.MappingRuleID] {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_mapping_rule_id"))
		}
		relationshipRuleIDs[mapping.MappingRuleID] = true
		if prior := relationshipKinds[mapping.SourceRelationshipKind]; prior != "" {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_source_relationship_kind_mapping"))
		}
		relationshipKinds[mapping.SourceRelationshipKind] = mapping.MappingRuleID
		edgeKinds[mapping.ProjectedEdgeKind] = true
		if mapping.EmitReverseEdge {
			edgeKinds[mapping.ReverseEdgeKind] = true
		}
		if mapping.DirectionPolicy != "preserve" && mapping.DirectionPolicy != "normalize_forward" && mapping.DirectionPolicy != "normalize_reverse" && mapping.DirectionPolicy != "undirected" && mapping.DirectionPolicy != "bidirectional" {
			issues = append(issues, run.issue("fatal", "invalid_direction_policy", "mapping_rule", mapping.MappingRuleID, nil, map[string]any{"mapping_rule_id": mapping.MappingRuleID, "supplied_value": mapping.DirectionPolicy}))
		}
		if mapping.ReverseEdgeKindSupplied && !mapping.EmitReverseEdge {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "reverse_edge_kind_without_reverse"))
		}
		if mapping.LabelPolicy != "mapping_only" && mapping.LabelPolicy != "preserve_source" && mapping.LabelPolicy != "mapping_then_source" {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "label_invalid"))
		}
		if mapping.InclusionFilter == nil && mapping.InclusionPredicate != "always" {
			issues = append(issues, run.issue("fatal", "invalid_filter", "filter", mapping.MappingRuleID, nil, map[string]any{"field": "$.projection_config.relationship_mappings.inclusion_predicate", "reason_code": "invalid_operator"}))
		}
	}
	for _, rule := range request.projectionConfig.AggregationRules {
		if rule.TargetScope == "vertex" {
			vertexKinds[rule.ProjectedKind] = true
		} else if rule.TargetScope == "edge" {
			edgeKinds[rule.ProjectedKind] = true
		}
	}
	issues = append(issues, validateAggregationRules(run)...)
	for _, kind := range request.projectionConfig.DeclaredSourceEntityKinds {
		if entityKinds[kind] == "" {
			issues = append(issues, run.issue("fatal", "missing_entity_mapping_rule", "mapping_rule", kind, nil, map[string]any{"source_entity_kind": kind}))
		}
	}
	for _, kind := range request.projectionConfig.DeclaredSourceRelationshipKinds {
		if relationshipKinds[kind] == "" {
			issues = append(issues, run.issue("error", "missing_relationship_mapping_rule", "mapping_rule", kind, nil, map[string]any{"source_relationship_kind": kind}))
		}
	}
	propertyExpansions := map[string]string{}
	for _, definition := range request.PropertyDefinitions {
		if !validPropertyKey(definition.ProjectedKey) || !validIdentifierV2(definition.PropertyDefinitionID) {
			issues = append(issues, run.issue("fatal", "invalid_property_definition", "property_definition", definition.PropertyDefinitionID, nil, map[string]any{"property_definition_id": definition.PropertyDefinitionID, "reason_code": "invalid_projected_type"}))
		}
		if !validProjectedType(definition.ProjectedType) {
			issues = append(issues, invalidPropertyDefinitionIssue(run, definition, "invalid_projected_type"))
		}
		if definition.HasDefaultValue && (definition.DefaultValue == nil && definition.NullOutputPolicy != "emit_null" || definition.DefaultValue != nil && !valueMatchesType(definition.ProjectedType, definition.DefaultValue)) {
			issues = append(issues, invalidPropertyDefinitionIssue(run, definition, "invalid_default_value"))
		}
		if (definition.MissingBehavior == "default" || definition.SourceNullBehavior == "default") && !definition.HasDefaultValue {
			issues = append(issues, invalidPropertyDefinitionIssue(run, definition, "invalid_default_value"))
		}
		if definition.SourceNullBehavior == "emit_null" && definition.NullOutputPolicy != "emit_null" {
			issues = append(issues, invalidPropertyDefinitionIssue(run, definition, "invalid_null_policy"))
		}
		if !validProjectedMergeV2(definition.ProjectedType, definition.MergeBehavior) {
			issues = append(issues, invalidPropertyDefinitionIssue(run, definition, "invalid_merge_behavior_type"))
		}
		if !validDefinitionFieldPath(definition.SourceFieldPath, definition.TargetScope) {
			issues = append(issues, run.issue("fatal", "invalid_field_path", "projection_config", "projection_config", definition.SourceFieldPath, map[string]any{"field_path": definition.SourceFieldPath, "scope": definition.TargetScope}))
		}
		for _, kind := range expansionKinds(definition.TargetScope, definition.TargetKind, vertexKinds, edgeKinds) {
			key := definition.TargetScope + "\n" + kind + "\n" + definition.ProjectedKey
			if propertyExpansions[key] != "" {
				issues = append(issues, run.issue("fatal", "invalid_property_definition", "property_definition", definition.PropertyDefinitionID, nil, map[string]any{"property_definition_id": definition.PropertyDefinitionID, "reason_code": "duplicate_after_wildcard_expansion"}))
			}
			propertyExpansions[key] = definition.PropertyDefinitionID
		}
	}
	metadataExpansions := map[string]string{}
	for _, mapping := range request.projectionConfig.MetadataMappings {
		if reservedSystemMetadataTerminal(mapping.ProjectedMetadataKey) {
			issues = append(issues, run.issue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, nil, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "reserved_metadata_key"}))
		}
		if !validDefinitionFieldPath(mapping.SourceFieldPath, mapping.TargetScope) {
			issues = append(issues, run.issue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, mapping.SourceFieldPath, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "invalid_source_scope"}))
		}
		if !validProjectedType(mapping.ProjectedType) {
			issues = append(issues, invalidMetadataMappingIssue(run, mapping, "invalid_projected_type"))
		}
		if mapping.HasDefaultValue && (mapping.DefaultValue == nil && mapping.NullOutputPolicy != "emit_null" || mapping.DefaultValue != nil && !valueMatchesType(mapping.ProjectedType, mapping.DefaultValue)) {
			issues = append(issues, invalidMetadataMappingIssue(run, mapping, "invalid_default_value"))
		}
		if (mapping.MissingBehavior == "default" || mapping.SourceNullBehavior == "default") && !mapping.HasDefaultValue {
			issues = append(issues, invalidMetadataMappingIssue(run, mapping, "invalid_default_value"))
		}
		if !validProjectedMergeV2(mapping.ProjectedType, mapping.MergeBehavior) {
			issues = append(issues, invalidMetadataMappingIssue(run, mapping, "invalid_merge_behavior_type"))
		}
		for _, kind := range expansionKinds(mapping.TargetScope, mapping.TargetKind, vertexKinds, edgeKinds) {
			key := mapping.TargetScope + "\n" + kind + "\n" + mapping.ProjectedMetadataKey
			if metadataExpansions[key] != "" {
				issues = append(issues, run.issue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, nil, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "duplicate_after_wildcard_expansion"}))
			}
			metadataExpansions[key] = mapping.MetadataMappingID
		}
	}
	return issues
}

func validateAggregationRules(run projectionWork) []ValidationIssue {
	rules := run.Request.projectionConfig.AggregationRules
	issues := []ValidationIssue{}
	indexes := map[string]int{}
	byID := map[string]aggregationRule{}
	for index, rule := range rules {
		if _, exists := indexes[rule.AggregationRuleID]; exists {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "aggregation_cycle"))
		}
		indexes[rule.AggregationRuleID] = index
		byID[rule.AggregationRuleID] = rule
	}
	for index, rule := range rules {
		validInputScope := rule.InputScope == "source_entity" || rule.InputScope == "source_relationship" || rule.InputScope == "projected_vertex" || rule.InputScope == "projected_edge"
		if !validInputScope || rule.TargetScope != "vertex" && rule.TargetScope != "edge" {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "input_scope_invalid"))
		}
		if len(rule.GroupingKeys) == 0 || len(rule.GroupingKeys) > 32 || hasDuplicateStringsLocal(rule.GroupingKeys) {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "grouping_key_invalid"))
		}
		for _, fieldPath := range rule.GroupingKeys {
			if !validAggregationFieldPath(fieldPath, rule.InputScope) {
				issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "grouping_key_invalid"))
			}
		}
		if rule.MissingGroupingKeyBehavior != "error" && rule.MissingGroupingKeyBehavior != "exclude" && rule.MissingGroupingKeyBehavior != "use_null" {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "grouping_key_invalid"))
		}
		for projectedKey, mergeBehavior := range rule.PropertyMergeBehavior {
			matched := false
			for _, definition := range run.Request.PropertyDefinitions {
				if definition.ProjectedKey == projectedKey && propertyApplies(definition, rule.TargetScope, rule.ProjectedKind) {
					matched = true
					if !validProjectedMergeV2(definition.ProjectedType, mergeBehavior) {
						issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_merge_behavior_type"))
					}
				}
			}
			if !matched {
				issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_merge_behavior_type"))
			}
		}
		if rule.TargetScope == "vertex" {
			if rule.endpointGrouping != nil || rule.EdgeDirection != "directed" {
				issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_edge_direction"))
			}
			continue
		}
		if rule.EdgeDirection != "directed" && rule.EdgeDirection != "undirected" && rule.EdgeDirection != "bidirectional" {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_edge_direction"))
		}
		if rule.endpointGrouping == nil {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_rule_not_vertex_rule"))
			continue
		}
		endpoint := rule.endpointGrouping
		if endpoint.MissingEndpointBehavior != "error" && endpoint.MissingEndpointBehavior != "exclude" {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_endpoint_behavior"))
		}
		for _, reference := range []struct {
			id   string
			keys []string
		}{{id: endpoint.SourceVertexAggregationRuleID, keys: endpoint.SourceGroupingKeys}, {id: endpoint.DestinationVertexAggregationRuleID, keys: endpoint.DestinationGroupingKeys}} {
			referenced, exists := byID[reference.id]
			if !exists || referenced.TargetScope != "vertex" {
				issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_rule_not_vertex_rule"))
				continue
			}
			if indexes[reference.id] >= index {
				issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "dependency_on_later_rule"))
			}
			if len(reference.keys) != len(referenced.GroupingKeys) {
				issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_grouping_key_count_mismatch"))
			}
			for _, fieldPath := range reference.keys {
				if !validAggregationFieldPath(fieldPath, rule.InputScope) {
					issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_field_scope_invalid"))
				}
			}
		}
	}
	return issues
}

func invalidAggregationIssue(run projectionWork, aggregationRuleID, reason string) ValidationIssue {
	return run.issue("fatal", "invalid_aggregation_rule", "mapping_rule", aggregationRuleID, nil, map[string]any{"aggregation_rule_id": aggregationRuleID, "reason_code": reason})
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

func validAggregationFieldPath(path, inputScope string) bool {
	if !validFieldPath(path) {
		return false
	}
	switch inputScope {
	case "source_entity":
		return path == "source_entity_id" || path == "kind" || strings.HasPrefix(path, "properties.") || strings.HasPrefix(path, "metadata.")
	case "source_relationship":
		return path == "source_relationship_id" || path == "src_source_entity_id" || path == "dst_source_entity_id" || path == "kind" || path == "direction" || strings.HasPrefix(path, "properties.") || strings.HasPrefix(path, "metadata.")
	case "projected_vertex":
		return path == "kind" || path == "projected.vertex_id" || strings.HasPrefix(path, "projected.properties.") || strings.HasPrefix(path, "projected.metadata.")
	case "projected_edge":
		return path == "kind" || path == "projected.edge_id" || path == "projected.src_vertex_id" || path == "projected.dst_vertex_id" || path == "projected.direction" || strings.HasPrefix(path, "projected.properties.") || strings.HasPrefix(path, "projected.metadata.")
	default:
		return false
	}
}

func invalidPropertyDefinitionIssue(run projectionWork, definition propertyDefinition, reason string) ValidationIssue {
	return run.issue("fatal", "invalid_property_definition", "property_definition", definition.PropertyDefinitionID, nil, map[string]any{"property_definition_id": definition.PropertyDefinitionID, "reason_code": reason})
}

func invalidMetadataMappingIssue(run projectionWork, mapping metadataMapping, reason string) ValidationIssue {
	return run.issue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, nil, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": reason})
}

func invalidMappingIssue(run projectionWork, mappingRuleID, reason string) ValidationIssue {
	return run.issue("fatal", "invalid_mapping_rule", "mapping_rule", mappingRuleID, nil, map[string]any{"mapping_rule_id": mappingRuleID, "reason_code": reason})
}

func expansionKinds(scope, target string, vertexKinds, edgeKinds map[string]bool) []string {
	if target != "*" || scope == "graph_view" {
		return []string{target}
	}
	set := vertexKinds
	if scope == "edge" {
		set = edgeKinds
	}
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func validDefinitionFieldPath(path, scope string) bool {
	if !validFieldPath(path) {
		return false
	}
	if scope == "graph_view" {
		return strings.HasPrefix(path, "source_metadata.")
	}
	if scope == "vertex" {
		return path == "source_entity_id" || path == "kind" || path == "projected.vertex_id" || strings.HasPrefix(path, "properties.") || strings.HasPrefix(path, "metadata.") || strings.HasPrefix(path, "projected.properties.") || strings.HasPrefix(path, "projected.metadata.")
	}
	if scope == "edge" {
		return path == "source_relationship_id" || path == "src_source_entity_id" || path == "dst_source_entity_id" || path == "kind" || path == "direction" || path == "projected.edge_id" || path == "projected.src_vertex_id" || path == "projected.dst_vertex_id" || path == "projected.direction" || strings.HasPrefix(path, "properties.") || strings.HasPrefix(path, "metadata.") || strings.HasPrefix(path, "projected.properties.") || strings.HasPrefix(path, "projected.metadata.")
	}
	return false
}

func admittedResourceLimitIssues(run projectionWork) []ValidationIssue {
	request := run.Request
	issues := []ValidationIssue{}
	whole := func(key string, observed, limit int) {
		if observed > limit {
			issues = append(issues, run.issue("fatal", "resource_limit_exceeded", "projection_input", "projection_input", nil, map[string]any{"limit_key": key, "limit": limit, "observed": observed}))
		}
	}
	whole("maximum_source_vertices", len(request.SourceEntities), graphProjectionLimits.MaxSourceEntities)
	whole("maximum_source_edges", len(request.SourceRelationships), graphProjectionLimits.MaxSourceRelationships)
	whole("maximum_entity_filters", len(request.filters.EntityFilters), graphProjectionLimits.MaxEntityFilters)
	whole("maximum_relationship_filters", len(request.filters.RelationshipFilters), graphProjectionLimits.MaxRelationshipFilters)
	whole("maximum_entity_mappings", len(request.projectionConfig.DeclaredSourceEntityKinds), graphProjectionLimits.MaxEntityMappings)
	whole("maximum_relationship_mappings", len(request.projectionConfig.DeclaredSourceRelationshipKinds), graphProjectionLimits.MaxRelationshipMappings)
	whole("maximum_entity_mappings", len(request.projectionConfig.EntityMappings), graphProjectionLimits.MaxEntityMappings)
	whole("maximum_relationship_mappings", len(request.projectionConfig.RelationshipMappings), graphProjectionLimits.MaxRelationshipMappings)
	whole("maximum_property_definitions", len(request.PropertyDefinitions), graphProjectionLimits.MaxPropertyDefinitions)
	whole("maximum_metadata_mappings", len(request.projectionConfig.MetadataMappings), graphProjectionLimits.MaxMetadataMappings)
	whole("maximum_aggregation_definitions", len(request.projectionConfig.AggregationRules), graphProjectionLimits.MaxAggregationRules)
	whole("maximum_labels_per_object", len(request.projectionConfig.DefaultVertexLabels), graphProjectionLimits.MaxLabelsPerObject)
	whole("maximum_labels_per_object", len(request.projectionConfig.DefaultEdgeLabels), graphProjectionLimits.MaxLabelsPerObject)
	whole("maximum_property_keys", len(request.SourceMetadata), graphProjectionLimits.MaxPropertyKeys)
	for _, mapping := range request.projectionConfig.EntityMappings {
		whole("maximum_labels_per_object", len(mapping.MappingLabels), graphProjectionLimits.MaxLabelsPerObject)
		whole("maximum_property_keys", len(mapping.RequiredPropertyKeys), graphProjectionLimits.MaxPropertyKeys)
		whole("maximum_property_keys", len(mapping.OptionalPropertyKeys), graphProjectionLimits.MaxPropertyKeys)
	}
	for _, mapping := range request.projectionConfig.RelationshipMappings {
		whole("maximum_labels_per_object", len(mapping.MappingLabels), graphProjectionLimits.MaxLabelsPerObject)
		whole("maximum_property_keys", len(mapping.RequiredPropertyKeys), graphProjectionLimits.MaxPropertyKeys)
		whole("maximum_property_keys", len(mapping.OptionalPropertyKeys), graphProjectionLimits.MaxPropertyKeys)
	}
	for _, entity := range request.SourceEntities {
		issues = append(issues, sourceItemResourceLimitIssues(run, entity.SourceEntityID, len(entity.Labels), len(entity.Properties), len(entity.Metadata))...)
		issues = append(issues, sourceItemStringLimitIssues(run, entity.SourceEntityID, entity.Properties, entity.Metadata)...)
	}
	for _, relationship := range request.SourceRelationships {
		issues = append(issues, sourceItemResourceLimitIssues(run, relationship.SourceRelationshipID, len(relationship.Labels), len(relationship.Properties), len(relationship.Metadata))...)
		issues = append(issues, sourceItemStringLimitIssues(run, relationship.SourceRelationshipID, relationship.Properties, relationship.Metadata)...)
	}
	return issues
}

func sourceItemStringLimitIssues(run projectionWork, itemID string, objects ...map[string]any) []ValidationIssue {
	observed := 0
	for _, object := range objects {
		for _, value := range object {
			if length := longestPropertyString(value); length > observed {
				observed = length
			}
		}
	}
	if observed <= graphProjectionLimits.MaxStringBytes {
		return nil
	}
	return []ValidationIssue{run.issue("error", "source_item_resource_limit_exceeded", "source_item", itemID, nil, map[string]any{"source_item_id": itemID, "limit_key": "maximum_string_bytes", "limit": graphProjectionLimits.MaxStringBytes, "observed": observed})}
}

func longestPropertyString(value any) int {
	switch typed := value.(type) {
	case string:
		return len(typed)
	case []any:
		longest := 0
		for _, item := range typed {
			if length := longestPropertyString(item); length > longest {
				longest = length
			}
		}
		return longest
	default:
		return 0
	}
}

func validateLabels(run projectionWork) []ValidationIssue {
	issues := []ValidationIssue{}
	check := func(values []string, field string) {
		for _, value := range values {
			if value == "" || len(value) > graphProjectionLimits.MaxLabelBytes {
				issues = append(issues, run.issue("fatal", "invalid_input_shape", "projection_input", "projection_input", field, map[string]any{"field": field, "reason_code": "invalid_label"}))
			}
		}
	}
	check(run.Request.projectionConfig.DefaultVertexLabels, "$.projection_config.default_vertex_labels")
	check(run.Request.projectionConfig.DefaultEdgeLabels, "$.projection_config.default_edge_labels")
	for index, mapping := range run.Request.projectionConfig.EntityMappings {
		check(mapping.MappingLabels, fmt.Sprintf("$.projection_config.entity_mappings[%d].mapping_labels", index))
	}
	for index, mapping := range run.Request.projectionConfig.RelationshipMappings {
		check(mapping.MappingLabels, fmt.Sprintf("$.projection_config.relationship_mappings[%d].mapping_labels", index))
	}
	for index, entity := range run.Request.SourceEntities {
		check(entity.Labels, fmt.Sprintf("$.source_entities[%d].labels", index))
	}
	for index, relationship := range run.Request.SourceRelationships {
		check(relationship.Labels, fmt.Sprintf("$.source_relationships[%d].labels", index))
	}
	return issues
}

func sourceItemResourceLimitIssues(run projectionWork, itemID string, labelCount, propertyCount, metadataCount int) []ValidationIssue {
	issues := []ValidationIssue{}
	add := func(key string, observed, limit int) {
		if observed > limit {
			issues = append(issues, run.issue("error", "source_item_resource_limit_exceeded", "source_item", itemID, nil, map[string]any{"source_item_id": itemID, "limit_key": key, "limit": limit, "observed": observed}))
		}
	}
	add("maximum_labels_per_object", labelCount, graphProjectionLimits.MaxLabelsPerObject)
	add("maximum_property_keys", propertyCount, graphProjectionLimits.MaxPropertyKeys)
	add("maximum_property_keys", metadataCount, graphProjectionLimits.MaxPropertyKeys)
	return issues
}

func sourceItemWithinLimits(labelCount, propertyCount, metadataCount int) bool {
	return labelCount <= graphProjectionLimits.MaxLabelsPerObject &&
		propertyCount <= graphProjectionLimits.MaxPropertyKeys &&
		metadataCount <= graphProjectionLimits.MaxPropertyKeys
}

func sourceItemValuesWithinLimits(objects ...map[string]any) bool {
	for _, object := range objects {
		for _, value := range object {
			if longestPropertyString(value) > graphProjectionLimits.MaxStringBytes {
				return false
			}
		}
	}
	return true
}

func projectedOutputLimitIssues(run projectionWork, vertices []Vertex, edges []Edge) []ValidationIssue {
	issues := []ValidationIssue{}
	add := func(key string, observed, limit int) {
		if observed > limit {
			issues = append(issues, run.issue("fatal", "projected_output_limit_exceeded", "graph_view", run.GraphViewID, nil, map[string]any{"limit_key": key, "limit": limit, "observed": observed}))
		}
	}
	add("maximum_projected_vertices", len(vertices), graphProjectionLimits.MaxProjectedVertices)
	add("maximum_projected_edges", len(edges), graphProjectionLimits.MaxProjectedEdges)
	for _, vertex := range vertices {
		add("maximum_property_keys", len(vertex.Properties), graphProjectionLimits.MaxPropertyKeys)
		add("maximum_property_keys", len(vertex.Metadata.MappedMetadata), graphProjectionLimits.MaxPropertyKeys)
	}
	for _, edge := range edges {
		add("maximum_property_keys", len(edge.Properties), graphProjectionLimits.MaxPropertyKeys)
		add("maximum_property_keys", len(edge.Metadata.MappedMetadata), graphProjectionLimits.MaxPropertyKeys)
	}
	return issues
}
