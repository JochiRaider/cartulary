package revisionsupport

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

// SnapshotRecords observes authoritative source values through the application's
// native owner providers. The short transaction releases capture locks before
// returning; callers compare the opaque native snapshots without copying schemas.
func SnapshotRecords(t testing.TB, pool *pgxpool.Pool, appender *revisions.Appender, ids ...uuid.UUID) []revisions.RecordSnapshot {
	t.Helper()
	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	snapshots := make([]revisions.RecordSnapshot, 0, len(ids))
	for _, id := range ids {
		snapshot, err := appender.CaptureRecordSnapshotTx(context.Background(), tx, id)
		if err != nil {
			t.Fatalf("capture authoritative record %s: %v", id, err)
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots
}
