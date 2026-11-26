package childenv

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestChildEnvironment(t *testing.T) {
	root := t.TempDir()
	for path, body := range map[string]string{
		"tools/harness_environment_registry.json": `{"inherited":{"names":["HOME","PATH"],"prefixes":["LC_"]}}`,
		"tools/task_surface_manifest.json":        `{"global_inputs":[{"name":"GO_CACHE_DIR"}]}`,
	} {
		if err := os.MkdirAll(filepath.Join(root, filepath.Dir(path)), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// Harness and test context names are assembled so this file names none.
	prefix := "CARTULARY" + "_"
	parent := []string{"PATH=/bin", "HOME=/home/operator", "LC_ALL=C.UTF-8", "GO_CACHE_DIR=/cache", "CGO_ENABLED=0",
		"AGENT_TOKEN=operator secret", prefix + "PGTEST_DSN_TEMPLATE=postgres://fixture", prefix + "TEST_RUN_ID=unit", "malformed"}
	child, err := Build(root, parent, []string{prefix + "TEST_RUN_ID"}, map[string]string{"HOME": "/srv", prefix + "HTTP_PORT": "0"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{prefix + "HTTP_PORT=0", prefix + "TEST_RUN_ID=unit", "GO_CACHE_DIR=/cache", "HOME=/srv", "LC_ALL=C.UTF-8", "PATH=/bin"}
	if !slices.Equal(child, want) {
		t.Fatalf("child environment %q, want %q", child, want)
	}
	if _, err := Build(t.TempDir(), parent, nil, nil); err == nil {
		t.Fatal("a root without declarations must not build an environment")
	}
}
