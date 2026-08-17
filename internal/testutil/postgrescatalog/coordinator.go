// Package postgrescatalog coordinates harness-owned PostgreSQL catalog
// mutations. Same-target work is exclusive; global catalog capacity belongs to
// the harness scheduler's typed PostgreSQL claims.
package postgrescatalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	catalogTargetLockIdentity  = "cartulary.test-harness.postgres-catalog-target.v1"
	catalogMutationUnlockLimit = 2 * time.Second
)

// WithMutation admits one target-exclusive catalog mutation. The admission
// budget ends when the target lock is acquired; the caller owns the separately
// bounded operation context used inside operation.
func WithMutation(
	ctx context.Context,
	adminDSN string,
	resource string,
	admissionLimit time.Duration,
	operation func(*sql.DB) error,
) error {
	if ctx == nil {
		return errors.New("postgres catalog mutation requires a parent context")
	}
	if adminDSN == "" || resource == "" {
		return errors.New("postgres catalog mutation identity is incomplete")
	}
	if admissionLimit <= 0 {
		return errors.New("postgres catalog mutation requires a positive admission limit")
	}
	if operation == nil {
		return errors.New("postgres catalog mutation requires an operation")
	}
	return withMutation(ctx, adminDSN, resource, admissionLimit, operation)
}

func withMutation(
	ctx context.Context,
	adminDSN string,
	resource string,
	admissionLimit time.Duration,
	operation func(*sql.DB) error,
) (resultErr error) {

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("open postgres catalog coordination handle: %w", err)
	}
	defer func() {
		if closeErr := admin.Close(); closeErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close postgres catalog coordination handle for %s: %w", resource, closeErr))
		}
	}()
	admin.SetMaxOpenConns(2)

	admissionCtx, cancelAdmission := context.WithTimeout(ctx, admissionLimit)
	connection, err := admin.Conn(admissionCtx)
	if err != nil {
		cancelAdmission()
		return fmt.Errorf("acquire postgres catalog coordination connection for %s: %w", resource, err)
	}
	targetLockStatement := `SELECT pg_catalog.pg_advisory_lock(pg_catalog.hashtext($1), pg_catalog.hashtext($2))`
	if _, err := connection.ExecContext(
		admissionCtx,
		targetLockStatement,
		catalogTargetLockIdentity,
		resource,
	); err != nil {
		cancelAdmission()
		_ = connection.Close()
		return fmt.Errorf("acquire postgres catalog target lock for %s: %w", resource, err)
	}
	cancelAdmission()

	defer func() {
		resultErr = errors.Join(resultErr, releaseCatalogLock(
			connection,
			resource,
			"target",
			`SELECT pg_catalog.pg_advisory_unlock(pg_catalog.hashtext($1), pg_catalog.hashtext($2))`,
			catalogTargetLockIdentity,
			resource,
		))
		if closeErr := connection.Close(); closeErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close postgres catalog lock connection for %s: %w", resource, closeErr))
		}
	}()

	return operation(admin)
}

func releaseCatalogLock(
	connection *sql.Conn,
	resource string,
	kind string,
	statement string,
	arguments ...any,
) error {
	unlockCtx, cancelUnlock := context.WithTimeout(context.Background(), catalogMutationUnlockLimit)
	defer cancelUnlock()
	var unlocked bool
	if err := connection.QueryRowContext(unlockCtx, statement, arguments...).Scan(&unlocked); err != nil {
		return fmt.Errorf("release postgres catalog %s lock for %s: %w", kind, resource, err)
	}
	if !unlocked {
		return fmt.Errorf("release postgres catalog %s lock for %s: session did not own lock", kind, resource)
	}
	return nil
}
