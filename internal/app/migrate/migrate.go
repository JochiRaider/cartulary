package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type migrateRunner struct {
	stderr     io.Writer
	loadConfig func() (config.Config, error)
	openSQL    func(config.Config) (*sql.DB, error)
	migrate    func(context.Context, *sql.DB, postgres.MigrationSource, string, ...string) (postgres.MigrationStatus, error)
	source     postgres.MigrationSource
}

func RunMigrateCLIContext(ctx context.Context, args []string, stderr io.Writer) int {
	return newMigrateRunner(stderr).runCLI(ctx, args)
}

func newMigrateRunner(stderr io.Writer) migrateRunner {
	return migrateRunner{
		stderr: normalizeMigrateWriter(stderr),
		loadConfig: func() (config.Config, error) {
			catalog, err := extensionassembly.GeneratedInactiveConfigurationCatalog()
			if err != nil {
				return config.Config{}, err
			}
			return config.LoadWithOptions(config.LoadOptions{ExtensionInactiveCatalog: catalog})
		},
		openSQL: postgres.OpenSQL,
		migrate: postgres.Migrate,
		source:  dbmigrations.Source(),
	}
}

func (runner migrateRunner) runCLI(ctx context.Context, args []string) int {
	if !isExactMigrateUp(args) {
		_, _ = fmt.Fprintln(runner.stderr, "usage: migrate up")
		return 2
	}

	if err := runner.run(ctx); err != nil {
		var remediation *postgres.MigrationRemediationError
		if errors.As(err, &remediation) {
			_, _ = fmt.Fprintln(runner.stderr, remediation.ReportJSON())
		}
		runner.logger().Error("migrate failed", "error", err)
		return 1
	}

	return 0
}

func (runner migrateRunner) run(ctx context.Context) error {
	cfg, err := runner.loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := runner.openSQL(cfg)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	if db != nil {
		defer db.Close()
	}

	_, err = runner.migrate(ctx, db, runner.source, "up")
	return err
}

func (runner migrateRunner) logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(runner.stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func isExactMigrateUp(args []string) bool {
	return len(args) == 1 && args[0] == "up"
}

func normalizeMigrateWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}
