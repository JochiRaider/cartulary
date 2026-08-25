package stream

import (
	"context"
	"fmt"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

const (
	maxRetainedEvents = 100_000
	maxRetainedBytes  = 256 * 1024 * 1024
	maxRetentionAge   = 24 * time.Hour
)

func (s *PostgresStream) PruneReplay(ctx context.Context, now time.Time) error {
	_, err := s.db.Exec(ctx, `
DELETE FROM collaboration_replay_events
 WHERE emitted_at < $1
`, now.Add(-maxRetentionAge))
	if err != nil {
		return fmt.Errorf("prune collaboration replay maximum age: %w", err)
	}
	_, err = s.db.Exec(ctx, `
WITH ranked AS (
    SELECT
        event_id,
        emitted_at,
        row_number() OVER (
            PARTITION BY incident_id
            ORDER BY stream_seq DESC
        ) AS recency_rank,
        sum(octet_length(canonical_payload::text)) OVER (
            PARTITION BY incident_id
            ORDER BY stream_seq DESC
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
        ) AS retained_bytes
      FROM collaboration_replay_events
)
DELETE FROM collaboration_replay_events AS event
 USING ranked
 WHERE event.event_id = ranked.event_id
   AND ranked.emitted_at < $1
   AND (
       ranked.recency_rank > $2
       OR ranked.retained_bytes > $3
   )
`, now.Add(-protocol.ResumeWindow), maxRetainedEvents, maxRetainedBytes)
	if err != nil {
		return fmt.Errorf("prune collaboration replay capacity: %w", err)
	}
	_, err = s.db.Exec(ctx, `
WITH expired AS (
    SELECT intent.intent_id
      FROM collaboration_event_intents AS intent
     WHERE intent.dispatch_state = 'sequenced'
       AND NOT EXISTS (
           SELECT 1
             FROM collaboration_replay_events AS replay
            WHERE replay.event_id = intent.sequenced_event_id
       )
     ORDER BY intent.sequenced_at, intent.intent_id
     LIMIT 1000
)
DELETE FROM collaboration_event_intents AS intent
 USING expired
 WHERE intent.intent_id = expired.intent_id
`)
	if err != nil {
		return fmt.Errorf("prune sequenced collaboration intents: %w", err)
	}
	_, err = s.db.Exec(ctx, `
WITH expired AS (
    SELECT token_hash
      FROM collaboration_resume_tokens
     WHERE expires_at <= $1
     ORDER BY expires_at, token_hash
     LIMIT 1000
)
DELETE FROM collaboration_resume_tokens AS token
 USING expired
 WHERE token.token_hash = expired.token_hash
`, now)
	if err != nil {
		return fmt.Errorf("prune collaboration resume tokens: %w", err)
	}
	return nil
}
