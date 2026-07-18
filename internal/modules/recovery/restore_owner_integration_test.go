package recovery_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
)

func TestRestoreReadinessAndCoherentStoreOrder_Unit(t *testing.T) {
	ctx := context.Background()
	failing := newRestoreProjectionContractFixture(t, ctx, "backup_restore-u-10-02-failing", uuid.MustParse("00000000-0000-0000-0000-000000102001"))
	failingReadiness := &recordingRestoreReadinessGate{}
	failingObserver := &restoreStepRecorder{}
	failing.Target.Projections = failingProjectionRebuilder{}
	failing.Target.Readiness = failingReadiness
	failing.Target.Observer = failingObserver
	_, err := failing.Runner.RestoreBackupSet(ctx, failing.Target, failing.BackupSet)
	if !errors.Is(err, errProjectionRebuildBlocked) {
		t.Fatalf("projection failure got %v want %v", err, errProjectionRebuildBlocked)
	}
	if failingReadiness.Calls != 0 {
		t.Fatalf("readiness marked after projection failure: calls=%d", failingReadiness.Calls)
	}
	if got, want := failingObserver.Steps, []recovery.RestoreStep{
		recovery.RestoreStepPostgresRestore,
		recovery.RestoreStepObjectStoreRestore,
		recovery.RestoreStepProjectionRebuild,
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("failing restore steps got %v want %v", got, want)
	}

	success := newRestoreProjectionContractFixture(t, ctx, "backup_restore-u-10-02-success", uuid.MustParse("00000000-0000-0000-0000-000000102002"))
	readiness := &recordingRestoreReadinessGate{}
	observer := &restoreStepRecorder{}
	success.Target.Projections = &recordingProjectionRebuilder{}
	success.Target.Readiness = readiness
	success.Target.Observer = observer
	result, err := success.Runner.RestoreBackupSet(ctx, success.Target, success.BackupSet)
	if err != nil {
		t.Fatalf("restore coherent backup: %v", err)
	}
	if readiness.Calls != 1 || !result.ProjectionRebuildResult.ReadinessSatisfied() {
		t.Fatalf("successful restore did not mark readiness after projection completion: readiness=%#v result=%#v", readiness, result)
	}
	if got, want := observer.Steps, []recovery.RestoreStep{
		recovery.RestoreStepPostgresRestore,
		recovery.RestoreStepObjectStoreRestore,
		recovery.RestoreStepProjectionRebuild,
		recovery.RestoreStepConsistencyCheck,
		recovery.RestoreStepReadiness,
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("restore steps got %v want %v", got, want)
	}
}

func TestMissingArtifactFailsBeforeReadinessBlocked_Integration(t *testing.T) {
	ctx := context.Background()
	fixture := newRestoreProjectionContractFixture(t, ctx, "backup_restore-i-10-03", uuid.MustParse("00000000-0000-0000-0000-000000103003"))
	readiness := &recordingRestoreReadinessGate{}
	observer := &restoreStepRecorder{}
	fixture.Target.Projections = &recordingProjectionRebuilder{}
	fixture.Target.Readiness = readiness
	fixture.Target.Observer = observer

	runner := recovery.NewRestoreRunner(fixture.Store, tamperedBackupStorage{
		Inner:   fixture.BackupStorage,
		Missing: map[string]bool{fixture.BackupSet.ObjectStoreArtifactKey: true},
	})
	_, err := runner.RestoreBackupSet(ctx, fixture.Target, fixture.BackupSet)
	if err == nil {
		t.Fatal("restore with missing object artifact unexpectedly succeeded")
	}
	if readiness.Calls != 0 {
		t.Fatalf("restore with missing artifact marked readiness: calls=%d", readiness.Calls)
	}
	if len(observer.Steps) != 0 {
		t.Fatalf("missing artifact reached target mutation steps: %v", observer.Steps)
	}
}

func TestFailClosedRestoreVerificationBlocked_Unit(t *testing.T) {
	ctx := context.Background()
	fixture := newRestoreProjectionContractFixture(t, ctx, "backup_restore-u-10-03", uuid.MustParse("00000000-0000-0000-0000-000000103001"))

	for _, tc := range []struct {
		name    string
		storage recovery.BackupStorage
	}{
		{
			name: "missing integrity manifest",
			storage: tamperedBackupStorage{
				Inner:   fixture.BackupStorage,
				Missing: map[string]bool{fixture.BackupSet.IntegrityManifestKey: true},
			},
		},
		{
			name: "corrupt integrity manifest",
			storage: tamperedBackupStorage{
				Inner:        fixture.BackupStorage,
				Replacements: map[string][]byte{fixture.BackupSet.IntegrityManifestKey: []byte(`{"schema_id":"cartulary.backup_integrity_manifest.v1"}`)},
			},
		},
		{
			name: "missing postgres artifact",
			storage: tamperedBackupStorage{
				Inner:   fixture.BackupStorage,
				Missing: map[string]bool{fixture.BackupSet.PostgresArtifactKey: true},
			},
		},
		{
			name: "missing object artifact",
			storage: tamperedBackupStorage{
				Inner:   fixture.BackupStorage,
				Missing: map[string]bool{fixture.BackupSet.ObjectStoreArtifactKey: true},
			},
		},
		{
			name: "corrupt postgres artifact",
			storage: tamperedBackupStorage{
				Inner:        fixture.BackupStorage,
				Replacements: map[string][]byte{fixture.BackupSet.PostgresArtifactKey: []byte(`{"schema_id":"cartulary.postgres_snapshot_artifact.v1","tables":[]}`)},
			},
		},
		{
			name: "corrupt object artifact",
			storage: tamperedBackupStorage{
				Inner:        fixture.BackupStorage,
				Replacements: map[string][]byte{fixture.BackupSet.ObjectStoreArtifactKey: []byte(`{}`)},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			readiness := &recordingRestoreReadinessGate{}
			observer := &restoreStepRecorder{}
			target := fixture.Target
			target.Projections = &recordingProjectionRebuilder{}
			target.Readiness = readiness
			target.Observer = observer
			_, err := recovery.NewRestoreRunner(fixture.Store, tc.storage).RestoreBackupSet(ctx, target, fixture.BackupSet)
			if err == nil {
				t.Fatalf("tampered backup %q unexpectedly restored", tc.name)
			}
			if readiness.Calls != 0 {
				t.Fatalf("tampered backup %q marked readiness", tc.name)
			}
			for _, step := range observer.Steps {
				if step == recovery.RestoreStepReadiness {
					t.Fatalf("tampered backup %q reached readiness: %v", tc.name, observer.Steps)
				}
			}
		})
	}

	basis, err := recovery.RestoreVerificationBasisSHA256(map[string]string{
		"backup_mechanism": "backup_restore.recovery.restore.v1",
		"fixture":          "backup_restore-u-10-03",
	})
	if err != nil {
		t.Fatalf("restore verification basis: %v", err)
	}
	successTarget := fixture.Target
	successTarget.Projections = &recordingProjectionRebuilder{}
	verified, err := recovery.NewRestoreVerificationService(fixture.Store, fixture.Runner).VerifyBackupSet(ctx, recovery.RestoreVerificationTarget{
		RestoreTarget: successTarget,
	}, fixture.BackupSet, basis)
	if err != nil {
		t.Fatalf("successful restore verification: %v", err)
	}
	if verified.BackupSet.VerificationState != recovery.VerificationVerified ||
		verified.BackupSet.LastVerifiedRestoreAt == nil ||
		verified.BackupSet.LastVerificationBasisSHA256 != basis ||
		verified.Run.VerificationState != recovery.VerificationVerified {
		t.Fatalf("successful verification state got %#v", verified)
	}

	failureFixture := newRestoreProjectionContractFixture(t, ctx, "backup_restore-u-10-03-failure", uuid.MustParse("00000000-0000-0000-0000-000000103002"))
	failureTarget := failureFixture.Target
	failureTarget.Projections = failingProjectionRebuilder{}
	_, err = recovery.NewRestoreVerificationService(failureFixture.Store, failureFixture.Runner).VerifyBackupSet(ctx, recovery.RestoreVerificationTarget{
		RestoreTarget: failureTarget,
	}, failureFixture.BackupSet, basis)
	if !errors.Is(err, errProjectionRebuildBlocked) {
		t.Fatalf("failed restore verification got %v want %v", err, errProjectionRebuildBlocked)
	}
	failed, err := failureFixture.Store.GetBackupSet(ctx, failureFixture.BackupSet.BackupSetID)
	if err != nil {
		t.Fatalf("reload failed restore verification state: %v", err)
	}
	if failed.VerificationState != recovery.VerificationFailed ||
		failed.LastVerifiedRestoreAt == nil ||
		failed.LastVerificationBasisSHA256 != basis {
		t.Fatalf("failed verification state got %#v", failed)
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
