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
	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/importassembly"
	"github.com/JochiRaider/cartulary/internal/app/incidentportabilityassembly"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
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
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	timelineadmission "github.com/JochiRaider/cartulary/internal/modules/timeline/admission"
	"github.com/JochiRaider/cartulary/internal/modules/viewschemas"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
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
	"github.com/JochiRaider/cartulary/internal/platform/securefile"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

var (
	newJobsManager      = jobs.NewManager
	setupPostgres       = postgres.Setup
	ensureSchemaReady   = postgres.EnsureSchemaReady
	setupObjectStore    = objectstore.Setup
	runBootstrap        = bootstrap.Preflight
	newCollaborationHub = collaboration.NewHub
	newHTTPHandler      = httpapi.NewHandler
	readSecureFile      = securefile.Read
)

const stagedObjectJanitorAdvisoryKey int64 = 4850189438622597894

type Options struct {
	Env         map[string]string
	HTTP        httpapi.Options
	Postgres    *pgxpool.Pool
	ObjectStore objectstore.Store
	Now         func() time.Time
}

type Runtime struct {
	Settings                RuntimeSettings
	Handler                 http.Handler
	Extensions              *extensions.Coordinator
	ExtensionState          *extensions.StateRuntime
	ExtensionJobFinalizer   *extensionstore.OwnerFinalizer
	StagedObjects           *stagedobjects.Service
	StagedJanitor           *stagedobjects.Janitor
	StagedJanitorLeader     *componentLeader
	StagedHealth            *stagedobjects.Health
	CrossOwnerTransactions  *crossownertransaction.Coordinator
	Postgres                *pgxpool.Pool
	ObjectStore             objectstore.Store
	Jobs                    *jobs.Manager
	JobTransactions         *jobs.TransactionService
	JobRunner               *jobs.Runner
	CollaborationHub        *collaboration.Hub
	CollaborationDispatcher *collaboration.Dispatcher
	CollaborationIntents    collaboration.IntentAppender
	Timeline                *timelineassembly.Bundle
	Revisions               *revisionassembly.Runtime
	Telemetry               *telemetry.Runtime
	ProcessLease            *processlease.Lease
	Lifecycle               *processlifecycle.Controller
	Publication             *PublicationController
	PublicHTTP              httpapi.RouteDiagnostics

	closeOnce            sync.Once
	publicationOnce      sync.Once
	cleanups             []func()
	stagedJanitorContext context.Context
}

type RuntimeSettings struct {
	TelemetryFlushTimeoutMS  int64
	ReconciliationSeconds    int64
	StagedObjectSweepSeconds int64
	ShutdownDrainSeconds     int64
	ProcessModel             string
}

func NewRuntime(ctx context.Context, deployment configassembly.Deployment, options Options) (*Runtime, error) {
	loaded, err := configassembly.Admit(deployment)
	if err != nil {
		return nil, err
	}
	return newRuntime(ctx, loaded, options)
}

func newRuntime(ctx context.Context, loadedConfiguration configassembly.Loaded, options Options) (*Runtime, error) {
	extensionCoordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		return nil, fmt.Errorf("admit packaged extension registry: %w", err)
	}
	enterpriseAuthenticationConfiguration, err := loadedConfiguration.EnterpriseAuthentication()
	if err != nil {
		return nil, fmt.Errorf("project Enterprise Authentication configuration: %w", err)
	}
	networkFlowConfiguration, err := loadedConfiguration.NetworkFlow()
	if err != nil {
		return nil, fmt.Errorf("project Network Flow configuration: %w", err)
	}
	if err := loadedConfiguration.ValidateForStartup(); err != nil {
		return nil, err
	}
	normalizedCfg := loadedConfiguration.Deployment()

	runtime := &Runtime{
		Settings:  runtimeSettings(normalizedCfg),
		Lifecycle: processlifecycle.New(),
	}
	var incidentBundleStorage incidentbundles.BundleStorage
	var referencePackStorage reference_data.Storage
	if normalizedCfg.Application.ProcessModel == config.ProcessModelSingle {
		rootIncidentBundleStorage, storageErr := newIncidentBundleRootStorage(
			normalizedCfg.Roots.TemporaryWork.Path,
			normalizedCfg.Roots.ExportOutputs.Path,
		)
		if storageErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose incident bundle storage: %w", storageErr)
		}
		incidentBundleStorage = rootIncidentBundleStorage
		runtime.own(rootIncidentBundleStorage.Close)
		rootReferencePackStorage, storageErr := newReferencePackRootStorage(
			normalizedCfg.Roots.TemporaryWork.Path,
			normalizedCfg.Roots.ReferencePackStorage.Path,
		)
		if storageErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose Reference Pack storage: %w", storageErr)
		}
		referencePackStorage = rootReferencePackStorage
		runtime.own(rootReferencePackStorage.Close)
	}
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	if options.Postgres != nil {
		runtime.Postgres = options.Postgres
	} else {
		postgresSettings, settingsErr := postgres.ResolveSettings(configassembly.PostgresBinding(normalizedCfg), options.Env)
		if settingsErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup postgres: %w", settingsErr)
		}
		pool, err := setupPostgres(ctx, postgresSettings)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup postgres: %w", err)
		}
		runtime.Postgres = pool
		if pool != nil {
			runtime.own(pool.Close)
		}
	}
	if runtime.Postgres != nil && normalizedCfg.Application.ProcessModel == config.ProcessModelSingle {
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
	descriptors := extensionCoordinator.Descriptors()
	claimPaths, err := extensionassembly.ClaimConfigurationPaths(descriptors)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project extension claim configuration: %w", err)
	}
	claimValues, err := loadedConfiguration.BooleanValuesAtPaths(claimPaths)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project extension claim configuration: %w", err)
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
	snapshotReportingRoutesAdmitted, err := publicationCatalog.ExactProfileContributionSet(
		reporting.ProfileID,
		"http_route_family",
		[]string{
			reporting.ReleasesRouteContributionID,
			reporting.ReportCompositionsRouteContributionID,
			reporting.SnapshotsRouteContributionID,
		},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project Snapshot/Reporting route application plan: %w", err)
	}
	snapshotReportingParticipantAdmitted, err := publicationCatalog.ExactProfileContributionSet(
		reporting.ProfileID,
		"snapshot_reporting_participant",
		[]string{reporting.RenderExportContributionID},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project Snapshot/Reporting participant application plan: %w", err)
	}
	if snapshotReportingRoutesAdmitted != snapshotReportingParticipantAdmitted {
		runtime.Close()
		return nil, fmt.Errorf("project Snapshot/Reporting application plan: routes and participant admission disagree")
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
	if err := telemetry.RegisterSecretPurposes(normalizedCfg.Telemetry, options.Env, secretPurposes); err != nil {
		runtime.Close()
		return nil, err
	}
	var enterpriseProviderDefinitions []authn.EnterpriseAuthProviderDefinition
	if enterpriseAuthenticationAdmitted {
		enterpriseProviderDefinitions, err = loadEnterpriseProviderManifest(enterpriseAuthenticationConfiguration, options.Env)
		if err != nil {
			runtime.Close()
			return nil, err
		}
		if err := enterpriseauth.RegisterProviderSecretPurposes(enterpriseProviderDefinitions, options.Env, secretPurposes); err != nil {
			runtime.Close()
			return nil, deploymentEnterpriseAuthenticationError(err)
		}
	}
	telemetryRuntime, err := telemetry.Bootstrap(ctx, normalizedCfg.Telemetry, normalizedCfg.DeploymentProfile, options.Env, telemetry.WithResolvedClaimIdentity(telemetry.ResolvedClaimIdentity{
		ProfileIDs: resolvedClaims.ProfileIDs(),
		SHA256:     resolvedClaims.SHA256(),
	}))
	if err != nil {
		runtime.Close()
		return nil, err
	}
	runtime.Telemetry = telemetryRuntime
	runtime.own(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(runtime.Settings.TelemetryFlushTimeoutMS)*time.Millisecond)
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
		runtime.ObjectStore = instrumentedObjectStore(
			normalizedCfg.Telemetry.Enabled,
			normalizedCfg.Telemetry.Resource.ServiceVersion,
			options.ObjectStore,
		)
	} else {
		objectStoreSettings, settingsErr := objectstore.ResolveSettings(configassembly.ObjectStoreBinding(normalizedCfg), options.Env)
		if settingsErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup object store: %w", settingsErr)
		}
		client, err := setupObjectStore(ctx, objectStoreSettings, configassembly.ObjectStoreInstrumentation(normalizedCfg))
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup object store: %w", err)
		}
		runtime.ObjectStore = client
		if client != nil {
			runtime.own(func() { _ = client.Close() })
		}
	}
	if normalizedCfg.Application.ProcessModel == config.ProcessModelReplicated {
		if runtime.ObjectStore == nil {
			runtime.Close()
			return nil, errors.New("replicated process model requires shared object storage")
		}
		incidentBundleStorage = newSharedIncidentBundleStorage(runtime.ObjectStore)
		referencePackStorage = newSharedReferencePackStorage(runtime.ObjectStore)
		agreement, agreementErr := admitPublicationPlanAgreement(
			ctx,
			runtime.Postgres,
			runtime.ObjectStore,
			extensionPlan.Summary(),
			normalizedCfg.Telemetry.Resource.ServiceInstanceID,
			now,
		)
		if agreementErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("admit replicated publication plan: %w", agreementErr)
		}
		runtime.own(agreement.Close)
	}
	if incidentBundleStorage == nil || referencePackStorage == nil {
		runtime.Close()
		return nil, errors.New("publication storage is unavailable for the admitted process model")
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
		runtime.StagedObjects = stagedService
		runtime.StagedJanitor = janitor
		runtime.StagedHealth = stagedHealth
		if normalizedCfg.Application.ProcessModel == config.ProcessModelSingle {
			cleanupCtx, cancelCleanup := context.WithTimeout(ctx, time.Duration(normalizedCfg.Timeouts.Extensions.StagedObjectCleanupSeconds)*time.Second)
			cleanupErr := janitor.Sweep(cleanupCtx)
			cancelCleanup()
			if cleanupErr != nil {
				runtime.Close()
				return nil, fmt.Errorf("initial staged-object cleanup: %w", cleanupErr)
			}
			janitorCtx, cancelJanitor := context.WithCancel(context.Background())
			runtime.stagedJanitorContext = janitorCtx
			runtime.own(cancelJanitor)
		} else {
			runtime.StagedJanitorLeader = newComponentLeader(
				processlease.PostgresBackend{
					Pool:        runtime.Postgres,
					AdvisoryKey: stagedObjectJanitorAdvisoryKey,
					Purpose:     "staged-object janitor",
				},
				500*time.Millisecond,
				time.Duration(normalizedCfg.Timeouts.Extensions.ProcessLeaseLossDetectionSeconds)*time.Second,
				func(componentCtx context.Context) error {
					if err := janitor.Sweep(componentCtx); err != nil {
						return err
					}
					return janitor.Run(
						componentCtx,
						time.Duration(normalizedCfg.Intervals.Extensions.StagedObjectSweepSeconds)*time.Second,
					)
				},
				func() {
					if runtime.Publication != nil {
						runtime.Publication.ComponentLost("staged_object_janitor")
					}
				},
			)
			runtime.own(runtime.StagedJanitorLeader.Close)
		}
	}
	postgresHandle := instrumentedPostgres(
		normalizedCfg.Telemetry.Enabled,
		normalizedCfg.Telemetry.Resource.ServiceVersion,
		runtime.Postgres,
	)
	postgresHandle, err = decoratePostgresForTestRuntime(
		options.Env,
		options.HTTP.Dependencies.ModuleOverrides,
		postgresHandle,
	)
	if err != nil {
		runtime.Close()
		return nil, err
	}

	if err := runBootstrap(ctx, configassembly.BootstrapSettings(normalizedCfg), runtime.Postgres); err != nil {
		runtime.Close()
		return nil, err
	}
	if enterpriseAuthenticationAdmitted {
		if err := enterpriseauth.ReconcileProviderDefinitions(ctx, enterpriseProviderDefinitions, authn.NewStore(postgresHandle), now()); err != nil {
			runtime.Close()
			return nil, deploymentEnterpriseAuthenticationError(err)
		}
	}
	networkFlowKeyRings, err := loadNetworkFlowKeyRings(networkFlowConfiguration, options.Env, now(), secretPurposes)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	if referencePackRouteAdmitted {
		referenceLimits := referenceDataLimits(normalizedCfg)
		if err := reference_data.EnsureMinimumDisconnectedBundle(ctx, reference_data.MinimumDisconnectedBundleOptions{
			DeploymentProfile: normalizedCfg.DeploymentProfile,
			ArchiveLimits:     referenceLimits.Archives,
			ReferenceLimits:   referenceLimits.ReferencePacks,
			Storage:           referencePackStorage,
		}, runtime.Postgres, now()); err != nil {
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
	hub := newCollaborationHub()
	runtime.CollaborationHub = hub
	intentAppender := collaboration.NewIntentAppender()
	runtime.CollaborationIntents = intentAppender
	runtime.CollaborationDispatcher = collaboration.NewDispatcher(postgresHandle, hub, now)
	dispatcher := runtime.CollaborationDispatcher
	runtime.own(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = dispatcher.Close(closeCtx)
	})
	intentAdapters := collaborationIntentAdapters{appender: intentAppender}
	runtime.JobTransactions, err = jobs.NewTransactionService(intentAdapters)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Jobs transaction service: %w", err)
	}
	runtime.Jobs.Configure(runtime.Postgres, runtime.JobTransactions, now)
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
			runtime.JobTransactions,
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
		CollaborationSession: collaboration.NewIncidentSessionNotifier(postgresHandle, hub),
	})
	incidentBundleImportFinalizer := incidents.NewIncidentBundleImportFinalizer()
	historicalIntentPolicy := collaboration.NewHistoricalIntentPolicy()
	revisionRuntime, err := revisionassembly.Build(
		revisionassembly.Dependencies{
			HistoricalIntentPolicy: historicalIntentPolicy,
			IntentAppender:         intentAppender,
		},
		revisionassembly.CurrentProviderContributions()...,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Revisions runtime: %w", err)
	}
	runtime.Revisions = revisionRuntime
	timelineBundle := timelineassembly.NewBundle(
		postgresHandle,
		conflicttokens.NewConflictTokenCodec(keys),
		revisionRuntime.Appender(),
		intentAppender,
	)
	runtime.Timeline = timelineBundle
	revisionCommands, err := revisionRuntime.NewCommandService(
		postgresHandle,
		attributionResolvers.ImportedAttributionResolver(incidentbundles.IncidentPortabilityProfileID),
		timelineBundle.ProjectionCoordinator,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose revisions command service: %w", err)
	}
	revisionRoutes := revisions.RegisterRoutes(revisionCommands)
	importStore := imports.NewStore(
		runtime.Postgres,
		revisionRuntime.Appender(),
		runtime.JobTransactions,
	)
	timelineFacade := timelineBundle.Facade
	networkFlowModule, err := networkflow.NewModule(networkflow.ModuleDependencies{
		Postgres:        postgresHandle,
		ImportSources:   importStore,
		KeyRings:        networkFlowKeyRings,
		Now:             now,
		Transactions:    postgres.NewTransactionRunner(postgresHandle),
		IncidentLocks:   incidents.NewTransactionParticipant(),
		AuditAppender:   authn.NewAdministrativeAuditAppender(),
		Indicators:      indicators.NewStore(postgresHandle, revisionRuntime.Appender()),
		ResourceIntents: intentAdapters,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Network Flow module: %w", err)
	}
	incidentBundleImportTransactions, err := incidentbundles.NewImportTransactionProvider(
		runtime.Postgres,
		runtime.ObjectStore,
		incidentBundleImportFinalizer,
		timelineBundle.ProjectionCatalog.Rebuild,
		historicalIntentPolicy,
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
	incidentSourceCatalog, err := incidentportabilityassembly.NewCatalog()
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Incident Portability source catalog: %w", err)
	}
	incidentBundleRoutes := incidentbundles.RegisterRoutes(
		incidentbundles.WithStorage(incidentBundleStorage),
		incidentbundles.WithLimits(incidentBundleLimits(normalizedCfg)),
		incidentbundles.WithImportFinalizer(incidentBundleImportFinalizer),
		incidentbundles.WithJobSuccessFinalizer(
			extensionassembly.NewIncidentBundleJobSuccessFinalizer(runtime.ExtensionJobFinalizer, now),
		),
		incidentbundles.WithPortability(portability, crossOwnerCoordinator),
		incidentbundles.WithProjectionRebuild(timelineBundle.ProjectionCatalog.Rebuild),
		incidentbundles.WithSourceCatalog(incidentSourceCatalog),
		incidentbundles.WithHistoricalIntentPolicy(historicalIntentPolicy),
	)
	importOwnerLimits, importArchiveLimits := importLimits(normalizedCfg)
	importOwnerRegistry, err := importassembly.NewOwnerCreateRegistry(
		importassembly.OwnerRegistryDependencies{
			Postgres:         postgresHandle,
			RevisionAppender: revisionRuntime.Appender(),
			Intents:          intentAppender,
			Timeline:         timelineFacade,
		},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Imports owner registry: %w", err)
	}
	importRoutes := imports.RegisterRoutes(
		imports.WithLimits(importOwnerLimits, importArchiveLimits),
		imports.WithOwnerCreateRegistry(importOwnerRegistry),
		imports.WithRevisionAppender(revisionRuntime.Appender()),
		imports.WithExtensionProfileAdmission(func(profileID string) bool {
			return profileID == networkflow.ProfileID && networkFlowRouteAdmitted
		}),
		imports.WithJobSuccessFinalizer(extensionassembly.NewImportJobSuccessFinalizer(
			runtime.ExtensionJobFinalizer,
			postgresHandle,
			runtime.JobTransactions,
			now,
		)),
	)
	referencePackRoutes := reference_data.RegisterRoutes(
		reference_data.WithStorage(referencePackStorage),
		reference_data.WithLimits(referenceDataLimits(normalizedCfg)),
		reference_data.WithJobSuccessFinalizer(
			extensionassembly.NewReferencePackJobSuccessFinalizer(runtime.ExtensionJobFinalizer),
		),
	)
	var renderExportInvoker reporting.RenderExportInvoker
	if snapshotReportingParticipantAdmitted {
		renderExportInvoker, err = extensionassembly.NewAdmittedRenderExportInvoker(
			publicationCatalog,
			reporting.BuiltInRenderExportParticipant{},
			time.Duration(normalizedCfg.Timeouts.Extensions.TransactionParticipantSeconds)*time.Second,
		)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose Snapshot/Reporting participant: %w", err)
		}
	}
	reportingRoutes := reporting.RegisterRoutes(reporting.RouteOptions{
		JobSuccessFinalizer: extensionassembly.NewReportingJobSuccessFinalizer(runtime.ExtensionJobFinalizer),
		RenderExportInvoker: renderExportInvoker,
	})
	var compositionPreviewJobs reportcomposition.PreviewJobPort
	if snapshotReportingRoutesAdmitted {
		compositionPreviewJobs, err = reporting.NewCompositionPreviewJobPort(runtime.JobRunner, runtime.JobTransactions)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose Report Composition preview job port: %w", err)
		}
	}
	reportCompositionRoutes := reportcomposition.RegisterRoutes(reportcomposition.RouteOptions{
		PreviewJobs: compositionPreviewJobs,
	})
	moduleOverrides := mergeNetworkFlowImportFacadeOverride(testRuntimeDeps.ModuleOverrides, networkFlowModule.ImportOwner())
	delete(moduleOverrides, networkflow.KeyRingsOverrideKey)
	delete(moduleOverrides, postgresDBDecoratorOverrideKey)
	authRouteOptions := []auth.RouteOption{}
	authRouteOptions = append(
		authRouteOptions,
		auth.WithPublicOrigin(normalizedCfg.Application.PublicOrigin),
		auth.WithSessionRevocations(hub),
	)
	if enterpriseAuthenticationAdmitted {
		authRouteOptions = append(authRouteOptions, auth.WithEnterpriseAuthBindings())
	}
	workbookConflictTokens := conflicttokens.NewConflictTokenCodec(keys)
	workbookContributionCatalog, err := workbookassembly.NewContributionCatalog(
		postgresHandle,
		timelineBundle.ProjectionCatalog.Catalog,
		timelineBundle.ProjectionCatalog.Query,
		timelineFacade,
		workbookConflictTokens,
		revisionRuntime.Appender(),
		intentAppender,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Workbook contribution catalog: %w", err)
	}
	builtInRoutes, err := applicationRouteRegistrars([]routeContribution{
		{id: "auth", registrar: auth.RegisterRoutes(authRouteOptions...)},
		{id: "incidents", registrar: incidentRoutes},
		{id: "extensions", registrar: extensiondiscovery.RegisterRoutes()},
		{id: "jobs", registrar: jobapi.RegisterRoutes()},
		{id: "saved_views", registrar: savedviews.RegisterRoutes()},
		{id: "view_schemas", registrar: viewschemas.RegisterRoutes()},
		{id: "collaboration", registrar: collaboration.RegisterRoutes(collaborationSettings(normalizedCfg, hub))},
		{id: "entities", registrar: entities.RegisterRoutes(entities.RouteOptions{
			MergeStore:   timelineBundle.EntityMergeStore,
			MentionStore: timelineBundle.EntityMentionStore,
		})},
		{id: "evidence", registrar: evidence.RegisterRoutes(
			evidenceSettings(normalizedCfg),
			evidence.WithStore(timelineBundle.EvidenceStore),
		)},
		{
			id: "workbook",
			registrar: workbook.RegisterRoutes(workbook.RouteDependencies{
				TimelineOwner: timelineFacade,
				MutationStore: workbookassembly.NewMutationStore(
					postgresHandle,
					workbookContributionCatalog,
					revisionRuntime.Appender(),
				),
				EntityOwner:         hostidentity.NewStore(postgresHandle, revisionRuntime.Appender()),
				ConflictTokens:      workbookConflictTokens,
				StartupStoreFactory: workbookassembly.NewStartupStoreFromDependencies,
			}),
		},
		{id: "timeline", registrar: timelineadmission.RegisterRoutes(timelineadmission.RouteOptions{
			Facade: timelineFacade,
		})},
		{id: "revisions", registrar: revisionRoutes},
	}, []extensionRouteBinding{
		{
			id: "enterprise_authentication_routes",
			contributionIDs: []string{
				auth.EnterpriseOIDCRouteContributionID,
				auth.EnterpriseProvidersRouteContributionID,
				auth.EnterpriseSAMLRouteContributionID,
			},
			registrar: auth.RegisterEnterpriseRoutes(auth.WithPublicOrigin(normalizedCfg.Application.PublicOrigin)),
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
			contributionIDs: []string{reporting.ReleasesRouteContributionID, reporting.SnapshotsRouteContributionID},
			registrar:       reportingRoutes,
		},
		{
			id:              "snapshot_reporting_compositions",
			contributionIDs: []string{reporting.ReportCompositionsRouteContributionID},
			registrar:       reportCompositionRoutes,
		},
	}, publicationCatalog)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose built-in routes: %w", err)
	}
	httpOptions.AdditionalRoutes = append(builtInRoutes, httpOptions.AdditionalRoutes...)
	httpOptions.ValidatePublicRoutes = true
	readinessProbes := []httpapi.DependencyReadinessProbe{}
	if runtime.StagedHealth != nil {
		readinessProbes = append(readinessProbes, stagedCleanupReadinessProbe{health: runtime.StagedHealth})
	}
	httpOptions.Dependencies = httpapi.DependencySet{
		Telemetry:           httpapi.TelemetrySettings{Enabled: normalizedCfg.Telemetry.Enabled, ServiceVersion: normalizedCfg.Telemetry.Resource.ServiceVersion},
		Env:                 options.Env,
		Postgres:            runtime.Postgres,
		PostgresDB:          postgresHandle,
		ObjectStore:         runtime.ObjectStore,
		Jobs:                runtime.Jobs,
		JobTransactions:     runtime.JobTransactions,
		JobRunner:           runtime.JobRunner,
		CursorCodec:         cursorCodec,
		Readiness:           httpapi.NewDependencyReadinessChecker(runtime.Postgres, runtime.ObjectStore, readinessProbes...),
		Admission:           runtime.Lifecycle,
		PublicErrorFaults:   testRuntimeDeps.PublicErrorFaults,
		TestResetBootstrap:  testResetBootstrap(normalizedCfg),
		ModuleOverrides:     moduleOverrides,
		ExtensionDiscovery:  publicationHTTPProjections{publication: publication},
		ExtensionClaims:     publicationHTTPProjections{publication: publication},
		ExtensionRoutes:     publicationHTTPProjections{publication: publication},
		ExtensionWorkspaces: publicationHTTPProjections{publication: publication},
		Now:                 now,
	}

	if err := publication.Commit(); err != nil {
		runtime.Close()
		return nil, err
	}
	if httpOptions.Dependencies.PublicRoutes == nil {
		httpOptions.Dependencies.PublicRoutes, err = httpapi.NewRouteRegistry(
			httpapi.ExtensionClaimsFromDependencies(httpOptions.Dependencies),
		)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("initialize public route registry: %w", err)
		}
	}
	handler, err := newHTTPHandler(httpOptions)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("setup http handler: %w", err)
	}

	runtime.Handler = handler
	runtime.PublicHTTP = httpOptions.Dependencies.PublicRoutes.Diagnostics()
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

func loadNetworkFlowKeyRings(
	configuration networkflow.Configuration,
	env map[string]string,
	now time.Time,
	registry *secretpurpose.Registry,
) (*networkflow.KeyRings, error) {
	if !configuration.Claimed {
		return nil, nil
	}
	document, err := readSecureFile(configuration.KeyRingManifestPath, networkflow.KeyRingManifestMaximumSize)
	if err != nil {
		failure := networkflow.ManifestUnreadable
		var secureError *securefile.Error
		if errors.As(err, &secureError) {
			switch secureError.Kind {
			case securefile.FailureTooLarge:
				failure = networkflow.ManifestTooLarge
			case securefile.FailureInvalidPath, securefile.FailureUnsafeObject, securefile.FailureChanged:
				failure = networkflow.ManifestUnsafe
			}
		}
		return nil, deploymentNetworkFlowError(networkflow.KeyRingManifestReadError(failure))
	}
	rings, err := networkflow.ParseKeyRingsWithRegistry(document.Bytes(), env, now, registry)
	if err != nil {
		return nil, deploymentNetworkFlowError(err)
	}
	return rings, nil
}

type enterpriseDocumentReader struct{}

func loadEnterpriseProviderManifest(
	configuration enterpriseauth.Configuration,
	env map[string]string,
) ([]authn.EnterpriseAuthProviderDefinition, error) {
	definitions, err := enterpriseauth.LoadProviderManifest(configuration, env, enterpriseDocumentReader{})
	if err != nil {
		return nil, deploymentEnterpriseAuthenticationError(err)
	}
	return definitions, nil
}

func (enterpriseDocumentReader) ReadDocument(absolutePath string, maximumBytes int64) ([]byte, enterpriseauth.DocumentReadFailure) {
	document, err := readSecureFile(absolutePath, maximumBytes)
	if err == nil {
		return document.Bytes(), ""
	}
	failure := enterpriseauth.DocumentUnavailable
	var secureError *securefile.Error
	if errors.As(err, &secureError) {
		switch secureError.Kind {
		case securefile.FailureTooLarge:
			failure = enterpriseauth.DocumentTooLarge
		case securefile.FailureInvalidPath, securefile.FailureUnsafeObject, securefile.FailureChanged:
			failure = enterpriseauth.DocumentUnsafe
		}
	}
	return nil, failure
}

func deploymentEnterpriseAuthenticationError(err error) error {
	finding, ok := enterpriseauth.ConfigurationFindingFromError(err)
	if !ok {
		return err
	}
	return config.NewDiagnosticsError(config.Diagnostic{
		Path:       finding.Path,
		ReasonCode: finding.ReasonCode,
		Message:    finding.Message,
	})
}

func deploymentNetworkFlowError(err error) error {
	finding, ok := networkflow.ConfigurationFindingFromError(err)
	if !ok {
		return err
	}
	return config.NewDiagnosticsError(config.Diagnostic{
		Path:       finding.Path,
		ReasonCode: finding.ReasonCode,
		Message:    finding.Message,
	})
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

func instrumentedPostgres(enabled bool, serviceVersion string, pool *pgxpool.Pool) postgres.DB {
	if pool == nil || !enabled {
		return pool
	}
	return postgres.InstrumentDB(pool, serviceVersion)
}

func instrumentedObjectStore(enabled bool, serviceVersion string, store objectstore.Store) objectstore.Store {
	if store == nil || !enabled {
		return store
	}
	return objectstore.InstrumentStore(store, serviceVersion)
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

type publicationHTTPProjections struct {
	publication *PublicationController
}

func (provider publicationHTTPProjections) ExtensionDiscoveryProfiles() []httpapi.ExtensionProfile {
	if provider.publication == nil {
		return nil
	}
	return publicationHTTPProfiles(provider.publication.Discovery())
}

func (provider publicationHTTPProjections) ExtensionClaims() []httpapi.ExtensionClaim {
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

func (provider publicationHTTPProjections) ExtensionRoutes() []httpapi.ExtensionRoute {
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

func (provider publicationHTTPProjections) ExtensionWorkspaces() []httpapi.ExtensionWorkspacePublication {
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
		recoveryCtx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Settings.ReconciliationSeconds)*time.Second)
		defer cancel()
		if err := r.JobRunner.Activate(recoveryCtx); err != nil {
			r.Publication.ComponentLost("job_dequeue")
			return fmt.Errorf("activate extension job recovery: %w", err)
		}
	}
	if r.CollaborationDispatcher != nil {
		if err := r.CollaborationDispatcher.Start(context.Background()); err != nil {
			r.Publication.ComponentLost("collaboration_dispatcher")
			return fmt.Errorf("activate collaboration dispatcher: %w", err)
		}
	}
	r.publicationOnce.Do(func() {
		if r.StagedJanitorLeader != nil {
			r.StagedJanitorLeader.Start(context.Background())
		} else if r.StagedJanitor != nil && r.stagedJanitorContext != nil {
			go func() {
				defer func() {
					if recovered := recover(); recovered != nil && r.Lifecycle != nil {
						r.Publication.ComponentLost("staged_object_janitor")
					}
				}()
				if err := r.StagedJanitor.Run(r.stagedJanitorContext, time.Duration(r.Settings.StagedObjectSweepSeconds)*time.Second); err != nil && r.Lifecycle != nil {
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
