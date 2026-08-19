package codegenboundary_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	repoImportPrefix = "github.com/JochiRaider/cartulary/"
	sqlcImportPath   = repoImportPrefix + "internal/gen/sql"
)

func TestSQLCImportsStayBehindPersistenceAdapters(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	allowed := map[string]bool{
		"internal/modules/incidents/import_finalization.go":        true,
		"internal/modules/incidents/open_guard.go":                 true,
		"internal/modules/incidents/store.go":                      true,
		"internal/modules/workbook/startup/postgres/repository.go": true,
		"internal/modules/workbook/startup/postgres/writer.go":     true,
		"internal/modules/recovery/store.go":                       true,
		"internal/modules/reporting/store.go":                      true,
		"internal/modules/savedviews/store.go":                     true,
		"internal/modules/timeline/query_projection_store.go":      true,
		"internal/modules/timeline/store.go":                       true,
		"internal/modules/timeline/workbookprojection/store.go":    true,
	}

	var offenders []string
	err := filepath.WalkDir(repoRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch filepath.ToSlash(path) {
			case filepath.ToSlash(filepath.Join(repoRoot, ".git")),
				filepath.ToSlash(filepath.Join(repoRoot, ".cartulary")),
				filepath.ToSlash(filepath.Join(repoRoot, "internal", "gen")),
				filepath.ToSlash(filepath.Join(repoRoot, "tmp")):
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)
		sqlcAliases := sqlcImportAliases(t, path)
		if len(sqlcAliases) == 0 {
			return nil
		}
		if !allowed[relPath] {
			offenders = append(offenders, relPath)
			return nil
		}
		if exportedDeclExposesSQLC(t, path, sqlcAliases) {
			offenders = append(offenders, relPath+" exposes sqlc types through an exported declaration")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan repository Go files: %v", err)
	}
	if len(offenders) > 0 {
		sort.Strings(offenders)
		t.Fatalf("SQLC imports must stay behind owner persistence adapters:\n%s", strings.Join(offenders, "\n"))
	}
}

func sqlcImportAliases(t *testing.T, filePath string) map[string]bool {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", filePath, err)
	}
	aliases := map[string]bool{}
	for _, importSpec := range parsed.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import in %s: %v", filePath, err)
		}
		if importPath != sqlcImportPath {
			continue
		}
		alias := "sqlc"
		if importSpec.Name != nil {
			alias = importSpec.Name.Name
		}
		aliases[alias] = true
	}
	return aliases
}

func exportedDeclExposesSQLC(t *testing.T, filePath string, aliases map[string]bool) bool {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse file for %s: %v", filePath, err)
	}
	for _, decl := range parsed.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.IsExported() && nodeReferencesSQLCAlias(d.Type, aliases) {
				return true
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.IsExported() && nodeReferencesSQLCAlias(s.Type, aliases) {
						return true
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.IsExported() && nodeReferencesSQLCAlias(s.Type, aliases) {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func nodeReferencesSQLCAlias(node ast.Node, aliases map[string]bool) bool {
	if node == nil {
		return false
	}
	found := false
	ast.Inspect(node, func(child ast.Node) bool {
		if found || child == nil {
			return false
		}
		selector, ok := child.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if ok && aliases[ident.Name] {
			found = true
			return false
		}
		return true
	})
	return found
}
