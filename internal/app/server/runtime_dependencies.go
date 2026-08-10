package server

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/platform/securefile"
)

type runtimeDependencies struct {
	newJobsManager                 func(jobs.ManagerOptions) (*jobs.Manager, error)
	setupPostgres                  func(context.Context, postgres.Settings) (*pgxpool.Pool, error)
	ensureSchemaReady              func(context.Context, *pgxpool.Pool, *database_migrations.Source) error
	setupObjectStore               func(context.Context, objectstore.Settings, objectstore.Instrumentation) (objectstore.Store, error)
	runBootstrap                   func(context.Context, bootstrap.Settings, *pgxpool.Pool) error
	newCollaborationHub            func() *collaboration.Hub
	newHTTPHandler                 func(...httpapi.Options) (http.Handler, error)
	readSecureFile                 secureDocumentReader
	acquireApplicationProcessLease func(context.Context, *pgxpool.Pool, time.Duration, time.Duration) (*processlease.ApplicationProcessLease, error)
	acquireRecoveryServingLease    func(context.Context, *pgxpool.Pool, time.Duration, time.Duration) (*processlease.ApplicationRecoveryServingLease, error)
}

type secureDocumentReader func(string, int64) (securefile.Document, error)

func productionRuntimeDependencies() runtimeDependencies {
	return runtimeDependencies{
		newJobsManager:                 jobs.NewManager,
		setupPostgres:                  postgres.Setup,
		ensureSchemaReady:              database_migrations.EnsureSchemaReady,
		setupObjectStore:               objectstore.Setup,
		runBootstrap:                   bootstrap.Preflight,
		newCollaborationHub:            collaboration.NewHub,
		newHTTPHandler:                 httpapi.NewHandler,
		readSecureFile:                 securefile.Read,
		acquireApplicationProcessLease: processlease.AcquireApplicationProcess,
		acquireRecoveryServingLease:    processlease.AcquireApplicationRecoveryServing,
	}
}
