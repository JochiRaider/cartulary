package operator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestOperatorExportedSurfaceIsCompositionOnly(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve operator package path")
	}
	operatorDir := filepath.Dir(currentFile)
	assertOperatorExportedSurface(t, operatorDir, []string{"RunOperatorCLIContext"})
	assertOperatorExportedSurface(
		t,
		filepath.Join(operatorDir, "internal", "recoverycli"),
		[]string{"FailureEvidenceFields", "Run"},
	)

	t.Run("negative fixture detects an unapproved export", func(t *testing.T) {
		file, err := parser.ParseFile(
			token.NewFileSet(),
			"negative_fixture.go",
			"package negativefixture\nfunc UnexpectedExport() {}\n",
			0,
		)
		if err != nil {
			t.Fatalf("parse negative fixture: %v", err)
		}
		unexpected, _ := operatorSurfaceDelta(exportedOperatorDeclarations(file), nil)
		if !reflect.DeepEqual(unexpected, []string{"UnexpectedExport"}) {
			t.Fatalf("negative fixture was not rejected: unexpected=%#v", unexpected)
		}
	})
}

func assertOperatorExportedSurface(t testing.TB, directory string, allowed []string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read package directory %s: %v", directory, err)
	}
	fileSet := token.NewFileSet()
	var exported []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		filePath := filepath.Join(directory, entry.Name())
		file, parseErr := parser.ParseFile(fileSet, filePath, nil, 0)
		if parseErr != nil {
			t.Fatalf("parse production Go file %s: %v", filePath, parseErr)
		}
		exported = append(exported, exportedOperatorDeclarations(file)...)
	}
	unapproved, missing := operatorSurfaceDelta(exported, allowed)
	if len(unapproved) != 0 || len(missing) != 0 {
		t.Fatalf("exported surface mismatch for %s: unapproved=%v missing=%v", directory, unapproved, missing)
	}
}

func exportedOperatorDeclarations(file *ast.File) []string {
	var names []string
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			if ast.IsExported(declaration.Name.Name) {
				names = append(names, declaration.Name.Name)
			}
		case *ast.GenDecl:
			for _, spec := range declaration.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					if ast.IsExported(spec.Name.Name) {
						names = append(names, spec.Name.Name)
					}
				case *ast.ValueSpec:
					for _, name := range spec.Names {
						if ast.IsExported(name.Name) {
							names = append(names, name.Name)
						}
					}
				}
			}
		}
	}
	return names
}

func operatorSurfaceDelta(exported []string, allowed []string) ([]string, []string) {
	exportedSet := make(map[string]struct{}, len(exported))
	for _, name := range exported {
		exportedSet[name] = struct{}{}
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = struct{}{}
	}
	var unapproved []string
	for name := range exportedSet {
		if _, ok := allowedSet[name]; !ok {
			unapproved = append(unapproved, name)
		}
	}
	var missing []string
	for name := range allowedSet {
		if _, ok := exportedSet[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(unapproved)
	sort.Strings(missing)
	return unapproved, missing
}
