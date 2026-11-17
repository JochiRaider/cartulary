package tasksdecisions_test

import (
	"context"
	"errors"
	"testing"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
)

// REQ-02-109a: a persisted Task whose legal status violates REQ-02-103 is
// inconsistent. A mutation that leaves it illegal fails closed with
// inconsistent_machine_state and commits nothing; a mutation whose result is
// legal is the corrective action and commits. A blank blocker counts as absent.
func TestTaskInconsistentStateFailsClosedUntilCorrected_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-task-decision-task-inconsistent")
	owner := appsupport.NewTaskDecisionOwner(harness.DB, conflicttest.NewCodec("workbook"))
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "task-inconsistent@example.test", DisplayName: "Task Inconsistent"})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "task-inconsistent-incident", "IR-TASK-INCONSISTENT", "Task inconsistent state")
	for name, blocker := range map[string]any{"null-blocker": nil} {
		t.Run(name, func(t *testing.T) {
			task := mustCreateTask(t, owner, actor, incident.ID, "create-"+name, map[string]tasksdecisions.FieldValue{
				"task.title":          {Text: stringPtr("Inconsistent " + name)},
				"task.task_kind":      {Text: stringPtr("request")},
				"task.status":         {Text: stringPtr("blocked")},
				"task.blocked_reason": {Text: stringPtr("Waiting")},
			}, nil)
			// Persist a contradiction that no legal mutation can produce.
			if _, err := harness.DB.Exec(context.Background(), "UPDATE task_requests SET blocked_reason = $2 WHERE record_id = $1", task.RecordID, blocker); err != nil {
				t.Fatal(err)
			}
			before := taskSnapshot(t, harness.DB, task.RecordID)
			_, err := patchRecord(owner, actor, task.RecordID, tasksdecisions.TaskRequestsViewSchemaID, before.RowVersion, "uncorrected-"+name,
				valueChange("task.title", tasksdecisions.FieldValue{Text: stringPtr("Still inconsistent")}))
			requireLifecycleGuards(t, err, "inconsistent_machine_state", "blocked_requires_reason")
			requireLifecycleStatuses(t, err, "blocked", "blocked")
			requireTaskSnapshot(t, taskSnapshot(t, harness.DB, task.RecordID), before, name)
			corrected, err := patchRecord(owner, actor, task.RecordID, tasksdecisions.TaskRequestsViewSchemaID, before.RowVersion, "corrective-"+name,
				valueChange("task.blocked_reason", tasksdecisions.FieldValue{Text: stringPtr("Waiting on vendor")}))
			if err != nil {
				t.Fatalf("corrective patch rejected: %v", err)
			}
			requireCellValue(t, corrected.Row, "task.blocked_reason", "Waiting on vendor")
		})
	}
	// A whitespace-only stored blocker is not a legal stored representation of
	// the text contract (writes normalize it to null), so it is a storage
	// invariant failure: unclassified (internal_error), not illegal_transition.
	t.Run("blank-stored-blocker", func(t *testing.T) {
		task := mustCreateTask(t, owner, actor, incident.ID, "create-blank", map[string]tasksdecisions.FieldValue{
			"task.title":          {Text: stringPtr("Blank stored blocker")},
			"task.task_kind":      {Text: stringPtr("request")},
			"task.status":         {Text: stringPtr("blocked")},
			"task.blocked_reason": {Text: stringPtr("Waiting")},
		}, nil)
		if _, err := harness.DB.Exec(context.Background(), "UPDATE task_requests SET blocked_reason = '   ' WHERE record_id = $1", task.RecordID); err != nil {
			t.Fatal(err)
		}
		before := taskSnapshot(t, harness.DB, task.RecordID)
		_, err := patchRecord(owner, actor, task.RecordID, tasksdecisions.TaskRequestsViewSchemaID, before.RowVersion, "blank-patch",
			valueChange("task.title", tasksdecisions.FieldValue{Text: stringPtr("Changed")}))
		var lifecycle *tasksdecisions.LifecycleValidationError
		if err == nil || errors.As(err, &lifecycle) {
			t.Fatalf("blank stored blocker: want an unclassified invariant error, got %v", err)
		}
		requireTaskSnapshot(t, taskSnapshot(t, harness.DB, task.RecordID), before, "blank-stored-blocker")
	})
}
