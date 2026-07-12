// Package migrationtest owns database-contract-only migration application helpers.
package migrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// ApplyThrough applies the migration source through one positive Goose version.
// It is test support only and MUST NOT be exposed through the migrate executable.
func ApplyThrough(ctx context.Context, db *sql.DB, source postgres.MigrationSource, version int) (postgres.MigrationStatus, error) {
	if version < 1 {
		return postgres.MigrationStatus{}, fmt.Errorf("migration version must be positive")
	}
	return postgres.Migrate(ctx, db, source, "up-to", strconv.Itoa(version))
}
