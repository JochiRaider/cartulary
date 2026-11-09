package canonicaljson

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"
)

var (
	ErrInvalidJSON     = errors.New("canonicaljson: invalid JSON")
	ErrInvalidUnicode  = errors.New("canonicaljson: invalid Unicode")
	ErrDuplicateMember = errors.New("canonicaljson: duplicate object member")
)

// DecodeStrict admits one JSON value without repairing invalid Unicode or
// discarding duplicate members. Numbers retain their lexical form for the
// caller's schema validation. It deliberately applies no application-specific
// number bounds, normalization, unknown-member policy, or default values.
// Errors never include untrusted keys or values.
func DecodeStrict(document []byte) (any, error) {
	if !utf8.Valid(document) {
		return nil, ErrInvalidUnicode
	}
	// Validate the complete document before looking for duplicate members. A
	// duplicate near the start must not conceal invalid syntax later in the
	// input. Valid scans the bytes without constructing or repairing an object.
	if !json.Valid(document) {
		return nil, ErrInvalidJSON
	}
	if !validStringEscapes(document) {
		return nil, ErrInvalidUnicode
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	value, err := decodeStrictValue(decoder, 0)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, ErrInvalidJSON
	}
	return value, nil
}

func decodeStrictValue(decoder *json.Decoder, depth int) (any, error) {
	// Match encoding/json's maximum nesting depth while bounding recursion even
	// when Token is used rather than Decode.
	if depth > 10000 {
		return nil, ErrInvalidJSON
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, ErrInvalidJSON
	}
	delimiter, composite := token.(json.Delim)
	if !composite {
		return token, nil
	}
	switch delimiter {
	case '{':
		object := make(map[string]any)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, ErrInvalidJSON
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, ErrInvalidJSON
			}
			if _, exists := object[key]; exists {
				return nil, ErrDuplicateMember
			}
			value, err := decodeStrictValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			object[key] = value
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return nil, ErrInvalidJSON
		}
		return object, nil
	case '[':
		array := make([]any, 0)
		for decoder.More() {
			value, err := decodeStrictValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return nil, ErrInvalidJSON
		}
		return array, nil
	default:
		return nil, ErrInvalidJSON
	}
}

// ValidateUnicode examines original bytes and escaped UTF-16 without decoding
// or repairing strings. JSON syntax is a separate admission obligation.
func ValidateUnicode(document []byte) error {
	if !utf8.Valid(document) || !validStringEscapes(document) {
		return ErrInvalidUnicode
	}
	return nil
}

// encoding/json replaces lone UTF-16 surrogates with U+FFFD. Check escapes in
// the original bytes before letting that decoder construct strings.
func validStringEscapes(document []byte) bool {
	for i := 0; i < len(document); i++ {
		if document[i] != '"' {
			continue
		}
		for i++; i < len(document) && document[i] != '"'; i++ {
			if document[i] != '\\' {
				continue
			}
			i++
			if i >= len(document) {
				return true // The JSON parser reports the syntax error.
			}
			if document[i] != 'u' {
				continue
			}
			unit, ok := hexUnit(document, i+1)
			if !ok {
				return true // Invalid escape syntax, not a repaired surrogate.
			}
			i += 4
			if unit >= 0xdc00 && unit <= 0xdfff {
				return false
			}
			if unit < 0xd800 || unit > 0xdbff {
				continue
			}
			if i+2 >= len(document) || document[i+1] != '\\' || document[i+2] != 'u' {
				return false
			}
			low, ok := hexUnit(document, i+3)
			if !ok || low < 0xdc00 || low > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return true
}

func hexUnit(document []byte, start int) (uint16, bool) {
	if start+4 > len(document) {
		return 0, false
	}
	var value uint16
	for _, b := range document[start : start+4] {
		value <<= 4
		switch {
		case b >= '0' && b <= '9':
			value |= uint16(b - '0')
		case b >= 'a' && b <= 'f':
			value |= uint16(b-'a') + 10
		case b >= 'A' && b <= 'F':
			value |= uint16(b-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}
