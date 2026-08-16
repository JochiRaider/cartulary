package graphprojection

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

type projectionWork struct {
	Request                       projectionRequest
	GraphViewID                   string
	NormalizedConfigurationSHA256 string
	NormalizedSourceSHA256        string
	IdentityDigest                string
}

type projectedGraph struct {
	Properties           map[string]any
	MappedMetadata       map[string]any
	SchemaRegistry       SchemaRegistry
	Vertices             []Vertex
	Edges                []Edge
	ValidationSummary    ValidationSummary
	ConsumerCapabilities ConsumerCapabilities
}

func projectSemanticGraph(ctx context.Context, run projectionWork) (projectedGraph, error) {
	if err := ctx.Err(); err != nil {
		return projectedGraph{}, err
	}
	issues := validateAdmittedRequest(run)
	if hasFatalIssue(issues) || len(issues) > graphProjectionLimits.MaxValidationIssues {
		return projectedGraph{ValidationSummary: validationSummary(run, issues)}, nil
	}

	vertices, vertexBySource, vertexIssues, err := emitDirectVertices(ctx, run)
	if err != nil {
		return projectedGraph{}, err
	}
	issues = append(issues, vertexIssues...)
	edges, edgeIssues, err := emitDirectEdges(ctx, run, vertexBySource)
	if err != nil {
		return projectedGraph{}, err
	}
	issues = append(issues, edgeIssues...)
	aggregatedVertices, aggregatedEdges, aggregationIssues, err := emitAggregations(ctx, run, vertices, edges)
	if err != nil {
		return projectedGraph{}, err
	}
	issues = append(issues, aggregationIssues...)
	vertices = append(vertices, aggregatedVertices...)
	edges = append(edges, aggregatedEdges...)
	sortVertices(vertices)
	sortEdges(edges)
	issues = append(issues, projectedOutputLimitIssues(run, vertices, edges)...)
	if hasFatalIssue(issues) || len(issues) > graphProjectionLimits.MaxValidationIssues {
		return projectedGraph{ValidationSummary: validationSummary(run, issues)}, nil
	}

	properties := deriveGraphProperties(run, &issues)
	mappedMetadata, metadataIssues := deriveMetadata(run, "graph_view", "*", sourceEntity{}, nil, run.GraphViewID)
	issues = append(issues, metadataIssues...)
	if len(properties) > graphProjectionLimits.MaxPropertiesPerObject {
		issues = append(issues, run.issue("fatal", "projected_output_limit_exceeded", "graph_view", run.GraphViewID, nil, map[string]any{"limit_key": "max_properties_per_object", "limit": graphProjectionLimits.MaxPropertiesPerObject, "observed": len(properties)}))
	}
	if len(mappedMetadata) > graphProjectionLimits.MaxMetadataKeysPerObject {
		issues = append(issues, run.issue("fatal", "projected_output_limit_exceeded", "graph_view", run.GraphViewID, nil, map[string]any{"limit_key": "max_metadata_keys_per_object", "limit": graphProjectionLimits.MaxMetadataKeysPerObject, "observed": len(mappedMetadata)}))
	}
	summary := validationSummary(run, issues)
	if hasFatalIssue(issues) || len(issues) > graphProjectionLimits.MaxValidationIssues {
		return projectedGraph{ValidationSummary: summary}, nil
	}

	graph := projectedGraph{
		Properties:           properties,
		MappedMetadata:       mappedMetadata,
		SchemaRegistry:       buildSchemaRegistry(run, vertices, edges),
		Vertices:             vertices,
		Edges:                edges,
		ValidationSummary:    summary,
		ConsumerCapabilities: defaultConsumerCapabilities(),
	}
	if err := validateProjectedGraph(run, graph); err != nil {
		issue := run.issue("fatal", "output_schema_violation", "graph_view", run.GraphViewID, "$", map[string]any{"field": "$", "reason_code": err.Error()})
		return projectedGraph{ValidationSummary: validationSummary(run, append(issues, issue))}, nil
	}
	return graph, nil
}

func validateProjectedGraph(run projectionWork, graph projectedGraph) error {
	vertexKinds := make(map[string]bool, len(graph.SchemaRegistry.VertexKinds))
	for _, schema := range graph.SchemaRegistry.VertexKinds {
		vertexKinds[schema.VertexKind] = true
	}
	edgeKinds := make(map[string]bool, len(graph.SchemaRegistry.EdgeKinds))
	for _, schema := range graph.SchemaRegistry.EdgeKinds {
		edgeKinds[schema.EdgeKind] = true
	}
	vertexIDs := make(map[string]bool, len(graph.Vertices))
	for index, vertex := range graph.Vertices {
		if !generatedIDPattern.MatchString(vertex.VertexID) || !strings.HasPrefix(vertex.VertexID, "vx_") || vertexIDs[vertex.VertexID] {
			return fmt.Errorf("id_mismatch")
		}
		if !vertexKinds[vertex.VertexKind] {
			return fmt.Errorf("schema_registry_mismatch")
		}
		if index > 0 && (graph.Vertices[index-1].SortKey > vertex.SortKey || (graph.Vertices[index-1].SortKey == vertex.SortKey && graph.Vertices[index-1].VertexID > vertex.VertexID)) {
			return fmt.Errorf("sort_order_invalid")
		}
		vertexIDs[vertex.VertexID] = true
	}
	edgeIDs := make(map[string]bool, len(graph.Edges))
	for index, edge := range graph.Edges {
		if !generatedIDPattern.MatchString(edge.EdgeID) || !strings.HasPrefix(edge.EdgeID, "ed_") || edgeIDs[edge.EdgeID] {
			return fmt.Errorf("id_mismatch")
		}
		if !vertexIDs[edge.SrcVertexID] || !vertexIDs[edge.DstVertexID] {
			return fmt.Errorf("reference_missing")
		}
		if !edgeKinds[edge.EdgeKind] {
			return fmt.Errorf("schema_registry_mismatch")
		}
		if index > 0 && (graph.Edges[index-1].SortKey > edge.SortKey || (graph.Edges[index-1].SortKey == edge.SortKey && graph.Edges[index-1].EdgeID > edge.EdgeID)) {
			return fmt.Errorf("sort_order_invalid")
		}
		edgeIDs[edge.EdgeID] = true
	}
	if _, err := canonicalJSON(projectedGraphResource(graph)); err != nil {
		return fmt.Errorf("canonical_serialization_invalid")
	}
	return nil
}

func projectedGraphResource(graph projectedGraph) map[string]any {
	vertices := make([]any, 0, len(graph.Vertices))
	for _, vertex := range graph.Vertices {
		vertices = append(vertices, vertexResource(vertex))
	}
	edges := make([]any, 0, len(graph.Edges))
	for _, edge := range graph.Edges {
		edges = append(edges, edgeResource(edge))
	}
	return map[string]any{
		"properties":            nonNilMap(graph.Properties),
		"mapped_metadata":       nonNilMap(graph.MappedMetadata),
		"schema_registry":       schemaRegistryResource(graph.SchemaRegistry),
		"vertices":              vertices,
		"edges":                 edges,
		"validation_summary":    validationSummaryResource(graph.ValidationSummary),
		"consumer_capabilities": consumerCapabilitiesResource(graph.ConsumerCapabilities),
	}
}

func buildSchemaRegistry(run projectionWork, vertices []Vertex, edges []Edge) SchemaRegistry {
	vertexKinds := map[string]*VertexKindSchema{}
	for _, mapping := range run.Request.projectionConfig.EntityMappings {
		item := ensureVertexKind(vertexKinds, mapping.ProjectedVertexKind)
		item.SourceEntityKinds = append(item.SourceEntityKinds, mapping.SourceEntityKind)
		item.Labels = append(item.Labels, run.Request.projectionConfig.DefaultVertexLabels...)
		if mapping.LabelPolicy == "mapping_only" || mapping.LabelPolicy == "mapping_then_source" {
			item.Labels = append(item.Labels, mapping.MappingLabels...)
		}
		if mapping.LabelPolicy == "preserve_source" || mapping.LabelPolicy == "mapping_then_source" {
			item.SourceLabelsPreserved = true
		}
	}
	for _, rule := range run.Request.projectionConfig.AggregationRules {
		if rule.TargetScope == "vertex" {
			item := ensureVertexKind(vertexKinds, rule.ProjectedKind)
			item.AggregationRuleIDs = append(item.AggregationRuleIDs, rule.AggregationRuleID)
			item.Labels = append(item.Labels, run.Request.projectionConfig.DefaultVertexLabels...)
		}
	}
	edgeKinds := map[string]*EdgeKindSchema{}
	for _, mapping := range run.Request.RelationshipMappings {
		item := ensureEdgeKind(edgeKinds, mapping.ProjectedEdgeKind)
		item.SourceRelationshipKinds = append(item.SourceRelationshipKinds, mapping.SourceRelationshipKind)
		item.Directions = append(item.Directions, projectedDirectionsForPolicy(mapping.DirectionPolicy)...)
		item.Labels = append(item.Labels, run.Request.projectionConfig.DefaultEdgeLabels...)
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
	for _, rule := range run.Request.projectionConfig.AggregationRules {
		if rule.TargetScope == "edge" {
			item := ensureEdgeKind(edgeKinds, rule.ProjectedKind)
			item.AggregationRuleIDs = append(item.AggregationRuleIDs, rule.AggregationRuleID)
			item.Directions = append(item.Directions, rule.EdgeDirection)
			item.Labels = append(item.Labels, run.Request.projectionConfig.DefaultEdgeLabels...)
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
	for _, mapping := range run.Request.projectionConfig.MetadataMappings {
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

func entityIDs(entities []sourceEntity) []string {
	out := make([]string, 0, len(entities))
	for _, entity := range entities {
		out = append(out, entity.SourceEntityID)
	}
	return out
}

func relationshipIDs(relationships []sourceRelationship) []string {
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
	case "mapping_rule_id", "aggregation_rule_id", "aggregation_source_refs", "is_reverse_edge", "reverse_of_edge_id", "mapped_metadata", "projection_result_id", "normalized_configuration_sha256", "normalized_source_sha256", "canonical_output_sha256":
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

func propertyApplies(definition propertyDefinition, targetScope, targetKind string) bool {
	return definition.TargetScope == targetScope && (definition.TargetKind == "*" || definition.TargetKind == targetKind)
}

func filtersMatchEntity(filters []filterPredicate, entity sourceEntity) bool {
	for _, filter := range filters {
		value, ok := sourceField(entity, nil, nil, filter.FieldPath)
		if !filterMatches(value, ok, filter) {
			return false
		}
	}
	return true
}

func filtersMatchRelationship(filters []filterPredicate, relationship sourceRelationship) bool {
	for _, filter := range filters {
		value, ok := sourceField(sourceEntity{}, &relationship, nil, filter.FieldPath)
		if !filterMatches(value, ok, filter) {
			return false
		}
	}
	return true
}

func filterMatches(value any, present bool, filter filterPredicate) bool {
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
		QueryShapes:                     []string{"read_edges", "read_exact_result", "read_vertices", "traverse"},
		SupportsDirectVertexLookup:      true,
		SupportsDirectEdgeLookup:        true,
		SupportsBreadthFirstTraversal:   true,
		SupportsAlternateTraversalOrder: []string{},
		MaxTraversalDepth:               graphProjectionLimits.MaxTraversalDepth,
		MaxTraversalSeedVertices:        graphProjectionLimits.MaxTraversalSeedVertices,
		MaxKindFilters:                  graphProjectionLimits.MaxTraversalKindFilters,
	}
}
