package configassembly

import (
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func RecoveryTargetBindings(cfg Deployment, env map[string]string) (recovery.TargetBindingDigests, error) {
	database, err := postgres.ResolveSettings(PostgresBinding(cfg), postgres.PurposeRuntime, env)
	if err != nil {
		return recovery.TargetBindingDigests{}, err
	}
	objects, err := objectstore.ResolveSettings(ObjectStoreBinding(cfg), env)
	if err != nil {
		return recovery.TargetBindingDigests{}, err
	}
	return recovery.TargetBindingDigestsFor(recovery.TargetBindingInputs{
		DatabaseStorage:      recovery.RootBinding{BindingKind: cfg.Roots.DatabaseStorage.BindingKind, Path: cfg.Roots.DatabaseStorage.Path, ServiceRef: cfg.Roots.DatabaseStorage.ServiceRef},
		ObjectStorage:        recovery.RootBinding{BindingKind: cfg.Roots.ObjectStorage.BindingKind, Path: cfg.Roots.ObjectStorage.Path, ServiceRef: cfg.Roots.ObjectStorage.ServiceRef},
		ReferencePackStorage: recovery.RootBinding{BindingKind: cfg.Roots.ReferencePackStorage.BindingKind, Path: cfg.Roots.ReferencePackStorage.Path, ServiceRef: cfg.Roots.ReferencePackStorage.ServiceRef},
		ExportOutputs:        recovery.RootBinding{BindingKind: cfg.Roots.ExportOutputs.BindingKind, Path: cfg.Roots.ExportOutputs.Path, ServiceRef: cfg.Roots.ExportOutputs.ServiceRef},
		PostgresSettings:     database, ObjectSettings: objects,
	}), nil
}
