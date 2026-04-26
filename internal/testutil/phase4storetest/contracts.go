package phase4storetest

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

type viewContractDocument struct {
	TechnicalFields []string `json:"technical_fields"`
	Fields          []struct {
		FieldKey          string `json:"field_key"`
		EntityBindingMode string `json:"entity_binding_mode"`
	} `json:"fields"`
}

func AllowedFieldKeys(t testing.TB, testID string, viewSchemaID string) []string {
	t.Helper()

	document := loadViewContract(t, testID, viewSchemaID)
	allowed := make([]string, 0, len(document.TechnicalFields)+len(document.Fields))
	allowed = append(allowed, document.TechnicalFields...)
	for _, field := range document.Fields {
		allowed = append(allowed, field.FieldKey)
	}
	slices.Sort(allowed)
	return allowed
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

func loadViewContract(t testing.TB, testID string, viewSchemaID string) viewContractDocument {
	t.Helper()

	RequireViewContract(t, testID, viewSchemaID)

	data, err := os.ReadFile(viewContractPath(viewSchemaID))
	if err != nil {
		t.Fatalf("Phase 4 %s failed to read view contract %s: %v", testID, viewSchemaID, err)
	}

	var document viewContractDocument
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("Phase 4 %s failed to parse view contract %s: %v", testID, viewSchemaID, err)
	}

	return document
}
