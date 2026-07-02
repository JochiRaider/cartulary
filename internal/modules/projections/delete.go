package projections

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) DeleteRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	var query string
	switch viewSchemaID {
	case "cartulary.view.hosts.v1":
		query = "DELETE FROM host_grid_projection WHERE record_id = $1"
	case "cartulary.view.identities.v1":
		query = "DELETE FROM identity_grid_projection WHERE record_id = $1"
	case "cartulary.view.indicators.v1":
		query = "DELETE FROM indicator_grid_projection WHERE record_id = $1"
	default:
		return fmt.Errorf("projection delete surface %q not mapped", viewSchemaID)
	}
	if _, err := tx.Exec(ctx, query, recordID); err != nil {
		return fmt.Errorf("delete projection row %s: %w", viewSchemaID, err)
	}
	return nil
}
