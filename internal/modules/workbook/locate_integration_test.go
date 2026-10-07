package workbook_test

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func TestWorkbookLocatorBoundedWindowAndOrdinaryCursor_Integration(t *testing.T) {
	h, login, actorID, incidentID := ConflictFixture(t, "workbook-locator", "IR-LOCATOR")
	ids := make([]string, 106)
	for i := range ids {
		row := CreateNote(t, h, login, incidentID, fmt.Sprintf("locate-%d", i), fmt.Sprintf("Note %03d", i), "Source body")
		ids[i] = row["record_id"].(string)
	}
	route := h.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + artifacts.NotesViewSchemaID
	sort := []map[string]any{{"field_key": "note.title", "direction": "asc"}}
	request := func(path string, body any) map[string]any {
		t.Helper()
		return httptestx.RequireSuccessEnvelope(t, appsupport.DoJSON(t, http.MethodPost, route+path, body, appsupport.WithCookies(login.SessionCookie)), 200)
	}
	located := request("/locate", map[string]any{"record_id": ids[3], "sort": sort})
	data := located["data"].(map[string]any)
	rows := data["rows"].([]any)
	if data["outcome"] != "located" || len(rows) != 100 || rows[0].(map[string]any)["record_id"] != ids[3] || rows[99].(map[string]any)["record_id"] != ids[102] {
		t.Fatalf("unexpected bounded locator page: %v, %d", data["outcome"], len(rows))
	}
	cursor := data["window_start_cursor"].(string)
	reread := request("/query", map[string]any{"sort": sort, "cursor_token": cursor})
	if !reflect.DeepEqual(rows, reread["data"].(map[string]any)["rows"]) {
		t.Fatal("ordinary query did not reproduce located window")
	}
	next := located["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor"].(string)
	continuation := request("/query", map[string]any{"sort": sort, "cursor_token": next})["data"].(map[string]any)["rows"].([]any)
	if len(continuation) != 3 || continuation[0].(map[string]any)["record_id"] != ids[103] {
		t.Fatalf("invalid continuation: %#v", continuation)
	}
	first := request("/locate", map[string]any{"record_id": ids[0], "sort": sort})["data"].(map[string]any)
	if first["window_start_cursor"] != nil {
		t.Fatal("first record requires a null start cursor")
	}
	outside := request("/locate", map[string]any{"record_id": ids[3], "filters": []map[string]any{{"field_key": "note.created_by_user_id", "op": "eq", "arg": map[string]any{"value": uuid.NewString()}}}})["data"]
	if !reflect.DeepEqual(outside, map[string]any{"outcome": "outside_query", "target_record_id": ids[3]}) {
		t.Fatalf("outside response disclosed rows: %#v", outside)
	}
	unavailable := request("/locate", map[string]any{"record_id": uuid.NewString()})["data"]
	if !reflect.DeepEqual(unavailable, map[string]any{"outcome": "unavailable"}) {
		t.Fatalf("unavailable response disclosed target: %#v", unavailable)
	}
	httptestx.RequireErrorEnvelope(t, appsupport.DoJSON(t, http.MethodPost, route+"/locate", map[string]any{"record_id": ids[3]}), 401, "session_required")
	for _, resource := range viewschema.ListPublicResources() {
		url := h.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + resource.ViewSchemaID + "/locate"
		missing := httptestx.RequireSuccessEnvelope(t, appsupport.DoJSON(t, http.MethodPost, url, map[string]any{"record_id": uuid.NewString()}, appsupport.WithCookies(login.SessionCookie)), 200)
		if !reflect.DeepEqual(missing["data"], map[string]any{"outcome": "unavailable"}) {
			t.Fatalf("%s exposed unavailable metadata: %#v", resource.ViewSchemaID, missing)
		}
	}
	for name, body := range map[string]map[string]any{
		"query": {"cursor_token": cursor, "sort": []map[string]any{{"field_key": "note.title", "direction": "desc"}}},
		"limit": {"cursor_token": cursor, "sort": sort, "limit": 99},
	} {
		failed := httptestx.RequireErrorEnvelope(t, appsupport.DoJSON(t, http.MethodPost, route+"/query", body, appsupport.WithCookies(login.SessionCookie)), 400, "invalid_view_query")
		if httptestx.RequireErrorDetails(t, failed)["reason_code"] != "cursor_query_mismatch" {
			t.Fatalf("%s did not bind locator cursor", name)
		}
	}
	// The separate entity provider supports a grouped, mixed-direction order and ties.
	for i := 0; i < 3; i++ {
		id := uuid.New()
		seedHostForPaging(t, h, incidentID, actorID, id, fmt.Sprintf("Host %d", i))
		updateHostDisplayNameForPaging(t, h, id, "Same host", 2)
	}
	hostRoute := h.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + entitycontract.HostsViewSchemaID
	hostQuery := map[string]any{"sort": []map[string]any{{"field_key": "host.display_name", "direction": "desc"}, {"field_key": "host.hostname", "direction": "asc"}}, "group_by": "host.host_state"}
	hostRows := queryWorkbook(t, h, login, hostRoute+"/query", hostQuery)["data"].(map[string]any)["rows"].([]any)
	hostQuery["record_id"] = hostRows[1].(map[string]any)["record_id"]
	hostLocated := httptestx.RequireSuccessEnvelope(t, appsupport.DoJSON(t, http.MethodPost, hostRoute+"/locate", hostQuery, appsupport.WithCookies(login.SessionCookie)), 200)["data"].(map[string]any)
	if !reflect.DeepEqual(hostLocated["rows"], hostRows[1:]) {
		t.Fatal("grouped entity locator lost tied order")
	}
	// Null source time stays null while the row remains locatable in chronology.
	source := CreateTimelineRow(t, h, login, incidentID, "locator-null-time", "Uncertain source time")
	timelineURL := h.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + TimelineViewSchemaID + "/locate"
	timeline := httptestx.RequireSuccessEnvelope(t, appsupport.DoJSON(t, http.MethodPost, timelineURL, map[string]any{"record_id": source["record_id"]}, appsupport.WithCookies(login.SessionCookie)), 200)["data"].(map[string]any)
	if timeline["outcome"] != "located" || timeline["window_start_cursor"] != nil {
		t.Fatal("null-time Timeline row was not located at its authorized chronology position")
	}

	// Cursor positions remain values, even after their predecessor no longer occupies that position.
	requireWorkbookPatch(t, h, login, appsupport.MustUUID(t, ids[2]), map[string]any{"view_schema_id": artifacts.NotesViewSchemaID, "base_row_version": 1, "client_txn_id": "move-predecessor", "changes": []map[string]any{{"field_key": "note.title", "value": "AAA"}}})
	afterMove := request("/query", map[string]any{"sort": sort, "cursor_token": cursor})["data"].(map[string]any)["rows"].([]any)
	if afterMove[0].(map[string]any)["record_id"] != ids[3] {
		t.Fatal("cursor replay looked up the moved predecessor")
	}
}
