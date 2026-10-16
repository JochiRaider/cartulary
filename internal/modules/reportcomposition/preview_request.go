package reportcomposition

import (
	"bytes"
	"encoding/json"
	"errors"
)

// DecodeNormalizedPreviewRequest reads the private idempotency representation,
// whose absent composition version is explicitly null. Public requests still
// reject an explicitly null version. Exact normalization rejects missing,
// additional, noncanonical or altered retained members rather than repairing them.
func DecodeNormalizedPreviewRequest(raw []byte) (PreviewRequest, error) {
	invalid := errors.New("reportcomposition: invalid normalized preview request")
	object, apiErr := decodeJSONObject(bytes.NewReader(raw))
	if apiErr != nil {
		return PreviewRequest{}, invalid
	}
	if bytes.Equal(object["composition_version"], []byte("null")) {
		delete(object, "composition_version")
	}
	public, err := json.Marshal(object)
	if err != nil {
		return PreviewRequest{}, invalid
	}
	request, apiErr := DecodePreviewRequest(bytes.NewReader(public))
	if apiErr != nil || !bytes.Equal(request.Normalized, raw) {
		return PreviewRequest{}, invalid
	}
	return request, nil
}
