package reporting

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const reportingRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestReportingProductionImportBoundaries(t *testing.T) {
	allowedSiblingImports := map[string]map[string]bool{
		reportingRepoImportPrefix + "internal/modules/artifacts/reportingprovider": {
			"export_materializer.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/entities/hostidentity/reportingprovider": {
			"export_materializer.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/entities/reportingprovider": {
			"export_materializer.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/evidence/reportingprovider": {
			"export_materializer.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/graphprojection": {
			"store.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/incidents": {
			"application_service.go": true,
			"routes.go":              true,
		},
		reportingRepoImportPrefix + "internal/modules/incidents/reportingprovider": {
			"export_materializer.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/links/reportingprovider": {
			"export_materializer.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/parties/reportingprovider": {
			"export_materializer.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/records/reportingprovider": {
			"export_materializer.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/reportcomposition": {
			"store.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/reporting/exportprovider": {
			"export_materializer.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/tasksdecisions/reportingprovider": {
			"export_materializer.go": true,
		},
		reportingRepoImportPrefix + "internal/modules/timeline/reportingprovider": {
			"export_materializer.go": true,
		},
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read reporting package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		for _, importPath := range reportingProductionImports(t, fileName) {
			if !strings.HasPrefix(importPath, reportingRepoImportPrefix+"internal/modules/") {
				continue
			}
			allowedFiles, ok := allowedSiblingImports[importPath]
			if !ok {
				t.Fatalf("%s imports sibling module %s without an explicit reporting boundary allowance", fileName, importPath)
			}
			if !allowedFiles[fileName] {
				t.Fatalf("%s imports %s; allowed files are %v", fileName, importPath, reportingAllowedFileNames(allowedFiles))
			}
		}
	}
}

func TestReportingProductionDoesNotReadOwnerTablesDirectly(t *testing.T) {
	forbiddenFragments := []string{
		"graph_projection_runs",
		"report_compositions",
		"report_composition_versions",
		"report_composition_release_bindings",
		"FROM incidents",
		"FROM change_sets",
		"FROM records",
		"JOIN records",
		"FROM timeline_events",
		"FROM host_grid_projection",
		"FROM identity_grid_projection",
		"FROM parties",
		"FROM evidence",
		"FROM task_request_grid_projection",
		"FROM decision_grid_projection",
		"FROM artifact_grid_projection",
		"FROM record_links",
		"FROM record_tags",
		"FROM entity_mentions",
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read reporting package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		data, err := os.ReadFile(fileName)
		if err != nil {
			t.Fatalf("read %s: %v", fileName, err)
		}
		for _, fragment := range forbiddenFragments {
			if strings.Contains(string(data), fragment) {
				t.Fatalf("%s contains direct owner table reference %q", fileName, fragment)
			}
		}
	}
}

func TestReportingProductionDoesNotReintroduceLegacyExportOrRenderPaths(t *testing.T) {
	forbiddenSymbols := []string{
		"collectWorkbookExportFieldsTx",
		"renderSlidevSource",
		"renderMermaidSource",
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read reporting package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		data, err := os.ReadFile(fileName)
		if err != nil {
			t.Fatalf("read %s: %v", fileName, err)
		}
		for _, symbol := range forbiddenSymbols {
			if strings.Contains(string(data), symbol) {
				t.Fatalf("%s reintroduces legacy reporting path %q", fileName, symbol)
			}
		}
	}
}

func reportingProductionImports(t testing.TB, fileName string) []string {
	t.Helper()

	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(fileName), nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", fileName, err)
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		imports = append(imports, strings.Trim(spec.Path.Value, `"`))
	}
	return imports
}

func reportingAllowedFileNames(allowed map[string]bool) []string {
	files := make([]string, 0, len(allowed))
	for fileName := range allowed {
		files = append(files, fileName)
	}
	return files
}
