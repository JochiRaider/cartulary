package revisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// importedHistoryAttributionDecorator enriches portable history after row
// materialization. The resolver remains the sole owner of imported identity
// remapping; history does not infer source actors using storage projections.
type importedHistoryAttributionDecorator struct {
	resolver ImportedAttributionResolver
}

func (d importedHistoryAttributionDecorator) DecorateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, items []RecordHistoryItem) error {
	if len(items) == 0 || d.resolver == nil {
		return nil
	}
	rowIDs := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		rowID := item.ChangeSetID.String()
		if _, ok := seen[rowID]; ok {
			continue
		}
		seen[rowID] = struct{}{}
		rowIDs = append(rowIDs, rowID)
	}
	sourceActors, err := d.resolver.ResolveImportedSourceActorsTx(ctx, tx, incidentID, "change_sets", "actor_user_id", rowIDs)
	if err != nil {
		return err
	}
	for index := range items {
		sourceActorID := sourceActors[items[index].ChangeSetID.String()]
		if sourceActorID != "" {
			items[index].SourceActorID = &sourceActorID
		}
	}
	return nil
}
