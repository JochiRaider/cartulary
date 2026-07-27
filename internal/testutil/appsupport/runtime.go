package appsupport

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
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
	Pool   *pgxpool.Pool
}

type ServerOptions struct {
	Prefix           string
	Database         *pgtest.TestDatabase
	Env              map[string]string
	Dependencies     httpapi.DependencySet
	AdditionalRoutes []httpapi.RouteRegistrar
	ObjectStore      objectstore.Store
	TestRouteMode    httptestx.TestRouteMode
}

func StartRuntime(t testing.TB) *Runtime {
	t.Helper()
	return &Runtime{
		Postgres: pgtest.Start(t),
		S3:       s3test.Start(t),
	}
}

func StartServer(t testing.TB, prefix string) *ServerHarness {
	t.Helper()
	return StartRuntime(t).startServer(t, ServerOptions{
		Prefix:        prefix,
		Dependencies:  httpapi.DependencySet{},
		TestRouteMode: httptestx.TestRouteModeDisabled,
	}, false)
}

func StartServerWithDependencies(
	t testing.TB,
	prefix string,
	dependencies httpapi.DependencySet,
) *ServerHarness {
	t.Helper()
	return StartRuntime(t).startServer(t, ServerOptions{
		Prefix:        prefix,
		Dependencies:  dependencies,
		TestRouteMode: httptestx.TestRouteModeDisabled,
	}, false)
}

func StartServerWithConfig(
	t testing.TB,
	prefix string,
	mutate func(*configassembly.Deployment),
) *ServerHarness {
	t.Helper()
	return StartRuntime(t).StartServerWithConfig(t, prefix, mutate)
}

func (r *Runtime) PrepareIsolatedDatabase(t testing.TB, prefix string) *pgtest.TestDatabase {
	t.Helper()
	return r.Postgres.PrepareIsolatedDatabaseT(t, prefix)
}

func (r *Runtime) PrepareGroupDatabase(t testing.TB, prefix string, groupKey string) *pgtest.TestDatabase {
	t.Helper()
	return r.Postgres.PrepareGroupDatabaseT(t, prefix, groupKey)
}

func (r *Runtime) PrepareServerDatabase(t testing.TB, prefix string) *pgtest.TestDatabase {
	t.Helper()
	return r.PrepareIsolatedDatabase(t, prefix)
}

func (r *Runtime) PrepareGroupServerDatabase(
	t testing.TB,
	prefix string,
	groupKey string,
) *pgtest.TestDatabase {
	t.Helper()
	return r.PrepareGroupDatabase(t, prefix, groupKey)
}

func (r *Runtime) StartServerWithDependencies(
	t testing.TB,
	prefix string,
	dependencies httpapi.DependencySet,
) *ServerHarness {
	t.Helper()
	return r.startServer(t, ServerOptions{
		Prefix:        prefix,
		Dependencies:  dependencies,
		TestRouteMode: httptestx.TestRouteModeDisabled,
	}, false)
}

func (r *Runtime) StartDefaultServer(
	t testing.TB,
	prefix string,
) *ServerHarness {
	t.Helper()
	return r.StartServerWithDependencies(t, prefix, httpapi.DependencySet{})
}

func (r *Runtime) StartServerWithDatabase(
	t testing.TB,
	prefix string,
	database *pgtest.TestDatabase,
) *ServerHarness {
	t.Helper()
	return r.StartServerWithDatabaseAndDependencies(
		t,
		prefix,
		database,
		httpapi.DependencySet{},
	)
}

func (r *Runtime) StartServerWithDatabaseAndDependencies(
	t testing.TB,
	prefix string,
	database *pgtest.TestDatabase,
	dependencies httpapi.DependencySet,
) *ServerHarness {
	t.Helper()
	return r.startServer(t, ServerOptions{
		Prefix:        prefix,
		Database:      database,
		Dependencies:  dependencies,
		TestRouteMode: httptestx.TestRouteModeDisabled,
	}, false)
}

func (r *Runtime) StartServerWithConfig(
	t testing.TB,
	prefix string,
	mutate func(*configassembly.Deployment),
) *ServerHarness {
	t.Helper()
	testDB := r.PrepareIsolatedDatabase(t, prefix)
	bucket := r.S3.PreparePackageBucketT(t, prefix)
	env := testDB.Env()
	for key, value := range r.S3.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path(
		"bootstrap-admin",
		"canonical.json",
	)
	tempRoots := configtest.SetupTempRoots(t)
	for key, value := range tempRoots.Paths {
		if _, exists := env[key]; !exists {
			env[key] = value
		}
	}
	configtest.BindPostgresEnvToDatabaseRoot(
		t,
		tempRoots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"],
		env,
	)
	cfg := configtest.LoadEffectiveFixture(t, []string{"config", "valid.toml"}, env)
	if mutate != nil {
		mutate(&cfg)
	}
	server := httptestx.StartServer(t, httptestx.ServerOptions{
		Config:        cfg,
		Env:           env,
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
	return serverHarnessForDatabase(t, testDB, server, false)
}

func (r *Runtime) StartServerWithObjectStore(
	t testing.TB,
	prefix string,
	store objectstore.Store,
) *ServerHarness {
	t.Helper()
	return r.startServer(t, ServerOptions{
		Prefix:        prefix,
		ObjectStore:   store,
		TestRouteMode: httptestx.TestRouteModeDisabled,
	}, false)
}

func (r *Runtime) StartServerWithDatabaseAndObjectStore(
	t testing.TB,
	prefix string,
	database *pgtest.TestDatabase,
	store objectstore.Store,
) *ServerHarness {
	t.Helper()
	return r.startServer(t, ServerOptions{
		Prefix:        prefix,
		Database:      database,
		ObjectStore:   store,
		TestRouteMode: httptestx.TestRouteModeDisabled,
	}, false)
}

func (r *Runtime) StartServer(t testing.TB, options ServerOptions) *ServerHarness {
	t.Helper()
	return r.startServer(t, options, true)
}

func (r *Runtime) startServer(
	t testing.TB,
	options ServerOptions,
	openPool bool,
) *ServerHarness {
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
	env := testDB.Env()
	if options.ObjectStore == nil {
		bucket := r.S3.PreparePackageBucketT(t, options.Prefix)
		for key, value := range r.S3.Env(bucket) {
			env[key] = value
		}
	}
	for key, value := range options.Env {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	server := httptestx.StartServer(t, httptestx.ServerOptions{
		Env:              env,
		Dependencies:     options.Dependencies,
		AdditionalRoutes: append([]httpapi.RouteRegistrar(nil), options.AdditionalRoutes...),
		ObjectStore:      options.ObjectStore,
		TestRouteMode:    options.TestRouteMode,
	})

	return serverHarnessForDatabase(t, testDB, server, openPool)
}

func serverHarnessForDatabase(
	t testing.TB,
	testDB *pgtest.TestDatabase,
	server *httptestx.Server,
	openPool bool,
) *ServerHarness {
	t.Helper()
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open postgres sql handle: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	var pool *pgxpool.Pool
	if openPool {
		pool, err = pgxpool.New(context.Background(), testDB.DSN)
		if err != nil {
			t.Fatalf("open postgres pool: %v", err)
		}
		t.Cleanup(pool.Close)
	}

	return &ServerHarness{Server: server, DB: db, Pool: pool}
}
