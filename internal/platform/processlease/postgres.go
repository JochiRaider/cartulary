package processlease

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	applicationProcessAdvisoryKey int64 = 4850189438622597893
	ServingAdvisoryKey            int64 = 4850189438622597896
)

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
	connection  *pgxpool.Conn
	identity    string
	advisoryKey int64
	purpose     string
	mode        LockMode
	acquired    bool
	closed      bool
}

func (s *postgresSession) Identity() string { return s.identity }

func (s *postgresSession) TryAcquire(ctx context.Context) (bool, error) {
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
	if s == nil || s.closed {
		return
	}
	s.closed = true
	if s.connection != nil {
		if s.acquired {
			connection := s.connection.Hijack()
			_ = connection.Close(context.Background())
		} else {
			s.connection.Release()
		}
	}
}
