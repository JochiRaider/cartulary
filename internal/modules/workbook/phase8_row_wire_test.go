package workbook_test

import (
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestPhase8_RowWireFamilies_U_8_10(t *testing.T) {
	harness := phase4test.StartServer(t, "phase8-u-8-10-row-wire")
	adminLogin, actorID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-u-8-10-incident",
		"incident_key":  "IR-PHASE8-U-8-10",
		"title":         "Phase 8 row wire",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	recordID := seedNoteArtifact(t, harness, incidentID, actorID, "Visible title", nil)

	notesQueryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/cartulary.view.notes.v1/query"
	envelope := queryWorkbook(t, harness, adminLogin, notesQueryURL, map[string]any{})
	rows := responseRows(envelope)
	if len(rows) != 1 || rows[0]["record_id"] != recordID.String() {
		t.Fatalf("expected queried note row, got %#v", rows)
	}
	cells := rows[0]["cells"].(map[string]any)
	for _, fieldKey := range []string{
		"note.title",
		"note.body",
		"note.tags",
		"note.linked_record_count",
		"note.updated_at",
		"note.created_by_user_id",
	} {
		if _, ok := cells[fieldKey]; !ok {
			t.Fatalf("full row omitted schema field %s: %#v", fieldKey, cells)
		}
	}
	if got := cells["note.body"].(map[string]any)["value"]; got != nil {
		t.Fatalf("expected null body to remain authoritative null, got %#v", got)
	}
	if got := cells["note.created_by_user_id"].(map[string]any)["value"]; got != actorID.String() {
		t.Fatalf("hidden writable/technical-adjacent field must be present, got %#v", got)
	}

	rowPatch := platformws.BuildViewRowPatch(rows[0], []string{"note.title"})
	patchCells := rowPatch["cells"].(map[string]any)
	if !reflect.DeepEqual(patchCells, map[string]any{"note.title": cells["note.title"]}) {
		t.Fatalf("sparse patch must include only changed cells, got %#v", patchCells)
	}
	directPayload := platformws.RecordChangePayload(platformws.RecordChange{
		IncidentID:       incidentID,
		RecordID:         recordID,
		RowVersion:       2,
		ChangeSetID:      uuid.New(),
		ClientTxnID:      "txn-phase8-u-8-10-patch",
		ActorUserID:      actorID,
		ChangedFieldKeys: []string{"note.title", "note.body", "note.title"},
		ViewSchemaID:     "cartulary.view.notes.v1",
		PatchCells:       platformws.BuildViewRowPatch(rows[0], []string{"note.title", "note.body"}),
	})
	if got := directPayload["changed_field_keys"]; !reflect.DeepEqual(got, []string{"note.body", "note.title"}) {
		t.Fatalf("changed_field_keys must be canonical, got %#v", got)
	}
	affectedViews := directPayload["affected_views"].([]map[string]any)
	if len(affectedViews) != 1 || affectedViews[0]["view_schema_id"] != "cartulary.view.notes.v1" || affectedViews[0]["change_kind"] != "patch" {
		t.Fatalf("unexpected affected_views: %#v", affectedViews)
	}

	timelineCreated := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-phase8-u-8-10-timeline-create",
		"timeline.activity_synopsis_text": "Phase 8 operational row",
		"timeline.raw_activity_text":      "Hidden source",
		"timeline.activity_utc_text":      "2026-05-16T12:00:00Z",
	})
	timelineRow := timelineCreated["row"].(map[string]any)
	timelineRecordID := phase4test.MustUUID(t, timelineRow["record_id"].(string))
	timelineQueryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/cartulary.view.timeline.v2/query"
	timelineRows := responseRows(queryWorkbook(t, harness, adminLogin, timelineQueryURL, map[string]any{}))
	var queriedTimelineRow map[string]any
	for _, row := range timelineRows {
		if row["record_id"] == timelineRecordID.String() {
			queriedTimelineRow = row
			break
		}
	}
	if queriedTimelineRow == nil {
		t.Fatalf("expected queried timeline row, got %#v", timelineRows)
	}
	timelineCells := queriedTimelineRow["cells"].(map[string]any)
	for _, fieldKey := range []string{"timeline.activity_synopsis_text", "timeline.raw_activity_text"} {
		cell, ok := timelineCells[fieldKey].(map[string]any)
		if !ok {
			t.Fatalf("full row omitted visible operational field %s: %#v", fieldKey, timelineCells)
		}
		if got := cell["value"]; got == nil || got == "" {
			t.Fatalf("full row did not preserve operational %s value: %#v", fieldKey, cell)
		}
	}

	patchEvidenceCreated := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":               "txn-phase8-u-8-10-evidence-create",
		"evidence.title":              "Phase 8 sparse evidence",
		"evidence.lifecycle_state":    "received",
		"evidence.received_at":        "2026-05-16T12:05:00Z",
		"evidence.collector_party_id": nil,
	})
	patchEvidenceRow := patchEvidenceCreated["row"].(map[string]any)
	patchEvidenceRecordID := phase4test.MustUUID(t, patchEvidenceRow["record_id"].(string))
	hubChanges, unsubscribe := harness.Server.Runtime.WSHub.SubscribeRecordChanges(4)
	defer unsubscribe()
	patched := requireWorkbookPatch(t, harness, adminLogin, patchEvidenceRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": patchEvidenceRow["row_version"],
		"client_txn_id":    "txn-phase8-u-8-10-evidence-null-patch",
		"changes": []map[string]any{
			{"field_key": "evidence.received_at", "value": nil},
		},
	})
	patchedRow := patched["row"].(map[string]any)
	change := requireHubRecordChange(t, hubChanges, patchEvidenceRecordID, int64(patchedRow["row_version"].(float64)))
	payload := platformws.RecordChangePayload(change)
	changedKeys := payload["changed_field_keys"].([]string)
	if !slices.IsSorted(changedKeys) || len(changedKeys) != len(slices.Compact(append([]string(nil), changedKeys...))) || !slices.Contains(changedKeys, "evidence.received_at") {
		t.Fatalf("route-backed changed_field_keys must be canonical public keys, got %#v", changedKeys)
	}
	routeAffectedViews := payload["affected_views"].([]map[string]any)
	if len(routeAffectedViews) != 1 {
		t.Fatalf("expected one affected view, got %#v", routeAffectedViews)
	}
	affectedView := routeAffectedViews[0]
	if affectedView["view_schema_id"] != "cartulary.view.evidence.v1" || affectedView["change_kind"] != "patch" {
		t.Fatalf("affected_views must use public deterministic metadata, got %#v", affectedView)
	}
	patch := affectedView["patch_cells"].(map[string]any)
	if patch["record_id"] != patchEvidenceRecordID.String() || int64Value(patch["row_version"]) != int64Value(patchedRow["row_version"]) {
		t.Fatalf("patch_cells must carry authoritative row identity/version, got %#v", patch)
	}
	routePatchCells := patch["cells"].(map[string]any)
	if _, ok := routePatchCells["evidence.title"]; ok {
		t.Fatalf("sparse patch included unchanged title cell: %#v", routePatchCells)
	}
	receivedAtCell := routePatchCells["evidence.received_at"].(map[string]any)
	if got := receivedAtCell["value"]; got != nil {
		t.Fatalf("sparse patch must preserve authoritative null received_at, got %#v", got)
	}
}

func TestPhase8_RecordChangedSparsePatchPayloads_U_8_10_AC368(t *testing.T) {
	harness := phase4test.StartServer(t, "phase8-ac368-sparse-patches")
	adminLogin, actorID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-ac368-incident",
		"incident_key":  "IR-PHASE8-AC368",
		"title":         "Phase 8 sparse patch coverage",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	partyData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-phase8-ac368-party",
		"party.display_name": "Sparse Patch Party",
		"party.party_kind":   "team",
	})
	partyID := phase4test.MustUUID(t, partyData["row"].(map[string]any)["record_id"].(string))

	hubChanges, unsubscribe := harness.Server.Runtime.WSHub.SubscribeRecordChanges(16)
	defer unsubscribe()

	evidenceData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":                 "txn-phase8-ac368-evidence",
		"evidence.title":                "Sparse reference evidence",
		"evidence.lifecycle_state":      "received",
		"evidence.collector_party_text": "Collector text stays",
		"evidence.collector_party_id":   partyID.String(),
		"evidence.source_party_text":    "Source text stays",
		"evidence.source_party_id":      partyID.String(),
	})
	evidenceRow := evidenceData["row"].(map[string]any)
	evidenceID := phase4test.MustUUID(t, evidenceRow["record_id"].(string))
	clearedCollector := requireWorkbookPatch(t, harness, adminLogin, evidenceID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": evidenceRow["row_version"],
		"client_txn_id":    "txn-phase8-ac368-clear-collector",
		"changes": []map[string]any{
			{"field_key": "evidence.collector_party_id", "value": nil},
		},
	})["row"].(map[string]any)
	collectorPatch := requireSparsePatchForChange(t, hubChanges, evidenceID, clearedCollector, "cartulary.view.evidence.v1")
	requireSparsePatchNullCell(t, collectorPatch, "evidence.collector_party_id")
	requireSparsePatchOmitsCells(t, collectorPatch, "evidence.title", "evidence.source_party_id")

	taskData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.task_requests.v1", map[string]any{
		"client_txn_id":             "txn-phase8-ac368-task",
		"task.title":                "Sparse requester task",
		"task.task_kind":            "request",
		"task.requester_party_text": "Requester text stays",
		"task.requester_party_id":   partyID.String(),
	})
	taskRow := taskData["row"].(map[string]any)
	taskID := phase4test.MustUUID(t, taskRow["record_id"].(string))
	clearedRequester := requireWorkbookPatch(t, harness, adminLogin, taskID, map[string]any{
		"view_schema_id":   "cartulary.view.task_requests.v1",
		"base_row_version": taskRow["row_version"],
		"client_txn_id":    "txn-phase8-ac368-clear-requester",
		"changes": []map[string]any{
			{"field_key": "task.requester_party_id", "value": nil},
		},
	})["row"].(map[string]any)
	requesterPatch := requireSparsePatchForChange(t, hubChanges, taskID, clearedRequester, "cartulary.view.task_requests.v1")
	requireSparsePatchNullCell(t, requesterPatch, "task.requester_party_id")
	requireSparsePatchOmitsCells(t, requesterPatch, "task.title", "task.requester_party_text")

	noteData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.notes.v1", map[string]any{
		"client_txn_id": "txn-phase8-ac368-note",
		"note.title":    "Sparse collection note",
		"note.body":     "Collection patch body remains unchanged",
	})
	noteRow := noteData["row"].(map[string]any)
	noteID := phase4test.MustUUID(t, noteRow["record_id"].(string))
	patchedNote := requireWorkbookPatch(t, harness, adminLogin, noteID, map[string]any{
		"view_schema_id":   "cartulary.view.notes.v1",
		"base_row_version": noteRow["row_version"],
		"client_txn_id":    "txn-phase8-ac368-note-tags",
		"changes": []map[string]any{
			{"field_key": "note.tags", "action_payload": collectionActions(addToken("triage"))},
		},
	})["row"].(map[string]any)
	notePatch := requireSparsePatchForChange(t, hubChanges, noteID, patchedNote, "cartulary.view.notes.v1")
	noteTags := requireSparsePatchCellValue(t, notePatch, "note.tags").(map[string]any)
	if noteTags["kind"] != "collection_value_v1" || noteTags["ordered"] != false {
		t.Fatalf("note.tags sparse patch must carry collection_value_v1, got %#v", noteTags)
	}
	if got := collectionItemCount(noteTags["items"]); got != 1 {
		t.Fatalf("note.tags sparse patch item count got %d want 1: %#v", got, noteTags)
	}
	requireSparsePatchOmitsCells(t, notePatch, "note.title", "note.body")

	commData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.comm_log.v1", map[string]any{
		"client_txn_id":               "txn-phase8-ac368-comm",
		"comm_log.comm_type":          "briefing",
		"comm_log.audience":           "IR leadership",
		"comm_log.channel_or_meeting": "Bridge",
		"comm_log.summary":            "Grouped sparse patch source",
	})
	commID := phase4test.MustUUID(t, commData["row"].(map[string]any)["record_id"].(string))
	commQueryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/cartulary.view.comm_log.v1/query"
	groupedRows := responseRows(queryWorkbook(t, harness, adminLogin, commQueryURL, map[string]any{
		"group_by": "comm_log.comm_type",
	}))
	groupedComm := findRowByID(t, groupedRows, commID)
	groupPatch := platformws.BuildViewRowPatch(groupedComm, []string{"comm_log.comm_type", "comm_log.summary"})
	groupValues := groupPatch["group_values"].(map[string]any)
	if !reflect.DeepEqual(groupValues, map[string]any{"comm_log.comm_type": "briefing"}) {
		t.Fatalf("grouped sparse patch must include only changed grouping scalars, got %#v", groupValues)
	}
	groupPatchCells := groupPatch["cells"].(map[string]any)
	if _, ok := groupPatchCells["comm_log.audience"]; ok {
		t.Fatalf("grouped sparse patch included unchanged sibling cell: %#v", groupPatchCells)
	}

	invalidatePayload := platformws.RecordChangePayload(platformws.RecordChange{
		IncidentID:       incidentID,
		RecordID:         uuid.New(),
		RowVersion:       3,
		ChangeSetID:      uuid.New(),
		ClientTxnID:      "txn-phase8-ac368-invalidate",
		ActorUserID:      actorID,
		ChangedFieldKeys: []string{"note.title"},
		ViewSchemaID:     "cartulary.view.notes.v1",
	})
	affectedViews := invalidatePayload["affected_views"].([]map[string]any)
	if len(affectedViews) != 1 || affectedViews[0]["change_kind"] != "invalidate" {
		t.Fatalf("missing patch data must fall back to invalidate, got %#v", affectedViews)
	}
	if _, ok := affectedViews[0]["patch_cells"]; ok {
		t.Fatalf("invalidate affected view must not guess patch_cells: %#v", affectedViews[0])
	}
}

func int64Value(value any) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case float64:
		return int64(typed)
	default:
		return 0
	}
}

func requireSparsePatchForChange(t testing.TB, changes <-chan platformws.RecordChange, recordID uuid.UUID, row map[string]any, viewSchemaID string) map[string]any {
	t.Helper()
	change := requireHubRecordChange(t, changes, recordID, int64Value(row["row_version"]))
	payload := platformws.RecordChangePayload(change)
	changedKeys := payload["changed_field_keys"].([]string)
	if !slices.IsSorted(changedKeys) || len(changedKeys) != len(slices.Compact(append([]string(nil), changedKeys...))) {
		t.Fatalf("changed_field_keys must be canonical, got %#v", changedKeys)
	}
	affectedViews := payload["affected_views"].([]map[string]any)
	if len(affectedViews) != 1 || affectedViews[0]["view_schema_id"] != viewSchemaID || affectedViews[0]["change_kind"] != "patch" {
		t.Fatalf("unexpected affected_views for %s: %#v", viewSchemaID, affectedViews)
	}
	patch := affectedViews[0]["patch_cells"].(map[string]any)
	if patch["record_id"] != recordID.String() || int64Value(patch["row_version"]) != int64Value(row["row_version"]) {
		t.Fatalf("patch identity/version mismatch got %#v row %#v", patch, row)
	}
	patchCells := patch["cells"].(map[string]any)
	for fieldKey := range patchCells {
		if !slices.Contains(changedKeys, fieldKey) {
			t.Fatalf("patch cell %s missing from changed_field_keys %#v", fieldKey, changedKeys)
		}
	}
	return patch
}

func requireSparsePatchCellValue(t testing.TB, patch map[string]any, fieldKey string) any {
	t.Helper()
	cells := patch["cells"].(map[string]any)
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		t.Fatalf("sparse patch missing cell %s: %#v", fieldKey, cells)
	}
	return cell["value"]
}

func requireSparsePatchNullCell(t testing.TB, patch map[string]any, fieldKey string) {
	t.Helper()
	if value := requireSparsePatchCellValue(t, patch, fieldKey); value != nil {
		t.Fatalf("sparse patch cell %s got %#v want authoritative null", fieldKey, value)
	}
}

func requireSparsePatchOmitsCells(t testing.TB, patch map[string]any, fieldKeys ...string) {
	t.Helper()
	cells := patch["cells"].(map[string]any)
	for _, fieldKey := range fieldKeys {
		if _, ok := cells[fieldKey]; ok {
			t.Fatalf("sparse patch included unchanged cell %s: %#v", fieldKey, cells)
		}
	}
}

func collectionItemCount(items any) int {
	switch typed := items.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	default:
		return 0
	}
}

func findRowByID(t testing.TB, rows []map[string]any, recordID uuid.UUID) map[string]any {
	t.Helper()
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return row
		}
	}
	t.Fatalf("missing row %s in %#v", recordID, rows)
	return nil
}

func requireHubRecordChange(t testing.TB, changes <-chan platformws.RecordChange, recordID uuid.UUID, rowVersion int64) platformws.RecordChange {
	t.Helper()
	deadline := time.After(5 * time.Second)
	var last platformws.RecordChange
	for {
		select {
		case change := <-changes:
			last = change
			if change.RecordID == recordID && change.RowVersion == rowVersion {
				return change
			}
		case <-deadline:
			t.Fatalf("timed out waiting for record change record=%s version=%d after %#v", recordID, rowVersion, last)
		}
	}
}

func TestPhase8_LiveAuthorizedCursorPagination_I_8_04(t *testing.T) {
	harness := phase4test.StartServer(t, "phase8-i-8-04-cursor")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-i-8-04-incident",
		"incident_key":  "IR-PHASE8-I-8-04",
		"title":         "Phase 8 cursor",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	hostA := uuid.New()
	hostC := uuid.New()
	seedHostForPaging(t, harness, incidentID, adminUserID, hostA, "Alpha")
	seedHostForPaging(t, harness, incidentID, adminUserID, hostC, "Charlie")

	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + entities.HostsViewSchemaID + "/query"
	sortByName := []map[string]any{{"field_key": "host.display_name", "direction": "asc"}}
	pageOne := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{"limit": 1, "sort": sortByName})
	cursor := responsePaging(pageOne)["next_cursor"].(string)
	if rows := responseRows(pageOne); len(rows) != 1 || rows[0]["record_id"] != hostA.String() {
		t.Fatalf("expected first page host A, got %#v", rows)
	}

	hostB := uuid.New()
	seedHostForPaging(t, harness, incidentID, adminUserID, hostB, "Bravo")
	pageTwo := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{"cursor_token": cursor, "sort": sortByName})
	if rows := responseRows(pageTwo); len(rows) != 1 || rows[0]["record_id"] != hostB.String() {
		t.Fatalf("cursor continuation must evaluate fetch-time rows, got %#v", rows)
	}

	resp := phase4test.DoJSON(t, http.MethodPost, queryURL, map[string]any{
		"cursor_token": tamperCursor(t, cursor),
		"sort":         sortByName,
	}, phase4test.WithCookies(adminLogin.SessionCookie))
	errBody := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_view_query")
	if details := errBody["error"].(map[string]any)["details"].(map[string]any); details["reason_code"] != "invalid_cursor_token" {
		t.Fatalf("expected invalid_cursor_token, got %#v", details)
	}

	changedQuery := phase4test.DoJSON(t, http.MethodPost, queryURL, map[string]any{
		"cursor_token": cursor,
		"sort":         []map[string]any{{"field_key": "host.hostname", "direction": "asc"}},
	}, phase4test.WithCookies(adminLogin.SessionCookie))
	errBody = httptestx.RequireErrorEnvelope(t, changedQuery, http.StatusBadRequest, "invalid_view_query")
	if details := errBody["error"].(map[string]any)["details"].(map[string]any); details["reason_code"] != "cursor_query_mismatch" {
		t.Fatalf("expected cursor_query_mismatch, got %#v", details)
	}

	otherView := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.notes.v1/query", map[string]any{
		"cursor_token": cursor,
	}, phase4test.WithCookies(adminLogin.SessionCookie))
	errBody = httptestx.RequireErrorEnvelope(t, otherView, http.StatusBadRequest, "invalid_view_query")
	if details := errBody["error"].(map[string]any)["details"].(map[string]any); details["reason_code"] != "cursor_query_mismatch" {
		t.Fatalf("expected view_schema cursor_query_mismatch, got %#v", details)
	}

	otherIncident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-i-8-04-other-incident",
		"incident_key":  "IR-PHASE8-I-8-04-OTHER",
		"title":         "Phase 8 cursor other",
	})
	otherIncidentID := phase4test.MustUUID(t, otherIncident["incident_id"].(string))
	otherRoute := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+otherIncidentID.String()+"/views/"+entities.HostsViewSchemaID+"/query", map[string]any{
		"cursor_token": cursor,
		"sort":         sortByName,
	}, phase4test.WithCookies(adminLogin.SessionCookie))
	errBody = httptestx.RequireErrorEnvelope(t, otherRoute, http.StatusBadRequest, "invalid_view_query")
	if details := errBody["error"].(map[string]any)["details"].(map[string]any); details["reason_code"] != "cursor_query_mismatch" {
		t.Fatalf("expected incident cursor_query_mismatch, got %#v", details)
	}
}

func TestPhase8_CursorContinuationRechecksAuthorization_I_8_04(t *testing.T) {
	harness := phase4test.StartServer(t, "phase8-i-8-04-cursor-auth")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-i-8-04-auth-incident",
		"incident_key":  "IR-PHASE8-I-8-04-AUTH",
		"title":         "Phase 8 cursor auth",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	seedHostForPaging(t, harness, incidentID, adminUserID, uuid.New(), "Alpha")
	seedHostForPaging(t, harness, incidentID, adminUserID, uuid.New(), "Bravo")

	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + entities.HostsViewSchemaID + "/query"
	sortByName := []map[string]any{{"field_key": "host.display_name", "direction": "asc"}}
	pageOne := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{"limit": 1, "sort": sortByName})
	cursor := responsePaging(pageOne)["next_cursor"].(string)

	execSeed(t, harness, `
UPDATE user_sessions
   SET revoked_at = now(),
       revoke_reason_code = 'phase8_test_revoked',
       updated_at = now()
 WHERE user_id = $1
   AND revoked_at IS NULL
`, adminUserID)
	sessionResp := phase4test.DoJSON(t, http.MethodPost, queryURL, map[string]any{
		"cursor_token": cursor,
		"sort":         sortByName,
	}, phase4test.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireErrorEnvelope(t, sessionResp, http.StatusUnauthorized, "session_required")
}

func TestPhase8_CursorContinuationRechecksMembership_I_8_04(t *testing.T) {
	harness := phase4test.StartServer(t, "phase8-i-8-04-cursor-membership")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-i-8-04-membership-incident",
		"incident_key":  "IR-PHASE8-I-8-04-MEMBER",
		"title":         "Phase 8 cursor membership",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	seedHostForPaging(t, harness, incidentID, adminUserID, uuid.New(), "Alpha")
	seedHostForPaging(t, harness, incidentID, adminUserID, uuid.New(), "Bravo")

	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + entities.HostsViewSchemaID + "/query"
	sortByName := []map[string]any{{"field_key": "host.display_name", "direction": "asc"}}
	pageOne := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{"limit": 1, "sort": sortByName})
	cursor := responsePaging(pageOne)["next_cursor"].(string)

	execSeed(t, harness, `DELETE FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`, incidentID, adminUserID)
	membershipResp := phase4test.DoJSON(t, http.MethodPost, queryURL, map[string]any{
		"cursor_token": cursor,
		"sort":         sortByName,
	}, phase4test.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireErrorEnvelope(t, membershipResp, http.StatusNotFound, "incident_not_found")
}

func TestSupportPhase8Integration_NotesFullTextExactSearch_AC185(t *testing.T) {
	harness := phase4test.StartServer(t, "phase8-ac185-notes")
	adminLogin, actorID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-ac185-incident",
		"incident_key":  "IR-PHASE8-AC185",
		"title":         "Phase 8 AC185",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	alpha := seedNoteArtifact(t, harness, incidentID, actorID, "Alpha Delta", ptrString("Responder contained shell"))
	bravo := seedNoteArtifact(t, harness, incidentID, actorID, "Bravo Alpha", ptrString("Shell token appears earlier alphabetically"))
	seedNoteArtifact(t, harness, incidentID, actorID, "Powershell only", nil)
	phrase := seedNoteArtifact(t, harness, incidentID, actorID, "Phrase note", ptrString("alpha middle shell"))
	seedNoteArtifact(t, harness, incidentID, actorID, "Wildcard note", ptrString("shells alpha"))
	seedNoteArtifact(t, harness, incidentID, actorID, "Cafe note", ptrString("cafe token"))
	seedNoteArtifact(t, harness, incidentID, actorID, "Accent note", ptrString("café token"))

	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/cartulary.view.notes.v1/query"
	exact := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"sort":    []map[string]any{{"field_key": "note.title", "direction": "asc"}},
		"filters": []map[string]any{{"field_key": "note.full_text", "op": "full_text", "arg": map[string]any{"query": "shell alpha shell"}}},
	})
	if ids := rowIDs(responseRows(exact)); !reflect.DeepEqual(ids, []string{alpha.String(), bravo.String(), phrase.String()}) {
		t.Fatalf("exact full_text must match every unique token, got %v", ids)
	}

	substring := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"sort":    []map[string]any{{"field_key": "note.title", "direction": "asc"}},
		"filters": []map[string]any{{"field_key": "note.full_text", "op": "full_text", "arg": map[string]any{"query": "shell"}}},
	})
	if ids := rowIDs(responseRows(substring)); !reflect.DeepEqual(ids, []string{alpha.String(), bravo.String(), phrase.String()}) {
		t.Fatalf("full_text must not match powershell by substring, got %v", ids)
	}

	reordered := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"sort":    []map[string]any{{"field_key": "note.title", "direction": "asc"}},
		"filters": []map[string]any{{"field_key": "note.full_text", "op": "full_text", "arg": map[string]any{"query": "shell alpha shell"}}},
	})
	if ids := rowIDs(responseRows(reordered)); !reflect.DeepEqual(ids, []string{alpha.String(), bravo.String(), phrase.String()}) {
		t.Fatalf("full_text must preserve applied sort order rather than relevance, got %v", ids)
	}

	phraseQuery := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"sort":    []map[string]any{{"field_key": "note.title", "direction": "asc"}},
		"filters": []map[string]any{{"field_key": "note.full_text", "op": "full_text", "arg": map[string]any{"query": "alpha shell"}}},
	})
	if ids := rowIDs(responseRows(phraseQuery)); !reflect.DeepEqual(ids, []string{alpha.String(), bravo.String(), phrase.String()}) {
		t.Fatalf("full_text must use all-token membership without phrase ordering, got %v", ids)
	}

	accent := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"filters": []map[string]any{{"field_key": "note.full_text", "op": "full_text", "arg": map[string]any{"query": "cafe"}}},
	})
	if len(responseRows(accent)) != 1 {
		t.Fatalf("diacritics must remain significant; expected only cafe row, got %#v", responseRows(accent))
	}

	invalid := phase4test.DoJSON(t, http.MethodPost, queryURL, map[string]any{
		"filters": []map[string]any{{"field_key": "note.full_text", "op": "full_text", "arg": map[string]any{"query": " -- "}}},
	}, phase4test.WithCookies(adminLogin.SessionCookie))
	errBody := httptestx.RequireErrorEnvelope(t, invalid, http.StatusBadRequest, "invalid_view_query")
	if details := errBody["error"].(map[string]any)["details"].(map[string]any); details["reason_code"] != "empty_full_text_after_tokenization" {
		t.Fatalf("expected zero-token rejection, got %#v", details)
	}
}

func TestSupportPhase8Integration_PrefixAndNullLastOrdering(t *testing.T) {
	harness := phase4test.StartServer(t, "phase8-e-8-04-prefix-null-last")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-e-8-04-prefix-incident",
		"incident_key":  "IR-PHASE8-E-8-04-PREFIX",
		"title":         "Phase 8 prefix",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	alpha := uuid.New()
	infix := uuid.New()
	wildcard := uuid.New()
	nullHost := uuid.New()
	seedHostForPaging(t, harness, incidentID, adminUserID, alpha, "Host A")
	seedHostForPaging(t, harness, incidentID, adminUserID, infix, "Host B")
	seedHostForPaging(t, harness, incidentID, adminUserID, wildcard, "Host C")
	seedHostForPaging(t, harness, incidentID, adminUserID, nullHost, "Zulu")
	execSeed(t, harness, `
UPDATE host_grid_projection
   SET location = CASE record_id
       WHEN $1 THEN 'Alpha'
       WHEN $2 THEN 'XAlpha'
       WHEN $3 THEN 'Al%ha'
       WHEN $4 THEN NULL
       END
 WHERE record_id IN ($1, $2, $3, $4)
`, alpha, infix, wildcard, nullHost)

	queryURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + entities.HostsViewSchemaID + "/query"
	prefix := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"sort":    []map[string]any{{"field_key": "host.display_name", "direction": "asc"}},
		"filters": []map[string]any{{"field_key": "host.location", "op": "prefix", "arg": map[string]any{"value": "al"}}},
	})
	if ids := rowIDs(responseRows(prefix)); !reflect.DeepEqual(ids, []string{alpha.String(), wildcard.String()}) {
		t.Fatalf("prefix must be anchored and treat wildcard chars literally, got %v", ids)
	}

	desc := queryWorkbook(t, harness, adminLogin, queryURL, map[string]any{
		"sort": []map[string]any{{"field_key": "host.location", "direction": "desc"}},
	})
	ids := rowIDs(responseRows(desc))
	if len(ids) != 4 || ids[len(ids)-1] != nullHost.String() {
		t.Fatalf("null sort values must remain last for descending sort, got %v", ids)
	}
}

func tamperCursor(t testing.TB, cursor string) string {
	t.Helper()
	replacement := "A"
	if strings.HasPrefix(cursor, replacement) {
		replacement = "B"
	}
	return replacement + cursor[1:]
}

func seedNoteArtifact(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, title string, body *string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	seedRecordEnvelope(t, harness, incidentID, actorID, recordID, "artifact")
	execSeed(t, harness, `
INSERT INTO artifacts (record_id, incident_id, artifact_type, title, body, created_by_user_id)
VALUES ($1, $2, 'note', $3, $4, $5)
`, recordID, incidentID, title, body, actorID)
	execSeed(t, harness, `
INSERT INTO artifact_grid_projection (
    record_id,
    incident_id,
    row_version,
    artifact_type,
    title,
    body,
    updated_at,
    created_at,
    created_by_user_id,
    timestamp_day,
    next_report_day,
    ack_state,
    linked_record_count
)
SELECT
    a.record_id,
    a.incident_id,
    r.row_version,
    a.artifact_type,
    a.title,
    a.body,
    a.updated_at,
    a.created_at,
    a.created_by_user_id,
    a.timestamp_utc::date,
    a.next_report_at::date,
    CASE WHEN a.acknowledged_at IS NULL THEN 'pending' ELSE 'acknowledged' END,
    0
  FROM artifacts a
  JOIN records r
    ON r.incident_id = a.incident_id
   AND r.record_id = a.record_id
   AND r.deleted_at IS NULL
 WHERE a.record_id = $1
`, recordID)
	return recordID
}

func ptrString(value string) *string {
	return &value
}
