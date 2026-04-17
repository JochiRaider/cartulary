package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"

	"example.com/todo/cartulary/internal/platform/config"
	"example.com/todo/cartulary/internal/platform/httpapi"
	"example.com/todo/cartulary/internal/platform/jobs"
	"example.com/todo/cartulary/internal/platform/objectstore"
	"example.com/todo/cartulary/internal/platform/postgres"
)

type Options struct {
	Env         map[string]string
	HTTP        httpapi.Options
	Postgres    *pgxpool.Pool
	ObjectStore *minio.Client
}

type Runtime struct {
	Config      config.Config
	Handler     http.Handler
	Postgres    *pgxpool.Pool
	ObjectStore *minio.Client
	Jobs        *jobs.Manager
}

func NewRuntime(ctx context.Context, cfg config.Config, options Options) (*Runtime, error) {
	runtime := &Runtime{
		Config: cfg,
		Jobs:   jobs.NewManager(),
	}

	if options.Postgres != nil {
		runtime.Postgres = options.Postgres
	} else {
		pool, err := postgres.SetupWithEnv(ctx, cfg, options.Env)
		if err != nil {
			return nil, fmt.Errorf("setup postgres: %w", err)
		}
		runtime.Postgres = pool
	}

	if options.ObjectStore != nil {
		runtime.ObjectStore = options.ObjectStore
	} else {
		client, err := objectstore.SetupWithEnv(ctx, cfg, options.Env)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup object store: %w", err)
		}
		runtime.ObjectStore = client
	}

	httpOptions := options.HTTP
	httpOptions.Dependencies = httpapi.DependencySet{
		Config:      cfg,
		Postgres:    runtime.Postgres,
		ObjectStore: runtime.ObjectStore,
		Jobs:        runtime.Jobs,
	}

	handler, err := httpapi.NewHandler(httpOptions)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("setup http handler: %w", err)
	}

	runtime.Handler = handler
	return runtime, nil
}

func (r *Runtime) Close() {
	if r == nil {
		return
	}
	if r.Postgres != nil {
		r.Postgres.Close()
	}
}
