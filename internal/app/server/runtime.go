package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/jobapi"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
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
	"github.com/JochiRaider/cartulary/internal/platform/httpapi/extensiondiscovery"
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
	Config                 config.Config
	Handler                http.Handler
	Extensions             *extensions.Coordinator
	ExtensionState         *extensions.StateRuntime
	StagedObjects          *stagedobjects.Service
	StagedJanitor          *stagedobjects.Janitor
	StagedHealth           *stagedobjects.Health
	CrossOwnerTransactions *crossownertransaction.Coordinator
	Postgres               *pgxpool.Pool
	ObjectStore            objectstore.Store
	Jobs                   *jobs.Manager
	JobRunner              *jobs.Runner
	WSHub                  *platformws.Hub
	Telemetry              *telemetry.Runtime
	ProcessLease           *processlease.Lease
	Lifecycle              *processlifecycle.Controller
	Publication            *PublicationController

	closeOnce            sync.Once
	publicationOnce      sync.Once
	cleanups             []func()
	stagedJanitorContext context.Context
}

func NewRuntime(ctx context.Context, cfg config.Config, options Options) (*Runtime, error) {
	extensionCoordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		return nil, fmt.Errorf("admit packaged extension registry: %w", err)
	}
	inactiveCatalog, err := extensionassembly.InactiveConfigurationCatalog(extensionCoordinator)
	if err != nil {
		return nil, fmt.Errorf("project extension configuration policy: %w", err)
	}
	normalizedCfg, err := config.ValidateForStartupWithExtensionInactiveCatalog(cfg, inactiveCatalog)
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
	var claimOverride []httpapi.ExtensionProfile
	if options.HTTP.Dependencies.ExtensionEpoch != nil {
		claimOverride = httpapi.ExtensionProfilesFromEpoch(options.HTTP.Dependencies.ExtensionEpoch)
	}
	requestedClaims := extensionClaimRequest(extensionCoordinator.Descriptors(), normalizedCfg, claimOverride)
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
	publication := NewPublicationController(runtime.Lifecycle)
	if err := publication.Prepare(extensionPlan); err != nil {
		runtime.Close()
		return nil, err
	}
	runtime.Publication = publication
	resolvedClaims := publication.ResolvedClaims()
	profiles := publicationHTTPProfiles(publication.Discovery())
	runtime.Extensions = extensionCoordinator

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
		logicalStateStore, stateStoreAdapterErr := extensionassembly.NewStateStore(stateStore)
		if stateStoreAdapterErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose extension state store port: %w", stateStoreAdapterErr)
		}
		stateRuntime, stateRuntimeErr := extensions.NewStateRuntime(extensions.StateRuntimeOptions{
			Store: logicalStateStore,
			FinalValidators: map[string]extensions.FinalStateValidator{
				"network_flow_activity.validate_state_v1": func(ctx context.Context, _ extensions.FinalStateValidationContext, reader extensions.StateReadCapability) (extensions.StateValidationResult, error) {
					if err := networkflow.ValidateExtensionState(ctx, reader); err != nil {
						return extensions.StateValidationResult{
							SchemaID: "cartulary.extension_final_state_validation_result.v1",
							Status:   "invalid",
							Findings: []extensions.StateFinding{{
								Code: "network_flow_activity_state_invalid",
								Path: "/",
							}},
						}, nil
					}
					return extensions.ValidFinalStateValidationResult(), nil
				},
			},
			Now:               now,
			LockTimeout:       time.Duration(normalizedCfg.Timeouts.Extensions.MigrationLockSeconds) * time.Second,
			StepTimeout:       time.Duration(normalizedCfg.Timeouts.Extensions.MigrationStepSeconds) * time.Second,
			ProfileTimeout:    time.Duration(normalizedCfg.Timeouts.Extensions.ProfileMigrationSeconds) * time.Second,
			ValidationTimeout: time.Duration(normalizedCfg.Timeouts.Extensions.ValidationSeconds) * time.Second,
			FatalIntegritySink: func(cause error) {
				reason := "indeterminate_database_commit"
				if errors.Is(cause, extensions.ErrStateReadbackMismatch) {
					reason = "migration_ledger_state_mismatch"
				}
				runtime.Lifecycle.Fatal(reason)
			},
		})
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
		stagedRepository, stagedRepositoryErr := extensionassembly.NewStagedObjectRepository(extensionStateStore)
		if stagedRepositoryErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose staged-object repository: %w", stagedRepositoryErr)
		}
		stagedBytes, stagedBytesErr := extensionassembly.NewStagedObjectBytes(runtime.ObjectStore)
		if stagedBytesErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose staged-object byte store: %w", stagedBytesErr)
		}
		stagedService, stagedServiceErr := stagedobjects.NewService(stagedobjects.ServiceOptions{
			Repository: stagedRepository,
			Bytes:      stagedBytes,
			Now:        now,
			FatalSink: func(error) {
				runtime.Lifecycle.Fatal("staged_object_publication_mismatch")
			},
		})
		if stagedServiceErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose staged-object service: %w", stagedServiceErr)
		}
		stagedHealth := stagedobjects.NewHealth()
		janitor, janitorErr := stagedobjects.NewJanitor(stagedobjects.JanitorOptions{
			Repository:       stagedRepository,
			Bytes:            stagedBytes,
			Health:           stagedHealth,
			Now:              now,
			BatchLimit:       int(normalizedCfg.Limits.Extensions.StagedObjectCleanupBatch),
			OperationTimeout: time.Duration(normalizedCfg.Timeouts.Extensions.StagedObjectCleanupSeconds) * time.Second,
			FatalSink: func(error) {
				runtime.Lifecycle.Fatal("staged_object_publication_mismatch")
			},
		})
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
		runtime.StagedObjects = stagedService
		runtime.StagedJanitor = janitor
		runtime.StagedHealth = stagedHealth
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
	listenerPlanSHA256 := extensionPlan.Summary().ListenerPlanSHA256
	if err := publication.Acknowledge("websocket", listenerPlanSHA256, nil); err != nil {
		runtime.Close()
		return nil, err
	}
	if err := publication.Acknowledge("job_dequeue", listenerPlanSHA256, nil); err != nil {
		runtime.Close()
		return nil, err
	}
	for _, worker := range extensionPlan.Workers() {
		if err := publication.Acknowledge(
			"worker:"+worker.ProfileID+":"+worker.WorkerKind,
			extensionPlan.Summary().WorkerPlanSHA256,
			nil,
		); err != nil {
			runtime.Close()
			return nil, err
		}
	}

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
	if err := attributionResolvers.ValidateAttributionResolvers(revisionPublicationClaims(publication.Claims())); err != nil {
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
	revisionCommands, err := revisionassembly.NewCommandService(postgresHandle, attributionResolvers.ImportedAttributionResolver(incidentbundles.IncidentPortabilityProfileID))
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose revisions command service: %w", err)
	}
	revisionRoutes := revisions.RegisterRoutes(revisionCommands)
	importStore := imports.NewStore(runtime.Postgres)
	networkFlowModule, err := networkflow.NewModule(networkflow.ModuleDependencies{
		Postgres:      postgresHandle,
		ImportSources: importStore,
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
	incidentBundleImportTransactions, err := incidentbundles.NewImportTransactionProvider(
		runtime.Postgres,
		runtime.ObjectStore,
		incidentBundleImportFinalizer,
		projectionadapters.NewIncidentImportRebuilder(runtime.Postgres),
		now,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Incident Bundles import transaction provider: %w", err)
	}
	crossOwnerBackend, err := extensionassembly.NewCrossOwnerBackend(postgresHandle, extensionassembly.TransactionCapabilityMux{
		NetworkFlow: networkFlowModule, IncidentBundles: incidentBundleImportTransactions,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose cross-owner transaction backend: %w", err)
	}
	crossOwnerCoordinator, err := crossownertransaction.New(crossownertransaction.Options{
		Backend: crossOwnerBackend,
		Catalog: extensionassembly.CrossOwnerDescriptors(extensionCoordinator.ParticipantContracts()),
		Timeout: time.Duration(normalizedCfg.Timeouts.Extensions.TransactionParticipantSeconds) * time.Second,
		FatalSink: func(error) {
			runtime.Lifecycle.Fatal("indeterminate_database_commit")
		},
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose cross-owner transaction coordinator: %w", err)
	}
	if err := networkFlowModule.InstallCrossOwnerCoordinator(crossOwnerCoordinator); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("install Network Flow cross-owner transactions: %w", err)
	}
	runtime.CrossOwnerTransactions = crossOwnerCoordinator
	networkFlowPortabilityState, err := networkflow.NewPortabilityStateBinding(postgresHandle)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Network Flow portability state binding: %w", err)
	}
	portabilityPresence, err := extensionassembly.NewIncidentPortabilityStatePresence(networkFlowPortabilityState)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Incident Portability state presence: %w", err)
	}
	portability, err := incidentbundles.NewPortabilityOrchestrator(
		extensionassembly.IncidentPortabilityPolicies(extensionCoordinator.PortabilityPolicies(), publication.ResolvedClaims()),
		portabilityPresence,
		nil,
		runtime.StagedObjects,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Incident Portability: %w", err)
	}
	incidentBundleRoutes := incidentbundles.RegisterRoutes(
		incidentbundles.WithImportFinalizer(incidentBundleImportFinalizer),
		incidentbundles.WithPortability(portability, crossOwnerCoordinator),
	)
	moduleOverrides := mergeNetworkFlowImportFacadeOverride(testRuntimeDeps.ModuleOverrides, networkFlowModule.ImportOwner())
	delete(moduleOverrides, networkflow.KeyRingsOverrideKey)
	builtInRoutes, err := builtInRouteRegistrars([]routeContribution{
		{id: "auth", registrar: auth.RegisterRoutes()},
		{id: "incidents", registrar: incidentRoutes},
		{id: "extensions", registrar: extensiondiscovery.RegisterRoutes()},
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
	readinessProbes := []httpapi.DependencyReadinessProbe{}
	if runtime.StagedHealth != nil {
		readinessProbes = append(readinessProbes, stagedCleanupReadinessProbe{health: runtime.StagedHealth})
	}
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
		Readiness:         httpapi.NewDependencyReadinessChecker(runtime.Postgres, runtime.ObjectStore, readinessProbes...),
		Admission:         runtime.Lifecycle,
		PublicErrorFaults: testRuntimeDeps.PublicErrorFaults,
		ModuleOverrides:   moduleOverrides,
		ExtensionEpoch:    publicationExtensionEpoch{publication: publication},
		Now:               now,
	}

	handler, err := newHTTPHandler(httpOptions)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("setup http handler: %w", err)
	}

	runtime.Handler = handler
	if err := publication.Acknowledge("http", listenerPlanSHA256, nil); err != nil {
		runtime.Close()
		return nil, err
	}
	if err := publication.Commit(); err != nil {
		runtime.Close()
		return nil, err
	}
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

type stagedCleanupReadinessProbe struct {
	health *stagedobjects.Health
}

func (stagedCleanupReadinessProbe) ReadinessName() string {
	return "staged_object_cleanup"
}

func (probe stagedCleanupReadinessProbe) CheckReadinessDependency(context.Context) error {
	if probe.health == nil {
		return nil
	}
	state := probe.health.State()
	if state.Available {
		return nil
	}
	return errors.New(state.ReasonCode)
}

func extensionClaimRequest(descriptors []extensions.Descriptor, cfg config.Config, override []httpapi.ExtensionProfile) []string {
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

func publicationHTTPProfiles(discovery []extensions.DiscoveryProfile) []httpapi.ExtensionProfile {
	profiles := make([]httpapi.ExtensionProfile, 0, len(discovery))
	for _, profile := range discovery {
		workspaces := make([]httpapi.ExtensionWorkspace, len(profile.Workspaces))
		for index, workspace := range profile.Workspaces {
			workspaces[index] = httpapi.ExtensionWorkspace{WorkspaceKey: workspace.WorkspaceKey, MinimumRole: workspace.MinimumRole}
		}
		profiles = append(profiles, httpapi.ExtensionProfile{
			ProfileID: profile.ProfileID, Claimable: profile.Claimable, Claimed: profile.Claimed,
			ContractMajor: profile.ContractMajor, RouteFamilies: profile.RouteFamilies,
			WorkspaceKeys: profile.WorkspaceKeys, Capabilities: profile.Capabilities, Workspaces: workspaces,
		})
	}
	return profiles
}

type publicationExtensionEpoch struct {
	publication *PublicationController
}

func (provider publicationExtensionEpoch) ExtensionProfiles() []httpapi.ExtensionProfile {
	if provider.publication == nil {
		return nil
	}
	return publicationHTTPProfiles(provider.publication.Discovery())
}

func revisionPublicationClaims(publication []extensions.ClaimPublication) []revisions.ExtensionClaim {
	claims := make([]revisions.ExtensionClaim, 0, len(publication))
	for _, profile := range publication {
		claims = append(claims, revisions.ExtensionClaim{
			ProfileID: profile.ProfileID,
			Claimed:   profile.Claimed,
		})
	}
	return claims
}

func (r *Runtime) ActivatePublication() error {
	if r == nil || r.Publication == nil {
		return fmt.Errorf("extension_publication_failed")
	}
	if r.ProcessLease != nil && r.ProcessLease.State() != processlease.StateHeld {
		return fmt.Errorf("extension_publication_failed")
	}
	if err := r.Publication.Serve(); err != nil {
		return err
	}
	r.publicationOnce.Do(func() {
		if r.StagedJanitor != nil && r.stagedJanitorContext != nil {
			go func() {
				defer func() {
					if recovered := recover(); recovered != nil && r.Lifecycle != nil {
						r.Publication.ComponentLost("staged_object_janitor")
					}
				}()
				if err := r.StagedJanitor.Run(r.stagedJanitorContext, time.Duration(r.Config.Intervals.Extensions.StagedObjectSweepSeconds)*time.Second); err != nil && r.Lifecycle != nil {
					r.Publication.ComponentLost("staged_object_janitor")
				}
			}()
		}
	})
	return nil
}

func (r *Runtime) PublishedComponentLost(componentID string) bool {
	if r == nil || r.Publication == nil {
		return false
	}
	return r.Publication.ComponentLost(componentID)
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
		if r.Publication != nil {
			r.Publication.AbortStartup()
		}
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
