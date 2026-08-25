package collaborationsupport

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type IntentSelector struct {
	IntentKey         string
	IncidentID        string
	EventFamily       string
	SourceIdentity    string
	SourceRecordID    string
	SourceChangeSetID string
	SourceRowVersion  *int64
	DispatchState     string
}

type IntentRecord struct {
	IntentKey        string
	IncidentID       uuid.UUID
	EventFamily      string
	SourceIdentity   string
	SourceRecordID   *uuid.UUID
	SourceChangeSet  *uuid.UUID
	SourceRowVersion *int64
	CanonicalPayload []byte
}

type rowScanner interface {
	Scan(...any) error
}

type postgresQueryRower interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type sqlQueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func CountIntents(t testing.TB, db any, selector IntentSelector) int {
	t.Helper()
	where, args := intentWhere(selector)
	var count int
	if err := queryRow(t, db, "SELECT count(*) FROM collaboration_event_intents"+where, args...).Scan(&count); err != nil {
		t.Fatalf("count Collaboration intents: %v", err)
	}
	return count
}

func RequireIntentCount(t testing.TB, db any, selector IntentSelector, want int) {
	t.Helper()
	if got := CountIntents(t, db, selector); got != want {
		t.Fatalf("Collaboration intent count = %d, want %d for %#v", got, want, selector)
	}
}

func RequireSingleRecordChangedIntent(
	t testing.TB,
	db any,
	changeSetID string,
	recordID string,
	changeKind string,
) {
	t.Helper()
	selector := IntentSelector{
		EventFamily: "record_changed", SourceChangeSetID: changeSetID, SourceRecordID: recordID,
	}
	RequireIntentCount(t, db, selector, 1)
	var payload struct {
		AffectedViews []struct {
			ChangeKind string `json:"change_kind"`
		} `json:"affected_views"`
	}
	intent := LoadLatestIntent(t, db, selector)
	if err := json.Unmarshal(intent.CanonicalPayload, &payload); err != nil {
		t.Fatalf("decode record_changed intent: %v", err)
	}
	if len(payload.AffectedViews) == 0 || payload.AffectedViews[0].ChangeKind != changeKind {
		t.Fatalf("record_changed affected view kind = %#v, want %q", payload.AffectedViews, changeKind)
	}
}

func LoadLatestIntent(t testing.TB, db any, selector IntentSelector) IntentRecord {
	t.Helper()
	where, args := intentWhere(selector)
	var record IntentRecord
	if err := queryRow(t, db, `
SELECT intent_key, incident_id, event_family, source_identity,
       source_record_id, source_change_set_id, source_row_version,
       canonical_payload
  FROM collaboration_event_intents`+where+`
 ORDER BY created_at DESC, intent_key DESC
 LIMIT 1
`, args...).Scan(
		&record.IntentKey,
		&record.IncidentID,
		&record.EventFamily,
		&record.SourceIdentity,
		&record.SourceRecordID,
		&record.SourceChangeSet,
		&record.SourceRowVersion,
		&record.CanonicalPayload,
	); err != nil {
		t.Fatalf("load Collaboration intent: %v", err)
	}
	return record
}

func CountIntentsForChangeSetSources(t testing.TB, db any, incidentID string, sources ...string) int {
	t.Helper()
	if strings.TrimSpace(incidentID) == "" || len(sources) == 0 {
		t.Fatal("Collaboration intent source count requires an incident and at least one source")
	}
	args := make([]any, 0, len(sources)+1)
	args = append(args, incidentID)
	placeholders := make([]string, 0, len(sources))
	for _, source := range sources {
		args = append(args, source)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
	}
	var count int
	if err := queryRow(t, db, `
SELECT count(*)
  FROM collaboration_event_intents AS intent
  JOIN change_sets AS change_set
    ON change_set.change_set_id = intent.source_change_set_id
 WHERE change_set.incident_id::text = $1
   AND change_set.source IN (`+strings.Join(placeholders, ", ")+")", args...).Scan(&count); err != nil {
		t.Fatalf("count Collaboration intents for change-set sources: %v", err)
	}
	return count
}

func CountIntentsForChangeSet(
	t testing.TB,
	db any,
	incidentID string,
	source string,
	clientTxnID string,
	eventFamily string,
) int {
	t.Helper()
	var count int
	if err := queryRow(t, db, `
SELECT count(*)
  FROM collaboration_event_intents AS intent
  JOIN change_sets AS change_set
    ON change_set.change_set_id = intent.source_change_set_id
 WHERE change_set.incident_id::text = $1
   AND change_set.source = $2
   AND change_set.client_txn_id = $3
   AND intent.event_family = $4
`, incidentID, source, clientTxnID, eventFamily).Scan(&count); err != nil {
		t.Fatalf("count Collaboration intents for change set: %v", err)
	}
	return count
}

func CountImportApplyIntents(t testing.TB, db any, incidentID string, clientTxnID string) int {
	t.Helper()
	return CountIntentsForChangeSet(t, db, incidentID, "imports.apply", clientTxnID, "record_changed")
}

func RequireOneRecordChangeIntentPerRevision(t testing.TB, db any, changeSetID string) {
	t.Helper()
	var revisedRecords int
	if err := queryRow(t, db, `
SELECT count(DISTINCT record_id)
  FROM record_revisions
 WHERE change_set_id::text = $1
`, changeSetID).Scan(&revisedRecords); err != nil {
		t.Fatalf("count revised records for change set %s: %v", changeSetID, err)
	}
	if revisedRecords == 0 {
		t.Fatalf("change set %s has no revised records", changeSetID)
	}
	var mismatched int
	if err := queryRow(t, db, `
SELECT count(*)
  FROM (
        SELECT revised.record_id
          FROM (
                SELECT DISTINCT change_set_id, record_id
                  FROM record_revisions
                 WHERE change_set_id::text = $1
               ) AS revised
          LEFT JOIN collaboration_event_intents AS intent
            ON intent.source_change_set_id = revised.change_set_id
           AND intent.source_record_id = revised.record_id
           AND intent.event_family = 'record_changed'
         GROUP BY revised.record_id
        HAVING count(intent.intent_id) <> 1
       ) AS mismatched
`, changeSetID).Scan(&mismatched); err != nil {
		t.Fatalf("count record-change intents for change set %s: %v", changeSetID, err)
	}
	if mismatched != 0 {
		t.Fatalf("change set %s has %d revised records without exactly one record_changed intent", changeSetID, mismatched)
	}
}

func AwaitIncidentStreamIdle(t testing.TB, db any, incidentID string, timeout time.Duration) {
	t.Helper()
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	deadline := time.Now().Add(timeout)
	selector := IntentSelector{IncidentID: incidentID, DispatchState: "pending"}
	for {
		pending := CountIntents(t, db, selector)
		if pending == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("Collaboration stream for incident %s did not become idle; pending=%d", incidentID, pending)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func intentWhere(selector IntentSelector) (string, []any) {
	clauses := make([]string, 0, 8)
	args := make([]any, 0, 8)
	add := func(column string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	if selector.IntentKey != "" {
		add("intent_key", selector.IntentKey)
	}
	if selector.IncidentID != "" {
		add("incident_id::text", selector.IncidentID)
	}
	if selector.EventFamily != "" {
		add("event_family", selector.EventFamily)
	}
	if selector.SourceIdentity != "" {
		add("source_identity", selector.SourceIdentity)
	}
	if selector.SourceRecordID != "" {
		add("source_record_id::text", selector.SourceRecordID)
	}
	if selector.SourceChangeSetID != "" {
		add("source_change_set_id::text", selector.SourceChangeSetID)
	}
	if selector.SourceRowVersion != nil {
		add("source_row_version", *selector.SourceRowVersion)
	}
	if selector.DispatchState != "" {
		add("dispatch_state", selector.DispatchState)
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func queryRow(t testing.TB, db any, query string, args ...any) rowScanner {
	t.Helper()
	switch typed := db.(type) {
	case sqlQueryRower:
		return typed.QueryRowContext(context.Background(), query, args...)
	case postgresQueryRower:
		return typed.QueryRow(context.Background(), query, args...)
	default:
		t.Fatalf("unsupported Collaboration assertion database %T", db)
		return nil
	}
}
