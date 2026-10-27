package revisions

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/sourcecheck"
)

func TestRevisionsApplicationAndHTTPBoundaries_Unit(t *testing.T) {
	files, err := sourcecheck.ReadImports(os.DirFS("."))
	if err != nil {
		t.Fatal(err)
	}
	selected := map[string]int{".": 0, "conflicts": 0}
	for _, file := range files {
		scope := path.Dir(file.Path)
		if scope != "." && !strings.HasPrefix(file.Path, "conflicts/") {
			continue
		}
		if scope != "." {
			scope = "conflicts"
		}
		selected[scope]++
		for _, imported := range file.Imports {
			if !revisionsApplicationImportAllowed(imported) {
				t.Errorf("%s imports transport or concrete storage dependency %s", file.Path, imported)
			}
		}
	}
	for scope, count := range selected {
		if count == 0 {
			t.Errorf("no production files for %s", scope)
		}
	}
	for _, fixture := range []struct {
		dependency string
		allowed    bool
	}{
		{"net/http", false}, {"github.com/JochiRaider/cartulary/internal/platform/authn", false},
		{"github.com/JochiRaider/cartulary/internal/platform/postgres", false},
		{"github.com/JochiRaider/cartulary/internal/modules/records", false},
		{"github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract", true},
		{"github.com/jackc/pgx/v5", true},
	} {
		for _, name := range []string{"renamed.go", "nested/extracted.go"} {
			imports, err := sourcecheck.ParseImports(name, []byte("package application\nimport renamed `"+fixture.dependency+"`"))
			if err != nil {
				t.Fatal(err)
			}
			if revisionsApplicationImportAllowed(imports[0]) != fixture.allowed {
				t.Errorf("%s imports %s: want allowed=%v", name, fixture.dependency, fixture.allowed)
			}
		}
	}
}
func revisionsApplicationImportAllowed(imported string) bool {
	switch imported {
	case "net/http", "github.com/JochiRaider/cartulary/internal/platform/authn", "github.com/JochiRaider/cartulary/internal/platform/httpapi", "github.com/JochiRaider/cartulary/internal/platform/pagination", "github.com/JochiRaider/cartulary/internal/platform/postgres", "github.com/JochiRaider/cartulary/internal/modules/records":
		return false
	default:
		return true
	}
}
