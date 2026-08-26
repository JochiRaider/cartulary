package rollback

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
)

type DecisionProvider struct{}

var _ rollbackcontract.RowSourceProvider = DecisionProvider{}

func NewDecisionProvider() DecisionProvider { return DecisionProvider{} }

func (DecisionProvider) ValidateRollbackValue(value map[string]any) error {
	source, ok := decisionSourceForRollbackValue(value)
	if !ok || !validDecisionSource(source) {
		return rollbackcontract.ErrTargetNotReversible
	}
	return nil
}

type decisionMachine struct {
	incidentID         uuid.UUID
	status             string
	ownerUserID        *uuid.UUID
	decidedAt          *time.Time
	incomingSupersedes int
	outgoingSupersedes int
}

func (DecisionProvider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := decisionSourceForRollbackValue(request.RetainedValue)
	if !ok || !validDecisionSource(source) {
		return rollbackcontract.ErrTargetNotReversible
	}
	var machine decisionMachine
	if err := tx.QueryRow(ctx, `
SELECT d.incident_id, d.status, d.owner_user_id, d.decided_at,
       (SELECT COUNT(*) FROM active_record_links_v1 rl WHERE rl.incident_id = d.incident_id AND rl.dst_record_id = d.record_id AND rl.link_type = 'supersedes'),
       (SELECT COUNT(*) FROM active_record_links_v1 rl WHERE rl.incident_id = d.incident_id AND rl.src_record_id = d.record_id AND rl.link_type = 'supersedes')
  FROM decisions d
 WHERE d.record_id = $1
`, request.RecordID).Scan(&machine.incidentID, &machine.status, &machine.ownerUserID, &machine.decidedAt, &machine.incomingSupersedes, &machine.outgoingSupersedes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rollbackcontract.ErrTargetNotFound
		}
		return err
	}
	if raw, present := source["status"]; present {
		machine.status = raw.(string)
	}
	if value, present, err := nullableUUID(source, "owner_user_id"); err != nil {
		return rollbackcontract.ErrTargetNotReversible
	} else if present {
		machine.ownerUserID = value
	}
	if value, present, err := nullableTime(source, "decided_at"); err != nil {
		return rollbackcontract.ErrTargetNotReversible
	} else if present {
		machine.decidedAt = value
	}
	if !validDecisionMachine(machine) {
		return rollbackcontract.ErrTargetNotReversible
	}
	values, err := typedValues(source, []fieldSpec{
		{"summary", fieldText}, {"status", fieldText}, {"owner_user_id", fieldUUID},
		{"decision_type", fieldText}, {"decided_at", fieldTime}, {"rationale", fieldText},
	})
	if err != nil {
		return rollbackcontract.ErrTargetNotReversible
	}
	_, err = tx.Exec(ctx, `
UPDATE decisions
   SET summary = CASE WHEN $2 THEN $3::text ELSE summary END,
       status = CASE WHEN $4 THEN $5::text ELSE status END,
       owner_user_id = CASE WHEN $6 THEN $7::uuid ELSE owner_user_id END,
       decision_type = CASE WHEN $8 THEN $9::text ELSE decision_type END,
       decided_at = CASE WHEN $10 THEN $11::timestamptz ELSE decided_at END,
       rationale = CASE WHEN $12 THEN $13::text ELSE rationale END,
       updated_at = $14
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	return err
}

func decisionSourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	return sourceForRollbackValue(value)
}

func validDecisionSource(source map[string]any) bool {
	if raw, present := source["summary"]; present && !nonEmptyText(raw) {
		return false
	}
	if raw, present := source["status"]; present && !policyText(raw, policy.ValidDecisionStatus) {
		return false
	}
	if raw, present := source["decision_type"]; present && !policyText(raw, policy.ValidDecisionType) {
		return false
	}
	return validTypedFields(source, []fieldSpec{{"owner_user_id", fieldUUID}, {"decided_at", fieldTime}})
}

func validDecisionMachine(state decisionMachine) bool {
	return policy.ValidateDecisionMachineState(policy.DecisionMachineState{
		Status: state.status, OwnerUserID: nullableSQLUUID(state.ownerUserID),
		DecidedAt: nullableSQLTime(state.decidedAt), IncomingSupersedes: state.incomingSupersedes,
		OutgoingSupersedes: state.outgoingSupersedes,
	}) == nil
}
