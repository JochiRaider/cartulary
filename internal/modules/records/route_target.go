package records

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type RouteTarget struct {
	IncidentID uuid.UUID
	RecordType string
	Deleted    bool
	RowVersion int64
}

type RouteTargetResolver struct {
	db postgres.DB
}

func NewRouteTargetResolver(db postgres.DB) *RouteTargetResolver {
	return &RouteTargetResolver{db: db}
}

func (r *RouteTargetResolver) Resolve(ctx context.Context, recordID uuid.UUID) (RouteTarget, error) {
	envelope, err := NewStore(r.db).LoadEnvelope(ctx, recordID)
	if err != nil {
		return RouteTarget{}, err
	}
	return routeTargetFromEnvelope(envelope), nil
}

func (r *RouteTargetResolver) ResolveIncident(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	target, err := r.Resolve(ctx, recordID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return target.IncidentID, nil
}

func (r *RouteTargetResolver) ResolveTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (RouteTarget, error) {
	envelope, err := NewStore().LoadEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		return RouteTarget{}, err
	}
	return routeTargetFromEnvelope(envelope), nil
}

func routeTargetFromEnvelope(envelope Envelope) RouteTarget {
	return RouteTarget{
		IncidentID: envelope.IncidentID,
		RecordType: envelope.RecordType,
		Deleted:    envelope.DeletedAt != nil,
		RowVersion: envelope.RowVersion,
	}
}
