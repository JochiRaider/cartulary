package dataset

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"golang.org/x/text/unicode/norm"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
var typePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._-]{0,63}$`)

func closed(v any, names ...string) (map[string]any, bool) {
	m, ok := v.(map[string]any)
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

func integer(v any, min, max int64) (int64, bool) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, false
	}
	s := n.String()
	if strings.ContainsAny(s, ".eE+") || s == "-0" {
		return 0, false
	}
	x, err := strconv.ParseInt(s, 10, 64)
	return x, err == nil && x >= min && x <= max
}

func safeNumbers(v any) bool {
	switch n := v.(type) {
	case json.Number:
		_, ok := integer(n, -MaxSafeInteger, MaxSafeInteger)
		return ok
	case map[string]any:
		for _, child := range n {
			if !safeNumbers(child) {
				return false
			}
		}
	case []any:
		for _, child := range n {
			if !safeNumbers(child) {
				return false
			}
		}
	}
	return true
}

func document(raw []byte) (map[string]any, error) {
	v, err := canonicaljson.DecodeStrictBounded(raw, MaxDepth)
	if err != nil {
		if err == canonicaljson.ErrDepth {
			return nil, excess("max_dataset_json_depth", MaxDepth, MaxDepth+1, true)
		}
		return nil, invalid()
	}
	m, ok := v.(map[string]any)
	if !ok || !safeNumbers(v) {
		return nil, invalid()
	}
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil || !bytes.Equal(raw, canonical) {
		return nil, invalid()
	}
	return m, nil
}

func DecodeDescriptor(raw []byte) (Descriptor, error) {
	return decodeDescriptor(raw, true)
}

func decodeDescriptor(raw []byte, semantics bool) (Descriptor, error) {
	var d Descriptor
	if len(raw) > MaxDescriptorBytes {
		return d, excess("max_dataset_descriptor_bytes", MaxDescriptorBytes, uint64(len(raw)), true)
	}
	if err := decodeTyped(raw, "descriptor", &d); err != nil {
		return Descriptor{}, err
	}
	if err := descriptorRules(d, false, firstFailure("content_schema_invalid")); err != nil {
		return Descriptor{}, err
	}
	if semantics {
		if err := descriptorRules(d, true, firstFailure("content_semantic_invalid")); err != nil {
			return Descriptor{}, err
		}
	}
	return d, nil
}

// Normalize returns the injectively framed canonical key for an admitted
// descriptor. It distinguishes an unsupported name from malformed values.
func Normalize(d Descriptor, name string, raw []byte) (string, error) {
	for _, lookup := range d.Lookups {
		if lookup.Name != name {
			continue
		}
		if len(raw) > MaxRowBytes {
			return "", ErrInvalidInput
		}
		v, err := canonicaljson.DecodeStrictBounded(raw, 2)
		if err != nil {
			return "", ErrInvalidInput
		}
		key, err := normalize(lookup, v)
		if err != nil {
			return "", ErrInvalidInput
		}
		return key, nil
	}
	return "", ErrUnsupportedLookup
}

func normalize(lookup Lookup, v any) (string, error) {
	return normalizeBudget(lookup, v, MaxTotalKeyBytes)
}

func normalizeBudget(lookup Lookup, v any, remaining int64) (string, error) {
	values := make([]any, len(lookup.Components))
	if len(values) == 0 || len(values) > MaxComponents {
		return "", invalid()
	}
	if lookup.InputType == "scalar" {
		if len(values) != 1 {
			return "", invalid()
		}
		values[0] = v
	} else if lookup.InputType == "composite" {
		m, ok := v.(map[string]any)
		if !ok || len(m) != len(values) {
			return "", invalid()
		}
		for i, c := range lookup.Components {
			value, ok := m[c.Name]
			if !ok {
				return "", invalid()
			}
			values[i] = value
		}
	} else {
		return "", invalid()
	}
	// Measure the exact framed key before allocating its encoded representation.
	size := 2 + len(values) - 1
	for i, c := range lookup.Components {
		if c.Type == "integer" && c.Normalization == "safe_integer" {
			x, ok := integer(values[i], -MaxSafeInteger, MaxSafeInteger)
			if !ok {
				return "", invalid()
			}
			values[i] = x
			size += len(strconv.FormatInt(x, 10))
		} else if c.Type == "string" {
			s, ok := values[i].(string)
			if !ok || s == "" || !utf8.ValidString(s) || utf8.RuneCountInString(s) > 8192 || !lookupString(s) {
				return "", invalid()
			}
			switch c.Normalization {
			case "exact_nfc":
				s = norm.NFC.String(s)
			case "ascii_case_insensitive":
				s = asciiLower(norm.NFC.String(s))
			case "uuid":
				if !uuid(s) {
					return "", invalid()
				}
				s = asciiLower(s)
			default:
				return "", invalid()
			}
			values[i] = s
			size += 2 + len(s) + strings.Count(s, `"`) + strings.Count(s, `\`)
		} else {
			return "", invalid()
		}
		if size > MaxKeyBytes {
			return "", excess("max_dataset_key_bytes", MaxKeyBytes, uint64(size), false)
		}
	}
	if int64(size) > remaining {
		return "", excess("max_dataset_total_key_bytes", MaxTotalKeyBytes, uint64(MaxTotalKeyBytes-remaining)+uint64(size), false)
	}
	var out strings.Builder
	out.Grow(size)
	out.WriteByte('[')
	for i, value := range values {
		if i > 0 {
			out.WriteByte(',')
		}
		switch value := value.(type) {
		case int64:
			out.WriteString(strconv.FormatInt(value, 10))
		case string:
			out.WriteByte('"')
			for _, b := range []byte(value) {
				if b == '"' || b == '\\' {
					out.WriteByte('\\')
				}
				out.WriteByte(b)
			}
			out.WriteByte('"')
		}
	}
	out.WriteByte(']')
	return out.String(), nil
}

func lookupString(s string) bool {
	for _, r := range s {
		if r < 32 || r >= 127 && r <= 159 || r == 0x2028 || r == 0x2029 {
			return false
		}
	}
	return true
}
func asciiLower(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + 'a' - 'A'
		}
		return r
	}, s)
}
func uuid(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, b := range []byte(s) {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if b != '-' {
				return false
			}
			continue
		}
		if !(b >= '0' && b <= '9' || b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F') {
			return false
		}
	}
	return true
}
