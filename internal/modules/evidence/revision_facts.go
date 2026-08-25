package evidence

import "github.com/JochiRaider/cartulary/internal/modules/revisions"

func evidenceRevisionFacts(beforeRow map[string]any, afterRow map[string]any, fieldKeys []string) []revisions.RevisionConflictFact {
	beforeCells, _ := beforeRow["cells"].(map[string]any)
	afterCells, _ := afterRow["cells"].(map[string]any)
	facts := make([]revisions.RevisionConflictFact, 0, len(fieldKeys))
	for _, key := range fieldKeys {
		beforeValue, beforePresent := beforeCells[key]
		afterValue, afterPresent := afterCells[key]
		facts = append(facts, revisions.RevisionConflictFact{
			FieldKey: key, BeforePresent: beforePresent, BeforeValue: beforeValue,
			AfterPresent: afterPresent, AfterValue: afterValue,
		})
	}
	return facts
}
