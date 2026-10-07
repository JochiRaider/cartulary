package querypage

import (
	"encoding/json"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

// Location is a classified, bounded provider result. Authorization and token
// issuance remain outside the storage provider.
type Location struct {
	Outcome     string
	Page        Result
	Predecessor map[string]string
}

func PositionForRow(row map[string]any, sort []viewschema.SortEntry) (map[string]string, error) {
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

// Locate evaluates membership and a target-first window through one caller-owned consistent read.
func Locate(recordID string, query viewschema.QueryMeta, read func(viewschema.QueryMeta, Window) (Result, error)) (Location, error) {
	membership := query
	membership.Filters = nil
	targetWindow := Window{Limit: 1, TargetRecordID: recordID}
	target, err := read(membership, targetWindow)
	if err != nil {
		return Location{}, err
	}
	if len(target.Rows) == 0 {
		return Location{Outcome: "unavailable"}, nil
	}
	matching, err := read(query, targetWindow)
	if err != nil {
		return Location{}, err
	}
	if len(matching.Rows) == 0 {
		return Location{Outcome: "outside_query"}, nil
	}
	position, err := PositionForRow(matching.Rows[0], query.Sort)
	if err != nil {
		return Location{}, err
	}
	predecessor, err := read(query, Window{Limit: 1, Before: true, Position: position})
	if err != nil {
		return Location{}, err
	}
	var start map[string]string
	if len(predecessor.Rows) != 0 {
		start, err = PositionForRow(predecessor.Rows[0], query.Sort)
		if err != nil {
			return Location{}, err
		}
	}
	page, err := read(query, Window{Limit: 100, Position: start})
	if err != nil {
		return Location{}, err
	}
	if len(page.Rows) == 0 || fmt.Sprint(page.Rows[0]["record_id"]) != recordID {
		return Location{}, fmt.Errorf("locator lost target in consistent read")
	}
	return Location{Outcome: "located", Page: page, Predecessor: start}, nil
}
