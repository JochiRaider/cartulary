package workbook_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

const phase6NotesViewSchemaID = "cartulary.view.notes.v1"

func TestPhase6_GridWriteConcurrencyRoute_U_6_01(t *testing.T) {
	harness, login, _, incidentID := phase6ConflictFixture(t, "phase6-u-6-01-grid-write-concurrency", "IR-PHASE6-U-6-01")
	note := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-01-create", "Base title", "Base body")
	recordID := phase4test.MustUUID(t, note["record_id"].(string))

	missingBase := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id": phase6NotesViewSchemaID,
		"client_txn_id":  "txn-phase6-u-6-01-missing-base",
		"changes":        []map[string]any{{"field_key": "note.body", "value": "Client body"}},
	})
	httptestx.RequireErrorEnvelope(t, missingBase, http.StatusBadRequest, "invalid_mutation_payload")
	fullRowPatch := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"record_id":        recordID.String(),
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-01-record-id-in-body",
		"note.body":        "Client body",
	})
	httptestx.RequireErrorEnvelope(t, fullRowPatch, http.StatusBadRequest, "invalid_mutation_payload")

	titlePatch := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-01-server-title",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Server title"}},
	})
	requireCellValue(t, titlePatch["row"].(map[string]any), "note.title", "Server title")

	bodyPatch := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-01-client-body",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "Client body"}},
	})
	bodyRow := bodyPatch["row"].(map[string]any)
	requireCellValue(t, bodyRow, "note.title", "Server title")
	requireCellValue(t, bodyRow, "note.body", "Client body")
	if got := int64(bodyRow["row_version"].(float64)); got != 3 {
		t.Fatalf("different-field stale edit row_version = %d want 3", got)
	}
	phase6RequireMutationChangedFields(t, harness, bodyPatch["change_set_id"].(string), []string{"note.body"})

	beforeConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	sameField := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-01-client-title",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Client title"}},
	})
	body := httptestx.RequireErrorEnvelope(t, sameField, http.StatusConflict, "same_field_conflict")
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["field_key"] != "note.title" || conflict["client_value"] != "Client title" || conflict["server_value"] != "Server title" {
		t.Fatalf("unexpected same-field conflict payload: %#v", conflict)
	}
	afterConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if beforeConflict != afterConflict {
		t.Fatalf("same-field stale edit wrote durable side effects: before=%+v after=%+v", beforeConflict, afterConflict)
	}
}

func TestPhase6_SameFieldConflictHTTP_U_6_02(t *testing.T) {
	harness, login, actorID, incidentID := phase6ConflictFixture(t, "phase6-u-6-02-same-field-http", "IR-PHASE6-U-6-02")
	note := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-02-create", "Base title", "Base body")
	recordID := phase4test.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-02-server",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Server title"}},
	})

	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-02-client",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Client title"}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	errorObject := body["error"].(map[string]any)
	if errorObject["status"] != float64(http.StatusConflict) {
		t.Fatalf("same-field conflict error.status = %#v want 409", errorObject["status"])
	}
	conflict := errorObject["conflict"].(map[string]any)
	phase6RequireNoLegacyConflictAliases(t, conflict)
	if conflict["conflict_token"] == "" ||
		conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != "note.title" ||
		conflict["conflict_resolution_class"] != "text_compare_merge" ||
		conflict["base_value"] != "Base title" ||
		conflict["server_value"] != "Server title" ||
		conflict["client_value"] != "Client title" ||
		conflict["server_updated_by"] != actorID.String() {
		t.Fatalf("unexpected same-field conflict identity/value fields: %#v", conflict)
	}
	phase6RequireConflictVersion(t, conflict, "base_row_version", 1)
	phase6RequireConflictVersion(t, conflict, "current_row_version", 2)
	if _, err := time.Parse(time.RFC3339Nano, conflict["server_updated_at"].(string)); err != nil {
		t.Fatalf("server_updated_at was not RFC3339Nano: %v conflict=%#v", err, conflict)
	}
}

func TestPhase6_TextCompareMergeDurability_U_6_03(t *testing.T) {
	harness, login, _, incidentID := phase6ConflictFixture(t, "phase6-u-6-03-text-merge-durability", "IR-PHASE6-U-6-03")
	note := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-03-clean-create", "Merge note", "one\ntwo\nthree")
	recordID := phase4test.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-03-clean-server",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\nTWO\nthree"}},
	})

	beforeConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-03-clean-client",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\ntwo\nTHREE"}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["suggested_merged_value"] != "one\nTWO\nTHREE" {
		t.Fatalf("expected clean text suggestion without accepting patch, got %#v", conflict)
	}
	afterConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if beforeConflict != afterConflict {
		t.Fatalf("clean text conflict wrote durable side effects: before=%+v after=%+v", beforeConflict, afterConflict)
	}
	current := phase6RequireQueriedRow(t, harness, login, incidentID, phase6NotesViewSchemaID, recordID)
	requireCellValue(t, current, "note.body", "one\nTWO\nthree")
	if got := int64(current["row_version"].(float64)); got != 2 {
		t.Fatalf("rejected clean text conflict changed row_version = %d want 2", got)
	}

	resolved := phase6ResolveConflict(t, harness, login, recordID, conflict["conflict_token"].(string), map[string]any{
		"conflict_token":  conflict["conflict_token"].(string),
		"resolution_kind": "merged_value",
		"client_txn_id":   "txn-phase6-u-6-03-clean-resolve",
		"resolved_value":  "one\nTWO\nTHREE",
	})
	requireCellValue(t, resolved["row"].(map[string]any), "note.body", "one\nTWO\nTHREE")
	afterResolve := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if afterResolve.ChangeSets != beforeConflict.ChangeSets+1 || afterResolve.RecordRevisions != beforeConflict.RecordRevisions+1 {
		t.Fatalf("explicit text resolution should be the next durable revision: before=%+v after=%+v", beforeConflict, afterResolve)
	}

	overlap := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-03-overlap-create", "Overlap note", "one\ntwo")
	overlapID := phase4test.MustUUID(t, overlap["record_id"].(string))
	requireWorkbookPatch(t, harness, login, overlapID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-03-overlap-server",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\nserver"}},
	})
	overlapResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", overlapID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-03-overlap-client",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\nclient"}},
	})
	overlapBody := httptestx.RequireErrorEnvelope(t, overlapResp, http.StatusConflict, "same_field_conflict")
	overlapConflict := overlapBody["error"].(map[string]any)["conflict"].(map[string]any)
	if _, ok := overlapConflict["suggested_merged_value"]; ok {
		t.Fatalf("overlapping text conflict must omit suggested_merged_value: %#v", overlapConflict)
	}
}

func TestPhase6_CollectionReviewRouteResolve_U_6_04(t *testing.T) {
	harness, login, _, incidentID := phase6ConflictFixture(t, "phase6-u-6-04-collection-review-route", "IR-PHASE6-U-6-04")
	noteData := requireWorkbookCreate(t, harness, login, incidentID, phase6NotesViewSchemaID, map[string]any{
		"client_txn_id": "txn-phase6-u-6-04-create",
		"note.title":    "Collection note",
		"note.tags":     collectionActions(addToken("base-tag")),
	})
	note := noteData["row"].(map[string]any)
	recordID := phase4test.MustUUID(t, note["record_id"].(string))
	requireCollectionValueHasItemKind(t, cellMapValue(t, note, "note.tags"), "tag")
	queried := phase6RequireQueriedRow(t, harness, login, incidentID, phase6NotesViewSchemaID, recordID)
	requireCollectionValueHasItemKind(t, cellMapValue(t, queried, "note.tags"), "tag")

	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-04-server",
		"changes": []map[string]any{{
			"field_key":      "note.tags",
			"action_payload": collectionActions(addToken("server-tag")),
		}},
	})
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-04-client",
		"changes": []map[string]any{{
			"field_key":      "note.tags",
			"action_payload": collectionActions(addToken("client-tag")),
		}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["field_key"] != "note.tags" || conflict["conflict_resolution_class"] != "collection_review" {
		t.Fatalf("unexpected collection conflict identity fields: %#v", conflict)
	}
	for _, key := range []string{"base_value", "server_value", "client_value"} {
		value := conflict[key].(map[string]any)
		if value["kind"] != "collection_value_v1" {
			t.Fatalf("%s must use collection_value_v1, got %#v", key, value)
		}
		requireCollectionValueHasItemKind(t, value, "tag")
	}

	resolved := phase6ResolveConflict(t, harness, login, recordID, conflict["conflict_token"].(string), map[string]any{
		"conflict_token":  conflict["conflict_token"].(string),
		"resolution_kind": "merged_value",
		"client_txn_id":   "txn-phase6-u-6-04-resolve",
		"resolved_value":  collectionActions(addToken("client-tag")),
	})
	resolvedTags := cellMapValue(t, resolved["row"].(map[string]any), "note.tags")
	requireCollectionValueHasDisplayText(t, resolvedTags, "base-tag")
	requireCollectionValueHasDisplayText(t, resolvedTags, "server-tag")
	requireCollectionValueHasDisplayText(t, resolvedTags, "client-tag")

	for _, invalid := range []struct {
		name    string
		payload any
	}{
		{name: "raw-array", payload: []any{}},
		{name: "raw-string", payload: "client-tag"},
		{name: "replace-all", payload: map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "replace_all", "raw_text": "client-tag"}}}},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			invalidResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
				"view_schema_id":   phase6NotesViewSchemaID,
				"base_row_version": 3,
				"client_txn_id":    "txn-phase6-u-6-04-invalid-" + invalid.name,
				"changes": []map[string]any{{
					"field_key":      "note.tags",
					"action_payload": invalid.payload,
				}},
			})
			httptestx.RequireErrorEnvelope(t, invalidResp, http.StatusBadRequest, "invalid_mutation_payload")
		})
	}
}

func phase6ConflictFixture(t testing.TB, name string, incidentKey string) (*phase4test.ServerHarness, phase4test.LoginResult, uuid.UUID, uuid.UUID) {
	t.Helper()
	harness := phase4test.StartServer(t, name)
	login, actorID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-" + incidentKey,
		"incident_key":  incidentKey,
		"title":         name,
	})
	return harness, login, actorID, phase4test.MustUUID(t, incident["incident_id"].(string))
}

func phase6CreateNote(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, clientTxnID string, title string, body string) map[string]any {
	t.Helper()
	data := requireWorkbookCreate(t, harness, login, incidentID, phase6NotesViewSchemaID, map[string]any{
		"client_txn_id": clientTxnID,
		"note.title":    title,
		"note.body":     body,
	})
	return data["row"].(map[string]any)
}

func phase6RequireMutationChangedFields(t testing.TB, harness *phase4test.ServerHarness, changeSetID string, want []string) {
	t.Helper()
	var beforeJSON, afterJSON []byte
	row := harness.DB.QueryRowContext(context.Background(), `
SELECT before_value, after_value
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND sequence_no = 1
`, changeSetID)
	if err := row.Scan(&beforeJSON, &afterJSON); err != nil {
		t.Fatalf("load change-set mutation %s: %v", changeSetID, err)
	}
	beforeRow := phase6DecodeRowJSON(t, beforeJSON)
	afterRow := phase6DecodeRowJSON(t, afterJSON)
	got := phase6ChangedCellKeys(beforeRow, afterRow)
	if !phase6StringSlicesEqual(got, want) {
		t.Fatalf("change-set mutation changed fields = %#v want %#v", got, want)
	}
}

func phase6DecodeRowJSON(t testing.TB, payload []byte) map[string]any {
	t.Helper()
	var row map[string]any
	if err := json.Unmarshal(payload, &row); err != nil {
		t.Fatalf("decode row json: %v payload=%s", err, payload)
	}
	return row
}

func phase6ChangedCellKeys(beforeRow map[string]any, afterRow map[string]any) []string {
	beforeCells, _ := beforeRow["cells"].(map[string]any)
	afterCells, _ := afterRow["cells"].(map[string]any)
	keys := make([]string, 0)
	for fieldKey, afterCell := range afterCells {
		if phase6ServerManagedRevisionCell(fieldKey) {
			continue
		}
		beforeCell, ok := beforeCells[fieldKey]
		if !ok || !phase6JSONEqual(beforeCell, afterCell) {
			keys = append(keys, fieldKey)
		}
	}
	sort.Strings(keys)
	return keys
}

func phase6ServerManagedRevisionCell(fieldKey string) bool {
	switch fieldKey {
	case "note.updated_at":
		return true
	default:
		return false
	}
}

func phase6JSONEqual(left any, right any) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return string(leftJSON) == string(rightJSON)
}

func phase6StringSlicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func phase6RequireNoLegacyConflictAliases(t testing.TB, conflict map[string]any) {
	t.Helper()
	for _, key := range []string{"current_field_value", "conflict_resolution"} {
		if _, ok := conflict[key]; ok {
			t.Fatalf("same-field conflict preserved legacy alias %q: %#v", key, conflict)
		}
	}
}

func phase6RequireConflictVersion(t testing.TB, conflict map[string]any, key string, want int64) {
	t.Helper()
	got, ok := conflict[key].(float64)
	if !ok || int64(got) != want {
		t.Fatalf("unexpected %s: got %#v want %d conflict=%#v", key, conflict[key], want, conflict)
	}
}

func phase6RequireQueriedRow(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, viewSchemaID string, recordID uuid.UUID) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/query",
		map[string]any{},
		phase4test.WithCookies(login.SessionCookie),
	)
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	rows := body["data"].(map[string]any)["rows"].([]any)
	for _, row := range rows {
		rowMap := row.(map[string]any)
		if rowMap["record_id"] == recordID.String() {
			return rowMap
		}
	}
	t.Fatalf("query did not return record %s in %s rows=%#v", recordID, viewSchemaID, rows)
	return nil
}

func requireCollectionValueHasDisplayText(t testing.TB, value map[string]any, displayText string) {
	t.Helper()
	items := value["items"].([]any)
	for _, item := range items {
		itemMap := item.(map[string]any)
		if itemMap["display_text"] == displayText {
			return
		}
	}
	t.Fatalf("expected collection to contain display_text %q, got %#v", displayText, value)
}
