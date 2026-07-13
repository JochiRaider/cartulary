package graphprojection

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type projectOptions struct {
	ProjectionRunNonce      string
	AcceptedAt              time.Time
	GeneratedAt             time.Time
	PreviousProjectionRunID *string
	InvocationIDPrefix      string
	InvocationDomain        string
	Operation               string
}

func project(data []byte, options projectOptions) (ProjectionRun, error) {
	run, err := admitProjectionInput(data, admitOptions{
		ProjectionRunNonce: options.ProjectionRunNonce,
		AcceptedAt:         options.AcceptedAt,
		InvocationIDPrefix: options.InvocationIDPrefix,
		InvocationDomain:   options.InvocationDomain,
		Operation:          options.Operation,
	})
	if err != nil {
		return ProjectionRun{}, err
	}
	run.State = RunStateComputing
	startedAt := options.AcceptedAt.UTC()
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	run.StartedAt = &startedAt
	projected := projectAdmittedRun(run, options)
	return projected, nil
}

func projectAdmittedRun(run ProjectionRun, options projectOptions) ProjectionRun {
	issues := validateAdmittedRequest(run)
	result := run
	generatedAt := options.GeneratedAt.UTC()
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}
	result.PreviousProjectionRunID = options.PreviousProjectionRunID
	if hasFatalIssue(issues) || len(issues) > graphProjectionLimits.MaxValidationIssues {
		result.State = RunStateFailed
		result.FailureReason = "fatal_validation"
		now := generatedAt
		result.CompletedAt = &now
		result.ValidationSummary = validationSummary(run, issues)
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
	issues = append(issues, projectedOutputLimitIssues(run, vertices, edges)...)

	if hasFatalIssue(issues) || len(issues) > graphProjectionLimits.MaxValidationIssues {
		result.State = RunStateFailed
		result.FailureReason = "fatal_validation"
		now := generatedAt
		result.CompletedAt = &now
		result.ValidationSummary = validationSummary(run, issues)
		return result
	}

	graphProperties := deriveGraphProperties(run, &issues)
	graphMappedMetadata, graphMetadataIssues := deriveMetadata(run, "graph_view", "*", SourceEntity{}, nil, run.GraphViewID)
	issues = append(issues, graphMetadataIssues...)
	if len(graphProperties) > graphProjectionLimits.MaxPropertiesPerObject {
		issues = append(issues, run.issue("fatal", "projected_output_limit_exceeded", "graph_view", run.GraphViewID, nil, map[string]any{"limit_key": "max_properties_per_object", "limit": graphProjectionLimits.MaxPropertiesPerObject, "observed": len(graphProperties)}))
	}
	if len(graphMappedMetadata) > graphProjectionLimits.MaxMetadataKeysPerObject {
		issues = append(issues, run.issue("fatal", "projected_output_limit_exceeded", "graph_view", run.GraphViewID, nil, map[string]any{"limit_key": "max_metadata_keys_per_object", "limit": graphProjectionLimits.MaxMetadataKeysPerObject, "observed": len(graphMappedMetadata)}))
	}
	if hasFatalIssue(issues) || len(issues) > graphProjectionLimits.MaxValidationIssues {
		result.State = RunStateFailed
		result.FailureReason = "fatal_validation"
		now := generatedAt
		result.CompletedAt = &now
		result.ValidationSummary = validationSummary(run, issues)
		return result
	}
	summary := validationSummary(run, issues)
	result.State = RunStateAvailable
	now := generatedAt
	result.GeneratedAt = &now
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
		Properties:         graphProperties,
		Metadata: GraphMetadata{
			ProjectionConfigDigest:  run.ProjectionConfigDigest,
			ProjectionSourceDigest:  run.ProjectionSourceDigest,
			PreviousProjectionRunID: options.PreviousProjectionRunID,
			MappedMetadata:          graphMappedMetadata,
		},
		SchemaRegistry:       buildSchemaRegistry(run, vertices, edges),
		Vertices:             vertices,
		Edges:                edges,
		ValidationSummary:    summary,
		ConsumerCapabilities: defaultConsumerCapabilities(),
	}
	result.ValidationSummary = summary
	result.GraphView = graphView
	if graphBytes, err := canonicalJSON(graphViewCanonicalResource(*graphView)); err == nil {
		result.ProjectionOutputDigest = sha256Hex(graphBytes)
	}
	return result
}

func validateAdmittedRequest(run ProjectionRun) []ValidationIssue {
	request := run.Request
	issues := []ValidationIssue{}
	scalar := func(valid bool, field string) {
		if !valid {
			issues = append(issues, run.issue("fatal", "invalid_input_shape", "projection_input", "projection_input", field, map[string]any{"field": field, "reason_code": "scalar_contract_violation"}))
		}
	}
	scalar(validIdentifier(request.ProjectionConfig.GraphViewKey), "$.projection_config.graph_view_key")
	scalar(validIdentifier(request.SourceSnapshotID), "$.source_snapshot_id")
	scalar(validIdentifier(request.RequestedBy), "$.requested_by")
	_, timestampErr := parseTimestamp(request.RequestedAt)
	scalar(timestampErr == nil, "$.requested_at")
	for index, entity := range request.SourceEntities {
		base := fmt.Sprintf("$.source_entities[%d]", index)
		scalar(validIdentifier(entity.SourceEntityID), base+".source_entity_id")
		scalar(validIdentifier(entity.SourceEntityKind), base+".source_entity_kind")
	}
	for index, relationship := range request.SourceRelationships {
		base := fmt.Sprintf("$.source_relationships[%d]", index)
		scalar(validIdentifier(relationship.SourceRelationshipID), base+".source_relationship_id")
		scalar(validIdentifier(relationship.SourceRelationshipKind), base+".source_relationship_kind")
		if relationship.SrcSourceEntityID != "" {
			scalar(validIdentifier(relationship.SrcSourceEntityID), base+".src_source_entity_id")
		}
		if relationship.DstSourceEntityID != "" {
			scalar(validIdentifier(relationship.DstSourceEntityID), base+".dst_source_entity_id")
		}
	}
	if len(request.ProjectionConfig.DeclaredSourceEntityKinds) == 0 && len(request.ProjectionConfig.DeclaredSourceRelationshipKinds) == 0 && !request.ProjectionConfig.AllowEmptyKindRegistry {
		issues = append(issues, run.issue("fatal", "invalid_projection_config", "graph_view", run.GraphViewID, "$.projection_config", map[string]any{"reason_code": "empty_kind_registry_not_allowed"}))
	}
	checkDuplicates := func(values []string, collection string) {
		seen := map[string]struct{}{}
		for _, value := range values {
			if _, ok := seen[value]; ok {
				issues = append(issues, run.issue("fatal", "duplicate_identifier", "projection_input", "projection_input", nil, map[string]any{"identifier_value": value, "collection": collection}))
			}
			seen[value] = struct{}{}
		}
	}
	checkDuplicates(entityIDs(request.SourceEntities), "$.source_entities")
	checkDuplicates(relationshipIDs(request.SourceRelationships), "$.source_relationships")
	checkDuplicateKinds := func(values []string) {
		for index := 1; index < len(values); index++ {
			if values[index] == values[index-1] {
				issues = append(issues, run.issue("fatal", "invalid_projection_config", "projection_config", "projection_config", nil, map[string]any{"field": "$.projection_config", "reason_code": "declared_kind_duplicate"}))
			}
		}
	}
	checkDuplicateKinds(request.ProjectionConfig.DeclaredSourceEntityKinds)
	checkDuplicateKinds(request.ProjectionConfig.DeclaredSourceRelationshipKinds)
	issues = append(issues, admittedResourceLimitIssues(run)...)
	issues = append(issues, validateLabels(run)...)
	issues = append(issues, validateFilters(run)...)
	issues = append(issues, validateMappingsAndDefinitions(run)...)
	issues = append(issues, validateRetentionPolicy(run)...)
	for _, mapping := range request.RelationshipMappings {
		if mapping.EmitReverseEdge && mapping.DirectionPolicy != "normalize_forward" && mapping.DirectionPolicy != "normalize_reverse" {
			issues = append(issues, run.issue("fatal", "invalid_reverse_edge_policy", "mapping_rule", mapping.MappingRuleID, nil, map[string]any{"mapping_rule_id": mapping.MappingRuleID, "projected_direction": nil}))
		}
	}
	return issues
}

func validateRetentionPolicy(run ProjectionRun) []ValidationIssue {
	policy := run.Request.ProjectionConfig.RetentionPolicy
	issues := []ValidationIssue{}
	fields := []struct {
		name string
		max  int64
	}{
		{name: "retention_count", max: 100},
		{name: "retention_duration_seconds", max: 31536000},
		{name: "failed_retention_count", max: 100},
		{name: "failed_retention_duration_seconds", max: 31536000},
	}
	for _, field := range fields {
		lexeme, supplied := policy.RawIntegerLexemes[field.name]
		if !supplied {
			continue
		}
		path := "$.projection_config.retention_policy." + field.name
		if !validFiniteInteger(lexeme) {
			issues = append(issues, run.issue("fatal", "invalid_retention_policy", "projection_config", "projection_config", path, map[string]any{"field": path, "reason_code": "invalid_type"}))
			continue
		}
		value, _ := strconv.ParseInt(lexeme, 10, 64)
		if value < 0 || value > field.max {
			issues = append(issues, run.issue("fatal", "invalid_retention_policy", "projection_config", "projection_config", path, map[string]any{"field": path, "reason_code": "out_of_bounds"}))
		}
	}
	return issues
}

func validateFilters(run ProjectionRun) []ValidationIssue {
	issues := []ValidationIssue{}
	if run.Request.Filters.Logic != "and" {
		issues = append(issues, run.issue("fatal", "invalid_filter", "filter", "$.filters.logic", "$.filters.logic", map[string]any{"field": "$.filters.logic", "reason_code": "unsupported_logic"}))
	}
	validate := func(predicates []FilterPredicate, collection, scope string) {
		for index, predicate := range predicates {
			base := fmt.Sprintf("%s[%d]", collection, index)
			if !validFilterFieldPath(predicate.FieldPath, scope) {
				issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".field_path", map[string]any{"field": base + ".field_path", "reason_code": "invalid_field_scope"}))
			}
			switch predicate.Operator {
			case "exists":
				if predicate.HasValue {
					issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "value_forbidden"}))
				}
			case "equals", "not_equals":
				if !predicate.HasValue {
					issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "value_required"}))
				}
			case "in":
				values, ok := predicate.Value.([]any)
				if !predicate.HasValue {
					issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "value_required"}))
				} else if !ok || len(values) == 0 || !scalarArray(values) {
					issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".value", map[string]any{"field": base + ".value", "reason_code": "invalid_value_shape"}))
				}
			default:
				issues = append(issues, run.issue("fatal", "invalid_filter", "filter", base, base+".op", map[string]any{"field": base + ".op", "reason_code": "invalid_operator"}))
			}
		}
	}
	validate(run.Request.Filters.EntityFilters, "$.filters.entity_filters", "source_entity")
	validate(run.Request.Filters.RelationshipFilters, "$.filters.relationship_filters", "source_relationship")
	return issues
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

func validFilterFieldPath(path, scope string) bool {
	if strings.HasPrefix(path, "projected.") || strings.HasPrefix(path, "source_metadata.") {
		return false
	}
	parts := strings.Split(path, ".")
	if scope == "source_entity" {
		if len(parts) == 1 && (parts[0] == "source_entity_id" || parts[0] == "kind") {
			return true
		}
	} else if len(parts) == 1 && (parts[0] == "source_relationship_id" || parts[0] == "src_source_entity_id" || parts[0] == "dst_source_entity_id" || parts[0] == "kind" || parts[0] == "direction") {
		return true
	}
	return len(parts) == 2 && (parts[0] == "properties" || parts[0] == "metadata") && validPropertyKey(parts[1])
}

func validateMappingsAndDefinitions(run ProjectionRun) []ValidationIssue {
	request := run.Request
	issues := []ValidationIssue{}
	if request.RelationshipMappingSourceConflict {
		issues = append(issues, run.issue("fatal", "invalid_projection_config", "projection_config", "projection_config", "$.projection_config.relationship_mappings", map[string]any{"field": "$.projection_config.relationship_mappings", "reason_code": "relationship_mapping_source_conflict"}))
	}
	entityRuleIDs := map[string]bool{}
	entityKinds := map[string]string{}
	vertexKinds := map[string]bool{}
	for _, mapping := range request.ProjectionConfig.EntityMappings {
		if entityRuleIDs[mapping.MappingRuleID] {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_mapping_rule_id"))
		}
		entityRuleIDs[mapping.MappingRuleID] = true
		if prior := entityKinds[mapping.SourceEntityKind]; prior != "" {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_source_entity_kind_mapping"))
		}
		entityKinds[mapping.SourceEntityKind] = mapping.MappingRuleID
		vertexKinds[mapping.ProjectedVertexKind] = true
		if mapping.LabelPolicy != "mapping_only" && mapping.LabelPolicy != "preserve_source" && mapping.LabelPolicy != "mapping_then_source" {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "label_invalid"))
		}
		if mapping.InclusionFilter == nil && mapping.InclusionPredicate != "always" {
			issues = append(issues, run.issue("fatal", "invalid_filter", "filter", mapping.MappingRuleID, nil, map[string]any{"field": "$.projection_config.entity_mappings.inclusion_predicate", "reason_code": "invalid_operator"}))
		}
	}
	relationshipRuleIDs := map[string]bool{}
	relationshipKinds := map[string]string{}
	edgeKinds := map[string]bool{}
	for _, mapping := range request.RelationshipMappings {
		if entityRuleIDs[mapping.MappingRuleID] || relationshipRuleIDs[mapping.MappingRuleID] {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_mapping_rule_id"))
		}
		relationshipRuleIDs[mapping.MappingRuleID] = true
		if prior := relationshipKinds[mapping.SourceRelationshipKind]; prior != "" {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "duplicate_source_relationship_kind_mapping"))
		}
		relationshipKinds[mapping.SourceRelationshipKind] = mapping.MappingRuleID
		edgeKinds[mapping.ProjectedEdgeKind] = true
		if mapping.EmitReverseEdge {
			edgeKinds[mapping.ReverseEdgeKind] = true
		}
		if mapping.DirectionPolicy != "preserve" && mapping.DirectionPolicy != "normalize_forward" && mapping.DirectionPolicy != "normalize_reverse" && mapping.DirectionPolicy != "undirected" && mapping.DirectionPolicy != "bidirectional" {
			issues = append(issues, run.issue("fatal", "invalid_direction_policy", "mapping_rule", mapping.MappingRuleID, nil, map[string]any{"mapping_rule_id": mapping.MappingRuleID, "supplied_value": mapping.DirectionPolicy}))
		}
		if mapping.ReverseEdgeKindSupplied && !mapping.EmitReverseEdge {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "reverse_edge_kind_without_reverse"))
		}
		if mapping.LabelPolicy != "mapping_only" && mapping.LabelPolicy != "preserve_source" && mapping.LabelPolicy != "mapping_then_source" {
			issues = append(issues, invalidMappingIssue(run, mapping.MappingRuleID, "label_invalid"))
		}
		if mapping.InclusionFilter == nil && mapping.InclusionPredicate != "always" {
			issues = append(issues, run.issue("fatal", "invalid_filter", "filter", mapping.MappingRuleID, nil, map[string]any{"field": "$.projection_config.relationship_mappings.inclusion_predicate", "reason_code": "invalid_operator"}))
		}
	}
	for _, rule := range request.ProjectionConfig.AggregationRules {
		if rule.TargetScope == "vertex" {
			vertexKinds[rule.ProjectedKind] = true
		} else if rule.TargetScope == "edge" {
			edgeKinds[rule.ProjectedKind] = true
		}
	}
	issues = append(issues, validateAggregationRules(run)...)
	for _, kind := range request.ProjectionConfig.DeclaredSourceEntityKinds {
		if entityKinds[kind] == "" {
			issues = append(issues, run.issue("fatal", "missing_entity_mapping_rule", "mapping_rule", kind, nil, map[string]any{"source_entity_kind": kind}))
		}
	}
	for _, kind := range request.ProjectionConfig.DeclaredSourceRelationshipKinds {
		if relationshipKinds[kind] == "" {
			issues = append(issues, run.issue("error", "missing_relationship_mapping_rule", "mapping_rule", kind, nil, map[string]any{"source_relationship_kind": kind}))
		}
	}
	propertyExpansions := map[string]string{}
	for _, definition := range request.PropertyDefinitions {
		if !validPropertyKey(definition.ProjectedKey) || !validIdentifier(definition.PropertyDefinitionID) {
			issues = append(issues, run.issue("fatal", "invalid_property_definition", "property_definition", definition.PropertyDefinitionID, nil, map[string]any{"property_definition_id": definition.PropertyDefinitionID, "reason_code": "invalid_projected_type"}))
		}
		if !validProjectedType(definition.ProjectedType) {
			issues = append(issues, invalidPropertyDefinitionIssue(run, definition, "invalid_projected_type"))
		}
		if definition.HasDefaultValue && (definition.DefaultValue == nil && definition.NullOutputPolicy != "emit_null" || definition.DefaultValue != nil && !valueMatchesType(definition.ProjectedType, definition.DefaultValue)) {
			issues = append(issues, invalidPropertyDefinitionIssue(run, definition, "invalid_default_value"))
		}
		if (definition.MissingBehavior == "default" || definition.SourceNullBehavior == "default") && !definition.HasDefaultValue {
			issues = append(issues, invalidPropertyDefinitionIssue(run, definition, "invalid_default_value"))
		}
		if definition.SourceNullBehavior == "emit_null" && definition.NullOutputPolicy != "emit_null" {
			issues = append(issues, invalidPropertyDefinitionIssue(run, definition, "invalid_null_policy"))
		}
		if !validMergeBehavior(definition.MergeBehavior) || definition.MergeBehavior == "count" && definition.ProjectedType != "integer" {
			issues = append(issues, invalidPropertyDefinitionIssue(run, definition, "invalid_merge_behavior_type"))
		}
		if !validDefinitionFieldPath(definition.SourceFieldPath, definition.TargetScope) {
			issues = append(issues, run.issue("fatal", "invalid_field_path", "projection_config", "projection_config", definition.SourceFieldPath, map[string]any{"field_path": definition.SourceFieldPath, "scope": definition.TargetScope}))
		}
		for _, kind := range expansionKinds(definition.TargetScope, definition.TargetKind, vertexKinds, edgeKinds) {
			key := definition.TargetScope + "\n" + kind + "\n" + definition.ProjectedKey
			if propertyExpansions[key] != "" {
				issues = append(issues, run.issue("fatal", "invalid_property_definition", "property_definition", definition.PropertyDefinitionID, nil, map[string]any{"property_definition_id": definition.PropertyDefinitionID, "reason_code": "duplicate_after_wildcard_expansion"}))
			}
			propertyExpansions[key] = definition.PropertyDefinitionID
		}
	}
	metadataExpansions := map[string]string{}
	for _, mapping := range request.ProjectionConfig.MetadataMappings {
		if reservedSystemMetadataTerminal(mapping.ProjectedMetadataKey) {
			issues = append(issues, run.issue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, nil, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "reserved_metadata_key"}))
		}
		if !validDefinitionFieldPath(mapping.SourceFieldPath, mapping.TargetScope) {
			issues = append(issues, run.issue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, mapping.SourceFieldPath, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "invalid_source_scope"}))
		}
		if !validProjectedType(mapping.ProjectedType) {
			issues = append(issues, invalidMetadataMappingIssue(run, mapping, "invalid_projected_type"))
		}
		if mapping.HasDefaultValue && (mapping.DefaultValue == nil && mapping.NullOutputPolicy != "emit_null" || mapping.DefaultValue != nil && !valueMatchesType(mapping.ProjectedType, mapping.DefaultValue)) {
			issues = append(issues, invalidMetadataMappingIssue(run, mapping, "invalid_default_value"))
		}
		if (mapping.MissingBehavior == "default" || mapping.SourceNullBehavior == "default") && !mapping.HasDefaultValue {
			issues = append(issues, invalidMetadataMappingIssue(run, mapping, "invalid_default_value"))
		}
		if !validMergeBehavior(mapping.MergeBehavior) || mapping.MergeBehavior == "count" && mapping.ProjectedType != "integer" {
			issues = append(issues, invalidMetadataMappingIssue(run, mapping, "invalid_merge_behavior_type"))
		}
		for _, kind := range expansionKinds(mapping.TargetScope, mapping.TargetKind, vertexKinds, edgeKinds) {
			key := mapping.TargetScope + "\n" + kind + "\n" + mapping.ProjectedMetadataKey
			if metadataExpansions[key] != "" {
				issues = append(issues, run.issue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, nil, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": "duplicate_after_wildcard_expansion"}))
			}
			metadataExpansions[key] = mapping.MetadataMappingID
		}
	}
	return issues
}

func validateAggregationRules(run ProjectionRun) []ValidationIssue {
	rules := run.Request.ProjectionConfig.AggregationRules
	issues := []ValidationIssue{}
	indexes := map[string]int{}
	byID := map[string]AggregationRule{}
	for index, rule := range rules {
		if _, exists := indexes[rule.AggregationRuleID]; exists {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "aggregation_cycle"))
		}
		indexes[rule.AggregationRuleID] = index
		byID[rule.AggregationRuleID] = rule
	}
	for index, rule := range rules {
		validInputScope := rule.InputScope == "source_entity" || rule.InputScope == "source_relationship" || rule.InputScope == "projected_vertex" || rule.InputScope == "projected_edge"
		if !validInputScope || rule.TargetScope != "vertex" && rule.TargetScope != "edge" {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "input_scope_invalid"))
		}
		if len(rule.GroupingKeys) == 0 || len(rule.GroupingKeys) > 32 || hasDuplicateStringsLocal(rule.GroupingKeys) {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "grouping_key_invalid"))
		}
		for _, fieldPath := range rule.GroupingKeys {
			if !validAggregationFieldPath(fieldPath, rule.InputScope) {
				issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "grouping_key_invalid"))
			}
		}
		if rule.MissingGroupingKeyBehavior != "error" && rule.MissingGroupingKeyBehavior != "exclude" && rule.MissingGroupingKeyBehavior != "use_null" {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "grouping_key_invalid"))
		}
		if rule.TargetScope == "vertex" {
			if rule.EndpointGrouping != nil || rule.EdgeDirection != "directed" {
				issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_edge_direction"))
			}
			continue
		}
		if rule.EdgeDirection != "directed" && rule.EdgeDirection != "undirected" && rule.EdgeDirection != "bidirectional" {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_edge_direction"))
		}
		if rule.EndpointGrouping == nil {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_rule_not_vertex_rule"))
			continue
		}
		endpoint := rule.EndpointGrouping
		if endpoint.MissingEndpointBehavior != "error" && endpoint.MissingEndpointBehavior != "exclude" {
			issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "invalid_endpoint_behavior"))
		}
		for _, reference := range []struct {
			id   string
			keys []string
		}{{id: endpoint.SourceVertexAggregationRuleID, keys: endpoint.SourceGroupingKeys}, {id: endpoint.DestinationVertexAggregationRuleID, keys: endpoint.DestinationGroupingKeys}} {
			referenced, exists := byID[reference.id]
			if !exists || referenced.TargetScope != "vertex" {
				issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_rule_not_vertex_rule"))
				continue
			}
			if indexes[reference.id] >= index {
				issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "dependency_on_later_rule"))
			}
			if len(reference.keys) != len(referenced.GroupingKeys) {
				issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_grouping_key_count_mismatch"))
			}
			for _, fieldPath := range reference.keys {
				if !validAggregationFieldPath(fieldPath, rule.InputScope) {
					issues = append(issues, invalidAggregationIssue(run, rule.AggregationRuleID, "endpoint_field_scope_invalid"))
				}
			}
		}
	}
	return issues
}

func invalidAggregationIssue(run ProjectionRun, aggregationRuleID, reason string) ValidationIssue {
	return run.issue("fatal", "invalid_aggregation_rule", "mapping_rule", aggregationRuleID, nil, map[string]any{"aggregation_rule_id": aggregationRuleID, "reason_code": reason})
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

func validAggregationFieldPath(path, inputScope string) bool {
	if !validFieldPath(path) {
		return false
	}
	switch inputScope {
	case "source_entity":
		return path == "source_entity_id" || path == "kind" || strings.HasPrefix(path, "properties.") || strings.HasPrefix(path, "metadata.")
	case "source_relationship":
		return path == "source_relationship_id" || path == "src_source_entity_id" || path == "dst_source_entity_id" || path == "kind" || path == "direction" || strings.HasPrefix(path, "properties.") || strings.HasPrefix(path, "metadata.")
	case "projected_vertex":
		return path == "kind" || path == "projected.vertex_id" || strings.HasPrefix(path, "projected.properties.") || strings.HasPrefix(path, "projected.metadata.")
	case "projected_edge":
		return path == "kind" || path == "projected.edge_id" || path == "projected.src_vertex_id" || path == "projected.dst_vertex_id" || path == "projected.direction" || strings.HasPrefix(path, "projected.properties.") || strings.HasPrefix(path, "projected.metadata.")
	default:
		return false
	}
}

func invalidPropertyDefinitionIssue(run ProjectionRun, definition PropertyDefinition, reason string) ValidationIssue {
	return run.issue("fatal", "invalid_property_definition", "property_definition", definition.PropertyDefinitionID, nil, map[string]any{"property_definition_id": definition.PropertyDefinitionID, "reason_code": reason})
}

func invalidMetadataMappingIssue(run ProjectionRun, mapping MetadataMapping, reason string) ValidationIssue {
	return run.issue("fatal", "invalid_metadata_mapping", "mapping_rule", mapping.MetadataMappingID, nil, map[string]any{"metadata_mapping_id": mapping.MetadataMappingID, "reason_code": reason})
}

func validProjectedType(value string) bool {
	switch value {
	case "string", "integer", "boolean", "timestamp", "identifier", "string_array", "identifier_array":
		return true
	default:
		return false
	}
}

func validMergeBehavior(value string) bool {
	switch value {
	case "single_value", "first_by_sort", "last_by_sort", "distinct_sorted_array", "count", "omit":
		return true
	default:
		return false
	}
}

func invalidMappingIssue(run ProjectionRun, mappingRuleID, reason string) ValidationIssue {
	return run.issue("fatal", "invalid_mapping_rule", "mapping_rule", mappingRuleID, nil, map[string]any{"mapping_rule_id": mappingRuleID, "reason_code": reason})
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

func validDefinitionFieldPath(path, scope string) bool {
	if !validFieldPath(path) {
		return false
	}
	if scope == "graph_view" {
		return strings.HasPrefix(path, "source_metadata.")
	}
	if scope == "vertex" {
		return path == "source_entity_id" || path == "kind" || path == "projected.vertex_id" || strings.HasPrefix(path, "properties.") || strings.HasPrefix(path, "metadata.") || strings.HasPrefix(path, "projected.properties.") || strings.HasPrefix(path, "projected.metadata.")
	}
	if scope == "edge" {
		return path == "source_relationship_id" || path == "src_source_entity_id" || path == "dst_source_entity_id" || path == "kind" || path == "direction" || path == "projected.edge_id" || path == "projected.src_vertex_id" || path == "projected.dst_vertex_id" || path == "projected.direction" || strings.HasPrefix(path, "properties.") || strings.HasPrefix(path, "metadata.") || strings.HasPrefix(path, "projected.properties.") || strings.HasPrefix(path, "projected.metadata.")
	}
	return false
}

func admittedResourceLimitIssues(run ProjectionRun) []ValidationIssue {
	request := run.Request
	issues := []ValidationIssue{}
	whole := func(key string, observed, limit int) {
		if observed > limit {
			issues = append(issues, run.issue("fatal", "resource_limit_exceeded", "projection_input", "projection_input", nil, map[string]any{"limit_key": key, "limit": limit, "observed": observed}))
		}
	}
	whole("max_source_entities", len(request.SourceEntities), graphProjectionLimits.MaxSourceEntities)
	whole("max_source_relationships", len(request.SourceRelationships), graphProjectionLimits.MaxSourceRelationships)
	whole("max_entity_filters", len(request.Filters.EntityFilters), graphProjectionLimits.MaxEntityFilters)
	whole("max_relationship_filters", len(request.Filters.RelationshipFilters), graphProjectionLimits.MaxRelationshipFilters)
	whole("max_declared_source_entity_kinds", len(request.ProjectionConfig.DeclaredSourceEntityKinds), graphProjectionLimits.MaxDeclaredSourceEntityKinds)
	whole("max_declared_source_relationship_kinds", len(request.ProjectionConfig.DeclaredSourceRelationshipKinds), graphProjectionLimits.MaxDeclaredSourceRelationshipKinds)
	whole("max_entity_mappings", len(request.ProjectionConfig.EntityMappings), graphProjectionLimits.MaxEntityMappings)
	whole("max_relationship_mappings", len(request.RelationshipMappings), graphProjectionLimits.MaxRelationshipMappings)
	whole("max_property_definitions", len(request.PropertyDefinitions), graphProjectionLimits.MaxPropertyDefinitions)
	whole("max_metadata_mappings", len(request.ProjectionConfig.MetadataMappings), graphProjectionLimits.MaxMetadataMappings)
	whole("max_aggregation_rules", len(request.ProjectionConfig.AggregationRules), graphProjectionLimits.MaxAggregationRules)
	whole("max_default_vertex_labels", len(request.ProjectionConfig.DefaultVertexLabels), graphProjectionLimits.MaxDefaultVertexLabels)
	whole("max_default_edge_labels", len(request.ProjectionConfig.DefaultEdgeLabels), graphProjectionLimits.MaxDefaultEdgeLabels)
	whole("max_metadata_keys_per_object", len(request.SourceMetadata), graphProjectionLimits.MaxMetadataKeysPerObject)
	whole("max_custom_config_keys", len(request.ProjectionConfig.CustomConfig), graphProjectionLimits.MaxCustomConfigKeys)
	for _, mapping := range request.ProjectionConfig.EntityMappings {
		whole("max_mapping_labels_per_rule", len(mapping.MappingLabels), graphProjectionLimits.MaxMappingLabelsPerRule)
		whole("max_mapping_property_key_refs", len(mapping.RequiredPropertyKeys), graphProjectionLimits.MaxMappingPropertyKeyRefs)
		whole("max_mapping_property_key_refs", len(mapping.OptionalPropertyKeys), graphProjectionLimits.MaxMappingPropertyKeyRefs)
	}
	for _, mapping := range request.RelationshipMappings {
		whole("max_mapping_labels_per_rule", len(mapping.MappingLabels), graphProjectionLimits.MaxMappingLabelsPerRule)
		whole("max_mapping_property_key_refs", len(mapping.RequiredPropertyKeys), graphProjectionLimits.MaxMappingPropertyKeyRefs)
		whole("max_mapping_property_key_refs", len(mapping.OptionalPropertyKeys), graphProjectionLimits.MaxMappingPropertyKeyRefs)
	}
	for _, entity := range request.SourceEntities {
		issues = append(issues, sourceItemResourceLimitIssues(run, entity.SourceEntityID, len(entity.Labels), len(entity.Properties), len(entity.Metadata))...)
		issues = append(issues, sourceItemStringLimitIssues(run, entity.SourceEntityID, entity.Properties, entity.Metadata)...)
	}
	for _, relationship := range request.SourceRelationships {
		issues = append(issues, sourceItemResourceLimitIssues(run, relationship.SourceRelationshipID, len(relationship.Labels), len(relationship.Properties), len(relationship.Metadata))...)
		issues = append(issues, sourceItemStringLimitIssues(run, relationship.SourceRelationshipID, relationship.Properties, relationship.Metadata)...)
	}
	return issues
}

func sourceItemStringLimitIssues(run ProjectionRun, itemID string, objects ...map[string]any) []ValidationIssue {
	observed := 0
	for _, object := range objects {
		for _, value := range object {
			if length := longestPropertyString(value); length > observed {
				observed = length
			}
		}
	}
	if observed <= graphProjectionLimits.MaxStringPropertyValueLength {
		return nil
	}
	return []ValidationIssue{run.issue("error", "source_item_resource_limit_exceeded", "source_item", itemID, nil, map[string]any{"source_item_id": itemID, "limit_key": "max_string_property_value_length", "limit": graphProjectionLimits.MaxStringPropertyValueLength, "observed": observed})}
}

func longestPropertyString(value any) int {
	switch typed := value.(type) {
	case string:
		return utf8.RuneCountInString(typed)
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

func validateLabels(run ProjectionRun) []ValidationIssue {
	issues := []ValidationIssue{}
	check := func(values []string, field string) {
		for _, value := range values {
			if value == "" || utf8.RuneCountInString(value) > graphProjectionLimits.MaxLabelLength {
				issues = append(issues, run.issue("fatal", "invalid_input_shape", "projection_input", "projection_input", field, map[string]any{"field": field, "reason_code": "invalid_label"}))
			}
		}
	}
	check(run.Request.ProjectionConfig.DefaultVertexLabels, "$.projection_config.default_vertex_labels")
	check(run.Request.ProjectionConfig.DefaultEdgeLabels, "$.projection_config.default_edge_labels")
	for index, mapping := range run.Request.ProjectionConfig.EntityMappings {
		check(mapping.MappingLabels, fmt.Sprintf("$.projection_config.entity_mappings[%d].mapping_labels", index))
	}
	for index, mapping := range run.Request.RelationshipMappings {
		check(mapping.MappingLabels, fmt.Sprintf("$.projection_config.relationship_mappings[%d].mapping_labels", index))
	}
	for index, entity := range run.Request.SourceEntities {
		check(entity.Labels, fmt.Sprintf("$.source_entities[%d].labels", index))
	}
	for index, relationship := range run.Request.SourceRelationships {
		check(relationship.Labels, fmt.Sprintf("$.source_relationships[%d].labels", index))
	}
	return issues
}

func sourceItemResourceLimitIssues(run ProjectionRun, itemID string, labelCount, propertyCount, metadataCount int) []ValidationIssue {
	issues := []ValidationIssue{}
	add := func(key string, observed, limit int) {
		if observed > limit {
			issues = append(issues, run.issue("error", "source_item_resource_limit_exceeded", "source_item", itemID, nil, map[string]any{"source_item_id": itemID, "limit_key": key, "limit": limit, "observed": observed}))
		}
	}
	add("max_labels_per_source_item", labelCount, graphProjectionLimits.MaxLabelsPerSourceItem)
	add("max_properties_per_object", propertyCount, graphProjectionLimits.MaxPropertiesPerObject)
	add("max_metadata_keys_per_object", metadataCount, graphProjectionLimits.MaxMetadataKeysPerObject)
	return issues
}

func sourceItemWithinLimits(labelCount, propertyCount, metadataCount int) bool {
	return labelCount <= graphProjectionLimits.MaxLabelsPerSourceItem &&
		propertyCount <= graphProjectionLimits.MaxPropertiesPerObject &&
		metadataCount <= graphProjectionLimits.MaxMetadataKeysPerObject
}

func sourceItemValuesWithinLimits(objects ...map[string]any) bool {
	for _, object := range objects {
		for _, value := range object {
			if longestPropertyString(value) > graphProjectionLimits.MaxStringPropertyValueLength {
				return false
			}
		}
	}
	return true
}

func projectedOutputLimitIssues(run ProjectionRun, vertices []Vertex, edges []Edge) []ValidationIssue {
	issues := []ValidationIssue{}
	add := func(key string, observed, limit int) {
		if observed > limit {
			issues = append(issues, run.issue("fatal", "projected_output_limit_exceeded", "graph_view", run.GraphViewID, nil, map[string]any{"limit_key": key, "limit": limit, "observed": observed}))
		}
	}
	add("max_projected_vertices", len(vertices), graphProjectionLimits.MaxProjectedVertices)
	add("max_projected_edges", len(edges), graphProjectionLimits.MaxProjectedEdges)
	for _, vertex := range vertices {
		add("max_properties_per_object", len(vertex.Properties), graphProjectionLimits.MaxPropertiesPerObject)
		add("max_metadata_keys_per_object", len(vertex.Metadata.MappedMetadata), graphProjectionLimits.MaxMetadataKeysPerObject)
	}
	for _, edge := range edges {
		add("max_properties_per_object", len(edge.Properties), graphProjectionLimits.MaxPropertiesPerObject)
		add("max_metadata_keys_per_object", len(edge.Metadata.MappedMetadata), graphProjectionLimits.MaxMetadataKeysPerObject)
	}
	return issues
}

func emitDirectVertices(run ProjectionRun) ([]Vertex, map[string][]Vertex, []ValidationIssue) {
	vertices := []Vertex{}
	issues := []ValidationIssue{}
	bySource := map[string][]Vertex{}
	declaredKinds := stringSet(run.Request.ProjectionConfig.DeclaredSourceEntityKinds)
	for _, entity := range run.Request.SourceEntities {
		if !sourceItemWithinLimits(len(entity.Labels), len(entity.Properties), len(entity.Metadata)) || !sourceItemValuesWithinLimits(entity.Properties, entity.Metadata) {
			continue
		}
		if !validIdentifier(entity.SourceEntityID) || !validIdentifier(entity.SourceEntityKind) {
			issues = append(issues, run.issue("fatal", "invalid_source_entity", "source_entity", entity.SourceEntityID, nil, map[string]any{"reason_code": "invalid_identifier"}))
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
	return vertices, bySource, issues
}

func emitDirectEdges(run ProjectionRun, vertexBySource map[string][]Vertex) ([]Edge, []ValidationIssue) {
	edges := []Edge{}
	issues := []ValidationIssue{}
	declaredKinds := stringSet(run.Request.ProjectionConfig.DeclaredSourceRelationshipKinds)
	for _, relationship := range run.Request.SourceRelationships {
		if !sourceItemWithinLimits(len(relationship.Labels), len(relationship.Properties), len(relationship.Metadata)) || !sourceItemValuesWithinLimits(relationship.Properties, relationship.Metadata) {
			continue
		}
		if !validIdentifier(relationship.SourceRelationshipID) || !validIdentifier(relationship.SourceRelationshipKind) {
			issues = append(issues, run.issue("fatal", "invalid_source_relationship", "source_relationship", relationship.SourceRelationshipID, nil, map[string]any{"reason_code": "invalid_identifier"}))
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
	return edges, issues
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

func emitAggregations(run ProjectionRun, directVertices []Vertex, directEdges []Edge) ([]Vertex, []Edge, []ValidationIssue) {
	vertices := []Vertex{}
	edges := []Edge{}
	availableVertices := append([]Vertex(nil), directVertices...)
	availableEdges := append([]Edge(nil), directEdges...)
	issues := []ValidationIssue{}
	aggregatedVerticesByRuleAndDigest := map[string]map[string]Vertex{}
	for _, rule := range run.Request.ProjectionConfig.AggregationRules {
		if rule.TargetScope != "vertex" {
			continue
		}
		groups := groupContributors(run, rule, availableVertices, availableEdges, &issues)
		for _, digest := range sortedGroupDigests(groups) {
			contributors := groups[digest]
			vertexID, _ := generatedID("vx_", "GPVERTEX1\n", "aggregated_vertex", ProjectionSchemaID, run.GraphViewID, rule.AggregationIdentityDigest, digest)
			props := mergeAggregateProperties(run, rule, digest, contributors, "vertex", rule.ProjectedKind, vertexID, &issues)
			mappedMetadata := mergeAggregateMetadata(run, rule, digest, contributors, "vertex", rule.ProjectedKind, vertexID, &issues)
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
	for _, rule := range run.Request.ProjectionConfig.AggregationRules {
		if rule.TargetScope != "edge" || rule.EndpointGrouping == nil {
			continue
		}
		aggregationKinds := aggregationProjectedKinds(run.Request.ProjectionConfig.AggregationRules)
		groups := groupContributors(run, rule, availableVertices, availableEdges, &issues)
		for _, digest := range sortedGroupDigests(groups) {
			contributors := groups[digest]
			srcRuleID := rule.EndpointGrouping.SourceVertexAggregationRuleID
			dstRuleID := rule.EndpointGrouping.DestinationVertexAggregationRuleID
			srcDigest, srcMissingField, srcOK := endpointDigest(srcRuleID, "vertex", aggregationKinds[srcRuleID], rule.EndpointGrouping.SourceGroupingKeys, contributors)
			dstDigest, dstMissingField, dstOK := endpointDigest(dstRuleID, "vertex", aggregationKinds[dstRuleID], rule.EndpointGrouping.DestinationGroupingKeys, contributors)
			src := aggregatedVerticesByRuleAndDigest[rule.EndpointGrouping.SourceVertexAggregationRuleID][srcDigest]
			dst := aggregatedVerticesByRuleAndDigest[rule.EndpointGrouping.DestinationVertexAggregationRuleID][dstDigest]
			if !srcOK || !dstOK || src.VertexID == "" || dst.VertexID == "" {
				if rule.EndpointGrouping.MissingEndpointBehavior == "error" {
					issues = append(issues, endpointResolutionIssues(run, rule, "src", srcDigest, srcMissingField, srcOK, src.VertexID != "")...)
					issues = append(issues, endpointResolutionIssues(run, rule, "dst", dstDigest, dstMissingField, dstOK, dst.VertexID != "")...)
				}
				continue
			}
			edgeID, _ := generatedID("ed_", "GPEDGE1\n", "aggregated_edge", ProjectionSchemaID, run.GraphViewID, rule.AggregationIdentityDigest, src.VertexID, dst.VertexID, rule.EdgeDirection, digest)
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
				Labels:      run.Request.ProjectionConfig.DefaultEdgeLabels,
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
	return vertices, edges, issues
}

func sortedGroupDigests(groups map[string][]contributor) []string {
	digests := make([]string, 0, len(groups))
	for digest := range groups {
		digests = append(digests, digest)
	}
	sort.Strings(digests)
	return digests
}

func endpointResolutionIssues(run ProjectionRun, rule AggregationRule, side, digest, missingField string, digestOK, matched bool) []ValidationIssue {
	if !digestOK {
		return []ValidationIssue{run.issue("error", "aggregation_endpoint_missing", "mapping_rule", rule.AggregationRuleID, missingField, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "endpoint_side": side, "reason_code": "endpoint_key_missing", "endpoint_digest": nil, "field_path": missingField})}
	}
	if !matched {
		return []ValidationIssue{run.issue("error", "aggregation_endpoint_missing", "mapping_rule", rule.AggregationRuleID, nil, map[string]any{"aggregation_rule_id": rule.AggregationRuleID, "endpoint_side": side, "reason_code": "endpoint_vertex_not_found", "endpoint_digest": digest, "field_path": nil})}
	}
	return nil
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
	KindName     string
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
				if rule.MissingGroupingKeyBehavior == "use_null" {
					values = append(values, nil)
					continue
				}
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
	declaredEntityKinds := stringSet(run.Request.ProjectionConfig.DeclaredSourceEntityKinds)
	declaredRelationshipKinds := stringSet(run.Request.ProjectionConfig.DeclaredSourceRelationshipKinds)
	switch rule.InputScope {
	case "source_entity":
		for i := range run.Request.SourceEntities {
			entity := &run.Request.SourceEntities[i]
			if entity.SourceEntityKind != rule.InputKind || !validIdentifier(entity.SourceEntityID) || !validIdentifier(entity.SourceEntityKind) ||
				!sourceItemWithinLimits(len(entity.Labels), len(entity.Properties), len(entity.Metadata)) ||
				!sourceItemValuesWithinLimits(entity.Properties, entity.Metadata) ||
				!declaredEntityKinds[entity.SourceEntityKind] ||
				!filtersMatchEntity(run.Request.Filters.EntityFilters, *entity) {
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
				!filtersMatchRelationship(run.Request.Filters.RelationshipFilters, *relationship) {
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

func mergeAggregateProperties(run ProjectionRun, rule AggregationRule, groupingDigest string, contributors []contributor, targetScope, targetKind, outputID string, issues *[]ValidationIssue) map[string]any {
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

func mergeAggregateMetadata(run ProjectionRun, rule AggregationRule, groupingDigest string, contributors []contributor, targetScope, targetKind, outputID string, issues *[]ValidationIssue) map[string]any {
	metadata := map[string]any{}
	for _, mapping := range run.Request.ProjectionConfig.MetadataMappings {
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

func deriveProperties(run ProjectionRun, targetScope, targetKind string, entity SourceEntity, relationship *SourceRelationship, graphSource map[string]any, outputID string) (map[string]any, []ValidationIssue) {
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

func deriveGraphProperties(run ProjectionRun, issues *[]ValidationIssue) map[string]any {
	properties, propertyIssues := deriveProperties(run, "graph_view", "*", SourceEntity{}, nil, run.Request.SourceMetadata, run.GraphViewID)
	*issues = append(*issues, propertyIssues...)
	return properties
}

func deriveMetadata(run ProjectionRun, targetScope, targetKind string, entity SourceEntity, relationship *SourceRelationship, outputID string) (map[string]any, []ValidationIssue) {
	metadata := map[string]any{}
	issues := []ValidationIssue{}
	for _, mapping := range run.Request.ProjectionConfig.MetadataMappings {
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

func candidateDefinitionFromProperty(definition PropertyDefinition) candidateDefinition {
	return candidateDefinition{ProjectedType: definition.ProjectedType, DefaultValue: definition.DefaultValue, HasDefaultValue: definition.HasDefaultValue, MissingBehavior: definition.MissingBehavior, SourceNullBehavior: definition.SourceNullBehavior, NullOutputPolicy: definition.NullOutputPolicy}
}

func candidateDefinitionFromMetadata(mapping MetadataMapping) candidateDefinition {
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

func propertyValueIssue(run ProjectionRun, code, projectedKey, expectedType, sourceFieldPath, outputID string, value any, aggregationRuleID, groupingDigest, contributorID string) ValidationIssue {
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

func buildSchemaRegistry(run ProjectionRun, vertices []Vertex, edges []Edge) SchemaRegistry {
	vertexKinds := map[string]*VertexKindSchema{}
	for _, mapping := range run.Request.ProjectionConfig.EntityMappings {
		item := ensureVertexKind(vertexKinds, mapping.ProjectedVertexKind)
		item.SourceEntityKinds = append(item.SourceEntityKinds, mapping.SourceEntityKind)
		item.Labels = append(item.Labels, run.Request.ProjectionConfig.DefaultVertexLabels...)
		if mapping.LabelPolicy == "mapping_only" || mapping.LabelPolicy == "mapping_then_source" {
			item.Labels = append(item.Labels, mapping.MappingLabels...)
		}
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
		if mapping.LabelPolicy == "mapping_only" || mapping.LabelPolicy == "mapping_then_source" {
			item.Labels = append(item.Labels, mapping.MappingLabels...)
		}
		if mapping.LabelPolicy == "preserve_source" || mapping.LabelPolicy == "mapping_then_source" {
			item.SourceLabelsPreserved = true
		}
		if mapping.EmitReverseEdge {
			reverse := ensureEdgeKind(edgeKinds, mapping.ReverseEdgeKind)
			reverse.SourceRelationshipKinds = append(reverse.SourceRelationshipKinds, mapping.SourceRelationshipKind)
			reverse.Directions = append(reverse.Directions, "directed")
			reverse.Labels = append(reverse.Labels, item.Labels...)
			reverse.SourceLabelsPreserved = item.SourceLabelsPreserved
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
		reference := PropertySchemaReference{ProjectedKey: definition.ProjectedKey, ProjectedType: definition.ProjectedType, Required: definition.Required, NullableOutput: definition.NullOutputPolicy == "emit_null"}
		if definition.TargetScope == "vertex" {
			for key, item := range vertexKinds {
				if definition.TargetKind == "*" || definition.TargetKind == key {
					item.Properties = append(item.Properties, reference)
				}
			}
		}
		if definition.TargetScope == "edge" {
			for key, item := range edgeKinds {
				if definition.TargetKind == "*" || definition.TargetKind == key {
					item.Properties = append(item.Properties, reference)
				}
			}
		}
	}
	registry := SchemaRegistry{}
	for _, item := range vertexKinds {
		item.SourceEntityKinds = uniqueSortedStrings(item.SourceEntityKinds)
		item.AggregationRuleIDs = uniqueSortedStrings(item.AggregationRuleIDs)
		item.Labels = uniqueSortedStrings(item.Labels)
		sort.Slice(item.Properties, func(i, j int) bool { return item.Properties[i].ProjectedKey < item.Properties[j].ProjectedKey })
		registry.VertexKinds = append(registry.VertexKinds, *item)
	}
	for _, item := range edgeKinds {
		item.SourceRelationshipKinds = uniqueSortedStrings(item.SourceRelationshipKinds)
		item.AggregationRuleIDs = uniqueSortedStrings(item.AggregationRuleIDs)
		item.Directions = sortDirections(item.Directions)
		item.Labels = uniqueSortedStrings(item.Labels)
		sort.Slice(item.Properties, func(i, j int) bool { return item.Properties[i].ProjectedKey < item.Properties[j].ProjectedKey })
		registry.EdgeKinds = append(registry.EdgeKinds, *item)
	}
	for _, definition := range run.Request.PropertyDefinitions {
		for _, targetKind := range expandedTargetKinds(definition.TargetScope, definition.TargetKind, vertexKinds, edgeKinds) {
			registry.PropertyKeys = append(registry.PropertyKeys, PropertySchema{TargetScope: definition.TargetScope, TargetKind: targetKind, ProjectedKey: definition.ProjectedKey, ProjectedType: definition.ProjectedType, Required: definition.Required, NullableOutput: definition.NullOutputPolicy == "emit_null", MissingBehavior: definition.MissingBehavior, SourceNullBehavior: definition.SourceNullBehavior})
		}
	}
	for _, mapping := range run.Request.ProjectionConfig.MetadataMappings {
		for _, targetKind := range expandedTargetKinds(mapping.TargetScope, mapping.TargetKind, vertexKinds, edgeKinds) {
			registry.MetadataKeys = append(registry.MetadataKeys, MetadataSchema{TargetScope: mapping.TargetScope, TargetKind: targetKind, ProjectedMetadataKey: mapping.ProjectedMetadataKey, ProjectedType: mapping.ProjectedType, Required: mapping.Required, NullableOutput: mapping.NullOutputPolicy == "emit_null", MissingBehavior: mapping.MissingBehavior, SourceNullBehavior: mapping.SourceNullBehavior})
		}
	}
	sort.Slice(registry.VertexKinds, func(i, j int) bool { return registry.VertexKinds[i].VertexKind < registry.VertexKinds[j].VertexKind })
	sort.Slice(registry.EdgeKinds, func(i, j int) bool { return registry.EdgeKinds[i].EdgeKind < registry.EdgeKinds[j].EdgeKind })
	sort.Slice(registry.PropertyKeys, func(i, j int) bool {
		a, b := registry.PropertyKeys[i], registry.PropertyKeys[j]
		return a.TargetScope+"|"+a.TargetKind+"|"+a.ProjectedKey < b.TargetScope+"|"+b.TargetKind+"|"+b.ProjectedKey
	})
	sort.Slice(registry.MetadataKeys, func(i, j int) bool {
		a, b := registry.MetadataKeys[i], registry.MetadataKeys[j]
		return a.TargetScope+"|"+a.TargetKind+"|"+a.ProjectedMetadataKey < b.TargetScope+"|"+b.TargetKind+"|"+b.ProjectedMetadataKey
	})
	return registry
}

func expandedTargetKinds(scope, targetKind string, vertexKinds map[string]*VertexKindSchema, edgeKinds map[string]*EdgeKindSchema) []string {
	if targetKind != "*" || scope == "graph_view" {
		return []string{targetKind}
	}
	kinds := []string{}
	if scope == "vertex" {
		for kind := range vertexKinds {
			kinds = append(kinds, kind)
		}
	} else if scope == "edge" {
		for kind := range edgeKinds {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	return kinds
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

func validationSummary(run ProjectionRun, discovered []ValidationIssue) ValidationSummary {
	issues := append([]ValidationIssue(nil), discovered...)
	if len(issues) > graphProjectionLimits.MaxValidationIssues {
		issues = issues[:graphProjectionLimits.MaxValidationIssues-1]
		issues = append(issues, run.issue("fatal", "validation_issue_limit_exceeded", "projection_input", "projection_input", nil, map[string]any{"limit": graphProjectionLimits.MaxValidationIssues}))
	}
	capIssue := len(issues) > 0 && issues[len(issues)-1].Code == "validation_issue_limit_exceeded"
	sortEnd := len(issues)
	if capIssue {
		sortEnd--
	}
	severityRank := map[string]int{"fatal": 0, "error": 1, "warning": 2, "info": 3}
	sort.Slice(issues[:sortEnd], func(i, j int) bool {
		left, right := issues[i], issues[j]
		leftKey := fmt.Sprintf("%d|%s|%s|%s|%s", severityRank[left.Severity], left.Code, left.TargetKind, left.TargetID, left.IssueID)
		rightKey := fmt.Sprintf("%d|%s|%s|%s|%s", severityRank[right.Severity], right.Code, right.TargetKind, right.TargetID, right.IssueID)
		return leftKey < rightKey
	})
	status := "passed"
	for _, issue := range issues {
		if issue.Severity == "warning" && status == "passed" {
			status = "passed_with_warnings"
		}
		if issue.Severity == "error" {
			status = "passed_with_errors"
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
		refs = append(refs, SourceRef{RefKind: contributor.Kind, RefID: contributor.ID, RefKindName: contributor.KindName, ContributorSortKey: contributor.SortKey})
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
		if len(parts) == 2 {
			switch parts[1] {
			case "vertex_id", "edge_id", "src_vertex_id", "dst_vertex_id", "direction":
				return true
			}
		}
		if len(parts) == 3 && parts[1] == "metadata" {
			if !validPropertyKey(parts[2]) || reservedSystemMetadataTerminal(parts[2]) {
				return false
			}
			return true
		}
		return len(parts) == 3 && parts[1] == "properties" && validPropertyKey(parts[2])
	default:
		return false
	}
}

func reservedSystemMetadataTerminal(value string) bool {
	switch value {
	case "mapping_rule_id", "aggregation_rule_id", "aggregation_source_refs", "is_reverse_edge", "reverse_of_edge_id", "mapped_metadata", "previous_projection_run_id", "projection_config_digest", "projection_source_digest", "invalidation":
		return true
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
		stringValue, ok := value.(string)
		if !ok || utf8.RuneCountInString(stringValue) > graphProjectionLimits.MaxStringPropertyValueLength {
			return false
		}
		if projectedType == "identifier" && ok {
			return validIdentifier(stringValue)
		}
		if projectedType == "timestamp" {
			_, err := parseTimestamp(stringValue)
			return err == nil
		}
		return true
	case "integer":
		switch typed := value.(type) {
		case int, int64:
			return true
		case fmt.Stringer:
			return validFiniteInteger(typed.String())
		default:
			return validFiniteInteger(fmt.Sprint(value))
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
			if !ok || utf8.RuneCountInString(stringValue) > graphProjectionLimits.MaxStringPropertyValueLength {
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
	if !present {
		return filter.IncludeIfMissing
	}
	switch filter.Operator {
	case "exists":
		return true
	case "equals":
		return canonicalValueKey(value) == canonicalValueKey(filter.Value)
	case "not_equals":
		return canonicalValueKey(value) != canonicalValueKey(filter.Value)
	case "in":
		candidates, ok := filter.Value.([]any)
		if !ok {
			return false
		}
		if array, ok := value.([]any); ok {
			for _, item := range array {
				for _, candidate := range candidates {
					if canonicalValueKey(item) == canonicalValueKey(candidate) {
						return true
					}
				}
			}
			return false
		}
		for _, candidate := range candidates {
			if canonicalValueKey(value) == canonicalValueKey(candidate) {
				return true
			}
		}
		return false
	default:
		return false
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
		MaxTraversalDepth:               graphProjectionLimits.MaxTraversalDepth,
		MaxTraversalSeedVertices:        graphProjectionLimits.MaxTraversalSeedVertices,
		MaxKindFilters:                  graphProjectionLimits.MaxTraversalKindFilters,
	}
}

func formatLifecycleTimestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000000Z")
}
