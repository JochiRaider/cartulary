package sourcecheck

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/sqltemplate"
)

// SQLExpression is a bounded construction result, not a SQL access decision.
// Unknown operands remain visible; consumers must never treat them as no SQL.
type SQLExpression struct {
	Kind        string   `json:"kind"`
	Text        string   `json:"text"`
	Dynamic     bool     `json:"dynamic"`
	Unresolved  []string `json:"unresolved"`
	Line        int      `json:"line"`
	Column      int      `json:"column"`
	Disposition string   `json:"disposition"`
}

type SQLDeclaration struct {
	Name string          `json:"name"`
	SQL  []SQLExpression `json:"sql"`
}

const closedSQLText = "\x00closed-sql-text"

type sqlValue struct {
	text    string
	unknown []string
}
type sqlEnv map[string][]sqlValue
type sqlAnalyzer struct {
	packageConstantBindings map[*ast.Object]bool
	bindings                map[string]*sqlBindingFacts
	packageConstants        map[string]map[string]ast.Expr
	constantImports         map[ast.Expr]map[string]string

	functions       map[string]*ast.FuncDecl
	functionImports map[*ast.FuncDecl]map[string]string
	fset            *token.FileSet
	current         ast.Node
	observed        map[ast.Expr]bool
	directCalls     map[ast.Expr]bool
	caseTypes       map[*ast.CaseClause]map[string]sqlTypeRef
	types           sqlTypes
	scope           string
	variableTypes   map[string]sqlTypeRef
	collections     map[string]sqlCollection
	calls           map[string]bool
	returnSummary   *sqlReturnSummary
	breaks          []sqlState
	continues       []sqlState

	imports   map[string]string
	constants map[string]ast.Expr
	env       sqlEnv
	results   []SQLExpression
	seen      map[string]bool
}

var sqlPreparedName = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)

var sqlStart = regexp.MustCompile(`(?is)^\s*(?:/\*|--|(?:select|with|insert|update|delete|from|join|left\s+join|right\s+join|cross\s+join|lock\s+table|truncate|table)\b)`)
var sqlAnywhere = regexp.MustCompile(`(?is)\b(?:select|with|insert\s+into|update\s+\w+\s+set|delete\s+from|for\s+(?:update|share|no\s+key\s+update|key\s+share))\b`)
var sqlNested = regexp.MustCompile(`(?is)\(\s*(?:select|with|table|values)\b`)

func unknownSQL(reason string) []sqlValue { return []sqlValue{{"%s", []string{reason}}} }
func copySQLEnv(env sqlEnv) sqlEnv {
	out := sqlEnv{}
	for k, v := range env {
		out[k] = append([]sqlValue(nil), v...)
		for i := range out[k] {
			out[k][i].unknown = append([]string(nil), v[i].unknown...)
		}
	}
	return out
}
func mergeSQLValues(a, b []sqlValue) []sqlValue {
	seen := map[string]bool{}
	out := []sqlValue{}
	for _, v := range append(append([]sqlValue{}, a...), b...) {
		key := v.text + "\x00" + strings.Join(v.unknown, "\x00")
		if !seen[key] {
			seen[key] = true
			out = append(out, v)
		}
	}
	if len(out) > 256 {
		return unknownSQL("construction exceeds 256 alternatives")
	}
	return out
}
func joinSQLValues(a, b []sqlValue) []sqlValue {
	if len(a)*len(b) > 256 {
		return unknownSQL("construction exceeds 256 alternatives")
	}
	out := []sqlValue{}
	for _, x := range a {
		for _, y := range b {
			if x.text == closedSQLText || y.text == closedSQLText {
				return []sqlValue{{"SELECT %s", []string{"closed SQL text must remain a typed fragment during composition"}}}
			}
			out = append(out, sqlValue{x.text + y.text, append(append([]string{}, x.unknown...), y.unknown...)})
		}
	}
	return out
}

// AnalyzeSQL parses the complete batch so package constants survive file splits.
// It does not load packages, evaluate build tags, or acquire files itself.
func AnalyzeSQL(sources map[string]string) (map[string][]SQLDeclaration, error) {
	files := map[string]*ast.File{}
	fset := token.NewFileSet()
	importsByFile := map[string]map[string]string{}
	functionImports := map[*ast.FuncDecl]map[string]string{}
	types := sqlTypes{}
	constants := map[string]map[string]ast.Expr{}
	packageConstants := map[string]map[string]ast.Expr{}
	constantImports := map[ast.Expr]map[string]string{}
	constantBindings := map[*ast.Object]bool{}
	functions := map[string]map[string]*ast.FuncDecl{}
	names := []string{}
	for name := range sources {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		file, err := parser.ParseFile(fset, name, sources[name], 0)
		if err != nil {
			return nil, fmt.Errorf("parse SQL %s: %w", name, err)
		}
		files[name] = file
		imports := map[string]string{}
		for _, spec := range file.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return nil, err
			}
			alias := path.Base(imported)
			if imported == "github.com/jackc/pgx/v5" {
				alias = "pgx"
			}
			if spec.Name != nil {
				alias = spec.Name.Name
			}
			if alias == "." {
				return nil, fmt.Errorf("%s: dot import %q is unsupported in SQL analysis; use a named import", fset.Position(spec.Pos()), imported)
			}
			imports[alias] = imported
		}

		importsByFile[name] = imports

		pkg := path.Dir(name) + "/" + file.Name.Name
		if constants[pkg] == nil {
			constants[pkg] = map[string]ast.Expr{}
			packageConstants["github.com/JochiRaider/cartulary/"+path.Dir(name)] = constants[pkg]
			functions[pkg] = map[string]*ast.FuncDecl{}
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil {
				if _, duplicate := functions[pkg][fn.Name.Name]; duplicate {
					functions[pkg][fn.Name.Name] = nil
				} else {
					functions[pkg][fn.Name.Name] = fn
				}
				functionImports[fn] = imports
			}
			scope := "github.com/JochiRaider/cartulary/" + path.Dir(name)
			types.collect(decl, scope, imports)
			if scope == "github.com/JochiRaider/cartulary/internal/platform/sqlshape" {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.IsExported() {
					allowed := map[string]bool{"Command": true, "CommandText": true, "Identifier": true, "Literal": true, "Quote": true, "Number": true, "Parameter": true, "Text": true, "Empty": true, "Concat": true, "Join": true, "Format": true, "Write": true, "Fragment": true, "Operator": true}
					if !allowed[fn.Name.Name] {
						return nil, fmt.Errorf("%s: unqualified SQL constructor surface %s", name, fn.Name.Name)
					}
				}
			}

			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.CONST {
				var previous []ast.Expr
				for _, spec := range gen.Specs {
					v := spec.(*ast.ValueSpec)
					values := v.Values
					if len(values) == 0 {
						values = previous
					}
					previous = values
					for i, n := range v.Names {
						if n.Obj != nil {
							constantBindings[n.Obj] = true
						}
						if i < len(values) {
							if _, duplicate := constants[pkg][n.Name]; duplicate {
								constants[pkg][n.Name] = &ast.BadExpr{}
							} else {
								constants[pkg][n.Name] = values[i]
							}
							constantImports[values[i]] = imports
						}
					}
				}
			}
		}
	}
	globals := map[string]map[string]sqlTypeRef{}
	for _, name := range names {
		scope := "github.com/JochiRaider/cartulary/" + path.Dir(name)
		if globals[scope] == nil {
			globals[scope] = map[string]sqlTypeRef{}
		}
		for _, decl := range files[name].Decls {
			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
				for _, spec := range gen.Specs {
					v := spec.(*ast.ValueSpec)
					for _, id := range v.Names {
						if v.Type != nil {
							ref := sqlTypeRef{expr: v.Type, scope: scope, imports: importsByFile[name]}
							if previous, exists := globals[scope][id.Name]; exists && !types.sameType(previous, ref) {
								ref = sqlTypeRef{expr: &ast.BadExpr{}}
							}
							globals[scope][id.Name] = ref
							globals[scope][sqlBinding(id)] = ref
						}
					}
				}
			}
		}
	}
	// Infer package bindings from declared factory results. Unknown factories do
	// not acquire a database identity; explicit types remain authoritative.
	for pass := 0; pass < len(names); pass++ {
		changed := false
		for _, name := range names {
			scope := "github.com/JochiRaider/cartulary/" + path.Dir(name)
			resolver := &sqlAnalyzer{types: types, scope: scope, imports: importsByFile[name], variableTypes: globals[scope]}
			for _, decl := range files[name].Decls {
				if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
					for _, spec := range gen.Specs {
						v := spec.(*ast.ValueSpec)
						if v.Type != nil {
							continue
						}
						for i, id := range v.Names {
							var ref sqlTypeRef
							if len(v.Values) == 1 {
								ref = resolver.expressionResult(v.Values[0], i)
							} else if i < len(v.Values) {
								ref = resolver.expressionType(v.Values[i])
							}
							ref.qualified = false // package state is not local preparation or batch provenance
							if ref.expr != nil && globals[scope][id.Name].expr == nil {
								globals[scope][id.Name] = ref
								globals[scope][sqlBinding(id)] = ref
								changed = true
							}
						}
					}
				}
			}
		}
		if !changed {
			break
		}
	}
	out := map[string][]SQLDeclaration{}
	for _, name := range names {
		file := files[name]
		imports := importsByFile[name]
		direct := map[ast.Expr]bool{}
		ast.Inspect(file, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				fun := unparenSQL(call.Fun)
				direct[fun] = true
				switch instance := fun.(type) {
				case *ast.IndexExpr:
					direct[unparenSQL(instance.X)] = true
				case *ast.IndexListExpr:
					direct[unparenSQL(instance.X)] = true
				}
			}
			return true
		})
		var escaped error
		ast.Inspect(file, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			id, ok := selector.X.(*ast.Ident)
			if !ok || id.Obj != nil {
				return true
			}
			if imports[id.Name] == "github.com/JochiRaider/cartulary/internal/platform/sqlshape" && (selector.Sel.Name == "Literal" || selector.Sel.Name == "Format" || selector.Sel.Name == "Command") && !direct[selector] {
				escaped = fmt.Errorf("%s: SQL constructors must be called directly", name)
			}
			return true
		})
		if escaped != nil {
			return nil, escaped
		}
		out[name] = []SQLDeclaration{}
		for _, decl := range file.Decls {
			a := &sqlAnalyzer{packageConstantBindings: constantBindings, bindings: sqlBindingInventory(decl), fset: fset, current: decl, observed: map[ast.Expr]bool{}, directCalls: direct, packageConstants: packageConstants, constantImports: constantImports, types: types, scope: "github.com/JochiRaider/cartulary/" + path.Dir(name), variableTypes: map[string]sqlTypeRef{}, functionImports: functionImports, functions: functions[path.Dir(name)+"/"+file.Name.Name], collections: map[string]sqlCollection{}, calls: map[string]bool{}, imports: imports, constants: constants[path.Dir(name)+"/"+file.Name.Name], env: sqlEnv{}, seen: map[string]bool{}, results: []SQLExpression{}}
			for key, ref := range globals[a.scope] {
				a.variableTypes[key] = ref
			}
			switch d := decl.(type) {
			case *ast.FuncDecl:
				identity := d.Name.Name
				if d.Recv != nil {
					for _, field := range d.Recv.List {
						for _, name := range field.Names {
							a.variableTypes[sqlBinding(name)] = a.typeRef(field.Type)
						}
					}
					var receiver func(ast.Expr) string
					receiver = func(e ast.Expr) string {
						switch n := e.(type) {
						case *ast.StarExpr:
							return receiver(n.X)
						case *ast.IndexExpr:
							return receiver(n.X)
						case *ast.IndexListExpr:
							return receiver(n.X)
						case *ast.Ident:
							return n.Name
						}
						return ""
					}
					identity = receiver(d.Recv.List[0].Type) + "." + identity
				}
				if d.Type.Params != nil {
					for _, f := range d.Type.Params.List {
						for _, n := range f.Names {
							a.variableTypes[sqlBinding(n)] = a.typeRef(f.Type)
							a.env[sqlBinding(n)] = unknownSQL("parameter " + n.Name)
							if a.databaseType(f.Type) {
								a.env[sqlBinding(n)] = []sqlValue{{"\x00database", nil}}
							}

						}
					}
				}
				if d.Body != nil {
					// Named results have static identity before any branch writes
					// them, but no inferred raw SQL or preparation provenance.
					if d.Type.Results != nil {
						for _, field := range d.Type.Results.List {
							for _, name := range field.Names {
								a.variableTypes[sqlBinding(name)] = a.typeRef(field.Type)
								a.env[sqlBinding(name)] = unknownSQL("named result " + name.Name)
							}
						}
					}
					a.block(d.Body.List, false)
				}
				keys := []string{}
				for key := range a.env {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					a.emit(a.env[key], false)
				}
				a.discoverUnvisited(d)
				out[name] = append(out[name], SQLDeclaration{identity, a.results})
			case *ast.GenDecl:
				if d.Tok != token.CONST && d.Tok != token.VAR {
					continue
				}
				for _, spec := range d.Specs {
					v := spec.(*ast.ValueSpec)
					a.results = []SQLExpression{}
					a.seen = map[string]bool{}
					a.current = v
					ids := []string{}
					for _, n := range v.Names {
						ids = append(ids, n.Name)
					}
					for _, expr := range v.Values {
						a.observe(expr)
					}
					if len(v.Values) == 0 && d.Tok == token.CONST {
						for _, name := range v.Names {
							a.emit(a.value(name, map[string]bool{}), false)
						}
					}
					out[name] = append(out[name], SQLDeclaration{strings.Join(ids, ","), a.results})
				}
			}
		}
	}
	return out, nil
}

func (a *sqlAnalyzer) value(expr ast.Expr, visiting map[string]bool) []sqlValue {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			s, err := strconv.Unquote(e.Value)
			if err == nil {
				return []sqlValue{{s, nil}}
			}
		}
		return unknownSQL("non-string expression")
	case *ast.ParenExpr:
		return a.value(e.X, visiting)
	case *ast.Ident:
		if value, ok := a.env[sqlBinding(e)]; ok {
			if a.bindings[sqlBinding(e)].externallyMutable() && !preservedSQLValue(value) {
				return unknownSQL("SQL binding has a nested writer or address escape: " + e.Name)
			}
			return value
		}
		if definition, ok := a.constants[e.Name]; ok && !visiting[e.Name] && (e.Obj == nil || a.packageConstantBindings[e.Obj]) {
			next := map[string]bool{}
			for k, v := range visiting {
				next[k] = v
			}
			next[e.Name] = true
			return a.value(definition, next)
		}
		return unknownSQL("binding " + e.Name)
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			return joinSQLValues(a.value(e.X, visiting), a.value(e.Y, visiting))
		}
		return unknownSQL("binary expression")
	case *ast.TypeAssertExpr:
		if a.databaseType(e.Type) {
			return []sqlValue{{"\x00database", nil}}
		}
		return unknownSQL("unsupported type assertion")
	case *ast.CallExpr:
		if a.types.database(a.expressionType(e), map[string]bool{}) {
			return []sqlValue{{"\x00database", nil}}
		}
		if selector, ok := e.Fun.(*ast.SelectorExpr); ok {
			if id, ok := selector.X.(*ast.Ident); ok {
				if a.imported(id, "github.com/JochiRaider/cartulary/internal/platform/sqlshape") && (selector.Sel.Name == "Text" || selector.Sel.Name == "CommandText") {
					return []sqlValue{{closedSQLText, nil}} // sealed fragments are inspected at construction
				}
				if a.imported(id, "fmt") && selector.Sel.Name == "Sprintf" && len(e.Args) > 0 {
					return a.formatted(e.Args)
				}
			}
		}
		if id, ok := e.Fun.(*ast.Ident); ok {
			if fn := a.functions[id.Name]; fn != nil && fn.Body != nil && !a.calls[id.Name] && (id.Obj == nil || id.Obj.Kind == ast.Fun) {
				return a.helperValue(fn, e.Args, visiting)
			}
		}
		return unknownSQL("call " + sqlExprName(e.Fun))
	case *ast.SelectorExpr:
		if id, ok := e.X.(*ast.Ident); ok && id.Obj == nil {
			imported := a.imports[id.Name]
			if definition := a.packageConstants[imported][e.Sel.Name]; definition != nil {
				key := imported + "." + e.Sel.Name
				if visiting[key] {
					return unknownSQL("recursive package constant " + key)
				}
				next := map[string]bool{}
				for k, v := range visiting {
					next[k] = v
				}
				next[key] = true
				child := *a
				child.env = sqlEnv{}
				child.constants = a.packageConstants[imported]
				child.imports = a.constantImports[definition]
				return child.value(definition, next)
			}
		}
		if a.databaseField(e) {
			return []sqlValue{{"\x00database", nil}}
		}
		if value, ok := a.env[sqlExpressionBinding(e)]; ok {
			if root := sqlRootBinding(e); root != nil && a.bindings[sqlBinding(root)].externallyMutable() {
				return unknownSQL("mutable field state cannot supply inferred SQL")
			}
			return value
		}
		return unknownSQL("selector " + sqlExprName(e))
	default:
		return unknownSQL("unsupported expression")
	}
}

func sqlExprName(e ast.Expr) string {
	switch n := e.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.SelectorExpr:
		return sqlExprName(n.X) + "." + n.Sel.Name
	}
	return "expression"
}

func (a *sqlAnalyzer) formatted(args []ast.Expr) []sqlValue {
	formats := a.value(args[0], map[string]bool{})
	// Unsupported formatting must remain visible even when a SQL-producing
	// declaration has no execution call in the inspected package.
	failure := func(reason string) []sqlValue {
		for _, format := range formats {
			if sqlStart.MatchString(format.text) || sqlAnywhere.MatchString(format.text) || sqlNested.MatchString(format.text) {
				return []sqlValue{{"SELECT %s", []string{reason}}}
			}
		}
		return unknownSQL(reason)
	}
	out := []sqlValue{}
	for _, format := range formats {
		if len(format.unknown) > 0 {
			return failure("nonconstant format")
		}
		parts := []sqlValue{{"", nil}}
		argument := 1
		for i := 0; i < len(format.text); i++ {
			if format.text[i] != '%' {
				parts = joinSQLValues(parts, []sqlValue{{string(format.text[i]), nil}})
				continue
			}
			if i+1 < len(format.text) && format.text[i+1] == '%' {
				parts = joinSQLValues(parts, []sqlValue{{"%", nil}})
				i++
				continue
			}
			if i+1 == len(format.text) || format.text[i+1] != 's' {
				return failure("unsupported SQL format: only %s and %% are supported")
			}
			if argument >= len(args) {
				return failure("SQL format argument count mismatch")
			}
			parts = joinSQLValues(parts, a.value(args[argument], map[string]bool{}))
			argument++
			i++
		}
		if argument != len(args) {
			return failure("SQL format argument count mismatch")
		}
		out = mergeSQLValues(out, parts)
	}
	return out
}

func (a *sqlAnalyzer) emit(values []sqlValue, sink bool) {
	for _, value := range values {
		text := value.text
		if sink && text == "" && len(value.unknown) == 0 {
			a.results = append(a.results, a.result("empty", "", nil))
			continue
		}
		if text == closedSQLText {
			if sink {
				a.results = append(a.results, a.result("closed", "", nil))
			}
			continue
		}
		if !sqlStart.MatchString(text) {
			if sqlNested.MatchString(text) {
				text = "SELECT " + text
			} else if len(value.unknown) > 0 && (sink || sqlAnywhere.MatchString(text)) {
				text = ""
			} else if !sink {
				continue
			}
		}
		key := text + "\x00" + strings.Join(value.unknown, "\x00")
		if sink {
			key = fmt.Sprint(a.current.Pos(), sink) + key
		}
		if a.seen[key] {
			continue
		}
		a.seen[key] = true
		unknown := append([]string{}, value.unknown...)
		kind := "statement"
		if sink {
			kind = "execution"
			if value.text == "%s" && len(unknown) == 1 && (strings.HasPrefix(unknown[0], "parameter ") || strings.HasPrefix(unknown[0], "callback parameter ")) {
				kind = "forward"
			}
		}
		a.results = append(a.results, a.result(kind, text, unknown))
	}
}

func sqlExpressionBinding(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return sqlBinding(e)
	case *ast.SelectorExpr:
		return sqlExpressionBinding(e.X) + "." + e.Sel.Name
	}
	return ""
}
func sqlBinding(id *ast.Ident) string {
	if id.Obj != nil {
		if node, ok := id.Obj.Decl.(ast.Node); ok {
			return fmt.Sprintf("%s@%d", id.Name, node.Pos())
		}
	}
	return id.Name
}
func sqlTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return sqlTypeName(e.X)
	case *ast.Ident:
		return e.Name
	case *ast.ParenExpr:
		return sqlTypeName(e.X)
	case *ast.IndexExpr:
		return sqlTypeName(e.X)
	case *ast.IndexListExpr:
		return sqlTypeName(e.X)
	}
	return ""
}
func (a *sqlAnalyzer) imported(id *ast.Ident, imported string) bool {
	return id.Obj == nil && a.imports[id.Name] == imported
}
func (a *sqlAnalyzer) template(expr ast.Expr, kind string, format bool, count int, variadic bool) {
	for _, value := range a.value(expr, map[string]bool{}) {
		reasons := append([]string{}, value.unknown...)
		if len(reasons) == 0 {
			chunks, err := sqltemplate.Parse(value.text, format)
			if err != nil {
				reasons = append(reasons, err.Error())
			} else if format && (variadic || len(chunks) != count+1) {
				reasons = append(reasons, "closed SQL template argument count is not exact")
			}
		}
		a.results = append(a.results, a.result(kind, value.text, reasons))
	}
}

// result binds every candidate to its discovery site. Policy, not the parser,
// decides whether a transparent forwarding candidate has authority.
func (a *sqlAnalyzer) result(kind, text string, reasons []string) SQLExpression {
	pos := a.fset.Position(a.current.Pos())
	disposition := "analyzed"
	if len(reasons) > 0 {
		disposition = "rejected"
	}
	if kind == "forward" {
		disposition = "transport"
	}
	return SQLExpression{Kind: kind, Text: text, Dynamic: len(reasons) > 0, Unresolved: append([]string{}, reasons...), Line: pos.Line, Column: pos.Column, Disposition: disposition}
}

// Discovery is exhaustive even when bounded evaluation stops at a return or
// cannot interpret a parent. Unvisited calls are inspected without guessed state.
func (a *sqlAnalyzer) discoverUnvisited(node ast.Node) {
	saved := a.state()
	a.env = sqlEnv{}
	a.collections = map[string]sqlCollection{}
	for key, ref := range a.variableTypes {
		ref.qualified = false
		a.variableTypes[key] = ref
	}
	defer func() { a.setState(saved) }()
	ast.Inspect(node, func(n ast.Node) bool {
		switch e := n.(type) {
		case *ast.AssignStmt:
			for _, expr := range e.Rhs {
				if !a.observed[expr] {
					a.observe(expr)
				}
			}
		case *ast.ValueSpec:
			for _, expr := range e.Values {
				if !a.observed[expr] {
					a.observe(expr)
				}
			}
		case *ast.ReturnStmt:
			for _, expr := range e.Results {
				if !a.observed[expr] {
					a.observe(expr)
				}
			}
		case *ast.CallExpr:
			if !a.observed[e] {
				a.observe(e)
			}
		case *ast.SelectorExpr:
			if !a.directCalls[e] && !a.observed[e] {
				a.observe(e)
			}
		case *ast.FuncLit:
			if !a.observed[e] {
				a.observe(e)
			}
		}
		return true
	})
}

func (a *sqlAnalyzer) observe(expr ast.Expr) {
	if expr == nil {
		return
	}
	previous := a.current
	a.current = expr
	a.observed[expr] = true
	defer func() { a.current = previous }()
	if selector, ok := expr.(*ast.SelectorExpr); ok && a.rejectedNativeValue(selector) {
		a.results = append(a.results, a.result("execution", "", []string{"raw native protocol construction or access is not qualified"}))
		return
	}
	if !a.directCalls[unparenSQL(expr)] {
		if boundary, known := a.callableBoundary(expr); known && boundary.kind != "none" {
			a.results = append(a.results, a.result("execution", "", []string{"SQL execution function values are unsupported; use a direct call or narrow interface port"}))
		}
	}
	switch e := expr.(type) {
	case *ast.CallExpr:
		fun := unparenSQL(e.Fun)
		if _, selector := fun.(*ast.SelectorExpr); !selector {
			if boundary, known := a.callableBoundary(fun); known && boundary.kind != "none" {
				if id, declared := fun.(*ast.Ident); !declared || id.Obj != nil && id.Obj.Kind != ast.Fun {
					a.results = append(a.results, a.result("execution", "", []string{"SQL execution function values are unsupported; use a direct call or narrow interface port"}))
				}
			}
			a.observe(fun)
		}
		if selector, ok := fun.(*ast.SelectorExpr); ok && a.rejectedNativeValue(selector) {
			a.observe(selector)
		}
		if literal, ok := fun.(*ast.FuncLit); ok {
			a.observe(literal)
		}
		if selector, ok := fun.(*ast.SelectorExpr); ok {
			if id, ok := selector.X.(*ast.Ident); ok && a.imported(id, "fmt") && selector.Sel.Name == "Errorf" {
				// Diagnostic templates are data, while calls in their arguments can
				// still execute SQL and must be visited.
				for _, arg := range e.Args {
					if _, literal := arg.(*ast.BasicLit); !literal {
						a.observe(arg)
					}
				}
				return
			}
		}
		if selector, ok := fun.(*ast.SelectorExpr); ok {
			if id, ok := selector.X.(*ast.Ident); ok && a.imported(id, "github.com/JochiRaider/cartulary/internal/platform/sqlshape") {
				if (selector.Sel.Name == "Literal" || selector.Sel.Name == "Format" || selector.Sel.Name == "Command") && len(e.Args) > 0 {
					kind := "fragment"
					if selector.Sel.Name == "Command" {
						kind = "command"
					}
					a.template(e.Args[0], kind, selector.Sel.Name != "Literal", len(e.Args)-1, e.Ellipsis.IsValid())
					for _, arg := range e.Args[1:] {
						a.observe(arg)
					}
					return
				}
			}
			if nested, ok := selector.X.(*ast.CallExpr); ok {
				a.observe(nested)
			}
			if boundary, known := a.callableBoundary(selector); known {
				if a.typeExpression(selector.X) {
					boundary.kind = "reject"
				}
				switch boundary.kind {
				case "retire":
					if a.connectionRetirement(e) {
						a.results = append(a.results, a.result("closed", "", nil))
					} else {
						a.results = append(a.results, a.result("execution", "", []string{"unsupported native SQL API " + selector.Sel.Name}))
					}
				case "reject":
					a.results = append(a.results, a.result("execution", "", []string{"unsupported native SQL API " + selector.Sel.Name}))
				case "sql", "copy":
					if selector.Sel.Name == "Prepare" && boundary.argument == 2 && len(e.Args) > 1 {
						for _, name := range a.value(e.Args[1], map[string]bool{}) {
							if len(name.unknown) > 0 || name.text != "" && !sqlPreparedName.MatchString(name.text) {
								a.results = append(a.results, a.result("execution", "", []string{"prepared names must be finite identifiers, never SQL text"}))
							}
						}
					}
					if boundary.argument >= len(e.Args) {
						a.results = append(a.results, a.result("execution", "", []string{"native SQL argument missing"}))
					} else if boundary.kind == "copy" {
						a.copyFrom(e.Args[boundary.argument])
					} else {
						a.emit(a.value(e.Args[boundary.argument], map[string]bool{}), true)
					}
				case "prepared":
					if !a.expressionType(selector.X).qualified {
						a.results = append(a.results, a.result("execution", "", []string{"prepared handle lacks qualified preparation provenance"}))
					} else {
						a.results = append(a.results, a.result("closed", "", nil))
					}
				case "batch":
					if boundary.argument >= len(e.Args) || !a.expressionType(e.Args[boundary.argument]).qualified {
						a.results = append(a.results, a.result("execution", "", []string{"batch lacks local construction provenance"}))
					} else {
						a.results = append(a.results, a.result("closed", "", nil))
					}
				}
			} else if potentialSQLBoundary(selector.Sel.Name) && !a.knownFunction(selector) {
				a.results = append(a.results, a.result("execution", "", []string{"unresolved SQL boundary receiver for " + selector.Sel.Name}))
			}

		}
		if selector, ok := fun.(*ast.SelectorExpr); ok {
			if id, ok := selector.X.(*ast.Ident); ok && a.imported(id, "fmt") && selector.Sel.Name == "Sprintf" {
				a.emit(a.value(e, map[string]bool{}), false)
				return
			}
		}
		batchConsumer := false
		if selector, ok := fun.(*ast.SelectorExpr); ok {
			boundary, known := a.callableBoundary(selector)
			batchConsumer = known && boundary.kind == "batch"
		}
		for _, arg := range e.Args {
			if a.types.isBatch(a.expressionType(arg)) && !batchConsumer {
				a.results = append(a.results, a.result("execution", "", []string{"native SQL batch escapes qualified execution"}))
			}
			if pointer, ok := arg.(*ast.UnaryExpr); ok && pointer.Op == token.AND {
				if id, ok := pointer.X.(*ast.Ident); ok {
					for _, v := range a.env[sqlBinding(id)] {
						if sqlStart.MatchString(v.text) {
							a.env[sqlBinding(id)] = unknownSQL("escaped SQL binding " + id.Name)
							break
						}
					}
				}
			}
			a.observe(arg)
		}
	case *ast.SelectorExpr:
		if !a.directCalls[e] {
			if _, known := a.callableBoundary(e); !known && potentialSQLBoundary(e.Sel.Name) && !a.knownFunction(e) && !a.knownType(e) {
				a.results = append(a.results, a.result("execution", "", []string{"unresolved escaped SQL boundary " + e.Sel.Name}))
			}
		}
		a.observe(e.X)
	case *ast.CompositeLit:
		if a.types.rawProtocolType(a.typeRef(e.Type), map[string]bool{}) {
			a.results = append(a.results, a.result("execution", "", []string{"raw protocol message or handle construction is not qualified"}))
		}
		if a.types.nativeIdentity(a.typeRef(e.Type), map[string]bool{}) != "" && len(e.Elts) > 0 {
			a.results = append(a.results, a.result("execution", "", []string{"native SQL handles require zero-value or qualified factory construction"}))
		}
		for _, el := range e.Elts {
			if kv, ok := el.(*ast.KeyValueExpr); ok {
				a.observe(kv.Value)
			} else {
				a.observe(el)
			}
		}
	case *ast.FuncLit:
		if e.Body != nil {
			a.observeClosure(e)
		}
	default:
		a.emit(a.value(expr, map[string]bool{}), false)
		ast.Inspect(expr, func(node ast.Node) bool {
			if node == expr {
				return true
			}
			switch expression := node.(type) {
			case *ast.CallExpr, *ast.SelectorExpr:
				// Observe expression children while lexical refinements are
				// available, rather than after range/type-switch scope exits.
				a.observe(expression.(ast.Expr))
				return false
			case *ast.FuncLit:
				a.observe(expression)
				return false
			}
			return true
		})
	}
}

// block reports abrupt completion. Return summaries, break exits and continue
// exits are separate channels; only ordinary continuation joins into later code.
func (a *sqlAnalyzer) block(statements []ast.Stmt, scoped bool) bool {
	outer := a.state()
	breakStart, continueStart := len(a.breaks), len(a.continues)
	locals := map[string]bool{}
	defer func() {
		if scoped {
			for name := range locals {
				a.restoreLocal(name, outer)
			}
			current := a.state()
			for i := breakStart; i < len(a.breaks); i++ {
				a.setState(a.breaks[i])
				for name := range locals {
					a.restoreLocal(name, outer)
				}
				a.breaks[i] = a.state()
			}
			for i := continueStart; i < len(a.continues); i++ {
				a.setState(a.continues[i])
				for name := range locals {
					a.restoreLocal(name, outer)
				}
				a.continues[i] = a.state()
			}
			a.setState(current)
		}
	}()
	for _, statement := range statements {
		switch s := statement.(type) {
		case *ast.AssignStmt:
			a.assign(s.Lhs, s.Rhs, s.Tok, nil, s, locals)
		case *ast.DeclStmt:
			if gen, ok := s.Decl.(*ast.GenDecl); ok {
				for _, spec := range gen.Specs {
					if value, ok := spec.(*ast.ValueSpec); ok {
						left := make([]ast.Expr, len(value.Names))
						for i, name := range value.Names {
							left[i] = name
						}
						a.assign(left, value.Values, gen.Tok, value.Type, value, locals)
					}
				}
			}
		case *ast.ExprStmt:
			a.observe(s.X)
		case *ast.ReturnStmt:
			a.recordReturn(s)
			for _, expr := range s.Results {
				a.observe(expr)
			}
			return true
		case *ast.BlockStmt:
			if a.block(s.List, true) {
				return true
			}
		case *ast.IfStmt:
			if s.Init != nil {
				inner := *s
				inner.Init = nil
				if a.block([]ast.Stmt{s.Init, &inner}, true) {
					return true
				}
				continue
			}
			a.observe(s.Cond)
			before := a.state()
			yesReturn := a.block(s.Body.List, true)
			yes := a.state()
			a.setState(before)
			noReturn := false
			if s.Else != nil {
				noReturn = a.block([]ast.Stmt{s.Else}, true)
			}
			if yesReturn && noReturn {
				return true
			}
			if noReturn {
				a.setState(yes)
			} else if !yesReturn {
				a.setState(a.joinStates(yes, a.state()))
			}
		case *ast.SwitchStmt:
			if s.Init != nil {
				inner := *s
				inner.Init = nil
				if a.block([]ast.Stmt{s.Init, &inner}, true) {
					return true
				}
				continue
			}
			a.observe(s.Tag)
			if hasSQLFallthrough(s.Body) {
				a.invalidate(sqlAssignedBindings(s.Body), "SQL switch fallthrough is unsupported")
			}
			savedBreaks := a.breaks
			a.breaks = nil
			before := a.state()
			paths := []sqlState{}
			hasDefault := false
			for _, item := range s.Body.List {
				c := item.(*ast.CaseClause)
				a.setState(before)
				// Each case starts with an independent copy, including provenance.
				a.setState(a.state())
				for _, expression := range c.List {
					a.observe(expression)
				}
				for key, ref := range a.caseTypes[c] {
					a.variableTypes[key] = ref
				}
				if c.List == nil {
					hasDefault = true
				}
				if !a.block(c.Body, true) {
					paths = append(paths, a.state())
				}
			}
			if !hasDefault {
				paths = append(paths, before)
			}
			paths = append(paths, a.breaks...)
			a.breaks = savedBreaks
			a.setState(a.joinStates(paths...))
			if len(paths) == 0 {
				return true
			}

		case *ast.ForStmt:
			if s.Init != nil {
				inner := *s
				inner.Init = nil
				if a.block([]ast.Stmt{s.Init, &inner}, true) {
					return true
				}
				continue
			}
			a.forLoop(s)
		case *ast.RangeStmt:
			a.rangeLoop(s)
		case *ast.TypeSwitchStmt:
			if s.Init != nil {
				inner := *s
				inner.Init = nil
				if a.block([]ast.Stmt{s.Init, &inner}, true) {
					return true
				}
				continue
			}
			if assignment, ok := s.Assign.(*ast.AssignStmt); ok && len(assignment.Lhs) == 1 {
				if id, ok := assignment.Lhs[0].(*ast.Ident); ok {
					if a.caseTypes == nil {
						a.caseTypes = map[*ast.CaseClause]map[string]sqlTypeRef{}
					}
					for _, item := range s.Body.List {
						c := item.(*ast.CaseClause)
						if len(c.List) == 1 {
							a.caseTypes[c] = map[string]sqlTypeRef{sqlBinding(id): a.typeRef(c.List[0])}
						}
					}
				}
			}
			if s.Assign != nil {
				inner := *s
				inner.Assign = nil
				if a.block([]ast.Stmt{s.Assign, &inner}, true) {
					return true
				}
				continue
			}
			if a.block([]ast.Stmt{&ast.SwitchStmt{Body: s.Body}}, false) {
				return true
			}
		case *ast.SelectStmt, *ast.LabeledStmt:
			// Taint before discovering any boundary, including calls preceding a write.
			a.invalidate(sqlAssignedBindings(statement), "unsupported SQL control flow")
			ast.Inspect(statement, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok {
					a.observe(call)
				}
				return true
			})
		case *ast.BranchStmt:
			switch s.Tok {
			case token.BREAK:
				a.breaks = append(a.breaks, a.state())
			case token.CONTINUE:
				a.continues = append(a.continues, a.state())
			}
			return true
		case *ast.DeferStmt:
			a.observe(s.Call)
		case *ast.GoStmt:
			a.observe(s.Call)
		default:
			ast.Inspect(statement, func(node ast.Node) bool {
				if expression, ok := node.(ast.Expr); ok {
					a.observe(expression)
					return false
				}
				return true
			})
		}
	}
	return false
}

func (a *sqlAnalyzer) copyFrom(expr ast.Expr) {
	literal, ok := expr.(*ast.CompositeLit)
	if !ok {
		a.results = append(a.results, a.result("copy", "", []string{"CopyFrom requires a literal pgx.Identifier"}))
		return
	}
	selector, ok := literal.Type.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Identifier" {
		a.results = append(a.results, a.result("copy", "", []string{"unsupported CopyFrom relation"}))
		return
	}
	id, ok := selector.X.(*ast.Ident)
	if !ok || !a.imported(id, "github.com/jackc/pgx/v5") || len(literal.Elts) < 1 || len(literal.Elts) > 2 {
		a.results = append(a.results, a.result("copy", "", []string{"unsupported CopyFrom identifier"}))
		return
	}
	values := []sqlValue{{"", nil}}
	for i, part := range literal.Elts {
		next := append([]sqlValue(nil), a.value(part, map[string]bool{})...)
		for j := range next {
			next[j].text = `"` + strings.ReplaceAll(next[j].text, `"`, `""`) + `"`
			if i > 0 {
				next[j].text = "." + next[j].text
			}
		}
		values = joinSQLValues(values, next)
	}
	for _, value := range values {
		a.results = append(a.results, a.result("copy", value.text, value.unknown))
	}
}
