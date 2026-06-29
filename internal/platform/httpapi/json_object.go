package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var (
	ErrStrictJSONNotObject       = errors.New("strict json: request body is not one object")
	ErrStrictJSONDuplicateMember = errors.New("strict json: duplicate object member")
	ErrStrictJSONTrailingData    = errors.New("strict json: trailing data after object")
	ErrStrictJSONMalformed       = errors.New("strict json: malformed json")
)

func DecodeStrictJSONObject(reader io.Reader) (map[string]json.RawMessage, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStrictJSONMalformed, err)
	}
	if err := ValidateStrictJSONObject(data); err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStrictJSONMalformed, err)
	}
	if raw == nil {
		return nil, ErrStrictJSONNotObject
	}
	return raw, nil
}

func ValidateStrictJSONObject(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	token, err := decoder.Token()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return ErrStrictJSONNotObject
		}
		return fmt.Errorf("%w: %v", ErrStrictJSONMalformed, err)
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != '{' {
		return ErrStrictJSONNotObject
	}
	if err := consumeStrictJSONObject(decoder); err != nil {
		return err
	}
	if token, err := decoder.Token(); err == nil {
		if token != nil {
			return ErrStrictJSONTrailingData
		}
		return ErrStrictJSONTrailingData
	} else if !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: %v", ErrStrictJSONMalformed, err)
	}
	return nil
}

func consumeStrictJSONObject(decoder *json.Decoder) error {
	seen := make(map[string]struct{})
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("%w: %v", ErrStrictJSONMalformed, err)
		}
		key, ok := token.(string)
		if !ok {
			return ErrStrictJSONMalformed
		}
		if _, ok := seen[key]; ok {
			return ErrStrictJSONDuplicateMember
		}
		seen[key] = struct{}{}
		if err := consumeStrictJSONValue(decoder); err != nil {
			return err
		}
	}
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStrictJSONMalformed, err)
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != '}' {
		return ErrStrictJSONMalformed
	}
	return nil
}

func consumeStrictJSONArray(decoder *json.Decoder) error {
	for decoder.More() {
		if err := consumeStrictJSONValue(decoder); err != nil {
			return err
		}
	}
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStrictJSONMalformed, err)
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != ']' {
		return ErrStrictJSONMalformed
	}
	return nil
}

func consumeStrictJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStrictJSONMalformed, err)
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		return consumeStrictJSONObject(decoder)
	case '[':
		return consumeStrictJSONArray(decoder)
	default:
		return ErrStrictJSONMalformed
	}
}
