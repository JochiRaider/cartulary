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

// IncidentCreateCommitPort owns the final commit boundary for incident create.
// Implementations are invoked only after the incident, creator membership,
// workbook preferences, audit events, and idempotency result have been staged
// in the same transaction.
type IncidentCreateCommitPort interface {
	CommitIncidentCreate(context.Context, pgx.Tx) error
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

type ApplicationOptions struct {
	PreferenceBootstrap  PreferenceBootstrapPort
	IncidentCreateCommit IncidentCreateCommitPort
}

type RouteOptions struct {
	CollaborationSession CollaborationSessionPort
}
