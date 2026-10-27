package entities

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const entitiesRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestEntitiesProductionImportBoundaries(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list Entities root: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		if forbidden := entitiesUnapprovedProductionImports(entitiesProductionImports(t, file)); len(forbidden) != 0 {
			t.Errorf("%s imports dependencies outside the Entities root boundary: %v", file, forbidden)
		}
	}

	t.Run("package boundaries permit file reorganization", func(t *testing.T) {
		for _, file := range []string{"renamed.go", "split_part.go"} {
			for _, test := range []struct {
				dependency string
				forbidden  bool
			}{
				{"modules/entities/candidates", false},
				{"modules/entities/mutationadmission", false},
				{"modules/revisions/historycontract", false},
				{"modules/entities/hostidentity", true},
				{"modules/assessments", true},
				{"platform/ws", true},
				{"app/server", true},
			} {
				path := filepath.Join(t.TempDir(), file)
				source := "package entities\nimport \"" + entitiesRepoImportPrefix + "internal/" + test.dependency + "\"\n"
				if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
					t.Fatal(err)
				}
				got := entitiesUnapprovedProductionImports(entitiesProductionImports(t, path))
				if (len(got) != 0) != test.forbidden {
					t.Fatalf("%s importing %s: forbidden=%v, want %v", file, test.dependency, got, test.forbidden)
				}
			}
		}
	})

	t.Run("merge uses consumer-owned cross-owner ports", func(t *testing.T) {
		entries, err := filepath.Glob(filepath.Join("merge", "*.go"))
		if err != nil {
			t.Fatalf("list merge package: %v", err)
		}
		for _, path := range entries {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			for _, importPath := range entitiesProductionImports(t, path) {
				if importPath == entitiesRepoImportPrefix+"internal/modules/assessments" {
					t.Fatalf("%s imports concrete Assessments instead of the merge-owned port", path)
				}
			}
		}
	})
}

func entitiesUnapprovedProductionImports(imports []string) []string {
	// These are package capabilities; declaration and file placement are unconstrained.
	allowed := map[string]bool{
		entitiesRepoImportPrefix + "internal/modules/entities/candidates":                    true,
		entitiesRepoImportPrefix + "internal/modules/entities/hostidentity/deleterestore":    true,
		entitiesRepoImportPrefix + "internal/modules/entities/hostidentity/rollbackprovider": true,
		entitiesRepoImportPrefix + "internal/modules/entities/mentions":                      true,
		entitiesRepoImportPrefix + "internal/modules/entities/mentions/rollbackprovider":     true,
		entitiesRepoImportPrefix + "internal/modules/entities/merge":                         true,
		entitiesRepoImportPrefix + "internal/modules/entities/mutationadmission":             true,
		entitiesRepoImportPrefix + "internal/modules/incidentbundles/sourceport":             true,
		entitiesRepoImportPrefix + "internal/modules/incidentportability":                    true,
		entitiesRepoImportPrefix + "internal/modules/incidents/admission":                    true,
		entitiesRepoImportPrefix + "internal/modules/records/subtypepresence":                true,
		entitiesRepoImportPrefix + "internal/modules/revisions":                              true,
		entitiesRepoImportPrefix + "internal/modules/revisions/historycontract":              true,
		entitiesRepoImportPrefix + "internal/platform/authn":                                 true,
		entitiesRepoImportPrefix + "internal/platform/fieldnorm":                             true,
		entitiesRepoImportPrefix + "internal/platform/httpapi":                               true,
		entitiesRepoImportPrefix + "internal/platform/httpauth":                              true,
		entitiesRepoImportPrefix + "internal/platform/listquery":                             true,
		entitiesRepoImportPrefix + "internal/platform/pagination":                            true,
		entitiesRepoImportPrefix + "internal/platform/recoverystate":                         true,
	}
	var forbidden []string
	for _, dependency := range imports {
		if strings.HasPrefix(dependency, entitiesRepoImportPrefix+"internal/") && !allowed[dependency] {
			forbidden = append(forbidden, dependency)
		}
	}
	return forbidden
}

func TestEntitiesDoNotRegisterWorkbookRowCreateRoutes(t *testing.T) {
	content := entitiesProductionSource(t, ".")
	for _, route := range []string{
		"views/cartulary.view.hosts.v1/rows",
		"views/cartulary.view.identities.v1/rows",
		"views/cartulary.view.indicators.v1/rows",
	} {
		if strings.Contains(content, route) {
			t.Fatalf("Entities root registers Workbook row-create route %s", route)
		}
	}
}

func TestEntitiesDoNotBuildClipboardPastePlans(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read entities package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Clean(fileName))
		if err != nil {
			t.Fatalf("read %s: %v", fileName, err)
		}
		if strings.Contains(string(body), "BuildBatchPlan(") || strings.Contains(string(body), "BuildTabularRowPlanV1(") {
			t.Fatalf("%s builds clipboard paste plans inside entities", fileName)
		}
	}
}

func TestMergeDoesNotWriteMentionOrProjectionTablesDirectly(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("merge", "*.go"))
	if err != nil {
		t.Fatalf("list merge production files: %v", err)
	}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		content := string(body)
		for _, disallowed := range []string{
			"UPDATE entity_mentions",
			"INSERT INTO entity_mentions",
			"DELETE FROM entity_mentions",
			"DELETE FROM host_grid_projection",
			"DELETE FROM identity_grid_projection",
		} {
			if strings.Contains(content, disallowed) {
				t.Fatalf("%s contains direct cross-owner write %q", path, disallowed)
			}
		}
	}

	t.Run("carry-forward planning is read-only", func(t *testing.T) {
		for _, path := range paths {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			for _, declaration := range parsed.Decls {
				function, ok := declaration.(*ast.FuncDecl)
				if !ok || function.Body == nil || !strings.HasPrefix(function.Name.Name, "plan") {
					continue
				}
				ast.Inspect(function.Body, func(node ast.Node) bool {
					call, ok := node.(*ast.CallExpr)
					if !ok {
						return true
					}
					selector, ok := call.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					for _, mutationPrefix := range []string{"Sync", "Update", "Insert", "Delete", "Exec"} {
						if strings.HasPrefix(selector.Sel.Name, mutationPrefix) {
							t.Fatalf("%s calls mutation-capable method %s", function.Name.Name, selector.Sel.Name)
						}
					}
					return true
				})
			}
		}
	})
}

func TestMentionsUseCommandLevelTimelineEffectsPort(t *testing.T) {
	content := entitiesProductionSource(t, "mentions")
	if !strings.Contains(content, "type TimelineEffectsPort interface") {
		t.Fatalf("mentions ports must expose command-level timelineEffectsPort")
	}
	for _, disallowed := range []string{
		"type timelinePort interface",
		"LoadSourceRecordTx(context.Context, pgx.Tx, uuid.UUID)",
		"UpdateSourceRecordTx(context.Context, pgx.Tx",
		"BuildRecordRowTx(context.Context, pgx.Tx, uuid.UUID)",
		"RebuildTimelineProjectionTx(context.Context, pgx.Tx, uuid.UUID)",
		"VersionID(uuid.UUID, int64)",
	} {
		if strings.Contains(content, disallowed) {
			t.Fatalf("mentions timeline effects port still exposes timeline mechanics %q", disallowed)
		}
	}
}

func entitiesProductionImports(t testing.TB, fileName string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(fileName), nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", fileName, err)
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import for %s: %v", fileName, err)
		}
		imports = append(imports, importPath)
	}
	return imports
}

func entitiesProductionSource(t testing.TB, directory string) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(directory, "*.go"))
	if err != nil {
		t.Fatalf("list %s: %v", directory, err)
	}
	var source strings.Builder
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		source.Write(body)
		source.WriteByte('\n')
	}
	return source.String()
}
