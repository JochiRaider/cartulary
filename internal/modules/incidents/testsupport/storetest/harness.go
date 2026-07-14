package storetest

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

type StoreHarness struct {
	DB        postgres.DB
	Incidents *incidents.Store
}

func StartStore(t testing.TB, prefix string) *StoreHarness {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	db := postgresHarness.BeginRollbackDBT(t, prefix)
	return &StoreHarness{DB: db, Incidents: appsupport.NewIncidentStore(db)}
}
