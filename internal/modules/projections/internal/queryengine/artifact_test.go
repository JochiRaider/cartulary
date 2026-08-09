package queryengine

import (
	"fmt"
	"testing"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestArtifactProjectionProviderSurfaceContractMatrix(t *testing.T) {
	t.Parallel()

	want := make(map[string]string)
	intents, err := artifactprojection.SurfaceIntents()
	if err != nil {
		t.Fatalf("Artifact semantic intents: %v", err)
	}
	for _, intent := range intents {
		if intent.CanonicalSourceFilter == nil {
			t.Fatalf("Artifact intent %s has no canonical source filter", intent.ViewSchemaID)
		}
		want[intent.ViewSchemaID] = intent.CanonicalSourceFilter.Value
	}
	surfaces := ArtifactPlans()
	if len(surfaces) != len(want) {
		t.Fatalf("Plans() returned %d surfaces, want %d", len(surfaces), len(want))
	}
	seen := make(map[string]struct{}, len(surfaces))
	for _, surface := range surfaces {
		artifactType, ok := want[surface.ViewSchemaID]
		if !ok {
			t.Fatalf("projection provider admitted unknown artifact surface %q", surface.ViewSchemaID)
		}
		if _, duplicate := seen[surface.ViewSchemaID]; duplicate {
			t.Fatalf("projection provider duplicated artifact surface %q", surface.ViewSchemaID)
		}
		seen[surface.ViewSchemaID] = struct{}{}
		if got, expected := surface.WhereSQL, fmt.Sprintf("p.artifact_type = '%s'", artifactType); got != expected {
			t.Fatalf("%s WhereSQL = %q, want %q", surface.ViewSchemaID, got, expected)
		}
		schema, ok := viewschema.Lookup(surface.ViewSchemaID)
		if !ok {
			t.Fatalf("viewschema.Lookup(%q) not found", surface.ViewSchemaID)
		}
		filter, ok := schema.CanonicalSourceFilter()
		if !ok || filter.Value != artifactType {
			t.Fatalf("%s canonical filter = %#v, %v; want %q", surface.ViewSchemaID, filter, ok, artifactType)
		}
		if len(surface.Fields) == 0 {
			t.Fatalf("%s has no query fields", surface.ViewSchemaID)
		}
	}
}
