package networkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ResourceIntent is Network Flow's source-owned Collaboration producer
// contract. Network Flow owns the payload and stable key; Collaboration owns
// validation, persistence, replay comparison, and collision diagnostics.
type ResourceIntent struct {
	IntentKey        string
	IncidentID       uuid.UUID
	CanonicalPayload json.RawMessage
	SourceIdentity   string
	CreatedAt        time.Time
}

type ResourceIntentAppender interface {
	AppendResourceIntentTx(context.Context, pgx.Tx, ResourceIntent) error
}

func (s *Store) appendTableResourceIntentTx(ctx context.Context, tx pgx.Tx, table TableRecord, changeKind string, reasonCode string) error {
	if s == nil || s.resourceIntents == nil {
		return fmt.Errorf("network flow resource intent appender is not configured")
	}
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
	return s.resourceIntents.AppendResourceIntentTx(ctx, tx, ResourceIntent{
		IntentKey: fmt.Sprintf(
			"extension_resource_changed:network_flow_table:%s:%d:%s",
			table.TableID,
			table.TableVersion,
			reasonCode,
		),
		IncidentID:       table.IncidentID,
		CanonicalPayload: encoded,
		SourceIdentity:   "network_flow_table:" + table.TableID,
		CreatedAt:        table.UpdatedAt.UTC(),
	})
}
