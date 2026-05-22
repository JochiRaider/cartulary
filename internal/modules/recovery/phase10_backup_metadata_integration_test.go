package recovery_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestPhase10_I_10_01_RealBackingStorageMetadataPersistsAndLatestLookup(t *testing.T) {
	harness := phase4test.StartServer(t, "phase10-i-10-01-metadata")
	store := recovery.NewStore(harness.Server.Runtime.Postgres)
	ctx := context.Background()
	backupStorage, err := recovery.NewBackupStorageFromConfig(harness.Server.Runtime.Config, map[string]string{
		recovery.RecoveryMasterKeyEnv: phase10RecoveryMasterKey,
	})
	if err != nil {
		t.Fatalf("create backup storage from runtime config: %v", err)
	}
	capture := recovery.NewCaptureService(store, backupStorage)

	adminLogin, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase10-i-10-01-incident",
		"incident_key":  "phase10-i-10-01",
		"title":         "Phase 10 I-10-01 Backup Metadata",
	})
	objectKey := "phase10/i-10-01/" + incident["incident_id"].(string) + "/proof.txt"
	if err := harness.Server.Runtime.ObjectStore.PutObject(ctx, objectKey, bytes.NewReader([]byte("phase10 backup proof")), int64(len("phase10 backup proof")), "text/plain"); err != nil {
		t.Fatalf("write object-store proof before capture: %v", err)
	}
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, harness.Server.Runtime.Postgres)
	if err != nil {
		t.Fatalf("capture postgres snapshot artifact: %v", err)
	}
	if !bytes.Contains(postgresArtifact, []byte("phase10-i-10-01")) {
		t.Fatalf("postgres snapshot artifact does not contain seeded incident data: %s", postgresArtifact)
	}
	objectStoreArtifact, err := recovery.CaptureObjectStoreSnapshotArtifact(ctx, harness.Server.Runtime.ObjectStore, "phase10/i-10-01/")
	if err != nil {
		t.Fatalf("capture object-store snapshot artifact: %v", err)
	}

	asOf := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	olderCreatedAt := asOf.Add(-6 * time.Hour)
	if _, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        uuid.MustParse("00000000-0000-0000-0000-000000101001"),
		ConsistencyPointAt: asOf.Add(-5 * time.Hour),
		CreatedAt:          olderCreatedAt,
		RetainedUntil:      olderCreatedAt.Add(31 * 24 * time.Hour),
		PostgresArtifact: recovery.BackupArtifact{
			Body:        postgresArtifact,
			ContentType: "application/json",
		},
		ObjectStoreArtifact: recovery.BackupArtifact{
			Body:        objectStoreArtifact,
			ContentType: "application/json",
		},
	})); err != nil {
		t.Fatalf("create older backup metadata: %v", err)
	}

	latestCreatedAt := asOf.Add(-2 * time.Hour)
	latestID := uuid.MustParse("00000000-0000-0000-0000-000000101002")
	latestCreated, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        latestID,
		ConsistencyPointAt: asOf.Add(-time.Hour),
		CreatedAt:          latestCreatedAt,
		RetainedUntil:      latestCreatedAt.Add(31 * 24 * time.Hour),
		PostgresArtifact: recovery.BackupArtifact{
			Body:        postgresArtifact,
			ContentType: "application/json",
		},
		ObjectStoreArtifact: recovery.BackupArtifact{
			Body:        objectStoreArtifact,
			ContentType: "application/json",
		},
	}))
	if err != nil {
		t.Fatalf("create latest backup metadata: %v", err)
	}

	reopenedStore := recovery.NewStore(harness.Server.Runtime.Postgres)
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
	rawPostgresArtifact := filepath.Join(harness.Server.Runtime.Config.Roots.BackupStorage.Path, filepath.FromSlash(reloaded.PostgresArtifactKey))
	rawBody, err := os.ReadFile(rawPostgresArtifact)
	if err != nil {
		t.Fatalf("read raw encrypted postgres artifact: %v", err)
	}
	if bytes.Contains(rawBody, []byte("phase10-i-10-01")) {
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
	if !bytes.Equal(reloadedObjectArtifact, objectStoreArtifact) {
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

func requireStoredArtifactProof(t *testing.T, storage recovery.BackupStorage, proof recovery.BackupArtifactProof) []byte {
	t.Helper()
	body, err := recovery.VerifyArtifactProof(context.Background(), storage, proof)
	if err != nil {
		t.Fatalf("verify stored artifact proof for %s: %v", proof.Key, err)
	}
	return body
}
