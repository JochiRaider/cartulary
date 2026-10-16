package reporting

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const reportingRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestReportingProductionImportBoundaries(t *testing.T) {
	// Owner facades and shared wire contracts constrain dependency direction;
	// private file placement is not an ownership boundary.
	allowedModuleImports := map[string]bool{
		reportingRepoImportPrefix + "internal/modules/entities/mentions/reportingprovider": true,
		reportingRepoImportPrefix + "internal/modules/graphprojection":                     true,
		reportingRepoImportPrefix + "internal/modules/graphprojection/postgresresult":      true,
		reportingRepoImportPrefix + "internal/modules/incidentbundles/artifactport":        true,
		reportingRepoImportPrefix + "internal/modules/incidents":                           true,
		reportingRepoImportPrefix + "internal/modules/incidents/admission":                 true,
		reportingRepoImportPrefix + "internal/modules/incidents/reportingprovider":         true,
		reportingRepoImportPrefix + "internal/modules/parties":                             true,
		reportingRepoImportPrefix + "internal/modules/records/reportingprovider":           true,
		reportingRepoImportPrefix + "internal/modules/recovery":                            true,
		reportingRepoImportPrefix + "internal/modules/reference_data":                      true,
		reportingRepoImportPrefix + "internal/modules/reportcomposition":                   true,
		reportingRepoImportPrefix + "internal/modules/reporting/canonicaljson":             true,
		reportingRepoImportPrefix + "internal/modules/reporting/diagramcontract":           true,
		reportingRepoImportPrefix + "internal/modules/reporting/exportprovider":            true,
		reportingRepoImportPrefix + "internal/modules/reporting/graphsourcecontract":       true,
		reportingRepoImportPrefix + "internal/modules/reporting/outputcontract":            true,
		reportingRepoImportPrefix + "internal/modules/revisions/sourceboundary":            true,
		reportingRepoImportPrefix + "internal/modules/timeline/reportingprovider":          true,
	}

	for _, fixture := range []struct {
		name, source string
		allowed      bool
	}{
		{"renamed facade alias", "package reporting\nimport renamed \"github.com/JochiRaider/cartulary/internal/modules/reportcomposition\"", true},
		{"extracted wire consumer", "package reporting\nimport \"github.com/JochiRaider/cartulary/internal/modules/reporting/outputcontract\"", true},
		{"source owner internals", "package reporting\nimport renamed \"github.com/JochiRaider/cartulary/internal/modules/entities/mentions/internal/storage\"", false},
		{"unadmitted source root", "package reporting\nimport \"github.com/JochiRaider/cartulary/internal/modules/records\"", false},
		{"raw quoted forbidden import", "package reporting\nimport `github.com/JochiRaider/cartulary/internal/modules/timeline`", false},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			imports := reportingProductionImports(t, "extracted.go", fixture.source)
			if len(imports) != 1 || allowedModuleImports[imports[0]] != fixture.allowed {
				t.Fatal("owner boundary failed to distinguish a facade from source internals")
			}
		})
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
		for _, importPath := range reportingProductionImports(t, fileName, nil) {
			if !strings.HasPrefix(importPath, reportingRepoImportPrefix+"internal/modules/") {
				continue
			}
			if !allowedModuleImports[importPath] {
				t.Fatalf("%s imports module %s outside a Reporting owner facade", fileName, importPath)
			}
		}
	}
}

func TestReportingProductionDoesNotReadOwnerTablesDirectly(t *testing.T) {
	forbiddenFragments := []string{
		"FROM reference_pack_", "JOIN reference_pack_",
		"FROM graph_projection_runs", "JOIN graph_projection_runs",
		"FROM report_compositions", "JOIN report_compositions",
		"FROM report_composition_versions", "JOIN report_composition_versions",
		"FROM report_composition_release_bindings", "JOIN report_composition_release_bindings",
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

func reportingProductionImports(t testing.TB, fileName string, source any) []string {
	t.Helper()

	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(fileName), source, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", fileName, err)
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Fatal(err)
		}
		imports = append(imports, importPath)
	}
	return imports
}
