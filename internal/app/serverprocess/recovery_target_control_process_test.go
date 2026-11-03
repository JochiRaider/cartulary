package serverprocess

import (
	"encoding/json"
	"errors"
	"os/exec"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestRecoveryTargetControlRejectsProcessPublication_Process(t *testing.T) {
	pg, s3 := sharedProcessHarnesses(t)
	database := pg.PrepareIsolatedDatabaseT(t, "recovery-control-process")
	control, err := pgtest.OpenPurposeDatabase(t.Context(), database.DSN, postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close()
	bucket := bucketName("recovery-control-process")
	defer cleanupBucket(t, s3, bucket)
	env := newProcessEnv(t, processEnvOptions{Database: database.Env(), ObjectStore: s3.Env(bucket), BootstrapPath: fixtures.Path("bootstrap-admin", "canonical.json"), ConfigPath: writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))})
	healthy := processtest.StartServer(t, processtest.ServerOptions{Env: env})
	t.Cleanup(func() { healthy.Stop(t) })
	healthy.WaitForReady(t)
	healthy.Stop(t)
	const counts = `SELECT (SELECT count(*) FROM extension_state_metadata)+(SELECT count(*) FROM deployment_bootstrap_state)+(SELECT count(*) FROM jobs)`
	var before int
	if err := control.QueryRowContext(t.Context(), counts).Scan(&before); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"unproven", "restoring", "resetting", "verification_ready", "complete"} {
		t.Run(phase, func(t *testing.T) {
			_, err := control.ExecContext(t.Context(), `UPDATE recovery_target_state SET
 state=$1,revision=revision+1,
 target_generation_id=CASE WHEN $1='unproven' THEN NULL ELSE '10000000-0000-4000-8000-000000000001'::uuid END,
 purpose=CASE WHEN $1='unproven' THEN NULL WHEN $1 IN ('restoring','complete') THEN 'restore_target' ELSE 'restore_verification_target' END,
 operation=CASE WHEN $1 IN ('restoring','complete') THEN 'restore_latest' ELSE NULL END,
 operation_id=CASE WHEN $1 IN ('restoring','complete') THEN '20000000-0000-4000-8000-000000000001'::uuid ELSE NULL END,
 backup_set_id=CASE WHEN $1 IN ('restoring','complete') THEN '30000000-0000-4000-8000-000000000001'::uuid ELSE NULL END,
 recovery_state_catalog_sha256=CASE WHEN $1 IN ('restoring','complete') THEN repeat('1',64) ELSE NULL END,
 codec_registry_sha256=CASE WHEN $1 IN ('restoring','complete') THEN repeat('1',64) ELSE NULL END,
 database_sha256=CASE WHEN $1='unproven' THEN NULL ELSE repeat('1',64) END,
 object_store_sha256=CASE WHEN $1='unproven' THEN NULL ELSE repeat('1',64) END,
 reference_pack_storage_sha256=CASE WHEN $1='unproven' THEN NULL ELSE repeat('1',64) END,
 export_outputs_sha256=CASE WHEN $1='unproven' THEN NULL ELSE repeat('1',64) END,
 completion_evidence_sha256=CASE WHEN $1='complete' THEN repeat('1',64) ELSE NULL END`, phase)
			if err != nil {
				t.Fatal(err)
			}
			env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"] = t.TempDir()
			failed := processtest.StartServer(t, processtest.ServerOptions{Env: env})
			var exit *exec.ExitError
			if err := failed.WaitForExit(t); !errors.As(err, &exit) || exit.ExitCode() != 2 {
				t.Fatalf("incomplete target exit: %v", err)
			}
			failed.RequireConnectionRefused(t, "/healthz")
			failed.RequireConnectionRefused(t, "/readyz")
			failed.RequireConnectionRefused(t, "/api/v1/network-flow-tables")
			failed.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000")
			failed.RequireDiagnosticsCode(t, "recovery_target_incomplete")
			failed.RequireDiagnosticsField(t, "$", "recovery_target_incomplete")
			var envelope struct {
				Error struct {
					Details struct {
						Items []struct {
							Message string         `json:"message"`
							Details map[string]any `json:"details"`
						} `json:"items"`
					} `json:"details"`
				} `json:"error"`
			}
			if err := json.Unmarshal([]byte(failed.DiagnosticsJSON(t)), &envelope); err != nil || len(envelope.Error.Details.Items) != 1 {
				t.Fatalf("invalid safe diagnostic: %v", err)
			}
			finding := envelope.Error.Details.Items[0]
			if finding.Message != "Recovery target completion is not established; application startup is closed." || len(finding.Details) != 0 {
				t.Fatal("Recovery control escaped its safe diagnostic")
			}
			var after int
			if err := control.QueryRowContext(t.Context(), counts).Scan(&after); err != nil || after != before {
				t.Fatalf("rejected process initialized state: %d/%d %v", before, after, err)
			}
		})
	}
}
