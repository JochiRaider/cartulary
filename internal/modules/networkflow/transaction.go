package networkflow

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func withinTransaction(ctx context.Context, db postgres.DB, options pgx.TxOptions, operation func(pgx.Tx) error) error {
	if db == nil {
		return fmt.Errorf("network flow PostgreSQL transaction unavailable")
	}
	tx, err := db.BeginTx(ctx, options)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := operation(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
