package asserttest

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/versionid"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

type rowScanner interface {
	Scan(dest ...any) error
}

type rows interface {
	Err() error
	Next() bool
	Scan(dest ...any) error
}

type Database struct {
	queryRow        func(context.Context, string, ...any) rowScanner
	query           func(context.Context, string, ...any) (rows, error)
	close           func(rows)
	collaborationDB any
}

func SQLDatabase(db *sql.DB) Database {
	return Database{
		collaborationDB: db,
		queryRow: func(ctx context.Context, query string, args ...any) rowScanner {
			return db.QueryRowContext(ctx, query, args...)
		},
		query: func(ctx context.Context, query string, args ...any) (rows, error) {
			return db.QueryContext(ctx, query, args...)
		},
		close: func(result rows) { _ = result.(*sql.Rows).Close() },
	}
}

func PostgresDatabase(db postgres.DB) Database {
	return Database{
		collaborationDB: db,
		queryRow: func(ctx context.Context, query string, args ...any) rowScanner {
			return db.QueryRow(ctx, query, args...)
		},
		query: func(ctx context.Context, query string, args ...any) (rows, error) {
			return db.Query(ctx, query, args...)
		},
		close: func(result rows) { result.(pgx.Rows).Close() },
	}
}

type Counters struct {
	ChangeSets           int
	MutationRows         int
	Revisions            int
	ProjectionRows       int
	CollaborationIntents int
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

type TimelineRecordMutationExpectation struct {
	SequenceNo       int
	RecordID         string
	OperationKind    string
	BeforeRowVersion *int64
	AfterRowVersion  *int64
	BeforeCells      map[string]any
	AfterCells       map[string]any
}

type RecordLinkMutationExpectation struct {
	SequenceNo          int
	IncidentID          string
	SourceRecordID      string
	DestinationRecordID string
	LinkType            string
}

func RowVersion(value int64) *int64 {
	return &value
}

// AwaitIncidentStreamIdle waits for durable per-incident sequencing. Local API
// tailers independently deliver the committed replay log to their own hubs.
func AwaitIncidentStreamIdle(t testing.TB, db Database, incidentID string) {
	t.Helper()
	collaborationsupport.AwaitIncidentStreamIdle(t, db.collaborationDB, incidentID, 5*time.Second)
}

func SnapshotCounters(t testing.TB, db Database, incidentID string, recordID string) Counters {
	t.Helper()

	return Counters{
		ChangeSets:     queryCount(t, db, `SELECT COUNT(*) FROM change_sets WHERE incident_id::text = $1`, incidentID),
		MutationRows:   queryCount(t, db, `SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id::text = $1`, incidentID),
		Revisions:      queryCount(t, db, `SELECT COUNT(*) FROM record_revisions WHERE record_id::text = $1`, recordID),
		ProjectionRows: queryCount(t, db, `SELECT COUNT(*) FROM timeline_grid_projection WHERE incident_id::text = $1`, incidentID),
		CollaborationIntents: collaborationsupport.CountIntents(t, db.collaborationDB, collaborationsupport.IntentSelector{
			IncidentID:     incidentID,
			SourceRecordID: recordID,
		}),
	}
}

func LookupChangeSet(t testing.TB, db Database, changeSetID string) ChangeSetRow {
	t.Helper()

	var row ChangeSetRow
	if err := db.queryRow(context.Background(), `
SELECT actor_user_id::text, source, client_txn_id, request_id, created_at
FROM change_sets
WHERE change_set_id::text = $1
`, changeSetID).Scan(&row.ActorUserID, &row.Source, &row.ClientTxnID, &row.RequestID, &row.CreatedAt); err != nil {
		t.Fatalf("lookup change set: %v", err)
	}
	return row
}

func CountChangeSetMutations(t testing.TB, db Database, changeSetID string) int {
	t.Helper()
	return queryCount(t, db, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1`, changeSetID)
}

func LookupChangeSetMutations(t testing.TB, db Database, changeSetID string) []ChangeSetMutationRow {
	t.Helper()

	rows, err := db.query(context.Background(), `
SELECT sequence_no, target_kind, target_id, operation_kind, before_version_id, after_version_id, before_value, after_value
FROM change_set_mutations
WHERE change_set_id::text = $1
ORDER BY sequence_no ASC
`, changeSetID)
	if err != nil {
		t.Fatalf("query change-set mutations: %v", err)
	}
	defer db.close(rows)

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

func RequireTimelineRecordMutation(t testing.TB, db Database, changeSetID string, expectation TimelineRecordMutationExpectation) ChangeSetMutationRow {
	t.Helper()

	mutation := requireMutationAtSequence(t, LookupChangeSetMutations(t, db, changeSetID), expectation.SequenceNo)
	if mutation.TargetKind != "timeline_record" || mutation.TargetID != expectation.RecordID || mutation.OperationKind != expectation.OperationKind {
		t.Fatalf("unexpected timeline mutation identity: %#v", mutation)
	}
	requireVersionID(t, "before_version_id", mutation.BeforeVersionID, expectation.RecordID, expectation.BeforeRowVersion)
	requireVersionID(t, "after_version_id", mutation.AfterVersionID, expectation.RecordID, expectation.AfterRowVersion)
	requireRowVersionValue(t, "before_value", mutation.BeforeValue, expectation.BeforeRowVersion)
	requireRowVersionValue(t, "after_value", mutation.AfterValue, expectation.AfterRowVersion)
	requireCellValues(t, "before_value", mutation.BeforeValue, expectation.BeforeCells)
	requireCellValues(t, "after_value", mutation.AfterValue, expectation.AfterCells)
	return mutation
}

func RequireRecordLinkCreateMutation(t testing.TB, db Database, changeSetID string, expectation RecordLinkMutationExpectation) ChangeSetMutationRow {
	t.Helper()

	mutation := requireMutationAtSequence(t, LookupChangeSetMutations(t, db, changeSetID), expectation.SequenceNo)
	if mutation.TargetKind != "record_link" || mutation.OperationKind != "create" || mutation.TargetID == "" {
		t.Fatalf("unexpected record-link mutation identity: %#v", mutation)
	}
	if mutation.BeforeVersionID != nil || mutation.BeforeValue != nil {
		t.Fatalf("record-link create mutation must not have before state: %#v", mutation)
	}
	wants := map[string]any{
		"src_record_id": expectation.SourceRecordID,
		"dst_record_id": expectation.DestinationRecordID,
		"link_type":     expectation.LinkType,
	}
	if expectation.IncidentID != "" {
		wants["incident_id"] = expectation.IncidentID
	}
	for key, want := range wants {
		if got := mutation.AfterValue[key]; got != want {
			t.Fatalf("unexpected record-link mutation %s: got %#v want %#v in %#v", key, got, want, mutation.AfterValue)
		}
	}
	if _, ok := mutation.AfterValue["record_link_id"].(string); !ok {
		t.Fatalf("record-link mutation missing record_link_id: %#v", mutation.AfterValue)
	}
	return mutation
}

func RequireSupersedeCoupledChangeSet(t testing.TB, db Database, changeSetID string, recordID string, replacementRecordID string, wantRowVersion int64) {
	t.Helper()

	mutations := LookupChangeSetMutations(t, db, changeSetID)
	if len(mutations) != 2 {
		t.Fatalf("expected supersede change set %s to write exactly two mutations, got %#v", changeSetID, mutations)
	}

	RequireTimelineRecordMutation(t, db, changeSetID, TimelineRecordMutationExpectation{
		SequenceNo:       1,
		RecordID:         recordID,
		OperationKind:    "patch",
		BeforeRowVersion: RowVersion(wantRowVersion - 1),
		AfterRowVersion:  RowVersion(wantRowVersion),
		AfterCells:       map[string]any{"timeline.capture_state": "superseded"},
	})
	RequireRecordLinkCreateMutation(t, db, changeSetID, RecordLinkMutationExpectation{
		SequenceNo:          2,
		SourceRecordID:      replacementRecordID,
		DestinationRecordID: recordID,
		LinkType:            "supersedes",
	})
}

func CountRecordRevisions(t testing.TB, db Database, recordID string) int {
	t.Helper()
	return queryCount(t, db, `SELECT COUNT(*) FROM record_revisions WHERE record_id::text = $1`, recordID)
}

func CountActiveSupersedesLinks(t testing.TB, db Database, incidentID string, replacementRecordID string, supersededRecordID string) int {
	t.Helper()
	return queryCount(t, db, `SELECT COUNT(*) FROM record_links WHERE incident_id::text = $1 AND src_record_id::text = $2 AND dst_record_id::text = $3 AND link_type = 'supersedes' AND deleted_at IS NULL`, incidentID, replacementRecordID, supersededRecordID)
}

func LookupProjectionRow(t testing.TB, db Database, recordID string) ProjectionRow {
	t.Helper()

	var row ProjectionRow
	var replacement sql.NullString
	if err := db.queryRow(context.Background(), `
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

func AwaitRecordChange(t testing.TB, messages <-chan platformws.Message, timeout time.Duration) platformws.RecordChangedEvent {
	t.Helper()

	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	deadline := time.After(timeout)
	for {
		select {
		case message, ok := <-messages:
			if !ok {
				t.Fatal("incident stream closed before record change")
			}
			if message.Type != "record_changed" {
				continue
			}
			change, err := platformws.RecordChangeFromSequencedMessage(message)
			if err != nil {
				t.Fatalf("decode record change: %v", err)
			}
			return change
		case <-deadline:
			t.Fatal("timed out waiting for record change")
			return platformws.RecordChangedEvent{}
		}
	}
}

func RequireNoRecordChange(t testing.TB, messages <-chan platformws.Message, timeout time.Duration) {
	t.Helper()

	if timeout <= 0 {
		timeout = 300 * time.Millisecond
	}

	deadline := time.After(timeout)
	for {
		select {
		case message, ok := <-messages:
			if !ok {
				return
			}
			if message.Type != "record_changed" {
				continue
			}
			change, err := platformws.RecordChangeFromSequencedMessage(message)
			if err != nil {
				t.Fatalf("decode unexpected record change: %v", err)
			}
			t.Fatalf("expected no record change, got %+v", change)
		case <-deadline:
			return
		}
	}
}

func queryCount(t testing.TB, db Database, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.queryRow(context.Background(), query, args...).Scan(&count); err != nil {
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

func requireMutationAtSequence(t testing.TB, mutations []ChangeSetMutationRow, sequenceNo int) ChangeSetMutationRow {
	t.Helper()
	for _, mutation := range mutations {
		if mutation.SequenceNo == sequenceNo {
			return mutation
		}
	}
	t.Fatalf("mutation sequence %d not found in %#v", sequenceNo, mutations)
	return ChangeSetMutationRow{}
}

func requireVersionID(t testing.TB, field string, got *string, recordID string, wantVersion *int64) {
	t.Helper()
	if wantVersion == nil {
		if got != nil {
			t.Fatalf("expected nil %s, got %q", field, *got)
		}
		return
	}
	want := versionid.Format(uuid.MustParse(recordID), *wantVersion)
	if got == nil || *got != want {
		var gotValue any
		if got != nil {
			gotValue = *got
		}
		t.Fatalf("unexpected %s: got %v want %s", field, gotValue, want)
	}
}

func requireRowVersionValue(t testing.TB, field string, row map[string]any, wantVersion *int64) {
	t.Helper()
	if wantVersion == nil {
		if row != nil {
			t.Fatalf("expected nil %s row, got %#v", field, row)
		}
		return
	}
	if row == nil {
		t.Fatalf("expected %s row_version %d, got nil row", field, *wantVersion)
	}
	if row["snapshot_schema_id"] != "cartulary.revisions.snapshot.timeline_event.v1" {
		t.Fatalf("unexpected %s snapshot schema: %#v", field, row)
	}
	record, ok := row["record"].(map[string]any)
	if !ok {
		t.Fatalf("expected %s canonical record object, got %#v", field, row)
	}
	if got, ok := numberAsInt64(record["row_version"]); !ok || got != *wantVersion {
		t.Fatalf("unexpected %s row_version: got %#v want %d in %#v", field, record["row_version"], *wantVersion, row)
	}
}

func requireCellValues(t testing.TB, field string, row map[string]any, wants map[string]any) {
	t.Helper()
	if len(wants) == 0 {
		return
	}
	if row == nil {
		t.Fatalf("expected %s row with cells %#v, got nil", field, wants)
	}
	source, ok := row["source"].(map[string]any)
	if !ok {
		t.Fatalf("expected %s canonical source object, got %#v", field, row)
	}
	for fieldKey, want := range wants {
		sourceKey := strings.TrimPrefix(fieldKey, "timeline.")
		if got := source[sourceKey]; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected %s source %s: got %#v want %#v", field, fieldKey, got, want)
		}
	}
}

func numberAsInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case float64:
		return int64(typed), typed == float64(int64(typed))
	default:
		return 0, false
	}
}
