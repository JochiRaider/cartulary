package graphprojection

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
)

type jsonKind uint8

const (
	kindAny jsonKind = iota
	kindString
	kindObject
	kindArray
	kindBoolean
	kindNumber
)

type memberSpec struct {
	kind     jsonKind
	required bool
	nullable bool
}

var asciiPathIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func canonicalInputMemberPath(parent, member string) string {
	if asciiPathIdentifier.MatchString(member) {
		return parent + "." + member
	}
	encoded, _ := canonicalJSON(member)
	return parent + "[" + string(encoded) + "]"
}

func schemaError(reason, field, validationCode string) error {
	details := map[string]any{"reason_code": reason, "field": nullableField(field), "validation_code": nullableCode(validationCode)}
	return &OperationError{Code: "invalid_projection_request", ReasonCode: reason, Field: field, Details: details}
}

func nullableField(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableCode(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func validateClosedObject(object map[string]any, path string, members map[string]memberSpec, validationCode string) error {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		spec, ok := members[key]
		field := canonicalInputMemberPath(path, key)
		if !ok {
			return schemaError("unknown_member", field, "")
		}
		value := object[key]
		if value == nil {
			if !spec.nullable {
				return schemaError("explicit_null_not_allowed", field, validationCode)
			}
			continue
		}
		if !matchesJSONKind(value, spec.kind) {
			return schemaError("schema_type_mismatch", field, validationCode)
		}
	}
	requiredKeys := make([]string, 0, len(members))
	for key, spec := range members {
		if spec.required {
			requiredKeys = append(requiredKeys, key)
		}
	}
	sort.Strings(requiredKeys)
	for _, key := range requiredKeys {
		spec := members[key]
		if !spec.required {
			continue
		}
		if _, ok := object[key]; !ok {
			return schemaError("missing_required_member", canonicalInputMemberPath(path, key), validationCode)
		}
	}
	return nil
}

func matchesJSONKind(value any, kind jsonKind) bool {
	switch kind {
	case kindAny:
		return true
	case kindString:
		_, ok := value.(string)
		return ok
	case kindObject:
		_, ok := value.(map[string]any)
		return ok
	case kindArray:
		_, ok := value.([]any)
		return ok
	case kindBoolean:
		_, ok := value.(bool)
		return ok
	case kindNumber:
		_, ok := value.(json.Number)
		return ok
	default:
		return false
	}
}

func validateProjectionInputSchema(root map[string]any) error {
	top := map[string]memberSpec{
		"projection_schema_id":     {kind: kindString, required: true},
		"graph_view_id":            {kind: kindString, required: true},
		"source_snapshot_id":       {kind: kindString, required: true},
		"projection_config":        {kind: kindObject, required: true},
		"source_entities":          {kind: kindArray, required: true},
		"source_relationships":     {kind: kindArray, required: true},
		"source_metadata":          {kind: kindObject},
		"filters":                  {kind: kindObject},
		"relationship_definitions": {kind: kindArray},
		"property_definitions":     {kind: kindArray},
		"requested_at":             {kind: kindString, required: true},
		"requested_by":             {kind: kindString, required: true},
	}
	if err := validateClosedObject(root, "$", top, "invalid_input_shape"); err != nil {
		return err
	}
	if err := validateProjectionConfigSchema(root["projection_config"].(map[string]any), "$.projection_config"); err != nil {
		return err
	}
	if err := validateObjectArray(root["source_entities"], "$.source_entities", validateSourceEntitySchema, "invalid_input_shape"); err != nil {
		return err
	}
	if err := validateObjectArray(root["source_relationships"], "$.source_relationships", validateSourceRelationshipSchema, "invalid_input_shape"); err != nil {
		return err
	}
	if value, ok := root["relationship_definitions"]; ok {
		if err := validateObjectArray(value, "$.relationship_definitions", validateRelationshipMappingSchema, "invalid_projection_config"); err != nil {
			return err
		}
	}
	if value, ok := root["property_definitions"]; ok {
		if err := validateObjectArray(value, "$.property_definitions", validatePropertyDefinitionSchema, "invalid_property_definition"); err != nil {
			return err
		}
	}
	if value, ok := root["filters"]; ok {
		if err := validateFiltersSchema(value.(map[string]any), "$.filters"); err != nil {
			return err
		}
	}
	return nil
}

func validateProjectionConfigSchema(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"graph_view_key":                     {kind: kindString, required: true},
		"projection_version":                 {kind: kindString},
		"declared_source_entity_kinds":       {kind: kindArray, required: true},
		"declared_source_relationship_kinds": {kind: kindArray},
		"entity_mappings":                    {kind: kindArray, required: true},
		"relationship_mappings":              {kind: kindArray},
		"metadata_mappings":                  {kind: kindArray},
		"aggregation_rules":                  {kind: kindArray},
		"default_vertex_labels":              {kind: kindArray},
		"default_edge_labels":                {kind: kindArray},
		"allow_empty_kind_registry":          {kind: kindBoolean},
		"retention_policy":                   {kind: kindObject},
		"custom_config":                      {kind: kindObject},
	}
	if err := validateClosedObject(object, path, members, "invalid_projection_config"); err != nil {
		return err
	}
	for _, key := range []string{"declared_source_entity_kinds", "declared_source_relationship_kinds", "default_vertex_labels", "default_edge_labels"} {
		if value, ok := object[key]; ok {
			if err := validateStringArray(value, canonicalInputMemberPath(path, key), "invalid_projection_config"); err != nil {
				return err
			}
		}
	}
	if err := validateObjectArray(object["entity_mappings"], path+".entity_mappings", validateEntityMappingSchema, "invalid_projection_config"); err != nil {
		return err
	}
	for key, validator := range map[string]func(map[string]any, string) error{
		"relationship_mappings": validateRelationshipMappingSchema,
		"metadata_mappings":     validateMetadataMappingSchema,
		"aggregation_rules":     validateAggregationRuleSchema,
	} {
		if value, ok := object[key]; ok {
			if err := validateObjectArray(value, canonicalInputMemberPath(path, key), validator, "invalid_projection_config"); err != nil {
				return err
			}
		}
	}
	if value, ok := object["retention_policy"]; ok {
		if err := validateRetentionPolicySchema(value.(map[string]any), path+".retention_policy"); err != nil {
			return err
		}
	}
	return nil
}

func validateEntityMappingSchema(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"mapping_rule_id": {kind: kindString, required: true}, "source_entity_kind": {kind: kindString, required: true},
		"projected_vertex_kind": {kind: kindString, required: true}, "inclusion_predicate": {kind: kindAny},
		"label_policy": {kind: kindString}, "mapping_labels": {kind: kindArray},
		"required_property_keys": {kind: kindArray}, "optional_property_keys": {kind: kindArray},
	}
	if err := validateClosedObject(object, path, members, "invalid_mapping_rule"); err != nil {
		return err
	}
	return validateMappingArraysAndPredicate(object, path)
}

func validateRelationshipMappingSchema(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"mapping_rule_id": {kind: kindString, required: true}, "source_relationship_kind": {kind: kindString, required: true},
		"projected_edge_kind": {kind: kindString, required: true}, "inclusion_predicate": {kind: kindAny},
		"direction_policy": {kind: kindString}, "emit_reverse_edge": {kind: kindBoolean}, "reverse_edge_kind": {kind: kindString},
		"label_policy": {kind: kindString}, "mapping_labels": {kind: kindArray},
		"required_property_keys": {kind: kindArray}, "optional_property_keys": {kind: kindArray},
	}
	if err := validateClosedObject(object, path, members, "invalid_mapping_rule"); err != nil {
		return err
	}
	return validateMappingArraysAndPredicate(object, path)
}

func validateMappingArraysAndPredicate(object map[string]any, path string) error {
	for _, key := range []string{"mapping_labels", "required_property_keys", "optional_property_keys"} {
		if value, ok := object[key]; ok {
			if err := validateStringArray(value, canonicalInputMemberPath(path, key), "invalid_mapping_rule"); err != nil {
				return err
			}
		}
	}
	if value, ok := object["inclusion_predicate"]; ok {
		switch typed := value.(type) {
		case string:
		case map[string]any:
			if err := validateFilterPredicateSchema(typed, path+".inclusion_predicate"); err != nil {
				return err
			}
		default:
			return schemaError("schema_type_mismatch", path+".inclusion_predicate", "invalid_mapping_rule")
		}
	}
	return nil
}

func validateMetadataMappingSchema(object map[string]any, path string) error {
	members := definitionMembers("metadata_mapping_id", "projected_metadata_key")
	return validateClosedObject(object, path, members, "invalid_metadata_mapping")
}

func validatePropertyDefinitionSchema(object map[string]any, path string) error {
	members := definitionMembers("property_definition_id", "projected_key")
	return validateClosedObject(object, path, members, "invalid_property_definition")
}

func definitionMembers(idKey, outputKey string) map[string]memberSpec {
	return map[string]memberSpec{
		idKey: {kind: kindString, required: true}, "target_scope": {kind: kindString, required: true}, "target_kind": {kind: kindString, required: true},
		"source_field_path": {kind: kindString, required: true}, outputKey: {kind: kindString, required: true}, "projected_type": {kind: kindString, required: true},
		"required": {kind: kindBoolean}, "default_value": {kind: kindAny, nullable: true}, "missing_behavior": {kind: kindString},
		"source_null_behavior": {kind: kindString}, "null_output_policy": {kind: kindString}, "merge_behavior": {kind: kindString},
	}
}

func validateAggregationRuleSchema(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"aggregation_rule_id": {kind: kindString, required: true}, "target_scope": {kind: kindString, required: true},
		"input_scope": {kind: kindString, required: true}, "input_kind": {kind: kindString, required: true}, "projected_kind": {kind: kindString, required: true},
		"grouping_keys": {kind: kindArray, required: true}, "missing_grouping_key_behavior": {kind: kindString}, "source_reference_policy": {kind: kindString},
		"property_merge_behavior": {kind: kindObject}, "edge_direction": {kind: kindString}, "endpoint_grouping": {kind: kindObject},
	}
	if err := validateClosedObject(object, path, members, "invalid_aggregation_rule"); err != nil {
		return err
	}
	if err := validateStringArray(object["grouping_keys"], path+".grouping_keys", "invalid_aggregation_rule"); err != nil {
		return err
	}
	if value, ok := object["property_merge_behavior"]; ok {
		for key, item := range value.(map[string]any) {
			if _, ok := item.(string); !ok {
				return schemaError("schema_type_mismatch", canonicalInputMemberPath(path+".property_merge_behavior", key), "invalid_aggregation_rule")
			}
		}
	}
	if value, ok := object["endpoint_grouping"]; ok {
		return validateEndpointGroupingSchema(value.(map[string]any), path+".endpoint_grouping")
	}
	return nil
}

func validateEndpointGroupingSchema(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"src_vertex_aggregation_rule_id": {kind: kindString, required: true}, "src_grouping_keys": {kind: kindArray, required: true},
		"dst_vertex_aggregation_rule_id": {kind: kindString, required: true}, "dst_grouping_keys": {kind: kindArray, required: true},
		"missing_endpoint_behavior": {kind: kindString},
	}
	if err := validateClosedObject(object, path, members, "invalid_aggregation_rule"); err != nil {
		return err
	}
	for _, key := range []string{"src_grouping_keys", "dst_grouping_keys"} {
		if err := validateStringArray(object[key], canonicalInputMemberPath(path, key), "invalid_aggregation_rule"); err != nil {
			return err
		}
	}
	return nil
}

func validateSourceEntitySchema(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"source_entity_id": {kind: kindString, required: true}, "source_entity_kind": {kind: kindString, required: true},
		"properties": {kind: kindObject}, "metadata": {kind: kindObject}, "labels": {kind: kindArray},
	}
	if err := validateClosedObject(object, path, members, "invalid_input_shape"); err != nil {
		return err
	}
	if value, ok := object["labels"]; ok {
		return validateStringArray(value, path+".labels", "invalid_input_shape")
	}
	return nil
}

func validateSourceRelationshipSchema(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"source_relationship_id": {kind: kindString, required: true}, "source_relationship_kind": {kind: kindString, required: true},
		"src_source_entity_id": {kind: kindString}, "dst_source_entity_id": {kind: kindString}, "direction": {kind: kindString},
		"properties": {kind: kindObject}, "metadata": {kind: kindObject}, "labels": {kind: kindArray},
	}
	if err := validateClosedObject(object, path, members, "invalid_input_shape"); err != nil {
		return err
	}
	if value, ok := object["labels"]; ok {
		return validateStringArray(value, path+".labels", "invalid_input_shape")
	}
	return nil
}

func validateRetentionPolicySchema(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"retain_replaced_results": {kind: kindBoolean}, "retention_count": {kind: kindNumber}, "retention_duration_seconds": {kind: kindNumber},
		"retain_failed_results": {kind: kindBoolean}, "failed_retention_count": {kind: kindNumber}, "failed_retention_duration_seconds": {kind: kindNumber},
	}
	return validateClosedObject(object, path, members, "invalid_retention_policy")
}

func validateFiltersSchema(object map[string]any, path string) error {
	members := map[string]memberSpec{"entity_filters": {kind: kindArray}, "relationship_filters": {kind: kindArray}, "logic": {kind: kindString}}
	if err := validateClosedObject(object, path, members, "invalid_filter"); err != nil {
		return err
	}
	for _, key := range []string{"entity_filters", "relationship_filters"} {
		if value, ok := object[key]; ok {
			if err := validateObjectArray(value, canonicalInputMemberPath(path, key), validateFilterPredicateSchema, "invalid_filter"); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateFilterPredicateSchema(object map[string]any, path string) error {
	members := map[string]memberSpec{
		"field_path": {kind: kindString, required: true}, "op": {kind: kindString, required: true},
		"value": {kind: kindAny, nullable: true}, "include_if_missing": {kind: kindBoolean},
	}
	return validateClosedObject(object, path, members, "invalid_filter")
}

func validateObjectArray(value any, path string, validator func(map[string]any, string) error, validationCode string) error {
	if value == nil {
		return nil
	}
	array, ok := value.([]any)
	if !ok {
		return schemaError("schema_type_mismatch", path, validationCode)
	}
	for index, item := range array {
		itemPath := path + "[" + strconv.Itoa(index) + "]"
		object, ok := item.(map[string]any)
		if !ok {
			return schemaError("schema_type_mismatch", itemPath, validationCode)
		}
		if err := validator(object, itemPath); err != nil {
			return err
		}
	}
	return nil
}

func validateStringArray(value any, path, validationCode string) error {
	array, ok := value.([]any)
	if !ok {
		return schemaError("schema_type_mismatch", path, validationCode)
	}
	for index, item := range array {
		if _, ok := item.(string); !ok {
			return schemaError("schema_type_mismatch", fmt.Sprintf("%s[%d]", path, index), validationCode)
		}
	}
	return nil
}

func annotateAdmissionError(err error, operation string) error {
	opErr, ok := err.(*OperationError)
	if !ok {
		return err
	}
	if opErr.Details == nil {
		opErr.Details = map[string]any{}
	}
	opErr.Details["operation"] = operation
	opErr.Details["reason_code"] = opErr.ReasonCode
	if _, ok := opErr.Details["field"]; !ok {
		opErr.Details["field"] = nullableField(opErr.Field)
	}
	if _, ok := opErr.Details["validation_code"]; !ok {
		opErr.Details["validation_code"] = nil
	}
	return opErr
}
