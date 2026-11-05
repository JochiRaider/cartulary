package graphprojection

func entityMappingsObject(mappings []entityMapping) []any {
	out := make([]any, 0, len(mappings))
	for _, mapping := range mappings {
		inclusion := any(mapping.InclusionPredicate)
		if mapping.InclusionFilter != nil {
			inclusion = filterPredicateObject(*mapping.InclusionFilter)
		}
		out = append(out, map[string]any{
			"mapping_rule_id":        mapping.MappingRuleID,
			"source_entity_kind":     mapping.SourceEntityKind,
			"projected_vertex_kind":  mapping.ProjectedVertexKind,
			"inclusion_predicate":    inclusion,
			"label_policy":           mapping.LabelPolicy,
			"mapping_labels":         mapping.MappingLabels,
			"required_property_keys": mapping.RequiredPropertyKeys,
			"optional_property_keys": mapping.OptionalPropertyKeys,
		})
	}
	return out
}

func relationshipMappingsObject(mappings []relationshipMapping) []any {
	out := make([]any, 0, len(mappings))
	for _, mapping := range mappings {
		inclusion := any(mapping.InclusionPredicate)
		if mapping.InclusionFilter != nil {
			inclusion = filterPredicateObject(*mapping.InclusionFilter)
		}
		entry := map[string]any{
			"mapping_rule_id":          mapping.MappingRuleID,
			"source_relationship_kind": mapping.SourceRelationshipKind,
			"projected_edge_kind":      mapping.ProjectedEdgeKind,
			"inclusion_predicate":      inclusion,
			"direction_policy":         mapping.DirectionPolicy,
			"emit_reverse_edge":        mapping.EmitReverseEdge,
			"label_policy":             mapping.LabelPolicy,
			"mapping_labels":           mapping.MappingLabels,
			"required_property_keys":   mapping.RequiredPropertyKeys,
			"optional_property_keys":   mapping.OptionalPropertyKeys,
		}
		if mapping.ReverseEdgeKindSupplied {
			entry["reverse_edge_kind"] = mapping.ReverseEdgeKind
		}
		out = append(out, entry)
	}
	return out
}

func metadataMappingsObject(mappings []metadataMapping) []any {
	out := make([]any, 0, len(mappings))
	for _, mapping := range mappings {
		entry := map[string]any{
			"metadata_mapping_id":    mapping.MetadataMappingID,
			"target_scope":           mapping.TargetScope,
			"target_kind":            mapping.TargetKind,
			"source_field_path":      mapping.SourceFieldPath,
			"projected_metadata_key": mapping.ProjectedMetadataKey,
			"projected_type":         mapping.ProjectedType,
			"required":               mapping.Required,
			"missing_behavior":       mapping.MissingBehavior,
			"source_null_behavior":   mapping.SourceNullBehavior,
			"null_output_policy":     mapping.NullOutputPolicy,
			"merge_behavior":         mapping.MergeBehavior,
		}
		if mapping.HasDefaultValue {
			entry["default_value"] = mapping.DefaultValue
		}
		out = append(out, entry)
	}
	return out
}

func aggregationRulesObject(rules []aggregationRule) []any {
	out := make([]any, 0, len(rules))
	for _, rule := range rules {
		entry := map[string]any{
			"aggregation_rule_id":           rule.AggregationRuleID,
			"target_scope":                  rule.TargetScope,
			"input_scope":                   rule.InputScope,
			"input_kind":                    rule.InputKind,
			"projected_kind":                rule.ProjectedKind,
			"grouping_keys":                 rule.GroupingKeys,
			"missing_grouping_key_behavior": rule.MissingGroupingKeyBehavior,
			"property_merge_behavior":       rule.PropertyMergeBehavior,
			"edge_direction":                rule.EdgeDirection,
		}
		if rule.endpointGrouping != nil {
			entry["endpoint_grouping"] = map[string]any{
				"src_vertex_aggregation_rule_id": rule.endpointGrouping.SourceVertexAggregationRuleID,
				"src_grouping_keys":              rule.endpointGrouping.SourceGroupingKeys,
				"dst_vertex_aggregation_rule_id": rule.endpointGrouping.DestinationVertexAggregationRuleID,
				"dst_grouping_keys":              rule.endpointGrouping.DestinationGroupingKeys,
				"missing_endpoint_behavior":      rule.endpointGrouping.MissingEndpointBehavior,
			}
		} else {
			entry["endpoint_grouping"] = nil
		}
		out = append(out, entry)
	}
	return out
}

func propertyDefinitionsObject(definitions []propertyDefinition) []any {
	out := make([]any, 0, len(definitions))
	for _, definition := range definitions {
		entry := map[string]any{
			"property_definition_id": definition.PropertyDefinitionID,
			"target_scope":           definition.TargetScope,
			"target_kind":            definition.TargetKind,
			"source_field_path":      definition.SourceFieldPath,
			"projected_key":          definition.ProjectedKey,
			"projected_type":         definition.ProjectedType,
			"required":               definition.Required,
			"missing_behavior":       definition.MissingBehavior,
			"source_null_behavior":   definition.SourceNullBehavior,
			"null_output_policy":     definition.NullOutputPolicy,
			"merge_behavior":         definition.MergeBehavior,
		}
		if definition.HasDefaultValue {
			entry["default_value"] = definition.DefaultValue
		}
		out = append(out, entry)
	}
	return out
}

func sourceEntitiesObject(entities []sourceEntity) []any {
	out := make([]any, 0, len(entities))
	for _, entity := range entities {
		out = append(out, map[string]any{
			"source_entity_id":   entity.SourceEntityID,
			"source_entity_kind": entity.SourceEntityKind,
			"properties":         entity.Properties,
			"metadata":           entity.Metadata,
			"labels":             entity.Labels,
		})
	}
	return out
}

func sourceRelationshipsObject(relationships []sourceRelationship) []any {
	out := make([]any, 0, len(relationships))
	for _, relationship := range relationships {
		out = append(out, map[string]any{
			"source_relationship_id":   relationship.SourceRelationshipID,
			"source_relationship_kind": relationship.SourceRelationshipKind,
			"src_source_entity_id":     relationship.SrcSourceEntityID,
			"dst_source_entity_id":     relationship.DstSourceEntityID,
			"direction":                relationship.Direction,
			"properties":               relationship.Properties,
			"metadata":                 relationship.Metadata,
			"labels":                   relationship.Labels,
		})
	}
	return out
}

func filtersObject(filters filters) map[string]any {
	return map[string]any{
		"entity_filters":       filterPredicatesObject(filters.EntityFilters),
		"relationship_filters": filterPredicatesObject(filters.RelationshipFilters),
		"logic":                filters.Logic,
	}
}

func filterPredicatesObject(predicates []filterPredicate) []any {
	out := make([]any, 0, len(predicates))
	for _, predicate := range predicates {
		out = append(out, filterPredicateObject(predicate))
	}
	return out
}

func filterPredicateObject(predicate filterPredicate) map[string]any {
	entry := map[string]any{
		"field_path":         predicate.FieldPath,
		"operator":           predicate.Operator,
		"include_if_missing": predicate.IncludeIfMissing,
	}
	if predicate.HasValue {
		entry["value"] = predicate.Value
	}
	return entry
}
