// Package postgrescleanup contains cross-process coordination for destructive
// PostgreSQL test-resource cleanup. Callers remain responsible for proving
// exact ownership before invoking these mechanics.
package postgrescleanup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/testutil/postgrescatalog"
)

const databaseDropAdmissionLimit = 15 * time.Second

type databaseDropExecutor func(context.Context, string, bool) error
type databaseDropLock func(context.Context, string, string, func(databaseDropExecutor)) error

type dropOperations struct {
	withNormalAdmission databaseDropLock
	withForcedAdmission databaseDropLock
}

var productionDropOperations = dropOperations{
	withForcedAdmission: func(ctx context.Context, adminDSN string, name string, operation func(databaseDropExecutor)) error {
		return postgrescatalog.WithMutation(
			ctx,
			adminDSN,
			name,
			databaseDropAdmissionLimit,
			func(admin *sql.DB) error {
				operation(func(operationCtx context.Context, database string, force bool) error {
					return executeDatabaseDrop(operationCtx, admin, database, force)
				})
				return nil
			},
		)
	},
	withNormalAdmission: func(ctx context.Context, adminDSN string, name string, operation func(databaseDropExecutor)) error {
		return postgrescatalog.WithMutation(
			ctx,
			adminDSN,
			name,
			databaseDropAdmissionLimit,
			func(admin *sql.DB) error {
				operation(func(operationCtx context.Context, database string, force bool) error {
					return executeDatabaseDrop(operationCtx, admin, database, force)
				})
				return nil
			},
		)
	},
}

// DropOwnedDatabase attempts an ordinary drop with its own budget and uses one
// separately bounded atomic forced drop only for an active database or an
// expired ordinary operation. Each attempt independently acquires target and
// stripe admission. The caller must prove exact ownership first.
func DropOwnedDatabase(
	ctx context.Context,
	adminDSN string,
	name string,
	normalLimit time.Duration,
	forcedLimit time.Duration,
) (bool, error) {
	return dropOwnedDatabase(ctx, adminDSN, name, normalLimit, forcedLimit, productionDropOperations)
}

func dropOwnedDatabase(
	ctx context.Context,
	adminDSN string,
	name string,
	normalLimit time.Duration,
	forcedLimit time.Duration,
	operations dropOperations,
) (forced bool, resultErr error) {
	if err := validateDatabaseIdentifier(name); err != nil {
		return false, err
	}
	if adminDSN == "" {
		return false, errors.New("database drop requires an administrator DSN")
	}
	if normalLimit <= 0 || forcedLimit <= 0 {
		return false, errors.New("database drop requires positive normal and forced limits")
	}

	var normalErr error
	normalAdmissionErr := operations.withNormalAdmission(ctx, adminDSN, name, func(drop databaseDropExecutor) {
		normalCtx, cancelNormal := context.WithTimeout(ctx, normalLimit)
		normalErr = drop(normalCtx, name, false)
		cancelNormal()
	})
	if normalAdmissionErr != nil {
		return false, errors.Join(normalErr, normalAdmissionErr)
	}
	if normalErr == nil {
		return false, nil
	}
	if ctx.Err() != nil {
		return false, normalErr
	}
	var pgErr *pgconn.PgError
	normalTimedOut := errors.Is(normalErr, context.DeadlineExceeded)
	if (!errors.As(normalErr, &pgErr) || pgErr.Code != "55006") && !normalTimedOut {
		return false, normalErr
	}

	forced = true
	var forcedErr error
	forcedAdmissionErr := operations.withForcedAdmission(ctx, adminDSN, name, func(drop databaseDropExecutor) {
		forcedCtx, cancelForced := context.WithTimeout(ctx, forcedLimit)
		defer cancelForced()
		forcedErr = drop(forcedCtx, name, true)
	})
	return forced, errors.Join(forcedErr, forcedAdmissionErr)
}

// ForceDropDatabase coordinates one atomic forced drop. It exists for
// explicitly owned fixture paths whose lifecycle has no ordinary borrower
// phase; per-test database finalizers use DropOwnedDatabase instead.
func ForceDropDatabase(ctx context.Context, adminDSN string, name string) error {
	if err := validateDatabaseIdentifier(name); err != nil {
		return err
	}
	if adminDSN == "" {
		return errors.New("database drop requires an administrator DSN")
	}
	var dropErr error
	lockErr := productionDropOperations.withForcedAdmission(ctx, adminDSN, name, func(drop databaseDropExecutor) {
		dropErr = drop(ctx, name, true)
	})
	return errors.Join(dropErr, lockErr)
}

func executeDatabaseDrop(ctx context.Context, admin *sql.DB, name string, force bool) error {
	statement := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, name)
	action := "drop"
	if force {
		statement += " WITH (FORCE)"
		action = "force drop"
	}
	if _, err := admin.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("%s postgres database %s: %w", action, name, err)
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
