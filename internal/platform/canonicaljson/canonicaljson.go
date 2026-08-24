// Package canonicaljson encodes JSON-domain values using RFC 8785 (JCS).
//
// Marshal first projects the supplied value through encoding/json so structs,
// UUIDs, and json.Marshaler implementations retain their ordinary JSON shape.
// It then emits the closed JSON value with UTF-16 object-key ordering and the
// ECMAScript number representation required by RFC 8785.
package canonicaljson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

var ErrUnsupportedValue = errors.New("canonicaljson: unsupported JSON value")

// Marshal returns RFC-8785 canonical JSON without a trailing newline.
func Marshal(value any) ([]byte, error) {
	projected, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("project canonical JSON value: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(projected))
	decoder.UseNumber()
	var document any
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode projected canonical JSON value: %w", err)
	}
	var output bytes.Buffer
	if err := appendValue(&output, document); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

// Canonicalize validates one complete JSON document and returns its RFC-8785
// representation.
func Canonicalize(document []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode canonical JSON input: %w", err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		if err == nil {
			err = ErrUnsupportedValue
		}
		return nil, fmt.Errorf("decode canonical JSON input: trailing value: %w", err)
	}
	var output bytes.Buffer
	if err := appendValue(&output, value); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func appendValue(output *bytes.Buffer, value any) error {
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
		number, err := canonicalNumber(typed)
		if err != nil {
			return err
		}
		output.WriteString(number)
	case []any:
		output.WriteByte('[')
		for index, member := range typed {
			if index != 0 {
				output.WriteByte(',')
			}
			if err := appendValue(output, member); err != nil {
				return err
			}
		}
		output.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		slices.SortFunc(keys, compareUTF16)
		output.WriteByte('{')
		for index, key := range keys {
			if index != 0 {
				output.WriteByte(',')
			}
			if err := appendString(output, key); err != nil {
				return err
			}
			output.WriteByte(':')
			if err := appendValue(output, typed[key]); err != nil {
				return err
			}
		}
		output.WriteByte('}')
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedValue, value)
	}
	return nil
}

func appendString(output *bytes.Buffer, value string) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("%w: invalid UTF-8 string", ErrUnsupportedValue)
	}
	output.WriteByte('"')
	for _, character := range value {
		switch character {
		case '"', '\\':
			output.WriteByte('\\')
			output.WriteRune(character)
		case '\b':
			output.WriteString(`\b`)
		case '\t':
			output.WriteString(`\t`)
		case '\n':
			output.WriteString(`\n`)
		case '\f':
			output.WriteString(`\f`)
		case '\r':
			output.WriteString(`\r`)
		default:
			if character < 0x20 {
				output.WriteString(`\u00`)
				output.WriteString(strconv.FormatInt(int64(character), 16))
				if character < 0x10 {
					bytes := output.Bytes()
					last := len(bytes) - 1
					bytes[last] = '0'
					output.WriteByte("0123456789abcdef"[character])
				}
				continue
			}
			output.WriteRune(character)
		}
	}
	output.WriteByte('"')
	return nil
}

func canonicalNumber(number json.Number) (string, error) {
	value, err := number.Float64()
	if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
		return "", fmt.Errorf("%w: invalid number %q", ErrUnsupportedValue, number.String())
	}
	if value == 0 {
		return "0", nil
	}
	rendered := strconv.FormatFloat(value, 'g', -1, 64)
	absolute := math.Abs(value)
	if absolute >= 1e-6 && absolute < 1e21 {
		if strings.ContainsAny(rendered, "eE") {
			rendered = strconv.FormatFloat(value, 'f', -1, 64)
		}
		return rendered, nil
	}
	rendered = strings.ToLower(rendered)
	parts := strings.SplitN(rendered, "e", 2)
	if len(parts) != 2 {
		return rendered, nil
	}
	exponent, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("%w: invalid exponent %q", ErrUnsupportedValue, parts[1])
	}
	sign := ""
	if exponent >= 0 {
		sign = "+"
	}
	return parts[0] + "e" + sign + strconv.Itoa(exponent), nil
}

func compareUTF16(left, right string) int {
	leftUnits := utf16.Encode([]rune(left))
	rightUnits := utf16.Encode([]rune(right))
	for index := 0; index < len(leftUnits) && index < len(rightUnits); index++ {
		if leftUnits[index] < rightUnits[index] {
			return -1
		}
		if leftUnits[index] > rightUnits[index] {
			return 1
		}
	}
	return len(leftUnits) - len(rightUnits)
}
