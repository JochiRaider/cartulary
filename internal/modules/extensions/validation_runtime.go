package extensions

import "errors"

type ValidationCondition struct {
	ConditionID       string
	Phase             string
	ConditionClass    string
	PathAlgorithmID   string
	ReasonCode        string
	ExpectedFormatter string
	ActualFormatter   string
	Multiplicity      string
	SecretPolicy      string
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
			ConditionID:       stringValue(row["condition_id"]),
			Phase:             stringValue(row["phase"]),
			ConditionClass:    stringValue(row["condition_class"]),
			PathAlgorithmID:   stringValue(row["path_algorithm_id"]),
			ReasonCode:        stringValue(row["reason_code"]),
			ExpectedFormatter: stringValue(row["expected_formatter_id"]),
			ActualFormatter:   stringValue(row["actual_formatter_id"]),
			Multiplicity:      stringValue(row["multiplicity"]),
			SecretPolicy:      stringValue(row["secret_policy"]),
		}
		orderKey := condition.Phase + "\x00" + condition.ConditionID
		if condition.ConditionID == "" || condition.Phase == "" || orderKey <= previous {
			return nil, errors.New("condition identity is empty, duplicate, or unsorted")
		}
		if _, exists := result[condition.ConditionID]; exists {
			return nil, errors.New("condition identity is duplicate")
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
