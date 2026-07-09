package artifacts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/riskrefs"
)

const HandoffOpenRiskRefsFieldKey = "handoff.open_risk_refs"

type RiskRefActionPayload struct {
	Actions []RiskRefAction
}

type RiskRefAction struct {
	Op             string
	ItemRef        string
	RiskRefText    string
	NormalizedText string
}

func ValidateHandoffRiskRefPayload(payload RiskRefActionPayload) error {
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_risk_ref":
			if strings.TrimSpace(action.RiskRefText) == "" || strings.TrimSpace(action.NormalizedText) == "" {
				return riskRefValidationError()
			}
		case "remove_risk_ref":
			if _, err := ParseRiskRefItemRef(action.ItemRef); err != nil {
				return riskRefValidationError()
			}
		default:
			return riskRefValidationError()
		}
	}
	return nil
}

func (s *Store) ApplyHandoffRiskRefPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, handoffRecordID uuid.UUID, actorID uuid.UUID, payload RiskRefActionPayload, now time.Time) (bool, error) {
	if err := ValidateHandoffRiskRefPayload(payload); err != nil {
		return false, err
	}
	if err := validateHandoffRiskRefOwnerTx(ctx, tx, incidentID, handoffRecordID); err != nil {
		return false, err
	}
	changed := false
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_risk_ref":
			applied, err := s.UpsertHandoffRiskRefTx(ctx, tx, incidentID, handoffRecordID, action.RiskRefText, action.NormalizedText, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_risk_ref":
			riskRefID, err := ParseRiskRefItemRef(action.ItemRef)
			if err != nil {
				return false, riskRefValidationError()
			}
			applied, err := s.TombstoneHandoffRiskRefTx(ctx, tx, incidentID, handoffRecordID, riskRefID, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		}
	}
	return changed, nil
}

func (s *Store) UpsertHandoffRiskRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, handoffRecordID uuid.UUID, text string, normalized string, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
INSERT INTO handoff_risk_refs (
    incident_id, handoff_record_id, risk_ref_text, normalized_risk_ref_text,
    created_by_user_id, created_at
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (handoff_record_id, normalized_risk_ref_text)
WHERE deleted_at IS NULL
DO NOTHING
`, incidentID, handoffRecordID, text, normalized, actorID, now.UTC())
	if err != nil {
		return false, fmt.Errorf("upsert handoff risk ref: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) TombstoneHandoffRiskRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, handoffRecordID uuid.UUID, riskRefID uuid.UUID, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
UPDATE handoff_risk_refs
   SET deleted_at = $5,
       deleted_by_user_id = $4
 WHERE incident_id = $1
   AND handoff_record_id = $2
   AND risk_ref_id = $3
   AND deleted_at IS NULL
`, incidentID, handoffRecordID, riskRefID, actorID, now.UTC())
	if err != nil {
		return false, fmt.Errorf("remove handoff risk ref: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, riskRefValidationError()
	}
	return true, nil
}

func validateHandoffRiskRefOwnerTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, handoffRecordID uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records r
      JOIN artifacts a
        ON a.incident_id = r.incident_id
       AND a.record_id = r.record_id
     WHERE r.incident_id = $1
       AND r.record_id = $2
       AND r.record_type = 'artifact'
       AND r.deleted_at IS NULL
       AND a.artifact_type = 'handoff'
)
`, incidentID, handoffRecordID).Scan(&exists); err != nil {
		return fmt.Errorf("validate handoff risk ref owner: %w", err)
	}
	if !exists {
		return riskRefValidationError()
	}
	return nil
}

func RiskRefItemRef(riskRefID uuid.UUID) string {
	return riskrefs.RiskRefItemRef(riskRefID)
}

func ParseRiskRefItemRef(itemRef string) (uuid.UUID, error) {
	return riskrefs.ParseRiskRefItemRef(itemRef)
}

func riskRefValidationError() *ValidationError {
	return &ValidationError{Field: HandoffOpenRiskRefsFieldKey, ReasonCode: "invalid_value"}
}
