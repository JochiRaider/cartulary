package appsupport

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func NewIncidentStore(pool postgres.DB) *incidents.Store {
	return incidents.NewStore(pool)
}
