package entities

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/sourcecheck"
)

const entitiesRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestEntitiesProductionImportBoundaries(t *testing.T) {
	files, err := sourcecheck.ReadImports(os.DirFS("."))
	if err != nil {
		t.Fatal(err)
	}
	selected := 0
	for _, file := range files {
		if path.Dir(file.Path) != "." {
			continue
		}
		selected++
		if forbidden := entitiesUnapprovedProductionImports(file.Imports); len(forbidden) != 0 {
			t.Errorf("%s imports dependencies outside the Entities root boundary: %v", file.Path, forbidden)
		}
	}
	if selected == 0 {
		t.Fatal("no Entities root production files")
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
				source := "package entities\nimport \"" + entitiesRepoImportPrefix + "internal/" + test.dependency + "\"\n"
				imports, err := sourcecheck.ParseImports(file, []byte(source))
				if err != nil {
					t.Fatal(err)
				}
				got := entitiesUnapprovedProductionImports(imports)
				if (len(got) != 0) != test.forbidden {
					t.Fatalf("%s importing %s: forbidden=%v, want %v", file, test.dependency, got, test.forbidden)
				}
			}
		}
	})

	t.Run("merge uses consumer-owned cross-owner ports", func(t *testing.T) {
		selected := 0
		for _, file := range files {
			if path.Dir(file.Path) != "merge" {
				continue
			}
			selected++
			for _, imported := range file.Imports {
				if imported == entitiesRepoImportPrefix+"internal/modules/assessments" {
					t.Errorf("%s imports concrete Assessments", file.Path)
				}
			}
		}
		if selected == 0 {
			t.Fatal("no merge production files")
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
