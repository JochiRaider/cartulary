package graphprojection

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
)

type projectionWork struct {
	budget         *resultBudgetV2
	defaultValues  map[defaultKeyV2]any
	issues         issueCollector
	Request        projectionRequest
	GraphViewID    string
	IdentityDigest string
}

func projectSemanticGraph(ctx context.Context, run *projectionWork, skeleton ProjectionResultV2) (ProjectionResultV2, error) {
	if err := ctx.Err(); err != nil {
		return ProjectionResultV2{}, err
	}
	validateAdmittedRequest(run)
	if run.issues.err != nil {
		return ProjectionResultV2{}, run.issues.err
	}
	if hasFatalIssue(run.issues.entries) {
		return ProjectionResultV2{ValidationSummary: validationSummary(run.issues.entries)}, nil
	}

	// Mapping admission has priority over unresolved source references.
	if len(run.issues.entries) == 0 {
		if err := validateSubmittedEndpointsV2(ctx, run.Request); err != nil {
			return ProjectionResultV2{}, err
		}
	}

	skeleton.SchemaRegistry = buildSchemaRegistry(run)
	if err := run.budget.charge(skeleton.Resource()); err != nil {
		return ProjectionResultV2{}, err
	}
	state := projectionState{budget: run.budget, kinds: aggregationProjectedKinds(run.Request.projectionConfig.AggregationRules)}
	if err := state.emitDirectVertices(ctx, run); err != nil {
		return ProjectionResultV2{}, err
	}
	if err := state.emitDirectEdges(ctx, run); err != nil {
		return ProjectionResultV2{}, err
	}
	if err := phaseCheckpointV2(ctx); err != nil {
		return ProjectionResultV2{}, err
	}
	if err := state.emitAggregations(ctx, run); err != nil {
		return ProjectionResultV2{}, err
	}
	vertices, edges := state.vertices, state.edges
	sortVertices(vertices)
	sortEdges(edges)
	projectedOutputLimitIssues(run, vertices, edges)
	if run.issues.err != nil {
		return ProjectionResultV2{}, run.issues.err
	}
	if hasFatalIssue(run.issues.entries) {
		return ProjectionResultV2{ValidationSummary: validationSummary(run.issues.entries)}, nil
	}

	mappedMetadata := deriveMetadata(run, "graph_view", "*", sourceEntity{}, nil, run.GraphViewID)
	if len(mappedMetadata) > graphProjectionLimits.MaxPropertyKeys {
		run.recordIssue("fatal", "projected_output_limit_exceeded", "graph_view", run.GraphViewID, nil, map[string]any{"limit_key": "maximum_property_keys", "limit": graphProjectionLimits.MaxPropertyKeys, "observed": len(mappedMetadata)})
	}
	if run.issues.err != nil {
		return ProjectionResultV2{}, run.issues.err
	}
	summary := validationSummary(run.issues.entries)
	if hasFatalIssue(run.issues.entries) {
		return ProjectionResultV2{ValidationSummary: summary}, nil
	}

	graph := ProjectionResultV2{
		MappedMetadata:       mappedMetadata,
		SchemaRegistry:       skeleton.SchemaRegistry,
		Vertices:             vertices,
		Edges:                edges,
		ValidationSummary:    summary,
		ConsumerCapabilities: defaultConsumerCapabilities(),
	}
	if err := validateProjectedGraph(graph); err != nil {
		run.recordIssue("fatal", "output_schema_violation", "graph_view", run.GraphViewID, "$", map[string]any{"field": "$", "reason_code": err.Error()})
		if run.issues.err != nil {
			return ProjectionResultV2{}, run.issues.err
		}
		return ProjectionResultV2{ValidationSummary: validationSummary(run.issues.entries)}, nil
	}
	return graph, nil
}

func validateProjectedGraph(graph ProjectionResultV2) error {
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
	if _, err := canonicalJSON(graph.outputResource()); err != nil {
		return fmt.Errorf("canonical_serialization_invalid")
	}
	return nil
}

func buildSchemaRegistry(run *projectionWork) SchemaRegistry {
	vertexKinds := map[string]*VertexKindSchema{}
	for _, mapping := range run.Request.projectionConfig.EntityMappings {
		item := ensureVertexKind(vertexKinds, mapping.ProjectedVertexKind)
		item.SourceEntityKinds = append(item.SourceEntityKinds, mapping.SourceEntityKind)
		labels, preserve := mappingLabelsV2(mapping.MappingLabels, mapping.LabelPolicy)
		item.Labels = append(item.Labels, labels...)
		item.SourceLabelsPreserved = item.SourceLabelsPreserved || preserve
	}
	for _, rule := range run.Request.projectionConfig.AggregationRules {
		if rule.TargetScope == "vertex" {
			item := ensureVertexKind(vertexKinds, rule.ProjectedKind)
			item.AggregationRuleIDs = append(item.AggregationRuleIDs, rule.AggregationRuleID)
		}
	}
	edgeKinds := map[string]*EdgeKindSchema{}
	for _, mapping := range run.Request.projectionConfig.RelationshipMappings {
		item := ensureEdgeKind(edgeKinds, mapping.ProjectedEdgeKind)
		item.SourceRelationshipKinds = append(item.SourceRelationshipKinds, mapping.SourceRelationshipKind)
		item.Directions = append(item.Directions, mappingDirectionPoliciesV2[mapping.DirectionPolicy].directions...)
		labels, preserve := mappingLabelsV2(mapping.MappingLabels, mapping.LabelPolicy)
		item.Labels = append(item.Labels, labels...)
		item.SourceLabelsPreserved = item.SourceLabelsPreserved || preserve
		if mapping.EmitReverseEdge {
			reverse := ensureEdgeKind(edgeKinds, mapping.ReverseEdgeKind)
			reverse.SourceRelationshipKinds = append(reverse.SourceRelationshipKinds, mapping.SourceRelationshipKind)
			reverse.Directions = append(reverse.Directions, "directed")
			reverse.Labels = append(reverse.Labels, labels...)
			reverse.SourceLabelsPreserved = reverse.SourceLabelsPreserved || preserve
		}
	}
	for _, rule := range run.Request.projectionConfig.AggregationRules {
		if rule.TargetScope == "edge" {
			item := ensureEdgeKind(edgeKinds, rule.ProjectedKind)
			item.AggregationRuleIDs = append(item.AggregationRuleIDs, rule.AggregationRuleID)
			item.Directions = append(item.Directions, rule.EdgeDirection)
		}
	}
	for _, definition := range run.Request.PropertyDefinitions {
		reference := PropertySchemaReference{ProjectedKey: definition.ProjectedKey, ProjectedType: definition.ProjectedType, Required: definition.Required, NullableOutput: definition.NullOutputPolicy == "emit_null"}
		for _, kind := range run.Request.targets.expand(definition.TargetScope, definition.TargetKind) {
			switch definition.TargetScope {
			case "vertex":
				if item := vertexKinds[kind]; item != nil {
					item.Properties = append(item.Properties, reference)
				}
			case "edge":
				if item := edgeKinds[kind]; item != nil {
					item.Properties = append(item.Properties, reference)
				}
			}
		}
	}
	registry := SchemaRegistry{}
	for _, item := range vertexKinds {
		item.SourceEntityKinds = uniqueSortedStrings(item.SourceEntityKinds)
		item.AggregationRuleIDs = uniqueSortedStrings(item.AggregationRuleIDs)
		item.Labels = uniqueSortedStrings(append(item.Labels, run.Request.projectionConfig.DefaultVertexLabels...))
		sort.Slice(item.Properties, func(i, j int) bool { return item.Properties[i].ProjectedKey < item.Properties[j].ProjectedKey })
		registry.VertexKinds = append(registry.VertexKinds, *item)
	}
	for _, item := range edgeKinds {
		item.SourceRelationshipKinds = uniqueSortedStrings(item.SourceRelationshipKinds)
		item.AggregationRuleIDs = uniqueSortedStrings(item.AggregationRuleIDs)
		item.Directions = sortDirections(item.Directions)
		item.Labels = uniqueSortedStrings(append(item.Labels, run.Request.projectionConfig.DefaultEdgeLabels...))
		sort.Slice(item.Properties, func(i, j int) bool { return item.Properties[i].ProjectedKey < item.Properties[j].ProjectedKey })
		registry.EdgeKinds = append(registry.EdgeKinds, *item)
	}
	for _, definition := range run.Request.PropertyDefinitions {
		for _, targetKind := range run.Request.targets.expand(definition.TargetScope, definition.TargetKind) {
			registry.PropertyKeys = append(registry.PropertyKeys, PropertySchema{TargetScope: definition.TargetScope, TargetKind: targetKind, ProjectedKey: definition.ProjectedKey, ProjectedType: definition.ProjectedType, Required: definition.Required, NullableOutput: definition.NullOutputPolicy == "emit_null", MissingBehavior: definition.MissingBehavior, SourceNullBehavior: definition.SourceNullBehavior})
		}
	}
	for _, mapping := range run.Request.projectionConfig.MetadataMappings {
		for _, targetKind := range run.Request.targets.expand(mapping.TargetScope, mapping.TargetKind) {
			registry.MetadataKeys = append(registry.MetadataKeys, MetadataSchema{TargetScope: mapping.TargetScope, TargetKind: targetKind, ProjectedMetadataKey: mapping.ProjectedMetadataKey, ProjectedType: mapping.ProjectedType, Required: mapping.Required, NullableOutput: mapping.NullOutputPolicy == "emit_null", MissingBehavior: mapping.MissingBehavior, SourceNullBehavior: mapping.SourceNullBehavior})
		}
	}
	sort.Slice(registry.VertexKinds, func(i, j int) bool { return registry.VertexKinds[i].VertexKind < registry.VertexKinds[j].VertexKind })
	sort.Slice(registry.EdgeKinds, func(i, j int) bool { return registry.EdgeKinds[i].EdgeKind < registry.EdgeKinds[j].EdgeKind })
	sort.Slice(registry.PropertyKeys, func(i, j int) bool {
		a, b := registry.PropertyKeys[i], registry.PropertyKeys[j]
		return (declarationKeyV2{a.TargetScope, a.TargetKind, a.ProjectedKey}).less(declarationKeyV2{b.TargetScope, b.TargetKind, b.ProjectedKey})
	})
	sort.Slice(registry.MetadataKeys, func(i, j int) bool {
		a, b := registry.MetadataKeys[i], registry.MetadataKeys[j]
		return (declarationKeyV2{a.TargetScope, a.TargetKind, a.ProjectedMetadataKey}).less(declarationKeyV2{b.TargetScope, b.TargetKind, b.ProjectedMetadataKey})
	})
	return registry
}

type declarationKeyV2 struct{ scope, kind, key string }

func (a declarationKeyV2) less(b declarationKeyV2) bool {
	if a.scope != b.scope {
		return a.scope < b.scope
	}
	if a.kind != b.kind {
		return a.kind < b.kind
	}
	return a.key < b.key
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

func (run *projectionWork) sourceRefs(contributors []contributor) []SourceRef {
	var refs []SourceRef
	for _, contributor := range contributors {
		ref := SourceRef{RefKind: contributor.Kind, RefID: contributor.ID, RefKindName: contributor.KindName, ContributorSortKey: contributor.Order.key()}
		if run.budget != nil {
			if err := run.budget.arrayEntry(sourceRefResource(ref), len(refs)); err != nil {
				run.issues.err = err
				return nil
			}
		}
		refs = append(refs, ref)
	}
	return refs
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
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

func mappingLabelsV2(labels []string, policy string) ([]string, bool) {
	if policy == "preserve_source" {
		return nil, true
	}
	return labels, policy == "mapping_then_source"
}

func deriveLabels(defaults, mappingLabels, sourceLabels []string, policy string) ([]string, error) {
	mapped, preserve := mappingLabelsV2(mappingLabels, policy)
	if !preserve {
		sourceLabels = nil
	}
	seen := make(map[string]bool)
	labels := []string{}
	for _, group := range [][]string{defaults, mapped, sourceLabels} {
		for _, label := range group {
			if seen[label] {
				continue
			}
			if len(labels) == graphProjectionLimits.MaxLabelsPerObject {
				return nil, resourceProjectionErrorV2("maximum_labels_per_object", len(labels), len(labels)+1)
			}
			seen[label] = true
			labels = append(labels, label)
		}
	}
	slices.Sort(labels)
	return labels, nil
}

func valueMatchesType(projectedType string, value any) bool {
	_, ok := normalizeProjectedValueV2(projectedType, value)
	return ok
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

var registryDirectionsV2 = [...]string{"directed", "undirected", "bidirectional"}

// Admission, registry construction and stored identity proof use the same policy capabilities.
var mappingDirectionPoliciesV2 = map[string]struct {
	directions    []string
	allowsReverse bool
}{
	"preserve":          {directions: registryDirectionsV2[:]},
	"normalize_forward": {directions: []string{"directed"}, allowsReverse: true},
	"normalize_reverse": {directions: []string{"directed"}, allowsReverse: true},
	"undirected":        {directions: []string{"undirected"}},
	"bidirectional":     {directions: []string{"bidirectional"}},
}

func sortDirections(values []string) []string {
	seen := stringSet(values)
	ordered := make([]string, 0, len(registryDirectionsV2))
	for _, direction := range registryDirectionsV2 {
		if seen[direction] {
			ordered = append(ordered, direction)
		}
	}
	return ordered
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
