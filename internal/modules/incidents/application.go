package incidents

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/workbook/startup/bootstrapport"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// Application is the Incidents application boundary. It owns transaction and
// policy coordination while the private repository owns persistence details.
type Application struct {
	pool                 postgres.DB
	authStore            *authn.Store
	repository           *repository
	admission            *admission.Checker
	preferenceBootstrap  bootstrapport.Writer
	incidentCreateCommit IncidentCreateCommitPort
}

// These values are persisted with public idempotency payloads and therefore
// remain stable even though HTTP response selection belongs to httpapi.
const (
	persistedSuccessStatus = 200
	persistedCreatedStatus = 201
)

func NewApplication(pool postgres.DB, preferenceBootstrap bootstrapport.Writer) *Application {
	return NewApplicationWithOptions(pool, ApplicationOptions{PreferenceBootstrap: preferenceBootstrap})
}

func NewApplicationWithOptions(pool postgres.DB, options ApplicationOptions) *Application {
	incidentCreateCommit := options.IncidentCreateCommit
	if incidentCreateCommit == nil {
		incidentCreateCommit = directIncidentCreateCommit{}
	}
	return &Application{
		pool:                 pool,
		authStore:            authn.NewStore(pool),
		repository:           newRepository(pool),
		admission:            admission.NewChecker(pool),
		preferenceBootstrap:  options.PreferenceBootstrap,
		incidentCreateCommit: incidentCreateCommit,
	}
}

type directIncidentCreateCommit struct{}

func (directIncidentCreateCommit) CommitIncidentCreate(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}

func (a *Application) ListAdministrativeAuditEvents(
	ctx context.Context,
	filter administrativeaudit.ListFilter,
) ([]administrativeaudit.Record, error) {
	return administrativeaudit.List(ctx, a.pool, filter)
}
