package testsupport

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	workbookstartupbootstrap "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func NewIncidentStore(pool postgres.DB) *incidents.Store {
	return incidents.NewStoreWithOptions(pool, incidents.StoreOptions{
		WorkbookBootstrap: workbookstartupbootstrap.NewIncidentCreatePreferencesPort(),
	})
}
