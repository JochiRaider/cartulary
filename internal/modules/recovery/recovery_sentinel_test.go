package recovery

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPublicRouteAbsenceStaticInventory_Unit(t *testing.T) {
	forbiddenFamilies := []string{
		"/api/v1/backups",
		"/api/v1/restores",
		"/api/v1/restore-verifications",
		"/ws/v1/backups",
		"/ws/v1/restores",
		"/ws/v1/restore-verifications",
	}
	scanRoots := []string{
		filepath.Join(repoRoot(), "cmd", "server"),
		filepath.Join(repoRoot(), "internal", "app"),
		filepath.Join(repoRoot(), "internal", "modules"),
		filepath.Join(repoRoot(), "internal", "platform", "httpapi"),
	}
	for _, root := range scanRoots {
		root := root
		if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			content, err := os.ReadFile(path) // #nosec G304 -- test scans repo-local authored Go files for route-family literals.
			if err != nil {
				return err
			}
			for _, forbidden := range forbiddenFamilies {
				if strings.Contains(string(content), forbidden) {
					t.Fatalf("public backup/restore route family %q found in %s", forbidden, path)
				}
			}
			return nil
		}); err != nil {
			t.Fatalf("scan route inventory root %s: %v", root, err)
		}
	}
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}
