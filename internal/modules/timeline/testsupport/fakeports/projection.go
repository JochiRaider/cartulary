package fakeports

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

// Projection decorates Timeline's projection port for transaction-boundary
// tests. FailApply runs after the delegate applies the mutation so callers can
// prove that source, peer-owner, projection, and history writes roll back
// together.
type Projection struct {
	Delegate  timeline.ProjectionPort
	FailApply func(workbookprojection.ProjectionMutation) error
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

func (p Projection) RefreshHostTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if p.Delegate == nil {
		return nil
	}
	return p.Delegate.RefreshHostTx(ctx, tx, recordID)
}

func (p Projection) RefreshIdentityTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	if p.Delegate == nil {
		return nil
	}
	return p.Delegate.RefreshIdentityTx(ctx, tx, recordID)
}
