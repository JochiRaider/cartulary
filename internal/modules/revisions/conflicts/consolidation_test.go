package conflicts

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestConflictCapabilitiesAreConsolidated_Unit(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve conflict consolidation test source")
	}
	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(source), "../../../.."))
	retiredPackages := []string{"conflictmerge", "conflictresolution", "conflicttokens", "conflictwindows"}
	for _, retiredPackage := range retiredPackages {
		path := filepath.Join(repositoryRoot, "internal/modules/revisions", retiredPackage)
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("retired conflict package %s still exists: %v", path, err)
		}
	}

	for _, scanRoot := range []string{filepath.Join(repositoryRoot, "internal/modules"), filepath.Join(repositoryRoot, "internal/app")} {
		err := filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || filepath.Ext(path) != ".go" {
				return nil
			}
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, retiredPackage := range retiredPackages {
				retiredImport := "internal/modules/revisions/" + retiredPackage
				if strings.Contains(string(contents), retiredImport) {
					t.Errorf("%s imports retired conflict package %s", path, retiredPackage)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan production imports under %s: %v", scanRoot, err)
		}
	}

	entries, err := os.ReadDir(filepath.Join(repositoryRoot, "internal/modules/revisions/conflicts"))
	if err != nil {
		t.Fatalf("read consolidated conflicts package: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(repositoryRoot, "internal/modules/revisions/conflicts", entry.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		forbiddenOwnerImport := "github.com/JochiRaider/cartulary/internal/modules/"
		if strings.Contains(string(contents), forbiddenOwnerImport) {
			t.Errorf("generic conflict capability %s imports a source-owner module", entry.Name())
		}
	}
}
