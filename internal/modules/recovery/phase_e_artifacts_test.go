package recovery_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func TestSupportPhaseE_ObjectStoreBackupManifestCanonicalAndSummaryRedacted(t *testing.T) {
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000100501")
	point := time.Date(2026, 6, 3, 14, 0, 0, 0, time.UTC)
	snapshot := recovery.ObjectStoreSnapshotArtifact{
		SchemaID: recovery.ObjectStoreSnapshotArtifactSchemaID,
		Objects: []recovery.ObjectStoreSnapshotItem{
			{
				Key:         "objects/b.txt",
				SizeBytes:   int64(len("second")),
				ContentType: "text/plain",
				SHA256:      "16367aacb67a4a017c5874470d245fa44f44595a93e22f5235a2fc6e12691f73",
				BodyBase64:  "c2Vjb25k",
			},
			{
				Key:         "objects/a.txt",
				SizeBytes:   int64(len("first")),
				ContentType: "text/plain",
				SHA256:      "a7937b64b8c27183e9ef17e047b74546e89fb62e012415bf7f1ff4953e220a71",
				BodyBase64:  "Zmlyc3Q=",
			},
		},
	}
	objectBlobID := uuid.MustParse("00000000-0000-0000-0000-000000100502")
	manifest, manifestBody, err := recovery.BuildSeaweedFSS3ObjectStoreBackupManifest(snapshot, recovery.ObjectStoreBackupManifestParams{
		BackupSetID:        backupSetID,
		ConsistencyPointAt: point,
		Bucket:             "private-bucket",
		BlobObjectIDsByStorageRef: map[string]uuid.UUID{
			"objects/a.txt": objectBlobID,
		},
	})
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	if !bytes.HasSuffix(manifestBody, []byte("\n")) || bytes.Contains(manifestBody, []byte(" ")) {
		t.Fatalf("manifest is not compact LF-terminated canonical JSON: %s", manifestBody)
	}
	decoded, err := recovery.DecodeObjectStoreBackupManifestArtifact(manifestBody)
	if err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if decoded.Objects[0].StorageRef != "objects/a.txt" ||
		decoded.Objects[0].ObjectBlobID != objectBlobID.String() ||
		decoded.Objects[1].StorageRef != "objects/b.txt" {
		t.Fatalf("manifest objects are not sorted or blob-linked correctly: %#v", decoded.Objects)
	}
	if decoded.ManifestSHA256 != manifest.ManifestSHA256 {
		t.Fatalf("manifest sha changed: got %s want %s", decoded.ManifestSHA256, manifest.ManifestSHA256)
	}

	summary, summaryBody, err := recovery.BuildObjectStoreBackupSummary(decoded)
	if err != nil {
		t.Fatalf("build summary: %v", err)
	}
	if summary.BucketRef.RedactionClass != "bucket" || !summary.BucketRef.Redacted {
		t.Fatalf("summary bucket ref is not redacted: %#v", summary.BucketRef)
	}
	if bytes.Contains(summaryBody, []byte("private-bucket")) ||
		bytes.Contains(summaryBody, []byte("objects/a.txt")) ||
		bytes.Contains(summaryBody, []byte("objects/b.txt")) {
		t.Fatalf("summary leaked raw bucket or storage refs: %s", summaryBody)
	}
	if _, err := recovery.DecodeObjectStoreBackupSummaryArtifact(summaryBody); err != nil {
		t.Fatalf("decode summary: %v", err)
	}
}

func TestSupportPhaseE_DuplicateArtifactKeysRejected(t *testing.T) {
	body := []byte(`{"schema_id":"cartulary.object_store_backup_manifest.v1","schema_id":"cartulary.object_store_backup_manifest.v1"}` + "\n")
	if _, err := recovery.DecodeObjectStoreBackupManifestArtifact(body); err == nil || !strings.Contains(err.Error(), "unique") {
		t.Fatalf("duplicate manifest keys error got %v", err)
	}
}

func TestSupportPhaseE_CaptureSeaweedFSS3BackupArtifactsFromObjectStore(t *testing.T) {
	ctx := context.Background()
	store, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("create object store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	payload := []byte("phase e object backup bytes")
	if err := store.PutObject(ctx, "objects/proof.txt", bytes.NewReader(payload), int64(len(payload)), "text/plain"); err != nil {
		t.Fatalf("put object: %v", err)
	}
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000100503")
	objectBlobID := uuid.MustParse("00000000-0000-0000-0000-000000100504")
	artifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, store, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:        backupSetID,
		ConsistencyPointAt: time.Date(2026, 6, 3, 15, 0, 0, 0, time.UTC),
		Bucket:             "phase-e-bucket",
		BlobObjectIDsByStorageRef: map[string]uuid.UUID{
			"objects/proof.txt": objectBlobID,
		},
	})
	if err != nil {
		t.Fatalf("capture artifacts: %v", err)
	}
	manifest, err := recovery.DecodeObjectStoreBackupManifestArtifact(artifacts.ManifestBody)
	if err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if manifest.ObjectCount != 1 ||
		manifest.Objects[0].ObjectBlobID != objectBlobID.String() ||
		manifest.Objects[0].SHA256 == "" ||
		manifest.Objects[0].BackupMemberSHA256 != manifest.Objects[0].SHA256 {
		t.Fatalf("captured manifest missing object proof: %#v", manifest)
	}
	if _, err := recovery.DecodeObjectStoreSnapshotArtifact(artifacts.SnapshotBody); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if _, err := recovery.DecodeObjectStoreBackupSummaryArtifact(artifacts.SummaryBody); err != nil {
		t.Fatalf("decode summary: %v", err)
	}
}
