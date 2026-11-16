package recovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRestoreCandidateUsesObjectStoreBackupLatestTieBreakers(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "backup_restore-u-10-02-latest-tiebreaker")
	store := recovery.NewStore(db)
	capture := newCurrentMetadataCapture(t, store)
	ctx := context.Background()

	asOf := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	createdAt := asOf.Add(-2 * time.Hour)
	consistencyPointAt := asOf.Add(-time.Hour)
	for _, backupSetID := range []uuid.UUID{
		uuid.MustParse("00000000-0000-0000-0000-000000100203"),
		uuid.MustParse("00000000-0000-0000-0000-000000100204"),
	} {
		if _, err := capture.CaptureAndPublish(ctx, recovery.VNextCaptureParams{
			BackupSetID:        backupSetID,
			ConsistencyPointAt: consistencyPointAt,
			CreatedAt:          createdAt,
			RetainedUntil:      createdAt.Add(31 * 24 * time.Hour),
		}); err != nil {
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

	// A later created_at outranks a higher backup_set_id at the same point, and
	// a later consistency point outranks both.
	for _, step := range []struct {
		id                 uuid.UUID
		consistencyPointAt time.Time
		createdAt          time.Time
		reason             string
	}{
		{uuid.MustParse("00000000-0000-0000-0000-000000100201"), consistencyPointAt, createdAt.Add(time.Minute), "later created_at"},
		{uuid.MustParse("00000000-0000-0000-0000-000000100200"), consistencyPointAt.Add(time.Minute), createdAt.Add(-time.Minute), "later consistency point"},
	} {
		if _, err := capture.CaptureAndPublish(ctx, recovery.VNextCaptureParams{
			BackupSetID:        step.id,
			ConsistencyPointAt: step.consistencyPointAt,
			CreatedAt:          step.createdAt,
			RetainedUntil:      step.createdAt.Add(31 * 24 * time.Hour),
		}); err != nil {
			t.Fatalf("capture %s backup set: %v", step.reason, err)
		}
		selected, err := store.RestoreCandidateBackup(ctx, asOf)
		if err != nil {
			t.Fatalf("select restore candidate after %s: %v", step.reason, err)
		}
		if selected.BackupSetID != step.id {
			t.Fatalf("restore candidate got %s want %s selected by %s", selected.BackupSetID, step.id, step.reason)
		}
	}
}

func TestObjectStoreSnapshotCarriesRestorableBytes(t *testing.T) {
	assertCurrentObjectRoundTrip(t, []byte("backup_restore restorable object bytes"))
}
