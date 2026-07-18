package recovery_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRestoreCandidateUsesObjectStoreBackupLatestTieBreakers(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "backup_restore-u-10-02-latest-tiebreaker")
	store := recovery.NewStore(db)
	capture := newCaptureService(t, store)
	ctx := context.Background()

	asOf := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	createdAt := asOf.Add(-2 * time.Hour)
	consistencyPointAt := asOf.Add(-time.Hour)
	for _, backupSetID := range []uuid.UUID{
		uuid.MustParse("00000000-0000-0000-0000-000000100203"),
		uuid.MustParse("00000000-0000-0000-0000-000000100204"),
	} {
		if _, err := capture.CaptureBackupSet(ctx, captureParams(recovery.CaptureBackupSetParams{
			BackupSetID:        backupSetID,
			ConsistencyPointAt: consistencyPointAt,
			CreatedAt:          createdAt,
			RetainedUntil:      createdAt.Add(31 * 24 * time.Hour),
		})); err != nil {
			t.Fatalf("capture same-point backup set: %v", err)
		}
	}

	selected, err := store.RestoreCandidateBackup(ctx, asOf)
	if err != nil {
		t.Fatalf("select restore candidate: %v", err)
	}
	want := uuid.MustParse("00000000-0000-0000-0000-000000100204")
	if selected.BackupSetID != want {
		t.Fatalf("restore candidate got %s want highest backup_set_id tie-breaker %s", selected.BackupSetID, want)
	}
}

func TestObjectStoreSnapshotCarriesRestorableBytes(t *testing.T) {
	ctx := context.Background()
	store, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("create object store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	payload := []byte("backup_restore restorable object bytes")
	if err := store.PutObject(ctx, "objects/proof.txt", bytes.NewReader(payload), int64(len(payload)), "text/plain"); err != nil {
		t.Fatalf("put object fixture: %v", err)
	}
	body, err := recovery.CaptureObjectStoreSnapshotArtifact(ctx, store, "")
	if err != nil {
		t.Fatalf("capture object-store snapshot: %v", err)
	}
	artifact, err := recovery.DecodeObjectStoreSnapshotArtifact(body)
	if err != nil {
		t.Fatalf("decode object-store snapshot: %v", err)
	}
	if artifact.SchemaID != recovery.ObjectStoreSnapshotArtifactSchemaID || len(artifact.Objects) != 1 {
		t.Fatalf("unexpected object-store snapshot metadata: %#v", artifact)
	}
	decoded, err := base64.StdEncoding.DecodeString(artifact.Objects[0].BodyBase64)
	if err != nil {
		t.Fatalf("decode object-store body: %v", err)
	}
	if !bytes.Equal(decoded, payload) {
		t.Fatalf("restorable body got %q want %q", decoded, payload)
	}
}
