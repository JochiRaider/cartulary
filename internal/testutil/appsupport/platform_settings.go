package appsupport

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func OpenPostgres(ctx context.Context, cfg configassembly.Deployment, env map[string]string) (*pgxpool.Pool, error) {
	settings, err := postgres.ResolveSettings(configassembly.PostgresBinding(cfg), env)
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
