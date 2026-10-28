package parties_test

import (
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"testing"
)

func TestPartyBaselineUsesEnvelopeClocks_Integration(t *testing.T) {

	db := pgtest.Start(t).MigrationDatabaseT(t).SQL()
	var clocks int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM pg_catalog.pg_attribute WHERE attrelid='public.parties'::regclass AND attname IN ('created_at','updated_at') AND NOT attisdropped`).Scan(&clocks); err != nil || clocks != 0 {
		t.Fatal("duplicate clocks retained", clocks, err)
	}

}
