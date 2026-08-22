package entities

import (
	"encoding/json"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const entitiesRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestEntitiesProductionImportBoundaries(t *testing.T) {
	allowedSiblingImports := map[string]map[string]bool{
		entitiesRepoImportPrefix + "internal/modules/entities/hostidentity/deleterestore": {
			"revision_provider_contribution.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/entities/hostidentity/rollbackprovider": {
			"revision_provider_contribution.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/entities/merge": {
			"http_helpers.go": true,
			"routes.go":       true,
		},
		entitiesRepoImportPrefix + "internal/modules/entities/mentions": {
			"routes.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/entities/mentions/rollbackprovider": {
			"revision_provider_contribution.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/incidentbundles/sourceport": {
			"incident_bundle_source_port.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/incidentportability": {
			"incident_bundle_portability.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/incidents/admission": {
			"routes.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/records/subtypepresence": {
			"incident_bundle_subtype_presence.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/revisions": {
			"revision_provider_contribution.go": true,
		},
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read entities package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		for _, importPath := range entitiesProductionImports(t, fileName) {
			if !strings.HasPrefix(importPath, entitiesRepoImportPrefix+"internal/modules/") {
				continue
			}
			allowedFiles, ok := allowedSiblingImports[importPath]
			if !ok {
				t.Fatalf("%s imports unapproved sibling module %s", fileName, importPath)
			}
			if !allowedFiles[fileName] {
				t.Fatalf("%s imports %s; allowed files are %v", fileName, importPath, entitiesAllowedFileNames(allowedFiles))
			}
		}
	}

	t.Run("active tests have one exact owner selector", func(t *testing.T) {
		entitiesReconcileExactTestSelectors(t)
	})
}

func TestEntitiesDoNotRegisterWorkbookRowCreateRoutes(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean("routes.go"))
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}
	content := string(body)
	for _, route := range []string{
		"views/cartulary.view.hosts.v1/rows",
		"views/cartulary.view.identities.v1/rows",
		"views/cartulary.view.indicators.v1/rows",
	} {
		if strings.Contains(content, route) {
			t.Fatalf("entities routes.go still registers workbook row-create route %s", route)
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

func TestEntitiesRoutesUseCollaborationPublisher(t *testing.T) {
	imports := entitiesProductionImports(t, "routes.go")
	for _, importPath := range imports {
		if importPath == entitiesRepoImportPrefix+"internal/platform/ws" {
			t.Fatalf("routes.go imports platform/ws directly instead of collaboration publisher")
		}
	}
}

func TestEntitiesRootDoesNotImportHostIdentity(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read entities package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		for _, importPath := range entitiesProductionImports(t, fileName) {
			if importPath == entitiesRepoImportPrefix+"internal/modules/entities/hostidentity" {
				t.Fatalf("%s imports the hostidentity application surface; root imports must remain contribution-specific", fileName)
			}
		}
	}
}

func TestMergeDoesNotWriteMentionOrProjectionTablesDirectly(t *testing.T) {
	for _, path := range []string{
		filepath.Join("merge", "merge_store.go"),
		filepath.Join("merge", "ports.go"),
	} {
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
}

func TestMentionsUseCommandLevelTimelineEffectsPort(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean(filepath.Join("mentions", "ports.go")))
	if err != nil {
		t.Fatalf("read mentions/ports.go: %v", err)
	}
	content := string(body)
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

func entitiesAllowedFileNames(files map[string]bool) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type entitiesTestFamilyManifest struct {
	Rows []struct {
		Runner   string `json:"runner"`
		Selector struct {
			Package string   `json:"package"`
			Tests   []string `json:"tests"`
		} `json:"selector"`
	} `json:"rows"`
}

func entitiesReconcileExactTestSelectors(t testing.TB) {
	t.Helper()

	discovered := map[string]bool{}
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		active, err := build.Default.MatchFile(filepath.Dir(path), entry.Name())
		if err != nil {
			return err
		}
		if !active {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		packagePath := "./internal/modules/entities"
		if dir := filepath.ToSlash(filepath.Dir(path)); dir != "." {
			packagePath += "/" + dir
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv != nil || !entitiesIsTopLevelTestName(function.Name.Name) {
				continue
			}
			discovered[packagePath+"\x00"+function.Name.Name] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("discover active entities tests: %v", err)
	}

	manifestPath := filepath.Join("..", "..", "..", "tools", "test_families", "module.entities.json")
	body, err := os.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		t.Fatalf("read entities test-family manifest: %v", err)
	}
	var manifest entitiesTestFamilyManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatalf("decode entities test-family manifest: %v", err)
	}

	selected := map[string]int{}
	for _, row := range manifest.Rows {
		if row.Runner != "go" || (row.Selector.Package != "./internal/modules/entities" &&
			!strings.HasPrefix(row.Selector.Package, "./internal/modules/entities/")) {
			continue
		}
		for _, testName := range row.Selector.Tests {
			selected[row.Selector.Package+"\x00"+testName]++
		}
	}

	var anomalies []string
	for key := range discovered {
		switch selected[key] {
		case 0:
			anomalies = append(anomalies, "missing exact selector: "+entitiesDisplayTestKey(key))
		case 1:
		default:
			anomalies = append(anomalies, "duplicate exact selector: "+entitiesDisplayTestKey(key))
		}
	}
	for key, count := range selected {
		if !discovered[key] {
			anomalies = append(anomalies, "stale or inactive selector: "+entitiesDisplayTestKey(key))
		} else if count > 1 {
			anomalies = append(anomalies, "selector appears more than once: "+entitiesDisplayTestKey(key))
		}
	}
	sort.Strings(anomalies)
	if len(anomalies) > 0 {
		t.Fatalf("entities exact-selector reconciliation failed (discovered=%d selected=%d):\n%s",
			len(discovered), len(selected), strings.Join(anomalies, "\n"))
	}
	if len(discovered) != len(selected) {
		t.Fatalf("entities exact-selector count mismatch: discovered=%d selected=%d", len(discovered), len(selected))
	}
}

func entitiesIsTopLevelTestName(name string) bool {
	if !strings.HasPrefix(name, "Test") || len(name) == len("Test") {
		return false
	}
	next := name[len("Test")]
	return next < 'a' || next > 'z'
}

func entitiesDisplayTestKey(key string) string {
	return strings.Replace(key, "\x00", "::", 1)
}
