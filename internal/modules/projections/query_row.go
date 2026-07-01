package projections

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func buildGenericRow(definition genericSurface, groupBy *string, values []any) (map[string]any, error) {
	recordID, err := uuidValue(values[0])
	if err != nil {
		return nil, err
	}
	cells := make(map[string]any, len(definition.fields))
	fieldValues := make(map[string]any, len(definition.fields))
	for index, field := range definition.fields {
		value := genericCellValue(field, values[index+2])
		fieldValues[field.key] = value
		cells[field.key] = map[string]any{"value": value}
	}
	row := map[string]any{
		"record_id":   recordID.String(),
		"row_version": values[1],
		"cells":       cells,
	}
	groupValues := map[string]any{}
	if groupBy != nil {
		groupValues[*groupBy] = fieldValues[*groupBy]
	}
	row["group_values"] = groupValues
	return row, nil
}

func uuidValue(value any) (uuid.UUID, error) {
	switch typed := value.(type) {
	case uuid.UUID:
		return typed, nil
	case string:
		parsed, err := uuid.Parse(typed)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("query workbook rows: invalid record_id %q", typed)
		}
		return parsed, nil
	case []byte:
		if len(typed) == 16 {
			parsed, err := uuid.FromBytes(typed)
			if err != nil {
				return uuid.UUID{}, fmt.Errorf("query workbook rows: invalid binary record_id: %w", err)
			}
			return parsed, nil
		}
		parsed, err := uuid.Parse(string(typed))
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("query workbook rows: invalid record_id bytes")
		}
		return parsed, nil
	case [16]byte:
		return uuid.UUID(typed), nil
	default:
		return uuid.UUID{}, fmt.Errorf("query workbook rows: record_id was %T", value)
	}
}

func genericCellValue(field genericField, value any) any {
	if field.kind == fieldKindCollection {
		if value != nil {
			if items, ok := collectionItemsFromValue(value); ok {
				return map[string]any{
					"kind":    "collection_value_v1",
					"ordered": field.ordered,
					"items":   items,
				}
			}
		}
		return map[string]any{
			"kind":    "collection_value_v1",
			"ordered": field.ordered,
			"items":   []map[string]any{},
		}
	}
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case time.Time:
		if field.kind == fieldKindDate {
			return typed.UTC().Format("2006-01-02")
		}
		return typed.UTC().Format(time.RFC3339Nano)
	case uuid.UUID:
		return typed.String()
	case []byte:
		if field.kind == fieldKindText && len(typed) == 16 {
			if parsed, err := uuid.FromBytes(typed); err == nil {
				return parsed.String()
			}
		}
		return string(typed)
	case [16]byte:
		if field.kind == fieldKindText {
			return uuid.UUID(typed).String()
		}
		return typed
	default:
		return typed
	}
}

func collectionItemsFromValue(value any) ([]map[string]any, bool) {
	var data []byte
	switch typed := value.(type) {
	case []byte:
		data = typed
	case string:
		data = []byte(typed)
	default:
		return nil, false
	}
	var items []map[string]any
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, false
	}
	if items == nil {
		items = []map[string]any{}
	}
	return items, true
}
