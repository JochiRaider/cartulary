package packformat

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// These independently specified graph inputs exercise the dependency boundary
// after authenticated tuple resolution. They do not purport to be signed packs
// or to establish that a digest-bound cyclic archive is constructible.
func runDependencyManifest(t *testing.T, input []byte, expected map[string]any) {
	t.Helper()
	value, err := canonicaljson.DecodeStrict(input)
	if err != nil || !compileFixtureProjection("dependency_fixture.v1.schema.json").matches(value) {
		t.Fatal("invalid dependency fixture", err)
	}
	var fixture struct {
		Candidate Manifest   `json:"candidate"`
		Retained  []Manifest `json:"retained"`
	}
	if err := json.Unmarshal(input, &fixture); err != nil {
		t.Fatal(err)
	}
	resolve := func(_ context.Context, dependency Dependency) (Manifest, bool, error) {
		for _, m := range fixture.Retained {
			if m.Key == dependency.Key && m.Version == dependency.Version {
				return m, true, nil
			}
		}
		return Manifest{}, false, nil
	}
	err = VerifyDependencies(context.Background(), fixture.Candidate, resolve, newDiagnosticTestScratch())
	assertManifestVerdict(t, expected, err)
}
