package migrate

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestMigrateFacadeFreshStateAdmission_Integration(t *testing.T) {
	h := pgtest.Start(t)
	for _, retained := range []string{"", "database", "backups", "packs", "temporary", "exports", "objects"} {
		t.Run("retained_"+retained, func(t *testing.T) {
			target := h.NewDatabaseT(t, "facade_admission")
			roots := configtest.SetupTempRoots(t)
			configtest.BindPostgresDSNToDatabaseRoot(t, roots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], target.DSN, postgres.PurposeMigration)
			loaded := configtest.LoadFixture(t, []string{"config", "valid.toml"}, roots.Paths)
			cfg := loaded.Deployment()
			db, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if retained == "database" {
				if _, err := db.Exec(`CREATE TABLE public.retained_fixture (id integer)`); err != nil {
					t.Fatal(err)
				}
			} else if retained != "" {
				root := map[string]string{"backups": cfg.Roots.BackupStorage.Path, "packs": cfg.Roots.ReferencePackStorage.Path, "temporary": cfg.Roots.TemporaryWork.Path, "exports": cfg.Roots.ExportOutputs.Path, "objects": cfg.Roots.ObjectStorage.Path}[retained]
				if err := os.WriteFile(filepath.Join(root, "retained"), []byte("keep"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			runner := newMigrateRunner(&bytes.Buffer{})
			runner.loadConfig = func() (configassembly.Loaded, error) { return loaded, nil }
			err = runner.run(context.Background())
			if retained != "" {
				if !errors.Is(err, database_migrations.ErrIncompatibleCryptoState) {
					t.Fatalf("retained admission = %v", err)
				}
				var ledger bool
				if err := db.QueryRow(`SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&ledger); err != nil || ledger {
					t.Fatalf("rejection created ledger: %t, %v", ledger, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := runner.run(context.Background()); err != nil {
				t.Fatalf("current-state retry: %v", err)
			}
			var count int
			var format string
			if err := db.QueryRow(`SELECT count(*),min(format_id) FROM public.application_crypto_format`).Scan(&count, &format); err != nil || count != 1 || format != database_migrations.ApplicationCryptoFormat {
				t.Fatalf("identity: %d %q %v", count, format, err)
			}
			if _, err := db.Exec(`DELETE FROM public.application_crypto_format`); err != nil {
				t.Fatal(err)
			}
			if err := runner.run(context.Background()); !errors.Is(err, database_migrations.ErrIncompatibleCryptoState) {
				t.Fatalf("unstamped retained database accepted: %v", err)
			}
		})
	}
}
