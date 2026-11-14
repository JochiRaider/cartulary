package server

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/app/collaborationassembly/incidenteffects"
	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/importassembly"
	"github.com/JochiRaider/cartulary/internal/app/incidentportabilityassembly"
	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/app/jobassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/app/reportingassembly"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/entities/candidates"
	hostidentityreporting "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/reportingprovider"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	incidentshttpapi "github.com/JochiRaider/cartulary/internal/modules/incidents/httpapi"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatorshttpapi "github.com/JochiRaider/cartulary/internal/modules/indicators/httpapi"
	"github.com/JochiRaider/cartulary/internal/modules/jobapi"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	revisionshttpapi "github.com/JochiRaider/cartulary/internal/modules/revisions/httpapi"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/sourceboundary"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	timelineadmission "github.com/JochiRaider/cartulary/internal/modules/timeline/admission"
	"github.com/JochiRaider/cartulary/internal/modules/viewschemas"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
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
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
	"github.com/JochiRaider/cartulary/internal/platform/secretpurpose"
	"github.com/JochiRaider/cartulary/internal/platform/securefile"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

// networkFlowComposition decorates owner dependencies before any worker starts.
// Ordinary composition leaves this nil; no control registry is a runtime dependency.
type networkFlowComposition interface {
	WrapDatabase(dbport.DB) dbport.DB
	ConfigureNetworkFlow(networkflow.ModuleDependencies) networkflow.ModuleDependencies
	WrapImportFacade(imports.ExtensionImportFacade) imports.ExtensionImportFacade
}

// referenceDataComposition decorates this runtime's owner database before construction.
type referenceDataComposition interface{ WrapDatabase(dbport.DB) dbport.DB }

type Options struct {
	immediateFatal            func(processlifecycle.FatalSignal)
	BrowserBuild              *webassets.Build
	ReportingRuntime          reporting.RenderRuntime
	ReportingTemplatePacks    *reporting.TemplatePackRegistry
	ReferenceDataComposition  referenceDataComposition
	NetworkFlowComposition    networkFlowComposition
	Env                       map[string]string
	HTTP                      httpapi.Options
	Postgres                  postgres.AdmittedPool
	ObjectStore               objectstore.Store
	Now                       func() time.Time
	ObserveJobs               func(*jobs.Manager, *jobs.TransactionService, *jobs.Runner, *pgxpool.Pool)
	ObserveCollaboration      func(*collaboration.Runtime)
	ObserveEvidenceCleanup    func(*evidence.CleanupDispatcher)
	ObserveNetworkFlowCleanup func(*networkflow.GraphResultCleanupDispatcher)
	ObserveProjections        func(*projectionassembly.Runtime)
	ObserveTimeline           func(*timelineassembly.Bundle)
	ObserveRevisions          func(*revisionassembly.Runtime)
}

type evidenceCleanupLifecycle interface {
	Prepare(context.Context, <-chan struct{}) error
	Close(context.Context) error
}

type Runtime struct {
	publicationOperation         *publicationOperation
	immediateFatal               func(processlifecycle.FatalSignal)
	handler                      http.Handler
	stagedJanitor                *stagedobjects.Janitor
	referenceCollector           *reference_data.ObjectCollector
	jobRunner                    *jobs.Runner
	collaborationRuntime         *collaboration.Runtime
	evidenceCleanupDispatcher    evidenceCleanupLifecycle
	networkFlowCleanupDispatcher evidenceCleanupLifecycle
	graphActivation              *claimedGraphActivation
	processLease                 *processlease.ApplicationProcessLease
	servingLease                 *processlease.ApplicationRecoveryServingLease
	lifecycle                    *processlifecycle.Controller
	publication                  *publicationController
	publicHTTP                   httpapi.RouteDiagnostics
	shutdownDrainTimeout         time.Duration
	reconciliationTimeout        time.Duration
	stagedObjectSweepPeriod      time.Duration

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

func (assembly runtimeAssembly) build(ctx context.Context) (_ *Runtime, resultErr error) {
	loadedConfiguration := assembly.loadedConfiguration
	options := assembly.options
	dependencies := assembly.dependencies
	browserBuild := options.BrowserBuild
	if browserBuild == nil {
		var err error
		browserBuild, err = webassets.LoadEmbedded()
		if err != nil {
			return nil, fmt.Errorf("admit browser build: %w", err)
		}
	}
	extensionCoordinator, err := extensionassembly.BrowserCoordinator(browserBuild)
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
		immediateFatal:          options.immediateFatal,
		shutdownDrainTimeout:    time.Duration(runtimeSettings.ShutdownDrainSeconds) * time.Second,
		reconciliationTimeout:   time.Duration(runtimeSettings.ReconciliationSeconds) * time.Second,
		stagedObjectSweepPeriod: time.Duration(runtimeSettings.StagedObjectSweepSeconds) * time.Second,
	}
	referenceLifetime, err := reference_data.NewVerificationLifetime(time.Duration(normalizedCfg.Timeouts.Extensions.CancellationGraceSeconds)*time.Second, runtime.fatalIntegrity)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resultErr != nil && runtime.lifecycle.FatalReason() != "" {
			resultErr = publicationStartupFatal{reason: runtime.lifecycle.FatalReason()}
		}
	}()

	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	// Configuration material is admitted before any service or lease acquisition.
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
	if enterpriseAuthenticationConfiguration.Claimed {
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
	keys, err := authn.LoadMasterKeys(options.Env)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	var objectStoreSettings objectstore.Settings
	if options.ObjectStore == nil {
		objectStoreSettings, err = objectstore.ResolveSettings(configassembly.ObjectStoreBinding(normalizedCfg), options.Env)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup object store: %w", err)
		}
	}
	postgresAdmission := options.Postgres
	if postgresAdmission == nil {
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
		postgresAdmission = pool
		if pool != nil {
			runtime.own(pool.Close)
		}
	}
	var postgresPool *pgxpool.Pool
	if postgresAdmission != nil {
		postgresPool = postgresAdmission.Pool()
	}
	if dependencies.requireCryptoState == nil {
		runtime.Close()
		return nil, database_migrations.ErrIncompatibleCryptoState
	}
	if err := dependencies.requireCryptoState(ctx, postgresPool); err != nil {
		runtime.Close()
		return nil, err
	}

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
	if postgresPool != nil {
		if err := referenceassembly.ValidateRetainedRuntimeState(ctx, postgresPool); err != nil {
			runtime.Close()
			return nil, err
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
		if errors.Is(leaseErr, processlease.ErrApplicationProcessActive) {
			return nil, extensionCoordinator.ApplicationLeaseAdmissionFailure(normalizedCfg.Timeouts.Extensions.ProcessLeaseAcquireSeconds, leaseErr)
		}
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
		if errors.Is(servingLeaseErr, processlease.ErrRecoveryServingLeaseActive) {
			return nil, extensionCoordinator.RecoveryLeaseAdmissionFailure(servingLeaseErr)
		}
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

	control, controlErr := dependencies.readRecoveryTargetControl(ctx, postgresPool)
	var targetBindings recovery.TargetBindingDigests
	if controlErr == nil && control.State == recovery.TargetComplete && *control.Purpose == "restore_target" {
		targetBindings, controlErr = configassembly.RecoveryTargetBindings(normalizedCfg, options.Env)
	}
	if controlErr == nil {
		controlErr = control.RequireServing(targetBindings)
	}
	if controlErr != nil {
		runtime.Close()
		return nil, extensionCoordinator.RecoveryTargetAdmissionFailure(controlErr)
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
	var referencePackStorage reference_data.ArtifactStorage = rootReferencePackStorage
	runtime.own(rootReferencePackStorage.Close)
	if err := rootReferencePackStorage.ReconcileWorkspaces(ctx); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("reconcile Reference Pack scratch: %w", err)
	}

	// Both application-owned leases are held before claim resolution and every
	// Stage-1-or-later effect. Their separate typed identities prevent the
	// Recovery serving vocabulary from collapsing into the process-lease one.

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
	incidentBundleRouteAdmitted, err := publicationCatalog.ExactProfileContributionSet(
		incidentbundles.ProfileID,
		"http_route_family",
		[]string{incidentbundles.BundlesRouteContributionID},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project Incident Bundles route application plan: %w", err)
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
	var referencePackBootstrap reference_data.TrustBootstrap
	if referencePackRouteAdmitted {
		referencePackBootstrap, err = referenceassembly.ReadTrustBootstrap(normalizedCfg.ReferencePacks)
		if err != nil {
			runtime.Close()
			return nil, err
		}
	}
	importWorkersAdmitted, err := publicationCatalog.ExactProfileWorkerJobSet(
		imports.ProfileID,
		[]string{imports.ApplyJobKind, imports.DiscoveryJobKind},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project Imports worker application plan: %w", err)
	}
	snapshotReportingRoutesAdmitted, err := publicationCatalog.ExactProfileContributionSet(
		reporting.ProfileID,
		"http_route_family",
		[]string{
			reporting.ReleasesRouteContributionID,
			reporting.ReportCompositionsRouteContributionID,
			reporting.RetainedArtifactsRouteContributionID,
			reporting.SnapshotsRouteContributionID,
		},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("project Snapshot/Reporting route application plan: %w", err)
	}
	var reportRenderer *reporting.Renderer
	if snapshotReportingRoutesAdmitted {
		if options.ReportingRuntime != nil {
			reportRenderer, err = reporting.NewRenderer(options.ReportingRuntime)
		} else {
			reportRenderer, err = reporting.NewInstalledRenderer(filepath.Dir(normalizedCfg.Reporting.RuntimeReceiptPath), normalizedCfg.Roots.TemporaryWork.Path)
		}
		if err != nil {
			runtime.Close()
			return nil, config.NewDiagnosticsError(config.Diagnostic{Path: "snapshot_reporting.runtime_receipt_path", ReasonCode: "reporting_runtime_unavailable", Message: "Reporting runtime installation admission failed."})
		}
	}
	publication := &publicationOrchestrator{controller: newPublicationController(runtime.lifecycle)}
	runtime.publication = publication.controller
	resolvedClaims := extensionPlan.ResolvedClaims()

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

	recognizedJobDefinitions, err := jobassembly.RecognizedDefinitions(extensionCoordinator.JobKindContracts(), extensionCoordinator.WorkerRuntimeContracts())
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose recognized extension job definitions: %w", err)
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
			Store:       logicalStateStore,
			Coordinator: extensionCoordinator,
			FinalValidators: map[string]extensions.FinalStateValidator{
				"network_flow_activity.validate_state_v7": func(ctx context.Context, input extensions.FinalStateValidationContext, reader extensions.StateReadCapability) (extensions.StateValidationResult, error) {
					if err := networkflow.ValidateExtensionState(ctx, reader); err != nil {
						return extensions.StateValidationResult{
							SchemaID: "cartulary.extension_final_state_validation_result.v1",
							Status:   "invalid",
							Findings: []extensions.StartupFinding{extensionCoordinator.StateValidationFailure(input.ProfileID, "state_validation", "network_flow_activity.validate_state_v7")},
						}, nil
					}
					return extensions.ValidFinalStateValidationResult(), nil
				},
			},
			Now:                now,
			LockTimeout:        time.Duration(normalizedCfg.Timeouts.Extensions.MigrationLockSeconds) * time.Second,
			StepTimeout:        time.Duration(normalizedCfg.Timeouts.Extensions.MigrationStepSeconds) * time.Second,
			ProfileTimeout:     time.Duration(normalizedCfg.Timeouts.Extensions.ProfileMigrationSeconds) * time.Second,
			ValidationTimeout:  time.Duration(normalizedCfg.Timeouts.Extensions.ValidationSeconds) * time.Second,
			CancellationGrace:  time.Duration(normalizedCfg.Timeouts.Extensions.CancellationGraceSeconds) * time.Second,
			FatalIntegritySink: runtime.fatalIntegrity,
		})
		if stateRuntimeErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose extension state runtime: %w", stateRuntimeErr)
		}
		admissionViews, admissionAlgorithms, admissionErr := networkFlowAdmissionPorts(extensionCoordinator, stateStore, networkFlowConfiguration, recognizedJobDefinitions)
		if admissionErr != nil {
			runtime.Close()
			return nil, admissionErr
		}
		if referencePackRouteAdmitted {
			values := []extensions.ProfileConfigurationValue{}
			for _, value := range configassembly.ReferencePackAdmissionConfiguration(normalizedCfg.ReferencePacks) {
				values = append(values, extensions.ProfileConfigurationValue{Key: value.Key, Source: value.Source, Value: value.Value})
			}
			view, viewErr := extensionCoordinator.ConfigurationView(reference_data.ProfileID, values)
			if viewErr != nil {
				runtime.Close()
				return nil, viewErr
			}
			admissionViews[reference_data.ProfileID] = view
		}
		if snapshotReportingRoutesAdmitted {
			view, viewErr := extensionCoordinator.ConfigurationView(reporting.ProfileID, []extensions.ProfileConfigurationValue{{Key: "snapshot_reporting.runtime_receipt_path", Source: "explicit", Value: map[string]any{"kind": "regular_file_ref", "key": "snapshot_reporting.runtime_receipt_path"}}})
			if viewErr != nil {
				runtime.Close()
				return nil, viewErr
			}
			admissionViews[reporting.ProfileID] = view
		}
		if admissionErr := stateRuntime.ValidateClaimAdmission(ctx, extensionCoordinator, claimResolution, "preflight", admissionViews, admissionAlgorithms); admissionErr != nil {
			runtime.Close()
			return nil, admissionErr
		}
		if stateAdmissionErr := stateRuntime.AdmitClaims(ctx, extensionCoordinator, claimResolution); stateAdmissionErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("admit extension state: %w", stateAdmissionErr)
		}
		if admissionErr := stateRuntime.ValidateClaimAdmission(ctx, extensionCoordinator, claimResolution, "post_migration", admissionViews, admissionAlgorithms); admissionErr != nil {
			runtime.Close()
			return nil, admissionErr
		}
		extensionStateStore = stateStore
	}
	var rawObjectStore objectstore.Store
	if options.ObjectStore != nil {
		rawObjectStore = options.ObjectStore
	} else {
		client, err := dependencies.setupObjectStore(ctx, objectStoreSettings, objectstore.Instrumentation{})
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("setup object store: %w", err)
		}
		rawObjectStore = client
		if client != nil {
			runtime.own(func() { _ = client.Close() })
		}
	}
	objectStore := instrumentedObjectStore(
		normalizedCfg.Telemetry.Enabled,
		normalizedCfg.Telemetry.Resource.ServiceVersion,
		rawObjectStore,
	)
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
			Repository:       stagedRepository,
			ResourceIdentity: evidence.IncidentBundleResourceIdentity,
			Bytes:            stagedBytes,
			Now:              now,
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
		postgresAdmission,
	)
	if options.NetworkFlowComposition != nil {
		postgresHandle = options.NetworkFlowComposition.WrapDatabase(postgresHandle)
	}
	referenceDatabase := postgresHandle
	if options.ReferenceDataComposition != nil {
		referenceDatabase = options.ReferenceDataComposition.WrapDatabase(referenceDatabase)
		if referenceDatabase == nil {
			runtime.Close()
			return nil, errors.New("reference pack: database decorator returned nil")
		}
	}
	if err := dependencies.runBootstrap(ctx, configassembly.BootstrapSettings(normalizedCfg), postgresPool); err != nil {
		runtime.Close()
		return nil, err
	}
	if _, ok := rawObjectStore.(objectstore.TypedStore); !ok {
		runtime.Close()
		return nil, fmt.Errorf("compose Evidence owner: typed object store is required")
	}
	typedObjectStore, ok := objectStore.(objectstore.TypedStore)
	if !ok {
		runtime.Close()
		return nil, fmt.Errorf("compose Evidence owner: typed object store is required")
	}
	evidenceBlobPort, err := evidence.NewIncidentBundleBlobPortability(typedObjectStore, stagedObjectService)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	if enterpriseAuthenticationAdmitted {
		if err := enterpriseauth.ReconcileProviderDefinitions(ctx, enterpriseProviderDefinitions, authn.NewStore(postgresHandle), now()); err != nil {
			runtime.Close()
			return nil, deploymentEnterpriseAuthenticationError(err)
		}
	}
	if referencePackRouteAdmitted {
		if err := reference_data.ReconcileTrustBootstrap(ctx, referenceDatabase, referencePackBootstrap, now()); err != nil {
			runtime.Close()
			return nil, fmt.Errorf("reconcile Reference Pack trust bootstrap: %w", err)
		}
	}
	referenceMutationFinalizer, err := extensionassembly.NewReferencePackMutationFinalizer(runtime.fatalIntegrity)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	referenceObserver := referenceassembly.NewOperationObserver(normalizedCfg.Telemetry.Resource.ServiceVersion)
	if err := reference_data.ReconcileBaseRelease(ctx, referenceDatabase, rootReferencePackStorage, reference_data.BaseReleaseOptions{Finalizer: referenceMutationFinalizer, Observer: referenceObserver, ProfileClaimed: referencePackRouteAdmitted, ClockTrusted: normalizedCfg.ReferencePacks.ClockTrusted, Limits: settingsProjection.ReferenceData()}, now()); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("reconcile Base reference registries: %w", err)
	}
	if err := reference_data.ValidateRequiredState(ctx, postgresHandle, rootReferencePackStorage, settingsProjection.ReferenceData(), referenceLifetime); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("validate retained Reference Pack state: %w", err)
	}
	cursorKey, err := authn.DerivePurposeKey(keys, "pagination-cursor-v2")
	if err != nil {
		runtime.Close()
		return nil, err
	}
	cursorCodec := pagination.NewCodec(cursorKey)
	referenceIntegrity := reference_data.IntegrityOptions{Observer: referenceObserver, Finalizer: referenceMutationFinalizer, Limits: settingsProjection.ReferenceData()}
	referenceConsumer, err := reference_data.NewConsumer(referenceDatabase, rootReferencePackStorage, cursorCodec, now, referenceIntegrity)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	referenceAssignments, err := reference_data.NewRegistryAssignments(referenceDatabase, rootReferencePackStorage, cursorCodec, now, referenceIntegrity)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	providerContributions, err := revisionassembly.CurrentProviderContributions(referenceAssignments)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Revisions provider contributions: %w", err)
	}
	projectionRuntime, err := projectionassembly.Build(postgresHandle)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose projections runtime: %w", err)
	}
	collaborationPublicationCatalog, err := buildCollaborationPublicationCatalog(
		providerContributions,
		projectionRuntime.DescriptorSet(),
	)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	collaborationOptions := settingsProjection.Collaboration(postgresHandle, socketTransport, now)
	collaborationOptions.PublicationCatalog = collaborationPublicationCatalog
	for _, workspace := range extensionPlan.Workspaces() {
		collaborationOptions.ClaimedWorkspaces = append(collaborationOptions.ClaimedWorkspaces, collaboration.ClaimedWorkspace{
			ExtensionProfileID: workspace.ProfileID, WorkspaceKey: workspace.WorkspaceKey,
		})
	}
	collaborationOptions.OnUnexpectedDispatcherLoss = func() {
		runtime.publication.componentLost("websocket")
	}
	collaborationRuntime, err := dependencies.newCollaborationRuntime(collaborationOptions)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Collaboration runtime: %w", err)
	}
	runtime.collaborationRuntime = collaborationRuntime
	runtime.own(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = collaborationRuntime.Close(closeCtx)
	})
	recordChanges := collaborationRuntime.RecordChanges()
	jobProgressIntents := newCollaborationJobProgressTranslator(collaborationRuntime.JobProgress())
	extensionResourceChangeIntents := newCollaborationExtensionResourceChangeTranslator(collaborationRuntime.ExtensionResourceChanges())
	jobOwnerPorts := extensionassembly.JobOwnerTransactionAdapters{}
	extensionJobDefinitions, err := jobassembly.SelectedDefinitions(publicationCatalog)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose extension job definitions: %w", err)
	}
	jobCatalog, err := jobs.NewCatalog(recognizedJobDefinitions)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Jobs catalog: %w", err)
	}
	workerRuntimeContracts, err := jobassembly.WorkerContracts(publicationCatalog)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Jobs worker runtime contracts: %w", err)
	}
	jobSelection, err := jobs.NewRuntimeSelection(jobCatalog, extensionassembly.JobKinds(extensionJobDefinitions), workerRuntimeContracts)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Jobs runtime selection: %w", err)
	}
	jobTransactions, err := jobs.NewTransactionService(jobProgressIntents, jobs.OwnerTransactionPorts{
		RouteIdempotency:      jobOwnerPorts,
		ExtensionCancellation: jobOwnerPorts,
		TerminalEffects:       reference_data.JobTerminalEffects{},
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
		ctx, cancel := context.WithTimeout(context.Background(), runtime.shutdownDrainTimeout)
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
		reconciliationErr := extensions.ReconcileInactiveExtensionJobs(ctx, inactiveJobStore,
			inactiveExtensionProfileIDs(extensionPlan.Claims()), extensionCoordinator.JobKindContracts(),
			extensions.ReconciliationOptions{Coordinator: extensionCoordinator,
				MaxNonterminalJobs: int(normalizedCfg.Limits.Extensions.MaxNonterminalJobsPerProfile),
				Timeout:            time.Duration(normalizedCfg.Timeouts.Extensions.ReconciliationSeconds) * time.Second,
				FatalIntegritySink: runtime.fatalIntegrity,
			})
		if reconciliationErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("reconcile inactive extension jobs: %w", reconciliationErr)
		}
	}
	if err := reference_data.ReconcileTerminalJobs(ctx, referenceDatabase, jobTransactions, referenceMutationFinalizer); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("reconcile terminal Reference Pack jobs: %w", err)
	}
	collector, err := reference_data.NewObjectCollector(referenceDatabase, rootReferencePackStorage, rootReferencePackStorage, referenceObserver)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	// Startup reconciles bytes after terminal Jobs are classified. Periodic
	// maintenance consumes abandoned scratch, removal tombstones and publications.
	cleanupCtx, cancelCollection := context.WithTimeout(ctx, 30*time.Second)
	cleanupErr := collector.Sweep(cleanupCtx)
	cancelCollection()
	if cleanupErr != nil {
		runtime.Close()
		return nil, fmt.Errorf("reconcile Reference Pack objects: %w", cleanupErr)
	}
	runtime.referenceCollector = collector
	runtime.own(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = collector.Close(closeCtx)
	})
	var extensionJobFinalizer *extensionstore.OwnerFinalizer
	var networkFlowJobFinalizer *extensionstore.OwnerFinalizer
	var reportingJobFinalizer *extensionstore.OwnerFinalizer
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
		networkFlowJobFinalizer, err = extensionstore.NewOwnerFinalizer(
			extensionStateStore, jobTransactions, networkflow.GraphViewReceiptReconciler{}, now,
			runtime.fatalIntegrity,
		)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose saved graph receipt finalizer: %w", err)
		}
		reportingJobFinalizer, err = extensionstore.NewOwnerFinalizer(
			extensionStateStore, jobTransactions, extensionassembly.ReportingReceiptReconciler{}, now,
			runtime.fatalIntegrity,
		)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose Reporting receipt finalizer: %w", err)
		}

	}
	listenerPlanSHA256 := extensionPlan.Summary().ListenerPlanSHA256

	httpOptions := options.HTTP
	testRuntimeDeps := httpOptions.Dependencies
	if testRings, ok := testRuntimeDeps.ModuleOverrides[networkflow.KeyRingsOverrideKey].(*networkflow.KeyRings); ok && testRings != nil {
		networkFlowKeyRings = testRings
	}
	workbookPreferenceBootstrap := workbookstartuppostgres.NewWriter()
	incidentApplication, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            postgresHandle,
		PreferenceBootstrap: workbookPreferenceBootstrap,
		Now:                 now,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Incidents application: %w", err)
	}
	incidentEffects, err := incidenteffects.New(
		incidentApplication,
		collaborationRuntime,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose incident terminal effects: %w", err)
	}
	incidentAdmission := admission.NewChecker(postgresHandle)
	incidentRoutes := incidentshttpapi.RegisterRoutes(incidentshttpapi.Dependencies{
		Application:                 incidentApplication,
		AdmissionChecker:            incidentAdmission,
		TerminalMutationCoordinator: incidentEffects,
	})
	incidentBundleImportFinalizer, err := incidents.NewIncidentBundleImportFinalizer(workbookPreferenceBootstrap)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose incident bundle import finalizer: %w", err)
	}
	revisionRuntime, err := revisionassembly.Build(providerContributions...)
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
	cleanupObserver, err := newEvidenceCleanupTelemetryObserver(normalizedCfg.Telemetry.Resource.ServiceVersion)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Evidence cleanup telemetry: %w", err)
	}
	evidenceMutationDependencies, err := workbookassembly.NewEvidenceMutationDependencies(
		postgresHandle,
		revisionRuntime.Appender(),
		revisionRuntime.ConflictFieldResolver(),
		workbookassembly.NewConflictIdempotencyPort(postgresHandle),
		projectionRuntime.EvidenceMutationRows(),
		projectionRuntime.EvidenceAssociationEffects(),
		recordChanges,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Evidence mutation dependencies: %w", err)
	}
	evidenceOwner, err := evidence.NewOwnerRuntime(evidence.OwnerRuntimeDependencies{
		JobTransactions: jobTransactions, JobManager: jobManager, JobRunner: jobRunner,
		Postgres:        postgresHandle,
		ConflictTokens:  &workbookConflictTokens,
		Revisions:       revisionRuntime.Appender(),
		Collaboration:   recordChanges,
		ObjectStore:     typedObjectStore,
		Mutation:        evidenceMutationDependencies,
		CleanupObserver: cleanupObserver,
		Now:             now,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Evidence owner: %w", err)
	}
	cleanupDispatcher := evidenceOwner.CleanupDispatcher()
	runtime.evidenceCleanupDispatcher = cleanupDispatcher
	runtime.own(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		_ = cleanupDispatcher.Close(closeCtx)
	})
	timelineBundle, err := timelineassembly.NewBundle(timelineassembly.Dependencies{
		Postgres:            postgresHandle,
		ConflictTokens:      workbookConflictTokens,
		ConflictFields:      revisionRuntime.ConflictFieldResolver(),
		Revisions:           revisionRuntime.Appender(),
		Collaboration:       recordChanges,
		EvidenceAttachments: evidenceOwner.TimelineAttachmentContribution(),
		TimelineProjection:  projectionRuntime.TimelinePorts().Writer,
		EntityProjection:    projectionRuntime.EntityMutationRows(),
		AssessmentRows:      projectionRuntime.AssessmentPorts().Rows,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose timeline bundle: %w", err)
	}
	indicatorRecords := records.NewStore(postgresHandle)
	indicatorAuth := authn.NewStore(postgresHandle)
	indicatorOwner, err := indicators.NewApplication(indicators.ApplicationDependencies{
		ReferencePacks:  referenceAssignments,
		Postgres:        postgresHandle,
		Idempotency:     indicatorassembly.NewIdempotencyPort(indicatorAuth),
		IncidentState:   incidentAdmission,
		Revisions:       revisionRuntime.Appender(),
		RecordEnvelopes: indicatorRecords,
		Projections:     projectionRuntime.IndicatorPorts().Rows,
		SourceText:      indicatorassembly.NewSourceTextPort(projectionRuntime.SourceTextRows()),
		Collaboration:   recordChanges,
		Clock:           now,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Indicators owner: %w", err)
	}
	registryTransactions, err := referenceassembly.NewRegistryTransactions(postgresHandle, referenceAssignments)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose registry mutation transactions: %w", err)
	}
	revisionCommands, err := revisionRuntime.NewCommandService(
		postgresHandle,
		registryTransactions,
		incidentbundles.ImportedAttributionResolver(),
		projectionRuntime.RevisionRows(),
		projectionRuntime.RevisionLiveRecords(),
		now,
		recordChanges,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose revisions command service: %w", err)
	}
	revisionRoutes := revisionshttpapi.RegisterRoutes(
		revisionCommands,
		revisionassembly.NewRecordEnvelopeReader(postgresHandle),
	)
	importSourcePort, err := imports.NewExtensionSourcePort(postgresHandle)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Imports extension source port: %w", err)
	}
	timelineFacade := timelineBundle.Facade
	var networkFlowTelemetry networkflow.GraphTelemetryObserver
	if networkFlowRouteAdmitted {
		networkFlowTelemetry = newNetworkFlowTelemetryObserver(normalizedCfg.Telemetry.Resource.ServiceVersion)
	}
	incidentTransactionParticipant := incidents.NewTransactionParticipant()
	graphActivation, err := newClaimedGraphActivation(postgresHandle, control, networkFlowRouteAdmitted, now, func() {
		runtime.lifecycle.Fatal("graph_reconstruction_indeterminate")
	})
	if err != nil {
		runtime.Close()
		return nil, err
	}
	runtime.graphActivation = graphActivation
	runtime.own(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		_ = graphActivation.Close(closeCtx)
	})
	var networkFlowModule *networkflow.Module
	if networkFlowRouteAdmitted {
		networkFlowDependencies := networkflow.ModuleDependencies{
			Postgres:             postgresHandle,
			ImportSources:        importSourcePort,
			ImportedAttributions: incidentbundles.PortableSourceAttributions(),
			KeyRings:             networkFlowKeyRings,
			EffectiveLimits:      networkFlowConfiguration.EffectiveResourceLimits(),
			Now:                  now,
			IncidentLocks:        incidentTransactionParticipant,
			AuditAppender:        authn.NewAdministrativeAuditAppender(),
			Indicators:           indicatorOwner,
			ResourceIntents:      extensionResourceChangeIntents,
			GraphViewJobs:        jobTransactions,
			JobManager:           jobManager,
			JobRunner:            runtime.jobRunner,
			JobFinalizer:         extensionassembly.NewNetworkFlowGraphViewJobFinalizer(networkFlowJobFinalizer),
			GraphTelemetry:       networkFlowTelemetry,
			GraphStateReady:      graphActivation.Ready,
		}
		if options.NetworkFlowComposition != nil {
			networkFlowDependencies = options.NetworkFlowComposition.ConfigureNetworkFlow(networkFlowDependencies)
		}
		networkFlowModule, err = networkflow.NewModule(networkFlowDependencies)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose Network Flow module: %w", err)
		}
	}

	networkFlowPortabilityState := networkflow.NewPortabilityStateBinding()
	portabilityPresence, err := extensionassembly.NewIncidentPortabilityStatePresence(networkFlowPortabilityState)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Incident Portability state presence: %w", err)
	}
	var portabilityParticipants []incidentbundles.ExtensionParticipant
	if networkFlowModule != nil {
		portabilityParticipants = append(portabilityParticipants, networkFlowModule.IncidentPortability())
	}
	portability, err := incidentbundles.NewPortabilityOrchestrator(incidentbundles.PortabilityOptions{
		Policies:     extensionassembly.IncidentPortabilityPolicies(extensionCoordinator.PortabilityPolicies(), resolvedClaims),
		Participants: portabilityParticipants, SnapshotReads: extensionassembly.IncidentPortabilityReads{}, ImportDependencies: extensionassembly.IncidentPortabilityReads{},
		Presence: portabilityPresence, Allocator: stagedObjectService, Limits: settingsProjection.IncidentBundles(),
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Incident Portability: %w", err)
	}
	incidentSourceCatalog, err := incidentportabilityassembly.NewCatalog(
		revisionRuntime.IncidentBundleSourcePort(), referenceConsumer, referenceAssignments,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Incident Portability source catalog: %w", err)
	}
	incidentReferences, err := reference_data.NewIncidentReferences(reference_data.IncidentReferenceOptions{Lifetime: referenceLifetime, Postgres: referenceDatabase, Storage: rootReferencePackStorage, Configuration: normalizedCfg.ReferencePacks, Limits: settingsProjection.ReferenceData(), JobExecutions: jobTransactions, JobOperations: jobManager, Observer: referenceObserver, Now: now})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Incident Portability reference retention: %w", err)
	}
	referenceRetention, err := reference_data.NewRetention(referenceDatabase, rootReferencePackStorage, now, referenceIntegrity)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Reference Pack retention: %w", err)
	}
	if err := reporting.ValidateRetainedSnapshotModels(ctx, jobTransactions, postgresPool, referenceRetention); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("validate retained reporting snapshots: %w", err)
	}
	if err := reporting.ValidateRetainedJobInventory(ctx, jobTransactions, postgresPool); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("validate retained Reporting job inventory: %w", err)
	}
	if err := reporting.ValidateRetainedReleases(ctx, jobTransactions, postgresPool, referenceRetention); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("validate retained reporting releases: %w", err)
	}
	if err := reporting.ValidateRetainedPreviews(ctx, jobTransactions, postgresPool, referenceRetention); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("validate retained Reporting previews: %w", err)
	}

	if err := reporting.ValidateRetainedSensitiveArtifacts(ctx, postgresPool); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("validate retained reporting sensitive artifacts: %w", err)
	}
	if err := reporting.ValidateRetainedArtifacts(ctx, postgresPool, incidentReferences); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("validate retained reporting artifacts: %w", err)
	}
	incidentBundleModule, err := incidentbundles.NewModule(incidentbundles.ModuleDependencies{
		Staging:                 stagedObjectService,
		Postgres:                postgresPool,
		JobTransactions:         jobTransactions,
		JobOperations:           jobManager,
		JobRunner:               runtime.jobRunner,
		Storage:                 incidentBundleStorage,
		Limits:                  settingsProjection.IncidentBundles(),
		ImportFinalizer:         incidentBundleImportFinalizer,
		JobFinalizer:            extensionassembly.NewIncidentBundleJobSuccessFinalizer(extensionJobFinalizer, now),
		Portability:             portability,
		IncidentPublicationLock: incidentTransactionParticipant,
		ProjectionRebuilder:     projectionRuntime.ImportRebuilder(),
		SourceCatalog:           incidentSourceCatalog,
		BlobPortability:         evidenceBlobPort,
		ReferencePacks:          incidentReferences,
		ArtifactReferences:      reporting.PortableReferenceSource{Outputs: typedObjectStore},
		Now:                     now,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Incident Bundles module: %w", err)
	}
	crossOwnerBackend, err := extensionassembly.NewCrossOwnerBackend(registryTransactions, extensionassembly.TransactionCapabilityMux{
		NetworkFlow: networkFlowModule, IncidentBundles: incidentBundleModule,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose cross-owner transaction backend: %w", err)
	}
	crossOwnerCoordinator, err := crossownertransaction.New(crossownertransaction.Options{
		Limits:            extensionassembly.CrossOwnerLimits(extensionCoordinator),
		ValidationPolicy:  extensionassembly.CrossOwnerValidationPolicy{Coordinator: extensionCoordinator, NetworkFlow: networkFlowModule},
		Backend:           crossOwnerBackend,
		Catalog:           extensionassembly.CrossOwnerDescriptors(extensionCoordinator.ParticipantContracts()),
		Timeout:           time.Duration(normalizedCfg.Timeouts.Extensions.TransactionParticipantSeconds) * time.Second,
		CancellationGrace: time.Duration(normalizedCfg.Timeouts.Extensions.CancellationGraceSeconds) * time.Second,
		FatalSink:         runtime.fatalIntegrity,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose cross-owner transaction coordinator: %w", err)
	}
	var networkFlowRoutes httpapi.RouteRegistrar
	var reportingGraphSources []reporting.GraphSourceProvider
	if networkFlowModule != nil {
		if err := networkFlowModule.InstallCrossOwnerCoordinator(crossOwnerCoordinator); err != nil {
			runtime.Close()
			return nil, fmt.Errorf("install Network Flow cross-owner transactions: %w", err)
		}
		networkFlowRoutes = networkFlowModule.RegisterRoutes()
		reportingSource, err := extensionassembly.NewAdmittedNetworkFlowReportingSource(publicationCatalog, networkFlowModule)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose Network Flow Reporting source: %w", err)
		}
		reportingGraphSources = append(reportingGraphSources, reportingSource)
	}
	if err := incidentBundleModule.InstallCrossOwnerCoordinator(crossOwnerCoordinator); err != nil {
		runtime.Close()
		return nil, fmt.Errorf("install Incident Bundles cross-owner transactions: %w", err)
	}
	var networkFlowCleanupDispatcher *networkflow.GraphResultCleanupDispatcher
	if networkFlowRouteAdmitted {
		var cleanupErr error
		networkFlowCleanupDispatcher, cleanupErr = networkFlowModule.NewGraphResultCleanupDispatcher(func() {
			runtime.publication.privateComponentLost()
		})
		if cleanupErr != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose Network Flow graph-result cleanup dispatcher: %w", cleanupErr)
		}
		runtime.networkFlowCleanupDispatcher = networkFlowCleanupDispatcher
		runtime.own(func() {
			closeCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			_ = networkFlowCleanupDispatcher.Close(closeCtx)
		})
	}
	if _, graphViewJobsAdmitted := publicationCatalog.Job(networkflow.GraphViewMaterializationJobKind); graphViewJobsAdmitted {
		if err := networkFlowModule.RegisterGraphViewWorker(); err != nil {
			runtime.Close()
			return nil, fmt.Errorf("register Network Flow graph view worker: %w", err)
		}
	}
	var incidentBundleRoutes httpapi.RouteRegistrar
	if incidentBundleRouteAdmitted {
		if err := incidentBundleModule.RegisterBundleWorker(); err != nil {
			runtime.Close()
			return nil, fmt.Errorf("register Incident Bundles worker: %w", err)
		}
		incidentBundleRoutes = incidentBundleModule.RegisterRoutes()
	}
	importOwnerLimits, importArchiveLimits := settingsProjection.Imports()
	importOwnerRegistry, err := importassembly.NewOwnerCreateRegistry(
		importassembly.OwnerRegistryDependencies{
			Postgres:                postgresHandle,
			RevisionAppender:        revisionRuntime.Appender(),
			Collaboration:           recordChanges,
			Timeline:                timelineFacade,
			EntityProjections:       projectionRuntime.EntityMutationRows(),
			AssessmentProjections:   projectionRuntime.AssessmentPorts().Rows,
			ArtifactProjections:     projectionRuntime.ArtifactPorts().Rows,
			Evidence:                evidenceOwner.ImportCreateFacade(),
			PartyProjections:        projectionRuntime.PartyPorts().Rows,
			TaskDecisionProjections: projectionRuntime.TaskDecisionMutationRows(),
			Indicators:              indicatorOwner,
		},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose imports owner registry: %w", err)
	}
	analyticalImportFacades := []imports.ExtensionImportFacade{}
	if networkFlowRouteAdmitted {
		facade, err := networkFlowModule.ImportOwner()
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose Network Flow import contribution: %w", err)
		}
		if options.NetworkFlowComposition != nil {
			facade = options.NetworkFlowComposition.WrapImportFacade(facade)
		}
		analyticalImportFacades = append(analyticalImportFacades, facade)
	}
	importModule, err := imports.NewModule(imports.ModuleDependencies{
		ImportedAttributions:   incidentbundles.ImportedAttributionResolver(),
		ApplyTransactions:      registryTransactions,
		AnalyticalTransactions: crossOwnerCoordinator,
		Postgres:               postgresHandle,
		JobTransactions:        jobTransactions,
		JobOperations:          jobManager,
		JobRunner:              runtime.jobRunner,
		Limits:                 importOwnerLimits,
		ArchiveLimits:          importArchiveLimits,
		OwnerCreateRegistry:    importOwnerRegistry,
		RevisionAppender:       revisionRuntime.Appender(),
		ExtensionProfileAdmission: func(profileID string) bool {
			return profileID == networkflow.ProfileID && networkFlowRouteAdmitted
		},
		JobSuccessFinalizer: extensionassembly.NewImportJobSuccessFinalizer(
			extensionJobFinalizer,
			postgresHandle,
			jobTransactions,
			now,
		),
		ExtensionImportFacades: analyticalImportFacades,
		Env:                    options.Env,
		CursorCodec:            cursorCodec,
		Now:                    now,
	})
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Imports module: %w", err)
	}
	if importWorkersAdmitted {
		if err := importModule.RegisterWorkers(); err != nil {
			runtime.Close()
			return nil, fmt.Errorf("register Imports workers: %w", err)
		}
	}
	importRoutes := importModule.RegisterRoutes()
	var referenceAdministration reference_data.AdministrativeApplication
	if referencePackRouteAdmitted {
		referenceAdministration, err = referenceassembly.NewAdministration(reference_data.CoordinatorOptions{Lifetime: referenceLifetime,
			Observer: referenceObserver, Postgres: referenceDatabase, Storage: referencePackStorage,
			Configuration: normalizedCfg.ReferencePacks, Limits: settingsProjection.ReferenceData(),
			JobAdmission: jobTransactions, JobOperations: jobManager, JobExecutionGuard: jobTransactions,
			JobFinalizer:  extensionassembly.NewReferencePackJobSuccessFinalizer(extensionJobFinalizer),
			RegistryUsage: referenceassembly.RegistryUsage{}, Now: now,
		}, runtime.jobRunner)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose Reference Pack administration: %w", err)
		}
	}
	referencePackRoutes := reference_data.RegisterRoutes(referenceAdministration, runtime.jobRunner)
	hostIdentityReporting, err := hostidentityreporting.New(projectionRuntime.EntityReportingReader())
	if err != nil {
		runtime.Close()
		return nil, err
	}
	artifactReporting, err := artifacts.NewReportingContribution(projectionRuntime.ArtifactPorts().Reader)
	if err != nil {
		runtime.Close()
		return nil, err
	}
	taskDecisionReporting, err := tasksdecisions.NewReportingContribution(projectionRuntime.TaskDecisionReportingReader())
	if err != nil {
		runtime.Close()
		return nil, err
	}
	linksReporting, err := reportingassembly.NewLinksProvider(evidence.NewReportingLogicalTargetContribution())
	if err != nil {
		runtime.Close()
		return nil, err
	}
	templatePacks := options.ReportingTemplatePacks
	if templatePacks == nil {
		templatePacks, err = reporting.NewTemplatePackRegistry()
		if err != nil {
			runtime.Close()
			return nil, err
		}
	}
	reportingRouteOptions := reporting.WithJobs(reporting.RouteOptions{
		Renderer:             reportRenderer,
		TemplatePacks:        templatePacks,
		ReferencePacks:       referenceRetention,
		JobSuccessFinalizer:  extensionassembly.NewReportingJobSuccessFinalizer(reportingJobFinalizer),
		SourceBoundary:       sourceboundary.NewResolver(),
		SupportRefProvider:   linksReporting,
		GraphSourceProviders: reportingGraphSources,
		ExportFieldProviders: []exportprovider.FieldProvider{
			artifactReporting,
			evidence.NewReportingFieldContribution(),
			hostIdentityReporting,
			linksReporting,
			taskDecisionReporting,
		},
	}, jobTransactions, jobManager, runtime.jobRunner)
	reportingRoutes := reporting.RegisterRoutes(reportingRouteOptions)
	var compositionPreviewJobs reportcomposition.PreviewJobPort
	if snapshotReportingRoutesAdmitted {
		compositionPreviewJobs, err = reporting.NewCompositionPreviewJobPort(runtime.jobRunner, jobTransactions, referenceRetention, templatePacks, reportingRouteOptions.GraphSourceProviders...)
		if err != nil {
			runtime.Close()
			return nil, fmt.Errorf("compose Report Composition preview job port: %w", err)
		}
	}
	reportCompositionRoutes := reportcomposition.RegisterRoutes(reportcomposition.RouteOptions{
		PreviewJobs: compositionPreviewJobs,
	})
	moduleOverrides := maps.Clone(testRuntimeDeps.ModuleOverrides)
	delete(moduleOverrides, networkflow.KeyRingsOverrideKey)
	authRouteOptions := []auth.RouteOption{}
	authRouteOptions = append(
		authRouteOptions,
		auth.WithPublicOrigin(normalizedCfg.Application.PublicOrigin),
		auth.WithSessionRevocations(collaborationRuntime),
	)
	if enterpriseAuthenticationAdmitted {
		authRouteOptions = append(authRouteOptions, auth.WithEnterpriseAuthBindings())
	}
	taskDecisionMutation, err := workbookassembly.NewTaskDecisionMutationContribution(
		postgresHandle,
		workbookConflictTokens,
		revisionRuntime.Appender(),
		revisionRuntime.ConflictFieldResolver(),
		projectionRuntime.TaskDecisionMutationRows(),
		projectionRuntime.LinkRows(),
		recordChanges,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose workbook tasks/decisions mutation contribution: %w", err)
	}
	artifactMutation, err := workbookassembly.NewArtifactMutationContribution(
		postgresHandle,
		workbookConflictTokens,
		revisionRuntime.Appender(),
		revisionRuntime.ConflictFieldResolver(),
		projectionRuntime.ArtifactPorts().Rows,
		projectionRuntime.LinkRows(),
		recordChanges,
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose workbook artifacts mutation contribution: %w", err)
	}
	noteAssociationProvider, err := workbookassembly.NewNoteAssociationProvider(postgresHandle, artifactMutation, projectionRuntime.SourceTextRows())
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose Note associations: %w", err)
	}
	workbookContributionCatalog, err := workbookassembly.NewContributionCatalog(
		workbookassembly.ContributionDependencies{
			Postgres:              postgresHandle,
			ProjectionDescriptors: projectionRuntime.DescriptorSet(),
			ProjectionQueries:     projectionRuntime,
			EntityMutationRows:    projectionRuntime.EntityMutationRows(),
			EntityQueryReader:     projectionRuntime.EntityQueryReader(),
			AssessmentProjections: projectionRuntime.AssessmentPorts().Rows,
			LinkProjections:       projectionRuntime.LinkRows(),
			PartyProjections:      projectionRuntime.PartyPorts().Rows,
			IndicatorOwner:        indicatorOwner,
			TimelineOwner:         timelineFacade,
			EvidenceOwner:         evidenceOwner.MutationContribution(),
			ArtifactOwner:         artifactMutation,
			TaskDecisionOwner:     taskDecisionMutation,
			ConflictTokens:        workbookConflictTokens,
			ConflictFields:        revisionRuntime.ConflictFieldResolver(),
			Revisions:             revisionRuntime.Appender(),
			CollaborationIntents:  recordChanges,
		},
	)
	if err != nil {
		runtime.Close()
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	builtInRoutes, err := newApplicationRouteCatalog(publicationCatalog).Bind([]routeContribution{
		{id: "auth", registrar: auth.RegisterRoutes(authRouteOptions...)},
		{id: "incidents", registrar: incidentRoutes},
		{id: "extensions", registrar: extensiondiscovery.RegisterRoutes()},
		{id: "jobs", registrar: jobapi.RegisterRoutes(jobManager)},
		{id: "saved_views", registrar: savedviews.RegisterRoutes()},
		{id: "view_schemas", registrar: viewschemas.RegisterRoutes()},
		{id: "collaboration", registrar: collaboration.RegisterRoutes(collaborationRuntime)},
		{id: "entities", registrar: entities.RegisterRoutes(entities.RouteOptions{
			CandidateReader: candidates.NewReader(postgresHandle),
			MergeStore:      timelineBundle.EntityMergeStore,
			MentionStore:    timelineBundle.EntityMentionStore,
		})},
		{id: "evidence", registrar: evidenceOwner.RouteRegistrar(settingsProjection.Evidence())},
		{
			id: "workbook",
			registrar: workbook.RegisterRoutes(workbook.RouteDependencies{
				NoteAssociations:    noteAssociationProvider,
				Catalog:             workbookContributionCatalog,
				RecordTargets:       workbookassembly.NewRecordTargetResolver(postgresHandle),
				ConflictTokens:      workbookassembly.NewConflictTokenDecoder(workbookConflictTokens),
				StartupStoreFactory: workbookassembly.NewStartupStoreFromDependencies,
			}),
		},
		{id: "timeline", registrar: timelineadmission.RegisterRoutes(timelineadmission.RouteOptions{
			Owner: timelineFacade,
		})},
		{id: "revisions", registrar: revisionRoutes},
		{id: "indicators", registrar: indicatorshttpapi.RegisterRoutes(indicatorOwner, indicatorRecords, incidentAdmission, indicatorAuth)},
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
		{id: "import", contributionIDs: []string{"import.sessions_route", "import.committed_sources_route"}, registrar: importRoutes},
		{id: "incident_portability", contributionIDs: []string{incidentbundles.BundlesRouteContributionID}, registrar: incidentBundleRoutes},
		{id: "network_flow_activity", contributionIDs: []string{networkflow.RouteContributionID}, registrar: networkFlowRoutes},
		{id: "reference_pack", contributionIDs: []string{reference_data.PacksRouteContributionID}, registrar: referencePackRoutes},
		{
			id:              "snapshot_reporting_resources",
			contributionIDs: []string{reporting.ReleasesRouteContributionID, reporting.SnapshotsRouteContributionID, reporting.RetainedArtifactsRouteContributionID},
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
	readinessProbes := []httpapi.DependencyReadinessProbe{referencePackReadinessProbe{db: postgresHandle}}
	if stagedHealth != nil {
		readinessProbes = append(readinessProbes, stagedCleanupReadinessProbe{health: stagedHealth})
	}
	publicationProjections := publication.httpProjections()
	httpOptions.Dependencies = httpapi.DependencySet{
		BrowserBuild:        browserBuild,
		Telemetry:           httpapi.TelemetrySettings{Enabled: normalizedCfg.Telemetry.Enabled, ServiceVersion: normalizedCfg.Telemetry.Resource.ServiceVersion},
		Env:                 options.Env,
		Postgres:            postgresPool,
		PostgresDB:          postgresHandle,
		ObjectStore:         objectStore,
		CursorCodec:         cursorCodec,
		Readiness:           httpapi.NewDependencyReadinessChecker(postgresPool, objectStore, readinessProbes...),
		Admission:           runtime.lifecycle,
		TestResetBootstrap:  settingsProjection.TestResetBootstrap(),
		ModuleOverrides:     moduleOverrides,
		ExtensionDiscovery:  publicationProjections,
		ExtensionClaims:     publicationProjections,
		ExtensionRoutes:     publicationProjections,
		ExtensionWorkspaces: publicationProjections,
		Now:                 now,
	}

	if err := runtime.beginPublication(ctx, extensionCoordinator, publicationTiming{
		seconds: normalizedCfg.Timeouts.Extensions.PublicationSeconds,
		grace:   time.Duration(normalizedCfg.Timeouts.Extensions.CancellationGraceSeconds) * time.Second,
	}, func(publicationCtx context.Context) error {
		if err := publication.controller.prepare(extensionPlan); err != nil {
			return err
		}
		if err := publication.controller.commit(); err != nil {
			return err
		}
		if httpOptions.Dependencies.PublicRoutes == nil {
			httpOptions.Dependencies.PublicRoutes, err = httpapi.NewRouteRegistry(
				httpapi.ExtensionClaimsFromDependencies(httpOptions.Dependencies),
			)
			if err != nil {
				return fmt.Errorf("initialize public route registry: %w", err)
			}
		}
		handler, err := dependencies.newHTTPHandler(httpOptions)
		if err != nil {
			return fmt.Errorf("setup http handler: %w", err)
		}

		runtime.handler = handler
		runtime.publicHTTP = httpOptions.Dependencies.PublicRoutes.Diagnostics()
		jobActivationCtx, cancelJobActivation := context.WithTimeout(publicationCtx, runtime.reconciliationTimeout)
		jobActivationErr := runtime.jobRunner.Activate(jobActivationCtx)
		cancelJobActivation()
		if jobActivationErr != nil {
			return fmt.Errorf("activate initial Jobs recovery: %w", jobActivationErr)
		}
		if err := publication.controller.acknowledge("job_dequeue", listenerPlanSHA256, nil); err != nil {
			return err
		}
		for _, worker := range extensionPlan.Workers() {
			if err := publication.controller.acknowledge(
				"worker:"+worker.ProfileID+":"+worker.WorkerKind,
				extensionPlan.Summary().WorkerPlanSHA256,
				nil,
			); err != nil {
				return err
			}
		}
		if runtime.processLease != nil && runtime.processLease.State() == processlease.StateLost {
			return processlease.ErrApplicationProcessLeaseLost
		}
		if runtime.servingLease != nil && runtime.servingLease.State() == processlease.StateLost {
			return processlease.ErrRecoveryServingLeaseLost
		}
		if options.ObserveJobs != nil {
			options.ObserveJobs(jobManager, jobTransactions, runtime.jobRunner, postgresPool)
		}
		if options.ObserveCollaboration != nil {
			options.ObserveCollaboration(collaborationRuntime)
		}
		if options.ObserveEvidenceCleanup != nil {
			options.ObserveEvidenceCleanup(cleanupDispatcher)
		}
		if options.ObserveNetworkFlowCleanup != nil && networkFlowCleanupDispatcher != nil {
			options.ObserveNetworkFlowCleanup(networkFlowCleanupDispatcher)
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
		return nil
	}); err != nil {
		runtime.Close()
		return nil, err
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

func instrumentedPostgres(enabled bool, serviceVersion string, pool dbport.DB) dbport.DB {
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

type referencePackReadinessProbe struct{ db dbport.DB }

func (referencePackReadinessProbe) ReadinessName() string { return "reference_packs" }
func (probe referencePackReadinessProbe) CheckReadinessDependency(ctx context.Context) error {
	return reference_data.CheckReadyState(ctx, probe.db)
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

type publicationStartupFatal struct{ reason string }

func (e publicationStartupFatal) Error() string           { return e.reason }
func (e publicationStartupFatal) FatalReasonCode() string { return e.reason }

func (r *Runtime) ActivatePublication() error {
	if r == nil || r.publicationOperation == nil {
		return errors.New("publication operation unavailable")
	}
	return r.publicationOperation.activate()
}

func (r *Runtime) activatePreparedPublication(ctx context.Context) (resultErr error) {
	if r == nil || r.publication == nil {
		return fmt.Errorf("extension_publication_failed")
	}
	defer func() {
		if resultErr != nil {
			r.publication.abortStartup()
		}
	}()
	if r.processLease != nil && r.processLease.State() != processlease.StateHeld {
		return fmt.Errorf("extension_publication_failed")
	}
	if r.servingLease != nil && r.servingLease.State() != processlease.StateHeld {
		return fmt.Errorf("extension_publication_failed")
	}
	if r.collaborationRuntime != nil {
		if err := r.collaborationRuntime.Prepare(ctx, r.publication.servingReady); err != nil {
			return fmt.Errorf("activate collaboration dispatcher: %w", err)
		}
	}
	if r.evidenceCleanupDispatcher != nil {
		if err := r.evidenceCleanupDispatcher.Prepare(ctx, r.publication.servingReady); err != nil {
			return fmt.Errorf("activate Evidence cleanup dispatcher: %w", err)
		}
	}
	if r.graphActivation != nil {
		if err := r.graphActivation.Prepare(ctx, r.publication.servingReady); err != nil {
			return fmt.Errorf("activate Graph reconstruction: %w", err)
		}
	}
	if r.networkFlowCleanupDispatcher != nil {
		if r.graphActivation == nil {
			return errors.New("graph cleanup readiness unavailable")
		}
		if err := r.networkFlowCleanupDispatcher.Prepare(ctx, r.graphActivation.dependentReady); err != nil {
			return fmt.Errorf("activate Network Flow graph-result cleanup dispatcher: %w", err)
		}
	}
	if r.referenceCollector != nil {
		if err := r.referenceCollector.Prepare(ctx, r.publication.servingReady); err != nil {
			return fmt.Errorf("activate Reference Pack collection: %w", err)
		}
	}
	r.publicationOnce.Do(func() {
		if r.stagedJanitor != nil && r.stagedJanitorContext != nil {
			go func() {
				defer func() {
					if recovered := recover(); recovered != nil && r.lifecycle != nil {
						r.publication.privateComponentLost()
					}
				}()
				select {
				case <-r.stagedJanitorContext.Done():
					return
				case <-r.publication.servingReady:
				}
				if r.stagedJanitorContext.Err() != nil {
					return
				}
				if err := r.stagedJanitor.Run(r.stagedJanitorContext, r.stagedObjectSweepPeriod); err != nil && r.lifecycle != nil {
					r.publication.privateComponentLost()
				}
			}()
		}
	})
	// The HTTP runtime calls this only after its accept loop is installed and
	// quiescent. Both HTTP and WebSocket admission share that physical listener.
	summary, installed := r.publication.summary()
	if !installed {
		return fmt.Errorf("extension_publication_failed")
	}
	for _, component := range []string{"http", "websocket"} {
		if err := r.publication.acknowledge(component, summary.ListenerPlanSHA256, nil); err != nil {
			return err
		}
	}

	return r.publication.serve()
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
		if r.publicationOperation != nil {
			r.publicationOperation.stop()
		}
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

// fatalIntegrity imports the selected owner outcome into the process lifecycle.
// Grace exhaustion is observed independently of the blocked operation so the
// runner can exit without joining callbacks or releasing their capabilities.
func (r *Runtime) fatalIntegrity(cause error) {
	reason := "indeterminate_database_commit"
	var classified interface{ FatalReasonCode() string }
	if errors.As(cause, &classified) {
		reason = classified.FatalReasonCode()
	} else if errors.Is(cause, extensions.ErrStateReadbackMismatch) {
		reason = "migration_ledger_state_mismatch"
	}
	r.lifecycle.Fatal(reason)
	if reason == "operation_cancellation_grace_exhausted" && r.immediateFatal != nil {
		r.immediateFatal(processlifecycle.FatalSignal{ReasonCode: r.lifecycle.FatalReason(), ExitCode: 70})
	}
}
