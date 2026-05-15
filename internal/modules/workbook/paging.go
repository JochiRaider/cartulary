package workbook

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func pageWorkbookResources(binding pagination.Binding, cursor *pagination.Cursor, query viewschema.QueryMeta, resources []map[string]any) ([]json.RawMessage, *pagination.Cursor, error) {
	start := 0
	if cursor != nil {
		if cursor.Mode != pagination.ModeKeyset || len(cursor.Position) == 0 {
			return nil, nil, pagination.ErrInvalidCursorToken
		}
		var found bool
		for index, resource := range resources {
			cmp, err := compareRowToCursorPosition(resource, query.Sort, cursor.Position)
			if err != nil {
				return nil, nil, err
			}
			if cmp > 0 {
				start = index
				found = true
				break
			}
		}
		if !found {
			return []json.RawMessage{}, nil, nil
		}
	}

	if binding.Limit < 1 {
		return nil, nil, pagination.ErrInvalidCursorToken
	}
	end := start + binding.Limit
	hasMore := end < len(resources)
	if end > len(resources) {
		end = len(resources)
	}
	if start >= end {
		return []json.RawMessage{}, nil, nil
	}

	rows, err := pagination.MarshalResources(resources[start:end])
	if err != nil {
		return nil, nil, err
	}
	if !hasMore {
		return rows, nil, nil
	}
	position, err := cursorPositionForRow(resources[end-1], query.Sort)
	if err != nil {
		return nil, nil, err
	}
	return rows, &pagination.Cursor{
		Mode:        pagination.ModeKeyset,
		Route:       binding.Route,
		ActorUserID: binding.ActorUserID,
		Limit:       binding.Limit,
		Scope:       binding.Scope,
		Position:    position,
	}, nil
}

func compareRowToCursorPosition(row map[string]any, sort []viewschema.SortEntry, position map[string]string) (int, error) {
	for _, entry := range sort {
		encoded, ok := position[entry.FieldKey]
		if !ok {
			return 0, pagination.ErrInvalidCursorToken
		}
		cursorValue, err := decodeCursorSortValue(encoded)
		if err != nil {
			return 0, err
		}
		rowValue, ok := rowSortValue(row, entry.FieldKey)
		if !ok {
			return 0, fmt.Errorf("workbook cursor sort field %q missing from row", entry.FieldKey)
		}
		if cmp := compareSortValues(rowValue, cursorValue, entry.Direction); cmp != 0 {
			return cmp, nil
		}
	}
	return 0, nil
}

func cursorPositionForRow(row map[string]any, sort []viewschema.SortEntry) (map[string]string, error) {
	position := make(map[string]string, len(sort))
	for _, entry := range sort {
		value, ok := rowSortValue(row, entry.FieldKey)
		if !ok {
			return nil, fmt.Errorf("workbook cursor sort field %q missing from row", entry.FieldKey)
		}
		payload, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		position[entry.FieldKey] = string(payload)
	}
	return position, nil
}

func rowSortValue(row map[string]any, fieldKey string) (any, bool) {
	switch fieldKey {
	case "record_id", "row_version":
		value, ok := row[fieldKey]
		return value, ok
	default:
		cells, ok := row["cells"].(map[string]any)
		if !ok {
			return nil, false
		}
		cell, ok := cells[fieldKey].(map[string]any)
		if !ok {
			return nil, false
		}
		value, ok := cell["value"]
		return value, ok
	}
}

func decodeCursorSortValue(encoded string) (any, error) {
	var value any
	decoder := json.NewDecoder(strings.NewReader(encoded))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, pagination.ErrInvalidCursorToken
	}
	return value, nil
}

func compareSortValues(left any, right any, direction string) int {
	if left == nil && right == nil {
		return 0
	}
	if left == nil {
		return 1
	}
	if right == nil {
		return -1
	}

	cmp := compareNonNullSortValues(left, right)
	if direction == "desc" {
		cmp = -cmp
	}
	return cmp
}

func compareNonNullSortValues(left any, right any) int {
	if leftBool, ok := left.(bool); ok {
		rightBool, _ := right.(bool)
		switch {
		case leftBool == rightBool:
			return 0
		case !leftBool && rightBool:
			return -1
		default:
			return 1
		}
	}
	if leftNumber, ok := sortNumber(left); ok {
		rightNumber, _ := sortNumber(right)
		switch {
		case leftNumber < rightNumber:
			return -1
		case leftNumber > rightNumber:
			return 1
		default:
			return 0
		}
	}
	return strings.Compare(fmt.Sprint(left), fmt.Sprint(right))
}

func sortNumber(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case float64:
		return typed, true
	case json.Number:
		parsed, err := typed.Float64()
		if err == nil && !math.IsNaN(parsed) {
			return parsed, true
		}
	}
	return 0, false
}
