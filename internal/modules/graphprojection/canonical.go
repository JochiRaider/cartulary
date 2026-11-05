package graphprojection

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
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

type canonicalSink interface {
	WriteString(string) (int, error)
	WriteByte(byte) error
}

// Buffering, hashing and byte accounting use one canonical transcript.
type canonicalWriter struct {
	out canonicalSink
	err error
}

func (w *canonicalWriter) text(s string) {
	if w.err == nil {
		_, w.err = w.out.WriteString(s)
	}
}
func (w *canonicalWriter) byte(b byte) {
	if w.err == nil {
		w.err = w.out.WriteByte(b)
	}
}
func (w *canonicalWriter) quoted(s string) {
	if w.err == nil {
		w.text(canonicalJSONString(s))
	}
}
func writeCanonicalJSON(out canonicalSink, value any) error {
	w := canonicalWriter{out: out}
	return w.value(value)
}
func (w *canonicalWriter) value(value any) error {
	if w.err != nil {
		return w.err
	}
	switch typed := value.(type) {
	case nil:
		w.text("null")
	case bool:
		if typed {
			w.text("true")
		} else {
			w.text("false")
		}
	case string:
		w.quoted(typed)
	case json.Number:
		normalized, err := normalizedNumberV2(typed)
		if err != nil {
			return fmt.Errorf("invalid canonical number")
		}
		return w.value(normalized)
	case int:
		w.text(strconv.Itoa(typed))
	case int64:
		w.text(strconv.FormatInt(typed, 10))
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return fmt.Errorf("non-finite canonical number")
		}
		formatted := strconv.FormatFloat(typed, 'g', -1, 64)
		formatted = strings.Replace(formatted, "e+", "e", 1)
		w.text(formatted)
	case []any:
		w.byte('[')
		for i, entry := range typed {
			if w.err != nil {
				return w.err
			}
			if i > 0 {
				w.byte(',')
			}
			if err := w.value(entry); err != nil {
				return err
			}
		}
		w.byte(']')
	case []string:
		w.byte('[')
		for i, entry := range typed {
			if w.err != nil {
				return w.err
			}
			if i > 0 {
				w.byte(',')
			}
			w.quoted(entry)
		}
		w.byte(']')
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		w.byte('{')
		for i, key := range keys {
			if w.err != nil {
				return w.err
			}
			if i > 0 {
				w.byte(',')
			}
			w.quoted(key)
			w.byte(':')
			if err := w.value(typed[key]); err != nil {
				return err
			}
		}
		w.byte('}')
	case map[string]string:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		w.byte('{')
		for i, key := range keys {
			if w.err != nil {
				return w.err
			}
			if i > 0 {
				w.byte(',')
			}
			w.quoted(key)
			w.byte(':')
			w.quoted(typed[key])
		}
		w.byte('}')
	default:
		return fmt.Errorf("unsupported canonical JSON type %T", typed)
	}
	return w.err
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
