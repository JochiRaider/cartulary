package database_migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// A session lock must be released on the exact connection that acquired it.
// The returned cleanup is idempotent so normal assertions and early exits share it.
func migrationFixtureUnlock(conn *sql.Conn, key int64) func() error {
	locked := true
	return func() error {
		if !locked {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var released bool
		if err := conn.QueryRowContext(ctx, `SELECT pg_advisory_unlock($1)`, key).Scan(&released); err != nil {
			return err
		}
		if !released {
			return fmt.Errorf("fixture session did not own advisory lock %d", key)
		}
		locked = false
		return nil
	}
}
