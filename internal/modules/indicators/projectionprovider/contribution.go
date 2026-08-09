package projectionprovider

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/providers/projection"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
)

func NewContribution() (workbookprojection.Contribution, error) {
	return workbookprojection.NewContribution(source{})
}

type source struct{}

func (source) LoadProjectionInputTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (workbookprojection.ProjectionInput, bool, error) {
	return indicatorprojection.LoadProjectionInputTx(ctx, tx, recordID)
}

func (source) ListProjectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) (workbookprojection.ProjectionInputPage, error) {
	return indicatorprojection.ListProjectionInputsTx(ctx, tx, incidentID, afterRecordID, limit)
}
