package recovery_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestRealBackingStorageMetadataPersistsAndLatestLookup_Integration(t *testing.T) {
	runtimeHarness := appsupport.StartRuntime(t)
	harness := runtimeHarness.StartDefaultServer(
		t,
		"backup_restore-i-10-01-metadata",
	)
	store := recovery.NewStore(harness.Server.Runtime.Postgres)
	ctx := context.Background()
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
	capture := recovery.NewCaptureService(store, backupStorage, testExtensionBackupCatalog(t))

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
	if _, err := harness.Server.Runtime.Postgres.Exec(ctx, `
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
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, harness.Server.Runtime.Postgres)
	if err != nil {
		t.Fatalf("capture postgres snapshot artifact: %v", err)
	}
	if !bytes.Contains(postgresArtifact, []byte("backup_restore-i-10-01")) {
		t.Fatalf("postgres snapshot artifact does not contain seeded incident data: %s", postgresArtifact)
	}
	blobIndex, err := recovery.AvailableBlobObjectIDsByStorageRef(ctx, harness.Server.Runtime.Postgres)
	if err != nil {
		t.Fatalf("index source blob storage refs: %v", err)
	}

	asOf := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	olderCreatedAt := asOf.Add(-6 * time.Hour)
	olderID := uuid.MustParse("00000000-0000-0000-0000-000000101001")
	olderPoint := asOf.Add(-5 * time.Hour)
	olderObjectArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, sourceObjectStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               olderID,
		ConsistencyPointAt:        olderPoint,
		Bucket:                    sourceBucket,
		Prefix:                    "backup_restore/i-10-01/",
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		t.Fatalf("capture older SeaweedFS object-store backup artifacts: %v", err)
	}
	if _, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        olderID,
		ConsistencyPointAt: olderPoint,
		CreatedAt:          olderCreatedAt,
		RetainedUntil:      olderCreatedAt.Add(31 * 24 * time.Hour),
		PostgresArtifact: recovery.BackupArtifact{
			Body:        postgresArtifact,
			ContentType: "application/json",
		},
		ObjectStoreArtifact: recovery.BackupArtifact{
			Body:        olderObjectArtifacts.SnapshotBody,
			ContentType: "application/json",
		},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: olderObjectArtifacts.ManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: olderObjectArtifacts.SummaryBody, ContentType: "application/json"},
	})); err != nil {
		t.Fatalf("create older backup metadata: %v", err)
	}

	latestCreatedAt := asOf.Add(-2 * time.Hour)
	latestID := uuid.MustParse("00000000-0000-0000-0000-000000101002")
	latestPoint := asOf.Add(-time.Hour)
	latestObjectArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, sourceObjectStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               latestID,
		ConsistencyPointAt:        latestPoint,
		Bucket:                    sourceBucket,
		Prefix:                    "backup_restore/i-10-01/",
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		t.Fatalf("capture latest SeaweedFS object-store backup artifacts: %v", err)
	}
	latestCreated, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        latestID,
		ConsistencyPointAt: latestPoint,
		CreatedAt:          latestCreatedAt,
		RetainedUntil:      latestCreatedAt.Add(31 * 24 * time.Hour),
		PostgresArtifact: recovery.BackupArtifact{
			Body:        postgresArtifact,
			ContentType: "application/json",
		},
		ObjectStoreArtifact: recovery.BackupArtifact{
			Body:        latestObjectArtifacts.SnapshotBody,
			ContentType: "application/json",
		},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: latestObjectArtifacts.ManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: latestObjectArtifacts.SummaryBody, ContentType: "application/json"},
	}))
	if err != nil {
		t.Fatalf("create latest backup metadata: %v", err)
	}
	decodedObjectManifest, err := recovery.DecodeObjectStoreBackupManifestArtifact(latestObjectArtifacts.ManifestBody)
	if err != nil {
		t.Fatalf("decode latest object-store backup manifest: %v", err)
	}
	if decodedObjectManifest.Bucket != sourceBucket ||
		decodedObjectManifest.ObjectStoreBackend != recovery.ObjectStoreBackendSeaweedFSS3 ||
		decodedObjectManifest.ObjectCount != 1 ||
		decodedObjectManifest.Objects[0].ObjectBlobID != objectBlobID.String() ||
		decodedObjectManifest.Objects[0].SHA256 != objectSHA {
		t.Fatalf("latest SeaweedFS backup manifest does not prove the durable blob: %#v", decodedObjectManifest)
	}

	reopenedStore := recovery.NewStore(harness.Server.Runtime.Postgres)
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
	serviceBackedRestore, err := recovery.NewRestoreRunner(reopenedStore, backupStorage, testExtensionBackupCatalog(t)).RestoreLatestSuccessfulRetained(ctx, recovery.RestoreTarget{
		Stopped:     true,
		Postgres:    targetPool,
		ObjectStore: targetObjectStore,
		Projections: timelineassembly.NewRestoreRebuilder(targetPool),
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
	rawPostgresArtifact := filepath.Join(harness.Server.Config.Roots.BackupStorage.Path, filepath.FromSlash(reloaded.PostgresArtifactKey))
	rawBody, err := os.ReadFile(rawPostgresArtifact)
	if err != nil {
		t.Fatalf("read raw encrypted postgres artifact: %v", err)
	}
	if bytes.Contains(rawBody, []byte("backup_restore-i-10-01")) {
		t.Fatalf("raw backup storage artifact contains incident marker plaintext: %s", rawBody)
	}
	requireArtifactProof(t, reloaded)
	reloadedPostgresArtifact := requireStoredArtifactProof(t, backupStorage, recovery.BackupArtifactProof{
		Key:       reloaded.PostgresArtifactKey,
		SHA256:    reloaded.PostgresArtifactSHA256,
		SizeBytes: reloaded.PostgresArtifactSizeBytes,
	})
	if !bytes.Equal(reloadedPostgresArtifact, postgresArtifact) {
		t.Fatalf("reloaded postgres artifact body changed")
	}
	reloadedObjectArtifact := requireStoredArtifactProof(t, backupStorage, recovery.BackupArtifactProof{
		Key:       reloaded.ObjectStoreArtifactKey,
		SHA256:    reloaded.ObjectStoreArtifactSHA256,
		SizeBytes: reloaded.ObjectStoreArtifactSizeBytes,
	})
	if !bytes.Equal(reloadedObjectArtifact, latestObjectArtifacts.SnapshotBody) {
		t.Fatalf("reloaded object-store artifact body changed")
	}
	manifestBody := requireStoredArtifactProof(t, backupStorage, recovery.BackupArtifactProof{
		Key:       reloaded.IntegrityManifestKey,
		SHA256:    reloaded.IntegrityManifestSHA256,
		SizeBytes: reloaded.IntegrityManifestSizeBytes,
	})
	manifest, err := recovery.DecodeIntegrityManifest(manifestBody)
	if err != nil {
		t.Fatalf("decode persisted integrity manifest: %v", err)
	}
	if manifest.SchemaID != recovery.BackupIntegrityManifestSchemaID ||
		manifest.BackupSetID != latestID.String() ||
		manifest.PostgresArtifact.Key != reloaded.PostgresArtifactKey ||
		manifest.ObjectStoreArtifact.Key != reloaded.ObjectStoreArtifactKey {
		t.Fatalf("persisted integrity manifest does not match reloaded metadata: %#v", manifest)
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

func requireStoredArtifactProof(t *testing.T, storage recovery.BackupStorage, proof recovery.BackupArtifactProof) []byte {
	t.Helper()
	body, err := recovery.VerifyArtifactProof(context.Background(), storage, proof)
	if err != nil {
		t.Fatalf("verify stored artifact proof for %s: %v", proof.Key, err)
	}
	return body
}
