package operator

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/operatorcli"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/operatorops"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
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
			LoadDeployment: runner.loadRecoveryDeployment,
			ReadTargetMarker: func(bindingKind string, rootPath string) ([]byte, error) {
				if bindingKind != "filesystem_root" {
					return nil, operatorops.ErrTargetMarkerRequiresFilesystemStorage
				}
				storage, err := recoveryassembly.NewFilesystemStorage(rootPath)
				if err != nil {
					return nil, err
				}
				defer storage.Close()
				return storage.ReadMarker(operatorops.RestoreVerificationTargetMarkerMaximumBytes)
			},
			NewProjectionServices: func(db postgres.DB) (restorecontract.ProjectionRebuilder, recovery.WorkbookProjectionQuery) {
				return timelineassembly.NewRecoveryProjectionServices(db)
			},
			LoadJournalKey: func() (recovery.RecoveryEncryptionKey, error) {
				return recovery.LoadRecoveryEncryptionKey(nil)
			},
			ExtensionBackups: extensionBackups,
			Now:              runner.now,
		},
	}.Run(ctx, args)
}

func (runner operatorRunner) loadRecoveryDeployment(path string) (operatorops.Deployment, error) {
	loaded, err := runner.loadConfig(path)
	if err != nil {
		return operatorops.Deployment{}, err
	}
	cfg := loaded.Deployment()
	postgresSettings, err := postgres.ResolveSettings(configassembly.PostgresBinding(cfg), nil)
	if err != nil {
		return operatorops.Deployment{}, err
	}
	objectSettings, err := objectstore.ResolveSettings(configassembly.ObjectStoreBinding(cfg), nil)
	if err != nil {
		return operatorops.Deployment{}, err
	}
	return operatorops.Deployment{
		DatabaseStorage: operatorops.RootBinding{
			BindingKind: cfg.Roots.DatabaseStorage.BindingKind,
			Path:        cfg.Roots.DatabaseStorage.Path,
			ServiceRef:  cfg.Roots.DatabaseStorage.ServiceRef,
		},
		ObjectStorage: operatorops.RootBinding{
			BindingKind: cfg.Roots.ObjectStorage.BindingKind,
			Path:        cfg.Roots.ObjectStorage.Path,
			ServiceRef:  cfg.Roots.ObjectStorage.ServiceRef,
		},
		BackupStorage: operatorops.RootBinding{
			BindingKind: cfg.Roots.BackupStorage.BindingKind,
			Path:        cfg.Roots.BackupStorage.Path,
			ServiceRef:  cfg.Roots.BackupStorage.ServiceRef,
		},
		PostgresSettings: postgresSettings,
		ObjectSettings:   objectSettings,
		OpenPostgres: func(ctx context.Context) (operatorops.PostgresPool, error) {
			return runner.setupPostgres(ctx, postgresSettings)
		},
		OpenObjectStore: func(ctx context.Context) (objectstore.Store, error) {
			return runner.setupObjectStore(ctx, objectSettings, configassembly.ObjectStoreInstrumentation(cfg))
		},
		OpenBackup: func() (recovery.BackupStorage, error) {
			return runner.newBackupStorage(
				cfg.Roots.BackupStorage.BindingKind,
				cfg.Roots.BackupStorage.Path,
			)
		},
	}, nil
}
