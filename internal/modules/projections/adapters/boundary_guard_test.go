package adapters

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLegacyProjectionMaintenanceHasNoProductionCallerOutsideProjections(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", "..", ".."))
	legacySelectors := map[string]struct{}{
		"Delete" + "IndicatorTx":      {},
		"Rebuild" + "Artifacts":       {},
		"Rebuild" + "ArtifactsTx":     {},
		"Rebuild" + "Assessments":     {},
		"Rebuild" + "AssessmentsTx":   {},
		"Rebuild" + "Evidence":        {},
		"Rebuild" + "EvidenceTx":      {},
		"Rebuild" + "Hosts":           {},
		"Rebuild" + "HostsTx":         {},
		"Rebuild" + "Identities":      {},
		"Rebuild" + "IdentitiesTx":    {},
		"Rebuild" + "Indicators":      {},
		"Rebuild" + "IndicatorsTx":    {},
		"Rebuild" + "Parties":         {},
		"Rebuild" + "Timeline":        {},
		"Rebuild" + "IncidentViewsTx": {},
	}
	for _, root := range []string{filepath.Join(repoRoot, "cmd"), filepath.Join(repoRoot, "internal")} {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			relative, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}
			if strings.HasPrefix(filepath.ToSlash(relative), "internal/modules/projections/") {
				return nil
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				return err
			}
			ast.Inspect(parsed, func(node ast.Node) bool {
				selector, ok := node.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if _, legacy := legacySelectors[selector.Sel.Name]; legacy {
					t.Errorf("%s retains production call to legacy projection maintenance selector %s", filepath.ToSlash(relative), selector.Sel.Name)
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("scan production maintenance callers under %s: %v", root, err)
		}
	}
}

func TestProjectionTestsupportUsesProductionAssembly(t *testing.T) {
	if err := projectionTestsupportAssemblyBoundary(filepath.Join("..", "testsupport")); err != nil {
		t.Fatal(err)
	}
	t.Run("delegation survives file reorganization", func(t *testing.T) {
		dir := t.TempDir()
		for i, name := range []string{"renamed.go", "split_part.go"} {
			source := fmt.Sprintf(`package testsupport
import assembly "github.com/JochiRaider/cartulary/internal/app/projectionassembly"
import adapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
func build%d() { assembly.Build(nil) }
type Port%d = adapters.MaintenanceRebuilder
`, i, i)
			if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		if err := projectionTestsupportAssemblyBoundary(dir); err != nil {
			t.Fatal(err)
		}
	})
	for _, forbidden := range []string{
		projectionAdaptersPath,
		cartularyImportPrefix + "internal/modules/timeline/projectionprovider",
		cartularyImportPrefix + "internal/modules/evidence/internal/providers/projection",
	} {
		t.Run("reject direct construction through "+forbidden, func(t *testing.T) {
			dir := t.TempDir()
			source := fmt.Sprintf("package testsupport\nimport forbidden %q\nvar construct = forbidden.New\n", forbidden)
			if err := os.WriteFile(filepath.Join(dir, "renamed.go"), []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := projectionTestsupportAssemblyBoundary(dir); err == nil || !strings.Contains(err.Error(), "reconstructs") {
				t.Fatalf("direct graph construction must be rejected: %v", err)
			}
		})
	}
	t.Run("requires production delegation", func(t *testing.T) {
		if err := projectionTestsupportAssemblyBoundary(t.TempDir()); err == nil {
			t.Fatal("missing production assembly delegation accepted")
		}
	})
}

func projectionTestsupportAssemblyBoundary(directory string) error {
	paths, err := filepath.Glob(filepath.Join(directory, "*.go"))
	if err != nil {
		return err
	}
	delegates := false
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		assemblyAlias := ""
		adapterAlias := ""
		for _, spec := range parsed.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if strings.Contains(importPath, "/projectionprovider") || strings.Contains(importPath, "/internal/providers/projection") {
				return fmt.Errorf("%s reconstructs the projection graph through %s", path, importPath)
			}
			alias := filepath.Base(importPath)
			if spec.Name != nil {
				alias = spec.Name.Name
			}
			if importPath == cartularyImportPrefix+"internal/app/projectionassembly" {
				assemblyAlias = alias
			}
			if importPath == projectionAdaptersPath {
				adapterAlias = alias
			}
		}
		var boundaryErr error
		ast.Inspect(parsed, func(node ast.Node) bool {
			if selector, ok := node.(*ast.SelectorExpr); ok && selector.Sel.Name == "New" {
				if qualifier, ok := selector.X.(*ast.Ident); ok && qualifier.Name == adapterAlias {
					boundaryErr = fmt.Errorf("%s reconstructs the projection graph through adapters.New", path)
				}
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Build" {
				return true
			}
			qualifier, ok := selector.X.(*ast.Ident)
			if ok && qualifier.Name == assemblyAlias {
				delegates = true
			}
			return true
		})
		if boundaryErr != nil {
			return boundaryErr
		}
	}
	if !delegates {
		return fmt.Errorf("testsupport does not delegate to production projection assembly")
	}
	return nil
}

const (
	cartularyImportPrefix      = "github.com/JochiRaider/cartulary/"
	projectionsImportPath      = cartularyImportPrefix + "internal/modules/projections"
	projectionAdaptersPath     = projectionsImportPath + "/adapters"
	providerContractImportPath = projectionsImportPath + "/providercontract"
)

func TestProductionProjectionImportsUseApprovedFacades(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", "..", ".."))
	scanRoots := []string{
		filepath.Join(repoRoot, "cmd"),
		filepath.Join(repoRoot, "internal", "app"),
		filepath.Join(repoRoot, "internal", "modules"),
	}
	projectionsDir := filepath.Clean(filepath.Join(repoRoot, "internal", "modules", "projections"))
	testSupportRoots := loadRuntimeExcludedSupportRoots(t, repoRoot)
	entries, err := os.ReadDir(projectionsDir)
	if err != nil {
		t.Fatalf("read Projections root: %v", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			t.Fatalf("Projections root retains Go file %q", entry.Name())
		}
	}

	for _, root := range scanRoots {
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("stat scan root %s: %v", root, err)
		}
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if filepath.Clean(path) == projectionsDir {
					return filepath.SkipDir
				}
				relative, relativeErr := filepath.Rel(repoRoot, path)
				if relativeErr != nil {
					return relativeErr
				}
				if _, excluded := testSupportRoots[filepath.ToSlash(relative)]; excluded {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			checkProjectionImports(t, repoRoot, path)
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
}

func checkProjectionImports(t *testing.T, repoRoot, filePath string) {
	t.Helper()

	relPath, err := filepath.Rel(repoRoot, filePath)
	if err != nil {
		t.Fatalf("rel path for %s: %v", filePath, err)
	}
	relPath = filepath.ToSlash(relPath)

	parsed, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", relPath, err)
	}

	for _, importSpec := range parsed.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import in %s: %v", relPath, err)
		}
		switch {
		case importPath == projectionsImportPath:
			t.Fatalf("%s imports the removed Projections root package", relPath)
		case importPath == projectionAdaptersPath:
			if filepath.ToSlash(filepath.Dir(relPath)) != "internal/app/projectionassembly" {
				t.Fatalf("%s imports projection adapters outside the projection assembly package", relPath)
			}
		case importPath == providerContractImportPath:
			continue
		case strings.HasPrefix(importPath, projectionsImportPath+"/"):
			t.Fatalf("%s imports projection internal package %s", relPath, importPath)
		case strings.HasPrefix(importPath, cartularyImportPrefix+"internal/modules/") &&
			(strings.Contains(importPath, "/projectionprovider") ||
				strings.Contains(importPath, "/internal/providers/projection")):
			if projectionProviderAssemblyImportAllowed(relPath, importPath) {
				continue
			}
			t.Fatalf("%s imports projection provider internal package %s", relPath, importPath)
		case strings.Contains(importPath, "/projections/testfixtures") ||
			strings.Contains(importPath, "/projections/test"):
			t.Fatalf("%s imports projection test fixture package %s", relPath, importPath)
		}
	}
}

func projectionProviderAssemblyImportAllowed(relPath string, importPath string) bool {
	allowed := map[string]map[string]struct{}{
		"internal/modules/artifacts": {
			cartularyImportPrefix + "internal/modules/artifacts/internal/providers/projection": {},
		},
		"internal/modules/assessments": {
			cartularyImportPrefix + "internal/modules/assessments/internal/providers/projection": {},
		},
		"internal/modules/evidence": {
			cartularyImportPrefix + "internal/modules/evidence/internal/providers/projection": {},
		},
		"internal/modules/indicators": {
			cartularyImportPrefix + "internal/modules/indicators/internal/providers/projection": {},
		},
		"internal/modules/parties": {
			cartularyImportPrefix + "internal/modules/parties/internal/providers/projection": {},
		},
		"internal/modules/tasksdecisions": {
			cartularyImportPrefix + "internal/modules/tasksdecisions/internal/providers/projection": {},
		},
		"internal/app/projectionassembly": {
			cartularyImportPrefix + "internal/modules/entities/hostidentity/projectionprovider": {},
			cartularyImportPrefix + "internal/modules/evidence":                                 {},
			cartularyImportPrefix + "internal/modules/indicators":                               {},
			cartularyImportPrefix + "internal/modules/parties":                                  {},
			cartularyImportPrefix + "internal/modules/tasksdecisions":                           {},
			cartularyImportPrefix + "internal/modules/timeline/projectionprovider":              {},
		},
	}
	_, ok := allowed[filepath.ToSlash(filepath.Dir(relPath))][importPath]
	return ok
}

func TestProjectionProviderAssemblyAllowlistMatchesFinalTopology(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		path       string
		importPath string
		want       bool
	}{
		"global projection assembly": {
			path:       "internal/app/projectionassembly/build.go",
			importPath: cartularyImportPrefix + "internal/modules/timeline/projectionprovider",
			want:       true,
		},
		"assessment root contribution": {
			path:       "internal/modules/assessments/projection_provider_contribution.go",
			importPath: cartularyImportPrefix + "internal/modules/assessments/internal/providers/projection",
			want:       true,
		},
		"evidence root contribution": {
			path:       "internal/modules/evidence/projection_contribution.go",
			importPath: cartularyImportPrefix + "internal/modules/evidence/internal/providers/projection",
			want:       true,
		},
		"projection assembly consumes Evidence root": {
			path:       "internal/app/projectionassembly/build.go",
			importPath: cartularyImportPrefix + "internal/modules/evidence",
			want:       true,
		},
		"projection assembly cannot bypass Evidence root": {
			path:       "internal/app/projectionassembly/build.go",
			importPath: cartularyImportPrefix + "internal/modules/evidence/internal/providers/projection",
			want:       false,
		},
		"party root contribution": {
			path:       "internal/modules/parties/provider_contributions.go",
			importPath: cartularyImportPrefix + "internal/modules/parties/internal/providers/projection",
			want:       true,
		},
		"indicator root contribution": {
			path:       "internal/modules/indicators/projection_provider_contribution.go",
			importPath: cartularyImportPrefix + "internal/modules/indicators/internal/providers/projection",
			want:       true,
		},
		"projection assembly uses Indicator root": {
			path:       "internal/app/projectionassembly/build.go",
			importPath: cartularyImportPrefix + "internal/modules/indicators/internal/providers/projection",
			want:       false,
		},
		"projection assembly consumes Indicator root": {
			path:       "internal/app/projectionassembly/build.go",
			importPath: cartularyImportPrefix + "internal/modules/indicators",
			want:       true,
		},
		"tasks decisions root contribution": {
			path:       "internal/modules/tasksdecisions/projection_contribution.go",
			importPath: cartularyImportPrefix + "internal/modules/tasksdecisions/internal/providers/projection",
			want:       true,
		},
		"projection assembly consumes Tasks Decisions root": {
			path:       "internal/app/projectionassembly/build.go",
			importPath: cartularyImportPrefix + "internal/modules/tasksdecisions",
			want:       true,
		},
		"projection assembly cannot bypass Tasks Decisions root": {
			path:       "internal/app/projectionassembly/build.go",
			importPath: cartularyImportPrefix + "internal/modules/tasksdecisions/internal/providers/projection",
			want:       false,
		},
		"projection assembly uses Party root": {
			path:       "internal/app/projectionassembly/build.go",
			importPath: cartularyImportPrefix + "internal/modules/parties/internal/providers/projection",
			want:       false,
		},
		"assessment application assembly uses root": {
			path:       "internal/app/assessmentassembly/projection_source.go",
			importPath: cartularyImportPrefix + "internal/modules/assessments/internal/providers/projection",
			want:       false,
		},
		"timeline is not a composition root": {
			path:       "internal/app/timelineassembly/assembly.go",
			importPath: cartularyImportPrefix + "internal/modules/timeline/projectionprovider",
			want:       false,
		},
		"projection assembly cannot bypass assessment seam": {
			path:       "internal/app/projectionassembly/build.go",
			importPath: cartularyImportPrefix + "internal/modules/assessments/internal/providers/projection",
			want:       false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := projectionProviderAssemblyImportAllowed(test.path, test.importPath); got != test.want {
				t.Fatalf("allowance for %s importing %s = %v, want %v", test.path, test.importPath, got, test.want)
			}
		})
	}
}

func loadRuntimeExcludedSupportRoots(t testing.TB, repoRoot string) map[string]struct{} {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(repoRoot, "tools", "test_support_inventory.json"))
	if err != nil {
		t.Fatalf("read test-support inventory: %v", err)
	}
	var inventory struct {
		Roots []struct {
			Path        string `json:"path"`
			RuntimeScan string `json:"runtime_scan"`
		} `json:"go_support_roots"`
	}
	if err := json.Unmarshal(body, &inventory); err != nil {
		t.Fatalf("decode test-support inventory: %v", err)
	}
	result := map[string]struct{}{}
	for _, root := range inventory.Roots {
		if root.RuntimeScan == "excluded" {
			result[root.Path] = struct{}{}
		}
	}
	return result
}
