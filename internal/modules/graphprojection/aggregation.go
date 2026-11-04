package graphprojection

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math"
	"sort"
	"strings"
)

func (state *projectionState) emitAggregations(ctx context.Context, run *projectionWork) error {
	for _, rule := range run.Request.projectionConfig.AggregationRules {
		if err := ctx.Err(); err != nil {
			return err
		}
		// Snapshot contributors before the rule emits; only subsequent rules
		// can consume its output, regardless of either rule's target scope.
		groups, err := groupContributors(ctx, run, rule, state.vertices, state.edges)
		if err != nil {
			return err
		}
		if rule.TargetScope == "vertex" {
			err = state.emitVertices(ctx, run, rule, groups)
		} else {
			err = state.emitEdges(ctx, run, rule, groups)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (state *projectionState) emitVertices(ctx context.Context, run *projectionWork, rule aggregationRule, groups map[string][]contributor) error {
	for _, digest := range sortedGroupDigests(groups) {
		if run.issues.err != nil {
			return run.issues.err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		contributors := groups[digest]
		err := state.emitVertex(ctx, &run.issues, digest, func() (Vertex, error) {
			vertexID, _ := generatedID("vx_", "GPVERTEX1\n", "aggregated_vertex", ProjectionSchemaIDV2, run.GraphViewID, rule.AggregationIdentityDigest, digest)
			props, err := mergeAggregateProperties(run, rule, digest, contributors, "vertex", rule.ProjectedKind, vertexID)
			if err != nil {
				return Vertex{}, err
			}
			mappedMetadata, err := mergeAggregateMetadata(run, rule, digest, contributors, "vertex", rule.ProjectedKind, vertexID)
			if err != nil {
				return Vertex{}, err
			}
			ruleID := rule.AggregationRuleID
			vertex := Vertex{
				VertexID:     vertexID,
				VertexKind:   rule.ProjectedKind,
				VertexFamily: "aggregated",
				Labels:       run.Request.projectionConfig.DefaultVertexLabels,
				Properties:   props,
				Metadata: VertexMetadata{
					AggregationRuleID:     &ruleID,
					AggregationSourceRefs: sourceRefs(contributors),
					MappedMetadata:        mappedMetadata,
				},
				SortKey: sortKey("vertex", "aggregated", rule.ProjectedKind, rule.AggregationRuleID, digest, vertexID),
			}
			return vertex, run.issues.err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (state *projectionState) emitEdges(ctx context.Context, run *projectionWork, rule aggregationRule, groups map[string][]contributor) error {
	for _, digest := range sortedGroupDigests(groups) {
		if run.issues.err != nil {
			return run.issues.err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		contributors := groups[digest]
		srcRuleID := rule.endpointGrouping.SourceVertexAggregationRuleID
		dstRuleID := rule.endpointGrouping.DestinationVertexAggregationRuleID
		pair, missing, err := resolveEndpointPair(ctx, rule, state.kinds, contributors)
		if err != nil {
			return err
		}
		src := state.verticesByRuleAndDigest[srcRuleID][pair[0]]
		dst := state.verticesByRuleAndDigest[dstRuleID][pair[1]]
		if missing[0] != "" || missing[1] != "" || src.VertexID == "" || dst.VertexID == "" {
			if rule.endpointGrouping.MissingEndpointBehavior == "error" {
				if missing[0] != "" || pair[0] != "" && src.VertexID == "" {
					recordMissingEndpoint(run, rule, "src", pair[0], missing[0])
				}
				if missing[1] != "" || pair[1] != "" && dst.VertexID == "" {
					recordMissingEndpoint(run, rule, "dst", pair[1], missing[1])
				}
			}
			continue
		}
		err = state.emitEdge(ctx, &run.issues, func() (Edge, error) {
			edgeID, _ := generatedID("ed_", "GPEDGE1\n", "aggregated_edge", ProjectionSchemaIDV2, run.GraphViewID, rule.AggregationIdentityDigest, src.VertexID, dst.VertexID, rule.EdgeDirection, digest)
			props, err := mergeAggregateProperties(run, rule, digest, contributors, "edge", rule.ProjectedKind, edgeID)
			if err != nil {
				return Edge{}, err
			}
			mappedMetadata, err := mergeAggregateMetadata(run, rule, digest, contributors, "edge", rule.ProjectedKind, edgeID)
			if err != nil {
				return Edge{}, err
			}
			ruleID := rule.AggregationRuleID
			edge := Edge{
				EdgeID:      edgeID,
				EdgeKind:    rule.ProjectedKind,
				EdgeFamily:  "aggregated",
				SrcVertexID: src.VertexID,
				DstVertexID: dst.VertexID,
				Direction:   rule.EdgeDirection,
				Labels:      run.Request.projectionConfig.DefaultEdgeLabels,
				Properties:  props,
				Metadata: EdgeMetadata{
					AggregationRuleID:     &ruleID,
					AggregationSourceRefs: sourceRefs(contributors),
					MappedMetadata:        mappedMetadata,
				},
				SortKey: sortKey("edge", "aggregated", rule.ProjectedKind, rule.AggregationRuleID, src.VertexID, dst.VertexID, rule.EdgeDirection, digest, edgeID),
			}
			return edge, run.issues.err
		})
		if err != nil {
			return err
		}
	}
	return run.issues.err
}

func sortedGroupDigests(groups map[string][]contributor) []string {
	digests := make([]string, 0, len(groups))
	for digest := range groups {
		digests = append(digests, digest)
	}
	sort.Strings(digests)
	return digests
}

func recordMissingEndpoint(run *projectionWork, rule aggregationRule, side, digest, missingField string) {
	if missingField != "" {
		run.recordIssue("error", "aggregation_endpoint_missing", "mapping_rule", rule.AggregationRuleID, missingField, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "endpoint_side": side, "reason_code": "endpoint_key_missing", "endpoint_digest": nil, "field_path": missingField})
		return
	}
	run.recordIssue("error", "aggregation_endpoint_missing", "mapping_rule", rule.AggregationRuleID, nil, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "endpoint_side": side, "reason_code": "endpoint_vertex_not_found", "endpoint_digest": digest, "field_path": nil})
}
func aggregationProjectedKinds(rules []aggregationRule) map[string]string {
	out := map[string]string{}
	for _, rule := range rules {
		out[rule.AggregationRuleID] = rule.ProjectedKind
	}
	return out
}

type contributor struct {
	Kind         string
	ID           string
	KindName     string
	Order        contributorOrder
	Entity       *sourceEntity
	Relationship *sourceRelationship
	Vertex       *Vertex
	Edge         *Edge
}

func groupContributors(ctx context.Context, run *projectionWork, rule aggregationRule, vertices []Vertex, edges []Edge) (map[string][]contributor, error) {
	groups := map[string][]contributor{}
	contributors, err := contributorsForRule(ctx, run, rule, vertices, edges)
	if err != nil {
		return nil, err
	}
	for _, contributor := range contributors {
		if err := itemCheckpointV2(ctx); err != nil {
			return nil, err
		}
		values := make([]any, 0, len(rule.GroupingKeys))
		missing := false
		for _, fieldPath := range rule.GroupingKeys {
			value, ok := contributorField(contributor, fieldPath)
			if !ok {
				if rule.MissingGroupingKeyBehavior == "use_null" {
					values = append(values, nil)
					continue
				}
				missing = true
				if rule.MissingGroupingKeyBehavior == "error" {
					run.recordIssue("error", "aggregation_grouping_key_missing", "source_or_projected_item", contributor.ID, fieldPath, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "field_path": fieldPath, "contributor_id": contributor.ID})
				}
				break
			}
			values = append(values, value)
		}
		if run.issues.err != nil {
			return nil, run.issues.err
		}
		if missing {
			continue
		}
		digest, err := groupingTupleDigest(rule.AggregationRuleID, rule.TargetScope, rule.ProjectedKind, values)
		if err != nil {
			return nil, err
		}
		groups[digest] = append(groups[digest], contributor)
	}
	for digest := range groups {
		sort.Slice(groups[digest], func(i, j int) bool {
			return groups[digest][i].Order.less(groups[digest][j].Order)
		})
	}
	return groups, nil
}

func contributorsForRule(ctx context.Context, run *projectionWork, rule aggregationRule, vertices []Vertex, edges []Edge) ([]contributor, error) {
	out := []contributor{}
	add := func(entry contributor) error {
		order, err := entry.orderTuple()
		if err != nil {
			return err
		}
		entry.Order = order
		out = append(out, entry)
		return nil
	}
	declaredEntityKinds := stringSet(run.Request.projectionConfig.DeclaredSourceEntityKinds)
	declaredRelationshipKinds := stringSet(run.Request.projectionConfig.DeclaredSourceRelationshipKinds)
	switch rule.InputScope {
	case "source_entity":
		for i := range run.Request.SourceEntities {
			if err := itemCheckpointV2(ctx); err != nil {
				return nil, err
			}
			entity := &run.Request.SourceEntities[i]
			if entity.SourceEntityKind != rule.InputKind || !validIdentifierV2(entity.SourceEntityID) || !validIdentifierV2(entity.SourceEntityKind) ||
				!sourceItemWithinLimits(len(entity.Labels), len(entity.Properties), len(entity.Metadata)) ||
				!sourceItemValuesWithinLimits(entity.Properties, entity.Metadata) ||
				!declaredEntityKinds[entity.SourceEntityKind] ||
				!filtersMatchEntity(run.Request.filters.EntityFilters, *entity) {
				continue
			}
			if err := add(contributor{Kind: "source_entity", ID: entity.SourceEntityID, KindName: entity.SourceEntityKind, Entity: entity}); err != nil {
				return nil, err
			}
		}
	case "source_relationship":
		for i := range run.Request.SourceRelationships {
			if err := itemCheckpointV2(ctx); err != nil {
				return nil, err
			}
			relationship := &run.Request.SourceRelationships[i]
			if relationship.SourceRelationshipKind != rule.InputKind || !validIdentifierV2(relationship.SourceRelationshipID) || !validIdentifierV2(relationship.SourceRelationshipKind) ||
				!sourceItemWithinLimits(len(relationship.Labels), len(relationship.Properties), len(relationship.Metadata)) ||
				!sourceItemValuesWithinLimits(relationship.Properties, relationship.Metadata) ||
				!declaredRelationshipKinds[relationship.SourceRelationshipKind] ||
				relationship.SrcSourceEntityID == "" || relationship.DstSourceEntityID == "" || !validSourceDirection(relationship.Direction) ||
				!filtersMatchRelationship(run.Request.filters.RelationshipFilters, *relationship) {
				continue
			}
			if err := add(contributor{Kind: "source_relationship", ID: relationship.SourceRelationshipID, KindName: relationship.SourceRelationshipKind, Relationship: relationship}); err != nil {
				return nil, err
			}
		}
	case "projected_vertex":
		for i := range vertices {
			if err := itemCheckpointV2(ctx); err != nil {
				return nil, err
			}
			vertex := &vertices[i]
			if vertex.VertexKind != rule.InputKind {
				continue
			}
			if err := add(contributor{Kind: "projected_vertex", ID: vertex.VertexID, KindName: vertex.VertexKind, Vertex: vertex}); err != nil {
				return nil, err
			}
		}
	case "projected_edge":
		for i := range edges {
			if err := itemCheckpointV2(ctx); err != nil {
				return nil, err
			}
			edge := &edges[i]
			if edge.EdgeKind != rule.InputKind {
				continue
			}
			if err := add(contributor{Kind: "projected_edge", ID: edge.EdgeID, KindName: edge.EdgeKind, Edge: edge}); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func mergeAggregateProperties(run *projectionWork, rule aggregationRule, groupingDigest string, contributors []contributor, targetScope, targetKind, outputID string) (map[string]any, error) {
	properties := map[string]any{}
	for _, definition := range run.Request.PropertyDefinitions {
		if run.issues.err != nil {
			return nil, run.issues.err
		}
		if !propertyApplies(definition, targetScope, targetKind) {
			continue
		}
		mergeBehavior := definition.MergeBehavior
		if override := rule.PropertyMergeBehavior[definition.ProjectedKey]; override != "" {
			mergeBehavior = override
		}
		merge := projectedMerge{behavior: mergeBehavior}
		for _, contributor := range contributors {
			if run.issues.err != nil {
				return nil, run.issues.err
			}
			value, found := contributorField(contributor, definition.SourceFieldPath)
			normalized, include, issueCode := evaluateCandidate(candidateDefinitionFromProperty(definition), value, found)
			if issueCode != "" {
				if issueCode == "required_property_missing" {
					run.recordIssue("error", "required_property_missing", "property", outputID, definition.SourceFieldPath, map[string]any{"projected_key": definition.ProjectedKey, "source_field_path": definition.SourceFieldPath, "output_object_id": outputID, "aggregation_rule_id": rule.AggregationRuleID})
				} else {
					propertyValueIssue(run, issueCode, definition.ProjectedKey, definition.ProjectedType, definition.SourceFieldPath, outputID, value, rule.AggregationRuleID, groupingDigest, contributor.ID)
				}
				continue
			}
			if include {
				if err := merge.add(normalized); err != nil {
					return nil, err
				}
			}
		}
		merged, ok, conflict := merge.result(definition.ProjectedType)
		if conflict {
			run.recordIssue("error", "aggregation_merge_conflict", "mapping_rule", rule.AggregationRuleID, nil, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "canonical_grouping_key_digest": groupingDigest, "projected_key": definition.ProjectedKey})
			continue
		}
		if ok {
			properties[definition.ProjectedKey] = merged
		}
	}
	return properties, run.issues.err
}

func mergeAggregateMetadata(run *projectionWork, rule aggregationRule, groupingDigest string, contributors []contributor, targetScope, targetKind, outputID string) (map[string]any, error) {
	metadata := map[string]any{}
	for _, mapping := range run.Request.projectionConfig.MetadataMappings {
		if run.issues.err != nil {
			return nil, run.issues.err
		}
		if mapping.TargetScope != targetScope || (mapping.TargetKind != "*" && mapping.TargetKind != targetKind) {
			continue
		}
		merge := projectedMerge{behavior: mapping.MergeBehavior}
		for _, contributor := range contributors {
			if run.issues.err != nil {
				return nil, run.issues.err
			}
			value, found := contributorField(contributor, mapping.SourceFieldPath)
			normalized, include, issueCode := evaluateCandidate(candidateDefinitionFromMetadata(mapping), value, found)
			if issueCode != "" {
				if issueCode == "required_property_missing" {
					run.recordIssue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, mapping.SourceFieldPath, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "required_metadata_missing"})
				} else {
					propertyValueIssue(run, issueCode, mapping.ProjectedMetadataKey, mapping.ProjectedType, mapping.SourceFieldPath, outputID, value, rule.AggregationRuleID, groupingDigest, contributor.ID)
				}
				continue
			}
			if include {
				if err := merge.add(normalized); err != nil {
					return nil, err
				}
			}
		}
		merged, ok, conflict := merge.result(mapping.ProjectedType)
		if conflict {
			run.recordIssue("error", "aggregation_merge_conflict", "mapping_rule", rule.AggregationRuleID, nil, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "canonical_grouping_key_digest": groupingDigest, "projected_key": mapping.ProjectedMetadataKey})
			continue
		}
		if ok {
			metadata[mapping.ProjectedMetadataKey] = merged
		}
	}
	return metadata, run.issues.err
}

func deriveProperties(run *projectionWork, targetScope, targetKind string, entity sourceEntity, relationship *sourceRelationship, graphSource map[string]any, outputID string) map[string]any {
	properties := map[string]any{}
	for _, definition := range run.Request.PropertyDefinitions {
		if run.issues.err != nil {
			return nil
		}
		if !propertyApplies(definition, targetScope, targetKind) {
			continue
		}
		value, found := sourceField(entity, relationship, graphSource, definition.SourceFieldPath)
		normalized, include, issueCode := evaluateCandidate(candidateDefinitionFromProperty(definition), value, found)
		if issueCode != "" {
			if issueCode == "required_property_missing" {
				run.recordIssue("error", issueCode, "property", outputID+"#"+definition.ProjectedKey, definition.SourceFieldPath, map[string]any{"projected_key": definition.ProjectedKey, "source_field_path": definition.SourceFieldPath, "output_object_id": outputID})
			} else {
				propertyValueIssue(run, issueCode, definition.ProjectedKey, definition.ProjectedType, definition.SourceFieldPath, outputID, value, "", "", "")
			}
			continue
		}
		if include {
			properties[definition.ProjectedKey] = normalized
		}
	}
	return properties
}

func deriveMetadata(run *projectionWork, targetScope, targetKind string, entity sourceEntity, relationship *sourceRelationship, outputID string) map[string]any {
	metadata := map[string]any{}
	for _, mapping := range run.Request.projectionConfig.MetadataMappings {
		if run.issues.err != nil {
			return nil
		}
		if mapping.TargetScope != targetScope || (mapping.TargetKind != "*" && mapping.TargetKind != targetKind) {
			continue
		}
		value, found := sourceField(entity, relationship, run.Request.SourceMetadata, mapping.SourceFieldPath)
		normalized, include, issueCode := evaluateCandidate(candidateDefinitionFromMetadata(mapping), value, found)
		if issueCode != "" {
			if issueCode == "required_property_missing" {
				run.recordIssue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, mapping.SourceFieldPath, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "required_metadata_missing"})
			} else {
				propertyValueIssue(run, issueCode, mapping.ProjectedMetadataKey, mapping.ProjectedType, mapping.SourceFieldPath, outputID, value, "", "", "")
			}
			continue
		}
		if include {
			metadata[mapping.ProjectedMetadataKey] = normalized
		}
	}
	return metadata
}

type candidateDefinition struct {
	ProjectedType      string
	DefaultValue       any
	HasDefaultValue    bool
	MissingBehavior    string
	SourceNullBehavior string
	NullOutputPolicy   string
}

func candidateDefinitionFromProperty(definition propertyDefinition) candidateDefinition {
	return candidateDefinition{ProjectedType: definition.ProjectedType, DefaultValue: definition.DefaultValue, HasDefaultValue: definition.HasDefaultValue, MissingBehavior: definition.MissingBehavior, SourceNullBehavior: definition.SourceNullBehavior, NullOutputPolicy: definition.NullOutputPolicy}
}

func candidateDefinitionFromMetadata(mapping metadataMapping) candidateDefinition {
	return candidateDefinition{ProjectedType: mapping.ProjectedType, DefaultValue: mapping.DefaultValue, HasDefaultValue: mapping.HasDefaultValue, MissingBehavior: mapping.MissingBehavior, SourceNullBehavior: mapping.SourceNullBehavior, NullOutputPolicy: mapping.NullOutputPolicy}
}

func evaluateCandidate(definition candidateDefinition, value any, found bool) (any, bool, string) {
	usedDefault := false
	if !found {
		switch definition.MissingBehavior {
		case "default":
			if !definition.HasDefaultValue {
				return nil, false, "required_property_missing"
			}
			value = definition.DefaultValue
			usedDefault = true
		case "error":
			return nil, false, "required_property_missing"
		default:
			return nil, false, ""
		}
	}
	if value == nil && !usedDefault {
		switch definition.SourceNullBehavior {
		case "default":
			if !definition.HasDefaultValue {
				return nil, false, "source_null_for_required_property"
			}
			value = definition.DefaultValue
			usedDefault = true
		case "emit_null":
			return nil, definition.NullOutputPolicy == "emit_null", ""
		case "error":
			return nil, false, "source_null_for_required_property"
		default:
			return nil, false, ""
		}
	}
	if value == nil {
		return nil, definition.NullOutputPolicy == "emit_null", ""
	}
	normalized, valid := normalizeProjectedValueV2(definition.ProjectedType, value)
	if !valid {
		return nil, false, "invalid_property_type"
	}
	return normalized, true, ""
}

func propertyValueIssue(run *projectionWork, code, projectedKey, expectedType, sourceFieldPath, outputID string, value any, aggregationRuleID, groupingDigest, contributorID string) {
	details := map[string]any{"projected_key": projectedKey, "expected_type": expectedType, "actual_type": jsonValueType(value), "source_field_path": sourceFieldPath, "output_object_id": outputID}
	if aggregationRuleID != "" {
		details["aggregation_rule_id"] = aggregationRuleID
		details["canonical_grouping_key_digest"] = groupingDigest
	}
	if contributorID != "" {
		details["contributor_id"] = contributorID
	}
	run.recordIssue("error", code, "property", outputID+"#"+projectedKey, sourceFieldPath, details)
}

func jsonValueType(value any) string {
	if value == nil {
		return "null"
	}
	switch typed := value.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case []any, []string:
		return "array"
	case map[string]any:
		return "object"
	case int, int64:
		return "integer"
	case json.Number:
		if validFiniteInteger(typed.String()) {
			return "integer"
		}
		return "number"
	default:
		return "number"
	}
}

func sourceField(entity sourceEntity, relationship *sourceRelationship, graphSource map[string]any, path string) (any, bool) {
	field, ok := classifyFieldPath(path)
	if !ok {
		return nil, false
	}
	if field.scopes == fieldGraphMetadata {
		value, found := graphSource[field.key]
		return value, found
	}
	if relationship != nil {
		if field.scopes&fieldSourceRelationship == 0 {
			return nil, false
		}
		switch field.member {
		case "source_relationship_id":
			return relationship.SourceRelationshipID, true
		case "source_relationship_kind":
			return relationship.SourceRelationshipKind, relationship.SourceRelationshipKind != ""
		case "src_source_entity_id":
			return relationship.SrcSourceEntityID, relationship.SrcSourceEntityID != ""
		case "dst_source_entity_id":
			return relationship.DstSourceEntityID, relationship.DstSourceEntityID != ""
		case "direction":
			return relationship.Direction, true
		case "properties":
			value, found := relationship.Properties[field.key]
			return value, found
		case "metadata":
			value, found := relationship.Metadata[field.key]
			return value, found
		}
	} else {
		if field.scopes&fieldSourceEntity == 0 {
			return nil, false
		}
		switch field.member {
		case "source_entity_id":
			return entity.SourceEntityID, true
		case "source_entity_kind":
			return entity.SourceEntityKind, entity.SourceEntityKind != ""
		case "properties":
			value, found := entity.Properties[field.key]
			return value, found
		case "metadata":
			value, found := entity.Metadata[field.key]
			return value, found
		}
	}
	return nil, false
}

func contributorField(entry contributor, fieldPath string) (any, bool) {
	if entry.Entity != nil {
		return sourceField(*entry.Entity, nil, nil, fieldPath)
	}
	if entry.Relationship != nil {
		return sourceField(sourceEntity{}, entry.Relationship, nil, fieldPath)
	}
	if entry.Vertex != nil {
		return projectedVertexField(*entry.Vertex, fieldPath)
	}
	if entry.Edge != nil {
		return projectedEdgeField(*entry.Edge, fieldPath)
	}
	return nil, false
}

func projectedVertexField(vertex Vertex, path string) (any, bool) {
	field, ok := classifyFieldPath(path)
	if !ok || field.scopes&fieldProjectedVertex == 0 {
		return nil, false
	}
	switch field.member {
	case "vertex_id":
		return vertex.VertexID, true
	case "properties":
		value, found := vertex.Properties[field.key]
		return value, found
	case "metadata":
		value, found := vertex.Metadata.MappedMetadata[field.key]
		return value, found
	}
	return nil, false
}

func projectedEdgeField(edge Edge, path string) (any, bool) {
	field, ok := classifyFieldPath(path)
	if !ok || field.scopes&fieldProjectedEdge == 0 {
		return nil, false
	}
	switch field.member {
	case "edge_id":
		return edge.EdgeID, true
	case "src_vertex_id":
		return edge.SrcVertexID, true
	case "dst_vertex_id":
		return edge.DstVertexID, true
	case "direction":
		return edge.Direction, true
	case "properties":
		value, found := edge.Properties[field.key]
		return value, found
	case "metadata":
		value, found := edge.Metadata.MappedMetadata[field.key]
		return value, found
	}
	return nil, false
}

// Compare complete ordered pairs. A missing member cannot suppress validation
// of another supplied member or a conflict between other complete pairs.
func resolveEndpointPair(ctx context.Context, rule aggregationRule, kinds map[string]string, contributors []contributor) (pair, missing [2]string, err error) {
	ep := rule.endpointGrouping
	rules := [2]string{ep.SourceVertexAggregationRuleID, ep.DestinationVertexAggregationRuleID}
	keys := [2][]string{ep.SourceGroupingKeys, ep.DestinationGroupingKeys}
	conflict := false
	for _, entry := range contributors {
		if err := itemCheckpointV2(ctx); err != nil {
			return pair, missing, err
		}
		var current [2]string
		for side, id := range rules {
			values := make([]any, 0, len(keys[side]))
			complete := true
			for _, key := range keys[side] {
				value, ok := contributorField(entry, key)
				if !ok {
					if missing[side] == "" {
						missing[side] = key
					}
					complete = false
					continue
				}
				value, err = groupingScalar(value)
				if err != nil {
					return pair, missing, err
				}
				values = append(values, value)
			}
			if complete {
				current[side], err = scalarTupleDigest(id, "vertex", kinds[id], values)
				if err != nil {
					return pair, missing, err
				}
			}
		}
		if current[0] == "" || current[1] == "" {
			continue
		}
		if pair[0] == "" {
			pair = current
		} else if pair != current {
			conflict = true
		}
	}
	if conflict {
		return pair, missing, &ProjectionErrorV2{Code: "projection_validation_failed", ReasonCode: "aggregation_endpoint_conflict", RetryAction: "do_not_retry", Details: map[string]any{}}
	}
	return pair, missing, nil
}

// groupingTupleDigest is shared by group creation and endpoint lookup. Missing
// fields are handled by the caller; supplied values must be canonical scalars.
func groupingTupleDigest(ruleID, scope, kind string, values []any) (string, error) {
	tuple := make([]any, len(values))
	for i, value := range values {
		scalar, err := groupingScalar(value)
		if err != nil {
			return "", err
		}
		tuple[i] = scalar
	}
	return scalarTupleDigest(ruleID, scope, kind, tuple)
}

func groupingScalar(value any) (any, error) {
	if number, ok := value.(json.Number); ok {
		var err error
		value, err = normalizedNumberV2(number)
		if err != nil {
			return nil, invalidProjectionErrorV2("invalid_number", "")
		}
	}
	switch scalar := value.(type) {
	case nil, bool, string, int, int64:
	case float64:
		const upper = 1 << 63
		if scalar >= -upper && scalar < upper && math.Trunc(scalar) == scalar {
			return int64(scalar), nil
		}
		if math.IsInf(scalar, 0) || math.IsNaN(scalar) {
			return nil, invalidProjectionErrorV2("invalid_number", "")
		}
	default:
		return nil, &ProjectionErrorV2{Code: "projection_validation_failed", ReasonCode: "invalid_aggregation_rule", RetryAction: "do_not_retry", Details: map[string]any{"reason_code": "grouping_value_not_scalar"}}
	}
	return value, nil
}

func scalarTupleDigest(ruleID, scope, kind string, tuple []any) (string, error) {
	keyJSON, err := canonicalJSON(tuple)
	if err != nil {
		return "", computationProjectionErrorV2("canonical_grouping_failed", err)
	}
	digest, err := digestTuple("GPGROUP1\n", ruleID, scope, kind, keyJSON)
	if err != nil {
		return "", computationProjectionErrorV2("canonical_grouping_failed", err)
	}
	return digest, nil
}

// contributorOrder is immediate provenance, not transitive ancestry or a hash.
type contributorOrder [4]string

func (order contributorOrder) less(other contributorOrder) bool {
	for i := range order {
		if order[i] != other[i] {
			return order[i] < other[i]
		}
	}
	return false
}

func (order contributorOrder) key() string {
	var parts [4]string
	for i, value := range order {
		parts[i] = hex.EncodeToString([]byte(value))
	}
	return "csk_" + strings.Join(parts[:], ".")
}

func (entry contributor) orderTuple() (contributorOrder, error) {
	if source := entry.Entity; source != nil {
		return contributorOrder{source.SourceEntityKind, source.SourceEntityID, "", ""}, nil
	}
	if source := entry.Relationship; source != nil {
		return contributorOrder{source.SourceRelationshipKind, source.SourceRelationshipID, "", ""}, nil
	}
	if vertex := entry.Vertex; vertex != nil {
		if ref := vertex.SourceEntityRef; ref != nil {
			return contributorOrder{ref.SourceEntityKind, ref.SourceEntityID, ref.MappingRuleID, vertex.VertexID}, nil
		}
		if rule := vertex.Metadata.AggregationRuleID; rule != nil {
			return contributorOrder{vertex.VertexKind, vertex.VertexID, *rule, vertex.VertexID}, nil
		}
	}
	if edge := entry.Edge; edge != nil {
		if ref := edge.SourceRelationshipRef; ref != nil {
			return contributorOrder{ref.SourceRelationshipKind, ref.SourceRelationshipID, ref.MappingRuleID, edge.EdgeID}, nil
		}
		if rule := edge.Metadata.AggregationRuleID; rule != nil {
			return contributorOrder{edge.EdgeKind, edge.EdgeID, *rule, edge.EdgeID}, nil
		}
	}
	return contributorOrder{}, &ProjectionErrorV2{Code: "projection_computation_failed", ReasonCode: "invalid_contributor_provenance", RetryAction: "do_not_retry", Details: map[string]any{}}
}
