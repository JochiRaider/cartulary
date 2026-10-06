package testsupport

import (
	"maps"
	"os"
	"path/filepath"
	"testing"
)

func TestTargetFixtureCopiesCallerEnvironment_Unit(t *testing.T) {
	env := map[string]string{"DATABASE_URL": "borrowed", "SECRET": "ephemeral"}
	fixture := NewTargetFixture(env, nil, nil)

	env["DATABASE_URL"] = "mutated"
	delete(env, "SECRET")
	if maps.Equal(fixture.Env, env) || fixture.Env["DATABASE_URL"] != "borrowed" || fixture.Env["SECRET"] != "ephemeral" {
		t.Fatalf("target fixture did not retain an independent environment copy: %#v", fixture.Env)
	}
	fixture.Cleanup()
	fixture.Cleanup()
	if len(fixture.Env) != 0 {
		t.Fatalf("idempotent cleanup retained copied environment values: %#v", fixture.Env)
	}
}

func TestEvidenceArtifactUsesPrivatePermissions_Unit(t *testing.T) {
	location := EvidenceLocation{
		ResultsRoot: t.TempDir(),
		RunID:       "run",
		Target:      "backend-process",
		Group:       "backup-restore",
	}
	path := WriteEvidenceArtifact(t, location, "proof.json", []byte("safe"))

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat evidence directory: %v", err)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat evidence file: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("evidence directory mode got %#o want 0700", got)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("evidence file mode got %#o want 0600", got)
	}
}

func TestEvidenceLocationRejectsNestedExecutionSegments_Unit(t *testing.T) {
	location := EvidenceLocation{
		ResultsRoot: t.TempDir(),
		RunID:       "nested/run",
		Target:      "backend-process",
		Group:       "backup-restore",
	}
	if _, err := location.Dir(); err == nil {
		t.Fatal("nested run ID was accepted")
	}
}
