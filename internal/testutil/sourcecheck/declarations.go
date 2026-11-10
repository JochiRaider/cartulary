package sourcecheck

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// Declaration identifies a top-level Go binding independently of its filename.
// Source retains literal expressions for the shared SQL analyzer.
type Declaration struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

func ParseDeclarations(path string, source []byte) ([]Declaration, error) {
	positions := token.NewFileSet()
	file, err := parser.ParseFile(positions, path, source, 0)
	if err != nil {
		return nil, fmt.Errorf("parse declarations %s: %w", path, err)
	}
	result := []Declaration{}
	text := func(node ast.Node) string {
		return string(source[positions.Position(node.Pos()).Offset:positions.Position(node.End()).Offset])
	}
	for _, declaration := range file.Decls {
		switch decl := declaration.(type) {
		case *ast.FuncDecl:
			name := decl.Name.Name
			if decl.Recv != nil {
				receiver := decl.Recv.List[0].Type
				for {
					switch value := receiver.(type) {
					case *ast.StarExpr:
						receiver = value.X
					case *ast.IndexExpr:
						receiver = value.X
					case *ast.IndexListExpr:
						receiver = value.X
					case *ast.Ident:
						name = value.Name + "." + name
						receiver = nil
					default:
						return nil, fmt.Errorf("unsupported receiver in %s", path)
					}
					if receiver == nil {
						break
					}
				}
			}
			result = append(result, Declaration{Name: name, Source: text(decl)})
		case *ast.GenDecl:
			if decl.Tok != token.CONST && decl.Tok != token.VAR {
				continue
			}
			for _, spec := range decl.Specs {
				value := spec.(*ast.ValueSpec)
				names := ""
				for i, name := range value.Names {
					if i > 0 {
						names += ","
					}
					names += name.Name
				}
				result = append(result, Declaration{Name: names, Source: decl.Tok.String() + " " + text(value)})
			}
		}
	}
	return result, nil
}
