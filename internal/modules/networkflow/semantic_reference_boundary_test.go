package networkflow

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type semanticDeclaration struct {
	file, name, receiver string
	node                 ast.Node
	imports              map[string]string
	refs                 map[string]bool
	forbidden            bool
}

// Trace package-local declarations as well as imports. A function alias, local
// type alias, or helper cannot launder a route service or public API error into
// an application. Calls across an explicitly typed owner port remain ports;
// their adapter implementations are separate boundaries.
func semanticReferenceViolations(files map[string]*ast.File, roots map[string]bool) []string {
	declarations := map[string]*semanticDeclaration{}
	for path, file := range files {
		imports := map[string]string{}
		for _, imp := range file.Imports {
			value, _ := strconv.Unquote(imp.Path.Value)
			name := filepath.Base(value)
			if imp.Name != nil {
				name = imp.Name.Name
			}
			imports[name] = value
		}
		add := func(name, receiver string, node ast.Node) {
			declarations[name] = &semanticDeclaration{file: path, name: name, receiver: receiver, node: node, imports: imports, refs: map[string]bool{}}
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				name, receiver := d.Name.Name, ""
				if d.Recv != nil {
					typ := d.Recv.List[0].Type
					if ptr, ok := typ.(*ast.StarExpr); ok {
						typ = ptr.X
					}
					id, ok := typ.(*ast.Ident)
					if !ok {
						continue
					}
					name = id.Name + "." + name
					if len(d.Recv.List[0].Names) > 0 {
						receiver = d.Recv.List[0].Names[0].Name
					}
				}
				add(name, receiver, d)
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						add(spec.Name.Name, "", spec.Type)
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							add(name.Name, "", spec)
						}
					}
				}
			}
		}
	}
	forbiddenImport := func(path string) bool {
		return path == "net/http" || strings.HasPrefix(path, "net/http/") || strings.HasSuffix(path, "/httpapi") || strings.HasSuffix(path, "/httpauth")
	}
	for _, decl := range declarations {
		if decl.name == "routeService" {
			decl.forbidden = true
		}
		ast.Inspect(decl.node, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.Ident:
				if _, ok := declarations[node.Name]; ok && node.Name != decl.name {
					decl.refs[node.Name] = true
				}
			case *ast.SelectorExpr:
				if id, ok := node.X.(*ast.Ident); ok {
					if forbiddenImport(decl.imports[id.Name]) {
						decl.forbidden = true
					}
					if decl.receiver != "" && id.Name == decl.receiver {
						owner := strings.SplitN(decl.name, ".", 2)[0]
						target := owner + "." + node.Sel.Name
						if _, ok := declarations[target]; ok {
							decl.refs[target] = true
						}
					}
				}
			}
			return true
		})
	}
	for changed := true; changed; {
		changed = false
		for _, decl := range declarations {
			if decl.forbidden {
				continue
			}
			for ref := range decl.refs {
				if declarations[ref].forbidden {
					decl.forbidden = true
					changed = true
					break
				}
			}
		}
	}
	var violations []string
	for _, decl := range declarations {
		if roots[decl.file] && decl.forbidden {
			violations = append(violations, fmt.Sprintf("%s: %s reaches transport", decl.file, decl.name))
		}
	}
	sort.Strings(violations)
	return violations
}

func assertSemanticReferences(t *testing.T, rootPaths []string) {
	files := map[string]*ast.File{}
	roots := map[string]bool{}
	for _, path := range rootPaths {
		roots[path] = true
	}
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files[path] = file
	}
	if violations := semanticReferenceViolations(files, roots); len(violations) != 0 {
		t.Fatalf("local semantic dependency violations:\n%s", strings.Join(violations, "\n"))
	}
	for name, source := range map[string]string{
		"type alias":     `package fixture; type Sneaky = routeService; type Application struct{ transport *Sneaky }`,
		"error alias":    `package fixture; import h "github.com/JochiRaider/cartulary/internal/platform/httpapi"; type Failure = h.APIError; func operation() *Failure{return nil}`,
		"helper chain":   `package fixture; func operation(){helper()}; func helper(){transportHelper()}`,
		"function alias": `package fixture; var hidden = transportHelper; func operation(){hidden()}`,
		"local method":   `package fixture; type Application struct{}; func(a Application) operation(){a.hidden()}; func(a Application) hidden(){transportHelper()}`,
	} {
		t.Run(name, func(t *testing.T) {
			fixture, err := parser.ParseFile(token.NewFileSet(), "application.go", source, 0)
			if err != nil {
				t.Fatal(err)
			}
			transport, err := parser.ParseFile(token.NewFileSet(), "transport.go", `package fixture; import "net/http"; type routeService struct{request *http.Request}; func transportHelper()*http.Request{return nil}`, 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(semanticReferenceViolations(map[string]*ast.File{"application.go": fixture, "transport.go": transport}, map[string]bool{"application.go": true})) == 0 {
				t.Fatal("forbidden local coupling escaped guard")
			}
		})
	}
}
