package operator

import (
	"context"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/operator/recoverycli"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/recoveryprovider"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/platform/workbookprobe"
)

type recoveryExecutor struct {
	transport        operatorTransport
	loadConfig       func(string) (configassembly.Loaded, error)
	setupPostgres    func(context.Context, postgres.Settings) (operatorPostgresPool, error)
	setupObjectStore func(context.Context, objectstore.Settings, objectstore.Instrumentation) (objectstore.Store, error)
	newBackupStorage func(string, string) (recovery.BackupStorage, error)
	now              func() time.Time
}

func (executor recoveryExecutor) runCLI(ctx context.Context, args []string) (bool, int) {
	extensionBackups, err := extensionassembly.GeneratedRecoveryCatalog()
	if err != nil {
		executor.transport.logger().Error("extension recovery catalog is invalid", "error", err)
		return true, 1
	}
	recoveryStateCatalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		executor.transport.logger().Error("recovery state catalog is invalid", "error", err)
		return true, 1
	}
	return recoverycli.Runner{
		Stdout: executor.transport.stdout,
		Stderr: executor.transport.stderr,
		Now:    executor.now,
		Facade: application.Service{
			LoadDeployment: executor.loadDeployment,
			ReadTargetMarker: func(bindingKind string, rootPath string) (application.TargetMarkerMaterial, error) {
				if bindingKind != "filesystem_root" {
					return application.TargetMarkerMaterial{}, application.ErrTargetMarkerRequiresFilesystemStorage
				}
				storage, err := recoveryassembly.NewFilesystemStorage(rootPath)
				if err != nil {
					return application.TargetMarkerMaterial{}, err
				}
				defer storage.Close()
				markerBody, generationBody, err := storage.ReadTargetMarker(
					application.RestoreTargetMarkerMaximumBytes,
					application.RestoreTargetGenerationMaximumBytes,
				)
				return application.TargetMarkerMaterial{
					MarkerBody:     markerBody,
					GenerationBody: generationBody,
				}, err
			},
			NewProjectionServices: func(db postgres.DB) (restorecontract.ProjectionRebuilder, workbookprobe.Executor, error) {
				return timelineassembly.NewRecoveryProjectionServices(db)
			},
			NewEvidenceProvider: func(db postgres.DB) recovery.EvidenceRecoveryProvider {
				return recoveryprovider.New(db)
			},
			NewEvidenceRepository: func(pool application.PostgresPool) (application.RecoveryEvidenceRepository, error) {
				return recoveryassembly.NewRecoveryEvidenceRepository(pool, func() (recovery.RecoveryEncryptionKey, error) {
					return recovery.LoadRecoveryEncryptionKey(nil)
				})
			},
			NewTargetAdmission:     recoveryassembly.AcquireTargetServingAdmission,
			ProjectFailureEvidence: recoverycli.FailureEvidenceFields,
			ExtensionBackups:       extensionBackups,
			RecoveryStateCatalog:   recoveryStateCatalog,
			ValidateStateCoverage:  recoveryassembly.ValidateRecoveryStateDatabaseCoverage,
			NewVNextCapture: func(
				pool application.PostgresPool,
				objects objectstore.Store,
				storage recovery.BackupStorage,
				state *recoverystate.Catalog,
			) (*recovery.VNextCaptureService, error) {
				streaming, err := recovery.RequireStreamingBackupStorage(storage)
				if err != nil {
					return nil, err
				}
				inventories, err := recoveryassembly.CurrentVNextObjectInventoryCatalog(
					recoveryassembly.NewVNextObjectSource(objects),
				)
				if err != nil {
					return nil, err
				}
				return recovery.NewVNextCaptureService(
					recoveryassembly.NewVNextSnapshotRepository(pool),
					streaming,
					state,
					inventories,
				)
			},
			Now: executor.now,
		},
	}.Run(ctx, args)
}

func (executor recoveryExecutor) loadDeployment(path string) (application.Deployment, error) {
	loaded, err := executor.loadConfig(path)
	if err != nil {
		return application.Deployment{}, err
	}
	cfg := loaded.Deployment()
	postgresSettings, err := postgres.ResolveSettings(configassembly.PostgresBinding(cfg), nil)
	if err != nil {
		return application.Deployment{}, err
	}
	objectSettings, err := objectstore.ResolveSettings(configassembly.ObjectStoreBinding(cfg), nil)
	if err != nil {
		return application.Deployment{}, err
	}
	return application.Deployment{
		DatabaseStorage: application.RootBinding{
			BindingKind: cfg.Roots.DatabaseStorage.BindingKind,
			Path:        cfg.Roots.DatabaseStorage.Path,
			ServiceRef:  cfg.Roots.DatabaseStorage.ServiceRef,
		},
		ObjectStorage: application.RootBinding{
			BindingKind: cfg.Roots.ObjectStorage.BindingKind,
			Path:        cfg.Roots.ObjectStorage.Path,
			ServiceRef:  cfg.Roots.ObjectStorage.ServiceRef,
		},
		BackupStorage: application.RootBinding{
			BindingKind: cfg.Roots.BackupStorage.BindingKind,
			Path:        cfg.Roots.BackupStorage.Path,
			ServiceRef:  cfg.Roots.BackupStorage.ServiceRef,
		},
		PostgresSettings: postgresSettings,
		ObjectSettings:   objectSettings,
		OpenPostgres: func(ctx context.Context) (application.PostgresPool, error) {
			return executor.setupPostgres(ctx, postgresSettings)
		},
		OpenObjectStore: func(ctx context.Context) (objectstore.Store, error) {
			return executor.setupObjectStore(ctx, objectSettings, configassembly.ObjectStoreInstrumentation(cfg))
		},
		OpenBackup: func() (recovery.BackupStorage, error) {
			return executor.newBackupStorage(
				cfg.Roots.BackupStorage.BindingKind,
				cfg.Roots.BackupStorage.Path,
			)
		},
		ServingLeaseAcquireTimeout: time.Duration(cfg.Timeouts.Extensions.ProcessLeaseAcquireSeconds) * time.Second,
		ServingLeaseLossDetection:  time.Duration(cfg.Timeouts.Extensions.ProcessLeaseLossDetectionSeconds) * time.Second,
	}, nil
}
