// Package extensioninactive validates and discards configuration values for
// extension profiles that are not claimed. It intentionally has no resolver,
// filesystem, network, secret, view, or profile callback.
package extensioninactive

import (
	"fmt"
	"regexp"
	"sort"
	"unicode/utf8"
)

type PolicyKind string

const (
	PolicyForbidden  PolicyKind = "forbidden"
	PolicySyntaxOnly PolicyKind = "syntax_only"
)

type Policy struct {
	ProfileID string
	ClaimKey  string
	Key       string
	Kind      PolicyKind
	Schema    map[string]any
}

type Finding struct {
	Key        string
	ReasonCode string
}

// Catalog is an immutable, read-only policy projection. Schemas are copied at
// construction and query boundaries so callers cannot change admitted policy.
type Catalog struct {
	byKey map[string]Policy
	keys  []string
}

func NewCatalog(policies []Policy) (Catalog, error) {
	catalog := Catalog{byKey: make(map[string]Policy, len(policies))}
	for _, policy := range policies {
		if policy.ProfileID == "" || policy.ClaimKey == "" || policy.Key == "" {
			return Catalog{}, fmt.Errorf("inactive extension policy identity is incomplete")
		}
		if policy.Kind != PolicyForbidden && policy.Kind != PolicySyntaxOnly {
			return Catalog{}, fmt.Errorf("inactive extension policy %s has unsupported kind %q", policy.Key, policy.Kind)
		}
		if policy.Kind == PolicyForbidden && policy.Schema != nil {
			return Catalog{}, fmt.Errorf("forbidden inactive extension policy %s has a schema", policy.Key)
		}
		if policy.Kind == PolicySyntaxOnly {
			if policy.Schema == nil {
				return Catalog{}, fmt.Errorf("syntax-only inactive extension policy %s has no schema", policy.Key)
			}
			if err := validateInertSchema(policy.Schema, 0); err != nil {
				return Catalog{}, fmt.Errorf("syntax-only inactive extension policy %s: %w", policy.Key, err)
			}
		}
		if _, duplicate := catalog.byKey[policy.Key]; duplicate {
			return Catalog{}, fmt.Errorf("duplicate inactive extension policy %s", policy.Key)
		}
		policy.Schema = cloneMap(policy.Schema)
		catalog.byKey[policy.Key] = policy
		catalog.keys = append(catalog.keys, policy.Key)
	}
	sort.Strings(catalog.keys)
	return catalog, nil
}

func (c Catalog) Policy(key string) (Policy, bool) {
	policy, ok := c.byKey[key]
	policy.Schema = cloneMap(policy.Schema)
	return policy, ok
}

func (c Catalog) Keys() []string {
	return append([]string(nil), c.keys...)
}

// ValidateAndDiscard returns redacted diagnostics only. It deliberately returns
// no normalized values, so successfully validated syntax-only values cannot
// cross this boundary.
func (c Catalog) ValidateAndDiscard(values map[string]any) []Finding {
	findings := make([]Finding, 0)
	for key, value := range values {
		policy, ok := c.byKey[key]
		if !ok || policy.Kind == PolicyForbidden {
			findings = append(findings, Finding{Key: key, ReasonCode: "extension_config_without_claim"})
			continue
		}
		if value == nil || validateInertValue(policy.Schema, value, 0) != nil {
			findings = append(findings, Finding{Key: key, ReasonCode: "extension_validation_result_invalid"})
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Key != findings[j].Key {
			return findings[i].Key < findings[j].Key
		}
		return findings[i].ReasonCode < findings[j].ReasonCode
	})
	return findings
}

func validateInertSchema(schema map[string]any, depth int) error {
	if depth > 16 {
		return fmt.Errorf("schema nesting depth exceeded")
	}
	allowed := map[string]struct{}{
		"additionalProperties": {}, "enum": {}, "items": {}, "maxItems": {},
		"maxLength": {}, "maxProperties": {}, "maximum": {}, "minItems": {},
		"minLength": {}, "minProperties": {}, "minimum": {}, "pattern": {},
		"properties": {}, "required": {}, "type": {},
	}
	for keyword := range schema {
		if _, ok := allowed[keyword]; !ok {
			return fmt.Errorf("unsupported inert schema keyword %q", keyword)
		}
	}
	typeName, _ := schema["type"].(string)
	switch typeName {
	case "object":
		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			return fmt.Errorf("object schema properties are missing")
		}
		for _, raw := range properties {
			member, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("object member schema is invalid")
			}
			if err := validateInertSchema(member, depth+1); err != nil {
				return err
			}
		}
	case "array":
		item, ok := schema["items"].(map[string]any)
		if !ok {
			return fmt.Errorf("array item schema is missing")
		}
		return validateInertSchema(item, depth+1)
	case "string":
		if pattern, ok := schema["pattern"].(string); ok {
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("invalid pattern")
			}
		}
	case "integer", "boolean":
	default:
		return fmt.Errorf("unsupported inert type")
	}
	return nil
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
		integer, ok := integerValue(value)
		if !ok {
			return fmt.Errorf("wrong type")
		}
		if minimum, ok := integerValue(schema["minimum"]); ok && integer < minimum {
			return fmt.Errorf("below minimum")
		}
		if maximum, ok := integerValue(schema["maximum"]); ok && integer > maximum {
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
	if minimum, ok := integerValue(schema["min"+suffix]); ok && int64(count) < minimum {
		return fmt.Errorf("below minimum")
	}
	if maximum, ok := integerValue(schema["max"+suffix]); ok && int64(count) > maximum {
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

func integerValue(value any) (int64, bool) {
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

func cloneMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	result := make(map[string]any, len(source))
	for key, value := range source {
		switch typed := value.(type) {
		case map[string]any:
			result[key] = cloneMap(typed)
		case []any:
			result[key] = cloneSlice(typed)
		default:
			result[key] = typed
		}
	}
	return result
}

func cloneSlice(source []any) []any {
	result := make([]any, len(source))
	for index, value := range source {
		switch typed := value.(type) {
		case map[string]any:
			result[index] = cloneMap(typed)
		case []any:
			result[index] = cloneSlice(typed)
		default:
			result[index] = typed
		}
	}
	return result
}
