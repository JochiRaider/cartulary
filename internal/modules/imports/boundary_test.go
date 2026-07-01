package imports

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportsProductionPackageDoesNotImportWorkbook(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob imports package files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(data), "internal/modules/workbook") {
			t.Fatalf("%s imports workbook; import apply must dispatch through target owners", file)
		}
	}
}
