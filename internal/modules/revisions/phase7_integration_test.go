package revisions_test

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestPhase7_DeleteRestoreRollbackAtomicConsequences_I_7_01(t *testing.T) {
	requirePhase7LaterSprintScope(t, "I-7-01", "atomic delete, restore, rollback, projection, history, and collaboration consequences")
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
	requirePhase7LaterSprintScope(t, "I-7-03", "stale restore or rollback preconditions failing closed")
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
