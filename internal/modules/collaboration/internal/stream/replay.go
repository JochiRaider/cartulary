package stream

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

const (
	replayPageLimit = 250
	maxReplayEvents = 10_000
	maxReplayBytes  = 16 * 1024 * 1024
)

type ReplayResult struct {
	Status    string
	Messages  []protocol.Message
	HighWater int64
}

func (s *PostgresStream) CurrentHighWater(ctx context.Context, incidentID uuid.UUID) (int64, error) {
	if s == nil || s.db == nil || incidentID == uuid.Nil {
		return 0, errors.New("collaboration stream store is not configured")
	}
	var highWater int64
	if err := s.db.QueryRow(ctx, `
SELECT COALESCE((
    SELECT high_water_stream_seq
      FROM collaboration_incident_stream_cursors
     WHERE incident_id = $1
), 0)
`, incidentID).Scan(&highWater); err != nil {
		return 0, fmt.Errorf("read collaboration stream high water: %w", err)
	}
	return highWater, nil
}

func (s *PostgresStream) IssueResumeToken(
	ctx context.Context,
	sessionID uuid.UUID,
	incidentID uuid.UUID,
	clientInstanceID string,
	sessionExpiresAt time.Time,
	now time.Time,
) (string, time.Time, error) {
	if s == nil || s.db == nil || sessionID == uuid.Nil || incidentID == uuid.Nil || strings.TrimSpace(clientInstanceID) == "" {
		return "", time.Time{}, errors.New("collaboration resume-token identity is invalid")
	}
	token, err := opaqueToken()
	if err != nil {
		return "", time.Time{}, err
	}
	now = now.UTC()
	expiresAt := now.Add(protocol.ResumeWindow)
	if sessionExpiresAt.Before(expiresAt) {
		expiresAt = sessionExpiresAt.UTC()
	}
	if !expiresAt.After(now) {
		return "", time.Time{}, errors.New("collaboration resume-token expiry is invalid")
	}
	hash := resumeTokenHash(token)
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", time.Time{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
DELETE FROM collaboration_resume_tokens
 WHERE expires_at <= $1
`, now); err != nil {
		return "", time.Time{}, fmt.Errorf("prune collaboration resume tokens: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO collaboration_resume_tokens (
    token_hash, session_id, incident_id, client_instance_id, issued_at, expires_at
) VALUES ($1, $2, $3, $4, $5, $6)
`, hash[:], sessionID, incidentID, clientInstanceID, now, expiresAt); err != nil {
		return "", time.Time{}, fmt.Errorf("insert collaboration resume token: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return "", time.Time{}, fmt.Errorf("commit collaboration resume token: %w", err)
	}
	return token, expiresAt, nil
}

func (s *PostgresStream) ReplayMessages(
	ctx context.Context,
	sessionID uuid.UUID,
	incidentID uuid.UUID,
	clientInstanceID string,
	token string,
	lastSeenStreamSeq int64,
	now time.Time,
) (ReplayResult, error) {
	result := ReplayResult{Status: protocol.ResumeStatusResetNeeded, Messages: []protocol.Message{}}
	if s == nil || s.db == nil || lastSeenStreamSeq < 0 {
		return result, nil
	}
	hash := resumeTokenHash(token)
	var tokenValid bool
	if err := s.db.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM collaboration_resume_tokens
     WHERE token_hash = $1
       AND session_id = $2
       AND incident_id = $3
       AND client_instance_id = $4
       AND expires_at > $5
)
`, hash[:], sessionID, incidentID, clientInstanceID, now.UTC()).Scan(&tokenValid); err != nil {
		return result, fmt.Errorf("validate collaboration resume token: %w", err)
	}
	if err := s.db.QueryRow(ctx, `
SELECT COALESCE((
    SELECT high_water_stream_seq
      FROM collaboration_incident_stream_cursors
     WHERE incident_id = $1
), 0)
`, incidentID).Scan(&result.HighWater); err != nil {
		return result, fmt.Errorf("read collaboration stream high water: %w", err)
	}
	if !tokenValid || lastSeenStreamSeq > result.HighWater {
		return result, nil
	}

	var firstRetained *int64
	if err := s.db.QueryRow(ctx, `
SELECT min(stream_seq)
  FROM collaboration_replay_events
 WHERE incident_id = $1
`, incidentID).Scan(&firstRetained); err != nil {
		return result, fmt.Errorf("read collaboration replay lower bound: %w", err)
	}
	if (firstRetained != nil && lastSeenStreamSeq < *firstRetained-1) ||
		(firstRetained == nil && lastSeenStreamSeq < result.HighWater) {
		return result, nil
	}

	afterStreamSeq := lastSeenStreamSeq
	replayBytes := 0
	for afterStreamSeq < result.HighWater {
		rows, err := s.db.Query(ctx, `
SELECT event_id, event_family, stream_seq, canonical_payload, emitted_at
  FROM collaboration_replay_events
 WHERE incident_id = $1
   AND stream_seq > $2
   AND stream_seq <= $3
 ORDER BY stream_seq
 LIMIT $4
`, incidentID, afterStreamSeq, result.HighWater, replayPageLimit)
		if err != nil {
			return result, fmt.Errorf("query collaboration replay events: %w", err)
		}
		pageCount := 0
		for rows.Next() {
			var (
				eventID   uuid.UUID
				family    string
				streamSeq int64
				payload   []byte
				emittedAt time.Time
			)
			if err := rows.Scan(&eventID, &family, &streamSeq, &payload, &emittedAt); err != nil {
				rows.Close()
				return result, fmt.Errorf("scan collaboration replay event: %w", err)
			}
			pageCount++
			afterStreamSeq = streamSeq
			replayBytes += len(payload)
			if len(result.Messages) >= maxReplayEvents || replayBytes > maxReplayBytes {
				rows.Close()
				result.Messages = []protocol.Message{}
				return result, nil
			}
			result.Messages = append(result.Messages, replayMessage(eventID, incidentID, family, streamSeq, payload, emittedAt))
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return result, fmt.Errorf("iterate collaboration replay events: %w", err)
		}
		rows.Close()
		if pageCount < replayPageLimit {
			break
		}
	}
	result.Status = protocol.ResumeStatusReplayed
	return result, nil
}

func replayMessage(eventID uuid.UUID, incidentID uuid.UUID, family string, streamSeq int64, payload []byte, emittedAt time.Time) protocol.Message {
	return protocol.Message{
		Type:       family,
		IncidentID: incidentID.String(),
		EventID:    eventID.String(),
		EmittedAt:  emittedAt.UTC().Format(time.RFC3339Nano),
		StreamSeq:  &streamSeq,
		Payload:    append(json.RawMessage(nil), payload...),
	}
}

func opaqueToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func resumeTokenHash(token string) [32]byte {
	return sha256.Sum256([]byte(token))
}
