package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"strconv"
)

type declarationCatalog map[string]map[int][]string

// readCatalog reads each operation's declared error codes from the generated
// operation catalog source. Every element of its composite literals that names
// an OperationID contributes that operation's ErrorCodes values.
func readCatalog(path string) (declarationCatalog, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse operation catalog: %w", err)
	}
	catalog := declarationCatalog{}
	var failure error
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok || failure != nil {
			return failure == nil
		}
		operationID := ""
		var codes map[int][]string
		hasCodes := false
		for _, element := range literal.Elts {
			pair, ok := element.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := pair.Key.(*ast.Ident)
			if !ok {
				continue
			}
			switch key.Name {
			case "OperationID":
				value, err := stringLiteral(pair.Value)
				if err != nil {
					failure = fmt.Errorf("operation catalog OperationID: %w", err)
					return false
				}
				operationID = value
			case "ErrorCodes":
				hasCodes = true
				codes, err = errorCodes(pair.Value)
				if err != nil {
					failure = fmt.Errorf("operation catalog ErrorCodes: %w", err)
					return false
				}
			}
		}
		if operationID == "" {
			return true
		}
		if !hasCodes {
			failure = fmt.Errorf("operation %s declares no ErrorCodes", operationID)
			return false
		}
		if _, duplicate := catalog[operationID]; duplicate {
			failure = fmt.Errorf("operation %s is declared twice", operationID)
			return false
		}
		catalog[operationID] = codes
		return false
	})
	if failure != nil {
		return nil, failure
	}
	if len(catalog) == 0 {
		return nil, fmt.Errorf("operation catalog %s declares no operations", path)
	}
	return catalog, nil
}

func errorCodes(expression ast.Expr) (map[int][]string, error) {
	if ident, ok := expression.(*ast.Ident); ok && ident.Name == "nil" {
		return map[int][]string{}, nil
	}
	literal, ok := expression.(*ast.CompositeLit)
	if !ok {
		return nil, fmt.Errorf("expected a map literal")
	}
	codes := map[int][]string{}
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return nil, fmt.Errorf("expected status: codes pairs")
		}
		key, ok := pair.Key.(*ast.BasicLit)
		if !ok || key.Kind != token.INT {
			return nil, fmt.Errorf("expected a numeric HTTP status")
		}
		status, err := strconv.Atoi(key.Value)
		if err != nil || status < http.StatusBadRequest || status > 599 {
			return nil, fmt.Errorf("invalid error status %q", key.Value)
		}
		if _, exists := codes[status]; exists {
			return nil, fmt.Errorf("duplicate error status %d", status)
		}
		codes[status] = nil
		list, ok := pair.Value.(*ast.CompositeLit)
		if !ok {
			return nil, fmt.Errorf("expected a code list")
		}
		for _, item := range list.Elts {
			code, err := stringLiteral(item)
			if err != nil {
				return nil, err
			}
			codes[status] = append(codes[status], code)
		}
	}
	return codes, nil
}

func stringLiteral(expression ast.Expr) (string, error) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", fmt.Errorf("expected a string literal")
	}
	return strconv.Unquote(literal.Value)
}
