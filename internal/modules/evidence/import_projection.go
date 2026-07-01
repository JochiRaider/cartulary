package evidence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
)

func (s *Store) RefreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	if viewSchemaID != evidenceViewSchemaID {
		return nil, fmt.Errorf("evidence import projection surface %q not mapped", viewSchemaID)
	}
	projectionStore := projections.NewStore(nil)
	if err := projectionStore.RefreshEvidenceTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return projectionStore.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}
