package queryengine

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	taskprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionports"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

const taskRequestsViewSchemaID = "cartulary.view.task_requests.v1"

type TaskReader struct{}

func NewTaskReader() *TaskReader { return &TaskReader{} }

func (*TaskReader) CollectTaskDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]taskprojection.TaskDerivedFact, error) {
	rows, err := tx.Query(ctx, `
SELECT t.record_id, jsonb_build_object(
       'record_id', t.record_id,
       'row_version', t.row_version,
       'title', t.title,
       'status', t.status,
       'owner_user_id', t.owner_user_id,
       'priority', t.priority,
       'task_kind', t.task_kind,
       'workstream', t.workstream,
       'due_at', t.due_at,
       'requester_party_text', t.requester_party_text,
       'requester_party_id', t.requester_party_id,
       'blocked_reason', t.blocked_reason,
       'completed_at', t.completed_at,
       'external_ticket_ref', t.external_ticket_ref,
       'closure_summary', t.closure_summary,
       'decision_record_id', t.decision_record_id,
       'linked_record_count', t.linked_record_count,
       'updated_at', t.updated_at,
       'no_owner', t.no_owner
   )
  FROM task_request_grid_projection t
  JOIN records r
    ON r.incident_id = t.incident_id
   AND r.record_id = t.record_id
   AND r.deleted_at IS NULL
 WHERE t.incident_id = $1
 ORDER BY t.record_id
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("collect Task-request projection facts: %w", err)
	}
	defer rows.Close()
	facts := make([]taskprojection.TaskDerivedFact, 0)
	for rows.Next() {
		var fact taskprojection.TaskDerivedFact
		var raw []byte
		if err := rows.Scan(&fact.RecordID, &raw); err != nil {
			return nil, fmt.Errorf("scan Task-request projection fact: %w", err)
		}
		fact.Value, err = decodeReportingFactObject(raw)
		if err != nil {
			return nil, fmt.Errorf("decode Task-request projection fact: %w", err)
		}
		facts = append(facts, fact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Task-request projection facts: %w", err)
	}
	return facts, nil
}

func TaskRequestPlans() []Surface {
	return []Surface{{
		ViewSchemaID: taskRequestsViewSchemaID,
		FromSQL:      sqlshape.Literal("FROM task_request_grid_projection p JOIN records r ON r.record_id = p.record_id"),
		RecordExpr:   sqlshape.Literal("p.record_id"),
		IncidentExpr: sqlshape.Literal("p.incident_id"),
		Fields: []Field{
			{Key: "task.title", Expr: sqlshape.Literal("p.title"), Kind: FieldKindText},
			{Key: "task.status", Expr: sqlshape.Literal("p.status"), Kind: FieldKindText},
			{Key: "task.owner_user_id", Expr: sqlshape.Literal("p.owner_user_id"), Kind: FieldKindText},
			{Key: "task.priority", Expr: sqlshape.Literal("p.priority"), Kind: FieldKindText},
			{Key: "task.task_kind", Expr: sqlshape.Literal("p.task_kind"), Kind: FieldKindText},
			{Key: "task.workstream", Expr: sqlshape.Literal("p.workstream"), Kind: FieldKindText},
			{Key: "task.due_at", Expr: sqlshape.Literal("p.due_at"), Kind: FieldKindTimestamp},
			{Key: "task.requester_party_text", Expr: sqlshape.Literal("p.requester_party_text"), Kind: FieldKindText},
			{Key: "task.requester_party_id", Expr: sqlshape.Literal("p.requester_party_id"), Kind: FieldKindText},
			{Key: "task.blocked_reason", Expr: sqlshape.Literal("p.blocked_reason"), Kind: FieldKindText},
			{Key: "task.completed_at", Expr: sqlshape.Literal("p.completed_at"), Kind: FieldKindTimestamp},
			{Key: "task.external_ticket_ref", Expr: sqlshape.Literal("p.external_ticket_ref"), Kind: FieldKindText},
			{Key: "task.closure_summary", Expr: sqlshape.Literal("p.closure_summary"), Kind: FieldKindText},
			{Key: "task.linked_record_ids", Expr: taskRecordRefCollectionExpr("task.linked_record_ids", "references_record"), Kind: FieldKindCollection},
			{Key: "task.decision_record_id", Expr: sqlshape.Literal("p.decision_record_id"), Kind: FieldKindText},
			{Key: "task.linked_record_count", Expr: sqlshape.Literal("p.linked_record_count"), Kind: FieldKindNumber},
			{Key: "task.updated_at", Expr: sqlshape.Literal("p.updated_at"), Kind: FieldKindTimestamp},
			{Key: "task.no_owner", Expr: sqlshape.Literal("p.no_owner"), Kind: FieldKindBool},
		},
	}}
}

func taskRecordRefCollectionExpr(fieldKey string, linkType string) sqlshape.Fragment {
	return sqlshape.Format(`(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', %s,
        'item_kind', 'record_ref',
        'display_text', dst.record_type || ':' || dst.record_id::text,
        'linked_record_id', dst.record_id::text
    ) ORDER BY dst.record_type ASC, dst.record_id ASC), '[]'::jsonb)
      FROM active_record_links_v1 rl
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = p.incident_id
       AND rl.src_record_id = p.record_id
       AND rl.link_type = %s
       AND rl.field_key = %s)::text`, recordRefItemRefSQL(sqlshape.Literal("dst.record_id")), sqlshape.Quote(linkType), sqlshape.Quote(fieldKey))
}
