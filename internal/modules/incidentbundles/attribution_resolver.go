package incidentbundles

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func ImportedAttributionResolver() revisions.ImportedAttributionResolver {
	return importedAttributionResolver{}
}

type importedAttributionResolver struct{}

func (importedAttributionResolver) ResolveImportedHistoryActorsTx(ctx context.Context, tx pgx.Tx, incidentID, changeSetID uuid.UUID, sourceActorIDs []string) (map[string]string, error) {
	var localUserID uuid.UUID
	err := tx.QueryRow(ctx, `
SELECT local_user_id
  FROM incident_bundle_imported_attributions
 WHERE incident_id=$1 AND source_table='change_sets'
   AND source_column='actor_user_id' AND source_row_id=$2
`, incidentID, changeSetID.String()).Scan(&localUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `
SELECT source_actor_id
  FROM incident_bundle_imported_actors
 WHERE incident_id=$1 AND source_actor_id=ANY($2::text[])
 ORDER BY source_actor_id
`, incidentID, sourceActorIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]string, len(sourceActorIDs))
	for rows.Next() {
		var actorID string
		if err := rows.Scan(&actorID); err != nil {
			return nil, err
		}
		result[actorID] = localUserID.String()
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, actorID := range sourceActorIDs {
		if _, ok := result[actorID]; !ok {
			return nil, errors.New("incident bundle historical actor is unavailable")
		}
	}
	return result, nil
}

type portableAttributionResolver struct{}

// PortableSourceAttributions exposes the owner resolver to application snapshot
// composition; each consumer remains confined by its logical read capability.
func PortableSourceAttributions() sourceport.PortableAttributionResolver {
	return portableAttributionResolver{}
}

func (portableAttributionResolver) ResolvePortableSourceActors(
	ctx context.Context,
	q incidentportability.Queryer,
	incidentID uuid.UUID,
	sourceTable string,
	sourceColumn string,
	sourceRowIDs []string,
) (map[string]string, error) {
	if len(sourceRowIDs) == 0 {
		return map[string]string{}, nil
	}
	rows, err := q.Query(ctx, `
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
