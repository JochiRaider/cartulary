package timeline

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
)

func buildClearCellsOwnerRows(fieldKeys []string, rowCount int) ([]ownerBatchRowPlanV1, error) {
	if rowCount < 1 || rowCount > mutationpolicy.MaxOwnerBatchTargets || !mutationpolicy.ValidClearFields(fieldKeys) {
		return nil, fmt.Errorf("invalid clear-cells fields or target count")
	}
	rows := make([]ownerBatchRowPlanV1, rowCount)
	for index := range rows {
		rows[index].RowOrdinal = index + 1
		for _, fieldKey := range fieldKeys {
			// Presence is explicit in the plan; a nil TextValue is authoritative null.
			rows[index].Cells = append(rows[index].Cells, ownerBatchCellV1{
				FieldKey: fieldKey,
				Change:   PatchChange{FieldKey: fieldKey},
			})
		}
	}
	return rows, nil
}
