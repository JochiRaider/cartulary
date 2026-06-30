package workbook_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

const phase6TimelineViewSchemaID = "cartulary.view.timeline.v2"

func TestPhase6_ConcurrentEditsResolverPath_I_6_03(t *testing.T) {
	harness := phase4test.StartServer(t, "phase6-i-6-03-concurrent-resolver")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase6-i-6-03-incident",
		"incident_key":  "IR-PHASE6-I-6-03",
		"title":         "Phase 6 I-6-03 concurrent resolver path",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	firstUser := phase4test.SeedLocalUserFlags(t, harness.DB, "phase6-i-6-03-first@example.test", "Phase 6 First", "Phase6FirstPass1!", false, false, true)
	secondUser := phase4test.SeedLocalUserFlags(t, harness.DB, "phase6-i-6-03-second@example.test", "Phase 6 Second", "Phase6SecondPass1!", false, false, true)
	phase4test.SeedIncidentMembership(t, harness.DB, incidentID, firstUser.ID, firstUser.DisplayName, "editor", adminUserID)
	phase4test.SeedIncidentMembership(t, harness.DB, incidentID, secondUser.ID, secondUser.DisplayName, "editor", adminUserID)
	firstLogin := phase6LoginLocalUserNoMFA(t, harness, firstUser.Email, "Phase6FirstPass1!")
	secondLogin := phase6LoginLocalUserNoMFA(t, harness, secondUser.Email, "Phase6SecondPass1!")

	differentRow := phase6CreateTimelineRow(t, harness, firstLogin, incidentID, "txn-phase6-i-6-03-different-create", "Different base")
	differentID := phase4test.MustUUID(t, differentRow["record_id"].(string))
	summaryPatch := requireWorkbookPatch(t, harness, firstLogin, differentID, map[string]any{
		"view_schema_id":   phase6TimelineViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-i-6-03-different-summary",
		"changes": []map[string]any{{
			"field_key": "timeline.activity_synopsis_text",
			"value":     "Different summary",
		}},
	})
	requireCellValue(t, summaryPatch["row"].(map[string]any), "timeline.activity_synopsis_text", "Different summary")
	detailsPatch := requireWorkbookPatch(t, harness, secondLogin, differentID, map[string]any{
		"view_schema_id":   phase6TimelineViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-i-6-03-different-details",
		"changes": []map[string]any{{
			"field_key": "timeline.raw_activity_text",
			"value":     "Different details",
		}},
	})
	differentAfter := detailsPatch["row"].(map[string]any)
	requireCellValue(t, differentAfter, "timeline.activity_synopsis_text", "Different summary")
	requireCellValue(t, differentAfter, "timeline.raw_activity_text", "Different details")
	if got := int64(differentAfter["row_version"].(float64)); got != 3 {
		t.Fatalf("different-field stale edit row_version = %d want 3", got)
	}

	keepRow := phase6CreateTimelineRow(t, harness, firstLogin, incidentID, "txn-phase6-i-6-03-keep-create", "Keep base")
	keepID := phase4test.MustUUID(t, keepRow["record_id"].(string))
	keepConflict := phase6CreateTimelineSameFieldConflict(t, harness, firstLogin, secondLogin, keepID, "keep", "Keep saved", "Keep local")
	if token, ok := keepConflict["conflict_token"].(string); !ok || token == "" {
		t.Fatalf("test received empty timeline conflict token: %q", keepConflict["conflict_token"])
	}
	phase6RequireConflictValues(t, keepConflict, keepID, "Keep base", "Keep saved", "Keep local")
	beforeClearChanges := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID)
	beforeClearRevisions := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, keepID)
	keepData := phase6ResolveConflict(t, harness, secondLogin, keepID, keepConflict["conflict_token"].(string), map[string]any{
		"conflict_token":  keepConflict["conflict_token"].(string),
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-phase6-i-6-03-keep-resolve",
	})
	phase6RequireNoChangeSet(t, keepData)
	requireCellValue(t, keepData["row"].(map[string]any), "timeline.activity_synopsis_text", "Keep saved")
	phase6RequireCount(t, "keep_saved change_sets", phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID), beforeClearChanges)
	phase6RequireCount(t, "keep_saved revisions", phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, keepID), beforeClearRevisions)

	useRow := phase6CreateTimelineRow(t, harness, firstLogin, incidentID, "txn-phase6-i-6-03-use-create", "Use base")
	useID := phase4test.MustUUID(t, useRow["record_id"].(string))
	useConflict := phase6CreateTimelineSameFieldConflict(t, harness, firstLogin, secondLogin, useID, "use", "Use saved", "Use local")
	socket := incidentwstest.ConnectAndHello(t, harness.Server.HTTP.URL, incidentID.String(), incidentwstest.ConnectOptions{
		SessionToken:     adminLogin.SessionCookie.Value,
		ClientInstanceID: "phase6-i-6-03-record-change-listener",
		Presence: platformws.PresenceInput{
			SheetRef: map[string]string{"kind": "view_schema", "id": phase6TimelineViewSchemaID},
			Mode:     "viewing",
		},
	})
	defer socket.Close(websocket.StatusNormalClosure, "test_complete")
	beforeUseChanges := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID)
	beforeUseRevisions := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, useID)
	useData := phase6ResolveConflict(t, harness, secondLogin, useID, useConflict["conflict_token"].(string), map[string]any{
		"conflict_token":  useConflict["conflict_token"].(string),
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-phase6-i-6-03-use-resolve",
		"resolved_value":  "Use local",
	})
	requireCellValue(t, useData["row"].(map[string]any), "timeline.activity_synopsis_text", "Use local")
	phase6RequireCount(t, "use_unsaved change_sets", phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID), beforeUseChanges+1)
	phase6RequireCount(t, "use_unsaved revisions", phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, useID), beforeUseRevisions+1)
	phase4test.RequireChangeSetAttribution(t, harness.DB, useData["change_set_id"].(string), secondUser.ID.String(), "timeline.records.conflicts.resolve", "txn-phase6-i-6-03-use-resolve")
	phase6RequireRecordChanged(t, socket, useID, int64(useData["row"].(map[string]any)["row_version"].(float64)))
	phase6ResolveConflict(t, harness, secondLogin, useID, useConflict["conflict_token"].(string), map[string]any{
		"conflict_token":  useConflict["conflict_token"].(string),
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-phase6-i-6-03-use-resolve",
		"resolved_value":  "Use local",
	})
	phase6RequireCount(t, "use_unsaved replay change_sets", phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID), beforeUseChanges+1)
	phase6RequireCount(t, "use_unsaved replay revisions", phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, useID), beforeUseRevisions+1)

	mergedRow := phase6CreateTimelineRow(t, harness, firstLogin, incidentID, "txn-phase6-i-6-03-merged-create", "Merged base")
	mergedID := phase4test.MustUUID(t, mergedRow["record_id"].(string))
	mergedConflict := phase6CreateTimelineSameFieldConflict(t, harness, firstLogin, secondLogin, mergedID, "merged", "Merged saved", "Merged local")
	beforeMergedChanges := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID)
	beforeMergedRevisions := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, mergedID)
	mergedData := phase6ResolveConflict(t, harness, secondLogin, mergedID, mergedConflict["conflict_token"].(string), map[string]any{
		"conflict_token":  mergedConflict["conflict_token"].(string),
		"resolution_kind": "merged_value",
		"client_txn_id":   "txn-phase6-i-6-03-merged-resolve",
		"resolved_value":  "Merged final",
	})
	requireCellValue(t, mergedData["row"].(map[string]any), "timeline.activity_synopsis_text", "Merged final")
	phase6RequireCount(t, "merged_value change_sets", phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID), beforeMergedChanges+1)
	phase6RequireCount(t, "merged_value revisions", phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, mergedID), beforeMergedRevisions+1)
	phase4test.RequireChangeSetAttribution(t, harness.DB, mergedData["change_set_id"].(string), secondUser.ID.String(), "timeline.records.conflicts.resolve", "txn-phase6-i-6-03-merged-resolve")
	phase6RequireRecordChanged(t, socket, mergedID, int64(mergedData["row"].(map[string]any)["row_version"].(float64)))
}

func phase6CreateTimelineRow(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, clientTxnID string, summary string) map[string]any {
	t.Helper()
	data := requireWorkbookCreate(t, harness, login, incidentID, phase6TimelineViewSchemaID, map[string]any{
		"client_txn_id":                   clientTxnID,
		"timeline.activity_synopsis_text": summary,
	})
	return data["row"].(map[string]any)
}

func phase6CreateTimelineSameFieldConflict(t testing.TB, harness *phase4test.ServerHarness, firstLogin phase4test.LoginResult, secondLogin phase4test.LoginResult, recordID uuid.UUID, prefix string, savedValue string, localValue string) map[string]any {
	t.Helper()
	requireWorkbookPatch(t, harness, firstLogin, recordID, map[string]any{
		"view_schema_id":   phase6TimelineViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-i-6-03-" + prefix + "-server",
		"changes": []map[string]any{{
			"field_key": "timeline.activity_synopsis_text",
			"value":     savedValue,
		}},
	})
	resp := doWorkbookJSON(t, harness, secondLogin, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6TimelineViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-i-6-03-" + prefix + "-client",
		"changes": []map[string]any{{
			"field_key": "timeline.activity_synopsis_text",
			"value":     localValue,
		}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	return body["error"].(map[string]any)["conflict"].(map[string]any)
}

func phase6ResolveConflict(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, recordID uuid.UUID, conflictToken string, body map[string]any) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/conflicts/"+conflictToken+"/resolve",
		body,
		phase4test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resolve conflict failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func phase6LoginLocalUserNoMFA(t testing.TB, harness *phase4test.ServerHarness, username string, password string) phase4test.LoginResult {
	t.Helper()
	resp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected login cookies, got %#v", resp.Cookies())
	}
	return phase4test.LoginResult{SessionCookie: sessionCookie, CSRFCookie: csrfCookie}
}

func phase6RequireConflictValues(t testing.TB, conflict map[string]any, recordID uuid.UUID, baseValue string, serverValue string, clientValue string) {
	t.Helper()
	if conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != "timeline.activity_synopsis_text" ||
		conflict["conflict_resolution_class"] != "text_compare_merge" ||
		conflict["base_value"] != baseValue ||
		conflict["server_value"] != serverValue ||
		conflict["client_value"] != clientValue ||
		conflict["conflict_token"] == "" {
		t.Fatalf("unexpected same-field conflict payload: %#v", conflict)
	}
}

func phase6RequireNoChangeSet(t testing.TB, data map[string]any) {
	t.Helper()
	if _, ok := data["change_set_id"]; ok {
		t.Fatalf("keep_saved response must not carry change_set_id: %#v", data)
	}
}

func phase6RequireCount(t testing.TB, label string, got int, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s count = %d want %d", label, got, want)
	}
}

func phase6RequireRecordChanged(t testing.TB, client *incidentwstest.Client, recordID uuid.UUID, rowVersion int64) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		message, err := client.AwaitNextMessage(time.Until(deadline))
		if err != nil {
			t.Fatalf("wait for record_changed for %s: %v", recordID, err)
		}
		if message.Type != "record_changed" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			t.Fatalf("decode record_changed payload: %v", err)
		}
		if payload["record_id"] == recordID.String() && int64(payload["row_version"].(float64)) == rowVersion {
			return
		}
	}
	t.Fatalf("timed out waiting for record_changed for %s version %d", recordID, rowVersion)
}
