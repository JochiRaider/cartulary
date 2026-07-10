package records

import (
	"context"
	"fmt"
	"time"

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
	var target RouteTarget
	var deletedAt *time.Time
	err := r.db.QueryRow(ctx, `
SELECT incident_id, record_type, row_version, deleted_at
  FROM records
 WHERE record_id = $1
`, recordID).Scan(&target.IncidentID, &target.RecordType, &target.RowVersion, &deletedAt)
	if err != nil {
		return RouteTarget{}, fmt.Errorf("resolve record route target: %w", err)
	}
	target.Deleted = deletedAt != nil
	return target, nil
}

func (r *RouteTargetResolver) ResolveIncident(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	target, err := r.Resolve(ctx, recordID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return target.IncidentID, nil
}

func (r *RouteTargetResolver) RecordIncident(ctx context.Context, recordID uuid.UUID, _ string) (uuid.UUID, error) {
	return r.ResolveIncident(ctx, recordID)
}

func (r *RouteTargetResolver) RecordRouteTarget(ctx context.Context, recordID uuid.UUID) (RouteTarget, error) {
	return r.Resolve(ctx, recordID)
}

func (r *RouteTargetResolver) ResolveTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (RouteTarget, error) {
	var target RouteTarget
	var deletedAt *time.Time
	err := tx.QueryRow(ctx, `
SELECT incident_id, record_type, row_version, deleted_at
  FROM records
 WHERE record_id = $1
`, recordID).Scan(&target.IncidentID, &target.RecordType, &target.RowVersion, &deletedAt)
	if err != nil {
		return RouteTarget{}, fmt.Errorf("resolve record route target: %w", err)
	}
	target.Deleted = deletedAt != nil
	return target, nil
}
