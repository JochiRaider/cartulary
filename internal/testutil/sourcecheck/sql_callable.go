package sourcecheck

import (
	"fmt"
	"go/ast"
	"go/token"
)

// callable follows the underlying signature, including defined function types.
// It is shared by invocation and value-transfer admission. A named type that is
// not callable stays data; missing/cyclic/conflicting identity cannot prove that.
func (types sqlTypes) callable(ref sqlTypeRef, seen map[string]bool) (nativeSQLMethod, bool) {
	ref = types.resolveAliases(ref, cloneSQLSeen(seen))
	switch e := ref.expr.(type) {
	case nil:
		return nativeSQLMethod{}, false
	case *ast.BadExpr:
		return nativeSQLMethod{kind: "reject"}, true
	case *ast.FuncType:
		if types.ambiguousGenericSQLSignature(ref, e) {
			return nativeSQLMethod{kind: "reject"}, true
		}
		return types.portBoundary(ref, "")
	case *ast.UnaryExpr:
		if e.Op == token.TILDE {
			return types.callable(ref.with(e.X), seen)
		}
	case *ast.BinaryExpr:
		if e.Op == token.OR {
			var candidates []nativeSQLMethod
			for _, term := range []ast.Expr{e.X, e.Y} {
				if b, ok := types.callable(ref.with(term), cloneSQLSeen(seen)); ok {
					candidates = append(candidates, b)
				}
			}
			return mergeSQLBoundaries(candidates)
		}
	case *ast.InterfaceType:
		var candidates []nativeSQLMethod
		for _, field := range e.Methods.List {
			if len(field.Names) == 0 {
				if b, ok := types.callable(ref.with(field.Type), cloneSQLSeen(seen)); ok {
					candidates = append(candidates, b)
				}
			}
		}
		return mergeSQLBoundaries(candidates)
	case *ast.ParenExpr:
		return types.callable(ref.with(e.X), seen)
	case *ast.StarExpr:
		return types.callable(ref.with(e.X), seen)
	case *ast.IndexExpr:
		return types.callable(ref.with(e.X), seen)
	case *ast.IndexListExpr:
		return types.callable(ref.with(e.X), seen)
	}
	if parameter := types.typeParameter(ref); parameter != nil {
		key := fmt.Sprintf("callable-parameter:%d", parameter.Pos())
		if seen[key] {
			return nativeSQLMethod{kind: "reject"}, true
		}
		seen[key] = true
		return types.callable(ref.with(parameter.Type), seen)
	}
	if local := types.localType(ref); local != nil {
		key := fmt.Sprintf("callable-local:%d", local.Pos())
		if seen[key] {
			return nativeSQLMethod{kind: "reject"}, true
		}
		seen[key] = true
		return types.callable(ref.with(local.Type), seen)
	}
	key := ref.key()
	if key == "" {
		return nativeSQLMethod{}, false
	}
	if seen[key] {
		return nativeSQLMethod{kind: "reject"}, true
	}
	seen[key] = true
	if definition := types[key]; definition != nil {
		var candidates []nativeSQLMethod
		nonCallable := false
		for _, alternative := range definition.definitions {
			if candidate, known := types.callable(alternative, cloneSQLSeen(seen)); known {
				candidates = append(candidates, candidate)
			} else {
				nonCallable = true
			}
		}
		if nonCallable && len(candidates) > 0 {
			return nativeSQLMethod{kind: "reject"}, true
		}
		return mergeSQLBoundaries(candidates)
	}
	if _, known := types.typeIdentity(ref, map[string]bool{}); !known {
		return nativeSQLMethod{kind: "reject"}, true
	}
	return nativeSQLMethod{}, false
}

func unparenSQL(expr ast.Expr) ast.Expr {
	for {
		p, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = p.X
	}
}

// callableBoundary returns the same identity for direct calls and extracted
// values. SQL function values have no transport authority. Only native/port
// methods may be qualified as direct execution boundaries.
func (a *sqlAnalyzer) callableBoundary(expr ast.Expr) (nativeSQLMethod, bool) {
	expr = unparenSQL(expr)
	if a.directCalls[expr] && a.declaredFunction(expr) {
		// Repository function bodies are inspected under their own declaration
		// authority. Calling one directly is not extracting a function value.
		return nativeSQLMethod{kind: "none"}, true
	}
	if selector, ok := expr.(*ast.SelectorExpr); ok {
		receiver := a.expressionType(selector.X)
		if a.types.field(receiver, selector.Sel.Name, map[string]bool{}).expr == nil {
			if boundary, known := a.types.boundary(receiver, selector.Sel.Name, map[string]bool{}); known {
				if boundary.kind != "none" && a.typeExpression(selector.X) {
					boundary.kind = "reject"
				}
				return boundary, true
			}
		}
	}
	if candidate, known := a.types.callable(a.expressionType(expr), map[string]bool{}); known {
		if candidate.kind != "none" {
			candidate.kind = "reject"
		}
		return candidate, true
	}
	if selector, ok := expr.(*ast.SelectorExpr); ok {
		if a.rejectedNativeValue(selector) {
			return nativeSQLMethod{kind: "reject"}, true
		}
		if field := a.types.field(a.expressionType(selector.X), selector.Sel.Name, map[string]bool{}); field.expr != nil {
			// callable fields were classified above; ordinary fields are data even if
			// they are named Query, Exec or Raw.
			return nativeSQLMethod{kind: "none"}, true
		}
		boundary, known := a.types.boundary(a.expressionType(selector.X), selector.Sel.Name, map[string]bool{})
		if known && boundary.kind != "none" && a.typeExpression(selector.X) {
			boundary.kind = "reject"
		}
		return boundary, known
	}
	return nativeSQLMethod{}, false
}

// A symbolic result can instantiate to a database result. Without generic
// specialization, a signature that could accept raw SQL must remain rejected.
// Domain callbacks that cannot accept a SQL string do not acquire SQL identity.
func (types sqlTypes) ambiguousGenericSQLSignature(ref sqlTypeRef, fn *ast.FuncType) bool {
	unbound := false
	if fn.Results != nil {
		for _, field := range fn.Results.List {
			unbound = unbound || types.unboundSQLResult(ref.with(field.Type))
		}
	}
	if !unbound || fn.Params == nil {
		return false
	}
	for _, field := range fn.Params.List {
		param := ref.with(field.Type)
		if types.typeParameter(param) != nil {
			return true
		}
		identity, known := types.typeIdentity(param, map[string]bool{})
		if !known || identity == "builtin.string" || identity == "...builtin.string" {
			return true
		}
	}
	return false
}
func (types sqlTypes) unboundSQLResult(ref sqlTypeRef) bool {
	ref = types.resolveAliases(ref, map[string]bool{})
	if types.typeParameter(ref) != nil {
		return true
	}
	if pointer, ok := ref.expr.(*ast.StarExpr); ok {
		return types.unboundSQLResult(ref.with(pointer.X))
	}
	return false
}

func (a *sqlAnalyzer) declaredFunction(expr ast.Expr) bool {
	switch e := unparenSQL(expr).(type) {
	case *ast.IndexExpr:
		return a.declaredFunction(e.X)
	case *ast.IndexListExpr:
		return a.declaredFunction(e.X)
	case *ast.Ident:
		return (e.Obj == nil || e.Obj.Kind == ast.Fun) && a.types[a.scope+".$function."+e.Name] != nil
	case *ast.SelectorExpr:
		return a.knownFunction(e)
	}
	return false
}
