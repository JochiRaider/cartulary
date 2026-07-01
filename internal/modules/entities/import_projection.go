package entities

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
)

const (
	partiesImportViewSchemaID     = "cartulary.view.parties.v1"
	assessmentsImportViewSchemaID = "cartulary.view.assessments.v1"
)

func (s *Store) RefreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	projectionStore := projections.NewStore(nil)
	switch viewSchemaID {
	case partiesImportViewSchemaID:
		if err := projectionStore.RefreshPartyTx(ctx, tx, recordID); err != nil {
			return nil, err
		}
	case assessmentsImportViewSchemaID:
		if err := projectionStore.RefreshAssessmentTx(ctx, tx, recordID); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("entity import projection surface %q not mapped", viewSchemaID)
	}
	return projectionStore.LoadRowTx(ctx, tx, viewSchemaID, recordID)
}
