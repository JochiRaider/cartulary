package incidents

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PreferenceBootstrapPort interface {
	BootstrapIncidentPreferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, actorUserID uuid.UUID, now time.Time) error
}

type IncidentBundleImportFinalizationParams struct {
	IncidentID        uuid.UUID
	SubmittedByUserID uuid.UUID
	PublishedAt       time.Time
	RequestID         *string
	ClientTxnID       *string
}

type IncidentBundleImportFinalizer interface {
	FinalizeIncidentBundleImportTx(ctx context.Context, tx pgx.Tx, params IncidentBundleImportFinalizationParams) error
}

type CollaborationSessionPort interface {
	NotifyIncidentClosed(ctx context.Context, incidentID uuid.UUID)
	NotifyIncidentMembershipRevoked(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID)
}

type StoreOptions struct {
	PreferenceBootstrap PreferenceBootstrapPort
}

type RouteOptions struct {
	CollaborationSession CollaborationSessionPort
}
