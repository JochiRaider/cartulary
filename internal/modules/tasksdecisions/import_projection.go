package tasksdecisions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const taskRequestsImportViewSchemaID = "cartulary.view.task_requests.v1"

func (s *Store) RefreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	rows := newTaskDecisionProjectionRows(nil)
	switch viewSchemaID {
	case taskRequestsImportViewSchemaID:
		if err := rows.RefreshTaskRequestTx(ctx, tx, recordID); err != nil {
			return nil, err
		}
	case DecisionsViewSchemaID:
		if err := rows.RefreshDecisionTx(ctx, tx, recordID); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("tasks/decisions import projection surface %q not mapped", viewSchemaID)
	}
	return rows.LoadTx(ctx, tx, viewSchemaID, recordID)
}

func newTaskDecisionProjectionRows(pool postgres.DB) *projections.TaskDecisionRows {
	surfaces := taskdecisionprojection.TaskRequestQuerySurfaces()
	surfaces = append(surfaces, taskdecisionprojection.DecisionQuerySurfaces()...)
	return projections.NewTaskDecisionRows(pool, surfaces...)
}
