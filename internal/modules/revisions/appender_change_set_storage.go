package revisions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (*Appender) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params AppendChangeSetParams) (uuid.UUID, error) {
	if params.ChangeSetID != nil {
		if _, err := tx.Exec(ctx, `
INSERT INTO change_sets (
    change_set_id,
    incident_id,
    actor_user_id,
    source,
    reason,
    client_txn_id,
    request_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, *params.ChangeSetID, params.IncidentID, params.ActorUserID, params.Source, params.Reason, params.ClientTxnID, params.RequestID, params.CreatedAt.UTC()); err != nil {
			return uuid.UUID{}, fmt.Errorf("append change set: %w", err)
		}
		return *params.ChangeSetID, nil
	}
	var changeSetID uuid.UUID
	if err := tx.QueryRow(ctx, `
INSERT INTO change_sets (
    incident_id,
    actor_user_id,
    source,
    reason,
    client_txn_id,
    request_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING change_set_id
`, params.IncidentID, params.ActorUserID, params.Source, params.Reason, params.ClientTxnID, params.RequestID, params.CreatedAt.UTC()).Scan(&changeSetID); err != nil {
		return uuid.UUID{}, fmt.Errorf("append change set: %w", err)
	}
	return changeSetID, nil
}
