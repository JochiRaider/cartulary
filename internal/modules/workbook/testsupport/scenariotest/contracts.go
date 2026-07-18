package scenariotest

import (
	"slices"
	"testing"

	viewschematest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
)

func AllowedFieldKeys(t testing.TB, testID string, viewSchemaID string) []string {
	t.Helper()
	return viewschematest.AllowedFieldKeys(t, "Record relationships "+testID, viewSchemaID)
}

func SortedRowFieldKeys(t testing.TB, row map[string]any) []string {
	t.Helper()

	cells, ok := row["cells"].(map[string]any)
	if !ok {
		t.Fatalf("expected row cells object, got %#v", row["cells"])
	}
	fieldKeys := make([]string, 0, len(cells))
	for fieldKey := range cells {
		fieldKeys = append(fieldKeys, fieldKey)
	}
	slices.Sort(fieldKeys)
	return fieldKeys
}
