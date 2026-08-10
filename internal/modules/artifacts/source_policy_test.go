package artifacts

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractartifacts"
)

func TestArtifactSourceCatalogRuntimeRejectsDrift(t *testing.T) {
	t.Parallel()
	if catalog, err := buildArtifactSourcePolicyCatalog(cloneGeneratedArtifactCatalog()); err != nil || len(catalog.surfaces) != 8 || len(catalog.fields) != 51 {
		t.Fatalf("canonical generated catalog = %d surfaces, %d fields, err=%v", len(catalog.surfaces), len(catalog.fields), err)
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
		{name: "unknown collection operation", mutate: func(catalog []contractartifacts.SourceSurface) []contractartifacts.SourceSurface {
			catalog[0].CollectionFields[0].AllowedOperations[0] = "future_operation"
			return catalog
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := buildArtifactSourcePolicyCatalog(tc.mutate(cloneGeneratedArtifactCatalog())); err == nil {
				t.Fatalf("runtime catalog accepted %s drift", tc.name)
			}
		})
	}
}

func cloneGeneratedArtifactCatalog() []contractartifacts.SourceSurface {
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
