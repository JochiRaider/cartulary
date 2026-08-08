package jobs

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RuntimePolicy struct {
	HandlerLease       time.Duration
	LeaseRenewal       time.Duration
	RecoveryScan       time.Duration
	RecoveryBatch      int
	HandlerConcurrency int
	MaximumFailures    int
	RetryDelays        []time.Duration
	ExpirySweep        time.Duration
	ExpiryBatch        int
}

func ProductionRuntimePolicy() RuntimePolicy {
	return RuntimePolicy{
		HandlerLease:       30 * time.Second,
		LeaseRenewal:       10 * time.Second,
		RecoveryScan:       5 * time.Second,
		RecoveryBatch:      100,
		HandlerConcurrency: 8,
		MaximumFailures:    3,
		RetryDelays:        []time.Duration{5 * time.Second, 30 * time.Second},
		ExpirySweep:        5 * time.Minute,
		ExpiryBatch:        1000,
	}
}

func (policy RuntimePolicy) validate() error {
	if policy.HandlerLease <= 0 || policy.LeaseRenewal <= 0 || policy.LeaseRenewal >= policy.HandlerLease ||
		policy.RecoveryScan <= 0 || policy.RecoveryBatch <= 0 || policy.HandlerConcurrency <= 0 ||
		policy.MaximumFailures < 1 || len(policy.RetryDelays) != policy.MaximumFailures-1 ||
		policy.ExpirySweep <= 0 || policy.ExpiryBatch <= 0 {
		return fmt.Errorf("%w: invalid runtime policy", ErrInvalidJobDefinition)
	}
	for _, delay := range policy.RetryDelays {
		if delay <= 0 {
			return fmt.Errorf("%w: invalid retry delay", ErrInvalidJobDefinition)
		}
	}
	return nil
}

func cloneRuntimePolicy(policy RuntimePolicy) RuntimePolicy {
	policy.RetryDelays = append([]time.Duration(nil), policy.RetryDelays...)
	return policy
}

type ManagerOptions struct {
	Postgres                *pgxpool.Pool
	Transactions            *TransactionService
	Catalog                 *Catalog
	Policy                  RuntimePolicy
	Now                     func() time.Time
	TelemetryServiceVersion string
}

// Manager is the immutable Jobs facade. Persistence and policy remain private
// package concerns; consumers receive narrower interfaces during assembly.
type Manager struct {
	pool                  *pgxpool.Pool
	now                   func() time.Time
	transactions          *TransactionService
	serviceVersion        string
	activeGaugeRegistered bool
	catalog               *Catalog
	policy                RuntimePolicy
}

func NewManager(options ManagerOptions) (*Manager, error) {
	if options.Postgres == nil || options.Transactions == nil || options.Catalog == nil {
		return nil, ErrNotConfigured
	}
	if options.Transactions.catalog != options.Catalog {
		return nil, fmt.Errorf("%w: transaction service catalog mismatch", ErrInvalidJobDefinition)
	}
	if err := options.Policy.validate(); err != nil {
		return nil, err
	}
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	manager := &Manager{
		pool:           options.Postgres,
		transactions:   options.Transactions,
		catalog:        options.Catalog,
		policy:         cloneRuntimePolicy(options.Policy),
		now:            now,
		serviceVersion: strings.TrimSpace(options.TelemetryServiceVersion),
	}
	manager.registerActiveGauge()
	return manager, nil
}

func (m *Manager) ensureConfigured() error {
	if m == nil || m.pool == nil || m.transactions == nil || m.catalog == nil || m.now == nil {
		return ErrNotConfigured
	}
	return m.policy.validate()
}
