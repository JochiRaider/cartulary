package appsupport

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testfailure"
)

// Runtime owns local deployment bytes for a parent test. Managed services remain
// with their providers; each server borrows explicit fixtures or allocates its own.
type Runtime struct {
	referencePackRoot string
}

type ServerHarness struct {
	Server                    *httptestx.Server
	Database                  *pgtest.TestDatabase
	DB                        *sql.DB
	Pool                      *pgxpool.Pool
	ObjectStore               objectstore.Store
	Jobs                      *httptestx.JobsCapability
	Collaboration             *httptestx.CollaborationCapability
	Revisions                 *httptestx.RevisionsCapability
	Projections               *httptestx.ProjectionCapability
	IndicatorSourceText       indicators.SourceTextPort
	IncidentCreateCommitFault *IncidentCreateCommitFaultCapability
}

// ServerOptions constructs one application. Database and ObjectStore are borrowed
// when supplied and must outlive the server. Missing fixtures are isolated and
// owned by the calling test. TestRouteMode must be explicit.
type ServerOptions struct {
	ConfigureRuntime          func(*server.Options)
	Prefix                    string
	Database                  *pgtest.TestDatabase
	Env                       map[string]string
	Dependencies              httpapi.DependencySet
	AdditionalRoutes          []httpapi.RouteRegistrar
	ObjectStore               objectstore.Store
	TestRouteMode             httptestx.TestRouteMode
	IncidentCreateCommitFault bool
}

func StartRuntime(t testing.TB) *Runtime {
	t.Helper()
	return &Runtime{referencePackRoot: t.TempDir()}
}

// These per-call operations are private composition seams, not service owners.
type fixtureStarters struct {
	database    func(testing.TB, string) *pgtest.TestDatabase
	objectStore func(testing.TB, string) (objectstore.Store, map[string]string)
}

type serverFixtures struct {
	database  *pgtest.TestDatabase
	store     objectstore.Store
	objectEnv map[string]string
}

func acquireServerFixtures(t testing.TB, options ServerOptions, declaration map[string]string, starters fixtureStarters) (serverFixtures, error) {
	t.Helper()
	if options.Prefix == "" {
		return serverFixtures{}, errors.New("app test server prefix is required")
	}
	switch options.TestRouteMode {
	case httptestx.TestRouteModeDisabled, httptestx.TestRouteModeHarnessOwned, httptestx.TestRouteModeCustomEnv:
	default:
		return serverFixtures{}, fmt.Errorf("app test server route mode must be explicit and valid, got %q", options.TestRouteMode)
	}
	required := []string{"postgres"}
	if options.ObjectStore == nil {
		required = []string{"object_store", "postgres"}
	}
	// Application Env overrides are not harness service assignments. Admission
	// covers the whole construction before either provider can acquire anything.
	if err := suiteservices.CheckServiceDependencies(declaration, required...); err != nil {
		return serverFixtures{}, err
	}
	result := serverFixtures{database: options.Database, store: options.ObjectStore}
	if result.database == nil {
		result.database = starters.database(t, options.Prefix)
	}
	if result.store == nil {
		result.store, result.objectEnv = starters.objectStore(t, options.Prefix)
	}
	return result, nil
}

func prepareApplicationDatabase(t testing.TB, prefix string) *pgtest.TestDatabase {
	t.Helper()
	suiteservices.RequireServiceDependencies(t, "appsupport", "postgres")
	return pgtest.Start(t).PrepareIsolatedDatabaseT(t, prefix)
}

func prepareApplicationObjectStore(t testing.TB, prefix string) (objectstore.Store, map[string]string) {
	t.Helper()
	service := s3test.Start(t)
	bucket := service.BootstrapBucketT(t, prefix)
	store, err := objectstore.Setup(context.Background(), objectstore.Settings{
		BindingKind:         "managed_service",
		Endpoint:            service.Endpoint,
		AccessKey:           service.AccessKey,
		SecretKey:           service.SecretKey,
		Secure:              service.Secure,
		Bucket:              bucket,
		RootCertificatePath: service.RootCertificatePath(),
	}, objectstore.Instrumentation{})
	if err != nil {
		t.Fatalf("open object store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			testfailure.CleanupFailure(t, "appsupport", "object_store", "attach", err)
		}
	})
	return store, service.Env(bucket)
}

func runtimeDependencyEnvelope(err error) testfailure.Envelope {
	service := "object_store"
	var dependencyErr *suiteservices.ServiceDependencyError
	if errors.As(err, &dependencyErr) && dependencyErr.Service != "" {
		service = dependencyErr.Service
	}
	return testfailure.NewEnvelope(
		"harness",
		"fixture_error",
		"appsupport",
		service,
		"dependency_guard",
		0,
		"not_required",
	)
}

// StartServer constructs one application with an independent local lifetime.
func StartServer(t testing.TB, options ServerOptions) *ServerHarness {
	t.Helper()
	return StartRuntime(t).StartServer(t, options)
}

func (r *Runtime) PrepareIsolatedDatabase(t testing.TB, prefix string) *pgtest.TestDatabase {
	t.Helper()
	return prepareApplicationDatabase(t, prefix)
}

func (r *Runtime) StartServer(t testing.TB, options ServerOptions) *ServerHarness {
	t.Helper()
	resources, err := acquireServerFixtures(t, options, nil, fixtureStarters{
		database:    prepareApplicationDatabase,
		objectStore: prepareApplicationObjectStore,
	})
	if err != nil {
		var dependencyErr *suiteservices.ServiceDependencyError
		if errors.As(err, &dependencyErr) {
			testfailure.Fail(t, runtimeDependencyEnvelope(err))
		}
		t.Fatalf("configure app test server: %v", err)
	}
	testDB, store := resources.database, resources.store
	env := testDB.Env()
	// A restarted deployment keeps its authoritative local bytes with its
	// database. Allocate under the runtime's parent test, not a shorter-lived
	// subtest that merely replaces one server or dependency adapter.
	packRoot := filepath.Join(r.referencePackRoot, testDB.Name, "reference-packs")
	if err := os.MkdirAll(packRoot, 0o700); err != nil {
		t.Fatalf("create deployment Reference Pack root: %v", err)
	}
	env["CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH"] = packRoot
	for key, value := range resources.objectEnv {
		env[key] = value
	}
	for key, value := range options.Env {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")
	admittedPool := openServerPool(t, testDB)
	pool := admittedPool.Pool()
	var incidentCreateCommitFault *IncidentCreateCommitFaultCapability
	if options.IncidentCreateCommitFault {
		var err error
		incidentCreateCommitFault, err = newIncidentCreateCommitFaultCapability(env, options.TestRouteMode, pool)
		if err != nil {
			t.Fatalf("install incident create commit fault: %v", err)
		}
	}

	server := httptestx.StartServer(t, httptestx.ServerOptions{
		ConfigureRuntime: options.ConfigureRuntime,
		Env:              env,
		Dependencies:     options.Dependencies,
		AdditionalRoutes: append([]httpapi.RouteRegistrar(nil), options.AdditionalRoutes...),
		Postgres:         admittedPool,
		ObjectStore:      store,
		TestRouteMode:    options.TestRouteMode,
	})

	harness := serverHarnessForDatabase(t, testDB, server, pool, store)
	harness.IncidentCreateCommitFault = incidentCreateCommitFault
	return harness
}

func openServerPool(t testing.TB, testDB *pgtest.TestDatabase) postgres.AdmittedPool {
	t.Helper()
	dsn, err := testDB.DSNForPurpose(postgres.PurposeRuntime)
	if err != nil {
		t.Fatalf("resolve postgres runtime DSN: %v", err)
	}
	pool, err := postgres.Setup(context.Background(), postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          dsn,
		Purpose:      postgres.PurposeRuntime,
		ExpectedRole: "cartulary_runtime",
	})
	if err != nil {
		t.Fatalf("open postgres pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func serverHarnessForDatabase(
	t testing.TB,
	testDB *pgtest.TestDatabase,
	server *httptestx.Server,
	pool *pgxpool.Pool,
	store objectstore.Store,
) *ServerHarness {
	t.Helper()
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open postgres sql handle: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			testfailure.CleanupFailure(t, "appsupport", "postgres", "attach", err)
		}
	})
	return &ServerHarness{
		Server:              server,
		Database:            testDB,
		DB:                  db,
		Pool:                pool,
		ObjectStore:         store,
		Jobs:                server.JobsCapability(),
		Collaboration:       server.CollaborationCapability(),
		Revisions:           server.RevisionsCapability(),
		Projections:         server.ProjectionCapability(),
		IndicatorSourceText: server.IndicatorSourceTextPort(),
	}
}
