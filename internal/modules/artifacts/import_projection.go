package artifacts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
)

func (s *Store) RefreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	if !IsArtifactBackedView(viewSchemaID) {
		return nil, fmt.Errorf("artifact import projection surface %q not mapped", viewSchemaID)
	}
	rows := projections.NewArtifactRows(nil, artifactprojection.QuerySurfaces()...)
	if err := rows.RefreshTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return rows.LoadTx(ctx, tx, viewSchemaID, recordID)
}
