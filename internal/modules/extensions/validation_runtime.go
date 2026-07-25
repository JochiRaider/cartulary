package extensions

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"unicode/utf8"
)

const (
	OwnerFindingSchemaLimit = 256
	OwnerFindingOverflowAt  = 4097
)

type ValidationCondition struct {
	ConditionID        string
	Phase              string
	ConditionClass     string
	PathAlgorithmID    string
	ReasonCode         string
	ExpectedFormatter  string
	ActualFormatter    string
	Multiplicity       string
	SecretPolicy       string
	OwnerRequirementID string
}

type OwnerFinding struct {
	Path       string
	ReasonCode string
	Message    string
	Details    map[string]any
}

type OwnerValidationDisposition string

const (
	OwnerValidationInvocationFailure OwnerValidationDisposition = "invocation_failure"
	OwnerValidationResultInvalid     OwnerValidationDisposition = "validation_result_invalid"
	OwnerValidationOverflow          OwnerValidationDisposition = "diagnostic_overflow"
	OwnerValidationFindings          OwnerValidationDisposition = "valid_findings"
	OwnerValidationSuccess           OwnerValidationDisposition = "valid_empty"
)

type OwnerValidationOutcome struct {
	Disposition OwnerValidationDisposition
	ReasonCode  string
	Findings    []OwnerFinding
	Limit       int
	Actual      int
}

// ClassifyOwnerValidationResult implements EXT-REQ-225 before any owner-local
// semantic interpretation. The caller supplies invocationErr separately so an
// invocation failure always wins over any bytes that happened to be returned.
func ClassifyOwnerValidationResult(invocationErr error, result []byte) OwnerValidationOutcome {
	if invocationErr != nil {
		return OwnerValidationOutcome{Disposition: OwnerValidationInvocationFailure, ReasonCode: "extension_admission_validation_failed"}
	}
	if !utf8.Valid(result) {
		return invalidOwnerValidationResult()
	}
	var raw any
	decoder := json.NewDecoder(bytes.NewReader(result))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return invalidOwnerValidationResult()
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return invalidOwnerValidationResult()
	}
	object, ok := raw.(map[string]any)
	if !ok || len(object) != 1 {
		return invalidOwnerValidationResult()
	}
	rawFindings, exists := object["findings"]
	findings, ok := rawFindings.([]any)
	if !exists || !ok {
		return invalidOwnerValidationResult()
	}
	if len(findings) >= OwnerFindingOverflowAt {
		return OwnerValidationOutcome{
			Disposition: OwnerValidationOverflow,
			ReasonCode:  "extension_diagnostic_overflow",
			Limit:       OwnerFindingOverflowAt - 1,
			Actual:      OwnerFindingOverflowAt,
		}
	}
	if len(findings) > OwnerFindingSchemaLimit {
		return invalidOwnerValidationResult()
	}
	normalized := make([]OwnerFinding, 0, len(findings))
	for _, rawFinding := range findings {
		finding, ok := normalizeOwnerFinding(rawFinding)
		if !ok {
			return invalidOwnerValidationResult()
		}
		normalized = append(normalized, finding)
	}
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].Path != normalized[j].Path {
			return normalized[i].Path < normalized[j].Path
		}
		if normalized[i].ReasonCode != normalized[j].ReasonCode {
			return normalized[i].ReasonCode < normalized[j].ReasonCode
		}
		left, _ := canonicalJSON(normalized[i].Details, false)
		right, _ := canonicalJSON(normalized[j].Details, false)
		return bytes.Compare(left, right) < 0
	})
	if len(normalized) == 0 {
		return OwnerValidationOutcome{Disposition: OwnerValidationSuccess}
	}
	return OwnerValidationOutcome{Disposition: OwnerValidationFindings, Findings: normalized}
}

func invalidOwnerValidationResult() OwnerValidationOutcome {
	return OwnerValidationOutcome{Disposition: OwnerValidationResultInvalid, ReasonCode: "extension_validation_result_invalid"}
}

func normalizeOwnerFinding(raw any) (OwnerFinding, bool) {
	object, ok := raw.(map[string]any)
	if !ok || len(object) != 4 {
		return OwnerFinding{}, false
	}
	path, pathOK := object["path"].(string)
	reasonCode, reasonOK := object["reason_code"].(string)
	message, messageOK := object["message"].(string)
	details, detailsOK := object["details"].(map[string]any)
	if !pathOK || !reasonOK || !messageOK || !detailsOK || path == "" || path[0] != '$' || reasonCode == "" || message == "" {
		return OwnerFinding{}, false
	}
	return OwnerFinding{Path: path, ReasonCode: reasonCode, Message: message, Details: cloneObject(details)}, true
}

func parseValidationConditions(registry map[string]any) (map[string]ValidationCondition, error) {
	if registry["schema_id"] != "cartulary.extension_validation_condition_registry.v2" {
		return nil, errors.New("unexpected schema_id")
	}
	rows, ok := objectSlice(registry["conditions"])
	if !ok || len(rows) == 0 || len(rows) > 16384 {
		return nil, errors.New("conditions must contain 1..16384 rows")
	}
	result := make(map[string]ValidationCondition, len(rows))
	previous := ""
	for _, row := range rows {
		condition := ValidationCondition{
			ConditionID:        stringValue(row["condition_id"]),
			Phase:              stringValue(row["phase"]),
			ConditionClass:     stringValue(row["condition_class"]),
			PathAlgorithmID:    stringValue(row["path_algorithm_id"]),
			ReasonCode:         stringValue(row["reason_code"]),
			ExpectedFormatter:  stringValue(row["expected_formatter_id"]),
			ActualFormatter:    stringValue(row["actual_formatter_id"]),
			Multiplicity:       stringValue(row["multiplicity"]),
			SecretPolicy:       stringValue(row["secret_policy"]),
			OwnerRequirementID: stringValue(row["owner_requirement_id"]),
		}
		orderKey := condition.ConditionID
		if orderKey == "" || orderKey <= previous {
			return nil, errors.New("condition identity is empty, duplicate, or unsorted")
		}
		previous = orderKey
		result[condition.ConditionID] = condition
	}
	return result, nil
}

// RequireRegisteredCondition fails closed without exposing a local validation
// library error or an invented condition.
func (c *Coordinator) RequireRegisteredCondition(conditionID string) (ValidationCondition, error) {
	condition, ok := c.ValidationCondition(conditionID)
	if !ok {
		return ValidationCondition{}, validationFailure(Finding{Code: "extension_validation_result_invalid", Phase: "profile_preflight"})
	}
	return condition, nil
}
