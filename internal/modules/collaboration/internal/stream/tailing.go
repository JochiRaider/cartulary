package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

const tailBatchLimit = 100

// SeedTailCursor starts a newly launched API process at the current durable
// high-water mark. Existing events remain available through authenticated
// resume; only events committed after process startup are fanned out as live
// messages.
func (s *PostgresStream) SeedTailCursor(ctx context.Context) (map[uuid.UUID]int64, error) {
	rows, err := s.db.Query(ctx, `
SELECT incident_id, high_water_stream_seq
  FROM collaboration_incident_stream_cursors
`)
	if err != nil {
		return nil, fmt.Errorf("seed collaboration replay tail cursor: %w", err)
	}
	defer rows.Close()
	cursor := make(map[uuid.UUID]int64)
	for rows.Next() {
		var incidentID uuid.UUID
		var highWater int64
		if err := rows.Scan(&incidentID, &highWater); err != nil {
			return nil, fmt.Errorf("scan collaboration replay tail cursor: %w", err)
		}
		cursor[incidentID] = highWater
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate collaboration replay tail cursor: %w", err)
	}
	return cursor, nil
}

func (s *PostgresStream) TailReplay(ctx context.Context, cursor map[uuid.UUID]int64) ([]protocol.Message, error) {
	cursorJSON, err := json.Marshal(cursor)
	if err != nil {
		return nil, fmt.Errorf("encode collaboration replay tail cursor: %w", err)
	}
	rows, err := s.db.Query(ctx, `
SELECT event_id, incident_id, event_family, stream_seq, canonical_payload, emitted_at
  FROM collaboration_replay_events AS event
 WHERE event.stream_seq > COALESCE(
       (($1::jsonb ->> event.incident_id::text)::bigint),
       0
 )
 ORDER BY event.emitted_at, event.incident_id, event.stream_seq
 LIMIT $2
`, cursorJSON, tailBatchLimit)
	if err != nil {
		return nil, fmt.Errorf("tail collaboration replay events: %w", err)
	}
	defer rows.Close()

	messages := make([]protocol.Message, 0, tailBatchLimit)
	for rows.Next() {
		var (
			eventID    uuid.UUID
			incidentID uuid.UUID
			family     string
			streamSeq  int64
			payload    []byte
			emittedAt  time.Time
		)
		if err := rows.Scan(&eventID, &incidentID, &family, &streamSeq, &payload, &emittedAt); err != nil {
			return nil, fmt.Errorf("scan collaboration replay tail event: %w", err)
		}
		message := replayMessage(eventID, incidentID, family, streamSeq, payload, emittedAt)
		if err := protocol.ValidateSequencedReplayableMessage(message); err != nil {
			return nil, fmt.Errorf("validate collaboration replay tail event: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate collaboration replay tail events: %w", err)
	}
	return messages, nil
}
