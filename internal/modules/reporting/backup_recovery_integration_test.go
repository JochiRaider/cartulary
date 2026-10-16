package reporting_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/google/uuid"
)

// The preview owner path supplies actual completed release and preview Jobs.
// Both databases and filesystem roots are disposable harness-owned fixtures.
func assertReportingPrivateBackupRestore(t *testing.T, runtime *appsupport.Runtime, source *appsupport.ServerHarness, releaseID, previewID string) {
	t.Helper()
	ctx := context.Background()
	sourceDSN, err := source.Database.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	sourceConnection, err := postgres.Setup(ctx, postgres.Settings{BindingKind: "managed_service", DSN: sourceDSN, Purpose: postgres.PurposeRecovery, ExpectedRole: "cartulary_recovery"})
	if err != nil {
		t.Fatal(err)
	}
	defer sourceConnection.Close()
	sourcePool := sourceConnection.Pool()
	sourcePacks, err := referenceassembly.NewRecoveryStorage(source.Server.Config.Roots.TemporaryWork.Path, source.Server.Config.Roots.ReferencePackStorage.Path, reference_data.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	defer sourcePacks.Close()
	rawBackup, err := recoveryassembly.NewFilesystemStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	key, err := recovery.ParseRecoveryEncryptionKey("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")
	if err != nil {
		t.Fatal(err)
	}
	backup, err := recovery.NewEncryptedBackupStorage(rawBackup, key)
	if err != nil {
		t.Fatal(err)
	}
	streaming, err := recovery.RequireStreamingBackupStorage(backup)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	inventories, err := recoveryassembly.CurrentVNextObjectInventoryCatalog(recoveryassembly.NewVNextObjectSource(source.ObjectStore), sourcePacks, nil)
	if err != nil {
		t.Fatal(err)
	}
	capture, err := recovery.NewVNextCaptureService(recoveryassembly.NewVNextSnapshotRepository(sourcePool), streaming, catalog, inventories)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Now().UTC()
	captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{BackupSetID: uuid.New(), ConsistencyPointAt: at, CreatedAt: at, RetainedUntil: at.Add(31 * 24 * time.Hour)})
	if err != nil {
		t.Fatal("capture Reporting history", err)
	}
	store := recovery.NewStore(sourcePool)
	saved, err := store.PublishVNextCapturedBackup(ctx, captured)
	if err != nil {
		t.Fatal(err)
	}
	targetDB := runtime.PrepareIsolatedDatabase(t, "reporting-private-artifact-restore")
	targetDSN, err := targetDB.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	targetConnection, err := postgres.Setup(ctx, postgres.Settings{BindingKind: "managed_service", DSN: targetDSN, Purpose: postgres.PurposeRecovery, ExpectedRole: "cartulary_recovery"})
	if err != nil {
		t.Fatal(err)
	}
	defer targetConnection.Close()
	targetPool := targetConnection.Pool()
	temporaryRoot, publishedRoot := t.TempDir(), t.TempDir()
	targetPacks, err := referenceassembly.NewRecoveryStorage(temporaryRoot, publishedRoot, reference_data.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	defer targetPacks.Close()
	targetObjects, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer targetObjects.Close()
	extensions, err := extensionassembly.GeneratedRecoveryCatalog()
	if err != nil {
		t.Fatal(err)
	}
	graph, err := recoveryassembly.NewGraphProjectionRestoreParticipant(targetPool)
	if err != nil {
		t.Fatal(err)
	}
	projections, _, err := projectionassembly.NewRecoveryServices(targetPool)
	if err != nil {
		t.Fatal(err)
	}
	_, err = recovery.NewVersionedRestoreRunner(store, backup, extensions, catalog).RestoreBackupSet(ctx, recovery.RestoreTarget{RestoreOperationID: uuid.New(), TargetGenerationID: uuid.New(), Postgres: targetPool, ObjectStore: targetObjects, ReferencePacks: targetPacks, EvidenceObjects: evidence.NewRecoveryProvider(targetPool), GraphProjection: graph, Projections: projections}, saved)
	if err != nil {
		t.Fatal("restore Reporting history", err)
	}
	if err := reporting.ValidateRetainedSensitiveArtifacts(ctx, targetPool); err != nil {
		t.Fatal("restored private artifacts invalid", err)
	}
	restoredStorage, err := referenceassembly.NewRootStorage(temporaryRoot, publishedRoot)
	if err != nil {
		t.Fatal(err)
	}
	defer restoredStorage.Close()
	finalizer, err := extensionassembly.NewReferencePackMutationFinalizer(func(err error) { t.Errorf("unexpected reference mutation: %v", err) })
	if err != nil {
		t.Fatal(err)
	}
	retained, err := reference_data.NewRetention(targetPool, restoredStorage, time.Now, reference_data.IntegrityOptions{Finalizer: finalizer, Limits: reference_data.DefaultLimits()})
	if err != nil {
		t.Fatal(err)
	}
	if err := reporting.ValidateRetainedSnapshotModels(ctx, targetPool, retained); err != nil {
		t.Fatal("restored snapshot admission invalid", err)
	}
	if err := reporting.ValidateRetainedReleases(ctx, targetPool, retained); err != nil {
		t.Fatal("restored release admission invalid", err)
	}
	if err := reporting.ValidateRetainedPreviews(ctx, targetPool, retained); err != nil {
		t.Fatal("restored preview admission invalid", err)
	}

	for _, item := range []struct{ id, selectSQL, deleteSQL, insertSQL string }{
		{releaseID, `SELECT canonical_bytes,reveal_map_sha256,created_at FROM reporting_release_reveal_maps WHERE release_id=$1`, `DELETE FROM reporting_release_reveal_maps WHERE release_id=$1`, `INSERT INTO reporting_release_reveal_maps(release_id,canonical_bytes,reveal_map_sha256,created_at) VALUES($1,$2,$3,$4)`},
		{previewID, `SELECT canonical_bytes,reveal_map_sha256,created_at FROM reporting_preview_reveal_maps WHERE preview_attempt_id=$1`, `DELETE FROM reporting_preview_reveal_maps WHERE preview_attempt_id=$1`, `INSERT INTO reporting_preview_reveal_maps(preview_attempt_id,canonical_bytes,reveal_map_sha256,created_at) VALUES($1,$2,$3,$4)`},
	} {
		var original, restored []byte
		var originalSHA, restoredSHA string
		var originalAt, restoredAt time.Time
		if err := sourcePool.QueryRow(ctx, item.selectSQL, item.id).Scan(&original, &originalSHA, &originalAt); err != nil {
			t.Fatal(err)
		}
		if err := targetPool.QueryRow(ctx, item.selectSQL, item.id).Scan(&restored, &restoredSHA, &restoredAt); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(original, restored) || originalSHA != restoredSHA || !originalAt.Equal(restoredAt) {
			t.Fatal("backup changed private bytes or bindings")
		}
		if _, err := targetPool.Exec(ctx, item.deleteSQL, item.id); err != nil {
			t.Fatal(err)
		}
		if reporting.ValidateRetainedSensitiveArtifacts(ctx, targetPool) == nil {
			t.Fatal("missing restored adjunct did not block readiness")
		}
		if _, err := targetPool.Exec(ctx, item.insertSQL, item.id, original, originalSHA, originalAt); err != nil {
			t.Fatal(err)
		}
	}
	if err := reporting.ValidateRetainedSensitiveArtifacts(ctx, targetPool); err != nil {
		t.Fatal("original private artifacts did not restore readiness", err)
	}
}
