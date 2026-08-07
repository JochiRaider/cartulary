package operator

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type operatorRunner struct {
	stderr            io.Writer
	recovery          recoveryExecutor
	migrationEvidence migrationEvidenceExecutor
	objectStore       objectStoreExecutor
	collaboration     collaborationExecutor
}

func RunOperatorCLIContext(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	return newOperatorRunner(stdout, stderr).runCLI(ctx, args)
}

func newOperatorRunner(stdout io.Writer, stderr io.Writer) operatorRunner {
	stdout = normalizeOperatorWriter(stdout)
	stderr = normalizeOperatorWriter(stderr)
	transport := operatorTransport{stdout: stdout, stderr: stderr}
	loadConfig := func(path string) (configassembly.Loaded, error) {
		options := config.LoadOptions{}
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
	}
	setupPostgres := func(ctx context.Context, settings postgres.Settings) (operatorPostgresPool, error) {
		return postgres.Setup(ctx, settings)
	}
	setupObjectStore := func(ctx context.Context, settings objectstore.Settings, instrumentation objectstore.Instrumentation) (objectstore.Store, error) {
		return objectstore.Setup(ctx, settings, instrumentation)
	}
	ensureObjectStoreBucket := func(ctx context.Context, settings objectstore.Settings) (objectstore.EnsureBucketResult, error) {
		return objectstore.EnsureBucket(ctx, settings)
	}
	newBackupStorage := func(bindingKind string, rootPath string) (recovery.BackupStorage, error) {
		return recoveryassembly.NewBackupStorage(
			bindingKind,
			rootPath,
			nil,
		)
	}
	now := func() time.Time {
		return time.Now().UTC()
	}
	return operatorRunner{
		stderr: stderr,
		recovery: recoveryExecutor{
			transport:        transport,
			loadConfig:       loadConfig,
			setupPostgres:    setupPostgres,
			setupObjectStore: setupObjectStore,
			newBackupStorage: newBackupStorage,
			now:              now,
		},
		migrationEvidence: migrationEvidenceExecutor{
			transport:     transport,
			loadConfig:    loadConfig,
			setupPostgres: setupPostgres,
			now:           now,
		},
		objectStore: objectStoreExecutor{
			transport:               transport,
			loadConfig:              loadConfig,
			ensureObjectStoreBucket: ensureObjectStoreBucket,
		},
		collaboration: collaborationExecutor{
			transport:       transport,
			loadConfig:      loadConfig,
			setupPostgres:   setupPostgres,
			now:             now,
			newOperationID:  uuid.New,
			newRecoveryPort: newCollaborationRecoveryPort,
		},
	}
}

func (runner operatorRunner) runCLI(ctx context.Context, args []string) int {
	registry, err := runner.commandRegistry()
	if err != nil {
		operatorLogger(runner.stderr).Error("operator command registry is invalid", "error", err)
		return 1
	}
	return registry.run(ctx, args)
}

func (runner operatorRunner) commandRegistry() (operatorCommandRegistry, error) {
	recoveryHandler := func(ctx context.Context, args []string) int {
		if handled, exitCode := runner.recovery.runCLI(ctx, args); handled {
			return exitCode
		}
		return 2
	}
	return newOperatorCommandRegistry(runner.stderr, []operatorCommandDescriptor{
		{
			Tokens:           []string{"backup", "inspect", "latest"},
			Usage:            "operator backup inspect latest [--source-config-file <path>] [--progress=jsonl]",
			Run:              recoveryHandler,
			InvalidNamespace: recoveryHandler,
		},
		{
			Tokens:           []string{"backup", "create"},
			Usage:            "operator backup create [--source-config-file <path>] [--progress=jsonl]",
			Run:              recoveryHandler,
			InvalidNamespace: recoveryHandler,
		},
		{
			Tokens:           []string{"restore", "latest"},
			Usage:            "operator restore latest --source-config-file <path> --target-config-file <path> --confirm-backup-set-id <uuid> [--progress=jsonl]",
			Run:              recoveryHandler,
			InvalidNamespace: recoveryHandler,
		},
		{
			Tokens:           []string{"restore-verify", "latest"},
			Usage:            "operator restore-verify latest --source-config-file <path> --target-config-file <path> [--progress=jsonl]",
			Run:              recoveryHandler,
			InvalidNamespace: recoveryHandler,
		},
		{
			Tokens:           []string{"restore-verify", "due"},
			Usage:            "operator restore-verify due --source-config-file <path> --target-config-file <path> [--progress=jsonl]",
			Run:              recoveryHandler,
			InvalidNamespace: recoveryHandler,
		},
		{
			Tokens: []string{"migration-evidence", "capture"},
			Usage:  "operator migration-evidence capture [-source-config <path>] [-manifest <path>] [-as-of <RFC3339>]",
			Run:    runner.migrationEvidence.runCommand,
		},
		{
			Tokens: []string{"object-store", "init"},
			Usage:  "operator object-store init [-config <path>]",
			Run:    runner.objectStore.runCommand,
		},
		{
			Tokens: []string{"collaboration", "requeue"},
			Usage:  collaborationRequeueUsage,
			Run:    runner.collaboration.runCommand,
		},
	})
}
