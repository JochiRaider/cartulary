package graphprojection

import (
	"encoding/json"
	"fmt"
	"sort"
)

func validateAdmittedRequest(run *projectionWork) {
	if run.issues.err != nil {
		return
	}
	request := run.Request
	identifierValid := validIdentifierV2
	scalar := func(valid bool, field string) {
		if !valid {
			run.recordIssue("fatal", "invalid_input_shape", "projection_input", "projection_input", field, map[string]any{"field": field, "reason_code": "scalar_contract_violation"})
		}
	}
	scalar(identifierValid(request.SourceSnapshotID), "$.source_snapshot_id")
	for index, entity := range request.SourceEntities {
		if run.issues.err != nil {
			return
		}
		base := fmt.Sprintf("$.source_entities[%d]", index)
		scalar(identifierValid(entity.SourceEntityID), base+".source_entity_id")
		scalar(identifierValid(entity.SourceEntityKind), base+".source_entity_kind")
	}
	for index, relationship := range request.SourceRelationships {
		if run.issues.err != nil {
			return
		}
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
		run.recordIssue("fatal", "invalid_projection_config", "projection_config", "projection_config", "$.projection_config", map[string]any{"field": "$.projection_config", "reason_code": "empty_kind_registry_not_allowed"})
	}
	checkDuplicates := func(values []string, collection string) {
		seen := map[string]struct{}{}
		for _, value := range values {
			if run.issues.err != nil {
				return
			}
			if _, ok := seen[value]; ok {
				run.recordIssue("fatal", "duplicate_identifier", "projection_input", "projection_input", nil, map[string]any{"identifier_value": value, "collection": collection})
			}
			seen[value] = struct{}{}
		}
	}
	checkDuplicates(entityIDs(request.SourceEntities), "$.source_entities")
	checkDuplicates(relationshipIDs(request.SourceRelationships), "$.source_relationships")
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
	admittedResourceLimitIssues(run)
	validateLabels(run)
	validateFilters(run)
	validateMappingsAndDefinitions(run)
	for _, mapping := range request.projectionConfig.RelationshipMappings {
		if run.issues.err != nil {
			return
		}
		if mapping.EmitReverseEdge && mapping.DirectionPolicy != "normalize_forward" && mapping.DirectionPolicy != "normalize_reverse" {
			run.recordIssue("fatal", "invalid_reverse_edge_policy", "mapping_rule", mapping.MappingRuleID, nil, map[string]any{"mapping_rule_id": mapping.MappingRuleID, "projected_direction": nil})
		}
	}
}

func validateFilters(run *projectionWork) {
	if run.issues.err != nil {
		return
	}
	if run.Request.filters.Logic != "and" {
		run.recordIssue("fatal", "invalid_filter", "filter", "$.filters.logic", "$.filters.logic", map[string]any{"field": "$.filters.logic", "reason_code": "unsupported_logic"})
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
	switch predicate.Operator {
	case "exists":
		if predicate.HasValue {
			run.recordIssue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "value_forbidden"})
		}
	case "equals", "not_equals":
		if !predicate.HasValue {
			run.recordIssue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "value_required"})
		}
	case "in":
		values, ok := predicate.Value.([]any)
		if !predicate.HasValue {
			run.recordIssue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "value_required"})
		} else if !ok || len(values) == 0 || !scalarArray(values) {
			run.recordIssue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "invalid_value_shape"})
		}
	default:
		run.recordIssue("fatal", "invalid_filter", "filter", base, base+".operator", map[string]any{"field": base + ".operator", "reason_code": "invalid_operator"})
	}
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

func validateMappingsAndDefinitions(run *projectionWork) {
	if run.issues.err != nil {
		return
	}
	request := run.Request
	entityRuleIDs := map[string]bool{}
	entityKinds := map[string]string{}
	vertexKinds := map[string]bool{}
	for _, mapping := range request.projectionConfig.EntityMappings {
		if run.issues.err != nil {
			return
		}
		if entityRuleIDs[mapping.MappingRuleID] {
			invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_mapping_rule_id")
		}
		entityRuleIDs[mapping.MappingRuleID] = true
		if prior := entityKinds[mapping.SourceEntityKind]; prior != "" {
			invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_source_entity_kind_mapping")
		}
		entityKinds[mapping.SourceEntityKind] = mapping.MappingRuleID
		vertexKinds[mapping.ProjectedVertexKind] = true
		if mapping.LabelPolicy != "mapping_only" && mapping.LabelPolicy != "preserve_source" && mapping.LabelPolicy != "mapping_then_source" {
			invalidMappingIssue(run, mapping.MappingRuleID, "label_invalid")
		}
		if mapping.InclusionFilter != nil {
			validateFilterPredicate(run, *mapping.InclusionFilter, "$.projection_config.entity_mappings."+mapping.MappingRuleID+".inclusion_predicate", "source_entity")
		}
		if mapping.InclusionFilter == nil && mapping.InclusionPredicate != "always" {
			run.recordIssue("fatal", "invalid_filter", "filter", mapping.MappingRuleID, nil, map[string]any{"field": "$.projection_config.entity_mappings.inclusion_predicate", "reason_code": "invalid_operator"})
		}
	}
	relationshipRuleIDs := map[string]bool{}
	relationshipKinds := map[string]string{}
	edgeKinds := map[string]bool{}
	for _, mapping := range request.projectionConfig.RelationshipMappings {
		if run.issues.err != nil {
			return
		}
		if entityRuleIDs[mapping.MappingRuleID] || relationshipRuleIDs[mapping.MappingRuleID] {
			invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_mapping_rule_id")
		}
		relationshipRuleIDs[mapping.MappingRuleID] = true
		if prior := relationshipKinds[mapping.SourceRelationshipKind]; prior != "" {
			invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_source_relationship_kind_mapping")
		}
		relationshipKinds[mapping.SourceRelationshipKind] = mapping.MappingRuleID
		edgeKinds[mapping.ProjectedEdgeKind] = true
		if mapping.EmitReverseEdge {
			edgeKinds[mapping.ReverseEdgeKind] = true
		}
		if mapping.DirectionPolicy != "preserve" && mapping.DirectionPolicy != "normalize_forward" && mapping.DirectionPolicy != "normalize_reverse" && mapping.DirectionPolicy != "undirected" && mapping.DirectionPolicy != "bidirectional" {
			run.recordIssue("fatal", "invalid_direction_policy", "mapping_rule", mapping.MappingRuleID, nil, map[string]any{"mapping_rule_id": mapping.MappingRuleID, "supplied_value": mapping.DirectionPolicy})
		}
		if mapping.ReverseEdgeKindSupplied && !mapping.EmitReverseEdge {
			invalidMappingIssue(run, mapping.MappingRuleID, "reverse_edge_kind_without_reverse")
		}
		if mapping.LabelPolicy != "mapping_only" && mapping.LabelPolicy != "preserve_source" && mapping.LabelPolicy != "mapping_then_source" {
			invalidMappingIssue(run, mapping.MappingRuleID, "label_invalid")
		}
		if mapping.InclusionFilter != nil {
			validateFilterPredicate(run, *mapping.InclusionFilter, "$.projection_config.relationship_mappings."+mapping.MappingRuleID+".inclusion_predicate", "source_relationship")
		}
		if mapping.InclusionFilter == nil && mapping.InclusionPredicate != "always" {
			run.recordIssue("fatal", "invalid_filter", "filter", mapping.MappingRuleID, nil, map[string]any{"field": "$.projection_config.relationship_mappings.inclusion_predicate", "reason_code": "invalid_operator"})
		}
	}
	for _, rule := range request.projectionConfig.AggregationRules {
		if run.issues.err != nil {
			return
		}
		if rule.TargetScope == "vertex" {
			vertexKinds[rule.ProjectedKind] = true
		} else if rule.TargetScope == "edge" {
			edgeKinds[rule.ProjectedKind] = true
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
	definitionContexts := configuredDefinitionReadContexts(request.projectionConfig)
	for _, definition := range request.PropertyDefinitions {
		if run.issues.err != nil {
			return
		}
		if !validPropertyKey(definition.ProjectedKey) || !validIdentifierV2(definition.PropertyDefinitionID) {
			run.recordIssue("fatal", "invalid_property_definition", "property_definition", definition.PropertyDefinitionID, nil, map[string]any{"property_definition_id": definition.PropertyDefinitionID, "reason_code": "invalid_projected_type"})
		}
		if !validProjectedType(definition.ProjectedType) {
			invalidPropertyDefinitionIssue(run, definition, "invalid_projected_type")
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
		if !definitionContexts.allows(definition.SourceFieldPath, definition.TargetScope, definition.TargetKind) {
			run.recordIssue("fatal", "invalid_field_path", "projection_config", "projection_config", definition.SourceFieldPath, map[string]any{"field_path": definition.SourceFieldPath, "scope": definition.TargetScope})
		}
		for _, kind := range expansionKinds(definition.TargetScope, definition.TargetKind, vertexKinds, edgeKinds) {
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
		if reservedSystemMetadataTerminal(mapping.ProjectedMetadataKey) {
			run.recordIssue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, nil, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "reserved_metadata_key"})
		}
		if !definitionContexts.allows(mapping.SourceFieldPath, mapping.TargetScope, mapping.TargetKind) {
			run.recordIssue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, mapping.SourceFieldPath, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "invalid_source_scope"})
		}
		if !validProjectedType(mapping.ProjectedType) {
			invalidMetadataMappingIssue(run, mapping, "invalid_projected_type")
		}
		if mapping.HasDefaultValue && (mapping.DefaultValue == nil && mapping.NullOutputPolicy != "emit_null" || mapping.DefaultValue != nil && !valueMatchesType(mapping.ProjectedType, mapping.DefaultValue)) {
			invalidMetadataMappingIssue(run, mapping, "invalid_default_value")
		}
		if (mapping.MissingBehavior == "default" || mapping.SourceNullBehavior == "default") && !mapping.HasDefaultValue {
			invalidMetadataMappingIssue(run, mapping, "invalid_default_value")
		}
		if !validProjectedMergeV2(mapping.ProjectedType, mapping.MergeBehavior) {
			invalidMetadataMappingIssue(run, mapping, "invalid_merge_behavior_type")
		}
		for _, kind := range expansionKinds(mapping.TargetScope, mapping.TargetKind, vertexKinds, edgeKinds) {
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
		validInputScope := rule.InputScope == "source_entity" || rule.InputScope == "source_relationship" || rule.InputScope == "projected_vertex" || rule.InputScope == "projected_edge"
		if !validInputScope || rule.TargetScope != "vertex" && rule.TargetScope != "edge" {
			invalidAggregationIssue(run, rule.AggregationRuleID, "input_scope_invalid")
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
		if rule.MissingGroupingKeyBehavior != "error" && rule.MissingGroupingKeyBehavior != "exclude" && rule.MissingGroupingKeyBehavior != "use_null" {
			invalidAggregationIssue(run, rule.AggregationRuleID, "grouping_key_invalid")
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
			if rule.endpointGrouping != nil || rule.EdgeDirection != "directed" {
				invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_edge_direction")
			}
			continue
		}
		if rule.EdgeDirection != "directed" && rule.EdgeDirection != "undirected" && rule.EdgeDirection != "bidirectional" {
			invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_edge_direction")
		}
		if rule.endpointGrouping == nil {
			invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_rule_not_vertex_rule")
			continue
		}
		endpoint := rule.endpointGrouping
		if endpoint.MissingEndpointBehavior != "error" && endpoint.MissingEndpointBehavior != "exclude" {
			invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_endpoint_behavior")
		}
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

func invalidMappingIssue(run *projectionWork, mappingRuleID, reason string) {
	run.recordIssue("fatal", "invalid_mapping_rule", "mapping_rule", mappingRuleID, nil, map[string]any{"mapping_rule_id": mappingRuleID, "reason_code": reason})
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

func admittedResourceLimitIssues(run *projectionWork) {
	if run.issues.err != nil {
		return
	}
	request := run.Request
	whole := func(key string, observed, limit int) {
		if observed > limit {
			run.recordIssue("fatal", "resource_limit_exceeded", "projection_input", "projection_input", nil, map[string]any{"limit_key": key, "limit": limit, "observed": observed})
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
		if run.issues.err != nil {
			return
		}
		whole("maximum_labels_per_object", len(mapping.MappingLabels), graphProjectionLimits.MaxLabelsPerObject)
		whole("maximum_property_keys", len(mapping.RequiredPropertyKeys), graphProjectionLimits.MaxPropertyKeys)
		whole("maximum_property_keys", len(mapping.OptionalPropertyKeys), graphProjectionLimits.MaxPropertyKeys)
	}
	for _, mapping := range request.projectionConfig.RelationshipMappings {
		if run.issues.err != nil {
			return
		}
		whole("maximum_labels_per_object", len(mapping.MappingLabels), graphProjectionLimits.MaxLabelsPerObject)
		whole("maximum_property_keys", len(mapping.RequiredPropertyKeys), graphProjectionLimits.MaxPropertyKeys)
		whole("maximum_property_keys", len(mapping.OptionalPropertyKeys), graphProjectionLimits.MaxPropertyKeys)
	}
	for _, entity := range request.SourceEntities {
		if run.issues.err != nil {
			return
		}
		sourceItemResourceLimitIssues(run, entity.SourceEntityID, len(entity.Labels), len(entity.Properties), len(entity.Metadata))
		sourceItemStringLimitIssues(run, entity.SourceEntityID, entity.Properties, entity.Metadata)
	}
	for _, relationship := range request.SourceRelationships {
		if run.issues.err != nil {
			return
		}
		sourceItemResourceLimitIssues(run, relationship.SourceRelationshipID, len(relationship.Labels), len(relationship.Properties), len(relationship.Metadata))
		sourceItemStringLimitIssues(run, relationship.SourceRelationshipID, relationship.Properties, relationship.Metadata)
	}
}

func sourceItemStringLimitIssues(run *projectionWork, itemID string, objects ...map[string]any) {
	if run.issues.err != nil {
		return
	}
	observed := 0
	for _, object := range objects {
		if run.issues.err != nil {
			return
		}
		for _, value := range object {
			if run.issues.err != nil {
				return
			}
			if length := longestPropertyString(value); length > observed {
				observed = length
			}
		}
	}
	if observed <= graphProjectionLimits.MaxStringBytes {
		return
	}
	run.recordIssue("error", "source_item_resource_limit_exceeded", "source_item", itemID, nil, map[string]any{"source_item_id": itemID, "limit_key": "maximum_string_bytes", "limit": graphProjectionLimits.MaxStringBytes, "observed": observed})
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

func validateLabels(run *projectionWork) {
	if run.issues.err != nil {
		return
	}
	check := func(values []string, field string) {
		for _, value := range values {
			if run.issues.err != nil {
				return
			}
			if value == "" || len(value) > graphProjectionLimits.MaxLabelBytes {
				run.recordIssue("fatal", "invalid_input_shape", "projection_input", "projection_input", field, map[string]any{"field": field, "reason_code": "invalid_label"})
			}
		}
	}
	check(run.Request.projectionConfig.DefaultVertexLabels, "$.projection_config.default_vertex_labels")
	check(run.Request.projectionConfig.DefaultEdgeLabels, "$.projection_config.default_edge_labels")
	for index, mapping := range run.Request.projectionConfig.EntityMappings {
		if run.issues.err != nil {
			return
		}
		check(mapping.MappingLabels, fmt.Sprintf("$.projection_config.entity_mappings[%d].mapping_labels", index))
	}
	for index, mapping := range run.Request.projectionConfig.RelationshipMappings {
		if run.issues.err != nil {
			return
		}
		check(mapping.MappingLabels, fmt.Sprintf("$.projection_config.relationship_mappings[%d].mapping_labels", index))
	}
	for index, entity := range run.Request.SourceEntities {
		if run.issues.err != nil {
			return
		}
		check(entity.Labels, fmt.Sprintf("$.source_entities[%d].labels", index))
	}
	for index, relationship := range run.Request.SourceRelationships {
		if run.issues.err != nil {
			return
		}
		check(relationship.Labels, fmt.Sprintf("$.source_relationships[%d].labels", index))
	}
}

func sourceItemResourceLimitIssues(run *projectionWork, itemID string, labelCount, propertyCount, metadataCount int) {
	if run.issues.err != nil {
		return
	}
	add := func(key string, observed, limit int) {
		if observed > limit {
			run.recordIssue("error", "source_item_resource_limit_exceeded", "source_item", itemID, nil, map[string]any{"source_item_id": itemID, "limit_key": key, "limit": limit, "observed": observed})
		}
	}
	add("maximum_labels_per_object", labelCount, graphProjectionLimits.MaxLabelsPerObject)
	add("maximum_property_keys", propertyCount, graphProjectionLimits.MaxPropertyKeys)
	add("maximum_property_keys", metadataCount, graphProjectionLimits.MaxPropertyKeys)
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
