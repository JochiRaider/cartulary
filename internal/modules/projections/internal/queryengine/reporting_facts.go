package queryengine

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// Reporting source facts are selected explicitly by each owner query. Decode
// them without a float64 round trip; canonical field admission belongs to the
// downstream Reporting owner.
func decodeReportingFactObject(raw []byte) (map[string]any, error) {
	value, err := canonicaljson.DecodeStrict(raw)
	if err != nil {
		return nil, errors.New("invalid reporting source fact JSON")
	}
	object, ok := value.(map[string]any)
	if !ok || object == nil {
		return nil, errors.New("invalid reporting source fact object")
	}
	return object, nil
}
