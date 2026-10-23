// Package postgrescleanup contains cross-process coordination for destructive
// PostgreSQL test-resource cleanup. Callers remain responsible for proving
// exact ownership before invoking these mechanics.
package postgrescleanup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/testutil/postgrescatalog"
)

const (
	databaseDropAdmissionLimit = 15 * time.Second
	databaseDropOwnerLimit     = 35 * time.Second
)

type databaseDropExecutor func(context.Context, string) error
type databaseDropLock func(context.Context, string, string, func(databaseDropExecutor) error) error

type dropOperations struct {
	withAdmission databaseDropLock
	observe       func(DropAttempt)
}

var productionDropOperations = dropOperations{withAdmission: withDropAdmission}

func withDropAdmission(ctx context.Context, adminDSN, name string, operation func(databaseDropExecutor) error) error {
	config, err := pgx.ParseConfig(adminDSN)
	if err != nil {
		return &postgrescatalog.MutationError{Stage: "catalog_open", Err: err}
	}
	var participant atomic.Int64
	config.RuntimeParams["client_min_messages"] = "log"
	config.OnNotice = func(_ *pgconn.PgConn, notice *pgconn.Notice) {
		if pid := barrierNoticePID(notice); pid > 0 {
			participant.Store(pid)
		}
	}
	// pgx owns this short-lived connection configuration token. It is never a
	// retained identity and is unregistered after WithMutation closes its pool.
	registered := stdlib.RegisterConnConfig(config)
	defer stdlib.UnregisterConnConfig(registered)
	return postgrescatalog.WithMutation(ctx, registered, name, databaseDropAdmissionLimit, func(admin *sql.Conn) error {
		return operation(func(operationCtx context.Context, database string) error {
			return executeObservedDatabaseDrop(operationCtx, admin, adminDSN, database, participant.Load)
		})
	})
}

// DropOwnedDatabase performs one bounded atomic forced drop after target
// admission. The caller must prove exact ownership and stop its runtime, jobs,
// pools and handles first. Admission, drop and settlement consume one bounded
// owner lifetime, clamped by a shorter parent deadline.
// A timeout is uncertain settlement, never authority to retry or release proof.
func DropOwnedDatabase(ctx context.Context, adminDSN, name string) error {
	return dropOwnedDatabase(ctx, adminDSN, name, productionDropOperations)
}

// DropOwnedDatabaseWithObservation publishes the same bounded attempt on
// success and failure. Observation grants no authority and adds no SQL probes.
func DropOwnedDatabaseWithObservation(ctx context.Context, adminDSN, name string, observe func(DropAttempt)) error {
	operations := productionDropOperations
	operations.observe = observe
	return dropOwnedDatabase(ctx, adminDSN, name, operations)
}

func dropOwnedDatabase(ctx context.Context, adminDSN, name string, operations dropOperations) (resultErr error) {
	if ctx == nil {
		return errors.New("database drop requires a parent context")
	}
	if err := validateDatabaseIdentifier(name); err != nil {
		return err
	}
	if adminDSN == "" {
		return errors.New("database drop requires an administrator DSN")
	}
	// Bound even a longer parent, and never restart this budget after admission.
	// The catalog owner also clamps lock settlement to this exact deadline.
	ctx, cancel := context.WithTimeout(ctx, databaseDropOwnerLimit)
	defer cancel()
	started := time.Now()
	observation := DropAttempt{Admission: "not_attempted", AdmissionBudgetMS: effectiveBudget(ctx, databaseDropAdmissionLimit).Milliseconds(), Wait: DropObservation{Status: "not_observed"}}
	defer func() {
		observation.DurationMS = time.Since(started).Milliseconds()
		observation.Outcome = "completed"
		defer func() {
			if operations.observe != nil {
				operations.observe(observation)
			}
		}()
		if resultErr == nil {
			return
		}
		observation.Outcome = "failed"
		if errors.Is(resultErr, context.DeadlineExceeded) {
			observation.Outcome = "deadline_exceeded"
		} else if errors.Is(resultErr, context.Canceled) {
			observation.Outcome = "cancelled"
		}
		resultErr = &attemptError{err: resultErr, observation: observation}
	}()
	if err := ctx.Err(); err != nil {
		return err
	}
	observation.Admission = "rejected"
	defer func() {
		if observation.Admission != "acquired" {
			observation.AdmissionMS = time.Since(started).Milliseconds()
		}
	}()
	return operations.withAdmission(ctx, adminDSN, name, func(drop databaseDropExecutor) error {
		observation.Admission = "acquired"
		observation.AdmissionMS = time.Since(started).Milliseconds()
		operationStarted := time.Now()
		defer func() { observation.OperationMS = time.Since(operationStarted).Milliseconds() }()
		observation.OperationBudgetMS = effectiveBudget(ctx, databaseDropOwnerLimit).Milliseconds()
		if err := ctx.Err(); err != nil {
			return err
		}
		err := drop(ctx, name)
		var observed *observedDropError
		if errors.As(err, &observed) {
			observation.Wait = observed.observation
			observation.ServiceGeneration = observed.observation.ServiceGeneration
		}
		return err
	})
}

func executeDatabaseDrop(ctx context.Context, admin *sql.Conn, name string) error {
	statement := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, name)
	if _, err := admin.ExecContext(ctx, statement); err != nil {
		return &postgrescatalog.MutationError{Stage: "forced_drop", Err: err}
	}
	return nil
}

func validateDatabaseIdentifier(name string) error {
	if name == "" || len(name) > 63 {
		return fmt.Errorf("invalid database-drop identifier %q", name)
	}
	for _, character := range name {
		if character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9' ||
			character == '_' {
			continue
		}
		return fmt.Errorf("invalid database-drop identifier %q", name)
	}
	return nil
}
