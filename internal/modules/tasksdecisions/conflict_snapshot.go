package tasksdecisions

import conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"

func taskDecisionConflictSnapshotProjector(viewSchemaID string) (conflicttokens.RevisionSnapshotProjector, bool) {
	var fieldSourceKeys map[string]string
	var snapshotSchemaID string
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		snapshotSchemaID = "cartulary.revisions.snapshot.task_request.v1"
		fieldSourceKeys = map[string]string{
			"task.title":                "title",
			"task.status":               "status",
			"task.owner_user_id":        "owner_user_id",
			"task.priority":             "priority",
			"task.task_kind":            "task_kind",
			"task.workstream":           "workstream",
			"task.due_at":               "due_at",
			"task.requester_party_text": "requester_party_text",
			"task.requester_party_id":   "requester_party_id",
			"task.blocked_reason":       "blocked_reason",
			"task.completed_at":         "completed_at",
			"task.external_ticket_ref":  "external_ticket_ref",
			"task.closure_summary":      "closure_summary",
			"task.decision_record_id":   "decision_record_id",
		}
	case DecisionsViewSchemaID:
		snapshotSchemaID = "cartulary.revisions.snapshot.decision.v1"
		fieldSourceKeys = map[string]string{
			"decision.summary":       "summary",
			"decision.status":        "status",
			"decision.owner_user_id": "owner_user_id",
			"decision.decision_type": "decision_type",
			"decision.decided_at":    "decided_at",
			"decision.rationale":     "rationale",
		}
	default:
		return conflicttokens.RevisionSnapshotProjector{}, false
	}
	projector, err := conflicttokens.NewRevisionSnapshotProjector(snapshotSchemaID, fieldSourceKeys)
	if err != nil {
		panic(err)
	}
	return projector, true
}
