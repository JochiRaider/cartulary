package indicators

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/sourcecheck"
)

const indicatorsRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestIndicatorsProductionImportBoundaries(t *testing.T) {
	assertIndicatorsProductionImportBoundaries(t)
	t.Run("package permissions survive file reorganization", func(t *testing.T) {
		for _, test := range []struct {
			path       string
			dependency string
			allowed    bool
		}{
			{"renamed.go", "records", true},
			{"split_part.go", "records", true},
			{"httpapi/renamed.go", "indicators", true},
			{"admission/renamed.go", "records", false},
			{"internal/providers/rollback/renamed.go", "records", false},
			{"internal/providers/rollback/renamed.go", "revisions/rollbackcontract", true},
		} {
			dependency := indicatorsRepoImportPrefix + "internal/modules/" + test.dependency
			if got := indicatorPackageImportAllowed(filepath.Dir(test.path), dependency); got != test.allowed {
				t.Errorf("%s importing %s: allowed=%v, want %v", test.path, dependency, got, test.allowed)
			}
		}
	})
}

func assertIndicatorsProductionImportBoundaries(t testing.TB) {
	t.Helper()

	files, err := sourcecheck.ReadImports(os.DirFS("."))
	if err != nil {
		t.Fatal(err)
	}
	selected := 0
	for _, file := range files {
		if strings.HasPrefix(file.Path, "testsupport/") {
			continue
		}
		selected++
		for _, imported := range file.Imports {
			if strings.HasPrefix(imported, indicatorsRepoImportPrefix+"internal/app") {
				t.Errorf("%s imports application assembly %s", file.Path, imported)
			}
			if strings.HasPrefix(imported, indicatorsRepoImportPrefix+"internal/modules/indicators/") {
				continue
			}
			if strings.HasPrefix(imported, indicatorsRepoImportPrefix+"internal/modules/") && !indicatorPackageImportAllowed(path.Dir(file.Path), imported) {
				t.Errorf("%s imports %s outside its owner-local capability boundary", file.Path, imported)
			}
		}
	}
	if selected == 0 {
		t.Fatal("no Indicator production files")
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
	files, err := sourcecheck.ReadImports(os.DirFS("admission"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no admission production files")
	}
	for _, file := range files {
		for _, imported := range file.Imports {
			if !indicatorAdmissionImportAllowed(imported) {
				t.Errorf("%s imports transport dependency %s", file.Path, imported)
			}
		}
	}
	for imported, allowed := range map[string]bool{"net/http": false, indicatorsRepoImportPrefix + "internal/platform/httpapi": false, indicatorsRepoImportPrefix + "internal/platform/viewschema": true, indicatorsRepoImportPrefix + "internal/platform/strictjson": true} {
		if indicatorAdmissionImportAllowed(imported) != allowed {
			t.Errorf("admission capability %s: want allowed=%v", imported, allowed)
		}
	}
}

func indicatorAdmissionImportAllowed(imported string) bool {
	return imported != "net/http" && imported != indicatorsRepoImportPrefix+"internal/platform/httpapi"
}

func indicatorPackageImportAllowed(directory, dependency string) bool {
	// Capability ownership is package-local; source filenames have no policy role.
	allowed := map[string][]string{
		".":                                 {"collaboration", "collaboration/protocol", "imports/ownerfacade", "incidentbundles/sourceport", "records", "records/subtypepresence", "reference_data", "revisions", "revisions/historycontract"},
		"admission":                         {"indicators"},
		"httpapi":                           {"incidents/admission", "indicators", "records"},
		"internal/identity":                 {"reference_data"},
		"internal/providers/deleterestore":  {"revisions/deleterestorecontract"},
		"internal/providers/incidentbundle": {"incidentbundles/sourceport", "incidentportability", "records/subtypepresence", "reference_data"},
		"internal/providers/rollback":       {"reference_data", "revisions/rollbackcontract"},
		"workbookprojection":                {"projections/providercontract"},
	}
	for _, capability := range allowed[directory] {
		if dependency == indicatorsRepoImportPrefix+"internal/modules/"+capability {
			return true
		}
	}
	return false
}
