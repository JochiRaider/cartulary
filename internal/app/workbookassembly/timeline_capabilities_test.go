package workbookassembly

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

var _ TimelineOperations = (*timeline.Facade)(nil)

func TestTimelineOperationsCapabilityBoundary_Unit(t *testing.T) {
	wantMethods := []string{
		"ApplyClipboardPaste",
		"ApplyFillDown",
		"ApplyMultiRowTagAssignment",
		"CreateRow",
		"PatchRow",
		"ResolveConflict",
		"SupersedeRow",
	}
	capability := reflect.TypeOf((*TimelineOperations)(nil)).Elem()
	gotMethods := make([]string, 0, capability.NumMethod())
	for index := range capability.NumMethod() {
		gotMethods = append(gotMethods, capability.Method(index).Name)
	}
	if !slices.Equal(gotMethods, wantMethods) {
		t.Fatalf("TimelineOperations methods = %#v, want %#v", gotMethods, wantMethods)
	}

	var facade *timeline.Facade
	var typedNil TimelineOperations = facade
	if _, err := newTimelineProviderSet(typedNil); err == nil || err.Error() != "compose Timeline Workbook adapters: owner is required" {
		t.Fatalf("typed-nil Timeline owner must preserve the construction error, got %v", err)
	}
	if !isNilContributionDependency(typedNil) {
		t.Fatal("typed-nil Timeline capability was not rejected by contribution validation")
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Workbook assembly package: %v", err)
	}
	files := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(files, entry.Name(), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			pointer, ok := node.(*ast.StarExpr)
			if !ok {
				return true
			}
			selector, ok := pointer.X.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Facade" {
				return true
			}
			owner, ok := selector.X.(*ast.Ident)
			if ok && owner.Name == "timeline" {
				t.Errorf("production Workbook assembly contains concrete *timeline.Facade in %s", entry.Name())
			}
			return true
		})
	}
}
