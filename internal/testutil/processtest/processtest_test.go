package processtest

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestStartServerAssignsAddressAndReachesReady(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	testDB := postgresHarness.PrepareDatabaseT(t, "processtest-ready")
	bucket, err := s3Harness.BootstrapBucket(context.Background(), "processtest-ready")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	t.Cleanup(func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	})

	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	env := processEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))

	server := StartServer(t, ServerOptions{Env: env})
	t.Cleanup(func() {
		server.Stop(t)
	})

	if server.Address == "" || server.BaseURL == "" {
		t.Fatalf("expected assigned process address, got address=%q base_url=%q", server.Address, server.BaseURL)
	}

	server.WaitForReady(t)
	server.RequireStatus(t, "/healthz", 200)
	server.RequireStatus(t, "/readyz", 200)
}

func TestStartServerExitsBeforeReadyAndParsesDiagnostics(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	testDB := postgresHarness.PrepareDatabaseT(t, "processtest-invalid")
	bucket, err := s3Harness.BootstrapBucket(context.Background(), "processtest-invalid")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	t.Cleanup(func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	})

	configPath := writeConfig(t, string(fixtures.MustRead("config", "invalid_missing_required.toml")))
	env := processEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, "")
	delete(env, "CARTULARY__ROOTS__DATABASE_STORAGE__PATH")
	delete(env, "CARTULARY__ROOTS__OBJECT_STORAGE__PATH")
	delete(env, "CARTULARY__ROOTS__BACKUP_STORAGE__PATH")
	delete(env, "CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH")
	delete(env, "CARTULARY__ROOTS__TEMPORARY_WORK__PATH")
	delete(env, "CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH")

	server := StartServer(t, ServerOptions{Env: env})

	err = server.WaitForExit(t)
	if err == nil {
		t.Fatal("expected invalid startup to exit non-zero")
	}
	server.RequireConnectionRefused(t, "/healthz")
	server.RequireConnectionRefused(t, "/readyz")
	server.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000")
	server.RequireDiagnosticsMatchGolden(t, []string{"config", "invalid_missing_required.json"})
	if diagnostics := server.Diagnostics(t); diagnostics["error"] == nil {
		t.Fatalf("expected error diagnostics payload, got %#v", diagnostics)
	}
}

func TestDiagnosticsJSONStripsGoRunExitSuffix(t *testing.T) {
	server := &Server{}
	server.stderr.WriteString("{\"error\":{\"code\":\"invalid_deployment_config\",\"details\":{\"items\":[{\"path\":\"x\",\"reason_code\":\"y\"}]}}}\nexit status 1\n")

	if got := server.DiagnosticsJSON(t); got != "{\"error\":{\"code\":\"invalid_deployment_config\",\"details\":{\"items\":[{\"path\":\"x\",\"reason_code\":\"y\"}]}}}" {
		t.Fatalf("unexpected diagnostics JSON: %q", got)
	}
}

func TestRequireConnectionRefusedBeforeListenerStartup(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve address: %v", err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("close reserved listener: %v", err)
	}

	server := &Server{
		Address: address,
		BaseURL: "http://" + address,
	}
	server.RequireConnectionRefused(t, "/healthz")
	server.RequireWebsocketConnectionRefused(t, "/ws/v1/bootstrap-harness")
}

func processEnv(t testing.TB, databaseEnv map[string]string, objectStoreEnv map[string]string, configPath string, bootstrapPath string) map[string]string {
	t.Helper()

	tempRoots := configtest.SetupTempRoots(t)
	env := make(map[string]string)
	for key, value := range databaseEnv {
		env[key] = value
	}
	for key, value := range objectStoreEnv {
		env[key] = value
	}
	for key, value := range tempRoots.Paths {
		env[key] = value
	}
	env["CARTULARY_CONFIG_FILE"] = configPath
	if bootstrapPath != "" {
		env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = bootstrapPath
	}
	return env
}

func writeConfig(t testing.TB, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
