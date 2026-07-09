package projectionprovider

import "github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"

const (
	taskRequestsViewSchemaID = "cartulary.view.task_requests.v1"
	decisionsViewSchemaID    = "cartulary.view.decisions.v1"
)

func TaskRequestQuerySurfaces() []providercontract.QuerySurface {
	return []providercontract.QuerySurface{{
		ViewSchemaID: taskRequestsViewSchemaID,
		FromSQL:      "FROM task_request_grid_projection p JOIN records r ON r.record_id = p.record_id",
		RecordExpr:   "p.record_id",
		IncidentExpr: "p.incident_id",
		Fields: []providercontract.QueryField{
			{Key: "task.title", Expr: "p.title", Kind: providercontract.FieldKindText},
			{Key: "task.status", Expr: "p.status", Kind: providercontract.FieldKindText},
			{Key: "task.owner_user_id", Expr: "p.owner_user_id", Kind: providercontract.FieldKindText},
			{Key: "task.priority", Expr: "p.priority", Kind: providercontract.FieldKindText},
			{Key: "task.task_kind", Expr: "p.task_kind", Kind: providercontract.FieldKindText},
			{Key: "task.workstream", Expr: "p.workstream", Kind: providercontract.FieldKindText},
			{Key: "task.due_at", Expr: "p.due_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "task.requester_party_text", Expr: "p.requester_party_text", Kind: providercontract.FieldKindText},
			{Key: "task.requester_party_id", Expr: "p.requester_party_id", Kind: providercontract.FieldKindText},
			{Key: "task.blocked_reason", Expr: "p.blocked_reason", Kind: providercontract.FieldKindText},
			{Key: "task.completed_at", Expr: "p.completed_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "task.external_ticket_ref", Expr: "p.external_ticket_ref", Kind: providercontract.FieldKindText},
			{Key: "task.closure_summary", Expr: "p.closure_summary", Kind: providercontract.FieldKindText},
			{Key: "task.linked_record_ids", Expr: recordRefCollectionExprFor("p", "task.linked_record_ids", "references_record"), Kind: providercontract.FieldKindCollection},
			{Key: "task.decision_record_id", Expr: "p.decision_record_id", Kind: providercontract.FieldKindText},
			{Key: "task.linked_record_count", Expr: "p.linked_record_count", Kind: providercontract.FieldKindNumber},
			{Key: "task.updated_at", Expr: "p.updated_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "task.no_owner", Expr: "p.no_owner", Kind: providercontract.FieldKindBool},
		},
	}}
}

func DecisionQuerySurfaces() []providercontract.QuerySurface {
	return []providercontract.QuerySurface{{
		ViewSchemaID: decisionsViewSchemaID,
		FromSQL:      "FROM decision_grid_projection p JOIN records r ON r.record_id = p.record_id",
		RecordExpr:   "p.record_id",
		IncidentExpr: "p.incident_id",
		Fields: []providercontract.QueryField{
			{Key: "decision.summary", Expr: "p.summary", Kind: providercontract.FieldKindText},
			{Key: "decision.status", Expr: "p.status", Kind: providercontract.FieldKindText},
			{Key: "decision.owner_user_id", Expr: "p.owner_user_id", Kind: providercontract.FieldKindText},
			{Key: "decision.decision_type", Expr: "p.decision_type", Kind: providercontract.FieldKindText},
			{Key: "decision.decided_at", Expr: "p.decided_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "decision.rationale", Expr: "p.rationale", Kind: providercontract.FieldKindText},
			{Key: "decision.support_refs", Expr: recordRefCollectionExprFor("p", "decision.support_refs", "supported_by"), Kind: providercontract.FieldKindCollection},
			{Key: "decision.affected_record_ids", Expr: recordRefCollectionExprFor("p", "decision.affected_record_ids", "references_record"), Kind: providercontract.FieldKindCollection},
			{Key: "decision.affected_record_count", Expr: "p.affected_record_count", Kind: providercontract.FieldKindNumber},
			{Key: "decision.supersedes_record_id", Expr: "p.supersedes_record_id", Kind: providercontract.FieldKindText},
			{Key: "decision.updated_at", Expr: "p.updated_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "decision.is_superseded", Expr: "p.is_superseded", Kind: providercontract.FieldKindBool},
		},
	}}
}

func recordRefCollectionExprFor(alias string, fieldKey string, linkType string) string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'record_ref:' || dst.record_id::text,
        'item_kind', 'record_ref',
        'display_text', dst.record_type || ':' || dst.record_id::text,
        'linked_record_id', dst.record_id::text
    ) ORDER BY dst.record_type ASC, dst.record_id ASC), '[]'::jsonb)
      FROM active_record_links_v1 rl
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = ` + alias + `.incident_id
       AND rl.src_record_id = ` + alias + `.record_id
       AND rl.link_type = '` + linkType + `'
       AND rl.field_key = '` + fieldKey + `'
       AND rl.deleted_at IS NULL)::text`
}
