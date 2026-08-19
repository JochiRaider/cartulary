package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/modules/workbook/startup/bootstrapport"
)

// Writer implements Workbook's insert-only incident bootstrap capability.
type Writer struct{}

func NewWriter() *Writer {
	return &Writer{}
}

func (*Writer) InsertInitialTx(
	ctx context.Context,
	tx pgx.Tx,
	input bootstrapport.InitialPreferenceInput,
) error {
	q := sqlc.New(tx)
	timestamp := pgTimestamptz(input.CommitTimestamp)
	if err := q.InsertIncidentWorkbookPreferencesBootstrap(ctx, sqlc.InsertIncidentWorkbookPreferencesBootstrapParams{
		IncidentID:      pgUUID(input.IncidentID),
		CreatedAt:       timestamp,
		UpdatedByUserID: pgUUID(input.UserID),
	}); err != nil {
		return fmt.Errorf("insert initial incident workbook preferences: %w", err)
	}
	if err := q.InsertUserWorkbookPreferencesBootstrap(ctx, sqlc.InsertUserWorkbookPreferencesBootstrapParams{
		IncidentID: pgUUID(input.IncidentID),
		UserID:     pgUUID(input.UserID),
		CreatedAt:  timestamp,
	}); err != nil {
		return fmt.Errorf("insert initial user workbook preferences: %w", err)
	}
	return nil
}
