package indicators

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/providers/projection"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

// ProjectionContribution is the complete Indicator-owned projection boundary.
// Generic projection coordinators receive this contract without importing the
// source owner's implementation package.
type ProjectionContribution struct {
	source   projections.IndicatorSource
	surfaces []providercontract.QuerySurface
}

func NewProjectionContribution() ProjectionContribution {
	return ProjectionContribution{
		source:   indicatorProjectionSource{},
		surfaces: indicatorprojection.QuerySurfaces(),
	}
}

func (c ProjectionContribution) Source() projections.IndicatorSource {
	return c.source
}

func (c ProjectionContribution) QuerySurfaces() []providercontract.QuerySurface {
	return append([]providercontract.QuerySurface(nil), c.surfaces...)
}

type indicatorProjectionSource struct{}

func (indicatorProjectionSource) RefreshIndicatorTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return indicatorprojection.RefreshIndicatorTx(ctx, tx, recordID)
}

func (indicatorProjectionSource) RebuildIncidentIndicatorsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return indicatorprojection.RebuildIncidentIndicatorsTx(ctx, tx, incidentID)
}
