package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type migrateRunner struct {
	stderr     io.Writer
	loadConfig func() (configassembly.Loaded, error)
	openSQL    func(context.Context, postgres.Settings) (*sql.DB, error)
	apply      func(context.Context, *sql.DB, *sql.DB, *database_migrations.Source, func(context.Context) error) error
	source     func() (*database_migrations.Source, error)
}

func RunMigrateCLIContext(ctx context.Context, args []string, stderr io.Writer) int {
	return newMigrateRunner(stderr).runCLI(ctx, args)
}

func newMigrateRunner(stderr io.Writer) migrateRunner {
	return migrateRunner{
		stderr: normalizeMigrateWriter(stderr),
		loadConfig: func() (configassembly.Loaded, error) {
			loaded, err := configassembly.Load(configassembly.LoadOptions{})
			if err != nil {
				return configassembly.Loaded{}, err
			}
			return loaded, nil
		},
		openSQL: postgres.OpenSQL,
		apply: func(ctx context.Context, db, guard *sql.DB, source *database_migrations.Source, inspect func(context.Context) error) error {
			if err := reference_data.PreflightCutover(ctx, db); err != nil {
				return err
			}
			if err := reporting.PreflightCutover(ctx, db); err != nil {
				return err
			}
			return database_migrations.ApplyWithCryptoAdmission(ctx, db, guard, source, inspect)
		},
		source: dbmigrations.Source,
	}
}

func (runner migrateRunner) runCLI(ctx context.Context, args []string) int {
	if !isExactMigrateUp(args) {
		_, _ = fmt.Fprintln(runner.stderr, "usage: migrate up")
		return 2
	}

	if err := runner.run(ctx); err != nil {
		if errors.Is(err, cryptography.ErrExecutionPolicy) || errors.Is(err, database_migrations.ErrIncompatibleCryptoState) {
			_, _ = fmt.Fprintln(runner.stderr, err)
			return 1
		}
		var postgresFailure *postgres.ConfigurationError
		if errors.As(err, &postgresFailure) {
			_, _ = fmt.Fprintln(runner.stderr, postgresFailure.Reason())
			return 1
		}
		var remediation database_migrations.RemediationReporter
		if errors.As(err, &remediation) {
			_, _ = io.WriteString(runner.stderr, remediation.RemediationReportJSON())
			_, _ = io.WriteString(runner.stderr, "\n")
			return 1
		}
		var migrationFailure database_migrations.MigrationFailure
		if errors.As(err, &migrationFailure) {
			_, _ = fmt.Fprintln(runner.stderr, migrationFailure.ReasonCode())
			return 1
		}
		_, _ = fmt.Fprintln(runner.stderr, "migration_operation_failed")
		return 1
	}

	return 0
}

func (runner migrateRunner) run(ctx context.Context) error {
	if err := cryptography.AdmitExecution(); err != nil {
		return err
	}
	source, err := runner.source()
	if err != nil {
		return fmt.Errorf("load migration source: %w", err)
	}
	loaded, err := runner.loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := loaded.ValidateForInspection(); err != nil {
		return err
	}
	cfg := loaded.Deployment()
	inspect, err := freshStorageInspector(cfg)
	if err != nil {
		return err
	}
	settings, err := postgres.ResolveSettings(configassembly.PostgresBinding(cfg), postgres.PurposeMigration, nil)
	if err != nil {
		return fmt.Errorf("resolve postgres settings: %w", err)
	}
	db, err := runner.openSQL(ctx, settings)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	if db != nil {
		defer db.Close()
	}

	guard, err := runner.openSQL(ctx, settings)
	if err != nil {
		return fmt.Errorf("open initialization guard: %w", err)
	}
	if guard != nil {
		defer guard.Close()
	}
	return runner.apply(ctx, db, guard, source, inspect)
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
