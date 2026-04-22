package phase4test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

type ServerHarness struct {
	Server *httptestx.Server
	DB     *sql.DB
}

type RuntimeHarness struct {
	Postgres *pgtest.Harness
	S3       *s3test.Harness
}

func StartServer(t testing.TB, prefix string) *ServerHarness {
	t.Helper()

	return StartRuntime(t).StartServer(t, prefix)
}

func StartRuntime(t testing.TB) *RuntimeHarness {
	t.Helper()

	return &RuntimeHarness{
		Postgres: pgtest.Start(t),
		S3:       s3test.Start(t),
	}
}

func (h *RuntimeHarness) StartServer(t testing.TB, prefix string) *ServerHarness {
	t.Helper()

	testDB := h.Postgres.PrepareDatabaseT(t, prefix)

	bucket, err := h.S3.BootstrapBucket(context.Background(), prefix)
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	t.Cleanup(func() {
		if err := h.S3.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	})

	env := testDB.Env()
	for key, value := range h.S3.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env})
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return &ServerHarness{
		Server: server,
		DB:     db,
	}
}

func RequireMigrationTables(t testing.TB, testID string, tables ...string) {
	t.Helper()

	missing := make([]string, 0)
	migrationFiles, err := filepath.Glob(filepath.Join(repoRoot(), "db", "migrations", "*.sql"))
	if err != nil {
		t.Fatalf("Phase 4 %s failed to enumerate migrations: %v", testID, err)
	}
	contents := make([]string, 0, len(migrationFiles))
	for _, path := range migrationFiles {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("Phase 4 %s failed to read migration %s: %v", testID, path, readErr)
		}
		contents = append(contents, string(data))
	}

	for _, table := range tables {
		if !migrationDeclaresTable(contents, table) {
			missing = append(missing, table)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("Phase 4 %s missing migration-backed tables: %s", testID, strings.Join(missing, ", "))
	}
}

func RequireSchemaTables(t testing.TB, db *sql.DB, testID string, tables ...string) {
	t.Helper()
	if db == nil {
		t.Fatalf("Phase 4 %s requires a live sql.DB for schema table assertions", testID)
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
			t.Fatalf("Phase 4 %s failed schema lookup for table %s: %v", testID, table, err)
		}
		if !exists {
			missing = append(missing, table)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("Phase 4 %s missing runtime schema tables: %s", testID, strings.Join(missing, ", "))
	}
}

func RequireViewContract(t testing.TB, testID string, viewSchemaIDs ...string) {
	t.Helper()

	missing := make([]string, 0)
	for _, viewSchemaID := range viewSchemaIDs {
		if _, err := os.Stat(viewContractPath(viewSchemaID)); err != nil {
			if os.IsNotExist(err) {
				missing = append(missing, viewSchemaID)
				continue
			}
			t.Fatalf("Phase 4 %s failed to stat view contract %s: %v", testID, viewSchemaID, err)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("Phase 4 %s missing view contracts: %s", testID, strings.Join(missing, ", "))
	}
}

func RequireViewFieldBindingMode(t testing.TB, testID string, viewSchemaID string, fieldKey string, wantBindingMode string) {
	t.Helper()
	document := loadViewContract(t, testID, viewSchemaID)

	for _, field := range document.Fields {
		if field.FieldKey != fieldKey {
			continue
		}
		if field.EntityBindingMode != wantBindingMode {
			t.Fatalf("Phase 4 %s expected %s %s entity_binding_mode=%s, got %s", testID, viewSchemaID, fieldKey, wantBindingMode, field.EntityBindingMode)
		}
		return
	}
	t.Fatalf("Phase 4 %s missing field %s in view contract %s", testID, fieldKey, viewSchemaID)
}

func RequireRouteSurface(t testing.TB, testID string, server *httptestx.Server, method string, path string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()
	if server == nil || server.HTTP == nil {
		t.Fatalf("Phase 4 %s requires a running HTTP server harness", testID)
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
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
		defer resp.Body.Close()
		t.Fatalf("Phase 4 %s missing route surface %s %s: got HTTP %d", testID, method, path, resp.StatusCode)
	}
	return resp
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}

func viewContractPath(viewSchemaID string) string {
	return filepath.Join(repoRoot(), "contracts", "view-schemas", fmt.Sprintf("%s.json", viewSchemaID))
}

func migrationDeclaresTable(contents []string, table string) bool {
	pattern := regexp.MustCompile(`(?i)create\s+table\s+(if\s+not\s+exists\s+)?` + regexp.QuoteMeta(table) + `\b`)
	for _, content := range contents {
		if pattern.FindStringIndex(content) != nil {
			return true
		}
	}
	return false
}
