package jobs_test

import (
	"fmt"
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
	// The removed API belongs to Jobs. Inspect its declarations; compilation
	// rejects callers of absent APIs regardless of import aliases or extraction.
	// An unrelated owner's identically named method is not a Jobs capability.
	root := filepath.Dir(currentFile)
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(root, entry.Name())
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, violation := range removedJobsAPIViolations(parsed) {
			t.Errorf("%s: %s", path, violation)
		}
	}
	for _, fixture := range []struct {
		name, source string
		violations   int
	}{
		{"removed type", "type CreateParams struct{}", 1},
		{"removed alias", "type ExtensionJobContract = Replacement", 1},
		{"manager mutation", "func (m *Manager) ConfigureTelemetry() {}", 1},
		{"value receiver mutation", "func (m Manager) ValidateConfiguration() {}", 1},
		{"runner dispatch", "func (r *Runner) DispatchJobID() {}", 1},
		{"removed function", "func DispatchJob() {}", 1},
		{"removed function alias", "var DispatchJob = replacement", 1},
		{"mutable constructor alias", "var NewManager = replacement", 1},
		{"mutable constructor", "func NewManager(a, b ManagerOptions) {}", 1},
		{"variadic constructor", "func NewRunner(options ...RunnerOptions) {}", 1},
		{"wrong options", "func NewManager(options OtherOptions) {}", 1},
		{"renamed parameter", "func NewManager(renamed ManagerOptions) {}", 0},
		{"unnamed parameter", "func NewRunner(RunnerOptions) {}", 0},
		{"unrelated call", "func assemble() { reporting.ValidateConfiguration(value) }", 0},
		{"unrelated construction", "func assemble() { other.NewManager(a, b) }", 0},
		{"store creation", "func (s *Store) Create() {}", 0},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			parsed, err := parser.ParseFile(token.NewFileSet(), "fixture.go", "package jobs\n"+fixture.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			if got := removedJobsAPIViolations(parsed); len(got) != fixture.violations {
				t.Fatalf("violations = %v; want %d", got, fixture.violations)
			}
		})
	}
}

func removedJobsAPIViolations(file *ast.File) []string {
	legacyTypes := map[string]bool{
		"CreateParams": true, "TransitionParams": true,
		"ExtensionJobContract": true, "RecoverableHandlerJobs": true,
	}
	legacySurfaces := map[string]bool{
		"Configure": true, "ConfigureTelemetry": true, "ConfigureDequeueGate": true,
		"Dispatch": true, "DispatchJob": true, "DispatchJobID": true,
		"ExtensionContract": true, "MarkRunning": true, "RecoverHandler": true,
		"ReleaseHandlerLease": true, "RecordHandlerFailure": true,
		"RecordHandlerIncomplete": true, "ValidateConfiguration": true,
	}
	var violations []string
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *ast.GenDecl:
			for _, spec := range value.Specs {
				switch declared := spec.(type) {
				case *ast.TypeSpec:
					if legacyTypes[declared.Name.Name] {
						violations = append(violations, "removed Jobs type "+declared.Name.Name)
					}
				case *ast.ValueSpec:
					for _, name := range declared.Names {
						if legacySurfaces[name.Name] || name.Name == "NewManager" || name.Name == "NewRunner" {
							violations = append(violations, "removed Jobs surface "+name.Name)
						}
					}
				}
			}
		case *ast.FuncDecl:
			name := value.Name.Name
			if legacySurfaces[name] || (name == "Create" && value.Recv != nil && len(value.Recv.List) == 1 && receiverTypeName(value.Recv.List[0].Type) == "Manager") {
				violations = append(violations, "removed Jobs surface "+name)
			}
			if value.Recv == nil && (name == "NewManager" || name == "NewRunner") {
				want := strings.TrimPrefix(name, "New") + "Options"
				parameters := value.Type.Params.List
				valid := len(parameters) == 1
				if valid {
					kind, named := parameters[0].Type.(*ast.Ident)
					valid = named && kind.Name == want && len(parameters[0].Names) <= 1
				}
				if !valid {
					violations = append(violations, fmt.Sprintf("%s requires one %s value", name, want))
				}
			}
		}
	}
	return violations
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
