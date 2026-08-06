package rootedfs_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestProductionFilesystemEffectBoundary_Unit(t *testing.T) {
	repositoryRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	scanRoots := []string{
		"internal/app/recoveryassembly",
		"internal/modules/imports",
		"internal/modules/incidentbundles",
		"internal/modules/recovery",
		"internal/modules/reference_data",
		"internal/modules/reporting",
		"internal/platform/bootstrap",
		"internal/platform/objectstore",
		"internal/platform/postgres",
	}
	allowedRawEffects := map[string]map[string]bool{
		"internal/platform/postgres/migrationevidence/evidence.go": {
			"ReadFile": true,
		},
		"internal/platform/postgres/postgres.go": {
			"MkdirAll": true,
			"OpenFile": true,
		},
	}
	rawEffectCalls := map[string]bool{
		"Create":     true,
		"CreateTemp": true,
		"Lstat":      true,
		"Mkdir":      true,
		"MkdirAll":   true,
		"MkdirTemp":  true,
		"Open":       true,
		"OpenFile":   true,
		"OpenRoot":   true,
		"ReadDir":    true,
		"ReadFile":   true,
		"Remove":     true,
		"RemoveAll":  true,
		"Rename":     true,
		"Stat":       true,
		"WriteFile":  true,
	}

	violations := make([]string, 0)
	for _, relativeRoot := range scanRoots {
		absoluteRoot := filepath.Join(repositoryRoot, relativeRoot)
		err := filepath.WalkDir(absoluteRoot, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				return nil
			}
			relativePath, err := filepath.Rel(repositoryRoot, path)
			if err != nil {
				return err
			}
			relativePath = filepath.ToSlash(relativePath)
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				return err
			}
			osAliases := map[string]bool{}
			for _, spec := range file.Imports {
				if strings.Trim(spec.Path.Value, `"`) != "os" {
					continue
				}
				alias := "os"
				if spec.Name != nil {
					alias = spec.Name.Name
				}
				osAliases[alias] = true
			}
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || !rawEffectCalls[selector.Sel.Name] {
					return true
				}
				identifier, ok := selector.X.(*ast.Ident)
				if !ok || !osAliases[identifier.Name] {
					return true
				}
				if allowedRawEffects[relativePath][selector.Sel.Name] {
					return true
				}
				violations = append(violations, relativePath+": os."+selector.Sel.Name)
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("scan production filesystem effects below %s: %v", relativeRoot, err)
		}
	}
	sort.Strings(violations)
	if len(violations) != 0 {
		t.Fatalf("unapproved production filesystem effects:\n%s", strings.Join(violations, "\n"))
	}
}
