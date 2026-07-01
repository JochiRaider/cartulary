package graphprojection

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type ProjectOptions struct {
	ProjectionRunNonce      string
	AcceptedAt              time.Time
	GeneratedAt             time.Time
	PreviousProjectionRunID *string
}

func Project(data []byte, options ProjectOptions) (ProjectionRun, error) {
	run, err := AdmitProjectionInput(data, AdmitOptions{
		ProjectionRunNonce: options.ProjectionRunNonce,
		AcceptedAt:         options.AcceptedAt,
	})
	if err != nil {
		return ProjectionRun{}, err
	}
	run.State = RunStateComputing
	projected := projectAdmittedRun(run, options)
	return projected, nil
}

func projectAdmittedRun(run ProjectionRun, options ProjectOptions) ProjectionRun {
	issues := validateAdmittedRequest(run)
	result := run
	generatedAt := options.GeneratedAt.UTC()
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}
	result.PreviousProjectionRunID = options.PreviousProjectionRunID
	if hasFatalIssue(issues) {
		result.State = RunStateFailed
		result.FailureReason = "fatal_validation"
		now := generatedAt
		result.CompletedAt = &now
		result.ValidationSummary = validationSummary(issues)
		return result
	}

	vertices, vertexBySource, vertexIssues := emitDirectVertices(run)
	issues = append(issues, vertexIssues...)
	edges, edgeIssues := emitDirectEdges(run, vertexBySource)
	issues = append(issues, edgeIssues...)
	aggVertices, aggEdges, aggIssues := emitAggregations(run, vertices, edges)
	issues = append(issues, aggIssues...)
	vertices = append(vertices, aggVertices...)
	edges = append(edges, aggEdges...)
	sortVertices(vertices)
	sortEdges(edges)

	if hasFatalIssue(issues) {
		result.State = RunStateFailed
		result.FailureReason = "fatal_validation"
		now := generatedAt
		result.CompletedAt = &now
		result.ValidationSummary = validationSummary(issues)
		return result
	}

	summary := validationSummary(issues)
	result.State = RunStateAvailable
	now := generatedAt
	result.CompletedAt = &now
	graphView := &GraphView{
		ProjectionSchemaID: ProjectionSchemaID,
		GraphViewID:        result.GraphViewID,
		GraphViewKey:       run.Request.ProjectionConfig.GraphViewKey,
		ProjectionRunID:    result.ProjectionRunID,
		SourceSnapshotID:   run.Request.SourceSnapshotID,
		ProjectionVersion:  run.Request.ProjectionConfig.ProjectionVersion,
		GeneratedAt:        formatLifecycleTimestamp(generatedAt),
		State:              RunStateAvailable,
		Properties:         deriveGraphProperties(run, &issues),
		Metadata: GraphMetadata{
			ProjectionConfigDigest:  run.ProjectionConfigDigest,
			ProjectionSourceDigest:  run.ProjectionSourceDigest,
			PreviousProjectionRunID: options.PreviousProjectionRunID,
			MappedMetadata:          map[string]any{},
		},
		SchemaRegistry:       buildSchemaRegistry(run, vertices, edges),
		Vertices:             vertices,
		Edges:                edges,
		ValidationSummary:    summary,
		ConsumerCapabilities: defaultConsumerCapabilities(),
	}
	result.ValidationSummary = summary
	result.GraphView = graphView
	return result
}

func validateAdmittedRequest(run ProjectionRun) []ValidationIssue {
	request := run.Request
	issues := []ValidationIssue{}
	if request.ProjectionConfig.GraphViewKey == "" || !validIdentifier(request.ProjectionConfig.GraphViewKey) {
		issues = append(issues, run.issue("fatal", "invalid_projection_config", "graph_view", run.GraphViewID, "$.projection_config.graph_view_key", map[string]any{"reason_code": "invalid_graph_view_key"}))
	}
	if len(request.ProjectionConfig.DeclaredSourceEntityKinds) == 0 && len(request.ProjectionConfig.DeclaredSourceRelationshipKinds) == 0 && !request.ProjectionConfig.AllowEmptyKindRegistry {
		issues = append(issues, run.issue("fatal", "invalid_projection_config", "graph_view", run.GraphViewID, "$.projection_config", map[string]any{"reason_code": "empty_kind_registry_not_allowed"}))
	}
	checkDuplicates := func(values []string, code, targetKind string) {
		seen := map[string]struct{}{}
		for _, value := range values {
			if _, ok := seen[value]; ok {
				issues = append(issues, run.issue("fatal", code, targetKind, value, nil, map[string]any{"reason_code": "duplicate"}))
			}
			seen[value] = struct{}{}
		}
	}
	checkDuplicates(entityIDs(request.SourceEntities), "duplicate_source_entity_id", "source_entity")
	checkDuplicates(relationshipIDs(request.SourceRelationships), "duplicate_source_relationship_id", "source_relationship")
	for _, mapping := range request.RelationshipMappings {
		if mapping.EmitReverseEdge && mapping.DirectionPolicy != "normalize_forward" && mapping.DirectionPolicy != "normalize_reverse" {
			issues = append(issues, run.issue("fatal", "invalid_reverse_edge_policy", "mapping_rule", mapping.MappingRuleID, nil, map[string]any{"mapping_rule_id": mapping.MappingRuleID, "projected_direction": nil}))
		}
	}
	for _, definition := range request.PropertyDefinitions {
		if !validPropertyKey(definition.ProjectedKey) {
			issues = append(issues, run.issue("fatal", "invalid_property_definition", "property", definition.PropertyDefinitionID, nil, map[string]any{"projected_key": definition.ProjectedKey, "reason_code": "invalid_projected_key"}))
		}
		if !validFieldPath(definition.SourceFieldPath) {
			issues = append(issues, run.issue("fatal", "invalid_field_path", "property", definition.PropertyDefinitionID, nil, map[string]any{"source_field_path": definition.SourceFieldPath}))
		}
	}
	return issues
}

func emitDirectVertices(run ProjectionRun) ([]Vertex, map[string][]Vertex, []ValidationIssue) {
	vertices := []Vertex{}
	issues := []ValidationIssue{}
	bySource := map[string][]Vertex{}
	declaredKinds := stringSet(run.Request.ProjectionConfig.DeclaredSourceEntityKinds)
	for _, entity := range run.Request.SourceEntities {
		if !validIdentifier(entity.SourceEntityID) || !validIdentifier(entity.SourceEntityKind) {
			issues = append(issues, run.issue("fatal", "invalid_source_entity", "source_entity", entity.SourceEntityID, nil, map[string]any{"reason_code": "invalid_identifier"}))
			continue
		}
		if !declaredKinds[entity.SourceEntityKind] {
			issues = append(issues, run.issue("error", "undeclared_source_kind", "source_entity", entity.SourceEntityID, nil, map[string]any{"source_entity_kind": entity.SourceEntityKind}))
			continue
		}
		if !filtersMatchEntity(run.Request.Filters.EntityFilters, entity) {
			continue
		}
		for _, mapping := range run.Request.ProjectionConfig.EntityMappings {
			if mapping.SourceEntityKind != entity.SourceEntityKind || mapping.InclusionPredicate != "always" {
				continue
			}
			vertexID, _ := generatedID("vx_", "GPVERTEX1\n", "direct_vertex", ProjectionSchemaID, run.GraphViewID, entity.SourceEntityKind, entity.SourceEntityID, mapping.MappingIdentityDigest)
			properties, propertyIssues := deriveProperties(run, "vertex", mapping.ProjectedVertexKind, entity, nil, nil, vertexID)
			issues = append(issues, propertyIssues...)
			mappedMetadata := deriveMetadata(run, "vertex", mapping.ProjectedVertexKind, entity, nil)
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
	return vertices, bySource, issues
}

func emitDirectEdges(run ProjectionRun, vertexBySource map[string][]Vertex) ([]Edge, []ValidationIssue) {
	edges := []Edge{}
	issues := []ValidationIssue{}
	declaredKinds := stringSet(run.Request.ProjectionConfig.DeclaredSourceRelationshipKinds)
	for _, relationship := range run.Request.SourceRelationships {
		if !validIdentifier(relationship.SourceRelationshipID) || !validIdentifier(relationship.SourceRelationshipKind) {
			issues = append(issues, run.issue("fatal", "invalid_source_relationship", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"reason_code": "invalid_identifier"}))
			continue
		}
		if !declaredKinds[relationship.SourceRelationshipKind] {
			issues = append(issues, run.issue("error", "undeclared_source_kind", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"source_relationship_kind": relationship.SourceRelationshipKind}))
			continue
		}
		if relationship.SrcSourceEntityID == "" || relationship.DstSourceEntityID == "" {
			issues = append(issues, run.issue("error", "missing_relationship_endpoint", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"source_relationship_id": relationship.SourceRelationshipID}))
			continue
		}
		if !validSourceDirection(relationship.Direction) {
			issues = append(issues, run.issue("error", "invalid_relationship_direction", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"direction": relationship.Direction}))
			continue
		}
		if !filtersMatchRelationship(run.Request.Filters.RelationshipFilters, relationship) {
			continue
		}
		for _, mapping := range run.Request.RelationshipMappings {
			if mapping.SourceRelationshipKind != relationship.SourceRelationshipKind || mapping.InclusionPredicate != "always" {
				continue
			}
			srcVertices := vertexBySource[relationship.SrcSourceEntityID]
			dstVertices := vertexBySource[relationship.DstSourceEntityID]
			if len(srcVertices) == 0 || len(dstVertices) == 0 {
				issues = append(issues, run.issue("error", "relationship_endpoint_not_projected", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"source_relationship_id": relationship.SourceRelationshipID}))
				continue
			}
			srcVertex, dstVertex, direction := projectDirection(mapping.DirectionPolicy, relationship.Direction, srcVertices[0], dstVertices[0])
			edgeID, _ := generatedID("ed_", "GPEDGE1\n", "direct_edge", ProjectionSchemaID, run.GraphViewID, relationship.SourceRelationshipKind, relationship.SourceRelationshipID, mapping.ProjectedEdgeKind, srcVertex.VertexID, dstVertex.VertexID, direction, mapping.MappingIdentityDigest)
			properties, propertyIssues := deriveProperties(run, "edge", mapping.ProjectedEdgeKind, SourceEntity{}, &relationship, nil, edgeID)
			issues = append(issues, propertyIssues...)
			mappedMetadata := deriveMetadata(run, "edge", mapping.ProjectedEdgeKind, SourceEntity{}, &relationship)
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
	return edges, issues
}

func emitAggregations(run ProjectionRun, directVertices []Vertex, directEdges []Edge) ([]Vertex, []Edge, []ValidationIssue) {
	vertices := []Vertex{}
	edges := []Edge{}
	issues := []ValidationIssue{}
	aggregatedVerticesByRuleAndDigest := map[string]map[string]Vertex{}
	for _, rule := range run.Request.ProjectionConfig.AggregationRules {
		if rule.TargetScope != "vertex" {
			continue
		}
		groups := groupContributors(run, rule, directVertices, directEdges, &issues)
		for digest, contributors := range groups {
			vertexID, _ := generatedID("vx_", "GPVERTEX1\n", "aggregated_vertex", ProjectionSchemaID, run.GraphViewID, rule.AggregationIdentityDigest, digest)
			props := mergeAggregateProperties(run, rule, contributors, "vertex", rule.ProjectedKind, vertexID, &issues)
			ruleID := rule.AggregationRuleID
			vertex := Vertex{
				VertexID:     vertexID,
				VertexKind:   rule.ProjectedKind,
				VertexFamily: "aggregated",
				Labels:       run.Request.ProjectionConfig.DefaultVertexLabels,
				Properties:   props,
				Metadata: VertexMetadata{
					AggregationRuleID:     &ruleID,
					AggregationSourceRefs: sourceRefs(contributors),
					MappedMetadata:        map[string]any{},
				},
				SortKey: sortKey("vertex", "aggregated", rule.ProjectedKind, rule.AggregationRuleID, digest, vertexID),
			}
			if aggregatedVerticesByRuleAndDigest[rule.AggregationRuleID] == nil {
				aggregatedVerticesByRuleAndDigest[rule.AggregationRuleID] = map[string]Vertex{}
			}
			aggregatedVerticesByRuleAndDigest[rule.AggregationRuleID][digest] = vertex
			vertices = append(vertices, vertex)
		}
	}
	for _, rule := range run.Request.ProjectionConfig.AggregationRules {
		if rule.TargetScope != "edge" || rule.EndpointGrouping == nil {
			continue
		}
		aggregationKinds := aggregationProjectedKinds(run.Request.ProjectionConfig.AggregationRules)
		groups := groupContributors(run, rule, directVertices, directEdges, &issues)
		for digest, contributors := range groups {
			srcRuleID := rule.EndpointGrouping.SourceVertexAggregationRuleID
			dstRuleID := rule.EndpointGrouping.DestinationVertexAggregationRuleID
			srcDigest, srcOK := endpointDigest(srcRuleID, "vertex", aggregationKinds[srcRuleID], rule.EndpointGrouping.SourceGroupingKeys, contributors)
			dstDigest, dstOK := endpointDigest(dstRuleID, "vertex", aggregationKinds[dstRuleID], rule.EndpointGrouping.DestinationGroupingKeys, contributors)
			src := aggregatedVerticesByRuleAndDigest[rule.EndpointGrouping.SourceVertexAggregationRuleID][srcDigest]
			dst := aggregatedVerticesByRuleAndDigest[rule.EndpointGrouping.DestinationVertexAggregationRuleID][dstDigest]
			if !srcOK || !dstOK || src.VertexID == "" || dst.VertexID == "" {
				if rule.EndpointGrouping.MissingEndpointBehavior == "error" {
					issues = append(issues, run.issue("error", "aggregation_endpoint_missing", "mapping_rule", rule.AggregationRuleID, nil, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "endpoint_digest": digest, "reason_code": "endpoint_vertex_not_found"}))
				}
				continue
			}
			edgeID, _ := generatedID("ed_", "GPEDGE1\n", "aggregated_edge", ProjectionSchemaID, run.GraphViewID, rule.AggregationIdentityDigest, src.VertexID, dst.VertexID, rule.EdgeDirection, digest)
			props := mergeAggregateProperties(run, rule, contributors, "edge", rule.ProjectedKind, edgeID, &issues)
			ruleID := rule.AggregationRuleID
			edge := Edge{
				EdgeID:      edgeID,
				EdgeKind:    rule.ProjectedKind,
				EdgeFamily:  "aggregated",
				SrcVertexID: src.VertexID,
				DstVertexID: dst.VertexID,
				Direction:   rule.EdgeDirection,
				Labels:      run.Request.ProjectionConfig.DefaultEdgeLabels,
				Properties:  props,
				Metadata: EdgeMetadata{
					AggregationRuleID:     &ruleID,
					AggregationSourceRefs: sourceRefs(contributors),
					MappedMetadata:        map[string]any{},
				},
				SortKey: sortKey("edge", "aggregated", rule.ProjectedKind, rule.AggregationRuleID, src.VertexID, dst.VertexID, rule.EdgeDirection, digest, edgeID),
			}
			edges = append(edges, edge)
		}
	}
	return vertices, edges, issues
}

func aggregationProjectedKinds(rules []AggregationRule) map[string]string {
	out := map[string]string{}
	for _, rule := range rules {
		out[rule.AggregationRuleID] = rule.ProjectedKind
	}
	return out
}

type contributor struct {
	Kind         string
	ID           string
	SortKey      string
	Entity       *SourceEntity
	Relationship *SourceRelationship
	Vertex       *Vertex
	Edge         *Edge
}

func groupContributors(run ProjectionRun, rule AggregationRule, vertices []Vertex, edges []Edge, issues *[]ValidationIssue) map[string][]contributor {
	groups := map[string][]contributor{}
	for _, contributor := range contributorsForRule(run, rule, vertices, edges) {
		values := make([]any, 0, len(rule.GroupingKeys))
		missing := false
		for _, fieldPath := range rule.GroupingKeys {
			value, ok := contributorField(contributor, fieldPath)
			if !ok {
				missing = true
				if rule.MissingGroupingKeyBehavior == "error" {
					*issues = append(*issues, run.issue("error", "aggregation_grouping_key_missing", "mapping_rule", rule.AggregationRuleID, fieldPath, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "field_path": fieldPath, "contributor_id": contributor.ID}))
				}
				break
			}
			values = append(values, value)
		}
		if missing {
			continue
		}
		keyJSON, _ := canonicalJSON(values)
		digest, _ := digestTuple("GPGROUP1\n", rule.AggregationRuleID, rule.TargetScope, rule.ProjectedKind, keyJSON)
		groups[digest] = append(groups[digest], contributor)
	}
	for digest := range groups {
		sort.Slice(groups[digest], func(i, j int) bool {
			return groups[digest][i].SortKey < groups[digest][j].SortKey
		})
	}
	return groups
}

func contributorsForRule(run ProjectionRun, rule AggregationRule, vertices []Vertex, edges []Edge) []contributor {
	out := []contributor{}
	switch rule.InputScope {
	case "source_entity":
		for i := range run.Request.SourceEntities {
			entity := &run.Request.SourceEntities[i]
			if entity.SourceEntityKind != rule.InputKind {
				continue
			}
			out = append(out, contributor{Kind: "source_entity", ID: entity.SourceEntityID, SortKey: contributorSortKey("source_entity", entity.SourceEntityKind, entity.SourceEntityID), Entity: entity})
		}
	case "source_relationship":
		for i := range run.Request.SourceRelationships {
			relationship := &run.Request.SourceRelationships[i]
			if relationship.SourceRelationshipKind != rule.InputKind {
				continue
			}
			out = append(out, contributor{Kind: "source_relationship", ID: relationship.SourceRelationshipID, SortKey: contributorSortKey("source_relationship", relationship.SourceRelationshipKind, relationship.SourceRelationshipID), Relationship: relationship})
		}
	case "projected_vertex":
		for i := range vertices {
			vertex := &vertices[i]
			if vertex.VertexKind != rule.InputKind {
				continue
			}
			out = append(out, contributor{Kind: "projected_vertex", ID: vertex.VertexID, SortKey: contributorSortKey("projected_vertex", vertex.VertexKind, vertex.SortKey, vertex.VertexID), Vertex: vertex})
		}
	case "projected_edge":
		for i := range edges {
			edge := &edges[i]
			if edge.EdgeKind != rule.InputKind {
				continue
			}
			out = append(out, contributor{Kind: "projected_edge", ID: edge.EdgeID, SortKey: contributorSortKey("projected_edge", edge.EdgeKind, edge.SortKey, edge.EdgeID), Edge: edge})
		}
	}
	return out
}

func mergeAggregateProperties(run ProjectionRun, rule AggregationRule, contributors []contributor, targetScope, targetKind, outputID string, issues *[]ValidationIssue) map[string]any {
	properties := map[string]any{}
	for _, definition := range run.Request.PropertyDefinitions {
		if !propertyApplies(definition, targetScope, targetKind) {
			continue
		}
		mergeBehavior := definition.MergeBehavior
		if override := rule.PropertyMergeBehavior[definition.ProjectedKey]; override != "" {
			mergeBehavior = override
		}
		if mergeBehavior == "omit" {
			continue
		}
		if mergeBehavior == "count" {
			properties[definition.ProjectedKey] = len(contributors)
			continue
		}
		candidates := []any{}
		for _, contributor := range contributors {
			value, ok := contributorField(contributor, definition.SourceFieldPath)
			if !ok {
				if definition.MissingBehavior == "default" && definition.HasDefaultValue {
					candidates = append(candidates, definition.DefaultValue)
				} else if definition.MissingBehavior == "error" {
					*issues = append(*issues, run.issue("error", "required_property_missing", "property", outputID, definition.SourceFieldPath, map[string]any{"projected_key": definition.ProjectedKey, "source_field_path": definition.SourceFieldPath, "output_object_id": outputID, "aggregation_rule_id": rule.AggregationRuleID}))
				}
				continue
			}
			normalized, include, issueCode := normalizePropertyValue(definition, value)
			if issueCode != "" {
				*issues = append(*issues, run.issue("error", issueCode, "property", outputID, definition.SourceFieldPath, map[string]any{"projected_key": definition.ProjectedKey, "source_field_path": definition.SourceFieldPath, "output_object_id": outputID, "aggregation_rule_id": rule.AggregationRuleID}))
				continue
			}
			if include {
				candidates = append(candidates, normalized)
			}
		}
		merged, ok, conflict := mergeValues(mergeBehavior, candidates)
		if conflict {
			*issues = append(*issues, run.issue("error", "aggregation_merge_conflict", "mapping_rule", rule.AggregationRuleID, nil, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "projected_key": definition.ProjectedKey}))
			continue
		}
		if ok {
			properties[definition.ProjectedKey] = merged
		}
	}
	return properties
}

func deriveProperties(run ProjectionRun, targetScope, targetKind string, entity SourceEntity, relationship *SourceRelationship, graphSource map[string]any, outputID string) (map[string]any, []ValidationIssue) {
	properties := map[string]any{}
	issues := []ValidationIssue{}
	for _, definition := range run.Request.PropertyDefinitions {
		if !propertyApplies(definition, targetScope, targetKind) {
			continue
		}
		value, ok := sourceField(entity, relationship, graphSource, definition.SourceFieldPath)
		if !ok {
			if definition.MissingBehavior == "default" && definition.HasDefaultValue {
				properties[definition.ProjectedKey] = definition.DefaultValue
			} else if definition.MissingBehavior == "error" {
				issues = append(issues, run.issue("error", "required_property_missing", "property", outputID, definition.SourceFieldPath, map[string]any{"projected_key": definition.ProjectedKey, "source_field_path": definition.SourceFieldPath, "output_object_id": outputID}))
			}
			continue
		}
		normalized, include, issueCode := normalizePropertyValue(definition, value)
		if issueCode != "" {
			issues = append(issues, run.issue("error", issueCode, "property", outputID, definition.SourceFieldPath, map[string]any{"projected_key": definition.ProjectedKey, "source_field_path": definition.SourceFieldPath, "output_object_id": outputID}))
			continue
		}
		if include {
			properties[definition.ProjectedKey] = normalized
		}
	}
	return properties, issues
}

func deriveGraphProperties(run ProjectionRun, issues *[]ValidationIssue) map[string]any {
	properties, propertyIssues := deriveProperties(run, "graph_view", "*", SourceEntity{}, nil, run.Request.SourceMetadata, run.GraphViewID)
	*issues = append(*issues, propertyIssues...)
	return properties
}

func deriveMetadata(run ProjectionRun, targetScope, targetKind string, entity SourceEntity, relationship *SourceRelationship) map[string]any {
	metadata := map[string]any{}
	for _, mapping := range run.Request.ProjectionConfig.MetadataMappings {
		if mapping.TargetScope != targetScope || (mapping.TargetKind != "*" && mapping.TargetKind != targetKind) {
			continue
		}
		value, ok := sourceField(entity, relationship, run.Request.SourceMetadata, mapping.SourceFieldPath)
		if !ok || value == nil {
			continue
		}
		metadata[mapping.ProjectedMetadataKey] = value
	}
	return metadata
}

func normalizePropertyValue(definition PropertyDefinition, value any) (any, bool, string) {
	if value == nil {
		switch definition.SourceNullBehavior {
		case "default":
			return definition.DefaultValue, definition.HasDefaultValue, ""
		case "emit_null":
			return nil, definition.NullOutputPolicy == "emit_null", ""
		case "error":
			return nil, false, "source_null_for_required_property"
		default:
			return nil, false, ""
		}
	}
	if !valueMatchesType(definition.ProjectedType, value) {
		return nil, false, "invalid_property_type"
	}
	return value, true, ""
}

func mergeValues(behavior string, values []any) (any, bool, bool) {
	if len(values) == 0 {
		return nil, false, false
	}
	switch behavior {
	case "first_by_sort":
		return values[0], true, false
	case "last_by_sort":
		return values[len(values)-1], true, false
	case "distinct_sorted_array":
		seen := map[string]any{}
		for _, value := range values {
			if array, ok := value.([]any); ok {
				for _, entry := range array {
					seen[canonicalValueKey(entry)] = entry
				}
				continue
			}
			seen[canonicalValueKey(value)] = value
		}
		keys := make([]string, 0, len(seen))
		for key := range seen {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out := make([]any, 0, len(keys))
		for _, key := range keys {
			out = append(out, seen[key])
		}
		return out, true, false
	default:
		seen := map[string]any{}
		for _, value := range values {
			seen[canonicalValueKey(value)] = value
		}
		if len(seen) > 1 {
			return nil, false, true
		}
		for _, value := range seen {
			return value, true, false
		}
	}
	return nil, false, false
}

func sourceField(entity SourceEntity, relationship *SourceRelationship, graphSource map[string]any, fieldPath string) (any, bool) {
	parts := strings.Split(fieldPath, ".")
	switch parts[0] {
	case "source_entity_id":
		if entity.SourceEntityID == "" {
			return nil, false
		}
		return entity.SourceEntityID, true
	case "source_relationship_id":
		if relationship == nil {
			return nil, false
		}
		return relationship.SourceRelationshipID, true
	case "src_source_entity_id":
		if relationship == nil {
			return nil, false
		}
		return relationship.SrcSourceEntityID, relationship.SrcSourceEntityID != ""
	case "dst_source_entity_id":
		if relationship == nil {
			return nil, false
		}
		return relationship.DstSourceEntityID, relationship.DstSourceEntityID != ""
	case "kind":
		if relationship != nil {
			return relationship.SourceRelationshipKind, true
		}
		if entity.SourceEntityKind != "" {
			return entity.SourceEntityKind, true
		}
		return nil, false
	case "direction":
		if relationship == nil {
			return nil, false
		}
		return relationship.Direction, true
	case "properties":
		if len(parts) != 2 {
			return nil, false
		}
		if relationship != nil {
			value, ok := relationship.Properties[parts[1]]
			return value, ok
		}
		value, ok := entity.Properties[parts[1]]
		return value, ok
	case "metadata":
		if len(parts) != 2 {
			return nil, false
		}
		if relationship != nil {
			value, ok := relationship.Metadata[parts[1]]
			return value, ok
		}
		value, ok := entity.Metadata[parts[1]]
		return value, ok
	case "source_metadata":
		if len(parts) != 2 {
			return nil, false
		}
		value, ok := graphSource[parts[1]]
		return value, ok
	default:
		return nil, false
	}
}

func contributorField(entry contributor, fieldPath string) (any, bool) {
	if entry.Entity != nil {
		return sourceField(*entry.Entity, nil, nil, fieldPath)
	}
	if entry.Relationship != nil {
		return sourceField(SourceEntity{}, entry.Relationship, nil, fieldPath)
	}
	if entry.Vertex != nil {
		return projectedVertexField(*entry.Vertex, fieldPath)
	}
	if entry.Edge != nil {
		return projectedEdgeField(*entry.Edge, fieldPath)
	}
	return nil, false
}

func projectedVertexField(vertex Vertex, fieldPath string) (any, bool) {
	parts := strings.Split(fieldPath, ".")
	switch parts[0] {
	case "projected":
		if len(parts) < 2 {
			return nil, false
		}
		switch parts[1] {
		case "vertex_id":
			return vertex.VertexID, true
		case "properties":
			if len(parts) != 3 {
				return nil, false
			}
			value, ok := vertex.Properties[parts[2]]
			return value, ok
		case "metadata":
			if len(parts) != 3 {
				return nil, false
			}
			value, ok := vertex.Metadata.MappedMetadata[parts[2]]
			return value, ok
		}
	case "kind":
		return vertex.VertexKind, true
	}
	return nil, false
}

func projectedEdgeField(edge Edge, fieldPath string) (any, bool) {
	parts := strings.Split(fieldPath, ".")
	switch parts[0] {
	case "projected":
		if len(parts) < 2 {
			return nil, false
		}
		switch parts[1] {
		case "edge_id":
			return edge.EdgeID, true
		case "src_vertex_id":
			return edge.SrcVertexID, true
		case "dst_vertex_id":
			return edge.DstVertexID, true
		case "direction":
			return edge.Direction, true
		case "properties":
			if len(parts) != 3 {
				return nil, false
			}
			value, ok := edge.Properties[parts[2]]
			return value, ok
		case "metadata":
			if len(parts) != 3 {
				return nil, false
			}
			value, ok := edge.Metadata.MappedMetadata[parts[2]]
			return value, ok
		}
	case "kind":
		return edge.EdgeKind, true
	}
	return nil, false
}

func endpointDigest(ruleID, targetScope, projectedKind string, keys []string, contributors []contributor) (string, bool) {
	if len(contributors) == 0 {
		return "", false
	}
	values := make([]any, 0, len(keys))
	for _, key := range keys {
		value, ok := contributorField(contributors[0], key)
		if !ok {
			return "", false
		}
		values = append(values, value)
	}
	keyJSON, _ := canonicalJSON(values)
	digest, _ := digestTuple("GPGROUP1\n", ruleID, targetScope, projectedKind, keyJSON)
	return digest, true
}

func buildSchemaRegistry(run ProjectionRun, vertices []Vertex, edges []Edge) SchemaRegistry {
	vertexKinds := map[string]*VertexKindSchema{}
	for _, mapping := range run.Request.ProjectionConfig.EntityMappings {
		item := ensureVertexKind(vertexKinds, mapping.ProjectedVertexKind)
		item.SourceEntityKinds = append(item.SourceEntityKinds, mapping.SourceEntityKind)
		item.Labels = append(item.Labels, run.Request.ProjectionConfig.DefaultVertexLabels...)
		item.Labels = append(item.Labels, mapping.MappingLabels...)
		if mapping.LabelPolicy == "preserve_source" || mapping.LabelPolicy == "mapping_then_source" {
			item.SourceLabelsPreserved = true
		}
	}
	for _, rule := range run.Request.ProjectionConfig.AggregationRules {
		if rule.TargetScope == "vertex" {
			item := ensureVertexKind(vertexKinds, rule.ProjectedKind)
			item.AggregationRuleIDs = append(item.AggregationRuleIDs, rule.AggregationRuleID)
			item.Labels = append(item.Labels, run.Request.ProjectionConfig.DefaultVertexLabels...)
		}
	}
	edgeKinds := map[string]*EdgeKindSchema{}
	for _, mapping := range run.Request.RelationshipMappings {
		item := ensureEdgeKind(edgeKinds, mapping.ProjectedEdgeKind)
		item.SourceRelationshipKinds = append(item.SourceRelationshipKinds, mapping.SourceRelationshipKind)
		item.Directions = append(item.Directions, projectedDirectionsForPolicy(mapping.DirectionPolicy)...)
		item.Labels = append(item.Labels, run.Request.ProjectionConfig.DefaultEdgeLabels...)
		item.Labels = append(item.Labels, mapping.MappingLabels...)
		if mapping.LabelPolicy == "preserve_source" || mapping.LabelPolicy == "mapping_then_source" {
			item.SourceLabelsPreserved = true
		}
		if mapping.EmitReverseEdge {
			reverse := ensureEdgeKind(edgeKinds, mapping.ReverseEdgeKind)
			reverse.SourceRelationshipKinds = append(reverse.SourceRelationshipKinds, mapping.SourceRelationshipKind)
			reverse.Directions = append(reverse.Directions, "directed")
			reverse.Labels = append(reverse.Labels, item.Labels...)
		}
	}
	for _, rule := range run.Request.ProjectionConfig.AggregationRules {
		if rule.TargetScope == "edge" {
			item := ensureEdgeKind(edgeKinds, rule.ProjectedKind)
			item.AggregationRuleIDs = append(item.AggregationRuleIDs, rule.AggregationRuleID)
			item.Directions = append(item.Directions, rule.EdgeDirection)
			item.Labels = append(item.Labels, run.Request.ProjectionConfig.DefaultEdgeLabels...)
		}
	}
	for _, definition := range run.Request.PropertyDefinitions {
		if definition.TargetScope == "vertex" {
			for key, item := range vertexKinds {
				if definition.TargetKind == "*" || definition.TargetKind == key {
					item.Properties = append(item.Properties, definition.ProjectedKey)
				}
			}
		}
		if definition.TargetScope == "edge" {
			for key, item := range edgeKinds {
				if definition.TargetKind == "*" || definition.TargetKind == key {
					item.Properties = append(item.Properties, definition.ProjectedKey)
				}
			}
		}
	}
	registry := SchemaRegistry{}
	for _, item := range vertexKinds {
		item.SourceEntityKinds = uniqueSortedStrings(item.SourceEntityKinds)
		item.AggregationRuleIDs = uniqueSortedStrings(item.AggregationRuleIDs)
		item.Labels = uniqueSortedStrings(item.Labels)
		item.Properties = uniqueSortedStrings(item.Properties)
		registry.VertexKinds = append(registry.VertexKinds, *item)
	}
	for _, item := range edgeKinds {
		item.SourceRelationshipKinds = uniqueSortedStrings(item.SourceRelationshipKinds)
		item.AggregationRuleIDs = uniqueSortedStrings(item.AggregationRuleIDs)
		item.Directions = sortDirections(item.Directions)
		item.Labels = uniqueSortedStrings(item.Labels)
		item.Properties = uniqueSortedStrings(item.Properties)
		registry.EdgeKinds = append(registry.EdgeKinds, *item)
	}
	for _, definition := range run.Request.PropertyDefinitions {
		registry.PropertyKeys = append(registry.PropertyKeys, PropertySchema{TargetScope: definition.TargetScope, TargetKind: definition.TargetKind, ProjectedKey: definition.ProjectedKey, ProjectedType: definition.ProjectedType, Required: definition.Required})
	}
	for _, mapping := range run.Request.ProjectionConfig.MetadataMappings {
		registry.MetadataKeys = append(registry.MetadataKeys, MetadataSchema{TargetScope: mapping.TargetScope, TargetKind: mapping.TargetKind, ProjectedMetadataKey: mapping.ProjectedMetadataKey, ProjectedType: mapping.ProjectedType, Required: mapping.Required})
	}
	sort.Slice(registry.VertexKinds, func(i, j int) bool { return registry.VertexKinds[i].VertexKind < registry.VertexKinds[j].VertexKind })
	sort.Slice(registry.EdgeKinds, func(i, j int) bool { return registry.EdgeKinds[i].EdgeKind < registry.EdgeKinds[j].EdgeKind })
	sort.Slice(registry.PropertyKeys, func(i, j int) bool {
		a, b := registry.PropertyKeys[i], registry.PropertyKeys[j]
		return a.TargetScope+"|"+a.TargetKind+"|"+a.ProjectedKey < b.TargetScope+"|"+b.TargetKind+"|"+b.ProjectedKey
	})
	return registry
}

func ensureVertexKind(items map[string]*VertexKindSchema, kind string) *VertexKindSchema {
	if items[kind] == nil {
		items[kind] = &VertexKindSchema{VertexKind: kind}
	}
	return items[kind]
}

func ensureEdgeKind(items map[string]*EdgeKindSchema, kind string) *EdgeKindSchema {
	if items[kind] == nil {
		items[kind] = &EdgeKindSchema{EdgeKind: kind}
	}
	return items[kind]
}

func (run ProjectionRun) issue(severity, code, targetKind, targetID string, field any, details map[string]any) ValidationIssue {
	if details == nil {
		details = map[string]any{}
	}
	var fieldPtr *string
	if fieldString, ok := field.(string); ok && fieldString != "" {
		fieldPtr = &fieldString
	}
	issueID, _ := generatedID("gpi_", "GPISSUE1\n", ProjectionSchemaID, run.GraphViewID, run.ProjectionRunID, severity, code, targetKind, targetID, details)
	return ValidationIssue{
		IssueID:    issueID,
		Severity:   severity,
		Code:       code,
		TargetKind: targetKind,
		TargetID:   targetID,
		Field:      fieldPtr,
		Message:    code,
		Details:    details,
	}
}

func validationSummary(issues []ValidationIssue) ValidationSummary {
	sort.Slice(issues, func(i, j int) bool {
		left, right := issues[i], issues[j]
		leftKey := left.Severity + "|" + left.Code + "|" + left.TargetKind + "|" + left.TargetID + "|" + left.IssueID
		rightKey := right.Severity + "|" + right.Code + "|" + right.TargetKind + "|" + right.TargetID + "|" + right.IssueID
		return leftKey < rightKey
	})
	status := "valid"
	if len(issues) > 0 {
		status = "warnings"
	}
	for _, issue := range issues {
		if issue.Severity == "error" {
			status = "errors"
		}
		if issue.Severity == "fatal" {
			status = "failed"
			break
		}
	}
	return ValidationSummary{Status: status, IssueCount: len(issues), Issues: issues}
}

func hasFatalIssue(issues []ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == "fatal" {
			return true
		}
	}
	return false
}

func sortVertices(vertices []Vertex) {
	sort.Slice(vertices, func(i, j int) bool {
		if vertices[i].SortKey == vertices[j].SortKey {
			return vertices[i].VertexID < vertices[j].VertexID
		}
		return vertices[i].SortKey < vertices[j].SortKey
	})
}

func sortEdges(edges []Edge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].SortKey == edges[j].SortKey {
			return edges[i].EdgeID < edges[j].EdgeID
		}
		return edges[i].SortKey < edges[j].SortKey
	})
}

func sortKey(fields ...any) string {
	digest, _ := digestTuple("GPSORT1\n", fields...)
	return "sk_" + digest
}

func contributorSortKey(fields ...any) string {
	digest, _ := digestTuple("GPCONTRIB1\n", fields...)
	return "csk_" + digest
}

func sourceRefs(contributors []contributor) []SourceRef {
	refs := make([]SourceRef, 0, len(contributors))
	for _, contributor := range contributors {
		refs = append(refs, SourceRef{RefKind: contributor.Kind, RefID: contributor.ID, ContributorSortKey: contributor.SortKey})
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].ContributorSortKey == refs[j].ContributorSortKey {
			return refs[i].RefKind+"|"+refs[i].RefID < refs[j].RefKind+"|"+refs[j].RefID
		}
		return refs[i].ContributorSortKey < refs[j].ContributorSortKey
	})
	return refs
}

func entityIDs(entities []SourceEntity) []string {
	out := make([]string, 0, len(entities))
	for _, entity := range entities {
		out = append(out, entity.SourceEntityID)
	}
	return out
}

func relationshipIDs(relationships []SourceRelationship) []string {
	out := make([]string, 0, len(relationships))
	for _, relationship := range relationships {
		out = append(out, relationship.SourceRelationshipID)
	}
	return out
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func validFieldPath(path string) bool {
	if path == "" {
		return false
	}
	parts := strings.Split(path, ".")
	switch parts[0] {
	case "source_entity_id", "source_relationship_id", "src_source_entity_id", "dst_source_entity_id", "kind", "direction":
		return len(parts) == 1
	case "properties", "metadata", "source_metadata":
		return len(parts) == 2 && validPropertyKey(parts[1])
	case "projected":
		return len(parts) >= 2
	default:
		return false
	}
}

func validSourceDirection(direction string) bool {
	switch direction {
	case "forward", "reverse", "undirected", "bidirectional":
		return true
	default:
		return false
	}
}

func projectDirection(policy, sourceDirection string, srcVertex, dstVertex Vertex) (Vertex, Vertex, string) {
	switch policy {
	case "normalize_reverse":
		return dstVertex, srcVertex, "directed"
	case "undirected":
		return srcVertex, dstVertex, "undirected"
	case "bidirectional":
		return srcVertex, dstVertex, "bidirectional"
	case "preserve":
		switch sourceDirection {
		case "reverse":
			return dstVertex, srcVertex, "directed"
		case "undirected":
			return srcVertex, dstVertex, "undirected"
		case "bidirectional":
			return srcVertex, dstVertex, "bidirectional"
		}
	}
	return srcVertex, dstVertex, "directed"
}

func deriveLabels(defaults, mappingLabels, sourceLabels []string, policy string) []string {
	labels := append([]string{}, defaults...)
	switch policy {
	case "preserve_source":
		labels = append(labels, sourceLabels...)
	case "mapping_then_source":
		labels = append(labels, mappingLabels...)
		labels = append(labels, sourceLabels...)
	default:
		labels = append(labels, mappingLabels...)
	}
	return uniqueSortedStrings(labels)
}

func valueMatchesType(projectedType string, value any) bool {
	switch projectedType {
	case "string", "timestamp", "identifier":
		_, ok := value.(string)
		if projectedType == "identifier" && ok {
			return validIdentifier(value.(string))
		}
		return ok
	case "integer":
		switch typed := value.(type) {
		case int, int64:
			return true
		case fmt.Stringer:
			return finiteIntegerPattern.MatchString(typed.String())
		default:
			return finiteIntegerPattern.MatchString(fmt.Sprint(value))
		}
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "string_array", "identifier_array":
		array, ok := value.([]any)
		if !ok {
			return false
		}
		for _, entry := range array {
			stringValue, ok := entry.(string)
			if !ok {
				return false
			}
			if projectedType == "identifier_array" && !validIdentifier(stringValue) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func propertyApplies(definition PropertyDefinition, targetScope, targetKind string) bool {
	return definition.TargetScope == targetScope && (definition.TargetKind == "*" || definition.TargetKind == targetKind)
}

func filtersMatchEntity(filters []FilterPredicate, entity SourceEntity) bool {
	for _, filter := range filters {
		value, ok := sourceField(entity, nil, nil, filter.FieldPath)
		if !filterMatches(value, ok, filter) {
			return false
		}
	}
	return true
}

func filtersMatchRelationship(filters []FilterPredicate, relationship SourceRelationship) bool {
	for _, filter := range filters {
		value, ok := sourceField(SourceEntity{}, &relationship, nil, filter.FieldPath)
		if !filterMatches(value, ok, filter) {
			return false
		}
	}
	return true
}

func filterMatches(value any, present bool, filter FilterPredicate) bool {
	switch filter.Operator {
	case "exists":
		return present
	case "not_exists":
		return !present
	case "eq":
		return present && canonicalValueKey(value) == canonicalValueKey(filter.Value)
	case "neq":
		return !present || canonicalValueKey(value) != canonicalValueKey(filter.Value)
	default:
		return true
	}
}

func projectedDirectionsForPolicy(policy string) []string {
	switch policy {
	case "undirected":
		return []string{"undirected"}
	case "bidirectional":
		return []string{"bidirectional"}
	default:
		return []string{"directed", "undirected", "bidirectional"}
	}
}

func sortDirections(values []string) []string {
	order := map[string]int{"directed": 0, "undirected": 1, "bidirectional": 2}
	values = uniqueSortedStrings(values)
	sort.Slice(values, func(i, j int) bool { return order[values[i]] < order[values[j]] })
	return values
}

func defaultConsumerCapabilities() ConsumerCapabilities {
	return ConsumerCapabilities{
		QueryShapes:                     []string{"get_edge", "get_graph_view", "get_projection_run", "get_vertex", "list_graph_views", "traverse"},
		SupportsDirectVertexLookup:      true,
		SupportsDirectEdgeLookup:        true,
		SupportsBreadthFirstTraversal:   true,
		SupportsAlternateTraversalOrder: []string{},
		MaxTraversalDepth:               16,
		MaxTraversalSeedVertices:        1024,
		MaxKindFilters:                  1024,
	}
}

func formatLifecycleTimestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000000Z")
}
