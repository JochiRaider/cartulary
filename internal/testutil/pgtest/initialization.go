package pgtest

import (
	"context"
	"database/sql"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// InitializeFreshDatabase uses production admission for a newly created,
// database-only fixture. Its owner has not attached object or filesystem state.
// The work handle is borrowed; the independent exclusion handle is owned here.
// Historical migration-scratch fixtures deliberately use the lower-level engine.
func InitializeFreshDatabase(ctx context.Context, db *sql.DB, dsn string, source *database_migrations.Source) error {
	guard, err := OpenPurposeDatabase(dsn, postgres.PurposeMigration)
	if err != nil {
		return err
	}
	defer guard.Close()
	return database_migrations.ApplyWithCryptoAdmission(ctx, db, guard, source, func(context.Context) error { return nil })
}
