package workbook

import (
	"encoding/json"

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
	return querypage.PositionForRow(row, sort)
}
