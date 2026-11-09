package packformat

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
)

func object(value any) (map[string]any, bool) { v, ok := value.(map[string]any); return v, ok }
func exact(value any, names ...string) (map[string]any, bool) {
	m, ok := object(value)
	if !ok || len(m) != len(names) {
		return nil, false
	}
	for _, name := range names {
		if _, ok := m[name]; !ok {
			return nil, false
		}
	}
	return m, true
}
func integer(value any, min, max int64) (int64, bool) {
	n, ok := value.(json.Number)
	if !ok {
		return 0, false
	}
	text := n.String()
	if text == "" || strings.ContainsAny(text, ".eE+") || (len(text) > 1 && text[0] == '0') || text == "-0" {
		return 0, false
	}
	v, err := strconv.ParseInt(text, 10, 64)
	return v, err == nil && v >= min && v <= max
}
func hexBytes(value any, size int) ([]byte, bool) {
	s, ok := value.(string)
	if !ok || len(s) != size*2 || s != strings.ToLower(s) {
		return nil, false
	}
	b, err := hex.DecodeString(s)
	return b, err == nil
}
func validPackNumbers(value any) bool {
	switch v := value.(type) {
	case json.Number:
		_, ok := integer(v, -9007199254740991, 9007199254740991)
		return ok
	case map[string]any:
		for _, child := range v {
			if !validPackNumbers(child) {
				return false
			}
		}
	case []any:
		for _, child := range v {
			if !validPackNumbers(child) {
				return false
			}
		}
	}
	return true
}
