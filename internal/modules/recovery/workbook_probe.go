package recovery

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type RestoreVerificationWorkbookProbe struct {
	Postgres postgres.DB
}

func (probe RestoreVerificationWorkbookProbe) ProbeRestoredBackup(ctx context.Context, _ RestoreResult) error {
	if probe.Postgres == nil {
		return fmt.Errorf("restore verification workbook probe requires postgres")
	}
	var incidentID uuid.UUID
	if err := probe.Postgres.QueryRow(ctx, `
SELECT id
FROM incidents
ORDER BY id::text ASC
LIMIT 1
`).Scan(&incidentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("restore verification workbook probe incident lookup: %w", err)
	}
	schema, ok := viewschema.Lookup(projectionadapters.TimelineViewSchemaID)
	if !ok {
		return fmt.Errorf("restore verification workbook probe missing timeline view schema")
	}
	if _, err := projectionadapters.NewWorkbookRows(probe.Postgres).QueryRows(ctx, incidentID, projectionadapters.TimelineViewSchemaID, schema.DefaultQueryMeta()); err != nil {
		return fmt.Errorf("restore verification workbook probe timeline query: %w", err)
	}
	return nil
}
