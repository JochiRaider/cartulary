package imports

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportsIndicatorTargetUsesIndicatorOwner(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean("targets.go"))
	if err != nil {
		t.Fatalf("read targets.go: %v", err)
	}
	content := string(body)
	for _, required := range []string{
		`createFacadeIndicator    = "indicators.import_create"`,
		`Owner:           "indicators"`,
		`ViewSchemaID:    indicators.ViewSchemaID`,
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("targets.go missing indicator owner mapping %q", required)
		}
	}
	if strings.Contains(content, `createFacadeIndicator    = "entities.import_create"`) ||
		strings.Contains(content, `ViewSchemaID:    entities.IndicatorsViewSchemaID`) {
		t.Fatalf("targets.go still maps indicator imports through entities")
	}
}

func TestImportsIndicatorApplyUsesOwnerFacade(t *testing.T) {
	routes, err := os.ReadFile(filepath.Clean("routes.go"))
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}
	if strings.Contains(string(routes), "CreateIndicatorRow") {
		t.Fatalf("imports routes.go must not call indicator store row create directly")
	}

	ownerApply, err := os.ReadFile(filepath.Clean("owner_apply.go"))
	if err != nil {
		t.Fatalf("read owner_apply.go: %v", err)
	}
	content := string(ownerApply)
	if !strings.Contains(content, "stores.indicators.CreateImportRowTx") {
		t.Fatalf("owner_apply.go must dispatch indicator imports through indicators.CreateImportRowTx")
	}
}
