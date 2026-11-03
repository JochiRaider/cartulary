package recovery_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestRestoreControlFencesAdmissionCompletionAndReset_Integration(t *testing.T) {
	for _, mode := range []string{"journal_failure", "ordinary_completion", "verification_reset", "lease_loss", "revision_exhausted", "identity_conflict"} {
		t.Run(mode, func(t *testing.T) {
			ctx := t.Context()
			fixture := newRestoreProjectionContractFixture(t, ctx, "restore-control-"+mode, uuid.New())
			projection, err := projectionassembly.Build(fixture.Target.Postgres)
			if err != nil {
				t.Fatal(err)
			}
			fixture.Target.Projections = projection.RecoveryPorts().Rebuilder
			pool := fixture.Target.Postgres.(postgres.AdmittedPool)
			admission, err := recoveryassembly.AcquireTargetServingAdmission(ctx, pool, 5*time.Second, 5*time.Second)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				settle, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = admission.Release(settle)
			})
			generation, err := recovery.NewBackupCatalog(fixture.Store, fixture.BackupStorage, testExtensionBackupCatalog(t), fixture.state).RecoveryGenerationIdentity(ctx, fixture.BackupSet)
			if err != nil {
				t.Fatal(err)
			}
			identity := recovery.RestoreTargetIdentity{TargetGenerationID: uuid.New(), Purpose: "restore_target", Operation: "restore_latest", OperationID: uuid.New(), BackupSetID: fixture.BackupSet.BackupSetID, RecoveryStateCatalogSHA256: generation.RecoveryStateCatalogSHA256, CodecRegistrySHA256: generation.CodecRegistrySHA256, BindingDigests: recovery.TargetBindingDigests{DatabaseSHA256: strings.Repeat("1", 64), ObjectStoreSHA256: strings.Repeat("2", 64), ReferencePackStorageSHA256: strings.Repeat("3", 64), ExportOutputsSHA256: strings.Repeat("4", 64)}}
			if mode == "verification_reset" {
				identity.Purpose = "restore_verification_target"
				identity.Operation = "restore_verify_latest"
			}
			read := func() recovery.TargetControl {
				t.Helper()
				control, err := recovery.ReadTargetControl(ctx, pool)
				if err != nil {
					t.Fatal(err)
				}
				return control
			}
			if _, err := fixture.Runner.RestoreBackupSet(ctx, fixture.Target, fixture.BackupSet); !errors.Is(err, recovery.ErrTargetIncomplete) {
				t.Fatalf("unfenced restore admitted: %v", err)
			}
			if read().State != recovery.TargetInitialized {
				t.Fatal("unfenced rejection changed control")
			}
			if mode == "revision_exhausted" {
				if _, err := pool.Exec(ctx, `UPDATE recovery_target_state SET revision=9223372036854775807`); err != nil {
					t.Fatal(err)
				}
				called := false
				fence, err := fixture.Runner.AdmitRestore(ctx, admission, fixture.Target, fixture.BackupSet, identity, func(context.Context) error { called = true; return nil })
				if !errors.Is(err, recovery.ErrTargetControlConflict) || fence != nil || called {
					t.Fatal("exhausted revision admitted mutation", err)
				}
				if control := read(); control.State != recovery.TargetInitialized || control.Revision != 9223372036854775807 {
					t.Fatal("control revision wrapped", control)
				}
				return
			}
			key, err := recovery.ParseRecoveryEncryptionKey("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")
			if err != nil {
				t.Fatal(err)
			}
			repository, err := recoveryassembly.NewRecoveryEvidenceRepository(fixture.sourcePool, func() (recovery.RecoveryEncryptionKey, error) { return key, nil })
			if err != nil {
				t.Fatal(err)
			}
			started := time.Now().UTC()
			journalFailure := errors.New("injected admission journal failure")
			fence, err := fixture.Runner.AdmitRestore(ctx, admission, fixture.Target, fixture.BackupSet, identity, func(work context.Context) error {
				control := read()
				if control.State != recovery.TargetRestoring || control.Revision != 2 || *control.OperationID != identity.OperationID.String() {
					t.Fatal("journal ran before durable restoring admission")
				}
				if mode == "journal_failure" {
					writer, err := pool.BeginTx(work, pgx.TxOptions{})
					if err != nil {
						return err
					}
					defer writer.Rollback(context.Background())
					if _, err := writer.Exec(work, `SET LOCAL lock_timeout='100ms'`); err != nil {
						return err
					}
					_, err = writer.Exec(work, `INSERT INTO users(email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES ('blocked@example.test','Blocked','fixture',false,true,false)`)
					var pgErr *pgconn.PgError
					if !errors.As(err, &pgErr) || pgErr.Code != "55P03" {
						t.Fatalf("pending data fence did not block writer: %v", err)
					}
					return journalFailure
				}
				return repository.AppendAdmission(work, application.RecoveryAdmissionRecord{OperationID: identity.OperationID, Operation: application.Operation(identity.Operation), StartedAt: started, BackupSetID: &identity.BackupSetID, ArtifactKinds: []string{}})
			})
			if mode == "journal_failure" {
				if !errors.Is(err, journalFailure) || fence != nil {
					t.Fatalf("journal failure admitted payload: %v", err)
				}
				if read().State != recovery.TargetRestoring {
					t.Fatal("payload rollback removed quarantine")
				}
				var users int
				if err := pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&users); err != nil || users != 0 {
					t.Fatalf("failed admission changed payload: %d %v", users, err)
				}
				called := false
				if _, err := fixture.Runner.AdmitRestore(ctx, admission, fixture.Target, fixture.BackupSet, identity, func(context.Context) error { called = true; return nil }); !errors.Is(err, recovery.ErrTargetIncomplete) || called {
					t.Fatalf("quarantined target was reused: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer fence.Close()
			if mode == "lease_loss" {
				var pid int
				if err := admission.WithControlTransaction(ctx, func(work context.Context, tx recovery.TargetControlTransaction) error {
					return tx.QueryRow(work, `SELECT pg_backend_pid()`).Scan(&pid)
				}); err != nil {
					t.Fatal(err)
				}
				admin, err := pgxpool.New(ctx, fixture.targetDatabase.DSN)
				if err != nil {
					t.Fatal(err)
				}
				defer admin.Close()
				var terminated bool
				if err := admin.QueryRow(ctx, `SELECT pg_terminate_backend($1,5000)`, pid).Scan(&terminated); err != nil || !terminated {
					t.Fatal("terminate original lease session", terminated, err)
				}
				select {
				case <-admission.Context().Done():
				case <-time.After(6 * time.Second):
					t.Fatal("original session loss did not cancel admission")
				}
				if _, err := fixture.Runner.RestoreBackupSet(ctx, fence.Target(), fixture.BackupSet); err == nil {
					t.Fatal("lost lease admitted payload")
				}
				if err := fence.Close(); err != nil {
					t.Fatal(err)
				}
				if control := read(); control.State != recovery.TargetRestoring || control.CompletionEvidenceSHA256 != nil {
					t.Fatal("lease loss reopened target", control)
				}
				var users int
				if err := pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&users); err != nil || users != 0 {
					t.Fatal("lost lease changed payload", users, err)
				}
				return
			}
			sealed := fence.Target()
			sealed.Postgres = nil // Replacement on a copy cannot replace admitted resources.
			result, err := fixture.Runner.RestoreBackupSet(ctx, sealed, fixture.BackupSet)
			if err != nil {
				t.Fatal(err)
			}
			if read().State != recovery.TargetRestoring {
				t.Fatal("payload alone published complete")
			}
			point := fixture.BackupSet.ConsistencyPointAt
			record := application.RecoveryCompletionRecord{OperationID: identity.OperationID, Operation: application.Operation(identity.Operation), StartedAt: started, CompletedAt: time.Now().UTC(), Result: application.ResultSucceeded, BackupSetID: &identity.BackupSetID, ConsistencyPointAt: &point, ArtifactCounts: []application.ArtifactCount{}, RestoreGeneration: result.RestoreGeneration, TargetBindings: &identity.BindingDigests}
			if err := repository.AppendCompletion(ctx, record); err != nil {
				t.Fatal(err)
			}
			stored, err := repository.FindCompletion(ctx, identity.OperationID, record.Operation, nil)
			if err != nil || stored == nil {
				t.Fatalf("authenticated terminal lookup: %v", err)
			}
			canonical, err := application.EncodeRecoveryCompletion(*stored)
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(canonical)
			proof := recovery.TargetCompletionProof{Identity: identity, EvidenceSHA256: hex.EncodeToString(digest[:])}
			var persistedDigest string
			if err := fixture.sourcePool.QueryRow(ctx, `SELECT payload_sha256 FROM operator_recovery_journal WHERE operation_id=$1 AND result='succeeded'`, identity.OperationID).Scan(&persistedDigest); err != nil || persistedDigest != proof.EvidenceSHA256 {
				t.Fatalf("terminal bytes differ from completion digest: %v", err)
			}
			if mode == "identity_conflict" {
				if _, err := pool.Exec(ctx, `UPDATE recovery_target_state SET operation_id=$1`, uuid.New()); err != nil {
					t.Fatal(err)
				}
				if err := fence.Complete(ctx, proof); !errors.Is(err, recovery.ErrTargetControlConflict) {
					t.Fatal("same-revision identity change was overwritten", err)
				}
				if err := recovery.ReconcileTargetCompletion(ctx, admission, proof); !errors.Is(err, recovery.ErrTargetIncomplete) {
					t.Fatal("conflicting operation inherited terminal proof", err)
				}
				if read().State != recovery.TargetRestoring {
					t.Fatal("identity conflict opened target")
				}
				return
			}
			if mode == "ordinary_completion" {
				if _, err := pool.Exec(ctx, `UPDATE recovery_target_state SET revision=revision+1`); err != nil {
					t.Fatal(err)
				}
				if err := fence.Complete(ctx, proof); !errors.Is(err, recovery.ErrTargetControlConflict) {
					t.Fatalf("stale revision overwrote control: %v", err)
				}
				if read().State != recovery.TargetRestoring {
					t.Fatal("stale completion changed state")
				}
				if err := recovery.ReconcileTargetCompletion(ctx, admission, proof); err != nil {
					t.Fatal(err)
				}
			} else if err := fence.Complete(ctx, proof); err != nil {
				t.Fatal(err)
			}
			complete := read()
			if complete.State != recovery.TargetComplete || *complete.CompletionEvidenceSHA256 != proof.EvidenceSHA256 {
				t.Fatal("completion lacks exact proof")
			}
			if err := recovery.ReconcileTargetCompletion(ctx, admission, proof); err != nil || read().Revision != complete.Revision {
				t.Fatalf("exact retry mutated complete state: %v", err)
			}
			wrong := proof
			wrong.EvidenceSHA256 = strings.Repeat("0", 64)
			if err := recovery.ReconcileTargetCompletion(ctx, admission, wrong); !errors.Is(err, recovery.ErrTargetIncomplete) {
				t.Fatalf("conflicting terminal accepted: %v", err)
			}
			if mode == "ordinary_completion" {
				if err := complete.RequireServing(identity.BindingDigests); err != nil {
					t.Fatal(err)
				}
				if err := recovery.ResetRestoreVerificationTarget(ctx, admission, fixture.Target, identity.TargetGenerationID, identity.BindingDigests, testExtensionBackupCatalog(t), fixture.state); !errors.Is(err, recovery.ErrTargetIncomplete) {
					t.Fatalf("ordinary reset admitted: %v", err)
				}
				return
			}
			if err := complete.RequireServing(identity.BindingDigests); !errors.Is(err, recovery.ErrTargetIncomplete) {
				t.Fatal("verification target became serving")
			}
			broken := fixture.Target
			broken.ObjectStore = resetListingFailure{Store: broken.ObjectStore}
			if err := recovery.ResetRestoreVerificationTarget(ctx, admission, broken, identity.TargetGenerationID, identity.BindingDigests, testExtensionBackupCatalog(t), fixture.state); err == nil {
				t.Fatal("injected reset failure ignored")
			}
			resetting := read()
			if resetting.State != recovery.TargetResetting || resetting.OperationID != nil || resetting.CompletionEvidenceSHA256 != nil {
				t.Fatal("reset did not invalidate old completion before mutation")
			}
			if err := recovery.ReconcileTargetCompletion(ctx, admission, proof); !errors.Is(err, recovery.ErrTargetIncomplete) {
				t.Fatal("old completion reopened resetting target")
			}
			if err := recovery.ResetRestoreVerificationTarget(ctx, admission, fixture.Target, identity.TargetGenerationID, identity.BindingDigests, testExtensionBackupCatalog(t), fixture.state); err != nil {
				t.Fatal(err)
			}
			if read().State != recovery.TargetVerificationReady {
				t.Fatal("successful reset did not publish verification readiness")
			}
			if err := recovery.ReconcileTargetCompletion(ctx, admission, proof); !errors.Is(err, recovery.ErrTargetIncomplete) {
				t.Fatal("old completion reopened reset target")
			}
		})
	}
}

type resetListingFailure struct{ objectstore.Store }

func (resetListingFailure) ListObjects(context.Context, string) ([]objectstore.ObjectInfo, error) {
	return nil, errors.New("injected reset listing failure")
}
