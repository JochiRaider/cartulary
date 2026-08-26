package projection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	decisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectioncontract"
)

type DecisionSource struct{}

func NewDecisionSource() *DecisionSource { return &DecisionSource{} }

func (*DecisionSource) LoadDecisionProjectionInputTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (decisionprojection.DecisionProjectionInput, bool, error) {
	input, err := scanDecisionProjectionInput(tx.QueryRow(ctx, decisionProjectionSourceSQL+`
 WHERE d.record_id = $1
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return decisionprojection.DecisionProjectionInput{}, false, nil
	}
	if err != nil {
		return decisionprojection.DecisionProjectionInput{}, false, fmt.Errorf("load Decision projection input: %w", err)
	}
	return input, true, nil
}

func (*DecisionSource) ListDecisionProjectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) (decisionprojection.DecisionProjectionInputPage, error) {
	if limit < 1 || limit > 1000 {
		return decisionprojection.DecisionProjectionInputPage{}, fmt.Errorf("decision projection source page limit %d is outside 1..1000", limit)
	}
	rows, err := tx.Query(ctx, decisionProjectionSourceSQL+`
 WHERE d.incident_id = $1
   AND ($2::uuid IS NULL OR d.record_id > $2)
 ORDER BY d.record_id
 LIMIT $3
`, incidentID, afterRecordID, limit+1)
	if err != nil {
		return decisionprojection.DecisionProjectionInputPage{}, fmt.Errorf("list decision projection inputs: %w", err)
	}
	defer rows.Close()
	inputs := make([]decisionprojection.DecisionProjectionInput, 0, limit+1)
	for rows.Next() {
		input, scanErr := scanDecisionProjectionInput(rows)
		if scanErr != nil {
			return decisionprojection.DecisionProjectionInputPage{}, fmt.Errorf("scan Decision projection input: %w", scanErr)
		}
		inputs = append(inputs, input)
	}
	if err := rows.Err(); err != nil {
		return decisionprojection.DecisionProjectionInputPage{}, fmt.Errorf("iterate Decision projection inputs: %w", err)
	}
	page := decisionprojection.DecisionProjectionInputPage{Inputs: inputs}
	if len(inputs) > limit {
		page.Inputs = inputs[:limit]
		next := page.Inputs[len(page.Inputs)-1].RecordID
		page.NextRecordID = &next
	}
	return page, nil
}

func scanDecisionProjectionInput(scanner interface{ Scan(...any) error }) (decisionprojection.DecisionProjectionInput, error) {
	var (
		input               decisionprojection.DecisionProjectionInput
		summary             pgtype.Text
		ownerUserID         pgtype.UUID
		decisionType        pgtype.Text
		decidedAt           pgtype.Timestamptz
		rationale           pgtype.Text
		affectedRecordCount int32
		supersedesRecordID  pgtype.UUID
	)
	if err := scanner.Scan(
		&input.RecordID, &input.IncidentID, &input.RowVersion, &summary,
		&input.Status, &ownerUserID, &decisionType, &decidedAt, &rationale,
		&affectedRecordCount, &supersedesRecordID, &input.UpdatedAt,
		&input.IsSuperseded,
	); err != nil {
		return decisionprojection.DecisionProjectionInput{}, err
	}
	input.Summary = decisionTextPointer(summary)
	input.OwnerUserID = decisionUUIDPointer(ownerUserID)
	input.DecisionType = decisionTextPointer(decisionType)
	input.DecidedAt = decisionTimePointer(decidedAt)
	input.Rationale = decisionTextPointer(rationale)
	input.AffectedRecordCount = int(affectedRecordCount)
	input.SupersedesRecordID = decisionUUIDPointer(supersedesRecordID)
	input.UpdatedAt = input.UpdatedAt.UTC()
	return input, nil
}

func decisionTextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func decisionUUIDPointer(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	result := uuid.UUID(value.Bytes)
	return &result
}

func decisionTimePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time.UTC()
	return &result
}

const decisionProjectionSourceSQL = `
SELECT
    d.record_id,
    d.incident_id,
    r.row_version,
    d.summary,
    d.status,
    d.owner_user_id,
    d.decision_type,
    d.decided_at,
    d.rationale,
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 rl
         WHERE rl.incident_id = d.incident_id
           AND rl.src_record_id = d.record_id
           AND rl.link_type = 'references_record'
           AND rl.field_key = 'decision.affected_record_ids'
    ),
    supersedes.supersedes_record_id,
    d.updated_at,
    EXISTS (
        SELECT 1
          FROM active_record_links_v1 rl
          JOIN records src
            ON src.incident_id = rl.incident_id
           AND src.record_id = rl.src_record_id
           AND src.record_type = 'decision'
           AND src.deleted_at IS NULL
         WHERE rl.incident_id = d.incident_id
           AND rl.dst_record_id = d.record_id
           AND rl.link_type = 'supersedes'
    )
  FROM decisions d
  JOIN records r
    ON r.incident_id = d.incident_id
   AND r.record_id = d.record_id
   AND r.deleted_at IS NULL
  LEFT JOIN LATERAL (
        SELECT rl.dst_record_id AS supersedes_record_id
          FROM active_record_links_v1 rl
          JOIN records dst
            ON dst.incident_id = rl.incident_id
           AND dst.record_id = rl.dst_record_id
           AND dst.record_type = 'decision'
           AND dst.deleted_at IS NULL
         WHERE rl.incident_id = d.incident_id
           AND rl.src_record_id = d.record_id
           AND rl.link_type = 'supersedes'
         ORDER BY rl.created_at DESC, rl.record_link_id DESC
         LIMIT 1
  ) supersedes ON true`
