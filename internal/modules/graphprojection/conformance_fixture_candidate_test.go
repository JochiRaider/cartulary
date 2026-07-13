package graphprojection

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/fixturetest"
)

// TestGraphProjectionFixtureCandidate is intentionally excluded from normal
// backend targets. The Make wrapper supplies one fixture ID and writes only a
// disposable candidate under the test-results root.
func TestGraphProjectionFixtureCandidate(t *testing.T) {
	fixtureID := os.Getenv("GRAPH_PROJECTION_FIXTURE")
	if fixtureID == "" {
		t.Skip("candidate mode is explicit")
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := fixturetest.RepoRoot(workingDirectory)
	if err != nil {
		t.Fatal(err)
	}
	manifest, directory, err := fixturetest.Load(root, fixtureID)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Steps[0].Operation != "project_ephemeral" {
		t.Fatalf("candidate mode does not yet support %s", manifest.Steps[0].Operation)
	}
	input, err := fixturetest.ReadArtifact(directory, manifest.Steps[0].InputArtifact)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(ServiceOptions{
		Now:      func() time.Time { return time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC) },
		NewNonce: func() (string, error) { return manifest.Determinism.Nonce, nil },
	})
	result, err := service.ProjectEphemeral(context.Background(), EphemeralProjectionRequest{ProjectionInput: input})
	if err != nil {
		t.Fatalf("candidate projection: %v", err)
	}
	body, err := json.MarshalIndent(result.Resource(), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	resultsRoot := os.Getenv("CARTULARY_TEST_RESULTS_DIR")
	if resultsRoot == "" {
		t.Fatal("CARTULARY_TEST_RESULTS_DIR is required for candidate mode")
	}
	directory = filepath.Join(resultsRoot, "graph-projection-fixture-candidate", fixtureID)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "candidate-response.json"), append(body, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}
