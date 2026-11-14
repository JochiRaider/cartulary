package stream

import (
	"context"
	"fmt"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

const minRetainedEvents = 10_000

func (s *PostgresStream) PruneReplay(ctx context.Context, now time.Time) error {
	// Retain one contiguous suffix covering both the count and time guarantees.
	// Sequence order is authoritative even when publisher clocks move backwards.
	_, err := s.db.Exec(ctx, `
WITH ranked AS (
    SELECT
        incident_id,
        stream_seq,
        emitted_at,
        row_number() OVER (
            PARTITION BY incident_id
            ORDER BY stream_seq DESC
        ) AS recency_rank
      FROM collaboration_replay_events
), retained AS (
    SELECT incident_id, min(stream_seq) AS first_sequence
      FROM ranked
     WHERE recency_rank <= $2 OR emitted_at >= $1
     GROUP BY incident_id
)
DELETE FROM collaboration_replay_events AS event
 USING retained
 WHERE event.incident_id = retained.incident_id
   AND event.stream_seq < retained.first_sequence
`, now.Add(-protocol.ResumeWindow), minRetainedEvents)
	if err != nil {
		return fmt.Errorf("prune collaboration replay window: %w", err)
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
