package indicators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const indicatorsRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestIndicatorsProductionImportBoundaries(t *testing.T) {
	allowedSiblingImports := map[string]map[string]bool{
		indicatorsRepoImportPrefix + "internal/modules/imports/ownerfacade": {
			"import_create.go": true,
		},
		indicatorsRepoImportPrefix + "internal/modules/imports/tabularingest": {
			"import_create.go": true,
			"query.go":         true,
		},
		indicatorsRepoImportPrefix + "internal/modules/incidents": {
			"store.go": true,
		},
		indicatorsRepoImportPrefix + "internal/modules/records": {
			"api.go":   true,
			"store.go": true,
		},
		indicatorsRepoImportPrefix + "internal/modules/revisions": {
			"api.go":   true,
			"store.go": true,
		},
	}

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
		for _, importPath := range indicatorsProductionImports(t, path) {
			if !strings.HasPrefix(importPath, indicatorsRepoImportPrefix+"internal/modules/") {
				continue
			}
			allowedFiles, ok := allowedSiblingImports[importPath]
			if !ok {
				t.Fatalf("%s imports unapproved sibling module %s", filepath.ToSlash(path), importPath)
			}
			if !allowedFiles[fileName] {
				t.Fatalf("%s imports %s; allowed files are %v", filepath.ToSlash(path), importPath, indicatorsAllowedFileNames(allowedFiles))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk indicators package: %v", err)
	}
}

func TestIndicatorsDoNotUseEntitiesSourcePrefixes(t *testing.T) {
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
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(content), "entities.indicators.") {
			t.Fatalf("%s uses legacy entities indicator source prefix", filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk indicators package: %v", err)
	}
}

func indicatorsProductionImports(t testing.TB, fileName string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(fileName), nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", fileName, err)
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import for %s: %v", fileName, err)
		}
		imports = append(imports, importPath)
	}
	return imports
}

func indicatorsAllowedFileNames(files map[string]bool) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	return names
}
