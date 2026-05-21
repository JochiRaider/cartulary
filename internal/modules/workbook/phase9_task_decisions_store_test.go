package workbook_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	phase4storetest "github.com/JochiRaider/cartulary/internal/testutil/phase4storetest"
)

func TestPhase9Sprint6_TaskRequestLifecycleDecisionLinksAndProjection_U_9_07(t *testing.T) {
	ctx := context.Background()
	harness := phase4storetest.StartStore(t, "phase9-sprint6-task-requests")
	store := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "sprint6-task@example.test", "Sprint6 Task", "Sprint6Task1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint6-task-incident", "IR-S6-TASK", "Phase 9 Sprint 6 task requests")

	beforeRecords := countSprint6Records(t, harness.DB, incident.ID)
	_, err := store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.TaskRequestsViewSchemaID,
		ClientTxnID:  "txn-phase9-sprint6-task-minimum-fail",
		Values: map[string]workbook.ValueChange{
			"task.task_kind": {Kind: "text", Text: stringPtrU911("collection")},
		},
	}, []byte("txn-phase9-sprint6-task-minimum-fail"), "req-phase9-sprint6-task-minimum-fail", sprint6Time(0))
	requireSprint6MutationValidation(t, err, "task.title", "missing_required_field")
	if got := countSprint6Records(t, harness.DB, incident.ID); got != beforeRecords {
		t.Fatalf("rejected minimum task create wrote records: got %d want %d", got, beforeRecords)
	}

	decision := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-task-decision", "approved", "Task-linked decision")
	support := mustCreateEvidenceForU911(t, store, actor, incident.ID, "txn-phase9-sprint6-task-support", "Task support record")
	dueAt := sprint6Time(24 * time.Hour)
	task := mustCreateSprint6Task(t, store, actor, incident.ID, "txn-phase9-sprint6-task-create", map[string]workbook.ValueChange{
		"task.title":               {Kind: "text", Text: stringPtrU911("Collect endpoint logs")},
		"task.task_kind":           {Kind: "text", Text: stringPtrU911("collection")},
		"task.due_at":              {Kind: "timestamp", Timestamp: &dueAt},
		"task.workstream":          {Kind: "text", Text: stringPtrU911("forensics")},
		"task.external_ticket_ref": {Kind: "text", Text: stringPtrU911("SOC-123")},
		"task.decision_record_id":  {Kind: "uuid", UUID: &decision},
	}, map[string]workbook.CollectionActionPayload{
		"task.linked_record_ids": sprint6Collection(addSprint6RecordRef(support)),
	})
	taskID := task.RecordID
	requireSprint6CellValue(t, task.Payload["row"].(map[string]any), "task.priority", "normal")
	requireSprint6CellValue(t, task.Payload["row"].(map[string]any), "task.decision_record_id", decision.String())
	if got := countSprint6TaskDecisionLinks(t, harness.DB, taskID); got != 1 {
		t.Fatalf("task decision direct link count: got %d want 1", got)
	}
	requireSprint6ManualReferenceLink(t, harness.DB, taskID, decision, "task.decision_record_id", "references_record")
	requireSprint6ManualReferenceLink(t, harness.DB, taskID, support, "task.linked_record_ids", "references_record")

	taskRows, err := store.QueryRows(ctx, incident.ID, workbook.TaskRequestsViewSchemaID, viewschema.QueryMeta{
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

	queueTask := mustCreateSprint6Task(t, store, actor, incident.ID, "txn-phase9-sprint6-task-queue-default", map[string]workbook.ValueChange{
		"task.title":     {Kind: "text", Text: stringPtrU911("Queue priority default")},
		"task.task_kind": {Kind: "text", Text: stringPtrU911("follow_up")},
	}, nil)
	queueTaskID := queueTask.RecordID
	requireSprint6CellValue(t, queueTask.Payload["row"].(map[string]any), "task.priority", "normal")
	requireSprint6CellValue(t, queueTask.Payload["row"].(map[string]any), "task.due_at", nil)
	queueDueAt := sprint6Time(48 * time.Hour)
	queueTask = mustSprint6Patch(t, store, actor, queueTaskID, workbook.TaskRequestsViewSchemaID, 1, "txn-phase9-sprint6-task-queue-edit",
		sprint6ValueChange("task.priority", workbook.ValueChange{Kind: "text", Text: stringPtrU911("high")}),
		sprint6ValueChange("task.due_at", workbook.ValueChange{Kind: "timestamp", Timestamp: &queueDueAt}))
	requireSprint6CellValue(t, queueTask.Payload["row"].(map[string]any), "task.priority", "high")
	requireSprint6CellNonEmpty(t, queueTask.Payload["row"].(map[string]any), "task.due_at")
	priorityRows, err := store.QueryRows(ctx, incident.ID, workbook.TaskRequestsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "task.priority", Op: "eq", Arg: map[string]any{"value": "high"}},
			{FieldKey: "task.due_at", Op: "range", Arg: map[string]any{"lte": queueDueAt.Format(time.RFC3339Nano)}},
		},
		Sort: []viewschema.SortEntry{{FieldKey: "task.priority", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query task priority and due filters: %v", err)
	}
	if !sprint6RowsContain(priorityRows, queueTaskID) {
		t.Fatalf("priority/due projection query missing queue task: %#v", priorityRows)
	}
	urgentTask := mustCreateSprint6Task(t, store, actor, incident.ID, "txn-phase9-sprint6-task-queue-urgent", map[string]workbook.ValueChange{
		"task.title":     {Kind: "text", Text: stringPtrU911("Queue priority urgent")},
		"task.task_kind": {Kind: "text", Text: stringPtrU911("follow_up")},
		"task.priority":  {Kind: "text", Text: stringPtrU911("urgent")},
	}, nil)
	sortedRows, err := store.QueryRows(ctx, incident.ID, workbook.TaskRequestsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "task.status", Op: "eq", Arg: map[string]any{"value": "open"}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "task.priority", Direction: "desc"}, {FieldKey: "record_id", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("query task priority sort: %v", err)
	}
	if len(sortedRows) == 0 || sortedRows[0]["record_id"] != urgentTask.RecordID.String() {
		t.Fatalf("priority sort did not put urgent task first: %#v", sortedRows)
	}
	queueTask = mustSprint6Patch(t, store, actor, queueTaskID, workbook.TaskRequestsViewSchemaID, 2, "txn-phase9-sprint6-task-queue-clear-due",
		sprint6ValueChange("task.due_at", workbook.ValueChange{Kind: "null"}))
	requireSprint6CellValue(t, queueTask.Payload["row"].(map[string]any), "task.due_at", nil)
	clearedDueRows, err := store.QueryRows(ctx, incident.ID, workbook.TaskRequestsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{
			{FieldKey: "task.priority", Op: "eq", Arg: map[string]any{"value": "high"}},
			{FieldKey: "task.due_at", Op: "range", Arg: map[string]any{"lte": queueDueAt.Format(time.RFC3339Nano)}},
		},
		Sort: []viewschema.SortEntry{{FieldKey: "task.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query task due after clear: %v", err)
	}
	if sprint6RowsContain(clearedDueRows, queueTaskID) {
		t.Fatalf("cleared due date still matched due range: %#v", clearedDueRows)
	}
	beforeInvalidPriorityVersion := sprint6RecordVersion(t, harness.DB, queueTaskID)
	_, err = sprint6Patch(store, actor, queueTaskID, workbook.TaskRequestsViewSchemaID, 3, "txn-phase9-sprint6-task-invalid-priority",
		sprint6ValueChange("task.priority", workbook.ValueChange{Kind: "text", Text: stringPtrU911("critical")}))
	requireSprint6MutationValidation(t, err, "task.priority", "invalid_value")
	if got := sprint6RecordVersion(t, harness.DB, queueTaskID); got != beforeInvalidPriorityVersion {
		t.Fatalf("invalid priority changed row version: got %d want %d", got, beforeInvalidPriorityVersion)
	}

	task = mustSprint6Patch(t, store, actor, taskID, workbook.TaskRequestsViewSchemaID, 1, "txn-phase9-sprint6-task-in-progress",
		sprint6ValueChange("task.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("in_progress")}))
	task = mustSprint6Patch(t, store, actor, taskID, workbook.TaskRequestsViewSchemaID, 2, "txn-phase9-sprint6-task-blocked",
		sprint6ValueChange("task.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("blocked")}),
		sprint6ValueChange("task.blocked_reason", workbook.ValueChange{Kind: "text", Text: stringPtrU911("Waiting for host owner")}))
	requireSprint6CellValue(t, task.Payload["row"].(map[string]any), "task.status", "blocked")
	requireSprint6CellValue(t, task.Payload["row"].(map[string]any), "task.blocked_reason", "Waiting for host owner")
	task = mustSprint6Patch(t, store, actor, taskID, workbook.TaskRequestsViewSchemaID, 3, "txn-phase9-sprint6-task-unblocked",
		sprint6ValueChange("task.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("in_progress")}))
	requireSprint6CellValue(t, task.Payload["row"].(map[string]any), "task.blocked_reason", nil)
	task = mustSprint6Patch(t, store, actor, taskID, workbook.TaskRequestsViewSchemaID, 4, "txn-phase9-sprint6-task-done",
		sprint6ValueChange("task.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("done")}))
	requireSprint6CellNonEmpty(t, task.Payload["row"].(map[string]any), "task.completed_at")
	task = mustSprint6Patch(t, store, actor, taskID, workbook.TaskRequestsViewSchemaID, 5, "txn-phase9-sprint6-task-reopen",
		sprint6ValueChange("task.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("open")}))
	requireSprint6CellValue(t, task.Payload["row"].(map[string]any), "task.completed_at", nil)

	cleared := mustSprint6Patch(t, store, actor, taskID, workbook.TaskRequestsViewSchemaID, 6, "txn-phase9-sprint6-task-clear-decision",
		sprint6ValueChange("task.decision_record_id", workbook.ValueChange{Kind: "null"}))
	requireSprint6CellValue(t, cleared.Payload["row"].(map[string]any), "task.decision_record_id", nil)
	if got := countSprint6TaskDecisionLinks(t, harness.DB, taskID); got != 0 {
		t.Fatalf("task decision direct link after clear: got %d want 0", got)
	}
	reset := mustSprint6Patch(t, store, actor, taskID, workbook.TaskRequestsViewSchemaID, 7, "txn-phase9-sprint6-task-reset-decision",
		sprint6ValueChange("task.decision_record_id", workbook.ValueChange{Kind: "uuid", UUID: &decision}))
	requireSprint6CellValue(t, reset.Payload["row"].(map[string]any), "task.decision_record_id", decision.String())
	if got := countSprint6TaskDecisionLinks(t, harness.DB, taskID); got != 1 {
		t.Fatalf("task decision direct link after reset: got %d want 1", got)
	}
	requireSprint6ManualReferenceLink(t, harness.DB, taskID, decision, "task.decision_record_id", "references_record")
	otherIncident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint6-task-other-incident", "IR-S6-TASK-OTHER", "Phase 9 Sprint 6 task requests other")
	foreignDecision := mustCreateSprint6Decision(t, store, actor, otherIncident.ID, "txn-phase9-sprint6-task-foreign-decision", "approved", "Foreign decision")
	deletedDecision := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-task-deleted-decision", "approved", "Deleted decision")
	if _, err := harness.DB.Exec(ctx, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, deletedDecision, sprint6Time(20*time.Minute), actor.ID); err != nil {
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
		_, err = sprint6Patch(store, actor, taskID, workbook.TaskRequestsViewSchemaID, 8, "txn-phase9-sprint6-task-invalid-decision-"+invalid.name,
			sprint6ValueChange("task.decision_record_id", workbook.ValueChange{Kind: "uuid", UUID: &invalid.id}))
		requireSprint6MutationValidation(t, err, "task.decision_record_id", "invalid_value")
		if got := countSprint6TaskDecisionLinks(t, harness.DB, taskID); got != 1 {
			t.Fatalf("invalid %s decision ref changed links: got %d want 1", invalid.name, got)
		}
	}

	done := mustSprint6Patch(t, store, actor, taskID, workbook.TaskRequestsViewSchemaID, 8, "txn-phase9-sprint6-task-done-again",
		sprint6ValueChange("task.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("done")}))
	requireSprint6CellValue(t, done.Payload["row"].(map[string]any), "task.status", "done")
	_, err = sprint6Patch(store, actor, taskID, workbook.TaskRequestsViewSchemaID, 9, "txn-phase9-sprint6-task-done-canceled",
		sprint6ValueChange("task.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("canceled")}))
	requireSprint6Lifecycle(t, err)

	canceled := mustCreateSprint6Task(t, store, actor, incident.ID, "txn-phase9-sprint6-task-canceled", map[string]workbook.ValueChange{
		"task.title":     {Kind: "text", Text: stringPtrU911("Canceled task")},
		"task.task_kind": {Kind: "text", Text: stringPtrU911("request")},
		"task.status":    {Kind: "text", Text: stringPtrU911("canceled")},
	}, nil)
	_, err = sprint6Patch(store, actor, canceled.RecordID, workbook.TaskRequestsViewSchemaID, 1, "txn-phase9-sprint6-task-canceled-done",
		sprint6ValueChange("task.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("done")}))
	requireSprint6Lifecycle(t, err)

	ownerless := mustCreateSprint6Task(t, store, actor, incident.ID, "txn-phase9-sprint6-task-ownerless", map[string]workbook.ValueChange{
		"task.title":     {Kind: "text", Text: stringPtrU911("Owner guard")},
		"task.task_kind": {Kind: "text", Text: stringPtrU911("request")},
	}, nil)
	_, err = sprint6Patch(store, actor, ownerless.RecordID, workbook.TaskRequestsViewSchemaID, 1, "txn-phase9-sprint6-task-ownerless-open",
		sprint6ValueChange("task.owner_user_id", workbook.ValueChange{Kind: "null"}))
	requireSprint6Lifecycle(t, err)
}

func TestPhase9Sprint6_TaskLifecycleGuardFailures_U_9_07(t *testing.T) {
	ctx := context.Background()
	harness := phase4storetest.StartStore(t, "phase9-sprint6-task-guard-failures")
	store := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "sprint6-task-guards@example.test", "Sprint6 Task Guards", "Sprint6TaskGuards1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint6-task-guard-incident", "IR-S6-TASK-GUARDS", "Phase 9 Sprint 6 task guard failures")

	beforeCreatedAt := sprint6Time(-time.Hour)
	for _, tc := range []struct {
		name   string
		values map[string]workbook.ValueChange
	}{
		{
			name: "blocked-without-reason-create",
			values: map[string]workbook.ValueChange{
				"task.title":     {Kind: "text", Text: stringPtrU911("Blocked without reason")},
				"task.task_kind": {Kind: "text", Text: stringPtrU911("request")},
				"task.status":    {Kind: "text", Text: stringPtrU911("blocked")},
			},
		},
		{
			name: "non-blocked-with-reason-create",
			values: map[string]workbook.ValueChange{
				"task.title":          {Kind: "text", Text: stringPtrU911("Open with blocked reason")},
				"task.task_kind":      {Kind: "text", Text: stringPtrU911("request")},
				"task.blocked_reason": {Kind: "text", Text: stringPtrU911("Reason is only legal while blocked")},
			},
		},
		{
			name: "non-done-with-completed-at-create",
			values: map[string]workbook.ValueChange{
				"task.title":        {Kind: "text", Text: stringPtrU911("Open with completion time")},
				"task.task_kind":    {Kind: "text", Text: stringPtrU911("request")},
				"task.completed_at": {Kind: "timestamp", Timestamp: &beforeCreatedAt},
			},
		},
		{
			name: "done-before-created-create",
			values: map[string]workbook.ValueChange{
				"task.title":        {Kind: "text", Text: stringPtrU911("Done before created")},
				"task.task_kind":    {Kind: "text", Text: stringPtrU911("request")},
				"task.status":       {Kind: "text", Text: stringPtrU911("done")},
				"task.completed_at": {Kind: "timestamp", Timestamp: &beforeCreatedAt},
			},
		},
	} {
		beforeRecords := countSprint6Records(t, harness.DB, incident.ID)
		_, err := store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
			ViewSchemaID: workbook.TaskRequestsViewSchemaID,
			ClientTxnID:  "txn-phase9-sprint6-task-guard-" + tc.name,
			Values:       tc.values,
		}, []byte("txn-phase9-sprint6-task-guard-"+tc.name), "req-phase9-sprint6-task-guard-"+tc.name, sprint6Time(0))
		requireSprint6Lifecycle(t, err)
		if got := countSprint6Records(t, harness.DB, incident.ID); got != beforeRecords {
			t.Fatalf("%s wrote partial records: got %d want %d", tc.name, got, beforeRecords)
		}
	}

	for _, tc := range []struct {
		name    string
		changes []workbook.PatchChange
	}{
		{
			name: "blocked-without-reason-patch",
			changes: []workbook.PatchChange{
				sprint6ValueChange("task.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("blocked")}),
			},
		},
		{
			name: "non-blocked-with-reason-patch",
			changes: []workbook.PatchChange{
				sprint6ValueChange("task.blocked_reason", workbook.ValueChange{Kind: "text", Text: stringPtrU911("Reason is only legal while blocked")}),
			},
		},
		{
			name: "non-done-with-completed-at-patch",
			changes: []workbook.PatchChange{
				sprint6ValueChange("task.completed_at", workbook.ValueChange{Kind: "timestamp", Timestamp: &beforeCreatedAt}),
			},
		},
		{
			name: "done-before-created-patch",
			changes: []workbook.PatchChange{
				sprint6ValueChange("task.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("done")}),
				sprint6ValueChange("task.completed_at", workbook.ValueChange{Kind: "timestamp", Timestamp: &beforeCreatedAt}),
			},
		},
	} {
		task := mustCreateSprint6Task(t, store, actor, incident.ID, "txn-phase9-sprint6-task-guard-base-"+tc.name, map[string]workbook.ValueChange{
			"task.title":     {Kind: "text", Text: stringPtrU911("Guard base " + tc.name)},
			"task.task_kind": {Kind: "text", Text: stringPtrU911("request")},
		}, nil)
		before := sprint6TaskSnapshot(t, harness.DB, task.RecordID)
		_, err := sprint6Patch(store, actor, task.RecordID, workbook.TaskRequestsViewSchemaID, before.RowVersion, "txn-phase9-sprint6-task-guard-"+tc.name, tc.changes...)
		requireSprint6Lifecycle(t, err)
		requireSprint6TaskSnapshot(t, sprint6TaskSnapshot(t, harness.DB, task.RecordID), before, tc.name)
		if got := countSprint6ReferenceLinks(t, harness.DB, task.RecordID, "task.linked_record_ids"); got != 0 {
			t.Fatalf("%s wrote partial task links: got %d want 0", tc.name, got)
		}
	}
}

func TestPhase9Sprint6_DecisionLifecycleSupersessionAndConsistency_U_9_07(t *testing.T) {
	ctx := context.Background()
	harness := phase4storetest.StartStore(t, "phase9-sprint6-decisions")
	store := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "sprint6-decision@example.test", "Sprint6 Decision", "Sprint6Decision1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint6-decision-incident", "IR-S6-DECISION", "Phase 9 Sprint 6 decisions")

	beforeRecords := countSprint6Records(t, harness.DB, incident.ID)
	_, err := store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.DecisionsViewSchemaID,
		ClientTxnID:  "txn-phase9-sprint6-decision-minimum-fail",
		Values: map[string]workbook.ValueChange{
			"decision.decision_type": {Kind: "text", Text: stringPtrU911("scope")},
		},
	}, []byte("txn-phase9-sprint6-decision-minimum-fail"), "req-phase9-sprint6-decision-minimum-fail", sprint6Time(0))
	requireSprint6MutationValidation(t, err, "decision.summary", "missing_required_field")
	if got := countSprint6Records(t, harness.DB, incident.ID); got != beforeRecords {
		t.Fatalf("rejected minimum decision create wrote records: got %d want %d", got, beforeRecords)
	}
	_, err = store.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.DecisionsViewSchemaID,
		ClientTxnID:  "txn-phase9-sprint6-decision-create-superseded",
		Values: map[string]workbook.ValueChange{
			"decision.summary":       {Kind: "text", Text: stringPtrU911("Bad superseded create")},
			"decision.decision_type": {Kind: "text", Text: stringPtrU911("scope")},
			"decision.rationale":     {Kind: "text", Text: stringPtrU911("Superseded must be explicit.")},
			"decision.status":        {Kind: "text", Text: stringPtrU911("superseded")},
		},
	}, []byte("txn-phase9-sprint6-decision-create-superseded"), "req-phase9-sprint6-decision-create-superseded", sprint6Time(0))
	requireSprint6Lifecycle(t, err)

	support := mustCreateEvidenceForU911(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-support", "Decision support record")
	affectedOne := mustCreateEvidenceForU911(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-affected-one", "Decision affected record one")
	affectedTwo := mustCreateEvidenceForU911(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-affected-two", "Decision affected record two")
	relationshipDecision := mustCreateSprint6DecisionWithCollections(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-relationships", "proposed", "Relationship decision", map[string]workbook.CollectionActionPayload{
		"decision.support_refs":        sprint6Collection(addSprint6RecordRef(support)),
		"decision.affected_record_ids": sprint6Collection(addSprint6RecordRef(affectedOne), addSprint6RecordRef(affectedOne)),
	})
	relationshipRow := relationshipDecision.Payload["row"].(map[string]any)
	requireSprint6CollectionItemCount(t, relationshipRow, "decision.support_refs", 1)
	requireSprint6CollectionItemCount(t, relationshipRow, "decision.affected_record_ids", 1)
	requireSprint6CellNumericValue(t, relationshipRow, "decision.affected_record_count", 1)
	requireSprint6ManualReferenceLink(t, harness.DB, relationshipDecision.RecordID, support, "decision.support_refs", "supported_by")
	requireSprint6ManualReferenceLink(t, harness.DB, relationshipDecision.RecordID, affectedOne, "decision.affected_record_ids", "references_record")

	relationshipDecision = mustSprint6Patch(t, store, actor, relationshipDecision.RecordID, workbook.DecisionsViewSchemaID, 1, "txn-phase9-sprint6-decision-affected-add",
		sprint6CollectionChange("decision.affected_record_ids", sprint6Collection(addSprint6RecordRef(affectedTwo))))
	relationshipRow = relationshipDecision.Payload["row"].(map[string]any)
	requireSprint6CollectionItemCount(t, relationshipRow, "decision.affected_record_ids", 2)
	requireSprint6CellNumericValue(t, relationshipRow, "decision.affected_record_count", 2)
	requireSprint6ManualReferenceLink(t, harness.DB, relationshipDecision.RecordID, affectedTwo, "decision.affected_record_ids", "references_record")

	relationshipDecision = mustSprint6Patch(t, store, actor, relationshipDecision.RecordID, workbook.DecisionsViewSchemaID, 2, "txn-phase9-sprint6-decision-affected-remove",
		sprint6CollectionChange("decision.affected_record_ids", sprint6Collection(removeSprint6RecordRef(affectedOne))))
	relationshipRow = relationshipDecision.Payload["row"].(map[string]any)
	requireSprint6CollectionItemCount(t, relationshipRow, "decision.affected_record_ids", 1)
	requireSprint6CellNumericValue(t, relationshipRow, "decision.affected_record_count", 1)
	if got := countSprint6ReferenceLinks(t, harness.DB, relationshipDecision.RecordID, "decision.affected_record_ids"); got != 1 {
		t.Fatalf("decision affected record links after remove: got %d want 1", got)
	}

	target := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-target", "proposed", "Target decision")
	source := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-source", "approved", "Superseding decision")
	executed := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-executed", "approved", "Executed decision")
	executedRow := mustSprint6Patch(t, store, actor, executed, workbook.DecisionsViewSchemaID, 1, "txn-phase9-sprint6-decision-approved-executed",
		sprint6ValueChange("decision.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("executed")}))
	requireSprint6CellValue(t, executedRow.Payload["row"].(map[string]any), "decision.status", "executed")

	_, err = sprint6Patch(store, actor, target, workbook.DecisionsViewSchemaID, 1, "txn-phase9-sprint6-decision-direct-superseded",
		sprint6ValueChange("decision.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("superseded")}))
	requireSprint6Lifecycle(t, err)
	rejected := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-rejected", "rejected", "Rejected decision")
	_, err = sprint6Patch(store, actor, rejected, workbook.DecisionsViewSchemaID, 1, "txn-phase9-sprint6-decision-rejected-proposed",
		sprint6ValueChange("decision.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("proposed")}))
	requireSprint6Lifecycle(t, err)
	_, err = sprint6Patch(store, actor, executed, workbook.DecisionsViewSchemaID, 2, "txn-phase9-sprint6-decision-executed-approved",
		sprint6ValueChange("decision.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("approved")}))
	requireSprint6Lifecycle(t, err)

	request := timeline.SupersedeRequest{
		BaseRowVersion:      1,
		ClientTxnID:         "txn-phase9-sprint6-decision-supersede",
		Reason:              "Supersede with better containment rationale.",
		ReplacementRecordID: &source,
	}
	result, err := store.SupersedeDecision(ctx, actor, target, request, timeline.TimelineActionRequestHash(request.BaseRowVersion, request.ClientTxnID, &request.Reason, request.ReplacementRecordID), "req-phase9-sprint6-decision-supersede", sprint6Time(time.Hour))
	if err != nil {
		t.Fatalf("supersede proposed target: %v", err)
	}
	if result.Payload["view_schema_id"] != workbook.DecisionsViewSchemaID || result.Payload["target_status"] != "superseded" {
		t.Fatalf("unexpected decision supersede payload: %#v", result.Payload)
	}
	if len(result.AdditionalRecordChanges) != 2 {
		t.Fatalf("expected two changed decision rows, got %d", len(result.AdditionalRecordChanges))
	}
	if got := countSprint6SupersedesLinks(t, harness.DB, source, target); got != 1 {
		t.Fatalf("decision supersedes link count: got %d want 1", got)
	}
	decisionRows, err := store.QueryRows(ctx, incident.ID, workbook.DecisionsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "decision.is_superseded", Op: "eq", Arg: map[string]any{"value": true}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "decision.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query decision superseded projection: %v", err)
	}
	if !sprint6RowsContain(decisionRows, target) {
		t.Fatalf("superseded target missing from projection rows: %#v", decisionRows)
	}
	sourceRows, err := store.QueryRows(ctx, incident.ID, workbook.DecisionsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "decision.supersedes_record_id", Op: "eq", Arg: map[string]any{"value": target.String()}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "decision.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query decision supersedes projection: %v", err)
	}
	if len(sourceRows) != 1 || sourceRows[0]["record_id"] != source.String() {
		t.Fatalf("superseding decision missing from projection rows: %#v", sourceRows)
	}
	replay, err := store.SupersedeDecision(ctx, actor, target, request, timeline.TimelineActionRequestHash(request.BaseRowVersion, request.ClientTxnID, &request.Reason, request.ReplacementRecordID), "req-phase9-sprint6-decision-supersede-replay", sprint6Time(2*time.Hour))
	if err != nil {
		t.Fatalf("replay decision supersede: %v", err)
	}
	if !replay.Replayed {
		t.Fatalf("expected idempotent replay")
	}
	if got := countSprint6SupersedesLinks(t, harness.DB, source, target); got != 1 {
		t.Fatalf("decision supersedes link count after replay: got %d want 1", got)
	}

	executedTarget := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-executed-target", "executed", "Executed target")
	executedSource := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-executed-source", "approved", "Executed target replacement")
	executedRequest := timeline.SupersedeRequest{
		BaseRowVersion:      1,
		ClientTxnID:         "txn-phase9-sprint6-decision-supersede-executed",
		Reason:              "Supersede executed decision.",
		ReplacementRecordID: &executedSource,
	}
	if _, err := store.SupersedeDecision(ctx, actor, executedTarget, executedRequest, timeline.TimelineActionRequestHash(executedRequest.BaseRowVersion, executedRequest.ClientTxnID, &executedRequest.Reason, executedRequest.ReplacementRecordID), "req-phase9-sprint6-decision-supersede-executed", sprint6Time(3*time.Hour)); err != nil {
		t.Fatalf("supersede executed target: %v", err)
	}
	row := sprint6QueryOne(t, store, incident.ID, workbook.DecisionsViewSchemaID, "decision.is_superseded", true, executedTarget)
	requireSprint6CellValue(t, row, "decision.status", "executed")
	requireSprint6CellValue(t, row, "decision.is_superseded", true)

	badSource := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-bad-source", "proposed", "Inconsistent source")
	badTarget := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-bad-target", "proposed", "Inconsistent target")
	if _, err := harness.DB.Exec(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'supersedes', NULL, 'manual', NULL, $4, $4, $5, $5)
`, incident.ID, badSource, badTarget, actor.ID, sprint6Time(4*time.Hour)); err != nil {
		t.Fatalf("seed inconsistent supersedes link: %v", err)
	}
	_, err = sprint6Patch(store, actor, badSource, workbook.DecisionsViewSchemaID, 1, "txn-phase9-sprint6-decision-inconsistent-fail",
		sprint6ValueChange("decision.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911("approved")}))
	requireSprint6Lifecycle(t, err)
	beforeInconsistentVersion := sprint6RecordVersion(t, harness.DB, badSource)
	_, err = sprint6Patch(store, actor, badSource, workbook.DecisionsViewSchemaID, 1, "txn-phase9-sprint6-decision-inconsistent-rationale-fail",
		sprint6ValueChange("decision.rationale", workbook.ValueChange{Kind: "text", Text: stringPtrU911("Ordinary scalar edits must fail closed.")}))
	requireSprint6Lifecycle(t, err)
	if got := sprint6RecordVersion(t, harness.DB, badSource); got != beforeInconsistentVersion {
		t.Fatalf("inconsistent decision scalar patch changed row version: got %d want %d", got, beforeInconsistentVersion)
	}
	beforeSupportLinks := countSprint6ReferenceLinks(t, harness.DB, badSource, "decision.support_refs")
	_, err = sprint6Patch(store, actor, badSource, workbook.DecisionsViewSchemaID, 1, "txn-phase9-sprint6-decision-inconsistent-support-fail",
		sprint6CollectionChange("decision.support_refs", sprint6Collection(addSprint6RecordRef(support))))
	requireSprint6Lifecycle(t, err)
	if got := countSprint6ReferenceLinks(t, harness.DB, badSource, "decision.support_refs"); got != beforeSupportLinks {
		t.Fatalf("inconsistent decision support patch changed links: got %d want %d", got, beforeSupportLinks)
	}
	if got := sprint6RecordVersion(t, harness.DB, badSource); got != beforeInconsistentVersion {
		t.Fatalf("inconsistent decision support patch changed row version: got %d want %d", got, beforeInconsistentVersion)
	}
	beforeAffectedLinks := countSprint6ReferenceLinks(t, harness.DB, badSource, "decision.affected_record_ids")
	_, err = sprint6Patch(store, actor, badSource, workbook.DecisionsViewSchemaID, 1, "txn-phase9-sprint6-decision-inconsistent-affected-fail",
		sprint6CollectionChange("decision.affected_record_ids", sprint6Collection(addSprint6RecordRef(affectedTwo))))
	requireSprint6Lifecycle(t, err)
	if got := countSprint6ReferenceLinks(t, harness.DB, badSource, "decision.affected_record_ids"); got != beforeAffectedLinks {
		t.Fatalf("inconsistent decision affected patch changed links: got %d want %d", got, beforeAffectedLinks)
	}
	if got := sprint6RecordVersion(t, harness.DB, badSource); got != beforeInconsistentVersion {
		t.Fatalf("inconsistent decision affected patch changed row version: got %d want %d", got, beforeInconsistentVersion)
	}
}

func TestPhase9Sprint6_SupersedeDecisionRejectsInconsistentSourceOrTarget_U_9_07(t *testing.T) {
	ctx := context.Background()
	harness := phase4storetest.StartStore(t, "phase9-sprint6-decision-supersede-inconsistent")
	store := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "sprint6-decision-supersede-inconsistent@example.test", "Sprint6 Decision Supersede Inconsistent", "Sprint6DecisionSupersede1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint6-decision-supersede-inconsistent-incident", "IR-S6-DECISION-SUPERSEDE-INCONSISTENT", "Phase 9 Sprint 6 decision supersede inconsistent")

	inconsistentSource := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-inconsistent-source", "proposed", "Inconsistent source")
	sourceExistingTarget := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-inconsistent-source-existing-target", "proposed", "Existing target")
	validTarget := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-source-fail-target", "proposed", "Source fail target")
	insertSprint6SupersedesLink(t, harness.DB, incident.ID, inconsistentSource, sourceExistingTarget, actor.ID, sprint6Time(time.Hour))

	sourceBefore := sprint6DecisionSnapshot(t, harness.DB, inconsistentSource)
	sourceExistingTargetBefore := sprint6DecisionSnapshot(t, harness.DB, sourceExistingTarget)
	validTargetBefore := sprint6DecisionSnapshot(t, harness.DB, validTarget)
	sourceRequest := timeline.SupersedeRequest{
		BaseRowVersion:      validTargetBefore.RowVersion,
		ClientTxnID:         "txn-phase9-sprint6-decision-inconsistent-source-route",
		Reason:              "Attempt explicit supersession with inconsistent source.",
		ReplacementRecordID: &inconsistentSource,
	}
	_, err := store.SupersedeDecision(ctx, actor, validTarget, sourceRequest, timeline.TimelineActionRequestHash(sourceRequest.BaseRowVersion, sourceRequest.ClientTxnID, &sourceRequest.Reason, sourceRequest.ReplacementRecordID), "req-phase9-sprint6-decision-inconsistent-source-route", sprint6Time(2*time.Hour))
	requireSprint6Lifecycle(t, err)
	requireSprint6DecisionSnapshot(t, sprint6DecisionSnapshot(t, harness.DB, inconsistentSource), sourceBefore, "inconsistent source supersede")
	requireSprint6DecisionSnapshot(t, sprint6DecisionSnapshot(t, harness.DB, sourceExistingTarget), sourceExistingTargetBefore, "inconsistent source existing target")
	requireSprint6DecisionSnapshot(t, sprint6DecisionSnapshot(t, harness.DB, validTarget), validTargetBefore, "valid target rejected by inconsistent source")
	if got := countSprint6SupersedesLinks(t, harness.DB, inconsistentSource, validTarget); got != 0 {
		t.Fatalf("inconsistent source supersede wrote link to attempted target: got %d want 0", got)
	}
	if got := countSprint6ReferenceLinks(t, harness.DB, inconsistentSource, "decision.support_refs"); got != 0 {
		t.Fatalf("inconsistent source supersede wrote support links: got %d want 0", got)
	}
	if got := countSprint6ReferenceLinks(t, harness.DB, validTarget, "decision.affected_record_ids"); got != 0 {
		t.Fatalf("inconsistent source supersede wrote affected links: got %d want 0", got)
	}

	validSource := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-target-fail-source", "approved", "Valid source")
	inconsistentTarget := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-inconsistent-target", "proposed", "Inconsistent target")
	targetExistingSource := mustCreateSprint6Decision(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-inconsistent-target-existing-source", "approved", "Existing superseding source")
	insertSprint6SupersedesLink(t, harness.DB, incident.ID, targetExistingSource, inconsistentTarget, actor.ID, sprint6Time(3*time.Hour))

	validSourceBefore := sprint6DecisionSnapshot(t, harness.DB, validSource)
	inconsistentTargetBefore := sprint6DecisionSnapshot(t, harness.DB, inconsistentTarget)
	targetExistingSourceBefore := sprint6DecisionSnapshot(t, harness.DB, targetExistingSource)
	targetRequest := timeline.SupersedeRequest{
		BaseRowVersion:      inconsistentTargetBefore.RowVersion,
		ClientTxnID:         "txn-phase9-sprint6-decision-inconsistent-target-route",
		Reason:              "Attempt explicit supersession against inconsistent target.",
		ReplacementRecordID: &validSource,
	}
	_, err = store.SupersedeDecision(ctx, actor, inconsistentTarget, targetRequest, timeline.TimelineActionRequestHash(targetRequest.BaseRowVersion, targetRequest.ClientTxnID, &targetRequest.Reason, targetRequest.ReplacementRecordID), "req-phase9-sprint6-decision-inconsistent-target-route", sprint6Time(4*time.Hour))
	requireSprint6Lifecycle(t, err)
	requireSprint6DecisionSnapshot(t, sprint6DecisionSnapshot(t, harness.DB, validSource), validSourceBefore, "valid source rejected by inconsistent target")
	requireSprint6DecisionSnapshot(t, sprint6DecisionSnapshot(t, harness.DB, inconsistentTarget), inconsistentTargetBefore, "inconsistent target supersede")
	requireSprint6DecisionSnapshot(t, sprint6DecisionSnapshot(t, harness.DB, targetExistingSource), targetExistingSourceBefore, "inconsistent target existing source")
	if got := countSprint6SupersedesLinks(t, harness.DB, validSource, inconsistentTarget); got != 0 {
		t.Fatalf("inconsistent target supersede wrote attempted link: got %d want 0", got)
	}
	if got := countSprint6ReferenceLinks(t, harness.DB, validSource, "decision.support_refs"); got != 0 {
		t.Fatalf("inconsistent target supersede wrote support links: got %d want 0", got)
	}
	if got := countSprint6ReferenceLinks(t, harness.DB, inconsistentTarget, "decision.affected_record_ids"); got != 0 {
		t.Fatalf("inconsistent target supersede wrote affected links: got %d want 0", got)
	}
}

func TestPhase9Sprint6_DecisionTerminalTransitionMatrix_U_9_07(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase9-sprint6-decision-terminal-matrix")
	store := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "sprint6-decision-terminal@example.test", "Sprint6 Decision Terminal", "Sprint6DecisionTerminal1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-sprint6-decision-terminal-incident", "IR-S6-DECISION-TERMINAL", "Phase 9 Sprint 6 decision terminal matrix")

	for _, from := range []string{"rejected", "executed", "superseded"} {
		for _, to := range []string{"proposed", "approved", "rejected", "executed", "superseded"} {
			name := from + "-to-" + to
			decisionID := mustCreateSprint6DecisionInTerminalState(t, store, actor, incident.ID, "txn-phase9-sprint6-decision-terminal-base-"+name, from)
			before := sprint6DecisionSnapshot(t, harness.DB, decisionID)
			changes := []workbook.PatchChange{
				sprint6ValueChange("decision.status", workbook.ValueChange{Kind: "text", Text: stringPtrU911(to)}),
			}
			if from == to {
				changes = append(changes, sprint6ValueChange("decision.rationale", workbook.ValueChange{Kind: "text", Text: stringPtrU911("Idempotent in-state terminal write remains ordinary scalar work.")}))
			}
			_, err := sprint6Patch(store, actor, decisionID, workbook.DecisionsViewSchemaID, before.RowVersion, "txn-phase9-sprint6-decision-terminal-"+name, changes...)
			if from == to && from != "superseded" {
				if err != nil {
					t.Fatalf("%s should allow in-state terminal write, got %v", name, err)
				}
				after := sprint6DecisionSnapshot(t, harness.DB, decisionID)
				if after.Status != from || after.Rationale != "Idempotent in-state terminal write remains ordinary scalar work." || after.RowVersion <= before.RowVersion {
					t.Fatalf("%s unexpected in-state result: before=%#v after=%#v", name, before, after)
				}
				continue
			}
			requireSprint6Lifecycle(t, err)
			requireSprint6DecisionSnapshot(t, sprint6DecisionSnapshot(t, harness.DB, decisionID), before, name)
		}
	}
}

func mustCreateSprint6Decision(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string, summary string) uuid.UUID {
	t.Helper()
	return mustCreateSprint6DecisionWithCollections(t, store, actor, incidentID, clientTxnID, status, summary, nil).RecordID
}

func mustCreateSprint6DecisionInTerminalState(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string) uuid.UUID {
	t.Helper()
	if status != "superseded" {
		return mustCreateSprint6Decision(t, store, actor, incidentID, clientTxnID, status, "Terminal "+status)
	}
	target := mustCreateSprint6Decision(t, store, actor, incidentID, clientTxnID+"-target", "proposed", "Superseded target")
	source := mustCreateSprint6Decision(t, store, actor, incidentID, clientTxnID+"-source", "approved", "Superseding source")
	request := timeline.SupersedeRequest{
		BaseRowVersion:      1,
		ClientTxnID:         clientTxnID + "-supersede",
		Reason:              "Create explicit superseded terminal state.",
		ReplacementRecordID: &source,
	}
	if _, err := store.SupersedeDecision(context.Background(), actor, target, request, timeline.TimelineActionRequestHash(request.BaseRowVersion, request.ClientTxnID, &request.Reason, request.ReplacementRecordID), "req-"+clientTxnID+"-supersede", sprint6Time(time.Hour)); err != nil {
		t.Fatalf("create superseded decision %s: %v", clientTxnID, err)
	}
	return target
}

func mustCreateSprint6DecisionWithCollections(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string, summary string, collections map[string]workbook.CollectionActionPayload) workbook.MutationResult {
	t.Helper()
	values := map[string]workbook.ValueChange{
		"decision.summary":       {Kind: "text", Text: &summary},
		"decision.decision_type": {Kind: "text", Text: stringPtrU911("containment")},
		"decision.rationale":     {Kind: "text", Text: stringPtrU911("The decision is needed for coordinated response.")},
	}
	if status != "" {
		values["decision.status"] = workbook.ValueChange{Kind: "text", Text: &status}
	}
	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, workbook.CreateRequest{
		ViewSchemaID: workbook.DecisionsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values:       values,
		Collections:  collections,
	}, []byte(clientTxnID), "req-"+clientTxnID, sprint6Time(0))
	if err != nil {
		t.Fatalf("create decision %s: %v", clientTxnID, err)
	}
	return result
}

type sprint6TaskState struct {
	RowVersion    int64
	Status        string
	BlockedReason sql.NullString
	CompletedAt   sql.NullTime
	OwnerUserID   sql.NullString
}

func sprint6TaskSnapshot(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, recordID uuid.UUID) sprint6TaskState {
	t.Helper()
	var state sprint6TaskState
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

func requireSprint6TaskSnapshot(t testing.TB, got sprint6TaskState, want sprint6TaskState, context string) {
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

type sprint6DecisionState struct {
	RowVersion         int64
	Status             string
	Rationale          string
	IncomingSupersedes int
	OutgoingSupersedes int
}

func sprint6DecisionSnapshot(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, recordID uuid.UUID) sprint6DecisionState {
	t.Helper()
	var state sprint6DecisionState
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

func requireSprint6DecisionSnapshot(t testing.TB, got sprint6DecisionState, want sprint6DecisionState, context string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s changed decision snapshot: got %#v want %#v", context, got, want)
	}
}

func mustCreateSprint6Task(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, values map[string]workbook.ValueChange, collections map[string]workbook.CollectionActionPayload) workbook.MutationResult {
	t.Helper()
	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, workbook.CreateRequest{
		ViewSchemaID: workbook.TaskRequestsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values:       values,
		Collections:  collections,
	}, []byte(clientTxnID), "req-"+clientTxnID, sprint6Time(0))
	if err != nil {
		t.Fatalf("create task %s: %v", clientTxnID, err)
	}
	return result
}

func sprint6Patch(store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...workbook.PatchChange) (workbook.MutationResult, error) {
	return store.PatchWorkbookRow(context.Background(), actor, recordID, workbook.PatchRequest{
		ViewSchemaID:   viewSchemaID,
		BaseRowVersion: baseRowVersion,
		ClientTxnID:    clientTxnID,
		Changes:        changes,
	}, []byte(clientTxnID), "req-"+clientTxnID, sprint6Time(30*time.Minute))
}

func mustSprint6Patch(t testing.TB, store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, clientTxnID string, changes ...workbook.PatchChange) workbook.MutationResult {
	t.Helper()
	result, err := sprint6Patch(store, actor, recordID, viewSchemaID, baseRowVersion, clientTxnID, changes...)
	if err != nil {
		t.Fatalf("patch %s: %v", clientTxnID, err)
	}
	return result
}

func sprint6ValueChange(fieldKey string, value workbook.ValueChange) workbook.PatchChange {
	return workbook.PatchChange{FieldKey: fieldKey, Value: &value}
}

func sprint6CollectionChange(fieldKey string, value workbook.CollectionActionPayload) workbook.PatchChange {
	return workbook.PatchChange{FieldKey: fieldKey, Collection: &value}
}

func sprint6Collection(actions ...workbook.CollectionAction) workbook.CollectionActionPayload {
	return workbook.CollectionActionPayload{Actions: actions}
}

func addSprint6RecordRef(recordID uuid.UUID) workbook.CollectionAction {
	return workbook.CollectionAction{Op: "add_record_ref", LinkedRecordID: &recordID}
}

func removeSprint6RecordRef(recordID uuid.UUID) workbook.CollectionAction {
	return workbook.CollectionAction{Op: "remove_record_ref", ItemRef: "record_ref:" + recordID.String()}
}

func countSprint6Records(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, incidentID uuid.UUID) int {
	t.Helper()
	var count int
	if err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM records WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		t.Fatalf("count records: %v", err)
	}
	return count
}

func sprint6RecordVersion(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, recordID uuid.UUID) int64 {
	t.Helper()
	var version int64
	if err := db.QueryRow(context.Background(), `SELECT row_version FROM records WHERE record_id = $1`, recordID).Scan(&version); err != nil {
		t.Fatalf("query record version: %v", err)
	}
	return version
}

func countSprint6TaskDecisionLinks(t testing.TB, db interface {
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

func countSprint6ReferenceLinks(t testing.TB, db interface {
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

func requireSprint6ManualReferenceLink(t testing.TB, db interface {
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

func countSprint6SupersedesLinks(t testing.TB, db interface {
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

func insertSprint6SupersedesLink(t testing.TB, db interface {
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

func sprint6QueryOne(t testing.TB, store *workbook.Store, incidentID uuid.UUID, viewSchemaID string, fieldKey string, value any, recordID uuid.UUID) map[string]any {
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

func sprint6RowsContain(rows []map[string]any, recordID uuid.UUID) bool {
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return true
		}
	}
	return false
}

func requireSprint6MutationValidation(t testing.TB, err error, field string, reason string) {
	t.Helper()
	var validationErr *workbook.MutationValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected mutation validation error, got %v", err)
	}
	if validationErr.Field != field || validationErr.ReasonCode != reason {
		t.Fatalf("unexpected validation error: %#v", validationErr)
	}
}

func requireSprint6Lifecycle(t testing.TB, err error) {
	t.Helper()
	var lifecycleErr *workbook.LifecycleValidationError
	if !errors.As(err, &lifecycleErr) {
		t.Fatalf("expected lifecycle validation error, got %v", err)
	}
}

func requireSprint6CellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("unexpected %s value: got %#v want %#v", fieldKey, got, want)
	}
}

func requireSprint6CellNonEmpty(t testing.TB, row map[string]any, fieldKey string) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	if got == nil || got == "" {
		t.Fatalf("expected non-empty %s value, got %#v", fieldKey, got)
	}
}

func requireSprint6CellNumericValue(t testing.TB, row map[string]any, fieldKey string, want int64) {
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

func requireSprint6CollectionItemCount(t testing.TB, row map[string]any, fieldKey string, want int) {
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

func sprint6Time(offset time.Duration) time.Time {
	return time.Date(2026, 5, 18, 16, 0, 0, 0, time.UTC).Add(offset)
}
