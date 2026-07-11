package scenariotest

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	viewschematest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
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

func StartServerWithDependencies(t testing.TB, prefix string, deps httpapi.DependencySet) *ServerHarness {
	t.Helper()

	return StartRuntime(t).StartServerWithDependencies(t, prefix, deps)
}

func StartServerWithConfig(t testing.TB, prefix string, mutate func(*config.Config)) *ServerHarness {
	t.Helper()

	return StartRuntime(t).StartServerWithConfig(t, prefix, mutate)
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

	return h.StartServerWithDependencies(t, prefix, httpapi.DependencySet{})
}

func (h *RuntimeHarness) StartServerWithDependencies(t testing.TB, prefix string, deps httpapi.DependencySet) *ServerHarness {
	t.Helper()

	testDB := h.PrepareServerDatabase(t, prefix)
	return h.StartServerWithDatabaseAndDependencies(t, prefix, testDB, deps)
}

func (h *RuntimeHarness) PrepareServerDatabase(t testing.TB, prefix string) *pgtest.TestDatabase {
	t.Helper()

	return h.Postgres.PrepareIsolatedDatabaseT(t, prefix)
}

func (h *RuntimeHarness) PrepareGroupServerDatabase(t testing.TB, prefix string, groupKey string) *pgtest.TestDatabase {
	t.Helper()

	return h.Postgres.PrepareGroupDatabaseT(t, prefix, groupKey)
}

func (h *RuntimeHarness) StartServerWithDatabase(t testing.TB, prefix string, testDB *pgtest.TestDatabase) *ServerHarness {
	t.Helper()

	return h.StartServerWithDatabaseAndDependencies(t, prefix, testDB, httpapi.DependencySet{})
}

func (h *RuntimeHarness) StartServerWithDatabaseAndDependencies(t testing.TB, prefix string, testDB *pgtest.TestDatabase, deps httpapi.DependencySet) *ServerHarness {
	t.Helper()

	bucket := h.S3.PreparePackageBucketT(t, prefix)

	env := serverDatabaseEnv(t, testDB)
	for key, value := range h.S3.Env(bucket) {
		env[key] = value
	}

	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env, Dependencies: deps, TestRouteMode: httptestx.TestRouteModeDisabled})
	return serverHarnessForDatabase(t, testDB, server)
}

func (h *RuntimeHarness) StartServerWithConfig(t testing.TB, prefix string, mutate func(*config.Config)) *ServerHarness {
	t.Helper()

	testDB := h.PrepareServerDatabase(t, prefix)

	bucket := h.S3.PreparePackageBucketT(t, prefix)

	env := serverDatabaseEnv(t, testDB)
	for key, value := range h.S3.Env(bucket) {
		env[key] = value
	}

	tempRoots := configtest.SetupTempRoots(t)
	for key, value := range tempRoots.Paths {
		if _, exists := env[key]; !exists {
			env[key] = value
		}
	}
	configtest.BindPostgresEnvToDatabaseRoot(t, tempRoots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], env)

	cfg := configtest.LoadEffectiveFixture(t, []string{"config", "valid.toml"}, env)
	if mutate != nil {
		mutate(&cfg)
	}

	server := httptestx.StartServer(t, httptestx.ServerOptions{Config: cfg, Env: env, TestRouteMode: httptestx.TestRouteModeDisabled})
	return serverHarnessForDatabase(t, testDB, server)
}

func (h *RuntimeHarness) StartServerWithObjectStore(t testing.TB, prefix string, store objectstore.Store) *ServerHarness {
	t.Helper()

	testDB := h.PrepareServerDatabase(t, prefix)
	return h.StartServerWithDatabaseAndObjectStore(t, prefix, testDB, store)
}

func (h *RuntimeHarness) StartServerWithDatabaseAndObjectStore(t testing.TB, prefix string, testDB *pgtest.TestDatabase, store objectstore.Store) *ServerHarness {
	t.Helper()

	env := serverDatabaseEnv(t, testDB)
	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env, ObjectStore: store, TestRouteMode: httptestx.TestRouteModeDisabled})
	return serverHarnessForDatabase(t, testDB, server)
}

func serverDatabaseEnv(t testing.TB, testDB *pgtest.TestDatabase) map[string]string {
	t.Helper()

	if testDB == nil {
		t.Fatal("test database is required")
	}
	env := testDB.Env()
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")
	return env
}

func serverHarnessForDatabase(t testing.TB, testDB *pgtest.TestDatabase, server *httptestx.Server) *ServerHarness {
	t.Helper()

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
		data, readErr := os.ReadFile(path) // #nosec G304 -- migration paths come from a repo-local db/migrations glob.
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
	viewschematest.RequireViewSchema(t, "Phase 4 "+testID, viewSchemaIDs...)
}

func RequireViewFieldBindingMode(t testing.TB, testID string, viewSchemaID string, fieldKey string, wantBindingMode string) {
	t.Helper()
	viewschematest.RequireFieldBindingMode(t, "Phase 4 "+testID, viewSchemaID, fieldKey, wantBindingMode)
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
	root, err := suiteservices.FindRepoRoot()
	if err != nil {
		panic(err)
	}
	return root
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
