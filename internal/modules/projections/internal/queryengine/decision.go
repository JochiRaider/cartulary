package queryengine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	decisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
)

const decisionsViewSchemaID = "cartulary.view.decisions.v1"

type DecisionReader struct{}

func NewDecisionReader() *DecisionReader { return &DecisionReader{} }

func (*DecisionReader) CollectDecisionDerivedFactsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
) ([]decisionprojection.DecisionDerivedFact, error) {
	rows, err := tx.Query(ctx, `
SELECT d.record_id, to_jsonb(d) - 'incident_id'
  FROM decision_grid_projection d
  JOIN records r
    ON r.incident_id = d.incident_id
   AND r.record_id = d.record_id
   AND r.deleted_at IS NULL
 WHERE d.incident_id = $1
 ORDER BY d.record_id
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("collect Decision projection facts: %w", err)
	}
	defer rows.Close()
	facts := make([]decisionprojection.DecisionDerivedFact, 0)
	for rows.Next() {
		var fact decisionprojection.DecisionDerivedFact
		var raw []byte
		if err := rows.Scan(&fact.RecordID, &raw); err != nil {
			return nil, fmt.Errorf("scan Decision projection fact: %w", err)
		}
		fact.Value = map[string]any{}
		if err := json.Unmarshal(raw, &fact.Value); err != nil {
			return nil, fmt.Errorf("decode Decision projection fact: %w", err)
		}
		facts = append(facts, fact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate Decision projection facts: %w", err)
	}
	return facts, nil
}

func DecisionPlans() []Surface {
	return []Surface{{
		ViewSchemaID: decisionsViewSchemaID,
		FromSQL:      "FROM decision_grid_projection p JOIN records r ON r.record_id = p.record_id",
		RecordExpr:   "p.record_id",
		IncidentExpr: "p.incident_id",
		Fields: []Field{
			{Key: "decision.summary", Expr: "p.summary", Kind: FieldKindText},
			{Key: "decision.status", Expr: "p.status", Kind: FieldKindText},
			{Key: "decision.owner_user_id", Expr: "p.owner_user_id", Kind: FieldKindText},
			{Key: "decision.decision_type", Expr: "p.decision_type", Kind: FieldKindText},
			{Key: "decision.decided_at", Expr: "p.decided_at", Kind: FieldKindTimestamp},
			{Key: "decision.rationale", Expr: "p.rationale", Kind: FieldKindText},
			{Key: "decision.support_refs", Expr: taskRecordRefCollectionExpr("p", "decision.support_refs", "supported_by"), Kind: FieldKindCollection},
			{Key: "decision.affected_record_ids", Expr: taskRecordRefCollectionExpr("p", "decision.affected_record_ids", "references_record"), Kind: FieldKindCollection},
			{Key: "decision.affected_record_count", Expr: "p.affected_record_count", Kind: FieldKindNumber},
			{Key: "decision.supersedes_record_id", Expr: "p.supersedes_record_id", Kind: FieldKindText},
			{Key: "decision.updated_at", Expr: "p.updated_at", Kind: FieldKindTimestamp},
			{Key: "decision.is_superseded", Expr: "p.is_superseded", Kind: FieldKindBool},
		},
	}}
}
