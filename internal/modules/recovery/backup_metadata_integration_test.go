package recovery_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRealBackingStorageMetadataPersistsAndLatestLookup_Integration(t *testing.T) {
	runtimeHarness := appsupport.StartRuntime(t)
	harness := runtimeHarness.StartDefaultServer(
		t,
		"backup_restore-i-10-01-metadata",
	)
	ctx := context.Background()
	sourceDB := &pgtest.TestDatabase{DSN: harness.Pool.Config().ConnConfig.ConnString()}
	recoveryDSN, err := sourceDB.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatalf("resolve source Recovery DSN: %v", err)
	}
	recoveryAdmission, err := postgres.Setup(ctx, postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          recoveryDSN,
		Purpose:      postgres.PurposeRecovery,
		ExpectedRole: "cartulary_recovery",
	})
	if err != nil {
		t.Fatalf("open admitted source Recovery pool: %v", err)
	}
	t.Cleanup(recoveryAdmission.Close)
	recoveryPool := recoveryAdmission.Pool()
	store := recovery.NewStore(recoveryPool)
	backupStorage, err := recoveryassembly.NewBackupStorage(
		harness.Server.Config.Roots.BackupStorage.BindingKind,
		harness.Server.Config.Roots.BackupStorage.Path,
		map[string]string{
			recovery.RecoveryMasterKeyEnv: RecoveryMasterKey,
		},
	)
	if err != nil {
		t.Fatalf("create backup storage from runtime config: %v", err)
	}

	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-backup_restore-i-10-01-incident",
		"incident_key":  "backup_restore-i-10-01",
		"title":         "Recovery and coordination recovery-metadata Backup Metadata",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	objectKey := "backup_restore/i-10-01/" + incident["incident_id"].(string) + "/proof.txt"
	objectPayload := []byte("backup_restore backup proof")
	sourceBucket := "backup-restore-i-10-01-source-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := runtimeHarness.S3.CreateBucket(ctx, sourceBucket); err != nil {
		t.Fatalf("create source SeaweedFS bucket: %v", err)
	}
	sourceObjectStore := S3StoreForBucket(t, harness, runtimeHarness, sourceBucket)
	if err := sourceObjectStore.PutObject(ctx, objectKey, bytes.NewReader(objectPayload), int64(len(objectPayload)), "text/plain"); err != nil {
		t.Fatalf("write object-store proof before capture: %v", err)
	}
	objectSHA := SHA256Hex(objectPayload)
	objectBlobID := uuid.MustParse("00000000-0000-0000-0000-000000101003")
	if _, err := harness.Pool.Exec(ctx, `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, expected_sha256_hex, observed_size, observed_content_type, observed_sha256_hex,
    target_expires_at, pending_expires_at, finalized_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'available',
    $5, $6, $5, 'text/plain', $6,
    $7, $7, $8, $8, $8
)
`, objectBlobID, incidentID, adminUserID, objectKey, int64(len(objectPayload)), objectSHA, asTime(t, "2026-05-22T13:00:00Z"), asTime(t, "2026-05-22T12:00:00Z")); err != nil {
		t.Fatalf("insert source durable object blob row: %v", err)
	}
	stateCatalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	sourcePacks, err := referenceassembly.NewRecoveryStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path, reference_data.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(sourcePacks.Close)
	inventories, err := recoveryassembly.CurrentVNextObjectInventoryCatalog(recoveryassembly.NewVNextObjectSource(sourceObjectStore), sourcePacks, nil)
	if err != nil {
		t.Fatal(err)
	}
	streaming, err := recovery.RequireStreamingBackupStorage(backupStorage)
	if err != nil {
		t.Fatal(err)
	}
	capture, err := recovery.NewVNextCaptureService(recoveryassembly.NewVNextSnapshotRepository(recoveryPool), streaming, stateCatalog, inventories)
	if err != nil {
		t.Fatal(err)
	}
	asOf := time.Now().UTC()
	captureAt := func(id uuid.UUID, created, point time.Time) (recovery.VNextCapturedBackup, recovery.BackupSet) {
		t.Helper()
		captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{BackupSetID: id, ConsistencyPointAt: point, CreatedAt: created, RetainedUntil: created.Add(31 * 24 * time.Hour)})
		if err != nil {
			t.Fatal(err)
		}
		published, err := store.PublishVNextCapturedBackup(ctx, captured)
		if err != nil {
			t.Fatal(err)
		}
		return captured, published
	}
	olderID := uuid.MustParse("00000000-0000-0000-0000-000000101001")
	captureAt(olderID, asOf.Add(-6*time.Hour), asOf.Add(-5*time.Hour))
	latestID := uuid.MustParse("00000000-0000-0000-0000-000000101002")
	latestCaptured, latestCreated := captureAt(latestID, asOf.Add(-2*time.Hour), asOf.Add(-time.Hour))
	var objectManifestBody bytes.Buffer
	if err := streaming.ReadArtifactStream(ctx, latestCaptured.ObjectManifestProof, &objectManifestBody); err != nil {
		t.Fatal(err)
	}
	var objectManifest recovery.VNextObjectStoreBackupManifest
	if err := json.Unmarshal(objectManifestBody.Bytes(), &objectManifest); err != nil {
		t.Fatal(err)
	}
	foundBlob := false
	for _, item := range objectManifest.Objects {
		if item.OwnerID == "module.evidence" && item.StorageKey == objectKey {
			if item.PlaintextSHA256 != objectSHA || item.PlaintextBytes != int64(len(objectPayload)) {
				t.Fatalf("backup blob identity changed: %#v", item)
			}
			foundBlob = true
		}
	}
	if !foundBlob || len(objectManifest.Objects) <= 1 {
		t.Fatal("backup omitted evidence blob or Reference Pack assets")
	}

	reopenedStore := recovery.NewStore(recoveryPool)
	targetDB := runtimeHarness.Postgres.PrepareIsolatedDatabaseT(t, "backup_restore-i-10-01-service-backed-target")
	targetPool, err := pgxpool.New(ctx, targetDB.DSN)
	if err != nil {
		t.Fatalf("open fresh target Postgres: %v", err)
	}
	t.Cleanup(targetPool.Close)
	targetBucket := "backup-restore-i-10-01-target-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := runtimeHarness.S3.CreateBucket(ctx, targetBucket); err != nil {
		t.Fatalf("create target SeaweedFS bucket: %v", err)
	}
	targetObjectStore := S3StoreForBucket(t, harness, runtimeHarness, targetBucket)
	if objects, err := targetObjectStore.ListObjects(ctx, ""); err != nil {
		t.Fatalf("list fresh target SeaweedFS bucket: %v", err)
	} else if len(objects) != 0 {
		t.Fatalf("fresh target SeaweedFS bucket is not empty before restore: %#v", objects)
	}
	projectionRuntime, err := projectionassembly.Build(targetPool)
	if err != nil {
		t.Fatalf("compose target projection runtime: %v", err)
	}
	targetPacks, err := referenceassembly.NewRecoveryStorage(t.TempDir(), t.TempDir(), reference_data.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(targetPacks.Close)
	graph, err := recoveryassembly.NewGraphProjectionRestoreParticipant(targetPool)
	if err != nil {
		t.Fatal(err)
	}
	serviceBackedRestore, err := recovery.NewVersionedRestoreRunner(reopenedStore, backupStorage, testExtensionBackupCatalog(t), stateCatalog).RestoreLatestSuccessfulRetained(ctx, recovery.RestoreTarget{
		RestoreOperationID: uuid.New(), TargetGenerationID: uuid.New(),
		Postgres: targetPool, ObjectStore: targetObjectStore, ReferencePacks: targetPacks,
		EvidenceObjects: evidence.NewRecoveryProvider(targetPool),
		GraphProjection: graph, Projections: projectionRuntime.RecoveryPorts().Rebuilder,
	}, asOf)
	if err != nil {
		t.Fatalf("restore latest retained backup into fresh SeaweedFS-backed target: %v", err)
	}
	if serviceBackedRestore.BackupSet.BackupSetID != latestID || serviceBackedRestore.ConsistencyReport.BlobCount != 1 {
		t.Fatalf("service-backed restore selected wrong backup or missed blob lifecycle: %#v", serviceBackedRestore)
	}
	restoredObject, _, err := targetObjectStore.ReadObject(ctx, objectKey, objectstore.ReadOptions{})
	if err != nil {
		t.Fatalf("read restored SeaweedFS object: %v", err)
	}
	restoredBytes, readErr := io.ReadAll(restoredObject)
	closeErr := restoredObject.Close()
	if readErr != nil {
		t.Fatalf("read restored SeaweedFS object body: %v", readErr)
	}
	if closeErr != nil {
		t.Fatalf("close restored SeaweedFS object body: %v", closeErr)
	}
	if !bytes.Equal(restoredBytes, objectPayload) {
		t.Fatalf("restored SeaweedFS object bytes changed: got %q want %q", restoredBytes, objectPayload)
	}
	var restoredBlobCount int
	if err := targetPool.QueryRow(ctx, `
SELECT count(*)
  FROM object_blobs
 WHERE object_blob_id = $1
   AND storage_key = $2
   AND upload_state = 'available'
   AND observed_sha256_hex = $3
`, objectBlobID, objectKey, objectSHA).Scan(&restoredBlobCount); err != nil {
		t.Fatalf("query restored object blob row: %v", err)
	}
	if restoredBlobCount != 1 {
		t.Fatalf("restored object blob lifecycle row count got %d want 1", restoredBlobCount)
	}

	reloaded, err := reopenedStore.GetBackupSet(ctx, latestID)
	if err != nil {
		t.Fatalf("reload committed latest backup metadata: %v", err)
	}
	if reloaded.BackupSetID != latestCreated.BackupSetID ||
		!reloaded.ConsistencyPointAt.Equal(latestCreated.ConsistencyPointAt) ||
		reloaded.PostgresRestoreAnchor != latestCreated.PostgresRestoreAnchor ||
		reloaded.ObjectStoreRestoreAnchor != latestCreated.ObjectStoreRestoreAnchor ||
		reloaded.IntegrityManifestSHA256 != latestCreated.IntegrityManifestSHA256 {
		t.Fatalf("committed metadata did not persist stable identity, point, and anchors:\ncreated=%#v\nreloaded=%#v", latestCreated, reloaded)
	}
	var incidentProof recovery.BackupArtifactStreamProof
	for _, artifact := range latestCaptured.IntegrityManifest.Artifacts {
		if strings.HasSuffix(artifact.LogicalRef, "/postgres/incidents.ndjson") {
			incidentProof = recovery.BackupArtifactStreamProof{LogicalRef: artifact.LogicalRef, ContentType: artifact.ContentType, PlaintextBytes: artifact.PlaintextBytes, PlaintextSHA256: artifact.PlaintextSHA256, EnvelopeRef: artifact.EnvelopeRef, EnvelopeSHA256: artifact.EnvelopeSHA256}
		}
	}
	if incidentProof.LogicalRef == "" {
		t.Fatal("backup omitted incident table unit")
	}
	rawBody, err := os.ReadFile(filepath.Join(harness.Server.Config.Roots.BackupStorage.Path, filepath.FromSlash(incidentProof.EnvelopeRef)))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(rawBody, []byte("backup_restore-i-10-01")) {
		t.Fatal("encrypted backup envelope contains incident plaintext")
	}
	var incidentRows bytes.Buffer
	if err := streaming.ReadArtifactStream(ctx, incidentProof, &incidentRows); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(incidentRows.Bytes(), []byte("backup_restore-i-10-01")) {
		t.Fatal("authenticated backup omitted incident content")
	}
	var manifestBody bytes.Buffer
	if err := streaming.ReadArtifactStream(ctx, latestCaptured.IntegrityProof, &manifestBody); err != nil {
		t.Fatal(err)
	}
	var manifest recovery.VNextBackupIntegrityManifest
	if err := json.Unmarshal(manifestBody.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaID != recovery.BackupIntegrityManifestV3SchemaID || manifest.BackupSetID != latestID.String() || reloaded.IntegrityManifestSHA256 != latestCaptured.IntegrityProof.PlaintextSHA256 {
		t.Fatalf("persisted integrity manifest does not match reloaded metadata: %#v", manifest)
	}
	if err := recovery.NewBackupCatalog(reopenedStore, backupStorage, testExtensionBackupCatalog(t), stateCatalog).VerifyBackupSetDurability(ctx, reloaded); err != nil {
		t.Fatal(err)
	}

	if reloaded.VerificationState != recovery.VerificationUnverified || reloaded.LastVerifiedRestoreAt != nil {
		t.Fatalf("committed backup metadata must remain unverified with null restore timestamp until verification: %#v", reloaded)
	}
	requireRetentionFloor(t, reloaded.CreatedAt, reloaded.RetainedUntil, "retained_until")
	requireRetentionFloor(t, reloaded.CreatedAt, reloaded.PostgresRestoreAnchorRetainedUntil, "postgres_restore_anchor_retained_until")
	requireRetentionFloor(t, reloaded.CreatedAt, reloaded.ObjectStoreRestoreAnchorRetainedUntil, "object_store_restore_anchor_retained_until")

	latest, err := reopenedStore.LatestSuccessfulRetainedBackup(ctx, asOf)
	if err != nil {
		t.Fatalf("latest successful retained lookup against real backing Postgres: %v", err)
	}
	if latest.BackupSetID != latestID {
		t.Fatalf("latest lookup got %s want %s", latest.BackupSetID, latestID)
	}
}

func S3StoreForBucket(t testing.TB, harness *appsupport.ServerHarness, runtimeHarness *appsupport.Runtime, bucket string) objectstore.Store {
	t.Helper()
	const serviceRef = "object_primary"
	cfg := harness.Server.Config
	cfg.Roots.ObjectStorage.BindingKind = "managed_service"
	cfg.Roots.ObjectStorage.Path = ""
	cfg.Roots.ObjectStorage.ServiceRef = serviceRef
	env := runtimeHarness.S3.EnvForServiceRef(serviceRef, bucket)
	store, err := appsupport.OpenObjectStore(context.Background(), cfg, env)
	if err != nil {
		t.Fatalf("open SeaweedFS object-store adapter for bucket %s: %v", bucket, err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}

func SHA256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func asTime(t testing.TB, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time fixture %q: %v", value, err)
	}
	return parsed
}
