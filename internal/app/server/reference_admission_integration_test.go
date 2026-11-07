package server

import (
	"errors"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/contractadmission"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/browserbuild"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	admissionfixture "github.com/JochiRaider/cartulary/internal/testutil/referenceadmissionfixture"
	"github.com/JochiRaider/cartulary/internal/testutil/reportingfixture"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestServerReferencePackRetainedAdmission_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	cases := append(slices.Clone(admissionfixture.Cases), "binding query failure", "reference table missing", "reporting table missing")
	cases = append(cases, admissionfixture.SchemaCases...)
	for _, kind := range cases {
		t.Run(kind, func(t *testing.T) {
			database := harness.PrepareIsolatedDatabaseT(t, "server-reference-admission")
			db, err := pgtest.OpenPurposeDatabase(t.Context(), database.DSN, postgres.PurposeMigration)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			switch kind {
			case "crypto state", "schema behind", "schema ahead", "schema lineage":
				admissionfixture.SeedSchemaFailure(t, db, kind)
			case "binding query failure":
				admissionfixture.SeedBindingQueryFailure(t, db)
			case "reference table missing", "reporting table missing":
				admissionfixture.SeedMissingTable(t, db, kind)
			default:
				admissionfixture.Seed(t, db, kind)
			}
			admitted := openAdmittedRuntimePool(t, t.Context(), database)
			defer admitted.Close()
			roots := configtest.SetupTempRoots(t)
			loaded := BindPostgres(t, RuntimeConfigWithOverlays(t, roots.Paths), database.Env())
			cfg := loaded.Deployment()
			admissionfixture.Sentinels(t, cfg.Roots.TemporaryWork.Path, cfg.Roots.ReferencePackStorage.Path)
			state, files := admissionfixture.State(t, db), admissionfixture.Files(t, roots.Base)
			runtime, err := NewRuntime(t.Context(), loaded, Options{BrowserBuild: browserbuild.New(t), ReportingRuntime: reportingfixture.Runtime{}, Postgres: admitted})
			if runtime != nil {
				runtime.Close()
			}
			if runtime != nil {
				t.Fatal("rejected state reached runtime")
			}
			if slices.Contains(admissionfixture.SchemaCases, kind) {
				admissionfixture.RequireSchemaFailure(t, err, kind)
			} else if kind == "binding query failure" || kind == "reference table missing" || kind == "reporting table missing" {
				code := "42P01"
				if kind == "binding query failure" {
					code = "42703"
				}
				var databaseError *pgconn.PgError
				if contractadmission.IsIncompatible(err) || !errors.As(err, &databaseError) || databaseError.Code != code {
					t.Fatal("query failure lost its operational classification", err)
				}
			} else if !contractadmission.IsIncompatible(err) {
				t.Fatalf("retained %s reached startup: %v", kind, err)
			}
			if admissionfixture.State(t, db) != state || admissionfixture.Files(t, roots.Base) != files {
				t.Fatal("rejected startup changed retained state")
			}
			if err := admitted.Pool().Ping(t.Context()); err != nil {
				t.Fatal("rejection closed borrowed pool", err)
			}
		})
	}
}
