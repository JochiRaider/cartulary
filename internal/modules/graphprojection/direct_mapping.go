package graphprojection

import "context"

func emitDirectVertices(ctx context.Context, run ProjectionRun) ([]Vertex, map[string][]Vertex, []ValidationIssue, error) {
	vertices := []Vertex{}
	issues := []ValidationIssue{}
	bySource := map[string][]Vertex{}
	declaredKinds := stringSet(run.Request.ProjectionConfig.DeclaredSourceEntityKinds)
	for _, entity := range run.Request.SourceEntities {
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}
		if !sourceItemWithinLimits(len(entity.Labels), len(entity.Properties), len(entity.Metadata)) || !sourceItemValuesWithinLimits(entity.Properties, entity.Metadata) {
			continue
		}
		if !validIdentifier(entity.SourceEntityID) || !validIdentifier(entity.SourceEntityKind) {
			field := "$.source_entities"
			issues = append(issues, run.issue("fatal", "invalid_input_shape", "projection_input", "projection_input", field, map[string]any{"field": field, "reason_code": "scalar_contract_violation"}))
			continue
		}
		if !declaredKinds[entity.SourceEntityKind] {
			issues = append(issues, run.issue("error", "undeclared_source_kind", "source_item", entity.SourceEntityID, nil, map[string]any{"source_item_id": entity.SourceEntityID, "source_kind": entity.SourceEntityKind}))
			continue
		}
		if !filtersMatchEntity(run.Request.Filters.EntityFilters, entity) {
			continue
		}
		for _, mapping := range run.Request.ProjectionConfig.EntityMappings {
			if mapping.SourceEntityKind != entity.SourceEntityKind || !entityMappingIncludes(mapping, entity) {
				continue
			}
			vertexID, _ := generatedID("vx_", "GPVERTEX1\n", "direct_vertex", ProjectionSchemaID, run.GraphViewID, entity.SourceEntityKind, entity.SourceEntityID, mapping.MappingIdentityDigest)
			properties, propertyIssues := deriveProperties(run, "vertex", mapping.ProjectedVertexKind, entity, nil, nil, vertexID)
			issues = append(issues, propertyIssues...)
			mappedMetadata, metadataIssues := deriveMetadata(run, "vertex", mapping.ProjectedVertexKind, entity, nil, vertexID)
			issues = append(issues, metadataIssues...)
			mappingID := mapping.MappingRuleID
			vertex := Vertex{
				VertexID:     vertexID,
				VertexKind:   mapping.ProjectedVertexKind,
				VertexFamily: "direct",
				Labels:       deriveLabels(run.Request.ProjectionConfig.DefaultVertexLabels, mapping.MappingLabels, entity.Labels, mapping.LabelPolicy),
				Properties:   properties,
				Metadata: VertexMetadata{
					MappingRuleID:  &mappingID,
					MappedMetadata: mappedMetadata,
				},
				SourceEntityRef: &SourceEntityRef{SourceEntityID: entity.SourceEntityID, SourceEntityKind: entity.SourceEntityKind, MappingRuleID: mapping.MappingRuleID},
			}
			vertex.SortKey = sortKey("vertex", "direct", vertex.VertexKind, entity.SourceEntityKind, entity.SourceEntityID, mapping.MappingIdentityDigest, vertex.VertexID)
			vertices = append(vertices, vertex)
			bySource[entity.SourceEntityID] = append(bySource[entity.SourceEntityID], vertex)
		}
	}
	return vertices, bySource, issues, nil
}

func emitDirectEdges(ctx context.Context, run ProjectionRun, vertexBySource map[string][]Vertex) ([]Edge, []ValidationIssue, error) {
	edges := []Edge{}
	issues := []ValidationIssue{}
	declaredKinds := stringSet(run.Request.ProjectionConfig.DeclaredSourceRelationshipKinds)
	for _, relationship := range run.Request.SourceRelationships {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		if !sourceItemWithinLimits(len(relationship.Labels), len(relationship.Properties), len(relationship.Metadata)) || !sourceItemValuesWithinLimits(relationship.Properties, relationship.Metadata) {
			continue
		}
		if !validIdentifier(relationship.SourceRelationshipID) || !validIdentifier(relationship.SourceRelationshipKind) {
			field := "$.source_relationships"
			issues = append(issues, run.issue("fatal", "invalid_input_shape", "projection_input", "projection_input", field, map[string]any{"field": field, "reason_code": "scalar_contract_violation"}))
			continue
		}
		if !declaredKinds[relationship.SourceRelationshipKind] {
			issues = append(issues, run.issue("error", "undeclared_source_kind", "source_item", relationship.SourceRelationshipID, nil, map[string]any{"source_item_id": relationship.SourceRelationshipID, "source_kind": relationship.SourceRelationshipKind}))
			continue
		}
		if relationship.SrcSourceEntityID == "" {
			issues = append(issues, run.issue("error", "missing_relationship_endpoint", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"source_relationship_id": relationship.SourceRelationshipID, "endpoint_field": "src_source_entity_id"}))
		}
		if relationship.DstSourceEntityID == "" {
			issues = append(issues, run.issue("error", "missing_relationship_endpoint", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"source_relationship_id": relationship.SourceRelationshipID, "endpoint_field": "dst_source_entity_id"}))
		}
		if relationship.SrcSourceEntityID == "" || relationship.DstSourceEntityID == "" {
			continue
		}
		if !validSourceDirection(relationship.Direction) {
			issues = append(issues, run.issue("error", "invalid_relationship_direction", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"source_relationship_id": relationship.SourceRelationshipID, "supplied_value": relationship.Direction}))
			continue
		}
		if !filtersMatchRelationship(run.Request.Filters.RelationshipFilters, relationship) {
			continue
		}
		for _, mapping := range run.Request.RelationshipMappings {
			if mapping.SourceRelationshipKind != relationship.SourceRelationshipKind || !relationshipMappingIncludes(mapping, relationship) {
				continue
			}
			srcVertices := vertexBySource[relationship.SrcSourceEntityID]
			dstVertices := vertexBySource[relationship.DstSourceEntityID]
			if len(srcVertices) == 0 {
				issues = append(issues, run.issue("error", "relationship_endpoint_not_projected", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"source_relationship_id": relationship.SourceRelationshipID, "endpoint_field": "src_source_entity_id", "endpoint_source_entity_id": relationship.SrcSourceEntityID}))
			}
			if len(dstVertices) == 0 {
				issues = append(issues, run.issue("error", "relationship_endpoint_not_projected", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"source_relationship_id": relationship.SourceRelationshipID, "endpoint_field": "dst_source_entity_id", "endpoint_source_entity_id": relationship.DstSourceEntityID}))
			}
			if len(srcVertices) == 0 || len(dstVertices) == 0 {
				continue
			}
			srcVertex, dstVertex, direction := projectDirection(mapping.DirectionPolicy, relationship.Direction, srcVertices[0], dstVertices[0])
			edgeID, _ := generatedID("ed_", "GPEDGE1\n", "direct_edge", ProjectionSchemaID, run.GraphViewID, relationship.SourceRelationshipKind, relationship.SourceRelationshipID, mapping.ProjectedEdgeKind, srcVertex.VertexID, dstVertex.VertexID, direction, mapping.MappingIdentityDigest)
			properties, propertyIssues := deriveProperties(run, "edge", mapping.ProjectedEdgeKind, SourceEntity{}, &relationship, nil, edgeID)
			issues = append(issues, propertyIssues...)
			mappedMetadata, metadataIssues := deriveMetadata(run, "edge", mapping.ProjectedEdgeKind, SourceEntity{}, &relationship, edgeID)
			issues = append(issues, metadataIssues...)
			mappingID := mapping.MappingRuleID
			edge := Edge{
				EdgeID:      edgeID,
				EdgeKind:    mapping.ProjectedEdgeKind,
				EdgeFamily:  "direct",
				SrcVertexID: srcVertex.VertexID,
				DstVertexID: dstVertex.VertexID,
				Direction:   direction,
				Labels:      deriveLabels(run.Request.ProjectionConfig.DefaultEdgeLabels, mapping.MappingLabels, relationship.Labels, mapping.LabelPolicy),
				Properties:  properties,
				Metadata: EdgeMetadata{
					MappingRuleID:  &mappingID,
					IsReverseEdge:  false,
					MappedMetadata: mappedMetadata,
				},
				SourceRelationshipRef: &SourceRelationshipRef{SourceRelationshipID: relationship.SourceRelationshipID, SourceRelationshipKind: relationship.SourceRelationshipKind, MappingRuleID: mapping.MappingRuleID},
			}
			edge.SortKey = sortKey("edge", "direct", edge.EdgeKind, relationship.SourceRelationshipKind, relationship.SourceRelationshipID, edge.SrcVertexID, edge.DstVertexID, edge.Direction, mapping.MappingIdentityDigest, false, edge.EdgeID)
			edges = append(edges, edge)
			if mapping.EmitReverseEdge {
				reverseID, _ := generatedID("ed_", "GPEDGE1\n", "reverse_edge", ProjectionSchemaID, run.GraphViewID, relationship.SourceRelationshipKind, relationship.SourceRelationshipID, mapping.ReverseEdgeKind, edge.DstVertexID, edge.SrcVertexID, "directed", mapping.MappingIdentityDigest)
				reverseOf := edge.EdgeID
				reverse := edge
				reverse.EdgeID = reverseID
				reverse.EdgeKind = mapping.ReverseEdgeKind
				reverse.EdgeFamily = "reverse"
				reverse.SrcVertexID = edge.DstVertexID
				reverse.DstVertexID = edge.SrcVertexID
				reverse.Direction = "directed"
				reverse.Metadata.IsReverseEdge = true
				reverse.Metadata.ReverseOfEdgeID = &reverseOf
				reverse.SortKey = sortKey("edge", "reverse", reverse.EdgeKind, relationship.SourceRelationshipKind, relationship.SourceRelationshipID, reverse.SrcVertexID, reverse.DstVertexID, reverse.Direction, mapping.MappingIdentityDigest, true, reverse.EdgeID)
				edges = append(edges, reverse)
			}
		}
	}
	return edges, issues, nil
}

func entityMappingIncludes(mapping EntityMapping, entity SourceEntity) bool {
	if mapping.InclusionFilter != nil {
		value, present := sourceField(entity, nil, nil, mapping.InclusionFilter.FieldPath)
		return filterMatches(value, present, *mapping.InclusionFilter)
	}
	return mapping.InclusionPredicate == "always"
}

func relationshipMappingIncludes(mapping RelationshipMapping, relationship SourceRelationship) bool {
	if mapping.InclusionFilter != nil {
		value, present := sourceField(SourceEntity{}, &relationship, nil, mapping.InclusionFilter.FieldPath)
		return filterMatches(value, present, *mapping.InclusionFilter)
	}
	return mapping.InclusionPredicate == "always"
}
