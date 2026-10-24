package recoveryassembly

import (
	"context"
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRecoveryStateDatabaseCoverageRejectsUnclassifiedTableBeforeAdmission_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	pool := harness.BeginRollbackDBT(t, "recovery-state-coverage")
	catalog, err := CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatalf("build current recovery state catalog: %v", err)
	}
	if err := ValidateRecoveryStateDatabaseCoverage(context.Background(), pool, catalog); err != nil {
		t.Fatalf("validate current database recovery state coverage: %v", err)
	}
	if _, err := pool.Exec(context.Background(), `CREATE TABLE unclassified_future_table (id bigint PRIMARY KEY)`); err != nil {
		t.Fatalf("create unclassified future table: %v", err)
	}
	if _, err := pool.Exec(context.Background(), `SET LOCAL ROLE cartulary_recovery`); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRecoveryStateDatabaseCoverage(context.Background(), pool, catalog); !errors.Is(err, recoverystate.ErrInvalidCatalog) {
		t.Fatalf("unclassified database table error = %v, want ErrInvalidCatalog", err)
	}
}
