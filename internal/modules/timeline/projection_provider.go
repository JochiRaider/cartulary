package timeline

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

type ProjectionSource struct {
	source *workbookprojection.Source
}

func NewProjectionSource(records RecordPort, collections CollectionFactPort) *ProjectionSource {
	return &ProjectionSource{
		source: workbookprojection.NewSource(records, collections),
	}
}

func (s *ProjectionSource) BuildProjectionMutationTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (workbookprojection.ProjectionMutation, error) {
	if s == nil {
		return workbookprojection.ProjectionMutation{}, errors.New("timeline projection source is required")
	}
	return s.source.BuildProjectionMutationTx(ctx, tx, recordID)
}

func (s *ProjectionSource) ListProjectionInputsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, afterRecordID *uuid.UUID, limit int) (workbookprojection.ProjectionInputPage, error) {
	if s == nil {
		return workbookprojection.ProjectionInputPage{}, errors.New("timeline projection source is required")
	}
	return s.source.ListProjectionInputsTx(ctx, tx, incidentID, afterRecordID, limit)
}

func ProjectionQuerySurfaces() []providercontract.QuerySurface {
	return workbookprojection.QuerySurfaces()
}
