// Package harnessschema owns strict JSON and generated harness schema validation.
// Its generated validators project the same foundation schemas as Node readers.
package harnessschema

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

var ErrInvalid = errors.New("invalid harness document")
var schemaPatterns sync.Map

// Decode applies the same interoperable-number boundary as the harness parser.
func Decode(data []byte) (any, error) {
	v, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return nil, ErrInvalid
	}
	return interoperable(v)
}
func interoperable(v any) (any, error) {
	switch x := v.(type) {
	case json.Number:
		n, err := x.Float64()
		if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || (n == 0 && math.Signbit(n)) || (math.Trunc(n) == n && math.Abs(n) > 9007199254740991) {
			return nil, ErrInvalid
		}
		return n, nil
	case map[string]any:
		for k, child := range x {
			var err error
			x[k], err = interoperable(child)
			if err != nil {
				return nil, err
			}
		}
	case []any:
		for i, child := range x {
			var err error
			x[i], err = interoperable(child)
			if err != nil {
				return nil, err
			}
		}
	}
	return v, nil
}
func Validate(id string, value any) error {
	if !validSchema(id, value) {
		return ErrInvalid
	}
	return nil
}
func schemaType(v any, name string) bool {
	switch name {
	case "null":
		return v == nil
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := v.([]any)
		return ok
	case "string":
		_, ok := v.(string)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "number":
		n, ok := v.(float64)
		return ok && !math.IsNaN(n) && !math.IsInf(n, 0)
	case "integer":
		n, ok := v.(float64)
		return ok && !math.IsNaN(n) && !math.IsInf(n, 0) && math.Trunc(n) == n
	default:
		return false
	}
}
func schemaLength(s string) int { return utf8.RuneCountInString(s) }
func schemaToken(value, prefix string, minimum, maximum int, firstLow, firstHigh, restLow, restHigh uint64) bool {
	if !strings.HasPrefix(value, prefix) {
		return false
	}
	value = value[len(prefix):]
	if len(value) < minimum || maximum >= 0 && len(value) > maximum {
		return false
	}
	for i := 0; i < len(value); i++ {
		low, high := restLow, restHigh
		if i == 0 {
			low, high = firstLow, firstHigh
		}
		c := value[i]
		if c >= 128 || c < 64 && low&(uint64(1)<<c) == 0 || c >= 64 && high&(uint64(1)<<(c-64)) == 0 {
			return false
		}
	}
	return true
}
func schemaEqual(a, b any) bool { return reflect.DeepEqual(a, b) }
func schemaUnique(values []any) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		bytes, err := json.Marshal(value)
		if err != nil || seen[string(bytes)] {
			return false
		}
		seen[string(bytes)] = true
	}
	return true
}
func MatchPattern(pattern, value string) bool {
	// The authored relative-reference pattern has negative lookaheads. Project
	// those predicates explicitly; all other authored patterns use RE2 syntax.
	const relative = `^(?!/)(?!.*(?:^|/)\.\.(?:/|$))(?!.*(?:^|/)\.(?:/|$))(?!.*//)[A-Za-z0-9._@+,:=-]+(?:/[A-Za-z0-9._@+,:=-]+)*$`
	const legacyReference = `^(?!/)(?!.*(?:^|/)\.\.(?:/|$)).+$`
	const cachePath = `^(?!/)(?!.*\\)(?!.*(?:^|/)\.{1,2}(?:/|$))(?!.*//)[A-Za-z0-9._@+/-]+$`
	if pattern == cachePath {
		if strings.HasPrefix(value, "/") || strings.Contains(value, "//") {
			return false
		}
		for _, part := range strings.Split(value, "/") {
			if part == "." || part == ".." {
				return false
			}
		}
		pattern = `^[A-Za-z0-9._@+/-]+$`
	}
	const reservedPart = `(?!phase[0-9]+\.|fe_p[0-9]+\.)`
	const reservedLast = `(?!phase[0-9]+$|fe_p[0-9]+$)`
	if strings.Contains(pattern, reservedPart) || strings.Contains(pattern, reservedLast) {
		for _, part := range strings.Split(value, ".") {
			for _, prefix := range []string{"phase", "fe_p"} {
				if !strings.HasPrefix(part, prefix) {
					continue
				}
				suffix := strings.TrimPrefix(part, prefix)
				digits := suffix != ""
				for _, c := range suffix {
					if c < '0' || c > '9' {
						digits = false
					}
				}
				if digits {
					return false
				}
			}
		}
		pattern = strings.ReplaceAll(strings.ReplaceAll(pattern, reservedPart, ""), reservedLast, "")
	}
	const reviewReference = `^(?!/)(?!.*(?:^|/)\.\.(?:/|$))(?!.*[\\\u0000]).+$`
	const repositoryPath = `^(?!/)(?!.*(?:^|/)\.\.(?:/|$))[A-Za-z0-9._@+-]+(?:/[A-Za-z0-9._@+-]+)*$`
	const fixturePath = `^(?!/)(?!.*(?:^|/)\.\.(?:/|$))(?!.*//)[A-Za-z0-9._@+-]+(?:/[A-Za-z0-9._@+-]+)*$`
	if pattern == repositoryPath || pattern == fixturePath {
		for _, part := range strings.Split(value, "/") {
			if part == ".." {
				return false
			}
		}
		pattern = `^[A-Za-z0-9._@+-]+(?:/[A-Za-z0-9._@+-]+)*$`
	}
	if pattern == reviewReference && strings.ContainsAny(value, "\\\x00") {
		return false
	}
	if pattern == legacyReference || pattern == reviewReference {
		if value == "" || strings.HasPrefix(value, "/") || strings.ContainsAny(value, "\n\r\u2028\u2029") {
			return false
		}
		for _, part := range strings.Split(value, "/") {
			if part == ".." {
				return false
			}
		}
		return true
	}
	if pattern == relative {
		for _, part := range strings.Split(value, "/") {
			if part == "." || part == ".." {
				return false
			}
		}
		pattern = `^[A-Za-z0-9._@+,:=-]+(?:/[A-Za-z0-9._@+,:=-]+)*$`
	}
	expression, found := schemaPatterns.Load(pattern)
	if !found {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		expression, _ = schemaPatterns.LoadOrStore(pattern, compiled)
	}
	return expression.(*regexp.Regexp).MatchString(value)
}

func schemaPattern(pattern, value string) bool { return MatchPattern(pattern, value) }
