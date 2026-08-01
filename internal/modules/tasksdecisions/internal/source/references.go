package source

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
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
	incidentID uuid.UUID,
	recordID uuid.UUID,
	recordType string,
	field string,
) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND record_type = $3
       AND deleted_at IS NULL
)`, incidentID, recordID, recordType).Scan(&exists); err != nil {
		return &Error{Operation: "validate record reference", Err: err}
	}
	if !exists {
		return &policy.ValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}
