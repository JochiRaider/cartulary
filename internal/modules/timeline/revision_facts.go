package timeline

import (
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func timelineRevisionFacts(beforeRow map[string]any, afterRow map[string]any, fieldKeys []string) []revisions.RevisionConflictFact {
	beforeCells := timelineRowCells(beforeRow)
	afterCells := timelineRowCells(afterRow)
	keys := slices.Clone(fieldKeys)
	slices.Sort(keys)
	keys = slices.Compact(keys)
	facts := make([]revisions.RevisionConflictFact, 0, len(keys))
	for _, key := range keys {
		beforeValue, beforePresent := beforeCells[key]
		afterValue, afterPresent := afterCells[key]
		facts = append(facts, revisions.RevisionConflictFact{
			FieldKey: key, BeforePresent: beforePresent, BeforeValue: beforeValue,
			AfterPresent: afterPresent, AfterValue: afterValue,
		})
	}
	return facts
}

func timelineRowCells(row map[string]any) map[string]any {
	if row == nil {
		return map[string]any{}
	}
	cells, _ := row["cells"].(map[string]any)
	if cells == nil {
		return map[string]any{}
	}
	return cells
}
