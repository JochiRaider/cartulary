package graphprojection

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

func canonicalJSON(value any) ([]byte, error) {
	var out bytes.Buffer
	if err := writeCanonicalJSON(&out, value); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func canonicalJSONString(value string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, r := range value {
		switch r {
		case '"':
			out.WriteString(`\"`)
		case '\\':
			out.WriteString(`\\`)
		case '\b':
			out.WriteString(`\b`)
		case '\t':
			out.WriteString(`\t`)
		case '\n':
			out.WriteString(`\n`)
		case '\f':
			out.WriteString(`\f`)
		case '\r':
			out.WriteString(`\r`)
		default:
			if r < 0x20 {
				out.WriteString(`\u00`)
				out.WriteString(hex.EncodeToString([]byte{byte(r)}))
				continue
			}
			out.WriteRune(r)
		}
	}
	out.WriteByte('"')
	return out.String()
}

func writeCanonicalJSON(out *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case nil:
		out.WriteString("null")
	case bool:
		if typed {
			out.WriteString("true")
		} else {
			out.WriteString("false")
		}
	case string:
		out.WriteString(canonicalJSONString(typed))
	case json.Number:
		if !finiteIntegerPattern.MatchString(typed.String()) {
			return fmt.Errorf("non-canonical number %q", typed.String())
		}
		out.WriteString(typed.String())
	case int:
		out.WriteString(strconv.Itoa(typed))
	case int64:
		out.WriteString(strconv.FormatInt(typed, 10))
	case []any:
		out.WriteByte('[')
		for i, entry := range typed {
			if i > 0 {
				out.WriteByte(',')
			}
			if err := writeCanonicalJSON(out, entry); err != nil {
				return err
			}
		}
		out.WriteByte(']')
	case []string:
		out.WriteByte('[')
		for i, entry := range typed {
			if i > 0 {
				out.WriteByte(',')
			}
			out.WriteString(canonicalJSONString(entry))
		}
		out.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out.WriteByte('{')
		for i, key := range keys {
			if i > 0 {
				out.WriteByte(',')
			}
			out.WriteString(canonicalJSONString(key))
			out.WriteByte(':')
			if err := writeCanonicalJSON(out, typed[key]); err != nil {
				return err
			}
		}
		out.WriteByte('}')
	case map[string]string:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out.WriteByte('{')
		for i, key := range keys {
			if i > 0 {
				out.WriteByte(',')
			}
			out.WriteString(canonicalJSONString(key))
			out.WriteByte(':')
			out.WriteString(canonicalJSONString(typed[key]))
		}
		out.WriteByte('}')
	default:
		return fmt.Errorf("unsupported canonical JSON type %T", typed)
	}
	return nil
}

func tupleBytes(prefix string, fields ...any) ([]byte, error) {
	var out bytes.Buffer
	out.WriteString(prefix)
	for _, field := range fields {
		fieldBytes, err := canonicalFieldBytes(field)
		if err != nil {
			return nil, err
		}
		out.WriteString(strconv.Itoa(len(fieldBytes)))
		out.WriteByte(':')
		out.Write(fieldBytes)
		out.WriteByte('\n')
	}
	return out.Bytes(), nil
}

func canonicalFieldBytes(value any) ([]byte, error) {
	switch typed := value.(type) {
	case string:
		if !utf8.ValidString(typed) {
			return nil, fmt.Errorf("invalid utf-8 string")
		}
		return []byte(typed), nil
	case []byte:
		return typed, nil
	default:
		return canonicalJSON(typed)
	}
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func digestTuple(prefix string, fields ...any) (string, error) {
	data, err := tupleBytes(prefix, fields...)
	if err != nil {
		return "", err
	}
	return sha256Hex(data), nil
}

func generatedID(idPrefix, tuplePrefix string, fields ...any) (string, error) {
	digest, err := digestTuple(tuplePrefix, fields...)
	if err != nil {
		return "", err
	}
	return idPrefix + digest, nil
}

func canonicalValueKey(value any) string {
	data, err := canonicalJSON(value)
	if err != nil {
		return fmt.Sprintf("%T:%v", value, value)
	}
	return string(data)
}
