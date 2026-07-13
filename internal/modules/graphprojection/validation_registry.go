package graphprojection

import (
	"fmt"
	"sort"
	"unicode/utf8"
)

type validationIssueContract struct {
	severity string
	target   string
	required []string
	optional []string
}

var validationIssueContracts = map[string]validationIssueContract{
	"invalid_input_shape":                 {"fatal", "projection_input", []string{"field", "reason_code"}, nil},
	"duplicate_identifier":                {"fatal", "projection_input", []string{"identifier_value", "collection"}, nil},
	"invalid_projection_schema":           {"fatal", "projection_input", []string{"field", "supplied_value"}, nil},
	"invalid_graph_view_id":               {"fatal", "graph_view", []string{"supplied_value", "expected_value"}, nil},
	"invalid_projection_config":           {"fatal", "projection_config", []string{"field", "reason_code"}, nil},
	"invalid_field_path":                  {"fatal", "projection_config", []string{"field_path", "scope"}, nil},
	"invalid_filter":                      {"fatal", "filter", []string{"field", "reason_code"}, nil},
	"invalid_mapping_rule":                {"fatal", "mapping_rule", []string{"mapping_rule_id", "reason_code"}, nil},
	"missing_entity_mapping_rule":         {"fatal", "mapping_rule", []string{"source_entity_kind"}, nil},
	"missing_relationship_mapping_rule":   {"error", "mapping_rule", []string{"source_relationship_kind"}, nil},
	"invalid_metadata_mapping":            {"fatal", "mapping_rule", []string{"metadata_mapping_id", "reason_code"}, nil},
	"invalid_property_definition":         {"fatal", "property_definition", []string{"property_definition_id", "reason_code"}, nil},
	"invalid_property_type":               {"error", "property", []string{"projected_key", "expected_type", "actual_type", "source_field_path", "output_object_id"}, []string{"aggregation_rule_id", "canonical_grouping_key_digest", "contributor_id"}},
	"required_property_missing":           {"error", "property", []string{"projected_key", "source_field_path", "output_object_id"}, []string{"aggregation_rule_id", "canonical_grouping_key_digest"}},
	"source_null_for_required_property":   {"error", "property", []string{"projected_key", "source_field_path", "output_object_id"}, []string{"aggregation_rule_id", "canonical_grouping_key_digest", "contributor_id"}},
	"undeclared_source_kind":              {"error", "source_item", []string{"source_item_id", "source_kind"}, nil},
	"source_item_resource_limit_exceeded": {"error", "source_item", []string{"source_item_id", "limit_key", "limit", "observed"}, nil},
	"missing_relationship_endpoint":       {"error", "source_relationship", []string{"source_relationship_id", "endpoint_field"}, nil},
	"relationship_endpoint_not_projected": {"error", "source_relationship", []string{"source_relationship_id", "endpoint_field", "endpoint_source_entity_id"}, nil},
	"invalid_relationship_direction":      {"error", "source_relationship", []string{"source_relationship_id", "supplied_value"}, nil},
	"invalid_direction_policy":            {"fatal", "mapping_rule", []string{"mapping_rule_id", "supplied_value"}, nil},
	"invalid_reverse_edge_policy":         {"fatal", "mapping_rule", []string{"mapping_rule_id", "projected_direction"}, nil},
	"invalid_aggregation_rule":            {"fatal", "mapping_rule", []string{"aggregation_rule_id", "reason_code"}, nil},
	"aggregation_grouping_key_missing":    {"error", "source_or_projected_item", []string{"aggregation_rule_id", "field_path", "contributor_id"}, nil},
	"aggregation_endpoint_missing":        {"error", "mapping_rule", []string{"aggregation_rule_id", "endpoint_side", "reason_code", "endpoint_digest", "field_path"}, nil},
	"aggregation_merge_conflict":          {"error", "mapping_rule", []string{"aggregation_rule_id", "canonical_grouping_key_digest", "projected_key"}, nil},
	"resource_limit_exceeded":             {"fatal", "projection_input", []string{"limit_key", "limit", "observed"}, nil},
	"projected_output_limit_exceeded":     {"fatal", "graph_view", []string{"limit_key", "limit", "observed"}, nil},
	"validation_issue_limit_exceeded":     {"fatal", "projection_input", []string{"limit"}, nil},
	"invalid_retention_policy":            {"fatal", "projection_config", []string{"field", "reason_code"}, nil},
	"output_schema_violation":             {"fatal", "graph_view", []string{"field", "reason_code"}, nil},
	"projection_computation_failed":       {"fatal", "graph_view", []string{"reason_code"}, nil},
}

var validationIssueReasonCodes = map[string][]string{
	"invalid_input_shape":           {"scalar_contract_violation", "property_value_too_long", "array_element_invalid", "array_length_exceeded", "invalid_label"},
	"invalid_projection_config":     {"custom_config_referenced", "relationship_mapping_source_conflict", "empty_kind_registry_not_allowed", "declared_kind_duplicate", "mapping_rule_duplicate", "metadata_mapping_duplicate", "aggregation_rule_duplicate", "invalid_default_materialization"},
	"invalid_filter":                {"invalid_operator", "value_required", "value_forbidden", "invalid_field_scope", "invalid_value_shape", "unsupported_logic"},
	"invalid_mapping_rule":          {"duplicate_mapping_rule_id", "duplicate_source_entity_kind_mapping", "duplicate_source_relationship_kind_mapping", "declared_source_kind_missing", "property_key_not_defined", "property_requiredness_mismatch", "required_optional_overlap", "reverse_edge_kind_without_reverse", "label_invalid"},
	"invalid_metadata_mapping":      {"reserved_metadata_key", "duplicate_after_wildcard_expansion", "invalid_source_scope", "invalid_default_value", "invalid_merge_behavior_type", "invalid_projected_type", "required_metadata_missing"},
	"invalid_property_definition":   {"duplicate_after_wildcard_expansion", "invalid_source_scope", "invalid_default_value", "invalid_null_policy", "invalid_merge_behavior_type", "invalid_projected_type"},
	"invalid_aggregation_rule":      {"dependency_on_later_rule", "aggregation_cycle", "endpoint_rule_not_vertex_rule", "endpoint_grouping_key_count_mismatch", "endpoint_grouping_key_invalid", "endpoint_field_scope_invalid", "grouping_key_invalid", "invalid_endpoint_behavior", "invalid_edge_direction", "input_scope_invalid", "invalid_merge_behavior_type"},
	"aggregation_endpoint_missing":  {"endpoint_key_missing", "endpoint_vertex_not_found"},
	"invalid_retention_policy":      {"out_of_bounds", "invalid_type"},
	"output_schema_violation":       {"id_mismatch", "reference_missing", "sort_order_invalid", "schema_registry_mismatch", "metadata_shape_invalid", "closed_schema_violation", "canonical_serialization_invalid"},
	"projection_computation_failed": {"internal_exception", "dependency_unavailable", "timeout", "resource_exhausted", "implementation_invariant_failed"},
}

func (run ProjectionRun) issue(severity, code, targetKind, targetID string, field any, details map[string]any) ValidationIssue {
	contract, ok := validationIssueContracts[code]
	if !ok || contract.severity != severity || contract.target != targetKind || !validIssueDetails(code, contract, details) {
		return run.computationFailureIssue("implementation_invariant_failed")
	}
	var fieldPtr *string
	if fieldString, ok := field.(string); ok && fieldString != "" {
		fieldPtr = &fieldString
	}
	identityDetails := make(map[string]any, len(contract.required))
	for _, key := range contract.required {
		identityDetails[key] = details[key]
	}
	issueID, err := generatedID("gpi_", "GPISSUE1\n", ProjectionSchemaID, run.GraphViewID, run.ProjectionRunID, severity, code, targetKind, targetID, identityDetails)
	if err != nil {
		return run.computationFailureIssue("implementation_invariant_failed")
	}
	return ValidationIssue{IssueID: issueID, Severity: severity, Code: code, TargetKind: targetKind, TargetID: targetID, Field: fieldPtr, Message: truncateScalars(code, graphProjectionLimits.MaxValidationMessageLength), Details: details}
}

func (run ProjectionRun) computationFailureIssue(reason string) ValidationIssue {
	details := map[string]any{"reason_code": reason}
	issueID, _ := generatedID("gpi_", "GPISSUE1\n", ProjectionSchemaID, run.GraphViewID, run.ProjectionRunID, "fatal", "projection_computation_failed", "graph_view", run.GraphViewID, details)
	return ValidationIssue{IssueID: issueID, Severity: "fatal", Code: "projection_computation_failed", TargetKind: "graph_view", TargetID: run.GraphViewID, Message: "projection_computation_failed", Details: details}
}

func validIssueDetails(code string, contract validationIssueContract, details map[string]any) bool {
	if details == nil {
		return len(contract.required) == 0
	}
	allowed := make(map[string]bool, len(contract.required)+len(contract.optional))
	for _, key := range contract.required {
		allowed[key] = true
		if _, ok := details[key]; !ok {
			return false
		}
	}
	for _, key := range contract.optional {
		allowed[key] = true
	}
	for key := range details {
		if !allowed[key] {
			return false
		}
	}
	if reasons := validationIssueReasonCodes[code]; len(reasons) > 0 {
		reason, ok := details["reason_code"].(string)
		if !ok || !containsRegistryValue(reasons, reason) {
			return false
		}
	}
	return true
}

func containsRegistryValue(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
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

func truncateScalars(value string, limit int) string {
	if limit < 0 || utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
}
