package pgtest

import (
	"context"
	"database/sql"
	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
)

// Each fixture carries its own database operations; contract faults cannot alter
// another fixture or the process attachment registry.
type databaseOperations struct {
	source          func() (*database_migrations.Source, error)
	create          func(context.Context, string, string, string, func() error) (bool, error)
	drop            func(context.Context, string, string, func(postgrescleanup.DropAttempt)) error
	migrate         func(context.Context, *sql.DB, *database_migrations.Source) error
	openMigration   func(context.Context, string, postgres.Purpose) (*sql.DB, error)
	applyThrough    func(context.Context, *sql.DB, *database_migrations.Source, int64) error
	rollbackThrough func(context.Context, *sql.DB, *database_migrations.Source, int64) error
}

func defaultDatabaseOperations() databaseOperations {
	return databaseOperations{
		source:          dbmigrations.Source,
		create:          createDatabase,
		drop:            postgrescleanup.DropOwnedDatabaseWithObservation,
		migrate:         database_migrations.Apply,
		openMigration:   OpenPurposeDatabase,
		applyThrough:    applyCanonicalMigrationsThrough,
		rollbackThrough: rollbackCanonicalMigrationsThrough,
	}
}
func (h *Harness) databaseOperations() databaseOperations {
	if h.operations != nil {
		return *h.operations
	}
	return defaultDatabaseOperations()
}
