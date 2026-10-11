// Package harnessdiag owns native harness diagnostic validation and observation.
// Its generated validators project the same foundation schemas as Node readers.
package harnessdiag

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

var errArtifact = errors.New("invalid diagnostic artifact")
var schemaPatterns sync.Map

// Decode applies the same interoperable-number boundary as the harness parser.
func Decode(data []byte) (any, error) {
	v, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return nil, errArtifact
	}
	return interoperable(v)
}
func interoperable(v any) (any, error) {
	switch x := v.(type) {
	case json.Number:
		n, err := x.Float64()
		if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || (n == 0 && math.Signbit(n)) || (math.Trunc(n) == n && math.Abs(n) > 9007199254740991) {
			return nil, errArtifact
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
		return errArtifact
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
func schemaPattern(pattern, value string) bool {
	// The authored relative-reference pattern has negative lookaheads. Project
	// those predicates explicitly; all other authored patterns use RE2 syntax.
	const relative = `^(?!/)(?!.*(?:^|/)\.\.(?:/|$))(?!.*(?:^|/)\.(?:/|$))(?!.*//)[A-Za-z0-9._@+,:=-]+(?:/[A-Za-z0-9._@+,:=-]+)*$`
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
