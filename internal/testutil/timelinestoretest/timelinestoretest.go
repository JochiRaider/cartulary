package timelinestoretest

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
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

type ChangeSetMutationRow struct {
	SequenceNo      int
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     map[string]any
	AfterValue      map[string]any
}

type ProjectionRow struct {
	RowVersion          int64
	CaptureState        string
	ReplacementRecordID *string
}

func SnapshotCounters(t testing.TB, db postgres.DB, incidentID string, recordID string) Counters {
	t.Helper()

	return Counters{
		ChangeSets:     queryCount(t, db, `SELECT COUNT(*) FROM change_sets WHERE incident_id::text = $1`, incidentID),
		MutationRows:   queryCount(t, db, `SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id::text = $1`, incidentID),
		Revisions:      queryCount(t, db, `SELECT COUNT(*) FROM record_revisions WHERE record_id::text = $1`, recordID),
		ProjectionRows: queryCount(t, db, `SELECT COUNT(*) FROM timeline_grid_projection WHERE incident_id::text = $1`, incidentID),
	}
}

func LookupChangeSet(t testing.TB, db postgres.DB, changeSetID string) ChangeSetRow {
	t.Helper()

	var row ChangeSetRow
	if err := db.QueryRow(context.Background(), `
SELECT actor_user_id::text, source, client_txn_id, request_id, created_at
FROM change_sets
WHERE change_set_id::text = $1
`, changeSetID).Scan(&row.ActorUserID, &row.Source, &row.ClientTxnID, &row.RequestID, &row.CreatedAt); err != nil {
		t.Fatalf("lookup change set: %v", err)
	}
	return row
}

func CountChangeSetMutations(t testing.TB, db postgres.DB, changeSetID string) int {
	t.Helper()
	return queryCount(t, db, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1`, changeSetID)
}

func LookupChangeSetMutations(t testing.TB, db postgres.DB, changeSetID string) []ChangeSetMutationRow {
	t.Helper()

	rows, err := db.Query(context.Background(), `
SELECT sequence_no, target_kind, target_id, operation_kind, before_version_id, after_version_id, before_value, after_value
FROM change_set_mutations
WHERE change_set_id::text = $1
ORDER BY sequence_no ASC
`, changeSetID)
	if err != nil {
		t.Fatalf("query change-set mutations: %v", err)
	}
	defer rows.Close()

	var mutations []ChangeSetMutationRow
	for rows.Next() {
		var mutation ChangeSetMutationRow
		var beforeVersion sql.NullString
		var afterVersion sql.NullString
		var beforeJSON []byte
		var afterJSON []byte
		if err := rows.Scan(
			&mutation.SequenceNo,
			&mutation.TargetKind,
			&mutation.TargetID,
			&mutation.OperationKind,
			&beforeVersion,
			&afterVersion,
			&beforeJSON,
			&afterJSON,
		); err != nil {
			t.Fatalf("scan change-set mutation: %v", err)
		}
		if beforeVersion.Valid {
			mutation.BeforeVersionID = &beforeVersion.String
		}
		if afterVersion.Valid {
			mutation.AfterVersionID = &afterVersion.String
		}
		mutation.BeforeValue = decodeJSONMap(t, beforeJSON)
		mutation.AfterValue = decodeJSONMap(t, afterJSON)
		mutations = append(mutations, mutation)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate change-set mutations: %v", err)
	}
	return mutations
}

func RequireSupersedeCoupledChangeSet(t testing.TB, db postgres.DB, changeSetID string, recordID string, replacementRecordID string, wantRowVersion int64) {
	t.Helper()

	mutations := LookupChangeSetMutations(t, db, changeSetID)
	if len(mutations) != 2 {
		t.Fatalf("expected supersede change set %s to write exactly two mutations, got %#v", changeSetID, mutations)
	}

	recordMutation := mutations[0]
	if recordMutation.SequenceNo != 1 || recordMutation.TargetKind != "timeline_record" || recordMutation.TargetID != recordID || recordMutation.OperationKind != "patch" {
		t.Fatalf("unexpected primary supersede mutation: %#v", recordMutation)
	}
	if recordMutation.AfterValue["record_id"] != recordID {
		t.Fatalf("expected supersede mutation row record_id %q, got %#v", recordID, recordMutation.AfterValue)
	}
	if got := recordMutation.AfterValue["row_version"]; got != float64(wantRowVersion) {
		t.Fatalf("expected supersede mutation row_version %d, got %#v", wantRowVersion, recordMutation.AfterValue)
	}
	cells, ok := recordMutation.AfterValue["cells"].(map[string]any)
	if !ok {
		t.Fatalf("expected supersede mutation cells map, got %#v", recordMutation.AfterValue)
	}
	captureStateCell, ok := cells["timeline.capture_state"].(map[string]any)
	if !ok || captureStateCell["value"] != "superseded" {
		t.Fatalf("expected supersede mutation capture_state cell, got %#v", recordMutation.AfterValue)
	}

	linkMutation := mutations[1]
	if linkMutation.SequenceNo != 2 || linkMutation.TargetKind != "record_link" || linkMutation.OperationKind != "create" {
		t.Fatalf("unexpected supersede link mutation: %#v", linkMutation)
	}
	if linkMutation.AfterValue["src_record_id"] != replacementRecordID || linkMutation.AfterValue["dst_record_id"] != recordID || linkMutation.AfterValue["link_type"] != "supersedes" {
		t.Fatalf("unexpected supersede link mutation payload: %#v", linkMutation.AfterValue)
	}
}

func CountRecordRevisions(t testing.TB, db postgres.DB, recordID string) int {
	t.Helper()
	return queryCount(t, db, `SELECT COUNT(*) FROM record_revisions WHERE record_id::text = $1`, recordID)
}

func CountActiveSupersedesLinks(t testing.TB, db postgres.DB, incidentID string, replacementRecordID string, supersededRecordID string) int {
	t.Helper()
	return queryCount(t, db, `SELECT COUNT(*) FROM record_links WHERE incident_id::text = $1 AND src_record_id::text = $2 AND dst_record_id::text = $3 AND link_type = 'supersedes' AND deleted_at IS NULL`, incidentID, replacementRecordID, supersededRecordID)
}

func LookupProjectionRow(t testing.TB, db postgres.DB, recordID string) ProjectionRow {
	t.Helper()

	var row ProjectionRow
	var replacement sql.NullString
	if err := db.QueryRow(context.Background(), `
SELECT row_version, capture_state, replacement_record_id::text
FROM timeline_grid_projection
WHERE record_id::text = $1
`, recordID).Scan(&row.RowVersion, &row.CaptureState, &replacement); err != nil {
		t.Fatalf("lookup projection row: %v", err)
	}
	if replacement.Valid {
		row.ReplacementRecordID = &replacement.String
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

func queryCount(t testing.TB, db postgres.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRow(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func decodeJSONMap(t testing.TB, payload []byte) map[string]any {
	t.Helper()

	if len(payload) == 0 {
		return nil
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode json map: %v", err)
	}
	return decoded
}
