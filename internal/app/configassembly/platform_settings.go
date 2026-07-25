package configassembly

import (
	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func PostgresBinding(cfg Deployment) postgres.Binding {
	return postgres.Binding{
		BindingKind: cfg.Roots.DatabaseStorage.BindingKind,
		RootPath:    cfg.Roots.DatabaseStorage.Path,
		ServiceRef:  cfg.Roots.DatabaseStorage.ServiceRef,
	}
}

func ObjectStoreBinding(cfg Deployment) objectstore.Binding {
	return objectstore.Binding{
		BindingKind: cfg.Roots.ObjectStorage.BindingKind,
		RootPath:    cfg.Roots.ObjectStorage.Path,
		ServiceRef:  cfg.Roots.ObjectStorage.ServiceRef,
	}
}

func ObjectStoreInstrumentation(cfg Deployment) objectstore.Instrumentation {
	return objectstore.Instrumentation{
		Enabled:        cfg.Telemetry.Enabled,
		ServiceVersion: cfg.Telemetry.Resource.ServiceVersion,
	}
}

func BootstrapSettings(cfg Deployment) bootstrap.Settings {
	return bootstrap.Settings{ManifestPath: cfg.Bootstrap.FirstAdminManifestPath}
}
