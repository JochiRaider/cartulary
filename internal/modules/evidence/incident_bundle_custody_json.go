package evidence

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// Custody JSON numbers are exact decimals. JCS/float64 would lose integers and
// fractions before either the fixed insert or the stored-state comparison.
func custodyJSON(raw any) (json.RawMessage, bool) {
	if _, ok := raw.(map[string]any); !ok {
		return nil, false
	}
	budget := incidentportability.MaxNDJSONLogicalLineBytes
	value, ok := normalizeCustodyJSON(raw, &budget)
	if !ok {
		return nil, false
	}
	encoded, err := json.Marshal(value)
	return encoded, err == nil && len(encoded) <= incidentportability.MaxNDJSONLogicalLineBytes
}
func custodyString(value string) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, '\x00')
}
func normalizeCustodyJSON(raw any, budget *int) (any, bool) {
	charge := func(n int) bool {
		if n > *budget {
			return false
		}
		*budget -= n
		return true
	}
	switch value := raw.(type) {
	case nil:
		return nil, charge(4)
	case bool:
		if value {
			return value, charge(4)
		}
		return value, charge(5)
	case string:
		if !custodyString(value) {
			return nil, false
		}
		encoded, err := json.Marshal(value)
		return value, err == nil && charge(len(encoded))
	case json.Number:
		normalized, ok := custodyDecimal(value.String(), *budget)
		if !ok {
			return nil, false
		}
		return json.Number(normalized), charge(len(normalized))
	case []any:
		if !charge(2 + max(0, len(value)-1)) {
			return nil, false
		}
		result := make([]any, 0, len(value))
		for _, child := range value {
			normalized, ok := normalizeCustodyJSON(child, budget)
			if !ok {
				return nil, false
			}
			result = append(result, normalized)
		}
		return result, true
	case map[string]any:
		if !charge(2 + max(0, 2*len(value)-1)) {
			return nil, false
		}
		result := make(map[string]any, len(value))
		for key, child := range value {
			if !custodyString(key) {
				return nil, false
			}
			encoded, err := json.Marshal(key)
			if err != nil || !charge(len(encoded)) {
				return nil, false
			}
			normalized, ok := normalizeCustodyJSON(child, budget)
			if !ok {
				return nil, false
			}
			result[key] = normalized
		}
		return result, true
	default:
		return nil, false
	}
}

func custodyDecimal(value string, budget int) (string, bool) {
	negative := strings.HasPrefix(value, "-")
	unsigned := strings.TrimPrefix(value, "-")
	mantissa, exponentText := unsigned, ""
	if index := strings.IndexAny(unsigned, "eE"); index >= 0 {
		mantissa, exponentText = unsigned[:index], unsigned[index+1:]
	}
	fractional := 0
	if index := strings.IndexByte(mantissa, '.'); index >= 0 {
		fractional = len(mantissa) - index - 1
	}
	digits := strings.ReplaceAll(mantissa, ".", "")
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		return "0", budget >= 1
	}
	exponent := int64(0)
	if exponentText != "" {
		parsed, err := strconv.ParseInt(exponentText, 10, 64)
		if err != nil {
			return "", false
		}
		exponent = parsed
	}
	// Bound arithmetic before expansion, including an exponent too large for the
	// original mantissa to cancel within one admitted logical row.
	limit := int64(len(value)) + 131072 + 16383
	if exponent > limit || exponent < -limit {
		return "", false
	}
	coefficient := strings.TrimRight(digits, "0")
	scale := int64(fractional) - exponent - int64(len(digits)-len(coefficient))
	integerDigits := int64(len(coefficient)) - scale
	if integerDigits > 131072 || scale > 16383 {
		return "", false
	}
	size := len(coefficient)
	if scale <= 0 {
		size += int(-scale)
	} else if integerDigits > 0 {
		size++
	} else {
		size = 2 + int(scale)
	}
	if negative {
		size++
	}
	if size > budget {
		return "", false
	}
	result := ""
	switch {
	case scale <= 0:
		result = coefficient + strings.Repeat("0", int(-scale))
	case integerDigits > 0:
		result = coefficient[:int(integerDigits)] + "." + coefficient[int(integerDigits):]
	default:
		result = "0." + strings.Repeat("0", int(-integerDigits)) + coefficient
	}
	if negative {
		result = "-" + result
	}
	return result, true
}
