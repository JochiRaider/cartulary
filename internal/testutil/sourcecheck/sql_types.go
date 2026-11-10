package sourcecheck

import (
	"go/ast"
	"go/token"
)

// A type keeps its declaration scope when followed through an imported field or
// alias. This is deliberately declaration resolution, not application call-graph
// inference. Every production file in the supplied dependency closure participates.
type sqlTypeRef struct {
	expr      ast.Expr
	scope     string
	imports   map[string]string
	qualified bool
	alias     bool // declaration RHS is an alias, not a new defined type
}
type sqlTypeDefinition struct {
	definitions []sqlTypeRef
	methods     map[string][]sqlTypeRef
}
type sqlTypes map[string]*sqlTypeDefinition

func (types sqlTypes) collect(decl ast.Decl, scope string, imports map[string]string) {
	add := func(name string) *sqlTypeDefinition {
		key := scope + "." + name
		if types[key] == nil {
			types[key] = &sqlTypeDefinition{}
		}
		return types[key]
	}
	if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
		for _, spec := range gen.Specs {
			t := spec.(*ast.TypeSpec)
			d := add(t.Name.Name)
			d.definitions = append(d.definitions, sqlTypeRef{expr: t.Type, scope: scope, imports: imports, alias: t.Assign.IsValid()})
		}
	}
	if fn, ok := decl.(*ast.FuncDecl); ok {
		if fn.Recv == nil {
			d := add("$function." + fn.Name.Name)
			d.definitions = append(d.definitions, sqlTypeRef{expr: fn.Type, scope: scope, imports: imports})
		} else {
			d := add(sqlTypeName(fn.Recv.List[0].Type))
			if d.methods == nil {
				d.methods = map[string][]sqlTypeRef{}
			}
			d.methods[fn.Name.Name] = append(d.methods[fn.Name.Name], sqlTypeRef{expr: fn.Type, scope: scope, imports: imports})
		}
	}
}
func (a *sqlAnalyzer) typeRef(expr ast.Expr) sqlTypeRef {
	return sqlTypeRef{expr: expr, scope: a.scope, imports: a.imports}
}
func (ref sqlTypeRef) with(expr ast.Expr) sqlTypeRef { ref.expr = expr; return ref }
func (ref sqlTypeRef) key() string {
	switch e := ref.expr.(type) {
	case *ast.Ident:
		return ref.scope + "." + e.Name
	case *ast.SelectorExpr:
		if id, ok := e.X.(*ast.Ident); ok && id.Obj == nil {
			return ref.imports[id.Name] + "." + e.Sel.Name
		}
	}
	return ""
}
func (types sqlTypes) database(ref sqlTypeRef, seen map[string]bool) bool {
	if types.nativeIdentity(ref, cloneSQLSeen(seen)) != "" {
		return true
	}
	for _, method := range []string{"Query", "QueryRow", "Exec", "QueryContext", "QueryRowContext", "ExecContext"} {
		if boundary, known := types.boundary(ref, method, cloneSQLSeen(seen)); known && boundary.kind == "sql" {
			return true
		}
	}
	return false
}
func (types sqlTypes) field(ref sqlTypeRef, name string, seen map[string]bool) sqlTypeRef {
	ref = types.resolveAliases(ref, map[string]bool{})
	switch e := ref.expr.(type) {
	case *ast.StarExpr:
		return types.field(ref.with(e.X), name, seen)
	case *ast.ParenExpr:
		return types.field(ref.with(e.X), name, seen)
	case *ast.StructType:
		for _, field := range e.Fields.List {
			for _, n := range field.Names {
				if n.Name == name {
					return ref.with(field.Type)
				}
			}
			if len(field.Names) == 0 {
				if embeddedSQLTypeName(field.Type) == name {
					return ref.with(field.Type)
				}
				if result := types.field(ref.with(field.Type), name, seen); result.expr != nil {
					return result
				}
			}
		}
	}
	if ref.key() == "net/http.Request" && name == "URL" {
		return qualifiedSQLType("net/url", "URL")
	}
	key := ref.key()
	if key == "" || seen[key] {
		return sqlTypeRef{}
	}
	seen[key] = true
	if definition := types[key]; definition != nil {
		// Conflicting build variants cannot silently select a field type.
		if len(definition.definitions) == 1 {
			return types.field(definition.definitions[0], name, seen)
		}
		var result sqlTypeRef
		for _, alternative := range definition.definitions {
			candidate := types.field(alternative, name, cloneSQLSeen(seen))
			if candidate.expr == nil {
				continue
			}
			if result.expr != nil && !types.sameType(candidate, result) {
				return sqlTypeRef{expr: &ast.BadExpr{}}
			}
			result = candidate
		}
		return result
	}
	return sqlTypeRef{}
}
func (a *sqlAnalyzer) expressionType(expr ast.Expr) sqlTypeRef {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Obj != nil && e.Obj.Kind == ast.Typ {
			return a.typeRef(e)
		}
		if ref := a.variableTypes[sqlBinding(e)]; ref.expr != nil {
			if facts := a.bindings[sqlBinding(e)]; facts != nil && facts.nestedWriter {
				ref.qualified = false
			}
			return ref
		}
		if e.Obj == nil || e.Obj.Kind == ast.Fun {
			if def := a.types[a.scope+".$function."+e.Name]; def != nil && len(def.definitions) == 1 {
				return def.definitions[0]
			}
		}
		return sqlTypeRef{}
	case *ast.FuncLit:
		return a.typeRef(e.Type)
	case *ast.ParenExpr:
		return a.expressionType(e.X)
	case *ast.UnaryExpr:
		return a.expressionType(e.X)
	case *ast.StarExpr:
		if ref := a.expressionType(e.X); ref.expr != nil {
			return ref
		}
		return sqlTypeRef{}
	case *ast.SelectorExpr:
		if id, ok := e.X.(*ast.Ident); ok && id.Obj == nil && a.imports[id.Name] != "" {
			if def := a.types[a.imports[id.Name]+".$function."+e.Sel.Name]; def != nil && len(def.definitions) == 1 {
				return def.definitions[0]
			}
			if _, native := nativeSQL[a.imports[id.Name]+"."+e.Sel.Name]; native {
				return a.typeRef(e)
			}
		}
		if field := a.types.field(a.expressionType(e.X), e.Sel.Name, map[string]bool{}); field.expr != nil {
			return field
		}
		return a.types.methodResultAt(a.expressionType(e.X), e.Sel.Name, -1, map[string]bool{})
	case *ast.IndexExpr:
		if signature := a.types.structural(a.expressionType(e.X), map[string]bool{}); signature.expr != nil {
			if fn, ok := signature.expr.(*ast.FuncType); ok && fn.TypeParams != nil {
				return signature
			}
		}
		return a.types.element(a.expressionType(e.X), map[string]bool{})
	case *ast.IndexListExpr:
		return a.expressionType(e.X)
	case *ast.CompositeLit:
		ref := a.typeRef(e.Type)
		ref.qualified = a.types.isBatch(ref) && len(e.Elts) == 0
		return ref
	case *ast.TypeAssertExpr:
		return a.typeRef(e.Type)
	case *ast.CallExpr:
		fun := unparenSQL(e.Fun)
		if a.declaredTypeExpression(fun) && len(e.Args) == 1 {
			if _, known := a.types.callable(a.typeRef(fun), map[string]bool{}); known {
				return a.typeRef(fun)
			}
		}
		if signature := a.types.structural(a.expressionType(fun), map[string]bool{}); signature.expr != nil {
			if _, ok := signature.expr.(*ast.FuncType); ok {
				return sqlResultType(signature)
			}
		}
		if id, ok := fun.(*ast.Ident); ok {
			if id.Name == "make" && id.Obj == nil && len(e.Args) > 0 {
				return a.typeRef(e.Args[0])
			}
			if id.Name == "new" && id.Obj == nil && len(e.Args) == 1 {
				ref := a.typeRef(e.Args[0])
				ref.qualified = a.types.isBatch(ref)
				return ref
			}
			if definition := a.types[a.scope+".$function."+id.Name]; definition != nil && len(definition.definitions) == 1 && (id.Obj == nil || id.Obj.Kind == ast.Fun) {
				return sqlResultType(definition.definitions[0])
			}
		}
		if sel, ok := fun.(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok && id.Obj == nil && a.imports[id.Name] != "" {
				scope := a.imports[id.Name]
				if definition := a.types[scope+".$function."+sel.Sel.Name]; definition != nil && len(definition.definitions) == 1 {
					return sqlResultType(definition.definitions[0])
				}
				switch scope + "." + sel.Sel.Name {
				case "net/url.Parse", "net/url.ParseRequestURI":
					return qualifiedSQLType("net/url", "URL")
				case "net/http.NewRequest", "net/http.NewRequestWithContext":
					return qualifiedSQLType("net/http", "Request")
				case "github.com/jackc/pgx/v5.Connect", "github.com/jackc/pgx/v5.ConnectConfig":
					return qualifiedSQLType("github.com/jackc/pgx/v5", "Conn")
				case "github.com/jackc/pgx/v5/pgxpool.New", "github.com/jackc/pgx/v5/pgxpool.NewWithConfig":
					return qualifiedSQLType("github.com/jackc/pgx/v5/pgxpool", "Pool")
				case "github.com/jackc/pgx/v5/pgconn.Connect", "github.com/jackc/pgx/v5/pgconn.ConnectConfig":
					return qualifiedSQLType(pgconnPackage, "PgConn")
				case "database/sql.Open", "database/sql.OpenDB":
					return qualifiedSQLType("database/sql", "DB")
				}
			}
			receiver := a.expressionType(sel.X)
			if boundary, known := a.types.boundary(receiver, sel.Sel.Name, map[string]bool{}); known && boundary.result != "" {
				ref := nativeTypeRef(boundary.result)
				ref.qualified = boundary.kind == "sql" && boundary.result == "database/sql.Stmt"
				return ref
			}
			return a.types.methodResult(receiver, sel.Sel.Name, map[string]bool{})
		}
	}
	return sqlTypeRef{}
}
func (a *sqlAnalyzer) databaseType(expr ast.Expr) bool {
	return a.types.database(a.typeRef(expr), map[string]bool{})
}
func (a *sqlAnalyzer) databaseField(expr *ast.SelectorExpr) bool {
	return a.types.database(a.expressionType(expr), map[string]bool{})
}

func qualifiedSQLType(scope, name string) sqlTypeRef {
	return sqlTypeRef{expr: &ast.Ident{Name: name}, scope: scope}
}
func sqlResultType(ref sqlTypeRef) sqlTypeRef { return sqlResultTypeAt(ref, 0) }
func sqlResultTypeAt(ref sqlTypeRef, index int) sqlTypeRef {
	if index < 0 {
		return ref
	}
	if signature, ok := ref.expr.(*ast.FuncType); ok && signature.Results != nil {
		position := 0
		for _, field := range signature.Results.List {
			count := len(field.Names)
			if count == 0 {
				count = 1
			}
			if index >= position && index < position+count {
				return ref.with(field.Type)
			}
			position += count
		}
	}
	return sqlTypeRef{}
}
func (types sqlTypes) methodResult(ref sqlTypeRef, name string, seen map[string]bool) sqlTypeRef {
	return types.methodResultAt(ref, name, 0, seen)
}
func (types sqlTypes) methodResultAt(ref sqlTypeRef, name string, index int, seen map[string]bool) sqlTypeRef {
	ref = types.resolveAliases(ref, map[string]bool{})
	switch e := ref.expr.(type) {
	case *ast.StarExpr:
		return types.methodResultAt(ref.with(e.X), name, index, seen)
	case *ast.ParenExpr:
		return types.methodResultAt(ref.with(e.X), name, index, seen)
	case *ast.InterfaceType:
		for _, field := range e.Methods.List {
			for _, id := range field.Names {
				if id.Name == name {
					return sqlResultTypeAt(ref.with(field.Type), index)
				}
			}
			if len(field.Names) == 0 {
				if result := types.methodResultAt(ref.with(field.Type), name, index, seen); result.expr != nil {
					return result
				}
			}
		}
	}
	key := ref.key()
	if key == "" || seen[key] {
		return sqlTypeRef{}
	}
	seen[key] = true
	if definition := types[key]; definition != nil {
		if methods, ok := definition.methods[name]; ok {
			var result sqlTypeRef
			for i, method := range methods {
				candidate := sqlResultTypeAt(method, index)
				if i > 0 && !types.sameType(candidate, result) {
					return sqlTypeRef{expr: &ast.BadExpr{}}
				}
				result = candidate
			}
			return result
		}
		for _, alternative := range definition.definitions {
			if result := types.methodResultAt(alternative, name, index, seen); result.expr != nil {
				return result
			}
		}
	}
	return sqlTypeRef{}
}

func embeddedSQLTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return embeddedSQLTypeName(e.X)
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.Ident:
		return e.Name
	}
	return ""
}

func (a *sqlAnalyzer) typeExpression(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.ParenExpr:
		return a.typeExpression(e.X)
	case *ast.StarExpr:
		return a.typeExpression(e.X)
	case *ast.Ident:
		return e.Obj != nil && e.Obj.Kind == ast.Typ
	case *ast.SelectorExpr:
		id, ok := e.X.(*ast.Ident)
		return ok && id.Obj == nil && a.imports[id.Name] != ""
	}
	return false
}

func (a *sqlAnalyzer) expressionResult(expr ast.Expr, index int) sqlTypeRef {
	if index == 0 {
		return a.expressionType(expr)
	}
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return sqlTypeRef{}
	}
	fun := unparenSQL(call.Fun)
	if ref := a.types.structural(a.expressionType(fun), map[string]bool{}); ref.expr != nil {
		if result := sqlResultTypeAt(ref, index); result.expr != nil {
			return result
		}
	}
	if selector, ok := fun.(*ast.SelectorExpr); ok {
		return a.types.methodResultAt(a.expressionType(selector.X), selector.Sel.Name, index, map[string]bool{})
	}
	return sqlTypeRef{}
}

func (types sqlTypes) element(ref sqlTypeRef, seen map[string]bool) sqlTypeRef {
	ref = types.resolveAliases(ref, map[string]bool{})
	switch e := ref.expr.(type) {
	case *ast.StarExpr:
		return types.element(ref.with(e.X), seen)
	case *ast.ArrayType:
		return ref.with(e.Elt)
	case *ast.MapType:
		return ref.with(e.Value)
	case *ast.ChanType:
		return ref.with(e.Value)
	}
	key := ref.key()
	if key == "" || seen[key] {
		return sqlTypeRef{}
	}
	seen[key] = true
	if definition := types[key]; definition != nil && len(definition.definitions) == 1 {
		return types.element(definition.definitions[0], seen)
	}
	return sqlTypeRef{}
}
func (a *sqlAnalyzer) knownFunction(selector *ast.SelectorExpr) bool {
	id, ok := selector.X.(*ast.Ident)
	if !ok || id.Obj != nil {
		return false
	}
	imported := a.imports[id.Name]
	if imported == "" {
		return false
	}
	if imported == "golang.org/x/sys/unix" && selector.Sel.Name == "Exec" {
		return true
	}
	return a.types[imported+".$function."+selector.Sel.Name] != nil
}

// A parsed type reference is not an escaped execution method, even when the
// imported type happens to be named Query or Exec.
func (a *sqlAnalyzer) knownType(selector *ast.SelectorExpr) bool {
	id, ok := selector.X.(*ast.Ident)
	return ok && id.Obj == nil && a.imports[id.Name] != "" && a.types[a.imports[id.Name]+"."+selector.Sel.Name] != nil
}

// A package selector may be a function, not a conversion. Only a declared type
// can supply the type of a converted function value.
func (a *sqlAnalyzer) declaredTypeExpression(expr ast.Expr) bool {
	switch e := unparenSQL(expr).(type) {
	case *ast.IndexExpr:
		return a.declaredTypeExpression(e.X)
	case *ast.IndexListExpr:
		return a.declaredTypeExpression(e.X)
	case *ast.Ident:
		return e.Obj != nil && e.Obj.Kind == ast.Typ || a.types[a.scope+"."+e.Name] != nil
	case *ast.SelectorExpr:
		return a.knownType(e)
	}
	return false
}
