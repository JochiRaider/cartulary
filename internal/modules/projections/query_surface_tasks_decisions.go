package projections

func taskRequestQuerySurfaces() []genericSurface {
	return []genericSurface{{
		viewSchemaID: taskRequestsViewSchemaID,
		fromSQL:      "FROM task_request_grid_projection p JOIN records r ON r.record_id = p.record_id",
		recordExpr:   "p.record_id",
		incidentExpr: "p.incident_id",
		fields: []genericField{
			{key: "task.title", expr: "p.title", kind: fieldKindText},
			{key: "task.status", expr: "p.status", kind: fieldKindText},
			{key: "task.owner_user_id", expr: "p.owner_user_id", kind: fieldKindText},
			{key: "task.priority", expr: "p.priority", kind: fieldKindText},
			{key: "task.task_kind", expr: "p.task_kind", kind: fieldKindText},
			{key: "task.workstream", expr: "p.workstream", kind: fieldKindText},
			{key: "task.due_at", expr: "p.due_at", kind: fieldKindTimestamp},
			{key: "task.requester_party_text", expr: "p.requester_party_text", kind: fieldKindText},
			{key: "task.requester_party_id", expr: "p.requester_party_id", kind: fieldKindText},
			{key: "task.blocked_reason", expr: "p.blocked_reason", kind: fieldKindText},
			{key: "task.completed_at", expr: "p.completed_at", kind: fieldKindTimestamp},
			{key: "task.external_ticket_ref", expr: "p.external_ticket_ref", kind: fieldKindText},
			{key: "task.closure_summary", expr: "p.closure_summary", kind: fieldKindText},
			{key: "task.linked_record_ids", expr: recordRefCollectionExprFor("p", "task.linked_record_ids", "references_record"), kind: fieldKindCollection},
			{key: "task.decision_record_id", expr: "p.decision_record_id", kind: fieldKindText},
			{key: "task.linked_record_count", expr: "p.linked_record_count", kind: fieldKindNumber},
			{key: "task.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
			{key: "task.no_owner", expr: "p.no_owner", kind: fieldKindBool},
		},
	}}
}

func decisionQuerySurfaces() []genericSurface {
	return []genericSurface{{
		viewSchemaID: decisionsViewSchemaID,
		fromSQL:      "FROM decision_grid_projection p JOIN records r ON r.record_id = p.record_id",
		recordExpr:   "p.record_id",
		incidentExpr: "p.incident_id",
		fields: []genericField{
			{key: "decision.summary", expr: "p.summary", kind: fieldKindText},
			{key: "decision.status", expr: "p.status", kind: fieldKindText},
			{key: "decision.owner_user_id", expr: "p.owner_user_id", kind: fieldKindText},
			{key: "decision.decision_type", expr: "p.decision_type", kind: fieldKindText},
			{key: "decision.decided_at", expr: "p.decided_at", kind: fieldKindTimestamp},
			{key: "decision.rationale", expr: "p.rationale", kind: fieldKindText},
			{key: "decision.support_refs", expr: recordRefCollectionExprFor("p", "decision.support_refs", "supported_by"), kind: fieldKindCollection},
			{key: "decision.affected_record_ids", expr: recordRefCollectionExprFor("p", "decision.affected_record_ids", "references_record"), kind: fieldKindCollection},
			{key: "decision.affected_record_count", expr: "p.affected_record_count", kind: fieldKindNumber},
			{key: "decision.supersedes_record_id", expr: "p.supersedes_record_id", kind: fieldKindText},
			{key: "decision.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
			{key: "decision.is_superseded", expr: "p.is_superseded", kind: fieldKindBool},
		},
	}}
}
