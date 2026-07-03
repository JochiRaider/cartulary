package incidentbundles

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

const IncidentPortabilityProfileID = "incident_portability"

func ImportedAttributionResolver() revisions.ImportedAttributionResolver {
	return importedAttributionResolver{}
}

type importedAttributionResolver struct{}

func (importedAttributionResolver) ResolveImportedSourceActorsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, sourceTable string, sourceColumn string, sourceRowIDs []string) (map[string]string, error) {
	if len(sourceRowIDs) == 0 {
		return map[string]string{}, nil
	}
	rows, err := tx.Query(ctx, `
SELECT source_row_id, source_actor_id
  FROM incident_bundle_imported_attributions
 WHERE incident_id = $1
   AND source_table = $2
   AND source_column = $3
   AND source_row_id IN (SELECT unnest($4::text[]))
 ORDER BY source_row_id
`, incidentID, sourceTable, sourceColumn, sourceRowIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resolved := map[string]string{}
	for rows.Next() {
		var rowID string
		var sourceActorID string
		if err := rows.Scan(&rowID, &sourceActorID); err != nil {
			return nil, err
		}
		resolved[rowID] = sourceActorID
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return resolved, nil
}
