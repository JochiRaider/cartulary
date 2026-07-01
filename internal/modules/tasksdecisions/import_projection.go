package tasksdecisions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
)

const taskRequestsImportViewSchemaID = "cartulary.view.task_requests.v1"

func (s *Store) RefreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	projector := projectionadapters.NewRowProjector(nil)
	switch viewSchemaID {
	case taskRequestsImportViewSchemaID:
		if err := projector.RefreshRowTx(ctx, tx, projectionadapters.TaskRequestsViewSchemaID, recordID); err != nil {
			return nil, err
		}
	case decisionsViewSchemaID:
		if err := projector.RefreshRowTx(ctx, tx, projectionadapters.DecisionsViewSchemaID, recordID); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("tasks/decisions import projection surface %q not mapped", viewSchemaID)
	}
	return projector.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}
