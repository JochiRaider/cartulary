package projections

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	cartularyImportPrefix = "github.com/JochiRaider/cartulary/"
	projectionsImportPath = cartularyImportPrefix + "internal/modules/projections"
)

var approvedProductionProjectionImporters = map[string]struct{}{
	"internal/app/operator.go":                             {},
	"internal/modules/artifacts/import_projection.go":      {},
	"internal/modules/artifacts/linkednotes/facade.go":     {},
	"internal/modules/assessments/store.go":                {},
	"internal/modules/entities/store.go":                   {},
	"internal/modules/evidence/import_projection.go":       {},
	"internal/modules/evidence/store.go":                   {},
	"internal/modules/incidentbundles/source.go":           {},
	"internal/modules/parties/store.go":                    {},
	"internal/modules/revisions/delete_restore_store.go":   {},
	"internal/modules/revisions/rollback_store.go":         {},
	"internal/modules/tasksdecisions/import_projection.go": {},
	"internal/modules/tasksdecisions/supersede_facade.go":  {},
	"internal/modules/timeline/ports.go":                   {},
	"internal/modules/workbook/mutation_store.go":          {},
	"internal/modules/workbook/store.go":                   {},
}

func TestProductionProjectionImportsUseApprovedFacades(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	scanRoots := []string{
		filepath.Join(repoRoot, "cmd"),
		filepath.Join(repoRoot, "internal", "app"),
		filepath.Join(repoRoot, "internal", "modules"),
	}
	projectionsDir := filepath.Clean(filepath.Join(repoRoot, "internal", "modules", "projections"))

	for _, root := range scanRoots {
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("stat scan root %s: %v", root, err)
		}
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if filepath.Clean(path) == projectionsDir {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			checkProjectionImports(t, repoRoot, path)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
}

func checkProjectionImports(t *testing.T, repoRoot, filePath string) {
	t.Helper()

	relPath, err := filepath.Rel(repoRoot, filePath)
	if err != nil {
		t.Fatalf("rel path for %s: %v", filePath, err)
	}
	relPath = filepath.ToSlash(relPath)

	parsed, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", relPath, err)
	}

	for _, importSpec := range parsed.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import in %s: %v", relPath, err)
		}
		switch {
		case importPath == projectionsImportPath:
			if _, ok := approvedProductionProjectionImporters[relPath]; !ok {
				t.Fatalf("%s imports projections directly without owner approval", relPath)
			}
		case strings.HasPrefix(importPath, projectionsImportPath+"/"):
			t.Fatalf("%s imports projection internal package %s", relPath, importPath)
		case strings.HasPrefix(importPath, cartularyImportPrefix+"internal/modules/") &&
			strings.Contains(importPath, "/projectionprovider"):
			t.Fatalf("%s imports projection provider internal package %s", relPath, importPath)
		case strings.Contains(importPath, "/projections/testfixtures") ||
			strings.Contains(importPath, "/projections/test"):
			t.Fatalf("%s imports projection test fixture package %s", relPath, importPath)
		}
	}
}
