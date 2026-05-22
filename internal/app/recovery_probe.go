package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type RestoreVerificationWorkbookProbe struct {
	Postgres postgres.DB
}

func (probe RestoreVerificationWorkbookProbe) ProbeRestoredBackup(ctx context.Context, _ recovery.RestoreResult) error {
	if probe.Postgres == nil {
		return fmt.Errorf("restore verification workbook probe requires postgres")
	}
	var incidentID uuid.UUID
	if err := probe.Postgres.QueryRow(ctx, `
SELECT id
FROM incidents
ORDER BY created_at ASC, id ASC
LIMIT 1
`).Scan(&incidentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("restore verification workbook probe incident lookup: %w", err)
	}
	schema, ok := viewschema.Lookup(timeline.TimelineViewSchemaID)
	if !ok {
		return fmt.Errorf("restore verification workbook probe missing timeline view schema")
	}
	rows, err := workbook.NewStore(probe.Postgres).QueryRows(ctx, incidentID, timeline.TimelineViewSchemaID, schema.DefaultQueryMeta())
	if err != nil {
		return fmt.Errorf("restore verification workbook probe timeline query: %w", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("restore verification workbook probe timeline query returned no rows")
	}
	return nil
}
