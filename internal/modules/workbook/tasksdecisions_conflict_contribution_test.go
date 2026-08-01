package workbook_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestTasksDecisionsConflictResolutionContribution_Integration(t *testing.T) {
	testCases := []tasksDecisionsConflictCase{
		{
			name:           "task_request",
			viewSchemaID:   "cartulary.view.task_requests.v1",
			fieldKey:       "task.title",
			baseValue:      "Task base",
			savedValue:     "Task saved",
			localValue:     "Task local",
			laterSaved:     "Task saved after token",
			createFieldKey: "task.title",
			createValues: map[string]any{
				"task.task_kind": "request",
			},
		},
		{
			name:           "decision",
			viewSchemaID:   "cartulary.view.decisions.v1",
			fieldKey:       "decision.summary",
			baseValue:      "Decision base",
			savedValue:     "Decision saved",
			localValue:     "Decision local",
			laterSaved:     "Decision saved after token",
			createFieldKey: "decision.summary",
			createValues: map[string]any{
				"decision.decision_type": "containment",
				"decision.rationale":     "Decision rationale",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			harness, login, actorID, incidentID := ConflictFixture(
				t,
				"tasksdecisions-conflict-contribution-"+testCase.name,
				"IR-TASKSDECISIONS-CONFLICT-"+testCase.name,
			)

			keepID, keepConflict := createTasksDecisionsConflict(
				t, harness, login, incidentID, testCase, "keep",
			)
			beforeKeep := snapshotTasksDecisionsConflictEffects(t, harness, incidentID, keepID)
			keepBody := map[string]any{
				"conflict_token":  keepConflict["conflict_token"].(string),
				"resolution_kind": "keep_saved",
				"client_txn_id":   "txn-tasksdecisions-conflict-" + testCase.name + "-keep-resolve",
			}
			keepData := ResolveConflict(
				t, harness, login, keepID, keepConflict["conflict_token"].(string), keepBody,
			)
			RequireNoChangeSetID(t, keepData)
			requireCellValue(t, keepData["row"].(map[string]any), testCase.fieldKey, testCase.savedValue)
			requireTasksDecisionsConflictDelta(
				t,
				testCase.name+" keep_saved",
				beforeKeep,
				snapshotTasksDecisionsConflictEffects(t, harness, incidentID, keepID),
				tasksDecisionsConflictEffects{RouteIdempotency: 1},
			)

			keepReplay := ResolveConflict(
				t, harness, login, keepID, keepConflict["conflict_token"].(string), keepBody,
			)
			if !reflect.DeepEqual(keepReplay, keepData) {
				t.Fatalf("keep_saved replay changed the committed response:\nfirst  %#v\nreplay %#v", keepData, keepReplay)
			}
			requireTasksDecisionsConflictDelta(
				t,
				testCase.name+" keep_saved replay",
				beforeKeep,
				snapshotTasksDecisionsConflictEffects(t, harness, incidentID, keepID),
				tasksDecisionsConflictEffects{RouteIdempotency: 1},
			)

			changedKeep := map[string]any{
				"conflict_token":  keepConflict["conflict_token"].(string),
				"resolution_kind": "use_unsaved",
				"client_txn_id":   keepBody["client_txn_id"],
				"resolved_value":  testCase.localValue,
			}
			changedResponse := ResolveConflictRaw(
				t, harness, login, keepID, keepConflict["conflict_token"].(string), changedKeep,
			)
			httptestx.RequireErrorEnvelope(t, changedResponse, http.StatusConflict, "client_txn_conflict")
			requireTasksDecisionsConflictDelta(
				t,
				testCase.name+" changed replay",
				beforeKeep,
				snapshotTasksDecisionsConflictEffects(t, harness, incidentID, keepID),
				tasksDecisionsConflictEffects{RouteIdempotency: 1},
			)

			useID, useConflict := createTasksDecisionsConflict(
				t, harness, login, incidentID, testCase, "use",
			)
			beforeUse := snapshotTasksDecisionsConflictEffects(t, harness, incidentID, useID)
			useBody := map[string]any{
				"conflict_token":  useConflict["conflict_token"].(string),
				"resolution_kind": "use_unsaved",
				"client_txn_id":   "txn-tasksdecisions-conflict-" + testCase.name + "-use-resolve",
				"resolved_value":  testCase.localValue,
			}
			useData := ResolveConflict(
				t, harness, login, useID, useConflict["conflict_token"].(string), useBody,
			)
			requireCellValue(t, useData["row"].(map[string]any), testCase.fieldKey, testCase.localValue)
			workbookscenariotest.RequireChangeSetAttribution(
				t,
				harness.DB,
				useData["change_set_id"].(string),
				actorID.String(),
				"workbook.records.conflicts.resolve",
				useBody["client_txn_id"].(string),
			)
			requireTasksDecisionsConflictDelta(
				t,
				testCase.name+" use_unsaved",
				beforeUse,
				snapshotTasksDecisionsConflictEffects(t, harness, incidentID, useID),
				tasksDecisionsConflictEffects{
					ChangeSets: 1, Mutations: 1, RecordRevisions: 1,
					RouteIdempotency: 1, CollaborationIntents: 1, RowVersion: 1,
				},
			)

			useReplay := ResolveConflict(
				t, harness, login, useID, useConflict["conflict_token"].(string), useBody,
			)
			if !reflect.DeepEqual(useReplay, useData) {
				t.Fatalf("use_unsaved replay changed the committed response:\nfirst  %#v\nreplay %#v", useData, useReplay)
			}
			requireTasksDecisionsConflictDelta(
				t,
				testCase.name+" use_unsaved replay",
				beforeUse,
				snapshotTasksDecisionsConflictEffects(t, harness, incidentID, useID),
				tasksDecisionsConflictEffects{
					ChangeSets: 1, Mutations: 1, RecordRevisions: 1,
					RouteIdempotency: 1, CollaborationIntents: 1, RowVersion: 1,
				},
			)

			staleID, staleConflict := createTasksDecisionsConflict(
				t, harness, login, incidentID, testCase, "stale",
			)
			requireWorkbookPatch(t, harness, login, staleID, map[string]any{
				"view_schema_id":   testCase.viewSchemaID,
				"base_row_version": 2,
				"client_txn_id":    "txn-tasksdecisions-conflict-" + testCase.name + "-stale-advance",
				"changes": []map[string]any{{
					"field_key": testCase.fieldKey,
					"value":     testCase.laterSaved,
				}},
			})
			beforeStale := snapshotTasksDecisionsConflictEffects(t, harness, incidentID, staleID)
			staleResponse := ResolveConflictRaw(
				t,
				harness,
				login,
				staleID,
				staleConflict["conflict_token"].(string),
				map[string]any{
					"conflict_token":  staleConflict["conflict_token"].(string),
					"resolution_kind": "use_unsaved",
					"client_txn_id":   "txn-tasksdecisions-conflict-" + testCase.name + "-stale-resolve",
					"resolved_value":  testCase.localValue,
				},
			)
			staleBody := httptestx.RequireErrorEnvelope(t, staleResponse, http.StatusConflict, "same_field_conflict")
			fresh := staleBody["error"].(map[string]any)["conflict"].(map[string]any)
			if fresh["conflict_token"] == staleConflict["conflict_token"] {
				t.Fatalf("stale resolution reused its conflict token: %#v", fresh)
			}
			if fresh["server_value"] != testCase.laterSaved {
				t.Fatalf("stale resolution did not revalidate the current value: %#v", fresh)
			}
			if beforeStale != snapshotTasksDecisionsConflictEffects(t, harness, incidentID, staleID) {
				t.Fatalf("stale conflict resolution wrote durable effects")
			}
		})
	}
}

type tasksDecisionsConflictCase struct {
	name           string
	viewSchemaID   string
	fieldKey       string
	baseValue      string
	savedValue     string
	localValue     string
	laterSaved     string
	createFieldKey string
	createValues   map[string]any
}

func createTasksDecisionsConflict(
	t testing.TB,
	harness *appsupport.ServerHarness,
	login appsupport.LoginResult,
	incidentID uuid.UUID,
	testCase tasksDecisionsConflictCase,
	suffix string,
) (uuid.UUID, map[string]any) {
	t.Helper()
	createBody := map[string]any{
		"client_txn_id": "txn-tasksdecisions-conflict-" + testCase.name + "-" + suffix + "-create",
	}
	createBody[testCase.createFieldKey] = testCase.baseValue
	for fieldKey, value := range testCase.createValues {
		createBody[fieldKey] = value
	}
	created := requireWorkbookCreate(
		t, harness, login, incidentID, testCase.viewSchemaID, createBody,
	)["row"].(map[string]any)
	recordID := appsupport.MustUUID(t, created["record_id"].(string))
	conflict := createAtomicConflict(
		t,
		harness,
		login,
		recordID,
		testCase.viewSchemaID,
		testCase.fieldKey,
		testCase.savedValue,
		testCase.localValue,
		"tasksdecisions-"+testCase.name+"-"+suffix,
	)
	return recordID, conflict
}

type tasksDecisionsConflictEffects struct {
	ChangeSets           int
	Mutations            int
	RecordRevisions      int
	RouteIdempotency     int
	CollaborationIntents int
	RowVersion           int
}

func snapshotTasksDecisionsConflictEffects(
	t testing.TB,
	harness *appsupport.ServerHarness,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) tasksDecisionsConflictEffects {
	t.Helper()
	var effects tasksDecisionsConflictEffects
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT
    (SELECT count(*) FROM change_sets WHERE incident_id = $1),
    (SELECT count(*) FROM change_set_mutations m JOIN change_sets c USING (change_set_id) WHERE c.incident_id = $1),
    (SELECT count(*) FROM record_revisions WHERE record_id = $2),
    (SELECT count(*) FROM route_idempotency WHERE scope_key = $2::text),
    (SELECT count(*) FROM collaboration_event_intents WHERE source_record_id = $2),
    (SELECT row_version FROM records WHERE record_id = $2)
`, incidentID, recordID).Scan(
		&effects.ChangeSets,
		&effects.Mutations,
		&effects.RecordRevisions,
		&effects.RouteIdempotency,
		&effects.CollaborationIntents,
		&effects.RowVersion,
	); err != nil {
		t.Fatalf("snapshot Tasks/Decisions conflict effects: %v", err)
	}
	return effects
}

func requireTasksDecisionsConflictDelta(
	t testing.TB,
	label string,
	before tasksDecisionsConflictEffects,
	after tasksDecisionsConflictEffects,
	want tasksDecisionsConflictEffects,
) {
	t.Helper()
	got := tasksDecisionsConflictEffects{
		ChangeSets:           after.ChangeSets - before.ChangeSets,
		Mutations:            after.Mutations - before.Mutations,
		RecordRevisions:      after.RecordRevisions - before.RecordRevisions,
		RouteIdempotency:     after.RouteIdempotency - before.RouteIdempotency,
		CollaborationIntents: after.CollaborationIntents - before.CollaborationIntents,
		RowVersion:           after.RowVersion - before.RowVersion,
	}
	if got != want {
		t.Fatalf("%s side-effect delta = %+v want %+v", label, got, want)
	}
}
