package incidents

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/workbook/startup/bootstrapport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

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

type TerminalMutationCoordinator interface {
	CoordinateIncidentLifecycle(
		context.Context,
		authn.UserRecord,
		uuid.UUID,
		string,
		IncidentLifecycleRequest,
		[]byte,
		string,
		time.Time,
	) (IncidentLifecycleResult, error)
	CoordinateMembershipDeletion(
		context.Context,
		authn.UserRecord,
		uuid.UUID,
		uuid.UUID,
		MembershipDeleteRequest,
		string,
	) (MembershipDeleteResult, error)
}

type ApplicationOptions struct {
	PreferenceBootstrap  bootstrapport.Writer
	IncidentCreateCommit IncidentCreateCommitPort
}
