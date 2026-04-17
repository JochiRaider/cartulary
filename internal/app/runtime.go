package app

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"

	"example.com/todo/cartulary/internal/modules/auth"
	"example.com/todo/cartulary/internal/modules/incidents"
	"example.com/todo/cartulary/internal/platform/config"
	"example.com/todo/cartulary/internal/platform/httpapi"
	"example.com/todo/cartulary/internal/platform/jobs"
	"example.com/todo/cartulary/internal/platform/objectstore"
	"example.com/todo/cartulary/internal/platform/postgres"
	platformws "example.com/todo/cartulary/internal/platform/ws"
)

var (
	newJobsManager   = jobs.NewManager
	setupPostgres    = postgres.SetupWithEnv
	setupObjectStore = objectstore.SetupWithEnv
	newHTTPHandler   = httpapi.NewHandler
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
	normalizedCfg, err := config.ValidateForStartup(cfg)
	if err != nil {
		return nil, err
	}

	runtime := &Runtime{
		Config: normalizedCfg,
	}

	if options.Postgres != nil {
		runtime.Postgres = options.Postgres
	} else {
		pool, err := setupPostgres(ctx, normalizedCfg, options.Env)
		if err != nil {
			return nil, fmt.Errorf("setup postgres: %w", err)
		}
		runtime.Postgres = pool
	}

	if options.ObjectStore != nil {
		runtime.ObjectStore = options.ObjectStore
	} else {
		client, err := setupObjectStore(ctx, normalizedCfg, options.Env)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup object store: %w", err)
		}
		runtime.ObjectStore = client
	}

	if err := bootstrapPreflight(ctx, normalizedCfg, postgresBootstrapStore{pool: runtime.Postgres}, os.ReadFile, deriveBootstrapPasswordHash); err != nil {
		runtime.Close()
		return nil, err
	}

	runtime.Jobs = newJobsManager()
	hub := platformws.NewHub()

	httpOptions := options.HTTP
	httpOptions.AdditionalRoutes = append([]httpapi.RouteRegistrar{auth.RegisterRoutes(), incidents.RegisterRoutes()}, httpOptions.AdditionalRoutes...)
	httpOptions.Dependencies = httpapi.DependencySet{
		Config:      normalizedCfg,
		Env:         options.Env,
		Postgres:    runtime.Postgres,
		ObjectStore: runtime.ObjectStore,
		Jobs:        runtime.Jobs,
		WSHub:       hub,
	}

	handler, err := newHTTPHandler(httpOptions)
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
