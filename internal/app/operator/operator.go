package operator

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

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	OperatorObjectStoreInitResultSchemaID      = "cartulary.operator.object_store_init_result.v1"
	OperatorCollaborationRequeueResultSchemaID = "cartulary.operator.collaboration_requeue_result.v1"
)

type operatorPostgresPool interface {
	postgres.DB
	Close()
}

type operatorRunner struct {
	stdout                  io.Writer
	stderr                  io.Writer
	loadConfig              func(string) (configassembly.Loaded, error)
	setupPostgres           func(context.Context, postgres.Settings) (operatorPostgresPool, error)
	setupObjectStore        func(context.Context, objectstore.Settings, objectstore.Instrumentation) (objectstore.Store, error)
	ensureObjectStoreBucket func(context.Context, objectstore.Settings) (objectstore.EnsureBucketResult, error)
	newBackupStorage        func(string, string) (recovery.BackupStorage, error)
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

type OperatorCollaborationRequeueResult struct {
	SchemaID   string `json:"schema_id"`
	Result     string `json:"result"`
	IncidentID string `json:"incident_id"`
}

type collaborationRequeueArgs struct {
	stop       bool
	exitCode   int
	configPath string
	incidentID uuid.UUID
}

func RunOperatorCLIContext(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	return newOperatorRunner(stdout, stderr).runCLI(ctx, args)
}

func newOperatorRunner(stdout io.Writer, stderr io.Writer) operatorRunner {
	return operatorRunner{
		stdout: normalizeOperatorWriter(stdout),
		stderr: normalizeOperatorWriter(stderr),
		loadConfig: func(path string) (configassembly.Loaded, error) {
			policy, err := extensionassembly.GeneratedInactiveConfigurationPolicy()
			if err != nil {
				return configassembly.Loaded{}, err
			}
			options := config.LoadOptions{InactivePolicy: policy}
			if strings.TrimSpace(path) == "" {
				loaded, err := configassembly.Load(options)
				if err != nil {
					return configassembly.Loaded{}, err
				}
				return loaded, nil
			}
			options.Path = path
			loaded, err := configassembly.Load(options)
			if err != nil {
				return configassembly.Loaded{}, err
			}
			return loaded, nil
		},
		setupPostgres: func(ctx context.Context, settings postgres.Settings) (operatorPostgresPool, error) {
			return postgres.Setup(ctx, settings)
		},
		setupObjectStore: func(ctx context.Context, settings objectstore.Settings, instrumentation objectstore.Instrumentation) (objectstore.Store, error) {
			return objectstore.Setup(ctx, settings, instrumentation)
		},
		ensureObjectStoreBucket: func(ctx context.Context, settings objectstore.Settings) (objectstore.EnsureBucketResult, error) {
			return objectstore.EnsureBucket(ctx, settings)
		},
		newBackupStorage: func(bindingKind string, rootPath string) (recovery.BackupStorage, error) {
			return recoveryassembly.NewBackupStorage(
				bindingKind,
				rootPath,
				nil,
			)
		},
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (runner operatorRunner) runCLI(ctx context.Context, args []string) int {
	registry, err := runner.commandRegistry()
	if err != nil {
		runner.logger().Error("operator command registry is invalid", "error", err)
		return 1
	}
	return registry.run(ctx, args)
}

func (runner operatorRunner) commandRegistry() (operatorCommandRegistry, error) {
	recoveryHandler := func(ctx context.Context, args []string) int {
		if handled, exitCode := runner.runRecoveryCLI(ctx, args); handled {
			return exitCode
		}
		return 2
	}
	return newOperatorCommandRegistry(runner.stderr, []operatorCommandDescriptor{
		{
			Tokens:           []string{"backup", "inspect", "latest"},
			Owner:            "recovery",
			Usage:            "operator backup inspect latest [--source-config-file <path>] [--progress=jsonl]",
			Run:              recoveryHandler,
			InvalidNamespace: recoveryHandler,
		},
		{
			Tokens:           []string{"backup", "create"},
			Owner:            "recovery",
			Usage:            "operator backup create [--source-config-file <path>] [--progress=jsonl]",
			Run:              recoveryHandler,
			InvalidNamespace: recoveryHandler,
		},
		{
			Tokens:           []string{"restore", "latest"},
			Owner:            "recovery",
			Usage:            "operator restore latest --source-config-file <path> --target-config-file <path> --confirm-backup-set-id <uuid> [--progress=jsonl]",
			Run:              recoveryHandler,
			InvalidNamespace: recoveryHandler,
		},
		{
			Tokens:           []string{"restore-verify", "latest"},
			Owner:            "recovery",
			Usage:            "operator restore-verify latest --source-config-file <path> --target-config-file <path> [--progress=jsonl]",
			Run:              recoveryHandler,
			InvalidNamespace: recoveryHandler,
		},
		{
			Tokens:           []string{"restore-verify", "due"},
			Owner:            "recovery",
			Usage:            "operator restore-verify due --source-config-file <path> --target-config-file <path> [--progress=jsonl]",
			Run:              recoveryHandler,
			InvalidNamespace: recoveryHandler,
		},
		{
			Tokens: []string{"migration-evidence", "capture"},
			Owner:  "postgres-migration-evidence",
			Usage:  "operator migration-evidence capture [-source-config <path>] [-manifest <path>] [-as-of <RFC3339>]",
			Run:    runner.runMigrationEvidenceCommand,
		},
		{
			Tokens: []string{"object-store", "init"},
			Owner:  "object-store",
			Usage:  "operator object-store init [-config <path>]",
			Run:    runner.runObjectStoreInitCommand,
		},
		{
			Tokens: []string{"collaboration", "requeue"},
			Owner:  "collaboration",
			Usage:  "operator collaboration requeue --incident-id <uuid> [-config <path>]",
			Run:    runner.runCollaborationRequeueCommand,
		},
	})
}

func (runner operatorRunner) runCollaborationRequeueCommand(ctx context.Context, args []string) int {
	parsed := parseCollaborationRequeueArgs(args[2:], runner.stderr)
	if parsed.stop {
		return parsed.exitCode
	}
	loaded, err := runner.loadConfig(parsed.configPath)
	if err != nil {
		runner.logger().Error("operator command failed", "error", err)
		return 1
	}
	settings, err := postgres.ResolveSettings(
		configassembly.PostgresBinding(loaded.Deployment()),
		nil,
	)
	if err != nil {
		runner.logger().Error("operator command failed", "error", err)
		return 1
	}
	pool, err := runner.setupPostgres(ctx, settings)
	if err != nil {
		runner.logger().Error("operator command failed", "error", err)
		return 1
	}
	defer pool.Close()
	if err := collaboration.NewRecoveryService(pool).RequeueIncident(
		ctx,
		parsed.incidentID,
		runner.now(),
	); err != nil {
		runner.logger().Error("operator command failed", "error", err)
		return 1
	}
	if err := runner.encodeJSON(OperatorCollaborationRequeueResult{
		SchemaID:   OperatorCollaborationRequeueResultSchemaID,
		Result:     "collaboration_incident_requeued",
		IncidentID: parsed.incidentID.String(),
	}); err != nil {
		runner.logger().Error("operator command failed", "error", err)
		return 1
	}
	return 0
}

func (runner operatorRunner) runMigrationEvidenceCommand(ctx context.Context, args []string) int {
	parsed := parseMigrationEvidenceCaptureArgs(args[2:], runner.stderr)
	if parsed.stop {
		return parsed.exitCode
	}
	if err := runner.runMigrationEvidenceCapture(ctx, parsed); err != nil {
		runner.logger().Error("operator command failed", "error", err)
		return 1
	}
	return 0
}

func (runner operatorRunner) runObjectStoreInitCommand(ctx context.Context, args []string) int {
	parsed := parseObjectStoreInitArgs(args[2:], runner.stderr)
	if parsed.stop {
		return parsed.exitCode
	}
	if err := runner.runObjectStoreInit(ctx, parsed); err != nil {
		runner.logger().Error("operator command failed", "error", err)
		return 1
	}
	return 0
}

func (runner operatorRunner) runObjectStoreInit(ctx context.Context, parsed operatorCLIResult) error {
	loaded, err := runner.loadConfig(parsed.sourceConfigPath)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	cfg := loaded.Deployment()
	settings, err := objectstore.ResolveSettings(configassembly.ObjectStoreBinding(cfg), nil)
	if err != nil {
		return sanitizeObjectStoreInitError(err)
	}
	result, err := runner.ensureObjectStoreBucket(ctx, settings)
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

func parseCollaborationRequeueArgs(args []string, stderr io.Writer) collaborationRequeueArgs {
	flags := flag.NewFlagSet("operator collaboration requeue", flag.ContinueOnError)
	flags.SetOutput(normalizeOperatorWriter(stderr))
	configPath := flags.String("config", "", "optional deployment config path")
	incidentIDText := flags.String("incident-id", "", "quarantined incident UUID")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return collaborationRequeueArgs{stop: true, exitCode: 0}
		}
		return collaborationRequeueArgs{stop: true, exitCode: 2}
	}
	incidentID, err := uuid.Parse(strings.TrimSpace(*incidentIDText))
	if err != nil || incidentID == uuid.Nil || flags.NArg() != 0 {
		fmt.Fprintln(normalizeOperatorWriter(stderr), "operator collaboration requeue requires exactly one non-zero --incident-id UUID")
		return collaborationRequeueArgs{stop: true, exitCode: 2}
	}
	return collaborationRequeueArgs{
		configPath: strings.TrimSpace(*configPath),
		incidentID: incidentID,
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

func (runner operatorRunner) encodeJSON(payload any) error {
	encoder := json.NewEncoder(runner.stdout)
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
