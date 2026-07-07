package graphprojection

import (
	"encoding/json"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const cartularyImportPrefix = "github.com/JochiRaider/cartulary/"

func TestGraphProjectionProductionImportBoundaries(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read graphprojection package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		for _, importPath := range productionImportsForFile(t, fileName) {
			if strings.HasPrefix(importPath, cartularyImportPrefix+"internal/modules/") &&
				!strings.HasPrefix(importPath, cartularyImportPrefix+"internal/modules/graphprojection") {
				t.Fatalf("%s imports sibling module %s; graphprojection v1 must consume explicit input bytes only", fileName, importPath)
			}
		}
	}
}

func TestWorkbookProjectionsAndPublicRoutesDoNotImportGraphProjection(t *testing.T) {
	root := repoRoot(t)
	roots := []string{
		filepath.Join(root, "cmd"),
		filepath.Join(root, "internal", "modules", "workbook"),
		filepath.Join(root, "internal", "modules", "projections"),
		filepath.Join(root, "internal", "platform", "httpapi"),
		filepath.Join(root, "internal", "platform", "ws"),
	}
	for _, scanRoot := range roots {
		if _, err := os.Stat(scanRoot); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("stat %s: %v", scanRoot, err)
		}
		err := filepath.WalkDir(scanRoot, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			for _, importPath := range productionImportsForFile(t, path) {
				if importPath == cartularyImportPrefix+"internal/modules/graphprojection" {
					t.Fatalf("%s imports graphprojection; activation slice must not expose public route or workbook projection coupling", path)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s: %v", scanRoot, err)
		}
	}
}

func TestNoPublicGraphProjectionRoutes(t *testing.T) {
	root := repoRoot(t)
	files := []string{
		filepath.Join(root, "contracts", "openapi", "cartulary.openapi.yaml"),
		filepath.Join(root, "contracts", "ws", "index.schema.json"),
	}
	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read contract %s: %v", file, err)
		}
		content := string(body)
		disallowedRoutes := []string{"/graph-projection", "/graph_projection", "/graphProjection", "/GraphProjection"}
		for _, route := range disallowedRoutes {
			if strings.Contains(content, route) {
				t.Fatalf("%s exposes graph projection public route term %q", file, route)
			}
		}
	}
}

func TestStoreSQLStaysWithinGraphProjectionTables(t *testing.T) {
	body, err := os.ReadFile("store.go")
	if err != nil {
		t.Fatalf("read store.go: %v", err)
	}
	content := string(body)
	forbiddenTables := []string{
		" incidents",
		" incident_",
		" workbook_",
		" timeline_",
		" record_",
		" evidence_",
		" assessment",
		" party_",
		" task_",
		" decision_",
	}
	for _, table := range forbiddenTables {
		if strings.Contains(content, table) {
			t.Fatalf("store.go references source/workbook table marker %q; graphprojection store must only use derived graph tables", table)
		}
	}
}

func TestOperationErrorsDoNotEchoSourceAuthoredValues(t *testing.T) {
	secret := "SECRET_SOURCE_VALUE_7812"
	input := []byte(`{
		"projection_schema_id":"graph_projection.v1",
		"graph_view_id":"not_a_graph_view_id",
		"source_snapshot_id":"snap_secret",
		"projection_config":{
			"graph_view_key":"secret_graph",
			"declared_source_entity_kinds":["host"],
			"entity_mappings":[{"mapping_rule_id":"map_host","source_entity_kind":"host","projected_vertex_kind":"host_vertex"}]
		},
		"source_entities":[{"source_entity_id":"host1","source_entity_kind":"host","properties":{"hostname":"` + secret + `"}}],
		"source_relationships":[],
		"requested_at":"2026-05-30T00:00:00Z",
		"requested_by":"fixture"
	}`)
	_, err := AdmitProjectionInput(input, AdmitOptions{})
	var opErr *OperationError
	if err == nil || !strings.Contains(err.Error(), "invalid_projection_request") {
		t.Fatalf("expected invalid projection request, got %v", err)
	}
	if !errors.As(err, &opErr) {
		t.Fatalf("expected operation error, got %T", err)
	}
	details, marshalErr := json.Marshal(opErr.Details)
	if marshalErr != nil {
		t.Fatalf("marshal operation error details: %v", marshalErr)
	}
	if strings.Contains(err.Error(), secret) || strings.Contains(string(details), secret) {
		t.Fatalf("operation error leaked source-authored value: error=%q details=%s", err.Error(), string(details))
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
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", "..", ".."))
}
