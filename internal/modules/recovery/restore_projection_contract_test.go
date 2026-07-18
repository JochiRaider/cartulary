package recovery_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRestoreProjectionRebuildReceivesStructuredRequest(t *testing.T) {
	ctx := context.Background()
	fixture := newRestoreProjectionContractFixture(t, ctx, "phase10-restore-projection-request", uuid.MustParse("00000000-0000-0000-0000-000000104101"))
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
	if request.RestoreOperationID == uuid.Nil {
		t.Fatalf("restore operation id was not generated")
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
	fixture := newRestoreProjectionContractFixture(t, ctx, "phase10-restore-projection-fail-closed", uuid.MustParse("00000000-0000-0000-0000-000000104102"))
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
	Store         *recovery.Store
	BackupStorage recovery.BackupStorage
	BackupSet     recovery.BackupSet
	Target        recovery.RestoreTarget
	AsOf          time.Time
}

func newRestoreProjectionContractFixture(t *testing.T, ctx context.Context, prefix string, backupSetID uuid.UUID) restoreProjectionContractFixture {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	sourceDB := postgresHarness.PrepareGroupDatabaseT(t, prefix+"-source", prefix+"-source")
	sourcePool, err := pgxpool.New(ctx, sourceDB.DSN)
	if err != nil {
		t.Fatalf("open source postgres fixture: %v", err)
	}
	t.Cleanup(sourcePool.Close)
	targetDB := postgresHarness.PrepareGroupDatabaseT(t, prefix+"-target", prefix+"-target")
	targetPool, err := pgxpool.New(ctx, targetDB.DSN)
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

	postgresBody, err := recovery.CapturePostgresSnapshotArtifact(ctx, sourcePool)
	if err != nil {
		t.Fatalf("capture postgres snapshot fixture: %v", err)
	}
	objectBody, err := recovery.CaptureObjectStoreSnapshotArtifact(ctx, sourceObjectStore, "")
	if err != nil {
		t.Fatalf("capture object-store snapshot fixture: %v", err)
	}

	sourceStore := recovery.NewStore(sourcePool)
	backupStorage := newEncryptedBackupStorage(t, t.TempDir())
	capture := recovery.NewCaptureService(sourceStore, backupStorage)
	asOf := time.Date(2026, 7, 1, 14, 0, 0, 0, time.UTC)
	backupSet, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        backupSetID,
		ConsistencyPointAt: asOf.Add(-time.Hour),
		CreatedAt:          asOf,
		RetainedUntil:      asOf.Add(31 * 24 * time.Hour),
		PostgresArtifact: recovery.BackupArtifact{
			Body:        postgresBody,
			ContentType: "application/json",
		},
		ObjectStoreArtifact: recovery.BackupArtifact{
			Body:        objectBody,
			ContentType: "application/json",
		},
	}))
	if err != nil {
		t.Fatalf("capture backup set fixture: %v", err)
	}

	return restoreProjectionContractFixture{
		Runner:        recovery.NewRestoreRunner(sourceStore, backupStorage),
		Store:         sourceStore,
		BackupStorage: backupStorage,
		BackupSet:     backupSet,
		AsOf:          asOf,
		Target: recovery.RestoreTarget{
			Postgres:    targetPool,
			ObjectStore: targetObjectStore,
		},
	}
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
