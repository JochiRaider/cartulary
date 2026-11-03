package recovery_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

type currentRestoreAdmissionFixture struct {
	Fence               *recovery.FencedRestore
	Admission           application.TargetServingAdmission
	Identity            recovery.RestoreTargetIdentity
	Repository          application.RecoveryEvidenceRepository
	AdmissionRepository application.RecoveryEvidenceRepository
	StartedAt           time.Time
}

func openCurrentRestorePool(t testing.TB, database *pgtest.TestDatabase) postgres.AdmittedPool {
	t.Helper()
	dsn, err := database.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := postgres.Setup(t.Context(), postgres.Settings{BindingKind: "managed_service", DSN: dsn, Purpose: postgres.PurposeRecovery, ExpectedRole: "cartulary_recovery"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// This fixture composes current owner entry points. It does not manufacture
// snapshot bytes, replace restore algorithms, or publish a healthy control row.
func admitCurrentRestore(t testing.TB, source capturedBackupContractFixture, runner *recovery.RestoreRunner, target recovery.RestoreTarget) currentRestoreAdmissionFixture {
	t.Helper()
	fixture, err := tryAdmitCurrentRestore(t, source, runner, target, source.BackupSet)
	if err != nil {
		t.Fatal(err)
	}
	return fixture
}

func tryAdmitCurrentRestore(t testing.TB, source capturedBackupContractFixture, runner *recovery.RestoreRunner, target recovery.RestoreTarget, selected recovery.BackupSet) (currentRestoreAdmissionFixture, error) {
	t.Helper()
	fixture := prepareCurrentRestoreAdmission(t, source, target)
	err := fixture.admit(t, runner, target, selected)
	return fixture, err
}

func prepareCurrentRestoreAdmission(t testing.TB, source capturedBackupContractFixture, target recovery.RestoreTarget) currentRestoreAdmissionFixture {
	t.Helper()
	ctx := t.Context()
	pool, ok := target.Postgres.(postgres.AdmittedPool)
	if !ok {
		t.Fatal("current restore fixture requires an admitted Recovery pool")
	}
	admission, err := recoveryassembly.AcquireTargetServingAdmission(ctx, pool, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		settle, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := admission.Release(settle); err != nil {
			t.Error(err)
		}
	})
	generation, err := recovery.NewBackupCatalog(source.Store, source.BackupStorage, testExtensionBackupCatalog(t), source.state).RecoveryGenerationIdentity(ctx, source.BackupSet)
	if err != nil {
		t.Fatal(err)
	}
	identity := recovery.RestoreTargetIdentity{TargetGenerationID: uuid.New(), Purpose: "restore_target", Operation: "restore_latest", OperationID: uuid.New(), BackupSetID: source.BackupSet.BackupSetID, RecoveryStateCatalogSHA256: generation.RecoveryStateCatalogSHA256, CodecRegistrySHA256: generation.CodecRegistrySHA256, BindingDigests: recovery.TargetBindingDigests{DatabaseSHA256: strings.Repeat("1", 64), ObjectStoreSHA256: strings.Repeat("2", 64), ReferencePackStorageSHA256: strings.Repeat("3", 64), ExportOutputsSHA256: strings.Repeat("4", 64)}}
	key, err := recovery.ParseRecoveryEncryptionKey("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")
	if err != nil {
		t.Fatal(err)
	}
	repository, err := recoveryassembly.NewRecoveryEvidenceRepository(source.sourcePool, func() (recovery.RecoveryEncryptionKey, error) { return key, nil })
	if err != nil {
		t.Fatal(err)
	}
	targetRepository, err := recoveryassembly.NewRecoveryEvidenceRepository(pool, func() (recovery.RecoveryEncryptionKey, error) { return key, nil })
	if err != nil {
		t.Fatal(err)
	}
	return currentRestoreAdmissionFixture{Admission: admission, Identity: identity, Repository: targetRepository, AdmissionRepository: repository, StartedAt: time.Now().UTC()}
}

func (fixture *currentRestoreAdmissionFixture) admit(t testing.TB, runner *recovery.RestoreRunner, target recovery.RestoreTarget, selected recovery.BackupSet) error {
	t.Helper()
	identity := fixture.Identity
	fence, err := runner.AdmitRestore(t.Context(), fixture.Admission, target, selected, identity, func(work context.Context) error {
		return fixture.AdmissionRepository.AppendAdmission(work, application.RecoveryAdmissionRecord{OperationID: identity.OperationID, Operation: application.Operation(identity.Operation), StartedAt: fixture.StartedAt, BackupSetID: &identity.BackupSetID, ArtifactKinds: []string{}})
	})
	if err != nil {
		return err
	}
	fixture.Fence = fence
	t.Cleanup(func() {
		if err := fence.Close(); err != nil {
			t.Error(err)
		}
	})
	return nil
}

func (fixture currentRestoreAdmissionFixture) Complete(t testing.TB, result recovery.RestoreResult) {
	t.Helper()
	point := result.BackupSet.ConsistencyPointAt
	record := application.RecoveryCompletionRecord{OperationID: fixture.Identity.OperationID, Operation: application.Operation(fixture.Identity.Operation), StartedAt: fixture.StartedAt, CompletedAt: time.Now().UTC(), Result: application.ResultSucceeded, BackupSetID: &fixture.Identity.BackupSetID, ConsistencyPointAt: &point, ArtifactCounts: []application.ArtifactCount{}, RestoreGeneration: result.RestoreGeneration, TargetBindings: &fixture.Identity.BindingDigests}
	if err := fixture.Repository.AppendCompletion(t.Context(), record); err != nil {
		t.Fatal(err)
	}
	body, err := application.EncodeRecoveryCompletion(record)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(body)
	if err := fixture.Fence.Complete(t.Context(), recovery.TargetCompletionProof{Identity: fixture.Identity, EvidenceSHA256: hex.EncodeToString(digest[:])}); err != nil {
		t.Fatal(err)
	}
}
