package parties_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestPartyClockCutoverPreservesIncompatibleHistory_Integration(t *testing.T) {
	h := pgtest.Start(t)
	for _, populated := range []bool{false, true} {
		name := "empty initialization"
		if populated {
			name = "retained party state"
		}
		t.Run(name, func(t *testing.T) {
			migration := h.MigrationDatabaseThroughT(t, 72)
			db, ctx := migration.SQL(), t.Context()
			if err := parties.PreflightClockCutover(ctx, db); err != nil {
				t.Fatal(err)
			}
			if populated {
				_, err := db.ExecContext(ctx, `
INSERT INTO users(id,email,display_name,password_hash) VALUES ('73000000-0000-4000-8000-000000000001','party-clock@example.test','Cutover','not-used');
INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id)
VALUES ('73000000-0000-4000-8000-000000000002','PARTY-CLOCK','party-clock','Cutover','active','73000000-0000-4000-8000-000000000001','73000000-0000-4000-8000-000000000001');
INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id)
VALUES ('73000000-0000-4000-8000-000000000003','73000000-0000-4000-8000-000000000002','party','73000000-0000-4000-8000-000000000001','73000000-0000-4000-8000-000000000001');
INSERT INTO parties(record_id,incident_id,display_name,party_kind,created_at,updated_at)
VALUES ('73000000-0000-4000-8000-000000000003','73000000-0000-4000-8000-000000000002','Private retained name','person','2020-01-01T00:00:00Z','2020-01-02T00:00:00Z');`)
				if err != nil {
					t.Fatal(err)
				}
				var before string
				if err := db.QueryRowContext(ctx, `SELECT to_jsonb(p)::text FROM parties p`).Scan(&before); err != nil {
					t.Fatal(err)
				}
				var rejected *parties.ClockCutoverRequiredError
				if err := parties.PreflightClockCutover(ctx, db); !errors.As(err, &rejected) {
					t.Fatal("history admitted", err)
				}
				if strings.Contains(rejected.RemediationReportJSON(), "Private retained") {
					t.Fatal("private data in diagnostic")
				}
				if err := migration.ApplyThrough(ctx, 73); err == nil {
					t.Fatal("migration discarded retained clocks")
				}
				var after string
				var head int64
				if err := db.QueryRowContext(ctx, `SELECT to_jsonb(p)::text FROM parties p`).Scan(&after); err != nil {
					t.Fatal(err)
				}
				if err := db.QueryRowContext(ctx, `SELECT max(version_id) FROM goose_db_version WHERE is_applied`).Scan(&head); err != nil {
					t.Fatal(err)
				}
				if before != after || head != 72 {
					t.Fatal("failed cutover mutated source or migration state")
				}
				return
			}
			if err := migration.ApplyThrough(ctx, 73); err != nil {
				t.Fatal(err)
			}
			if err := parties.PreflightClockCutover(ctx, db); err != nil {
				t.Fatal(err)
			}
			var clocks int
			if err := db.QueryRowContext(ctx, `SELECT count(*) FROM pg_attribute WHERE attrelid='public.parties'::regclass AND attname IN ('created_at','updated_at') AND NOT attisdropped`).Scan(&clocks); err != nil || clocks != 0 {
				t.Fatal("duplicate clocks retained", clocks, err)
			}
			if err := migration.RollbackThrough(ctx, 72); err == nil {
				t.Fatal("downgrade fabricated clocks")
			}
		})
	}
}
