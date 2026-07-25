package operator

import (
	"context"
	"fmt"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/migrationevidence"
)

const defaultMigrationEvidenceManifestPath = migrationevidence.DefaultManifestPath

func (runner operatorRunner) runMigrationEvidenceCapture(ctx context.Context, parsed operatorCLIResult) error {
	loaded, err := runner.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg := loaded.Deployment()
	settings, err := postgres.ResolveSettings(configassembly.PostgresBinding(cfg), nil)
	if err != nil {
		return fmt.Errorf("resolve postgres settings: %w", err)
	}
	pool, err := runner.setupPostgres(ctx, settings)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer pool.Close()

	collectedAt := parsed.asOf
	if collectedAt.IsZero() {
		collectedAt = runner.now()
	}
	result, err := migrationevidence.Build(ctx, migrationevidence.DatabaseBinding{
		BindingKind: cfg.Roots.DatabaseStorage.BindingKind,
		ServiceRef:  cfg.Roots.DatabaseStorage.ServiceRef,
	}, pool, collectedAt.UTC(), parsed.manifestPath, dbmigrations.Files)
	if err != nil {
		return err
	}
	return runner.encodeJSON(result)
}
