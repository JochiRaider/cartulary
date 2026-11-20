package networkflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

// ImportSourcePort is the Network Flow consumer-owned source capability. The
// Import module supplies the adapter; Network Flow does not construct or query
// an Import store directly.
type ImportSourcePort interface {
	OpenSourceStream(context.Context, string) (imports.ImportSourceStream, error)
	OpenSourceStreamTx(context.Context, pgx.Tx, string) (imports.ImportSourceStream, error)
	ValidateExtensionApplyPreconditionsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, string, string) error
}

type ModuleDependencies struct {
	TableIDEntropy       io.Reader
	Postgres             dbport.DB
	ImportSources        ImportSourcePort
	ImportedAttributions sourceport.PortableAttributionResolver
	KeyRings             *KeyRings
	EffectiveLimits      EffectiveLimits
	Now                  func() time.Time
	IncidentLocks        IncidentLockPort
	AuditAppender        AdministrativeAuditPort
	Indicators           IndicatorParticipationPort
	ResourceIntents      ResourceIntentAppender
	GraphViewJobs        GraphViewJobTransactions
	JobManager           GraphViewJobManager
	JobRunner            GraphViewJobRunner
	JobFinalizer         GraphViewJobFinalizer
	GraphTelemetry       GraphTelemetryObserver
	GraphStateReady      func() bool
}

// Module is the single Network Flow composition facade. Transport and generic
// import integration receive narrow views of this root rather than assembling
// sibling stores independently.
type Module struct {
	store           *store
	active          *activeApplications
	incidentAccess  incidentAdmissionChecker
	authStore       *authn.Store
	reportingSource *ReportingGraphSource
	graphVerifier   *graphSourceComposer
	importSources   ImportSourcePort
	transactions    *crossownertransaction.Coordinator
	cursorProtector cursorProtector
	safeDigester    safeDigester
	limits          EffectiveLimits
	now             func() time.Time
	graphComposer   *graphSourceComposer
	graphProjection graphProjectionPort
	graphViewJobs   GraphViewJobTransactions
	jobManager      GraphViewJobManager
	jobRunner       GraphViewJobRunner
	jobFinalizer    GraphViewJobFinalizer
	graphTelemetry  GraphTelemetryObserver
	graphStateReady func() bool
}

func NewModule(dependencies ModuleDependencies) (*Module, error) {
	if dependencies.Postgres == nil {
		return nil, errors.New("network flow module requires PostgreSQL")
	}
	if dependencies.ImportSources == nil {
		return nil, errors.New("network flow module requires Imports source capability")
	}
	if dependencies.ImportedAttributions == nil {
		return nil, errors.New("network flow module requires historical attribution resolver")
	}
	if dependencies.GraphStateReady == nil {
		return nil, errors.New("network flow module requires Graph state readiness")
	}
	limits := dependencies.EffectiveLimits
	if err := checkEffectiveLimits(limits); err != nil {
		return nil, fmt.Errorf("network flow module effective limits: %w", err)
	}
	now := dependencies.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	if dependencies.IncidentLocks == nil || dependencies.AuditAppender == nil || dependencies.Indicators == nil || dependencies.ResourceIntents == nil {
		return nil, errors.New("network flow owner transaction participants are required")
	}
	var safeDigester safeDigester
	var cursorProtector cursorProtector
	if dependencies.KeyRings != nil {
		var err error
		safeDigester, err = newSafeDigester(dependencies.KeyRings, now)
		if err != nil {
			return nil, err
		}
		cursorProtector, err = newCursorCodec(dependencies.KeyRings, now)
		if err != nil {
			return nil, err
		}
	}
	store := newStore(
		dependencies.Postgres,
		limits,
		withOwnerParticipants(dependencies.IncidentLocks, dependencies.AuditAppender, dependencies.Indicators),
		withSafeDigester(safeDigester),
		withTableIDEntropy(dependencies.TableIDEntropy),
		withResourceIntentAppender(dependencies.ResourceIntents),
		withImportedAttributions(dependencies.ImportedAttributions),
	)
	module := &Module{
		store: store, importSources: dependencies.ImportSources,
		incidentAccess:  admission.NewChecker(dependencies.Postgres),
		authStore:       authn.NewStore(dependencies.Postgres),
		cursorProtector: cursorProtector, safeDigester: safeDigester,
		limits: limits, now: now, graphProjection: newGraphProjectionAdapter(),
		graphViewJobs: dependencies.GraphViewJobs, jobManager: dependencies.JobManager,
		jobRunner: dependencies.JobRunner, jobFinalizer: dependencies.JobFinalizer,
		graphTelemetry: dependencies.GraphTelemetry, graphStateReady: dependencies.GraphStateReady,
	}
	composer, err := newGraphSourceComposer(store, limits, module.graphProjection, dependencies.GraphTelemetry)
	if err != nil {
		return nil, err
	}
	module.graphComposer = composer
	module.graphVerifier, err = newGraphSourceComposer(store, limits, module.graphProjection, nil)
	if err != nil {
		return nil, err
	}
	module.reportingSource, err = newReportingGraphSource(store, dependencies.GraphStateReady)
	if err != nil {
		return nil, err
	}
	return module, nil
}

// RegisterGraphViewWorker binds the one canonical materialization handler.
// Registration is explicit so constructing a module never starts work or
// hides process-level concurrency.
func (m *Module) RegisterGraphViewWorker() error {
	if err := m.prepareActiveApplications(); err != nil {
		return err
	}
	return m.jobRunner.RegisterHandlerWhenReady(graphViewWorkerKind, m.handleGraphViewMaterialization, func(context.Context, uuid.UUID) (bool, error) {
		return m.graphStateReady(), nil
	})
}

// InstallCrossOwnerCoordinator completes the application-owned composition
// edge while every route and worker remains quiescent. It is deliberately
// single-assignment so a serving process cannot switch transaction epochs.
func (m *Module) InstallCrossOwnerCoordinator(coordinator *crossownertransaction.Coordinator) error {
	if m == nil || coordinator == nil || m.importSources == nil {
		return errors.New("network flow cross-owner transaction composition unavailable")
	}
	if m.transactions != nil {
		return errors.New("network flow cross-owner transaction coordinator already installed")
	}
	m.transactions = coordinator
	return nil
}

func (m *Module) ImportOwner() (imports.ExtensionImportFacade, error) {
	if err := m.prepareActiveApplications(); err != nil {
		return nil, err
	}
	return m.active.importOwner, nil
}

func (m *Module) RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if m == nil {
			return errors.New("network flow module unavailable")
		}
		if m.transactions == nil {
			return errors.New("network flow cross-owner transaction coordinator unavailable")
		}
		service, err := newRouteService(m)
		if err != nil {
			return err
		}
		// Central admission in Network Flow's own order (NF NLSpec Table 21-C,
		// Core 01 REQ-01-707 family clause): the incident gate runs between the
		// incident identifier and the table or graph-view identifier.
		return httpapi.BindAdmittedOwnerOrderRoutes(mux, deps, httpapi.OwnerOrder{
			OwnerID: routeOwnerID, Gate: service.admitIncident, Handlers: service.operationHandlers(),
		})
	}
}

// TransactionCapabilities is the physical composition adapter used only by
// app-server's PostgreSQL backend. The returned values expose owner-logical
// methods to participants and never expose pgx to the shared coordinator.
func (m *Module) TransactionCapabilities(participantID string, tx pgx.Tx) (crossownertransaction.ReadCapability, crossownertransaction.WriteCapability, error) {
	if m == nil || m.store == nil || tx == nil {
		return nil, nil, crossownertransaction.ErrUnavailable
	}
	switch participantID {
	case ImportApplyParticipantID:
		if m.importSources == nil {
			return nil, nil, crossownertransaction.ErrUnavailable
		}
		return &importApplyReadAccess{tx: tx, sources: m.importSources},
			&importApplyWriteAccess{tx: tx, importOwner: newImportFacade(m.store, m.importSources, m.limits, m.now, m.safeDigester)}, nil
	case IndicatorLinkParticipantID:
		return &indicatorLinkReadAccess{tx: tx, store: m.store}, &indicatorLinkWriteAccess{tx: tx, store: m.store}, nil
	default:
		return nil, nil, fmt.Errorf("%w: %s", crossownertransaction.ErrParticipantSet, participantID)
	}
}
