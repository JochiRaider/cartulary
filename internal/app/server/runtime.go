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
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi/webassets"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
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
	Config         config.Config
	Handler        http.Handler
	Extensions     *extensions.Coordinator
	ExtensionState *extensions.StateRuntime
	StagedJanitor  *extensions.StagedObjectJanitor
	ExtensionPlan  extensions.PublicationPlan
	ResolvedClaims extensions.ResolvedClaimSet
	Postgres       *pgxpool.Pool
	ObjectStore    objectstore.Store
	Jobs           *jobs.Manager
	JobRunner      *jobs.Runner
	WSHub          *platformws.Hub
	Telemetry      *telemetry.Runtime
	ProcessLease   *processlease.Lease
	Lifecycle      *processlifecycle.Controller

	closeOnce            sync.Once
	publicationOnce      sync.Once
	cleanups             []func()
	stagedJanitorContext context.Context
}

func NewRuntime(ctx context.Context, cfg config.Config, options Options) (*Runtime, error) {
	normalizedCfg, err := config.ValidateForStartup(cfg)
	if err != nil {
		return nil, err
	}

	runtime := &Runtime{
		Config:    normalizedCfg,
		Lifecycle: processlifecycle.New(),
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
			runtime.Close()
			return nil, fmt.Errorf("setup postgres: %w", err)
		}
		runtime.Postgres = pool
		if pool != nil {
			runtime.own(pool.Close)
		}
	}
	if runtime.Postgres != nil {
		lease, leaseErr := processlease.Acquire(
			ctx,
			processlease.PostgresBackend{Pool: runtime.Postgres},
			time.Duration(normalizedCfg.Timeouts.Extensions.ProcessLeaseAcquireSeconds)*time.Second,
			time.Duration(normalizedCfg.Timeouts.Extensions.ProcessLeaseLossDetectionSeconds)*time.Second,
		)
		if leaseErr != nil {
			runtime.Close()
			return nil, leaseErr
		}
		runtime.ProcessLease = lease
		runtime.own(func() {
			releaseCtx, cancel := context.WithTimeout(context.Background(), time.Duration(normalizedCfg.Timeouts.Extensions.ProcessLeaseLossDetectionSeconds)*time.Second)
			defer cancel()
			if lease.State() == processlease.StateHeld {
				_ = lease.Release(releaseCtx)
			} else {
				lease.Close()
			}
		})
		monitorCtx, cancelMonitor := context.WithCancel(context.Background())
		runtime.own(cancelMonitor)
		lease.StartMonitor(monitorCtx)
		go runtime.watchProcessLease(monitorCtx)
	}

	extensionCoordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("admit packaged extension registry: %w", err)
	}
	if clientSupportDigest, present, digestErr := webassets.ClientSupportRegistrySHA256(); digestErr != nil {
		runtime.Close()
		return nil, fmt.Errorf("admit packaged browser contracts: %w", digestErr)
	} else if present {
		extensionCoordinator, err = extensionCoordinator.WithClientSupportRegistrySHA256(clientSupportDigest)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("bind packaged browser contracts: %w", err)
		}
	}
	requestedClaims := configuredExtensionClaimIDs(extensionCoordinator.Descriptors(), normalizedCfg, options.HTTP.Dependencies.ExtensionProfiles)
	claimResolution, err := extensionCoordinator.ResolveClaims(requestedClaims)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("resolve extension claims: %w", err)
	}
	extensionPlan, err := extensionCoordinator.BuildPublicationPlan(claimResolution)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("prepare extension publication: %w", err)
	}
	resolvedClaims := claimResolution.Claims()
	profiles := extensionHTTPProfiles(extensionCoordinator.Descriptors(), resolvedClaims.ProfileIDs())
	runtime.Extensions = extensionCoordinator
	runtime.ExtensionPlan = extensionPlan
	runtime.ResolvedClaims = resolvedClaims

	secretPurposes := secretpurpose.NewRegistry()
	if err := authn.RegisterMasterSecretPurpose(secretPurposes, options.Env); err != nil {
		runtime.Close()
		return nil, err
	}
	if err := config.RegisterTelemetrySecretPurposes(normalizedCfg, options.Env, secretPurposes); err != nil {
		runtime.Close()
		return nil, err
	}
	var enterpriseProviderDefinitions []authn.EnterpriseAuthProviderDefinition
	if normalizedCfg.EnterpriseAuthentication.Claimed {
		enterpriseProviderDefinitions, err = enterpriseauth.LoadProviderManifest(normalizedCfg, options.Env)
		if err != nil {
			runtime.Close()
			return nil, err
		}
		if err := enterpriseauth.RegisterProviderSecretPurposes(enterpriseProviderDefinitions, options.Env, secretPurposes); err != nil {
			runtime.Close()
			return nil, err
		}
	}
	telemetryRuntime, err := telemetry.Bootstrap(ctx, normalizedCfg, options.Env, telemetry.WithResolvedClaimIdentity(telemetry.ResolvedClaimIdentity{
		ProfileIDs: resolvedClaims.ProfileIDs(),
		SHA256:     resolvedClaims.SHA256(),
	}))
	if err != nil {
		runtime.Close()
		return nil, err
	}
	runtime.Telemetry = telemetryRuntime
	runtime.own(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(runtime.Config.Telemetry.Shutdown.FlushTimeoutMS)*time.Millisecond)
		defer cancel()
		_ = runtime.Telemetry.Shutdown(shutdownCtx)
	})

	if err := ensureSchemaReady(ctx, runtime.Postgres, dbmigrations.Source()); err != nil {
		runtime.Close()
		return nil, err
	}
	var extensionStateStore *extensionstore.Store
	if runtime.Postgres != nil {
		stateStore, stateStoreErr := extensionstore.New(runtime.Postgres, networkflow.ExtensionStateFamilyCounters())
		if stateStoreErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose extension state store: %w", stateStoreErr)
		}
		stateRuntime, stateRuntimeErr := extensions.NewStateRuntime(
			stateStore,
			map[string]extensions.StateValidator{
				"network_flow_activity.validate_state_v1": networkflow.ValidateExtensionState,
			},
			nil,
			now,
			time.Duration(normalizedCfg.Timeouts.Extensions.MigrationLockSeconds)*time.Second,
		)
		if stateRuntimeErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose extension state runtime: %w", stateRuntimeErr)
		}
		if stateAdmissionErr := stateRuntime.AdmitClaims(ctx, extensionCoordinator, resolvedClaims); stateAdmissionErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("admit extension state: %w", stateAdmissionErr)
		}
		runtime.ExtensionState = stateRuntime
		extensionStateStore = stateStore
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
	if extensionStateStore != nil && runtime.ObjectStore != nil {
		janitor, janitorErr := extensions.NewStagedObjectJanitor(
			extensionStateStore,
			stagedObjectDeleter{store: runtime.ObjectStore},
			nil,
			now,
			int(normalizedCfg.Limits.Extensions.StagedObjectCleanupBatch),
		)
		if janitorErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose staged-object janitor: %w", janitorErr)
		}
		cleanupCtx, cancelCleanup := context.WithTimeout(ctx, time.Duration(normalizedCfg.Timeouts.Extensions.StagedObjectCleanupSeconds)*time.Second)
		cleanupErr := janitor.Sweep(cleanupCtx)
		cancelCleanup()
		if cleanupErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("initial staged-object cleanup: %w", cleanupErr)
		}
		janitorCtx, cancelJanitor := context.WithCancel(context.Background())
		runtime.StagedJanitor = janitor
		runtime.stagedJanitorContext = janitorCtx
		runtime.own(cancelJanitor)
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
		Admission:         runtime.Lifecycle,
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
	if runtime.ProcessLease != nil && runtime.ProcessLease.State() == processlease.StateLost {
		runtime.Close()
		return nil, processlease.ErrLeaseLost
	}
	return runtime, nil
}

func mergeNetworkFlowImportFacadeOverride(overrides map[string]any, facade imports.ExtensionImportFacade) map[string]any {
	merged := map[string]any{}
	for key, value := range overrides {
		merged[key] = value
	}
	facades := map[string]imports.ExtensionImportFacade{}
	if existing, ok := overrides[imports.ExtensionImportFacadesOverrideKey]; ok && existing != nil {
		typed, ok := existing.(map[string]imports.ExtensionImportFacade)
		if !ok {
			return merged
		}
		for key, value := range typed {
			facades[key] = value
		}
	}
	if facade != nil {
		facades[imports.ExtensionImportFacadeKey(imports.ImportTargetKindNetworkFlowTable, imports.NetworkFlowExtensionProfileID)] = facade
	}
	merged[imports.ExtensionImportFacadesOverrideKey] = facades
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

type stagedObjectDeleter struct {
	store objectstore.Store
}

func (deleter stagedObjectDeleter) DeleteStagedObject(ctx context.Context, storageIdentity string) error {
	if deleter.store == nil {
		return fmt.Errorf("object store unavailable")
	}
	return deleter.store.DeleteObject(ctx, storageIdentity)
}

func configuredExtensionClaimIDs(descriptors []extensions.Descriptor, cfg config.Config, override []httpapi.ExtensionProfile) []string {
	if override != nil {
		claimed := make([]string, 0, len(override))
		for _, profile := range override {
			if profile.Claimed {
				claimed = append(claimed, profile.ProfileID)
			}
		}
		return claimed
	}
	claimed := make([]string, 0, len(descriptors))
	for _, descriptor := range descriptors {
		isClaimed := false
		switch descriptor.ProfileID {
		case "enterprise_authentication":
			isClaimed = cfg.EnterpriseAuthentication.Claimed
		case "import":
			isClaimed = cfg.Import.Claimed
		case "incident_portability":
			isClaimed = cfg.IncidentPortability.Claimed
		case "network_flow_activity":
			isClaimed = cfg.NetworkFlowActivity.Claimed
		case "reference_pack":
			isClaimed = cfg.ReferencePack.Claimed
		case "snapshot_reporting":
			isClaimed = cfg.SnapshotReporting.Claimed
		}
		if isClaimed {
			claimed = append(claimed, descriptor.ProfileID)
		}
	}
	return claimed
}

func extensionHTTPProfiles(descriptors []extensions.Descriptor, claimedProfileIDs []string) []httpapi.ExtensionProfile {
	claimed := make(map[string]struct{}, len(claimedProfileIDs))
	for _, profileID := range claimedProfileIDs {
		claimed[profileID] = struct{}{}
	}
	profiles := make([]httpapi.ExtensionProfile, 0, len(descriptors))
	for _, descriptor := range descriptors {
		_, isClaimed := claimed[descriptor.ProfileID]
		contractMajor := descriptor.ContractMajor
		workspaces := make([]httpapi.ExtensionWorkspace, 0, len(descriptor.WorkspaceKeys))
		for _, workspaceKey := range descriptor.WorkspaceKeys {
			workspaces = append(workspaces, httpapi.ExtensionWorkspace{WorkspaceKey: workspaceKey, MinimumRole: "viewer"})
		}
		profiles = append(profiles, httpapi.ExtensionProfile{
			ProfileID: descriptor.ProfileID, Claimable: descriptor.Claimable, Claimed: isClaimed,
			ContractMajor: &contractMajor, RouteFamilies: descriptor.RouteFamilies,
			WorkspaceKeys: descriptor.WorkspaceKeys, Capabilities: descriptor.CapabilityIDs, Workspaces: workspaces,
		})
	}
	return profiles
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

func (r *Runtime) ActivatePublication() error {
	if r == nil || r.Lifecycle == nil {
		return fmt.Errorf("extension_publication_failed")
	}
	if r.ProcessLease != nil && r.ProcessLease.State() != processlease.StateHeld {
		return fmt.Errorf("extension_publication_failed")
	}
	if err := r.Lifecycle.Publish(); err != nil {
		return err
	}
	r.publicationOnce.Do(func() {
		if r.StagedJanitor != nil && r.stagedJanitorContext != nil {
			go func() {
				defer func() {
					if recovered := recover(); recovered != nil && r.Lifecycle != nil {
						r.Lifecycle.Fatal("published_component_lost")
					}
				}()
				if err := r.StagedJanitor.Run(r.stagedJanitorContext, time.Duration(r.Config.Intervals.Extensions.StagedObjectSweepSeconds)*time.Second); err != nil && r.Lifecycle != nil {
					r.Lifecycle.Fatal("published_component_lost")
				}
			}()
		}
	})
	return nil
}

func (r *Runtime) FatalEvents() <-chan processlifecycle.FatalSignal {
	if r == nil || r.Lifecycle == nil {
		return nil
	}
	return r.Lifecycle.FatalEvents()
}

func (r *Runtime) watchProcessLease(ctx context.Context) {
	if r == nil || r.ProcessLease == nil || r.Lifecycle == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-r.ProcessLease.Events():
			switch event.State {
			case processlease.StateUncertain:
				r.Lifecycle.CloseAdmission()
			case processlease.StateHeld:
				if event.Previous == processlease.StateUncertain {
					r.Lifecycle.RestoreAdmission()
				}
			case processlease.StateLost:
				r.Lifecycle.Fatal("application_process_lease_lost")
				return
			}
		}
	}
}

func (r *Runtime) Close() {
	if r == nil {
		return
	}
	r.closeOnce.Do(func() {
		if r.Lifecycle != nil {
			r.Lifecycle.MarkTerminating()
		}
		for index := len(r.cleanups) - 1; index >= 0; index-- {
			r.cleanups[index]()
		}
		if r.Lifecycle != nil {
			r.Lifecycle.MarkExited()
		}
	})
}

func (r *Runtime) own(cleanup func()) {
	if r != nil && cleanup != nil {
		r.cleanups = append(r.cleanups, cleanup)
	}
}
