package jobs

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Manager is the Jobs facade. Persistence and policy remain private package
// concerns; consumers receive narrower interfaces during application assembly.
type Manager struct {
	pool                  *pgxpool.Pool
	now                   func() time.Time
	transactions          *TransactionService
	serviceVersion        string
	activeGaugeRegistered bool
	definitions           *jobDefinitionCatalog
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Configure(pool *pgxpool.Pool, transactions *TransactionService, now func() time.Time) {
	m.pool = pool
	m.transactions = transactions
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	m.now = now
	m.registerActiveGauge()
}

func (m *Manager) ConfigureTelemetry(serviceVersion string) {
	if m == nil {
		return
	}
	m.serviceVersion = serviceVersion
	m.registerActiveGauge()
}

func (m *Manager) ensureConfigured() error {
	if m == nil || m.pool == nil || m.transactions == nil {
		return ErrNotConfigured
	}
	if m.now == nil {
		m.now = func() time.Time { return time.Now().UTC() }
	}
	return nil
}
