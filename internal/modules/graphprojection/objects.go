package graphprojection

func entityMappingsObject(mappings []EntityMapping) []any {
	out := make([]any, 0, len(mappings))
	for _, mapping := range mappings {
		out = append(out, map[string]any{
			"mapping_rule_id":        mapping.MappingRuleID,
			"source_entity_kind":     mapping.SourceEntityKind,
			"projected_vertex_kind":  mapping.ProjectedVertexKind,
			"inclusion_predicate":    mapping.InclusionPredicate,
			"label_policy":           mapping.LabelPolicy,
			"mapping_labels":         mapping.MappingLabels,
			"required_property_keys": mapping.RequiredPropertyKeys,
			"optional_property_keys": mapping.OptionalPropertyKeys,
		})
	}
	return out
}

func relationshipMappingsObject(mappings []RelationshipMapping) []any {
	out := make([]any, 0, len(mappings))
	for _, mapping := range mappings {
		out = append(out, map[string]any{
			"mapping_rule_id":          mapping.MappingRuleID,
			"source_relationship_kind": mapping.SourceRelationshipKind,
			"projected_edge_kind":      mapping.ProjectedEdgeKind,
			"inclusion_predicate":      mapping.InclusionPredicate,
			"direction_policy":         mapping.DirectionPolicy,
			"emit_reverse_edge":        mapping.EmitReverseEdge,
			"reverse_edge_kind":        mapping.ReverseEdgeKind,
			"label_policy":             mapping.LabelPolicy,
			"mapping_labels":           mapping.MappingLabels,
			"required_property_keys":   mapping.RequiredPropertyKeys,
			"optional_property_keys":   mapping.OptionalPropertyKeys,
		})
	}
	return out
}

func metadataMappingsObject(mappings []MetadataMapping) []any {
	out := make([]any, 0, len(mappings))
	for _, mapping := range mappings {
		out = append(out, map[string]any{
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
		})
	}
	return out
}

func aggregationRulesObject(rules []AggregationRule) []any {
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
		if rule.EndpointGrouping != nil {
			entry["endpoint_grouping"] = map[string]any{
				"src_vertex_aggregation_rule_id": rule.EndpointGrouping.SourceVertexAggregationRuleID,
				"src_grouping_keys":              rule.EndpointGrouping.SourceGroupingKeys,
				"dst_vertex_aggregation_rule_id": rule.EndpointGrouping.DestinationVertexAggregationRuleID,
				"dst_grouping_keys":              rule.EndpointGrouping.DestinationGroupingKeys,
				"missing_endpoint_behavior":      rule.EndpointGrouping.MissingEndpointBehavior,
			}
		} else {
			entry["endpoint_grouping"] = nil
		}
		out = append(out, entry)
	}
	return out
}

func propertyDefinitionsObject(definitions []PropertyDefinition) []any {
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

func sourceEntitiesObject(entities []SourceEntity) []any {
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

func sourceRelationshipsObject(relationships []SourceRelationship) []any {
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

func filtersObject(filters Filters) map[string]any {
	return map[string]any{
		"entity_filters":       filterPredicatesObject(filters.EntityFilters),
		"relationship_filters": filterPredicatesObject(filters.RelationshipFilters),
		"logic":                filters.Logic,
	}
}

func filterPredicatesObject(predicates []FilterPredicate) []any {
	out := make([]any, 0, len(predicates))
	for _, predicate := range predicates {
		out = append(out, map[string]any{
			"field_path": predicate.FieldPath,
			"operator":   predicate.Operator,
			"value":      predicate.Value,
		})
	}
	return out
}
