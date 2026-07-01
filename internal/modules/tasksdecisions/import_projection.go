package tasksdecisions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
)

const taskRequestsImportViewSchemaID = "cartulary.view.task_requests.v1"

func (s *Store) RefreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	projectionStore := projections.NewStore(nil)
	switch viewSchemaID {
	case taskRequestsImportViewSchemaID:
		if err := projectionStore.RefreshTaskRequestTx(ctx, tx, recordID); err != nil {
			return nil, err
		}
	case decisionsViewSchemaID:
		if err := projectionStore.RefreshDecisionTx(ctx, tx, recordID); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("tasks/decisions import projection surface %q not mapped", viewSchemaID)
	}
	return projectionStore.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}
