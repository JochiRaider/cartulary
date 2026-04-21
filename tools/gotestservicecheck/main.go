package main

import (
	"encoding/json"
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
	unitTestPattern     = regexp.MustCompile(`^TestPhase[0-9][A-Za-z0-9_]*_U_[0-9]_`)
	disallowedSelectors = map[string]struct{}{
		"pgtest.Start":            {},
		"phase3test.StartStore":   {},
		"s3test.Start":            {},
		"phase1test.StartRuntime": {},
		"phase2test.StartRuntime": {},
		"phase2test.StartStore":   {},
		"phase4test.StartRuntime": {},
		"phase4test.StartServer":  {},
		"phase4test.StartStore":   {},
	}
)

type manifestEntry struct {
	Coverage            string   `json:"coverage"`
	ExecutionDependency string   `json:"execution_dependency"`
	Runner              string   `json:"runner"`
	Symbol              string   `json:"symbol"`
	Symbols             []string `json:"symbols"`
}

type phaseManifest struct {
	Unit []manifestEntry `json:"unit"`
}

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
	fmt.Fprintln(os.Stderr, "Only authoritative Go U-tests whose manifest entry declares execution_dependency=backend_store may start pgtest/s3test or runtime helpers directly.")
	for _, finding := range findings {
		fmt.Fprintf(os.Stderr, "  %s: %s:%d uses %s\n", finding.Test, finding.File, finding.Line, finding.Selector)
	}
	os.Exit(1)
}

func scan() ([]finding, error) {
	fset := token.NewFileSet()
	var findings []finding
	allowedTests, err := loadBackendStoreUnitTests()
	if err != nil {
		return nil, err
	}

	for _, root := range []string{"internal", filepath.Join("cmd", "server")} {
		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
				if _, ok := allowedTests[fn.Name.Name]; ok {
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

func loadBackendStoreUnitTests() (map[string]struct{}, error) {
	paths, err := filepath.Glob(filepath.Join("tools", "phase*_test_map.json"))
	if err != nil {
		return nil, fmt.Errorf("glob phase manifests: %w", err)
	}

	allowed := make(map[string]struct{})
	for _, manifestPath := range paths {
		raw, err := os.ReadFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", manifestPath, err)
		}

		var manifest phaseManifest
		if err := json.Unmarshal(raw, &manifest); err != nil {
			return nil, fmt.Errorf("decode %s: %w", manifestPath, err)
		}

		for _, entry := range manifest.Unit {
			if entry.Coverage != "authoritative" || entry.Runner != "go_test" || entry.ExecutionDependency != "backend_store" {
				continue
			}
			if entry.Symbol != "" {
				allowed[entry.Symbol] = struct{}{}
			}
			for _, symbol := range entry.Symbols {
				if symbol != "" {
					allowed[symbol] = struct{}{}
				}
			}
		}
	}
	return allowed, nil
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
