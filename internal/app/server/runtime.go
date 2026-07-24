package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
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
	ExtensionJobFinalizer  *extensionstore.OwnerFinalizer
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
	descriptors := extensionCoordinator.Descriptors()
	claimPaths, err := extensionassembly.ClaimConfigurationPaths(descriptors)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project extension claim configuration: %w", err)
	}
	claimValues, err := config.BooleanValuesAtPaths(normalizedCfg, claimPaths)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project extension claim configuration: %w", err)
	}
	if claimOverride != nil {
		claimValues, err = extensionClaimOverride(descriptors, claimOverride)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("project extension claim override: %w", err)
		}
	}
	requestedClaims, err := extensionassembly.ResolveClaimRequest(descriptors, claimValues)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("materialize extension claim request: %w", err)
	}
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
	publicationCatalog, err := extensionassembly.NewPublicationCatalog(extensionPlan, extensionCoordinator.ParticipantContracts())
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("prepare extension publication catalog: %w", err)
	}
	enterpriseAuthenticationAdmitted, err := publicationCatalog.ExactProfileContributionSet(
		auth.ProfileID,
		"http_route_family",
		auth.EnterpriseRouteContributionIDs(),
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project Enterprise Authentication application plan: %w", err)
	}
	networkFlowRouteAdmitted, err := publicationCatalog.ExactProfileContributionSet(
		networkflow.ProfileID,
		"http_route_family",
		[]string{networkflow.RouteContributionID},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project Network Flow route application plan: %w", err)
	}
	networkFlowWorkspaceAdmitted, err := publicationCatalog.ExactProfileContributionSet(
		networkflow.ProfileID,
		"incident_workspace",
		[]string{networkflow.WorkspaceContributionID},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project Network Flow workspace application plan: %w", err)
	}
	if networkFlowRouteAdmitted != networkFlowWorkspaceAdmitted {
		runtime.Close()
		return nil, fmt.Errorf("project Network Flow application plan: route and workspace admission disagree")
	}
	referencePackRouteAdmitted, err := publicationCatalog.ExactProfileContributionSet(
		reference_data.ProfileID,
		"http_route_family",
		[]string{reference_data.PacksRouteContributionID},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project Reference Pack application plan: %w", err)
	}
	publication := NewPublicationController(runtime.Lifecycle)
	if err := publication.Prepare(extensionPlan); err != nil {
		runtime.Close()
		return nil, err
	}
	runtime.Publication = publication
	resolvedClaims := extensionPlan.ResolvedClaims()
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
	if enterpriseAuthenticationAdmitted {
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
	if extensionStateStore != nil {
		inactiveJobStore, err := extensionassembly.NewInactiveJobStore(extensionStateStore, now)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose inactive extension job reconciliation: %w", err)
		}
		reconciliationCtx, cancelReconciliation := context.WithTimeout(
			ctx,
			time.Duration(normalizedCfg.Timeouts.Extensions.ReconciliationSeconds)*time.Second,
		)
		reconciliationErr := extensions.ReconcileInactiveExtensionJobs(
			reconciliationCtx,
			inactiveJobStore,
			inactiveExtensionProfileIDs(extensionPlan.Claims()),
			extensionCoordinator.JobKindContracts(),
			int(normalizedCfg.Limits.Extensions.MaxNonterminalJobsPerProfile),
			func(error) {
				runtime.Lifecycle.Fatal("indeterminate_database_commit")
			},
		)
		cancelReconciliation()
		if reconciliationErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("reconcile inactive extension jobs: %w", reconciliationErr)
		}
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
	if enterpriseAuthenticationAdmitted {
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
	if referencePackRouteAdmitted {
		if err := reference_data.EnsureMinimumDisconnectedBundle(ctx, normalizedCfg, runtime.Postgres, now()); err != nil {
			runtime.Close()
			return nil, fmt.Errorf("seed minimum disconnected reference packs: %w", err)
		}
	}
	runtime.Jobs = newJobsManager()
	runtime.Jobs.ConfigureTelemetry(normalizedCfg.Telemetry.Resource.ServiceVersion)
	runtime.JobRunner = jobs.NewRunner()
	runtime.JobRunner.ConfigureDequeueGate(runtime.Lifecycle)
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
	extensionJobContracts, err := extensionassembly.JobContracts(publicationCatalog)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose extension job contracts: %w", err)
	}
	if err := runtime.Jobs.ConfigureExtensionContracts(extensionJobContracts); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("configure extension job contracts: %w", err)
	}
	if extensionStateStore != nil {
		extensionJobFinalizer, err := extensionstore.NewOwnerFinalizer(
			extensionStateStore,
			runtime.Jobs,
			now,
			func(error) {
				runtime.Lifecycle.Fatal("indeterminate_database_commit")
			},
		)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose extension job finalizer: %w", err)
		}
		runtime.ExtensionJobFinalizer = extensionJobFinalizer
	}
	runtime.JobRunner.Configure(runtime.Jobs)
	hub.ConfigureTelemetry(normalizedCfg.Telemetry.Resource.ServiceVersion)
	listenerPlanSHA256 := extensionPlan.Summary().ListenerPlanSHA256

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
	if err := attributionResolvers.ValidateAttributionResolvers(revisionPublicationClaims(extensionPlan.Claims())); err != nil {
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
		extensionassembly.IncidentPortabilityPolicies(extensionCoordinator.PortabilityPolicies(), resolvedClaims),
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
		incidentbundles.WithJobSuccessFinalizer(
			extensionassembly.NewIncidentBundleJobSuccessFinalizer(runtime.ExtensionJobFinalizer, now),
		),
		incidentbundles.WithPortability(portability, crossOwnerCoordinator),
	)
	importRoutes := imports.RegisterRoutes(
		imports.WithExtensionProfileAdmission(func(profileID string) bool {
			return profileID == networkflow.ProfileID && networkFlowRouteAdmitted
		}),
		imports.WithJobSuccessFinalizer(extensionassembly.NewImportJobSuccessFinalizer(runtime.ExtensionJobFinalizer)),
	)
	referencePackRoutes := reference_data.RegisterRoutes(
		reference_data.WithJobSuccessFinalizer(
			extensionassembly.NewReferencePackJobSuccessFinalizer(runtime.ExtensionJobFinalizer),
		),
	)
	moduleOverrides := mergeNetworkFlowImportFacadeOverride(testRuntimeDeps.ModuleOverrides, networkFlowModule.ImportOwner())
	delete(moduleOverrides, networkflow.KeyRingsOverrideKey)
	authRouteOptions := []auth.RouteOption{}
	if enterpriseAuthenticationAdmitted {
		authRouteOptions = append(authRouteOptions, auth.WithEnterpriseAuthBindings())
	}
	builtInRoutes, err := applicationRouteRegistrars([]routeContribution{
		{id: "auth", registrar: auth.RegisterRoutes(authRouteOptions...)},
		{id: "incidents", registrar: incidentRoutes},
		{id: "extensions", registrar: extensiondiscovery.RegisterRoutes()},
		{id: "jobs", registrar: jobapi.RegisterRoutes()},
		{id: "saved_views", registrar: savedviews.RegisterRoutes()},
		{id: "view_schemas", registrar: viewschemas.RegisterRoutes()},
		{id: "collaboration", registrar: collaboration.RegisterRoutes()},
		{id: "entities", registrar: entities.RegisterRoutes()},
		{id: "evidence", registrar: evidence.RegisterRoutes()},
		{id: "assessments", registrar: assessments.RegisterRoutes()},
		{id: "workbook", registrar: workbook.RegisterRoutes()},
		{id: "timeline", registrar: timeline.RegisterRoutes()},
		{id: "revisions", registrar: revisionRoutes},
	}, []extensionRouteBinding{
		{
			id: "enterprise_authentication_routes",
			contributionIDs: []string{
				auth.EnterpriseOIDCRouteContributionID,
				auth.EnterpriseProvidersRouteContributionID,
				auth.EnterpriseSAMLRouteContributionID,
			},
			registrar: auth.RegisterEnterpriseRoutes(),
		},
		{
			id:              "enterprise_authentication_user_auth_bindings",
			contributionIDs: []string{auth.EnterpriseUserAuthBindingsRouteContributionID},
			baseRegistrarID: "auth",
		},
		{id: "import", contributionIDs: []string{"import.sessions_route"}, registrar: importRoutes},
		{id: "incident_portability", contributionIDs: []string{incidentbundles.BundlesRouteContributionID}, registrar: incidentBundleRoutes},
		{id: "network_flow_activity", contributionIDs: []string{networkflow.RouteContributionID}, registrar: networkFlowModule.RegisterRoutes()},
		{id: "reference_pack", contributionIDs: []string{reference_data.PacksRouteContributionID}, registrar: referencePackRoutes},
		{
			id:              "snapshot_reporting_resources",
			contributionIDs: []string{"snapshot_reporting.releases_route", "snapshot_reporting.snapshots_route"},
			registrar:       reporting.RegisterRoutes(),
		},
		{
			id:              "snapshot_reporting_compositions",
			contributionIDs: []string{"snapshot_reporting.report_compositions_route"},
			registrar:       reportcomposition.RegisterRoutes(),
		},
	}, publicationCatalog)
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
		Config:              normalizedCfg,
		Env:                 options.Env,
		Postgres:            runtime.Postgres,
		PostgresDB:          postgresHandle,
		ObjectStore:         runtime.ObjectStore,
		Jobs:                runtime.Jobs,
		JobRunner:           runtime.JobRunner,
		WSHub:               hub,
		CursorCodec:         cursorCodec,
		Readiness:           httpapi.NewDependencyReadinessChecker(runtime.Postgres, runtime.ObjectStore, readinessProbes...),
		Admission:           runtime.Lifecycle,
		PublicErrorFaults:   testRuntimeDeps.PublicErrorFaults,
		ModuleOverrides:     moduleOverrides,
		ExtensionEpoch:      publicationExtensionEpoch{publication: publication},
		ExtensionDiscovery:  publicationExtensionEpoch{publication: publication},
		ExtensionClaims:     publicationExtensionEpoch{publication: publication},
		ExtensionRoutes:     publicationExtensionEpoch{publication: publication},
		ExtensionWorkspaces: publicationExtensionEpoch{publication: publication},
		Now:                 now,
	}

	if err := publication.Commit(); err != nil {
		runtime.Close()
		return nil, err
	}
	handler, err := newHTTPHandler(httpOptions)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("setup http handler: %w", err)
	}

	runtime.Handler = handler
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
	if err := publication.Acknowledge("http", listenerPlanSHA256, nil); err != nil {
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

func extensionClaimOverride(descriptors []extensions.Descriptor, override []httpapi.ExtensionProfile) (map[string]bool, error) {
	descriptorByProfile := make(map[string]extensions.Descriptor, len(descriptors))
	values := make(map[string]bool, len(descriptors))
	for _, descriptor := range descriptors {
		if _, duplicate := descriptorByProfile[descriptor.ProfileID]; duplicate {
			return nil, fmt.Errorf("duplicate extension profile %q", descriptor.ProfileID)
		}
		descriptorByProfile[descriptor.ProfileID] = descriptor
		values[descriptor.ClaimConfigKey] = false
	}
	seen := make(map[string]struct{}, len(override))
	for _, profile := range override {
		descriptor, present := descriptorByProfile[profile.ProfileID]
		if !present {
			return nil, fmt.Errorf("extension profile %q is not generated", profile.ProfileID)
		}
		if _, duplicate := seen[profile.ProfileID]; duplicate {
			return nil, fmt.Errorf("duplicate extension profile %q", profile.ProfileID)
		}
		seen[profile.ProfileID] = struct{}{}
		values[descriptor.ClaimConfigKey] = profile.Claimed
	}
	return values, nil
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

func (provider publicationExtensionEpoch) ExtensionDiscoveryProfiles() []httpapi.ExtensionProfile {
	return provider.ExtensionProfiles()
}

func (provider publicationExtensionEpoch) ExtensionClaims() []httpapi.ExtensionClaim {
	if provider.publication == nil {
		return nil
	}
	publication := provider.publication.Claims()
	claims := make([]httpapi.ExtensionClaim, 0, len(publication))
	for _, claim := range publication {
		claims = append(claims, httpapi.ExtensionClaim{ProfileID: claim.ProfileID, Claimed: claim.Claimed})
	}
	return claims
}

func (provider publicationExtensionEpoch) ExtensionRoutes() []httpapi.ExtensionRoute {
	if provider.publication == nil {
		return nil
	}
	publication := provider.publication.Routes()
	routes := make([]httpapi.ExtensionRoute, 0, len(publication))
	for _, route := range publication {
		routes = append(routes, httpapi.ExtensionRoute{
			ProfileID: route.ProfileID, RouteFamily: route.RouteFamily, Claimed: route.DispatchState == "claimed",
		})
	}
	return routes
}

func (provider publicationExtensionEpoch) ExtensionWorkspaces() []httpapi.ExtensionWorkspacePublication {
	if provider.publication == nil {
		return nil
	}
	publication := provider.publication.Workspaces()
	workspaces := make([]httpapi.ExtensionWorkspacePublication, 0, len(publication))
	for _, workspace := range publication {
		workspaces = append(workspaces, httpapi.ExtensionWorkspacePublication{
			ProfileID: workspace.ProfileID, WorkspaceKey: workspace.WorkspaceKey, MinimumRole: "viewer",
		})
	}
	return workspaces
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

func inactiveExtensionProfileIDs(publication []extensions.ClaimPublication) []string {
	profileIDs := make([]string, 0, len(publication))
	for _, profile := range publication {
		if !profile.Claimed {
			profileIDs = append(profileIDs, profile.ProfileID)
		}
	}
	sort.Strings(profileIDs)
	return profileIDs
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
	if r.JobRunner != nil {
		recoveryCtx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Config.Timeouts.Extensions.ReconciliationSeconds)*time.Second)
		defer cancel()
		if err := r.JobRunner.Activate(recoveryCtx); err != nil {
			r.Publication.ComponentLost("job_dequeue")
			return fmt.Errorf("activate extension job recovery: %w", err)
		}
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
