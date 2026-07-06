package app

import (
	"context"
	"fmt"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/migrationevidence"
)

const defaultMigrationEvidenceManifestPath = migrationevidence.DefaultManifestPath

func (runner operatorRunner) runMigrationEvidenceCapture(ctx context.Context, parsed operatorCLIResult) error {
	cfg, err := runner.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	pool, err := runner.setupPostgres(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer pool.Close()

	collectedAt := parsed.asOf
	if collectedAt.IsZero() {
		collectedAt = runner.now()
	}
	result, err := migrationevidence.Build(ctx, cfg, pool, collectedAt.UTC(), parsed.manifestPath, dbmigrations.Files)
	if err != nil {
		return err
	}
	return runner.encodeJSON(result)
}
