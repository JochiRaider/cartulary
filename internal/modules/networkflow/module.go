package networkflow

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/gen/networkflowroutes"
	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// ImportSourcePort is the Network Flow consumer-owned source capability. The
// Import module supplies the adapter; Network Flow does not construct or query
// an Import store directly.
type ImportSourcePort interface {
	OpenSourceStream(context.Context, string) (imports.ImportSourceStream, error)
	GetSession(context.Context, uuid.UUID) (map[string]any, uuid.UUID, error)
}

type importTransactionPort interface {
	ValidateExtensionApplyPreconditionsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, string, string) error
}

type ModuleDependencies struct {
	Postgres        postgres.DB
	ImportSources   ImportSourcePort
	KeyRings        *KeyRings
	Limits          Limits
	Now             func() time.Time
	IncidentLocks   IncidentLockPort
	AuditAppender   AdministrativeAuditPort
	Indicators      IndicatorParticipationPort
	ResourceIntents ResourceIntentAppender
}

// Module is the single Network Flow composition facade. Transport and generic
// import integration receive narrow views of this root rather than assembling
// sibling stores independently.
type Module struct {
	store           *Store
	importOwner     imports.ExtensionImportFacade
	importSources   ImportSourcePort
	importTx        importTransactionPort
	transactions    *crossownertransaction.Coordinator
	cursorProtector CursorProtector
	safeDigester    SafeDigester
	limits          Limits
	now             func() time.Time
}

func NewModule(dependencies ModuleDependencies) (*Module, error) {
	if dependencies.Postgres == nil {
		return nil, errors.New("network flow module requires PostgreSQL")
	}
	limits := dependencies.Limits.normalized()
	if dependencies.Limits == (Limits{}) {
		limits = DefaultLimits()
	}
	now := dependencies.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	if dependencies.IncidentLocks == nil || dependencies.AuditAppender == nil || dependencies.Indicators == nil || dependencies.ResourceIntents == nil {
		return nil, errors.New("network flow owner transaction participants are required")
	}
	var safeDigester SafeDigester
	var cursorProtector CursorProtector
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
	store := NewStore(
		dependencies.Postgres,
		WithLimits(limits),
		WithOwnerParticipants(dependencies.IncidentLocks, dependencies.AuditAppender, dependencies.Indicators),
		WithSafeDigester(safeDigester),
		WithResourceIntentAppender(dependencies.ResourceIntents),
	)
	module := &Module{store: store, importSources: dependencies.ImportSources, cursorProtector: cursorProtector, safeDigester: safeDigester, limits: limits, now: now}
	if physical, ok := dependencies.ImportSources.(importTransactionPort); ok {
		module.importTx = physical
	}
	return module, nil
}

// InstallCrossOwnerCoordinator completes the application-owned composition
// edge while every route and worker remains quiescent. It is deliberately
// single-assignment so a serving process cannot switch transaction epochs.
func (m *Module) InstallCrossOwnerCoordinator(coordinator *crossownertransaction.Coordinator) error {
	if m == nil || coordinator == nil || m.importSources == nil || m.importTx == nil {
		return errors.New("network flow cross-owner transaction composition unavailable")
	}
	if m.transactions != nil {
		return errors.New("network flow cross-owner transaction coordinator already installed")
	}
	m.transactions = coordinator
	m.importOwner = newImportFacade(m.store, m.importSources, m.limits, m.now, m.safeDigester)
	return nil
}

func (m *Module) ImportOwner() imports.ExtensionImportFacade {
	if m == nil {
		return nil
	}
	return m.importOwner
}

func (m *Module) RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if m == nil {
			return errors.New("network flow module unavailable")
		}
		if m.transactions == nil {
			return errors.New("network flow cross-owner transaction coordinator unavailable")
		}
		service, err := newRouteService(deps, m)
		if err != nil {
			return err
		}
		return registerNetworkFlowRoutes(mux, service)
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
		if m.importTx == nil {
			return nil, nil, crossownertransaction.ErrUnavailable
		}
	case IndicatorLinkParticipantID:
	default:
		return nil, nil, fmt.Errorf("%w: %s", crossownertransaction.ErrParticipantSet, participantID)
	}
	capability := &transactionCapability{
		participantID: participantID,
		tx:            tx,
		store:         m.store,
		imports:       m.importTx,
	}
	return capability, capability, nil
}

func registerNetworkFlowRoutes(mux *http.ServeMux, service *Service) error {
	handlers := map[string]http.HandlerFunc{
		"nf.graphs.contributors.query": service.handleGraphContributorsQuery,
		"nf.graphs.query":              service.handleGraphQuery,
		"nf.indicator_links.create":    service.handleIndicatorLinks,
		"nf.rejected_rows.query":       service.handleRejectedRowsQuery,
		"nf.rows.query":                service.handleRowsQuery,
		"nf.source_profiles.list":      service.handleSourceProfiles,
		"nf.tables.delete":             service.handleTableResource,
		"nf.tables.get":                service.handleTableResource,
		"nf.tables.list":               service.handleTablesCollection,
		"nf.tables.patch":              service.handleTableResource,
		"nf.tables.query":              service.handleTableRowsQuery,
	}
	routes := networkflowroutes.All()
	if len(routes) != len(handlers) {
		return fmt.Errorf("network flow route parity failed: contract=%d handlers=%d", len(routes), len(handlers))
	}
	for _, route := range routes {
		handler, ok := handlers[route.RouteID]
		if !ok {
			return fmt.Errorf("network flow route %q has no handler", route.RouteID)
		}
		mux.HandleFunc(route.Pattern, handler)
		delete(handlers, route.RouteID)
	}
	if len(handlers) != 0 {
		return fmt.Errorf("network flow has handlers outside its route contract")
	}
	return nil
}
