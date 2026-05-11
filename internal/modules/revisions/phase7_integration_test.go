package revisions_test

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/timelinetest"
)

func TestPhase7_DeleteRestoreRollbackAtomicConsequences_I_7_01(t *testing.T) {
	harness := phase4test.StartServer(t, "phase7-i-7-01-delete-restore")
	login, actorID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedPhase7Record(t, harness.DB, harness.Server, login, actorID, "IR-P7-I701")
	seedHostProjection(t, harness.DB, incidentID, recordID)

	indicatorID := seedIndicatorRecord(t, harness.DB, incidentID, actorID)
	hubChanges, unsubscribe := harness.Server.Runtime.WSHub.SubscribeRecordChanges(8)
	defer unsubscribe()

	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 13, 0, 0, 0, time.UTC))
	deletePayload := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, recordID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-i-7-01-delete-host"}), http.StatusOK)["data"].(map[string]any)
	requireDeleteRestoreRecordChange(t, timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second), recordID, 2, "remove", "cartulary.view.hosts.v1")
	deleteChangeSetID := deletePayload["change_set_id"].(string)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'records.delete' AND actor_user_id = $2`, deleteChangeSetID, actorID) != 1 {
		t.Fatalf("delete did not create attributed change_set")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND target_kind = 'record' AND target_id = $2 AND operation_kind = 'soft_delete'`, deleteChangeSetID, recordID.String()) != 1 {
		t.Fatalf("delete did not create reversible soft_delete mutation")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1 AND record_id = $2 AND row_version = 2`, deleteChangeSetID, recordID) != 1 {
		t.Fatalf("delete did not append row revision")
	}
	if rows := phase4test.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), "cartulary.view.hosts.v1", login); len(rows) != 0 {
		t.Fatalf("deleted host remained in ordinary view rows: %#v", rows)
	}
	historyAfterDelete := historyItems(getHistory(t, harness.Server.HTTP.URL, login, recordID, ""))
	if historyAfterDelete[0].(map[string]any)["operation"] != "soft_delete" {
		t.Fatalf("latest history item must be soft_delete, got %#v", historyAfterDelete[0])
	}

	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 13, 1, 0, 0, time.UTC))
	restorePayload := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-i-7-01-restore-host"}), http.StatusOK)["data"].(map[string]any)
	requireDeleteRestoreRecordChange(t, timelinetest.AwaitRecordChange(t, hubChanges, 5*time.Second), recordID, 3, "invalidate", "cartulary.view.hosts.v1")
	restoreChangeSetID := restorePayload["change_set_id"].(string)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'records.restore' AND actor_user_id = $2`, restoreChangeSetID, actorID) != 1 {
		t.Fatalf("restore did not create attributed change_set")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND target_kind = 'record' AND target_id = $2 AND operation_kind = 'restore'`, restoreChangeSetID, recordID.String()) != 1 {
		t.Fatalf("restore did not create reversible restore mutation")
	}
	if rows := phase4test.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), "cartulary.view.hosts.v1", login); len(rows) != 1 || rows[0]["record_id"] != recordID.String() {
		t.Fatalf("restored host was not returned to ordinary view rows: %#v", rows)
	}
	historyAfterRestore := historyItems(getHistory(t, harness.Server.HTTP.URL, login, recordID, ""))
	if historyAfterRestore[0].(map[string]any)["operation"] != "restore" || historyAfterRestore[1].(map[string]any)["operation"] != "soft_delete" {
		t.Fatalf("delete/restore history was not append-only newest-first: %#v", historyAfterRestore)
	}

	indicatorDelete := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, indicatorID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-i-7-01-delete-indicator"}), http.StatusOK)["data"].(map[string]any)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM indicators WHERE record_id = $1 AND deleted_at IS NOT NULL AND deleted_by_user_id = $2`, indicatorID, actorID) != 1 {
		t.Fatalf("indicator source tombstone was not set")
	}
	httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, indicatorID, map[string]any{"base_row_version": int64(indicatorDelete["row_version"].(float64)), "client_txn_id": "txn-i-7-01-restore-indicator"}), http.StatusOK)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM indicators WHERE record_id = $1 AND deleted_at IS NULL AND deleted_by_user_id IS NULL`, indicatorID) != 1 {
		t.Fatalf("indicator source tombstone was not cleared")
	}
}

func TestPhase7_HistoryPaginationRecordBinding_I_7_02(t *testing.T) {
	harness := phase4test.StartServer(t, "phase7-i-7-02-pagination")
	login, actorID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordA := seedPhase7Record(t, harness.DB, harness.Server, login, actorID, "IR-P7-I702A")
	incidentB, recordB := seedPhase7Record(t, harness.DB, harness.Server, login, actorID, "IR-P7-I702B")
	base := time.Date(2026, 5, 10, 15, 0, 0, 0, time.UTC)
	firstChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000301")
	secondChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000302")
	thirdChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000303")

	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordA, ChangeSetID: firstChangeSet,
		CreatedAt: base, Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "host", Operation: "oldest", RowVersion: 2,
	})
	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordA, ChangeSetID: secondChangeSet,
		CreatedAt: base.Add(time.Minute), Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "record", Operation: "middle", RowVersion: 3,
	})
	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordA, ChangeSetID: thirdChangeSet,
		CreatedAt: base.Add(2 * time.Minute), Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "identity", Operation: "newest", RowVersion: 4,
	})
	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentB, ActorID: actorID, RecordID: recordB, ChangeSetID: mustUUID(t, "77777777-0000-4000-8000-000000000304"),
		CreatedAt: base.Add(3 * time.Minute), Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "host", Operation: "other-record", RowVersion: 2,
	})

	firstPage := getHistory(t, harness.Server.HTTP.URL, login, recordA, "?limit=1")
	firstItems := historyItems(firstPage)
	if len(firstItems) != 1 || firstItems[0].(map[string]any)["change_set_id"] != thirdChangeSet.String() {
		t.Fatalf("first page did not preserve newest-first order: %#v", firstItems)
	}
	paging := firstPage["meta"].(map[string]any)["paging"].(map[string]any)
	cursor := paging["next_cursor"].(string)
	if paging["limit"] != float64(1) || paging["has_more"] != true || cursor == "" {
		t.Fatalf("unexpected first page cursor metadata: %#v", paging)
	}

	secondPage := getHistory(t, harness.Server.HTTP.URL, login, recordA, "?cursor_token="+cursor)
	secondItems := historyItems(secondPage)
	if len(secondItems) != 1 || secondItems[0].(map[string]any)["change_set_id"] != secondChangeSet.String() {
		t.Fatalf("second page did not preserve order: %#v", secondItems)
	}

	crossRecord := phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/records/"+recordB.String()+"/history?cursor_token="+cursor, nil, phase4test.WithCookies(login.SessionCookie))
	errBody := httptestx.RequireErrorEnvelope(t, crossRecord, http.StatusBadRequest, "invalid_pagination_request")
	if errBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "invalid_cursor_token" {
		t.Fatalf("unexpected cross-record cursor reason: %#v", errBody)
	}

	for _, query := range []string{"?page=1", "?offset=1", "?page_size=1", "?block_size=1", "?limit=0"} {
		resp := phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/records/"+recordA.String()+"/history"+query, nil, phase4test.WithCookies(login.SessionCookie))
		httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_pagination_request")
	}
	for _, query := range []string{"?limit=-1", "?limit=abc", "?limit=501"} {
		resp := phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/records/"+recordA.String()+"/history"+query, nil, phase4test.WithCookies(login.SessionCookie))
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_pagination_request")
		if body["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "invalid_limit" {
			t.Fatalf("unexpected invalid limit reason for %s: %#v", query, body)
		}
	}
}

func TestPhase7_StaleRestoreRollbackFailsClosed_I_7_03(t *testing.T) {
	harness := phase4test.StartServer(t, "phase7-i-7-03-stale-restore")
	login, actorID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedPhase7Record(t, harness.DB, harness.Server, login, actorID, "IR-P7-I703")
	seedHostProjection(t, harness.DB, incidentID, recordID)

	deletePayload := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, recordID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-i-7-03-delete"}), http.StatusOK)["data"].(map[string]any)
	if deletePayload["row_version"] != float64(2) {
		t.Fatalf("unexpected delete payload: %#v", deletePayload)
	}
	before := phase7StateCounts(t, harness.DB, recordID)
	hubChanges, unsubscribe := harness.Server.Runtime.WSHub.SubscribeRecordChanges(4)
	defer unsubscribe()
	stale := restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-i-7-03-stale-restore"})
	httptestx.RequireErrorEnvelope(t, stale, http.StatusConflict, "row_version_conflict")
	timelinetest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
	after := phase7StateCounts(t, harness.DB, recordID)
	if before != after {
		t.Fatalf("stale restore mutated state: before=%+v after=%+v", before, after)
	}
}

func requireDeleteRestoreRecordChange(t testing.TB, change platformws.RecordChange, recordID uuid.UUID, rowVersion int64, changeKind string, viewSchemaID string) {
	t.Helper()
	if change.RecordID != recordID || change.RowVersion != rowVersion || change.ChangeKind != changeKind || change.ViewSchemaID != viewSchemaID {
		t.Fatalf("unexpected record_changed event: %+v", change)
	}
	if len(change.ChangedFieldKeys) != 0 {
		t.Fatalf("delete/restore changed_field_keys must be present and empty, got %#v", change.ChangedFieldKeys)
	}
	payload := platformws.RecordChangePayload(change)
	affectedViews, ok := payload["affected_views"].([]map[string]any)
	if !ok || len(affectedViews) != 1 {
		t.Fatalf("delete/restore affected_views must be a single view, got %#v", payload["affected_views"])
	}
	if affectedViews[0]["view_schema_id"] != viewSchemaID || affectedViews[0]["change_kind"] != changeKind {
		t.Fatalf("unexpected affected view payload: %#v", affectedViews[0])
	}
}

func TestPhase7_RetainedHistoryAcrossRestartAndClosure_I_7_04(t *testing.T) {
	server, db, env := startPhase7ReusableServer(t, "phase7-i-7-04-restart")
	login, actorID := phase4test.ProvisionBootstrapAdmin(t, server)
	incidentID, recordID := seedPhase7Record(t, db, server, login, actorID, "IR-P7-I704")
	base := time.Date(2026, 5, 10, 16, 0, 0, 0, time.UTC)
	originalChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000401")

	seedHistoryChangeSet(t, db, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: originalChangeSet,
		CreatedAt: base, Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "host", Operation: "field_update", RowVersion: 2,
	})
	refBefore := stringField(t, historyItems(getHistory(t, server.HTTP.URL, login, recordID, ""))[0], "history_entry_ref")
	server.Close()

	restarted := httptestx.StartServer(t, httptestx.ServerOptions{Env: env})
	refAfterRestart := stringField(t, historyItems(getHistory(t, restarted.HTTP.URL, login, recordID, ""))[0], "history_entry_ref")
	if refAfterRestart != refBefore {
		t.Fatalf("history_entry_ref changed across restart: before=%q after=%q", refBefore, refAfterRestart)
	}
	if _, err := db.Exec(`
UPDATE incidents
   SET closed_at = $1
 WHERE id = $2
`, base.Add(time.Minute), incidentID); err != nil {
		t.Fatalf("seed incident closure: %v", err)
	}
	seedHistoryChangeSet(t, db, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: mustUUID(t, "77777777-0000-4000-8000-000000000402"),
		CreatedAt: base.Add(2 * time.Minute), Source: "rollback", SequenceNo: 1,
		TargetKind: "host", Operation: "rollback", RowVersion: 3,
	})
	items := collectHistoryPages(t, restarted.HTTP.URL, login, recordID, 1)
	if len(items) != 2 {
		t.Fatalf("expected full retained history after restart and later change: %#v", items)
	}
	if stringField(t, items[1], "history_entry_ref") != refBefore {
		t.Fatalf("older selector changed after closure and later change: %#v", items[1])
	}
}

func startPhase7ReusableServer(t testing.TB, prefix string) (*httptestx.Server, *sql.DB, map[string]string) {
	t.Helper()
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PreparePackageDatabaseT(t, prefix)
	s3Harness := s3test.Start(t)
	bucket := s3Harness.PreparePackageBucketT(t, prefix)

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env})
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open reusable sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return server, db, env
}
