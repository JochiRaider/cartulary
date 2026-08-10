package sourcecatalog

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractartifacts"
)

func TestCatalogRejectsGeneratedOwnerFactDrift(t *testing.T) {
	t.Parallel()
	catalog, err := build(cloneGeneratedCatalog())
	if err != nil || len(catalog.Surfaces()) != 8 || len(catalog.Fields()) != 51 {
		t.Fatalf("canonical generated catalog = %d surfaces, %d fields, err=%v", len(catalog.Surfaces()), len(catalog.Fields()), err)
	}
	tests := []struct {
		name   string
		mutate func([]contractartifacts.SourceSurface) []contractartifacts.SourceSurface
	}{
		{name: "missing surface", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			return catalog[:len(catalog)-1]
		}},
		{name: "duplicate surface", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			catalog[len(catalog)-1] = catalog[0]
			return catalog
		}},
		{name: "duplicate artifact type", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			catalog[len(catalog)-1].ArtifactType = catalog[0].ArtifactType
			return catalog
		}},
		{name: "cross surface field", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			field := catalog[0].DirectFields[len(catalog[0].DirectFields)-1]
			catalog[0].DirectFields = catalog[0].DirectFields[:len(catalog[0].DirectFields)-1]
			catalog[1].DirectFields = append(catalog[1].DirectFields, field)
			return catalog
		}},
		{name: "read only field", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			catalog[0].DirectFields[0].FieldKey = "comm_log.comm_id"
			return catalog
		}},
		{name: "source filter mismatch", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			catalog[0].ArtifactType = "future_artifact"
			return catalog
		}},
		{name: "duplicate field", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			catalog[0].DirectFields[1] = catalog[0].DirectFields[0]
			return catalog
		}},
		{name: "stale view facts", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			catalog[0].DirectFields[0].View.Clearable = !catalog[0].DirectFields[0].View.Clearable
			return catalog
		}},
		{name: "unowned storage relation", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			catalog[0].DirectFields[0].Table = "records"
			return catalog
		}},
		{name: "unsafe storage identifier", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			catalog[0].DirectFields[0].Column = "title;drop_table"
			return catalog
		}},
		{name: "unknown collection operation", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			catalog[0].CollectionFields[0].AllowedOperations[0] = "future_operation"
			return catalog
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := build(tc.mutate(cloneGeneratedCatalog())); err == nil {
				t.Fatalf("runtime catalog accepted %s drift", tc.name)
			}
		})
	}
}

func TestCatalogLookupsAreIndexedAndDefensive(t *testing.T) {
	t.Parallel()
	catalog, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	loadedAgain, err := Load()
	if err != nil || loadedAgain != catalog {
		t.Fatalf("Load() did not return cached catalog: same=%v err=%v", loadedAgain == catalog, err)
	}
	surfaces := catalog.Surfaces()
	if len(surfaces) != 8 {
		t.Fatalf("surfaces = %d, want 8", len(surfaces))
	}
	projectionSurfaces := catalog.ProjectionSurfaces()
	if len(projectionSurfaces) != 8 || projectionSurfaces[0].ViewSchemaID != contractartifacts.NoteViewSchemaID {
		t.Fatalf("projection surface order = %#v", projectionSurfaces)
	}
	projectionSurfaces[0].ArtifactType = "mutated"
	if catalog.ProjectionSurfaces()[0].ArtifactType == "mutated" {
		t.Fatal("projection surface mutation escaped defensive copy")
	}
	surfaces[0].ArtifactType = "mutated"
	first, ok := catalog.SurfaceByViewID(contractartifacts.SourceCatalog[0].ViewSchemaID)
	if !ok || first.ArtifactType != contractartifacts.SourceCatalog[0].ArtifactType {
		t.Fatalf("surface mutation escaped defensive copy: %#v, ok=%v", first, ok)
	}
	byType, ok := catalog.SurfaceByArtifactType(first.ArtifactType)
	if !ok || byType != first {
		t.Fatalf("artifact type index = %#v, %v; want %#v", byType, ok, first)
	}
	fieldKey := contractartifacts.SourceCatalog[0].CollectionFields[0].FieldKey
	field, ok := catalog.Field(fieldKey)
	if !ok || len(field.Collection.AllowedOperations) == 0 {
		t.Fatalf("collection field %q missing", fieldKey)
	}
	field.Collection.AllowedOperations[0] = "mutated"
	loadedField, _ := catalog.Field(fieldKey)
	if loadedField.Collection.AllowedOperations[0] == "mutated" {
		t.Fatal("collection operation mutation escaped defensive copy")
	}
	fields := catalog.Fields()
	delete(fields, fieldKey)
	if _, ok := catalog.Field(fieldKey); !ok {
		t.Fatal("field map mutation escaped defensive copy")
	}
	mappings := catalog.WritableDirectStorageMappings()
	for key := range mappings {
		delete(mappings, key)
		break
	}
	if len(catalog.WritableDirectStorageMappings()) != 36 {
		t.Fatal("storage mapping mutation escaped defensive copy")
	}
}

func cloneGeneratedCatalog() []contractartifacts.SourceSurface {
	result := make([]contractartifacts.SourceSurface, len(contractartifacts.SourceCatalog))
	for index, surface := range contractartifacts.SourceCatalog {
		result[index] = surface
		result[index].DirectFields = append([]contractartifacts.DirectField(nil), surface.DirectFields...)
		result[index].CollectionFields = append([]contractartifacts.CollectionField(nil), surface.CollectionFields...)
		for fieldIndex := range result[index].CollectionFields {
			result[index].CollectionFields[fieldIndex].AllowedOperations = append([]string(nil), surface.CollectionFields[fieldIndex].AllowedOperations...)
		}
	}
	return result
}
