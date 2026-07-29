package recovery

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type WorkbookProjectionQuery interface {
	QueryRows(context.Context, uuid.UUID, string, viewschema.QueryMeta) ([]map[string]any, error)
}

var ErrWorkbookProbeFailed = errors.New("recovery: workbook probe failed")

type RestoreVerificationWorkbookProbe struct {
	Postgres postgres.DB
	Query    WorkbookProjectionQuery
}

func (probe RestoreVerificationWorkbookProbe) ProbeRestoredBackup(ctx context.Context, _ RestoreResult) error {
	if probe.Postgres == nil {
		return fmt.Errorf("%w: restore verification workbook probe requires postgres", ErrWorkbookProbeFailed)
	}
	if probe.Query == nil {
		return fmt.Errorf("%w: restore verification workbook probe requires projection query", ErrWorkbookProbeFailed)
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
		return fmt.Errorf("%w: incident lookup: %v", ErrWorkbookProbeFailed, err)
	}
	schema, ok := viewschema.Lookup(RestoreVerificationTimelineViewID)
	if !ok {
		return fmt.Errorf("%w: timeline view schema is missing", ErrWorkbookProbeFailed)
	}
	if _, err := probe.Query.QueryRows(ctx, incidentID, RestoreVerificationTimelineViewID, schema.DefaultQueryMeta()); err != nil {
		return fmt.Errorf("%w: timeline query: %v", ErrWorkbookProbeFailed, err)
	}
	return nil
}
