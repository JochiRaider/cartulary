package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// appendProgressIntentTx is the Jobs-owned durable Collaboration producer.
// It deliberately publishes only the public common-job resource and never
// handler leases, payloads, or other worker-private state.
func appendProgressIntentTx(ctx context.Context, tx pgx.Tx, resource Resource) error {
	if resource.Scope.Kind != ScopeKindIncident || resource.Scope.IncidentID == nil {
		return nil
	}
	payload := map[string]any{
		"job_id": resource.JobID,
		"scope": map[string]any{
			"kind":        ScopeKindIncident,
			"incident_id": resource.Scope.IncidentID.String(),
		},
		"status": resource.Status,
		"progress": map[string]any{
			"completed": resource.Progress.Completed,
			"total":     resource.Progress.Total,
		},
		"updated_at": resource.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"cancelable": resource.Cancelable,
	}
	if resource.Message != nil {
		payload["message"] = *resource.Message
	}
	if resource.ResultSummary != nil {
		payload["result_summary"] = resource.ResultSummary
	}
	if resource.ErrorSummary != nil {
		payload["error_summary"] = resource.ErrorSummary
	}
	if resource.RetainedUntil != nil {
		payload["retained_until"] = resource.RetainedUntil.UTC().Format(time.RFC3339Nano)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal job progress intent: %w", err)
	}
	var intentKey string
	if err := tx.QueryRow(ctx, `
SELECT 'job_progress:' || $1 || ':' || encode(digest(($2::jsonb)::text, 'sha256'), 'hex')
`, resource.JobID, encoded).Scan(&intentKey); err != nil {
		return fmt.Errorf("derive job progress intent key: %w", err)
	}
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
) VALUES ($1, $2, 'job_progress', $3::jsonb, $4, 0, $5, $5, $5)
ON CONFLICT (intent_key) DO NOTHING
`, intentKey, *resource.Scope.IncidentID, encoded, "job:"+resource.JobID, resource.UpdatedAt.UTC())
	if err != nil {
		return fmt.Errorf("append job progress intent: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	var exact bool
	if err := tx.QueryRow(ctx, `
SELECT incident_id = $2
   AND event_family = 'job_progress'
   AND canonical_payload = $3::jsonb
   AND source_identity = $4
   AND mutation_ordinal = 0
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, intentKey, *resource.Scope.IncidentID, encoded, "job:"+resource.JobID).Scan(&exact); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("job progress intent %q disappeared during replay verification", intentKey)
		}
		return fmt.Errorf("verify job progress intent replay: %w", err)
	}
	if !exact {
		return fmt.Errorf("job progress intent key collision: %s", intentKey)
	}
	return nil
}
