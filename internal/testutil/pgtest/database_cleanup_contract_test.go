package pgtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

// Cleanup guarantees vary by resource kind and borrowing context.
func TestDatabaseFixtureCleanup(t *testing.T) {
	for _, ownership := range []string{"standalone", "attached", "persistent_borrower"} {
		for _, kind := range []string{"clone", "scratch_head", "scratch_target"} {
			t.Run(ownership+"/"+kind, func(t *testing.T) {
				authorizePGSuite(t)
				t.Setenv(suiteservices.SuiteIDEnv, "")
				t.Setenv(suiteservices.PersistentBorrowerEnv, "")
				if ownership != "standalone" {
					t.Setenv(suiteservices.SuiteIDEnv, "cleanup-contract")
				}
				if ownership == "persistent_borrower" {
					t.Setenv(suiteservices.PersistentBorrowerEnv, "1")
				}
				t.Setenv(suiteservices.TargetEnv, "backend-unit")
				t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
				t.Setenv("CARTULARY_TEST_RUN_ID", "cleanup-contract")
				t.Setenv(postgresFixturePolicyTestsEnv, "")
				t.Setenv(postgresFixturePolicyPackagesEnv, "")
				t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyTemplateClone)
				ops := defaultDatabaseOperations()
				inputs := processFixtureInputs()
				inputs["CARTULARY_FIXTURE_CAPABILITY"] = "postgres_migration"
				inputs[suiteservices.PGIsolationEnv] = "exclusive"
				h := &Harness{binding: bindFixture(inputs), operations: &ops, adminDSN: "postgres://owner/cluster", dsnTemplate: "postgres://fixture/{database}", templateDB: "suite_template", suiteHash: "suitehash", processHash: "procaaaa"}
				var created, dropped string
				var opened *sql.DB
				creates, drops, migrations := 0, 0, 0
				ops.create = func(_ context.Context, dsn, name, template string, confirmed func() error) (bool, error) {
					creates++
					created = name
					expectedTemplate := ""
					if kind == "clone" {
						expectedTemplate = h.templateDB
					}
					if dsn != h.adminDSN || template != expectedTemplate {
						t.Fatal("wrong creation authority or strategy")
					}
					return true, confirmed()
				}
				ops.openMigration = func(_ context.Context, dsn string, purpose postgres.Purpose) (*sql.DB, error) {
					if kind == "clone" || purpose != postgres.PurposeMigration || dsn != h.dsnFor(created) {
						t.Fatal("unexpected migration handle")
					}
					var err error
					opened, err = sql.Open("pgx", "postgres://unused/never_connected")
					return opened, err
				}
				ops.migrate = func(_ context.Context, db *sql.DB, source *database_migrations.Source) error {
					migrations++
					if kind != "scratch_head" || db != opened || source == nil {
						t.Fatal("wrong scratch head migration")
					}
					return nil
				}
				ops.applyThrough = func(_ context.Context, db *sql.DB, source *database_migrations.Source, version int64) error {
					migrations++
					if kind != "scratch_target" || db != opened || source == nil || version != 1 {
						t.Fatal("wrong targeted migration")
					}
					return nil
				}
				ops.drop = func(ctx context.Context, dsn, name string, _ func(postgrescleanup.DropAttempt)) error {
					drops++
					dropped = name
					if ctx.Err() != nil || name != created || dsn != h.adminDSN || !h.ownsDatabase(name) {
						t.Fatal("cleanup lost exact ownership")
					}
					if opened != nil {
						if err := opened.PingContext(ctx); err == nil || !strings.Contains(err.Error(), "database is closed") {
							t.Fatal("drop preceded migration handle closure", err)
						}
					}
					return nil
				}
				t.Run("borrow", func(t *testing.T) {
					switch kind {
					case "clone":
						if db := h.PrepareIsolatedDatabaseT(t, "cleanup"); db.Name != created {
							t.Fatal("wrong clone")
						}
					case "scratch_head":
						if db := h.MigrationDatabaseT(t); !db.valid() || db.SQL() != opened {
							t.Fatal("scratch capability not issued")
						}
					case "scratch_target":
						if db := h.MigrationDatabaseThroughT(t, 1); !db.valid() || db.SQL() != opened {
							t.Fatal("targeted capability not issued")
						}
					}
				})
				wantMigrations := 1
				if kind == "clone" {
					wantMigrations = 0
				}
				if created == "" || dropped != created || creates != 1 || drops != 1 || migrations != wantMigrations {
					t.Fatal("fixture not eagerly and exactly settled", creates, drops, migrations)
				}
			})
		}
	}
}
