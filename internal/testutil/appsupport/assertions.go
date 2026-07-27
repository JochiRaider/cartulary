package appsupport

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	viewschematest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func RequireMigrationTables(t testing.TB, testID string, tables ...string) {
	t.Helper()

	missing := make([]string, 0)
	migrationFiles, err := filepath.Glob(
		filepath.Join(repositoryRoot(), "db", "migrations", "*.sql"),
	)
	if err != nil {
		t.Fatalf("%s failed to enumerate migrations: %v", testID, err)
	}
	contents := make([]string, 0, len(migrationFiles))
	for _, path := range migrationFiles {
		data, readErr := os.ReadFile(path) // #nosec G304 -- paths come from a repo-local migration glob.
		if readErr != nil {
			t.Fatalf("%s failed to read migration %s: %v", testID, path, readErr)
		}
		contents = append(contents, string(data))
	}

	for _, table := range tables {
		if !migrationDeclaresTable(contents, table) {
			missing = append(missing, table)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("%s missing migration-backed tables: %s", testID, strings.Join(missing, ", "))
	}
}

func RequireSchemaTables(
	t testing.TB,
	db *sql.DB,
	testID string,
	tables ...string,
) {
	t.Helper()
	if db == nil {
		t.Fatalf("%s requires a live sql.DB for schema table assertions", testID)
	}

	missing := make([]string, 0)
	for _, table := range tables {
		var exists bool
		if err := db.QueryRowContext(context.Background(), `
SELECT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = $1
)`, table).Scan(&exists); err != nil {
			t.Fatalf("%s failed schema lookup for table %s: %v", testID, table, err)
		}
		if !exists {
			missing = append(missing, table)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("%s missing runtime schema tables: %s", testID, strings.Join(missing, ", "))
	}
}

func RequireViewContract(t testing.TB, testID string, viewSchemaIDs ...string) {
	t.Helper()
	viewschematest.RequireViewSchema(t, testID, viewSchemaIDs...)
}

func RequireViewFieldBindingMode(
	t testing.TB,
	testID string,
	viewSchemaID string,
	fieldKey string,
	wantBindingMode string,
) {
	t.Helper()
	viewschematest.RequireFieldBindingMode(
		t,
		testID,
		viewSchemaID,
		fieldKey,
		wantBindingMode,
	)
}

func RequireRouteSurface(
	t testing.TB,
	testID string,
	server *httptestx.Server,
	method string,
	path string,
	body any,
	options ...func(*http.Request),
) *http.Response {
	t.Helper()
	if server == nil || server.HTTP == nil {
		t.Fatalf("%s requires a running HTTP server harness", testID)
	}

	requestBody := body
	if requestBody == nil && method != http.MethodGet {
		requestBody = map[string]any{}
	}
	req := httptestx.NewJSONRequest(t, method, server.HTTP.URL+path, requestBody)
	for _, option := range options {
		option(req)
	}
	resp := httptestx.Do(t, http.DefaultClient, req)
	if resp.StatusCode == http.StatusNotFound ||
		resp.StatusCode == http.StatusMethodNotAllowed {
		defer resp.Body.Close()
		t.Fatalf(
			"%s missing route surface %s %s: got HTTP %d",
			testID,
			method,
			path,
			resp.StatusCode,
		)
	}
	return resp
}

func repositoryRoot() string {
	root, err := suiteservices.FindRepoRoot()
	if err != nil {
		panic(err)
	}
	return root
}

func migrationDeclaresTable(contents []string, table string) bool {
	pattern := regexp.MustCompile(
		`(?i)create\s+table\s+(if\s+not\s+exists\s+)?` +
			regexp.QuoteMeta(table) +
			`\b`,
	)
	for _, content := range contents {
		if pattern.FindStringIndex(content) != nil {
			return true
		}
	}
	return false
}
