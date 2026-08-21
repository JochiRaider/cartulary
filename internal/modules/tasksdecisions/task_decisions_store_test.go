package tasksdecisions_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
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
	_, err := createTaskDecision(owner, actor, incident.ID, minimumRequest, "req-workbook_interaction-task-decision-task-minimum-fail", Time(0))
	requireMutationValidation(t, err, "task.title", "missing_required_field")
	if got := countRecords(t, harness.DB, incident.ID); got != beforeRecords {
		t.Fatalf("rejected minimum task create wrote records: got %d want %d", got, beforeRecords)
	}

	decision := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-decision", "approved", "Task-linked decision")
	support := mustCreateEvidence(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-task-decision-task-support", "Task support record")
	dueAt := Time(24 * time.Hour)
	task := mustCreateTask(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-create", map[string]tasksdecisions.FieldValue{
		"task.title":               {Text: stringPtr("Collect endpoint logs")},
		"task.task_kind":           {Text: stringPtr("collection")},
		"task.due_at":              {Timestamp: &dueAt},
		"task.workstream":          {Text: stringPtr("forensics")},
		"task.external_ticket_ref": {Text: stringPtr("SOC-123")},
		"task.decision_record_id":  {UUID: &decision},
	}, map[string]tasksdecisions.CollectionActionPayload{
		"task.linked_record_ids": Collection(addOptionalSurfaceRecordRef(support)),
	})
	taskID := task.RecordID
	requireCellValue(t, task.Row, "task.priority", "normal")
	requireCellValue(t, task.Row, "task.decision_record_id", decision.String())
	if got := countTaskDecisionLinks(t, harness.DB, taskID); got != 1 {
		t.Fatalf("task decision direct link count: got %d want 1", got)
	}
	requireManualReferenceLink(t, harness.DB, taskID, decision, "task.decision_record_id", "references_record")
	requireManualReferenceLink(t, harness.DB, taskID, support, "task.linked_record_ids", "references_record")

	taskRows, err := store.QueryRows(ctx, incident.ID, tasksdecisions.TaskRequestsViewSchemaID, viewschema.QueryMeta{
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
	queueDueAt := Time(48 * time.Hour)
	queueTask = mustPatch(t, owner, actor, queueTaskID, tasksdecisions.TaskRequestsViewSchemaID, 1, "txn-workbook_interaction-task-decision-task-queue-edit",
		ValueChange("task.priority", tasksdecisions.FieldValue{Text: stringPtr("high")}),
		ValueChange("task.due_at", tasksdecisions.FieldValue{Timestamp: &queueDueAt}))
	requireCellValue(t, queueTask.Row, "task.priority", "high")
	requireCellNonEmpty(t, queueTask.Row, "task.due_at")
	priorityRows, err := store.QueryRows(ctx, incident.ID, tasksdecisions.TaskRequestsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "task.priority", Op: "eq", Arg: map[string]any{"value": "high"}},
			{FieldKey: "task.due_at", Op: "range", Arg: map[string]any{"lte": queueDueAt.Format(time.RFC3339Nano)}},
		},
		Sort: []viewschema.SortEntry{{FieldKey: "task.priority", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query task priority and due filters: %v", err)
	}
	if !RowsContain(priorityRows, queueTaskID) {
		t.Fatalf("priority/due projection query missing queue task: %#v", priorityRows)
	}
	urgentTask := mustCreateTask(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-queue-urgent", map[string]tasksdecisions.FieldValue{
		"task.title":     {Text: stringPtr("Queue priority urgent")},
		"task.task_kind": {Text: stringPtr("follow_up")},
		"task.priority":  {Text: stringPtr("urgent")},
	}, nil)
	sortedRows, err := store.QueryRows(ctx, incident.ID, tasksdecisions.TaskRequestsViewSchemaID, viewschema.QueryMeta{
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
		ValueChange("task.due_at", tasksdecisions.FieldValue{}))
	requireCellValue(t, queueTask.Row, "task.due_at", nil)
	clearedDueRows, err := store.QueryRows(ctx, incident.ID, tasksdecisions.TaskRequestsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "task.priority", Op: "eq", Arg: map[string]any{"value": "high"}},
			{FieldKey: "task.due_at", Op: "range", Arg: map[string]any{"lte": queueDueAt.Format(time.RFC3339Nano)}},
		},
		Sort: []viewschema.SortEntry{{FieldKey: "task.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query task due after clear: %v", err)
	}
	if RowsContain(clearedDueRows, queueTaskID) {
		t.Fatalf("cleared due date still matched due range: %#v", clearedDueRows)
	}
	beforeInvalidPriorityVersion := RecordVersion(t, harness.DB, queueTaskID)
	_, err = Patch(owner, actor, queueTaskID, tasksdecisions.TaskRequestsViewSchemaID, 3, "txn-workbook_interaction-task-decision-task-invalid-priority",
		ValueChange("task.priority", tasksdecisions.FieldValue{Text: stringPtr("critical")}))
	requireMutationValidation(t, err, "task.priority", "invalid_value")
	if got := RecordVersion(t, harness.DB, queueTaskID); got != beforeInvalidPriorityVersion {
		t.Fatalf("invalid priority changed row version: got %d want %d", got, beforeInvalidPriorityVersion)
	}

	task = mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 1, "txn-workbook_interaction-task-decision-task-in-progress",
		ValueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("in_progress")}))
	task = mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 2, "txn-workbook_interaction-task-decision-task-blocked",
		ValueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("blocked")}),
		ValueChange("task.blocked_reason", tasksdecisions.FieldValue{Text: stringPtr("Waiting for host owner")}))
	requireCellValue(t, task.Row, "task.status", "blocked")
	requireCellValue(t, task.Row, "task.blocked_reason", "Waiting for host owner")
	task = mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 3, "txn-workbook_interaction-task-decision-task-unblocked",
		ValueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("in_progress")}))
	requireCellValue(t, task.Row, "task.blocked_reason", nil)
	task = mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 4, "txn-workbook_interaction-task-decision-task-done",
		ValueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("done")}))
	requireCellNonEmpty(t, task.Row, "task.completed_at")
	task = mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 5, "txn-workbook_interaction-task-decision-task-reopen",
		ValueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("open")}))
	requireCellValue(t, task.Row, "task.completed_at", nil)

	cleared := mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 6, "txn-workbook_interaction-task-decision-task-clear-decision",
		ValueChange("task.decision_record_id", tasksdecisions.FieldValue{}))
	requireCellValue(t, cleared.Row, "task.decision_record_id", nil)
	if got := countTaskDecisionLinks(t, harness.DB, taskID); got != 0 {
		t.Fatalf("task decision direct link after clear: got %d want 0", got)
	}
	reset := mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 7, "txn-workbook_interaction-task-decision-task-reset-decision",
		ValueChange("task.decision_record_id", tasksdecisions.FieldValue{UUID: &decision}))
	requireCellValue(t, reset.Row, "task.decision_record_id", decision.String())
	if got := countTaskDecisionLinks(t, harness.DB, taskID); got != 1 {
		t.Fatalf("task decision direct link after reset: got %d want 1", got)
	}
	requireManualReferenceLink(t, harness.DB, taskID, decision, "task.decision_record_id", "references_record")
	otherIncident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-task-decision-task-other-incident", "IR-TASK-DECISION-TASK-OTHER", "Workbook inspector task and decision workflow task requests other")
	foreignDecision := mustCreateDecision(t, owner, actor, otherIncident.ID, "txn-workbook_interaction-task-decision-task-foreign-decision", "approved", "Foreign decision")
	deletedDecision := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-deleted-decision", "approved", "Deleted decision")
	if _, err := harness.DB.Exec(ctx, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, deletedDecision, Time(20*time.Minute), actor.ID); err != nil {
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
		_, err = Patch(owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 8, "txn-workbook_interaction-task-decision-task-invalid-decision-"+invalid.name,
			ValueChange("task.decision_record_id", tasksdecisions.FieldValue{UUID: &invalid.id}))
		requireMutationValidation(t, err, "task.decision_record_id", "invalid_value")
		if got := countTaskDecisionLinks(t, harness.DB, taskID); got != 1 {
			t.Fatalf("invalid %s decision ref changed links: got %d want 1", invalid.name, got)
		}
	}

	done := mustPatch(t, owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 8, "txn-workbook_interaction-task-decision-task-done-again",
		ValueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("done")}))
	requireCellValue(t, done.Row, "task.status", "done")
	_, err = Patch(owner, actor, taskID, tasksdecisions.TaskRequestsViewSchemaID, 9, "txn-workbook_interaction-task-decision-task-done-canceled",
		ValueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("canceled")}))
	requireLifecycle(t, err)

	canceled := mustCreateTask(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-canceled", map[string]tasksdecisions.FieldValue{
		"task.title":     {Text: stringPtr("Canceled task")},
		"task.task_kind": {Text: stringPtr("request")},
		"task.status":    {Text: stringPtr("canceled")},
	}, nil)
	_, err = Patch(owner, actor, canceled.RecordID, tasksdecisions.TaskRequestsViewSchemaID, 1, "txn-workbook_interaction-task-decision-task-canceled-done",
		ValueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("done")}))
	requireLifecycle(t, err)

	ownerless := mustCreateTask(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-ownerless", map[string]tasksdecisions.FieldValue{
		"task.title":     {Text: stringPtr("Owner guard")},
		"task.task_kind": {Text: stringPtr("request")},
	}, nil)
	_, err = Patch(owner, actor, ownerless.RecordID, tasksdecisions.TaskRequestsViewSchemaID, 1, "txn-workbook_interaction-task-decision-task-ownerless-open",
		ValueChange("task.owner_user_id", tasksdecisions.FieldValue{}))
	requireLifecycle(t, err)
}

func TestTaskLifecycleGuardFailures_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-task-decision-task-guard-failures")
	codec := conflicttest.NewCodec("workbook")
	owner := appsupport.NewTaskDecisionOwner(harness.DB, codec)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "task-decision-task-guards@example.test", "TaskDecision Task Guards", "TaskDecisionTaskGuards1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-task-decision-task-guard-incident", "IR-TASK-DECISION-TASK-GUARDS", "Workbook inspector task and decision workflow task guard failures")

	beforeCreatedAt := Time(-time.Hour)
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
		_, err := createTaskDecision(owner, actor, incident.ID, request, "req-workbook_interaction-task-decision-task-guard-"+tc.name, Time(0))
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
				ValueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("blocked")}),
			},
		},
		{
			name: "non-blocked-with-reason-patch",
			changes: []tasksdecisions.PatchChange{
				ValueChange("task.blocked_reason", tasksdecisions.FieldValue{Text: stringPtr("Reason is only legal while blocked")}),
			},
		},
		{
			name: "non-done-with-completed-at-patch",
			changes: []tasksdecisions.PatchChange{
				ValueChange("task.completed_at", tasksdecisions.FieldValue{Timestamp: &beforeCreatedAt}),
			},
		},
		{
			name: "done-before-created-patch",
			changes: []tasksdecisions.PatchChange{
				ValueChange("task.status", tasksdecisions.FieldValue{Text: stringPtr("done")}),
				ValueChange("task.completed_at", tasksdecisions.FieldValue{Timestamp: &beforeCreatedAt}),
			},
		},
	} {
		task := mustCreateTask(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-task-guard-base-"+tc.name, map[string]tasksdecisions.FieldValue{
			"task.title":     {Text: stringPtr("Guard base " + tc.name)},
			"task.task_kind": {Text: stringPtr("request")},
		}, nil)
		before := TaskSnapshot(t, harness.DB, task.RecordID)
		_, err := Patch(owner, actor, task.RecordID, tasksdecisions.TaskRequestsViewSchemaID, before.RowVersion, "txn-workbook_interaction-task-decision-task-guard-"+tc.name, tc.changes...)
		requireLifecycle(t, err)
		requireTaskSnapshot(t, TaskSnapshot(t, harness.DB, task.RecordID), before, tc.name)
		if got := countReferenceLinks(t, harness.DB, task.RecordID, "task.linked_record_ids"); got != 0 {
			t.Fatalf("%s wrote partial task links: got %d want 0", tc.name, got)
		}
	}
}

func TestDecisionLifecycleSupersessionAndConsistency_Unit(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "workbook_interaction-task-decision-decisions")
	codec := conflicttest.NewCodec("workbook")
	store := appsupport.NewWorkbookCatalog(harness.DB, codec)
	owner := appsupport.NewTaskDecisionOwner(harness.DB, codec)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "task-decision-decision@example.test", "TaskDecision Decision", "TaskDecisionDecision1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-task-decision-decision-incident", "IR-TASK-DECISION-DECISION", "Workbook inspector task and decision workflow decisions")

	beforeRecords := countRecords(t, harness.DB, incident.ID)
	minimumRequest := tasksdecisions.CreateRequest{
		ViewSchemaID: tasksdecisions.DecisionsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-task-decision-decision-minimum-fail",
		Values: map[string]tasksdecisions.FieldValue{
			"decision.decision_type": {Text: stringPtr("scope")},
		},
	}
	_, err := createTaskDecision(owner, actor, incident.ID, minimumRequest, "req-workbook_interaction-task-decision-decision-minimum-fail", Time(0))
	requireMutationValidation(t, err, "decision.summary", "missing_required_field")
	if got := countRecords(t, harness.DB, incident.ID); got != beforeRecords {
		t.Fatalf("rejected minimum decision create wrote records: got %d want %d", got, beforeRecords)
	}
	supersededRequest := tasksdecisions.CreateRequest{
		ViewSchemaID: tasksdecisions.DecisionsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-task-decision-decision-create-superseded",
		Values: map[string]tasksdecisions.FieldValue{
			"decision.summary":       {Text: stringPtr("Bad superseded create")},
			"decision.decision_type": {Text: stringPtr("scope")},
			"decision.rationale":     {Text: stringPtr("Superseded must be explicit.")},
			"decision.status":        {Text: stringPtr("superseded")},
		},
	}
	_, err = createTaskDecision(owner, actor, incident.ID, supersededRequest, "req-workbook_interaction-task-decision-decision-create-superseded", Time(0))
	requireLifecycle(t, err)

	support := mustCreateEvidence(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-support", "Decision support record")
	affectedOne := mustCreateEvidence(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-affected-one", "Decision affected record one")
	affectedTwo := mustCreateEvidence(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-affected-two", "Decision affected record two")
	relationshipDecision := mustCreateDecisionWithCollections(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-relationships", "proposed", "Relationship decision", map[string]tasksdecisions.CollectionActionPayload{
		"decision.support_refs":        Collection(addOptionalSurfaceRecordRef(support)),
		"decision.affected_record_ids": Collection(addOptionalSurfaceRecordRef(affectedOne), addOptionalSurfaceRecordRef(affectedOne)),
	})
	relationshipRow := relationshipDecision.Row
	requireCollectionItemCount(t, relationshipRow, "decision.support_refs", 1)
	requireCollectionItemCount(t, relationshipRow, "decision.affected_record_ids", 1)
	requireCellNumericValue(t, relationshipRow, "decision.affected_record_count", 1)
	requireManualReferenceLink(t, harness.DB, relationshipDecision.RecordID, support, "decision.support_refs", "supported_by")
	requireManualReferenceLink(t, harness.DB, relationshipDecision.RecordID, affectedOne, "decision.affected_record_ids", "references_record")

	relationshipDecision = mustPatch(t, owner, actor, relationshipDecision.RecordID, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-affected-add",
		CollectionChange("decision.affected_record_ids", Collection(addOptionalSurfaceRecordRef(affectedTwo))))
	relationshipRow = relationshipDecision.Row
	requireCollectionItemCount(t, relationshipRow, "decision.affected_record_ids", 2)
	requireCellNumericValue(t, relationshipRow, "decision.affected_record_count", 2)
	requireManualReferenceLink(t, harness.DB, relationshipDecision.RecordID, affectedTwo, "decision.affected_record_ids", "references_record")

	relationshipDecision = mustPatch(t, owner, actor, relationshipDecision.RecordID, tasksdecisions.DecisionsViewSchemaID, 2, "txn-workbook_interaction-task-decision-decision-affected-remove",
		CollectionChange("decision.affected_record_ids", Collection(removeRecordRef(affectedOne))))
	relationshipRow = relationshipDecision.Row
	requireCollectionItemCount(t, relationshipRow, "decision.affected_record_ids", 1)
	requireCellNumericValue(t, relationshipRow, "decision.affected_record_count", 1)
	if got := countReferenceLinks(t, harness.DB, relationshipDecision.RecordID, "decision.affected_record_ids"); got != 1 {
		t.Fatalf("decision affected record links after remove: got %d want 1", got)
	}

	target := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-target", "proposed", "Target decision")
	source := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-source", "approved", "Superseding decision")
	executed := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-executed", "approved", "Executed decision")
	executedRow := mustPatch(t, owner, actor, executed, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-approved-executed",
		ValueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr("executed")}))
	requireCellValue(t, executedRow.Row, "decision.status", "executed")

	_, err = Patch(owner, actor, target, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-direct-superseded",
		ValueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr("superseded")}))
	requireLifecycle(t, err)
	rejected := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-rejected", "rejected", "Rejected decision")
	_, err = Patch(owner, actor, rejected, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-rejected-proposed",
		ValueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr("proposed")}))
	requireLifecycle(t, err)
	_, err = Patch(owner, actor, executed, tasksdecisions.DecisionsViewSchemaID, 2, "txn-workbook_interaction-task-decision-decision-executed-approved",
		ValueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr("approved")}))
	requireLifecycle(t, err)

	request := tasksdecisions.SupersedeRequest{
		BaseRowVersion:      1,
		ClientTxnID:         "txn-workbook_interaction-task-decision-decision-supersede",
		Reason:              "Supersede with better containment rationale.",
		ReplacementRecordID: &source,
	}
	result, err := supersedeDecision(ctx, owner, actor, target, request, "req-workbook_interaction-task-decision-decision-supersede", Time(time.Hour))
	if err != nil {
		t.Fatalf("supersede proposed target: %v", err)
	}
	if result.ViewSchemaID != tasksdecisions.DecisionsViewSchemaID || result.Facts.TargetStatus != "superseded" {
		t.Fatalf("unexpected decision supersede result: %#v", result)
	}
	if len(result.AdditionalRecordChanges) != 2 {
		t.Fatalf("expected two changed decision rows, got %d", len(result.AdditionalRecordChanges))
	}
	requireDecisionSupersessionChangeSetEffects(t, harness.DB, result.ChangeSetID)
	if got := countSupersedesLinks(t, harness.DB, source, target); got != 1 {
		t.Fatalf("decision supersedes link count: got %d want 1", got)
	}
	decisionRows, err := store.QueryRows(ctx, incident.ID, tasksdecisions.DecisionsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "decision.is_superseded", Op: "eq", Arg: map[string]any{"value": true}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "decision.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query decision superseded projection: %v", err)
	}
	if !RowsContain(decisionRows, target) {
		t.Fatalf("superseded target missing from projection rows: %#v", decisionRows)
	}
	sourceRows, err := store.QueryRows(ctx, incident.ID, tasksdecisions.DecisionsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "decision.supersedes_record_id", Op: "eq", Arg: map[string]any{"value": target.String()}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "decision.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query decision supersedes projection: %v", err)
	}
	if len(sourceRows) != 1 || sourceRows[0]["record_id"] != source.String() {
		t.Fatalf("superseding decision missing from projection rows: %#v", sourceRows)
	}
	afterSupersessionEffects := decisionSupersessionIncidentEffects(t, harness.DB, incident.ID)
	replay, err := supersedeDecision(ctx, owner, actor, target, request, "req-workbook_interaction-task-decision-decision-supersede-replay", Time(2*time.Hour))
	if err != nil {
		t.Fatalf("replay decision supersede: %v", err)
	}
	if !replay.Replayed {
		t.Fatalf("expected idempotent replay")
	}
	if got := countSupersedesLinks(t, harness.DB, source, target); got != 1 {
		t.Fatalf("decision supersedes link count after replay: got %d want 1", got)
	}
	if got := decisionSupersessionIncidentEffects(t, harness.DB, incident.ID); got != afterSupersessionEffects {
		t.Fatalf("decision supersede replay changed durable effects: before=%+v after=%+v", afterSupersessionEffects, got)
	}
	rejectedRequest := request
	rejectedRequest.BaseRowVersion = result.RowVersion
	rejectedRequest.ClientTxnID = "txn-workbook_interaction-task-decision-decision-supersede-rejected"
	_, err = supersedeDecision(ctx, owner, actor, target, rejectedRequest, "req-workbook_interaction-task-decision-decision-supersede-rejected", Time(2*time.Hour))
	requireLifecycle(t, err)
	if got := decisionSupersessionIncidentEffects(t, harness.DB, incident.ID); got != afterSupersessionEffects {
		t.Fatalf("rejected decision supersede changed durable effects: before=%+v after=%+v", afterSupersessionEffects, got)
	}

	executedTarget := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-executed-target", "executed", "Executed target")
	executedSource := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-executed-source", "approved", "Executed target replacement")
	executedRequest := tasksdecisions.SupersedeRequest{
		BaseRowVersion:      1,
		ClientTxnID:         "txn-workbook_interaction-task-decision-decision-supersede-executed",
		Reason:              "Supersede executed decision.",
		ReplacementRecordID: &executedSource,
	}
	if _, err := supersedeDecision(ctx, owner, actor, executedTarget, executedRequest, "req-workbook_interaction-task-decision-decision-supersede-executed", Time(3*time.Hour)); err != nil {
		t.Fatalf("supersede executed target: %v", err)
	}
	row := QueryOne(t, store, incident.ID, tasksdecisions.DecisionsViewSchemaID, "decision.is_superseded", true, executedTarget)
	requireCellValue(t, row, "decision.status", "executed")
	requireCellValue(t, row, "decision.is_superseded", true)

	badSource := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-bad-source", "proposed", "Inconsistent source")
	badTarget := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-bad-target", "proposed", "Inconsistent target")
	if _, err := harness.DB.Exec(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'supersedes', NULL, 'manual', NULL, $4, $4, $5, $5)
`, incident.ID, badSource, badTarget, actor.ID, Time(4*time.Hour)); err != nil {
		t.Fatalf("seed inconsistent supersedes link: %v", err)
	}
	_, err = Patch(owner, actor, badSource, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-inconsistent-fail",
		ValueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr("approved")}))
	requireLifecycle(t, err)
	beforeInconsistentVersion := RecordVersion(t, harness.DB, badSource)
	_, err = Patch(owner, actor, badSource, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-inconsistent-rationale-fail",
		ValueChange("decision.rationale", tasksdecisions.FieldValue{Text: stringPtr("Ordinary scalar edits must fail closed.")}))
	requireLifecycle(t, err)
	if got := RecordVersion(t, harness.DB, badSource); got != beforeInconsistentVersion {
		t.Fatalf("inconsistent decision scalar patch changed row version: got %d want %d", got, beforeInconsistentVersion)
	}
	beforeSupportLinks := countReferenceLinks(t, harness.DB, badSource, "decision.support_refs")
	_, err = Patch(owner, actor, badSource, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-inconsistent-support-fail",
		CollectionChange("decision.support_refs", Collection(addOptionalSurfaceRecordRef(support))))
	requireLifecycle(t, err)
	if got := countReferenceLinks(t, harness.DB, badSource, "decision.support_refs"); got != beforeSupportLinks {
		t.Fatalf("inconsistent decision support patch changed links: got %d want %d", got, beforeSupportLinks)
	}
	if got := RecordVersion(t, harness.DB, badSource); got != beforeInconsistentVersion {
		t.Fatalf("inconsistent decision support patch changed row version: got %d want %d", got, beforeInconsistentVersion)
	}
	beforeAffectedLinks := countReferenceLinks(t, harness.DB, badSource, "decision.affected_record_ids")
	_, err = Patch(owner, actor, badSource, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-inconsistent-affected-fail",
		CollectionChange("decision.affected_record_ids", Collection(addOptionalSurfaceRecordRef(affectedTwo))))
	requireLifecycle(t, err)
	if got := countReferenceLinks(t, harness.DB, badSource, "decision.affected_record_ids"); got != beforeAffectedLinks {
		t.Fatalf("inconsistent decision affected patch changed links: got %d want %d", got, beforeAffectedLinks)
	}
	if got := RecordVersion(t, harness.DB, badSource); got != beforeInconsistentVersion {
		t.Fatalf("inconsistent decision affected patch changed row version: got %d want %d", got, beforeInconsistentVersion)
	}
}

type decisionSupersessionEffects struct {
	changeSets int
	revisions  int
	intents    int
}

func requireDecisionSupersessionChangeSetEffects(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, changeSetID uuid.UUID) {
	t.Helper()
	var effects decisionSupersessionEffects
	if err := db.QueryRow(context.Background(), `
SELECT
    (SELECT count(*) FROM change_sets WHERE change_set_id = $1),
    (SELECT count(*) FROM record_revisions WHERE change_set_id = $1),
    (SELECT count(*) FROM collaboration_event_intents WHERE source_change_set_id = $1)
`, changeSetID).Scan(&effects.changeSets, &effects.revisions, &effects.intents); err != nil {
		t.Fatalf("query decision supersession change-set effects: %v", err)
	}
	if effects != (decisionSupersessionEffects{changeSets: 1, revisions: 2, intents: 2}) {
		t.Fatalf("decision supersession change-set effects = %+v; want one change set, two revisions, and two intents", effects)
	}
}

func decisionSupersessionIncidentEffects(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, incidentID uuid.UUID) decisionSupersessionEffects {
	t.Helper()
	var effects decisionSupersessionEffects
	if err := db.QueryRow(context.Background(), `
SELECT
    (SELECT count(*) FROM change_sets WHERE incident_id = $1),
    (SELECT count(*) FROM record_revisions revision
       JOIN records record USING (record_id)
      WHERE record.incident_id = $1),
    (SELECT count(*) FROM collaboration_event_intents WHERE incident_id = $1)
`, incidentID).Scan(&effects.changeSets, &effects.revisions, &effects.intents); err != nil {
		t.Fatalf("query decision supersession incident effects: %v", err)
	}
	return effects
}

func TestSupersedeDecisionRejectsInconsistentSourceOrTarget_Unit(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "workbook_interaction-task-decision-decision-supersede-inconsistent")
	codec := conflicttest.NewCodec("workbook")
	owner := appsupport.NewTaskDecisionOwner(harness.DB, codec)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "task-decision-decision-supersede-inconsistent@example.test", "TaskDecision Decision Supersede Inconsistent", "TaskDecisionDecisionSupersede1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-task-decision-decision-supersede-inconsistent-incident", "IR-TASK-DECISION-DECISION-SUPERSEDE-INCONSISTENT", "Workbook inspector task and decision workflow decision supersede inconsistent")

	inconsistentSource := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-inconsistent-source", "proposed", "Inconsistent source")
	sourceExistingTarget := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-inconsistent-source-existing-target", "proposed", "Existing target")
	validTarget := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-source-fail-target", "proposed", "Source fail target")
	insertSupersedesLink(t, harness.DB, incident.ID, inconsistentSource, sourceExistingTarget, actor.ID, Time(time.Hour))

	sourceBefore := DecisionSnapshot(t, harness.DB, inconsistentSource)
	sourceExistingTargetBefore := DecisionSnapshot(t, harness.DB, sourceExistingTarget)
	validTargetBefore := DecisionSnapshot(t, harness.DB, validTarget)
	sourceRequest := tasksdecisions.SupersedeRequest{
		BaseRowVersion:      validTargetBefore.RowVersion,
		ClientTxnID:         "txn-workbook_interaction-task-decision-decision-inconsistent-source-route",
		Reason:              "Attempt explicit supersession with inconsistent source.",
		ReplacementRecordID: &inconsistentSource,
	}
	_, err := supersedeDecision(ctx, owner, actor, validTarget, sourceRequest, "req-workbook_interaction-task-decision-decision-inconsistent-source-route", Time(2*time.Hour))
	requireLifecycle(t, err)
	requireDecisionSnapshot(t, DecisionSnapshot(t, harness.DB, inconsistentSource), sourceBefore, "inconsistent source supersede")
	requireDecisionSnapshot(t, DecisionSnapshot(t, harness.DB, sourceExistingTarget), sourceExistingTargetBefore, "inconsistent source existing target")
	requireDecisionSnapshot(t, DecisionSnapshot(t, harness.DB, validTarget), validTargetBefore, "valid target rejected by inconsistent source")
	if got := countSupersedesLinks(t, harness.DB, inconsistentSource, validTarget); got != 0 {
		t.Fatalf("inconsistent source supersede wrote link to attempted target: got %d want 0", got)
	}
	if got := countReferenceLinks(t, harness.DB, inconsistentSource, "decision.support_refs"); got != 0 {
		t.Fatalf("inconsistent source supersede wrote support links: got %d want 0", got)
	}
	if got := countReferenceLinks(t, harness.DB, validTarget, "decision.affected_record_ids"); got != 0 {
		t.Fatalf("inconsistent source supersede wrote affected links: got %d want 0", got)
	}

	validSource := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-target-fail-source", "approved", "Valid source")
	inconsistentTarget := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-inconsistent-target", "proposed", "Inconsistent target")
	targetExistingSource := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-inconsistent-target-existing-source", "approved", "Existing superseding source")
	insertSupersedesLink(t, harness.DB, incident.ID, targetExistingSource, inconsistentTarget, actor.ID, Time(3*time.Hour))

	validSourceBefore := DecisionSnapshot(t, harness.DB, validSource)
	inconsistentTargetBefore := DecisionSnapshot(t, harness.DB, inconsistentTarget)
	targetExistingSourceBefore := DecisionSnapshot(t, harness.DB, targetExistingSource)
	targetRequest := tasksdecisions.SupersedeRequest{
		BaseRowVersion:      inconsistentTargetBefore.RowVersion,
		ClientTxnID:         "txn-workbook_interaction-task-decision-decision-inconsistent-target-route",
		Reason:              "Attempt explicit supersession against inconsistent target.",
		ReplacementRecordID: &validSource,
	}
	_, err = supersedeDecision(ctx, owner, actor, inconsistentTarget, targetRequest, "req-workbook_interaction-task-decision-decision-inconsistent-target-route", Time(4*time.Hour))
	requireLifecycle(t, err)
	requireDecisionSnapshot(t, DecisionSnapshot(t, harness.DB, validSource), validSourceBefore, "valid source rejected by inconsistent target")
	requireDecisionSnapshot(t, DecisionSnapshot(t, harness.DB, inconsistentTarget), inconsistentTargetBefore, "inconsistent target supersede")
	requireDecisionSnapshot(t, DecisionSnapshot(t, harness.DB, targetExistingSource), targetExistingSourceBefore, "inconsistent target existing source")
	if got := countSupersedesLinks(t, harness.DB, validSource, inconsistentTarget); got != 0 {
		t.Fatalf("inconsistent target supersede wrote attempted link: got %d want 0", got)
	}
	if got := countReferenceLinks(t, harness.DB, validSource, "decision.support_refs"); got != 0 {
		t.Fatalf("inconsistent target supersede wrote support links: got %d want 0", got)
	}
	if got := countReferenceLinks(t, harness.DB, inconsistentTarget, "decision.affected_record_ids"); got != 0 {
		t.Fatalf("inconsistent target supersede wrote affected links: got %d want 0", got)
	}
}

func TestDecisionTerminalTransitionMatrix_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-task-decision-decision-terminal-matrix")
	codec := conflicttest.NewCodec("workbook")
	store := appsupport.NewWorkbookCatalog(harness.DB, codec)
	owner := appsupport.NewTaskDecisionOwner(harness.DB, codec)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "task-decision-decision-terminal@example.test", "TaskDecision Decision Terminal", "TaskDecisionDecisionTerminal1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-task-decision-decision-terminal-incident", "IR-TASK-DECISION-DECISION-TERMINAL", "Workbook inspector task and decision workflow decision terminal matrix")

	for _, from := range []string{"rejected", "executed", "superseded"} {
		for _, to := range []string{"proposed", "approved", "rejected", "executed", "superseded"} {
			name := from + "-to-" + to
			decisionID := mustCreateDecisionInTerminalState(t, store, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-terminal-base-"+name, from)
			before := DecisionSnapshot(t, harness.DB, decisionID)
			changes := []tasksdecisions.PatchChange{
				ValueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr(to)}),
			}
			if from == to {
				changes = append(changes, ValueChange("decision.rationale", tasksdecisions.FieldValue{Text: stringPtr("Idempotent in-state terminal write remains ordinary scalar work.")}))
			}
			_, err := Patch(owner, actor, decisionID, tasksdecisions.DecisionsViewSchemaID, before.RowVersion, "txn-workbook_interaction-task-decision-decision-terminal-"+name, changes...)
			if from == to && from != "superseded" {
				if err != nil {
					t.Fatalf("%s should allow in-state terminal write, got %v", name, err)
				}
				after := DecisionSnapshot(t, harness.DB, decisionID)
				if after.Status != from || after.Rationale != "Idempotent in-state terminal write remains ordinary scalar work." || after.RowVersion <= before.RowVersion {
					t.Fatalf("%s unexpected in-state result: before=%#v after=%#v", name, before, after)
				}
				continue
			}
			requireLifecycle(t, err)
			requireDecisionSnapshot(t, DecisionSnapshot(t, harness.DB, decisionID), before, name)
		}
	}
}

func mustCreateDecision(t testing.TB, owner *tasksdecisions.MutationFacade, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string, summary string) uuid.UUID {
	t.Helper()
	return mustCreateDecisionWithCollections(t, owner, actor, incidentID, clientTxnID, status, summary, nil).RecordID
}

func mustCreateEvidence(t testing.TB, pool postgres.DB, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, title string) uuid.UUID {
	t.Helper()
	request := evidence.CreateRequest{
		ViewSchemaID: evidence.ViewSchemaID, ClientTxnID: clientTxnID,
		Values: map[string]evidence.FieldValue{
			"evidence.title": {Text: &title},
		},
	}
	result, err := appsupport.NewEvidenceMutationOwner(pool, conflicttest.NewCodec("workbook")).Create(
		context.Background(),
		evidence.CreateCommand{
			Actor: actor, IncidentID: incidentID, Request: request,
			RequestHash: evidence.CreateRequestHash(request), RequestID: "req-" + clientTxnID,
			RouteKey: "workbook.rows.create", Now: Time(0),
		},
	)
	if err != nil {
		t.Fatalf("create evidence %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func supersedeDecision(
	ctx context.Context,
	owner *tasksdecisions.MutationFacade,
	actor authn.UserRecord,
	targetRecordID uuid.UUID,
	request tasksdecisions.SupersedeRequest,
	requestID string,
	now time.Time,
) (tasksdecisions.SupersedeMutationResult, error) {
	return owner.SupersedeDecision(ctx, tasksdecisions.SupersedeCommand{
		ActorUserID: actor.ID, TargetRecordID: targetRecordID, Request: request,
		RequestHash: tasksdecisions.SupersedeRequestHash(request), RequestID: requestID,
		RouteKey: "workbook.records.supersede", Now: now,
	})
}

func createTaskDecision(
	owner *tasksdecisions.MutationFacade,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	request tasksdecisions.CreateRequest,
	requestID string,
	now time.Time,
) (tasksdecisions.MutationResult, error) {
	return owner.Create(context.Background(), tasksdecisions.CreateCommand{
		ActorUserID: actor.ID, IncidentID: incidentID, Request: request,
		RequestHash: tasksdecisions.CreateRequestHash(request), RequestID: requestID,
		RouteKey: "workbook.rows.create", Now: now,
	})
}

func mustCreateDecisionInTerminalState(t testing.TB, _ *workbook.WorkbookContributionCatalog, owner *tasksdecisions.MutationFacade, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string) uuid.UUID {
	t.Helper()
	if status != "superseded" {
		return mustCreateDecision(t, owner, actor, incidentID, clientTxnID, status, "Terminal "+status)
	}
	target := mustCreateDecision(t, owner, actor, incidentID, clientTxnID+"-target", "proposed", "Superseded target")
	source := mustCreateDecision(t, owner, actor, incidentID, clientTxnID+"-source", "approved", "Superseding source")
	request := tasksdecisions.SupersedeRequest{
		BaseRowVersion:      1,
		ClientTxnID:         clientTxnID + "-supersede",
		Reason:              "Create explicit superseded terminal state.",
		ReplacementRecordID: &source,
	}
	if _, err := supersedeDecision(context.Background(), owner, actor, target, request, "req-"+clientTxnID+"-supersede", Time(time.Hour)); err != nil {
		t.Fatalf("create superseded decision %s: %v", clientTxnID, err)
	}
	return target
}

func mustCreateDecisionWithCollections(t testing.TB, owner *tasksdecisions.MutationFacade, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string, summary string, collections map[string]tasksdecisions.CollectionActionPayload) tasksdecisions.MutationResult {
	t.Helper()
	values := map[string]tasksdecisions.FieldValue{
		"decision.summary":       {Text: &summary},
		"decision.decision_type": {Text: stringPtr("containment")},
		"decision.rationale":     {Text: stringPtr("The decision is needed for coordinated response.")},
	}
	if status != "" {
		values["decision.status"] = tasksdecisions.FieldValue{Text: &status}
	}
	request := tasksdecisions.CreateRequest{
		ViewSchemaID: tasksdecisions.DecisionsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values:       values,
		Collections:  collections,
	}
	result, err := createTaskDecision(owner, actor, incidentID, request, "req-"+clientTxnID, Time(0))
	if err != nil {
		t.Fatalf("create decision %s: %v", clientTxnID, err)
	}
	return result
}

type TaskState struct {
	RowVersion    int64
	Status        string
	BlockedReason sql.NullString
	CompletedAt   sql.NullTime
	OwnerUserID   sql.NullString
}

func TaskSnapshot(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, recordID uuid.UUID) TaskState {
	t.Helper()
	var state TaskState
	if err := db.QueryRow(context.Background(), `
SELECT r.row_version, t.status, t.blocked_reason, t.completed_at, t.owner_user_id::text
  FROM task_requests t
  JOIN records r
    ON r.record_id = t.record_id
 WHERE t.record_id = $1
`, recordID).Scan(&state.RowVersion, &state.Status, &state.BlockedReason, &state.CompletedAt, &state.OwnerUserID); err != nil {
		t.Fatalf("query task snapshot: %v", err)
	}
	return state
}

func requireTaskSnapshot(t testing.TB, got TaskState, want TaskState, context string) {
	t.Helper()
	if got.RowVersion != want.RowVersion ||
		got.Status != want.Status ||
		got.BlockedReason != want.BlockedReason ||
		got.CompletedAt.Valid != want.CompletedAt.Valid ||
		(got.CompletedAt.Valid && !got.CompletedAt.Time.Equal(want.CompletedAt.Time)) ||
		got.OwnerUserID != want.OwnerUserID {
		t.Fatalf("%s changed task snapshot: got %#v want %#v", context, got, want)
	}
}

type DecisionState struct {
	RowVersion         int64
	Status             string
	Rationale          string
	IncomingSupersedes int
	OutgoingSupersedes int
}

func DecisionSnapshot(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, recordID uuid.UUID) DecisionState {
	t.Helper()
	var state DecisionState
	if err := db.QueryRow(context.Background(), `
SELECT r.row_version,
       d.status,
       d.rationale,
       COALESCE(incoming.count, 0)::integer,
       COALESCE(outgoing.count, 0)::integer
  FROM decisions d
  JOIN records r
    ON r.record_id = d.record_id
  LEFT JOIN (
        SELECT dst_record_id, COUNT(*) AS count
          FROM record_links
         WHERE link_type = 'supersedes'
           AND deleted_at IS NULL
         GROUP BY dst_record_id
  ) incoming
    ON incoming.dst_record_id = d.record_id
  LEFT JOIN (
        SELECT src_record_id, COUNT(*) AS count
          FROM record_links
         WHERE link_type = 'supersedes'
           AND deleted_at IS NULL
         GROUP BY src_record_id
  ) outgoing
    ON outgoing.src_record_id = d.record_id
 WHERE d.record_id = $1
`, recordID).Scan(&state.RowVersion, &state.Status, &state.Rationale, &state.IncomingSupersedes, &state.OutgoingSupersedes); err != nil {
		t.Fatalf("query decision snapshot: %v", err)
	}
	return state
}

func requireDecisionSnapshot(t testing.TB, got DecisionState, want DecisionState, context string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s changed decision snapshot: got %#v want %#v", context, got, want)
	}
}

func mustCreateTask(t testing.TB, owner *tasksdecisions.MutationFacade, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, values map[string]tasksdecisions.FieldValue, collections map[string]tasksdecisions.CollectionActionPayload) tasksdecisions.MutationResult {
	t.Helper()
	request := tasksdecisions.CreateRequest{
		ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values:       values,
		Collections:  collections,
	}
	result, err := createTaskDecision(owner, actor, incidentID, request, "req-"+clientTxnID, Time(0))
	if err != nil {
		t.Fatalf("create task %s: %v", clientTxnID, err)
	}
	return result
}

func Patch(owner *tasksdecisions.MutationFacade, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...tasksdecisions.PatchChange) (tasksdecisions.MutationResult, error) {
	request := tasksdecisions.PatchRequest{
		ViewSchemaID:   viewSchemaID,
		BaseRowVersion: baseRowVersion,
		ClientTxnID:    clientTxnID,
		Changes:        changes,
	}
	return owner.Patch(context.Background(), tasksdecisions.PatchCommand{
		ActorUserID: actor.ID, RecordID: recordID, Request: request,
		RequestHash: tasksdecisions.PatchRequestHash(request), RequestID: "req-" + clientTxnID,
		RouteKey: "workbook.records.patch", ConflictRouteKey: "workbook.records.conflicts.resolve",
		Now: Time(30 * time.Minute),
	})
}

func mustPatch(t testing.TB, owner *tasksdecisions.MutationFacade, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...tasksdecisions.PatchChange) tasksdecisions.MutationResult {
	t.Helper()
	result, err := Patch(owner, actor, recordID, viewSchemaID, baseRowVersion, clientTxnID, changes...)
	if err != nil {
		t.Fatalf("patch %s: %v", clientTxnID, err)
	}
	return result
}

func ValueChange(fieldKey string, value tasksdecisions.FieldValue) tasksdecisions.PatchChange {
	return tasksdecisions.PatchChange{FieldKey: fieldKey, Value: &value}
}

func CollectionChange(fieldKey string, value tasksdecisions.CollectionActionPayload) tasksdecisions.PatchChange {
	return tasksdecisions.PatchChange{FieldKey: fieldKey, Collection: &value}
}

func Collection(actions ...tasksdecisions.CollectionAction) tasksdecisions.CollectionActionPayload {
	return tasksdecisions.CollectionActionPayload{Actions: actions}
}

func addOptionalSurfaceRecordRef(recordID uuid.UUID) tasksdecisions.CollectionAction {
	return tasksdecisions.CollectionAction{Op: "add_record_ref", LinkedRecordID: &recordID}
}

func removeRecordRef(recordID uuid.UUID) tasksdecisions.CollectionAction {
	return tasksdecisions.CollectionAction{Op: "remove_record_ref", ItemRef: "record_ref:" + recordID.String()}
}

func countRecords(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, incidentID uuid.UUID) int {
	t.Helper()
	var count int
	if err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM records WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		t.Fatalf("count records: %v", err)
	}
	return count
}

func RecordVersion(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, recordID uuid.UUID) int64 {
	t.Helper()
	var version int64
	if err := db.QueryRow(context.Background(), `SELECT row_version FROM records WHERE record_id = $1`, recordID).Scan(&version); err != nil {
		t.Fatalf("query record version: %v", err)
	}
	return version
}

func countTaskDecisionLinks(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, taskID uuid.UUID) int {
	t.Helper()
	var count int
	if err := db.QueryRow(context.Background(), `
SELECT COUNT(*)
  FROM record_links
 WHERE src_record_id = $1
   AND link_type = 'references_record'
   AND field_key = 'task.decision_record_id'
   AND deleted_at IS NULL
`, taskID).Scan(&count); err != nil {
		t.Fatalf("count task decision links: %v", err)
	}
	return count
}

func countReferenceLinks(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, sourceID uuid.UUID, fieldKey string) int {
	t.Helper()
	var count int
	if err := db.QueryRow(context.Background(), `
SELECT COUNT(*)
  FROM record_links
 WHERE src_record_id = $1
   AND field_key = $2
   AND deleted_at IS NULL
`, sourceID, fieldKey).Scan(&count); err != nil {
		t.Fatalf("count reference links: %v", err)
	}
	return count
}

func requireManualReferenceLink(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, sourceID uuid.UUID, targetID uuid.UUID, fieldKey string, linkType string) {
	t.Helper()
	var provenance string
	var confidence sql.NullInt64
	if err := db.QueryRow(context.Background(), `
SELECT provenance, confidence
  FROM record_links
 WHERE src_record_id = $1
   AND dst_record_id = $2
   AND field_key = $3
   AND link_type = $4
   AND deleted_at IS NULL
`, sourceID, targetID, fieldKey, linkType).Scan(&provenance, &confidence); err != nil {
		t.Fatalf("query manual reference link %s %s -> %s: %v", fieldKey, sourceID, targetID, err)
	}
	if provenance != "manual" || confidence.Valid {
		t.Fatalf("manual link %s must preserve provenance=manual confidence=NULL, got provenance=%q confidence=%#v", fieldKey, provenance, confidence)
	}
}

func countSupersedesLinks(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, sourceID uuid.UUID, targetID uuid.UUID) int {
	t.Helper()
	var count int
	if err := db.QueryRow(context.Background(), `
SELECT COUNT(*)
  FROM record_links
 WHERE src_record_id = $1
   AND dst_record_id = $2
   AND link_type = 'supersedes'
   AND deleted_at IS NULL
`, sourceID, targetID).Scan(&count); err != nil {
		t.Fatalf("count supersedes links: %v", err)
	}
	return count
}

func insertSupersedesLink(t testing.TB, db interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, actorID uuid.UUID, now time.Time) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'supersedes', NULL, 'manual', NULL, $4, $4, $5, $5)
`, incidentID, sourceID, targetID, actorID, now); err != nil {
		t.Fatalf("insert supersedes link %s -> %s: %v", sourceID, targetID, err)
	}
}

func QueryOne(t testing.TB, store *workbook.WorkbookContributionCatalog, incidentID uuid.UUID, viewSchemaID string, fieldKey string, value any, recordID uuid.UUID) map[string]any {
	t.Helper()
	rows, err := store.QueryRows(context.Background(), incidentID, viewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": value}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "decision.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query one %s: %v", fieldKey, err)
	}
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return row
		}
	}
	t.Fatalf("record %s missing from rows %#v", recordID, rows)
	return nil
}

func RowsContain(rows []map[string]any, recordID uuid.UUID) bool {
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return true
		}
	}
	return false
}

func requireMutationValidation(t testing.TB, err error, field string, reason string) {
	t.Helper()
	var ownerValidation *tasksdecisions.ValidationError
	if errors.As(err, &ownerValidation) {
		if ownerValidation.Field != field || ownerValidation.ReasonCode != reason {
			t.Fatalf("unexpected owner validation error: %#v", ownerValidation)
		}
		return
	}
	t.Fatalf("expected owner validation error, got %v", err)
}

func requireLifecycle(t testing.TB, err error) {
	t.Helper()
	var ownerLifecycleErr *tasksdecisions.LifecycleValidationError
	if errors.As(err, &ownerLifecycleErr) {
		return
	}
	t.Fatalf("expected lifecycle validation error, got %v", err)
}

func requireCellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("unexpected %s value: got %#v want %#v", fieldKey, got, want)
	}
}

func requireCellNonEmpty(t testing.TB, row map[string]any, fieldKey string) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	if got == nil || got == "" {
		t.Fatalf("expected non-empty %s value, got %#v", fieldKey, got)
	}
}

func requireCellNumericValue(t testing.TB, row map[string]any, fieldKey string, want int64) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	switch value := got.(type) {
	case int:
		if int64(value) == want {
			return
		}
	case int32:
		if int64(value) == want {
			return
		}
	case int64:
		if value == want {
			return
		}
	case float64:
		if int64(value) == want {
			return
		}
	}
	t.Fatalf("unexpected %s value: got %#v want %d", fieldKey, got, want)
}

func requireCollectionItemCount(t testing.TB, row map[string]any, fieldKey string, want int) {
	t.Helper()
	value := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"].(map[string]any)
	if value["kind"] != "collection_value_v1" {
		t.Fatalf("expected %s to be collection_value_v1, got %#v", fieldKey, value)
	}
	var got int
	switch items := value["items"].(type) {
	case []any:
		got = len(items)
	case []map[string]any:
		got = len(items)
	default:
		t.Fatalf("unexpected %s items shape: %#v", fieldKey, value["items"])
	}
	if got != want {
		t.Fatalf("unexpected %s item count: got %d want %d items=%#v", fieldKey, got, want, value["items"])
	}
}

func Time(offset time.Duration) time.Time {
	return time.Date(2026, 5, 18, 16, 0, 0, 0, time.UTC).Add(offset)
}

func stringPtr(value string) *string {
	return &value
}
