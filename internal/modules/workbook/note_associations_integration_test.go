package workbook_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestNoteAssociationsPublicPagingAndHistory(t *testing.T) {
	h, login, _, incidentID := ConflictFixture(t, "note-association-public-contract", "IR-NOTE-ASSOCIATION-API")
	note := CreateNote(t, h, login, incidentID, "association-api-note", "Current note", "Body")
	noteID := appsupport.MustUUID(t, note["record_id"].(string))
	prefix := h.Server.HTTP.URL + "/api/v1/records/" + noteID.String()
	route := prefix + "/note-associations"
	request := func(method, path string, body any) *http.Response {
		t.Helper()
		return appsupport.DoJSON(t, method, path, body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	}
	get := func(query string) map[string]any {
		t.Helper()
		return httptestx.RequireSuccessEnvelope(t, request(http.MethodGet, route+query, nil), http.StatusOK)["data"].(map[string]any)
	}
	write := func(body map[string]any) map[string]any {
		t.Helper()
		return httptestx.RequireSuccessEnvelope(t, request(http.MethodPost, route, body), http.StatusOK)["data"].(map[string]any)
	}
	httptestx.RequireErrorEnvelope(t, appsupport.DoJSON(t, http.MethodGet, route+"?kind=source", nil), http.StatusUnauthorized, "session_required")
	httptestx.RequireErrorEnvelope(t, appsupport.DoJSON(t, http.MethodPost, route, map[string]any{}, appsupport.WithCookies(login.SessionCookie)), http.StatusForbidden, "csrf_verification_failed")
	httptestx.RequireErrorEnvelope(t, request(http.MethodGet, route, nil), http.StatusBadRequest, "invalid_mutation_payload")
	empty := get("?kind=source")
	if len(empty["items"].([]any)) != 0 || empty["next_cursor_token"] != nil {
		t.Fatalf("empty association page: %#v", empty)
	}
	actions := []map[string]any{}
	for i := 0; i < 3; i++ {
		other := CreateNote(t, h, login, incidentID, fmt.Sprintf("association-api-other-%d", i), fmt.Sprintf("Related note %d", i), "Body")
		actions = append(actions, map[string]any{"op": "add", "counterpart_record_id": other["record_id"]})
	}
	payload := map[string]any{"kind": "related_note", "base_row_version": 1, "client_txn_id": "association-api-add", "actions": actions}
	accepted := write(payload)
	if accepted["change_set_id"] == nil || accepted["row"].(map[string]any)["row_version"] != float64(2) {
		t.Fatalf("incomplete mutation receipt: %#v", accepted)
	}
	if replay := write(payload); !reflect.DeepEqual(replay, accepted) {
		t.Fatalf("replay differs: %#v %#v", accepted, replay)
	}
	page := get("?kind=related_note&limit=1")
	item := page["items"].([]any)[0].(map[string]any)
	if item["view_schema_id"] != artifacts.NotesViewSchemaID || item["direction"] != "outgoing" || item["display_label"] == "" {
		t.Fatalf("unsafe association item: %#v", item)
	}
	cursor := page["next_cursor_token"].(string)
	for _, query := range []string{"?kind=source&limit=1&cursor_token=", "?kind=related_note&limit=2&cursor_token="} {
		httptestx.RequireErrorEnvelope(t, request(http.MethodGet, route+query+url.QueryEscape(cursor), nil), http.StatusBadRequest, "invalid_view_query")
	}
	seen := map[string]bool{}
	for {
		for _, raw := range page["items"].([]any) {
			ref := raw.(map[string]any)["item_ref"].(string)
			if seen[ref] {
				t.Fatal("duplicate cursor result")
			}
			seen[ref] = true
		}
		if page["next_cursor_token"] == nil {
			break
		}
		page = get("?kind=related_note&limit=1&cursor_token=" + url.QueryEscape(page["next_cursor_token"].(string)))
	}
	if len(seen) != 3 {
		t.Fatalf("paged %d of 3 Notes", len(seen))
	}
	removed := write(map[string]any{"kind": "related_note", "base_row_version": 2, "client_txn_id": "association-api-remove", "actions": []map[string]any{{"op": "remove", "item_ref": item["item_ref"]}}})
	if removed["row"].(map[string]any)["row_version"] != float64(3) || len(get("?kind=related_note")["items"].([]any)) != 2 {
		t.Fatal("valid removal did not update Note")
	}
	// Roll back the removal through the existing History owner. It must restore
	// the original canonical link identity, not manufacture a replacement link.
	history := httptestx.RequireSuccessEnvelope(t, request(http.MethodGet, prefix+"/history", nil), http.StatusOK)["data"].(map[string]any)
	var historyRef string
	for _, raw := range history["items"].([]any) {
		entry := raw.(map[string]any)
		if entry["change_set_id"] == removed["change_set_id"] {
			for _, unit := range entry["diff_summary"].(map[string]any)["units"].([]any) {
				if unit.(map[string]any)["kind"] == "link" {
					historyRef, _ = entry["history_entry_ref"].(string)
				}
			}
		}
	}
	if historyRef == "" {
		t.Fatalf("missing association history entry: %#v", history)
	}
	rollback := map[string]any{"base_row_version": 3, "client_txn_id": "association-api-rollback", "target": map[string]any{"kind": "history_entry", "history_entry_ref": historyRef}}
	httptestx.RequireSuccessEnvelope(t, request(http.MethodPost, prefix+"/rollback", rollback), http.StatusOK)
	restored := get("?kind=related_note")
	found := false
	for _, raw := range restored["items"].([]any) {
		if raw.(map[string]any)["item_ref"] == item["item_ref"] {
			found = true
		}
	}
	if !found || len(restored["items"].([]any)) != 3 {
		t.Fatalf("rollback lost association identity: %#v", restored)
	}
	// A source-owned projection rebuild must agree with the versioned row.
	row := RequireQueriedRow(t, h, login, incidentID, artifacts.NotesViewSchemaID, noteID)
	requireCellValue(t, row, "note.linked_record_count", float64(3))
	var linkCount int
	if err := h.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM record_links WHERE src_record_id=$1 AND field_key IS NULL AND deleted_at IS NULL`, noteID).Scan(&linkCount); err != nil || linkCount != 3 {
		t.Fatalf("canonical links: %d %v", linkCount, err)
	}
	source := CreateTimelineRow(t, h, login, incidentID, "association-api-timeline", "Authored Timeline source")
	write(map[string]any{"kind": "source", "base_row_version": row["row_version"], "client_txn_id": "association-api-source", "actions": []map[string]any{{"op": "add", "counterpart_record_id": source["record_id"]}}})
	sources := get("?kind=source")["items"].([]any)
	if len(sources) != 1 || sources[0].(map[string]any)["display_label"] != "Authored Timeline source" {
		t.Fatalf("Timeline association must use its authored synopsis: %#v", sources)
	}
}
