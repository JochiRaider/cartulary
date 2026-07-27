package workbookpreferences

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
)

type Bootstrap struct{}

func NewBootstrap() Bootstrap {
	return Bootstrap{}
}

func (Bootstrap) BootstrapIncidentPreferencesTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	now time.Time,
) error {
	q := sqlc.New(tx)
	timestamp := pgTimestamptz(now)
	if err := q.InsertIncidentWorkbookPreferencesBootstrap(ctx, sqlc.InsertIncidentWorkbookPreferencesBootstrapParams{
		IncidentID:      pgUUID(incidentID),
		CreatedAt:       timestamp,
		UpdatedByUserID: pgUUID(actorUserID),
	}); err != nil {
		return fmt.Errorf("insert incident workbook preferences: %w", err)
	}
	if err := q.InsertUserWorkbookPreferencesBootstrap(ctx, sqlc.InsertUserWorkbookPreferencesBootstrapParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(actorUserID),
		CreatedAt:  timestamp,
	}); err != nil {
		return fmt.Errorf("insert user workbook preferences: %w", err)
	}
	return nil
}
