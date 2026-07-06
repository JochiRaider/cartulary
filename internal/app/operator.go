package app

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	OperatorObjectStoreInitResultSchemaID = "cartulary.operator.object_store_init_result.v1"
)

type operatorPostgresPool interface {
	postgres.DB
	Close()
}

type operatorRunner struct {
	stdout                  io.Writer
	stderr                  io.Writer
	loadConfig              func(string) (config.Config, error)
	setupPostgres           func(context.Context, config.Config) (operatorPostgresPool, error)
	setupObjectStore        func(context.Context, config.Config) (objectstore.Store, error)
	ensureObjectStoreBucket func(context.Context, config.Config) (objectstore.EnsureBucketResult, error)
	newBackupStorage        func(config.Config) (recovery.BackupStorage, error)
	now                     func() time.Time
}

type operatorCLIResult struct {
	stop             bool
	exitCode         int
	command          string
	asOf             time.Time
	sourceConfigPath string
	manifestPath     string
}

type OperatorObjectStoreInitResult struct {
	SchemaID      string `json:"schema_id"`
	Result        string `json:"result"`
	Created       bool   `json:"created"`
	AlreadyExists bool   `json:"already_exists"`
}

func RunOperatorCLIContext(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	runner := newOperatorRunnerForCLI(stdout, stderr)
	if handled, exitCode := runner.runRecoveryCLI(ctx, args); handled {
		return exitCode
	}
	return runner.runCLI(ctx, args)
}

func RunOperatorCLI(args []string, stdout io.Writer, stderr io.Writer) int {
	return RunOperatorCLIContext(context.Background(), args, stdout, stderr)
}

var newOperatorRunnerForCLI = newOperatorRunner

func newOperatorRunner(stdout io.Writer, stderr io.Writer) operatorRunner {
	return operatorRunner{
		stdout: normalizeOperatorWriter(stdout),
		stderr: normalizeOperatorWriter(stderr),
		loadConfig: func(path string) (config.Config, error) {
			if strings.TrimSpace(path) == "" {
				return config.Load()
			}
			return config.LoadWithOptions(config.LoadOptions{Path: path})
		},
		setupPostgres: func(ctx context.Context, cfg config.Config) (operatorPostgresPool, error) {
			return postgres.Setup(ctx, cfg)
		},
		setupObjectStore: func(ctx context.Context, cfg config.Config) (objectstore.Store, error) {
			return objectstore.Setup(ctx, cfg)
		},
		ensureObjectStoreBucket: func(ctx context.Context, cfg config.Config) (objectstore.EnsureBucketResult, error) {
			return objectstore.EnsureConfiguredBucket(ctx, cfg, nil)
		},
		newBackupStorage: func(cfg config.Config) (recovery.BackupStorage, error) {
			return recovery.NewBackupStorageFromConfig(cfg)
		},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (runner operatorRunner) runCLI(ctx context.Context, args []string) int {
	parsed := parseOperatorCLIArgs(args, runner.stderr)
	if parsed.stop {
		return parsed.exitCode
	}

	if err := runner.run(ctx, parsed); err != nil {
		runner.logger().Error("operator command failed", "error", err)
		return 1
	}
	return 0
}

func (runner operatorRunner) run(ctx context.Context, parsed operatorCLIResult) error {
	switch parsed.command {
	case "migration-evidence capture":
		return runner.runMigrationEvidenceCapture(ctx, parsed)
	case "object-store init":
		return runner.runObjectStoreInit(ctx, parsed)
	default:
		return fmt.Errorf("unsupported operator command %q", parsed.command)
	}
}

func (runner operatorRunner) runObjectStoreInit(ctx context.Context, parsed operatorCLIResult) error {
	cfg, err := runner.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	result, err := runner.ensureObjectStoreBucket(ctx, cfg)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	payload := OperatorObjectStoreInitResult{
		SchemaID:      OperatorObjectStoreInitResultSchemaID,
		Result:        operatorObjectStoreInitResultCode(result),
		Created:       result.Created,
		AlreadyExists: result.AlreadyExists,
	}
	return runner.encodeJSON(payload)
}

func parseOperatorCLIArgs(args []string, stderr io.Writer) operatorCLIResult {
	if len(args) < 2 {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), operatorUsage())
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	if len(args) >= 2 && args[0] == "object-store" && args[1] == "init" {
		return parseObjectStoreInitArgs(args[2:], stderr)
	}
	switch args[0] + " " + args[1] {
	case "migration-evidence capture":
		return parseMigrationEvidenceCaptureArgs(args[2:], stderr)
	default:
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), operatorUsage())
		return operatorCLIResult{stop: true, exitCode: 2}
	}
}

func parseObjectStoreInitArgs(args []string, stderr io.Writer) operatorCLIResult {
	flags := flag.NewFlagSet("operator object-store init", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	sourceConfig := flags.String("config", "", "optional deployment config path")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return operatorCLIResult{stop: true, exitCode: 0}
		}
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	return operatorCLIResult{
		command:          "object-store init",
		sourceConfigPath: strings.TrimSpace(*sourceConfig),
	}
}

func parseMigrationEvidenceCaptureArgs(args []string, stderr io.Writer) operatorCLIResult {
	flags := flag.NewFlagSet("operator migration-evidence capture", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	sourceConfig := flags.String("source-config", "", "optional source deployment config path")
	manifest := flags.String("manifest", defaultMigrationEvidenceManifestPath, "migration history manifest path")
	asOfRaw := flags.String("as-of", "", "RFC3339 evidence collection timestamp")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return operatorCLIResult{stop: true, exitCode: 0}
		}
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	asOf, ok := parseOperatorAsOfFlag(stderr, *asOfRaw)
	if !ok {
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	manifestPath := strings.TrimSpace(*manifest)
	if manifestPath == "" {
		_, _ = fmt.Fprintln(normalizeOperatorWriter(stderr), "manifest is required")
		return operatorCLIResult{stop: true, exitCode: 2}
	}
	return operatorCLIResult{
		command:          "migration-evidence capture",
		asOf:             asOf,
		sourceConfigPath: strings.TrimSpace(*sourceConfig),
		manifestPath:     manifestPath,
	}
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

func operatorUsage() string {
	return strings.Join([]string{
		"usage:",
		"  operator backup inspect latest [--source-config-file <path>] [--progress=jsonl]",
		"  operator backup create [--source-config-file <path>] [--progress=jsonl]",
		"  operator restore latest --source-config-file <path> --target-config-file <path> --confirm-backup-set-id <uuid> [--progress=jsonl]",
		"  operator restore-verify latest --source-config-file <path> --target-config-file <path> [--progress=jsonl]",
		"  operator restore-verify due --source-config-file <path> --target-config-file <path> [--progress=jsonl]",
		"  operator migration-evidence capture [-source-config <path>] [-manifest <path>] [-as-of <RFC3339>]",
		"  operator object-store init [-config <path>]",
	}, "\n")
}

func (runner operatorRunner) encodeJSON(payload any) error {
	encoder := json.NewEncoder(runner.stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		return fmt.Errorf("encode operator JSON: %w", err)
	}
	return nil
}

func operatorObjectStoreInitResultCode(result objectstore.EnsureBucketResult) string {
	if result.Created {
		return "created"
	}
	return "already_exists"
}

func sanitizeObjectStoreInitError(err error) error {
	if err == nil {
		return nil
	}
	reasonCode := "dependency_unavailable"
	var diagnosticsErr *config.DiagnosticsError
	if errors.As(err, &diagnosticsErr) && diagnosticsErr.Code != "" {
		reasonCode = diagnosticsErr.Code
	} else if adapterErr, ok := objectstore.AsAdapterError(err); ok && adapterErr.Reason != "" {
		reasonCode = string(adapterErr.Reason)
	} else {
		lower := strings.ToLower(err.Error())
		switch {
		case strings.Contains(lower, "service_ref"):
			reasonCode = "managed_service_ref_invalid"
		case strings.Contains(lower, "missing object-store"):
			reasonCode = "missing_service_binding"
		case strings.Contains(lower, "parse cartulary_s3_"):
			reasonCode = "invalid_service_binding"
		}
	}
	return fmt.Errorf("object-store init failed: reason_code=%s", reasonCode)
}

func (runner operatorRunner) logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(runner.stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func normalizeOperatorWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}
