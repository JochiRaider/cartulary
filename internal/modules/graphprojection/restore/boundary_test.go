package restore_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	. "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
)

func TestGraphRestorePackageProductionBoundaryAllowlist_Unit(t *testing.T) {
	wantFiles := map[string]bool{
		"contract.go":       false,
		"recovery_state.go": false,
		"service.go":        false,
	}
	wantImports := map[string]int{
		"github.com/JochiRaider/cartulary/internal/gen/contractrecovery":    2,
		"github.com/JochiRaider/cartulary/internal/modules/graphprojection": 2,
		"github.com/JochiRaider/cartulary/internal/platform/recoverystate":  1,
		"github.com/google/uuid": 2,
	}
	gotImports := make(map[string]int)
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Graph restore package: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if _, admitted := wantFiles[name]; !admitted {
			t.Fatalf("Graph restore package has unowned production file %s", name)
		}
		wantFiles[name] = true
		parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse Graph restore imports for %s: %v", name, err)
		}
		for _, spec := range parsed.Imports {
			importPath := strings.Trim(spec.Path.Value, "\"")
			if strings.Contains(strings.Split(importPath, "/")[0], ".") {
				gotImports[importPath]++
			}
		}
	}
	for name, present := range wantFiles {
		if !present {
			t.Fatalf("Graph restore required production file %s is missing", name)
		}
	}
	if len(gotImports) != len(wantImports) {
		t.Fatalf("Graph restore non-standard imports = %#v, want %#v", gotImports, wantImports)
	}
	for importPath, wantCount := range wantImports {
		if gotImports[importPath] != wantCount {
			t.Fatalf("Graph restore import %s count = %d, want %d", importPath, gotImports[importPath], wantCount)
		}
	}
}

func TestGraphProjectionCurrentRecoveryContributionIsV4Only_Unit(t *testing.T) {
	contribution := RecoveryStateContribution()
	if contribution.OwnerID != "module.graphprojection" {
		t.Fatalf("Recovery owner = %q", contribution.OwnerID)
	}
	wantTables := []string{
		"graph_projection_result_edges",
		"graph_projection_result_leases",
		"graph_projection_result_vertices",
		"graph_projection_results",
	}
	if len(contribution.Tables) != len(wantTables) {
		t.Fatalf("Recovery table count = %d, want %d", len(contribution.Tables), len(wantTables))
	}
	gotTables := make([]string, 0, len(contribution.Tables))
	for _, table := range contribution.Tables {
		gotTables = append(gotTables, table.TableName)
		if table.AlgorithmID == nil || *table.AlgorithmID != RestoreAlgorithmID {
			t.Fatalf("Recovery table %s algorithm = %v, want %s", table.TableName, table.AlgorithmID, RestoreAlgorithmID)
		}
	}
	slices.Sort(gotTables)
	if !slices.Equal(gotTables, wantTables) {
		t.Fatalf("Recovery tables = %v, want %v", gotTables, wantTables)
	}
}

func TestGraphRestoreExportedFunctionAllowlist_Unit(t *testing.T) {
	want := map[string]bool{
		"func CurrentRestoreImplementationBinding":       false,
		"func CurrentRestoreSourceRegistry":              false,
		"func FrozenRestoreImplementationBinding":        false,
		"func NewCurrentRestoreSourceRegistry":           false,
		"func NewRestoreError":                           false,
		"func NewRestoreService":                         false,
		"func RecoveryStateContribution":                 false,
		"func RestoreErrorCodeOf":                        false,
		"func RestoreGraphTableIDs":                      false,
		"method RestoreError.Error":                      false,
		"method RestorePublicationError.Error":           false,
		"method RestorePublicationError.Unwrap":          false,
		"method RestoreRebuildResult.ReadinessSatisfied": false,
		"method RestoreRebuildResult.Validate":           false,
		"method RestoreService.Rebuild":                  false,
		"method RestoreSourceRegistry.DigestSHA256":      false,
		"method RestoreSourceRegistry.Document":          false,
		"method RestoreSourceRegistry.Registrations":     false,
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Graph restore package: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(name), nil, 0)
		if err != nil {
			t.Fatalf("parse Graph restore declarations for %s: %v", name, err)
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || !function.Name.IsExported() {
				continue
			}
			key := restoreExportedFunctionKey(t, function)
			if key == "" {
				continue
			}
			if _, admitted := want[key]; !admitted {
				t.Fatalf("Graph restore exports unowned function %s", key)
			}
			want[key] = true
		}
	}
	for name, present := range want {
		if !present {
			t.Fatalf("Graph restore required export %s is missing", name)
		}
	}
}

func restoreExportedFunctionKey(t testing.TB, function *ast.FuncDecl) string {
	t.Helper()
	if function.Recv == nil {
		return "func " + function.Name.Name
	}
	receiver := function.Recv.List[0].Type
	if pointer, ok := receiver.(*ast.StarExpr); ok {
		receiver = pointer.X
	}
	identifier, ok := receiver.(*ast.Ident)
	if !ok {
		t.Fatalf("unsupported Graph restore receiver for %s", function.Name.Name)
	}
	if !token.IsExported(identifier.Name) {
		return ""
	}
	return "method " + identifier.Name + "." + function.Name.Name
}
