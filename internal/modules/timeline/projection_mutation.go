package timeline

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func (s *store) upsertProjectionTx(ctx context.Context, tx pgx.Tx, record projectedRecord) error {
	if s.projectionStore == nil {
		return errors.New("timeline projection port is required")
	}
	return s.projectionStore.ApplyTimelineMutationTx(ctx, tx, workbookprojection.ProjectionMutation{
		Kind:     workbookprojection.ProjectionMutationUpsert,
		RecordID: record.RecordID,
		Input:    record.ProjectionInput(),
	})
}
