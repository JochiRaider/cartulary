package artifacts_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestArtifactLinkedNoteAtomicity(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-linked-note-atomicity")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-linked-note-owner@example.test",
		"Artifact Linked Note Owner",
		"ArtifactLinkedNoteOwner1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-linked-note-incident",
		"IR-ARTIFACTS-LINKED-NOTE",
		"Artifact linked note contract",
	)
	conflictFields, err := revisionassembly.CurrentConflictFieldResolver()
	if err != nil {
		t.Fatalf("compose conflict field resolver: %v", err)
	}
	facade, err := workbookassembly.NewArtifactMutationContribution(
		harness.DB,
		conflicttest.NewCodec("artifacts-linked-notes"),
		revisionsupport.MustAppender(t),
		conflictFields,
		appsupport.ArtifactProjectionRows(harness.DB),
	)
	if err != nil {
		t.Fatalf("compose Artifacts mutation facade: %v", err)
	}
	now := time.Date(2026, 7, 30, 16, 0, 0, 0, time.UTC)

	for index, recordType := range []string{"timeline_event", "host", "identity", "evidence"} {
		recordType := recordType
		t.Run(recordType, func(t *testing.T) {
			sourceRecordID := seedLinkedNoteSource(t, harness, incident.ID, actor.ID, recordType, now)
			title := "Synthetic " + recordType + " note"
			clientTxnID := fmt.Sprintf("txn-artifacts-linked-note-%02d", index)
			command := artifacts.ContextualNoteCreateCommand{
				ActorUserID:    actor.ID,
				SourceRecordID: sourceRecordID,
				Request: artifacts.ContextualNoteCreateRequest{
					ClientTxnID: clientTxnID,
					Values: map[string]artifacts.FieldValue{
						"note.title": {Text: &title},
					},
					Collections: map[string]artifacts.WorkbookCollectionActionPayload{
						"note.tags": {Actions: []artifacts.WorkbookCollectionAction{{
							Op: "add_tag", RawText: "synthetic", NormalizedText: "synthetic",
						}}},
					},
				},
				RequestHash: []byte("hash-" + clientTxnID),
				RequestID:   "req-" + clientTxnID,
				OperationID: artifacts.OperationLinkedNoteCreate,
				Now:         now.Add(time.Duration(index) * time.Minute),
			}
			result, err := facade.CreateContextualNote(ctx, command)
			if err != nil {
				t.Fatalf("create linked note for %s: %v", recordType, err)
			}
			if result.RecordID == uuid.Nil || result.ChangeSetID == uuid.Nil || result.RowVersion != 1 {
				t.Fatalf("linked note result is incomplete: %#v", result)
			}
			requireLinkedNoteCount(t, harness, `
SELECT count(*)
  FROM artifacts
 WHERE record_id = $1
   AND incident_id = $2
   AND artifact_type = 'note'
   AND title = $3
`, result.RecordID, incident.ID, title, 1)
			requireLinkedNoteCount(t, harness, `
SELECT count(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'references_artifact'
   AND deleted_at IS NULL
`, incident.ID, sourceRecordID, result.RecordID, 1)
			requireLinkedNoteCount(t, harness, `SELECT count(*) FROM record_tags WHERE record_id = $1 AND normalized_tag_name = 'synthetic' AND deleted_at IS NULL`, result.RecordID, 1)
			requireLinkedNoteCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1 AND row_version = 1`, result.RecordID, 1)
			requireLinkedNoteCount(t, harness, `SELECT count(*) FROM artifact_grid_projection WHERE record_id = $1 AND artifact_type = 'note'`, result.RecordID, 1)
			requireLinkedNoteCount(t, harness, `SELECT count(*) FROM change_set_mutations WHERE change_set_id = $1`, result.ChangeSetID, 3)
			requireLinkedNoteCount(t, harness, `SELECT count(*) FROM change_set_mutations WHERE change_set_id = $1 AND target_kind = 'record_link'`, result.ChangeSetID, 1)
			requireLinkedNoteCount(t, harness, `SELECT count(*) FROM change_set_mutations WHERE change_set_id = $1 AND target_kind = 'record_tag'`, result.ChangeSetID, 1)

			replayed, err := facade.CreateContextualNote(ctx, command)
			if err != nil {
				t.Fatalf("replay linked note for %s: %v", recordType, err)
			}
			if !replayed.Replayed || replayed.RecordID != result.RecordID {
				t.Fatalf("linked note replay = %#v, want original record %s", replayed, result.RecordID)
			}
			requireLinkedNoteCount(t, harness, `SELECT count(*) FROM artifacts WHERE record_id = $1`, result.RecordID, 1)
			requireLinkedNoteCount(t, harness, `SELECT count(*) FROM record_links WHERE src_record_id = $1 AND dst_record_id = $2 AND deleted_at IS NULL`, sourceRecordID, result.RecordID, 1)

			conflicting := command
			conflicting.RequestHash = []byte("changed-" + clientTxnID)
			if _, err := facade.CreateContextualNote(ctx, conflicting); !errors.Is(err, artifacts.ErrClientTxnConflict) {
				t.Fatalf("changed linked-note replay error = %v, want client transaction conflict", err)
			}
		})
	}

	sourceRecordID := seedLinkedNoteSource(t, harness, incident.ID, actor.ID, "timeline_event", now.Add(time.Hour))
	before := linkedNoteCount(t, harness, `SELECT count(*) FROM records WHERE incident_id = $1`, incident.ID)
	_, err = facade.CreateContextualNote(ctx, artifacts.ContextualNoteCreateCommand{
		ActorUserID: actor.ID, SourceRecordID: sourceRecordID,
		Request:     artifacts.ContextualNoteCreateRequest{ClientTxnID: "txn-artifacts-linked-note-no-signal"},
		RequestHash: []byte("hash-artifacts-linked-note-no-signal"),
		RequestID:   "req-artifacts-linked-note-no-signal",
		OperationID: artifacts.OperationLinkedNoteCreate,
		Now:         now.Add(2 * time.Hour),
	})
	if err == nil {
		t.Fatal("linked note without title or body unexpectedly succeeded")
	}
	if got := linkedNoteCount(t, harness, `SELECT count(*) FROM records WHERE incident_id = $1`, incident.ID); got != before {
		t.Fatalf("rejected linked note changed record count: got %d want %d", got, before)
	}

	for index, fault := range []struct {
		name  string
		table string
	}{
		{name: "relationship", table: "record_links"},
		{name: "projection", table: "artifact_grid_projection"},
		{name: "revision", table: "change_sets"},
		{name: "idempotency", table: "route_idempotency"},
	} {
		fault := fault
		t.Run("rollback_"+fault.name, func(t *testing.T) {
			sourceRecordID := seedLinkedNoteSource(
				t,
				harness,
				incident.ID,
				actor.ID,
				"timeline_event",
				now.Add(time.Duration(index+3)*time.Hour),
			)
			clientTxnID := "txn-artifacts-linked-note-fault-" + fault.name
			beforeRecords := linkedNoteCount(
				t,
				harness,
				`SELECT count(*) FROM records WHERE incident_id = $1`,
				incident.ID,
			)
			installLinkedNoteFailureTrigger(t, harness, fault.table, fault.name)
			title := "Faulted " + fault.name + " note"
			_, err := facade.CreateContextualNote(ctx, artifacts.ContextualNoteCreateCommand{
				ActorUserID:    actor.ID,
				SourceRecordID: sourceRecordID,
				Request: artifacts.ContextualNoteCreateRequest{
					ClientTxnID: clientTxnID,
					Values: map[string]artifacts.FieldValue{
						"note.title": {Text: &title},
					},
				},
				RequestHash: []byte("hash-" + clientTxnID),
				RequestID:   "req-" + clientTxnID,
				OperationID: artifacts.OperationLinkedNoteCreate,
				Now:         now.Add(time.Duration(index+3) * time.Hour),
			})
			if err == nil {
				t.Fatalf("linked note with forced %s failure unexpectedly succeeded", fault.name)
			}
			requireLinkedNoteCount(
				t,
				harness,
				`SELECT count(*) FROM records WHERE incident_id = $1`,
				incident.ID,
				beforeRecords,
			)
			requireLinkedNoteCount(
				t,
				harness,
				`SELECT count(*) FROM change_sets WHERE client_txn_id = $1`,
				clientTxnID,
				0,
			)
			requireLinkedNoteCount(
				t,
				harness,
				`SELECT count(*) FROM route_idempotency WHERE client_txn_id = $1`,
				clientTxnID,
				0,
			)
		})
	}
}

func seedLinkedNoteSource(
	t testing.TB,
	harness *appsupport.StoreHarness,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	recordType string,
	now time.Time,
) uuid.UUID {
	t.Helper()
	tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin linked-note source seed: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	recordID, err := records.NewStore().InsertTx(context.Background(), tx, records.InsertParams{
		IncidentID: incidentID, RecordType: recordType,
		CreatedByUserID: actorID, CreatedAt: now,
		UpdatedByUserID: actorID, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("insert %s source record: %v", recordType, err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit %s source record: %v", recordType, err)
	}
	return recordID
}

func requireLinkedNoteCount(
	t testing.TB,
	harness *appsupport.StoreHarness,
	query string,
	argsAndWant ...any,
) {
	t.Helper()
	want := argsAndWant[len(argsAndWant)-1].(int)
	if got := linkedNoteCount(t, harness, query, argsAndWant[:len(argsAndWant)-1]...); got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
}

func linkedNoteCount(t testing.TB, harness *appsupport.StoreHarness, query string, args ...any) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRow(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query linked-note count: %v", err)
	}
	return count
}

func installLinkedNoteFailureTrigger(
	t testing.TB,
	harness *appsupport.StoreHarness,
	table string,
	suffix string,
) {
	t.Helper()
	functionName := pgx.Identifier{"artifacts_linked_note_fail_" + suffix}.Sanitize()
	triggerName := pgx.Identifier{"artifacts_linked_note_trigger_" + suffix}.Sanitize()
	tableName := pgx.Identifier{table}.Sanitize()
	if _, err := harness.DB.Exec(context.Background(), fmt.Sprintf(`
CREATE FUNCTION %s() RETURNS trigger
LANGUAGE plpgsql
AS $function$
BEGIN
    RAISE EXCEPTION 'forced linked-note %s failure';
END
$function$
`, functionName, suffix)); err != nil {
		t.Fatalf("create %s failure function: %v", suffix, err)
	}
	if _, err := harness.DB.Exec(context.Background(), fmt.Sprintf(`
CREATE TRIGGER %s
BEFORE INSERT OR UPDATE ON %s
FOR EACH ROW EXECUTE FUNCTION %s()
`, triggerName, tableName, functionName)); err != nil {
		_, _ = harness.DB.Exec(context.Background(), "DROP FUNCTION "+functionName+"()")
		t.Fatalf("create %s failure trigger: %v", suffix, err)
	}
	t.Cleanup(func() {
		_, _ = harness.DB.Exec(
			context.Background(),
			"DROP TRIGGER IF EXISTS "+triggerName+" ON "+tableName,
		)
		_, _ = harness.DB.Exec(
			context.Background(),
			"DROP FUNCTION IF EXISTS "+functionName+"()",
		)
	})
}
