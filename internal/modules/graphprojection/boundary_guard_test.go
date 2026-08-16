package graphprojection

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const cartularyImportPrefix = "github.com/JochiRaider/cartulary/"

func TestGraphProjectionV2RootProductionBoundary_Unit(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Graph Projection root: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		for _, importPath := range productionImportsForFile(t, name) {
			if importPath == "net/http" ||
				strings.HasPrefix(importPath, cartularyImportPrefix+"internal/platform/postgres") ||
				strings.HasPrefix(importPath, cartularyImportPrefix+"internal/platform/httpapi") ||
				strings.HasPrefix(importPath, cartularyImportPrefix+"internal/platform/jobs") ||
				strings.HasPrefix(importPath, cartularyImportPrefix+"internal/modules/auth") {
				t.Fatalf("root production file %s imports coordination dependency %s", name, importPath)
			}
			if strings.HasPrefix(importPath, cartularyImportPrefix+"internal/modules/") &&
				!strings.HasPrefix(importPath, cartularyImportPrefix+"internal/modules/graphprojection") {
				t.Fatalf("root production file %s imports sibling module %s", name, importPath)
			}
		}
	}
}

func TestGraphProjectionV1RuntimePackagesAndImportsAreRemoved_Unit(t *testing.T) {
	root := repoRoot(t)
	for _, relative := range []string{
		"internal/modules/graphprojection/postgresbinding",
		"internal/modules/graphprojection/postgresstore",
		"internal/modules/graphprojection/fixturetest",
	} {
		entries, err := os.ReadDir(filepath.Join(root, relative))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatalf("inspect removed v1 package %s: %v", relative, err)
		}
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".go") {
				t.Fatalf("removed v1 package %s still contains Go source %s", relative, entry.Name())
			}
		}
	}

	for _, scanRoot := range []string{filepath.Join(root, "cmd"), filepath.Join(root, "internal")} {
		err := filepath.WalkDir(scanRoot, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			for _, importPath := range productionImportsForFile(t, path) {
				if strings.HasPrefix(importPath, cartularyImportPrefix+"internal/modules/graphprojection/postgresbinding") ||
					strings.HasPrefix(importPath, cartularyImportPrefix+"internal/modules/graphprojection/postgresstore") ||
					strings.HasPrefix(importPath, cartularyImportPrefix+"internal/modules/graphprojection/fixturetest") {
					t.Fatalf("production file %s imports removed v1 package %s", path, importPath)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan production root %s: %v", scanRoot, err)
		}
	}
}

func TestGraphProjectionCurrentRecoveryContributionIsV2Only_Unit(t *testing.T) {
	contribution := RecoveryStateContribution()
	if contribution.OwnerID != "module.graphprojection" {
		t.Fatalf("Recovery owner = %q", contribution.OwnerID)
	}
	wantTables := []string{
		"graph_projection_result_edges",
		"graph_projection_result_leases",
		"graph_projection_result_vertices",
		"graph_projection_results",
	}
	if len(contribution.Tables) != len(wantTables) {
		t.Fatalf("Recovery table count = %d, want %d", len(contribution.Tables), len(wantTables))
	}
	gotTables := make([]string, 0, len(contribution.Tables))
	for _, table := range contribution.Tables {
		gotTables = append(gotTables, table.TableName)
		if table.AlgorithmID == nil || *table.AlgorithmID != RestoreAlgorithmID {
			t.Fatalf("Recovery table %s algorithm = %v, want %s", table.TableName, table.AlgorithmID, RestoreAlgorithmID)
		}
	}
	slices.Sort(gotTables)
	if !slices.Equal(gotTables, wantTables) {
		t.Fatalf("Recovery tables = %v, want %v", gotTables, wantTables)
	}
}

func productionImportsForFile(t testing.TB, path string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(path), nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", path, err)
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		imports = append(imports, strings.Trim(spec.Path.Value, "\""))
	}
	return imports
}

func repoRoot(t testing.TB) string {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	return filepath.Clean(filepath.Join(workingDirectory, "..", "..", ".."))
}
