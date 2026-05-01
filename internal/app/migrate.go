package app

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
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

type migrateCLIResult struct {
	command  string
	args     []string
	stop     bool
	exitCode int
}

type migrateCommandFlag struct {
	value string
	set   bool
}

func RunMigrateCLI(args []string, stderr io.Writer) int {
	return RunMigrateCLIContext(context.Background(), args, stderr)
}

func RunMigrateCLIContext(ctx context.Context, args []string, stderr io.Writer) int {
	return newMigrateRunnerForCLI(stderr).runCLI(ctx, args)
}

var newMigrateRunnerForCLI = newMigrateRunner

func newMigrateRunner(stderr io.Writer) migrateRunner {
	return migrateRunner{
		stderr:     normalizeMigrateWriter(stderr),
		loadConfig: config.Load,
		openSQL:    postgres.OpenSQL,
		migrate:    postgres.Migrate,
		source:     dbmigrations.Source(),
	}
}

func (runner migrateRunner) runCLI(ctx context.Context, args []string) int {
	parsed := parseMigrateCLIArgs(args, runner.stderr)
	if parsed.stop {
		return parsed.exitCode
	}

	if err := runner.run(ctx, parsed.command, parsed.args); err != nil {
		runner.logger().Error("migrate failed", "error", err)
		return 1
	}

	return 0
}

func (runner migrateRunner) run(ctx context.Context, command string, args []string) error {
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

	_, err = runner.migrate(ctx, db, runner.source, command, args...)
	return err
}

func (runner migrateRunner) logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(runner.stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func parseMigrateCLIArgs(args []string, stderr io.Writer) migrateCLIResult {
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	flags.SetOutput(normalizeMigrateWriter(stderr))

	command := &migrateCommandFlag{value: "up"}
	flags.Var(command, "command", "goose command to run")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return migrateCLIResult{stop: true, exitCode: 0}
		}
		return migrateCLIResult{stop: true, exitCode: 2}
	}

	remaining := flags.Args()
	if command.set {
		return migrateCLIResult{
			command: command.value,
			args:    append([]string(nil), remaining...),
		}
	}

	if len(remaining) > 0 && remaining[0] != "" {
		return migrateCLIResult{
			command: remaining[0],
			args:    append([]string(nil), remaining[1:]...),
		}
	}

	return migrateCLIResult{
		command: command.value,
		args:    nil,
	}
}

func normalizeMigrateWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}

func (flagValue *migrateCommandFlag) String() string {
	return flagValue.value
}

func (flagValue *migrateCommandFlag) Set(value string) error {
	flagValue.value = value
	flagValue.set = true
	return nil
}
