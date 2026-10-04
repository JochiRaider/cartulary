package packformat

import (
	"os"
	"path/filepath"
	"testing"
)

// Fixture scratch always lives under the test runner temporary directory.
func fixtureRelationScratch(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	if err := os.Mkdir(filepath.Join(base, "fixtures"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "fixtures", "input.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("fixtures", filepath.Join(base, "linked")); err != nil {
		t.Fatal(err)
	}

	return base
}
