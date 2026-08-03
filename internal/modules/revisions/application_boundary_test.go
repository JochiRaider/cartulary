package revisions

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRevisionsApplicationAndHTTPBoundaries_Unit(t *testing.T) {
	for _, retiredRootFile := range []string{"routes.go", "delete_restore_api.go", "rollback_api.go"} {
		if _, err := os.Stat(retiredRootFile); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("transport file remains in Revisions application root: %s (%v)", retiredRootFile, err)
		}
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Revisions application root: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		contents, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		for _, forbiddenImport := range []string{
			`"net/http"`,
			`internal/platform/authn`,
			`internal/platform/httpapi`,
			`internal/platform/pagination`,
		} {
			if strings.Contains(string(contents), forbiddenImport) {
				t.Errorf("Revisions application root file %s imports transport/platform concern %s", entry.Name(), forbiddenImport)
			}
		}
	}

	for _, applicationFile := range []string{"application_ports.go", "command_models.go", "command_service.go"} {
		contents, err := os.ReadFile(applicationFile)
		if err != nil {
			t.Fatalf("read %s: %v", applicationFile, err)
		}
		for _, concreteDependency := range []string{"postgres.DB", "authn.UserRecord", "records.Store"} {
			if strings.Contains(string(contents), concreteDependency) {
				t.Errorf("public application file %s exposes concrete dependency %s", applicationFile, concreteDependency)
			}
		}
	}

	routes, err := os.ReadFile(filepath.Join("httpapi", "routes.go"))
	if err != nil {
		t.Fatalf("read Revisions HTTP adapter: %v", err)
	}
	routeSource := string(routes)
	interfaceStart := strings.Index(routeSource, "type commandApplication interface {")
	if interfaceStart < 0 {
		t.Fatal("Revisions HTTP command application interface is missing")
	}
	interfaceEnd := strings.Index(routeSource[interfaceStart:], "\n}")
	if interfaceEnd < 0 {
		t.Fatal("Revisions HTTP command application interface is unterminated")
	}
	applicationInterface := routeSource[interfaceStart : interfaceStart+interfaceEnd]
	for _, operation := range []string{"GetHistory(", "RollbackRecord(", "SoftDeleteRecord(", "RestoreRecord("} {
		if strings.Count(applicationInterface, operation) != 1 {
			t.Errorf("HTTP application interface operation %s count = %d, want 1", operation, strings.Count(applicationInterface, operation))
		}
	}

	for _, authorizationBoundary := range []string{"delete_restore_store.go", "rollback_coordinator.go"} {
		contents, err := os.ReadFile(authorizationBoundary)
		if err != nil {
			t.Fatalf("read %s: %v", authorizationBoundary, err)
		}
		if !strings.Contains(string(contents), ".AuthorizeCommandTx(") {
			t.Errorf("%s does not recheck authorization inside its transaction", authorizationBoundary)
		}
	}

	conflictEntries, err := os.ReadDir("conflicts")
	if err != nil {
		t.Fatalf("read Revisions conflict capability: %v", err)
	}
	for _, entry := range conflictEntries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		contents, err := os.ReadFile(filepath.Join("conflicts", entry.Name()))
		if err != nil {
			t.Fatalf("read conflict capability file %s: %v", entry.Name(), err)
		}
		for _, forbiddenImport := range []string{`"net/http"`, `internal/platform/authn`, `internal/platform/postgres`} {
			if strings.Contains(string(contents), forbiddenImport) {
				t.Errorf("Revisions conflict capability file %s imports transport/platform concern %s", entry.Name(), forbiddenImport)
			}
		}
	}
}
