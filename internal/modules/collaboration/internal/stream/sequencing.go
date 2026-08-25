package stream

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	sequenceBatchLimit  = 100
	maxSequenceAttempts = 12
)

type pendingIntent struct {
	IntentID   uuid.UUID
	IntentKey  string
	IncidentID uuid.UUID
	Family     string
	Payload    []byte
}

func (s *PostgresStream) SequencePending(ctx context.Context, now time.Time) (int, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin collaboration intent sequencing: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// Ensure every incident with pending work has the cursor row that acts as
	// its sequencing lock. Different incidents may then progress concurrently.
	if _, err := tx.Exec(ctx, `
INSERT INTO collaboration_incident_stream_cursors (
    incident_id, high_water_stream_seq, updated_at
)
SELECT DISTINCT incident_id, 0, $1::timestamp with time zone
  FROM collaboration_event_intents
 WHERE dispatch_state = 'pending'
   AND next_attempt_at <= $1::timestamp with time zone
ON CONFLICT (incident_id) DO NOTHING
`, now); err != nil {
		return 0, fmt.Errorf("initialize collaboration incident stream cursors: %w", err)
	}

	var incidentID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT cursor.incident_id
  FROM collaboration_incident_stream_cursors AS cursor
 WHERE (
       SELECT intent.next_attempt_at
         FROM collaboration_event_intents AS intent
        WHERE intent.incident_id = cursor.incident_id
          AND intent.dispatch_state = 'pending'
        ORDER BY intent.created_at, intent.mutation_ordinal, intent.intent_key
        LIMIT 1
 ) <= $1
   AND cursor.quarantined_at IS NULL
 ORDER BY (
       SELECT min(intent.created_at)
         FROM collaboration_event_intents AS intent
        WHERE intent.incident_id = cursor.incident_id
          AND intent.dispatch_state = 'pending'
 ), cursor.incident_id
 FOR UPDATE OF cursor SKIP LOCKED
 LIMIT 1
`, now).Scan(&incidentID); errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("lock collaboration incident stream: %w", err)
	}

	rows, err := tx.Query(ctx, `
SELECT intent_id, intent_key, incident_id, event_family, canonical_payload
  FROM collaboration_event_intents
 WHERE dispatch_state = 'pending'
   AND incident_id = $2
 ORDER BY created_at, mutation_ordinal, intent_key
 FOR UPDATE
 LIMIT $1
`, sequenceBatchLimit, incidentID)
	if err != nil {
		return 0, fmt.Errorf("claim collaboration intents: %w", err)
	}
	pending := make([]pendingIntent, 0, sequenceBatchLimit)
	for rows.Next() {
		var intent pendingIntent
		if err := rows.Scan(&intent.IntentID, &intent.IntentKey, &intent.IncidentID, &intent.Family, &intent.Payload); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan collaboration intent: %w", err)
		}
		pending = append(pending, intent)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("iterate collaboration intents: %w", err)
	}
	rows.Close()

	for _, intent := range pending {
		if err := validateReplayPayload(intent); err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				return 0, fmt.Errorf("roll back invalid collaboration event sequencing: %w", rollbackErr)
			}
			if retryErr := s.retrySequencingFailure(ctx, intent, now); retryErr != nil {
				return 0, retryErr
			}
			return 0, nil
		}
		var streamSeq int64
		if err := tx.QueryRow(ctx, `
INSERT INTO collaboration_incident_stream_cursors (
    incident_id, high_water_stream_seq, updated_at
) VALUES ($1, 1, $2)
ON CONFLICT (incident_id) DO UPDATE
SET high_water_stream_seq = collaboration_incident_stream_cursors.high_water_stream_seq + 1,
    updated_at = EXCLUDED.updated_at
RETURNING high_water_stream_seq
`, intent.IncidentID, now).Scan(&streamSeq); err != nil {
			return 0, fmt.Errorf("advance collaboration incident stream: %w", err)
		}
		eventID := uuid.New()
		if _, err := tx.Exec(ctx, `
INSERT INTO collaboration_replay_events (
    event_id, incident_id, stream_seq, intent_key, event_family, canonical_payload, emitted_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
`, eventID, intent.IncidentID, streamSeq, intent.IntentKey, intent.Family, intent.Payload, now); err != nil {
			return 0, fmt.Errorf("insert collaboration replay event: %w", err)
		}
		if _, err := tx.Exec(ctx, `
UPDATE collaboration_event_intents
   SET dispatch_state = 'sequenced',
       sequenced_event_id = $2,
       sequenced_at = $3,
       last_error_code = NULL,
       updated_at = $3
 WHERE intent_id = $1
`, intent.IntentID, eventID, now); err != nil {
			return 0, fmt.Errorf("mark collaboration intent sequenced: %w", err)
		}
	}
	if len(pending) > 0 {
		if _, err := tx.Exec(ctx, `
SELECT pg_notify('cartulary_collaboration_replay', $1)
`, incidentID.String()); err != nil {
			return 0, fmt.Errorf("notify collaboration replay tailers: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit collaboration intent sequencing: %w", err)
	}
	return len(pending), nil
}

func validateReplayPayload(intent pendingIntent) error {
	return ValidateEventFamilyPayload(intent.IncidentID, intent.Family, intent.Payload)
}

func (s *PostgresStream) retrySequencingFailure(ctx context.Context, intent pendingIntent, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin collaboration sequencing retry: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
INSERT INTO collaboration_incident_stream_cursors (
    incident_id, high_water_stream_seq, updated_at
) VALUES ($1, 0, $2)
ON CONFLICT (incident_id) DO NOTHING
`, intent.IncidentID, now); err != nil {
		return fmt.Errorf("restore collaboration incident cursor for retry: %w", err)
	}
	var attemptCount int
	if err := tx.QueryRow(ctx, `
SELECT attempt_count
  FROM collaboration_event_intents
 WHERE intent_id = $1
   AND dispatch_state = 'pending'
 FOR UPDATE
`, intent.IntentID).Scan(&attemptCount); err != nil {
		return fmt.Errorf("lock collaboration sequencing retry: %w", err)
	}
	attemptCount++
	backoff := 100 * time.Millisecond
	for index := 1; index < attemptCount && backoff < dispatchBackoffMax; index++ {
		backoff *= 2
	}
	if backoff > dispatchBackoffMax {
		backoff = dispatchBackoffMax
	}
	jitterPercent := int(intent.IntentID[0])%41 - 20
	backoff += time.Duration(int64(backoff) * int64(jitterPercent) / 100)
	if _, err := tx.Exec(ctx, `
UPDATE collaboration_event_intents
   SET attempt_count = $2,
       next_attempt_at = $3,
       last_error_code = 'invalid_event_payload',
       updated_at = $4
 WHERE intent_id = $1
`, intent.IntentID, attemptCount, now.Add(backoff), now); err != nil {
		return fmt.Errorf("schedule collaboration sequencing retry: %w", err)
	}
	if attemptCount >= maxSequenceAttempts {
		if _, err := tx.Exec(ctx, `
UPDATE collaboration_incident_stream_cursors
   SET failure_count = $2,
       quarantined_at = $3,
       quarantine_reason = 'invalid_event_payload',
       updated_at = $3
 WHERE incident_id = $1
`, intent.IncidentID, maxSequenceAttempts, now); err != nil {
			return fmt.Errorf("quarantine collaboration incident stream: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit collaboration sequencing retry: %w", err)
	}
	return nil
}
