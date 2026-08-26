package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCanonicalTasksDecisionsSourceCatalogFamily(t *testing.T) {
	t.Parallel()
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("find repository root: %v", err)
	}
	if err := validateTasksDecisionsSourceCatalogFamily(root); err != nil {
		t.Fatalf("validate canonical Tasks/Decisions source catalog: %v", err)
	}
	catalog, err := loadTasksDecisionsSourceCatalog(root)
	if err != nil {
		t.Fatalf("load canonical Tasks/Decisions source catalog: %v", err)
	}
	directCount := 0
	collectionCount := 0
	for _, surface := range catalog.Surfaces {
		directCount += len(surface.DirectFields)
		collectionCount += len(surface.CollectionFields)
	}
	if len(catalog.Surfaces) != 2 || directCount != 20 || collectionCount != 3 {
		t.Fatalf("catalog counts = %d/%d/%d, want 2/20/3", len(catalog.Surfaces), directCount, collectionCount)
	}
}

func TestTasksDecisionsSourceCatalogGeneratorRejectsInvalidFixtures(t *testing.T) {
	t.Parallel()
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("find repository root: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, tasksDecisionsSourceCatalogPath))
	if err != nil {
		t.Fatalf("read canonical catalog: %v", err)
	}
	load := func(t testing.TB) map[string]any {
		t.Helper()
		value, err := decodeContract(data)
		if err != nil {
			t.Fatalf("decode canonical catalog: %v", err)
		}
		return value.(map[string]any)
	}
	t.Run("unsafe source table", func(t *testing.T) {
		catalog := load(t)
		catalog["surfaces"].([]any)[0].(map[string]any)["source_table"] = "decisions;drop"
		if _, err := parseTasksDecisionsSourceCatalog(catalog); err == nil {
			t.Fatal("catalog accepted unsafe source table")
		}
	})
	t.Run("duplicate field", func(t *testing.T) {
		catalog := load(t)
		fields := catalog["surfaces"].([]any)[0].(map[string]any)["direct_fields"].([]any)
		fields[1].(map[string]any)["field_key"] = fields[0].(map[string]any)["field_key"]
		if _, err := parseTasksDecisionsSourceCatalog(catalog); err == nil {
			t.Fatal("catalog accepted duplicate field")
		}
	})
	t.Run("cross surface field", func(t *testing.T) {
		catalog := load(t)
		fields := catalog["surfaces"].([]any)[0].(map[string]any)["direct_fields"].([]any)
		fields[len(fields)-1].(map[string]any)["field_key"] = "task.workstream"
		if _, err := parseTasksDecisionsSourceCatalog(catalog); err == nil {
			t.Fatal("generator accepted cross-surface field")
		}
	})
	t.Run("operation mismatch", func(t *testing.T) {
		catalog := load(t)
		field := catalog["surfaces"].([]any)[0].(map[string]any)["collection_fields"].([]any)[0].(map[string]any)
		field["allowed_operations"] = []any{"remove_record_ref", "add_record_ref"}
		if _, err := parseTasksDecisionsSourceCatalog(catalog); err == nil {
			t.Fatal("catalog accepted mismatched operation order")
		}
	})
	t.Run("reference role mismatch", func(t *testing.T) {
		catalog := load(t)
		field := catalog["surfaces"].([]any)[1].(map[string]any)["direct_fields"].([]any)[6].(map[string]any)
		field["reference_role"] = "same_incident_record"
		field["expected_target_record_type"] = "party"
		parsed, err := parseTasksDecisionsSourceCatalog(catalog)
		if err != nil {
			t.Fatalf("parse reference fixture: %v", err)
		}
		if err := enrichTasksDecisionsSurfaceFromViewSchema(root, &parsed.Surfaces[1]); err == nil {
			t.Fatal("generator accepted mismatched view reference contract")
		}
	})
	t.Run("read only field", func(t *testing.T) {
		catalog, err := loadTasksDecisionsSourceCatalog(root)
		if err != nil {
			t.Fatalf("load catalog: %v", err)
		}
		catalog.Surfaces[0].DirectFields[0].FieldKey = "decision.updated_at"
		if err := enrichTasksDecisionsSurfaceFromViewSchema(root, &catalog.Surfaces[0]); err == nil {
			t.Fatal("generator accepted read-only field")
		}
	})
}
