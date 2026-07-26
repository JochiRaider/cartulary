package evidence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
)

func (s *Store) RefreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	if viewSchemaID != ViewSchemaID {
		return nil, fmt.Errorf("evidence import projection surface %q not mapped", viewSchemaID)
	}
	rows := projections.NewEvidenceRows(nil, evidenceprojection.QuerySurfaces()...)
	if err := rows.RefreshTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return rows.LoadTx(ctx, tx, recordID)
}
