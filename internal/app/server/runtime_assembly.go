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
	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	hostidentityreporting "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/reportingprovider"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatorshttpapi "github.com/JochiRaider/cartulary/internal/modules/indicators/httpapi"
	"github.com/JochiRaider/cartulary/internal/modules/jobapi"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	revisionshttpapi "github.com/JochiRaider/cartulary/internal/modules/revisions/httpapi"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	timelineadmission "github.com/JochiRaider/cartulary/internal/modules/timeline/admission"
	"github.com/JochiRaider/cartulary/internal/modules/viewschemas"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
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

type Options struct {
	Env                    map[string]string
	HTTP                   httpapi.Options
	Postgres               *pgxpool.Pool
	ObjectStore            objectstore.Store
	Now                    func() time.Time
	ObserveJobs            func(*jobs.Manager, *jobs.TransactionService, *jobs.Runner, *pgxpool.Pool)
	ObserveCollaboration   func(*collaboration.Hub, *collaboration.Dispatcher, collaboration.IntentAppender)
	ObserveEvidenceCleanup func(*evidence.CleanupDispatcher)
	ObserveProjections     func(*projectionassembly.Runtime)
	ObserveTimeline        func(*timelineassembly.Bundle)
	ObserveRevisions       func(*revisionassembly.Runtime)
}

type Runtime struct {
	handler                   http.Handler
	stagedJanitor             *stagedobjects.Janitor
	jobRunner                 *jobs.Runner
	collaborationDispatcher   *collaboration.Dispatcher
	evidenceCleanupDispatcher *evidence.CleanupDispatcher
	processLease              *processlease.ApplicationProcessLease
	servingLease              *processlease.ApplicationRecoveryServingLease
	lifecycle                 *processlifecycle.Controller
	publication               *publicationController
	publicHTTP                httpapi.RouteDiagnostics
	shutdownDrainTimeout      time.Duration
	reconciliationTimeout     time.Duration
	stagedObjectSweepPeriod   time.Duration

	closeOnce            sync.Once
	publicationOnce      sync.Once
	cleanups             []func()
	stagedJanitorContext context.Context
}

const serverServingLeaseAcquireMax = time.Second

type runtimeSettings struct {
	TelemetryFlushTimeoutMS  int64
	ReconciliationSeconds    int64
	StagedObjectSweepSeconds int64
	ShutdownDrainSeconds     int64
}

type runtimeAssembly struct {
	loadedConfiguration configassembly.Loaded
	options             Options
	dependencies        runtimeDependencies
}

func (assembly runtimeAssembly) build(ctx context.Context) (*Runtime, error) {
	loadedConfiguration := assembly.loadedConfiguration
	options := assembly.options
	dependencies := assembly.dependencies
	extensionCoordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		return nil, fmt.Errorf("admit packaged extension registry: %w", err)
	}
	if err := loadedConfiguration.ValidateForStartup(); err != nil {
		return nil, err
	}
	normalizedCfg := loadedConfiguration.Deployment()
	enterpriseAuthenticationConfiguration := normalizedCfg.EnterpriseAuthentication
	networkFlowConfiguration := normalizedCfg.NetworkFlowActivity
	settingsProjection := newApplicationSettingsProjection(normalizedCfg)
	runtimeSettings := settingsProjection.Runtime()
	socketTransport := newCollaborationSocketTransport(normalizedCfg.Application.PublicOrigin)

	runtime := &Runtime{
		lifecycle:               processlifecycle.New(),
		shutdownDrainTimeout:    time.Duration(runtimeSettings.ShutdownDrainSeconds) * time.Second,
		reconciliationTimeout:   time.Duration(runtimeSettings.ReconciliationSeconds) * time.Second,
		stagedObjectSweepPeriod: time.Duration(runtimeSettings.StagedObjectSweepSeconds) * time.Second,
	}
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	postgresPool := options.Postgres
	if postgresPool == nil {
		postgresSettings, settingsErr := postgres.ResolveSettings(configassembly.PostgresBinding(normalizedCfg), postgres.PurposeRuntime, options.Env)
		if settingsErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup postgres: %w", settingsErr)
		}
		pool, err := dependencies.setupPostgres(ctx, postgresSettings)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup postgres: %w", err)
		}
		postgresPool = pool
		if pool != nil {
			runtime.own(pool.Close)
		}
	}
	lease, leaseErr := dependencies.acquireApplicationProcessLease(
		ctx,
		postgresPool,
		time.Duration(normalizedCfg.Timeouts.Extensions.ProcessLeaseAcquireSeconds)*time.Second,
		time.Duration(normalizedCfg.Timeouts.Extensions.ProcessLeaseLossDetectionSeconds)*time.Second,
	)
	if leaseErr != nil {
		runtime.Close()
		return nil, leaseErr
	}
	runtime.processLease = lease
	if lease != nil {
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

	servingLease, servingLeaseErr := dependencies.acquireRecoveryServingLease(
		ctx,
		postgresPool,
		min(
			time.Duration(normalizedCfg.Timeouts.Extensions.ProcessLeaseAcquireSeconds)*time.Second,
			serverServingLeaseAcquireMax,
		),
		time.Duration(normalizedCfg.Timeouts.Extensions.ProcessLeaseLossDetectionSeconds)*time.Second,
	)
	if servingLeaseErr != nil {
		runtime.Close()
		return nil, servingLeaseErr
	}
	runtime.servingLease = servingLease
	if servingLease != nil {
		runtime.own(func() {
			releaseCtx, cancel := context.WithTimeout(context.Background(), time.Duration(normalizedCfg.Timeouts.Extensions.ProcessLeaseLossDetectionSeconds)*time.Second)
			defer cancel()
			if servingLease.State() == processlease.StateHeld {
				_ = servingLease.Release(releaseCtx)
			} else {
				servingLease.Close()
			}
		})
		monitorCtx, cancelMonitor := context.WithCancel(context.Background())
		runtime.own(cancelMonitor)
		servingLease.StartMonitor(monitorCtx)
		go runtime.watchServingLease(monitorCtx)
	}

	rootIncidentBundleStorage, storageErr := incidentportabilityassembly.NewRootStorage(
		normalizedCfg.Roots.TemporaryWork.Path,
		normalizedCfg.Roots.ExportOutputs.Path,
	)
	if storageErr != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose incident bundle storage: %w", storageErr)
	}
	var incidentBundleStorage incidentbundles.BundleStorage = rootIncidentBundleStorage
	runtime.own(rootIncidentBundleStorage.Close)
	rootReferencePackStorage, storageErr := referenceassembly.NewRootStorage(
		normalizedCfg.Roots.TemporaryWork.Path,
		normalizedCfg.Roots.ReferencePackStorage.Path,
	)
	if storageErr != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Reference Pack storage: %w", storageErr)
	}
	var referencePackStorage reference_data.Storage = rootReferencePackStorage
	runtime.own(rootReferencePackStorage.Close)

	// Both application-owned leases are held before claim resolution and every
	// Stage-1-or-later effect. Their separate typed identities prevent the
	// Recovery serving vocabulary from collapsing into the process-lease one.

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
	requestedClaims := loadedConfiguration.RequestedClaims()
	claimResolution, err := extensionCoordinator.ResolveClaims(requestedClaims.ProfileIDs())
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
	publication, err := preparePublicationOrchestrator(runtime.lifecycle, extensionPlan)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	runtime.publication = publication.controller
	resolvedClaims := extensionPlan.ResolvedClaims()

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
		enterpriseProviderDefinitions, err = loadEnterpriseProviderManifest(
			enterpriseAuthenticationConfiguration,
			options.Env,
			dependencies.readSecureFile,
		)
		if err != nil {
			runtime.Close()
			return nil, err
		}
		if err := enterpriseauth.RegisterProviderSecretPurposes(enterpriseProviderDefinitions, options.Env, secretPurposes); err != nil {
			runtime.Close()
			return nil, deploymentEnterpriseAuthenticationError(err)
		}
	}
	revisionsConflictTokenKeyRing, err := loadRevisionsConflictTokenKeyRing(
		normalizedCfg.Revisions,
		options.Env,
		now(),
		secretPurposes,
		dependencies.readSecureFile,
	)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	telemetryRuntime, err := telemetry.Bootstrap(ctx, normalizedCfg.Telemetry, normalizedCfg.DeploymentProfile, options.Env, telemetry.WithResolvedClaimIdentity(telemetry.ResolvedClaimIdentity{
		ProfileIDs: resolvedClaims.ProfileIDs(),
		SHA256:     resolvedClaims.SHA256(),
	}))
	if err != nil {
		runtime.Close()
		return nil, err
	}
	telemetryFlushTimeout := time.Duration(runtimeSettings.TelemetryFlushTimeoutMS) * time.Millisecond
	runtime.own(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), telemetryFlushTimeout)
		defer cancel()
		_ = telemetryRuntime.Shutdown(shutdownCtx)
	})

	migrationSource, err := dbmigrations.Source()
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("load migration source: %w", err)
	}
	if err := dependencies.ensureSchemaReady(ctx, postgresPool, migrationSource); err != nil {
		runtime.Close()
		var remediation database_migrations.RemediationReporter
		if errors.As(err, &remediation) {
			return nil, remediation
		}
		var migrationFailure database_migrations.MigrationFailure
		if errors.As(err, &migrationFailure) {
			return nil, config.NewDiagnosticsError(config.Diagnostic{
				Path:       "database.schema_version",
				ReasonCode: migrationFailure.ReasonCode(),
				Message:    "Database migration validation failed.",
			})
		}
		return nil, err
	}
	var extensionStateStore *extensionstore.Store
	if postgresPool != nil {
		stateStore, stateStoreErr := extensionstore.New(postgresPool, networkflow.ExtensionStateFamilyCounters())
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
				runtime.lifecycle.Fatal(reason)
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
		extensionStateStore = stateStore
	}
	var objectStore objectstore.Store
	if options.ObjectStore != nil {
		objectStore = instrumentedObjectStore(
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
		client, err := dependencies.setupObjectStore(ctx, objectStoreSettings, configassembly.ObjectStoreInstrumentation(normalizedCfg))
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup object store: %w", err)
		}
		objectStore = client
		if client != nil {
			runtime.own(func() { _ = client.Close() })
		}
	}
	var stagedObjectService *stagedobjects.Service
	var stagedHealth *stagedobjects.Health
	if extensionStateStore != nil && objectStore != nil {
		stagedRepository, stagedRepositoryErr := extensionassembly.NewStagedObjectRepository(extensionStateStore)
		if stagedRepositoryErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose staged-object repository: %w", stagedRepositoryErr)
		}
		stagedBytes, stagedBytesErr := extensionassembly.NewStagedObjectBytes(objectStore)
		if stagedBytesErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose staged-object byte store: %w", stagedBytesErr)
		}
		stagedService, stagedServiceErr := stagedobjects.NewService(stagedobjects.ServiceOptions{
			Repository: stagedRepository,
			Bytes:      stagedBytes,
			Now:        now,
			FatalSink: func(error) {
				runtime.lifecycle.Fatal("staged_object_publication_mismatch")
			},
		})
		if stagedServiceErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose staged-object service: %w", stagedServiceErr)
		}
		stagedObjectService = stagedService
		stagedHealth = stagedobjects.NewHealth()
		janitor, janitorErr := stagedobjects.NewJanitor(stagedobjects.JanitorOptions{
			Repository:       stagedRepository,
			Bytes:            stagedBytes,
			Health:           stagedHealth,
			Now:              now,
			BatchLimit:       int(normalizedCfg.Limits.Extensions.StagedObjectCleanupBatch),
			OperationTimeout: time.Duration(normalizedCfg.Timeouts.Extensions.StagedObjectCleanupSeconds) * time.Second,
			FatalSink: func(error) {
				runtime.lifecycle.Fatal("staged_object_publication_mismatch")
			},
		})
		if janitorErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose staged-object janitor: %w", janitorErr)
		}
		runtime.stagedJanitor = janitor
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
	}
	postgresHandle := instrumentedPostgres(
		normalizedCfg.Telemetry.Enabled,
		normalizedCfg.Telemetry.Resource.ServiceVersion,
		postgresPool,
	)
	if err := dependencies.runBootstrap(ctx, configassembly.BootstrapSettings(normalizedCfg), postgresPool); err != nil {
		runtime.Close()
		return nil, err
	}
	if enterpriseAuthenticationAdmitted {
		if err := enterpriseauth.ReconcileProviderDefinitions(ctx, enterpriseProviderDefinitions, authn.NewStore(postgresHandle), now()); err != nil {
			runtime.Close()
			return nil, deploymentEnterpriseAuthenticationError(err)
		}
	}
	networkFlowKeyRings, err := loadNetworkFlowKeyRings(
		networkFlowConfiguration,
		options.Env,
		now(),
		secretPurposes,
		dependencies.readSecureFile,
	)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	if referencePackRouteAdmitted {
		referenceLimits := settingsProjection.ReferenceData()
		if err := reference_data.EnsureMinimumDisconnectedBundle(ctx, reference_data.MinimumDisconnectedBundleOptions{
			DeploymentProfile: normalizedCfg.DeploymentProfile,
			ArchiveLimits:     referenceLimits.Archives,
			ReferenceLimits:   referenceLimits.ReferencePacks,
			Storage:           referencePackStorage,
		}, postgresPool, now()); err != nil {
			runtime.Close()
			return nil, fmt.Errorf("seed minimum disconnected reference packs: %w", err)
		}
	}
	cleanupService, err := evidence.NewCleanupService(postgresHandle)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Evidence cleanup service: %w", err)
	}
	cleanupDeleter, err := evidence.NewCleanupObjectDeleter(objectStore)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	cleanupObserver, err := newEvidenceCleanupTelemetryObserver(normalizedCfg.Telemetry.Resource.ServiceVersion)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Evidence cleanup telemetry: %w", err)
	}
	runtime.evidenceCleanupDispatcher, err = evidence.NewCleanupDispatcher(
		cleanupService,
		cleanupDeleter,
		cleanupObserver,
		now,
	)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	cleanupDispatcher := runtime.evidenceCleanupDispatcher
	runtime.own(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		_ = cleanupDispatcher.Close(closeCtx)
	})
	hub := dependencies.newCollaborationHub()
	intentAppender := collaboration.NewIntentAppender()
	runtime.collaborationDispatcher = collaboration.NewDispatcher(postgresHandle, hub, now)
	dispatcher := runtime.collaborationDispatcher
	runtime.own(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = dispatcher.Close(closeCtx)
	})
	intentAdapters := newCollaborationIntentTranslator(intentAppender)
	jobOwnerPorts := jobOwnerTransactionAdapters{}
	extensionJobDefinitions, err := extensionassembly.JobDefinitions(publicationCatalog)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose extension job definitions: %w", err)
	}
	recognizedJobDefinitions, err := extensionassembly.RecognizedJobDefinitions(extensionCoordinator.JobKindContracts())
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose recognized extension job definitions: %w", err)
	}
	jobCatalog, err := jobs.NewCatalog(recognizedJobDefinitions)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Jobs catalog: %w", err)
	}
	jobSelection, err := jobs.NewRuntimeSelection(jobCatalog, extensionassembly.JobKinds(extensionJobDefinitions))
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Jobs runtime selection: %w", err)
	}
	jobTransactions, err := jobs.NewTransactionService(intentAdapters, jobs.OwnerTransactionPorts{
		RouteIdempotency:      jobOwnerPorts,
		ExtensionCancellation: jobOwnerPorts,
	}, jobCatalog, jobSelection)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Jobs transaction service: %w", err)
	}
	jobPolicy := jobs.ProductionRuntimePolicy()
	jobManager, err := dependencies.newJobsManager(jobs.ManagerOptions{
		Postgres:                postgresPool,
		Transactions:            jobTransactions,
		Catalog:                 jobCatalog,
		Policy:                  jobPolicy,
		Now:                     now,
		TelemetryServiceVersion: normalizedCfg.Telemetry.Resource.ServiceVersion,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Jobs manager: %w", err)
	}
	runtime.jobRunner, err = jobs.NewRunner(jobs.RunnerOptions{
		Manager:     jobManager,
		Catalog:     jobCatalog,
		Policy:      jobPolicy,
		DequeueGate: runtime.lifecycle,
		OnComponentLoss: func() {
			if runtime.publication != nil {
				runtime.publication.componentLost("job_dequeue")
			}
		},
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Jobs runner: %w", err)
	}
	jobRunner := runtime.jobRunner
	runtime.own(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = jobRunner.Close(ctx)
	})
	if err := jobManager.ValidateStorageCatalog(ctx); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("validate Jobs storage catalog: %w", err)
	}
	if extensionStateStore != nil {
		inactiveJobStore, err := extensionassembly.NewInactiveJobStore(extensionStateStore, jobTransactions, now)
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
				runtime.lifecycle.Fatal("indeterminate_database_commit")
			},
		)
		cancelReconciliation()
		if reconciliationErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("reconcile inactive extension jobs: %w", reconciliationErr)
		}
	}
	var extensionJobFinalizer *extensionstore.OwnerFinalizer
	if extensionStateStore != nil {
		extensionJobFinalizer, err = extensionstore.NewOwnerFinalizer(
			extensionStateStore,
			jobTransactions,
			jobOwnerPorts,
			now,
			func(error) {
				runtime.lifecycle.Fatal("indeterminate_database_commit")
			},
		)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose extension job finalizer: %w", err)
		}
	}
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
	providerContributions, err := revisionassembly.CurrentProviderContributions()
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Revisions provider contributions: %w", err)
	}
	revisionRuntime, err := revisionassembly.Build(
		revisionassembly.Dependencies{
			HistoricalIntentPolicy: historicalIntentPolicy,
			IntentAppender:         intentAppender,
		},
		providerContributions...,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Revisions runtime: %w", err)
	}
	workbookConflictTokens, err := conflicttokens.NewConflictTokenCodec(
		revisionsConflictTokenKeyRing,
		conflicttokens.WithClock(now),
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Revisions conflict-token codec: %w", err)
	}
	projectionRuntime, err := projectionassembly.Build(postgresHandle)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Projections runtime: %w", err)
	}
	evidenceOwner, err := evidence.NewOwnerRuntime(evidence.OwnerRuntimeDependencies{
		Postgres:            postgresHandle,
		ConflictTokens:      &workbookConflictTokens,
		Revisions:           revisionRuntime.Appender(),
		Collaboration:       intentAppender,
		ObjectStore:         objectStore,
		ConflictFields:      revisionRuntime.ConflictFieldResolver(),
		ConflictIdempotency: workbookassembly.NewConflictIdempotencyPort(postgresHandle),
		Projections:         projectionRuntime.EvidencePorts(),
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Evidence owner: %w", err)
	}
	timelineBundle, err := timelineassembly.NewBundle(timelineassembly.Dependencies{
		Postgres:            postgresHandle,
		ConflictTokens:      workbookConflictTokens,
		Revisions:           revisionRuntime.Appender(),
		Collaboration:       intentAppender,
		EvidenceAttachments: evidenceOwner.TimelineAttachmentContribution(),
		TimelineProjection:  projectionRuntime.TimelinePorts().Writer,
		EntityProjection:    projectionRuntime.EntityPorts().Writer,
		AssessmentRows:      projectionRuntime.AssessmentPorts().Rows,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Timeline bundle: %w", err)
	}
	indicatorOwner, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:    postgresHandle,
		Revisions:   revisionRuntime.Appender(),
		Projections: projectionRuntime.IndicatorPorts().Rows,
		SourceText:  indicatorassembly.NewSourceTextPort(projectionRuntime.SourceTextRows()),
		Clock:       now,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Indicators owner: %w", err)
	}
	revisionCommands, err := revisionRuntime.NewCommandService(
		postgresHandle,
		attributionResolvers.ImportedAttributionResolver(incidentbundles.IncidentPortabilityProfileID),
		projectionRuntime.RevisionRebuilder(),
		projectionRuntime.RevisionLiveRecords(),
		now,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose revisions command service: %w", err)
	}
	revisionRoutes := revisionshttpapi.RegisterRoutes(
		revisionCommands,
		revisionassembly.NewRecordEnvelopeReader(postgresHandle),
	)
	importStore := imports.NewStore(
		postgresPool,
		revisionRuntime.Appender(),
		jobTransactions,
	)
	timelineFacade := timelineBundle.Facade
	networkFlowModule, err := networkflow.NewModule(networkflow.ModuleDependencies{
		Postgres:        postgresHandle,
		ImportSources:   importStore,
		KeyRings:        networkFlowKeyRings,
		Now:             now,
		IncidentLocks:   incidents.NewTransactionParticipant(),
		AuditAppender:   authn.NewAdministrativeAuditAppender(),
		Indicators:      indicatorOwner,
		ResourceIntents: intentAdapters,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Network Flow module: %w", err)
	}
	incidentBundleImportTransactions, err := incidentbundles.NewImportTransactionProvider(
		postgresPool,
		objectStore,
		incidentBundleImportFinalizer,
		projectionRuntime.ImportRebuilder(),
		historicalIntentPolicy,
		jobTransactions,
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
			runtime.lifecycle.Fatal("indeterminate_database_commit")
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
		stagedObjectService,
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
		incidentbundles.WithJobs(jobTransactions, jobManager, runtime.jobRunner),
		incidentbundles.WithStorage(incidentBundleStorage),
		incidentbundles.WithLimits(settingsProjection.IncidentBundles()),
		incidentbundles.WithImportFinalizer(incidentBundleImportFinalizer),
		incidentbundles.WithJobSuccessFinalizer(
			extensionassembly.NewIncidentBundleJobSuccessFinalizer(extensionJobFinalizer, now),
		),
		incidentbundles.WithPortability(portability, crossOwnerCoordinator),
		incidentbundles.WithProjectionRebuild(projectionRuntime.ImportRebuilder()),
		incidentbundles.WithSourceCatalog(incidentSourceCatalog),
		incidentbundles.WithHistoricalIntentPolicy(historicalIntentPolicy),
	)
	importOwnerLimits, importArchiveLimits := settingsProjection.Imports()
	importOwnerRegistry, err := importassembly.NewOwnerCreateRegistry(
		importassembly.OwnerRegistryDependencies{
			Postgres:                postgresHandle,
			RevisionAppender:        revisionRuntime.Appender(),
			Intents:                 intentAppender,
			Timeline:                timelineFacade,
			EntityProjections:       projectionRuntime.EntityPorts().Writer,
			AssessmentProjections:   projectionRuntime.AssessmentPorts().Rows,
			ArtifactProjections:     projectionRuntime.ArtifactPorts().Rows,
			Evidence:                evidenceOwner.ImportCreateFacade(),
			PartyProjections:        projectionRuntime.PartyPorts().Rows,
			TaskDecisionProjections: projectionRuntime.TaskDecisionPorts().Rows,
			Indicators:              indicatorOwner,
		},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Imports owner registry: %w", err)
	}
	importRoutes := imports.RegisterRoutes(
		imports.WithJobs(jobTransactions, jobManager, runtime.jobRunner),
		imports.WithLimits(importOwnerLimits, importArchiveLimits),
		imports.WithOwnerCreateRegistry(importOwnerRegistry),
		imports.WithRevisionAppender(revisionRuntime.Appender()),
		imports.WithExtensionProfileAdmission(func(profileID string) bool {
			return profileID == networkflow.ProfileID && networkFlowRouteAdmitted
		}),
		imports.WithJobSuccessFinalizer(extensionassembly.NewImportJobSuccessFinalizer(
			extensionJobFinalizer,
			postgresHandle,
			jobTransactions,
			now,
		)),
	)
	referencePackRoutes := reference_data.RegisterRoutes(
		reference_data.WithJobs(jobTransactions, jobManager, runtime.jobRunner),
		reference_data.WithStorage(referencePackStorage),
		reference_data.WithLimits(settingsProjection.ReferenceData()),
		reference_data.WithJobSuccessFinalizer(
			extensionassembly.NewReferencePackJobSuccessFinalizer(extensionJobFinalizer),
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
	hostIdentityReporting, err := hostidentityreporting.New(projectionRuntime.EntityPorts().Reader)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	artifactReporting, err := artifacts.NewReportingContribution(projectionRuntime.ArtifactPorts().Reader)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	taskDecisionReporting, err := tasksdecisions.NewReportingContribution(projectionRuntime.TaskDecisionPorts().Reader)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	reportingRouteOptions := reporting.WithJobs(reporting.RouteOptions{
		JobSuccessFinalizer: extensionassembly.NewReportingJobSuccessFinalizer(extensionJobFinalizer),
		RenderExportInvoker: renderExportInvoker,
		ExportFieldProviders: []exportprovider.FieldProvider{
			artifactReporting,
			hostIdentityReporting,
			taskDecisionReporting,
		},
	}, jobTransactions, jobManager, runtime.jobRunner)
	reportingRoutes := reporting.RegisterRoutes(reportingRouteOptions)
	var compositionPreviewJobs reportcomposition.PreviewJobPort
	if snapshotReportingRoutesAdmitted {
		compositionPreviewJobs, err = reporting.NewCompositionPreviewJobPort(runtime.jobRunner, jobTransactions)
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
	authRouteOptions := []auth.RouteOption{}
	authRouteOptions = append(
		authRouteOptions,
		auth.WithPublicOrigin(normalizedCfg.Application.PublicOrigin),
		auth.WithSessionRevocations(hub),
	)
	if enterpriseAuthenticationAdmitted {
		authRouteOptions = append(authRouteOptions, auth.WithEnterpriseAuthBindings())
	}
	taskDecisionMutation, err := workbookassembly.NewTaskDecisionMutationContribution(
		postgresHandle,
		workbookConflictTokens,
		revisionRuntime.Appender(),
		revisionRuntime.ConflictFieldResolver(),
		projectionRuntime.TaskDecisionPorts().Rows,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Workbook Tasks/Decisions mutation contribution: %w", err)
	}
	artifactMutation, err := workbookassembly.NewArtifactMutationContribution(
		postgresHandle,
		workbookConflictTokens,
		revisionRuntime.Appender(),
		revisionRuntime.ConflictFieldResolver(),
		projectionRuntime.ArtifactPorts().Rows,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Workbook Artifacts mutation contribution: %w", err)
	}
	workbookContributionCatalog, err := workbookassembly.NewContributionCatalog(
		postgresHandle,
		projectionRuntime.DescriptorSet(),
		projectionRuntime,
		projectionRuntime.EntityPorts(),
		projectionRuntime.AssessmentPorts().Rows,
		projectionRuntime.PartyPorts().Rows,
		indicatorOwner,
		timelineFacade,
		evidenceOwner.WorkbookContribution(),
		artifactMutation,
		taskDecisionMutation,
		workbookConflictTokens,
		revisionRuntime.ConflictFieldResolver(),
		revisionRuntime.Appender(),
		intentAppender,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Workbook contribution catalog: %w", err)
	}
	workbookMutationStore, err := workbookassembly.NewMutationStore(
		postgresHandle,
		workbookContributionCatalog,
		artifactMutation,
		taskDecisionMutation,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Workbook mutation store: %w", err)
	}
	builtInRoutes, err := newApplicationRouteCatalog(publicationCatalog).Bind([]routeContribution{
		{id: "auth", registrar: auth.RegisterRoutes(authRouteOptions...)},
		{id: "incidents", registrar: incidentRoutes},
		{id: "extensions", registrar: extensiondiscovery.RegisterRoutes()},
		{id: "jobs", registrar: jobapi.RegisterRoutes(jobManager)},
		{id: "saved_views", registrar: savedviews.RegisterRoutes()},
		{id: "view_schemas", registrar: viewschemas.RegisterRoutes()},
		{id: "collaboration", registrar: collaboration.RegisterRoutes(settingsProjection.Collaboration(hub, socketTransport))},
		{id: "entities", registrar: entities.RegisterRoutes(entities.RouteOptions{
			MergeStore:   timelineBundle.EntityMergeStore,
			MentionStore: timelineBundle.EntityMentionStore,
		})},
		{id: "evidence", registrar: evidence.RegisterRoutes(
			settingsProjection.Evidence(),
			evidence.WithRouteService(evidenceOwner.RouteService()),
		)},
		{
			id: "workbook",
			registrar: workbook.RegisterRoutes(workbook.RouteDependencies{
				TimelineOwner: timelineFacade,
				MutationStore: workbookMutationStore,
				EntityOwner: hostidentity.NewStore(
					postgresHandle,
					revisionRuntime.Appender(),
					workbookassembly.NewConflictIdempotencyPort(postgresHandle),
					projectionRuntime.EntityPorts().Writer,
					hostidentity.WithProjectionReader(projectionRuntime.EntityPorts().Reader),
				),
				ConflictTokens:      workbookConflictTokens,
				StartupStoreFactory: workbookassembly.NewStartupStoreFromDependencies,
			}),
		},
		{id: "timeline", registrar: timelineadmission.RegisterRoutes(timelineadmission.RouteOptions{
			Facade: timelineFacade,
		})},
		{id: "revisions", registrar: revisionRoutes},
		{id: "indicators", registrar: indicatorshttpapi.RegisterRoutes(indicatorOwner)},
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
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose built-in routes: %w", err)
	}
	httpOptions.AdditionalRoutes = append(builtInRoutes, httpOptions.AdditionalRoutes...)
	httpOptions.ValidatePublicRoutes = true
	readinessProbes := []httpapi.DependencyReadinessProbe{}
	if stagedHealth != nil {
		readinessProbes = append(readinessProbes, stagedCleanupReadinessProbe{health: stagedHealth})
	}
	publicationProjections := publication.httpProjections()
	httpOptions.Dependencies = httpapi.DependencySet{
		Telemetry:           httpapi.TelemetrySettings{Enabled: normalizedCfg.Telemetry.Enabled, ServiceVersion: normalizedCfg.Telemetry.Resource.ServiceVersion},
		Env:                 options.Env,
		Postgres:            postgresPool,
		PostgresDB:          postgresHandle,
		ObjectStore:         objectStore,
		CursorCodec:         cursorCodec,
		Readiness:           httpapi.NewDependencyReadinessChecker(postgresPool, objectStore, readinessProbes...),
		Admission:           runtime.lifecycle,
		PublicErrorFaults:   testRuntimeDeps.PublicErrorFaults,
		TestResetBootstrap:  settingsProjection.TestResetBootstrap(),
		ModuleOverrides:     moduleOverrides,
		ExtensionDiscovery:  publicationProjections,
		ExtensionClaims:     publicationProjections,
		ExtensionRoutes:     publicationProjections,
		ExtensionWorkspaces: publicationProjections,
		Now:                 now,
	}

	if err := publication.controller.commit(); err != nil {
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
	handler, err := dependencies.newHTTPHandler(httpOptions)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("setup http handler: %w", err)
	}

	runtime.handler = handler
	runtime.publicHTTP = httpOptions.Dependencies.PublicRoutes.Diagnostics()
	if err := publication.controller.acknowledge("websocket", listenerPlanSHA256, nil); err != nil {
		runtime.Close()
		return nil, err
	}
	jobActivationCtx, cancelJobActivation := context.WithTimeout(ctx, runtime.reconciliationTimeout)
	jobActivationErr := runtime.jobRunner.Activate(jobActivationCtx)
	cancelJobActivation()
	if jobActivationErr != nil {
		runtime.Close()
		return nil, fmt.Errorf("activate initial Jobs recovery: %w", jobActivationErr)
	}
	if err := publication.controller.acknowledge("job_dequeue", listenerPlanSHA256, nil); err != nil {
		runtime.Close()
		return nil, err
	}
	for _, worker := range extensionPlan.Workers() {
		if err := publication.controller.acknowledge(
			"worker:"+worker.ProfileID+":"+worker.WorkerKind,
			extensionPlan.Summary().WorkerPlanSHA256,
			nil,
		); err != nil {
			runtime.Close()
			return nil, err
		}
	}
	if err := publication.controller.acknowledge("http", listenerPlanSHA256, nil); err != nil {
		runtime.Close()
		return nil, err
	}
	if runtime.processLease != nil && runtime.processLease.State() == processlease.StateLost {
		runtime.Close()
		return nil, processlease.ErrApplicationProcessLeaseLost
	}
	if runtime.servingLease != nil && runtime.servingLease.State() == processlease.StateLost {
		runtime.Close()
		return nil, processlease.ErrRecoveryServingLeaseLost
	}
	if options.ObserveJobs != nil {
		options.ObserveJobs(jobManager, jobTransactions, runtime.jobRunner, postgresPool)
	}
	if options.ObserveCollaboration != nil {
		options.ObserveCollaboration(hub, runtime.collaborationDispatcher, intentAppender)
	}
	if options.ObserveEvidenceCleanup != nil {
		options.ObserveEvidenceCleanup(runtime.evidenceCleanupDispatcher)
	}
	if options.ObserveProjections != nil {
		options.ObserveProjections(projectionRuntime)
	}
	if options.ObserveTimeline != nil {
		options.ObserveTimeline(timelineBundle)
	}
	if options.ObserveRevisions != nil {
		options.ObserveRevisions(revisionRuntime)
	}
	return runtime, nil
}

func loadNetworkFlowKeyRings(
	configuration networkflow.Configuration,
	env map[string]string,
	now time.Time,
	registry *secretpurpose.Registry,
	readDocument secureDocumentReader,
) (*networkflow.KeyRings, error) {
	if !configuration.Claimed {
		return nil, nil
	}
	document, err := readDocument(configuration.KeyRingManifestPath, networkflow.KeyRingManifestMaximumSize)
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

func loadRevisionsConflictTokenKeyRing(
	configuration conflicttokens.Configuration,
	env map[string]string,
	now time.Time,
	registry *secretpurpose.Registry,
	readDocument secureDocumentReader,
) (*conflicttokens.ConflictTokenKeyRing, error) {
	document, err := readDocument(configuration.ConflictTokenKeyRingManifestPath, conflicttokens.ConflictTokenKeyRingManifestMaximumSize)
	if err != nil {
		failure := conflicttokens.ManifestUnreadable
		var secureError *securefile.Error
		if errors.As(err, &secureError) {
			switch secureError.Kind {
			case securefile.FailureTooLarge:
				failure = conflicttokens.ManifestTooLarge
			case securefile.FailureInvalidPath, securefile.FailureUnsafeObject, securefile.FailureChanged:
				failure = conflicttokens.ManifestUnsafe
			}
		}
		return nil, deploymentRevisionsConflictTokenError(conflicttokens.ConflictTokenKeyRingManifestReadError(failure))
	}
	options := conflicttokens.KeyRingParseOptions{
		AllowFixtureKey: runtimeEnvironmentValue(env, conflicttokens.ConflictTokenFixtureRuntimeEnvName) == conflicttokens.ConflictTokenFixtureRuntimeMarker,
	}
	ring, err := conflicttokens.ParseConflictTokenKeyRingWithRegistry(document.Bytes(), env, now, registry, options)
	if err != nil {
		return nil, deploymentRevisionsConflictTokenError(err)
	}
	return ring, nil
}

func runtimeEnvironmentValue(env map[string]string, key string) string {
	return env[key]
}

type enterpriseDocumentReader struct {
	readDocument secureDocumentReader
}

func loadEnterpriseProviderManifest(
	configuration enterpriseauth.Configuration,
	env map[string]string,
	readDocument secureDocumentReader,
) ([]authn.EnterpriseAuthProviderDefinition, error) {
	definitions, err := enterpriseauth.LoadProviderManifest(configuration, env, enterpriseDocumentReader{readDocument: readDocument})
	if err != nil {
		return nil, deploymentEnterpriseAuthenticationError(err)
	}
	return definitions, nil
}

func (reader enterpriseDocumentReader) ReadDocument(absolutePath string, maximumBytes int64) ([]byte, enterpriseauth.DocumentReadFailure) {
	document, err := reader.readDocument(absolutePath, maximumBytes)
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

func deploymentRevisionsConflictTokenError(err error) error {
	finding, ok := conflicttokens.ConfigurationFindingFromError(err)
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
	publication *publicationController
}

func (provider publicationHTTPProjections) ExtensionDiscoveryProfiles() []httpapi.ExtensionProfile {
	if provider.publication == nil {
		return nil
	}
	return publicationHTTPProfiles(provider.publication.discovery())
}

func (provider publicationHTTPProjections) ExtensionClaims() []httpapi.ExtensionClaim {
	if provider.publication == nil {
		return nil
	}
	publication := provider.publication.claims()
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
	publication := provider.publication.routes()
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
	publication := provider.publication.workspaces()
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

func (r *Runtime) HTTPHandler() http.Handler {
	if r == nil {
		return nil
	}
	return r.handler
}

func (r *Runtime) drainTimeout() time.Duration {
	if r == nil {
		return 0
	}
	return r.shutdownDrainTimeout
}

func (r *Runtime) publicHTTPDiagnostics() httpapi.RouteDiagnostics {
	if r == nil {
		return httpapi.RouteDiagnostics{}
	}
	diagnostics := r.publicHTTP
	diagnostics.ClaimedProfiles = append([]string(nil), diagnostics.ClaimedProfiles...)
	return diagnostics
}

func (r *Runtime) ActivatePublication() error {
	if r == nil || r.publication == nil {
		return fmt.Errorf("extension_publication_failed")
	}
	if r.processLease != nil && r.processLease.State() != processlease.StateHeld {
		return fmt.Errorf("extension_publication_failed")
	}
	if r.servingLease != nil && r.servingLease.State() != processlease.StateHeld {
		return fmt.Errorf("extension_publication_failed")
	}
	if err := r.publication.serve(); err != nil {
		return err
	}
	if r.collaborationDispatcher != nil {
		if err := r.collaborationDispatcher.Start(context.Background()); err != nil {
			r.publication.componentLost("collaboration_dispatcher")
			return fmt.Errorf("activate collaboration dispatcher: %w", err)
		}
	}
	if r.evidenceCleanupDispatcher != nil {
		if err := r.evidenceCleanupDispatcher.Start(context.Background()); err != nil {
			r.publication.componentLost("evidence_cleanup_dispatcher")
			return fmt.Errorf("activate Evidence cleanup dispatcher: %w", err)
		}
	}
	r.publicationOnce.Do(func() {
		if r.stagedJanitor != nil && r.stagedJanitorContext != nil {
			go func() {
				defer func() {
					if recovered := recover(); recovered != nil && r.lifecycle != nil {
						r.publication.componentLost("staged_object_janitor")
					}
				}()
				if err := r.stagedJanitor.Run(r.stagedJanitorContext, r.stagedObjectSweepPeriod); err != nil && r.lifecycle != nil {
					r.publication.componentLost("staged_object_janitor")
				}
			}()
		}
	})
	return nil
}

func (r *Runtime) publishedComponentLost(componentID string) bool {
	if r == nil || r.publication == nil {
		return false
	}
	return r.publication.componentLost(componentID)
}

func (r *Runtime) fatalEvents() <-chan processlifecycle.FatalSignal {
	if r == nil || r.lifecycle == nil {
		return nil
	}
	return r.lifecycle.FatalEvents()
}

func (r *Runtime) watchProcessLease(ctx context.Context) {
	if r == nil || r.processLease == nil || r.lifecycle == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-r.processLease.Events():
			switch event.State {
			case processlease.StateUncertain:
				r.lifecycle.CloseAdmission()
			case processlease.StateHeld:
				if event.Previous == processlease.StateUncertain && r.servingLeasesHeld() {
					r.lifecycle.RestoreAdmission()
				}
			case processlease.StateLost:
				r.lifecycle.Fatal(processlease.ErrApplicationProcessLeaseLost.FatalReasonCode())
				return
			}
		}
	}
}

func (r *Runtime) watchServingLease(ctx context.Context) {
	if r == nil || r.servingLease == nil || r.lifecycle == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-r.servingLease.Events():
			switch event.State {
			case processlease.StateUncertain:
				r.lifecycle.CloseAdmission()
			case processlease.StateHeld:
				if event.Previous == processlease.StateUncertain && r.servingLeasesHeld() {
					r.lifecycle.RestoreAdmission()
				}
			case processlease.StateLost:
				r.lifecycle.Fatal(processlease.ErrRecoveryServingLeaseLost.FatalReasonCode())
				return
			}
		}
	}
}

func (r *Runtime) servingLeasesHeld() bool {
	if r == nil || r.servingLease == nil || r.servingLease.State() != processlease.StateHeld {
		return false
	}
	return r.processLease == nil || r.processLease.State() == processlease.StateHeld
}

func (r *Runtime) Close() {
	if r == nil {
		return
	}
	r.closeOnce.Do(func() {
		if r.publication != nil {
			r.publication.abortStartup()
		}
		if r.lifecycle != nil {
			r.lifecycle.MarkTerminating()
		}
		for index := len(r.cleanups) - 1; index >= 0; index-- {
			r.cleanups[index]()
		}
		if r.lifecycle != nil {
			r.lifecycle.MarkExited()
		}
	})
}

func (r *Runtime) own(cleanup func()) {
	if r != nil && cleanup != nil {
		r.cleanups = append(r.cleanups, cleanup)
	}
}
