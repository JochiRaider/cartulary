package jobs_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRemovedJobsSurfacesRemainAbsent_Architecture(t *testing.T) {
	t.Parallel()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture guard location")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	legacyJobsTypes := map[string]struct{}{
		"CreateParams":           {},
		"TransitionParams":       {},
		"ExtensionJobContract":   {},
		"RecoverableHandlerJobs": {},
	}
	legacySelectors := map[string]struct{}{
		"Configure":               {},
		"ConfigureDequeueGate":    {},
		"Dispatch":                {},
		"DispatchJob":             {},
		"DispatchJobID":           {},
		"ExtensionContract":       {},
		"MarkRunning":             {},
		"RecoverHandler":          {},
		"ReleaseHandlerLease":     {},
		"RecordHandlerFailure":    {},
		"RecordHandlerIncomplete": {},
		"ValidateConfiguration":   {},
	}
	legacyManagerMethods := map[string]struct{}{
		"Configure":             {},
		"ConfigureTelemetry":    {},
		"Create":                {},
		"ExtensionContract":     {},
		"MarkRunning":           {},
		"ValidateConfiguration": {},
	}

	err := filepath.WalkDir(filepath.Join(repoRoot, "internal"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		parsed, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		inJobsPackage := filepath.Dir(path) == filepath.Join(repoRoot, "internal", "platform", "jobs")
		ast.Inspect(parsed, func(node ast.Node) bool {
			switch value := node.(type) {
			case *ast.TypeSpec:
				if inJobsPackage {
					if _, forbidden := legacyJobsTypes[value.Name.Name]; forbidden {
						t.Errorf("removed Jobs type %s returned in %s", value.Name.Name, path)
					}
				}
			case *ast.FuncDecl:
				if !inJobsPackage || value.Recv == nil || len(value.Recv.List) != 1 {
					break
				}
				receiverName := receiverTypeName(value.Recv.List[0].Type)
				if receiverName == "Manager" {
					if _, forbidden := legacyManagerMethods[value.Name.Name]; forbidden {
						t.Errorf("removed Manager method %s returned in %s", value.Name.Name, path)
					}
				}
			case *ast.CallExpr:
				switch called := value.Fun.(type) {
				case *ast.Ident:
					if (called.Name == "NewManager" || called.Name == "NewRunner") && len(value.Args) != 1 {
						t.Errorf("%s must have exactly one immutable options argument in %s", called.Name, path)
					}
				case *ast.SelectorExpr:
					if _, forbidden := legacySelectors[called.Sel.Name]; forbidden {
						t.Errorf("removed Jobs call surface %s returned in %s", called.Sel.Name, path)
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func receiverTypeName(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return receiverTypeName(value.X)
	default:
		return ""
	}
}
