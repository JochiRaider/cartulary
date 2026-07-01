package artifacts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
)

func (s *Store) RefreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	if !IsArtifactBackedView(viewSchemaID) {
		return nil, fmt.Errorf("artifact import projection surface %q not mapped", viewSchemaID)
	}
	projector := projectionadapters.NewRowProjector(nil)
	if err := projector.RefreshRowTx(ctx, tx, viewSchemaID, recordID); err != nil {
		return nil, err
	}
	return projector.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}
