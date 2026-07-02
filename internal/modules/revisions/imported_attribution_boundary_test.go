package revisions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRevisionsDoNotJoinIncidentBundleAttributionSidecarsDirectly(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		fileName := filepath.Base(path)
		if !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(body), "incident_bundle_imported_attributions") {
			t.Fatalf("%s must use the imported attribution resolver port instead of joining incident bundle sidecars", filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk revisions files: %v", err)
	}
}
