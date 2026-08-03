package indicators_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func newIndicatorTestStore(t testing.TB, db postgres.DB, appender *revisions.Appender) *indicators.Store {
	t.Helper()
	store, err := indicators.NewStore(indicators.StoreDependencies{Postgres: db, Revisions: appender})
	if err != nil {
		t.Fatalf("compose Indicator test store: %v", err)
	}
	return store
}
