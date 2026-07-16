package networkflow

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

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

type ModuleDependencies struct {
	Postgres      postgres.DB
	ImportSources ImportSourcePort
	KeyRings      *KeyRings
	Limits        Limits
	Now           func() time.Time
	Transactions  postgres.TransactionRunner
	IncidentLocks IncidentLockPort
	AuditAppender AdministrativeAuditPort
	Indicators    IndicatorParticipationPort
}

// Module is the single Network Flow composition facade. Transport and generic
// import integration receive narrow views of this root rather than assembling
// sibling stores independently.
type Module struct {
	store           *Store
	importOwner     imports.ExtensionImportFacade
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
	transactions := dependencies.Transactions
	if transactions == nil {
		transactions = postgres.NewTransactionRunner(dependencies.Postgres)
	}
	if dependencies.IncidentLocks == nil || dependencies.AuditAppender == nil || dependencies.Indicators == nil {
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
		WithTransactionRunner(transactions),
		WithSafeDigester(safeDigester),
	)
	module := &Module{store: store, cursorProtector: cursorProtector, safeDigester: safeDigester, limits: limits, now: now}
	if dependencies.ImportSources != nil {
		module.importOwner = newImportFacade(store, dependencies.ImportSources, limits, now, safeDigester)
	}
	return module, nil
}

func (m *Module) ImportOwner() imports.ExtensionImportFacade {
	if m == nil {
		return nil
	}
	return m.importOwner
}

func (m *Module) RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.ExtensionProfileClaimedIn(deps.ExtensionProfiles, ProfileID) {
			return nil
		}
		if m == nil {
			return errors.New("network flow module unavailable")
		}
		service, err := newRouteService(deps, m)
		if err != nil {
			return err
		}
		registerNetworkFlowRoutes(mux, service)
		return nil
	}
}

func registerNetworkFlowRoutes(mux *http.ServeMux, service *Service) {
	mux.HandleFunc("GET "+routeRoot+"/source-profiles", service.handleSourceProfiles)
	mux.HandleFunc("GET "+routeRoot+"/tables", service.handleTablesCollection)
	mux.HandleFunc("GET "+routeRoot+"/tables/{network_flow_table_id}", service.handleTableResource)
	mux.HandleFunc("PATCH "+routeRoot+"/tables/{network_flow_table_id}", service.handleTableResource)
	mux.HandleFunc("DELETE "+routeRoot+"/tables/{network_flow_table_id}", service.handleTableResource)
	mux.HandleFunc("POST "+routeRoot+"/tables/{network_flow_table_id}/query", service.handleTableRowsQuery)
	mux.HandleFunc("POST "+routeRoot+"/tables/{network_flow_table_id}/rejected-rows/query", service.handleRejectedRowsQuery)
	mux.HandleFunc("POST "+routeRoot+"/rows/query", service.handleRowsQuery)
	mux.HandleFunc("POST "+routeRoot+"/graphs/query", service.handleGraphQuery)
	mux.HandleFunc("POST "+routeRoot+"/graphs/contributors/query", service.handleGraphContributorsQuery)
	mux.HandleFunc("POST "+routeRoot+"/indicator-links", service.handleIndicatorLinks)
}
