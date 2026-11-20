package sourcecheck

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"regexp"
	"sort"
	"strings"
)

// typesBinder resolves identifiers with go/types over one AnalyzeSQL batch.
// Each package (directory and package clause) is checked on its own; imports
// are never loaded and type errors are tolerated, because only declaration
// identity is needed. A binding is local: package-level, universe and package
// name objects are unbound, so package-level names always resolve through the
// analyzer's package maps, whichever file declares them.
type typesBinder struct {
	info       *types.Info
	fileScopes map[*types.Scope]bool
	decls      map[types.Object]ast.Node
	// go/types skips some operands in error paths (for example the variables of
	// a range assignment over an unresolved operand). Such references are
	// recovered by lexical lookup at their position.
	recovered map[*ast.Ident]types.Object
	// A type switch symbol (x in x := y.(type)) declares no object; each clause
	// declares its own implicit one. All of them bind to the switch assignment.
	switches map[*ast.Ident]*ast.AssignStmt
}

func newTypesBinder(fset *token.FileSet, files map[string]*ast.File) *typesBinder {
	b := &typesBinder{
		info: &types.Info{
			Defs:      map[*ast.Ident]types.Object{},
			Uses:      map[*ast.Ident]types.Object{},
			Implicits: map[ast.Node]types.Object{},
			Scopes:    map[ast.Node]*types.Scope{},
		},
		fileScopes: map[*types.Scope]bool{},
		decls:      map[types.Object]ast.Node{},
		recovered:  map[*ast.Ident]types.Object{},
		switches:   map[*ast.Ident]*ast.AssignStmt{},
	}
	declaring := map[*ast.Ident]ast.Node{}
	clauses := map[*ast.CaseClause]*ast.AssignStmt{}
	packages := map[string][]*ast.File{}
	importer := unloadedImporter{names: map[string]string{}}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		file := files[name]
		key := path.Dir(name) + "/" + file.Name.Name
		packages[key] = append(packages[key], file)
		b.collectDeclarations(file, declaring, clauses)
		if !strings.HasSuffix(file.Name.Name, "_test") {
			importPath := "github.com/JochiRaider/cartulary/" + path.Dir(name)
			if previous, seen := importer.names[importPath]; seen && previous != file.Name.Name {
				importer.names[importPath] = ""
			} else {
				importer.names[importPath] = file.Name.Name
			}
		}
	}
	keys := make([]string, 0, len(packages))
	for key := range packages {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		config := types.Config{Importer: importer, Error: func(error) {}, DisableUnusedImportCheck: true}
		pkg, _ := config.Check(key, fset, packages[key], b.info) // errors are expected: imports are never loaded
		for _, file := range packages[key] {
			b.checkSkippedLiterals(fset, pkg, file)
		}
	}
	for _, name := range names {
		if scope := b.info.Scopes[files[name]]; scope != nil {
			b.fileScopes[scope] = true
			b.recoverReferences(files[name], scope)
		}
	}
	for id, object := range b.info.Defs {
		if node := declaring[id]; object != nil && node != nil {
			b.decls[object] = node
		}
	}
	for node, object := range b.info.Implicits {
		if clause, ok := node.(*ast.CaseClause); ok && clauses[clause] != nil {
			b.decls[object] = clauses[clause]
		}
	}
	return b
}

// collectDeclarations maps each defining identifier to the node the analyzer
// treats as its declaration, as go/parser's object resolution did. Range
// variables use their range statement; receiver type parameters have none.
func (b *typesBinder) collectDeclarations(file *ast.File, declaring map[*ast.Ident]ast.Node, clauses map[*ast.CaseClause]*ast.AssignStmt) {
	fields := func(list *ast.FieldList) {
		if list != nil {
			for _, field := range list.List {
				for _, name := range field.Names {
					declaring[name] = field
				}
			}
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.ValueSpec:
			for _, name := range n.Names {
				declaring[name] = n
			}
		case *ast.TypeSpec:
			declaring[n.Name] = n
			fields(n.TypeParams)
		case *ast.FuncDecl:
			declaring[n.Name] = n
			fields(n.Recv)
		case *ast.FuncType:
			fields(n.TypeParams)
			fields(n.Params)
			fields(n.Results)
		case *ast.StructType:
			fields(n.Fields)
		case *ast.InterfaceType:
			fields(n.Methods)
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				for _, left := range n.Lhs {
					if id, ok := left.(*ast.Ident); ok {
						declaring[id] = n
					}
				}
			}
		case *ast.RangeStmt:
			if n.Tok == token.DEFINE {
				for _, left := range []ast.Expr{n.Key, n.Value} {
					if id, ok := left.(*ast.Ident); ok {
						declaring[id] = n
					}
				}
			}
		case *ast.TypeSwitchStmt:
			if assign, ok := n.Assign.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 {
				if id, ok := assign.Lhs[0].(*ast.Ident); ok {
					b.switches[id] = assign
					for _, clause := range n.Body.List {
						clauses[clause.(*ast.CaseClause)] = assign
					}
				}
			}
		case *ast.LabeledStmt:
			declaring[n.Label] = n
		}
		return true
	})
}

// checkSkippedLiterals checks function literals go/types never reached. An
// operand of an unloaded package is invalid, and go/types skips some operands
// next to one (for example a map literal value whose key is invalid), body and
// declarations included. Each skipped literal is checked where it appears.
func (b *typesBinder) checkSkippedLiterals(fset *token.FileSet, pkg *types.Package, file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		if literal, ok := node.(*ast.FuncLit); ok && b.info.Scopes[literal.Type] == nil {
			// Like the package check, this records what it can; an error ends it early.
			_ = types.CheckExpr(fset, pkg, literal.Pos(), literal, b.info)
		}
		return true
	})
}

// recoverReferences looks up identifiers go/types left unrecorded. Selector
// names and defining identifiers are never ordinary references.
func (b *typesBinder) recoverReferences(file *ast.File, scope *types.Scope) {
	selectors := map[*ast.Ident]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.SelectorExpr:
			selectors[n.Sel] = true
		case *ast.Ident:
			_, defined := b.info.Defs[n]
			if defined || selectors[n] || b.info.Uses[n] != nil || b.switches[n] != nil {
				return true
			}
			if inner := scope.Innermost(n.Pos()); inner != nil {
				if _, object := inner.LookupParent(n.Name, n.Pos()); object != nil {
					b.recovered[n] = object
				}
			}
		}
		return true
	})
}

func (b *typesBinder) bind(id *ast.Ident) binding {
	if assign := b.switches[id]; assign != nil {
		return binding{kind: variableBinding, decl: assign}
	}
	object := b.info.Uses[id]
	if object == nil {
		object = b.info.Defs[id]
	}
	if object == nil {
		object = b.recovered[id]
	}
	if object == nil || !b.local(object) {
		return binding{}
	}
	result := binding{kind: otherBinding, decl: b.decls[object]}
	switch object.(type) {
	case *types.Var:
		result.kind = variableBinding
	case *types.Const:
		result.kind = constantBinding
	case *types.TypeName:
		result.kind = typeBinding
	}
	return result
}

// local reports a declaration in a function or narrower scope. Functions,
// methods and struct fields are never local; blank declarations join no scope
// and are placed by position.
func (b *typesBinder) local(object types.Object) bool {
	if object.Pkg() == nil {
		return false
	}
	switch o := object.(type) {
	case *types.Func:
		return false
	case *types.Var:
		if o.IsField() {
			return false
		}
	}
	scope := object.Parent()
	if scope == nil && object.Name() == "_" {
		scope = object.Pkg().Scope().Innermost(object.Pos())
	}
	return scope != nil && scope != object.Pkg().Scope() && !b.fileScopes[scope]
}

// unloadedImporter supplies an empty package for every import without loading
// it. Returning it with an error makes go/types treat it as fake, so selectors
// on it raise no follow-up errors. It is named as an unaliased import is
// referenced, so package qualifiers resolve to package names: by the package
// clause of a repository package in the batch, otherwise by path convention.
type unloadedImporter struct {
	names map[string]string // import path -> package clause; "" when ambiguous
}

var errImportNotLoaded = errors.New("imports are not loaded for SQL analysis")
var majorVersionSuffix = regexp.MustCompile(`^v[0-9]+$`)

func (importer unloadedImporter) Import(importPath string) (*types.Package, error) {
	name := importer.names[importPath]
	if name == "" {
		name = path.Base(importPath)
		if majorVersionSuffix.MatchString(name) && path.Dir(importPath) != "." {
			name = path.Base(path.Dir(importPath))
		}
		if i := strings.Index(name, ".v"); i > 0 && majorVersionSuffix.MatchString(name[i+1:]) {
			name = name[:i] // gopkg.in/name.vN
		}
		name = strings.TrimSuffix(strings.TrimPrefix(name, "go-"), "-go")
	}
	return types.NewPackage(importPath, name), errImportNotLoaded
}
