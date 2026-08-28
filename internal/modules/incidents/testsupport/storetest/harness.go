package storetest

import (
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

type StoreHarness struct {
	DB        postgres.DB
	Incidents *incidents.Application
}

func StartStore(t testing.TB, prefix string) *StoreHarness {
	return StartStoreWithClock(t, prefix, time.Now)
}

func StartStoreWithClock(t testing.TB, prefix string, now func() time.Time) *StoreHarness {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	db := postgresHarness.BeginRollbackDBT(t, prefix)
	application, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            db,
		PreferenceBootstrap: workbookstartuppostgres.NewWriter(),
		Now:                 now,
	})
	if err != nil {
		t.Fatalf("construct Incidents application: %v", err)
	}
	return &StoreHarness{DB: db, Incidents: application}
}
