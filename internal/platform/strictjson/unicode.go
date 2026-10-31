package strictjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

var ErrTooLarge = errors.New("strict json: encoded object exceeds byte limit")

// DecodeUnicodeObject preserves source Unicode while admitting one strict JSON
// object. The caller retains encoded body-size ownership, as with DecodeObject.
func DecodeUnicodeObject(reader io.Reader) (map[string]json.RawMessage, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	return decodeUnicodeObject(data)
}

func decodeUnicodeObject(data []byte) (map[string]json.RawMessage, error) {
	if err := validateUnicode(data); err != nil {
		return nil, err
	}
	return DecodeObject(bytes.NewReader(data))
}

// DecodeBoundedUnicodeObject bounds encoded input before decoding and rejects
// malformed UTF-8 and unpaired UTF-16 escapes instead of replacing them.
func DecodeBoundedUnicodeObject(reader io.Reader, maxBytes int64) (map[string]json.RawMessage, error) {
	if maxBytes <= 0 {
		return nil, ErrTooLarge
	}
	data, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	if int64(len(data)) > maxBytes {
		return nil, ErrTooLarge
	}
	return decodeUnicodeObject(data)
}

func validateUnicode(data []byte) error {
	if !utf8.Valid(data) {
		return ErrMalformed
	}
	inString := false
	for i := 0; i < len(data); i++ {
		if data[i] == '"' {
			inString = !inString
			continue
		}
		if !inString || data[i] != '\\' {
			continue
		}
		i++
		if i >= len(data) {
			return ErrMalformed
		}
		if data[i] != 'u' {
			continue
		}
		value, ok := unicodeEscape(data, i)
		if !ok {
			return ErrMalformed
		}
		i += 4
		if value >= 0xdc00 && value <= 0xdfff {
			return ErrMalformed
		}
		if value < 0xd800 || value > 0xdbff {
			continue
		}
		if i+2 >= len(data) || data[i+1] != '\\' || data[i+2] != 'u' {
			return ErrMalformed
		}
		low, ok := unicodeEscape(data, i+2)
		if !ok || low < 0xdc00 || low > 0xdfff {
			return ErrMalformed
		}
		i += 6
	}
	return nil
}

func unicodeEscape(data []byte, u int) (uint16, bool) {
	if u+4 >= len(data) {
		return 0, false
	}
	var value uint16
	for _, digit := range data[u+1 : u+5] {
		value <<= 4
		switch {
		case digit >= '0' && digit <= '9':
			value += uint16(digit - '0')
		case digit >= 'a' && digit <= 'f':
			value += uint16(digit - 'a' + 10)
		case digit >= 'A' && digit <= 'F':
			value += uint16(digit - 'A' + 10)
		default:
			return 0, false
		}
	}
	return value, true
}
