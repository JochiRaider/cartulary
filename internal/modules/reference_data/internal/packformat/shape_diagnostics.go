package packformat

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
)

var errShapeMismatch = errors.New("reference pack schema mismatch")

func diagnosticMember(path, name string) string {
	for i, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_' || i > 0 && r >= '0' && r <= '9') {
			encoded, _ := canonicaljson.Marshal(name)
			return path + "[" + string(encoded) + "]"
		}
	}
	return path + "." + name
}

func diagnosticKind(value any) string {
	switch value.(type) {
	case nil:
		return "null"
	case string:
		return "string"
	case bool:
		return "boolean"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	case json.Number, float64:
		return "integer"
	default:
		return "type"
	}
}

// walk is the common schema evaluator for Boolean admission and complete
// diagnostics. It visits each declared node once and never copies arbitrary
// input names or values into findings. The sink controls bounded retention.
func (s *shape) walk(value any, path string, emit FindingSink) error {
	report := func(expected, actual string) error {
		var a *string
		if actual != "" {
			a = &actual
		}
		return emit(Finding{Path: path, Details: SafeDetails{ExpectedToken: &expected, ActualToken: a}})
	}
	if s.hasConstant && !equalJSON(s.constant, value) {
		if err := report("const", ""); err != nil {
			return err
		}
	}
	if s.enum != nil {
		found := false
		for _, candidate := range s.enum {
			found = found || equalJSON(candidate, value)
		}
		if !found {
			if err := report("enum", ""); err != nil {
				return err
			}
		}
	}
	if s.not != nil && s.not.matches(value) {
		if err := report("not", ""); err != nil {
			return err
		}
	}
	if s.anyOf != nil {
		found := false
		for _, child := range s.anyOf {
			found = found || child.matches(value)
		}
		if !found {
			if err := report("type", diagnosticKind(value)); err != nil {
				return err
			}
		}
	}
	if s.oneOf != nil {
		count := 0
		for _, child := range s.oneOf {
			if child.matches(value) {
				count++
			}
		}
		if count != 1 {
			if err := report("type", diagnosticKind(value)); err != nil {
				return err
			}
		}
	}
	if s.nullable && value == nil {
		return nil
	}
	bound := func(actual int, min, max *int) error {
		if min != nil && actual < *min {
			return report(strconv.Itoa(*min), strconv.Itoa(actual))
		}
		if max != nil && actual > *max {
			if s.limitID != "" {
				return emit(LimitFinding(path, s.limitID, int64(*max), int64(actual)))
			}
			return report(strconv.Itoa(*max), strconv.Itoa(actual))
		}
		return nil
	}
	switch s.kind {
	case "":
		return nil
	case "null":
		if value != nil {
			return report("null", diagnosticKind(value))
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return report("boolean", diagnosticKind(value))
		}
	case "integer":
		v, ok := integer(value, -9007199254740991, 9007199254740991)
		if !ok {
			return report("integer", diagnosticKind(value))
		}
		if s.minimum != nil && float64(v) < *s.minimum {
			if err := report("minimum", ""); err != nil {
				return err
			}
		}
		if s.maximum != nil && float64(v) > *s.maximum {
			if err := report("maximum", ""); err != nil {
				return err
			}
		}
	case "string":
		v, ok := value.(string)
		if !ok {
			return report("string", diagnosticKind(value))
		}
		if !utf8.ValidString(v) {
			return report("string", "")
		}
		if err := bound(len(v), nil, s.maxUTF8Bytes); err != nil {
			return err
		}
		if err := bound(utf8.RuneCountInString(v), s.minLength, s.maxLength); err != nil {
			return err
		}
		if s.pattern != nil && !s.pattern.MatchString(v) {
			if err := report("pattern", ""); err != nil {
				return err
			}
		}
		switch s.format {
		case "uuid":
			id, err := uuid.Parse(v)
			if err != nil || id == uuid.Nil || id.String() != v {
				return report("type", "")
			}
		case "date-time":
			at, err := time.Parse(time.RFC3339Nano, v)
			if err != nil || at.Year() < 1 || at.UTC().Format(time.RFC3339Nano) != v {
				return report("type", "")
			}
		}
	case "array":
		v, ok := value.([]any)
		if !ok {
			return report("array", diagnosticKind(value))
		}
		if err := bound(len(v), s.minItems, s.maxItems); err != nil {
			return err
		}
		var seen map[string]bool
		if s.unique {
			seen = map[string]bool{}
		}
		for i, child := range v {
			childPath := path + "[" + strconv.Itoa(i) + "]"
			if s.items != nil {
				if err := s.items.walk(child, childPath, emit); err != nil {
					return err
				}
			}
			if s.unique {
				encoded, err := canonicaljson.Marshal(child)
				if err != nil {
					return err
				}
				if seen[string(encoded)] {
					token := "unique"
					if err := emit(Finding{Path: childPath, Details: SafeDetails{ExpectedToken: &token}}); err != nil {
						return err
					}
				}
				seen[string(encoded)] = true
			}
		}
	case "object":
		v, ok := object(value)
		if !ok {
			return report("object", diagnosticKind(value))
		}
		if err := bound(len(v), s.minProperties, s.maxProperties); err != nil {
			return err
		}
		for _, name := range s.required {
			if _, exists := v[name]; !exists {
				expected, actual := "required", "missing"
				if err := emit(Finding{Path: diagnosticMember(path, name), Details: SafeDetails{ExpectedToken: &expected, ActualToken: &actual}}); err != nil {
					return err
				}
			}
		}
		for name, child := range v {
			rule, known := s.properties[name]
			childPath := path + `["<unknown>"]`
			if known {
				childPath = diagnosticMember(path, name)
			}
			if s.propertyNames != nil && !s.propertyNames.matches(name) {
				token := "pattern"
				if err := emit(Finding{Path: childPath, Details: SafeDetails{ExpectedToken: &token}}); err != nil {
					return err
				}
			}
			if !known {
				if s.closed {
					token := "unknown_member"
					if err := emit(Finding{Path: childPath, Details: SafeDetails{ActualToken: &token}}); err != nil {
						return err
					}
				}
				rule = s.additional
			}
			if rule != nil {
				if err := rule.walk(child, childPath, emit); err != nil {
					return err
				}
			}
		}
	default:
		panic("unsupported reference pack schema type: " + s.kind)
	}
	return nil
}
