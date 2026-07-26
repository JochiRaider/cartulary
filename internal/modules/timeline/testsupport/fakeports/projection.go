package fakeports

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// Projection decorates Timeline's projection port for transaction-boundary
// tests. FailApply runs after the delegate applies the mutation so callers can
// prove that source, peer-owner, projection, and history writes roll back
// together.
type Projection struct {
	Delegate  timeline.ProjectionPort
	FailApply func(workbookprojection.ProjectionMutation) error
}

func WithFailingProjection(base func(postgres.DB) timeline.Dependencies, fail func(workbookprojection.ProjectionMutation) error) func(postgres.DB) timeline.Dependencies {
	return func(pool postgres.DB) timeline.Dependencies {
		dependencies := base(pool)
		dependencies.Projections = Projection{
			Delegate:  dependencies.Projections,
			FailApply: fail,
		}
		return dependencies
	}
}

func (p Projection) ApplyTimelineMutationTx(ctx context.Context, tx pgx.Tx, mutation workbookprojection.ProjectionMutation) error {
	if p.Delegate != nil {
		if err := p.Delegate.ApplyTimelineMutationTx(ctx, tx, mutation); err != nil {
			return err
		}
	}
	if p.FailApply != nil {
		return p.FailApply(mutation)
	}
	return nil
}

func (p Projection) RebuildIncidentHostsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if p.Delegate == nil {
		return nil
	}
	return p.Delegate.RebuildIncidentHostsTx(ctx, tx, incidentID)
}

func (p Projection) RebuildIncidentIdentitiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	if p.Delegate == nil {
		return nil
	}
	return p.Delegate.RebuildIncidentIdentitiesTx(ctx, tx, incidentID)
}
