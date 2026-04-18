package timelinetest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

type Counters struct {
	ChangeSets     int
	MutationRows   int
	Revisions      int
	ProjectionRows int
}

type ChangeSetRow struct {
	ActorUserID string
	Source      string
	ClientTxnID string
	RequestID   string
	CreatedAt   time.Time
}

func SnapshotCounters(t testing.TB, db *sql.DB, incidentID string, recordID string) Counters {
	t.Helper()

	return Counters{
		ChangeSets:     queryCount(t, db, `SELECT COUNT(*) FROM change_sets WHERE incident_id::text = $1`, incidentID),
		MutationRows:   queryCount(t, db, `SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id::text = $1`, incidentID),
		Revisions:      queryCount(t, db, `SELECT COUNT(*) FROM record_revisions WHERE record_id::text = $1`, recordID),
		ProjectionRows: queryCount(t, db, `SELECT COUNT(*) FROM timeline_grid_projection WHERE incident_id::text = $1`, incidentID),
	}
}

func LookupChangeSet(t testing.TB, db *sql.DB, changeSetID string) ChangeSetRow {
	t.Helper()

	var row ChangeSetRow
	if err := db.QueryRowContext(context.Background(), `
SELECT actor_user_id::text, source, client_txn_id, request_id, created_at
FROM change_sets
WHERE change_set_id::text = $1
`, changeSetID).Scan(&row.ActorUserID, &row.Source, &row.ClientTxnID, &row.RequestID, &row.CreatedAt); err != nil {
		t.Fatalf("lookup change set: %v", err)
	}
	return row
}

func AwaitRecordChange(t testing.TB, changes <-chan platformws.RecordChange, timeout time.Duration) platformws.RecordChange {
	t.Helper()

	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	select {
	case change, ok := <-changes:
		if !ok {
			t.Fatal("record change channel closed before event")
		}
		return change
	case <-time.After(timeout):
		t.Fatal("timed out waiting for record change")
		return platformws.RecordChange{}
	}
}

func RequireNoRecordChange(t testing.TB, changes <-chan platformws.RecordChange, timeout time.Duration) {
	t.Helper()

	if timeout <= 0 {
		timeout = 300 * time.Millisecond
	}

	select {
	case change, ok := <-changes:
		if ok {
			t.Fatalf("expected no record change, got %+v", change)
		}
	case <-time.After(timeout):
	}
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}
