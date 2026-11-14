package source

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/referencefacts"
)

func ValidateMemberUserTx(ctx context.Context, tx pgx.Tx, incidentID, userID uuid.UUID, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM users u
    JOIN incident_memberships m ON m.user_id = u.id
   WHERE u.id = $1
     AND u.is_active = true
     AND m.incident_id = $2
)`, userID, incidentID).Scan(&exists); err != nil {
		return &Error{Operation: "validate member user reference", Err: err}
	}
	if !exists {
		return &policy.ValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}

func ValidateTargetRecordTx(
	ctx context.Context,
	tx pgx.Tx,
	targets referencefacts.Capability,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	recordType string,
	field string,
) error {
	valid, err := LockReferenceTargetTx(ctx, tx, targets, incidentID, recordID, recordType, false)
	if err != nil {
		return &Error{Operation: "validate record reference", Err: err}
	}
	if !valid {
		return &policy.ValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}

// LockReferenceTargetTx validates the current typed Task reference and retains
// its envelope/source locks in the caller transaction. Ordinary writes wait;
// destructive restore uses nowait and owns its safe contention error mapping.
func LockReferenceTargetTx(ctx context.Context, tx pgx.Tx, targets referencefacts.Capability, incidentID, recordID uuid.UUID, recordType string, nowait bool) (bool, error) {
	if recordType != "party" && recordType != "decision" {
		return false, nil
	}
	if targets == nil {
		return false, errors.New("tasks/decisions: reference target capability is required")
	}
	envelope, found, err := targets.LockEnvelopeFactTx(ctx, tx, recordID, recordType, nowait)
	if err != nil {
		return false, err
	}
	if !found || envelope.IncidentID != incidentID || envelope.RecordType != recordType || envelope.Deleted {
		return false, nil
	}
	if recordType == "party" {
		return targets.LockPartySourceTx(ctx, tx, incidentID, recordID, nowait)
	}
	query := "SELECT record_id FROM decisions WHERE incident_id=$1 AND record_id=$2 FOR SHARE"
	if nowait {
		query += " NOWAIT"
	}
	var id uuid.UUID
	err = tx.QueryRow(ctx, query, incidentID, recordID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
