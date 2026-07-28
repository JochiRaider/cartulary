package incidents

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/workbookpreferences"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// Application is the Incidents application boundary. It owns transaction and
// policy coordination while the private repository owns persistence details.
type Application struct {
	pool                postgres.DB
	authStore           *authn.Store
	repository          *repository
	preferenceBootstrap PreferenceBootstrapPort
}

func NewApplication(pool postgres.DB) *Application {
	return NewApplicationWithOptions(pool, ApplicationOptions{})
}

func NewApplicationWithOptions(pool postgres.DB, options ApplicationOptions) *Application {
	preferenceBootstrap := options.PreferenceBootstrap
	if preferenceBootstrap == nil {
		preferenceBootstrap = workbookpreferences.NewBootstrap()
	}
	return &Application{
		pool:                pool,
		authStore:           authn.NewStore(pool),
		repository:          newRepository(pool),
		preferenceBootstrap: preferenceBootstrap,
	}
}

func (a *Application) ListAdministrativeAuditEvents(
	ctx context.Context,
	filter administrativeaudit.ListFilter,
) ([]administrativeaudit.Record, error) {
	return administrativeaudit.List(ctx, a.pool, filter)
}
