package graphprojection

func entityMappingsObject(mappings []EntityMapping) []any {
	out := make([]any, 0, len(mappings))
	for _, mapping := range mappings {
		out = append(out, canonicalFields(
			canonicalMember{Name: "mapping_rule_id", Value: mapping.MappingRuleID},
			canonicalMember{Name: "source_entity_kind", Value: mapping.SourceEntityKind},
			canonicalMember{Name: "projected_vertex_kind", Value: mapping.ProjectedVertexKind},
			canonicalMember{Name: "inclusion_predicate", Value: mapping.InclusionPredicate},
			canonicalMember{Name: "label_policy", Value: mapping.LabelPolicy},
			canonicalMember{Name: "mapping_labels", Value: mapping.MappingLabels},
			canonicalMember{Name: "required_property_keys", Value: mapping.RequiredPropertyKeys},
			canonicalMember{Name: "optional_property_keys", Value: mapping.OptionalPropertyKeys},
		))
	}
	return out
}

func relationshipMappingsObject(mappings []RelationshipMapping) []any {
	out := make([]any, 0, len(mappings))
	for _, mapping := range mappings {
		out = append(out, canonicalFields(
			canonicalMember{Name: "mapping_rule_id", Value: mapping.MappingRuleID},
			canonicalMember{Name: "source_relationship_kind", Value: mapping.SourceRelationshipKind},
			canonicalMember{Name: "projected_edge_kind", Value: mapping.ProjectedEdgeKind},
			canonicalMember{Name: "inclusion_predicate", Value: mapping.InclusionPredicate},
			canonicalMember{Name: "direction_policy", Value: mapping.DirectionPolicy},
			canonicalMember{Name: "emit_reverse_edge", Value: mapping.EmitReverseEdge},
			canonicalMember{Name: "reverse_edge_kind", Value: mapping.ReverseEdgeKind},
			canonicalMember{Name: "label_policy", Value: mapping.LabelPolicy},
			canonicalMember{Name: "mapping_labels", Value: mapping.MappingLabels},
			canonicalMember{Name: "required_property_keys", Value: mapping.RequiredPropertyKeys},
			canonicalMember{Name: "optional_property_keys", Value: mapping.OptionalPropertyKeys},
		))
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
		entry := canonicalFields(
			canonicalMember{Name: "property_definition_id", Value: definition.PropertyDefinitionID},
			canonicalMember{Name: "target_scope", Value: definition.TargetScope},
			canonicalMember{Name: "target_kind", Value: definition.TargetKind},
			canonicalMember{Name: "source_field_path", Value: definition.SourceFieldPath},
			canonicalMember{Name: "projected_key", Value: definition.ProjectedKey},
			canonicalMember{Name: "projected_type", Value: definition.ProjectedType},
			canonicalMember{Name: "required", Value: definition.Required},
			canonicalMember{Name: "missing_behavior", Value: definition.MissingBehavior},
			canonicalMember{Name: "source_null_behavior", Value: definition.SourceNullBehavior},
			canonicalMember{Name: "null_output_policy", Value: definition.NullOutputPolicy},
			canonicalMember{Name: "merge_behavior", Value: definition.MergeBehavior},
		)
		if definition.HasDefaultValue {
			entry = append(entry, canonicalMember{Name: "default_value", Value: definition.DefaultValue})
		}
		out = append(out, entry)
	}
	return out
}

func sourceEntitiesObject(entities []SourceEntity) []any {
	out := make([]any, 0, len(entities))
	for _, entity := range entities {
		out = append(out, canonicalFields(
			canonicalMember{Name: "source_entity_id", Value: entity.SourceEntityID},
			canonicalMember{Name: "source_entity_kind", Value: entity.SourceEntityKind},
			canonicalMember{Name: "properties", Value: entity.Properties},
			canonicalMember{Name: "metadata", Value: entity.Metadata},
			canonicalMember{Name: "labels", Value: entity.Labels},
		))
	}
	return out
}

func sourceRelationshipsObject(relationships []SourceRelationship) []any {
	out := make([]any, 0, len(relationships))
	for _, relationship := range relationships {
		out = append(out, canonicalFields(
			canonicalMember{Name: "source_relationship_id", Value: relationship.SourceRelationshipID},
			canonicalMember{Name: "source_relationship_kind", Value: relationship.SourceRelationshipKind},
			canonicalMember{Name: "src_source_entity_id", Value: relationship.SrcSourceEntityID},
			canonicalMember{Name: "dst_source_entity_id", Value: relationship.DstSourceEntityID},
			canonicalMember{Name: "direction", Value: relationship.Direction},
			canonicalMember{Name: "properties", Value: relationship.Properties},
			canonicalMember{Name: "metadata", Value: relationship.Metadata},
			canonicalMember{Name: "labels", Value: relationship.Labels},
		))
	}
	return out
}

func filtersObject(filters Filters) canonicalObject {
	return canonicalFields(
		canonicalMember{Name: "entity_filters", Value: filterPredicatesObject(filters.EntityFilters)},
		canonicalMember{Name: "relationship_filters", Value: filterPredicatesObject(filters.RelationshipFilters)},
		canonicalMember{Name: "logic", Value: filters.Logic},
	)
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
