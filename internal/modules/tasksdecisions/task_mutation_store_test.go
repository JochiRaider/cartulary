package tasksdecisions_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"

	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookroutetest"
)

func TestTaskRequestLifecycleDecisionLinksAndProjection_Unit(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "workbook_interaction-task-decision-task-requests")
	codec := conflicttest.NewCodec("workbook")
	store := appsupport.NewWorkbookCatalog(harness.DB, codec)
	owner := appsupport.NewTaskDecisionOwner(harness.DB, codec)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "task-decision-task@example.test", "TaskDecision Task", "TaskDecisionTask1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-task-decision-task-incident", "IR-TASK-DECISION-TASK", "Workbook inspector task and decision workflow task requests")

	beforeRecords := countRecords(t, harness.DB, incident.ID)
	minimumRequest := tasksdecisions.CreateRequest{
		ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-task-decision-task-minimum-fail",
		Values: map[string]tasksdecisions.FieldValue{
			"task.task_kind": {Text: stringPtr("collection")},
		},
	}
	_, err := createTaskDecision(owner, actor, incident.ID, minimumRequest, "req-workbook_interaction-task-decision-task-minimum-fail", testTime(0))
	requireMutationValidation(t, err, "task.title", "missing_required_field")
	if got := countRecords(t, harness.DB, incident.ID); got != beforeRecords {
		t.Fatalf("rejected minimum task create wrote records: got %d want %d", got, beforeRecords)
	}

	decision := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-decision", "approved", "Task-linked decision")
	support := mustCreateEvidence(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-task-decision-task-support", "Task support record")
	dueAt := testTime(24 * time.Hour)
	task := mustCreateTask(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-create", map[string]tasksdecisions.FieldValue{
		"task.title":               {Text: stringPtr("Collect endpoint logs")},
		"task.task_kind":           {Text: stringPtr("collection")},
		"task.due_at":              {Timestamp: &dueAt},
		"task.workstream":          {Text: stringPtr("forensics")},
		"task.external_ticket_ref": {Text: stringPtr("SOC-123")},
		"task.decision_record_id":  {UUID: &decision},
	}, map[string]tasksdecisions.CollectionActionPayload{
		"task.linked_record_ids": collectionActions(addOptionalSurfaceRecordRef(support)),
	})
	taskID := task.RecordID
	requireCellValue(t, task.Row, "task.priority", "normal")
	requireCellValue(t, task.Row, "task.decision_record_id", decision.String())
	if got := countTaskDecisionLinks(t, harness.DB, taskID); got != 1 {
		t.Fatalf("task decision direct link count: got %d want 1", got)
	}
	requireManualReferenceLink(t, harness.DB, taskID, decision, "task.decision_record_id", "references_record")
	requireManualReferenceLink(t, harness.DB, taskID, support, "task.linked_record_ids", "references_record")

	taskRows, err := workbookroutetest.QueryRows(store, ctx, incident.ID, tasksdecisions.TaskRequestsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "task.status", Op: "eq", Arg: map[string]any{"value": "open"}},
			{FieldKey: "task.owner_user_id", Op: "eq", Arg: map[string]any{"value": actor.ID.String()}},
			{FieldKey: "task.due_at", Op: "range", Arg: map[string]any{"lte": dueAt.Format(time.RFC3339Nano)}},
		},
		Sort: []viewschema.SortEntry{{FieldKey: "task.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query task projection filters: %v", err)
	}
	if len(taskRows) != 1 || taskRows[0]["record_id"] != taskID.String() {
		t.Fatalf("projection-backed queue filters returned %#v", taskRows)
	}

	queueTask := mustCreateTask(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-queue-default", map[string]tasksdecisions.FieldValue{
		"task.title":     {Text: stringPtr("Queue priority default")},
		"task.task_kind": {Text: stringPtr("follow_up")},
	}, nil)
	queueTaskID := queueTask.RecordID
	requireCellValue(t, queueTask.Row, "task.priority", "normal")
	requireCellValue(t, queueTask.Row, "task.due_at", nil)
	queueDueAt := testTime(48 * time.Hour)
	queueTask = mustPatch(t, owner, actor, queueTaskID, tasksdecisions.TaskRequestsViewSchemaID, 1, "txn-workbook_interaction-task-decision-task-queue-edit",
		valueChange("task.priority", tasksdecisions.FieldValue{Text: stringPtr("high")}),
		valueChange("task.due_at", tasksdecisions.FieldValue{Timestamp: &queueDueAt}))
	requireCellValue(t, queueTask.Row, "task.priority", "high")
	requireCellNonEmpty(t, queueTask.Row, "task.due_at")
	priorityRows, err := workbookroutetest.QueryRows(store, ctx, incident.ID, tasksdecisions.TaskRequestsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "task.priority", Op: "eq", Arg: map[string]any{"value": "high"}},
			{FieldKey: "task.due_at", Op: "range", Arg: map[string]any{"lte": queueDueAt.Format(time.RFC3339Nano)}},
		},
		Sort: []viewschema.SortEntry{{FieldKey: "task.priority", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query task priority and due filters: %v", err)
	}
	if !rowsContain(priorityRows, queueTaskID) {
		t.Fatalf("priority/due projection query missing queue task: %#v", priorityRows)
	}
	urgentTask := mustCreateTask(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-queue-urgent", map[string]tasksdecisions.FieldValue{
		"task.title":     {Text: stringPtr("Queue priority urgent")},
		"task.task_kind": {Text: stringPtr("follow_up")},
		"task.priority":  {Text: stringPtr("urgent")},
	}, nil)
	sortedRows, err := workbookroutetest.QueryRows(store, ctx, incident.ID, tasksdecisions.TaskRequestsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "task.status", Op: "eq", Arg: map[string]any{"value": "open"}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "task.priority", Direction: "desc"}, {FieldKey: "record_id", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("query task priority sort: %v", err)
	}
	if len(sortedRows) == 0 || sortedRows[0]["record_id"] != urgentTask.RecordID.String() {
		t.Fatalf("priority sort did not put urgent task first: %#v", sortedRows)
	}
	queueTask = mustPatch(t, owner, actor, queueTaskID, tasksdecisions.TaskRequestsViewSchemaID, 2, "txn-workbook_interaction-task-decision-task-queue-clear-due",
		valueChange("task.due_at", tasksdecisions.FieldValue{}))
	requireCellValue(t, queueTask.Row, "task.due_at", nil)
	clearedDueRows, err := workbookroutetest.QueryRows(store, ctx, incident.ID, tasksdecisions.TaskRequestsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "task.priority", Op: "eq", Arg: map[string]any{"value": "high"}},
			{FieldKey: "task.due_at", Op: "range", Arg: map[string]any{"lte": queueDueAt.Format(time.RFC3339Nano)}},
		},
		Sort: []viewschema.SortEntry{{FieldKey: "task.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query task due after clear: %v", err)
	}
	if rowsContain(clearedDueRows, queueTaskID) {
		t.Fatalf("cleared due date still matched due range: %#v", clearedDueRows)
	}
	beforeInvalidPriorityVersion := recordVersion(t, harness.DB, queueTaskID)
	_, err = patchRecord(owner, actor, queueTaskID, tasksdecisions.TaskRequestsViewSchemaID, 3, "txn-workbook_interaction-task-decision-task-invalid-priority",
		valueChange("task.priority", tasksdecisions.FieldValue{Text: stringPtr("critical")}))
	requireMutationValidation(t, err, "task.priority", "invalid_value")
	if got := recordVersion(t, harness.DB, queueTaskID); got != beforeInvalidPriorityVersion {
		t.Fatalf("invalid priority changed row version: got %d want %d", got, beforeInvalidPriorityVersion)
	}

	task = mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 1, "txn-workbook_interaction-task-decision-task-in-progress",
		valueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("in_progress")}))
	task = mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 2, "txn-workbook_interaction-task-decision-task-blocked",
		valueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("blocked")}),
		valueChange("task.blocked_reason", tasksdecisions.FieldValue{Text: stringPtr("Waiting for host owner")}))
	requireCellValue(t, task.Row, "task.status", "blocked")
	requireCellValue(t, task.Row, "task.blocked_reason", "Waiting for host owner")
	task = mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 3, "txn-workbook_interaction-task-decision-task-unblocked",
		valueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("in_progress")}))
	requireCellValue(t, task.Row, "task.blocked_reason", nil)
	task = mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 4, "txn-workbook_interaction-task-decision-task-done",
		valueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("done")}))
	requireCellNonEmpty(t, task.Row, "task.completed_at")
	task = mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 5, "txn-workbook_interaction-task-decision-task-reopen",
		valueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("open")}))
	requireCellValue(t, task.Row, "task.completed_at", nil)

	cleared := mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 6, "txn-workbook_interaction-task-decision-task-clear-decision",
		valueChange("task.decision_record_id", tasksdecisions.FieldValue{}))
	requireCellValue(t, cleared.Row, "task.decision_record_id", nil)
	if got := countTaskDecisionLinks(t, harness.DB, taskID); got != 0 {
		t.Fatalf("task decision direct link after clear: got %d want 0", got)
	}
	reset := mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 7, "txn-workbook_interaction-task-decision-task-reset-decision",
		valueChange("task.decision_record_id", tasksdecisions.FieldValue{UUID: &decision}))
	requireCellValue(t, reset.Row, "task.decision_record_id", decision.String())
	if got := countTaskDecisionLinks(t, harness.DB, taskID); got != 1 {
		t.Fatalf("task decision direct link after reset: got %d want 1", got)
	}
	requireManualReferenceLink(t, harness.DB, taskID, decision, "task.decision_record_id", "references_record")
	otherIncident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-task-decision-task-other-incident", "IR-TASK-DECISION-TASK-OTHER", "Workbook inspector task and decision workflow task requests other")
	foreignDecision := mustCreateDecision(t, owner, actor, otherIncident.ID, "txn-workbook_interaction-task-decision-task-foreign-decision", "approved", "Foreign decision")
	deletedDecision := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-deleted-decision", "approved", "Deleted decision")
	if _, err := harness.DB.Exec(ctx, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, deletedDecision, testTime(20*time.Minute), actor.ID); err != nil {
		t.Fatalf("soft-delete decision target: %v", err)
	}
	for _, invalid := range []struct {
		name string
		id   uuid.UUID
	}{
		{name: "foreign", id: foreignDecision},
		{name: "deleted", id: deletedDecision},
		{name: "wrong-type", id: support},
	} {
		_, err = patchRecord(owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 8, "txn-workbook_interaction-task-decision-task-invalid-decision-"+invalid.name,
			valueChange("task.decision_record_id", tasksdecisions.FieldValue{UUID: &invalid.id}))
		requireMutationValidation(t, err, "task.decision_record_id", "invalid_value")
		if got := countTaskDecisionLinks(t, harness.DB, taskID); got != 1 {
			t.Fatalf("invalid %s decision ref changed links: got %d want 1", invalid.name, got)
		}
	}

	done := mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 8, "txn-workbook_interaction-task-decision-task-done-again",
		valueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("done")}))
	requireCellValue(t, done.Row, "task.status", "done")
	_, err = patchRecord(owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 9, "txn-workbook_interaction-task-decision-task-done-canceled",
		valueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("canceled")}))
	requireLifecycle(t, err)

	canceled := mustCreateTask(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-canceled", map[string]tasksdecisions.FieldValue{
		"task.title":     {Text: stringPtr("Canceled task")},
		"task.task_kind": {Text: stringPtr("request")},
		"task.status":    {Text: stringPtr("canceled")},
	}, nil)
	_, err = patchRecord(owner, actor, canceled.RecordID, tasksdecisions.TaskRequestsViewSchemaID, 1, "txn-workbook_interaction-task-decision-task-canceled-done",
		valueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("done")}))
	requireLifecycle(t, err)

	ownerless := mustCreateTask(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-ownerless", map[string]tasksdecisions.FieldValue{
		"task.title":     {Text: stringPtr("Owner guard")},
		"task.task_kind": {Text: stringPtr("request")},
	}, nil)
	_, err = patchRecord(owner, actor, ownerless.RecordID, tasksdecisions.TaskRequestsViewSchemaID, 1, "txn-workbook_interaction-task-decision-task-ownerless-open",
		valueChange("task.owner_user_id", tasksdecisions.FieldValue{}))
	requireLifecycle(t, err)
}

func TestTaskLifecycleGuardFailures_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-task-decision-task-guard-failures")
	codec := conflicttest.NewCodec("workbook")
	owner := appsupport.NewTaskDecisionOwner(harness.DB, codec)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "task-decision-task-guards@example.test", "TaskDecision Task Guards", "TaskDecisionTaskGuards1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-task-decision-task-guard-incident", "IR-TASK-DECISION-TASK-GUARDS", "Workbook inspector task and decision workflow task guard failures")

	beforeCreatedAt := testTime(-time.Hour)
	for _, tc := range []struct {
		name   string
		values map[string]tasksdecisions.FieldValue
	}{
		{
			name: "blocked-without-reason-create",
			values: map[string]tasksdecisions.FieldValue{
				"task.title":     {Text: stringPtr("Blocked without reason")},
				"task.task_kind": {Text: stringPtr("request")},
				"task.status":    {Text: stringPtr("blocked")},
			},
		},
		{
			name: "non-blocked-with-reason-create",
			values: map[string]tasksdecisions.FieldValue{
				"task.title":          {Text: stringPtr("Open with blocked reason")},
				"task.task_kind":      {Text: stringPtr("request")},
				"task.blocked_reason": {Text: stringPtr("Reason is only legal while blocked")},
			},
		},
		{
			name: "non-done-with-completed-at-create",
			values: map[string]tasksdecisions.FieldValue{
				"task.title":        {Text: stringPtr("Open with completion time")},
				"task.task_kind":    {Text: stringPtr("request")},
				"task.completed_at": {Timestamp: &beforeCreatedAt},
			},
		},
		{
			name: "done-before-created-create",
			values: map[string]tasksdecisions.FieldValue{
				"task.title":        {Text: stringPtr("Done before created")},
				"task.task_kind":    {Text: stringPtr("request")},
				"task.status":       {Text: stringPtr("done")},
				"task.completed_at": {Timestamp: &beforeCreatedAt},
			},
		},
	} {
		beforeRecords := countRecords(t, harness.DB, incident.ID)
		request := tasksdecisions.CreateRequest{
			ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID,
			ClientTxnID:  "txn-workbook_interaction-task-decision-task-guard-" + tc.name,
			Values:       tc.values,
		}
		_, err := createTaskDecision(owner, actor, incident.ID, request, "req-workbook_interaction-task-decision-task-guard-"+tc.name, testTime(0))
		requireLifecycle(t, err)
		if got := countRecords(t, harness.DB, incident.ID); got != beforeRecords {
			t.Fatalf("%s wrote partial records: got %d want %d", tc.name, got, beforeRecords)
		}
	}

	for _, tc := range []struct {
		name    string
		changes []tasksdecisions.PatchChange
	}{
		{
			name: "blocked-without-reason-patch",
			changes: []tasksdecisions.PatchChange{
				valueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("blocked")}),
			},
		},
		{
			name: "non-blocked-with-reason-patch",
			changes: []tasksdecisions.PatchChange{
				valueChange("task.blocked_reason", tasksdecisions.FieldValue{Text: stringPtr("Reason is only legal while blocked")}),
			},
		},
		{
			name: "non-done-with-completed-at-patch",
			changes: []tasksdecisions.PatchChange{
				valueChange("task.completed_at", tasksdecisions.FieldValue{Timestamp: &beforeCreatedAt}),
			},
		},
		{
			name: "done-before-created-patch",
			changes: []tasksdecisions.PatchChange{
				valueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("done")}),
				valueChange("task.completed_at", tasksdecisions.FieldValue{Timestamp: &beforeCreatedAt}),
			},
		},
	} {
		task := mustCreateTask(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-guard-base-"+tc.name, map[string]tasksdecisions.FieldValue{
			"task.title":     {Text: stringPtr("Guard base " + tc.name)},
			"task.task_kind": {Text: stringPtr("request")},
		}, nil)
		before := taskSnapshot(t, harness.DB, task.RecordID)
		_, err := patchRecord(owner, actor, task.RecordID, tasksdecisions.TaskRequestsViewSchemaID, before.RowVersion, "txn-workbook_interaction-task-decision-task-guard-"+tc.name, tc.changes...)
		requireLifecycle(t, err)
		requireTaskSnapshot(t, taskSnapshot(t, harness.DB, task.RecordID), before, tc.name)
		if got := countReferenceLinks(t, harness.DB, task.RecordID, "task.linked_record_ids"); got != 0 {
			t.Fatalf("%s wrote partial task links: got %d want 0", tc.name, got)
		}
	}
}
