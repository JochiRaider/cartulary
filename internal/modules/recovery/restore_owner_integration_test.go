package recovery_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
)

func TestRestoreReadinessAndCoherentStoreOrder_Unit(t *testing.T) {
	ctx := t.Context()
	failing := newRestoreProjectionContractFixture(t, ctx, "backup_restore-u-10-02-failing", uuid.New())
	failingObserver := &restoreStepRecorder{}
	failing.Target.Projections = failingProjectionRebuilder{}
	failing.Target.Observer = failingObserver
	admitted := admitCurrentRestore(t, failing.capturedBackupContractFixture, failing.Runner, failing.Target)
	_, err := failing.Runner.RestoreBackupSet(ctx, admitted.Fence.Target(), failing.BackupSet)
	if !errors.Is(err, errProjectionRebuildBlocked) {
		t.Fatalf("projection failure = %v", err)
	}
	assertRestoreQuarantined(t, failing.Target, admitted)
	if want := []recovery.RestoreStep{recovery.RestoreStepPostgresRestore, recovery.RestoreStepObjectStoreRestore, recovery.RestoreStepProjectionRebuild}; !reflect.DeepEqual(failingObserver.Steps, want) {
		t.Fatalf("failure phase order = %v, want %v", failingObserver.Steps, want)
	}
	success := newRestoreProjectionContractFixture(t, ctx, "backup_restore-u-10-02-success", uuid.New())
	observer := &restoreStepRecorder{}
	success.Target.Observer = observer
	completed := admitCurrentRestore(t, success.capturedBackupContractFixture, success.Runner, success.Target)
	result, err := success.Runner.RestoreBackupSet(ctx, completed.Fence.Target(), success.BackupSet)
	if err != nil {
		t.Fatal(err)
	}
	if !result.ProjectionRebuildResult.ReadinessSatisfied() {
		t.Fatal("Core rebuild incomplete")
	}
	assertRestoreQuarantined(t, success.Target, completed)
	if want := []recovery.RestoreStep{recovery.RestoreStepPostgresRestore, recovery.RestoreStepObjectStoreRestore, recovery.RestoreStepProjectionRebuild, recovery.RestoreStepConsistencyCheck}; !reflect.DeepEqual(observer.Steps, want) {
		t.Fatalf("success phase order = %v, want %v", observer.Steps, want)
	}
	completed.Complete(t, result)
	control, err := recovery.ReadTargetControl(ctx, success.Target.Postgres)
	if err != nil || control.RequireServing(completed.Identity.BindingDigests) != nil {
		t.Fatal("durable completion did not admit serving", err)
	}
}

func TestMissingArtifactFailsBeforeReadinessBlocked_Integration(t *testing.T) {
	fixture := newRestoreProjectionContractFixture(t, t.Context(), "backup_restore-i-10-03", uuid.New())
	observer := &restoreStepRecorder{}
	fixture.Target.Observer = observer
	runner := recovery.NewVersionedRestoreRunner(fixture.Store, tamperedBackupStorage{Inner: fixture.BackupStorage, Missing: map[string]bool{fixture.BackupSet.ObjectStoreArtifactKey: true}}, testExtensionBackupCatalog(t), currentStateCatalog(t))
	admitted, err := tryAdmitCurrentRestore(t, fixture.capturedBackupContractFixture, runner, fixture.Target, fixture.BackupSet)
	if !errors.Is(err, os.ErrNotExist) || admitted.Fence != nil || len(observer.Steps) != 0 {
		t.Fatalf("missing artifact passed admission: %v, steps %v", err, observer.Steps)
	}
	control, err := recovery.ReadTargetControl(t.Context(), fixture.Target.Postgres)
	if err != nil || control.State != recovery.TargetInitialized {
		t.Fatal("pre-admission failure changed target", err)
	}
}

func TestFailClosedRestoreVerificationBlocked_Unit(t *testing.T) {
	ctx := t.Context()
	harness := pgtest.Start(t)
	source := newCapturedBackupContractFixture(t, ctx, harness, "backup_restore-u-10-03", uuid.New())
	for _, tc := range []struct {
		name, key string
		body      []byte
	}{
		{"missing integrity manifest", source.BackupSet.IntegrityManifestKey, nil},
		{"corrupt integrity manifest", source.BackupSet.IntegrityManifestKey, []byte(`{"schema_id":"cartulary.backup_integrity_manifest.v1"}`)},
		{"missing postgres artifact", source.BackupSet.PostgresArtifactKey, nil},
		{"missing object artifact", source.BackupSet.ObjectStoreArtifactKey, nil},
		{"corrupt postgres artifact", source.BackupSet.PostgresArtifactKey, []byte(`{"schema_id":"cartulary.postgres_snapshot_artifact.v1","tables":[]}`)},
		{"corrupt object artifact", source.BackupSet.ObjectStoreArtifactKey, []byte(`{}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newRestoreTargetContractFixture(t, ctx, harness, "tampered", source)
			storage := tamperedBackupStorage{Inner: source.BackupStorage}
			if tc.body == nil {
				storage.Missing = map[string]bool{tc.key: true}
			} else {
				storage.Replacements = map[string][]byte{tc.key: tc.body}
			}
			observer := &restoreStepRecorder{}
			rebuilder := &recordingProjectionRebuilder{inner: fixture.Target.Projections}
			fixture.Target.Projections, fixture.Target.Observer = rebuilder, observer
			runner := recovery.NewVersionedRestoreRunner(source.Store, storage, testExtensionBackupCatalog(t), source.state)
			admitted, err := tryAdmitCurrentRestore(t, source, runner, fixture.Target, source.BackupSet)
			if err == nil {
				_, err = runner.RestoreBackupSet(ctx, admitted.Fence.Target(), source.BackupSet)
			}
			if err == nil || errors.Is(err, recovery.ErrTargetIncomplete) {
				t.Fatalf("artifact rejection bypassed current admission: %v", err)
			}
			if len(rebuilder.Requests) != 0 {
				t.Fatal("invalid artifact reached rebuild")
			}
			if admitted.Fence != nil {
				assertRestoreQuarantined(t, fixture.Target, admitted)
			} else if len(observer.Steps) != 0 {
				t.Fatalf("rejected admission mutated target: %v", observer.Steps)
			}
		})
	}
	basis := recovery.RestoreVerificationBasis{
		ApplicationCryptoFormat: recovery.ApplicationCryptoFormatID, MechanismID: "backup_restore.recovery.restore.v1",
		DatabaseBindingSHA256: recovery.SHA256String("database"), ObjectStoreBindingSHA256: recovery.SHA256String("objects"),
		ReferencePackStorageBindingSHA256: recovery.SHA256String("reference-packs"), ExportOutputsBindingSHA256: recovery.SHA256String("exports"),
		BackupStorageBindingSHA256: recovery.SHA256String("backups"), RecoveryStateCatalogSHA256: source.state.DigestSHA256(), CodecRegistrySHA256: recovery.VNextCodecRegistrySHA256(),
	}
	basisSHA256, err := basis.SHA256()
	if err != nil {
		t.Fatal(err)
	}
	success := newRestoreTargetContractFixture(t, ctx, harness, "verified", source)
	admitted := prepareCurrentRestoreAdmission(t, source, success.Target)
	admitted.Identity.Purpose, admitted.Identity.Operation = "restore_verification_target", string(application.OperationRestoreVerifyLatest)
	if err := admitted.admit(t, success.Runner, success.Target, source.BackupSet); err != nil {
		t.Fatal(err)
	}
	verified, err := recovery.NewRestoreVerificationService(source.Store, success.Runner).VerifyBackupSet(ctx, recovery.RestoreVerificationTarget{RestoreTarget: admitted.Fence.Target()}, source.BackupSet, basis)
	if err != nil {
		t.Fatal("current verification failed", err)
	}
	if verified.BackupSet.VerificationState != recovery.VerificationVerified || verified.BackupSet.LastVerifiedRestoreAt == nil || verified.BackupSet.LastVerificationBasisSHA256 != basisSHA256 || verified.Run.VerificationState != recovery.VerificationVerified {
		t.Fatalf("verification metadata = %#v", verified)
	}
	admitted.Complete(t, verified.RestoreResult)
	control, err := recovery.ReadTargetControl(ctx, success.Target.Postgres)
	if err != nil || control.State != recovery.TargetComplete || control.RequireServing(admitted.Identity.BindingDigests) == nil {
		t.Fatal("verification target became an ordinary serving target", err)
	}

	failure := newRestoreProjectionContractFixture(t, ctx, "backup_restore-u-10-03-failure", uuid.New())
	failure.Target.Projections = failingProjectionRebuilder{}
	rejected := prepareCurrentRestoreAdmission(t, failure.capturedBackupContractFixture, failure.Target)
	rejected.Identity.Purpose, rejected.Identity.Operation = "restore_verification_target", string(application.OperationRestoreVerifyLatest)
	if err := rejected.admit(t, failure.Runner, failure.Target, failure.BackupSet); err != nil {
		t.Fatal(err)
	}
	_, err = recovery.NewRestoreVerificationService(failure.Store, failure.Runner).VerifyBackupSet(ctx, recovery.RestoreVerificationTarget{RestoreTarget: rejected.Fence.Target()}, failure.BackupSet, basis)
	if !errors.Is(err, errProjectionRebuildBlocked) {
		t.Fatalf("failed verification = %v", err)
	}
	failed, err := failure.Store.GetBackupSet(ctx, failure.BackupSet.BackupSetID)
	if err != nil {
		t.Fatal(err)
	}
	if failed.VerificationState != recovery.VerificationFailed || failed.LastVerifiedRestoreAt == nil || failed.LastVerificationBasisSHA256 != basisSHA256 {
		t.Fatalf("failed verification metadata = %#v", failed)
	}
	assertRestoreQuarantined(t, failure.Target, rejected)
}

func assertRestoreQuarantined(t testing.TB, target recovery.RestoreTarget, admitted currentRestoreAdmissionFixture) {
	t.Helper()
	control, err := recovery.ReadTargetControl(t.Context(), target.Postgres)
	if err != nil || control.State != recovery.TargetRestoring || control.RequireServing(admitted.Identity.BindingDigests) == nil {
		t.Fatal("incomplete restore escaped durable quarantine", err)
	}
}

type restoreStepRecorder struct {
	Steps []recovery.RestoreStep
}

func (recorder *restoreStepRecorder) RecordRestoreStep(step recovery.RestoreStep) {
	recorder.Steps = append(recorder.Steps, step)
}

var errProjectionRebuildBlocked = errors.New("projection rebuild blocked")

type failingProjectionRebuilder struct{}

func (failingProjectionRebuilder) RebuildRestoreProjections(_ context.Context, request restorecontract.ProjectionRebuildRequest) (restorecontract.ProjectionRebuildResult, error) {
	return restorecontract.ProjectionRebuildResult{
		RestoreOperationID: request.RestoreOperationID,
		Status:             restorecontract.ProjectionRebuildStatusFailed,
		ReadinessOutcome:   restorecontract.ProjectionReadinessIncomplete,
		Errors: []restorecontract.ProjectionRebuildMessage{{
			Code:    "projection_rebuild_blocked",
			Message: errProjectionRebuildBlocked.Error(),
		}},
	}, errProjectionRebuildBlocked
}
