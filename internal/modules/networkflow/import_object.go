package networkflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"slices"
	"strings"
)

var errInvalidImportObject = errors.New("network flow import object failed owner validation")

// decodeClosedImportObject is private to the owner boundary. Required, optional and
// nullable members are explicit; Go zero values never stand in for presence.
func decodeClosedImportObject(raw json.RawMessage, required, optional, nullable string, target any) (map[string]json.RawMessage, error) {
	object, err := strictjson.DecodeUnicodeObject(bytes.NewReader(raw))
	if err != nil {
		return nil, errInvalidImportObject
	}
	for _, key := range strings.Fields(required) {
		if _, ok := object[key]; !ok {
			return nil, errInvalidImportObject
		}
	}
	allowed := strings.Fields(required + " " + optional)
	for key, value := range object {
		if !slices.Contains(allowed, key) || (bytes.Equal(bytes.TrimSpace(value), []byte("null")) && !slices.Contains(strings.Fields(nullable), key)) {
			return nil, errInvalidImportObject
		}
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return nil, errInvalidImportObject
	}
	return object, nil
}
