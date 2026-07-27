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
	manifest := `{"schema_id":"cartulary.graph_projection_fixture_manifest.v3","fixture_version":3,"fixture_id":"GP-FIX-001","execution_layer":"backend_unit","test_symbol":"TestGPFIX001","determinism":{"clock":"2026-05-30T00:00:00Z","nonce":"fixture","cursor_key_hex":"0000000000000000000000000000000000000000000000000000000000000000"},"comparison":{"scope":"run_specific","mode":"exact_artifacts","run_independent_validation":false},"state_effects":{"mode":"no_retained_state_change"},"steps":[{"id":"verify","operation":"project_ephemeral","input_artifact":"input.json","expected_artifact":"input.json"}],"artifacts":[{"path":"input.json","role":"input","sha256":"` + hex.EncodeToString(digest[:]) + `"}],"golden":{"review_status":"reviewed"}}`
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

func TestVerifyExecutesAndRejectsSemanticMismatch(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module fixturetest\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(root, "contracts", "graph-projection", "fixtures", "GP-FIX-001")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	input := []byte("{}\n")
	expected := []byte("{\"status\":\"ok\"}\n")
	inputDigest := sha256.Sum256(input)
	expectedDigest := sha256.Sum256(expected)
	manifest := `{"schema_id":"cartulary.graph_projection_fixture_manifest.v3","fixture_version":3,"fixture_id":"GP-FIX-001","execution_layer":"backend_unit","test_symbol":"TestGPFIX001","determinism":{"clock":"2026-05-30T00:00:00Z","nonce":"fixture","cursor_key_hex":"0000000000000000000000000000000000000000000000000000000000000000"},"comparison":{"scope":"run_specific","mode":"exact_artifacts","run_independent_validation":false},"state_effects":{"mode":"no_retained_state_change"},"steps":[{"id":"verify","operation":"project_ephemeral","input_artifact":"input.json","expected_artifact":"expected.json"}],"artifacts":[{"path":"input.json","role":"input","sha256":"` + hex.EncodeToString(inputDigest[:]) + `"},{"path":"expected.json","role":"expected_response","sha256":"` + hex.EncodeToString(expectedDigest[:]) + `"}],"golden":{"review_status":"reviewed"}}`
	for name, body := range map[string][]byte{"fixture.json": []byte(manifest), "input.json": input, "expected.json": expected} {
		if err := os.WriteFile(filepath.Join(directory, name), body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	executor := ExecutorFunc(func(_ Manifest, _ Step, got []byte) (StepExecution, error) {
		if string(got) != string(input) {
			t.Fatalf("executor input = %q", got)
		}
		return StepExecution{Artifact: []byte("{\"status\":\"ok\"}"), StateEffectMode: "no_retained_state_change"}, nil
	})
	if err := Verify(root, "GP-FIX-001", executor); err != nil {
		t.Fatalf("verify matching artifact: %v", err)
	}
	mismatch := ExecutorFunc(func(Manifest, Step, []byte) (StepExecution, error) {
		return StepExecution{Artifact: []byte("{\"status\":\"error\"}"), StateEffectMode: "no_retained_state_change"}, nil
	})
	if err := Verify(root, "GP-FIX-001", mismatch); err == nil {
		t.Fatal("semantic mismatch unexpectedly passed")
	}
}

func TestContainedPathRejectsTraversal(t *testing.T) {
	if _, err := cleanRelativePath("../outside.json"); err == nil {
		t.Fatal("traversal path unexpectedly accepted")
	}
}
