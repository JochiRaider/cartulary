package hostidentity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

func syncEntityAliasesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, entityType string, actions []CollectionAction, actorUserID uuid.UUID, now time.Time) (AliasSyncResult, error) {
	result := AliasSyncResult{}
	for _, action := range actions {
		normalized, ok := fieldnorm.NormalizeAliasText(action.NormalizedText)
		if !ok {
			return AliasSyncResult{}, fmt.Errorf("invalid entity alias")
		}
		aliasID := uuid.New()
		var insertedID uuid.UUID
		err := tx.QueryRow(ctx, `
INSERT INTO entity_aliases (
	entity_alias_id,
    incident_id,
    record_id,
    entity_type,
    raw_text,
    normalized_text,
    classification,
    created_by_user_id,
    created_at,
    deleted_at
)
VALUES ($1, $2, $3, $4, $5::text, $5::citext, 'suggestion_only', $6, $7, NULL)
ON CONFLICT (record_id, entity_type, normalized_text) WHERE deleted_at IS NULL
DO NOTHING
RETURNING entity_alias_id
`, aliasID, incidentID, recordID, entityType, normalized, actorUserID, now.UTC()).Scan(&insertedID)
		if errors.Is(err, pgx.ErrNoRows) {
			result.DuplicateNoopCount++
			continue
		}
		if err != nil {
			return AliasSyncResult{}, fmt.Errorf("insert entity alias: %w", err)
		}
		result.Added = append(result.Added, AliasMutationValue{
			EntityAliasID: insertedID,
			IncidentID:    incidentID,
			RecordID:      recordID,
			EntityType:    entityType,
			AliasText:     normalized,
			CreatedByUser: actorUserID,
			CreatedAt:     now.UTC(),
		})
	}
	return result, nil
}

func applyEntityAliasActionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, entityType string, actions []CollectionAction, actorUserID uuid.UUID, now time.Time) ([]AliasAppliedMutation, error) {
	mutations := make([]AliasAppliedMutation, 0, len(actions))
	for _, action := range actions {
		switch action.Op {
		case "add_alias":
			result, err := syncEntityAliasesTx(ctx, tx, incidentID, recordID, entityType, []CollectionAction{action}, actorUserID, now)
			if err != nil {
				return nil, err
			}
			for _, added := range result.Added {
				mutations = append(mutations, AliasAppliedMutation{
					OperationKind: "create",
					TargetID:      "entity_alias:" + added.EntityAliasID.String(),
					AfterValue:    added.MutationValue(),
				})
			}
		case "remove_alias":
			aliasID, err := parseEntityAliasItemRef(action.ItemRef)
			if err != nil {
				return nil, ErrInvalidAliasReference
			}
			var value AliasMutationValue
			var classification string
			err = tx.QueryRow(ctx, `
SELECT entity_alias_id, incident_id, record_id, entity_type, normalized_text::text,
       classification, created_by_user_id, created_at
  FROM entity_aliases
 WHERE entity_alias_id = $1
   AND deleted_at IS NULL
 FOR UPDATE
`, aliasID).Scan(
				&value.EntityAliasID,
				&value.IncidentID,
				&value.RecordID,
				&value.EntityType,
				&value.AliasText,
				&classification,
				&value.CreatedByUser,
				&value.CreatedAt,
			)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrInvalidAliasReference
			}
			if err != nil {
				return nil, err
			}
			if value.IncidentID != incidentID || value.RecordID != recordID || value.EntityType != entityType || classification != "suggestion_only" {
				return nil, ErrInvalidAliasReference
			}
			before := value.MutationValue()
			deletedAt := now.UTC()
			if tag, err := tx.Exec(ctx, `UPDATE entity_aliases SET deleted_at = $2 WHERE entity_alias_id = $1 AND deleted_at IS NULL`, aliasID, deletedAt); err != nil {
				return nil, err
			} else if tag.RowsAffected() != 1 {
				return nil, ErrInvalidAliasReference
			}
			value.DeletedAt = &deletedAt
			mutations = append(mutations, AliasAppliedMutation{
				OperationKind: "delete",
				TargetID:      action.ItemRef,
				BeforeValue:   before,
				AfterValue:    value.MutationValue(),
			})
		default:
			return nil, ErrInvalidAliasReference
		}
	}
	return mutations, nil
}

func loadEntityAliasesTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) ([]AliasValue, error) {
	rows, err := tx.Query(ctx, `
SELECT entity_alias_id, normalized_text::text
  FROM entity_aliases
 WHERE record_id = $1
   AND entity_type = $2
   AND deleted_at IS NULL
 ORDER BY normalized_text ASC, created_at ASC, entity_alias_id ASC
`, recordID, entityType)
	if err != nil {
		return nil, fmt.Errorf("load entity aliases: %w", err)
	}
	defer rows.Close()

	aliases := make([]AliasValue, 0)
	for rows.Next() {
		var alias AliasValue
		if err := rows.Scan(&alias.EntityAliasID, &alias.AliasText); err != nil {
			return nil, fmt.Errorf("scan entity alias: %w", err)
		}
		aliases = append(aliases, alias)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity aliases: %w", err)
	}
	return aliases, nil
}
