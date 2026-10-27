package reporting

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/sourcecheck"
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
			imports, err := sourcecheck.ParseImports("extracted.go", []byte(fixture.source))
			if err != nil {
				t.Fatal(err)
			}
			if len(imports) != 1 || allowedModuleImports[imports[0]] != fixture.allowed {
				t.Fatal("owner boundary failed to distinguish a facade from source internals")
			}
		})
	}
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
		for _, imported := range file.Imports {
			if strings.HasPrefix(imported, reportingRepoImportPrefix+"internal/modules/") && !allowedModuleImports[imported] {
				t.Errorf("%s imports %s outside a Reporting owner facade", file.Path, imported)
			}
		}
	}
	if selected == 0 {
		t.Fatal("no Reporting production files")
	}
}
