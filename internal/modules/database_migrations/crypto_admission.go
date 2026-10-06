package database_migrations

import (
	"context"
	"database/sql"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/database_migrations/sourcecatalog"
)

// ApplicationCryptoFormat identifies stored formats, independently of the binary's module version.
const ApplicationCryptoFormat = "cartulary.application_crypto_format.v1"

var ErrIncompatibleCryptoState = errors.New("application cryptographic state incompatible; provision a fresh target or use its matching release")

const cryptoIdentityPresentSQL = `SELECT to_regclass('public.application_crypto_format') IS NOT NULL`
const cryptoIdentityReadSQL = `SELECT count(*) = 1 AND COALESCE(bool_and(singleton AND format_id = $1), false) FROM public.application_crypto_format`

// RequireApplicationCryptoFormat inspects an already-open read-only capability.
// It never establishes identity, repairs a schema or acquires a serving lease.
func RequireApplicationCryptoFormat(ctx context.Context, reader LedgerReader) error {
	if ctx == nil || reader == nil {
		return ErrIncompatibleCryptoState
	}
	var present, current bool
	if err := reader.QueryRow(ctx, cryptoIdentityPresentSQL).Scan(&present); err != nil || !present {
		return ErrIncompatibleCryptoState
	}
	if err := reader.QueryRow(ctx, cryptoIdentityReadSQL, ApplicationCryptoFormat).Scan(&current); err != nil || !current {
		return ErrIncompatibleCryptoState
	}
	return nil
}

// ApplyWithCryptoAdmission holds the migration exclusion boundary on a dedicated
// admitted handle across Goose discovery, application, postconditions and identity
// commit. The work and admission handles must be distinct pools for the same
// database; keeping work separate prevents waiting initializers exhausting it.
// Both handles are borrowed. PostgreSQL connectivity remains caller-owned.
func ApplyWithCryptoAdmission(ctx context.Context, db, admissionDB *sql.DB, source *Source, inspectFreshDependencies func(context.Context) error) (retErr error) {
	if ctx == nil {
		return newMigrationFailure(reasonMigrationContextInvalid, errNilMigrateContext)
	}
	if err := ctx.Err(); err != nil {
		return newMigrationFailure(reasonSchemaMigrationExecutionFailed, err)
	}
	if err := validateSource(source); err != nil {
		return newMigrationFailure(reasonMigrationSourceInvalid, err)
	}
	if db == nil || admissionDB == nil || db == admissionDB || inspectFreshDependencies == nil {
		return newMigrationFailure(reasonMigrationDatabaseUnavailable, nil)
	}
	delegate, err := sourcecatalog.NewInitializationLocker()
	if err != nil {
		return newMigrationFailure(reasonMigrationLockAcquisitionFailed, err)
	}
	conn, err := admissionDB.Conn(ctx)
	if err != nil {
		return newMigrationFailure(reasonMigrationDatabaseUnavailable, err)
	}
	defer func() {
		if err := conn.Close(); retErr == nil && err != nil {
			retErr = newMigrationFailure(reasonSchemaMigrationCleanupFailed, err)
		}
	}()
	fresh := false
	guard := &validatingSessionLocker{delegate: delegate, validate: func(ctx context.Context, conn *sql.Conn) error {
		if err := sameCryptoTarget(ctx, conn, db); err != nil {
			return err
		}
		var err error
		fresh, err = inspectCryptoInitialization(ctx, conn)
		if err != nil {
			return err
		}
		_, err = migrationWorkNeeded(ctx, conn, source)
		return err
	}}
	if err := guard.SessionLock(ctx, conn); err != nil {
		return err
	}
	defer func() {
		if recover() != nil {
			retErr = newMigrationFailure(reasonSchemaMigrationExecutionFailed, nil)
		}
		if err := guard.SessionUnlock(ctx, conn); retErr == nil {
			retErr = err
		}
	}()
	// Establish the ledger on the admission session before handing work to the
	// provider. Lost admission ownership now leaves retained, incompatible state;
	// it cannot make the target appear empty to another initializer.
	if fresh {
		// Application composition supplies read-only inspection of its storage
		// bindings. Recheck inside exclusion, before even provider discovery writes.
		if err := inspectFreshDependencies(ctx); err != nil {
			return ErrIncompatibleCryptoState
		}
		if err := sourcecatalog.InitializeLedger(ctx, conn); err != nil {
			return newMigrationFailure(reasonSchemaMigrationExecutionFailed, err)
		}
	}
	// Keep the provider's own session lock and postconditions. It serializes work
	// even if the separate admission connection is lost during an operation.
	if err := Apply(ctx, db, source); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return newMigrationFailure(reasonSchemaMigrationExecutionFailed, err)
	}
	if err := verifyMigrationPostcondition(ctx, conn, source); err != nil {
		return err
	}
	if fresh {
		if err := inspectFreshDependencies(ctx); err != nil {
			return ErrIncompatibleCryptoState
		}
		if _, err := conn.ExecContext(ctx, `INSERT INTO public.application_crypto_format (singleton,format_id) VALUES (true,$1)`, ApplicationCryptoFormat); err != nil {
			return newMigrationFailure(reasonSchemaMigrationPostcondition, err)
		}
	}
	return nil
}

func inspectCryptoInitialization(ctx context.Context, conn *sql.Conn) (bool, error) {
	var present bool
	if err := conn.QueryRowContext(ctx, cryptoIdentityPresentSQL).Scan(&present); err != nil {
		return false, ErrIncompatibleCryptoState
	}
	if present {
		var current bool
		if err := conn.QueryRowContext(ctx, cryptoIdentityReadSQL, ApplicationCryptoFormat).Scan(&current); err != nil || !current {
			return false, ErrIncompatibleCryptoState
		}
		return false, nil
	}
	var ledger bool
	if err := conn.QueryRowContext(ctx, `SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&ledger); err != nil || ledger {
		return false, ErrIncompatibleCryptoState
	}
	contaminated, err := databaseHasPreexistingObjects(ctx, conn)
	if err != nil || contaminated {
		return false, ErrIncompatibleCryptoState
	}
	return true, nil
}

func sameCryptoTarget(ctx context.Context, conn *sql.Conn, db *sql.DB) error {
	const query = `SELECT current_database(), current_user, COALESCE(inet_server_addr()::text,''), COALESCE(inet_server_port(),0)`
	var left, right [4]string
	if err := conn.QueryRowContext(ctx, query).Scan(&left[0], &left[1], &left[2], &left[3]); err != nil {
		return ErrIncompatibleCryptoState
	}
	if err := db.QueryRowContext(ctx, query).Scan(&right[0], &right[1], &right[2], &right[3]); err != nil || left != right {
		return ErrIncompatibleCryptoState
	}
	return nil
}
