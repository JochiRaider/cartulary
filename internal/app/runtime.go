package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

var (
	newJobsManager   = jobs.NewManager
	setupPostgres    = postgres.SetupWithEnv
	setupObjectStore = objectstore.SetupWithEnv
	runBootstrap     = bootstrapPreflight
	newWSHub         = platformws.NewHub
	newHTTPHandler   = httpapi.NewHandler
)

type Options struct {
	Env         map[string]string
	HTTP        httpapi.Options
	Postgres    *pgxpool.Pool
	ObjectStore *minio.Client
	Now         func() time.Time
}

type Runtime struct {
	Config      config.Config
	Handler     http.Handler
	Postgres    *pgxpool.Pool
	ObjectStore *minio.Client
	Jobs        *jobs.Manager
	WSHub       *platformws.Hub
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

	if err := runBootstrap(ctx, normalizedCfg, postgresBootstrapStore{pool: runtime.Postgres}, osBootstrapManifestFS{}, deriveBootstrapPasswordHash); err != nil {
		runtime.Close()
		return nil, err
	}

	runtime.Jobs = newJobsManager()
	hub := newWSHub()
	runtime.WSHub = hub
	paginator := pagination.NewRegistry()

	httpOptions := options.HTTP
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	httpOptions.AdditionalRoutes = append([]httpapi.RouteRegistrar{auth.RegisterRoutes(), incidents.RegisterRoutes(), entities.RegisterRoutes(), timeline.RegisterRoutes()}, httpOptions.AdditionalRoutes...)
	httpOptions.Dependencies = httpapi.DependencySet{
		Config:      normalizedCfg,
		Env:         options.Env,
		Postgres:    runtime.Postgres,
		ObjectStore: runtime.ObjectStore,
		Jobs:        runtime.Jobs,
		WSHub:       hub,
		Pagination:  paginator,
		Now:         now,
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
