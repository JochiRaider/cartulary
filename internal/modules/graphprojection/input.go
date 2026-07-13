package graphprojection

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	finiteIntegerPattern = regexp.MustCompile(`^(0|-?[1-9][0-9]*)$`)
	generatedIDPattern   = regexp.MustCompile(`^(gv|gpr|vx|ed|gpi)_[0-9a-f]{64}$`)
	timestampPattern     = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,6})?Z$`)
)

const maxFiniteInteger = int64(9007199254740991)

const (
	defaultRetentionCount              = 5
	defaultRetentionDurationSeconds    = 2592000
	defaultFailedRetentionCount        = 20
	defaultFailedRetentionDurationSecs = 2592000
)

type admitOptions struct {
	ProjectionRunNonce string
	AcceptedAt         time.Time
	InvocationIDPrefix string
	InvocationDomain   string
	Operation          string
}

type projectionInputSource interface {
	Len() int
	Bytes() []byte
}

type byteProjectionInput []byte

func (input byteProjectionInput) Len() int      { return len(input) }
func (input byteProjectionInput) Bytes() []byte { return []byte(input) }

func admitProjectionInput(data []byte, options admitOptions) (run ProjectionRun, err error) {
	return admitProjectionInputSource(byteProjectionInput(data), options)
}

func admitProjectionInputSource(source projectionInputSource, options admitOptions) (run ProjectionRun, err error) {
	operation := options.Operation
	if operation == "" {
		operation = "create_projection"
	}
	defer func() { err = annotateAdmissionError(err, operation) }()
	if source.Len() > graphProjectionLimits.MaxInputBytes {
		return ProjectionRun{}, schemaError("whole_input_limit_exceeded", "$", "resource_limit_exceeded")
	}
	data := source.Bytes()
	request, normalizedRaw, err := parseProjectionInput(data)
	if err != nil {
		return ProjectionRun{}, err
	}
	if request.ProjectionSchemaID != ProjectionSchemaID {
		return ProjectionRun{}, invalidRequest("invalid_projection_schema", "$.projection_schema_id", map[string]any{"validation_code": "invalid_projection_schema"})
	}
	expectedGraphViewID, err := deriveGraphViewID(request.ProjectionConfig.GraphViewKey)
	if err != nil {
		return ProjectionRun{}, invalidRequest("normalization_failed", "$.projection_config.graph_view_key", nil)
	}
	if request.GraphViewID != expectedGraphViewID {
		return ProjectionRun{}, invalidRequest("invalid_graph_view_id", "$.graph_view_id", map[string]any{"validation_code": "invalid_graph_view_id"})
	}
	configDigest, err := projectionConfigDigest(request)
	if err != nil {
		return ProjectionRun{}, invalidRequest("normalization_failed", "", nil)
	}
	sourceDigest, err := projectionSourceDigest(request)
	if err != nil {
		return ProjectionRun{}, invalidRequest("normalization_failed", "", nil)
	}
	nonce := strings.TrimSpace(options.ProjectionRunNonce)
	if nonce == "" {
		return ProjectionRun{}, fmt.Errorf("graphprojection: projection nonce is required")
	}
	idPrefix := options.InvocationIDPrefix
	if idPrefix == "" {
		idPrefix = "gpr_"
	}
	domain := options.InvocationDomain
	if domain == "" {
		domain = "GPRUN1\n"
	}
	runID, err := generatedID(idPrefix, domain, "projection_run", ProjectionSchemaID, expectedGraphViewID, request.SourceSnapshotID, configDigest, sourceDigest, nonce)
	if err != nil {
		return ProjectionRun{}, invalidRequest("normalization_failed", "", nil)
	}
	acceptedAt := options.AcceptedAt.UTC()
	if acceptedAt.IsZero() {
		acceptedAt = time.Now().UTC()
	}
	request.Normalized = normalizedRaw
	return ProjectionRun{
		Request:                request,
		GraphViewID:            expectedGraphViewID,
		ProjectionRunID:        runID,
		ProjectionRunNonce:     nonce,
		ProjectionConfigDigest: configDigest,
		ProjectionSourceDigest: sourceDigest,
		AcceptedAt:             acceptedAt,
		State:                  RunStateAccepted,
	}, nil
}

func deriveGraphViewID(graphViewKey string) (string, error) {
	return generatedID("gv_", "GPID1\n", "graph_view", ProjectionSchemaID, graphViewKey)
}

func invalidRequest(reasonCode, field string, details map[string]any) *LifecycleError {
	if details == nil {
		details = map[string]any{}
	}
	details["reason_code"] = reasonCode
	if field != "" {
		details["field"] = field
	}
	return &LifecycleError{Code: "invalid_projection_request", ReasonCode: reasonCode, Field: field, Details: details}
}

func parseProjectionInput(data []byte) (ProjectionRequest, map[string]any, error) {
	if !utf8.Valid(data) {
		return ProjectionRequest{}, nil, invalidRequest("invalid_utf8", "", nil)
	}
	if err := rejectDuplicateObjectMembers(data); err != nil {
		var duplicate duplicateMemberError
		if errors.As(err, &duplicate) {
			return ProjectionRequest{}, nil, invalidRequest("duplicate_object_member", duplicate.path, nil)
		}
		return ProjectionRequest{}, nil, invalidRequest("invalid_json_syntax", "", nil)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return ProjectionRequest{}, nil, invalidRequest("invalid_json_syntax", "", nil)
	}
	root, ok := raw.(map[string]any)
	if !ok {
		return ProjectionRequest{}, nil, invalidRequest("top_level_not_object", "$", nil)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ProjectionRequest{}, nil, invalidRequest("invalid_json_syntax", "", nil)
	}
	if err := validateProjectionInputSchema(root); err != nil {
		return ProjectionRequest{}, nil, err
	}
	request := ProjectionRequest{
		ProjectionSchemaID:   mustString(root["projection_schema_id"], "$.projection_schema_id"),
		GraphViewID:          mustString(root["graph_view_id"], "$.graph_view_id"),
		SourceSnapshotID:     mustString(root["source_snapshot_id"], "$.source_snapshot_id"),
		SourceMetadata:       objectMap(defaultObject(root["source_metadata"]), "$.source_metadata"),
		RequestedAt:          mustString(root["requested_at"], "$.requested_at"),
		RequestedBy:          mustString(root["requested_by"], "$.requested_by"),
		Filters:              parseFilters(defaultObject(root["filters"])),
		RelationshipMappings: parseRelationshipMappings(arrayValue(root["relationship_definitions"], "$.relationship_definitions")),
		PropertyDefinitions:  parsePropertyDefinitions(arrayValue(root["property_definitions"], "$.property_definitions")),
	}
	if !generatedIDPattern.MatchString(request.GraphViewID) || !strings.HasPrefix(request.GraphViewID, "gv_") {
		return ProjectionRequest{}, nil, invalidRequest("invalid_graph_view_id", "$.graph_view_id", nil)
	}
	request.ProjectionConfig = parseProjectionConfig(mustObject(root["projection_config"], "$.projection_config"))
	request.RelationshipMappingSourceConflict = len(request.RelationshipMappings) > 0 && len(request.ProjectionConfig.RelationshipMappings) > 0
	request.SourceEntities = parseSourceEntities(arrayValue(root["source_entities"], "$.source_entities"))
	request.SourceRelationships = parseSourceRelationships(arrayValue(root["source_relationships"], "$.source_relationships"))
	normalizeProjectionRequest(&request)
	normalizedRaw := normalizedRequestObject(request)
	return request, normalizedRaw, nil
}

type duplicateMemberError struct{ path string }

func (err duplicateMemberError) Error() string { return err.path }

func rejectDuplicateObjectMembers(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var parseValue func(path string) error
	parseValue = func(path string) error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		if delimiter, ok := token.(json.Delim); ok {
			switch delimiter {
			case '{':
				seen := map[string]struct{}{}
				for decoder.More() {
					keyToken, err := decoder.Token()
					if err != nil {
						return err
					}
					key, ok := keyToken.(string)
					if !ok {
						return fmt.Errorf("object member at %s is not a string", path)
					}
					if _, exists := seen[key]; exists {
						return duplicateMemberError{path: canonicalInputMemberPath(path, key)}
					}
					seen[key] = struct{}{}
					childPath := canonicalInputMemberPath(path, key)
					if err := parseValue(childPath); err != nil {
						return err
					}
				}
				_, err := decoder.Token()
				return err
			case '[':
				index := 0
				for decoder.More() {
					if err := parseValue(fmt.Sprintf("%s[%d]", path, index)); err != nil {
						return err
					}
					index++
				}
				_, err := decoder.Token()
				return err
			default:
				return fmt.Errorf("unexpected delimiter at %s", path)
			}
		}
		return nil
	}
	if err := parseValue("$"); err != nil {
		return err
	}
	if decoder.More() {
		return fmt.Errorf("trailing input")
	}
	return nil
}

func parseProjectionConfig(raw map[string]any) ProjectionConfig {
	config := ProjectionConfig{
		GraphViewKey:                    mustString(raw["graph_view_key"], "$.projection_config.graph_view_key"),
		ProjectionVersion:               stringDefault(raw["projection_version"], "1"),
		DeclaredSourceEntityKinds:       stringArray(raw["declared_source_entity_kinds"]),
		DeclaredSourceRelationshipKinds: stringArray(raw["declared_source_relationship_kinds"]),
		EntityMappings:                  parseEntityMappings(arrayValue(raw["entity_mappings"], "$.projection_config.entity_mappings")),
		RelationshipMappings:            parseRelationshipMappings(arrayValue(raw["relationship_mappings"], "$.projection_config.relationship_mappings")),
		MetadataMappings:                parseMetadataMappings(arrayValue(raw["metadata_mappings"], "$.projection_config.metadata_mappings")),
		AggregationRules:                parseAggregationRules(arrayValue(raw["aggregation_rules"], "$.projection_config.aggregation_rules")),
		DefaultVertexLabels:             stringArray(raw["default_vertex_labels"]),
		DefaultEdgeLabels:               stringArray(raw["default_edge_labels"]),
		AllowEmptyKindRegistry:          boolDefault(raw["allow_empty_kind_registry"], false),
		RetentionPolicy:                 parseRetentionPolicy(defaultObject(raw["retention_policy"])),
		CustomConfig:                    objectMap(defaultObject(raw["custom_config"]), "$.projection_config.custom_config"),
	}
	return config
}

func parseRetentionPolicy(raw map[string]any) RetentionPolicy {
	policy := RetentionPolicy{
		RetainReplacedResults:       boolDefault(raw["retain_replaced_results"], true),
		RetentionCount:              intDefault(raw["retention_count"], defaultRetentionCount),
		RetentionDurationSeconds:    intDefault(raw["retention_duration_seconds"], defaultRetentionDurationSeconds),
		RetainFailedResults:         boolDefault(raw["retain_failed_results"], true),
		FailedRetentionCount:        intDefault(raw["failed_retention_count"], defaultFailedRetentionCount),
		FailedRetentionDurationSecs: intDefault(raw["failed_retention_duration_seconds"], defaultFailedRetentionDurationSecs),
		RawIntegerLexemes:           map[string]string{},
	}
	for _, key := range []string{"retention_count", "retention_duration_seconds", "failed_retention_count", "failed_retention_duration_seconds"} {
		if number, ok := raw[key].(json.Number); ok {
			policy.RawIntegerLexemes[key] = number.String()
		}
	}
	return policy
}

func parseEntityMappings(raw []any) []EntityMapping {
	mappings := make([]EntityMapping, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "entity_mapping")
		mapping := EntityMapping{
			MappingRuleID:        mustString(object["mapping_rule_id"], "mapping_rule_id"),
			SourceEntityKind:     mustString(object["source_entity_kind"], "source_entity_kind"),
			ProjectedVertexKind:  mustString(object["projected_vertex_kind"], "projected_vertex_kind"),
			InclusionPredicate:   stringDefault(object["inclusion_predicate"], "always"),
			LabelPolicy:          stringDefault(object["label_policy"], "mapping_only"),
			MappingLabels:        uniqueSortedStrings(stringArray(object["mapping_labels"])),
			RequiredPropertyKeys: uniqueSortedStrings(stringArray(object["required_property_keys"])),
			OptionalPropertyKeys: uniqueSortedStrings(stringArray(object["optional_property_keys"])),
		}
		mapping.InclusionPredicate, mapping.InclusionFilter = parseInclusionPredicate(object["inclusion_predicate"])
		mapping.MappingIdentityDigest, _ = digestTuple("GPMAPENTITY1\n", mapping.MappingRuleID, mapping.SourceEntityKind, mapping.ProjectedVertexKind)
		mappings = append(mappings, mapping)
	}
	sort.Slice(mappings, func(i, j int) bool { return mappings[i].MappingRuleID < mappings[j].MappingRuleID })
	return mappings
}

func parseRelationshipMappings(raw []any) []RelationshipMapping {
	mappings := make([]RelationshipMapping, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "relationship_mapping")
		projectedKind := mustString(object["projected_edge_kind"], "projected_edge_kind")
		mapping := RelationshipMapping{
			MappingRuleID:          mustString(object["mapping_rule_id"], "mapping_rule_id"),
			SourceRelationshipKind: mustString(object["source_relationship_kind"], "source_relationship_kind"),
			ProjectedEdgeKind:      projectedKind,
			InclusionPredicate:     stringDefault(object["inclusion_predicate"], "always"),
			DirectionPolicy:        stringDefault(object["direction_policy"], "preserve"),
			EmitReverseEdge:        boolDefault(object["emit_reverse_edge"], false),
			ReverseEdgeKind:        stringDefault(object["reverse_edge_kind"], projectedKind),
			LabelPolicy:            stringDefault(object["label_policy"], "mapping_only"),
			MappingLabels:          uniqueSortedStrings(stringArray(object["mapping_labels"])),
			RequiredPropertyKeys:   uniqueSortedStrings(stringArray(object["required_property_keys"])),
			OptionalPropertyKeys:   uniqueSortedStrings(stringArray(object["optional_property_keys"])),
		}
		_, mapping.ReverseEdgeKindSupplied = object["reverse_edge_kind"]
		mapping.InclusionPredicate, mapping.InclusionFilter = parseInclusionPredicate(object["inclusion_predicate"])
		mapping.MappingIdentityDigest, _ = digestTuple("GPMAPREL1\n", mapping.MappingRuleID, mapping.SourceRelationshipKind, mapping.ProjectedEdgeKind, mapping.DirectionPolicy, mapping.EmitReverseEdge, mapping.ReverseEdgeKind)
		mappings = append(mappings, mapping)
	}
	sort.Slice(mappings, func(i, j int) bool { return mappings[i].MappingRuleID < mappings[j].MappingRuleID })
	return mappings
}

func parsePropertyDefinitions(raw []any) []PropertyDefinition {
	definitions := make([]PropertyDefinition, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "property_definition")
		required := boolDefault(object["required"], false)
		definition := PropertyDefinition{
			PropertyDefinitionID: mustString(object["property_definition_id"], "property_definition_id"),
			TargetScope:          mustString(object["target_scope"], "target_scope"),
			TargetKind:           mustString(object["target_kind"], "target_kind"),
			SourceFieldPath:      mustString(object["source_field_path"], "source_field_path"),
			ProjectedKey:         mustString(object["projected_key"], "projected_key"),
			ProjectedType:        mustString(object["projected_type"], "projected_type"),
			Required:             required,
			MissingBehavior:      stringDefault(object["missing_behavior"], defaultMissingBehavior(required)),
			SourceNullBehavior:   stringDefault(object["source_null_behavior"], defaultMissingBehavior(required)),
			NullOutputPolicy:     stringDefault(object["null_output_policy"], "omit"),
			MergeBehavior:        stringDefault(object["merge_behavior"], "single_value"),
		}
		if value, ok := object["default_value"]; ok {
			definition.DefaultValue = value
			definition.HasDefaultValue = true
		}
		definitions = append(definitions, definition)
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].PropertyDefinitionID < definitions[j].PropertyDefinitionID })
	return definitions
}

func parseMetadataMappings(raw []any) []MetadataMapping {
	mappings := make([]MetadataMapping, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "metadata_mapping")
		required := boolDefault(object["required"], false)
		mapping := MetadataMapping{
			MetadataMappingID:    mustString(object["metadata_mapping_id"], "metadata_mapping_id"),
			TargetScope:          mustString(object["target_scope"], "target_scope"),
			TargetKind:           mustString(object["target_kind"], "target_kind"),
			SourceFieldPath:      mustString(object["source_field_path"], "source_field_path"),
			ProjectedMetadataKey: mustString(object["projected_metadata_key"], "projected_metadata_key"),
			ProjectedType:        mustString(object["projected_type"], "projected_type"),
			Required:             required,
			MissingBehavior:      stringDefault(object["missing_behavior"], defaultMissingBehavior(required)),
			SourceNullBehavior:   stringDefault(object["source_null_behavior"], defaultMissingBehavior(required)),
			NullOutputPolicy:     stringDefault(object["null_output_policy"], "omit"),
			MergeBehavior:        stringDefault(object["merge_behavior"], "single_value"),
		}
		if value, ok := object["default_value"]; ok {
			mapping.DefaultValue = value
			mapping.HasDefaultValue = true
		}
		mappings = append(mappings, mapping)
	}
	sort.Slice(mappings, func(i, j int) bool { return mappings[i].MetadataMappingID < mappings[j].MetadataMappingID })
	return mappings
}

func parseAggregationRules(raw []any) []AggregationRule {
	rules := make([]AggregationRule, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "aggregation_rule")
		rule := AggregationRule{
			AggregationRuleID:          mustString(object["aggregation_rule_id"], "aggregation_rule_id"),
			TargetScope:                mustString(object["target_scope"], "target_scope"),
			InputScope:                 mustString(object["input_scope"], "input_scope"),
			InputKind:                  mustString(object["input_kind"], "input_kind"),
			ProjectedKind:              mustString(object["projected_kind"], "projected_kind"),
			GroupingKeys:               stringArray(object["grouping_keys"]),
			MissingGroupingKeyBehavior: stringDefault(object["missing_grouping_key_behavior"], "error"),
			PropertyMergeBehavior:      stringMap(defaultObject(object["property_merge_behavior"])),
			EdgeDirection:              stringDefault(object["edge_direction"], "directed"),
		}
		if endpointRaw, ok := object["endpoint_grouping"]; ok && endpointRaw != nil {
			endpoint := mustObject(endpointRaw, "endpoint_grouping")
			rule.EndpointGrouping = &EndpointGrouping{
				SourceVertexAggregationRuleID:      mustString(endpoint["src_vertex_aggregation_rule_id"], "src_vertex_aggregation_rule_id"),
				SourceGroupingKeys:                 stringArray(endpoint["src_grouping_keys"]),
				DestinationVertexAggregationRuleID: mustString(endpoint["dst_vertex_aggregation_rule_id"], "dst_vertex_aggregation_rule_id"),
				DestinationGroupingKeys:            stringArray(endpoint["dst_grouping_keys"]),
				MissingEndpointBehavior:            stringDefault(endpoint["missing_endpoint_behavior"], "error"),
			}
		}
		var endpointGrouping any
		if rule.EndpointGrouping != nil {
			endpointGrouping = map[string]any{
				"source_vertex_aggregation_rule_id":      rule.EndpointGrouping.SourceVertexAggregationRuleID,
				"source_grouping_keys":                   rule.EndpointGrouping.SourceGroupingKeys,
				"destination_vertex_aggregation_rule_id": rule.EndpointGrouping.DestinationVertexAggregationRuleID,
				"destination_grouping_keys":              rule.EndpointGrouping.DestinationGroupingKeys,
				"missing_endpoint_behavior":              rule.EndpointGrouping.MissingEndpointBehavior,
			}
		}
		rule.AggregationIdentityDigest, _ = digestTuple("GPAGG1\n", rule.AggregationRuleID, rule.TargetScope, rule.InputScope, rule.InputKind, rule.ProjectedKind, rule.GroupingKeys, rule.MissingGroupingKeyBehavior, rule.EdgeDirection, endpointGrouping)
		rules = append(rules, rule)
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].AggregationRuleID < rules[j].AggregationRuleID })
	return rules
}

func parseSourceEntities(raw []any) []SourceEntity {
	entities := make([]SourceEntity, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "source_entity")
		entities = append(entities, SourceEntity{
			SourceEntityID:   mustString(object["source_entity_id"], "source_entity_id"),
			SourceEntityKind: mustString(object["source_entity_kind"], "source_entity_kind"),
			Properties:       objectMap(defaultObject(object["properties"]), "properties"),
			Metadata:         objectMap(defaultObject(object["metadata"]), "metadata"),
			Labels:           uniqueSortedStrings(stringArray(object["labels"])),
		})
	}
	sort.SliceStable(entities, func(i, j int) bool { return entities[i].SourceEntityID < entities[j].SourceEntityID })
	return entities
}

func parseSourceRelationships(raw []any) []SourceRelationship {
	relationships := make([]SourceRelationship, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "source_relationship")
		relationships = append(relationships, SourceRelationship{
			SourceRelationshipID:   mustString(object["source_relationship_id"], "source_relationship_id"),
			SourceRelationshipKind: mustString(object["source_relationship_kind"], "source_relationship_kind"),
			SrcSourceEntityID:      stringDefault(object["src_source_entity_id"], ""),
			DstSourceEntityID:      stringDefault(object["dst_source_entity_id"], ""),
			Direction:              stringDefault(object["direction"], "forward"),
			Properties:             objectMap(defaultObject(object["properties"]), "properties"),
			Metadata:               objectMap(defaultObject(object["metadata"]), "metadata"),
			Labels:                 uniqueSortedStrings(stringArray(object["labels"])),
		})
	}
	sort.SliceStable(relationships, func(i, j int) bool {
		return relationships[i].SourceRelationshipID < relationships[j].SourceRelationshipID
	})
	return relationships
}

func parseFilters(raw map[string]any) Filters {
	return Filters{
		EntityFilters:       parseFilterPredicates(arrayValue(raw["entity_filters"], "entity_filters")),
		RelationshipFilters: parseFilterPredicates(arrayValue(raw["relationship_filters"], "relationship_filters")),
		Logic:               stringDefault(raw["logic"], "and"),
	}
}

func parseFilterPredicates(raw []any) []FilterPredicate {
	predicates := make([]FilterPredicate, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "filter")
		value, hasValue := object["value"]
		predicates = append(predicates, FilterPredicate{
			FieldPath: mustString(object["field_path"], "field_path"), Operator: mustString(object["op"], "op"),
			Value: value, HasValue: hasValue, IncludeIfMissing: boolDefault(object["include_if_missing"], false),
		})
	}
	return predicates
}

func parseInclusionPredicate(value any) (string, *FilterPredicate) {
	if value == nil {
		return "always", nil
	}
	if token, ok := value.(string); ok {
		return token, nil
	}
	object := mustObject(value, "inclusion_predicate")
	parsed := parseFilterPredicates([]any{object})
	return "filter", &parsed[0]
}

func normalizeProjectionRequest(request *ProjectionRequest) {
	request.ProjectionConfig.DeclaredSourceEntityKinds = sortedStrings(request.ProjectionConfig.DeclaredSourceEntityKinds)
	request.ProjectionConfig.DeclaredSourceRelationshipKinds = sortedStrings(request.ProjectionConfig.DeclaredSourceRelationshipKinds)
	request.ProjectionConfig.DefaultVertexLabels = uniqueSortedStrings(request.ProjectionConfig.DefaultVertexLabels)
	request.ProjectionConfig.DefaultEdgeLabels = uniqueSortedStrings(request.ProjectionConfig.DefaultEdgeLabels)
	if len(request.ProjectionConfig.RelationshipMappings) > 0 {
		request.RelationshipMappings = request.ProjectionConfig.RelationshipMappings
	}
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func normalizedRequestObject(request ProjectionRequest) map[string]any {
	return map[string]any{
		"projection_schema_id":     request.ProjectionSchemaID,
		"graph_view_id":            request.GraphViewID,
		"source_snapshot_id":       request.SourceSnapshotID,
		"projection_config":        normalizedConfigObject(request.ProjectionConfig),
		"source_entities":          sourceEntitiesObject(request.SourceEntities),
		"source_relationships":     sourceRelationshipsObject(request.SourceRelationships),
		"source_metadata":          request.SourceMetadata,
		"filters":                  filtersObject(request.Filters),
		"relationship_definitions": relationshipMappingsObject(request.RelationshipMappings),
		"property_definitions":     propertyDefinitionsObject(request.PropertyDefinitions),
		"requested_at":             request.RequestedAt,
		"requested_by":             request.RequestedBy,
	}
}

func normalizedConfigObject(config ProjectionConfig) map[string]any {
	return map[string]any{
		"graph_view_key":                     config.GraphViewKey,
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
		"retention_policy": map[string]any{
			"retain_replaced_results":           config.RetentionPolicy.RetainReplacedResults,
			"retention_count":                   retentionInteger(config.RetentionPolicy, "retention_count", config.RetentionPolicy.RetentionCount),
			"retention_duration_seconds":        retentionInteger(config.RetentionPolicy, "retention_duration_seconds", config.RetentionPolicy.RetentionDurationSeconds),
			"retain_failed_results":             config.RetentionPolicy.RetainFailedResults,
			"failed_retention_count":            retentionInteger(config.RetentionPolicy, "failed_retention_count", config.RetentionPolicy.FailedRetentionCount),
			"failed_retention_duration_seconds": retentionInteger(config.RetentionPolicy, "failed_retention_duration_seconds", config.RetentionPolicy.FailedRetentionDurationSecs),
		},
		"custom_config": config.CustomConfig,
	}
}

func projectionConfigDigest(request ProjectionRequest) (string, error) {
	transcript, err := projectionConfigDigestTranscript(request)
	if err != nil {
		return "", err
	}
	return sha256Hex(transcript), nil
}

func projectionConfigDigestTranscript(request ProjectionRequest) ([]byte, error) {
	retention := request.ProjectionConfig.RetentionPolicy
	configCore := canonicalFields(
		canonicalMember{Name: "graph_view_key", Value: request.ProjectionConfig.GraphViewKey},
		canonicalMember{Name: "projection_version", Value: request.ProjectionConfig.ProjectionVersion},
		canonicalMember{Name: "declared_source_entity_kinds", Value: request.ProjectionConfig.DeclaredSourceEntityKinds},
		canonicalMember{Name: "declared_source_relationship_kinds", Value: request.ProjectionConfig.DeclaredSourceRelationshipKinds},
		canonicalMember{Name: "entity_mappings", Value: entityMappingsObject(request.ProjectionConfig.EntityMappings)},
		canonicalMember{Name: "default_vertex_labels", Value: request.ProjectionConfig.DefaultVertexLabels},
		canonicalMember{Name: "default_edge_labels", Value: request.ProjectionConfig.DefaultEdgeLabels},
		canonicalMember{Name: "allow_empty_kind_registry", Value: request.ProjectionConfig.AllowEmptyKindRegistry},
		canonicalMember{Name: "retention_policy", Value: canonicalFields(
			canonicalMember{Name: "retain_replaced_results", Value: retention.RetainReplacedResults},
			canonicalMember{Name: "retention_count", Value: retentionInteger(retention, "retention_count", retention.RetentionCount)},
			canonicalMember{Name: "retention_duration_seconds", Value: retentionInteger(retention, "retention_duration_seconds", retention.RetentionDurationSeconds)},
			canonicalMember{Name: "retain_failed_results", Value: retention.RetainFailedResults},
			canonicalMember{Name: "failed_retention_count", Value: retentionInteger(retention, "failed_retention_count", retention.FailedRetentionCount)},
			canonicalMember{Name: "failed_retention_duration_seconds", Value: retentionInteger(retention, "failed_retention_duration_seconds", retention.FailedRetentionDurationSecs)},
		)},
	)
	source := "none"
	if len(request.RelationshipMappings) > 0 {
		source = "top_level_relationship_definitions"
	}
	if len(request.ProjectionConfig.RelationshipMappings) > 0 {
		source = "projection_config_relationship_mappings"
	}
	return tupleBytesInput("GPCONFIG1\n", request.ProjectionSchemaID, configCore, source, relationshipMappingsObject(request.RelationshipMappings), filtersObject(request.Filters), propertyDefinitionsObject(request.PropertyDefinitions), metadataMappingsObject(request.ProjectionConfig.MetadataMappings), aggregationRulesObject(request.ProjectionConfig.AggregationRules))
}

func retentionInteger(policy RetentionPolicy, key string, value int) any {
	if lexeme, ok := policy.RawIntegerLexemes[key]; ok {
		return json.Number(lexeme)
	}
	return value
}

func projectionSourceDigest(request ProjectionRequest) (string, error) {
	transcript, err := projectionSourceDigestTranscript(request)
	if err != nil {
		return "", err
	}
	return sha256Hex(transcript), nil
}

func projectionSourceDigestTranscript(request ProjectionRequest) ([]byte, error) {
	return tupleBytesInput("GPSOURCE1\n", request.ProjectionSchemaID, request.SourceSnapshotID, sourceEntitiesObject(request.SourceEntities), sourceRelationshipsObject(request.SourceRelationships), request.SourceMetadata)
}

func defaultMissingBehavior(required bool) string {
	if required {
		return "error"
	}
	return "omit"
}

func validIdentifier(value string) bool {
	if value == "" || len([]rune(value)) > 128 || strings.ContainsAny(value, `/\#`) {
		return false
	}
	var first, last rune
	for index, r := range value {
		if r == 0 || unicode.IsControl(r) || (r >= 0x80 && r <= 0x9f) || (r >= 0xd800 && r <= 0xdfff) {
			return false
		}
		if index == 0 {
			first = r
		}
		last = r
	}
	return !isSpecWhitespace(first) && !isSpecWhitespace(last)
}

func validPropertyKey(value string) bool {
	if !validIdentifier(value) || strings.Contains(value, ".") {
		return false
	}
	switch value {
	case "kind", "properties", "metadata", "source_metadata", "projected":
		return false
	default:
		return true
	}
}

func isSpecWhitespace(r rune) bool {
	if r >= 0x09 && r <= 0x0d {
		return true
	}
	switch r {
	case 0x20, 0x85, 0xa0, 0x1680, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000:
		return true
	}
	return r >= 0x2000 && r <= 0x200a
}

func parseTimestamp(value string) (time.Time, error) {
	if !timestampPattern.MatchString(value) {
		return time.Time{}, fmt.Errorf("invalid timestamp")
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Year() < 1 || parsed.Year() > 9999 {
		return time.Time{}, fmt.Errorf("invalid timestamp")
	}
	return parsed, nil
}

func mustString(value any, _ string) string {
	stringValue, _ := value.(string)
	return stringValue
}

func mustObject(value any, _ string) map[string]any {
	object, _ := value.(map[string]any)
	if object == nil {
		return map[string]any{}
	}
	return object
}

func defaultObject(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return mustObject(value, "")
}

func objectMap(value map[string]any, _ string) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	out := map[string]any{}
	for key, entry := range value {
		out[key] = entry
	}
	return out
}

func stringMap(value map[string]any) map[string]string {
	out := map[string]string{}
	for key, entry := range value {
		if stringValue, ok := entry.(string); ok {
			out[key] = stringValue
		}
	}
	return out
}

func arrayValue(value any, _ string) []any {
	if value == nil {
		return []any{}
	}
	array, _ := value.([]any)
	if array == nil {
		return []any{}
	}
	return array
}

func stringArray(value any) []string {
	array, _ := value.([]any)
	if array == nil {
		return []string{}
	}
	out := make([]string, 0, len(array))
	for _, entry := range array {
		if stringValue, ok := entry.(string); ok {
			out = append(out, stringValue)
		}
	}
	return out
}

func stringDefault(value any, fallback string) string {
	if stringValue, ok := value.(string); ok {
		return stringValue
	}
	return fallback
}

func boolDefault(value any, fallback bool) bool {
	if boolValue, ok := value.(bool); ok {
		return boolValue
	}
	return fallback
}

func intDefault(value any, fallback int) int {
	if number, ok := value.(json.Number); ok && validFiniteInteger(number.String()) {
		parsed, err := number.Int64()
		if err == nil {
			return int(parsed)
		}
	}
	return fallback
}

func validFiniteInteger(value string) bool {
	if !finiteIntegerPattern.MatchString(value) {
		return false
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	return err == nil && parsed >= -maxFiniteInteger && parsed <= maxFiniteInteger
}

func uniqueSortedStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
