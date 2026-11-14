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

const replayPageLimit = 250

// PreparedReplay owns one read-only snapshot. Visit consumes it; Close also
// releases an unused snapshot. Callers must defer Close immediately after prepare.
type PreparedReplay struct {
	status         string
	highWater      int64
	incidentID     uuid.UUID
	afterStreamSeq int64
	tx             pgx.Tx
}

func (r *PreparedReplay) Status() string   { return r.status }
func (r *PreparedReplay) HighWater() int64 { return r.highWater }

func (r *PreparedReplay) Close() error {
	if r == nil || r.tx == nil {
		return nil
	}
	tx := r.tx
	r.tx = nil
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := tx.Rollback(ctx)
	if errors.Is(err, pgx.ErrTxClosed) {
		return nil
	}
	return err
}

func (r *PreparedReplay) Visit(ctx context.Context, consume func(protocol.Message) error) (err error) {
	if r == nil {
		return errors.New("collaboration replay is unavailable")
	}
	if r.status != protocol.ResumeStatusReplayed {
		return nil
	}
	if r.tx == nil {
		return errors.New("collaboration replay is closed")
	}
	defer func() { err = errors.Join(err, r.Close()) }()
	return r.visit(ctx, consume)
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

func (s *PostgresStream) PrepareReplay(
	ctx context.Context,
	sessionID uuid.UUID,
	incidentID uuid.UUID,
	clientInstanceID string,
	token string,
	lastSeenStreamSeq int64,
	now time.Time,
) (*PreparedReplay, error) {
	result := &PreparedReplay{status: protocol.ResumeStatusResetNeeded, incidentID: incidentID, afterStreamSeq: lastSeenStreamSeq}
	if s == nil || s.db == nil || lastSeenStreamSeq < 0 {
		return result, nil
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, fmt.Errorf("begin collaboration replay snapshot: %w", err)
	}
	result.tx = tx
	keep := false
	defer func() {
		if !keep {
			_ = result.Close()
		}
	}()
	hash := resumeTokenHash(token)
	var tokenValid bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM collaboration_resume_tokens
     WHERE token_hash = $1 AND session_id = $2 AND incident_id = $3
       AND client_instance_id = $4 AND expires_at > $5
)
`, hash[:], sessionID, incidentID, clientInstanceID, now.UTC()).Scan(&tokenValid); err != nil {
		return nil, fmt.Errorf("validate collaboration resume token: %w", err)
	}
	if err := tx.QueryRow(ctx, `
SELECT COALESCE((
    SELECT high_water_stream_seq FROM collaboration_incident_stream_cursors
     WHERE incident_id = $1
), 0)
`, incidentID).Scan(&result.highWater); err != nil {
		return nil, fmt.Errorf("read collaboration stream high water: %w", err)
	}
	if !tokenValid || lastSeenStreamSeq > result.highWater {
		return result, nil
	}
	// The unique incident/sequence key plus this bounded count proves a complete
	// integer range, including its first, interior and terminal positions.
	var retainedCount int64
	if err := tx.QueryRow(ctx, `
SELECT count(*) FROM collaboration_replay_events
 WHERE incident_id = $1 AND stream_seq > $2 AND stream_seq <= $3
`, incidentID, lastSeenStreamSeq, result.highWater).Scan(&retainedCount); err != nil {
		return nil, fmt.Errorf("check collaboration replay range: %w", err)
	}
	if retainedCount != result.highWater-lastSeenStreamSeq {
		return result, nil
	}
	// Validate the complete immutable snapshot before the route can acknowledge it.
	// A second bounded pass delivers the same events without retaining the range.
	if err := result.visit(ctx, func(protocol.Message) error { return nil }); err != nil {
		return nil, err
	}
	result.status = protocol.ResumeStatusReplayed
	keep = true
	return result, nil
}

func (r *PreparedReplay) visit(ctx context.Context, consume func(protocol.Message) error) error {
	afterStreamSeq := r.afterStreamSeq
	for afterStreamSeq < r.highWater {
		if err := ctx.Err(); err != nil {
			return err
		}
		rows, err := r.tx.Query(ctx, `
SELECT event_id, event_family, stream_seq, canonical_payload, emitted_at
  FROM collaboration_replay_events
 WHERE incident_id = $1 AND stream_seq > $2 AND stream_seq <= $3
 ORDER BY stream_seq
 LIMIT $4
`, r.incidentID, afterStreamSeq, r.highWater, replayPageLimit)
		if err != nil {
			return fmt.Errorf("query collaboration replay events: %w", err)
		}
		page := make([]protocol.Message, 0, replayPageLimit)
		for rows.Next() {
			var eventID uuid.UUID
			var family string
			var streamSeq int64
			var payload []byte
			var emittedAt time.Time
			if err := rows.Scan(&eventID, &family, &streamSeq, &payload, &emittedAt); err != nil {
				rows.Close()
				return fmt.Errorf("scan collaboration replay event: %w", err)
			}
			if streamSeq != afterStreamSeq+1 {
				rows.Close()
				return errors.New("collaboration replay snapshot has a sequence gap")
			}
			message := replayMessage(eventID, r.incidentID, family, streamSeq, payload, emittedAt)
			if err := protocol.ValidateSequencedReplayableMessage(message); err != nil {
				rows.Close()
				return fmt.Errorf("validate authenticated collaboration replay event: %w", err)
			}
			afterStreamSeq = streamSeq
			page = append(page, message)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return fmt.Errorf("iterate collaboration replay events: %w", err)
		}
		if len(page) == 0 {
			return errors.New("collaboration replay snapshot ends before high water")
		}
		for _, message := range page {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := consume(message); err != nil {
				return err
			}
		}
	}
	return ctx.Err()
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
