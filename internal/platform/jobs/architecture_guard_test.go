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

// lockTakingJobsCalls validate their execution under the Jobs transition lock
// in a transaction of their own (Core 01 REQ-01-453). Called while the
// attempt's own owner transaction is open, which holds that lock until commit,
// they would wait on themselves until their context ends.
var lockTakingJobsCalls = map[string]bool{
	"UpdateProgress": true, "CompleteSucceeded": true, "CompleteFailed": true,
	"CompleteCanceled": true, "ReleaseExecution": true, "RecordExecutionFailure": true,
}

func TestNoLockTakingJobsCallInsideOwnerTransaction_Architecture(t *testing.T) {
	t.Parallel()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture guard location")
	}
	modules := filepath.Join(filepath.Dir(currentFile), "..", "..", "modules")
	err := filepath.WalkDir(modules, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		fileSet := token.NewFileSet()
		parsed, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			return err
		}
		for _, violation := range ownerTransactionJobsCallViolations(fileSet, parsed) {
			t.Errorf("%s: %s", path, violation)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct {
		name, source string
		violations   int
	}{
		{"progress with a transaction parameter", "func apply(ctx context.Context, tx pgx.Tx) { manager.UpdateProgress(ctx, execution, progress, nil) }", 1},
		{"completion inside a begun transaction", "func apply(ctx context.Context) { tx, _ := pool.Begin(ctx); jobs.CompleteSucceeded(ctx, execution, done); tx.Commit(ctx) }", 1},
		{"closure inside a begun transaction", "func apply(ctx context.Context) { tx, _ := pool.BeginTx(ctx, options); run(func() { manager.RecordExecutionFailure(ctx, execution, false) }); _ = tx.Commit(ctx) }", 1},
		{"progress after commit", "func apply(ctx context.Context) { tx, _ := pool.Begin(ctx); _ = tx.Commit(ctx); manager.UpdateProgress(ctx, execution, progress, nil) }", 0},
		{"transactional variant", "func apply(ctx context.Context, tx pgx.Tx) { transactions.UpdateProgressTx(ctx, tx, execution, progress) }", 0},
		{"no transaction", "func apply(ctx context.Context) { manager.CompleteCanceled(ctx, execution, canceled) }", 0},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			fileSet := token.NewFileSet()
			parsed, err := parser.ParseFile(fileSet, "fixture.go", "package owner\n"+fixture.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			if got := ownerTransactionJobsCallViolations(fileSet, parsed); len(got) != fixture.violations {
				t.Fatalf("violations = %v; want %d", got, fixture.violations)
			}
		})
	}
}

// ownerTransactionJobsCallViolations reports lock-taking Jobs calls lexically
// inside an open transaction: anywhere in a function with a pgx.Tx parameter,
// or after a Begin/BeginTx assignment and before that variable's Commit.
func ownerTransactionJobsCallViolations(fileSet *token.FileSet, file *ast.File) []string {
	type region struct{ start, end token.Pos }
	var violations []string
	inspectBody := func(params *ast.FieldList, body *ast.BlockStmt) {
		if body == nil {
			return
		}
		var open []region
		if params != nil {
			for _, field := range params.List {
				if selector, ok := field.Type.(*ast.SelectorExpr); ok && selector.Sel.Name == "Tx" {
					if pkg, ok := selector.X.(*ast.Ident); ok && pkg.Name == "pgx" {
						open = append(open, region{body.Pos(), body.End()})
					}
				}
			}
		}
		ast.Inspect(body, func(node ast.Node) bool {
			assign, ok := node.(*ast.AssignStmt)
			if !ok || len(assign.Rhs) != 1 || len(assign.Lhs) == 0 {
				return true
			}
			call, ok := assign.Rhs[0].(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || (selector.Sel.Name != "Begin" && selector.Sel.Name != "BeginTx") {
				return true
			}
			name, ok := assign.Lhs[0].(*ast.Ident)
			if !ok {
				return true
			}
			end := body.End()
			ast.Inspect(body, func(inner ast.Node) bool {
				commit, ok := inner.(*ast.CallExpr)
				if !ok || commit.Pos() < assign.End() {
					return true
				}
				if sel, ok := commit.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Commit" {
					if receiver, ok := sel.X.(*ast.Ident); ok && receiver.Name == name.Name && commit.Pos() < end {
						end = commit.Pos()
					}
				}
				return true
			})
			open = append(open, region{assign.End(), end})
			return true
		})
		if len(open) == 0 {
			return
		}
		ast.Inspect(body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !lockTakingJobsCalls[selector.Sel.Name] {
				return true
			}
			for _, within := range open {
				if call.Pos() >= within.start && call.Pos() < within.end {
					violations = append(violations, fmt.Sprintf("%s: %s while an owner transaction is open", fileSet.Position(call.Pos()), selector.Sel.Name))
					break
				}
			}
			return true
		})
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.FuncDecl:
			inspectBody(value.Type.Params, value.Body)
		case *ast.FuncLit:
			inspectBody(value.Type.Params, value.Body)
		}
		return true
	})
	return violations
}
