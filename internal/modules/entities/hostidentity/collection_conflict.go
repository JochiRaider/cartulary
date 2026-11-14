package hostidentity

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func canonicalEntityConflictCollection(value any) (map[string]any, bool) {
	object, ok := value.(map[string]any)
	if !ok || object["kind"] != "collection_value_v1" || object["ordered"] != false {
		return nil, false
	}
	items := []map[string]any{}
	switch raw := object["items"].(type) {
	case []any:
		for _, entry := range raw {
			item, ok := entry.(map[string]any)
			if !ok {
				return nil, false
			}
			items = append(items, maps.Clone(item))
		}
	case []map[string]any:
		for _, item := range raw {
			items = append(items, maps.Clone(item))
		}
	default:
		return nil, false
	}
	seen := map[string]bool{}
	for _, item := range items {
		ref, ok := item["item_ref"].(string)
		if !ok || ref == "" || seen[ref] {
			return nil, false
		}
		seen[ref] = true
	}
	return collectionValue(false, items), true
}

func previewEntityCollectionTx(ctx context.Context, tx pgx.Tx, incidentID, recordID uuid.UUID, change PatchChange, base map[string]any, requestHash []byte) (any, error) {
	items := slices.Clone(base["items"].([]map[string]any))
	labels := map[uuid.UUID]string{}
	if isPartyAssignmentField(change.FieldKey) {
		var err error
		labels, err = partyAssignmentConflictLabelsTx(ctx, tx, incidentID, change.FieldKey, change.CollectionActions)
		if err != nil {
			return nil, err
		}
	}
	for index, action := range change.CollectionActions {
		switch action.Op {
		case "add_alias":
			text, ok := fieldnorm.NormalizeEntityAliasText(action.NormalizedText)
			if !ok {
				return nil, invalidMutationPayload(change.FieldKey, "invalid_value")
			}
			names := make([]string, 0, len(items))
			for _, item := range items {
				name, ok := item["alias_text"].(string)
				if !ok || item["item_kind"] != "alias" {
					return nil, fmt.Errorf("invalid retained entity alias collection")
				}
				names = append(names, name)
			}
			// Match the source relation's citext uniqueness, including its database
			// collation. Go's Unicode case folding is not an equivalent policy.
			var duplicate bool
			if err := tx.QueryRow(ctx, `SELECT $1::citext = ANY($2::text[]::citext[])`, text, names).Scan(&duplicate); err != nil {
				return nil, err
			}
			if duplicate {
				continue
			}
			id := uuid.NewSHA1(recordID, []byte(fmt.Sprintf("%x:%s:%d", requestHash, change.FieldKey, index)))
			items = append(items, map[string]any{"item_ref": "entity_alias:" + id.String(), "item_kind": "alias", "alias_text": text, "display_text": text})
		case "add_party_ref":
			ref := "party_ref:" + action.PartyID.String()
			if slices.ContainsFunc(items, func(item map[string]any) bool { return item["item_ref"] == ref }) {
				continue
			}
			label, ok := labels[action.PartyID]
			if !ok {
				return nil, fmt.Errorf("party assignment conflict label unavailable")
			}
			items = append(items, map[string]any{"item_ref": ref, "item_kind": "party_ref", "party_id": action.PartyID.String(), "display_text": label})
		case "remove_alias", "remove_party_ref":
			index := slices.IndexFunc(items, func(item map[string]any) bool { return item["item_ref"] == action.ItemRef })
			if index < 0 {
				return nil, invalidMutationPayload(change.FieldKey, "invalid_value")
			}
			items = slices.Delete(items, index, index+1)
		default:
			return nil, invalidMutationPayload(change.FieldKey, "invalid_value")
		}
	}
	return collectionValue(false, items), nil
}
