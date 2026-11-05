package graphprojection

import (
	"context"
	"maps"
	"slices"
)

// The member lists below are the request's traversal order. Descend while
// visiting each field so a later sibling cannot mask its structural errors.
type inputObjectV2 []inputMemberV2
type inputMemberV2 struct {
	name     string
	shape    inputValueShapeV2
	presence inputPresenceV2
}
type inputValueShapeV2 struct {
	kind     jsonKind
	children func(*inputShapeWalkV2, any, string) error
}
type inputShapeWalkV2 struct {
	ctx            context.Context
	oversizedArray bool
}

type inputPresenceV2 uint8

const (
	requiredInputV2 inputPresenceV2 = iota
	defaultInputV2
	predicateValueV2
	reverseKindV2
	edgeRuleMemberV2
)

func (s inputObjectV2) validate(w *inputShapeWalkV2, object map[string]any, path string) error {
	known := 0
	for _, member := range s {
		field := path + "." + member.name
		v, present := object[member.name]
		switch member.presence {
		case requiredInputV2:
			if !present {
				return invalidProjectionErrorV2("missing_required_member", field)
			}
		case predicateValueV2:
			if present != (object["operator"] != "exists") {
				return invalidProjectionErrorV2("filter_value_presence", field)
			}
		case reverseKindV2:
			if present != object["emit_reverse_edge"].(bool) {
				return invalidProjectionErrorV2("reverse_edge_kind_presence", field)
			}
		case edgeRuleMemberV2:
			if !present && object["target_scope"] == "edge" {
				return invalidProjectionErrorV2("missing_edge_aggregation_member", path)
			}
			if present && object["target_scope"] == "vertex" {
				return invalidProjectionErrorV2("vertex_aggregation_edge_member", path)
			}
		}
		if !present {
			continue
		}
		known++
		if v == nil {
			if member.presence != defaultInputV2 && member.presence != predicateValueV2 {
				return invalidProjectionErrorV2("explicit_null_not_allowed", field)
			}
			continue
		}
		if !matchesJSONKind(v, member.shape.kind) {
			return invalidProjectionErrorV2("schema_type_mismatch", field)
		}
		if member.shape.children != nil {
			if err := member.shape.children(w, v, field); err != nil {
				return err
			}
		}
	}
	if len(object) != known {
		for _, key := range slices.Sorted(maps.Keys(object)) {
			if !slices.ContainsFunc(s, func(m inputMemberV2) bool { return m.name == key }) {
				return invalidProjectionErrorV2("unknown_member", canonicalInputMemberPath(path, key))
			}
		}
	}
	return nil
}

func inputRecordV2(shape inputObjectV2) inputValueShapeV2 {
	return inputValueShapeV2{kindObject, func(w *inputShapeWalkV2, v any, path string) error {
		return shape.validate(w, v.(map[string]any), path)
	}}
}
func inputRowsV2(shape inputObjectV2) inputValueShapeV2 {
	return inputValueShapeV2{kindArray, func(w *inputShapeWalkV2, v any, path string) error {
		return validateObjectArrayV2(w.ctx, v, path, func(row map[string]any, p string) error { return shape.validate(w, row, p) })
	}}
}
func inputPropertyV2(w *inputShapeWalkV2, v any, path string) error {
	if a, ok := v.([]any); ok && len(a) > graphProjectionLimits.MaxArrayItems {
		w.oversizedArray = true
	}
	return validatePropertyShapeV2(v, path)
}

var (
	textInput       = inputValueShapeV2{kind: kindString}
	boolInput       = inputValueShapeV2{kind: kindBoolean}
	jsonMapInput    = inputValueShapeV2{kind: kindObject}
	textsInput      = inputValueShapeV2{kindArray, func(_ *inputShapeWalkV2, v any, path string) error { return validateStringArrayV2(v, path) }}
	propertyInput   = inputValueShapeV2{kindAny, inputPropertyV2}
	propertiesInput = inputValueShapeV2{kindObject, func(w *inputShapeWalkV2, v any, path string) error {
		props := v.(map[string]any)
		for _, key := range slices.Sorted(maps.Keys(props)) {
			if err := inputPropertyV2(w, props[key], canonicalInputMemberPath(path, key)); err != nil {
				return err
			}
		}
		return nil
	}}
	mergeInput = inputValueShapeV2{kindObject, func(_ *inputShapeWalkV2, v any, path string) error {
		props := v.(map[string]any)
		for _, key := range slices.Sorted(maps.Keys(props)) {
			if _, ok := props[key].(string); !ok {
				return invalidProjectionErrorV2("schema_type_mismatch", canonicalInputMemberPath(path, key))
			}
		}
		return nil
	}}
	inclusionInput = inputValueShapeV2{kindAny, func(w *inputShapeWalkV2, v any, path string) error {
		switch p := v.(type) {
		case string:
			return nil
		case map[string]any:
			return filterShapeV2.validate(w, p, path)
		default:
			return invalidProjectionErrorV2("schema_type_mismatch", path)
		}
	}}
)

var filterShapeV2 = inputObjectV2{
	{"field_path", textInput, requiredInputV2},
	{"operator", textInput, requiredInputV2},
	{"value", propertyInput, predicateValueV2},
	{"include_if_missing", boolInput, requiredInputV2},
}

var filtersShapeV2 = inputObjectV2{
	{"entity_filters", inputRowsV2(filterShapeV2), requiredInputV2},
	{"relationship_filters", inputRowsV2(filterShapeV2), requiredInputV2},
	{"logic", textInput, requiredInputV2},
}

func mappingShapeV2(source, output string) inputObjectV2 {
	return inputObjectV2{
		{"mapping_rule_id", textInput, requiredInputV2},
		{source, textInput, requiredInputV2},
		{output, textInput, requiredInputV2},
		{"inclusion_predicate", inclusionInput, requiredInputV2},
		{"label_policy", textInput, requiredInputV2},
		{"mapping_labels", textsInput, requiredInputV2},
		{"required_property_keys", textsInput, requiredInputV2},
		{"optional_property_keys", textsInput, requiredInputV2},
	}
}

var entityMappingShape = mappingShapeV2("source_entity_kind", "projected_vertex_kind")
var relationshipMappingShape = append(mappingShapeV2("source_relationship_kind", "projected_edge_kind"), inputMemberV2{"direction_policy", textInput, requiredInputV2}, inputMemberV2{"emit_reverse_edge", boolInput, requiredInputV2}, inputMemberV2{"reverse_edge_kind", textInput, reverseKindV2})

func definitionShapeV2(id, output string) inputObjectV2 {
	return inputObjectV2{
		{id, textInput, requiredInputV2},
		{"target_scope", textInput, requiredInputV2},
		{"target_kind", textInput, requiredInputV2},
		{"source_field_path", textInput, requiredInputV2},
		{output, textInput, requiredInputV2},
		{"projected_type", textInput, requiredInputV2},
		{"required", boolInput, requiredInputV2},
		{"default_value", propertyInput, defaultInputV2},
		{"missing_behavior", textInput, requiredInputV2},
		{"source_null_behavior", textInput, requiredInputV2},
		{"null_output_policy", textInput, requiredInputV2},
		{"merge_behavior", textInput, requiredInputV2},
	}
}

var endpointShapeV2 = inputObjectV2{
	{"src_vertex_aggregation_rule_id", textInput, requiredInputV2},
	{"src_grouping_keys", textsInput, requiredInputV2},
	{"dst_vertex_aggregation_rule_id", textInput, requiredInputV2},
	{"dst_grouping_keys", textsInput, requiredInputV2},
	{"missing_endpoint_behavior", textInput, requiredInputV2},
}

var aggregationShapeV2 = inputObjectV2{
	{"aggregation_rule_id", textInput, requiredInputV2},
	{"target_scope", textInput, requiredInputV2},
	{"input_scope", textInput, requiredInputV2},
	{"input_kind", textInput, requiredInputV2},
	{"projected_kind", textInput, requiredInputV2},
	{"grouping_keys", textsInput, requiredInputV2},
	{"missing_grouping_key_behavior", textInput, requiredInputV2},
	{"property_merge_behavior", mergeInput, requiredInputV2},
	{"edge_direction", textInput, edgeRuleMemberV2},
	{"endpoint_grouping", inputRecordV2(endpointShapeV2), edgeRuleMemberV2},
}

var entityShapeV2 = inputObjectV2{
	{"source_entity_id", textInput, requiredInputV2},
	{"source_entity_kind", textInput, requiredInputV2},
	{"properties", propertiesInput, requiredInputV2},
	{"metadata", jsonMapInput, requiredInputV2},
	{"labels", textsInput, requiredInputV2},
}

var relationshipShapeV2 = inputObjectV2{
	{"source_relationship_id", textInput, requiredInputV2},
	{"source_relationship_kind", textInput, requiredInputV2},
	{"src_source_entity_id", textInput, requiredInputV2},
	{"dst_source_entity_id", textInput, requiredInputV2},
	{"direction", textInput, requiredInputV2},
	{"properties", propertiesInput, requiredInputV2},
	{"metadata", jsonMapInput, requiredInputV2},
	{"labels", textsInput, requiredInputV2},
}

var configShapeV2 = inputObjectV2{
	{"projection_version", textInput, requiredInputV2},
	{"declared_source_entity_kinds", textsInput, requiredInputV2},
	{"declared_source_relationship_kinds", textsInput, requiredInputV2},
	{"entity_mappings", inputRowsV2(entityMappingShape), requiredInputV2},
	{"relationship_mappings", inputRowsV2(relationshipMappingShape), requiredInputV2},
	{"metadata_mappings", inputRowsV2(definitionShapeV2("metadata_mapping_id", "projected_metadata_key")), requiredInputV2},
	{"aggregation_rules", inputRowsV2(aggregationShapeV2), requiredInputV2},
	{"default_vertex_labels", textsInput, requiredInputV2},
	{"default_edge_labels", textsInput, requiredInputV2},
	{"allow_empty_kind_registry", boolInput, requiredInputV2},
}

var requestShapeV2 = inputObjectV2{
	{"projection_schema_id", textInput, requiredInputV2},
	{"source_snapshot_id", textInput, requiredInputV2},
	{"projection_config", inputRecordV2(configShapeV2), requiredInputV2},
	{"source_entities", inputRowsV2(entityShapeV2), requiredInputV2},
	{"source_relationships", inputRowsV2(relationshipShapeV2), requiredInputV2},
	{"source_metadata", jsonMapInput, requiredInputV2},
	{"filters", inputRecordV2(filtersShapeV2), requiredInputV2},
	{"property_definitions", inputRowsV2(definitionShapeV2("property_definition_id", "projected_key")), requiredInputV2},
}
