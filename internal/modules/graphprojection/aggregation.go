package graphprojection

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
)

func emitAggregations(ctx context.Context, run projectionWork, directVertices []Vertex, directEdges []Edge) ([]Vertex, []Edge, []ValidationIssue, error) {
	vertices := []Vertex{}
	edges := []Edge{}
	availableVertices := append([]Vertex(nil), directVertices...)
	availableEdges := append([]Edge(nil), directEdges...)
	issues := []ValidationIssue{}
	aggregatedVerticesByRuleAndDigest := map[string]map[string]Vertex{}
	for _, rule := range run.Request.projectionConfig.AggregationRules {
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}
		if rule.TargetScope != "vertex" {
			continue
		}
		groups, err := groupContributors(ctx, run, rule, availableVertices, availableEdges, &issues)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, digest := range sortedGroupDigests(groups) {
			if err := ctx.Err(); err != nil {
				return nil, nil, nil, err
			}
			contributors := groups[digest]
			vertexID, _ := generatedID("vx_", "GPVERTEX1\n", "aggregated_vertex", run.Request.ProjectionSchemaID, run.GraphViewID, rule.AggregationIdentityDigest, digest)
			props := mergeAggregateProperties(run, rule, digest, contributors, "vertex", rule.ProjectedKind, vertexID, &issues)
			mappedMetadata := mergeAggregateMetadata(run, rule, digest, contributors, "vertex", rule.ProjectedKind, vertexID, &issues)
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
			if aggregatedVerticesByRuleAndDigest[rule.AggregationRuleID] == nil {
				aggregatedVerticesByRuleAndDigest[rule.AggregationRuleID] = map[string]Vertex{}
			}
			aggregatedVerticesByRuleAndDigest[rule.AggregationRuleID][digest] = vertex
			vertices = append(vertices, vertex)
			availableVertices = append(availableVertices, vertex)
		}
	}
	for _, rule := range run.Request.projectionConfig.AggregationRules {
		if err := ctx.Err(); err != nil {
			return nil, nil, nil, err
		}
		if rule.TargetScope != "edge" || rule.endpointGrouping == nil {
			continue
		}
		aggregationKinds := aggregationProjectedKinds(run.Request.projectionConfig.AggregationRules)
		groups, err := groupContributors(ctx, run, rule, availableVertices, availableEdges, &issues)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, digest := range sortedGroupDigests(groups) {
			if err := ctx.Err(); err != nil {
				return nil, nil, nil, err
			}
			contributors := groups[digest]
			srcRuleID := rule.endpointGrouping.SourceVertexAggregationRuleID
			dstRuleID := rule.endpointGrouping.DestinationVertexAggregationRuleID
			srcDigest, srcMissingField, srcOK := endpointDigest(srcRuleID, "vertex", aggregationKinds[srcRuleID], rule.endpointGrouping.SourceGroupingKeys, contributors)
			dstDigest, dstMissingField, dstOK := endpointDigest(dstRuleID, "vertex", aggregationKinds[dstRuleID], rule.endpointGrouping.DestinationGroupingKeys, contributors)
			src := aggregatedVerticesByRuleAndDigest[rule.endpointGrouping.SourceVertexAggregationRuleID][srcDigest]
			dst := aggregatedVerticesByRuleAndDigest[rule.endpointGrouping.DestinationVertexAggregationRuleID][dstDigest]
			if !srcOK || !dstOK || src.VertexID == "" || dst.VertexID == "" {
				if rule.endpointGrouping.MissingEndpointBehavior == "error" {
					issues = append(issues, endpointResolutionIssues(run, rule, "src", srcDigest, srcMissingField, srcOK, src.VertexID != "")...)
					issues = append(issues, endpointResolutionIssues(run, rule, "dst", dstDigest, dstMissingField, dstOK, dst.VertexID != "")...)
				}
				continue
			}
			edgeID, _ := generatedID("ed_", "GPEDGE1\n", "aggregated_edge", run.Request.ProjectionSchemaID, run.GraphViewID, rule.AggregationIdentityDigest, src.VertexID, dst.VertexID, rule.EdgeDirection, digest)
			props := mergeAggregateProperties(run, rule, digest, contributors, "edge", rule.ProjectedKind, edgeID, &issues)
			mappedMetadata := mergeAggregateMetadata(run, rule, digest, contributors, "edge", rule.ProjectedKind, edgeID, &issues)
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
			edges = append(edges, edge)
			availableEdges = append(availableEdges, edge)
		}
	}
	return vertices, edges, issues, nil
}

func sortedGroupDigests(groups map[string][]contributor) []string {
	digests := make([]string, 0, len(groups))
	for digest := range groups {
		digests = append(digests, digest)
	}
	sort.Strings(digests)
	return digests
}

func endpointResolutionIssues(run projectionWork, rule aggregationRule, side, digest, missingField string, digestOK, matched bool) []ValidationIssue {
	if !digestOK {
		return []ValidationIssue{run.issue("error", "aggregation_endpoint_missing", "mapping_rule", rule.AggregationRuleID, missingField, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "endpoint_side": side, "reason_code": "endpoint_key_missing", "endpoint_digest": nil, "field_path": missingField})}
	}
	if !matched {
		return []ValidationIssue{run.issue("error", "aggregation_endpoint_missing", "mapping_rule", rule.AggregationRuleID, nil, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "endpoint_side": side, "reason_code": "endpoint_vertex_not_found", "endpoint_digest": digest, "field_path": nil})}
	}
	return nil
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
	SortKey      string
	Entity       *sourceEntity
	Relationship *sourceRelationship
	Vertex       *Vertex
	Edge         *Edge
}

func groupContributors(ctx context.Context, run projectionWork, rule aggregationRule, vertices []Vertex, edges []Edge, issues *[]ValidationIssue) (map[string][]contributor, error) {
	groups := map[string][]contributor{}
	for _, contributor := range contributorsForRule(run, rule, vertices, edges) {
		if err := ctx.Err(); err != nil {
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
					*issues = append(*issues, run.issue("error", "aggregation_grouping_key_missing", "source_or_projected_item", contributor.ID, fieldPath, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "field_path": fieldPath, "contributor_id": contributor.ID}))
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
	return groups, nil
}

func contributorsForRule(run projectionWork, rule aggregationRule, vertices []Vertex, edges []Edge) []contributor {
	out := []contributor{}
	declaredEntityKinds := stringSet(run.Request.projectionConfig.DeclaredSourceEntityKinds)
	declaredRelationshipKinds := stringSet(run.Request.projectionConfig.DeclaredSourceRelationshipKinds)
	switch rule.InputScope {
	case "source_entity":
		for i := range run.Request.SourceEntities {
			entity := &run.Request.SourceEntities[i]
			if entity.SourceEntityKind != rule.InputKind || !validIdentifier(entity.SourceEntityID) || !validIdentifier(entity.SourceEntityKind) ||
				!sourceItemWithinLimits(len(entity.Labels), len(entity.Properties), len(entity.Metadata)) ||
				!sourceItemValuesWithinLimits(entity.Properties, entity.Metadata) ||
				!declaredEntityKinds[entity.SourceEntityKind] ||
				!filtersMatchEntity(run.Request.filters.EntityFilters, *entity) {
				continue
			}
			out = append(out, contributor{Kind: "source_entity", ID: entity.SourceEntityID, KindName: entity.SourceEntityKind, SortKey: contributorSortKey("source_entity", entity.SourceEntityKind, entity.SourceEntityID), Entity: entity})
		}
	case "source_relationship":
		for i := range run.Request.SourceRelationships {
			relationship := &run.Request.SourceRelationships[i]
			if relationship.SourceRelationshipKind != rule.InputKind || !validIdentifier(relationship.SourceRelationshipID) || !validIdentifier(relationship.SourceRelationshipKind) ||
				!sourceItemWithinLimits(len(relationship.Labels), len(relationship.Properties), len(relationship.Metadata)) ||
				!sourceItemValuesWithinLimits(relationship.Properties, relationship.Metadata) ||
				!declaredRelationshipKinds[relationship.SourceRelationshipKind] ||
				relationship.SrcSourceEntityID == "" || relationship.DstSourceEntityID == "" || !validSourceDirection(relationship.Direction) ||
				!filtersMatchRelationship(run.Request.filters.RelationshipFilters, *relationship) {
				continue
			}
			out = append(out, contributor{Kind: "source_relationship", ID: relationship.SourceRelationshipID, KindName: relationship.SourceRelationshipKind, SortKey: contributorSortKey("source_relationship", relationship.SourceRelationshipKind, relationship.SourceRelationshipID), Relationship: relationship})
		}
	case "projected_vertex":
		for i := range vertices {
			vertex := &vertices[i]
			if vertex.VertexKind != rule.InputKind {
				continue
			}
			out = append(out, contributor{Kind: "projected_vertex", ID: vertex.VertexID, KindName: vertex.VertexKind, SortKey: contributorSortKey("projected_vertex", vertex.VertexKind, vertex.SortKey, vertex.VertexID), Vertex: vertex})
		}
	case "projected_edge":
		for i := range edges {
			edge := &edges[i]
			if edge.EdgeKind != rule.InputKind {
				continue
			}
			out = append(out, contributor{Kind: "projected_edge", ID: edge.EdgeID, KindName: edge.EdgeKind, SortKey: contributorSortKey("projected_edge", edge.EdgeKind, edge.SortKey, edge.EdgeID), Edge: edge})
		}
	}
	return out
}

func mergeAggregateProperties(run projectionWork, rule aggregationRule, groupingDigest string, contributors []contributor, targetScope, targetKind, outputID string, issues *[]ValidationIssue) map[string]any {
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
			value, found := contributorField(contributor, definition.SourceFieldPath)
			normalized, include, issueCode := evaluateCandidate(candidateDefinitionFromProperty(definition), value, found)
			if issueCode != "" {
				if issueCode == "required_property_missing" {
					*issues = append(*issues, run.issue("error", "required_property_missing", "property", outputID, definition.SourceFieldPath, map[string]any{"projected_key": definition.ProjectedKey, "source_field_path": definition.SourceFieldPath, "output_object_id": outputID, "aggregation_rule_id": rule.AggregationRuleID}))
				} else {
					*issues = append(*issues, propertyValueIssue(run, issueCode, definition.ProjectedKey, definition.ProjectedType, definition.SourceFieldPath, outputID, value, rule.AggregationRuleID, groupingDigest, contributor.ID))
				}
				continue
			}
			if include {
				candidates = append(candidates, normalized)
			}
		}
		merged, ok, conflict := mergeValues(mergeBehavior, candidates)
		if conflict {
			*issues = append(*issues, run.issue("error", "aggregation_merge_conflict", "mapping_rule", rule.AggregationRuleID, nil, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "canonical_grouping_key_digest": groupingDigest, "projected_key": definition.ProjectedKey}))
			continue
		}
		if ok {
			properties[definition.ProjectedKey] = merged
		}
	}
	return properties
}

func mergeAggregateMetadata(run projectionWork, rule aggregationRule, groupingDigest string, contributors []contributor, targetScope, targetKind, outputID string, issues *[]ValidationIssue) map[string]any {
	metadata := map[string]any{}
	for _, mapping := range run.Request.projectionConfig.MetadataMappings {
		if mapping.TargetScope != targetScope || (mapping.TargetKind != "*" && mapping.TargetKind != targetKind) {
			continue
		}
		if mapping.MergeBehavior == "omit" {
			continue
		}
		if mapping.MergeBehavior == "count" {
			metadata[mapping.ProjectedMetadataKey] = len(contributors)
			continue
		}
		candidates := []any{}
		for _, contributor := range contributors {
			value, found := contributorField(contributor, mapping.SourceFieldPath)
			normalized, include, issueCode := evaluateCandidate(candidateDefinitionFromMetadata(mapping), value, found)
			if issueCode != "" {
				if issueCode == "required_property_missing" {
					*issues = append(*issues, run.issue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, mapping.SourceFieldPath, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "required_metadata_missing"}))
				} else {
					*issues = append(*issues, propertyValueIssue(run, issueCode, mapping.ProjectedMetadataKey, mapping.ProjectedType, mapping.SourceFieldPath, outputID, value, rule.AggregationRuleID, groupingDigest, contributor.ID))
				}
				continue
			}
			if include {
				candidates = append(candidates, normalized)
			}
		}
		merged, ok, conflict := mergeValues(mapping.MergeBehavior, candidates)
		if conflict {
			*issues = append(*issues, run.issue("error", "aggregation_merge_conflict", "mapping_rule", rule.AggregationRuleID, nil, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "canonical_grouping_key_digest": groupingDigest, "projected_key": mapping.ProjectedMetadataKey}))
			continue
		}
		if ok {
			metadata[mapping.ProjectedMetadataKey] = merged
		}
	}
	return metadata
}

func deriveProperties(run projectionWork, targetScope, targetKind string, entity sourceEntity, relationship *sourceRelationship, graphSource map[string]any, outputID string) (map[string]any, []ValidationIssue) {
	properties := map[string]any{}
	issues := []ValidationIssue{}
	for _, definition := range run.Request.PropertyDefinitions {
		if !propertyApplies(definition, targetScope, targetKind) {
			continue
		}
		value, found := sourceField(entity, relationship, graphSource, definition.SourceFieldPath)
		normalized, include, issueCode := evaluateCandidate(candidateDefinitionFromProperty(definition), value, found)
		if issueCode != "" {
			if issueCode == "required_property_missing" {
				issues = append(issues, run.issue("error", issueCode, "property", outputID+"#"+definition.ProjectedKey, definition.SourceFieldPath, map[string]any{"projected_key": definition.ProjectedKey, "source_field_path": definition.SourceFieldPath, "output_object_id": outputID}))
			} else {
				issues = append(issues, propertyValueIssue(run, issueCode, definition.ProjectedKey, definition.ProjectedType, definition.SourceFieldPath, outputID, value, "", "", ""))
			}
			continue
		}
		if include {
			properties[definition.ProjectedKey] = normalized
		}
	}
	return properties, issues
}

func deriveGraphProperties(run projectionWork, issues *[]ValidationIssue) map[string]any {
	properties, propertyIssues := deriveProperties(run, "graph_view", "*", sourceEntity{}, nil, run.Request.SourceMetadata, run.GraphViewID)
	*issues = append(*issues, propertyIssues...)
	return properties
}

func deriveMetadata(run projectionWork, targetScope, targetKind string, entity sourceEntity, relationship *sourceRelationship, outputID string) (map[string]any, []ValidationIssue) {
	metadata := map[string]any{}
	issues := []ValidationIssue{}
	for _, mapping := range run.Request.projectionConfig.MetadataMappings {
		if mapping.TargetScope != targetScope || (mapping.TargetKind != "*" && mapping.TargetKind != targetKind) {
			continue
		}
		value, found := sourceField(entity, relationship, run.Request.SourceMetadata, mapping.SourceFieldPath)
		normalized, include, issueCode := evaluateCandidate(candidateDefinitionFromMetadata(mapping), value, found)
		if issueCode != "" {
			if issueCode == "required_property_missing" {
				issues = append(issues, run.issue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, mapping.SourceFieldPath, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "required_metadata_missing"}))
			} else {
				issues = append(issues, propertyValueIssue(run, issueCode, mapping.ProjectedMetadataKey, mapping.ProjectedType, mapping.SourceFieldPath, outputID, value, "", "", ""))
			}
			continue
		}
		if include {
			metadata[mapping.ProjectedMetadataKey] = normalized
		}
	}
	return metadata, issues
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
	if !valueMatchesType(definition.ProjectedType, value) {
		return nil, false, "invalid_property_type"
	}
	return value, true, ""
}

func propertyValueIssue(run projectionWork, code, projectedKey, expectedType, sourceFieldPath, outputID string, value any, aggregationRuleID, groupingDigest, contributorID string) ValidationIssue {
	details := map[string]any{"projected_key": projectedKey, "expected_type": expectedType, "actual_type": jsonValueType(value), "source_field_path": sourceFieldPath, "output_object_id": outputID}
	if aggregationRuleID != "" {
		details["aggregation_rule_id"] = aggregationRuleID
		details["canonical_grouping_key_digest"] = groupingDigest
	}
	if contributorID != "" {
		details["contributor_id"] = contributorID
	}
	return run.issue("error", code, "property", outputID+"#"+projectedKey, sourceFieldPath, details)
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

func sourceField(entity sourceEntity, relationship *sourceRelationship, graphSource map[string]any, fieldPath string) (any, bool) {
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

func endpointDigest(ruleID, targetScope, projectedKind string, keys []string, contributors []contributor) (string, string, bool) {
	if len(contributors) == 0 {
		return "", "", false
	}
	values := make([]any, 0, len(keys))
	for _, key := range keys {
		value, ok := contributorField(contributors[0], key)
		if !ok {
			return "", key, false
		}
		values = append(values, value)
	}
	keyJSON, _ := canonicalJSON(values)
	digest, _ := digestTuple("GPGROUP1\n", ruleID, targetScope, projectedKind, keyJSON)
	return digest, "", true
}
