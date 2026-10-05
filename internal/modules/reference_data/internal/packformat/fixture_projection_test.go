package packformat

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractreferencepackfixtures"
	"github.com/JochiRaider/cartulary/internal/gen/contractreferencepacks"
)

// Test assets have their own resolver. Production projection lookup never
// searches this collection, including when a runtime projection is missing.
func fixtureProjection(name string) []byte {
	artifact, ok := contractreferencepackfixtures.Index["contracts/reference-pack-fixtures/"+name]
	if !ok {
		panic("missing reference pack test projection: " + name)
	}
	return []byte(artifact.JSON)
}

func compileFixtureProjection(name string) *shape {
	var value map[string]any
	if err := json.Unmarshal(fixtureProjection(name), &value); err != nil {
		panic("invalid reference pack fixture schema")
	}
	return compileShape(value, map[string]bool{})
}

func testFixtureProjectionIsolation(t *testing.T) {
	t.Helper()
	for _, artifact := range contractreferencepacks.Artifacts {
		if strings.Contains(artifact.Path, "/fixtures/") || strings.Contains(artifact.Path, "/fixture-manifests/") || strings.Contains(artifact.Path, "_fixture.") {
			t.Fatal("test asset included in runtime projection", artifact.Path)
		}
	}
	// An existing test schema must remain unavailable to the runtime resolver.
	defer func() {
		if recover() == nil {
			t.Fatal("runtime projection lookup fell back to test assets")
		}
	}()
	projection("content_fixture.v1.schema.json")
}
