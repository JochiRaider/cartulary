package incidents

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const incidentsRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestIncidentsProductionImportBoundaries(t *testing.T) {
	forbiddenImports := map[string]string{
		incidentsRepoImportPrefix + "internal/modules/workbook/startup": "Incidents must own preference persistence and bootstrap",
		incidentsRepoImportPrefix + "internal/platform/ws":              "incident routes must use the collaboration/session port",
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read incidents package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		for _, importPath := range incidentsProductionImports(t, fileName) {
			if reason, ok := forbiddenImports[importPath]; ok {
				t.Fatalf("%s imports %s; %s", fileName, importPath, reason)
			}
		}
	}
}

func incidentsProductionImports(t testing.TB, fileName string) []string {
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
