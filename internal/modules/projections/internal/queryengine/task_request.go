package queryengine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	taskprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionports"
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
SELECT t.record_id, to_jsonb(t) - 'incident_id'
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
		fact.Value = map[string]any{}
		if err := json.Unmarshal(raw, &fact.Value); err != nil {
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
		FromSQL:      "FROM task_request_grid_projection p JOIN records r ON r.record_id = p.record_id",
		RecordExpr:   "p.record_id",
		IncidentExpr: "p.incident_id",
		Fields: []Field{
			{Key: "task.title", Expr: "p.title", Kind: FieldKindText},
			{Key: "task.status", Expr: "p.status", Kind: FieldKindText},
			{Key: "task.owner_user_id", Expr: "p.owner_user_id", Kind: FieldKindText},
			{Key: "task.priority", Expr: "p.priority", Kind: FieldKindText},
			{Key: "task.task_kind", Expr: "p.task_kind", Kind: FieldKindText},
			{Key: "task.workstream", Expr: "p.workstream", Kind: FieldKindText},
			{Key: "task.due_at", Expr: "p.due_at", Kind: FieldKindTimestamp},
			{Key: "task.requester_party_text", Expr: "p.requester_party_text", Kind: FieldKindText},
			{Key: "task.requester_party_id", Expr: "p.requester_party_id", Kind: FieldKindText},
			{Key: "task.blocked_reason", Expr: "p.blocked_reason", Kind: FieldKindText},
			{Key: "task.completed_at", Expr: "p.completed_at", Kind: FieldKindTimestamp},
			{Key: "task.external_ticket_ref", Expr: "p.external_ticket_ref", Kind: FieldKindText},
			{Key: "task.closure_summary", Expr: "p.closure_summary", Kind: FieldKindText},
			{Key: "task.linked_record_ids", Expr: taskRecordRefCollectionExpr("p", "task.linked_record_ids", "references_record"), Kind: FieldKindCollection},
			{Key: "task.decision_record_id", Expr: "p.decision_record_id", Kind: FieldKindText},
			{Key: "task.linked_record_count", Expr: "p.linked_record_count", Kind: FieldKindNumber},
			{Key: "task.updated_at", Expr: "p.updated_at", Kind: FieldKindTimestamp},
			{Key: "task.no_owner", Expr: "p.no_owner", Kind: FieldKindBool},
		},
	}}
}

func taskRecordRefCollectionExpr(alias string, fieldKey string, linkType string) string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', ` + recordRefItemRefSQL("dst.record_id") + `,
        'item_kind', 'record_ref',
        'display_text', dst.record_type || ':' || dst.record_id::text,
        'linked_record_id', dst.record_id::text
    ) ORDER BY dst.record_type ASC, dst.record_id ASC), '[]'::jsonb)
      FROM ` + activeRecordLinksAlias("rl") + `
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = ` + alias + `.incident_id
       AND rl.src_record_id = ` + alias + `.record_id
       AND rl.link_type = '` + linkType + `'
       AND rl.field_key = '` + fieldKey + `')::text`
}
