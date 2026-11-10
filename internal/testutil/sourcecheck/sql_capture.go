package sourcecheck

import (
	"go/ast"
	"go/token"
)

// Binding facts are collected before bounded evaluation. They describe lexical
// identity and all writes/escapes, never callback scheduling or heap contents.
type sqlBindingFacts struct {
	owner             ast.Node
	initialized       bool
	constant          bool
	writes            int
	nestedWriter      bool
	addressEscaped    bool
	aggregateMutated  bool
	collectionEscaped bool
	jumpMutable       bool
}

func (f *sqlBindingFacts) stableCapture() bool {
	return f != nil && (f.constant || f.initialized && f.writes == 0 && !f.nestedWriter && !f.addressEscaped && !f.aggregateMutated && !f.jumpMutable)
}
func (f *sqlBindingFacts) externallyMutable() bool {
	return f != nil && (f.nestedWriter || f.addressEscaped || f.aggregateMutated || f.jumpMutable)
}

func sqlBindingInventory(root ast.Node) map[string]*sqlBindingFacts {
	facts := map[string]*sqlBindingFacts{}
	add := func(id *ast.Ident, owner ast.Node, initialized bool) {
		key := sqlBinding(id)
		if facts[key] == nil {
			facts[key] = &sqlBindingFacts{owner: owner, initialized: initialized, constant: id.Obj != nil && id.Obj.Kind == ast.Con}
		}
	}
	fields := func(list *ast.FieldList, owner ast.Node) {
		if list != nil {
			for _, f := range list.List {
				for _, id := range f.Names {
					add(id, owner, false)
				}
			}
		}
	}
	var walk func(ast.Node, ast.Node)
	write := func(expr ast.Expr, owner ast.Node) {
		if id := sqlRootBinding(expr); id != nil {
			if f := facts[sqlBinding(id)]; f != nil {
				f.writes++
				f.nestedWriter = f.nestedWriter || f.owner != owner
				_, simple := unparenSQL(expr).(*ast.Ident)
				f.aggregateMutated = f.aggregateMutated || !simple
			}
		}
	}
	walk = func(node ast.Node, owner ast.Node) {
		ast.Inspect(node, func(n ast.Node) bool {
			switch e := n.(type) {
			case *ast.FuncLit:
				fields(e.Type.Params, e)
				fields(e.Type.Results, e)
				walk(e.Body, e)
				return false
			case *ast.ValueSpec:
				for _, id := range e.Names {
					add(id, owner, len(e.Values) > 0)
				}
			case *ast.AssignStmt:
				for _, left := range e.Lhs {
					if id, ok := left.(*ast.Ident); ok && e.Tok == token.DEFINE && id.Obj != nil && id.Obj.Decl == e {
						add(id, owner, true)
					} else {
						write(left, owner)
					}
				}
			case *ast.RangeStmt:
				for _, left := range []ast.Expr{e.Key, e.Value} {
					if left == nil {
						continue
					}
					if id, ok := left.(*ast.Ident); ok && e.Tok == token.DEFINE {
						add(id, owner, false)
					} else {
						write(left, owner)
					}
				}
			case *ast.IncDecStmt:
				write(e.X, owner)
			case *ast.UnaryExpr:
				if e.Op == token.AND {
					if id := sqlRootBinding(e.X); id != nil {
						if f := facts[sqlBinding(id)]; f != nil {
							f.addressEscaped = true
						}
					}
				}
			}
			return true
		})
	}
	if fn, ok := root.(*ast.FuncDecl); ok {
		fields(fn.Recv, fn)
		fields(fn.Type.Params, fn)
		fields(fn.Type.Results, fn)
		if fn.Body != nil {
			walk(fn.Body, fn)
		}
	} else {
		walk(root, nil)
	}
	// Literal collection inference is local and read-only. Any use beyond a
	// whole-binding assignment or direct range may create an alias; do not
	// guess whether that alias or a recipient mutates its backing storage.
	admitted := map[*ast.Ident]bool{}
	admit := func(expr ast.Expr) {
		if id, ok := expr.(*ast.Ident); ok {
			admitted[id] = true
		}
	}
	ast.Inspect(root, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.AssignStmt:
			for _, expr := range n.Lhs {
				admit(expr)
			}
		case *ast.ValueSpec:
			for _, id := range n.Names {
				admitted[id] = true
			}
		case *ast.Field:
			for _, id := range n.Names {
				admitted[id] = true
			}
		case *ast.RangeStmt:
			admit(n.Key)
			admit(n.Value)
			admit(n.X)
		}
		return true
	})
	ast.Inspect(root, func(node ast.Node) bool {
		if id, ok := node.(*ast.Ident); ok && !admitted[id] {
			if f := facts[sqlBinding(id)]; f != nil {
				f.collectionEscaped = true
			}
		}
		return true
	})

	// A goto may revisit an earlier boundary. Do not infer values for bindings
	// written in such a declaration; labeled blocks themselves remain bounded
	// unsupported-control discovery, without a jump-timing guess.
	hasJump := false
	ast.Inspect(root, func(node ast.Node) bool {
		if branch, ok := node.(*ast.BranchStmt); ok && branch.Tok == token.GOTO {
			hasJump = true
		}
		return true
	})
	if hasJump {
		for _, fact := range facts {
			fact.jumpMutable = fact.writes > 0
		}
	}

	return facts
}

func copySQLTypes(input map[string]sqlTypeRef) map[string]sqlTypeRef {
	out := map[string]sqlTypeRef{}
	for k, v := range input {
		out[k] = v
	}
	return out
}
func preservedSQLValue(values []sqlValue) bool {
	for _, v := range values {
		if v.text != "\x00database" {
			return false
		}
	}
	return len(values) > 0
}

// Captures retain a value only with an independent stability proof. Prepared
// and batch provenance is local, even for an otherwise immutable capture.
func (a *sqlAnalyzer) observeClosure(fn *ast.FuncLit) {
	saved, savedReturns := a.state(), a.returnSummary
	savedBreaks, savedContinues := a.breaks, a.continues
	a.breaks, a.continues = nil, nil
	a.setState(a.state())
	a.collections = map[string]sqlCollection{}
	a.returnSummary = nil
	defer func() {
		a.setState(saved)
		a.returnSummary = savedReturns
		a.breaks, a.continues = savedBreaks, savedContinues
	}()
	for key, values := range a.env {
		if !a.bindings[key].stableCapture() && !preservedSQLValue(values) {
			a.env[key] = unknownSQL("captured SQL binding is not proven immutable")
		}
	}
	for key, ref := range a.variableTypes {
		ref.qualified = false
		a.variableTypes[key] = ref
	}
	for _, fields := range []*ast.FieldList{fn.Type.Params, fn.Type.Results} {
		if fields == nil {
			continue
		}
		for _, field := range fields.List {
			for _, name := range field.Names {
				a.variableTypes[sqlBinding(name)] = a.typeRef(field.Type)
				a.env[sqlBinding(name)] = unknownSQL("callback parameter " + name.Name)
				if a.databaseType(field.Type) {
					a.env[sqlBinding(name)] = []sqlValue{{"\x00database", nil}}
				}
			}
		}
	}
	a.block(fn.Body.List, true)
}

func sqlRootBinding(expr ast.Expr) *ast.Ident {
	switch e := unparenSQL(expr).(type) {
	case *ast.Ident:
		return e
	case *ast.SelectorExpr:
		return sqlRootBinding(e.X)
	case *ast.IndexExpr:
		return sqlRootBinding(e.X)
	case *ast.StarExpr:
		return sqlRootBinding(e.X)
	}
	return nil
}
