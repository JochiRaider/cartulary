package networkflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
)

type stalledCleanupDatabase struct {
	postgres.DB
	tx *stalledCleanupTransaction
}

func (db stalledCleanupDatabase) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return db.tx, nil
}

type stalledCleanupTransaction struct {
	pgx.Tx
	deadline   time.Time
	rolledBack bool
}

func (tx *stalledCleanupTransaction) QueryRow(ctx context.Context, _ string, _ ...any) pgx.Row {
	return stalledCleanupRow{tx: tx, ctx: ctx}
}

type stalledCleanupRow struct {
	tx  *stalledCleanupTransaction
	ctx context.Context
}

func (row stalledCleanupRow) Scan(...any) error {
	var ok bool
	row.tx.deadline, ok = row.ctx.Deadline()
	if !ok {
		return errors.New("cleanup query has no deadline")
	}
	<-row.ctx.Done()
	return row.ctx.Err()
}
func (tx *stalledCleanupTransaction) Rollback(context.Context) error {
	tx.rolledBack = true
	return nil
}

func assertCleanupDatabaseDeadline(t *testing.T) {
	tx := &stalledCleanupTransaction{}
	db := stalledCleanupDatabase{tx: tx}
	service, err := newGraphResultCleanupService(db, &store{})
	if err != nil {
		t.Fatal(err)
	}
	service.maximumDuration = 10 * time.Millisecond
	started := time.Now()
	result, err := service.SweepGraphResults(context.Background(), started, nil)
	if !errors.Is(err, context.DeadlineExceeded) || tx.deadline.IsZero() || !tx.rolledBack || result.Examined != 0 || result.DeletedLeases != 0 || time.Since(started) > time.Second {
		t.Fatalf("stalled cleanup escaped deadline: %#v %v %#v", result, err, tx)
	}
}
