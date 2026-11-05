package graphprojection

import (
	"context"
	"fmt"
	"maps"
	"slices"
)

// The preceding shape walk guarantees every type assertion here. Walk in
// contract order so a request error can take precedence over counted issues
// without retaining the submitted values or allocating an issue per field.
type scalarAdmissionV2 struct {
	ctx    context.Context
	count  int
	reason string
}

func (a *scalarAdmissionV2) issue(reason string) error {
	if a.count == graphProjectionLimits.MaxValidationIssues {
		return validationIssueLimitV2()
	}
	a.count++
	// Source direction is the only nonfatal scalar-domain issue.
	if a.reason == "" || a.reason == "invalid_relationship_direction" || (reason != "invalid_relationship_direction" && reason < a.reason) {
		a.reason = reason
	}
	return nil
}

func (a *scalarAdmissionV2) labels(raw any) error {
	for _, value := range raw.([]any) {
		if !validLabelV2(value.(string)) {
			if err := a.issue("invalid_input_shape"); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRequestScalarsV2(ctx context.Context, root map[string]any) error {
	if root["projection_schema_id"] != ProjectionSchemaIDV2 {
		return invalidProjectionErrorV2("invalid_projection_schema", "$.projection_schema_id")
	}
	if !validIdentifierV2(root["source_snapshot_id"].(string)) {
		return invalidProjectionErrorV2("invalid_identifier", "$.source_snapshot_id")
	}
	a := scalarAdmissionV2{ctx: ctx}
	config := root["projection_config"].(map[string]any)
	if err := a.config(config, "$.projection_config"); err != nil {
		return err
	}
	for _, collection := range []string{"source_entities", "source_relationships"} {
		for i, value := range root[collection].([]any) {
			if err := itemCheckpointV2(ctx); err != nil {
				return cancelledProjectionErrorV2(err)
			}
			row := value.(map[string]any)
			keys := []string{"source_entity_id", "source_entity_kind"}
			if collection == "source_relationships" {
				keys = []string{"source_relationship_id", "source_relationship_kind", "src_source_entity_id", "dst_source_entity_id"}
			}
			for _, k := range keys {
				if !validIdentifierV2(row[k].(string)) {
					if err := a.issue("invalid_input_shape"); err != nil {
						return err
					}
				}
			}
			if collection == "source_relationships" && !validSourceDirection(row["direction"].(string)) {
				if err := a.issue("invalid_relationship_direction"); err != nil {
					return err
				}
			}
			path := fmt.Sprintf("$.%s[%d]", collection, i)
			if err := validatePropertyMapV2(row["properties"].(map[string]any), path+".properties"); err != nil {
				return err
			}
			if err := validateCanonicalJSONValueV2(row["metadata"], path+".metadata", false); err != nil {
				return err
			}
			if err := a.labels(row["labels"]); err != nil {
				return err
			}
		}
	}
	if err := validateCanonicalJSONValueV2(root["source_metadata"], "$.source_metadata", false); err != nil {
		return err
	}
	filters := root["filters"].(map[string]any)
	for _, k := range []string{"entity_filters", "relationship_filters"} {
		if err := a.rows(filters, k, "$.filters", a.predicate); err != nil {
			return err
		}
	}
	if filters["logic"] != "and" {
		if err := a.issue("invalid_filter"); err != nil {
			return err
		}
	}
	if err := a.rows(root, "property_definitions", "$", a.property); err != nil {
		return err
	}
	if a.count != 0 {
		return validationFailureV2(a.reason, a.count)
	}
	return nil
}

func (a *scalarAdmissionV2) rows(parent map[string]any, key, path string, visit func(map[string]any, string) error) error {
	for i, raw := range parent[key].([]any) {
		if err := itemCheckpointV2(a.ctx); err != nil {
			return cancelledProjectionErrorV2(err)
		}
		if err := visit(raw.(map[string]any), fmt.Sprintf("%s.%s[%d]", path, key, i)); err != nil {
			return err
		}
	}
	return nil
}

func scalarStringsV2(object map[string]any, key, path, reason string, valid func(string) bool) error {
	for i, raw := range object[key].([]any) {
		if !valid(raw.(string)) {
			return invalidProjectionErrorV2(reason, fmt.Sprintf("%s.%s[%d]", path, key, i))
		}
	}
	return nil
}

func (a *scalarAdmissionV2) config(config map[string]any, path string) error {
	if !validIdentifierV2(config["projection_version"].(string)) {
		return invalidProjectionErrorV2("invalid_identifier", path+".projection_version")
	}
	for _, key := range []string{"declared_source_entity_kinds", "declared_source_relationship_kinds"} {
		if err := scalarStringsV2(config, key, path, "invalid_identifier", validIdentifierV2); err != nil {
			return err
		}
	}
	for _, key := range []string{"entity_mappings", "relationship_mappings"} {
		if err := a.rows(config, key, path, a.mapping); err != nil {
			return err
		}
	}
	if err := a.rows(config, "metadata_mappings", path, a.metadata); err != nil {
		return err
	}
	if err := a.rows(config, "aggregation_rules", path, a.aggregation); err != nil {
		return err
	}
	if err := a.labels(config["default_vertex_labels"]); err != nil {
		return err
	}
	return a.labels(config["default_edge_labels"])
}

func (a *scalarAdmissionV2) mapping(row map[string]any, path string) error {
	for _, key := range []string{"mapping_rule_id", "source_entity_kind", "projected_vertex_kind", "source_relationship_kind", "projected_edge_kind"} {
		if v, ok := row[key]; ok && !validIdentifierV2(v.(string)) {
			return invalidProjectionErrorV2("invalid_identifier", path+"."+key)
		}
	}
	switch predicate := row["inclusion_predicate"].(type) {
	case string:
		if predicate != "always" {
			return invalidProjectionErrorV2("invalid_inclusion_predicate", path+".inclusion_predicate")
		}
	case map[string]any:
		if err := a.predicate(predicate, path+".inclusion_predicate"); err != nil {
			return err
		}
	}
	switch row["label_policy"] {
	case "mapping_only", "preserve_source", "mapping_then_source":
	default:
		if err := a.issue("invalid_mapping_rule"); err != nil {
			return err
		}
	}
	if err := a.labels(row["mapping_labels"]); err != nil {
		return err
	}
	for _, key := range []string{"required_property_keys", "optional_property_keys"} {
		if err := scalarStringsV2(row, key, path, "invalid_property_key", validPropertyKey); err != nil {
			return err
		}
	}
	if policy, ok := row["direction_policy"].(string); ok {
		if _, valid := mappingDirectionPoliciesV2[policy]; !valid {
			if err := a.issue("invalid_direction_policy"); err != nil {
				return err
			}
		}
	}
	if v, ok := row["reverse_edge_kind"]; ok && !validIdentifierV2(v.(string)) {
		return invalidProjectionErrorV2("invalid_identifier", path+".reverse_edge_kind")
	}
	return nil
}

func (a *scalarAdmissionV2) predicate(row map[string]any, path string) error {
	if !validFieldPath(row["field_path"].(string)) {
		return invalidProjectionErrorV2("invalid_field_path", path+".field_path")
	}
	switch row["operator"] {
	case "exists", "equals", "not_equals", "in":
	default:
		if err := a.issue("invalid_filter"); err != nil {
			return err
		}
	}
	return validateCanonicalJSONValueV2(row["value"], path+".value", true)
}

func (a *scalarAdmissionV2) property(row map[string]any, path string) error {
	return a.definition(row, path, "property_definition_id", "projected_key", "invalid_property_definition")
}
func (a *scalarAdmissionV2) metadata(row map[string]any, path string) error {
	return a.definition(row, path, "metadata_mapping_id", "projected_metadata_key", "invalid_metadata_mapping")
}

func (a *scalarAdmissionV2) definition(row map[string]any, path, idKey, keyField, reason string) error {
	if !validIdentifierV2(row[idKey].(string)) {
		return invalidProjectionErrorV2("invalid_identifier", path+"."+idKey)
	}
	scope := row["target_scope"]
	if scope != "vertex" && scope != "edge" && !(scope == "graph_view" && keyField == "projected_metadata_key") {
		if err := a.issue(reason); err != nil {
			return err
		}
	}
	if !validIdentifierV2(row["target_kind"].(string)) {
		return invalidProjectionErrorV2("invalid_identifier", path+".target_kind")
	}
	if !validFieldPath(row["source_field_path"].(string)) {
		return invalidProjectionErrorV2("invalid_field_path", path+".source_field_path")
	}
	if !validPropertyKey(row[keyField].(string)) {
		return invalidProjectionErrorV2("invalid_property_key", path+"."+keyField)
	}
	if keyField == "projected_metadata_key" && reservedSystemMetadataTerminal(row[keyField].(string)) {
		if err := a.issue(reason); err != nil {
			return err
		}
	}
	if !validProjectedType(row["projected_type"].(string)) {
		if err := a.issue(reason); err != nil {
			return err
		}
	}
	if err := validateCanonicalJSONValueV2(row["default_value"], path+".default_value", true); err != nil {
		return err
	}
	if !validMissingPolicyV2(row["missing_behavior"].(string)) {
		if err := a.issue(reason); err != nil {
			return err
		}
	}
	if p := row["source_null_behavior"].(string); !validMissingPolicyV2(p) && p != "emit_null" {
		if err := a.issue(reason); err != nil {
			return err
		}
	}
	if row["null_output_policy"] != "emit_null" && row["null_output_policy"] != "omit" {
		if err := a.issue(reason); err != nil {
			return err
		}
	}
	if !validMergeNameV2(row["merge_behavior"].(string)) {
		if err := a.issue(reason); err != nil {
			return err
		}
	}
	return nil
}

func validMergeNameV2(value string) bool {
	for _, behaviors := range projectedMergeMatrixV2 {
		if _, ok := behaviors[value]; ok {
			return true
		}
	}
	return false
}

func (a *scalarAdmissionV2) aggregation(row map[string]any, path string) error {
	if !validIdentifierV2(row["aggregation_rule_id"].(string)) {
		return invalidProjectionErrorV2("invalid_identifier", path+".aggregation_rule_id")
	}
	switch row["target_scope"] {
	case "vertex", "edge":
	default:
		if err := a.issue("invalid_aggregation_rule"); err != nil {
			return err
		}
	}
	switch row["input_scope"] {
	case "source_entity", "source_relationship", "projected_vertex", "projected_edge":
	default:
		if err := a.issue("invalid_aggregation_rule"); err != nil {
			return err
		}
	}
	for _, key := range []string{"input_kind", "projected_kind"} {
		if !validIdentifierV2(row[key].(string)) {
			return invalidProjectionErrorV2("invalid_identifier", path+"."+key)
		}
	}
	if err := scalarStringsV2(row, "grouping_keys", path, "invalid_field_path", validFieldPath); err != nil {
		return err
	}
	switch row["missing_grouping_key_behavior"] {
	case "error", "exclude", "use_null":
	default:
		if err := a.issue("invalid_aggregation_rule"); err != nil {
			return err
		}
	}
	merge := row["property_merge_behavior"].(map[string]any)
	for _, key := range slices.Sorted(maps.Keys(merge)) {
		if !validPropertyKey(key) {
			return invalidProjectionErrorV2("invalid_property_key", canonicalInputMemberPath(path+".property_merge_behavior", key))
		}
		if !validMergeNameV2(merge[key].(string)) {
			if err := a.issue("invalid_aggregation_rule"); err != nil {
				return err
			}
		}
	}
	if direction, ok := row["edge_direction"]; ok {
		switch direction {
		case "directed", "undirected", "bidirectional":
		default:
			if err := a.issue("invalid_aggregation_rule"); err != nil {
				return err
			}
		}
	}
	if endpoint, ok := row["endpoint_grouping"].(map[string]any); ok {
		for _, side := range []string{"src", "dst"} {
			key := side + "_vertex_aggregation_rule_id"
			if !validIdentifierV2(endpoint[key].(string)) {
				if err := a.issue("invalid_aggregation_rule"); err != nil {
					return err
				}
			}
			if err := scalarStringsV2(endpoint, side+"_grouping_keys", path+".endpoint_grouping", "invalid_field_path", validFieldPath); err != nil {
				return err
			}
		}
		if endpoint["missing_endpoint_behavior"] != "error" && endpoint["missing_endpoint_behavior"] != "exclude" {
			if err := a.issue("invalid_aggregation_rule"); err != nil {
				return err
			}
		}
	}
	return nil
}
