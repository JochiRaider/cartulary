package incidents

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WorkbookBootstrapPort interface {
	BootstrapIncidentCreatePreferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, actorUserID uuid.UUID, now time.Time) error
}

type CollaborationSessionPort interface {
	NotifyIncidentClosed(ctx context.Context, incidentID uuid.UUID)
	NotifyIncidentMembershipRevoked(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID)
}

type StoreOptions struct {
	WorkbookBootstrap WorkbookBootstrapPort
}

type RouteOptions struct {
	WorkbookBootstrap    WorkbookBootstrapPort
	CollaborationSession CollaborationSessionPort
}
