package main

import (
	"fmt"
	"go/ast"
	"go/constant"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ssa"
)

// A concrete session GET with ?limit=1 and at most one admitted string
// parameter. The sample is input data. Neither operation nor response is
// supplied to the evaluator.
type admittedProbe struct {
	owner, parameter, value, path string
	query                         url.Values
}

const samplePathValue = "cartulary.view.timeline.v2"

func singletonProbe(fields map[string]ast.Expr) *admittedProbe {
	if literalText(fields["Method"]) != "GET" || literalText(fields["Availability"]) != "base" || !admitsSession(fields["Security"]) {
		return nil
	}
	owner := literalText(fields["OwnerID"])
	template := literalText(fields["PathTemplate"])
	if owner == "" || !strings.HasPrefix(template, "/") || literalText(fields["Pattern"]) != "GET "+template {
		return nil
	}
	parameters, list := fields["PathParameters"].(*ast.CompositeLit)
	ident, named := fields["PathParameters"].(*ast.Ident)
	if fields["PathParameters"] == nil || (named && ident.Name == "nil") || (list && len(parameters.Elts) == 0) {
		if strings.ContainsAny(template, "{}") {
			return nil
		}
		return &admittedProbe{owner: owner, path: template}
	}
	if !list || len(parameters.Elts) != 1 {
		return nil
	}
	parameter, ok := parameters.Elts[0].(*ast.CompositeLit)
	if !ok {
		return nil
	}
	spec := literalFields(parameter)
	name := literalText(spec["Name"])
	pattern := literalText(spec["Pattern"])
	const value = samplePathValue
	minimum, minOK := literalBound(spec["MinLength"])
	maximum, maxOK := literalBound(spec["MaxLength"])
	match, err := regexp.MatchString(pattern, value)
	if expr := spec["Format"]; expr != nil {
		format, err := stringLiteral(expr)
		if err != nil || format != "" {
			return nil
		}
	}
	if name == "" || pattern == "" || err != nil || !match || !minOK || !maxOK || len(value) < minimum || maximum > 0 && len(value) > maximum {
		return nil
	}
	path := strings.Replace(template, "{"+name+"}", value, 1)
	if path == template || strings.ContainsAny(path, "{}") {
		return nil
	}
	return &admittedProbe{owner: owner, parameter: name, value: value, path: path}
}

// Resolve the concrete URL through the catalog's pattern set. A literal
// sibling can take precedence over an otherwise admissible parameter.
func resolveProbes(catalog declarationCatalog) (err error) {
	// ServeMux reports conflicting and malformed patterns by panic.
	defer func() {
		if failure := recover(); failure != nil {
			err = fmt.Errorf("resolve operation catalog patterns: %v", failure)
		}
	}()
	mux := http.NewServeMux()
	for _, op := range catalog {
		if strings.HasPrefix(op.pattern, "GET ") {
			mux.HandleFunc(op.pattern, func(http.ResponseWriter, *http.Request) {})
		}
	}
	for id, op := range catalog {
		if op.probe == nil {
			continue
		}
		req, err := http.NewRequest("GET", op.probe.path+"?limit=1", nil)
		if err == nil {
			if _, matched := mux.Handler(req); matched == op.pattern {
				op.probe.query = req.URL.Query()
				continue
			}
		}
		op.probe = nil
		catalog[id] = op
	}
	return nil
}

func admitsSession(expr ast.Expr) bool {
	security, ok := expr.(*ast.CompositeLit)
	if !ok {
		return false
	}
	session := false
	for _, elt := range security.Elts {
		alternative, ok := elt.(*ast.CompositeLit)
		if !ok || len(alternative.Elts) == 0 {
			return false
		}
		if len(alternative.Elts) == 1 && literalText(alternative.Elts[0]) == "sessionCookie" {
			session = true
		}
	}
	return session
}

func literalFields(literal *ast.CompositeLit) map[string]ast.Expr {
	fields := make(map[string]ast.Expr, len(literal.Elts))
	for _, elt := range literal.Elts {
		if pair, ok := elt.(*ast.KeyValueExpr); ok {
			if name, ok := pair.Key.(*ast.Ident); ok {
				fields[name.Name] = pair.Value
			}
		}
	}
	return fields
}

func literalText(expr ast.Expr) string {
	s, _ := stringLiteral(expr)
	return s
}

func literalBound(expr ast.Expr) (int, bool) {
	if expr == nil {
		return 0, true
	}
	lit, ok := expr.(*ast.BasicLit)
	if !ok {
		return 0, false
	}
	bound, err := strconv.Atoi(lit.Value)
	return bound, err == nil && bound >= 0
}

func stringConstant(v ssa.Value) string {
	if c, ok := v.(*ssa.Const); ok && c.Value != nil && c.Value.Kind() == constant.String {
		return constant.StringVal(c.Value)
	}
	return ""
}
