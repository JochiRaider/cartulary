package recovery_test

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/testutil/recoverytest"
	verificationtest "github.com/JochiRaider/cartulary/internal/testutil/referencetest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRestoreProjectionRebuildReceivesStructuredRequest(t *testing.T) {
	ctx := context.Background()
	fixture := newRestoreProjectionContractFixture(t, ctx, "backup_restore-restore-projection-request", uuid.MustParse("00000000-0000-0000-0000-000000104101"))
	rebuilder := &recordingProjectionRebuilder{inner: fixture.Target.Projections}
	fixture.Target.Projections = rebuilder

	admitted := admitCurrentRestore(t, fixture.capturedBackupContractFixture, fixture.Runner, fixture.Target)
	result, err := fixture.Runner.RestoreBackupSet(ctx, admitted.Fence.Target(), fixture.BackupSet)
	if err != nil {
		t.Fatalf("restore backup set with structured projection request: %v", err)
	}
	if len(rebuilder.Requests) != 1 {
		t.Fatalf("projection rebuilder request count got %d want 1", len(rebuilder.Requests))
	}
	request := rebuilder.Requests[0]
	if request.RestoreOperationID != admitted.Identity.OperationID {
		t.Fatalf("restore operation id differs from the admitted identity")
	}
	if request.RebuildScope != restorecontract.ProjectionRebuildScopeAllActiveProviders {
		t.Fatalf("rebuild scope got %q want %q", request.RebuildScope, restorecontract.ProjectionRebuildScopeAllActiveProviders)
	}
	if request.ProviderRegistryRef != restorecontract.ProviderRegistryRefCodeBacked {
		t.Fatalf("provider registry ref got %q want %q", request.ProviderRegistryRef, restorecontract.ProviderRegistryRefCodeBacked)
	}
	if !strings.Contains(request.RestoredSourceStateRef, fixture.BackupSet.BackupSetID.String()) ||
		!strings.Contains(request.RestoredSourceStateRef, fixture.BackupSet.PostgresArtifactSHA256) {
		t.Fatalf("source state ref %q does not identify backup set and postgres artifact", request.RestoredSourceStateRef)
	}
	if result.ProjectionRebuildResult.RestoreOperationID != request.RestoreOperationID ||
		!result.ProjectionRebuildResult.ReadinessSatisfied() {
		t.Fatalf("restore result did not preserve ready projection rebuild result: %#v", result.ProjectionRebuildResult)
	}
	control, err := recovery.ReadTargetControl(ctx, fixture.Target.Postgres)
	if err != nil || control.State != recovery.TargetRestoring || control.RequireServing(admitted.Identity.BindingDigests) == nil {
		t.Fatal("structural payload alone opened serving", err)
	}
	admitted.Complete(t, result)
	control, err = recovery.ReadTargetControl(ctx, fixture.Target.Postgres)
	if err != nil || control.RequireServing(admitted.Identity.BindingDigests) != nil {
		t.Fatal("durable completion did not admit serving", err)
	}

}

func TestRestoreProjectionRebuildReadinessFailsClosed(t *testing.T) {
	ctx := context.Background()
	fixture := newRestoreProjectionContractFixture(t, ctx, "backup_restore-restore-projection-fail-closed", uuid.MustParse("00000000-0000-0000-0000-000000104102"))
	rebuilder := &recordingProjectionRebuilder{
		respond: func(request restorecontract.ProjectionRebuildRequest) restorecontract.ProjectionRebuildResult {
			return restorecontract.ProjectionRebuildResult{
				RestoreOperationID: request.RestoreOperationID,
				Status:             restorecontract.ProjectionRebuildStatusFailed,
				ReadinessOutcome:   restorecontract.ProjectionReadinessIncomplete,
				Errors: []restorecontract.ProjectionRebuildMessage{{
					Code:    "test_projection_failure",
					Message: "projection rebuild did not complete",
				}},
			}
		},
	}
	fixture.Target.Projections = rebuilder

	admitted := admitCurrentRestore(t, fixture.capturedBackupContractFixture, fixture.Runner, fixture.Target)
	result, err := fixture.Runner.RestoreBackupSet(ctx, admitted.Fence.Target(), fixture.BackupSet)
	if err == nil || !strings.Contains(err.Error(), "independent Core rebuild did not complete") {
		t.Fatalf("restore error got %v want fail-closed projection readiness error", err)
	}
	if result.ProjectionRebuildResult.Status != restorecontract.ProjectionRebuildStatusFailed ||
		result.ProjectionRebuildResult.ReadinessOutcome != restorecontract.ProjectionReadinessIncomplete {
		t.Fatalf("partial restore result did not preserve failed projection metadata: %#v", result.ProjectionRebuildResult)
	}
	control, controlErr := recovery.ReadTargetControl(ctx, fixture.Target.Postgres)
	if controlErr != nil || control.State != recovery.TargetRestoring || control.RequireServing(admitted.Identity.BindingDigests) == nil {
		t.Fatal("failed Core rebuild escaped durable quarantine", controlErr)
	}

}

type restoreProjectionContractFixture struct {
	capturedBackupContractFixture
	targetDatabase *pgtest.TestDatabase
	Runner         *recovery.RestoreRunner
	Target         recovery.RestoreTarget
}

func newRestoreProjectionContractFixture(t *testing.T, ctx context.Context, prefix string, backupSetID uuid.UUID) restoreProjectionContractFixture {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	captured := newCapturedBackupContractFixture(t, ctx, postgresHarness, prefix, backupSetID)
	return newRestoreTargetContractFixture(t, ctx, postgresHarness, prefix, captured)
}

func newRestoreTargetContractFixture(t *testing.T, ctx context.Context, postgresHarness *pgtest.Harness, prefix string, captured capturedBackupContractFixture) restoreProjectionContractFixture {
	t.Helper()
	targetDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix+"-target")
	recoveryDSN, err := targetDB.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	targetPool, err := postgres.Setup(ctx, postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          recoveryDSN,
		Purpose:      postgres.PurposeRecovery,
		ExpectedRole: "cartulary_recovery",
	})
	if err != nil {
		t.Fatalf("open target postgres fixture: %v", err)
	}
	t.Cleanup(targetPool.Close)

	targetObjectStore, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("create target object store fixture: %v", err)
	}
	t.Cleanup(func() {
		_ = targetObjectStore.Close()
	})

	targetPacks, err := referenceassembly.NewRecoveryStorage(t.TempDir(), t.TempDir(), reference_data.DefaultLimits(), verificationtest.Lifetime(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(targetPacks.Close)

	core, err := projectionassembly.Build(targetPool)
	if err != nil {
		t.Fatal(err)
	}
	return restoreProjectionContractFixture{
		capturedBackupContractFixture: captured,
		targetDatabase:                targetDB,
		Runner:                        recovery.NewVersionedRestoreRunner(captured.Store, captured.BackupStorage, testExtensionBackupCatalog(t), captured.state),
		Target: recovery.RestoreTarget{ProofScratch: recoverytest.ProofScratch(t),
			RestoreOperationID: uuid.New(), TargetGenerationID: uuid.New(),
			ReferencePacks: targetPacks, Postgres: targetPool,
			ObjectStore:     targetObjectStore,
			EvidenceObjects: evidence.NewRecoveryProvider(targetPool),
			Projections:     core.RecoveryPorts().Rebuilder,
		},
	}
}

type recordingProjectionRebuilder struct {
	inner    restorecontract.ProjectionRebuilder
	Requests []restorecontract.ProjectionRebuildRequest
	respond  func(restorecontract.ProjectionRebuildRequest) restorecontract.ProjectionRebuildResult
	err      error
}

func (rebuilder *recordingProjectionRebuilder) RebuildRestoreProjections(ctx context.Context, request restorecontract.ProjectionRebuildRequest) (restorecontract.ProjectionRebuildResult, error) {
	rebuilder.Requests = append(rebuilder.Requests, request)
	if rebuilder.inner != nil {
		return rebuilder.inner.RebuildRestoreProjections(ctx, request)
	}
	if rebuilder.respond != nil {
		return rebuilder.respond(request), rebuilder.err
	}
	return readyProjectionRebuildResult(request), rebuilder.err
}

func readyProjectionRebuildResult(request restorecontract.ProjectionRebuildRequest) restorecontract.ProjectionRebuildResult {
	return restorecontract.ProjectionRebuildResult{
		RestoreOperationID: request.RestoreOperationID,
		Status:             restorecontract.ProjectionRebuildStatusSucceeded,
		ReadinessOutcome:   restorecontract.ProjectionReadinessReady,
		ProviderResults: []restorecontract.ProjectionProviderResult{{
			ProviderKey:             "test_projection_provider",
			Status:                  restorecontract.ProjectionProviderResultSucceeded,
			RebuiltViewSchemaIDs:    []string{"cartulary.view.timeline.v2"},
			RebuiltProjectionTables: []restorecontract.ProjectionTableResult{{ProjectionTableID: "timeline_grid_projection", RowCount: 1}},
		}},
	}
}
