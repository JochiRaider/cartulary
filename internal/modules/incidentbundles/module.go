package incidentbundles

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
)

// ModuleDependencies is the complete Incident Bundles production composition
// boundary. NewModule rejects every missing reference dependency before any
// worker or route can be published.
type ModuleDependencies struct {
	Postgres                *pgxpool.Pool
	JobTransactions         JobTransactions
	JobOperations           JobOperations
	JobRunner               JobRunner
	Storage                 BundleStorage
	Limits                  Limits
	ImportFinalizer         incidents.IncidentBundleImportFinalizer
	JobFinalizer            JobSuccessFinalizer
	Portability             *PortabilityOrchestrator
	IncidentPublicationLock IncidentPublicationLock
	ProjectionRebuilder     ImportProjectionRebuilder
	SourceCatalog           *sourceport.Catalog
	HistoricalIntentPolicy  HistoricalIntentPolicy
	BlobPortability         BlobPortability
	Now                     func() time.Time
}

// Module is the sole production construction facade for Incident Bundles.
// Application composition completes its explicit lifecycle before routes are
// admitted: install the coordinator, register the worker, then bind routes.
type Module struct {
	mu sync.Mutex

	pool             *pgxpool.Pool
	store            *store
	worker           *incidentBundleWorker
	importer         importer
	jobs             JobTransactions
	storage          BundleStorage
	now              func() time.Time
	transactions     *crossownertransaction.Coordinator
	workerRegistered bool
}

// NewModule validates the complete dependency graph and constructs quiescent
// owner state. It does not publish a worker or routes.
func NewModule(dependencies ModuleDependencies) (*Module, error) {
	switch {
	case dependencies.Postgres == nil:
		return nil, errors.New("incident bundle PostgreSQL dependency is required")
	case dependencies.JobTransactions == nil:
		return nil, errors.New("incident bundle Jobs transaction dependency is required")
	case dependencies.JobOperations == nil:
		return nil, errors.New("incident bundle Jobs operations dependency is required")
	case dependencies.JobRunner == nil:
		return nil, errors.New("incident bundle Jobs runner dependency is required")
	case dependencies.Storage == nil:
		return nil, errors.New("incident bundle storage dependency is required")
	case dependencies.ImportFinalizer == nil:
		return nil, errors.New("incident bundle import finalizer dependency is required")
	case dependencies.JobFinalizer == nil:
		return nil, errors.New("incident bundle job finalizer dependency is required")
	case dependencies.Portability == nil:
		return nil, errors.New("incident bundle portability dependency is required")
	case dependencies.IncidentPublicationLock == nil:
		return nil, errors.New("incident bundle publication lock dependency is required")
	case dependencies.ProjectionRebuilder == nil:
		return nil, errors.New("incident bundle projection rebuilder dependency is required")
	case dependencies.SourceCatalog == nil:
		return nil, errors.New("incident bundle source catalog dependency is required")
	case dependencies.HistoricalIntentPolicy == nil:
		return nil, errors.New("incident bundle historical intent policy dependency is required")
	case dependencies.BlobPortability == nil:
		return nil, errors.New("incident bundle blob portability dependency is required")
	}

	now := dependencies.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	store := newStore(dependencies.Postgres, dependencies.JobTransactions)
	importer := importer{
		pool:              dependencies.Postgres,
		blobPort:          dependencies.BlobPortability,
		finalizer:         dependencies.ImportFinalizer,
		projectionRebuild: dependencies.ProjectionRebuilder,
		sourceCatalog:     dependencies.SourceCatalog,
		historicalIntents: dependencies.HistoricalIntentPolicy,
	}
	worker := newIncidentBundleWorker(
		store,
		dependencies.Postgres,
		dependencies.JobOperations,
		dependencies.JobRunner,
		dependencies.Storage,
		dependencies.ImportFinalizer,
		dependencies.JobFinalizer,
		dependencies.Portability,
		dependencies.IncidentPublicationLock,
		dependencies.ProjectionRebuilder,
		dependencies.SourceCatalog,
		dependencies.HistoricalIntentPolicy,
		dependencies.BlobPortability,
		dependencies.Limits,
		now,
	)
	return &Module{
		pool:     dependencies.Postgres,
		store:    store,
		worker:   worker,
		importer: importer,
		jobs:     dependencies.JobTransactions,
		storage:  dependencies.Storage,
		now:      now,
	}, nil
}

// TransactionCapabilities exposes the Incident Bundles logical capability
// over an application-owned physical transaction.
func (m *Module) TransactionCapabilities(participantID string, tx pgx.Tx) (crossownertransaction.ReadCapability, crossownertransaction.WriteCapability, error) {
	if m == nil || m.store == nil || m.jobs == nil || tx == nil || participantID != ImportTransactionParticipantID {
		return nil, nil, fmt.Errorf("%w: %s", crossownertransaction.ErrParticipantSet, participantID)
	}
	capability := &importTransactionCapability{
		participantID: participantID,
		tx:            tx,
		importer:      m.importer,
		jobs:          m.jobs,
		now:           m.now,
	}
	return capability, capability, nil
}

// InstallCrossOwnerCoordinator completes the application-owned transaction
// edge exactly once while the module remains unpublished.
func (m *Module) InstallCrossOwnerCoordinator(coordinator *crossownertransaction.Coordinator) error {
	if m == nil || coordinator == nil {
		return errors.New("incident bundle cross-owner transaction coordinator is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.transactions != nil {
		return errors.New("incident bundle cross-owner transaction coordinator already installed")
	}
	if m.workerRegistered {
		return errors.New("incident bundle worker was registered before coordinator installation")
	}
	m.transactions = coordinator
	m.worker.transactions = coordinator
	return nil
}

// RegisterBundleWorker publishes the named worker exactly once and only after
// the coordinator has been installed.
func (m *Module) RegisterBundleWorker() error {
	if m == nil {
		return errors.New("incident bundle module is unavailable")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.transactions == nil {
		return errors.New("incident bundle cross-owner transaction coordinator is not installed")
	}
	if m.workerRegistered {
		return errors.New("incident bundle worker is already registered")
	}
	if err := m.worker.registerJobHandler(); err != nil {
		return fmt.Errorf("register incident bundle worker: %w", err)
	}
	m.workerRegistered = true
	return nil
}

func (m *Module) routeDependenciesReady() error {
	if m == nil {
		return errors.New("incident bundle module is unavailable")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.transactions == nil {
		return errors.New("incident bundle cross-owner transaction coordinator is not installed")
	}
	if !m.workerRegistered {
		return errors.New("incident bundle worker is not registered")
	}
	return nil
}
