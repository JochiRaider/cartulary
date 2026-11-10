package sourcecheck

import (
	"fmt"
	"go/ast"
	"go/token"
)

type sqlAssignmentValue struct {
	values     []sqlValue
	typ        sqlTypeRef
	collection sqlCollection
}

// Assignment evaluates all inputs before changing any destination. Result
// identity can be known independently of its raw string construction; unsupported
// result positions explicitly lose their previous scalar and collection values.
func (a *sqlAnalyzer) assign(left, right []ast.Expr, op token.Token, declaredType ast.Expr, declaration ast.Node, locals map[string]bool) {
	previousSite := a.current
	a.current = declaration
	defer func() { a.current = previousSite }()
	evaluated := make([]sqlAssignmentValue, len(left))
	for i := range left {
		value := sqlAssignmentValue{values: unknownSQL(fmt.Sprintf("unsupported assignment result position %d", i+1))}
		if len(right) == 1 {
			value.typ = a.expressionResult(right[0], i)
		} else if i < len(right) {
			value.typ = a.expressionType(right[i])
		}
		if i < len(right) {
			value.values = a.value(right[i], map[string]bool{})
			value.collection = a.collectionValue(right[i])
		}
		if declaredType != nil {
			value.typ = a.typeRef(declaredType)
			value.typ.qualified = len(right) == 0 && a.types.isBatch(value.typ)
		}
		if op == token.ADD_ASSIGN {
			if id, ok := left[i].(*ast.Ident); ok {
				value.values = joinSQLValues(a.value(id, map[string]bool{}), value.values)
			}
			value.collection = sqlCollection{}
		} else if op != token.ASSIGN && op != token.DEFINE && op != token.VAR && op != token.CONST {
			value.values = unknownSQL("unsupported compound assignment")
			value.collection = sqlCollection{}
		}
		evaluated[i] = value
	}
	for _, expr := range right {
		a.observe(expr)
	}
	for _, expr := range left {
		if _, simple := expr.(*ast.Ident); !simple {
			a.observe(expr)
		}
	}
	for i, expr := range left {
		id, ok := expr.(*ast.Ident)
		if !ok || id.Name == "_" {
			continue
		}
		key := sqlBinding(id)
		value := evaluated[i]
		previous, previouslyBound := a.variableTypes[key]
		if value.typ.expr == nil || previouslyBound {
			// Reassignment changes the value, never the binding's declared Go
			// type. In particular, concrete implementations cannot erase a port.
			value.typ = previous
		}
		if previouslyBound {
			value.typ.qualified = false
		}
		// Clear synthetic fields as well as the whole-binding collection.
		for member := range a.env {
			if sqlBindingMember(member, key) {
				delete(a.env, member)
			}
		}
		a.env[key] = value.values
		a.variableTypes[key] = value.typ
		a.collections[key] = value.collection
		// Preserve SQL-producing compound assignments even when a loop exit
		// discards their values and no execution call consumes the binding.
		a.emit(value.values, false)
		if id.Obj != nil && id.Obj.Decl == declaration && (op == token.DEFINE || op == token.VAR || op == token.CONST) {
			locals[key] = true
		}
	}
}
