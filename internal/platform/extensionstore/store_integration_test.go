package extensionstore

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestExtensionCoordinationStateAndStagedObjects_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "extension-coordination")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	store, err := New(pool, []FamilyCounter{{FamilyID: "network_flow_activity.tables", Count: func(ctx context.Context, querier Querier) (int64, error) {
		var count int64
		err := querier.QueryRow(ctx, `SELECT COUNT(*) FROM network_flow_tables`).Scan(&count)
		return count, err
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.WithProfileLock(context.Background(), "network_flow_activity", time.Second, func(session *Session) error {
		tx, err := session.Begin(context.Background())
		if err != nil {
			return err
		}
		metadata, err := tx.StateMetadata(context.Background(), "network_flow_activity")
		if err != nil {
			return err
		}
		if metadata == nil || metadata.MigrationLineageID != "network_flow_activity.state_v1" || metadata.StateVersion != 1 {
			t.Fatalf("migrated metadata = %#v", metadata)
		}
		counts, err := tx.FamilyCounts(context.Background(), []string{"network_flow_activity.tables"})
		if err != nil || counts["network_flow_activity.tables"] != 0 {
			t.Fatalf("authoritative counts = %v/%v", counts, err)
		}
		_, err = tx.Rollback(context.Background())
		return err
	}); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	staged := NewStagedObject("staging-1", "network_flow_activity", "staged/object-1", 3, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", now)
	if err := store.AllocateStagedObject(context.Background(), staged); err != nil {
		t.Fatal(err)
	}
	if err := store.MarkStagedReady(context.Background(), staged.StagingID, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := store.AbandonStagedObject(context.Background(), staged.StagingID, now.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	object, err := store.StagedObject(context.Background(), staged.StagingID)
	if err != nil || object.State != StagedAbandoned || object.ReadyAt != nil || object.DeleteState != DeletePending {
		t.Fatalf("abandoned staged object = %#v/%v", object, err)
	}
	batch, err := store.PrepareCleanupBatch(context.Background(), now.Add(2*time.Minute), now.Add(2*time.Minute), 10)
	if err != nil || len(batch) != 1 || batch[0].StagingID != staged.StagingID {
		t.Fatalf("cleanup batch = %#v/%v", batch, err)
	}
	if err := store.RecordDeletionFailure(context.Background(), staged.StagingID, "delete_failed", now.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	object, err = store.StagedObject(context.Background(), staged.StagingID)
	if err != nil || object.DeleteAttemptCount != 1 || object.NextDeleteAttempt == nil || !object.NextDeleteAttempt.Equal(now.Add(3*time.Minute)) {
		t.Fatalf("retry state = %#v/%v", object, err)
	}
	batch, err = store.PrepareCleanupBatch(context.Background(), now.Add(3*time.Minute), now.Add(3*time.Minute), 10)
	if err != nil || len(batch) != 1 {
		t.Fatalf("retry batch = %#v/%v", batch, err)
	}
	if err := store.RecordDeletionSuccess(context.Background(), staged.StagingID); err != nil {
		t.Fatal(err)
	}
	object, err = store.StagedObject(context.Background(), staged.StagingID)
	if err != nil || object.DeleteState != DeleteDeleted || object.NextDeleteAttempt != nil {
		t.Fatalf("deleted state = %#v/%v", object, err)
	}
}
