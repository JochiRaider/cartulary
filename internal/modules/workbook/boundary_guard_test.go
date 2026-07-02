package workbook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkbookOwnsGenericRowCreateRoute(t *testing.T) {
	workbookRoutes, err := os.ReadFile(filepath.Clean("routes.go"))
	if err != nil {
		t.Fatalf("read workbook routes.go: %v", err)
	}
	if !strings.Contains(string(workbookRoutes), "POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows") {
		t.Fatalf("workbook routes.go must register the generic workbook row-create route")
	}

	entityRoutes, err := os.ReadFile(filepath.Clean(filepath.Join("..", "entities", "routes.go")))
	if err != nil {
		t.Fatalf("read entities routes.go: %v", err)
	}
	entityContent := string(entityRoutes)
	for _, route := range []string{
		"views/cartulary.view.hosts.v1/rows",
		"views/cartulary.view.identities.v1/rows",
		"views/cartulary.view.indicators.v1/rows",
	} {
		if strings.Contains(entityContent, route) {
			t.Fatalf("entities routes.go still registers workbook row-create route %s", route)
		}
	}
}

func TestWorkbookIndicatorCreateDispatchUsesIndicatorOwner(t *testing.T) {
	for _, fileName := range []string{"routes.go", "store.go", "telemetry.go"} {
		body, err := os.ReadFile(filepath.Clean(fileName))
		if err != nil {
			t.Fatalf("read %s: %v", fileName, err)
		}
		if !strings.Contains(string(body), "internal/modules/indicators") {
			t.Fatalf("%s must import indicators for indicator-owned workbook behavior", fileName)
		}
	}
}
