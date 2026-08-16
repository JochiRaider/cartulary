package graphprojection

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
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
	Properties                    map[string]any
	MappedMetadata                map[string]any
	SchemaRegistry                SchemaRegistry
	Vertices                      []Vertex
	Edges                         []Edge
	ValidationSummary             ValidationSummary
	ConsumerCapabilities          ConsumerCapabilities
}

type EngineV2 struct{}

func NewEngineV2() *EngineV2 { return &EngineV2{} }

func DeriveGraphViewIDV2(sourceOwnerID, graphViewKey string) (string, error) {
	if !validIdentifierV2(sourceOwnerID) || !validIdentifierV2(graphViewKey) {
		return "", errors.New("graphprojection: invalid v2 graph view identity input")
	}
	binding := ResultBindingV2{GraphViewID: graphViewKey, SourceOwnerID: sourceOwnerID}
	fields := []string{"cartulary.graph_projection_graph_view_identity.v2", binding.SourceOwnerID, binding.GraphViewID}
	var transcript bytes.Buffer
	for _, field := range fields {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len([]byte(field))))
		transcript.Write(length[:])
		transcript.WriteString(field)
	}
	return "gv_" + sha256Hex(transcript.Bytes()), nil
}

func (engine *EngineV2) Project(ctx context.Context, invocation InvocationContextV2, semanticInput []byte) (ProjectionResultV2, error) {
	if invocation.CancellationCheck != nil {
		ctx = newProjectionCheckpointContextV2(ctx, invocation.CancellationCheck)
	}
	if err := contextProjectionErrorV2(ctx); err != nil {
		return ProjectionResultV2{}, err
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

	request, normalizedConfiguration, normalizedSource, err := parseProjectionInputV2(semanticInput, invocation)
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
	work := projectionWork{
		Request:                       request,
		GraphViewID:                   invocation.GraphViewID,
		NormalizedConfigurationSHA256: configurationDigest,
		NormalizedSourceSHA256:        sourceDigest,
		IdentityDigest:                workIdentity,
	}
	projected, err := projectSemanticGraph(ctx, work)
	if err != nil {
		if projectionErr := contextProjectionErrorV2(ctx); projectionErr != nil {
			return ProjectionResultV2{}, projectionErr
		}
		return ProjectionResultV2{}, computationProjectionErrorV2("projection_failed", err)
	}
	if projected.ValidationSummary.IssueCount != 0 {
		return ProjectionResultV2{}, &ProjectionErrorV2{
			Code:        "projection_validation_failed",
			ReasonCode:  firstValidationReasonV2(projected.ValidationSummary),
			RetryAction: "do_not_retry",
			Details:     map[string]any{"issue_count": projected.ValidationSummary.IssueCount},
		}
	}
	if len(projected.Vertices) > 100000 {
		return ProjectionResultV2{}, resourceProjectionErrorV2("maximum_vertices", 100000, boundedObserved(len(projected.Vertices), 100000))
	}
	if len(projected.Edges) > 250000 {
		return ProjectionResultV2{}, resourceProjectionErrorV2("maximum_edges", 250000, boundedObserved(len(projected.Edges), 250000))
	}

	result := ProjectionResultV2{
		ProjectionSchemaID:            ProjectionSchemaIDV2,
		GraphViewID:                   invocation.GraphViewID,
		SourceOwnerID:                 invocation.SourceOwnerID,
		SourceSnapshotID:              request.SourceSnapshotID,
		ProjectionVersion:             request.projectionConfig.ProjectionVersion,
		NormalizedConfigurationSHA256: configurationDigest,
		NormalizedSourceSHA256:        sourceDigest,
		Properties:                    normalizeMapV2(projected.Properties),
		MappedMetadata:                normalizeMapV2(projected.MappedMetadata),
		SchemaRegistry:                projected.SchemaRegistry,
		Vertices:                      normalizeVerticesV2(projected.Vertices),
		Edges:                         normalizeEdgesV2(projected.Edges),
		ValidationSummary:             projected.ValidationSummary,
		ConsumerCapabilities:          projected.ConsumerCapabilities,
	}
	outputBytes, err := canonicalJSON(result.outputResource())
	if err != nil {
		return ProjectionResultV2{}, computationProjectionErrorV2("canonical_output_failed", err)
	}
	result.CanonicalOutputSHA256 = sha256Hex(outputBytes)
	result.ProjectionResultID, err = DeriveProjectionResultIDV2(result.ResultBindingV2())
	if err != nil {
		return ProjectionResultV2{}, computationProjectionErrorV2("identity_failed", err)
	}
	if err := contextProjectionErrorV2(ctx); err != nil {
		return ProjectionResultV2{}, err
	}
	return result, nil
}

type projectionCheckpointContextV2 struct {
	context.Context
	check func(context.Context) error
	calls atomic.Uint64
}

func newProjectionCheckpointContextV2(parent context.Context, check func(context.Context) error) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return &projectionCheckpointContextV2{Context: parent, check: check}
}

func (ctx *projectionCheckpointContextV2) Err() error {
	if err := ctx.Context.Err(); err != nil {
		return err
	}
	calls := ctx.calls.Add(1)
	if calls == 1 || calls%1024 == 0 {
		if err := ctx.check(ctx.Context); err != nil {
			return context.Canceled
		}
	}
	return nil
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

func DeriveProjectionResultIDV2(binding ResultBindingV2) (string, error) {
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
	resultJSON, err := canonicalJSON(result.Resource())
	if err != nil {
		return CompletedResultV2{}, err
	}
	vertices := make([]ResultVertexV2, 0, len(result.Vertices))
	for _, vertex := range result.Vertices {
		body, err := canonicalJSON(vertexResource(vertex))
		if err != nil {
			return CompletedResultV2{}, err
		}
		vertices = append(vertices, ResultVertexV2{VertexID: vertex.VertexID, VertexKind: vertex.VertexKind, SortKey: vertex.SortKey, JSON: body})
	}
	edges := make([]ResultEdgeV2, 0, len(result.Edges))
	for _, edge := range result.Edges {
		body, err := canonicalJSON(edgeResource(edge))
		if err != nil {
			return CompletedResultV2{}, err
		}
		edges = append(edges, ResultEdgeV2{EdgeID: edge.EdgeID, EdgeKind: edge.EdgeKind, SrcVertexID: edge.SrcVertexID, DstVertexID: edge.DstVertexID, Direction: edge.Direction, SortKey: edge.SortKey, JSON: body})
	}
	return CompletedResultV2{Binding: result.ResultBindingV2(), ResultJSON: resultJSON, Vertices: vertices, Edges: edges}, nil
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
		"properties":            nonNilMap(result.Properties),
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

func parseProjectionInputV2(data []byte, invocation InvocationContextV2) (projectionRequest, map[string]any, map[string]any, error) {
	if !utf8.Valid(data) {
		return projectionRequest{}, nil, nil, invalidProjectionErrorV2("invalid_utf8", "")
	}
	if err := rejectDuplicateObjectMembers(data); err != nil {
		var duplicate duplicateMemberError
		if errors.As(err, &duplicate) {
			return projectionRequest{}, nil, nil, invalidProjectionErrorV2("duplicate_object_member", duplicate.path)
		}
		return projectionRequest{}, nil, nil, invalidProjectionErrorV2("invalid_json_syntax", "")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return projectionRequest{}, nil, nil, invalidProjectionErrorV2("invalid_json_syntax", "")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return projectionRequest{}, nil, nil, invalidProjectionErrorV2("invalid_json_syntax", "")
	}
	root, ok := decoded.(map[string]any)
	if !ok {
		return projectionRequest{}, nil, nil, invalidProjectionErrorV2("top_level_not_object", "$")
	}
	if err := validateProjectionInputSchemaV2(root); err != nil {
		return projectionRequest{}, nil, nil, err
	}
	if root["projection_schema_id"] != ProjectionSchemaIDV2 {
		return projectionRequest{}, nil, nil, invalidProjectionErrorV2("invalid_projection_schema", "$.projection_schema_id")
	}
	if err := validateV2ResourceBounds(root); err != nil {
		return projectionRequest{}, nil, nil, err
	}

	configRaw := cloneObjectV2(root["projection_config"].(map[string]any))
	configRaw["graph_view_key"] = invocation.GraphViewID
	adaptFilterOperatorsV2(configRaw)
	filtersRaw := cloneObjectV2(root["filters"].(map[string]any))
	adaptFilterOperatorsV2(filtersRaw)
	request := projectionRequest{
		ProjectionSchemaID:   ProjectionSchemaIDV2,
		SourceSnapshotID:     root["source_snapshot_id"].(string),
		projectionConfig:     parseProjectionConfig(configRaw),
		SourceEntities:       parseSourceEntities(root["source_entities"].([]any)),
		SourceRelationships:  parseSourceRelationships(root["source_relationships"].([]any)),
		SourceMetadata:       objectMap(root["source_metadata"].(map[string]any), "$.source_metadata"),
		filters:              parseFilters(filtersRaw),
		PropertyDefinitions:  parsePropertyDefinitions(root["property_definitions"].([]any)),
		RelationshipMappings: nil,
	}
	request.RelationshipMappings = request.projectionConfig.RelationshipMappings
	normalizeProjectionRequestV2(&request)
	configuration := map[string]any{
		"projection_config":    normalizedConfigObjectV2(request.projectionConfig),
		"filters":              filtersObjectV2(request.filters),
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

func normalizeProjectionRequestV2(request *projectionRequest) {
	normalizeProjectionRequest(request)
	sort.SliceStable(request.projectionConfig.EntityMappings, func(i, j int) bool {
		return request.projectionConfig.EntityMappings[i].MappingRuleID < request.projectionConfig.EntityMappings[j].MappingRuleID
	})
	sort.SliceStable(request.projectionConfig.RelationshipMappings, func(i, j int) bool {
		return request.projectionConfig.RelationshipMappings[i].MappingRuleID < request.projectionConfig.RelationshipMappings[j].MappingRuleID
	})
	request.RelationshipMappings = request.projectionConfig.RelationshipMappings
	sort.SliceStable(request.projectionConfig.MetadataMappings, func(i, j int) bool {
		return request.projectionConfig.MetadataMappings[i].MetadataMappingID < request.projectionConfig.MetadataMappings[j].MetadataMappingID
	})
	sort.SliceStable(request.PropertyDefinitions, func(i, j int) bool {
		return request.PropertyDefinitions[i].PropertyDefinitionID < request.PropertyDefinitions[j].PropertyDefinitionID
	})
	request.projectionConfig.AggregationRules = orderAggregationRulesV2(request.projectionConfig.AggregationRules)
}

func orderAggregationRulesV2(rules []aggregationRule) []aggregationRule {
	remaining := make(map[string]aggregationRule, len(rules))
	projectedKindOwners := make(map[string][]string)
	for _, rule := range rules {
		remaining[rule.AggregationRuleID] = rule
		projectedKindOwners[rule.ProjectedKind] = append(projectedKindOwners[rule.ProjectedKind], rule.AggregationRuleID)
	}
	dependencies := make(map[string]map[string]bool, len(rules))
	for _, rule := range rules {
		set := map[string]bool{}
		if rule.InputScope == "projected_vertex" || rule.InputScope == "projected_edge" {
			for _, ownerID := range projectedKindOwners[rule.InputKind] {
				if ownerID != rule.AggregationRuleID {
					set[ownerID] = true
				}
			}
		}
		if rule.endpointGrouping != nil {
			set[rule.endpointGrouping.SourceVertexAggregationRuleID] = true
			set[rule.endpointGrouping.DestinationVertexAggregationRuleID] = true
		}
		dependencies[rule.AggregationRuleID] = set
	}
	ordered := make([]aggregationRule, 0, len(rules))
	completed := map[string]bool{}
	for len(remaining) > 0 {
		ready := make([]string, 0, len(remaining))
		for ruleID := range remaining {
			allReady := true
			for dependencyID := range dependencies[ruleID] {
				if _, exists := remaining[dependencyID]; exists && !completed[dependencyID] {
					allReady = false
					break
				}
			}
			if allReady {
				ready = append(ready, ruleID)
			}
		}
		if len(ready) == 0 {
			for ruleID := range remaining {
				ready = append(ready, ruleID)
			}
		}
		sort.Strings(ready)
		for _, ruleID := range ready {
			ordered = append(ordered, remaining[ruleID])
			delete(remaining, ruleID)
			completed[ruleID] = true
		}
	}
	return ordered
}

func normalizedConfigObjectV2(config projectionConfig) map[string]any {
	return map[string]any{
		"projection_version":                 config.ProjectionVersion,
		"declared_source_entity_kinds":       config.DeclaredSourceEntityKinds,
		"declared_source_relationship_kinds": config.DeclaredSourceRelationshipKinds,
		"entity_mappings":                    entityMappingsObjectV2(config.EntityMappings),
		"relationship_mappings":              relationshipMappingsObjectV2(config.RelationshipMappings),
		"metadata_mappings":                  metadataMappingsObject(config.MetadataMappings),
		"aggregation_rules":                  aggregationRulesObject(config.AggregationRules),
		"default_vertex_labels":              config.DefaultVertexLabels,
		"default_edge_labels":                config.DefaultEdgeLabels,
		"allow_empty_kind_registry":          config.AllowEmptyKindRegistry,
	}
}

func entityMappingsObjectV2(mappings []entityMapping) []any {
	items := entityMappingsObject(mappings)
	for index, item := range items {
		object := canonicalObjectMapV2(item.(canonicalObject))
		adaptFilterOperatorsResourceV2(object)
		items[index] = object
	}
	return items
}

func relationshipMappingsObjectV2(mappings []relationshipMapping) []any {
	items := relationshipMappingsObject(mappings)
	for index, item := range items {
		object := canonicalObjectMapV2(item.(canonicalObject))
		if !mappings[index].ReverseEdgeKindSupplied {
			delete(object, "reverse_edge_kind")
		}
		adaptFilterOperatorsResourceV2(object)
		items[index] = object
	}
	return items
}

func filtersObjectV2(filters filters) map[string]any {
	object := canonicalObjectMapV2(filtersObject(filters))
	adaptFilterOperatorsResourceV2(object)
	return object
}

func canonicalObjectMapV2(input canonicalObject) map[string]any {
	output := make(map[string]any, len(input))
	for _, member := range input {
		output[member.Name] = canonicalObjectValueV2(member.Value)
	}
	return output
}

func canonicalObjectValueV2(input any) any {
	switch typed := input.(type) {
	case canonicalObject:
		return canonicalObjectMapV2(typed)
	case []any:
		output := make([]any, len(typed))
		for index, item := range typed {
			output[index] = canonicalObjectValueV2(item)
		}
		return output
	case map[string]any:
		output := make(map[string]any, len(typed))
		for key, item := range typed {
			output[key] = canonicalObjectValueV2(item)
		}
		return output
	default:
		return typed
	}
}

func adaptFilterOperatorsResourceV2(object map[string]any) {
	if op, ok := object["op"]; ok {
		delete(object, "op")
		object["operator"] = op
	}
	for _, key := range []string{"entity_filters", "relationship_filters"} {
		if items, ok := object[key].([]any); ok {
			for _, item := range items {
				adaptFilterOperatorsResourceV2(item.(map[string]any))
			}
		}
	}
	if predicate, ok := object["inclusion_predicate"].(map[string]any); ok {
		adaptFilterOperatorsResourceV2(predicate)
	}
}

func adaptFilterOperatorsV2(object map[string]any) {
	if operator, ok := object["operator"]; ok {
		delete(object, "operator")
		object["op"] = operator
	}
	for _, value := range object {
		switch typed := value.(type) {
		case map[string]any:
			adaptFilterOperatorsV2(typed)
		case []any:
			for _, item := range typed {
				if child, ok := item.(map[string]any); ok {
					adaptFilterOperatorsV2(child)
				}
			}
		}
	}
}

func cloneObjectV2(value map[string]any) map[string]any {
	cloned := make(map[string]any, len(value))
	for key, item := range value {
		switch typed := item.(type) {
		case map[string]any:
			cloned[key] = cloneObjectV2(typed)
		case []any:
			array := make([]any, len(typed))
			for index, entry := range typed {
				if object, ok := entry.(map[string]any); ok {
					array[index] = cloneObjectV2(object)
				} else {
					array[index] = entry
				}
			}
			cloned[key] = array
		default:
			cloned[key] = item
		}
	}
	return cloned
}

func validateProjectionInputSchemaV2(root map[string]any) error {
	top := map[string]memberSpec{
		"projection_schema_id": {kind: kindString, required: true},
		"source_snapshot_id":   {kind: kindString, required: true},
		"projection_config":    {kind: kindObject, required: true},
		"source_entities":      {kind: kindArray, required: true},
		"source_relationships": {kind: kindArray, required: true},
		"source_metadata":      {kind: kindObject, required: true},
		"filters":              {kind: kindObject, required: true},
		"property_definitions": {kind: kindArray, required: true},
	}
	if err := validateClosedObjectV2(root, "$", top); err != nil {
		return err
	}
	if err := validateProjectionConfigSchemaV2(root["projection_config"].(map[string]any), "$.projection_config"); err != nil {
		return err
	}
	if err := validateObjectArrayV2(root["source_entities"], "$.source_entities", validateSourceEntitySchemaV2); err != nil {
		return err
	}
	if err := validateObjectArrayV2(root["source_relationships"], "$.source_relationships", validateSourceRelationshipSchemaV2); err != nil {
		return err
	}
	if err := validateFiltersSchemaV2(root["filters"].(map[string]any), "$.filters"); err != nil {
		return err
	}
	if err := validateObjectArrayV2(root["property_definitions"], "$.property_definitions", validatePropertyDefinitionSchemaV2); err != nil {
		return err
	}
	if !validIdentifierV2(root["source_snapshot_id"].(string)) {
		return invalidProjectionErrorV2("invalid_identifier", "$.source_snapshot_id")
	}
	return validateCanonicalJSONValueV2(root["source_metadata"], "$.source_metadata", false)
}

func validateProjectionConfigSchemaV2(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"projection_version":                 {kind: kindString, required: true},
		"declared_source_entity_kinds":       {kind: kindArray, required: true},
		"declared_source_relationship_kinds": {kind: kindArray, required: true},
		"entity_mappings":                    {kind: kindArray, required: true},
		"relationship_mappings":              {kind: kindArray, required: true},
		"metadata_mappings":                  {kind: kindArray, required: true},
		"aggregation_rules":                  {kind: kindArray, required: true},
		"default_vertex_labels":              {kind: kindArray, required: true},
		"default_edge_labels":                {kind: kindArray, required: true},
		"allow_empty_kind_registry":          {kind: kindBoolean, required: true},
	}
	if err := validateClosedObjectV2(object, path, members); err != nil {
		return err
	}
	for _, key := range []string{"declared_source_entity_kinds", "declared_source_relationship_kinds", "default_vertex_labels", "default_edge_labels"} {
		if err := validateStringArrayV2(object[key], canonicalInputMemberPath(path, key)); err != nil {
			return err
		}
	}
	if err := validateObjectArrayV2(object["entity_mappings"], path+".entity_mappings", validateEntityMappingSchemaV2); err != nil {
		return err
	}
	if err := validateObjectArrayV2(object["relationship_mappings"], path+".relationship_mappings", validateRelationshipMappingSchemaV2); err != nil {
		return err
	}
	if err := validateObjectArrayV2(object["metadata_mappings"], path+".metadata_mappings", validateMetadataMappingSchemaV2); err != nil {
		return err
	}
	if err := validateObjectArrayV2(object["aggregation_rules"], path+".aggregation_rules", validateAggregationRuleSchemaV2); err != nil {
		return err
	}
	if !validIdentifierV2(object["projection_version"].(string)) {
		return invalidProjectionErrorV2("invalid_identifier", path+".projection_version")
	}
	return nil
}

func validateEntityMappingSchemaV2(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"mapping_rule_id": {kind: kindString, required: true}, "source_entity_kind": {kind: kindString, required: true},
		"projected_vertex_kind": {kind: kindString, required: true}, "inclusion_predicate": {kind: kindAny, required: true},
		"label_policy": {kind: kindString, required: true}, "mapping_labels": {kind: kindArray, required: true},
		"required_property_keys": {kind: kindArray, required: true}, "optional_property_keys": {kind: kindArray, required: true},
	}
	if err := validateClosedObjectV2(object, path, members); err != nil {
		return err
	}
	return validateMappingV2(object, path)
}

func validateRelationshipMappingSchemaV2(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"mapping_rule_id": {kind: kindString, required: true}, "source_relationship_kind": {kind: kindString, required: true},
		"projected_edge_kind": {kind: kindString, required: true}, "inclusion_predicate": {kind: kindAny, required: true},
		"direction_policy": {kind: kindString, required: true}, "emit_reverse_edge": {kind: kindBoolean, required: true}, "reverse_edge_kind": {kind: kindString},
		"label_policy": {kind: kindString, required: true}, "mapping_labels": {kind: kindArray, required: true},
		"required_property_keys": {kind: kindArray, required: true}, "optional_property_keys": {kind: kindArray, required: true},
	}
	if err := validateClosedObjectV2(object, path, members); err != nil {
		return err
	}
	emitReverse := object["emit_reverse_edge"].(bool)
	_, reversePresent := object["reverse_edge_kind"]
	if emitReverse != reversePresent {
		return invalidProjectionErrorV2("reverse_edge_kind_presence", path+".reverse_edge_kind")
	}
	return validateMappingV2(object, path)
}

func validateMappingV2(object map[string]any, path string) error {
	for _, key := range []string{"mapping_labels", "required_property_keys", "optional_property_keys"} {
		if err := validateStringArrayV2(object[key], path+"."+key); err != nil {
			return err
		}
	}
	switch predicate := object["inclusion_predicate"].(type) {
	case string:
		if predicate != "always" {
			return invalidProjectionErrorV2("invalid_inclusion_predicate", path+".inclusion_predicate")
		}
	case map[string]any:
		return validateFilterPredicateSchemaV2(predicate, path+".inclusion_predicate")
	default:
		return invalidProjectionErrorV2("schema_type_mismatch", path+".inclusion_predicate")
	}
	return nil
}

func validatePropertyDefinitionSchemaV2(object map[string]any, path string) error {
	return validateDefinitionSchemaV2(object, path, "property_definition_id", "projected_key")
}

func validateMetadataMappingSchemaV2(object map[string]any, path string) error {
	return validateDefinitionSchemaV2(object, path, "metadata_mapping_id", "projected_metadata_key")
}

func validateDefinitionSchemaV2(object map[string]any, path, idKey, outputKey string) error {
	members := definitionMembers(idKey, outputKey)
	for key, spec := range members {
		if key != "default_value" {
			spec.required = true
			members[key] = spec
		}
	}
	return validateClosedObjectV2(object, path, members)
}

func validateAggregationRuleSchemaV2(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"aggregation_rule_id": {kind: kindString, required: true}, "target_scope": {kind: kindString, required: true},
		"input_scope": {kind: kindString, required: true}, "input_kind": {kind: kindString, required: true}, "projected_kind": {kind: kindString, required: true},
		"grouping_keys": {kind: kindArray, required: true}, "missing_grouping_key_behavior": {kind: kindString, required: true},
		"property_merge_behavior": {kind: kindObject, required: true}, "edge_direction": {kind: kindString}, "endpoint_grouping": {kind: kindObject},
	}
	if err := validateClosedObjectV2(object, path, members); err != nil {
		return err
	}
	if err := validateStringArrayV2(object["grouping_keys"], path+".grouping_keys"); err != nil {
		return err
	}
	targetScope := object["target_scope"].(string)
	_, hasDirection := object["edge_direction"]
	_, hasEndpoints := object["endpoint_grouping"]
	if targetScope == "edge" {
		if !hasDirection || !hasEndpoints {
			return invalidProjectionErrorV2("missing_edge_aggregation_member", path)
		}
		endpoint := object["endpoint_grouping"].(map[string]any)
		endpointMembers := map[string]memberSpec{
			"src_vertex_aggregation_rule_id": {kind: kindString, required: true}, "src_grouping_keys": {kind: kindArray, required: true},
			"dst_vertex_aggregation_rule_id": {kind: kindString, required: true}, "dst_grouping_keys": {kind: kindArray, required: true},
			"missing_endpoint_behavior": {kind: kindString, required: true},
		}
		if err := validateClosedObjectV2(endpoint, path+".endpoint_grouping", endpointMembers); err != nil {
			return err
		}
	} else if hasDirection || hasEndpoints {
		return invalidProjectionErrorV2("vertex_aggregation_edge_member", path)
	}
	return nil
}

func validateSourceEntitySchemaV2(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"source_entity_id": {kind: kindString, required: true}, "source_entity_kind": {kind: kindString, required: true},
		"properties": {kind: kindObject, required: true}, "metadata": {kind: kindObject, required: true}, "labels": {kind: kindArray, required: true},
	}
	if err := validateClosedObjectV2(object, path, members); err != nil {
		return err
	}
	if err := validateStringArrayV2(object["labels"], path+".labels"); err != nil {
		return err
	}
	if err := validatePropertyMapV2(object["properties"].(map[string]any), path+".properties"); err != nil {
		return err
	}
	return validateCanonicalJSONValueV2(object["metadata"], path+".metadata", false)
}

func validateSourceRelationshipSchemaV2(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"source_relationship_id": {kind: kindString, required: true}, "source_relationship_kind": {kind: kindString, required: true},
		"src_source_entity_id": {kind: kindString, required: true}, "dst_source_entity_id": {kind: kindString, required: true}, "direction": {kind: kindString, required: true},
		"properties": {kind: kindObject, required: true}, "metadata": {kind: kindObject, required: true}, "labels": {kind: kindArray, required: true},
	}
	if err := validateClosedObjectV2(object, path, members); err != nil {
		return err
	}
	if err := validateStringArrayV2(object["labels"], path+".labels"); err != nil {
		return err
	}
	if err := validatePropertyMapV2(object["properties"].(map[string]any), path+".properties"); err != nil {
		return err
	}
	return validateCanonicalJSONValueV2(object["metadata"], path+".metadata", false)
}

func validateFiltersSchemaV2(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"entity_filters": {kind: kindArray, required: true}, "relationship_filters": {kind: kindArray, required: true}, "logic": {kind: kindString, required: true},
	}
	if err := validateClosedObjectV2(object, path, members); err != nil {
		return err
	}
	for _, key := range []string{"entity_filters", "relationship_filters"} {
		if err := validateObjectArrayV2(object[key], path+"."+key, validateFilterPredicateSchemaV2); err != nil {
			return err
		}
	}
	return nil
}

func validateFilterPredicateSchemaV2(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"field_path": {kind: kindString, required: true}, "operator": {kind: kindString, required: true},
		"value": {kind: kindAny, nullable: true}, "include_if_missing": {kind: kindBoolean, required: true},
	}
	if err := validateClosedObjectV2(object, path, members); err != nil {
		return err
	}
	operator := object["operator"].(string)
	_, hasValue := object["value"]
	if operator == "exists" && hasValue || operator != "exists" && !hasValue {
		return invalidProjectionErrorV2("filter_value_presence", path+".value")
	}
	if hasValue {
		return validateCanonicalJSONValueV2(object["value"], path+".value", true)
	}
	return nil
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

func validateObjectArrayV2(value any, path string, validator func(map[string]any, string) error) error {
	items, ok := value.([]any)
	if !ok {
		return invalidProjectionErrorV2("schema_type_mismatch", path)
	}
	for index, item := range items {
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

func validatePropertyMapV2(object map[string]any, path string) error {
	for key, value := range object {
		if !validIdentifierV2(key) {
			return invalidProjectionErrorV2("invalid_property_key", canonicalInputMemberPath(path, key))
		}
		if err := validateCanonicalJSONValueV2(value, canonicalInputMemberPath(path, key), true); err != nil {
			return err
		}
	}
	return nil
}

func validateCanonicalJSONValueV2(value any, path string, propertyValue bool) error {
	switch typed := value.(type) {
	case nil, bool, string:
		return nil
	case json.Number:
		if _, err := normalizedNumberV2(typed); err != nil {
			return invalidProjectionErrorV2("invalid_number", path)
		}
		return nil
	case []any:
		for index, entry := range typed {
			if propertyValue {
				switch entry.(type) {
				case map[string]any, []any:
					return invalidProjectionErrorV2("nested_property_value", fmt.Sprintf("%s[%d]", path, index))
				}
			}
			if err := validateCanonicalJSONValueV2(entry, fmt.Sprintf("%s[%d]", path, index), propertyValue); err != nil {
				return err
			}
		}
		return nil
	case map[string]any:
		if propertyValue {
			return invalidProjectionErrorV2("nested_property_value", path)
		}
		for key, entry := range typed {
			if err := validateCanonicalJSONValueV2(entry, canonicalInputMemberPath(path, key), false); err != nil {
				return err
			}
		}
		return nil
	default:
		return invalidProjectionErrorV2("invalid_json_value", path)
	}
}

func validateV2ResourceBounds(root map[string]any) error {
	checks := []struct {
		value any
		key   string
		limit int
	}{
		{root["source_entities"], "maximum_source_entities", 100000},
		{root["source_relationships"], "maximum_source_relationships", 250000},
		{root["property_definitions"], "maximum_property_definitions", 10000},
	}
	config := root["projection_config"].(map[string]any)
	filters := root["filters"].(map[string]any)
	checks = append(checks,
		struct {
			value any
			key   string
			limit int
		}{config["entity_mappings"], "maximum_entity_mappings", 10000},
		struct {
			value any
			key   string
			limit int
		}{config["relationship_mappings"], "maximum_relationship_mappings", 10000},
		struct {
			value any
			key   string
			limit int
		}{config["metadata_mappings"], "maximum_metadata_mappings", 10000},
		struct {
			value any
			key   string
			limit int
		}{config["aggregation_rules"], "maximum_aggregation_rules", 1000},
		struct {
			value any
			key   string
			limit int
		}{filters["entity_filters"], "maximum_entity_filters", 1000},
		struct {
			value any
			key   string
			limit int
		}{filters["relationship_filters"], "maximum_relationship_filters", 1000},
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
	case canonicalObject:
		return normalizeCanonicalValueV2(canonicalObjectMapV2(typed))
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
	if !utf8.ValidString(value) || len([]byte(value)) == 0 || len([]byte(value)) > 255 || strings.ContainsAny(value, "/\\") {
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
		return &ProjectionErrorV2{Code: "projection_cancelled", ReasonCode: "context_cancelled", RetryAction: "do_not_retry", Details: map[string]any{}, cause: err}
	}
	return nil
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
