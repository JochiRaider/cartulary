package rootedfs_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// osFunctionEffects classifies every exported os function: true for a
// filesystem effect that deployment-root consumers must reach only through
// rootedfs, false for an explicit non-effect. Classification is a policy
// decision; TestOSFunctionEffectClassificationIsComplete_Unit makes it
// complete, so a Go release that adds an os function forces a decision.
var osFunctionEffects = map[string]bool{
	"Chdir":      true,
	"Chmod":      true,
	"Chown":      true,
	"Chtimes":    true,
	"CopyFS":     true,
	"Create":     true,
	"CreateTemp": true,
	"DirFS":      true,
	"Lchown":     true,
	"Link":       true,
	"Lstat":      true,
	"Mkdir":      true,
	"MkdirAll":   true,
	"MkdirTemp":  true,
	"Open":       true,
	"OpenFile":   true,
	"OpenInRoot": true,
	"OpenRoot":   true,
	"ReadDir":    true,
	"ReadFile":   true,
	"Readlink":   true,
	"Remove":     true,
	"RemoveAll":  true,
	"Rename":     true,
	"Stat":       true,
	"Symlink":    true,
	"Truncate":   true,
	"WriteFile":  true,

	// Process, environment, identity, error and path computations touch no
	// deployment root.
	"Clearenv":        false,
	"Environ":         false,
	"Executable":      false,
	"Exit":            false,
	"Expand":          false,
	"ExpandEnv":       false,
	"FindProcess":     false,
	"Getegid":         false,
	"Getenv":          false,
	"Geteuid":         false,
	"Getgid":          false,
	"Getgroups":       false,
	"Getpagesize":     false,
	"Getpid":          false,
	"Getppid":         false,
	"Getuid":          false,
	"Getwd":           false,
	"Hostname":        false,
	"IsExist":         false,
	"IsNotExist":      false,
	"IsPathSeparator": false,
	"IsPermission":    false,
	"IsTimeout":       false,
	"LookupEnv":       false,
	"NewFile":         false,
	"NewSyscallError": false,
	"Pipe":            false,
	"SameFile":        false,
	"Setenv":          false,
	"StartProcess":    false,
	"TempDir":         false,
	"Unsetenv":        false,
	"UserCacheDir":    false,
	"UserConfigDir":   false,
	"UserHomeDir":     false,
}

func TestOSFunctionEffectClassificationIsComplete_Unit(t *testing.T) {
	pkg, err := importer.ForCompiler(token.NewFileSet(), "source", nil).Import("os")
	if err != nil {
		t.Fatalf("load os package: %v", err)
	}
	exported := map[string]bool{}
	for _, name := range pkg.Scope().Names() {
		if _, isFunc := pkg.Scope().Lookup(name).(*types.Func); isFunc && token.IsExported(name) {
			exported[name] = true
		}
	}
	var unclassified, stale []string
	for name := range exported {
		if _, classified := osFunctionEffects[name]; !classified {
			unclassified = append(unclassified, name)
		}
	}
	for name := range osFunctionEffects {
		if !exported[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(unclassified)
	sort.Strings(stale)
	if len(unclassified) != 0 || len(stale) != 0 {
		t.Fatalf("os function classification is incomplete: unclassified %v, no longer exported %v", unclassified, stale)
	}
}

func TestProductionFilesystemEffectBoundary_Unit(t *testing.T) {
	repositoryRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	scanRoots := []string{
		"internal/app/recoveryassembly",
		"internal/modules/database_migrations",
		"internal/modules/imports",
		"internal/modules/incidentbundles",
		"internal/modules/recovery",
		"internal/modules/reference_data",
		"internal/modules/reporting",
		"internal/platform/bootstrap",
		"internal/platform/objectstore",
		"internal/platform/postgres",
	}
	// Reference Data test support is fixture data, not a deployment root
	// consumer; runtime packages must never import it.
	nonRuntimeRoots := []string{
		"internal/modules/reference_data/testsupport/",
	}
	allowedRawEffects := map[string]map[string]bool{
		// Test-fixture evidence: it creates, secures and reads its own files.
		"internal/modules/recovery/testsupport/recovery.go": {
			"Chmod":     true,
			"MkdirAll":  true,
			"ReadFile":  true,
			"WriteFile": true,
		},
		"internal/modules/database_migrations/migrationevidence/evidence.go": {
			"ReadFile": true,
		},
		"internal/modules/database_migrations/migrations.go": {
			"MkdirAll": true,
			"OpenFile": true,
		},
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
			for _, prefix := range nonRuntimeRoots {
				if strings.HasPrefix(relativePath, prefix) {
					return nil
				}
			}
			for _, spec := range file.Imports {
				importPath := strings.Trim(spec.Path.Value, `"`)
				for _, prefix := range nonRuntimeRoots {
					boundary := "github.com/JochiRaider/cartulary/" + strings.TrimSuffix(prefix, "/")
					if importPath == boundary || strings.HasPrefix(importPath, boundary+"/") {
						violations = append(violations, relativePath+": runtime import of "+importPath)
					}
				}
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
			// Every os.<effect> selector counts, whether called or taken as a
			// function value.
			ast.Inspect(file, func(node ast.Node) bool {
				selector, ok := node.(*ast.SelectorExpr)
				if !ok || !osFunctionEffects[selector.Sel.Name] {
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
