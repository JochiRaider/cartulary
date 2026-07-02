package merge

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool           postgres.DB
	authStore      *authn.Store
	incidentAccess incidents.Access
	ports          entityStorePorts
}

func NewStore(pool postgres.DB) *Store {
	return &Store{
		pool:           pool,
		authStore:      authn.NewStore(pool),
		incidentAccess: incidents.NewAccess(pool),
		ports:          newEntityStorePorts(pool),
	}
}
