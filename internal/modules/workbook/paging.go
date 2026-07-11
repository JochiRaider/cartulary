package workbook

import (
	"encoding/json"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func pageBoundedWorkbookResources(binding pagination.Binding, query viewschema.QueryMeta, page querypage.Result) ([]json.RawMessage, *pagination.Cursor, error) {
	rows, err := pagination.MarshalResources(page.Rows)
	if err != nil {
		return nil, nil, err
	}
	if !page.HasMore || len(page.Rows) == 0 {
		return rows, nil, nil
	}
	position, err := cursorPositionForRow(page.Rows[len(page.Rows)-1], query.Sort)
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
