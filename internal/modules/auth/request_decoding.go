package auth

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return nil, invalidAuthRequest("", "request body must be one JSON object")
	}
	return raw, nil
}

func decodeEnterpriseObject(reader any) (map[string]json.RawMessage, *httpapi.APIError) {
	raw, ok := decodeRawObject(reader)
	if !ok {
		return nil, invalidEnterpriseAuthRequest("", "request_not_object")
	}
	return raw, nil
}

func decodeMutationObject(reader any) (map[string]json.RawMessage, *httpapi.APIError) {
	raw, ok := decodeRawObject(reader)
	if !ok {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
}

func decodeRawObject(reader any) (map[string]json.RawMessage, bool) {
	body, ok := reader.(interface{ Read([]byte) (int, error) })
	if !ok {
		return nil, false
	}
	raw, err := httpapi.DecodeStrictJSONObject(body)
	if err != nil {
		return nil, false
	}
	return raw, true
}

func requireInt64(raw map[string]json.RawMessage, field string, target *int64) *httpapi.APIError {
	value, ok := raw[field]
	if !ok {
		return invalidMutationPayload(field, "missing_required_field")
	}
	if string(value) == "null" || json.Unmarshal(value, target) != nil {
		return invalidMutationPayload(field, "field_not_nullable")
	}
	return nil
}

func requireNonEmptyString(raw map[string]json.RawMessage, field string, target *string) *httpapi.APIError {
	value, ok := raw[field]
	if !ok {
		return invalidMutationPayload(field, "missing_required_field")
	}
	if string(value) == "null" || json.Unmarshal(value, target) != nil || strings.TrimSpace(*target) == "" {
		return invalidMutationPayload(field, "field_not_nullable")
	}
	return nil
}

func decodeOptionalBindingReason(value json.RawMessage) (*string, *httpapi.APIError) {
	if len(value) == 0 || string(value) == "null" {
		return nil, nil
	}
	var reason string
	if err := json.Unmarshal(value, &reason); err != nil {
		return nil, invalidMutationPayload("reason", "field_not_nullable")
	}
	return authn.NormalizeReasonNote(&reason), nil
}
