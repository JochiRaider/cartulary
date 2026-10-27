package indicators

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestIndicatorCompositionUsesInjectedCapabilities(t *testing.T) {
	t.Helper()

	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	internalRoot := filepath.Join(repoRoot, "internal")
	indicatorRootImport := indicatorsRepoImportPrefix + "internal/modules/indicators"
	peerConstructors := map[string]map[string]struct{}{
		indicatorsRepoImportPrefix + "internal/platform/authn": {
			"NewStore": {},
		},
		indicatorsRepoImportPrefix + "internal/modules/incidents/admission": {
			"NewChecker": {},
		},
		indicatorsRepoImportPrefix + "internal/modules/records": {
			"NewStore": {},
		},
		indicatorsRepoImportPrefix + "internal/modules/revisions": {
			"NewAppender": {},
		},
	}
	err := filepath.WalkDir(internalRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		relativePath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		relativePath = filepath.ToSlash(relativePath)
		production := !strings.HasSuffix(path, "_test.go")
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		imports := indicatorParsedImportNames(t, parsed, relativePath)
		indicatorAlias := imports[indicatorRootImport]
		authnAlias := imports[indicatorsRepoImportPrefix+"internal/platform/authn"]
		rootProduction := production && filepath.ToSlash(filepath.Dir(relativePath)) == "internal/modules/indicators"

		ast.Inspect(parsed, func(node ast.Node) bool {
			literal, ok := node.(*ast.CompositeLit)
			if ok && production && indicatorAlias != "" && authnAlias != "" {
				if selector, ok := literal.Type.(*ast.SelectorExpr); ok {
					qualifier, qualified := selector.X.(*ast.Ident)
					if qualified && qualifier.Name == authnAlias && selector.Sel.Name == "UserRecord" {
						t.Fatalf("%s constructs a synthetic authn.UserRecord while consuming Indicators", relativePath)
					}
				}
			}

			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			called, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			qualifier, ok := called.X.(*ast.Ident)
			if !ok {
				return true
			}
			if !rootProduction {
				return true
			}
			for importPath, constructors := range peerConstructors {
				if qualifier.Name != imports[importPath] {
					continue
				}
				if _, forbidden := constructors[called.Sel.Name]; forbidden {
					t.Fatalf("%s constructs concrete peer %s.%s", relativePath, qualifier.Name, called.Sel.Name)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("audit Indicator composition: %v", err)
	}
}

func indicatorParsedImportNames(t testing.TB, file *ast.File, relativePath string) map[string]string {
	t.Helper()
	result := make(map[string]string, len(file.Imports))
	for _, importSpec := range file.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import in %s: %v", relativePath, err)
		}
		name := filepath.Base(importPath)
		if importSpec.Name != nil {
			name = importSpec.Name.Name
		}
		result[importPath] = name
	}
	return result
}

func TestIndicatorBorrowedTransactionsRemainCallerOwned(t *testing.T) {
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		pgxAlias := indicatorParsedImportNames(t, parsed, path)["github.com/jackc/pgx/v5"]
		if pgxAlias == "" {
			continue
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			borrowed := map[string]bool{}
			for _, parameter := range function.Type.Params.List {
				typeName, ok := parameter.Type.(*ast.SelectorExpr)
				if !ok || typeName.Sel.Name != "Tx" {
					continue
				}
				qualifier, ok := typeName.X.(*ast.Ident)
				if !ok || qualifier.Name != pgxAlias {
					continue
				}
				for _, name := range parameter.Names {
					borrowed[name.Name] = true
				}
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				receiver, ok := selector.X.(*ast.Ident)
				if !ok || !borrowed[receiver.Name] {
					return true
				}
				switch selector.Sel.Name {
				case "Begin", "BeginTx", "Commit", "Rollback":
					t.Errorf("%s: %s takes lifecycle ownership of borrowed transaction %s", path, function.Name.Name, receiver.Name)
				}
				return true
			})
		}
	}
}
