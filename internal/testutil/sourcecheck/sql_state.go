package sourcecheck

import (
	"go/ast"
	"strconv"
	"strings"
)

// sqlState is the complete path-dependent evaluation state. Declaration facts,
// discovered candidates, diagnostics and function return summaries are separate.
type sqlState struct {
	values      sqlEnv
	types       map[string]sqlTypeRef
	collections map[string]sqlCollection
}

// A collection stores construction-time element alternatives. The empty suffix
// denotes a scalar element; .field denotes an anonymous record field. A known
// empty collection has no members. Unknown collections never supply alternatives.
type sqlCollection struct {
	known   bool
	members sqlEnv
}

func (a *sqlAnalyzer) state() sqlState {
	collections := map[string]sqlCollection{}
	for key, collection := range a.collections {
		collection.members = copySQLEnv(collection.members)
		collections[key] = collection
	}
	return sqlState{copySQLEnv(a.env), copySQLTypes(a.variableTypes), collections}
}

func (a *sqlAnalyzer) setState(state sqlState) {
	a.env, a.variableTypes, a.collections = state.values, state.types, state.collections
}

func mergeSQLMembers(left, right sqlEnv) sqlEnv {
	joined := sqlEnv{}
	for key, values := range left {
		other, exists := right[key]
		if !exists {
			other = unknownSQL("value absent on a reaching path")
		}
		joined[key] = mergeSQLValues(values, other)
	}
	for key, values := range right {
		if _, exists := left[key]; !exists {
			joined[key] = mergeSQLValues(unknownSQL("value absent on a reaching path"), values)
		}
	}
	return joined
}

func mergeSQLCollections(left, right sqlCollection) sqlCollection {
	if !left.known || !right.known {
		return sqlCollection{}
	}
	if len(left.members) == 0 {
		return sqlCollection{true, copySQLEnv(right.members)}
	}
	if len(right.members) == 0 {
		return sqlCollection{true, copySQLEnv(left.members)}
	}
	return sqlCollection{true, mergeSQLMembers(left.members, right.members)}
}

func (a *sqlAnalyzer) joinStates(states ...sqlState) sqlState {
	if len(states) == 0 {
		return sqlState{sqlEnv{}, map[string]sqlTypeRef{}, map[string]sqlCollection{}}
	}
	joined := states[0]
	for _, next := range states[1:] {
		joined.values = mergeSQLMembers(joined.values, next.values)
		types := map[string]sqlTypeRef{}
		for key, left := range joined.types {
			right, exists := next.types[key]
			// Static declaration identity is independent of path reachability;
			// absence of an inferred type is not a conflicting declaration.
			if left.expr == nil {
				right.qualified = false
				types[key] = right
			} else if !exists || right.expr == nil {
				left.qualified = false
				types[key] = left
			} else if !(left.expr == right.expr && left.scope == right.scope || a.types.sameType(left, right)) {
				types[key] = sqlTypeRef{expr: &ast.BadExpr{}}
			} else {
				left.qualified = left.qualified && right.qualified
				types[key] = left
			}
		}
		for key, ref := range next.types {
			if _, exists := joined.types[key]; !exists {
				ref.qualified = false
				types[key] = ref
			}
		}
		joined.types = types
		collections := map[string]sqlCollection{}
		for key, left := range joined.collections {
			collections[key] = mergeSQLCollections(left, next.collections[key])
		}
		for key := range next.collections {
			if _, exists := joined.collections[key]; !exists {
				collections[key] = sqlCollection{}
			}
		}
		joined.collections = collections
	}
	return joined
}

func sqlBindingMember(key, binding string) bool {
	return key == binding || strings.HasPrefix(key, binding+".")
}

func (a *sqlAnalyzer) restoreLocal(binding string, outer sqlState) {
	for key, values := range a.env {
		if sqlBindingMember(key, binding) {
			a.emit(values, false)
			delete(a.env, key)
		}
	}
	delete(a.variableTypes, binding)
	delete(a.collections, binding)
	for key, values := range outer.values {
		if sqlBindingMember(key, binding) {
			a.env[key] = values
		}
	}
	if ref, exists := outer.types[binding]; exists {
		a.variableTypes[binding] = ref
	}
	if collection, exists := outer.collections[binding]; exists {
		a.collections[binding] = collection
	}
}

func (a *sqlAnalyzer) collectionValue(expression ast.Expr) sqlCollection {
	if id, ok := unparenSQL(expression).(*ast.Ident); ok {
		facts := a.bindings[sqlBinding(id)]
		if facts == nil || facts.externallyMutable() || facts.collectionEscaped {
			return sqlCollection{}
		}
		return a.collections[sqlBinding(id)]
	}
	literal, ok := unparenSQL(expression).(*ast.CompositeLit)
	if !ok || len(literal.Elts) > 64 {
		return sqlCollection{}
	}
	array, ok := literal.Type.(*ast.ArrayType)
	if !ok {
		return sqlCollection{}
	}
	if array.Len != nil {
		if _, inferred := array.Len.(*ast.Ellipsis); !inferred {
			length, ok := array.Len.(*ast.BasicLit)
			if !ok {
				return sqlCollection{}
			}
			n, err := strconv.Atoi(length.Value)
			if err != nil || n != len(literal.Elts) {
				return sqlCollection{}
			}
		}
	}
	fields := []string{}
	if structure, ok := array.Elt.(*ast.StructType); ok {
		for _, field := range structure.Fields.List {
			for _, name := range field.Names {
				fields = append(fields, name.Name)
			}
		}
	}
	result := sqlCollection{known: true, members: sqlEnv{}}
	for i, item := range literal.Elts {
		members := sqlEnv{}
		if record, ok := item.(*ast.CompositeLit); ok {
			if len(fields) == 0 {
				return sqlCollection{}
			}
			for _, field := range fields {
				members["."+field] = unknownSQL("omitted collection field")
			}
			for index, element := range record.Elts {
				name := ""
				value := element
				if pair, ok := element.(*ast.KeyValueExpr); ok {
					name = sqlExprName(pair.Key)
					value = pair.Value
				} else if index < len(fields) {
					name = fields[index]
				}
				if name == "" {
					return sqlCollection{}
				}
				members["."+name] = a.value(value, map[string]bool{})
			}
		} else {
			if _, keyed := item.(*ast.KeyValueExpr); keyed {
				return sqlCollection{}
			}
			members[""] = a.value(item, map[string]bool{})
		}
		if i == 0 {
			result.members = members
		} else {
			result.members = mergeSQLMembers(result.members, members)
		}
	}
	return result
}
