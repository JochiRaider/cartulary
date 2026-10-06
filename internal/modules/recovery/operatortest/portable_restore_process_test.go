package operatortest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/workbookprobe"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
)

func TestCanonicalOperatorRestoreBundle_Process(t *testing.T) {
	for _, scenario := range []struct {
		name string
		age  time.Duration
	}{{"fresh", time.Minute}, {"expired_source_retention", 40 * 24 * time.Hour}} {
		t.Run(scenario.name, func(t *testing.T) {
			ctx := context.Background()
			harness := pgtest.Start(t)
			sourceDB := harness.PrepareIsolatedDatabaseT(t, "portable-source")
			targetDB := harness.PrepareIsolatedDatabaseT(t, "portable-target")
			sourceCfg := operatorExplicitConfig(t, sourceDB.DSN, postgres.PurposeRecovery)
			targetCfg := operatorExplicitConfig(t, targetDB.DSN, postgres.PurposeRecovery)
			actor := seedOperatorUser(t, sourceDB.DSN, "portable@example.test", true, true)
			incident := seedOperatorIncident(t, sourceDB.DSN, actor, "PORTABLE")
			pool := mustOpenOperatorPool(t, sourceDB.DSN)
			storage := newOperatorEncryptedBackupStorage(t, sourceCfg.backupRoot)
			id := uuid.New()
			point := time.Now().UTC().Add(-scenario.age)
			backup := seedOperatorRecoveryBackupSet(t, ctx, pool, sourceCfg, storage, id, point, "portable")
			state, err := recoveryassembly.CurrentRecoveryStateCatalog()
			if err != nil {
				t.Fatal(err)
			}
			bin := injectedOperatorBinary(t)
			binary, err := os.ReadFile(bin)
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(binary)
			output := filepath.Join(t.TempDir(), "bundle")
			destination, err := recoveryassembly.NewTransferDirectory(output)
			if err != nil {
				t.Fatal(err)
			}
			// Simulate a completed historical export. Its original retention deadline is
			// already past when the real current-time operator is invoked below.
			catalog := recovery.NewBackupCatalog(recovery.NewStore(pool), storage, operatorExtensionBackupCatalog(t), state)
			if _, err := catalog.ExportTransfer(ctx, backup, hex.EncodeToString(digest[:]), point.Add(time.Second), destination); err != nil {
				t.Fatal(err)
			}
			if err := destination.Publish(ctx); err != nil {
				t.Fatal(err)
			}
			if err := destination.Close(); err != nil {
				t.Fatal(err)
			}
			sourceDeployment := loadOperatorConfig(t, sourceCfg.path)
			pool.Close()
			if err := recovery.CloseBackupStorage(storage); err != nil {
				t.Fatal(err)
			}
			if err := harness.DropDatabase(ctx, sourceDB.Name); err != nil {
				t.Fatal(err)
			}
			for _, root := range []string{sourceCfg.objectRoot, sourceCfg.backupRoot, sourceDeployment.Roots.ReferencePackStorage.Path, sourceDeployment.Roots.ExportOutputs.Path} {
				if err := os.RemoveAll(root); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Remove(sourceCfg.path); err != nil {
				t.Fatal(err)
			}
			op := uuid.NewString()
			args := []string{"restore", "bundle", "--bundle-directory", output, "--target-config-file", targetCfg.path, "--confirm-backup-set-id", id.String(), "--operation-id", op, "--progress", "jsonl"}
			wrong := append([]string{}, args...)
			wrong[7] = uuid.NewString()
			stdout, stderr, status := runOperatorBinary(t, bin, operatorRecoveryEnv(), portableWithoutProgress(wrong)...)
			requireOperatorRecoveryFailure(t, stdout, stderr, status, "restore_bundle", 2, "invalid_operator_request", "confirmation_mismatch")
			requireOperatorRestoreTargetUnmutated(t, targetDB.DSN)
			if scenario.age > 24*time.Hour {
				stdout, stderr, status = runOperatorBinary(t, bin, operatorRecoveryEnv(), portableWithoutProgress(args)...)
				requireOperatorRecoveryFailure(t, stdout, stderr, status, "restore_bundle", 2, "invalid_operator_request", "stale_backup_unacknowledged")
				requireOperatorRestoreTargetUnmutated(t, targetDB.DSN)
				args = append(args, "--acknowledge-stale-backup", id.String())
			}
			targetPool := mustOpenOperatorPool(t, targetDB.DSN)
			if _, err := targetPool.Exec(ctx, `CREATE FUNCTION fail_portable_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected audit publication failure'; END $$;
CREATE TRIGGER fail_portable_audit BEFORE INSERT ON deployment_admin_audit_events FOR EACH ROW EXECUTE FUNCTION fail_portable_audit();`); err != nil {
				t.Fatal(err)
			}
			stdout, stderr, status = runOperatorBinaryWithTimeout(t, 60*time.Second, bin, operatorRecoveryEnv(), portableWithoutProgress(args)...)
			requireOperatorRecoveryFailure(t, stdout, stderr, status, "restore_bundle", 4, "restore_failed", "journal_write_failed")
			if _, err := targetPool.Exec(ctx, `DROP TRIGGER fail_portable_audit ON deployment_admin_audit_events; DROP FUNCTION fail_portable_audit();`); err != nil {
				t.Fatal(err)
			}
			stdout, stderr, status = runOperatorBinaryWithTimeout(t, 60*time.Second, bin, operatorRecoveryEnv(), args...)
			if status != 0 {
				t.Fatalf("source-loss restore: exit=%d stdout=%s stderr=%s", status, stdout, stderr)
			}
			payload := decodeOperatorRecoveryResult(t, stdout)
			requireOperatorRecoverySuccess(t, payload, "restore_bundle", id.String())
			requireOperatorRecoveryProgress(t, stderr, op, []string{"preflight", "artifact_validate", "journal_write", "finalize"})
			var count int
			if err := targetPool.QueryRow(ctx, `SELECT count(*) FROM users WHERE email='portable@example.test'`).Scan(&count); err != nil || count != 1 {
				t.Fatal("restored user missing", count, err)
			}
			targetDeployment := loadOperatorConfig(t, targetCfg.path)
			packs, err := referenceassembly.NewRootStorage(targetDeployment.Roots.TemporaryWork.Path, targetDeployment.Roots.ReferencePackStorage.Path)
			if err != nil {
				t.Fatal(err)
			}
			defer packs.Close()
			if err := reference_data.ValidateRequiredState(ctx, targetPool, packs, reference_data.DefaultLimits()); err != nil {
				t.Fatal(err)
			}
			_, query, err := projectionassembly.NewRecoveryServices(targetPool)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := query.ExecuteDefault(ctx, workbookprobe.BaseProfile, incident); err != nil {
				t.Fatal("restored workbook probe", err)
			}
			localFile := filepath.Join(targetCfg.backupRoot, "portable-restore-journal", op, "completion.sealed")
			terminalBytes, err := os.ReadFile(localFile)
			if err != nil {
				t.Fatal(err)
			}
			// A target audit publication can be retried from the local terminal record.
			// Exact replay must not rerun engine writes or alter the encrypted journal.
			stdout, stderr, status = runOperatorBinaryWithTimeout(t, 60*time.Second, bin, operatorRecoveryEnv(), args...)
			if status != 0 {
				t.Fatalf("exact portable replay: exit=%d stdout=%s stderr=%s", status, stdout, stderr)
			}
			requireOperatorRecoveryProgress(t, stderr, op, []string{"preflight", "artifact_validate", "journal_write", "finalize"})
			replayPayload := decodeOperatorRecoveryResult(t, stdout)
			firstResult, _ := json.Marshal(payload)
			replayResult, _ := json.Marshal(replayPayload)
			if string(firstResult) != string(replayResult) {
				t.Fatal("terminal result replay changed admitted evidence")
			}

			if err := targetPool.QueryRow(ctx, `SELECT count(*) FROM operator_recovery_journal WHERE operation_id=$1 AND operation='restore_bundle'`, op).Scan(&count); err != nil || count != 1 {
				t.Fatal("portable replay duplicated terminal evidence", count, err)
			}
			after, err := os.ReadFile(localFile)
			if err != nil || string(after) != string(terminalBytes) {
				t.Fatal("portable replay changed encrypted completion", err)
			}
			if err := targetPool.QueryRow(ctx, `SELECT count(*) FROM deployment_admin_audit_events WHERE request_id=$1`, op).Scan(&count); err != nil || count != 1 {
				t.Fatal("portable replay did not reconcile safe target audit", count, err)
			}
			corrupt := append([]byte{}, terminalBytes...)
			corrupt[len(corrupt)/2] ^= 1
			if err := os.WriteFile(localFile, corrupt, 0600); err != nil {
				t.Fatal(err)
			}
			stdout, stderr, status = runOperatorBinary(t, bin, operatorRecoveryEnv(), portableWithoutProgress(args)...)
			requireOperatorRecoveryFailure(t, stdout, stderr, status, "restore_bundle", 3, "unsafe_restore_target", "target_database_not_fresh")
			if err := os.WriteFile(localFile, terminalBytes, 0600); err != nil {
				t.Fatal(err)
			}
			// Lost local completion is indeterminate, never permission to restore again.
			if err := os.Remove(localFile); err != nil {
				t.Fatal(err)
			}
			stdout, stderr, status = runOperatorBinary(t, bin, operatorRecoveryEnv(), portableWithoutProgress(args)...)
			requireOperatorRecoveryFailure(t, stdout, stderr, status, "restore_bundle", 3, "unsafe_restore_target", "target_database_not_fresh")
			if err := os.WriteFile(localFile, terminalBytes, 0600); err != nil {
				t.Fatal(err)
			}
			// The immutable journal envelope contains no plaintext completion record.
			var envelope recovery.OperatorRecoveryJournalEnvelope
			if err := json.Unmarshal(terminalBytes, &envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.SchemaID != recovery.OperatorRecoveryJournalSchemaID {
				t.Fatal("wrong encrypted journal schema")
			}
			requireOperatorRecoverySafeOutput(t, stdout, stderr, output, targetCfg.path, sourceDB.DSN, targetDB.DSN, operatorRecoveryMasterKey)
			// Completed restore does not manufacture a fresh operational backup.
			stdout, stderr, status = runOperatorBinary(t, bin, operatorRecoveryEnv(), "backup", "inspect", "latest", "--source-config-file", targetCfg.path)
			requireOperatorRecoveryFailure(t, stdout, stderr, status, "backup_inspect_latest", 3, "backup_set_not_found", "no_successful_retained_backup")
		})
	}
}

func portableWithoutProgress(args []string) []string {
	out := []string{}
	for index := 0; index < len(args); index++ {
		if args[index] == "--progress" {
			index++
			continue
		}
		out = append(out, args[index])
	}
	return out
}
