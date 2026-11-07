package migrate

import (
	"bytes"
	"errors"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/contractadmission"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	admissionfixture "github.com/JochiRaider/cartulary/internal/testutil/referenceadmissionfixture"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMigrateReferencePackRetainedAdmission_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	for _, kind := range append(slices.Clone(admissionfixture.Cases), "binding query failure") {
		t.Run(kind, func(t *testing.T) {
			database := harness.PrepareIsolatedDatabaseT(t, "migration-reference-admission")
			db, err := pgtest.OpenPurposeDatabase(t.Context(), database.DSN, postgres.PurposeMigration)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if kind == "binding query failure" {
				admissionfixture.SeedBindingQueryFailure(t, db)
			} else {
				admissionfixture.Seed(t, db, kind)
			}
			roots := configtest.SetupTempRoots(t)
			configtest.BindPostgresDSNToDatabaseRoot(t, roots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], database.DSN, postgres.PurposeMigration)
			loaded := configtest.LoadFixture(t, []string{"config", "valid.toml"}, roots.Paths)
			cfg := loaded.Deployment()
			admissionfixture.Sentinels(t, cfg.Roots.TemporaryWork.Path, cfg.Roots.ReferencePackStorage.Path)
			state, files := admissionfixture.State(t, db), admissionfixture.Files(t, roots.Base)
			runner := newMigrateRunner(&bytes.Buffer{})
			runner.loadConfig = func() (configassembly.Loaded, error) { return loaded, nil }
			err = runner.run(t.Context())
			if kind == "binding query failure" {
				var databaseError *pgconn.PgError
				if contractadmission.IsIncompatible(err) || !errors.As(err, &databaseError) || databaseError.Code != "42703" {
					t.Fatal("binding query failure lost its operational classification", err)
				}
			} else if !contractadmission.IsIncompatible(err) {
				t.Fatalf("retained %s reached migration: %v", kind, err)
			}
			if admissionfixture.State(t, db) != state || admissionfixture.Files(t, roots.Base) != files {
				t.Fatal("rejected migration changed retained state")
			}
		})
	}
}
