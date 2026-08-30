package graphprojection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	finiteIntegerPattern = regexp.MustCompile(`^(0|-?[1-9][0-9]*)$`)
	generatedIDPattern   = regexp.MustCompile(`^(gv|gpr|vx|ed|gpi)_[0-9a-f]{64}$`)
	timestampPattern     = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,6})?Z$`)
)

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

func parseProjectionConfig(raw map[string]any) projectionConfig {
	config := projectionConfig{
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
	}
	return config
}

func parseEntityMappings(raw []any) []entityMapping {
	mappings := make([]entityMapping, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "entity_mapping")
		mapping := entityMapping{
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

func parseRelationshipMappings(raw []any) []relationshipMapping {
	mappings := make([]relationshipMapping, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "relationship_mapping")
		projectedKind := mustString(object["projected_edge_kind"], "projected_edge_kind")
		mapping := relationshipMapping{
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

func parsePropertyDefinitions(raw []any) []propertyDefinition {
	definitions := make([]propertyDefinition, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "property_definition")
		required := boolDefault(object["required"], false)
		definition := propertyDefinition{
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

func parseMetadataMappings(raw []any) []metadataMapping {
	mappings := make([]metadataMapping, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "metadata_mapping")
		required := boolDefault(object["required"], false)
		mapping := metadataMapping{
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

func parseAggregationRules(raw []any) []aggregationRule {
	rules := make([]aggregationRule, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "aggregation_rule")
		rule := aggregationRule{
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
			rule.endpointGrouping = &endpointGrouping{
				SourceVertexAggregationRuleID:      mustString(endpoint["src_vertex_aggregation_rule_id"], "src_vertex_aggregation_rule_id"),
				SourceGroupingKeys:                 stringArray(endpoint["src_grouping_keys"]),
				DestinationVertexAggregationRuleID: mustString(endpoint["dst_vertex_aggregation_rule_id"], "dst_vertex_aggregation_rule_id"),
				DestinationGroupingKeys:            stringArray(endpoint["dst_grouping_keys"]),
				MissingEndpointBehavior:            stringDefault(endpoint["missing_endpoint_behavior"], "error"),
			}
		}
		var endpointGrouping any
		if rule.endpointGrouping != nil {
			endpointGrouping = map[string]any{
				"source_vertex_aggregation_rule_id":      rule.endpointGrouping.SourceVertexAggregationRuleID,
				"source_grouping_keys":                   rule.endpointGrouping.SourceGroupingKeys,
				"destination_vertex_aggregation_rule_id": rule.endpointGrouping.DestinationVertexAggregationRuleID,
				"destination_grouping_keys":              rule.endpointGrouping.DestinationGroupingKeys,
				"missing_endpoint_behavior":              rule.endpointGrouping.MissingEndpointBehavior,
			}
		}
		rule.AggregationIdentityDigest, _ = digestTuple("GPAGG1\n", rule.AggregationRuleID, rule.TargetScope, rule.InputScope, rule.InputKind, rule.ProjectedKind, rule.GroupingKeys, rule.MissingGroupingKeyBehavior, rule.EdgeDirection, endpointGrouping)
		rules = append(rules, rule)
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].AggregationRuleID < rules[j].AggregationRuleID })
	return rules
}

func parseSourceEntities(raw []any) []sourceEntity {
	entities := make([]sourceEntity, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "source_entity")
		entities = append(entities, sourceEntity{
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

func parseSourceRelationships(raw []any) []sourceRelationship {
	relationships := make([]sourceRelationship, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "source_relationship")
		relationships = append(relationships, sourceRelationship{
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

func parseFilters(raw map[string]any) filters {
	return filters{
		EntityFilters:       parseFilterPredicates(arrayValue(raw["entity_filters"], "entity_filters")),
		RelationshipFilters: parseFilterPredicates(arrayValue(raw["relationship_filters"], "relationship_filters")),
		Logic:               stringDefault(raw["logic"], "and"),
	}
}

func parseFilterPredicates(raw []any) []filterPredicate {
	predicates := make([]filterPredicate, 0, len(raw))
	for _, entry := range raw {
		object := mustObject(entry, "filter")
		value, hasValue := object["value"]
		predicates = append(predicates, filterPredicate{
			FieldPath: mustString(object["field_path"], "field_path"), Operator: mustString(object["operator"], "operator"),
			Value: value, HasValue: hasValue, IncludeIfMissing: boolDefault(object["include_if_missing"], false),
		})
	}
	return predicates
}

func parseInclusionPredicate(value any) (string, *filterPredicate) {
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

func normalizeProjectionRequest(request *projectionRequest) {
	request.projectionConfig.DeclaredSourceEntityKinds = sortedStrings(request.projectionConfig.DeclaredSourceEntityKinds)
	request.projectionConfig.DeclaredSourceRelationshipKinds = sortedStrings(request.projectionConfig.DeclaredSourceRelationshipKinds)
	request.projectionConfig.DefaultVertexLabels = uniqueSortedStrings(request.projectionConfig.DefaultVertexLabels)
	request.projectionConfig.DefaultEdgeLabels = uniqueSortedStrings(request.projectionConfig.DefaultEdgeLabels)
}

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func defaultMissingBehavior(required bool) string {
	if required {
		return "error"
	}
	return "omit"
}

func validPropertyKey(value string) bool {
	if !validIdentifierV2(value) || len(value) > graphProjectionLimits.MaxPropertyKeyBytes || strings.Contains(value, ".") {
		return false
	}
	switch value {
	case "kind", "properties", "metadata", "source_metadata", "projected":
		return false
	default:
		return true
	}
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

func validFiniteInteger(value string) bool {
	if !finiteIntegerPattern.MatchString(value) {
		return false
	}
	_, err := strconv.ParseInt(value, 10, 64)
	return err == nil
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
