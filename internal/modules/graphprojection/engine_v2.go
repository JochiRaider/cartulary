package graphprojection

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"maps"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const projectionResultIdentityDomainV2 = "cartulary.graph_projection_result_identity.v2"

type InvocationContextV2 struct {
	GraphViewID       string
	SourceOwnerID     string
	CancellationCheck func(context.Context) error
}

type ProjectionErrorV2 struct {
	Code        string         `json:"code"`
	ReasonCode  string         `json:"reason_code"`
	RetryAction string         `json:"retry_action"`
	Details     map[string]any `json:"details"`
	cause       error          `json:"-"`
}

func (err *ProjectionErrorV2) Error() string {
	if err == nil {
		return ""
	}
	if err.ReasonCode == "" {
		return err.Code
	}
	return err.Code + ": " + err.ReasonCode
}

func (err *ProjectionErrorV2) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.cause
}

type ProjectionResultV2 struct {
	ProjectionSchemaID            string
	ProjectionResultID            string
	GraphViewID                   string
	SourceOwnerID                 string
	SourceSnapshotID              string
	ProjectionVersion             string
	NormalizedConfigurationSHA256 string
	NormalizedSourceSHA256        string
	CanonicalOutputSHA256         string
	MappedMetadata                map[string]any
	SchemaRegistry                SchemaRegistry
	Vertices                      []Vertex
	Edges                         []Edge
	ValidationSummary             ValidationSummary
	ConsumerCapabilities          ConsumerCapabilities
}

func ProjectV2(ctx context.Context, invocation InvocationContextV2, semanticInput []byte) (ProjectionResultV2, error) {
	return projectWithResultLimitV2(ctx, invocation, semanticInput, graphProjectionLimits.MaxResultBytes)
}

func projectWithResultLimitV2(ctx context.Context, invocation InvocationContextV2, semanticInput []byte, resultLimit int) (ProjectionResultV2, error) {
	if invocation.CancellationCheck != nil {
		checkpoint := newProjectionCheckpointContextV2(ctx, invocation.CancellationCheck)
		defer checkpoint.cancel()
		ctx = checkpoint
	}
	if err := phaseCheckpointV2(ctx); err != nil {
		return ProjectionResultV2{}, cancelledProjectionErrorV2(err)
	}
	if !validIdentifierV2(invocation.GraphViewID) {
		return ProjectionResultV2{}, invalidProjectionErrorV2("invalid_graph_view_id", "graph_view_id")
	}
	if !validIdentifierV2(invocation.SourceOwnerID) {
		return ProjectionResultV2{}, invalidProjectionErrorV2("invalid_source_owner_id", "source_owner_id")
	}
	if len(semanticInput) > graphProjectionLimits.MaxInputBytes {
		return ProjectionResultV2{}, resourceProjectionErrorV2("maximum_input_bytes", graphProjectionLimits.MaxInputBytes, graphProjectionLimits.MaxInputBytes+1)
	}

	request, normalizedConfiguration, normalizedSource, err := parseProjectionInputV2(ctx, semanticInput)
	if err != nil {
		return ProjectionResultV2{}, err
	}
	if err := contextProjectionErrorV2(ctx); err != nil {
		return ProjectionResultV2{}, err
	}
	configurationBytes, err := canonicalJSON(normalizeCanonicalValueV2(normalizedConfiguration))
	if err != nil {
		return ProjectionResultV2{}, computationProjectionErrorV2("canonical_configuration_failed", err)
	}
	sourceBytes, err := canonicalJSON(normalizeCanonicalValueV2(normalizedSource))
	if err != nil {
		return ProjectionResultV2{}, computationProjectionErrorV2("canonical_source_failed", err)
	}
	configurationDigest := sha256Hex(configurationBytes)
	sourceDigest := sha256Hex(sourceBytes)
	workIdentity := sha256Hex(lengthFramedIdentityTranscriptV2([]string{
		"cartulary.graph_projection_semantic_work.v2",
		invocation.GraphViewID,
		invocation.SourceOwnerID,
		request.SourceSnapshotID,
		request.projectionConfig.ProjectionVersion,
		configurationDigest,
		sourceDigest,
	}))
	work := &projectionWork{
		Request:        request,
		GraphViewID:    invocation.GraphViewID,
		IdentityDigest: workIdentity,
	}
	skeleton := ProjectionResultV2{
		ProjectionSchemaID: ProjectionSchemaIDV2, ProjectionResultID: "gpres_" + strings.Repeat("0", 64),
		GraphViewID: invocation.GraphViewID, SourceOwnerID: invocation.SourceOwnerID, SourceSnapshotID: request.SourceSnapshotID,
		ProjectionVersion: request.projectionConfig.ProjectionVersion, NormalizedConfigurationSHA256: configurationDigest,
		NormalizedSourceSHA256: sourceDigest, CanonicalOutputSHA256: strings.Repeat("0", 64),
		ValidationSummary: ValidationSummary{Status: "passed"}, ConsumerCapabilities: defaultConsumerCapabilities(),
	}
	work.budget = newResultBudgetV2(resultLimit)
	result, err := projectSemanticGraph(ctx, work, skeleton)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return ProjectionResultV2{}, cancelledProjectionErrorV2(err)
		}
		var projectionErr *ProjectionErrorV2
		if errors.As(err, &projectionErr) {
			return ProjectionResultV2{}, projectionErr
		}
		return ProjectionResultV2{}, computationProjectionErrorV2("projection_failed", err)
	}
	if result.ValidationSummary.IssueCount != 0 {
		return ProjectionResultV2{}, validationFailureV2(firstValidationReasonV2(result.ValidationSummary), result.ValidationSummary.IssueCount)
	}
	if len(result.Vertices) > graphProjectionLimits.MaxProjectedVertices {
		return ProjectionResultV2{}, resourceProjectionErrorV2("maximum_projected_vertices", graphProjectionLimits.MaxProjectedVertices, boundedObserved(len(result.Vertices), graphProjectionLimits.MaxProjectedVertices))
	}
	if len(result.Edges) > graphProjectionLimits.MaxProjectedEdges {
		return ProjectionResultV2{}, resourceProjectionErrorV2("maximum_projected_edges", graphProjectionLimits.MaxProjectedEdges, boundedObserved(len(result.Edges), graphProjectionLimits.MaxProjectedEdges))
	}

	// Validate the producer's content before normalization, then bind the final bytes.
	result.MappedMetadata = normalizeMapV2(result.MappedMetadata)
	result.Vertices = normalizeVerticesV2(result.Vertices)
	result.Edges = normalizeEdgesV2(result.Edges)
	result.ProjectionSchemaID = ProjectionSchemaIDV2
	result.GraphViewID = invocation.GraphViewID
	result.SourceOwnerID = invocation.SourceOwnerID
	result.SourceSnapshotID = request.SourceSnapshotID
	result.ProjectionVersion = request.projectionConfig.ProjectionVersion
	result.NormalizedConfigurationSHA256 = configurationDigest
	result.NormalizedSourceSHA256 = sourceDigest
	outputDigest, err := canonicalOutputDigestV2(result.outputResource())
	if err != nil {
		return ProjectionResultV2{}, computationProjectionErrorV2("canonical_output_failed", err)
	}
	result.CanonicalOutputSHA256 = outputDigest
	result.ProjectionResultID, err = deriveProjectionResultIDV2(result.ResultBindingV2())
	if err != nil {
		return ProjectionResultV2{}, computationProjectionErrorV2("identity_failed", err)
	}
	if err := phaseCheckpointV2(ctx); err != nil {
		return ProjectionResultV2{}, cancelledProjectionErrorV2(err)
	}
	return result, nil
}

type projectionCheckpointContextV2 struct {
	context.Context
	check  func(context.Context) error
	items  int
	cancel context.CancelFunc
}

func newProjectionCheckpointContextV2(parent context.Context, check func(context.Context) error) *projectionCheckpointContextV2 {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	return &projectionCheckpointContextV2{Context: ctx, check: check, cancel: cancel}
}

// Only the projection goroutine advances the item counter. Ordinary context
// observations, including error classification, never consume a checkpoint.
func itemCheckpointV2(ctx context.Context) error {
	if checkpoint, ok := ctx.(*projectionCheckpointContextV2); ok {
		checkpoint.items++
		if checkpoint.items == graphProjectionLimits.CancellationCheckIntervalItems {
			checkpoint.items = 0
			return checkpoint.checkNow()
		}
	}
	return ctx.Err()
}

func (ctx *projectionCheckpointContextV2) checkNow() error {
	if err := ctx.Context.Err(); err != nil {
		return err
	}
	if ctx.check(ctx.Context) != nil {
		ctx.cancel()
	}
	return ctx.Context.Err()
}

func phaseCheckpointV2(ctx context.Context) error {
	if checkpoint, ok := ctx.(*projectionCheckpointContextV2); ok {
		return checkpoint.checkNow()
	}
	return ctx.Err()
}

func (result ProjectionResultV2) ResultBindingV2() ResultBindingV2 {
	return ResultBindingV2{
		ProjectionResultID:            result.ProjectionResultID,
		GraphViewID:                   result.GraphViewID,
		SourceOwnerID:                 result.SourceOwnerID,
		SourceSnapshotID:              result.SourceSnapshotID,
		ProjectionSchemaID:            result.ProjectionSchemaID,
		ProjectionVersion:             result.ProjectionVersion,
		NormalizedConfigurationSHA256: result.NormalizedConfigurationSHA256,
		NormalizedSourceSHA256:        result.NormalizedSourceSHA256,
		CanonicalOutputSHA256:         result.CanonicalOutputSHA256,
	}
}

func deriveProjectionResultIDV2(binding ResultBindingV2) (string, error) {
	fields := []string{
		projectionResultIdentityDomainV2,
		binding.GraphViewID,
		binding.SourceOwnerID,
		binding.SourceSnapshotID,
		binding.ProjectionVersion,
		binding.NormalizedConfigurationSHA256,
		binding.NormalizedSourceSHA256,
		binding.CanonicalOutputSHA256,
	}
	for _, field := range fields {
		if !utf8.ValidString(field) {
			return "", errors.New("graphprojection: result identity field is not UTF-8")
		}
	}
	return "gpres_" + sha256Hex(lengthFramedIdentityTranscriptV2(fields)), nil
}

func lengthFramedIdentityTranscriptV2(fields []string) []byte {
	var transcript bytes.Buffer
	for _, field := range fields {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len([]byte(field))))
		transcript.Write(length[:])
		transcript.WriteString(field)
	}
	return transcript.Bytes()
}

func (result ProjectionResultV2) Resource() map[string]any {
	resource := result.outputResource()
	resource["projection_schema_id"] = result.ProjectionSchemaID
	resource["projection_result_id"] = result.ProjectionResultID
	resource["graph_view_id"] = result.GraphViewID
	resource["source_owner_id"] = result.SourceOwnerID
	resource["source_snapshot_id"] = result.SourceSnapshotID
	resource["projection_version"] = result.ProjectionVersion
	resource["normalized_configuration_sha256"] = result.NormalizedConfigurationSHA256
	resource["normalized_source_sha256"] = result.NormalizedSourceSHA256
	resource["canonical_output_sha256"] = result.CanonicalOutputSHA256
	return resource
}

func (result ProjectionResultV2) CompletedResult() (CompletedResultV2, error) {
	if len(result.Vertices) > MaximumResultVerticesV2 || len(result.Edges) > MaximumResultEdgesV2 {
		return CompletedResultV2{}, ErrResultV2Invalid
	}
	shell := result
	shell.Vertices, shell.Edges = nil, nil
	envelope := shell.Resource()
	budget := newResultBudgetV2(graphProjectionLimits.MaxResultBytes)
	if err := budget.charge(envelope); err != nil {
		return CompletedResultV2{}, err
	}
	// Materialize each object only after its predecessors fit the envelope.
	for _, vertex := range result.Vertices {
		if err := appendCompletedObjectV2(budget, envelope, "vertices", vertexResource(vertex)); err != nil {
			return CompletedResultV2{}, err
		}
	}
	for _, edge := range result.Edges {
		if err := appendCompletedObjectV2(budget, envelope, "edges", edgeResource(edge)); err != nil {
			return CompletedResultV2{}, err
		}
	}
	resultJSON, err := canonicalJSON(envelope)
	if err != nil {
		return CompletedResultV2{}, err
	}
	vertices := make([]ResultVertexV2, 0, len(result.Vertices))
	for i, vertex := range result.Vertices {
		body, err := canonicalJSON(envelope["vertices"].([]any)[i])
		if err != nil {
			return CompletedResultV2{}, err
		}
		vertices = append(vertices, ResultVertexV2{VertexID: vertex.VertexID, VertexKind: vertex.VertexKind, SortKey: vertex.SortKey, JSON: body})
	}
	edges := make([]ResultEdgeV2, 0, len(result.Edges))
	for i, edge := range result.Edges {
		body, err := canonicalJSON(envelope["edges"].([]any)[i])
		if err != nil {
			return CompletedResultV2{}, err
		}
		edges = append(edges, ResultEdgeV2{EdgeID: edge.EdgeID, EdgeKind: edge.EdgeKind, SrcVertexID: edge.SrcVertexID, DstVertexID: edge.DstVertexID, Direction: edge.Direction, SortKey: edge.SortKey, JSON: body})
	}
	return CompletedResultV2{Binding: result.ResultBindingV2(), ResultJSON: resultJSON, Vertices: vertices, Edges: edges}, nil
}

func appendCompletedObjectV2(budget *resultBudgetV2, envelope map[string]any, scope string, row map[string]any) error {
	rows := envelope[scope].([]any)
	if err := budget.arrayEntry(row, len(rows)); err != nil {
		return err
	}
	envelope[scope] = append(rows, row)
	return nil
}

func (result ProjectionResultV2) outputResource() map[string]any {
	vertices := make([]any, 0, len(result.Vertices))
	for _, vertex := range result.Vertices {
		vertices = append(vertices, vertexResource(vertex))
	}
	edges := make([]any, 0, len(result.Edges))
	for _, edge := range result.Edges {
		edges = append(edges, edgeResource(edge))
	}
	return map[string]any{
		"mapped_metadata":       nonNilMap(result.MappedMetadata),
		"schema_registry":       schemaRegistryResource(result.SchemaRegistry),
		"vertices":              vertices,
		"edges":                 edges,
		"validation_summary":    validationSummaryResource(result.ValidationSummary),
		"consumer_capabilities": consumerCapabilitiesResource(result.ConsumerCapabilities),
	}
}

func consumerCapabilitiesResource(capabilities ConsumerCapabilities) map[string]any {
	return map[string]any{
		"query_shapes":                       nonNilStrings(capabilities.QueryShapes),
		"supports_direct_vertex_lookup":      capabilities.SupportsDirectVertexLookup,
		"supports_direct_edge_lookup":        capabilities.SupportsDirectEdgeLookup,
		"supports_breadth_first_traversal":   capabilities.SupportsBreadthFirstTraversal,
		"supports_alternate_traversal_order": nonNilStrings(capabilities.SupportsAlternateTraversalOrder),
		"max_traversal_depth":                capabilities.MaxTraversalDepth,
		"max_traversal_seed_vertices":        capabilities.MaxTraversalSeedVertices,
		"max_kind_filters":                   capabilities.MaxKindFilters,
	}
}

func parseProjectionInputV2(ctx context.Context, data []byte) (projectionRequest, map[string]any, map[string]any, error) {
	if !utf8.Valid(data) {
		return projectionRequest{}, nil, nil, invalidProjectionErrorV2("invalid_utf8", "")
	}
	decoded, err := decodeInputV2(ctx, data)
	if err != nil {
		return projectionRequest{}, nil, nil, err
	}
	root, ok := decoded.(map[string]any)
	if !ok {
		return projectionRequest{}, nil, nil, invalidProjectionErrorV2("top_level_not_object", "$")
	}
	shape := inputShapeWalkV2{ctx: ctx}
	if err := requestShapeV2.validate(&shape, root, "$"); err != nil {
		return projectionRequest{}, nil, nil, err
	}
	if err := validateRequestScalarsV2(ctx, root); err != nil {
		return projectionRequest{}, nil, nil, err
	}
	if err := admitUniqueIdentifiersV2(ctx, root); err != nil {
		return projectionRequest{}, nil, nil, err
	}
	if err := validateV2ResourceBounds(root); err != nil {
		return projectionRequest{}, nil, nil, err
	}

	if err := admitInputResourceCountsV2(ctx, root); err != nil {
		return projectionRequest{}, nil, nil, err
	}

	if shape.oversizedArray {
		return projectionRequest{}, nil, nil, arrayLimitErrorV2()
	}
	targets, err := admitDeclarationTargetsV2(ctx, root)
	if err != nil {
		return projectionRequest{}, nil, nil, err
	}
	request := projectionRequest{
		targets:             targets,
		SourceSnapshotID:    root["source_snapshot_id"].(string),
		projectionConfig:    parseProjectionConfig(root["projection_config"].(map[string]any)),
		SourceEntities:      parseSourceEntities(root["source_entities"].([]any)),
		SourceRelationships: parseSourceRelationships(root["source_relationships"].([]any)),
		SourceMetadata:      objectMap(root["source_metadata"].(map[string]any)),
		filters:             parseFilters(root["filters"].(map[string]any)),
		PropertyDefinitions: parsePropertyDefinitions(root["property_definitions"].([]any)),
	}
	if err := normalizeProjectionRequestV2(&request); err != nil {
		return projectionRequest{}, nil, nil, err
	}
	configuration := map[string]any{
		"projection_config":    normalizedConfigObjectV2(request.projectionConfig),
		"filters":              filtersObject(request.filters),
		"property_definitions": propertyDefinitionsObject(request.PropertyDefinitions),
	}
	source := map[string]any{
		"source_snapshot_id":   request.SourceSnapshotID,
		"source_entities":      sourceEntitiesObject(request.SourceEntities),
		"source_relationships": sourceRelationshipsObject(request.SourceRelationships),
		"source_metadata":      request.SourceMetadata,
	}
	return request, configuration, source, nil
}

func normalizeProjectionRequestV2(request *projectionRequest) error {
	ordered, err := orderAggregationRulesV2(request.projectionConfig.AggregationRules)
	if err != nil {
		return err
	}
	normalizeProjectionRequest(request)
	sort.SliceStable(request.projectionConfig.EntityMappings, func(i, j int) bool {
		return request.projectionConfig.EntityMappings[i].MappingRuleID < request.projectionConfig.EntityMappings[j].MappingRuleID
	})
	sort.SliceStable(request.projectionConfig.RelationshipMappings, func(i, j int) bool {
		return request.projectionConfig.RelationshipMappings[i].MappingRuleID < request.projectionConfig.RelationshipMappings[j].MappingRuleID
	})
	sort.SliceStable(request.projectionConfig.MetadataMappings, func(i, j int) bool {
		return request.projectionConfig.MetadataMappings[i].MetadataMappingID < request.projectionConfig.MetadataMappings[j].MetadataMappingID
	})
	sort.SliceStable(request.PropertyDefinitions, func(i, j int) bool {
		return request.PropertyDefinitions[i].PropertyDefinitionID < request.PropertyDefinitions[j].PropertyDefinitionID
	})
	request.projectionConfig.AggregationRules = ordered
	return nil
}

func orderAggregationRulesV2(rules []aggregationRule) ([]aggregationRule, error) {
	type producerKey struct{ scope, kind string }
	byID := make(map[string]aggregationRule, len(rules))
	producers := make(map[producerKey][]string)
	for _, rule := range rules {
		if _, exists := byID[rule.AggregationRuleID]; exists {
			return nil, invalidAggregationRulesErrorV2()
		}
		byID[rule.AggregationRuleID] = rule
		key := producerKey{rule.TargetScope, rule.ProjectedKind}
		producers[key] = append(producers[key], rule.AggregationRuleID)
	}
	prerequisites := make(ruleDependenciesV2, len(rules))
	for _, rule := range rules {
		dependencies := map[string]struct{}{}
		var scope string
		switch rule.InputScope {
		case "projected_vertex":
			scope = "vertex"
		case "projected_edge":
			scope = "edge"
		}
		if scope != "" {
			addOtherProducersV2(dependencies, rule.AggregationRuleID, slices.Values(producers[producerKey{scope, rule.InputKind}]))
		}
		if endpoint := rule.endpointGrouping; endpoint != nil {
			for _, id := range []string{endpoint.SourceVertexAggregationRuleID, endpoint.DestinationVertexAggregationRuleID} {
				producer, exists := byID[id]
				if !exists || producer.TargetScope != "vertex" {
					return nil, invalidAggregationRulesErrorV2()
				}
				dependencies[id] = struct{}{}
			}
		}
		prerequisites[rule.AggregationRuleID] = dependencies
	}
	ids := prerequisites.orderedIDs()
	if len(ids) != len(rules) {
		return nil, invalidAggregationRulesErrorV2()
	}
	ordered := make([]aggregationRule, len(ids))
	for i, id := range ids {
		ordered[i] = byID[id]
	}
	return ordered, nil
}

// Each rule executes once. A cycle leaves at least one rule outside the ready batches.
func addOtherProducersV2(prerequisites map[string]struct{}, rule string, producers iter.Seq[string]) {
	for producer := range producers {
		if producer != rule {
			prerequisites[producer] = struct{}{}
		}
	}
}

type ruleDependenciesV2 map[string]map[string]struct{}

func (rules ruleDependenciesV2) orderedIDs() []string {
	pending := make(map[string]int, len(rules))
	dependents := make(map[string][]string, len(rules))
	ready := make([]string, 0, len(rules))
	for rule, dependencies := range rules {
		pending[rule] = len(dependencies)
		if len(dependencies) == 0 {
			ready = append(ready, rule)
		}
		for dependency := range dependencies {
			dependents[dependency] = append(dependents[dependency], rule)
		}
	}
	ordered := make([]string, 0, len(rules))
	for len(ready) != 0 {
		sort.Strings(ready)
		// Finish the complete ready batch before considering newly ready rules.
		var next []string
		for _, id := range ready {
			ordered = append(ordered, id)
			for _, dependent := range dependents[id] {
				pending[dependent]--
				if pending[dependent] == 0 {
					next = append(next, dependent)
				}
			}
		}
		ready = next
	}
	return ordered
}

func invalidAggregationRulesErrorV2() *ProjectionErrorV2 {
	return &ProjectionErrorV2{
		Code: "projection_validation_failed", ReasonCode: "invalid_aggregation_rule",
		RetryAction: "do_not_retry", Details: map[string]any{"issue_count": 1},
	}
}

func normalizedConfigObjectV2(config projectionConfig) map[string]any {
	return map[string]any{
		"projection_version":                 config.ProjectionVersion,
		"declared_source_entity_kinds":       config.DeclaredSourceEntityKinds,
		"declared_source_relationship_kinds": config.DeclaredSourceRelationshipKinds,
		"entity_mappings":                    entityMappingsObject(config.EntityMappings),
		"relationship_mappings":              relationshipMappingsObject(config.RelationshipMappings),
		"metadata_mappings":                  metadataMappingsObject(config.MetadataMappings),
		"aggregation_rules":                  aggregationRulesObject(config.AggregationRules),
		"default_vertex_labels":              config.DefaultVertexLabels,
		"default_edge_labels":                config.DefaultEdgeLabels,
		"allow_empty_kind_registry":          config.AllowEmptyKindRegistry,
	}
}

func validateClosedObjectV2(object map[string]any, path string, members map[string]memberSpec) error {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		spec, ok := members[key]
		field := canonicalInputMemberPath(path, key)
		if !ok {
			return invalidProjectionErrorV2("unknown_member", field)
		}
		value := object[key]
		if value == nil {
			if !spec.nullable {
				return invalidProjectionErrorV2("explicit_null_not_allowed", field)
			}
			continue
		}
		if !matchesJSONKind(value, spec.kind) {
			return invalidProjectionErrorV2("schema_type_mismatch", field)
		}
	}
	required := make([]string, 0, len(members))
	for key, spec := range members {
		if spec.required {
			required = append(required, key)
		}
	}
	sort.Strings(required)
	for _, key := range required {
		if _, ok := object[key]; !ok {
			return invalidProjectionErrorV2("missing_required_member", canonicalInputMemberPath(path, key))
		}
	}
	return nil
}

func validateObjectArrayV2(ctx context.Context, value any, path string, validator func(map[string]any, string) error) error {
	items, ok := value.([]any)
	if !ok {
		return invalidProjectionErrorV2("schema_type_mismatch", path)
	}
	for index, item := range items {
		if err := itemCheckpointV2(ctx); err != nil {
			return cancelledProjectionErrorV2(err)
		}
		object, ok := item.(map[string]any)
		if !ok {
			return invalidProjectionErrorV2("schema_type_mismatch", fmt.Sprintf("%s[%d]", path, index))
		}
		if err := validator(object, fmt.Sprintf("%s[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateStringArrayV2(value any, path string) error {
	items, ok := value.([]any)
	if !ok {
		return invalidProjectionErrorV2("schema_type_mismatch", path)
	}
	for index, item := range items {
		text, ok := item.(string)
		if !ok || !utf8.ValidString(text) {
			return invalidProjectionErrorV2("schema_type_mismatch", fmt.Sprintf("%s[%d]", path, index))
		}
	}
	return nil
}

func validatePropertyShapeV2(value any, path string) error {
	switch v := value.(type) {
	case map[string]any:
		return invalidProjectionErrorV2("nested_property_value", path)
	case []any:
		for i, entry := range v {
			switch entry.(type) {
			case []any, map[string]any:
				return invalidProjectionErrorV2("nested_property_value", fmt.Sprintf("%s[%d]", path, i))
			}
		}
	}
	return nil
}

func validatePropertyMapV2(object map[string]any, path string) error {
	for _, key := range slices.Sorted(maps.Keys(object)) {
		value := object[key]
		if !validPropertyKey(key) {
			return invalidProjectionErrorV2("invalid_property_key", canonicalInputMemberPath(path, key))
		}
		if err := validateCanonicalJSONValueV2(value, canonicalInputMemberPath(path, key), true); err != nil {
			return err
		}
	}
	return nil
}

func validateCanonicalJSONValueV2(value any, path string, propertyValue bool) error {
	return validateJSONAtV2(value, path, nil, propertyValue)
}

func validateJSONAtV2(value any, path string, segments inputPathV2, propertyValue bool) error {
	failure := func(reason string) error { return invalidProjectionErrorV2(reason, segments.render(path)) }
	switch typed := value.(type) {
	case nil, bool:
		return nil
	case string:
		if !utf8.ValidString(typed) || len(typed) > graphProjectionLimits.MaxStringBytes {
			return failure("invalid_string")
		}
		return nil
	case json.Number:
		if _, err := normalizedNumberV2(typed); err != nil {
			return failure("invalid_number")
		}
		return nil
	case []any:
		for index, entry := range typed {
			if propertyValue {
				switch entry.(type) {
				case map[string]any, []any:
					return invalidProjectionErrorV2("nested_property_value", append(segments, index).render(path))
				}
			}
			if err := validateJSONAtV2(entry, path, append(segments, index), propertyValue); err != nil {
				return err
			}
		}
		return nil
	case map[string]any:
		if propertyValue {
			return failure("nested_property_value")
		}
		for _, key := range slices.Sorted(maps.Keys(typed)) {
			entry := typed[key]
			if err := validateJSONAtV2(entry, path, append(segments, key), false); err != nil {
				return err
			}
		}
		return nil
	default:
		return failure("invalid_json_value")
	}
}

func validateV2ResourceBounds(root map[string]any) error {
	checks := []struct {
		value any
		key   string
		limit int
	}{
		{root["source_entities"], "maximum_source_entities", graphProjectionLimits.MaxSourceEntities},
		{root["source_relationships"], "maximum_source_relationships", graphProjectionLimits.MaxSourceRelationships},
		{root["property_definitions"], "maximum_property_definitions", graphProjectionLimits.MaxPropertyDefinitions},
	}
	config := root["projection_config"].(map[string]any)
	filters := root["filters"].(map[string]any)
	checks = append(checks,
		struct {
			value any
			key   string
			limit int
		}{config["entity_mappings"], "maximum_entity_mappings", graphProjectionLimits.MaxEntityMappings},
		struct {
			value any
			key   string
			limit int
		}{config["relationship_mappings"], "maximum_relationship_mappings", graphProjectionLimits.MaxRelationshipMappings},
		struct {
			value any
			key   string
			limit int
		}{config["metadata_mappings"], "maximum_metadata_mappings", graphProjectionLimits.MaxMetadataMappings},
		struct {
			value any
			key   string
			limit int
		}{config["aggregation_rules"], "maximum_aggregation_rules", graphProjectionLimits.MaxAggregationRules},
		struct {
			value any
			key   string
			limit int
		}{filters["entity_filters"], "maximum_entity_filters", graphProjectionLimits.MaxEntityFilters},
		struct {
			value any
			key   string
			limit int
		}{filters["relationship_filters"], "maximum_relationship_filters", graphProjectionLimits.MaxRelationshipFilters},
	)
	for _, check := range checks {
		if got := len(check.value.([]any)); got > check.limit {
			return resourceProjectionErrorV2(check.key, check.limit, boundedObserved(got, check.limit))
		}
	}
	return nil
}

func normalizeCanonicalValueV2(value any) any {
	switch typed := value.(type) {
	case json.Number:
		normalized, _ := normalizedNumberV2(typed)
		return normalized
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = normalizeCanonicalValueV2(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = normalizeCanonicalValueV2(item)
		}
		return out
	default:
		return typed
	}
}

func normalizedNumberV2(number json.Number) (any, error) {
	lexeme := number.String()
	if !strings.ContainsAny(lexeme, ".eE") {
		value, err := strconv.ParseInt(lexeme, 10, 64)
		if err != nil {
			return nil, err
		}
		return value, nil
	}
	value, err := strconv.ParseFloat(lexeme, 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return nil, errors.New("non-finite number")
	}
	if value == 0 {
		value = 0
	}
	return value, nil
}

func normalizeMapV2(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	return normalizeCanonicalValueV2(input).(map[string]any)
}

func normalizeVerticesV2(input []Vertex) []Vertex {
	out := append([]Vertex(nil), input...)
	for index := range out {
		out[index].Properties = normalizeMapV2(out[index].Properties)
		out[index].Metadata.MappedMetadata = normalizeMapV2(out[index].Metadata.MappedMetadata)
	}
	return out
}

func normalizeEdgesV2(input []Edge) []Edge {
	out := append([]Edge(nil), input...)
	for index := range out {
		out[index].Properties = normalizeMapV2(out[index].Properties)
		out[index].Metadata.MappedMetadata = normalizeMapV2(out[index].Metadata.MappedMetadata)
	}
	return out
}

func validIdentifierV2(value string) bool {
	if !utf8.ValidString(value) || len(value) == 0 || len(value) > graphProjectionLimits.MaxIdentifierBytes || strings.ContainsAny(value, "/\\") {
		return false
	}
	var first, last rune
	for index, character := range value {
		if character == 0 || unicode.IsControl(character) || character >= 0x7f && character <= 0x9f {
			return false
		}
		if index == 0 {
			first = character
		}
		last = character
	}
	return !unicode.IsSpace(first) && !unicode.IsSpace(last)
}

func invalidProjectionErrorV2(reason, field string) *ProjectionErrorV2 {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	return &ProjectionErrorV2{Code: "invalid_projection_request", ReasonCode: reason, RetryAction: "do_not_retry", Details: details}
}

func resourceProjectionErrorV2(key string, limit, observed int) *ProjectionErrorV2 {
	return &ProjectionErrorV2{
		Code:        "projection_resource_limit_exceeded",
		ReasonCode:  key,
		RetryAction: "do_not_retry",
		Details:     map[string]any{"limit_key": key, "limit": limit, "observed": observed},
	}
}

func computationProjectionErrorV2(reason string, cause error) *ProjectionErrorV2 {
	return &ProjectionErrorV2{Code: "projection_computation_failed", ReasonCode: reason, RetryAction: "retry_with_backoff", Details: map[string]any{}, cause: cause}
}

func contextProjectionErrorV2(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return cancelledProjectionErrorV2(err)
	}
	return nil
}

func cancelledProjectionErrorV2(err error) *ProjectionErrorV2 {
	return &ProjectionErrorV2{Code: "projection_cancelled", ReasonCode: "context_cancelled", RetryAction: "do_not_retry", Details: map[string]any{}, cause: err}
}

func firstValidationReasonV2(summary ValidationSummary) string {
	if len(summary.Issues) == 0 {
		return "validation_failed"
	}
	return summary.Issues[0].Code
}

func boundedObserved(observed, limit int) int {
	if observed > limit {
		return limit + 1
	}
	return observed
}
