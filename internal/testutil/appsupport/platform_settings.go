package appsupport

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func OpenPostgres(ctx context.Context, cfg configassembly.Deployment, env map[string]string) (postgres.AdmittedPool, error) {
	settings, err := postgres.ResolveSettings(configassembly.PostgresBinding(cfg), postgres.PurposeRuntime, env)
	if err != nil {
		return nil, err
	}
	return postgres.Setup(ctx, settings)
}

func ObjectStoreSettings(cfg configassembly.Deployment, env map[string]string) (objectstore.Settings, error) {
	return objectstore.ResolveSettings(configassembly.ObjectStoreBinding(cfg), env)
}

func OpenObjectStore(ctx context.Context, cfg configassembly.Deployment, env map[string]string) (objectstore.Store, error) {
	settings, err := ObjectStoreSettings(cfg, env)
	if err != nil {
		return nil, err
	}
	return objectstore.Setup(ctx, settings, configassembly.ObjectStoreInstrumentation(cfg))
}

// RecognizedExtensionJobDefinitions uses the production composition and typed
// package registry for service-backed owner tests.
func RecognizedExtensionJobDefinitions(t testing.TB) []jobs.Definition {
	t.Helper()
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	definitions, err := extensionassembly.RecognizedJobDefinitions(coordinator.JobKindContracts(), coordinator.WorkerRuntimeContracts())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}
