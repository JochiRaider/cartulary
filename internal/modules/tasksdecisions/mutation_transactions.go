package tasksdecisions

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
)

func (f *MutationFacade) touchSourceRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID, now time.Time) error {
	return tasksource.TouchSourceRowTx(ctx, tx, f.catalog, viewSchemaID, recordID, now)
}

func (f *MutationFacade) refreshRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		return f.projectionRows.RefreshTaskRequestTx(ctx, tx, recordID)
	case DecisionsViewSchemaID:
		return f.projectionRows.RefreshDecisionTx(ctx, tx, recordID)
	default:
		return &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
}

func (f *MutationFacade) loadProjectionRowTx(
	ctx context.Context,
	tx pgx.Tx,
	viewSchemaID string,
	recordID uuid.UUID,
) (map[string]any, error) {
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		return f.projectionRows.LoadTaskRequestTx(ctx, tx, recordID)
	case DecisionsViewSchemaID:
		return f.projectionRows.LoadDecisionTx(ctx, tx, recordID)
	default:
		return nil, &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
}

func touchesField(changes []PatchChange, fieldKey string) bool {
	for _, change := range changes {
		if change.FieldKey == fieldKey {
			return true
		}
	}
	return false
}

func touchesAnyField(changes []PatchChange, fieldKeys ...string) bool {
	for _, fieldKey := range fieldKeys {
		if touchesField(changes, fieldKey) {
			return true
		}
	}
	return false
}

func recordTypeMatchesView(catalog *sourcecatalog.Catalog, recordType string, viewSchemaID string) bool {
	surface, ok := catalog.SurfaceByViewID(viewSchemaID)
	return ok && recordType == surface.RecordType
}
