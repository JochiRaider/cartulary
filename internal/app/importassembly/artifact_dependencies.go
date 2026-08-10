package importassembly

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
)

type artifactRecordInserter struct {
	store *records.Store
}

func (a artifactRecordInserter) InsertTx(ctx context.Context, tx pgx.Tx, params records.InsertParams) (uuid.UUID, error) {
	return a.store.InsertTx(ctx, tx, params)
}

type artifactActiveUserLookup struct{}

func (artifactActiveUserLookup) IsActiveUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (bool, error) {
	var active bool
	err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND is_active = true)`, userID).Scan(&active)
	return active, err
}

type artifactProjectionAdapter struct {
	rows artifactprojection.Rows
}

func (a artifactProjectionAdapter) RefreshArtifactTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return a.rows.RefreshArtifactTx(ctx, tx, recordID)
}

func (a artifactProjectionAdapter) LoadArtifactTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	return a.rows.LoadArtifactTx(ctx, tx, viewSchemaID, recordID)
}
