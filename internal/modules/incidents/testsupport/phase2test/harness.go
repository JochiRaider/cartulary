package phase2test

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

type RuntimeHarness struct {
	Postgres *pgtest.Harness
	S3       *s3test.Harness
}

type ServerHarness struct {
	Server *httptestx.Server
	DB     *sql.DB
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

	return h.StartServerWithDependenciesAndRoutes(t, prefix, httpapi.DependencySet{})
}

func (h *RuntimeHarness) StartServerWithRoutes(t testing.TB, prefix string, routes ...httpapi.RouteRegistrar) *ServerHarness {
	t.Helper()

	return h.StartServerWithDependenciesAndRoutes(t, prefix, httpapi.DependencySet{}, routes...)
}

func (h *RuntimeHarness) StartServerWithDependencies(t testing.TB, prefix string, deps httpapi.DependencySet) *ServerHarness {
	t.Helper()

	return h.StartServerWithDependenciesAndRoutes(t, prefix, deps)
}

func (h *RuntimeHarness) StartServerWithDependenciesAndRoutes(t testing.TB, prefix string, deps httpapi.DependencySet, routes ...httpapi.RouteRegistrar) *ServerHarness {
	t.Helper()

	testDB := h.prepareDatabase(t, prefix)
	bucket := h.prepareBucket(t, prefix)

	env := testDB.Env()
	for key, value := range h.S3.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env, Dependencies: deps, AdditionalRoutes: routes})
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

func (h *RuntimeHarness) prepareDatabase(t testing.TB, prefix string) *pgtest.TestDatabase {
	t.Helper()

	return h.Postgres.PreparePackageDatabaseT(t, prefix)
}

func (h *RuntimeHarness) prepareBucket(t testing.TB, prefix string) string {
	t.Helper()

	return h.S3.PreparePackageBucketT(t, prefix)
}
