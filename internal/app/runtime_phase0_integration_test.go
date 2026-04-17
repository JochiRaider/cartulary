package app

import (
	"context"
	"testing"

	"example.com/todo/cartulary/internal/platform/config"
	"example.com/todo/cartulary/internal/testutil/pgtest"
	"example.com/todo/cartulary/internal/testutil/s3test"
)

func TestPhase0_InvalidConfigNeverReachesReady_I_0_03(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-invalid-config")
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	defer func() {
		if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	}()

	s3Harness := s3test.Start(t)
	bucket, err := s3Harness.BootstrapBucket(context.Background(), "phase0-invalid-config")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}

	t.Run("rejects invalid filesystem roots even when services are healthy", func(t *testing.T) {
		cfg := phase0RuntimeConfig(t)
		cfg.Roots.DatabaseStorage.Path = "relative/postgres"

		_, err := NewRuntime(context.Background(), cfg, Options{Env: env})
		requireInvalidDeploymentConfig(t, err)
	})

	t.Run("rejects missing required runtime roots even when services are healthy", func(t *testing.T) {
		cfg := phase0RuntimeConfig(t)
		cfg.Roots.ExportOutputs = config.RootBinding{}

		_, err := NewRuntime(context.Background(), cfg, Options{Env: env})
		requireInvalidDeploymentConfig(t, err)
	})
}

func requireInvalidDeploymentConfig(t testing.TB, err error) {
	t.Helper()

	diagnosticsErr, ok := err.(*config.DiagnosticsError)
	if !ok {
		t.Fatalf("expected diagnostics error, got %T", err)
	}
	if diagnosticsErr.Code != config.InvalidDeploymentConfigCode {
		t.Fatalf("unexpected diagnostics code: got %q want %q", diagnosticsErr.Code, config.InvalidDeploymentConfigCode)
	}
}
