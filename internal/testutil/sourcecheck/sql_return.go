package sourcecheck

import (
	"fmt"
	"go/ast"
	"go/token"
)

// A nonempty set of observed returns is not proof of a complete helper summary.
// This context is local to one function invocation, never a nested callback.
type sqlReturnSummary struct {
	name       string
	binding    *ast.Ident
	values     []sqlValue
	complete   bool
	unresolved []string
}

func (s *sqlReturnSummary) reject(reason string) {
	s.complete = false
	s.unresolved = append(s.unresolved, "helper "+s.name+": "+reason)
}

func (s *sqlReturnSummary) result() []sqlValue {
	if !s.complete {
		return []sqlValue{{"%s", s.unresolved}}
	}
	if len(s.values) == 0 {
		return unknownSQL("helper " + s.name + ": no proven return value")
	}
	return s.values
}

func (a *sqlAnalyzer) helperValue(fn *ast.FuncDecl, args []ast.Expr, visiting map[string]bool) []sqlValue {
	summary := &sqlReturnSummary{name: fn.Name.Name, complete: true}
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 || len(fn.Type.Results.List[0].Names) > 1 || sqlExprName(fn.Type.Results.List[0].Type) != "string" {
		summary.reject("raw inference requires exactly one string result")
		return summary.result()
	}
	field := fn.Type.Results.List[0]
	if len(field.Names) == 1 {
		summary.binding = field.Names[0]
	}
	child := &sqlAnalyzer{packageConstantBindings: a.packageConstantBindings, bindings: sqlBindingInventory(fn), fset: a.fset, current: fn, observed: map[ast.Expr]bool{}, directCalls: a.directCalls, packageConstants: a.packageConstants, constantImports: a.constantImports, types: a.types, scope: a.scope, variableTypes: map[string]sqlTypeRef{}, constants: a.constants, functions: a.functions, functionImports: a.functionImports, imports: a.functionImports[fn], collections: map[string]sqlCollection{}, calls: map[string]bool{}, env: sqlEnv{}, seen: map[string]bool{}, returnSummary: summary}
	for key, val := range a.calls {
		child.calls[key] = val
	}
	child.calls[fn.Name.Name] = true
	argument := 0
	for _, field := range fn.Type.Params.List {
		for _, name := range field.Names {
			child.variableTypes[sqlBinding(name)] = child.typeRef(field.Type)
			child.env[sqlBinding(name)] = unknownSQL("missing argument " + name.Name)
			if argument < len(args) {
				child.env[sqlBinding(name)] = a.value(args[argument], visiting)
			}
			argument++
		}
	}
	if summary.binding != nil {
		key := sqlBinding(summary.binding)
		child.variableTypes[key] = child.typeRef(field.Type)
		child.env[key] = []sqlValue{{"", nil}}
		// Explicit returns assign the named destination before deferred calls.
		// Its effects matter even when the return expression is a safe literal.
		if child.bindings[key].externallyMutable() {
			summary.reject("named result has an external writer or address escape: " + summary.binding.Name)
		}
	}
	// Admission is separate from value traversal. A skipped unsupported return
	// must not be hidden by an earlier safe return. Nested bodies have their own
	// admission; their binding effects were independently inventoried above.
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if _, nested := node.(*ast.FuncLit); nested {
			return false
		}
		reason := ""
		switch statement := node.(type) {
		case *ast.DeferStmt:
			// Recovery can return the current result without executing a return
			// statement or writing the result. Do not infer deferred effects.
			summary.reject("deferred control flow is unsupported for raw helper inference at " + a.fset.Position(statement.Pos()).String())
			return false
		case *ast.SelectStmt:
			reason = "select"
		case *ast.LabeledStmt:
			reason = "label"
		case *ast.BranchStmt:
			if statement.Label != nil || (statement.Tok != token.BREAK && statement.Tok != token.CONTINUE) {
				reason = statement.Tok.String()
			}
		case *ast.BlockStmt, *ast.AssignStmt, *ast.DeclStmt, *ast.ExprStmt, *ast.ReturnStmt,
			*ast.IfStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.CaseClause,
			*ast.ForStmt, *ast.RangeStmt, *ast.IncDecStmt, *ast.EmptyStmt,
			*ast.GoStmt, *ast.SendStmt:
		default:
			if _, statement := node.(ast.Stmt); statement {
				reason = fmt.Sprintf("%T", node)
			}
		}
		if reason != "" {
			summary.reject("unsupported control flow " + reason + " at " + a.fset.Position(node.Pos()).String())
			return false
		}
		return true
	})
	if !summary.complete {
		return summary.result()
	}
	if !child.block(fn.Body.List, false) {
		summary.reject("function exit is not proven by bounded control flow")
	}
	return summary.result()
}

func (a *sqlAnalyzer) recordReturn(statement *ast.ReturnStmt) {
	summary := a.returnSummary
	if summary == nil {
		return
	}
	var values []sqlValue
	switch len(statement.Results) {
	case 0:
		if summary.binding == nil {
			summary.reject("bare return has no named result")
			return
		}
		values = a.value(summary.binding, map[string]bool{})
	case 1:
		values = a.value(statement.Results[0], map[string]bool{})
	default:
		summary.reject("raw tuple return is unsupported")
		return
	}
	summary.values = mergeSQLValues(summary.values, values)
}
