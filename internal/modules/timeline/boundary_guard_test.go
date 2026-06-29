package timeline

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const repoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestTimelineProductionImportBoundaries(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read timeline package directory: %v", err)
	}

	allowedImports := map[string]map[string]bool{
		"github.com/JochiRaider/cartulary/internal/modules/auth": {
			"api.go":             true,
			"clipboard_paste.go": true,
			"routes.go":          true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/entities": {
			"mentions_collections_store.go": true,
			"ports.go":                      true,
			"routes.go":                     true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/incidents": {
			"routes.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest": {
			"clipboard_paste.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/links": {
			"ports.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/projections": {
			"ports.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/records": {
			"ports.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/revisions": {
			"ports.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/platform/httpapi": {
			"routes.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/platform/ws": {
			"routes.go": true,
		},
	}
	disallowedParserImports := []string{
		"archive/zip",
		"encoding/xml",
		"github.com/xuri/excelize",
		"github.com/tealeg/xlsx",
		"openxml",
	}

	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		fileImports := productionImports(t, fileName)
		for _, importPath := range fileImports {
			if allowedFiles, ok := allowedImports[importPath]; ok {
				if !allowedFiles[fileName] {
					t.Fatalf("%s imports %s; allowed files are %v", fileName, importPath, allowedFileNames(allowedFiles))
				}
			} else if strings.HasPrefix(importPath, repoImportPrefix+"internal/modules/") {
				t.Fatalf("%s imports unapproved sibling module %s", fileName, importPath)
			}
			for _, disallowed := range disallowedParserImports {
				if strings.Contains(importPath, disallowed) {
					t.Fatalf("%s imports parser dependency %s; Timeline paste must stay on shared tabularingest", fileName, importPath)
				}
			}
		}
	}
}

func productionImports(t testing.TB, fileName string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(fileName), nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", fileName, err)
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		importPath := strings.Trim(spec.Path.Value, "\"")
		if !strings.HasPrefix(importPath, repoImportPrefix) {
			imports = append(imports, importPath)
			continue
		}
		imports = append(imports, importPath)
	}
	return imports
}

func allowedFileNames(files map[string]bool) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	return names
}
