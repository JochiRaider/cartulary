package reference_data

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
)

const byteRetentionNamespace = "cartulary/reference-data/byte-retention/v1"

// AcquireBackupRetention must precede creation of the repeatable-read backup
// snapshot, and remain held through object streaming. Acquiring inside that
// snapshot could select rows from before a collector that the lock waited for.
func AcquireBackupRetention(ctx context.Context, db postgres.DB) (io.Closer, error) {
	return acquireByteRetention(ctx, db, false)
}

type byteRetentionLease struct {
	tx   pgx.Tx
	once sync.Once
	err  error
}

func (l *byteRetentionLease) Close() error {
	l.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		l.err = l.tx.Rollback(ctx)
		if errors.Is(l.err, pgx.ErrTxClosed) {
			l.err = nil
		}
	})
	return l.err
}
func acquireByteRetention(ctx context.Context, db postgres.DB, exclusive bool) (io.Closer, error) {
	if db == nil {
		return nil, errors.New("reference pack: retention database required")
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}
	lease := &byteRetentionLease{tx: tx}
	if err := lockByteRetentionTx(ctx, tx, exclusive); err != nil {
		return nil, errors.Join(err, lease.Close())
	}
	return lease, nil
}

// Writers register the guard in the transaction which can publish the byte
// reference. If commit acknowledgement is lost, PostgreSQL retains this guard
// until the mutation actually commits or aborts, even after the storage lease
// closes. Collection therefore never guesses whether that reference exists.
func lockByteRetentionTx(ctx context.Context, tx pgx.Tx, exclusive bool) error {
	query := `SELECT pg_advisory_xact_lock_shared(hashtextextended($1,0))`
	if exclusive {
		query = `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`
	}
	_, err := tx.Exec(ctx, query, byteRetentionNamespace)
	return err
}
