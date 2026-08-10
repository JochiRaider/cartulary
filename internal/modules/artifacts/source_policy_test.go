package artifacts

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
)

func TestArtifactSourceCatalogRuntimeRejectsDrift(t *testing.T) {
	t.Parallel()
	catalog, err := sourcecatalog.Load()
	if err != nil {
		t.Fatalf("sourcecatalog.Load() error = %v", err)
	}
	if len(catalog.Surfaces()) != 8 || len(catalog.Fields()) != 51 || len(catalog.WritableDirectStorageMappings()) != 36 {
		t.Fatalf(
			"canonical source catalog = %d surfaces, %d fields, %d direct mappings",
			len(catalog.Surfaces()), len(catalog.Fields()), len(catalog.WritableDirectStorageMappings()),
		)
	}
}
