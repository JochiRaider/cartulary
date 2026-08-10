package revisionsupport

import (
	"context"
	"database/sql"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const recordChangeIntentPerRevisionQuery = `
SELECT count(*)
  FROM (
        SELECT revised.record_id
          FROM (
                SELECT DISTINCT change_set_id, record_id
                  FROM record_revisions
                 WHERE change_set_id::text = $1
               ) revised
          LEFT JOIN collaboration_event_intents intent
            ON intent.source_change_set_id = revised.change_set_id
           AND intent.source_record_id = revised.record_id
           AND intent.event_family = 'record_changed'
         GROUP BY revised.record_id
        HAVING count(intent.intent_id) <> 1
       ) mismatched
`

const revisedRecordCountQuery = `
SELECT count(DISTINCT record_id)
  FROM record_revisions
 WHERE change_set_id::text = $1
`

func RequireExactlyOneChangeSet(t testing.TB, before int, after int) {
	t.Helper()
	if after-before != 1 {
		t.Fatalf("expected exactly one change_set delta, before=%d after=%d", before, after)
	}
}

func RequireActorAttribution(
	t testing.TB,
	gotActorUserID string,
	wantActorUserID string,
	gotSource string,
	wantSource string,
) {
	t.Helper()
	if gotActorUserID != wantActorUserID {
		t.Fatalf("unexpected actor attribution: got %q want %q", gotActorUserID, wantActorUserID)
	}
	if gotSource != wantSource {
		t.Fatalf("unexpected mutation source: got %q want %q", gotSource, wantSource)
	}
}

func RequireOneRecordChangeIntentPerRevisionSQL(t testing.TB, db *sql.DB, changeSetID string) {
	t.Helper()
	var revisedRecords int
	if err := db.QueryRowContext(context.Background(), revisedRecordCountQuery, changeSetID).Scan(&revisedRecords); err != nil {
		t.Fatalf("count revised records for change set %s: %v", changeSetID, err)
	}
	if revisedRecords == 0 {
		t.Fatalf("change set %s has no revised records", changeSetID)
	}
	var mismatched int
	if err := db.QueryRowContext(context.Background(), recordChangeIntentPerRevisionQuery, changeSetID).Scan(&mismatched); err != nil {
		t.Fatalf("count record-change intents for change set %s: %v", changeSetID, err)
	}
	if mismatched != 0 {
		t.Fatalf("change set %s has %d revised records without exactly one record_changed intent", changeSetID, mismatched)
	}
}

func RequireOneRecordChangeIntentPerRevisionPostgres(t testing.TB, db postgres.DB, changeSetID string) {
	t.Helper()
	var revisedRecords int
	if err := db.QueryRow(context.Background(), revisedRecordCountQuery, changeSetID).Scan(&revisedRecords); err != nil {
		t.Fatalf("count revised records for change set %s: %v", changeSetID, err)
	}
	if revisedRecords == 0 {
		t.Fatalf("change set %s has no revised records", changeSetID)
	}
	var mismatched int
	if err := db.QueryRow(context.Background(), recordChangeIntentPerRevisionQuery, changeSetID).Scan(&mismatched); err != nil {
		t.Fatalf("count record-change intents for change set %s: %v", changeSetID, err)
	}
	if mismatched != 0 {
		t.Fatalf("change set %s has %d revised records without exactly one record_changed intent", changeSetID, mismatched)
	}
}
