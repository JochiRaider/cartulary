package incidents

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TransactionParticipant supplies the incident owner lock capability to
// consumer use cases without exposing incident SQL to those consumers.
type TransactionParticipant struct{}

func NewTransactionParticipant() *TransactionParticipant {
	return &TransactionParticipant{}
}

func (*TransactionParticipant) LockIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (bool, error) {
	var locked uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM incidents WHERE id = $1 FOR UPDATE`, incidentID).Scan(&locked)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("lock incident transaction participant: %w", err)
	}
	return true, nil
}
