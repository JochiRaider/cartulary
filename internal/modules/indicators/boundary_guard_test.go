package indicators

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

const indicatorsRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestIndicatorsProductionImportBoundaries(t *testing.T) {
	assertIndicatorsProductionImportBoundaries(t)
}

func assertIndicatorsProductionImportBoundaries(t testing.TB) {
	t.Helper()

	allowedSiblingImports := map[string]map[string]bool{
		indicatorsRepoImportPrefix + "internal/modules/indicators": {
			"create.go":   true,
			"decoding.go": true,
			"routes.go":   true,
		},
		indicatorsRepoImportPrefix + "internal/modules/imports/ownerfacade": {
			"import_create.go": true,
			"query.go":         true,
		},
		indicatorsRepoImportPrefix + "internal/modules/incidents/admission": {
			"routes.go": true,
		},
		indicatorsRepoImportPrefix + "internal/modules/incidentbundles/sourceport": {
			"incident_bundle_contribution.go": true,
			"portable_apply.go":               true,
			"portable_export.go":              true,
			"portable_model.go":               true,
			"portable_prepare.go":             true,
			"portable_validate.go":            true,
			"source_port.go":                  true,
		},
		indicatorsRepoImportPrefix + "internal/modules/incidentportability": {
			"portable_export.go":   true,
			"portable_prepare.go":  true,
			"portable_validate.go": true,
		},
		indicatorsRepoImportPrefix + "internal/modules/projections/providercontract": {
			"contribution.go": true,
		},
		indicatorsRepoImportPrefix + "internal/modules/records": {
			"application.go":         true,
			"child_coordination.go":  true,
			"create_service.go":      true,
			"observation_service.go": true,
			"routes.go":              true,
		},
		indicatorsRepoImportPrefix + "internal/modules/records/subtypepresence": {
			"incident_bundle_contribution.go": true,
			"subtype_presence.go":             true,
		},
		indicatorsRepoImportPrefix + "internal/modules/revisions": {
			"child_coordination.go":             true,
			"create_service.go":                 true,
			"lifecycle_service.go":              true,
			"observation_repository.go":         true,
			"observation_service.go":            true,
			"revision_append_port.go":           true,
			"revision_provider_contribution.go": true,
		},
		indicatorsRepoImportPrefix + "internal/modules/revisions/deleterestorecontract": {
			"provider.go": true,
		},
		indicatorsRepoImportPrefix + "internal/modules/revisions/rollbackcontract": {
			"children.go": true,
			"provider.go": true,
		},
	}

	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if filepath.ToSlash(path) == "testsupport" {
				return filepath.SkipDir
			}
			return nil
		}
		fileName := filepath.Base(path)
		if !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			return nil
		}
		for _, importPath := range indicatorsProductionImports(t, path) {
			if strings.HasPrefix(importPath, indicatorsRepoImportPrefix+"internal/app") {
				t.Fatalf("%s imports application assembly %s", filepath.ToSlash(path), importPath)
			}
			if strings.HasPrefix(importPath, indicatorsRepoImportPrefix+"internal/modules/indicators/") {
				continue
			}
			if !strings.HasPrefix(importPath, indicatorsRepoImportPrefix+"internal/modules/") {
				continue
			}
			allowedFiles, ok := allowedSiblingImports[importPath]
			if !ok {
				t.Fatalf("%s imports unapproved sibling module %s", filepath.ToSlash(path), importPath)
			}
			if !allowedFiles[fileName] {
				t.Fatalf("%s imports %s; allowed files are %v", filepath.ToSlash(path), importPath, indicatorsAllowedFileNames(allowedFiles))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk indicators package: %v", err)
	}
}

func TestIndicatorsDoNotUseEntitiesSourcePrefixes(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		fileName := filepath.Base(path)
		if !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(content), "entities.indicators.") {
			t.Fatalf("%s uses legacy entities indicator source prefix", filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk indicators package: %v", err)
	}
}

func TestIndicatorAdmissionIsTransportNeutral(t *testing.T) {
	for _, fileName := range []string{
		"application.go",
		"contracts.go",
		"persistence_scanners.go",
		"replay_codec.go",
		"value_serialization.go",
	} {
		for _, importPath := range indicatorsProductionImports(t, fileName) {
			switch importPath {
			case "net/http",
				indicatorsRepoImportPrefix + "internal/platform/httpapi",
				indicatorsRepoImportPrefix + "internal/platform/viewschema":
				t.Fatalf("%s imports transport/schema adapter %s", fileName, importPath)
			}
		}
	}
}

func TestIndicatorSourceSQLDoesNotUseEnvelopeMirrors(t *testing.T) {
	mirrorColumns := []string{
		"row_version", "created_at", "updated_at", "created_by_user_id",
		"updated_by_user_id", "deleted_at", "deleted_by_user_id",
	}
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			query, err := strconv.Unquote(literal.Value)
			if err != nil {
				return true
			}
			lower := strings.ToLower(query)
			for _, qualified := range []string{"i.", "indicator."} {
				for _, column := range mirrorColumns {
					if strings.Contains(lower, qualified+column) {
						t.Fatalf("%s reads removed Indicator envelope mirror %s%s", filepath.ToSlash(path), qualified, column)
					}
				}
			}
			for _, statement := range []string{"insert into indicators", "update indicators"} {
				start := strings.Index(lower, statement)
				if start < 0 {
					continue
				}
				writeClause := lower[start:]
				if end := strings.Index(writeClause, " where "); end >= 0 {
					writeClause = writeClause[:end]
				}
				for _, column := range mirrorColumns {
					if strings.Contains(writeClause, column) {
						t.Fatalf("%s writes removed Indicator envelope mirror %s", filepath.ToSlash(path), column)
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk Indicator production SQL: %v", err)
	}
}

func indicatorsProductionImports(t testing.TB, fileName string) []string {
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

func indicatorsAllowedFileNames(files map[string]bool) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	return names
}
