package reporting

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const reportingRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestReportingProductionImportBoundaries(t *testing.T) {
	allowedSiblingImports := map[string]map[string]bool{
		reportingRepoImportPrefix + "internal/modules/incidents": {
			"routes.go": true,
		},
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read reporting package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		for _, importPath := range reportingProductionImports(t, fileName) {
			if !strings.HasPrefix(importPath, reportingRepoImportPrefix+"internal/modules/") {
				continue
			}
			allowedFiles, ok := allowedSiblingImports[importPath]
			if !ok {
				continue
			}
			if !allowedFiles[fileName] {
				t.Fatalf("%s imports %s; allowed files are %v", fileName, importPath, reportingAllowedFileNames(allowedFiles))
			}
		}
	}
}

func reportingProductionImports(t testing.TB, fileName string) []string {
	t.Helper()

	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(fileName), nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", fileName, err)
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		imports = append(imports, strings.Trim(spec.Path.Value, `"`))
	}
	return imports
}

func reportingAllowedFileNames(allowed map[string]bool) []string {
	files := make([]string, 0, len(allowed))
	for fileName := range allowed {
		files = append(files, fileName)
	}
	return files
}
