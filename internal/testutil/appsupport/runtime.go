package appsupport

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

type Runtime struct {
	Postgres *pgtest.Harness
	S3       *s3test.Harness
}

type ServerHarness struct {
	Server *httptestx.Server
	DB     *sql.DB
}

type ServerOptions struct {
	Prefix               string
	Database             *pgtest.TestDatabase
	Env                  map[string]string
	Dependencies         httpapi.DependencySet
	AdditionalRoutes     []httpapi.RouteRegistrar
	ObjectStore          objectstore.Store
	TestRouteMode        httptestx.TestRouteMode
	TimelineDependencies func(postgres.DB) timeline.Dependencies
}

func StartRuntime(t testing.TB) *Runtime {
	t.Helper()
	return &Runtime{
		Postgres: pgtest.Start(t),
		S3:       s3test.Start(t),
	}
}

func (r *Runtime) PrepareIsolatedDatabase(t testing.TB, prefix string) *pgtest.TestDatabase {
	t.Helper()
	return r.Postgres.PrepareIsolatedDatabaseT(t, prefix)
}

func (r *Runtime) PrepareGroupDatabase(t testing.TB, prefix string, groupKey string) *pgtest.TestDatabase {
	t.Helper()
	return r.Postgres.PrepareGroupDatabaseT(t, prefix, groupKey)
}

func (r *Runtime) StartServer(t testing.TB, options ServerOptions) *ServerHarness {
	t.Helper()
	if options.Prefix == "" {
		t.Fatal("app test server prefix is required")
	}
	if options.TestRouteMode == "" {
		t.Fatal("app test server route mode is required")
	}

	testDB := options.Database
	if testDB == nil {
		testDB = r.PrepareIsolatedDatabase(t, options.Prefix)
	}
	bucket := r.S3.PreparePackageBucketT(t, options.Prefix)

	env := testDB.Env()
	for key, value := range r.S3.Env(bucket) {
		env[key] = value
	}
	for key, value := range options.Env {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	server := httptestx.StartServer(t, httptestx.ServerOptions{
		Env:                  env,
		Dependencies:         options.Dependencies,
		AdditionalRoutes:     append([]httpapi.RouteRegistrar(nil), options.AdditionalRoutes...),
		ObjectStore:          options.ObjectStore,
		TestRouteMode:        options.TestRouteMode,
		TimelineDependencies: options.TimelineDependencies,
	})

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open postgres sql handle: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return &ServerHarness{Server: server, DB: db}
}
