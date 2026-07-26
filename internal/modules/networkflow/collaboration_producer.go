package networkflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func appendTableResourceIntentTx(ctx context.Context, tx pgx.Tx, table TableRecord, changeKind string, reasonCode string) error {
	payload := map[string]any{
		"extension_profile_id": "network_flow_activity",
		"resource_kind":        "network_flow_table",
		"resource_id":          table.TableID,
		"change_kind":          changeKind,
		"reason_code":          reasonCode,
		"workspace_refs": []map[string]any{{
			"kind":                 "extension_workspace",
			"extension_profile_id": "network_flow_activity",
			"workspace_key":        "network_analysis",
		}},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal network flow collaboration intent: %w", err)
	}
	intentKey := fmt.Sprintf(
		"extension_resource_changed:network_flow_table:%s:%d:%s",
		table.TableID,
		table.TableVersion,
		reasonCode,
	)
	tag, err := tx.Exec(ctx, `
INSERT INTO collaboration_event_intents (
    intent_key,
    incident_id,
    event_family,
    canonical_payload,
    source_identity,
    mutation_ordinal,
    next_attempt_at,
    created_at,
    updated_at
) VALUES ($1, $2, 'extension_resource_changed', $3::jsonb, $4, 0, $5, $5, $5)
ON CONFLICT (intent_key) DO NOTHING
`, intentKey, table.IncidentID, encoded, "network_flow_table:"+table.TableID, table.UpdatedAt.UTC())
	if err != nil {
		return fmt.Errorf("append network flow collaboration intent: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	var exact bool
	if err := tx.QueryRow(ctx, `
SELECT incident_id = $2
   AND event_family = 'extension_resource_changed'
   AND canonical_payload = $3::jsonb
   AND source_identity = $4
   AND mutation_ordinal = 0
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, intentKey, table.IncidentID, encoded, "network_flow_table:"+table.TableID).Scan(&exact); err != nil {
		return fmt.Errorf("verify network flow collaboration intent replay: %w", err)
	}
	if !exact {
		return fmt.Errorf("network flow collaboration intent key collision: %s", intentKey)
	}
	return nil
}
