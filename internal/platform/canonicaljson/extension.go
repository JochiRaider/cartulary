package canonicaljson

import (
	"bytes"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"unicode/utf8"
)

// MarshalExtension implements extension_registry_canonical_json_v1, including
// its final LF. It uses UTF-8 key ordering and exact integers, independently of
// the RFC-8785 encoding exposed by Marshal.
func MarshalExtension(value any) ([]byte, error) {
	projected, err := json.Marshal(value)
	if err != nil || !extensionStringsValid(reflect.ValueOf(value), 0) {
		return nil, ErrUnsupportedValue
	}
	return CanonicalizeExtension(projected)
}

// CanonicalizeExtension admits strict JSON and emits the Extensions byte form.
// Fractions, exponents, negative zero and invalid Unicode are rejected.
func CanonicalizeExtension(document []byte) ([]byte, error) {
	value, err := DecodeStrict(document)
	if err != nil {
		return nil, err
	}
	return encodeExtension(value)
}

// DecodeExtension admits only the exact canonical byte form, including one LF.
// It never repairs duplicate members, invalid Unicode or noncanonical bytes.
func DecodeExtension(document []byte) (any, error) {
	value, err := DecodeStrict(document)
	if err != nil {
		return nil, err
	}
	canonical, err := encodeExtension(value)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(document, canonical) {
		return nil, ErrInvalidJSON
	}
	return value, nil
}

func encodeExtension(value any) ([]byte, error) {
	var output bytes.Buffer
	if err := appendExtension(&output, value); err != nil {
		return nil, err
	}
	output.WriteByte('\n')
	return output.Bytes(), nil
}

func appendExtension(output *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case nil:
		output.WriteString("null")
	case bool:
		if typed {
			output.WriteString("true")
		} else {
			output.WriteString("false")
		}
	case string:
		return appendString(output, typed)
	case json.Number:
		digits := strings.TrimPrefix(string(typed), "-")
		if digits == "" || typed == "-0" || (len(digits) > 1 && digits[0] == '0') {
			return ErrUnsupportedValue
		}
		for _, character := range digits {
			if character < '0' || character > '9' {
				return ErrUnsupportedValue
			}
		}
		output.WriteString(string(typed))
	case []any:
		output.WriteByte('[')
		for i, member := range typed {
			if i > 0 {
				output.WriteByte(',')
			}
			if err := appendExtension(output, member); err != nil {
				return err
			}
		}
		output.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		output.WriteByte('{')
		for i, key := range keys {
			if i > 0 {
				output.WriteByte(',')
			}
			if err := appendString(output, key); err != nil {
				return err
			}
			output.WriteByte(':')
			if err := appendExtension(output, typed[key]); err != nil {
				return err
			}
		}
		output.WriteByte('}')
	default:
		return ErrUnsupportedValue
	}
	return nil
}

// encoding/json repairs invalid Go strings. Inspect their original values,
// including typed slices/maps/structs, before accepting its JSON projection.
// Marshal has already rejected ordinary cycles; the depth bound also covers
// custom marshalers whose underlying Go value is cyclic.
func extensionStringsValid(value reflect.Value, depth int) bool {
	if !value.IsValid() {
		return true
	}
	if depth > 10000 {
		return false
	}
	switch value.Kind() {
	case reflect.String:
		return utf8.ValidString(value.String())
	case reflect.Interface, reflect.Pointer:
		return value.IsNil() || extensionStringsValid(value.Elem(), depth+1)
	case reflect.Array, reflect.Slice:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			return true
		}
		for i := 0; i < value.Len(); i++ {
			if !extensionStringsValid(value.Index(i), depth+1) {
				return false
			}
		}
	case reflect.Map:
		iterator := value.MapRange()
		for iterator.Next() {
			if !extensionStringsValid(iterator.Key(), depth+1) || !extensionStringsValid(iterator.Value(), depth+1) {
				return false
			}
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if value.Type().Field(i).IsExported() && !extensionStringsValid(value.Field(i), depth+1) {
				return false
			}
		}
	}
	return true
}
