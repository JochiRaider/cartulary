package entities

import (
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
	allowedSiblingImports := map[string]map[string]bool{
		entitiesRepoImportPrefix + "internal/modules/collaboration": {
			"routes.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/entities/entitycontract": {
			"routes.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/entities/merge": {
			"http_helpers.go": true,
			"routes.go":       true,
		},
		entitiesRepoImportPrefix + "internal/modules/entities/mentions": {
			"routes.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/incidents": {
			"routes.go": true,
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
		if strings.Contains(string(body), "BuildBatchPlan(") {
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
	for _, fileName := range []string{"routes.go", "http_helpers.go"} {
		for _, importPath := range entitiesProductionImports(t, fileName) {
			if importPath == entitiesRepoImportPrefix+"internal/modules/entities/hostidentity" {
				t.Fatalf("%s imports hostidentity; root entities must stay route composition only", fileName)
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
	if !strings.Contains(content, "type timelineEffectsPort interface") {
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
	return names
}
