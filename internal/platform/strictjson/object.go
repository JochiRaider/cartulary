package strictjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var (
	ErrNotObject       = errors.New("strict json: value is not one object")
	ErrDuplicateMember = errors.New("strict json: duplicate object member")
	ErrTrailingData    = errors.New("strict json: trailing data after object")
	ErrMalformed       = errors.New("strict json: malformed json")
)

func ValidateObject(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	token, err := decoder.Token()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return ErrNotObject
		}
		return fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != '{' {
		return ErrNotObject
	}
	if err := consumeObject(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err == nil {
		return ErrTrailingData
	} else if !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	return nil
}

// DecodeObject admits exactly one duplicate-free JSON object and preserves
// each member as raw JSON for owner-specific semantic admission.
func DecodeObject(reader io.Reader) (map[string]json.RawMessage, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	if err := ValidateObject(data); err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	if raw == nil {
		return nil, ErrNotObject
	}
	return raw, nil
}

func consumeObject(decoder *json.Decoder) error {
	seen := make(map[string]struct{})
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMalformed, err)
		}
		key, ok := token.(string)
		if !ok {
			return ErrMalformed
		}
		if _, duplicate := seen[key]; duplicate {
			return ErrDuplicateMember
		}
		seen[key] = struct{}{}
		if err := consumeValue(decoder); err != nil {
			return err
		}
	}
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != '}' {
		return ErrMalformed
	}
	return nil
}

func consumeArray(decoder *json.Decoder) error {
	for decoder.More() {
		if err := consumeValue(decoder); err != nil {
			return err
		}
	}
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != ']' {
		return ErrMalformed
	}
	return nil
}

func consumeValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		return consumeObject(decoder)
	case '[':
		return consumeArray(decoder)
	default:
		return ErrMalformed
	}
}
