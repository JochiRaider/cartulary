package operator

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/operatorcli"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/operatorops"
	"github.com/JochiRaider/cartulary/internal/platform/config"
)

func (runner operatorRunner) runRecoveryCLI(ctx context.Context, args []string) (bool, int) {
	extensionBackups, err := extensionassembly.GeneratedRecoveryCatalog()
	if err != nil {
		runner.logger().Error("extension recovery catalog is invalid", "error", err)
		return true, 1
	}
	return operatorcli.Runner{
		Stdout: runner.stdout,
		Stderr: runner.stderr,
		Now:    runner.now,
		Operations: operatorops.Service{
			LoadConfig: runner.loadConfig,
			SetupPostgres: func(ctx context.Context, cfg config.Config) (operatorops.PostgresPool, error) {
				return runner.setupPostgres(ctx, cfg)
			},
			SetupObjectStore:       runner.setupObjectStore,
			NewBackupStorage:       runner.newBackupStorage,
			NewProjectionRebuilder: projectionadapters.NewRestoreRebuilder,
			LoadJournalKey: func() (recovery.RecoveryEncryptionKey, error) {
				return recovery.LoadRecoveryEncryptionKey(nil)
			},
			ExtensionBackups: extensionBackups,
			Now:              runner.now,
		},
	}.Run(ctx, args)
}
