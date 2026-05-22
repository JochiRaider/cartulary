package main

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func phase10ProcessBlocked(t *testing.T) {
	t.Helper()
	t.Skip("Phase 10 process-bound operational recovery behavior is not implemented; blocker sentinel only")
}

func TestPhase10_U_10_02_RestoreReadinessBlocked(t *testing.T) {
	phase10ProcessBlocked(t)
}

func TestPhase10_U_10_03_FailClosedRestoreVerificationBlocked(t *testing.T) {
	phase10ProcessBlocked(t)
}

func TestPhase10_I_10_02_FreshEnvironmentRestoreWorkbookConsistencyBlocked(t *testing.T) {
	phase10ProcessBlocked(t)
}

func TestPhase10_I_10_03_MissingArtifactFailsBeforeReadinessBlocked(t *testing.T) {
	phase10ProcessBlocked(t)
}

func TestPhase10_E_10_03_PublicRouteInventoryAbsence(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)
	testDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-03-route-absence")

	bucket := phase0BucketName("phase10-e-10-03")
	defer cleanupPhase0Bucket(t, s3Harness, bucket)

	configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
	env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))
	server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
	defer server.Stop(t)
	server.WaitForReady(t)

	for _, path := range []string{
		"/api/v1/backups",
		"/api/v1/backups/latest",
		"/api/v1/restores",
		"/api/v1/restore-verifications",
		"/ws/v1/backups",
		"/ws/v1/restores",
		"/ws/v1/restore-verifications",
	} {
		server.RequireStatus(t, path, http.StatusNotFound)
	}
}
