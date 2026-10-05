package recovery_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRestoreProjectionRebuildReceivesStructuredRequest(t *testing.T) {
	ctx := context.Background()
	fixture := newRestoreProjectionContractFixture(t, ctx, "backup_restore-restore-projection-request", uuid.MustParse("00000000-0000-0000-0000-000000104101"))
	rebuilder := &recordingProjectionRebuilder{
		respond: func(request restorecontract.ProjectionRebuildRequest) restorecontract.ProjectionRebuildResult {
			return readyProjectionRebuildResult(request)
		},
	}
	readiness := &recordingRestoreReadinessGate{}
	fixture.Target.Projections = rebuilder
	fixture.Target.Readiness = readiness

	result, err := fixture.Runner.RestoreBackupSet(ctx, fixture.Target, fixture.BackupSet)
	if err != nil {
		t.Fatalf("restore backup set with structured projection request: %v", err)
	}
	if len(rebuilder.Requests) != 1 {
		t.Fatalf("projection rebuilder request count got %d want 1", len(rebuilder.Requests))
	}
	request := rebuilder.Requests[0]
	if request.RestoreOperationID != fixture.Target.RestoreOperationID {
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
	if readiness.Calls != 1 || readiness.Results[0].ProjectionRebuildResult.RestoreOperationID != request.RestoreOperationID {
		t.Fatalf("readiness did not receive restore result with projection metadata: %#v", readiness)
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
	readiness := &recordingRestoreReadinessGate{}
	fixture.Target.Projections = rebuilder
	fixture.Target.Readiness = readiness

	result, err := fixture.Runner.RestoreBackupSet(ctx, fixture.Target, fixture.BackupSet)
	if err == nil || !strings.Contains(err.Error(), "projection rebuild did not produce ready restore state") {
		t.Fatalf("restore error got %v want fail-closed projection readiness error", err)
	}
	if result.ProjectionRebuildResult.Status != restorecontract.ProjectionRebuildStatusFailed ||
		result.ProjectionRebuildResult.ReadinessOutcome != restorecontract.ProjectionReadinessIncomplete {
		t.Fatalf("partial restore result did not preserve failed projection metadata: %#v", result.ProjectionRebuildResult)
	}
	if readiness.Calls != 0 {
		t.Fatalf("readiness should not be marked after incomplete projection rebuild, calls=%d", readiness.Calls)
	}
}

type restoreProjectionContractFixture struct {
	Runner        *recovery.RestoreRunner
	Capture       *recovery.VNextCaptureService
	Store         *recovery.Store
	BackupStorage recovery.BackupStorage
	BackupSet     recovery.BackupSet
	Target        recovery.RestoreTarget
	AsOf          time.Time
}

func newRestoreProjectionContractFixture(t *testing.T, ctx context.Context, prefix string, backupSetID uuid.UUID) restoreProjectionContractFixture {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	sourceDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix+"-source")
	sourcePool, err := pgxpool.New(ctx, sourceDB.DSN)
	if err != nil {
		t.Fatalf("open source postgres fixture: %v", err)
	}
	t.Cleanup(sourcePool.Close)
	targetDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix+"-target")
	targetPool, err := postgres.Setup(ctx, postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          targetDB.DSN,
		Purpose:      postgres.PurposeRecovery,
		ExpectedRole: "cartulary_recovery",
	})
	if err != nil {
		t.Fatalf("open target postgres fixture: %v", err)
	}
	t.Cleanup(targetPool.Close)

	sourceObjectStore, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("create source object store fixture: %v", err)
	}
	t.Cleanup(func() {
		_ = sourceObjectStore.Close()
	})
	targetObjectStore, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("create target object store fixture: %v", err)
	}
	t.Cleanup(func() {
		_ = targetObjectStore.Close()
	})

	sourceTemporary, sourcePublished := t.TempDir(), t.TempDir()
	sourceLive, err := referenceassembly.NewRootStorage(sourceTemporary, sourcePublished)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(sourceLive.Close)
	if err := reference_data.ReconcileBaseRelease(ctx, sourcePool, sourceLive, reference_data.BaseReleaseOptions{Limits: reference_data.DefaultLimits()}, time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	sourcePacks, err := referenceassembly.NewRecoveryStorage(sourceTemporary, sourcePublished, reference_data.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(sourcePacks.Close)
	targetPacks, err := referenceassembly.NewRecoveryStorage(t.TempDir(), t.TempDir(), reference_data.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(targetPacks.Close)
	graph, err := recoveryassembly.NewGraphProjectionRestoreParticipant(targetPool)
	if err != nil {
		t.Fatal(err)
	}
	sourceStore := recovery.NewStore(sourcePool)
	backupStorage := newEncryptedBackupStorage(t, t.TempDir())
	streaming, err := recovery.RequireStreamingBackupStorage(backupStorage)
	if err != nil {
		t.Fatal(err)
	}
	state := currentStateCatalog(t)
	inventory, err := recoveryassembly.CurrentVNextObjectInventoryCatalog(recoveryassembly.NewVNextObjectSource(sourceObjectStore), sourcePacks, nil)
	if err != nil {
		t.Fatal(err)
	}
	capture, err := recovery.NewVNextCaptureService(recoveryassembly.NewVNextSnapshotRepository(sourcePool), streaming, state, inventory)
	if err != nil {
		t.Fatal(err)
	}
	asOf := time.Date(2026, 7, 1, 14, 0, 0, 0, time.UTC)
	captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{
		BackupSetID:        backupSetID,
		ConsistencyPointAt: asOf.Add(-time.Hour),
		CreatedAt:          asOf,
		RetainedUntil:      asOf.Add(31 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	backupSet, err := sourceStore.PublishVNextCapturedBackup(ctx, captured)
	if err != nil {
		t.Fatalf("capture backup set fixture: %v", err)
	}

	return restoreProjectionContractFixture{
		Runner:        recovery.NewVersionedRestoreRunner(sourceStore, backupStorage, testExtensionBackupCatalog(t), state),
		Capture:       capture,
		Store:         sourceStore,
		BackupStorage: backupStorage,
		BackupSet:     backupSet,
		AsOf:          asOf,
		Target: recovery.RestoreTarget{
			RestoreOperationID: uuid.New(), TargetGenerationID: uuid.New(),
			ReferencePacks: targetPacks, GraphProjection: graph,
			Postgres:        targetPool,
			ObjectStore:     targetObjectStore,
			EvidenceObjects: evidence.NewRecoveryProvider(targetPool),
		},
	}
}

func currentStateCatalog(t testing.TB) *recoverystate.Catalog {
	t.Helper()
	catalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

type recordingProjectionRebuilder struct {
	Requests []restorecontract.ProjectionRebuildRequest
	respond  func(restorecontract.ProjectionRebuildRequest) restorecontract.ProjectionRebuildResult
	err      error
}

func (rebuilder *recordingProjectionRebuilder) RebuildRestoreProjections(ctx context.Context, request restorecontract.ProjectionRebuildRequest) (restorecontract.ProjectionRebuildResult, error) {
	rebuilder.Requests = append(rebuilder.Requests, request)
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

type recordingRestoreReadinessGate struct {
	Calls   int
	Results []recovery.RestoreResult
}

func (gate *recordingRestoreReadinessGate) MarkRestoreReady(ctx context.Context, result recovery.RestoreResult) error {
	gate.Calls++
	gate.Results = append(gate.Results, result)
	return nil
}
