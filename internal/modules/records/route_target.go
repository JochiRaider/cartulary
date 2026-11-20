package records

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	recordenvelope "github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

type RouteTarget struct {
	IncidentID uuid.UUID
	RecordType string
	Deleted    bool
	RowVersion int64
}

type RouteTargetResolver struct {
	db dbport.DB
}

func NewRouteTargetResolver(db dbport.DB) *RouteTargetResolver {
	return &RouteTargetResolver{db: db}
}

func (r *RouteTargetResolver) Resolve(ctx context.Context, recordID uuid.UUID) (RouteTarget, error) {
	if r == nil || r.db == nil {
		return RouteTarget{}, errors.New("records: route target resolver database is required")
	}
	return routeTargetFromClassification(classifyTarget(ctx, r.db, recordID, TargetLockNone))
}

func (r *RouteTargetResolver) ResolveIncident(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	target, err := r.Resolve(ctx, recordID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return target.IncidentID, nil
}

func (r *RouteTargetResolver) ResolveTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (RouteTarget, error) {
	return routeTargetFromClassification(ClassifyTargetTx(ctx, tx, recordID, TargetLockNone))
}

// A deleted target still resolves, marked Deleted: the route decides incident
// visibility (REQ-01-704) before it reports the deleted state (REQ-01-705).
func routeTargetFromClassification(envelope recordenvelope.Envelope, err error) (RouteTarget, error) {
	deleted := errors.Is(err, ErrTargetDeleted)
	if err != nil && !deleted {
		return RouteTarget{}, err
	}
	return RouteTarget{
		IncidentID: envelope.IncidentID,
		RecordType: envelope.RecordType,
		Deleted:    deleted,
		RowVersion: envelope.RowVersion,
	}, nil
}
