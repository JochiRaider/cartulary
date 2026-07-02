package entities

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const entitiesRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestEntitiesProductionImportBoundaries(t *testing.T) {
	allowedSiblingImports := map[string]map[string]bool{
		entitiesRepoImportPrefix + "internal/modules/incidents": {
			"api_errors.go": true,
			"routes.go":     true,
			"store.go":      true,
		},
		entitiesRepoImportPrefix + "internal/modules/links": {
			"merge_store.go": true,
			"store.go":       true,
		},
		entitiesRepoImportPrefix + "internal/modules/projections/adapters": {
			"merge_store.go": true,
			"store.go":       true,
		},
		entitiesRepoImportPrefix + "internal/modules/records": {
			"store.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/revisions": {
			"mention_lifecycle.go": true,
			"merge_store.go":       true,
			"patch_store.go":       true,
			"store.go":             true,
		},
		entitiesRepoImportPrefix + "internal/modules/timeline/rowsnapshot": {
			"mention_lifecycle.go": true,
		},
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read entities package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		for _, importPath := range entitiesProductionImports(t, fileName) {
			if !strings.HasPrefix(importPath, entitiesRepoImportPrefix+"internal/modules/") {
				continue
			}
			allowedFiles, ok := allowedSiblingImports[importPath]
			if !ok {
				t.Fatalf("%s imports unapproved sibling module %s", fileName, importPath)
			}
			if !allowedFiles[fileName] {
				t.Fatalf("%s imports %s; allowed files are %v", fileName, importPath, entitiesAllowedFileNames(allowedFiles))
			}
		}
	}
}

func TestEntitiesDoNotRegisterWorkbookRowCreateRoutes(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean("routes.go"))
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}
	content := string(body)
	for _, route := range []string{
		"views/cartulary.view.hosts.v1/rows",
		"views/cartulary.view.identities.v1/rows",
		"views/cartulary.view.indicators.v1/rows",
	} {
		if strings.Contains(content, route) {
			t.Fatalf("entities routes.go still registers workbook row-create route %s", route)
		}
	}
}

func entitiesProductionImports(t testing.TB, fileName string) []string {
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

func entitiesAllowedFileNames(files map[string]bool) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	return names
}
