package timeline

import (
	"reflect"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/rowpresenter"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func buildRow(record workbookprojection.DerivedRecord) map[string]any {
	return rowpresenter.BuildRow(record.PresenterRecord())
}

func computeChangedFieldKeys(before *workbookprojection.DerivedRecord, after workbookprojection.DerivedRecord) []string {
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
