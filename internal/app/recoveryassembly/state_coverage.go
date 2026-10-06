package recoveryassembly

import (
	"context"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func ValidateRecoveryStateDatabaseCoverage(
	ctx context.Context,
	pool application.PostgresPool,
	catalog *recoverystate.Catalog,
) error {
	if pool == nil {
		return fmt.Errorf("%w: database is unavailable", recoverystate.ErrInvalidCatalog)
	}
	rows, err := pool.Query(ctx, `
SELECT c.relname
FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p', 'f')
ORDER BY c.relname ASC
`)
	if err != nil {
		return fmt.Errorf("list database tables for recovery state coverage: %w", err)
	}
	defer rows.Close()
	tableNames := make([]string, 0, recoverystate.AuthoredTableCount+1)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("scan database table for recovery state coverage: %w", err)
		}
		tableNames = append(tableNames, tableName)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate database tables for recovery state coverage: %w", err)
	}
	return catalog.ValidateDatabaseTableNames(tableNames)
}
