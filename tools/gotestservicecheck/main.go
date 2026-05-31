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
	"strings"
)

var (
	phaseUnitTestPattern      = regexp.MustCompile(`^TestPhase([0-9]+)[A-Za-z0-9_]*_U_([0-9]+)_`)
	phaseNamePattern          = regexp.MustCompile(`^phase(?:0|[1-9][0-9]*)$`)
	phaseManifestFilePattern  = regexp.MustCompile(`^(phase(?:0|[1-9][0-9]*))_test_map\.json$`)
	phaseHelperImportPattern  = regexp.MustCompile(`(?:^|/)internal/testutil/phase[0-9]+(?:store)?test$`)
	serviceHelperImportSuffix = regexp.MustCompile(`(?:^|/)internal/testutil/(?:pgtest|s3test)$`)
	startHelperPattern        = regexp.MustCompile(`^Start[A-Za-z0-9_]*$`)
)

const phaseTestMapSchemaID = "cartulary.phase_test_map.v2"
const phaseRegistrySchemaID = "cartulary.phase_registry.v1"

type manifestEntry struct {
	Coverage            string   `json:"coverage"`
	ExecutionDependency string   `json:"execution_dependency"`
	Runner              string   `json:"runner"`
	Symbol              string   `json:"symbol"`
	Symbols             []string `json:"symbols"`
}

type phaseManifest struct {
	SchemaID string          `json:"schema_id"`
	Phase    string          `json:"phase"`
	Unit     []manifestEntry `json:"unit"`
}

type phaseRegistry struct {
	SchemaID string               `json:"schema_id"`
	Phases   []phaseRegistryEntry `json:"phases"`
}

type phaseRegistryEntry struct {
	Phase        string `json:"phase"`
	Order        int    `json:"order"`
	Status       string `json:"status"`
	ManifestPath string `json:"manifest_path"`
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
	fmt.Fprintln(os.Stderr, "Only authoritative Go U-tests whose manifest entry declares a service-backed execution dependency may start pgtest/s3test or runtime helpers directly.")
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
	allowedTests, err := loadServiceBackedUnitTests(repoRoot)
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

func loadServiceBackedUnitTests(repoRoot string) (map[string]struct{}, error) {
	root := repoRoot
	if override := os.Getenv("CARTULARY_PHASE_MANIFEST_ROOT"); override != "" {
		root = override
	}
	entries, err := loadPhaseRegistry(root)
	if err != nil {
		return nil, err
	}

	allowed := make(map[string]struct{})
	seenPhases := make(map[string]string)
	for _, registryEntry := range entries {
		if registryEntry.Status != "active" {
			continue
		}
		raw, manifestPath, err := readTrustedRepoFile(root, registryEntry.ManifestPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", manifestPath, err)
		}

		var manifest phaseManifest
		if err := json.Unmarshal(raw, &manifest); err != nil {
			return nil, fmt.Errorf("decode %s: %w", manifestPath, err)
		}
		if err := validatePhaseManifestIdentity(manifestPath, manifest, seenPhases); err != nil {
			return nil, err
		}
		if manifest.Phase != registryEntry.Phase {
			return nil, fmt.Errorf("registry phase %s points at manifest declaring %s", registryEntry.Phase, manifest.Phase)
		}

		for _, entry := range manifest.Unit {
			if entry.Coverage != "authoritative" || entry.Runner != "go_test" || !isServiceBackedExecutionDependency(entry.ExecutionDependency) {
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

func isServiceBackedExecutionDependency(dependency string) bool {
	switch dependency {
	case "backend_store", "backend_integration", "backend_process":
		return true
	default:
		return false
	}
}

func loadPhaseRegistry(root string) ([]phaseRegistryEntry, error) {
	raw, registryPath, err := readTrustedRepoFile(root, "tools/phase_registry.json")
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", registryPath, err)
	}
	var registry phaseRegistry
	if err := json.Unmarshal(raw, &registry); err != nil {
		return nil, fmt.Errorf("decode %s: %w", registryPath, err)
	}
	if registry.SchemaID != phaseRegistrySchemaID {
		return nil, fmt.Errorf("%s must declare schema_id %s", registryPath, phaseRegistrySchemaID)
	}
	if len(registry.Phases) == 0 {
		return nil, fmt.Errorf("%s phases must be non-empty", registryPath)
	}
	seenPhases := make(map[string]struct{})
	seenOrders := make(map[int]string)
	for index, entry := range registry.Phases {
		label := fmt.Sprintf("%s phases[%d]", registryPath, index)
		if !phaseNamePattern.MatchString(entry.Phase) {
			return nil, fmt.Errorf("%s phase must match phase0 or phase[1-9][0-9]*", label)
		}
		switch entry.Status {
		case "active", "planned", "retired":
		default:
			return nil, fmt.Errorf("%s status must be active|planned|retired", label)
		}
		if entry.ManifestPath == "" {
			return nil, fmt.Errorf("%s manifest_path must be non-empty", label)
		}
		if filepath.IsAbs(entry.ManifestPath) {
			return nil, fmt.Errorf("%s manifest_path must be repo-relative", label)
		}
		clean := path.Clean(strings.ReplaceAll(entry.ManifestPath, "\\", "/"))
		if clean == "." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
			return nil, fmt.Errorf("%s manifest_path must not escape the repository root", label)
		}
		if clean != strings.ReplaceAll(entry.ManifestPath, "\\", "/") {
			return nil, fmt.Errorf("%s manifest_path must be normalized", label)
		}
		match := phaseManifestFilePattern.FindStringSubmatch(path.Base(clean))
		if match == nil {
			return nil, fmt.Errorf("%s manifest_path must end with phaseN_test_map.json", label)
		}
		if match[1] != entry.Phase {
			return nil, fmt.Errorf("%s manifest_path declares %s but phase is %s", label, match[1], entry.Phase)
		}
		if _, ok := seenPhases[entry.Phase]; ok {
			return nil, fmt.Errorf("%s declares duplicate phase %s", registryPath, entry.Phase)
		}
		if previous, ok := seenOrders[entry.Order]; ok {
			return nil, fmt.Errorf("%s declares duplicate order %d for %s and %s", registryPath, entry.Order, previous, entry.Phase)
		}
		seenPhases[entry.Phase] = struct{}{}
		seenOrders[entry.Order] = entry.Phase
	}
	sort.SliceStable(registry.Phases, func(i, j int) bool {
		if registry.Phases[i].Order != registry.Phases[j].Order {
			return registry.Phases[i].Order < registry.Phases[j].Order
		}
		return registry.Phases[i].Phase < registry.Phases[j].Phase
	})
	return registry.Phases, nil
}

func readTrustedRepoFile(root string, repoRelativePath string) ([]byte, string, error) {
	resolvedPath, err := trustedRepoFilePath(root, repoRelativePath)
	if err != nil {
		return nil, resolvedPath, err
	}
	raw, err := os.ReadFile(resolvedPath) // #nosec G304 -- resolvedPath is a normalized repo-relative path under the configured repository or manifest root.
	return raw, resolvedPath, err
}

func trustedRepoFilePath(root string, repoRelativePath string) (string, error) {
	clean, err := normalizedRepoRelativePath(repoRelativePath)
	if err != nil {
		return filepath.Join(root, filepath.FromSlash(repoRelativePath)), err
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve repository root %s: %w", root, err)
	}
	resolvedPath := filepath.Join(absRoot, filepath.FromSlash(clean))
	relative, err := filepath.Rel(absRoot, resolvedPath)
	if err != nil {
		return resolvedPath, fmt.Errorf("resolve repository-relative path %s: %w", repoRelativePath, err)
	}
	relative = filepath.ToSlash(relative)
	if relative == ".." || strings.HasPrefix(relative, "../") {
		return resolvedPath, fmt.Errorf("%s must not escape the repository root", repoRelativePath)
	}
	return resolvedPath, nil
}

func normalizedRepoRelativePath(repoRelativePath string) (string, error) {
	if strings.TrimSpace(repoRelativePath) == "" {
		return "", fmt.Errorf("repo-relative path must be non-empty")
	}
	if filepath.IsAbs(repoRelativePath) {
		return "", fmt.Errorf("%s must be repo-relative", repoRelativePath)
	}
	slashed := strings.ReplaceAll(repoRelativePath, "\\", "/")
	clean := path.Clean(slashed)
	if clean == "." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", fmt.Errorf("%s must not escape the repository root", repoRelativePath)
	}
	if clean != slashed {
		return "", fmt.Errorf("%s must be normalized", repoRelativePath)
	}
	return clean, nil
}

func validatePhaseManifestIdentity(manifestPath string, manifest phaseManifest, seenPhases map[string]string) error {
	if manifest.SchemaID != phaseTestMapSchemaID {
		return fmt.Errorf("manifest %s must declare schema_id %s", manifestPath, phaseTestMapSchemaID)
	}
	if manifest.Phase == "" {
		return fmt.Errorf("manifest %s must declare phase", manifestPath)
	}
	if !phaseNamePattern.MatchString(manifest.Phase) {
		return fmt.Errorf("invalid phase name %s; expected phase0 or phase[1-9][0-9]*", manifest.Phase)
	}
	match := phaseManifestFilePattern.FindStringSubmatch(filepath.Base(manifestPath))
	if match == nil {
		return fmt.Errorf("phase test map filename %s must match phase0_test_map.json or phase[1-9][0-9]*_test_map.json", filepath.Base(manifestPath))
	}
	if manifest.Phase != match[1] {
		return fmt.Errorf("manifest %s declares phase %s but filename declares %s", manifestPath, manifest.Phase, match[1])
	}
	if previous, ok := seenPhases[manifest.Phase]; ok {
		return fmt.Errorf("duplicate phase %s declared by %s and %s", manifest.Phase, previous, manifestPath)
	}
	seenPhases[manifest.Phase] = manifestPath
	return nil
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
