package parties

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) RefreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	if viewSchemaID != ViewSchemaID {
		return nil, fmt.Errorf("party import projection surface %q not mapped", viewSchemaID)
	}
	projector := s.rowProjections()
	if err := projector.RefreshPartyTx(ctx, tx, recordID); err != nil {
		return nil, err
	}
	return projector.LoadPartyTx(ctx, tx, recordID)
}
