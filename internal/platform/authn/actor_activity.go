package authn

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// IsActiveUserTx checks only actor activity and retains a shared row lock until
// the caller's transaction ends. It neither loads credentials nor grants any
// incident permission. Shared locks allow concurrent authorized publications
// while ordering them against account deactivation.
func (s *Store) IsActiveUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (bool, error) {
	if s == nil || tx == nil || userID == uuid.Nil {
		return false, ErrNotFound
	}
	var active bool
	err := tx.QueryRow(ctx, `SELECT is_active FROM users WHERE id=$1 FOR SHARE`, userID).Scan(&active)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return active, err
}
