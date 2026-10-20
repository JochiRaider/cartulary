package recovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

// Capturing and selecting a backup does not require a restore destination.
// Restore fixtures compose this source with their independently owned target.
type capturedBackupContractFixture struct {
	Capture       *recovery.VNextCaptureService
	Store         *recovery.Store
	BackupStorage recovery.BackupStorage
	BackupSet     recovery.BackupSet
	AsOf          time.Time
	state         *recoverystate.Catalog
}

func newCapturedBackupContractFixture(t *testing.T, ctx context.Context, postgresHarness *pgtest.Harness, prefix string, backupSetID uuid.UUID) capturedBackupContractFixture {
	t.Helper()

	sourceDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix+"-source")
	sourcePool, err := pgxpool.New(ctx, sourceDB.DSN)
	if err != nil {
		t.Fatalf("open source postgres fixture: %v", err)
	}
	t.Cleanup(sourcePool.Close)

	sourceObjectStore, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("create source object store fixture: %v", err)
	}
	t.Cleanup(func() {
		_ = sourceObjectStore.Close()
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

	return capturedBackupContractFixture{
		Capture: capture, Store: sourceStore, BackupStorage: backupStorage,
		BackupSet: backupSet, AsOf: asOf, state: state,
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
