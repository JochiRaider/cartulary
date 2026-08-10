package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCanonicalArtifactSourceCatalogFamily(t *testing.T) {
	t.Parallel()
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("find repository root: %v", err)
	}
	if err := validateArtifactSourceCatalogFamily(root); err != nil {
		t.Fatalf("validate canonical Artifacts source catalog: %v", err)
	}
	catalog, err := loadArtifactSourceCatalog(root)
	if err != nil {
		t.Fatalf("load canonical Artifacts source catalog: %v", err)
	}
	directCount := 0
	collectionCount := 0
	for _, surface := range catalog.Surfaces {
		directCount += len(surface.DirectFields)
		collectionCount += len(surface.CollectionFields)
	}
	if len(catalog.Surfaces) != 8 || directCount != 36 || collectionCount != 15 {
		t.Fatalf("catalog counts = %d/%d/%d, want 8/36/15", len(catalog.Surfaces), directCount, collectionCount)
	}
}

func TestArtifactSourceCatalogGeneratorRejectsInvalidFixtures(t *testing.T) {
	t.Parallel()
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("find repository root: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, artifactSourceCatalogPath))
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
	t.Run("unowned relation", func(t *testing.T) {
		catalog := load(t)
		catalog["surfaces"].([]any)[0].(map[string]any)["direct_fields"].([]any)[0].(map[string]any)["table"] = "records"
		if _, err := parseArtifactSourceCatalog(catalog); err == nil {
			t.Fatal("catalog accepted unowned relation")
		}
	})
	t.Run("duplicate field", func(t *testing.T) {
		catalog := load(t)
		fields := catalog["surfaces"].([]any)[0].(map[string]any)["direct_fields"].([]any)
		fields[1].(map[string]any)["field_key"] = fields[0].(map[string]any)["field_key"]
		if _, err := parseArtifactSourceCatalog(catalog); err == nil {
			t.Fatal("catalog accepted duplicate field")
		}
	})
	t.Run("operation family mismatch", func(t *testing.T) {
		catalog := load(t)
		field := catalog["surfaces"].([]any)[0].(map[string]any)["collection_fields"].([]any)[0].(map[string]any)
		field["allowed_operations"] = []any{"add_tag", "remove_tag"}
		if _, err := parseArtifactSourceCatalog(catalog); err == nil {
			t.Fatal("catalog accepted mismatched operation family")
		}
	})
	t.Run("view source mismatch", func(t *testing.T) {
		catalog, err := loadArtifactSourceCatalog(root)
		if err != nil {
			t.Fatalf("load catalog: %v", err)
		}
		catalog.Surfaces[0].ArtifactType = "future_artifact"
		if err := enrichArtifactSurfaceFromViewSchema(root, &catalog.Surfaces[0]); err == nil {
			t.Fatal("generator accepted mismatched view source filter")
		}
	})
	t.Run("read only field", func(t *testing.T) {
		catalog, err := loadArtifactSourceCatalog(root)
		if err != nil {
			t.Fatalf("load catalog: %v", err)
		}
		catalog.Surfaces[0].DirectFields[0].FieldKey = "comm_log.comm_id"
		if err := enrichArtifactSurfaceFromViewSchema(root, &catalog.Surfaces[0]); err == nil {
			t.Fatal("generator accepted read-only field")
		}
	})
}
