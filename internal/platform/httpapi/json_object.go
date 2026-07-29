package httpapi

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

var (
	ErrStrictJSONNotObject       = strictjson.ErrNotObject
	ErrStrictJSONDuplicateMember = strictjson.ErrDuplicateMember
	ErrStrictJSONTrailingData    = strictjson.ErrTrailingData
	ErrStrictJSONMalformed       = strictjson.ErrMalformed
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
	return strictjson.ValidateObject(data)
}
