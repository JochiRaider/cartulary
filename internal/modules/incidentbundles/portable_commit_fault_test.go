package incidentbundles_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// This decorator affects only the final portable-catalog transaction in a
// disposable test database. Admission, verification, and terminal-abort writes
// retain their normal transaction behavior.
type portableCommitFaultPool struct {
	postgres.AdmittedPool
	rollback bool
	injected atomic.Int32
}

func (p *portableCommitFaultPool) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := p.AdmittedPool.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return &portableCommitFaultTx{Tx: tx, pool: p}, nil
}

type portableCommitFaultTx struct {
	pgx.Tx
	pool        *portableCommitFaultPool
	publication bool
}

func (t *portableCommitFaultTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	result, err := t.Tx.Exec(ctx, sql, args...)
	if err == nil && strings.Contains(sql, "INSERT INTO reference_pack_portable_catalogs") {
		t.publication = true
	}
	return result, err
}

func (t *portableCommitFaultTx) Commit(ctx context.Context) error {
	if !t.publication {
		return t.Tx.Commit(ctx)
	}
	t.pool.injected.Add(1)
	if t.pool.rollback {
		if err := t.Tx.Rollback(context.WithoutCancel(ctx)); err != nil {
			return err
		}
		return pgx.ErrTxCommitRollback
	}
	if err := t.Tx.Commit(ctx); err != nil {
		return err
	}
	return errors.New("fixture: parent commit acknowledgement lost")
}
