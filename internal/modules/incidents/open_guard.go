package incidents

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func EnsureIncidentOpenTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	var status string
	err := tx.QueryRow(ctx, `
SELECT status
  FROM incidents
 WHERE id = $1
 FOR UPDATE
`, incidentID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrIncidentNotFound
		}
		return fmt.Errorf("ensure incident open: %w", err)
	}
	if status == "closed" {
		return ErrIncidentClosed
	}
	return nil
}
