package httpapi

import (
	"encoding/json"
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
	return strictjson.DecodeObject(reader)
}

func ValidateStrictJSONObject(data []byte) error {
	return strictjson.ValidateObject(data)
}
