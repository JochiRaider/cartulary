package phase2storetest

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

type StoreHarness struct {
	DB postgres.DB
}

func StartStore(t testing.TB, prefix string) *StoreHarness {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	return &StoreHarness{DB: postgresHarness.BeginRollbackDBT(t, prefix)}
}
