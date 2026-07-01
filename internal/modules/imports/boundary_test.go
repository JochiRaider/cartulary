package imports

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportsProductionPackageUsesOwnerFacadesForWorkbookRows(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob imports package files: %v", err)
	}
	forbiddenImports := map[string]string{
		"internal/modules/workbook":    "workbook",
		"internal/modules/projections": "projections",
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for importPath, moduleName := range forbiddenImports {
			if strings.Contains(string(data), importPath) {
				t.Fatalf("%s imports %s; import apply must dispatch through target owners", file, moduleName)
			}
		}
	}
}
