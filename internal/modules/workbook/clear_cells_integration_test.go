package workbook_test

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestTimelineClearCellsAuthoritativeBatch_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "timeline-clear-cells")
	login, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{"client_txn_id": "clear-incident", "incident_key": "IR-CLEAR", "title": "Clear cells"})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	fields := mutationpolicy.DirectWritableFieldKeys()
	ids := make([]string, 2)
	for index := range ids {
		body := map[string]any{"client_txn_id": fmt.Sprintf("clear-create-%d", index)}
		for fieldIndex, field := range fields {
			body[field] = []string{"", "  ", "line one\nline two", "Source"}[fieldIndex%4]
		}
		created := requireWorkbookCreate(t, harness, login, incidentID, timeline.TimelineViewSchemaID, body)
		ids[index] = created["row"].(map[string]any)["record_id"].(string)
	}
	request := map[string]any{"view_schema_id": timeline.TimelineViewSchemaID, "client_txn_id": "clear-all", "kind": "clear_cells_v1", "field_keys": fields, "targets": []map[string]any{{"record_id": ids[1], "base_row_version": 1}, {"record_id": ids[0], "base_row_version": 1}}}
	result := requireBulkMutation(t, harness, login, incidentID, timeline.TimelineViewSchemaID, request)
	rows := result["rows"].([]any)
	if len(rows) != 2 || len(result["conflicts"].([]any)) != 0 {
		t.Fatalf("wrong clear result: %#v", result)
	}
	for index, raw := range rows {
		row := raw.(map[string]any)
		if row["record_id"] != ids[1-index] || row["row_version"] != float64(2) {
			t.Fatalf("wrong row order/version: %#v", row)
		}
		for _, field := range fields {
			requireCellValue(t, row, field, nil)
		}
	}
	requireChangeSetSource(t, harness, result["change_set_id"].(string), "workbook.bulk_mutations", "clear-all")
	requireMutationTargets(t, harness, result["change_set_id"].(string), []string{ids[1], ids[0]})
	if count := appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM record_revisions WHERE change_set_id = $1`, result["change_set_id"]); count != 2 {
		t.Fatalf("multi-field batch created %d revisions, want 2", count)
	}
	if replay := requireBulkMutation(t, harness, login, incidentID, timeline.TimelineViewSchemaID, request); !reflect.DeepEqual(replay, result) {
		t.Fatalf("replay changed receipt: %#v", replay)
	}
	request["field_keys"] = []string{fields[0]}
	requireBulkMutationStatus(t, harness, login, incidentID, timeline.TimelineViewSchemaID, request, http.StatusConflict)
	request["field_keys"] = fields
	request["client_txn_id"] = "clear-noop"
	request["targets"] = []map[string]any{{"record_id": ids[0], "base_row_version": 2}}
	noop := requireBulkMutation(t, harness, login, incidentID, timeline.TimelineViewSchemaID, request)
	if len(noop["rows"].([]any)) != 0 || len(noop["conflicts"].([]any)) != 0 || noop["change_set_id"] != nil {
		t.Fatalf("null no-op manufactured changes: %#v", noop)
	}
	requireNoChangeSetForClientTxn(t, harness, "clear-noop")
	var nullCount int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM timeline_events WHERE record_id = $1 AND activity_synopsis_text IS NULL AND raw_activity_text IS NULL AND activity_utc_text IS NULL AND activity_local_text IS NULL`, ids[0]).Scan(&nullCount); err != nil {
		t.Fatal(err)
	}
	if nullCount != 1 {
		t.Fatal("stored values are not SQL null")
	}
	request["client_txn_id"] = "clear-foreign"
	request["targets"] = []map[string]any{{"record_id": ids[0], "base_row_version": 2}, {"record_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "base_row_version": 99}}
	rejected := requireBulkMutationStatus(t, harness, login, incidentID, timeline.TimelineViewSchemaID, request, http.StatusNotFound)
	requireNoVersionOracle(t, rejected)
	requireNoChangeSetForClientTxn(t, harness, "clear-foreign")
	closed := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/close", map[string]any{"base_incident_version": incident["incident_version"], "client_txn_id": "clear-close-incident", "reason": "Clear lifecycle boundary"}, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	httptestx.RequireSuccessEnvelope(t, closed, http.StatusOK)
	request["client_txn_id"] = "clear-closed"
	request["targets"] = []map[string]any{{"record_id": ids[0], "base_row_version": 2}}
	rejected = requireBulkMutationStatus(t, harness, login, incidentID, timeline.TimelineViewSchemaID, request, http.StatusConflict)
	if rejected["error"].(map[string]any)["code"] != "incident_closed" {
		t.Fatalf("wrong closed incident rejection: %#v", rejected)
	}
	requireNoChangeSetForClientTxn(t, harness, "clear-closed")
}

func TestTimelineClearDateIntent_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "timeline-clear-dates")
	login, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{"client_txn_id": "clear-date-incident", "incident_key": "IR-CLEAR-DATE", "title": "Clear dates"})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	response := appsupport.DoJSON(t, http.MethodPut, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/timeline-time-conversion-profile", map[string]any{"base_profile_version": 1, "enabled": true, "local_offset_minutes": -300, "local_label": "UTC-05"}, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
	created := requireWorkbookCreate(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"client_txn_id": "clear-date-create", "timeline.activity_utc_text": "2026-09-18T12:00:00Z", "timeline.activity_synopsis_text": "Keep this source"})
	id := appsupport.MustUUID(t, created["row"].(map[string]any)["record_id"].(string))
	cleared := requireBulkMutation(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"view_schema_id": timeline.TimelineViewSchemaID, "client_txn_id": "clear-date-utc", "kind": "clear_cells_v1", "field_keys": []string{"timeline.activity_utc_text"}, "targets": []map[string]any{{"record_id": id.String(), "base_row_version": 1}}})
	row := cleared["rows"].([]any)[0].(map[string]any)
	requireCellValue(t, row, "timeline.activity_utc_text", nil)
	requireCellValue(t, row, "timeline.activity_local_text", "2026-09-18T07:00:00-05:00")
	requireCellValue(t, row, "timeline.activity_time_pair_state", "conversion_unavailable")
	patch := func(txn, field string, value any, version int) map[string]any {
		return requireWorkbookPatch(t, harness, login, id, map[string]any{"view_schema_id": timeline.TimelineViewSchemaID, "base_row_version": version, "client_txn_id": txn, "changes": []map[string]any{{"field_key": field, "value": value}}})["row"].(map[string]any)
	}
	row = patch("clear-date-unrelated", "timeline.analyst_text", "Someone", 2)
	requireCellValue(t, row, "timeline.activity_utc_text", nil)
	row = patch("clear-date-explicit", "timeline.activity_local_text", "2026-09-18T08:00:00-05:00", 3)
	requireCellValue(t, row, "timeline.activity_utc_text", "2026-09-18T13:00:00Z")
	row = patch("clear-date-patch-null", "timeline.activity_utc_text", nil, 4)
	requireCellValue(t, row, "timeline.activity_utc_text", nil)
	requireCellValue(t, row, "timeline.activity_synopsis_text", "Keep this source")

	// A submitted counterpart remains submitted even when its cell conflicts.
	// Conversion must not overwrite that conflict's saved null as a side effect.
	paired := requireWorkbookCreate(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"client_txn_id": "clear-date-conflict-create", "timeline.activity_utc_text": "2026-09-18T12:00:00Z", "timeline.activity_local_text": "2026-09-18T07:00:00-05:00"})["row"].(map[string]any)
	pairedID := paired["record_id"].(string)
	requireBulkMutation(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"view_schema_id": timeline.TimelineViewSchemaID, "client_txn_id": "clear-date-conflict-local", "kind": "clear_cells_v1", "field_keys": []string{"timeline.activity_local_text"}, "targets": []map[string]any{{"record_id": pairedID, "base_row_version": 1}}})
	mixed := requireClipboardPaste(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"view_schema_id": timeline.TimelineViewSchemaID, "client_txn_id": "clear-date-conflict-paste", "clipboard_text": "2026-09-18T13:00:00Z\t2026-09-18T08:00:00-05:00", "format": "tsv", "header_mode": "none", "start_field_key": "timeline.activity_utc_text", "columns": []string{"timeline.activity_utc_text", "timeline.activity_local_text"}, "targets": []map[string]any{{"kind": "record", "record_id": pairedID, "base_row_version": 1}}}, http.StatusOK)
	if len(mixed["rows"].([]any)) != 1 || len(mixed["conflicts"].([]any)) != 1 {
		t.Fatalf("wrong date mixed result: %#v", mixed)
	}
	mixedRow := mixed["rows"].([]any)[0].(map[string]any)
	requireCellValue(t, mixedRow, "timeline.activity_utc_text", "2026-09-18T13:00:00Z")
	requireCellValue(t, mixedRow, "timeline.activity_local_text", nil)
	requireCellValue(t, mixedRow, "timeline.activity_time_pair_state", "conversion_unavailable")
}

func TestTimelineClearCellsConflictsRetainNull_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "timeline-clear-conflicts")
	login, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{"client_txn_id": "clear-conflict-incident", "incident_key": "IR-CLEAR-CONFLICT", "title": "Clear conflict"})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	created := requireWorkbookCreate(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"client_txn_id": "clear-conflict-create", "timeline.activity_synopsis_text": "Base", "timeline.raw_activity_text": "Source"})
	id := appsupport.MustUUID(t, created["row"].(map[string]any)["record_id"].(string))
	requireWorkbookPatch(t, harness, login, id, map[string]any{"view_schema_id": timeline.TimelineViewSchemaID, "base_row_version": 1, "client_txn_id": "clear-conflict-server", "changes": []map[string]any{{"field_key": "timeline.activity_synopsis_text", "value": "Saved"}}})
	request := map[string]any{"view_schema_id": timeline.TimelineViewSchemaID, "kind": "clear_cells_v1", "client_txn_id": "clear-conflict-mixed", "field_keys": []string{"timeline.activity_synopsis_text", "timeline.raw_activity_text"}, "targets": []map[string]any{{"record_id": id.String(), "base_row_version": 1}}}
	result := requireBulkMutation(t, harness, login, incidentID, timeline.TimelineViewSchemaID, request)
	if len(result["rows"].([]any)) != 1 || len(result["conflicts"].([]any)) != 1 {
		t.Fatalf("wrong mixed result: %#v", result)
	}
	row := result["rows"].([]any)[0].(map[string]any)
	requireCellValue(t, row, "timeline.activity_synopsis_text", "Saved")
	requireCellValue(t, row, "timeline.raw_activity_text", nil)
	conflict := result["conflicts"].([]any)[0].(map[string]any)
	if value, present := conflict["client_value"]; !present || value != nil {
		t.Fatalf("conflict lost explicit null: %#v", conflict)
	}
	request["client_txn_id"] = "clear-conflict-only"
	request["field_keys"] = []string{"timeline.activity_synopsis_text"}
	only := requireBulkMutation(t, harness, login, incidentID, timeline.TimelineViewSchemaID, request)
	if len(only["rows"].([]any)) != 0 || len(only["conflicts"].([]any)) != 1 || only["change_set_id"] != nil {
		t.Fatalf("wrong conflicts-only result: %#v", only)
	}
	conflict = only["conflicts"].([]any)[0].(map[string]any)
	resolved := resolveTimelineConflict(t, harness, login, id, conflict["conflict_token"].(string), map[string]any{"conflict_token": conflict["conflict_token"], "resolution_kind": "use_unsaved", "client_txn_id": "clear-null-resolve", "resolved_value": nil})
	requireCellValue(t, resolved["row"].(map[string]any), "timeline.activity_synopsis_text", nil)
}

func TestTimelineClearPreservesRelationshipsAndHistory_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "timeline-clear-history")
	login, actor := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{"client_txn_id": "clear-history-incident", "incident_key": "IR-CLEAR-HISTORY", "title": "Clear history"})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	pasted := requireClipboardPaste(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"view_schema_id": timeline.TimelineViewSchemaID, "client_txn_id": "clear-history-seed", "clipboard_text": "Source\tHost mention\tKeep tag", "format": "tsv", "header_mode": "none", "start_field_key": "timeline.activity_synopsis_text", "columns": []string{"timeline.activity_synopsis_text", "timeline.host_refs", "timeline.tags"}, "targets": []map[string]any{{"kind": "create"}}}, http.StatusOK)
	before := pasted["rows"].([]any)[0].(map[string]any)
	id := before["record_id"].(string)
	review := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/"+id+"/mark-reviewed", map[string]any{"base_row_version": 1, "client_txn_id": "clear-history-review", "reason": "Review source"}, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	httptestx.RequireSuccessEnvelope(t, review, http.StatusOK)
	clear := func(txn string, fields []string, version int) map[string]any {
		return requireBulkMutation(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"view_schema_id": timeline.TimelineViewSchemaID, "kind": "clear_cells_v1", "client_txn_id": txn, "field_keys": fields, "targets": []map[string]any{{"record_id": id, "base_row_version": version}}})
	}
	noop := clear("clear-history-noop", []string{"timeline.raw_activity_text"}, 2)
	if len(noop["rows"].([]any)) != 0 {
		t.Fatalf("already null reviewed row changed: %#v", noop)
	}
	result := clear("clear-history-material", []string{"timeline.activity_synopsis_text"}, 2)
	row := result["rows"].([]any)[0].(map[string]any)
	requireCellValue(t, row, "timeline.capture_state", "enriched")
	for _, field := range []string{"timeline.host_refs", "timeline.identity_refs", "timeline.tags", "timeline.attached_evidence_ids"} {
		if !reflect.DeepEqual(before["cells"].(map[string]any)[field], row["cells"].(map[string]any)[field]) {
			t.Fatalf("clear modified relationship %s", field)
		}
	}
	var beforeValue, afterValue, updatedBy string
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT before_json->'source'->'activity_synopsis_text', after_json->'source'->'activity_synopsis_text' FROM record_revisions WHERE change_set_id = $1 AND record_id = $2`, result["change_set_id"], id).Scan(&beforeValue, &afterValue); err != nil {
		t.Fatal(err)
	}
	if beforeValue != `"Source"` || afterValue != "null" {
		t.Fatalf("history lost exact clear: before=%s after=%s", beforeValue, afterValue)
	}
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT updated_by_user_id FROM timeline_events WHERE record_id = $1`, id).Scan(&updatedBy); err != nil {
		t.Fatal(err)
	}
	if updatedBy != actor.String() {
		t.Fatalf("wrong clear attribution: %s", updatedBy)
	}
	requireRecordTag(t, harness, id, "Keep tag")
	replacement := requireWorkbookCreate(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"client_txn_id": "clear-history-replacement", "timeline.activity_synopsis_text": "Replacement"})["row"].(map[string]any)
	for _, action := range []struct {
		path string
		body map[string]any
	}{
		{"mark-reviewed", map[string]any{"base_row_version": 3, "client_txn_id": "clear-history-rereview", "reason": "Reviewed"}},
		{"supersede", map[string]any{"base_row_version": 4, "client_txn_id": "clear-history-supersede", "reason": "Replaced", "replacement_record_id": replacement["record_id"]}},
	} {
		response := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/"+id+"/"+action.path, action.body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
		httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
	}
	rejected := requireBulkMutationStatus(t, harness, login, incidentID, timeline.TimelineViewSchemaID, map[string]any{"view_schema_id": timeline.TimelineViewSchemaID, "kind": "clear_cells_v1", "client_txn_id": "clear-history-superseded-reject", "field_keys": []string{"timeline.activity_synopsis_text"}, "targets": []map[string]any{{"record_id": replacement["record_id"], "base_row_version": 1}, {"record_id": id, "base_row_version": 1}}}, http.StatusConflict)
	if rejected["error"].(map[string]any)["code"] != "illegal_transition" {
		t.Fatalf("wrong superseded rejection: %#v", rejected)
	}
	requireNoChangeSetForClientTxn(t, harness, "clear-history-superseded-reject")
	if count := appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM records WHERE record_id = $1 AND row_version = 1`, replacement["record_id"]); count != 1 {
		t.Fatal("superseded admission partially changed another row")
	}
}
