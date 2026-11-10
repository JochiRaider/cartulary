package sourcecheck

import (
	"fmt"
	"go/ast"
	"sort"
	"strings"
)

// resolveAliases preserves the declaring scope and expands aliases only. A
// defined type keeps its identity; method-set traversal is a separate operation.
// All build alternatives must agree. BadExpr represents unresolved identity,
// never evidence that an expression or method is SQL-free.
func (types sqlTypes) resolveAliases(ref sqlTypeRef, seen map[string]bool) sqlTypeRef {
	switch e := ref.expr.(type) {
	case *ast.ParenExpr:
		return types.resolveAliases(ref.with(e.X), seen)
	case *ast.StarExpr:
		inner := types.resolveAliases(ref.with(e.X), seen)
		if _, bad := inner.expr.(*ast.BadExpr); bad {
			return inner
		}
		inner.expr = &ast.StarExpr{X: inner.expr}
		inner.qualified = ref.qualified
		return inner
	case *ast.Ident:
		if declaration := types.localType(ref); declaration != nil {
			if !declaration.Assign.IsValid() {
				return ref
			}
			key := fmt.Sprintf("alias-local:%d", declaration.Pos())
			if seen[key] {
				return sqlTypeRef{expr: &ast.BadExpr{}}
			}
			seen[key] = true
			return types.resolveAliases(ref.with(declaration.Type), seen)
		}
	}
	key := ref.key()
	definition := types[key]
	if definition == nil {
		return ref
	}
	aliases := false
	for _, candidate := range definition.definitions {
		aliases = aliases || candidate.alias
	}
	if !aliases {
		return ref
	}
	if seen[key] {
		return sqlTypeRef{expr: &ast.BadExpr{}}
	}
	seen[key] = true
	var result sqlTypeRef
	identity := ""
	for _, candidate := range definition.definitions {
		if !candidate.alias {
			return sqlTypeRef{expr: &ast.BadExpr{}}
		}
		resolved := types.resolveAliases(candidate, cloneSQLSeen(seen))
		current, known := types.typeIdentity(resolved, cloneSQLSeen(seen))
		if !known || identity != "" && identity != current {
			return sqlTypeRef{expr: &ast.BadExpr{}}
		}
		result, identity = resolved, current
	}
	result.qualified = ref.qualified
	return result
}

func (types sqlTypes) localType(ref sqlTypeRef) *ast.TypeSpec {
	id, ok := ref.expr.(*ast.Ident)
	if !ok || id.Obj == nil {
		return nil
	}
	decl, ok := id.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		return nil
	}
	if definition := types[ref.key()]; definition != nil {
		for _, candidate := range definition.definitions {
			if candidate.expr == decl.Type {
				return nil
			}
		}
	}
	return decl
}

// typeIdentity is shared by signature admission and field/result agreement.
// Unknown repository declarations are distinct from explicit foreign boundaries.
func (types sqlTypes) typeIdentity(ref sqlTypeRef, seen map[string]bool) (string, bool) {
	ref = types.resolveAliases(ref, cloneSQLSeen(seen))
	switch e := ref.expr.(type) {
	case nil, *ast.BadExpr:
		return "", false
	case *ast.StarExpr:
		key, known := types.typeIdentity(ref.with(e.X), seen)
		return "*" + key, known
	case *ast.Ident:
		if declaration := types.localType(ref); declaration != nil {
			return fmt.Sprintf("local:%d", declaration.Pos()), true
		}
		if e.Obj != nil && e.Obj.Kind == ast.Typ {
			if declaration, ok := e.Obj.Decl.(*ast.Field); ok {
				return fmt.Sprintf("parameter:%d", declaration.Pos()), true
			}
		}
		if types[ref.key()] == nil && e.Obj == nil {
			switch e.Name {
			case "any", "comparable", "bool", "byte", "complex64", "complex128", "error", "float32", "float64", "int", "int8", "int16", "int32", "int64", "rune", "string", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
				return "builtin." + e.Name, true
			}
		}
	case *ast.ArrayType:
		key, known := types.typeIdentity(ref.with(e.Elt), seen)
		if e.Len != nil {
			return "", false
		} // unqualified array length identity
		return "[]" + key, known
	case *ast.Ellipsis:
		key, known := types.typeIdentity(ref.with(e.Elt), seen)
		return "..." + key, known
	case *ast.MapType:
		key, known := types.typeIdentity(ref.with(e.Key), cloneSQLSeen(seen))
		value, valueKnown := types.typeIdentity(ref.with(e.Value), cloneSQLSeen(seen))
		return "map[" + key + "]" + value, known && valueKnown
	case *ast.ChanType:
		key, known := types.typeIdentity(ref.with(e.Value), seen)
		return fmt.Sprintf("chan%d:%s", e.Dir, key), known
	case *ast.FuncType:
		params, pKnown := types.fieldIdentity(ref, e.Params, false, false, seen)
		results, rKnown := types.fieldIdentity(ref, e.Results, false, false, seen)
		return "func(" + params + ")(" + results + ")", pKnown && rKnown
	case *ast.InterfaceType:
		key, known := types.fieldIdentity(ref, e.Methods, true, true, seen)
		return "interface{" + key + "}", known
	case *ast.StructType:
		key, known := types.fieldIdentity(ref, e.Fields, true, false, seen)
		return "struct{" + key + "}", known
	}
	key := ref.key()
	if key == "" {
		return "", false
	}
	if types[key] != nil {
		return key, true
	}
	return key, !repositorySQLType(key) && strings.Contains(key, ".")
}

func (types sqlTypes) fieldIdentity(ref sqlTypeRef, fields *ast.FieldList, names, unordered bool, seen map[string]bool) (string, bool) {
	if fields == nil {
		return "", true
	}
	parts := []string{}
	for _, field := range fields.List {
		key, known := types.typeIdentity(ref.with(field.Type), cloneSQLSeen(seen))
		if !known {
			return "", false
		}
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			part := key
			if names && len(field.Names) > 0 {
				part = field.Names[i].Name + ":" + part
			}
			if field.Tag != nil {
				part += " tag:" + field.Tag.Value
			}
			parts = append(parts, part)
		}
	}
	if unordered {
		sort.Strings(parts)
	}
	return strings.Join(parts, ";"), true
}

func (types sqlTypes) sameType(left, right sqlTypeRef) bool {
	a, aKnown := types.typeIdentity(left, map[string]bool{})
	b, bKnown := types.typeIdentity(right, map[string]bool{})
	return aKnown && bKnown && a == b
}

// Only structural declarations inherit fields/methods. A new named type whose
// underlying type is a native concrete handle does not inherit its methods.
func (types sqlTypes) structural(ref sqlTypeRef, seen map[string]bool) sqlTypeRef {
	ref = types.resolveAliases(ref, cloneSQLSeen(seen))
	switch ref.expr.(type) {
	case *ast.StructType, *ast.InterfaceType, *ast.FuncType, *ast.BadExpr:
		return ref
	case *ast.IndexExpr:
		return types.structural(ref.with(ref.expr.(*ast.IndexExpr).X), seen)
	case *ast.IndexListExpr:
		return types.structural(ref.with(ref.expr.(*ast.IndexListExpr).X), seen)
	}
	key := ref.key()
	if key == "" || seen[key] {
		return sqlTypeRef{}
	}
	seen[key] = true
	if definition := types[key]; definition != nil && len(definition.definitions) == 1 {
		return types.structural(definition.definitions[0], seen)
	}
	return sqlTypeRef{}
}

// A new defined result type does not acquire a native signature grant. If it
// derives from a database result, it also cannot prove a SQL-free boundary.
func nativeSQLResult(key string) bool {
	switch key {
	case pgxPackage + ".Rows", pgxPackage + ".Row", pgconnPackage + ".CommandTag", "*database/sql.Rows", "*database/sql.Row", "database/sql.Result", "*database/sql.Stmt", "*" + pgxPackage + ".QueuedQuery", "*" + pgconnPackage + ".MultiResultReader", "*" + pgconnPackage + ".ResultReader", "*" + pgconnPackage + ".StatementDescription":
		return true
	}
	return false
}
func (types sqlTypes) derivedSQLResult(ref sqlTypeRef, seen map[string]bool) bool {
	ref = types.resolveAliases(ref, map[string]bool{})
	identity, known := types.typeIdentity(ref, map[string]bool{})
	if known && nativeSQLResult(identity) {
		return true
	}
	if star, ok := ref.expr.(*ast.StarExpr); ok {
		return types.derivedSQLResult(ref.with(star.X), seen)
	}
	if local := types.localType(ref); local != nil {
		key := fmt.Sprintf("result-local:%d", local.Pos())
		if seen[key] {
			return true
		}
		seen[key] = true
		return types.derivedSQLResult(ref.with(local.Type), seen)
	}
	if e, ok := ref.expr.(*ast.InterfaceType); ok {
		for _, field := range e.Methods.List {
			if len(field.Names) == 0 && types.derivedSQLResult(ref.with(field.Type), cloneSQLSeen(seen)) {
				return true
			}
		}
	}
	key := ref.key()
	if key == "" {
		return false
	}
	if seen[key] {
		return true
	}
	seen[key] = true
	if definition := types[key]; definition != nil {
		for _, candidate := range definition.definitions {
			if types.derivedSQLResult(candidate, cloneSQLSeen(seen)) {
				return true
			}
		}
	}
	return false
}

// This separately versioned module shares the repository's namespace but is an
// explicit foreign-library boundary. The backend gate also enforces its runtime
// dependency policy; it is not a repository-local package or a SQL capability.
func repositorySQLType(key string) bool {
	return strings.HasPrefix(key, "github.com/JochiRaider/cartulary/") && !strings.HasPrefix(key, "github.com/JochiRaider/cartulary/referencepacks/format/")
}

// All result positions participate in signature admission. Data containers do
// not confer a native result capability; unresolved/derived identities reject.
func (types sqlTypes) portResult(ref sqlTypeRef) string {
	ref = types.resolveAliases(ref, map[string]bool{})
	if _, ok := types.structural(ref, map[string]bool{}).expr.(*ast.StructType); ok {
		return "none"
	}
	switch ref.expr.(type) {
	case *ast.ArrayType, *ast.MapType, *ast.ChanType, *ast.FuncType, *ast.StructType:
		return "none"
	}
	key, known := types.typeIdentity(ref, map[string]bool{})
	if !known {
		return "reject"
	}
	if nativeSQLResult(key) {
		return "sql"
	}
	if types.derivedSQLResult(ref, map[string]bool{}) {
		return "reject"
	}
	return "none"
}

func (types sqlTypes) typeParameter(ref sqlTypeRef) *ast.Field {
	ref = types.resolveAliases(ref, map[string]bool{})
	if id, ok := ref.expr.(*ast.Ident); ok && id.Obj != nil && id.Obj.Kind == ast.Typ {
		if field, ok := id.Obj.Decl.(*ast.Field); ok {
			return field
		}
	}
	return nil
}
