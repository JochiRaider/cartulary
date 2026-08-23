package merge

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) findThirdPartyExactMatchConflictTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, identifierClass string, normalizedValue string, survivorRecordID uuid.UUID, loserRecordID uuid.UUID) (uuid.UUID, bool, error) {
	var recordID uuid.UUID
	err := tx.QueryRow(ctx, `
SELECT record_id
  FROM entity_active_identifier_claims
 WHERE incident_id = $1
   AND entity_type = $2
   AND identifier_type = $3
   AND normalized_value = $4
   AND record_id <> $5
   AND record_id <> $6
`, incidentID, entityType, identifierClass, normalizedValue, survivorRecordID, loserRecordID).Scan(&recordID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("lookup merge identifier claim conflict: %w", err)
	}
	return recordID, true, nil
}
