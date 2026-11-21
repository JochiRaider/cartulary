package dbassert

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// RoundTripCounter wraps a transaction and counts the client round trips its
// caller makes through it. One statement, one pipelined batch or one COPY is
// one round trip, so a set-based reader or writer shows a count bounded by
// pages rather than by records. Work done through Conn() is not counted.
type RoundTripCounter struct {
	pgx.Tx
	trips *int
}

// CountRoundTrips returns a counter over tx starting at zero.
func CountRoundTrips(tx pgx.Tx) *RoundTripCounter {
	return &RoundTripCounter{Tx: tx, trips: new(int)}
}

// RoundTrips reports the round trips made since the counter was created,
// including those made through nested savepoint transactions.
func (c *RoundTripCounter) RoundTrips() int { return *c.trips }

func (c *RoundTripCounter) Begin(ctx context.Context) (pgx.Tx, error) {
	*c.trips++
	nested, err := c.Tx.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &RoundTripCounter{Tx: nested, trips: c.trips}, nil
}

func (c *RoundTripCounter) Commit(ctx context.Context) error {
	*c.trips++
	return c.Tx.Commit(ctx)
}

func (c *RoundTripCounter) Rollback(ctx context.Context) error {
	*c.trips++
	return c.Tx.Rollback(ctx)
}

func (c *RoundTripCounter) CopyFrom(ctx context.Context, table pgx.Identifier, columns []string, source pgx.CopyFromSource) (int64, error) {
	*c.trips++
	return c.Tx.CopyFrom(ctx, table, columns, source)
}

func (c *RoundTripCounter) SendBatch(ctx context.Context, batch *pgx.Batch) pgx.BatchResults {
	*c.trips++
	return c.Tx.SendBatch(ctx, batch)
}

func (c *RoundTripCounter) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	*c.trips++
	return c.Tx.Prepare(ctx, name, sql)
}

func (c *RoundTripCounter) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	*c.trips++
	return c.Tx.Exec(ctx, sql, arguments...)
}

func (c *RoundTripCounter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	*c.trips++
	return c.Tx.Query(ctx, sql, args...)
}

func (c *RoundTripCounter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	*c.trips++
	return c.Tx.QueryRow(ctx, sql, args...)
}
