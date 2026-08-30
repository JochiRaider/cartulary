package graphprojection

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const cartularyImportPrefix = "github.com/JochiRaider/cartulary/"

func TestGraphProjectionV2RootProductionBoundary_Unit(t *testing.T) {
	allowedPrivateImports := map[string]struct{}{
		cartularyImportPrefix + "internal/modules/graphprojection/internal/semanticlimits": {},
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Graph Projection root: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		for _, importPath := range productionImportsForFile(t, name) {
			firstSegment := strings.Split(importPath, "/")[0]
			if strings.Contains(firstSegment, ".") {
				if _, allowed := allowedPrivateImports[importPath]; !allowed {
					t.Fatalf("root production file %s imports non-standard or unowned private dependency %s", name, importPath)
				}
			}
		}
	}
}

func TestGraphProjectionPackageTopologyAllowlist_Unit(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Graph Projection root: %v", err)
	}
	wantDirectories := map[string]bool{
		"internal":        false,
		"postgresrestore": false,
		"postgresresult":  false,
		"restore":         false,
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, admitted := wantDirectories[entry.Name()]; !admitted {
			t.Fatalf("Graph Projection has unowned package directory %s", entry.Name())
		}
		wantDirectories[entry.Name()] = true
	}
	for directory, present := range wantDirectories {
		if !present {
			t.Fatalf("Graph Projection required package directory %s is missing", directory)
		}
	}
	privateEntries, err := os.ReadDir("internal")
	if err != nil {
		t.Fatalf("read Graph Projection private implementation root: %v", err)
	}
	wantPrivateDirectories := map[string]bool{"semanticlimits": false}
	for _, entry := range privateEntries {
		if !entry.IsDir() {
			t.Fatalf("Graph Projection private implementation root contains non-package file %s", entry.Name())
		}
		if _, admitted := wantPrivateDirectories[entry.Name()]; !admitted {
			t.Fatalf("Graph Projection has unowned private package directory internal/%s", entry.Name())
		}
		wantPrivateDirectories[entry.Name()] = true
	}
	for directory, present := range wantPrivateDirectories {
		if !present {
			t.Fatalf("Graph Projection required private package directory internal/%s is missing", directory)
		}
	}
}

func TestGraphProjectionRootExportedFunctionAllowlist_Unit(t *testing.T) {
	want := map[string]bool{
		"func ProjectV2":                            false,
		"method ProjectionErrorV2.Error":            false,
		"method ProjectionErrorV2.Unwrap":           false,
		"method ProjectionResultV2.CompletedResult": false,
		"method ProjectionResultV2.Resource":        false,
		"method ProjectionResultV2.ResultBindingV2": false,
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Graph Projection root: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(name), nil, 0)
		if err != nil {
			t.Fatalf("parse declarations for %s: %v", name, err)
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || !function.Name.IsExported() {
				continue
			}
			key := exportedFunctionKey(t, function)
			if key == "" {
				continue
			}
			if _, admitted := want[key]; !admitted {
				t.Fatalf("Graph Projection root exports unowned function %s", key)
			}
			want[key] = true
		}
	}
	for name, present := range want {
		if !present {
			t.Fatalf("Graph Projection root required export %s is missing", name)
		}
	}
}

func exportedFunctionKey(t testing.TB, function *ast.FuncDecl) string {
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
		t.Fatalf("unsupported exported receiver for %s", function.Name.Name)
	}
	if !token.IsExported(identifier.Name) {
		return ""
	}
	return "method " + identifier.Name + "." + function.Name.Name
}

func productionImportsForFile(t testing.TB, path string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(path), nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", path, err)
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		imports = append(imports, strings.Trim(spec.Path.Value, "\""))
	}
	return imports
}
