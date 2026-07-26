package timeline

import (
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/rowpresenter"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func buildRow(record workbookprojection.DerivedRecord) map[string]any {
	return rowpresenter.BuildRow(record.PresenterRecord())
}

func ComputeChangedFieldKeys(before *workbookprojection.DerivedRecord, after workbookprojection.DerivedRecord) []string {
	beforeCells := map[string]any{}
	if before != nil {
		beforeRow := buildRow(*before)
		beforeCells, _ = beforeRow["cells"].(map[string]any)
	}

	afterRow := buildRow(after)
	afterCells, _ := afterRow["cells"].(map[string]any)
	changed := make([]string, 0, len(afterCells))
	for fieldKey, afterValue := range afterCells {
		beforeValue, ok := beforeCells[fieldKey]
		if !ok || !reflect.DeepEqual(beforeValue, afterValue) {
			changed = append(changed, fieldKey)
		}
	}
	slices.Sort(changed)
	return changed
}

func collectionValue(ordered bool, items []map[string]any) map[string]any {
	if items == nil {
		items = []map[string]any{}
	}
	return map[string]any{
		"kind":    "collection_value_v1",
		"ordered": ordered,
		"items":   items,
	}
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func derefString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
