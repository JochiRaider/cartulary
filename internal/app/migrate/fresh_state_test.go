package migrate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
)

func TestFreshStorageAdmissionIsReadOnly(t *testing.T) {
	loaded, err := newTestMigrateRunner(t).loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg := loaded.Deployment()
	base := t.TempDir()
	cfg.Roots.BackupStorage.Path = filepath.Join(base, "backups")
	cfg.Roots.ReferencePackStorage.Path = filepath.Join(base, "packs")
	cfg.Roots.TemporaryWork.Path = filepath.Join(base, "temporary")
	cfg.Roots.ExportOutputs.Path = filepath.Join(base, "exports")
	cfg.Roots.ObjectStorage.BindingKind = "filesystem_root"
	cfg.Roots.ObjectStorage.ServiceRef = ""
	cfg.Roots.ObjectStorage.Path = filepath.Join(base, "objects")
	inspect, err := freshStorageInspector(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := inspect(context.Background()); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(base)
	if err != nil || len(entries) != 0 {
		t.Fatalf("inspection created state: %v, %v", entries, err)
	}
	for _, name := range []string{"backups", "packs", "temporary", "exports", "objects"} {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join(base, name)
			if err := os.Mkdir(root, 0700); err != nil {
				t.Fatal(err)
			}
			// Empty directories are fresh; even an empty child counts as retained state.
			if err := inspect(context.Background()); err != nil {
				t.Fatal(err)
			}
			child := filepath.Join(root, "retained")
			if err := os.Mkdir(child, 0700); err != nil {
				t.Fatal(err)
			}
			if err := inspect(context.Background()); !errors.Is(err, database_migrations.ErrIncompatibleCryptoState) {
				t.Fatalf("retained admission = %v", err)
			}
			if _, err := os.Stat(child); err != nil {
				t.Fatalf("inspection changed state: %v", err)
			}
			if err := os.Remove(child); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(root); err != nil {
				t.Fatal(err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := inspect(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled inspection = %v", err)
	}
}
