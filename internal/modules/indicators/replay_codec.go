package indicators

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	indicatorCreateRouteKey = "indicators.rows.create"
	httpStatusCreated       = 201
	httpStatusOK            = 200
)

func buildStoredCreateResponse(changeSetID uuid.UUID, row map[string]any) map[string]any {
	return map[string]any{
		"view_schema_id": ViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row":            row,
	}
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractUUIDFromPayload(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok || text == "" {
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsed, nil
}

func extractInt64FromPayload(payload map[string]any, path ...string) (int64, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	number, ok := current.(float64)
	if !ok || number < 1 || number != float64(int64(number)) {
		return 0, fmt.Errorf("decode payload integer path %q", strings.Join(path, "."))
	}
	return int64(number), nil
}
