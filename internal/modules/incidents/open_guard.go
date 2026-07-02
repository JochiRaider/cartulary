package incidents

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
)

func ensureIncidentOpenTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	status, err := sqlc.New(tx).EnsureIncidentOpenForUpdate(ctx, pgUUID(incidentID))
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
