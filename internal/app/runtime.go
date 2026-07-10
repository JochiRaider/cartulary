package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/jobapi"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/viewschemas"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	workbookstartupbootstrap "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/enterpriseauth"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

var (
	newJobsManager   = jobs.NewManager
	setupPostgres    = postgres.SetupWithEnv
	setupObjectStore = objectstore.SetupWithEnv
	runBootstrap     = bootstrap.Preflight
	newWSHub         = platformws.NewHub
	newHTTPHandler   = httpapi.NewHandler
)

type Options struct {
	Env         map[string]string
	HTTP        httpapi.Options
	Postgres    *pgxpool.Pool
	ObjectStore objectstore.Store
	Now         func() time.Time
}

type Runtime struct {
	Config      config.Config
	Handler     http.Handler
	Postgres    *pgxpool.Pool
	ObjectStore objectstore.Store
	Jobs        *jobs.Manager
	JobRunner   *jobs.Runner
	WSHub       *platformws.Hub
	Telemetry   *telemetry.Runtime
}

func NewRuntime(ctx context.Context, cfg config.Config, options Options) (*Runtime, error) {
	normalizedCfg, err := config.ValidateForStartup(cfg)
	if err != nil {
		return nil, err
	}

	profiles := httpapi.ResolveExtensionProfiles(options.HTTP.Dependencies.ExtensionProfiles)
	if options.HTTP.Dependencies.ExtensionProfiles == nil {
		profiles = applyConfigExtensionClaims(profiles, normalizedCfg)
	}
	telemetryRuntime, err := telemetry.Bootstrap(ctx, normalizedCfg, options.Env, telemetry.WithClaimedExtensionProfiles(claimedExtensionProfileIDs(profiles)))
	if err != nil {
		return nil, err
	}

	runtime := &Runtime{
		Config:    normalizedCfg,
		Telemetry: telemetryRuntime,
	}
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
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
		runtime.ObjectStore = instrumentedObjectStore(normalizedCfg, options.ObjectStore)
	} else {
		client, err := setupObjectStore(ctx, normalizedCfg, options.Env)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup object store: %w", err)
		}
		runtime.ObjectStore = client
	}
	postgresHandle := instrumentedPostgres(normalizedCfg, runtime.Postgres)

	if err := runBootstrap(ctx, normalizedCfg, runtime.Postgres); err != nil {
		runtime.Close()
		return nil, err
	}
	if err := enterpriseauth.ReconcileProviderManifest(ctx, normalizedCfg, options.Env, authn.NewStore(postgresHandle), now()); err != nil {
		runtime.Close()
		return nil, err
	}
	if err := reference_data.EnsureMinimumDisconnectedBundle(ctx, normalizedCfg, runtime.Postgres, profiles, now()); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("seed minimum disconnected reference packs: %w", err)
	}
	runtime.Jobs = newJobsManager()
	runtime.Jobs.ConfigureTelemetry(normalizedCfg.Telemetry.Resource.ServiceVersion)
	runtime.JobRunner = jobs.NewRunner()
	hub := newWSHub()
	runtime.WSHub = hub
	runtime.Jobs.Configure(runtime.Postgres, now)
	runtime.Jobs.ConfigureProgressHub(hub)
	hub.ConfigureTelemetry(normalizedCfg.Telemetry.Resource.ServiceVersion)

	httpOptions := options.HTTP
	testRuntimeDeps := httpOptions.Dependencies
	keys, err := authn.LoadMasterKeys(options.Env)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	cursorKey := authn.DerivePurposeKey(keys, "pagination-cursor-v1")
	cursorCodec := pagination.NewCodec(cursorKey[:])
	attributionResolvers := revisions.NewAttributionResolverRegistry()
	if err := attributionResolvers.RegisterImportedAttributionResolver(incidentbundles.IncidentPortabilityProfileID, incidentbundles.ImportedAttributionResolver()); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("register incident portability attribution resolver: %w", err)
	}
	if err := attributionResolvers.ValidateAttributionResolvers(revisionExtensionClaims(profiles)); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("validate attribution resolvers: %w", err)
	}
	incidentRoutes := incidents.RegisterRoutes(incidents.RouteOptions{
		WorkbookBootstrap:    workbookstartupbootstrap.NewIncidentCreatePreferencesPort(),
		CollaborationSession: collaboration.NewIncidentSessionNotifier(postgresHandle, hub),
	})
	incidentBundleImportFinalizer := incidents.NewStoreWithOptions(postgresHandle, incidents.StoreOptions{
		WorkbookBootstrap: workbookstartupbootstrap.NewIncidentCreatePreferencesPort(),
	})
	incidentBundleRoutes := incidentbundles.RegisterRoutes(
		incidentbundles.WithImportFinalizer(incidentBundleImportFinalizer),
	)
	revisionCommands, err := NewRevisionsCommandService(postgresHandle, attributionResolvers.ImportedAttributionResolver(incidentbundles.IncidentPortabilityProfileID))
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose revisions command service: %w", err)
	}
	revisionRoutes := revisions.RegisterRoutes(revisionCommands)
	httpOptions.AdditionalRoutes = append([]httpapi.RouteRegistrar{auth.RegisterRoutes(), incidentRoutes, extensions.RegisterRoutes(), jobapi.RegisterRoutes(), imports.RegisterRoutes(), reporting.RegisterRoutes(), reportcomposition.RegisterRoutes(), reference_data.RegisterRoutes(), incidentBundleRoutes, savedviews.RegisterRoutes(), viewschemas.RegisterRoutes(), collaboration.RegisterRoutes(), entities.RegisterRoutes(), evidence.RegisterRoutes(), assessments.RegisterRoutes(), workbook.RegisterRoutes(), timeline.RegisterRoutes(), revisionRoutes}, httpOptions.AdditionalRoutes...)
	httpOptions.Dependencies = httpapi.DependencySet{
		Config:            normalizedCfg,
		Env:               options.Env,
		Postgres:          runtime.Postgres,
		PostgresDB:        postgresHandle,
		ObjectStore:       runtime.ObjectStore,
		Jobs:              runtime.Jobs,
		JobRunner:         runtime.JobRunner,
		WSHub:             hub,
		CursorCodec:       cursorCodec,
		Readiness:         httpapi.NewDependencyReadinessChecker(runtime.Postgres, runtime.ObjectStore),
		PublicErrorFaults: testRuntimeDeps.PublicErrorFaults,
		ModuleOverrides:   testRuntimeDeps.ModuleOverrides,
		ExtensionProfiles: profiles,
		Now:               now,
	}

	handler, err := newHTTPHandler(httpOptions)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("setup http handler: %w", err)
	}

	runtime.Handler = handler
	return runtime, nil
}

func instrumentedPostgres(cfg config.Config, pool *pgxpool.Pool) postgres.DB {
	if pool == nil || !cfg.Telemetry.Enabled {
		return pool
	}
	return postgres.InstrumentDB(pool, cfg.Telemetry.Resource.ServiceVersion)
}

func instrumentedObjectStore(cfg config.Config, store objectstore.Store) objectstore.Store {
	if store == nil || !cfg.Telemetry.Enabled {
		return store
	}
	return objectstore.InstrumentStore(store, cfg.Telemetry.Resource.ServiceVersion)
}

func claimedExtensionProfileIDs(profiles []httpapi.ExtensionProfile) []string {
	claimed := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		if profile.Claimed {
			claimed = append(claimed, profile.ProfileID)
		}
	}
	return claimed
}

func applyConfigExtensionClaims(profiles []httpapi.ExtensionProfile, cfg config.Config) []httpapi.ExtensionProfile {
	claimed := append([]httpapi.ExtensionProfile(nil), profiles...)
	for index := range claimed {
		if claimed[index].ProfileID == "enterprise_authentication" {
			claimed[index].Claimed = cfg.EnterpriseAuthentication.Claimed
		}
	}
	return claimed
}

func revisionExtensionClaims(profiles []httpapi.ExtensionProfile) []revisions.ExtensionClaim {
	claims := make([]revisions.ExtensionClaim, 0, len(profiles))
	for _, profile := range profiles {
		claims = append(claims, revisions.ExtensionClaim{
			ProfileID: profile.ProfileID,
			Claimed:   profile.Claimed,
		})
	}
	return claims
}

func (r *Runtime) Close() {
	if r == nil {
		return
	}
	if r.JobRunner != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = r.JobRunner.Close(ctx)
		cancel()
	}
	if r.ObjectStore != nil {
		_ = r.ObjectStore.Close()
	}
	if r.Telemetry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Config.Telemetry.Shutdown.FlushTimeoutMS)*time.Millisecond)
		_ = r.Telemetry.Shutdown(ctx)
		cancel()
	}
	if r.Postgres != nil {
		r.Postgres.Close()
	}
}
