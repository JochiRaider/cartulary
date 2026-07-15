package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/jobapi"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
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
	"github.com/JochiRaider/cartulary/internal/platform/secretpurpose"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

var (
	newJobsManager    = jobs.NewManager
	setupPostgres     = postgres.SetupWithEnv
	ensureSchemaReady = postgres.EnsureSchemaReady
	setupObjectStore  = objectstore.SetupWithEnv
	runBootstrap      = bootstrap.Preflight
	newWSHub          = platformws.NewHub
	newHTTPHandler    = httpapi.NewHandler
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

	closeOnce sync.Once
	cleanups  []func()
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
	secretPurposes := secretpurpose.NewRegistry()
	if err := authn.RegisterMasterSecretPurpose(secretPurposes, options.Env); err != nil {
		return nil, err
	}
	if err := config.RegisterTelemetrySecretPurposes(normalizedCfg, options.Env, secretPurposes); err != nil {
		return nil, err
	}
	var enterpriseProviderDefinitions []authn.EnterpriseAuthProviderDefinition
	if normalizedCfg.EnterpriseAuthentication.Claimed {
		enterpriseProviderDefinitions, err = enterpriseauth.LoadProviderManifest(normalizedCfg, options.Env)
		if err != nil {
			return nil, err
		}
		if err := enterpriseauth.RegisterProviderSecretPurposes(enterpriseProviderDefinitions, options.Env, secretPurposes); err != nil {
			return nil, err
		}
	}
	telemetryRuntime, err := telemetry.Bootstrap(ctx, normalizedCfg, options.Env, telemetry.WithClaimedExtensionProfiles(claimedExtensionProfileIDs(profiles)))
	if err != nil {
		return nil, err
	}

	runtime := &Runtime{
		Config:    normalizedCfg,
		Telemetry: telemetryRuntime,
	}
	runtime.own(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runtime.Config.Telemetry.Shutdown.FlushTimeoutMS)*time.Millisecond)
		defer cancel()
		_ = runtime.Telemetry.Shutdown(ctx)
	})
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	if options.Postgres != nil {
		runtime.Postgres = options.Postgres
	} else {
		pool, err := setupPostgres(ctx, normalizedCfg, options.Env)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup postgres: %w", err)
		}
		runtime.Postgres = pool
		if pool != nil {
			runtime.own(pool.Close)
		}
	}

	if err := ensureSchemaReady(ctx, runtime.Postgres, dbmigrations.Source()); err != nil {
		runtime.Close()
		return nil, err
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
		if client != nil {
			runtime.own(func() { _ = client.Close() })
		}
	}
	postgresHandle := instrumentedPostgres(normalizedCfg, runtime.Postgres)

	if err := runBootstrap(ctx, normalizedCfg, runtime.Postgres); err != nil {
		runtime.Close()
		return nil, err
	}
	if normalizedCfg.EnterpriseAuthentication.Claimed {
		if err := enterpriseauth.ReconcileProviderDefinitions(ctx, enterpriseProviderDefinitions, authn.NewStore(postgresHandle), now()); err != nil {
			runtime.Close()
			return nil, err
		}
	}
	networkFlowKeyRings, err := networkflow.LoadKeyRingsWithRegistry(normalizedCfg, options.Env, now(), secretPurposes)
	if err != nil {
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
	jobRunner := runtime.JobRunner
	runtime.own(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = jobRunner.Close(ctx)
	})
	hub := newWSHub()
	runtime.WSHub = hub
	runtime.Jobs.Configure(runtime.Postgres, now)
	runtime.Jobs.ConfigureProgressHub(hub)
	runtime.JobRunner.Configure(runtime.Jobs)
	hub.ConfigureTelemetry(normalizedCfg.Telemetry.Resource.ServiceVersion)

	httpOptions := options.HTTP
	testRuntimeDeps := httpOptions.Dependencies
	if testRings, ok := testRuntimeDeps.ModuleOverrides[networkflow.KeyRingsOverrideKey].(*networkflow.KeyRings); ok && testRings != nil {
		networkFlowKeyRings = testRings
	}
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
	revisionCommands, err := revisionassembly.NewCommandService(postgresHandle, attributionResolvers.ImportedAttributionResolver(incidentbundles.IncidentPortabilityProfileID))
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose revisions command service: %w", err)
	}
	revisionRoutes := revisions.RegisterRoutes(revisionCommands)
	networkFlowModule, err := networkflow.NewModule(networkflow.ModuleDependencies{
		Postgres:      postgresHandle,
		ImportSources: imports.NewStore(runtime.Postgres),
		KeyRings:      networkFlowKeyRings,
		Now:           now,
		Transactions:  postgres.NewTransactionRunner(postgresHandle),
		IncidentLocks: incidents.NewTransactionParticipant(),
		AuditAppender: authn.NewAdministrativeAuditAppender(),
		Indicators:    indicators.NewStore(postgresHandle),
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Network Flow module: %w", err)
	}
	moduleOverrides := mergeNetworkFlowImportFacadeOverride(testRuntimeDeps.ModuleOverrides, networkFlowModule.ImportOwner())
	delete(moduleOverrides, networkflow.KeyRingsOverrideKey)
	builtInRoutes, err := builtInRouteRegistrars([]routeContribution{
		{id: "auth", registrar: auth.RegisterRoutes()},
		{id: "incidents", registrar: incidentRoutes},
		{id: "extensions", registrar: extensions.RegisterRoutes()},
		{id: "jobs", registrar: jobapi.RegisterRoutes()},
		{id: "imports", registrar: imports.RegisterRoutes()},
		{id: "network_flow", registrar: networkFlowModule.RegisterRoutes()},
		{id: "reporting", registrar: reporting.RegisterRoutes()},
		{id: "report_composition", registrar: reportcomposition.RegisterRoutes()},
		{id: "reference_data", registrar: reference_data.RegisterRoutes()},
		{id: "incident_bundles", registrar: incidentBundleRoutes},
		{id: "saved_views", registrar: savedviews.RegisterRoutes()},
		{id: "view_schemas", registrar: viewschemas.RegisterRoutes()},
		{id: "collaboration", registrar: collaboration.RegisterRoutes()},
		{id: "entities", registrar: entities.RegisterRoutes()},
		{id: "evidence", registrar: evidence.RegisterRoutes()},
		{id: "assessments", registrar: assessments.RegisterRoutes()},
		{id: "workbook", registrar: workbook.RegisterRoutes()},
		{id: "timeline", registrar: timeline.RegisterRoutes()},
		{id: "revisions", registrar: revisionRoutes},
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose built-in routes: %w", err)
	}
	httpOptions.AdditionalRoutes = append(builtInRoutes, httpOptions.AdditionalRoutes...)
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
		ModuleOverrides:   moduleOverrides,
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

func mergeNetworkFlowImportFacadeOverride(overrides map[string]any, facade imports.ExtensionImportApplyFacade) map[string]any {
	merged := map[string]any{}
	for key, value := range overrides {
		merged[key] = value
	}
	facades := map[string]imports.ExtensionImportApplyFacade{}
	if existing, ok := overrides[imports.ExtensionApplyFacadesOverrideKey]; ok && existing != nil {
		typed, ok := existing.(map[string]imports.ExtensionImportApplyFacade)
		if !ok {
			return merged
		}
		for key, value := range typed {
			facades[key] = value
		}
	}
	if facade != nil {
		facades[imports.ExtensionApplyFacadeKey(imports.ImportTargetKindNetworkFlowTable, imports.NetworkFlowExtensionProfileID)] = facade
	}
	merged[imports.ExtensionApplyFacadesOverrideKey] = facades
	return merged
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
		switch claimed[index].ProfileID {
		case "enterprise_authentication":
			claimed[index].Claimed = cfg.EnterpriseAuthentication.Claimed
		case "network_flow_activity":
			claimed[index].Claimed = cfg.NetworkFlowActivity.Claimed
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
	r.closeOnce.Do(func() {
		for index := len(r.cleanups) - 1; index >= 0; index-- {
			r.cleanups[index]()
		}
	})
}

func (r *Runtime) own(cleanup func()) {
	if r != nil && cleanup != nil {
		r.cleanups = append(r.cleanups, cleanup)
	}
}
