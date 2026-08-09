package serverprocess

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

const (
	processCleanupTimeout = 10 * time.Second
	processHTTPTimeout    = 30 * time.Second
)

type processEnvOptions struct {
	Database      map[string]string
	ObjectStore   map[string]string
	ConfigPath    string
	BootstrapPath string
	Overrides     map[string]string
}

type processCleanupReporter interface {
	Helper()
	Errorf(format string, args ...any)
}

func newProcessEnv(t testing.TB, options processEnvOptions) map[string]string {
	t.Helper()

	env := make(map[string]string, len(options.Database)+len(options.ObjectStore)+len(options.Overrides)+16)
	mergeProcessEnv(env, options.Database)
	mergeProcessEnv(env, options.ObjectStore)
	tempRoots := configtest.SetupTempRoots(t)
	mergeProcessEnv(env, tempRoots.Paths)
	if options.ConfigPath != "" {
		env["CARTULARY_CONFIG_FILE"] = options.ConfigPath
	}
	if options.BootstrapPath != "" {
		env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = options.BootstrapPath
	}
	configtest.EnsureRevisionsConflictTokenTestEnvironment(env)
	mergeProcessEnv(env, options.Overrides)
	configtest.BindPostgresEnvToDatabaseRoot(t, env["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], env)
	return env
}

func mergeProcessEnv(target map[string]string, source map[string]string) {
	for key, value := range source {
		target[key] = value
	}
}

func newProcessCleanupContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), processCleanupTimeout)
}

func newProcessHTTPClient() *http.Client {
	return &http.Client{Timeout: processHTTPTimeout}
}

func reportProcessCleanupFailure(reporter processCleanupReporter, operation string, err error) {
	reporter.Helper()
	if err != nil {
		reporter.Errorf("%s: %v", operation, err)
	}
}

func startServerProcess(t testing.TB, prefix string) *processtest.Server {
	t.Helper()

	server, _ := startServerProcessWithDB(t, prefix)
	return server
}

func startServerProcessWithDB(t testing.TB, prefix string) (*processtest.Server, *sql.DB) {
	t.Helper()

	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open postgres sql handle: %v", err)
	}
	t.Cleanup(func() { closeSQL(t, db) })

	bucket := bucketName(prefix)
	t.Cleanup(func() { cleanupBucket(t, s3Harness, bucket) })
	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	env := newProcessEnv(t, processEnvOptions{
		Database:      testDB.Env(),
		ObjectStore:   s3Harness.Env(bucket),
		ConfigPath:    configPath,
		BootstrapPath: fixtures.Path("bootstrap-admin", "canonical.json"),
		Overrides: map[string]string{
			"CARTULARY_ENABLE_TEST_ROUTES":  "1",
			"CARTULARY_TEST_RUNTIME_MARKER": "harness-owned",
			"CARTULARY_TEST_ROUTE_TOKEN":    httptestx.TestRouteToken,
		},
	})

	server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
	t.Cleanup(func() { server.Stop(t) })
	server.WaitForReady(t)
	return server, db
}

func doJSON(t testing.TB, server *processtest.Server, method string, path string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()
	return flowtest.DoJSON(t, method, server.BaseURL+path, body, options...)
}
