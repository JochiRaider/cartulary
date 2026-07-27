package revisions_test

import (
	"context"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestRecordHistoryEnvelope_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, "history_revision-u-7-01-history-envelope")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-U701")
	base := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	newerChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000020")
	olderChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000010")

	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: olderChangeSet,
		CreatedAt: base.Add(time.Minute), Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "host", Operation: "field_update", RowVersion: 2,
	})
	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: newerChangeSet,
		CreatedAt: base.Add(2 * time.Minute), Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "host", Operation: "hostname_update", RowVersion: 4,
	})
	seedHistoryMutation(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: newerChangeSet,
		CreatedAt: base.Add(2 * time.Minute), Source: "workbook.records.patch", SequenceNo: 2,
		TargetKind: "record", Operation: "envelope_update",
	})
	tombstone := base.Add(3 * time.Minute)
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE records
   SET row_version = 5,
       deleted_at = $1,
       deleted_by_user_id = $2
 WHERE record_id = $3
`, tombstone, actorID, recordID); err != nil {
		t.Fatalf("mark record deleted: %v", err)
	}

	body := getHistory(t, harness.Server.HTTP.URL, login, recordID, "")
	data := body["data"].(map[string]any)
	if data["incident_id"] != incidentID.String() || data["record_id"] != recordID.String() {
		t.Fatalf("unexpected history envelope ids: %#v", data)
	}
	if data["row_version"] != float64(5) || data["deleted"] != true {
		t.Fatalf("expected tombstone row_version and deleted=true, got %#v", data)
	}
	paging := body["meta"].(map[string]any)["paging"].(map[string]any)
	if paging["limit"] != float64(100) || paging["has_more"] != false || paging["next_cursor"] != nil {
		t.Fatalf("unexpected default terminal paging: %#v", paging)
	}
	items := data["items"].([]any)
	if len(items) != 3 {
		t.Fatalf("expected three history items, got %#v", items)
	}
	assertHistoryItem(t, items[0], newerChangeSet, "hostname_update", 1, []string{})
	assertHistoryItem(t, items[1], newerChangeSet, "envelope_update", 2, []string{})
	assertHistoryItem(t, items[2], olderChangeSet, "field_update", 1, []string{})

	unauthenticated := appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/history", nil)
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")
}

func TestRecordHistoryOpenAPIContract_Unit(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)

	operation := historyOpenAPIObjectAt(t, document, "paths", "/api/v1/records/{record_id}/history", "get")
	parameters, ok := operation["parameters"].([]any)
	if !ok {
		t.Fatalf("history operation parameters are %T, want array", operation["parameters"])
	}
	requireHistoryOpenAPIParameter(t, parameters, "record_id", "path", true, "uuid")
	requireHistoryOpenAPIParameter(t, parameters, "limit", "query", false, "")
	requireHistoryOpenAPIParameter(t, parameters, "cursor_token", "query", false, "")

	schemas := historyOpenAPIObjectAt(t, document, "components", "schemas")
	envelope := historyOpenAPIObjectAt(t, schemas, "RecordHistoryEnvelope")
	meta := historyOpenAPIObjectAt(t, envelope, "properties", "meta")
	if got := meta["$ref"]; got != "#/components/schemas/RecordHistoryEnvelopeMeta" {
		t.Fatalf("RecordHistoryEnvelope.meta ref = %v, want RecordHistoryEnvelopeMeta", got)
	}
	metaSchema := historyOpenAPIObjectAt(t, schemas, "RecordHistoryEnvelopeMeta")
	required := historyOpenAPIStringArrayAt(t, metaSchema, "required")
	for _, field := range []string{"request_id", "paging"} {
		if !slices.Contains(required, field) {
			t.Fatalf("RecordHistoryEnvelopeMeta missing required field %q: %v", field, required)
		}
	}
	itemSchema := historyOpenAPIObjectAt(t, schemas, "RecordHistoryItem")
	itemRequired := historyOpenAPIStringArrayAt(t, itemSchema, "required")
	for _, field := range []string{"actor_user_id", "committed_at", "history_item_ref", "operation", "diff_summary", "change_set_id", "reversible", "available_rollback_actions"} {
		if !slices.Contains(itemRequired, field) {
			t.Fatalf("RecordHistoryItem missing required field %q: %v", field, itemRequired)
		}
	}
	historyItemRef := historyOpenAPIObjectAt(t, itemSchema, "properties", "history_item_ref")
	if historyItemRef["type"] != "string" {
		t.Fatalf("RecordHistoryItem.history_item_ref type = %v, want string", historyItemRef["type"])
	}
}

func TestHistoryEntryRefStability_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, "history_revision-u-7-02-history-ref")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-U702")
	base := time.Date(2026, 5, 10, 13, 0, 0, 0, time.UTC)
	addressableChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000101")
	unsupportedChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000102")

	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: addressableChangeSet,
		CreatedAt: base.Add(2 * time.Minute), Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "host", Operation: "field_update", RowVersion: 2,
	})
	seedHistoryMutation(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: unsupportedChangeSet,
		CreatedAt: base.Add(time.Minute), Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "record_link", Operation: "link_update",
	})

	first := historyItems(getHistory(t, harness.Server.HTTP.URL, login, recordID, ""))
	second := historyItems(getHistory(t, harness.Server.HTTP.URL, login, recordID, ""))
	ref := stringField(t, first[0], "history_entry_ref")
	itemRef := stringField(t, first[0], "history_item_ref")
	unsupportedItemRef := stringField(t, first[1], "history_item_ref")
	if ref == "" || !strings.HasPrefix(ref, "href_") {
		t.Fatalf("expected opaque href_ selector, got %q", ref)
	}
	if _, err := uuid.Parse(ref); err == nil {
		t.Fatalf("history_entry_ref must not be a raw uuid: %q", ref)
	}
	if ref != stringField(t, second[0], "history_entry_ref") {
		t.Fatalf("history_entry_ref changed across repeated reads: first=%q second=%q", ref, stringField(t, second[0], "history_entry_ref"))
	}
	if itemRef == "" || !strings.HasPrefix(itemRef, "hitem_") {
		t.Fatalf("expected opaque hitem_ display selector, got %q", itemRef)
	}
	if _, err := uuid.Parse(itemRef); err == nil {
		t.Fatalf("history_item_ref must not be a raw uuid: %q", itemRef)
	}
	if itemRef != stringField(t, second[0], "history_item_ref") {
		t.Fatalf("history_item_ref changed across repeated reads: first=%q second=%q", itemRef, stringField(t, second[0], "history_item_ref"))
	}
	if unsupportedItemRef == "" || !strings.HasPrefix(unsupportedItemRef, "hitem_") || unsupportedItemRef == itemRef {
		t.Fatalf("unsupported item history_item_ref missing or not unique: supported=%q unsupported=%q", itemRef, unsupportedItemRef)
	}
	if _, ok := first[1].(map[string]any)["history_entry_ref"]; ok {
		t.Fatalf("unsupported multi-target-style item must not expose history_entry_ref: %#v", first[1])
	}
	assertActions(t, first[1], []string{})
}

func TestRetainedHistoryInvariants_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, "history_revision-u-7-07-retained-history")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-U707")
	base := time.Date(2026, 5, 10, 14, 0, 0, 0, time.UTC)
	originalChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000201")
	laterChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000202")

	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: originalChangeSet,
		CreatedAt: base, Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "host", Operation: "field_update", RowVersion: 2,
	})
	originalItems := historyItems(getHistory(t, harness.Server.HTTP.URL, login, recordID, ""))
	originalRef := stringField(t, originalItems[0], "history_entry_ref")

	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: laterChangeSet,
		CreatedAt: base.Add(time.Minute), Source: "rollback", SequenceNo: 1,
		TargetKind: "host", Operation: "rollback", RowVersion: 3,
	})
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE records
   SET deleted_at = $1,
       deleted_by_user_id = $2,
       row_version = 4
 WHERE record_id = $3
`, base.Add(2*time.Minute), actorID, recordID); err != nil {
		t.Fatalf("seed delete cycle tombstone: %v", err)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE records
   SET deleted_at = NULL,
       deleted_by_user_id = NULL,
       row_version = 5
 WHERE record_id = $1
`, recordID); err != nil {
		t.Fatalf("seed restore cycle: %v", err)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
	UPDATE incidents
	   SET status = 'closed',
	       closed_at = $1
	 WHERE id = $2
	`, base.Add(3*time.Minute), incidentID); err != nil {
		t.Fatalf("seed incident closure: %v", err)
	}

	paged := collectHistoryPages(t, harness.Server.HTTP.URL, login, recordID, 1)
	if len(paged) != 2 {
		t.Fatalf("expected retained paginated history, got %#v", paged)
	}
	older := paged[1].(map[string]any)
	if older["change_set_id"] != originalChangeSet.String() || stringField(t, older, "history_entry_ref") != originalRef {
		t.Fatalf("older history item or selector changed: %#v", older)
	}
	if older["reversible"] != false {
		t.Fatalf("older dependent item should remain visible but no longer reversible: %#v", older)
	}
	assertActions(t, older, []string{})
	requireNoRetainedHistoryNarrowingSurface(t, harness, login, incidentID, recordID)
}

func getHistory(t testing.TB, baseURL string, login appsupport.LoginResult, recordID uuid.UUID, query string) map[string]any {
	t.Helper()
	url := baseURL + "/api/v1/records/" + recordID.String() + "/history" + query
	resp := appsupport.DoJSON(t, http.MethodGet, url, nil, appsupport.WithCookies(login.SessionCookie))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("history request failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func historyItems(body map[string]any) []any {
	return body["data"].(map[string]any)["items"].([]any)
}

func collectHistoryPages(t testing.TB, baseURL string, login appsupport.LoginResult, recordID uuid.UUID, limit int) []any {
	t.Helper()
	query := fmt.Sprintf("?limit=%d", limit)
	collected := make([]any, 0)
	for {
		body := getHistory(t, baseURL, login, recordID, query)
		collected = append(collected, historyItems(body)...)
		paging := body["meta"].(map[string]any)["paging"].(map[string]any)
		cursor, _ := paging["next_cursor"].(string)
		if cursor == "" {
			return collected
		}
		query = "?cursor_token=" + cursor
	}
}

func assertHistoryItem(t testing.TB, raw any, changeSetID uuid.UUID, operation string, sequenceNo int, actions []string) {
	t.Helper()
	item := raw.(map[string]any)
	if item["change_set_id"] != changeSetID.String() || item["operation"] != operation {
		t.Fatalf("unexpected history item identity: %#v", item)
	}
	if ref := stringField(t, raw, "history_item_ref"); ref == "" || !strings.HasPrefix(ref, "hitem_") {
		t.Fatalf("history item missing opaque display selector: %#v", item)
	}
	if item["actor_user_id"] == "" || item["committed_at"] == "" {
		t.Fatalf("history item missing actor or timestamp: %#v", item)
	}
	diff := item["diff_summary"].(map[string]any)
	if diff["summary"] == "" {
		t.Fatalf("history item missing diff summary: %#v", item)
	}
	unit := diff["units"].([]any)[0].(map[string]any)
	if unit["sequence_no"] != float64(sequenceNo) {
		t.Fatalf("unexpected same-change-set order marker: got %#v want %d", unit["sequence_no"], sequenceNo)
	}
	assertActions(t, item, actions)
}

func assertActions(t testing.TB, raw any, want []string) {
	t.Helper()
	item := raw.(map[string]any)
	gotRaw := item["available_rollback_actions"].([]any)
	if len(gotRaw) != len(want) {
		t.Fatalf("unexpected rollback actions: got %#v want %#v item=%#v", gotRaw, want, item)
	}
	for i, action := range want {
		if gotRaw[i] != action {
			t.Fatalf("unexpected rollback action order: got %#v want %#v", gotRaw, want)
		}
	}
	if item["reversible"] != (len(want) > 0) {
		t.Fatalf("reversible did not match actions: %#v", item)
	}
}

func requireNoRetainedHistoryNarrowingSurface(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, recordID uuid.UUID) {
	t.Helper()
	document := contracttest.OpenAPIDocument(t)
	for path := range historyOpenAPIObjectAt(t, document, "paths") {
		if historyNarrowingSurfaceName(path) {
			t.Fatalf("OpenAPI exposes retained-history narrowing route %q", path)
		}
	}

	extensionsBody := httptestx.RequireSuccessEnvelope(t, appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/extensions", nil, appsupport.WithCookies(login.SessionCookie)), http.StatusOK)
	extensions := extensionsBody["data"].(map[string]any)["extensions"].([]any)
	for _, raw := range extensions {
		extension := raw.(map[string]any)
		for _, routeRaw := range extension["route_families"].([]any) {
			route := routeRaw.(string)
			if historyNarrowingSurfaceName(route) {
				t.Fatalf("current extension profile exposes retained-history narrowing route %q in %#v", route, extension)
			}
		}
	}

	for _, path := range []string{
		"/api/v1/records/" + recordID.String() + "/history/purge",
		"/api/v1/records/" + recordID.String() + "/history/retention",
		"/api/v1/incidents/" + incidentID.String() + "/history/purge",
		"/api/v1/incidents/" + incidentID.String() + "/history/retention",
		"/api/v1/incidents/" + incidentID.String() + "/records/history/purge",
	} {
		resp := appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+path, nil, appsupport.WithCookies(login.SessionCookie))
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("unexpected retained-history narrowing route at %s: status=%d body=%#v", path, resp.StatusCode, httptestx.ReadJSONBody(t, resp))
		}
	}

	for _, path := range []string{"../../../configs/dev/config.toml", "../../../internal/platform/config/config.go"} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read operator config surface %s: %v", path, err)
		}
		if historyNarrowingSurfaceName(string(content)) {
			t.Fatalf("operator config surface %s exposes retained-history narrowing setting", path)
		}
	}
}

func historyNarrowingSurfaceName(value string) bool {
	lower := strings.ToLower(value)
	if !strings.Contains(lower, "history") {
		return false
	}
	for _, marker := range []string{"purge", "retention", "retained", "truncate", "horizon"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func stringField(t testing.TB, raw any, key string) string {
	t.Helper()
	item := raw.(map[string]any)
	value, _ := item[key].(string)
	return value
}

func historyOpenAPIObjectAt(t testing.TB, root map[string]any, path ...string) map[string]any {
	t.Helper()
	current := any(root)
	for _, key := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("path %v parent for %q is %T, want object", path, key, current)
		}
		value, ok := object[key]
		if !ok {
			t.Fatalf("path %v missing key %q", path, key)
		}
		current = value
	}
	object, ok := current.(map[string]any)
	if !ok {
		t.Fatalf("path %v is %T, want object", path, current)
	}
	return object
}

func historyOpenAPIStringArrayAt(t testing.TB, root map[string]any, key string) []string {
	t.Helper()
	raw, ok := root[key].([]any)
	if !ok {
		t.Fatalf("key %q is %T, want array", key, root[key])
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("key %q contains %T, want string", key, item)
		}
		values = append(values, value)
	}
	return values
}

func requireHistoryOpenAPIParameter(t testing.TB, parameters []any, name string, in string, required bool, format string) {
	t.Helper()
	for _, raw := range parameters {
		parameter, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("OpenAPI parameter is %T, want object", raw)
		}
		if parameter["name"] != name || parameter["in"] != in {
			continue
		}
		if parameter["required"] != required {
			t.Fatalf("%s parameter required = %v, want %v", name, parameter["required"], required)
		}
		if format != "" {
			schema := historyOpenAPIObjectAt(t, parameter, "schema")
			if schema["format"] != format {
				t.Fatalf("%s parameter schema format = %v, want %s", name, schema["format"], format)
			}
		}
		return
	}
	t.Fatalf("missing OpenAPI parameter %s in %s", name, in)
}
