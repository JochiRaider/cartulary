package extensions

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

type InactiveKeyPolicy struct {
	Key    string
	Policy string
	Schema map[string]any
}

// ValidateInactiveConfiguration returns diagnostics only. It deliberately has no
// normalized-value result and accepts no resolver, filesystem, network, secret,
// view, or profile callback, so accepted syntax_only values cannot survive this
// boundary or cause effects.
func ValidateInactiveConfiguration(values map[string]any, policies []InactiveKeyPolicy) []Finding {
	byKey := make(map[string]InactiveKeyPolicy, len(policies))
	for _, policy := range policies {
		byKey[policy.Key] = policy
	}
	findings := []Finding{}
	for key, value := range values {
		policy, ok := byKey[key]
		if !ok || policy.Policy == "forbidden" {
			findings = append(findings, Finding{Code: "extension_config_without_claim", Phase: "claim_configuration", Actual: "redacted"})
			continue
		}
		if policy.Policy != "syntax_only" || value == nil || validateInertValue(policy.Schema, value, 0) != nil {
			findings = append(findings, Finding{Code: "extension_validation_result_invalid", Phase: "claim_configuration", Actual: "redacted"})
		}
	}
	sortFindings(findings)
	return findings
}

func validateInertValue(schema map[string]any, value any, depth int) error {
	if depth > 16 {
		return fmt.Errorf("nesting depth exceeded")
	}
	typeName, _ := schema["type"].(string)
	switch typeName {
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("wrong type")
		}
		if err := inertCountBounds(schema, len(object), "Properties"); err != nil {
			return err
		}
		properties, _ := schema["properties"].(map[string]any)
		required := map[string]struct{}{}
		if rows, ok := schema["required"].([]any); ok {
			for _, row := range rows {
				name, ok := row.(string)
				if !ok {
					return fmt.Errorf("invalid required declaration")
				}
				required[name] = struct{}{}
			}
		}
		for name := range required {
			if _, ok := object[name]; !ok {
				return fmt.Errorf("missing required member")
			}
		}
		additional, _ := schema["additionalProperties"].(bool)
		for name, member := range object {
			rawMemberSchema, declared := properties[name]
			if !declared {
				if !additional {
					return fmt.Errorf("unknown member")
				}
				continue
			}
			memberSchema, ok := rawMemberSchema.(map[string]any)
			if !ok || validateInertValue(memberSchema, member, depth+1) != nil {
				return fmt.Errorf("invalid member")
			}
		}
		return nil
	case "array":
		array, ok := value.([]any)
		if !ok {
			return fmt.Errorf("wrong type")
		}
		if err := inertCountBounds(schema, len(array), "Items"); err != nil {
			return err
		}
		itemSchema, _ := schema["items"].(map[string]any)
		for _, item := range array {
			if itemSchema == nil || validateInertValue(itemSchema, item, depth+1) != nil {
				return fmt.Errorf("invalid item")
			}
		}
		return nil
	case "string":
		text, ok := value.(string)
		if !ok || !utf8.ValidString(text) {
			return fmt.Errorf("wrong type")
		}
		if err := inertCountBounds(schema, len([]byte(text)), "Length"); err != nil {
			return err
		}
		if pattern, ok := schema["pattern"].(string); ok {
			compiled, err := regexp.Compile(pattern)
			if err != nil || !compiled.MatchString(text) {
				return fmt.Errorf("pattern mismatch")
			}
		}
		return inertEnum(schema, value)
	case "integer":
		integer, ok := inactiveIntegerValue(value)
		if !ok {
			return fmt.Errorf("wrong type")
		}
		if minimum, ok := inactiveIntegerValue(schema["minimum"]); ok && integer < minimum {
			return fmt.Errorf("below minimum")
		}
		if maximum, ok := inactiveIntegerValue(schema["maximum"]); ok && integer > maximum {
			return fmt.Errorf("above maximum")
		}
		return inertEnum(schema, value)
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("wrong type")
		}
		return inertEnum(schema, value)
	default:
		return fmt.Errorf("unsupported inert type")
	}
}

func inertCountBounds(schema map[string]any, count int, suffix string) error {
	if minimum, ok := inactiveIntegerValue(schema["min"+suffix]); ok && int64(count) < minimum {
		return fmt.Errorf("below minimum")
	}
	if maximum, ok := inactiveIntegerValue(schema["max"+suffix]); ok && int64(count) > maximum {
		return fmt.Errorf("above maximum")
	}
	return nil
}

func inertEnum(schema map[string]any, value any) error {
	rows, ok := schema["enum"].([]any)
	if !ok {
		return nil
	}
	for _, row := range rows {
		if fmt.Sprint(row) == fmt.Sprint(value) {
			return nil
		}
	}
	return fmt.Errorf("enum mismatch")
}

func inactiveIntegerValue(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	case interface{ Int64() (int64, error) }:
		integer, err := typed.Int64()
		return integer, err == nil
	default:
		return 0, false
	}
}
