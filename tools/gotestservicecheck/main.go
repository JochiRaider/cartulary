package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

var (
	unitTestPattern = regexp.MustCompile(`^TestPhase[0-9][A-Za-z0-9_]*_U_[0-9]_`)
	allowlisted     = map[string]struct{}{
		"TestPhase4_BindingMode_U_4_01":                    {},
		"TestPhase4_DuplicateMentionProvenance_U_4_02":     {},
		"TestPhase4_CreateFromMention_U_4_03":              {},
		"TestPhase4_DismissRestoreMentionLifecycle_U_4_04": {},
		"TestPhase4_ExactMatchPrecedence_U_4_05":           {},
		"TestPhase4_ExplicitEntityMerge_U_4_06":            {},
		"TestPhase4_IndicatorObservationSeparation_U_4_07": {},
	}
	disallowedSelectors = map[string]struct{}{
		"pgtest.Start":            {},
		"s3test.Start":            {},
		"phase1test.StartRuntime": {},
		"phase2test.StartRuntime": {},
		"phase2test.StartStore":   {},
		"phase4test.StartRuntime": {},
		"phase4test.StartServer":  {},
		"phase4test.StartStore":   {},
	}
)

type finding struct {
	Test     string
	File     string
	Line     int
	Selector string
}

func main() {
	findings, err := scan()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(findings) == 0 {
		return
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Test != findings[j].Test {
			return findings[i].Test < findings[j].Test
		}
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		if findings[i].Line != findings[j].Line {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].Selector < findings[j].Selector
	})

	fmt.Fprintln(os.Stderr, "Service-backed unit-test guard failed.")
	fmt.Fprintln(os.Stderr, "Only the allowlisted Phase 4 store-domain U-tests may start pgtest/s3test or runtime helpers directly; new service-backed U-tests must live under backend-store and be explicitly allowlisted.")
	for _, finding := range findings {
		fmt.Fprintf(os.Stderr, "  %s: %s:%d uses %s\n", finding.Test, finding.File, finding.Line, finding.Selector)
	}
	os.Exit(1)
}

func scan() ([]finding, error) {
	fset := token.NewFileSet()
	var findings []finding

	for _, root := range []string{"internal", filepath.Join("cmd", "server")} {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".go" || filepath.Base(path) == "" || filepath.Base(path)[0] == '.' {
				return nil
			}
			if len(path) < len("_test.go") || path[len(path)-len("_test.go"):] != "_test.go" {
				return nil
			}

			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}

			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil || fn.Name == nil {
					continue
				}
				if !unitTestPattern.MatchString(fn.Name.Name) {
					continue
				}
				if _, ok := allowlisted[fn.Name.Name]; ok {
					continue
				}

				ast.Inspect(fn.Body, func(node ast.Node) bool {
					call, ok := node.(*ast.CallExpr)
					if !ok {
						return true
					}
					selector := selectorString(call.Fun)
					if _, blocked := disallowedSelectors[selector]; !blocked {
						return true
					}
					position := fset.Position(call.Pos())
					findings = append(findings, finding{
						Test:     fn.Name.Name,
						File:     position.Filename,
						Line:     position.Line,
						Selector: selector,
					})
					return true
				})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return findings, nil
}

func selectorString(expr ast.Expr) string {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name + "." + selector.Sel.Name
}
