package operator

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/migrationevidence"
)

type migrationEvidenceExecutor struct {
	transport     operatorTransport
	loadConfig    func(string) (configassembly.Loaded, error)
	setupPostgres func(context.Context, postgres.Settings) (operatorPostgresPool, error)
	now           func() time.Time
}

type migrationEvidenceCaptureArgs struct {
	asOf             time.Time
	sourceConfigPath string
	manifestPath     string
}

func (executor migrationEvidenceExecutor) runCommand(ctx context.Context, args []string) int {
	parsed, stop, exitCode := parseMigrationEvidenceCaptureArgs(args[2:], executor.transport.stderr)
	if stop {
		return exitCode
	}
	if err := executor.capture(ctx, parsed); err != nil {
		executor.transport.logger().Error("operator command failed", "error", err)
		return 1
	}
	return 0
}

func (executor migrationEvidenceExecutor) capture(ctx context.Context, parsed migrationEvidenceCaptureArgs) error {
	loaded, err := executor.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cfg := loaded.Deployment()
	settings, err := postgres.ResolveSettings(configassembly.PostgresBinding(cfg), nil)
	if err != nil {
		return fmt.Errorf("resolve postgres settings: %w", err)
	}
	pool, err := executor.setupPostgres(ctx, settings)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer pool.Close()

	collectedAt := parsed.asOf
	if collectedAt.IsZero() {
		collectedAt = executor.now()
	}
	result, err := migrationevidence.Build(ctx, migrationevidence.DatabaseBinding{
		BindingKind: cfg.Roots.DatabaseStorage.BindingKind,
		ServiceRef:  cfg.Roots.DatabaseStorage.ServiceRef,
	}, pool, collectedAt.UTC(), parsed.manifestPath, dbmigrations.Files)
	if err != nil {
		return err
	}
	return executor.transport.encodeJSON(result)
}

func parseMigrationEvidenceCaptureArgs(args []string, stderr io.Writer) (migrationEvidenceCaptureArgs, bool, int) {
	flags := flag.NewFlagSet("operator migration-evidence capture", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	sourceConfig := flags.String("source-config", "", "optional source deployment config path")
	manifest := flags.String("manifest", migrationevidence.DefaultManifestPath, "migration history manifest path")
	asOfRaw := flags.String("as-of", "", "RFC3339 evidence collection timestamp")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return migrationEvidenceCaptureArgs{}, true, 0
		}
		return migrationEvidenceCaptureArgs{}, true, 2
	}
	asOf, ok := parseOperatorAsOfFlag(stderr, *asOfRaw)
	if !ok {
		return migrationEvidenceCaptureArgs{}, true, 2
	}
	manifestPath := strings.TrimSpace(*manifest)
	if manifestPath == "" {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), "manifest is required")
		return migrationEvidenceCaptureArgs{}, true, 2
	}
	return migrationEvidenceCaptureArgs{
		asOf:             asOf,
		sourceConfigPath: strings.TrimSpace(*sourceConfig),
		manifestPath:     manifestPath,
	}, false, 0
}

func parseOperatorAsOfFlag(stderr io.Writer, asOfRaw string) (time.Time, bool) {
	if strings.TrimSpace(asOfRaw) == "" {
		return time.Time{}, true
	}
	parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(asOfRaw))
	if err != nil {
		_, _ = fmt.Fprintf(normalizeOperatorWriter(stderr), "as-of must be RFC3339: %v\n", err)
		return time.Time{}, false
	}
	return parsed.UTC(), true
}
