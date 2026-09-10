package httpapi

import (
	"encoding/json"

	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

// History anchors are stable for the retained lifetime of a record. Resolving
// the anchor in the current ordered materialization preserves live continuation
// across new commits without retaining a snapshot or a numeric row position.
func pageRecordHistory(binding pagination.Binding, cursor *pagination.Cursor, resources []map[string]any) ([]json.RawMessage, *pagination.Cursor, error) {
	start := 0
	if cursor != nil {
		anchor := cursor.Position["after_history_item_ref"]
		if cursor.Mode != pagination.ModeKeyset || len(cursor.Position) != 1 || anchor == "" {
			return nil, nil, pagination.ErrInvalidCursorToken
		}
		start = -1
		for index, resource := range resources {
			if resource["history_item_ref"] == anchor {
				start = index + 1
				break
			}
		}
		if start < 0 {
			return nil, nil, pagination.ErrInvalidCursorToken
		}
	}
	end := min(start+binding.Limit, len(resources))
	page, err := pagination.MarshalResources(resources[start:end])
	if err != nil || end == len(resources) {
		return page, nil, err
	}
	anchor, ok := resources[end-1]["history_item_ref"].(string)
	if !ok || anchor == "" {
		return nil, nil, pagination.ErrInvalidCursorToken
	}
	return page, &pagination.Cursor{
		Version:     pagination.CursorVersion,
		Mode:        pagination.ModeKeyset,
		Route:       binding.Route,
		ActorUserID: binding.ActorUserID,
		Limit:       binding.Limit,
		Scope:       binding.Scope,
		Position:    map[string]string{"after_history_item_ref": anchor},
	}, nil
}
