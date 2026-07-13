package graphprojection

func graphViewResource(view GraphView) map[string]any {
	vertices := make([]any, 0, len(view.Vertices))
	for _, vertex := range view.Vertices {
		vertices = append(vertices, vertexResource(vertex))
	}
	edges := make([]any, 0, len(view.Edges))
	for _, edge := range view.Edges {
		edges = append(edges, edgeResource(edge))
	}
	return map[string]any{
		"projection_schema_id": view.ProjectionSchemaID,
		"graph_view_id":        view.GraphViewID,
		"graph_view_key":       view.GraphViewKey,
		"projection_run_id":    view.ProjectionRunID,
		"source_snapshot_id":   view.SourceSnapshotID,
		"projection_version":   view.ProjectionVersion,
		"generated_at":         view.GeneratedAt,
		"state":                string(view.State),
		"properties":           nonNilMap(view.Properties),
		"metadata":             graphMetadataResource(view.Metadata),
		"schema_registry":      schemaRegistryResource(view.SchemaRegistry),
		"vertices":             vertices,
		"edges":                edges,
		"validation_summary":   validationSummaryResource(view.ValidationSummary),
		"consumer_capabilities": map[string]any{
			"query_shapes":                       nonNilStrings(view.ConsumerCapabilities.QueryShapes),
			"supports_direct_vertex_lookup":      view.ConsumerCapabilities.SupportsDirectVertexLookup,
			"supports_direct_edge_lookup":        view.ConsumerCapabilities.SupportsDirectEdgeLookup,
			"supports_breadth_first_traversal":   view.ConsumerCapabilities.SupportsBreadthFirstTraversal,
			"supports_alternate_traversal_order": nonNilStrings(view.ConsumerCapabilities.SupportsAlternateTraversalOrder),
			"max_traversal_depth":                view.ConsumerCapabilities.MaxTraversalDepth,
			"max_traversal_seed_vertices":        view.ConsumerCapabilities.MaxTraversalSeedVertices,
			"max_kind_filters":                   view.ConsumerCapabilities.MaxKindFilters,
		},
	}
}

func graphMetadataResource(metadata GraphMetadata) map[string]any {
	return map[string]any{
		"previous_projection_run_id": nullableString(metadata.PreviousProjectionRunID),
		"projection_config_digest":   metadata.ProjectionConfigDigest,
		"projection_source_digest":   metadata.ProjectionSourceDigest,
		"mapped_metadata":            nonNilMap(metadata.MappedMetadata),
		"invalidation":               invalidationResource(metadata.Invalidation),
	}
}

func vertexResource(vertex Vertex) map[string]any {
	var sourceRef any
	if vertex.SourceEntityRef != nil {
		sourceRef = map[string]any{
			"source_entity_id":   vertex.SourceEntityRef.SourceEntityID,
			"source_entity_kind": vertex.SourceEntityRef.SourceEntityKind,
			"mapping_rule_id":    vertex.SourceEntityRef.MappingRuleID,
		}
	}
	refs := make([]any, 0, len(vertex.Metadata.AggregationSourceRefs))
	for _, ref := range vertex.Metadata.AggregationSourceRefs {
		refs = append(refs, sourceRefResource(ref))
	}
	return map[string]any{
		"vertex_id":     vertex.VertexID,
		"vertex_kind":   vertex.VertexKind,
		"vertex_family": vertex.VertexFamily,
		"labels":        nonNilStrings(vertex.Labels),
		"properties":    nonNilMap(vertex.Properties),
		"metadata": map[string]any{
			"mapping_rule_id":         nullableString(vertex.Metadata.MappingRuleID),
			"aggregation_rule_id":     nullableString(vertex.Metadata.AggregationRuleID),
			"aggregation_source_refs": refs,
			"mapped_metadata":         nonNilMap(vertex.Metadata.MappedMetadata),
		},
		"source_entity_ref": sourceRef,
		"sort_key":          vertex.SortKey,
	}
}

func edgeResource(edge Edge) map[string]any {
	var relationshipRef any
	if edge.SourceRelationshipRef != nil {
		relationshipRef = map[string]any{
			"source_relationship_id":   edge.SourceRelationshipRef.SourceRelationshipID,
			"source_relationship_kind": edge.SourceRelationshipRef.SourceRelationshipKind,
			"mapping_rule_id":          edge.SourceRelationshipRef.MappingRuleID,
		}
	}
	refs := make([]any, 0, len(edge.Metadata.AggregationSourceRefs))
	for _, ref := range edge.Metadata.AggregationSourceRefs {
		refs = append(refs, sourceRefResource(ref))
	}
	return map[string]any{
		"edge_id":       edge.EdgeID,
		"edge_kind":     edge.EdgeKind,
		"edge_family":   edge.EdgeFamily,
		"src_vertex_id": edge.SrcVertexID,
		"dst_vertex_id": edge.DstVertexID,
		"direction":     edge.Direction,
		"labels":        nonNilStrings(edge.Labels),
		"properties":    nonNilMap(edge.Properties),
		"metadata": map[string]any{
			"mapping_rule_id":         nullableString(edge.Metadata.MappingRuleID),
			"aggregation_rule_id":     nullableString(edge.Metadata.AggregationRuleID),
			"is_reverse_edge":         edge.Metadata.IsReverseEdge,
			"reverse_of_edge_id":      nullableString(edge.Metadata.ReverseOfEdgeID),
			"aggregation_source_refs": refs,
			"mapped_metadata":         nonNilMap(edge.Metadata.MappedMetadata),
		},
		"source_relationship_ref": relationshipRef,
		"sort_key":                edge.SortKey,
	}
}

func sourceRefResource(ref SourceRef) map[string]any {
	return map[string]any{"ref_kind": ref.RefKind, "ref_id": ref.RefID, "contributor_sort_key": ref.ContributorSortKey}
}

func validationSummaryResource(summary ValidationSummary) map[string]any {
	issues := make([]any, 0, len(summary.Issues))
	counts := map[string]int{"fatal": 0, "error": 0, "warning": 0, "info": 0}
	for _, issue := range summary.Issues {
		counts[issue.Severity]++
		issues = append(issues, map[string]any{
			"issue_id": issue.IssueID, "severity": issue.Severity, "code": issue.Code,
			"target_kind": issue.TargetKind, "target_id": issue.TargetID, "field": nullableString(issue.Field),
			"message": issue.Message, "details": nonNilMap(issue.Details),
		})
	}
	return map[string]any{
		"status": summary.Status, "fatal_count": counts["fatal"], "error_count": counts["error"],
		"warning_count": counts["warning"], "info_count": counts["info"], "issues": issues,
	}
}

func schemaRegistryResource(registry SchemaRegistry) map[string]any {
	vertices := make([]any, 0, len(registry.VertexKinds))
	for _, item := range registry.VertexKinds {
		properties := make([]any, 0, len(item.Properties))
		for _, property := range item.Properties {
			properties = append(properties, map[string]any{"projected_key": property.ProjectedKey, "projected_type": property.ProjectedType, "required": property.Required, "nullable_output": property.NullableOutput})
		}
		vertices = append(vertices, map[string]any{"vertex_kind": item.VertexKind, "source_entity_kinds": nonNilStrings(item.SourceEntityKinds), "aggregation_rule_ids": nonNilStrings(item.AggregationRuleIDs), "labels": nonNilStrings(item.Labels), "source_labels_preserved": item.SourceLabelsPreserved, "properties": properties})
	}
	edges := make([]any, 0, len(registry.EdgeKinds))
	for _, item := range registry.EdgeKinds {
		properties := make([]any, 0, len(item.Properties))
		for _, property := range item.Properties {
			properties = append(properties, map[string]any{"projected_key": property.ProjectedKey, "projected_type": property.ProjectedType, "required": property.Required, "nullable_output": property.NullableOutput})
		}
		edges = append(edges, map[string]any{"edge_kind": item.EdgeKind, "source_relationship_kinds": nonNilStrings(item.SourceRelationshipKinds), "aggregation_rule_ids": nonNilStrings(item.AggregationRuleIDs), "directions": nonNilStrings(item.Directions), "labels": nonNilStrings(item.Labels), "source_labels_preserved": item.SourceLabelsPreserved, "properties": properties})
	}
	properties := make([]any, 0, len(registry.PropertyKeys))
	for _, item := range registry.PropertyKeys {
		properties = append(properties, map[string]any{"target_scope": item.TargetScope, "target_kind": item.TargetKind, "projected_key": item.ProjectedKey, "projected_type": item.ProjectedType, "required": item.Required, "nullable_output": item.NullableOutput, "missing_behavior": item.MissingBehavior, "source_null_behavior": item.SourceNullBehavior})
	}
	metadata := make([]any, 0, len(registry.MetadataKeys))
	for _, item := range registry.MetadataKeys {
		metadata = append(metadata, map[string]any{"target_scope": item.TargetScope, "target_kind": item.TargetKind, "projected_metadata_key": item.ProjectedMetadataKey, "projected_type": item.ProjectedType, "required": item.Required, "nullable_output": item.NullableOutput, "missing_behavior": item.MissingBehavior, "source_null_behavior": item.SourceNullBehavior})
	}
	return map[string]any{"vertex_kinds": vertices, "edge_kinds": edges, "property_keys": properties, "metadata_keys": metadata}
}

func nonNilMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func nonNilStrings(value []string) []string {
	if value == nil {
		return []string{}
	}
	return value
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func invalidationResource(value *InvalidationSummary) any {
	if value == nil {
		return nil
	}
	return map[string]any{
		"graph_view_id":                  value.GraphViewID,
		"target_scope":                   value.TargetScope,
		"target_projection_run_id":       nullableString(value.TargetProjectionRunID),
		"invalidated_projection_run_ids": nonNilStrings(value.InvalidatedRunIDs),
		"graph_view_state_after":         string(value.GraphViewStateAfter),
		"invalidated_at":                 value.InvalidatedAt,
		"reason_code":                    value.ReasonCode,
		"requested_by":                   value.RequestedBy,
		"idempotency_expires_at":         nullableString(value.IdempotencyExpiresAt),
	}
}
