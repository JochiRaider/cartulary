// Package storetest contains DB-backed auth helpers that stay free of
// runtime and HTTP test dependencies so auth-package unit tests can import them
// without a package cycle.
package storetest

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
