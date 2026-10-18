// Package postgrescatalog coordinates harness-owned PostgreSQL catalog
// mutations. Same-target work is exclusive; global catalog capacity belongs to
// the harness scheduler's typed PostgreSQL claims.
package postgrescatalog

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	catalogTargetLockIdentity  = "cartulary.test-harness.postgres-catalog-target.v1"
	catalogMutationUnlockLimit = 2 * time.Second
)

// MutationError preserves a failed operation's stage without requiring callers
// to interpret error text. The wrapped error remains available to errors.Is/As.
type MutationError struct {
	Stage string
	Err   error
}

func (e *MutationError) Error() string        { return e.Stage + ": " + e.Err.Error() }
func (e *MutationError) Unwrap() error        { return e.Err }
func (e *MutationError) CleanupStage() string { return e.Stage }

// WithMutation admits one target-exclusive catalog mutation. The admission
// budget ends when the target lock is acquired; the caller owns the separately
// bounded operation context used inside operation. Admission and mutation share
// one physical session, so cancellation cannot release exclusion through a
// different connection while the original backend is still executing.
func WithMutation(
	ctx context.Context,
	adminDSN string,
	resource string,
	admissionLimit time.Duration,
	operation func(*sql.Conn) error,
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
	operation func(*sql.Conn) error,
) (resultErr error) {

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return &MutationError{Stage: "catalog_open", Err: err}
	}
	defer func() {
		if closeErr := admin.Close(); closeErr != nil {
			resultErr = errors.Join(resultErr, &MutationError{Stage: "catalog_close", Err: closeErr})
		}
	}()
	admin.SetMaxOpenConns(1)

	admissionCtx, cancelAdmission := context.WithTimeout(ctx, admissionLimit)
	connection, err := admin.Conn(admissionCtx)
	if err != nil {
		cancelAdmission()
		return &MutationError{Stage: "catalog_connection", Err: err}
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
		return &MutationError{Stage: "catalog_lock", Err: err}
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
		if closeErr := connection.Close(); closeErr != nil && !retiredConnection(closeErr) {
			resultErr = errors.Join(resultErr, &MutationError{Stage: "catalog_close", Err: closeErr})
		}
	}()

	return operation(connection)
}

func retiredConnection(err error) bool {
	return errors.Is(err, sql.ErrConnDone) || errors.Is(err, driver.ErrBadConn)
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
		// A cancelled query can retire its client connection. The server then
		// releases session locks only when that backend physically settles.
		// The next target admission still waits on PostgreSQL; no fresh session
		// may unlock this one or infer that the cancelled mutation succeeded.
		if retiredConnection(err) {
			return nil
		}
		return &MutationError{Stage: "catalog_unlock", Err: err}
	}
	if !unlocked {
		return &MutationError{Stage: "catalog_unlock", Err: fmt.Errorf("session did not own catalog lock")}
	}
	return nil
}
