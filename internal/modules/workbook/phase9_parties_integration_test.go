package workbook_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	phase4test "github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestPhase9_PartyLinkHelperFieldsPreserveTextIndependently_I_9_03(t *testing.T) {
	harness := phase4test.StartServer(t, "phase9-i-9-03-party-links")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incidentData := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-i-9-03-incident",
		"incident_key":  "IR-I903",
		"title":         "Phase 9 I-9-03 party links",
	})
	incidentID := phase4test.MustUUID(t, incidentData["incident_id"].(string))
	otherIncidentData := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-i-9-03-other-incident",
		"incident_key":  "IR-I903B",
		"title":         "Phase 9 I-9-03 other incident",
	})
	otherIncidentID := phase4test.MustUUID(t, otherIncidentData["incident_id"].(string))

	partyData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-phase9-i-9-03-party",
		"party.display_name": "IR Vendor",
		"party.party_kind":   "organization",
	})
	partyID := phase4test.MustUUID(t, partyData["row"].(map[string]any)["record_id"].(string))
	otherPartyData := requireWorkbookCreate(t, harness, adminLogin, otherIncidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-phase9-i-9-03-other-party",
		"party.display_name": "Foreign Vendor",
		"party.party_kind":   "organization",
	})
	otherPartyID := phase4test.MustUUID(t, otherPartyData["row"].(map[string]any)["record_id"].(string))
	deletedPartyData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-phase9-i-9-03-deleted-party",
		"party.display_name": "Deleted Vendor",
		"party.party_kind":   "organization",
	})
	deletedPartyID := phase4test.MustUUID(t, deletedPartyData["row"].(map[string]any)["record_id"].(string))
	deleteDeletedTarget := deleteRecordViaWorkbookRoute(t, harness, adminLogin, deletedPartyID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-i-9-03-delete-target-party",
	})
	httptestx.RequireSuccessEnvelope(t, deleteDeletedTarget, http.StatusOK)

	evidenceData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":                 "txn-phase9-i-9-03-evidence",
		"evidence.title":                "Phase 9 party text",
		"evidence.collector_party_text": "Raw collector label",
		"evidence.source_party_text":    "Raw source label",
	})
	evidenceRow := evidenceData["row"].(map[string]any)
	evidenceID := phase4test.MustUUID(t, evidenceRow["record_id"].(string))
	requireCellValue(t, evidenceRow, "evidence.collector_party_id", nil)
	requireCellValue(t, evidenceRow, "evidence.source_party_id", nil)

	linkedCollector := requireWorkbookPatch(t, harness, adminLogin, evidenceID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": evidenceRow["row_version"],
		"client_txn_id":    "txn-phase9-i-9-03-link-collector",
		"changes": []map[string]any{
			{"field_key": "evidence.collector_party_id", "value": partyID.String()},
		},
	})["row"].(map[string]any)
	requireCellValue(t, linkedCollector, "evidence.collector_party_text", "Raw collector label")
	requireCellValue(t, linkedCollector, "evidence.collector_party_id", partyID.String())

	clearedCollectorLink := requireWorkbookPatch(t, harness, adminLogin, evidenceID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": linkedCollector["row_version"],
		"client_txn_id":    "txn-phase9-i-9-03-clear-collector-link",
		"changes": []map[string]any{
			{"field_key": "evidence.collector_party_id", "value": nil},
		},
	})["row"].(map[string]any)
	requireCellValue(t, clearedCollectorLink, "evidence.collector_party_text", "Raw collector label")
	requireCellValue(t, clearedCollectorLink, "evidence.collector_party_id", nil)

	linkedSource := requireWorkbookPatch(t, harness, adminLogin, evidenceID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": clearedCollectorLink["row_version"],
		"client_txn_id":    "txn-phase9-i-9-03-link-source",
		"changes": []map[string]any{
			{"field_key": "evidence.source_party_id", "value": partyID.String()},
		},
	})["row"].(map[string]any)
	requireCellValue(t, linkedSource, "evidence.source_party_text", "Raw source label")
	requireCellValue(t, linkedSource, "evidence.source_party_id", partyID.String())

	clearedSourceText := requireWorkbookPatch(t, harness, adminLogin, evidenceID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": linkedSource["row_version"],
		"client_txn_id":    "txn-phase9-i-9-03-clear-source-text",
		"changes": []map[string]any{
			{"field_key": "evidence.source_party_text", "value": nil},
		},
	})["row"].(map[string]any)
	requireCellValue(t, clearedSourceText, "evidence.source_party_text", nil)
	requireCellValue(t, clearedSourceText, "evidence.source_party_id", partyID.String())

	restoredSource := requireWorkbookPatch(t, harness, adminLogin, evidenceID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": clearedSourceText["row_version"],
		"client_txn_id":    "txn-phase9-i-9-03-restore-source-both",
		"changes": []map[string]any{
			{"field_key": "evidence.source_party_text", "value": "Raw source label restored"},
			{"field_key": "evidence.source_party_id", "value": partyID.String()},
		},
	})["row"].(map[string]any)
	beforeClearBothVersion := rowVersionNumber(t, restoredSource)
	clearedBoth := requireWorkbookPatch(t, harness, adminLogin, evidenceID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": restoredSource["row_version"],
		"client_txn_id":    "txn-phase9-i-9-03-clear-source-both",
		"changes": []map[string]any{
			{"field_key": "evidence.source_party_text", "value": nil},
			{"field_key": "evidence.source_party_id", "value": nil},
		},
	})["row"].(map[string]any)
	requireCellValue(t, clearedBoth, "evidence.source_party_text", nil)
	requireCellValue(t, clearedBoth, "evidence.source_party_id", nil)
	if got, want := rowVersionNumber(t, clearedBoth), beforeClearBothVersion+1; got != want {
		t.Fatalf("clear both should commit one atomic row version: got %d want %d", got, want)
	}

	taskData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.task_requests.v1", map[string]any{
		"client_txn_id":             "txn-phase9-i-9-03-task",
		"task.title":                "Party requester task",
		"task.task_kind":            "request",
		"task.requester_party_text": "Requester raw text",
	})
	taskRow := taskData["row"].(map[string]any)
	taskID := phase4test.MustUUID(t, taskRow["record_id"].(string))
	linkedRequester := requireWorkbookPatch(t, harness, adminLogin, taskID, map[string]any{
		"view_schema_id":   "cartulary.view.task_requests.v1",
		"base_row_version": taskRow["row_version"],
		"client_txn_id":    "txn-phase9-i-9-03-link-requester",
		"changes": []map[string]any{
			{"field_key": "task.requester_party_id", "value": partyID.String()},
		},
	})["row"].(map[string]any)
	requireCellValue(t, linkedRequester, "task.requester_party_text", "Requester raw text")
	requireCellValue(t, linkedRequester, "task.requester_party_id", partyID.String())

	for _, tc := range []struct {
		name         string
		viewSchemaID string
		recordID     uuid.UUID
		baseRow      map[string]any
		fieldKey     string
		wantTextKey  string
		wantText     any
		wantRef      any
	}{
		{
			name:         "collector",
			viewSchemaID: "cartulary.view.evidence.v1",
			recordID:     evidenceID,
			baseRow:      clearedBoth,
			fieldKey:     "evidence.collector_party_id",
			wantTextKey:  "evidence.collector_party_text",
			wantText:     "Raw collector label",
			wantRef:      nil,
		},
		{
			name:         "source",
			viewSchemaID: "cartulary.view.evidence.v1",
			recordID:     evidenceID,
			baseRow:      clearedBoth,
			fieldKey:     "evidence.source_party_id",
			wantTextKey:  "evidence.source_party_text",
			wantText:     nil,
			wantRef:      nil,
		},
		{
			name:         "requester",
			viewSchemaID: "cartulary.view.task_requests.v1",
			recordID:     taskID,
			baseRow:      linkedRequester,
			fieldKey:     "task.requester_party_id",
			wantTextKey:  "task.requester_party_text",
			wantText:     "Requester raw text",
			wantRef:      partyID.String(),
		},
	} {
		for _, target := range []struct {
			name string
			id   uuid.UUID
		}{
			{name: "foreign", id: otherPartyID},
			{name: "deleted", id: deletedPartyID},
			{name: "wrong-type", id: evidenceID},
			{name: "user-id", id: adminUserID},
		} {
			resp := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", tc.recordID, map[string]any{
				"view_schema_id":   tc.viewSchemaID,
				"base_row_version": tc.baseRow["row_version"],
				"client_txn_id":    "txn-phase9-i-9-03-" + tc.name + "-" + target.name,
				"changes": []map[string]any{
					{"field_key": tc.fieldKey, "value": target.id.String()},
				},
			})
			httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
			rowAfterFailure := queryWorkbookRow(t, harness, adminLogin, incidentID, tc.viewSchemaID, tc.recordID)
			requireCellValue(t, rowAfterFailure, tc.wantTextKey, tc.wantText)
			requireCellValue(t, rowAfterFailure, tc.fieldKey, tc.wantRef)
			if got, want := rowVersionNumber(t, rowAfterFailure), rowVersionNumber(t, tc.baseRow); got != want {
				t.Fatalf("%s %s invalid write changed row version: got %d want %d", tc.name, target.name, got, want)
			}
		}
	}

	refOnlyTaskData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.task_requests.v1", map[string]any{
		"client_txn_id":             "txn-phase9-i-9-03-task-ref-only",
		"task.title":                "Party requester ref-only task",
		"task.task_kind":            "request",
		"task.requester_party_text": "Requester raw text for ref-only clear",
	})
	refOnlyTaskRow := refOnlyTaskData["row"].(map[string]any)
	refOnlyTaskID := phase4test.MustUUID(t, refOnlyTaskRow["record_id"].(string))
	linkedRefOnlyRequester := requireWorkbookPatch(t, harness, adminLogin, refOnlyTaskID, map[string]any{
		"view_schema_id":   "cartulary.view.task_requests.v1",
		"base_row_version": refOnlyTaskRow["row_version"],
		"client_txn_id":    "txn-phase9-i-9-03-link-ref-only-requester",
		"changes": []map[string]any{
			{"field_key": "task.requester_party_id", "value": partyID.String()},
		},
	})["row"].(map[string]any)
	requireCellValue(t, linkedRefOnlyRequester, "task.requester_party_text", "Requester raw text for ref-only clear")
	requireCellValue(t, linkedRefOnlyRequester, "task.requester_party_id", partyID.String())

	clearedRefOnlyText := requireWorkbookPatch(t, harness, adminLogin, refOnlyTaskID, map[string]any{
		"view_schema_id":   "cartulary.view.task_requests.v1",
		"base_row_version": linkedRefOnlyRequester["row_version"],
		"client_txn_id":    "txn-phase9-i-9-03-clear-ref-only-requester-text",
		"changes": []map[string]any{
			{"field_key": "task.requester_party_text", "value": nil},
		},
	})["row"].(map[string]any)
	requireCellValue(t, clearedRefOnlyText, "task.requester_party_text", nil)
	requireCellValue(t, clearedRefOnlyText, "task.requester_party_id", partyID.String())

	beforeTaskClearBothChangeSets := phase9PartiesQueryCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1`, incidentID)
	beforeTaskClearBothMutations := phase9PartiesQueryCount(t, harness, `SELECT count(*) FROM change_set_mutations WHERE target_id = $1`, refOnlyTaskID.String())
	beforeTaskClearBothRevisions := phase9PartiesQueryCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, refOnlyTaskID)
	beforeTaskClearBothVersion := rowVersionNumber(t, clearedRefOnlyText)
	clearedRefOnlyBoth := requireWorkbookPatch(t, harness, adminLogin, refOnlyTaskID, map[string]any{
		"view_schema_id":   "cartulary.view.task_requests.v1",
		"base_row_version": clearedRefOnlyText["row_version"],
		"client_txn_id":    "txn-phase9-i-9-03-clear-ref-only-requester-both",
		"changes": []map[string]any{
			{"field_key": "task.requester_party_text", "value": nil},
			{"field_key": "task.requester_party_id", "value": nil},
		},
	})["row"].(map[string]any)
	requireCellValue(t, clearedRefOnlyBoth, "task.requester_party_text", nil)
	requireCellValue(t, clearedRefOnlyBoth, "task.requester_party_id", nil)
	if got, want := rowVersionNumber(t, clearedRefOnlyBoth), beforeTaskClearBothVersion+1; got != want {
		t.Fatalf("task requester clear both should commit one atomic row version: got %d want %d", got, want)
	}
	if got, want := phase9PartiesQueryCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1`, incidentID), beforeTaskClearBothChangeSets+1; got != want {
		t.Fatalf("task requester clear both change set count: got %d want %d", got, want)
	}
	if got, want := phase9PartiesQueryCount(t, harness, `SELECT count(*) FROM change_set_mutations WHERE target_id = $1`, refOnlyTaskID.String()), beforeTaskClearBothMutations+1; got != want {
		t.Fatalf("task requester clear both mutation count: got %d want %d", got, want)
	}
	if got, want := phase9PartiesQueryCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, refOnlyTaskID), beforeTaskClearBothRevisions+1; got != want {
		t.Fatalf("task requester clear both revision count: got %d want %d", got, want)
	}
	queriedRefOnlyTask := queryWorkbookRow(t, harness, adminLogin, incidentID, "cartulary.view.task_requests.v1", refOnlyTaskID)
	requireCellValue(t, queriedRefOnlyTask, "task.requester_party_text", nil)
	requireCellValue(t, queriedRefOnlyTask, "task.requester_party_id", nil)
	if got, want := rowVersionNumber(t, queriedRefOnlyTask), rowVersionNumber(t, clearedRefOnlyBoth); got != want {
		t.Fatalf("task requester clear both query row version: got %d want %d", got, want)
	}

	beforeReferencedDeleteRevisions := partyRecordRevisionCount(t, harness, partyID)
	referencedDelete := deleteRecordViaWorkbookRoute(t, harness, adminLogin, partyID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-i-9-03-delete-referenced-party",
	})
	referencedDeleteBody := httptestx.RequireErrorEnvelope(t, referencedDelete, http.StatusConflict, "record_delete_blocked")
	referencedDeleteDetails := referencedDeleteBody["error"].(map[string]any)["details"].(map[string]any)
	if referencedDeleteDetails["reason_code"] != "active_incoming_party_reference" {
		t.Fatalf("referenced party delete reason_code = %#v", referencedDeleteDetails)
	}
	if got := partyRecordCount(t, harness, partyID, "deleted_at IS NULL AND row_version = 1"); got != 1 {
		t.Fatalf("referenced party delete mutated party record, count=%d", got)
	}
	if got := partyRecordRevisionCount(t, harness, partyID); got != beforeReferencedDeleteRevisions {
		t.Fatalf("blocked party delete wrote revisions: got %d want %d", got, beforeReferencedDeleteRevisions)
	}

	collectionPartyData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-phase9-i-9-03-collection-party",
		"party.display_name": "Collection Party",
		"party.party_kind":   "team",
	})
	collectionPartyID := phase4test.MustUUID(t, collectionPartyData["row"].(map[string]any)["record_id"].(string))

	commData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.comm_log.v1", map[string]any{
		"client_txn_id":               "txn-phase9-i-9-03-comm-preservation",
		"comm_log.comm_type":          "briefing",
		"comm_log.audience":           "Collection audience preserved",
		"comm_log.channel_or_meeting": "Bridge",
		"comm_log.summary":            "Collection party ref preservation",
	})
	commRow := commData["row"].(map[string]any)
	commID := phase4test.MustUUID(t, commRow["record_id"].(string))
	for _, fieldKey := range []string{"comm_log.audience_party_ids", "comm_log.attendee_party_ids"} {
		linkedComm := requireWorkbookPatch(t, harness, adminLogin, commID, map[string]any{
			"view_schema_id":   "cartulary.view.comm_log.v1",
			"base_row_version": commRow["row_version"],
			"client_txn_id":    "txn-phase9-i-9-03-add-" + fieldKey,
			"changes": []map[string]any{
				{"field_key": fieldKey, "action_payload": collectionActions(addPartyRef(collectionPartyID))},
			},
		})["row"].(map[string]any)
		requireCellValue(t, linkedComm, "comm_log.audience", "Collection audience preserved")
		requireCollectionItemCount(t, linkedComm, fieldKey, 1)

		clearedComm := requireWorkbookPatch(t, harness, adminLogin, commID, map[string]any{
			"view_schema_id":   "cartulary.view.comm_log.v1",
			"base_row_version": linkedComm["row_version"],
			"client_txn_id":    "txn-phase9-i-9-03-remove-" + fieldKey,
			"changes": []map[string]any{
				{"field_key": fieldKey, "action_payload": collectionActions(removePartyRef(collectionPartyID))},
			},
		})["row"].(map[string]any)
		requireCellValue(t, clearedComm, "comm_log.audience", "Collection audience preserved")
		requireCollectionItemCount(t, clearedComm, fieldKey, 0)
		commRow = clearedComm
	}

	for _, fieldKey := range []string{"comm_log.audience_party_ids", "comm_log.attendee_party_ids"} {
		for _, target := range []struct {
			name string
			id   uuid.UUID
		}{
			{name: "foreign", id: otherPartyID},
			{name: "deleted", id: deletedPartyID},
			{name: "wrong-type", id: evidenceID},
			{name: "user-id", id: adminUserID},
		} {
			resp := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
				"view_schema_id":   "cartulary.view.comm_log.v1",
				"base_row_version": commRow["row_version"],
				"client_txn_id":    "txn-phase9-i-9-03-" + fieldKey + "-" + target.name,
				"changes": []map[string]any{
					{"field_key": fieldKey, "action_payload": collectionActions(addPartyRef(target.id))},
				},
			})
			httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
			rowAfterFailure := queryWorkbookRow(t, harness, adminLogin, incidentID, "cartulary.view.comm_log.v1", commID)
			requireCellValue(t, rowAfterFailure, "comm_log.audience", "Collection audience preserved")
			requireCollectionItemCount(t, rowAfterFailure, fieldKey, 0)
			if got, want := rowVersionNumber(t, rowAfterFailure), rowVersionNumber(t, commRow); got != want {
				t.Fatalf("%s %s invalid write changed row version: got %d want %d", fieldKey, target.name, got, want)
			}
		}
	}

	requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.comm_log.v1", map[string]any{
		"client_txn_id":               "txn-phase9-i-9-03-comm-party-ref",
		"comm_log.comm_type":          "briefing",
		"comm_log.audience":           "Collection audience",
		"comm_log.channel_or_meeting": "Bridge",
		"comm_log.summary":            "Collection party ref delete guard",
		"comm_log.audience_party_ids": collectionActions(addPartyRef(collectionPartyID)),
	})
	collectionDelete := deleteRecordViaWorkbookRoute(t, harness, adminLogin, collectionPartyID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-i-9-03-delete-collection-party",
	})
	httptestx.RequireErrorEnvelope(t, collectionDelete, http.StatusConflict, "record_delete_blocked")
	if got := partyRecordCount(t, harness, collectionPartyID, "deleted_at IS NULL AND row_version = 1"); got != 1 {
		t.Fatalf("collection-referenced party delete mutated party record, count=%d", got)
	}

	clearedRequester := requireWorkbookPatch(t, harness, adminLogin, taskID, map[string]any{
		"view_schema_id":   "cartulary.view.task_requests.v1",
		"base_row_version": linkedRequester["row_version"],
		"client_txn_id":    "txn-phase9-i-9-03-clear-requester-before-delete",
		"changes": []map[string]any{
			{"field_key": "task.requester_party_id", "value": nil},
		},
	})["row"].(map[string]any)
	requireCellValue(t, clearedRequester, "task.requester_party_text", "Requester raw text")
	requireCellValue(t, clearedRequester, "task.requester_party_id", nil)
	unreferencedDelete := deleteRecordViaWorkbookRoute(t, harness, adminLogin, partyID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-i-9-03-delete-unreferenced-party",
	})
	httptestx.RequireSuccessEnvelope(t, unreferencedDelete, http.StatusOK)
	if got := partyRecordCount(t, harness, partyID, "deleted_at IS NOT NULL AND deleted_by_user_id = $2 AND row_version = 2", adminUserID); got != 1 {
		t.Fatalf("unreferenced party delete did not tombstone party record, count=%d", got)
	}
}

func queryWorkbookRow(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, viewSchemaID string, recordID uuid.UUID) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/query",
		map[string]any{},
		phase4test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	for _, rawRow := range data["rows"].([]any) {
		row := rawRow.(map[string]any)
		if row["record_id"] == recordID.String() {
			return row
		}
	}
	t.Fatalf("missing row %s in %s query", recordID, viewSchemaID)
	return nil
}

func rowVersionNumber(t testing.TB, row map[string]any) int {
	t.Helper()
	switch value := row["row_version"].(type) {
	case float64:
		return int(value)
	case int64:
		return int(value)
	case int:
		return value
	default:
		t.Fatalf("unexpected row_version type %T", value)
		return 0
	}
}

func phase9PartiesQueryCount(t testing.TB, harness *phase4test.ServerHarness, query string, args ...any) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func deleteRecordViaWorkbookRoute(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, recordID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	return phase4test.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(),
		body,
		phase4test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func partyRecordCount(t testing.TB, harness *phase4test.ServerHarness, partyID uuid.UUID, predicate string, args ...any) int {
	t.Helper()
	queryArgs := append([]any{partyID}, args...)
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM records WHERE record_id = $1 AND `+predicate, queryArgs...).Scan(&count); err != nil {
		t.Fatalf("count party record: %v", err)
	}
	return count
}

func partyRecordRevisionCount(t testing.TB, harness *phase4test.ServerHarness, partyID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM record_revisions WHERE record_id = $1`, partyID).Scan(&count); err != nil {
		t.Fatalf("count party record revisions: %v", err)
	}
	return count
}

func removePartyRef(partyID uuid.UUID) map[string]any {
	return map[string]any{"op": "remove_party_ref", "item_ref": "party_ref:" + partyID.String()}
}
