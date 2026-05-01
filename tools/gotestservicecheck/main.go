package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

var (
	phaseUnitTestPattern      = regexp.MustCompile(`^TestPhase([0-9]+)[A-Za-z0-9_]*_U_([0-9]+)_`)
	phaseHelperImportPattern  = regexp.MustCompile(`(?:^|/)internal/testutil/phase[0-9]+(?:store)?test$`)
	serviceHelperImportSuffix = regexp.MustCompile(`(?:^|/)internal/testutil/(?:pgtest|s3test)$`)
	startHelperPattern        = regexp.MustCompile(`^Start[A-Za-z0-9_]*$`)
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
	Test       string
	File       string
	Line       int
	Selector   string
	ImportPath string
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
		if findings[i].Selector != findings[j].Selector {
			return findings[i].Selector < findings[j].Selector
		}
		return findings[i].ImportPath < findings[j].ImportPath
	})

	fmt.Fprintln(os.Stderr, "Service-backed unit-test guard failed.")
	fmt.Fprintln(os.Stderr, "Only authoritative Go U-tests whose manifest entry declares execution_dependency=backend_store may start pgtest/s3test or runtime helpers directly.")
	for _, finding := range findings {
		fmt.Fprintf(os.Stderr, "  %s: %s:%d uses %s from %s\n", finding.Test, finding.File, finding.Line, finding.Selector, finding.ImportPath)
	}
	os.Exit(1)
}

func scan() ([]finding, error) {
	return scanRoot(".")
}

func scanRoot(repoRoot string) ([]finding, error) {
	fset := token.NewFileSet()
	var findings []finding
	allowedTests, err := loadBackendStoreUnitTests(repoRoot)
	if err != nil {
		return nil, err
	}

	for _, relativeRoot := range []string{"internal", filepath.Join("cmd", "server")} {
		root := filepath.Join(repoRoot, relativeRoot)
		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
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
			imports := importAliases(file)

			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil || fn.Name == nil {
					continue
				}
				if !isCanonicalPhaseUnitTest(fn.Name.Name) {
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
					selector, importPath, blocked := blockedServiceHelperCall(call.Fun, imports)
					if !blocked {
						return true
					}
					position := fset.Position(call.Pos())
					filePath := position.Filename
					if relativePath, err := filepath.Rel(repoRoot, filePath); err == nil {
						filePath = relativePath
					}
					findings = append(findings, finding{
						Test:       fn.Name.Name,
						File:       filepath.ToSlash(filePath),
						Line:       position.Line,
						Selector:   selector,
						ImportPath: importPath,
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

func loadBackendStoreUnitTests(repoRoot string) (map[string]struct{}, error) {
	root := repoRoot
	if override := os.Getenv("CARTULARY_PHASE_MANIFEST_ROOT"); override != "" {
		root = override
	}
	paths, err := filepath.Glob(filepath.Join(root, "tools", "phase*_test_map.json"))
	if err != nil {
		return nil, fmt.Errorf("glob phase manifests: %w", err)
	}
	sort.Strings(paths)

	allowed := make(map[string]struct{})
	for _, manifestPath := range paths {
		raw, err := os.ReadFile(manifestPath) // #nosec G304 -- phase manifest paths come from a repo-local glob or explicit manifest root override.
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

func isCanonicalPhaseUnitTest(name string) bool {
	match := phaseUnitTestPattern.FindStringSubmatch(name)
	if match == nil {
		return false
	}
	testPhase, err := strconv.Atoi(match[1])
	if err != nil {
		return false
	}
	unitPhase, err := strconv.Atoi(match[2])
	if err != nil {
		return false
	}
	return testPhase == unitPhase
}

func importAliases(file *ast.File) map[string]string {
	aliases := make(map[string]string)
	for _, importSpec := range file.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil || importPath == "" {
			continue
		}
		if importSpec.Name != nil {
			if importSpec.Name.Name == "." || importSpec.Name.Name == "_" {
				continue
			}
			aliases[importSpec.Name.Name] = importPath
			continue
		}
		aliases[path.Base(importPath)] = importPath
	}
	return aliases
}

func blockedServiceHelperCall(expr ast.Expr, imports map[string]string) (string, string, bool) {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", "", false
	}
	if !startHelperPattern.MatchString(selector.Sel.Name) {
		return "", "", false
	}
	importPath, ok := imports[ident.Name]
	if !ok || !isBlockedHelperImport(importPath) {
		return "", "", false
	}
	return ident.Name + "." + selector.Sel.Name, importPath, true
}

func isBlockedHelperImport(importPath string) bool {
	return serviceHelperImportSuffix.MatchString(importPath) || phaseHelperImportPattern.MatchString(importPath)
}
