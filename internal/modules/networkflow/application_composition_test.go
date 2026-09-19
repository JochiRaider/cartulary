package networkflow

import (
	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"net/http"
	"testing"
	"time"
)

// Nil embedded ports panic on any I/O: construction must only retain them.
type constructionPorts struct {
	postgres.DB
	ImportSourcePort
	IncidentLockPort
	AdministrativeAuditPort
	IndicatorParticipationPort
	ResourceIntentAppender
	GraphViewJobTransactions
	GraphViewJobManager
	GraphViewJobFinalizer
	GraphViewJobRunner
	registrations int
}

func (p *constructionPorts) RegisterHandler(_ string, _ jobs.HandlerFunc) error {
	p.registrations++
	return nil
}

func constructionModule(t *testing.T, active bool) (*Module, *constructionPorts) {
	t.Helper()
	ports := &constructionPorts{}
	deps := ModuleDependencies{Postgres: ports, ImportSources: ports, IncidentLocks: ports,
		AuditAppender: ports, Indicators: ports, ResourceIntents: ports, EffectiveLimits: defaultEffectiveLimits(),
		GraphViewJobs: ports, JobManager: ports, JobRunner: ports, JobFinalizer: ports,
		Now: func() time.Time { return time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC) },
	}
	if active {
		deps.KeyRings = &KeyRings{cursorActiveID: "cursor", cursorKeys: map[string]cursorKeyMaterial{"cursor": {key: make([]byte, 32), state: "active"}},
			safeActiveID: "safe", safeKeys: map[string]safeDigestKeyMaterial{"safe": {key: make([]byte, 32), state: "active"}}}
	}
	module, err := NewModule(deps)
	if err != nil {
		t.Fatal(err)
	}
	if module.ReportingGraphSource() == nil {
		t.Fatal("Reporting source unavailable")
	}
	if module.ReportingGraphSource().db != deps.Postgres || module.graphVerifier.graphTelemetry != nil {
		t.Fatal("incoherent source contribution")
	}
	return module, ports
}

func assertModuleComposition(t *testing.T) {
	inactive, ports := constructionModule(t, false)
	if err := inactive.InstallCrossOwnerCoordinator(&crossownertransaction.Coordinator{}); err != nil {
		t.Fatal(err)
	}
	if _, err := inactive.ImportOwner(); err == nil || inactive.active != nil || ports.registrations != 0 {
		t.Fatal("inactive active contribution admitted")
	}
	if _, err := NewGraphRestoreSourceRegistration(ports); err != nil {
		t.Fatal(err)
	}
	for name, remove := range map[string]func(*Module){
		"coordinator":      func(m *Module) { m.transactions = nil },
		"persistence":      func(m *Module) { m.store = nil },
		"source":           func(m *Module) { m.importSources = nil },
		"admission":        func(m *Module) { m.incidentAccess = nil },
		"auth":             func(m *Module) { m.authStore = nil },
		"clock":            func(m *Module) { m.now = nil },
		"digest":           func(m *Module) { m.safeDigester = nil },
		"cursor":           func(m *Module) { m.cursorProtector = nil },
		"composer":         func(m *Module) { m.graphComposer = nil },
		"verifier":         func(m *Module) { m.graphVerifier = nil },
		"projector":        func(m *Module) { m.graphProjection = nil },
		"reporting":        func(m *Module) { m.reportingSource = nil },
		"job transactions": func(m *Module) { m.graphViewJobs = nil },
		"job manager":      func(m *Module) { m.jobManager = nil },
		"job runner":       func(m *Module) { m.jobRunner = nil },
		"job finalizer":    func(m *Module) { m.jobFinalizer = nil },
	} {
		t.Run(name, func(t *testing.T) {
			m, ports := constructionModule(t, true)
			if err := m.InstallCrossOwnerCoordinator(&crossownertransaction.Coordinator{}); err != nil {
				t.Fatal(err)
			}
			remove(m)
			if err := m.RegisterGraphViewWorker(); err == nil {
				t.Fatal("incomplete worker admitted")
			}
			if _, err := m.ImportOwner(); err == nil {
				t.Fatal("incomplete import admitted")
			}
			if _, err := m.NewGraphResultCleanupDispatcher(func() {}); err == nil {
				t.Fatal("incomplete cleanup admitted")
			}
			if err := m.RegisterRoutes()(http.NewServeMux(), httpapi.DependencySet{}); err == nil {
				t.Fatal("incomplete routes admitted")
			}
			if m.active != nil || ports.registrations != 0 {
				t.Fatal("failed construction leaked active work")
			}
		})
	}
	m, ports := constructionModule(t, true)
	reportingSource := m.ReportingGraphSource()
	if _, err := m.ImportOwner(); err == nil {
		t.Fatal("import before coordinator admitted")
	}
	coordinator := &crossownertransaction.Coordinator{}
	if err := m.InstallCrossOwnerCoordinator(coordinator); err != nil {
		t.Fatal(err)
	}
	if err := m.InstallCrossOwnerCoordinator(coordinator); err == nil {
		t.Fatal("coordinator reinstallation admitted")
	}
	owner, err := m.ImportOwner()
	if err != nil {
		t.Fatal(err)
	}
	again, err := m.ImportOwner()
	if err != nil || again != owner || m.ReportingGraphSource() != reportingSource || ports.registrations != 0 {
		t.Fatal("construction was not retained/quiescent")
	}
	if m.active.tables.store != m.store || m.active.links.store != m.store || m.active.savedGraphs.store != m.store ||
		m.active.tables.incidentAccess != m.incidentAccess || m.active.links.incidentAccess != m.incidentAccess ||
		m.active.savedGraphs.incidentAccess != m.incidentAccess || !m.active.tables.now().Equal(m.now()) ||
		!m.active.links.now().Equal(m.now()) || !m.active.savedGraphs.now().Equal(m.now()) {
		t.Fatal("mixed database/clock/admission")
	}
	if err := m.RegisterGraphViewWorker(); err != nil || ports.registrations != 1 {
		t.Fatal("explicit registration failed")
	}
}
