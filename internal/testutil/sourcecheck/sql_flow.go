package sourcecheck

import (
	"go/ast"
	"go/token"
)

// Writes are inventoried before evaluating repeated or unsupported control flow.
// This is deliberately conservative: no iteration-count or callback-time proof.
func sqlAssignedBindings(nodes ...ast.Node) map[string]bool {
	writes := map[string]bool{}
	add := func(expr ast.Expr) {
		if id := sqlRootBinding(expr); id != nil && id.Name != "_" {
			writes[sqlBinding(id)] = true
		}
	}
	for _, node := range nodes {
		if node == nil {
			continue
		}
		ast.Inspect(node, func(n ast.Node) bool {
			switch s := n.(type) {
			case *ast.AssignStmt:
				for _, left := range s.Lhs {
					add(left)
				}
			case *ast.IncDecStmt:
				add(s.X)
			case *ast.ValueSpec:
				for _, id := range s.Names {
					add(id)
				}
			case *ast.RangeStmt:
				if s.Tok == token.ASSIGN {
					add(s.Key)
					add(s.Value)
				}
			}
			return true
		})
	}
	return writes
}

func (a *sqlAnalyzer) invalidate(bindings map[string]bool, reason string) {
	for binding := range bindings {
		for key := range a.env {
			if sqlBindingMember(key, binding) {
				a.env[key] = unknownSQL(reason)
			}
		}
		a.env[binding] = unknownSQL(reason)
		a.collections[binding] = sqlCollection{}
		if ref, exists := a.variableTypes[binding]; exists {
			ref.qualified = false
			a.variableTypes[binding] = ref
		}
	}
}

func (a *sqlAnalyzer) finishLoop(entry sqlState, writes map[string]bool) {
	a.setState(entry)
	outer := map[string]bool{}
	for key := range writes {
		_, value := entry.values[key]
		_, typ := entry.types[key]
		_, collection := entry.collections[key]
		if value || typ || collection {
			outer[key] = true
		}
	}
	a.invalidate(outer, "loop-carried SQL is unresolved after repeated or skipped execution")
}

func (a *sqlAnalyzer) forLoop(statement *ast.ForStmt) {
	entry := a.state()
	writes := sqlAssignedBindings(statement.Body, statement.Post, statement.Cond)
	a.invalidate(writes, "loop-carried SQL is unresolved at repeated execution")
	savedBreaks, savedContinues := a.breaks, a.continues
	a.breaks, a.continues = nil, nil
	defer func() { a.breaks, a.continues = savedBreaks, savedContinues }()
	a.observe(statement.Cond)
	stopped := a.block(statement.Body.List, true)
	postPaths := a.continues
	if !stopped {
		postPaths = append(postPaths, a.state())
	}
	if statement.Post != nil && len(postPaths) > 0 {
		a.setState(a.joinStates(postPaths...))
		a.block([]ast.Stmt{statement.Post}, false)
	}
	a.finishLoop(entry, writes)
}

func (a *sqlAnalyzer) rangeLoop(statement *ast.RangeStmt) {
	a.observe(statement.X)
	collection := a.collectionValue(statement.X)
	elementType := a.types.element(a.expressionType(statement.X), map[string]bool{})
	entry := a.state()
	writes := sqlAssignedBindings(statement)
	a.invalidate(writes, "loop-carried SQL is unresolved at repeated execution")
	for _, expr := range []ast.Expr{statement.Key, statement.Value} {
		if id, ok := expr.(*ast.Ident); ok && id.Name != "_" {
			a.invalidate(map[string]bool{sqlBinding(id): true}, "unresolved range binding "+id.Name)
		}
	}
	if id, ok := statement.Value.(*ast.Ident); ok {
		if elementType.expr != nil {
			a.variableTypes[sqlBinding(id)] = elementType
		}
		if collection.known {
			for suffix, values := range collection.members {
				a.env[sqlBinding(id)+suffix] = values
			}
		}
	}
	savedBreaks, savedContinues := a.breaks, a.continues
	a.breaks, a.continues = nil, nil
	a.block(statement.Body.List, true)
	a.breaks, a.continues = savedBreaks, savedContinues
	a.finishLoop(entry, writes)
}

func hasSQLFallthrough(node ast.Node) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if branch, ok := n.(*ast.BranchStmt); ok && branch.Tok == token.FALLTHROUGH {
			found = true
		}
		return true
	})
	return found
}
