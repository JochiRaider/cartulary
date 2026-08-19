package adapters

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

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
		"internal/modules/indicators/projectionprovider": {
			cartularyImportPrefix + "internal/modules/indicators/internal/providers/projection": {},
		},
		"internal/modules/tasksdecisions/projectionprovider": {
			cartularyImportPrefix + "internal/modules/tasksdecisions/internal/providers/projection": {},
		},
		"internal/app/projectionassembly": {
			cartularyImportPrefix + "internal/modules/entities/hostidentity/projectionprovider": {},
			cartularyImportPrefix + "internal/modules/evidence/projectionprovider":              {},
			cartularyImportPrefix + "internal/modules/indicators/projectionprovider":            {},
			cartularyImportPrefix + "internal/modules/parties/projectionprovider":               {},
			cartularyImportPrefix + "internal/modules/tasksdecisions/projectionprovider":        {},
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
