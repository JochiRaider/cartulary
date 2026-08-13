package evidence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *SourceMutationService) RefreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	if viewSchemaID != ViewSchemaID {
		return nil, fmt.Errorf("evidence import projection surface %q not mapped", viewSchemaID)
	}
	return s.source.refreshRowTx(ctx, tx, recordID)
}
