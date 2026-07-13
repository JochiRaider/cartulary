package fixturetest

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRejectsUnsafeAndCorruptArtifacts(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module fixturetest\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(root, "contracts", "graph-projection", "fixtures", "GP-FIX-001")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	input := []byte("{}\n")
	digest := sha256.Sum256(input)
	if err := os.WriteFile(filepath.Join(directory, "input.json"), input, 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_id":"cartulary.graph_projection_fixture_manifest.v1","fixture_version":1,"fixture_id":"GP-FIX-001","execution_layer":"backend_unit","owner_sections":["4"],"acceptance_ids":["GP-AC-001"],"test_symbol":"TestGPFIX001","determinism":{"clock":"2026-05-30T00:00:00Z","nonce":"fixture","cursor_key_hex":"0000000000000000000000000000000000000000000000000000000000000000"},"steps":[{"id":"verify","operation":"project_ephemeral","input_artifact":"input.json"}],"artifacts":[{"path":"input.json","role":"input","sha256":"` + hex.EncodeToString(digest[:]) + `"}],"golden":{"provenance":"nlspec_derived","review_status":"reviewed"}}`
	if err := os.WriteFile(filepath.Join(directory, "fixture.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(root, "GP-FIX-001"); err != nil {
		t.Fatalf("load valid fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "input.json"), []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(root, "GP-FIX-001"); err == nil {
		t.Fatal("corrupt fixture artifact unexpectedly loaded")
	}
}

func TestContainedPathRejectsTraversal(t *testing.T) {
	if _, err := containedPath("/tmp/fixture", "../outside.json"); err == nil {
		t.Fatal("traversal path unexpectedly accepted")
	}
}
