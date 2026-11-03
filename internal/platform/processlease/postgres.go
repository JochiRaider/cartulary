package processlease

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	applicationProcessAdvisoryKey int64 = 4850189438622597893
	ServingAdvisoryKey            int64 = 4850189438622597896
)

var ErrRecoveryTargetNotStopped = errors.New("recovery target is not stopped")

// RequireRecoveryTargetStopped is a read-only preflight observation. The
// exclusive Recovery serving lease must still be acquired and proved before
// mutation; this observation never substitutes for that exclusion.
func RequireRecoveryTargetStopped(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("recovery target pool is required")
	}
	for _, key := range []int64{applicationProcessAdvisoryKey, ServingAdvisoryKey} {
		var held bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS (
 SELECT 1 FROM pg_catalog.pg_locks
 WHERE locktype='advisory' AND granted AND objsubid=1
 AND database=(SELECT oid FROM pg_catalog.pg_database WHERE datname=current_database())
 AND classid::bigint=$1 AND objid::bigint=$2
)`, int64(uint64(key)>>32), int64(uint64(key)&0xffffffff)).Scan(&held); err != nil {
			return err
		}
		if held {
			return ErrRecoveryTargetNotStopped
		}
	}
	return nil
}

type LockMode string

const (
	LockExclusive LockMode = "exclusive"
	LockShared    LockMode = "shared"
)

type PostgresBackend struct {
	Pool        *pgxpool.Pool
	AdvisoryKey int64
	Purpose     string
	Mode        LockMode
}

// ApplicationProcessLease is the typed application-owned, deployment-global
// exclusive lease. Its concrete session mechanics remain private to this
// package so callers cannot accidentally change its lock identity or mode.
type ApplicationProcessLease struct {
	*Lease
}

// ApplicationRecoveryServingLease is the typed application side of the
// Recovery serving exclusion. Recovery owns the exclusive counterpart.
type ApplicationRecoveryServingLease struct {
	*Lease
}

// RecoveryTargetLease is the Recovery side of the serving exclusion. It is
// exclusive with every ApplicationRecoveryServingLease for the same database.
type RecoveryTargetLease struct {
	*Lease
}

func AcquireApplicationProcess(
	ctx context.Context,
	pool *pgxpool.Pool,
	acquireTimeout time.Duration,
	lossDetection time.Duration,
) (*ApplicationProcessLease, error) {
	lease, err := Acquire(
		ctx,
		PostgresBackend{Pool: pool},
		acquireTimeout,
		lossDetection,
	)
	if err != nil {
		return nil, err
	}
	return &ApplicationProcessLease{Lease: lease}, nil
}

func AcquireApplicationRecoveryServing(
	ctx context.Context,
	pool *pgxpool.Pool,
	acquireTimeout time.Duration,
	lossDetection time.Duration,
) (*ApplicationRecoveryServingLease, error) {
	lease, err := Acquire(
		ctx,
		PostgresBackend{
			Pool:        pool,
			AdvisoryKey: ServingAdvisoryKey,
			Purpose:     "application recovery serving",
			Mode:        LockShared,
		},
		acquireTimeout,
		lossDetection,
	)
	if err != nil {
		if errors.Is(err, ErrApplicationProcessActive) {
			return nil, ErrRecoveryServingLeaseActive
		}
		return nil, err
	}
	return &ApplicationRecoveryServingLease{Lease: lease}, nil
}

func AcquireRecoveryTargetLease(
	ctx context.Context,
	pool *pgxpool.Pool,
	acquireTimeout time.Duration,
	lossDetection time.Duration,
) (*RecoveryTargetLease, error) {
	lease, err := Acquire(
		ctx,
		PostgresBackend{
			Pool:        pool,
			AdvisoryKey: ServingAdvisoryKey,
			Purpose:     "recovery target",
			Mode:        LockExclusive,
		},
		acquireTimeout,
		lossDetection,
	)
	if err != nil {
		return nil, err
	}
	return &RecoveryTargetLease{Lease: lease}, nil
}

func (b PostgresBackend) Open(ctx context.Context) (Session, error) {
	if b.Pool == nil {
		return nil, fmt.Errorf("%s lease requires postgres", b.purpose())
	}
	mode := b.mode()
	if mode != LockExclusive && mode != LockShared {
		return nil, fmt.Errorf("%s lease has unsupported lock mode %q", b.purpose(), mode)
	}
	connection, err := b.Pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("open %s lease session: %w", b.purpose(), err)
	}
	var backendPID int64
	if err := connection.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&backendPID); err != nil {
		connection.Release()
		return nil, fmt.Errorf("identify %s lease session: %w", b.purpose(), err)
	}
	return &postgresSession{
		operations:  make(chan struct{}, 1),
		connection:  connection,
		identity:    strconv.FormatInt(backendPID, 10),
		advisoryKey: b.advisoryKey(),
		purpose:     b.purpose(),
		mode:        mode,
	}, nil
}

func (b PostgresBackend) advisoryKey() int64 {
	if b.AdvisoryKey == 0 {
		return applicationProcessAdvisoryKey
	}
	return b.AdvisoryKey
}

func (b PostgresBackend) purpose() string {
	if b.Purpose == "" {
		return "application process"
	}
	return b.Purpose
}

func (b PostgresBackend) mode() LockMode {
	if b.Mode == "" {
		return LockExclusive
	}
	return b.Mode
}

type postgresSession struct {
	operations  chan struct{}
	connection  *pgxpool.Conn
	identity    string
	advisoryKey int64
	purpose     string
	mode        LockMode
	acquired    bool
	closed      bool
}

func (s *postgresSession) Identity() string { return s.identity }

func (s *postgresSession) enter(ctx context.Context) error {
	if s == nil {
		return ErrLeaseLost
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.operations <- struct{}{}:
		return nil
	}
}

func (s *postgresSession) leave() { <-s.operations }

func (s *postgresSession) TryAcquire(ctx context.Context) (bool, error) {
	if err := s.enter(ctx); err != nil {
		return false, err
	}
	defer s.leave()
	if s == nil || s.connection == nil || s.closed {
		return false, fmt.Errorf("%s lease session is closed", s.purpose)
	}
	if s.acquired {
		return true, nil
	}
	query := "SELECT pg_try_advisory_lock($1)"
	if s.mode == LockShared {
		query = "SELECT pg_try_advisory_lock_shared($1)"
	}
	if err := s.connection.QueryRow(ctx, query, s.advisoryKey).Scan(&s.acquired); err != nil {
		return false, err
	}
	return s.acquired, nil
}

func (s *postgresSession) Prove(ctx context.Context) Proof {
	if err := s.enter(ctx); err != nil {
		return ProofUncertain
	}
	defer s.leave()
	return s.proveLocked(ctx)
}

func (s *postgresSession) proveLocked(ctx context.Context) Proof {
	if s == nil || s.connection == nil || s.closed || !s.acquired || s.connection.Conn().IsClosed() {
		return ProofLost
	}
	var backendPID int64
	if err := s.connection.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&backendPID); err != nil {
		if s.connection.Conn().IsClosed() {
			return ProofLost
		}
		return ProofUncertain
	}
	if strconv.FormatInt(backendPID, 10) != s.identity {
		return ProofLost
	}
	return ProofContinuous
}

func (s *postgresSession) Release(ctx context.Context) error {
	if err := s.enter(ctx); err != nil {
		return err
	}
	defer s.leave()
	if s == nil || s.connection == nil || s.closed || !s.acquired {
		return ErrInvalidTransition
	}
	var released bool
	query := "SELECT pg_advisory_unlock($1)"
	if s.mode == LockShared {
		query = "SELECT pg_advisory_unlock_shared($1)"
	}
	if err := s.connection.QueryRow(ctx, query, s.advisoryKey).Scan(&released); err != nil {
		return err
	}
	if !released {
		return ErrLeaseLost
	}
	s.acquired = false
	return nil
}

func (s *postgresSession) Close() {
	if s == nil {
		return
	}
	if err := s.enter(context.Background()); err != nil {
		return
	}
	defer s.leave()
	if s == nil || s.closed {
		return
	}
	s.closed = true
	if s.connection != nil {
		if s.acquired {
			connection := s.connection.Hijack()
			// An unreleased session cannot return to the pool. Close its transport
			// without extending the caller's already bounded settlement budget.
			closeContext, cancel := context.WithCancel(context.Background())
			cancel()
			_ = connection.Close(closeContext)
		} else {
			s.connection.Release()
		}
	}
}
