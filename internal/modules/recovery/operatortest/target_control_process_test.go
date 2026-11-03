package operatortest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRecoveryRestoreInterruptionAndCompletionReplay_Process(t *testing.T) {
	ctx := t.Context()
	harness := pgtest.Start(t)
	source := harness.PrepareIsolatedDatabaseT(t, "control-process-source")
	sourceConfig := operatorExplicitConfig(t, source.DSN, postgres.PurposeRecovery)
	seedOperatorUser(t, source.DSN, "control-process@example.test", true, true)
	sourcePool := mustOpenOperatorPool(t, source.DSN)
	storage := newOperatorEncryptedBackupStorage(t, sourceConfig.backupRoot)
	backupID := uuid.New()
	seedOperatorRecoveryBackupSet(t, ctx, sourcePool, sourceConfig, storage, backupID, time.Now().UTC().Add(-time.Minute), "control-process")
	bin := injectedOperatorBinary(t)

	for _, scenario := range []string{"interrupted_admission", "terminal_before_control"} {
		t.Run(scenario, func(t *testing.T) {
			ctx := t.Context()
			target := harness.PrepareIsolatedDatabaseT(t, scenario)
			targetConfig := operatorExplicitConfig(t, target.DSN, postgres.PurposeRecovery)
			targetPool := mustOpenOperatorPool(t, target.DSN)
			operationID := uuid.New()
			args := []string{"restore", "latest", "--source-config-file", sourceConfig.path, "--target-config-file", targetConfig.path, "--confirm-backup-set-id", backupID.String(), "--operation-id", operationID.String()}

			if scenario == "interrupted_admission" {
				// Block the real source admission journal after the independently
				// committed target transition. No production fault hook is installed.
				blocker, err := sourcePool.Acquire(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer blocker.Release()
				if _, err := blocker.Exec(ctx, `SELECT pg_advisory_lock(784193)`); err != nil {
					t.Fatal(err)
				}
				defer func() {
					settle, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if _, err := blocker.Exec(settle, `SELECT pg_advisory_unlock(784193)`); err != nil {
						t.Error(err)
					}
					if _, err := sourcePool.Exec(settle, `DROP TRIGGER block_control_admission ON operator_recovery_journal; DROP FUNCTION block_control_admission()`); err != nil {
						t.Error(err)
					}
				}()
				if _, err := sourcePool.Exec(ctx, `CREATE FUNCTION block_control_admission() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.operation='restore_latest' AND NEW.result='started' THEN PERFORM pg_advisory_xact_lock(784193); END IF; RETURN NEW; END $$;
CREATE TRIGGER block_control_admission BEFORE INSERT ON operator_recovery_journal FOR EACH ROW EXECUTE FUNCTION block_control_admission()`); err != nil {
					t.Fatal(err)
				}
				commandCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				defer cancel()
				cmd := exec.CommandContext(commandCtx, bin, args...)
				cmd.Dir, cmd.Env = repoRoot(), append(os.Environ(), envPairs(operatorRecoveryEnv())...)
				var stdout, stderr bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				if err := cmd.Start(); err != nil {
					t.Fatal(err)
				}
				waited := false
				defer func() {
					if !waited {
						_ = cmd.Process.Kill()
						_ = cmd.Wait()
					}
				}()
				deadline := time.NewTimer(15 * time.Second)
				defer deadline.Stop()
				tick := time.NewTicker(25 * time.Millisecond)
				defer tick.Stop()
				for {
					control, err := recovery.ReadTargetControl(ctx, targetPool)
					if err != nil {
						t.Fatal(err)
					}
					if control.State == recovery.TargetRestoring {
						break
					}
					select {
					case <-tick.C:
					case <-deadline.C:
						t.Fatal("operator did not durably admit target")
					case <-commandCtx.Done():
						t.Fatal(commandCtx.Err())
					}
				}
				if err := cmd.Process.Kill(); err != nil {
					t.Fatal(err)
				}
				if err := cmd.Wait(); err == nil {
					t.Fatal("killed operator unexpectedly succeeded")
				}
				waited = true
				control, err := recovery.ReadTargetControl(ctx, targetPool)
				if err != nil || control.State != recovery.TargetRestoring || control.CompletionEvidenceSHA256 != nil {
					t.Fatal("interruption lost quarantine", control, err)
				}
				var users int
				if err := targetPool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&users); err != nil || users != 0 {
					t.Fatal("interrupted admission mutated payload", users, err)
				}
				out, diagnostic, status := runOperatorBinary(t, bin, operatorRecoveryEnv(), args...)
				requireOperatorRecoveryFailure(t, out, diagnostic, status, "restore_latest", 3, "unsafe_restore_target", "target_database_not_fresh")
				after, err := recovery.ReadTargetControl(ctx, targetPool)
				if err != nil || after.State != recovery.TargetRestoring || after.Revision != control.Revision {
					t.Fatal("retry changed an unproven operation", after, err)
				}
				return
			}

			if _, err := targetPool.Exec(ctx, `CREATE FUNCTION reject_control_completion() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.state='complete' THEN RAISE EXCEPTION 'injected completion interruption'; END IF; RETURN NEW; END $$;
CREATE TRIGGER reject_control_completion BEFORE UPDATE ON recovery_target_state FOR EACH ROW EXECUTE FUNCTION reject_control_completion()`); err != nil {
				t.Fatal(err)
			}
			out, diagnostic, status := runOperatorBinaryWithTimeout(t, 60*time.Second, bin, operatorRecoveryEnv(), args...)
			requireOperatorRecoveryFailure(t, out, diagnostic, status, "restore_latest", 4, "restore_failed", "journal_write_failed")
			before, err := recovery.ReadTargetControl(ctx, targetPool)
			if err != nil || before.State != recovery.TargetRestoring || before.CompletionEvidenceSHA256 != nil {
				t.Fatal("failed completion opened target", before, err)
			}
			if _, err := targetPool.Exec(ctx, `DROP TRIGGER reject_control_completion ON recovery_target_state; DROP FUNCTION reject_control_completion();
CREATE FUNCTION reject_payload_replay() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'payload must not replay'; END $$;
CREATE TRIGGER reject_payload_replay BEFORE INSERT OR UPDATE OR DELETE ON users FOR EACH ROW EXECUTE FUNCTION reject_payload_replay()`); err != nil {
				t.Fatal(err)
			}
			out, diagnostic, status = runOperatorBinaryWithTimeout(t, 30*time.Second, bin, operatorRecoveryEnv(), args...)
			if status != 0 {
				t.Fatalf("completion reconciliation failed: exit=%d stdout=%s stderr=%s", status, out, diagnostic)
			}
			requireOperatorRecoverySuccess(t, decodeOperatorRecoveryResult(t, out), "restore_latest", backupID.String())
			completed, err := recovery.ReadTargetControl(ctx, targetPool)
			if err != nil || completed.State != recovery.TargetComplete || completed.Revision != before.Revision+1 || completed.CompletionEvidenceSHA256 == nil {
				t.Fatal("terminal proof did not complete target", completed, err)
			}
			key, err := recovery.ParseRecoveryEncryptionKey(operatorRecoveryMasterKey)
			if err != nil {
				t.Fatal(err)
			}
			repository, err := recoveryassembly.NewRecoveryEvidenceRepository(sourcePool, func() (recovery.RecoveryEncryptionKey, error) { return key, nil })
			if err != nil {
				t.Fatal(err)
			}
			record, err := repository.FindCompletion(ctx, operationID, application.OperationRestoreLatest, nil)
			if err != nil || record == nil {
				t.Fatal("missing authenticated terminal", err)
			}
			canonical, err := application.EncodeRecoveryCompletion(*record)
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(canonical)
			if *completed.CompletionEvidenceSHA256 != hex.EncodeToString(digest[:]) {
				t.Fatal("completion did not bind exact terminal plaintext")
			}
			out, diagnostic, status = runOperatorBinary(t, bin, operatorRecoveryEnv(), args...)
			if status != 0 {
				t.Fatalf("completed retry failed: exit=%d stdout=%s stderr=%s", status, out, diagnostic)
			}
			after, err := recovery.ReadTargetControl(ctx, targetPool)
			if err != nil || after.State != recovery.TargetComplete || after.Revision != completed.Revision {
				t.Fatal("exact retry changed completed control", after, err)
			}
		})
	}
}
