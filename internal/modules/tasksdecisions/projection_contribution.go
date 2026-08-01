package tasksdecisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionprovider"
)

// ProjectionContribution is the complete source-owned projection boundary for
// task requests and decisions. Generic projection coordinators receive only
// this contract; they do not import the source owner's implementation leaves.
type ProjectionContribution struct {
	source              projections.TaskDecisionSource
	taskRequestSurfaces []providercontract.QuerySurface
	decisionSurfaces    []providercontract.QuerySurface
}

func NewProjectionContribution() ProjectionContribution {
	return ProjectionContribution{
		source:              taskDecisionProjectionSource{},
		taskRequestSurfaces: taskdecisionprojection.TaskRequestQuerySurfaces(),
		decisionSurfaces:    taskdecisionprojection.DecisionQuerySurfaces(),
	}
}

func (c ProjectionContribution) Source() projections.TaskDecisionSource {
	return c.source
}

func (c ProjectionContribution) TaskRequestQuerySurfaces() []providercontract.QuerySurface {
	return append([]providercontract.QuerySurface(nil), c.taskRequestSurfaces...)
}

func (c ProjectionContribution) DecisionQuerySurfaces() []providercontract.QuerySurface {
	return append([]providercontract.QuerySurface(nil), c.decisionSurfaces...)
}

func (c ProjectionContribution) QuerySurfaces() []providercontract.QuerySurface {
	surfaces := c.TaskRequestQuerySurfaces()
	return append(surfaces, c.DecisionQuerySurfaces()...)
}

type taskDecisionProjectionSource struct{}

func (taskDecisionProjectionSource) RefreshTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return taskdecisionprojection.RefreshTaskRequestTx(ctx, tx, recordID)
}

func (taskDecisionProjectionSource) RefreshDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	return taskdecisionprojection.RefreshDecisionTx(ctx, tx, recordID)
}

func (taskDecisionProjectionSource) RebuildIncidentTaskRequestsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return taskdecisionprojection.RebuildIncidentTaskRequestsTx(ctx, tx, incidentID)
}

func (taskDecisionProjectionSource) RebuildIncidentDecisionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return taskdecisionprojection.RebuildIncidentDecisionsTx(ctx, tx, incidentID)
}
