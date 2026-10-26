package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// InitializeFreshDatabase uses production admission for a newly created,
// database-only fixture. Its owner has not attached object or filesystem state.
// The work handle is borrowed; the independent exclusion handle is owned here.
// Historical migration-scratch fixtures deliberately use the lower-level engine.
func InitializeFreshDatabase(ctx context.Context, db *sql.DB, dsn string, source *database_migrations.Source) (resultErr error) {
	return initializeFreshDatabase(ctx, db, dsn, source, OpenPurposeDatabase)
}

func initializeFreshDatabase(ctx context.Context, db *sql.DB, dsn string, source *database_migrations.Source, open func(context.Context, string, postgres.Purpose) (*sql.DB, error)) (resultErr error) {
	if ctx == nil {
		return errors.New("postgres initialization requires a context")
	}
	guard, err := open(ctx, dsn, postgres.PurposeMigration)
	if err != nil {
		return err
	}
	defer func() {
		if err := guard.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close migration admission handle: %w", err))
		}
	}()
	return database_migrations.ApplyWithCryptoAdmission(ctx, db, guard, source, func(context.Context) error { return nil })
}
