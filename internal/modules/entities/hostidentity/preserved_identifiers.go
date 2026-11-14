package hostidentity

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/preservedhistory"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

func syncPreservedIdentifiersTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, entityType string, seeds []identifierSeed, actorUserID uuid.UUID, now time.Time) ([]preservedhistory.Value, error) {
	var inserted []preservedhistory.Value
	for _, seed := range seeds {
		normalized, ok := fieldnorm.NormalizeIdentifier(seed.IdentifierType, seed.RawValue)
		if !ok {
			continue
		}
		var exists bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM entity_preserved_identifiers
     WHERE record_id = $1
       AND entity_type = $2
       AND identifier_type = $3
       AND normalized_value = $4
       AND classification = $5
       AND deleted_at IS NULL
)`, recordID, entityType, seed.IdentifierType, normalized, seed.Classification).Scan(&exists); err != nil {
			return nil, fmt.Errorf("query preserved identifier existence: %w", err)
		}
		if exists {
			continue
		}
		value := preservedhistory.Value{IncidentID: incidentID, RecordID: recordID, EntityType: entityType,
			IdentifierType: seed.IdentifierType, RawValue: seed.RawValue, NormalizedValue: normalized,
			Classification: seed.Classification, CreatedByUserID: actorUserID}
		if err := tx.QueryRow(ctx, `
INSERT INTO entity_preserved_identifiers (
    incident_id,
    record_id,
    entity_type,
    identifier_type,
    raw_value,
    normalized_value,
    classification,
    created_by_user_id,
    created_at,
    deleted_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULL)
RETURNING entity_preserved_identifier_id, created_at
`, incidentID, recordID, entityType, seed.IdentifierType, seed.RawValue, normalized, seed.Classification, actorUserID, now.UTC()).Scan(&value.ID, &value.CreatedAt); err != nil {
			return nil, fmt.Errorf("insert preserved identifier: %w", err)
		}
		inserted = append(inserted, value)
	}
	return inserted, nil
}
