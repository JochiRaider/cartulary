package appsupport

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

// StoreHarness supplies an owner-neutral transaction-scoped database for
// module tests that do not need a running application server.
type StoreHarness struct {
	DB postgres.DB
}

func StartStore(t testing.TB, prefix string) *StoreHarness {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	if pgtest.ExplicitPostgresFixturePolicyT(t) == pgtest.PostgresFixturePolicyTemplateClone {
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
		pool, err := pgxpool.New(context.Background(), testDB.DSN)
		if err != nil {
			t.Fatalf("open template-clone postgres pool: %v", err)
		}
		t.Cleanup(pool.Close)
		return &StoreHarness{DB: pool}
	}
	return &StoreHarness{DB: postgresHarness.BeginRollbackDBT(t, prefix)}
}
