package graphprojection

import "context"

// These tuples contain only admitted UTF-8 identifiers.
func entityMappingDigestV2(mappingID, sourceKind, vertexKind string) string {
	digest, _ := digestTuple("GPMAPENTITY1\n", mappingID, sourceKind, vertexKind)
	return digest
}

func directVertexIdentityV2(view, vertexKind, sourceKind, sourceID, mappingDigest string) (string, string) {
	id, _ := generatedID("vx_", "GPVERTEX1\n", "direct_vertex", ProjectionSchemaIDV2, view, sourceKind, sourceID, mappingDigest)
	return id, sortKey("vertex", "direct", vertexKind, sourceKind, sourceID, mappingDigest, id)
}

func relationshipMappingDigestV2(id, source, forward, policy string, reverse bool, reverseKind string) string {
	digest, _ := digestTuple("GPMAPREL1\n", id, source, forward, policy, reverse, reverseKind)
	return digest
}

func mappedEdgeIdentityV2(view, family, kind, sourceKind, sourceID, src, dst, direction, mappingDigest string) (string, string) {
	id, _ := generatedID("ed_", "GPEDGE1\n", family+"_edge", ProjectionSchemaIDV2, view, sourceKind, sourceID, kind, src, dst, direction, mappingDigest)
	return id, sortKey("edge", family, kind, sourceKind, sourceID, src, dst, direction, mappingDigest, family == "reverse", id)
}

func (state *projectionState) emitDirectVertices(ctx context.Context, run *projectionWork) error {
	declaredKinds := stringSet(run.Request.projectionConfig.DeclaredSourceEntityKinds)
	for _, entity := range run.Request.SourceEntities {
		if run.issues.err != nil {
			return run.issues.err
		}
		if err := itemCheckpointV2(ctx); err != nil {
			return err
		}
		if !declaredKinds[entity.SourceEntityKind] {
			run.recordIssue("error", "undeclared_source_kind", "source_item", entity.SourceEntityID, nil, map[string]any{"source_item_id": entity.SourceEntityID, "source_kind": entity.SourceEntityKind})
			continue
		}
		if !filtersMatchEntity(run.Request.filters.EntityFilters, entity) {
			continue
		}
		for _, mapping := range run.Request.projectionConfig.EntityMappings {
			if mapping.SourceEntityKind != entity.SourceEntityKind || !entityMappingIncludes(mapping, entity) {
				continue
			}
			err := state.emitVertex(ctx, &run.issues, "", func() (Vertex, error) {
				vertexID, key := directVertexIdentityV2(run.GraphViewID, mapping.ProjectedVertexKind, entity.SourceEntityKind, entity.SourceEntityID, mapping.MappingIdentityDigest)
				labels, err := deriveLabels(run.Request.projectionConfig.DefaultVertexLabels, mapping.MappingLabels, entity.Labels, mapping.LabelPolicy)
				if err != nil {
					return Vertex{}, err
				}
				properties := deriveProperties(run, "vertex", mapping.ProjectedVertexKind, entity, nil, vertexID)
				mappedMetadata := deriveMetadata(run, "vertex", mapping.ProjectedVertexKind, entity, nil, vertexID)
				mappingID := mapping.MappingRuleID
				vertex := Vertex{
					VertexID:     vertexID,
					VertexKind:   mapping.ProjectedVertexKind,
					VertexFamily: "direct",
					Labels:       labels,
					Properties:   properties,
					Metadata: VertexMetadata{
						MappingRuleID:  &mappingID,
						MappedMetadata: mappedMetadata,
					},
					SourceEntityRef: &SourceEntityRef{SourceEntityID: entity.SourceEntityID, SourceEntityKind: entity.SourceEntityKind, MappingRuleID: mapping.MappingRuleID},
				}
				vertex.SortKey = key
				return vertex, run.issues.err
			})
			if err != nil {
				return err
			}
		}
	}
	return run.issues.err
}

func (state *projectionState) emitDirectEdges(ctx context.Context, run *projectionWork) error {
	declaredKinds := stringSet(run.Request.projectionConfig.DeclaredSourceRelationshipKinds)
	for _, relationship := range run.Request.SourceRelationships {
		if run.issues.err != nil {
			return run.issues.err
		}
		if err := itemCheckpointV2(ctx); err != nil {
			return err
		}
		if !declaredKinds[relationship.SourceRelationshipKind] {
			run.recordIssue("error", "undeclared_source_kind", "source_item", relationship.SourceRelationshipID, nil, map[string]any{"source_item_id": relationship.SourceRelationshipID, "source_kind": relationship.SourceRelationshipKind})
			continue
		}
		if !filtersMatchRelationship(run.Request.filters.RelationshipFilters, relationship) {
			continue
		}
		for _, mapping := range run.Request.projectionConfig.RelationshipMappings {
			if mapping.SourceRelationshipKind != relationship.SourceRelationshipKind || !relationshipMappingIncludes(mapping, relationship) {
				continue
			}
			srcVertices := state.verticesBySource[relationship.SrcSourceEntityID]
			dstVertices := state.verticesBySource[relationship.DstSourceEntityID]
			if len(srcVertices) == 0 {
				run.recordIssue("error", "relationship_endpoint_not_projected", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"source_relationship_id": relationship.SourceRelationshipID, "endpoint_field": "src_source_entity_id", "endpoint_source_entity_id": relationship.SrcSourceEntityID})
			}
			if len(dstVertices) == 0 {
				run.recordIssue("error", "relationship_endpoint_not_projected", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"source_relationship_id": relationship.SourceRelationshipID, "endpoint_field": "dst_source_entity_id", "endpoint_source_entity_id": relationship.DstSourceEntityID})
			}
			if len(srcVertices) == 0 || len(dstVertices) == 0 {
				continue
			}
			srcVertex, dstVertex, direction := projectDirection(mapping.DirectionPolicy, relationship.Direction, srcVertices[0], dstVertices[0])
			// Both outputs read the relationship; applicability and diagnostics belong to the output being built.
			build := func(kind, family, src, dst, direction string, reverseOf *string) (Edge, error) {
				id, key := mappedEdgeIdentityV2(run.GraphViewID, family, kind, relationship.SourceRelationshipKind, relationship.SourceRelationshipID, src, dst, direction, mapping.MappingIdentityDigest)
				labels, err := deriveLabels(run.Request.projectionConfig.DefaultEdgeLabels, mapping.MappingLabels, relationship.Labels, mapping.LabelPolicy)
				if err != nil {
					return Edge{}, err
				}
				edge := Edge{
					EdgeID: id, EdgeKind: kind, EdgeFamily: family, SrcVertexID: src, DstVertexID: dst, Direction: direction,
					Labels:                labels,
					Properties:            deriveProperties(run, "edge", kind, sourceEntity{}, &relationship, id),
					Metadata:              EdgeMetadata{MappingRuleID: &mapping.MappingRuleID, IsReverseEdge: reverseOf != nil, ReverseOfEdgeID: reverseOf, MappedMetadata: deriveMetadata(run, "edge", kind, sourceEntity{}, &relationship, id)},
					SourceRelationshipRef: &SourceRelationshipRef{SourceRelationshipID: relationship.SourceRelationshipID, SourceRelationshipKind: relationship.SourceRelationshipKind, MappingRuleID: mapping.MappingRuleID},
					SortKey:               key,
				}
				return edge, run.issues.err
			}
			before := len(state.edges)
			if err := state.emitEdge(ctx, &run.issues, func() (Edge, error) {
				return build(mapping.ProjectedEdgeKind, "direct", srcVertex.VertexID, dstVertex.VertexID, direction, nil)
			}); err != nil {
				return err
			}
			if mapping.EmitReverseEdge && len(state.edges) > before {
				forwardID := state.edges[before].EdgeID
				if err := state.emitEdge(ctx, &run.issues, func() (Edge, error) {
					return build(mapping.ReverseEdgeKind, "reverse", dstVertex.VertexID, srcVertex.VertexID, "directed", &forwardID)
				}); err != nil {
					return err
				}
			}
		}
	}
	return run.issues.err
}

func entityMappingIncludes(mapping entityMapping, entity sourceEntity) bool {
	if mapping.InclusionFilter != nil {
		value, present := sourceField(entity, nil, nil, mapping.InclusionFilter.FieldPath)
		return filterMatches(value, present, *mapping.InclusionFilter)
	}
	return mapping.InclusionPredicate == "always"
}

func relationshipMappingIncludes(mapping relationshipMapping, relationship sourceRelationship) bool {
	if mapping.InclusionFilter != nil {
		value, present := sourceField(sourceEntity{}, &relationship, nil, mapping.InclusionFilter.FieldPath)
		return filterMatches(value, present, *mapping.InclusionFilter)
	}
	return mapping.InclusionPredicate == "always"
}
