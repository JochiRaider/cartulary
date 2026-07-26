package collaboration

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

const (
	EventFamilyRecordChanged           = "record_changed"
	EventFamilyJobProgress             = "job_progress"
	EventFamilyExtensionResourceChange = "extension_resource_changed"

	dispatchBatchLimit = 100
	dispatchLease      = 15 * time.Second
	dispatchInterval   = 100 * time.Millisecond
	dispatchBackoffMax = 30 * time.Second
)

type EventIntent struct {
	IntentKey         string
	IncidentID        uuid.UUID
	EventFamily       string
	CanonicalPayload  json.RawMessage
	SourceChangeSetID *uuid.UUID
	SourceRecordID    *uuid.UUID
	SourceRowVersion  *int64
	SourceIdentity    string
	MutationOrdinal   int
	CreatedAt         time.Time
}

type IntentAppender interface {
	AppendIntentTx(context.Context, pgx.Tx, EventIntent) error
}

type Store struct {
	db  postgres.DB
	now func() time.Time
}

func NewStore(db postgres.DB, now func() time.Time) *Store {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Store{db: db, now: now}
}

func NewEventIntent(
	intentKey string,
	incidentID uuid.UUID,
	eventFamily string,
	payload any,
	sourceIdentity string,
	mutationOrdinal int,
	createdAt time.Time,
) (EventIntent, error) {
	canonicalPayload, err := canonicalObject(payload)
	if err != nil {
		return EventIntent{}, err
	}
	intent := EventIntent{
		IntentKey:        intentKey,
		IncidentID:       incidentID,
		EventFamily:      eventFamily,
		CanonicalPayload: canonicalPayload,
		SourceIdentity:   sourceIdentity,
		MutationOrdinal:  mutationOrdinal,
		CreatedAt:        createdAt.UTC(),
	}
	if err := validateEventIntent(intent); err != nil {
		return EventIntent{}, err
	}
	return intent, nil
}

func (s *Store) AppendIntentTx(ctx context.Context, tx pgx.Tx, intent EventIntent) error {
	if s == nil || s.db == nil || tx == nil {
		return errors.New("collaboration intent store is not configured")
	}
	if err := validateEventIntent(intent); err != nil {
		return err
	}
	createdAt := intent.CreatedAt.UTC()
	_, err := tx.Exec(ctx, `
INSERT INTO collaboration_event_intents (
    intent_key,
    incident_id,
    event_family,
    canonical_payload,
    source_change_set_id,
    source_record_id,
    source_row_version,
    source_identity,
    mutation_ordinal,
    next_attempt_at,
    created_at,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10, $10)
ON CONFLICT (intent_key) DO UPDATE
SET canonical_payload = EXCLUDED.canonical_payload,
    source_change_set_id = EXCLUDED.source_change_set_id,
    source_record_id = EXCLUDED.source_record_id,
    source_row_version = EXCLUDED.source_row_version,
    source_identity = EXCLUDED.source_identity,
    mutation_ordinal = EXCLUDED.mutation_ordinal,
    next_attempt_at = EXCLUDED.next_attempt_at,
    updated_at = EXCLUDED.updated_at
WHERE collaboration_event_intents.dispatch_state = 'pending'
`, intent.IntentKey, intent.IncidentID, intent.EventFamily, []byte(intent.CanonicalPayload),
		intent.SourceChangeSetID, intent.SourceRecordID, intent.SourceRowVersion,
		intent.SourceIdentity, intent.MutationOrdinal, createdAt)
	if err != nil {
		return fmt.Errorf("append collaboration event intent: %w", err)
	}
	return nil
}

func validateEventIntent(intent EventIntent) error {
	if strings.TrimSpace(intent.IntentKey) == "" || len(intent.IntentKey) > 512 {
		return errors.New("collaboration event intent key is invalid")
	}
	if intent.IncidentID == uuid.Nil || strings.TrimSpace(intent.SourceIdentity) == "" || len(intent.SourceIdentity) > 512 {
		return errors.New("collaboration event intent source identity is invalid")
	}
	switch intent.EventFamily {
	case EventFamilyRecordChanged, EventFamilyJobProgress, EventFamilyExtensionResourceChange:
	default:
		return fmt.Errorf("unsupported collaboration event family %q", intent.EventFamily)
	}
	if intent.MutationOrdinal < 0 || intent.CreatedAt.IsZero() {
		return errors.New("collaboration event intent ordering is invalid")
	}
	var payload map[string]any
	if len(intent.CanonicalPayload) == 0 || json.Unmarshal(intent.CanonicalPayload, &payload) != nil || payload == nil {
		return errors.New("collaboration event intent payload must be a JSON object")
	}
	return nil
}

type ReplayResult struct {
	Status    string
	Messages  []platformws.Message
	HighWater int64
}

func (s *Store) CurrentHighWater(ctx context.Context, incidentID uuid.UUID) (int64, error) {
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

func (s *Store) IssueResumeToken(
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
	expiresAt := now.Add(platformws.ResumeWindow)
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

func (s *Store) ReplayMessages(
	ctx context.Context,
	sessionID uuid.UUID,
	incidentID uuid.UUID,
	clientInstanceID string,
	token string,
	lastSeenStreamSeq int64,
	now time.Time,
) (ReplayResult, error) {
	result := ReplayResult{Status: platformws.ResumeStatusResetNeeded, Messages: []platformws.Message{}}
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

	rows, err := s.db.Query(ctx, `
SELECT event_id, event_family, stream_seq, canonical_payload, emitted_at
  FROM collaboration_replay_events
 WHERE incident_id = $1
   AND stream_seq > $2
 ORDER BY stream_seq
`, incidentID, lastSeenStreamSeq)
	if err != nil {
		return result, fmt.Errorf("query collaboration replay events: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			eventID   uuid.UUID
			family    string
			streamSeq int64
			payload   []byte
			emittedAt time.Time
		)
		if err := rows.Scan(&eventID, &family, &streamSeq, &payload, &emittedAt); err != nil {
			return result, fmt.Errorf("scan collaboration replay event: %w", err)
		}
		result.Messages = append(result.Messages, replayMessage(eventID, incidentID, family, streamSeq, payload, emittedAt))
	}
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("iterate collaboration replay events: %w", err)
	}
	result.Status = platformws.ResumeStatusReplayed
	return result, nil
}

type replayBroadcaster interface {
	DeliverReplayable(platformws.Message) error
}

type Dispatcher struct {
	store       *Store
	broadcaster replayBroadcaster
	workerID    uuid.UUID
	now         func() time.Time
	interval    time.Duration

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func NewDispatcher(store *Store, broadcaster replayBroadcaster, now func() time.Time) *Dispatcher {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Dispatcher{
		store:       store,
		broadcaster: broadcaster,
		workerID:    uuid.New(),
		now:         now,
		interval:    dispatchInterval,
	}
}

func (d *Dispatcher) Start(parent context.Context) error {
	if d == nil || d.store == nil || d.store.db == nil || d.broadcaster == nil {
		return errors.New("collaboration dispatcher is not configured")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cancel != nil {
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	d.cancel = cancel
	d.done = make(chan struct{})
	go d.run(ctx, d.done)
	return nil
}

func (d *Dispatcher) Close(ctx context.Context) error {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	cancel := d.cancel
	done := d.done
	d.cancel = nil
	d.done = nil
	d.mu.Unlock()
	if cancel == nil {
		return nil
	}
	cancel()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *Dispatcher) run(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	backoff := d.interval
	for {
		_, err := d.RunOnce(ctx)
		delay := d.interval
		if err != nil {
			delay = backoff
			backoff *= 2
			if backoff > dispatchBackoffMax {
				backoff = dispatchBackoffMax
			}
		} else {
			backoff = d.interval
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (d *Dispatcher) RunOnce(ctx context.Context) (processed int, runErr error) {
	defer func() {
		d.recordDispatcherRun(ctx, processed, runErr)
	}()
	if d == nil || d.store == nil || d.store.db == nil || d.broadcaster == nil {
		return 0, errors.New("collaboration dispatcher is not configured")
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	now := d.now().UTC()
	sequenced, err := d.sequencePending(ctx, now)
	if err != nil {
		return 0, err
	}
	delivered := 0
	for delivered < dispatchBatchLimit {
		event, claimed, err := d.claimSequenced(ctx, now)
		if err != nil {
			return sequenced + delivered, err
		}
		if !claimed {
			break
		}
		if err := d.broadcaster.DeliverReplayable(event.Message); err != nil {
			if retryErr := d.retryDelivery(ctx, event.IntentID, now, err); retryErr != nil {
				return sequenced + delivered, retryErr
			}
			continue
		}
		if err := d.markDelivered(ctx, event.IntentID, now); err != nil {
			return sequenced + delivered, err
		}
		delivered++
	}
	if err := d.pruneReplay(ctx, now); err != nil {
		return sequenced + delivered, err
	}
	return sequenced + delivered, nil
}

type pendingIntent struct {
	IntentID   uuid.UUID
	IntentKey  string
	IncidentID uuid.UUID
	Family     string
	Payload    []byte
}

func (d *Dispatcher) sequencePending(ctx context.Context, now time.Time) (int, error) {
	tx, err := d.store.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin collaboration intent sequencing: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// Sequencing is globally serialized so two dispatcher processes cannot
	// assign a lower incident sequence to a later intent after SKIP LOCKED
	// bypasses an earlier row. Delivery remains independently lease-claimed.
	if _, err := tx.Exec(ctx, `
SELECT pg_advisory_xact_lock(hashtext('cartulary.collaboration.sequencer.v1'))
`); err != nil {
		return 0, fmt.Errorf("lock collaboration intent sequencer: %w", err)
	}
	rows, err := tx.Query(ctx, `
SELECT intent_id, intent_key, incident_id, event_family, canonical_payload
  FROM collaboration_event_intents
 WHERE dispatch_state = 'pending'
   AND next_attempt_at <= $1
   AND (lease_expires_at IS NULL OR lease_expires_at <= $1)
 ORDER BY created_at, mutation_ordinal, intent_key
 FOR UPDATE SKIP LOCKED
 LIMIT $2
`, now, dispatchBatchLimit)
	if err != nil {
		return 0, fmt.Errorf("claim collaboration intents: %w", err)
	}
	pending := make([]pendingIntent, 0, dispatchBatchLimit)
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
       lease_owner = NULL,
       lease_expires_at = NULL,
       last_error_code = NULL,
       updated_at = $3
 WHERE intent_id = $1
`, intent.IntentID, eventID, now); err != nil {
			return 0, fmt.Errorf("mark collaboration intent sequenced: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit collaboration intent sequencing: %w", err)
	}
	return len(pending), nil
}

type claimedEvent struct {
	IntentID uuid.UUID
	Message  platformws.Message
}

func (d *Dispatcher) claimSequenced(ctx context.Context, now time.Time) (claimedEvent, bool, error) {
	var (
		intentID   uuid.UUID
		eventID    uuid.UUID
		incidentID uuid.UUID
		family     string
		streamSeq  int64
		payload    []byte
		emittedAt  time.Time
	)
	err := d.store.db.QueryRow(ctx, `
WITH candidate AS (
    SELECT intent.intent_id
      FROM collaboration_event_intents AS intent
      JOIN collaboration_replay_events AS current_event
        ON current_event.event_id = intent.sequenced_event_id
     WHERE intent.dispatch_state = 'sequenced'
       AND intent.delivered_at IS NULL
       AND intent.next_attempt_at <= $1
       AND (intent.lease_expires_at IS NULL OR intent.lease_expires_at <= $1)
       AND NOT EXISTS (
           SELECT 1
             FROM collaboration_replay_events AS earlier_event
             JOIN collaboration_event_intents AS earlier_intent
               ON earlier_intent.sequenced_event_id = earlier_event.event_id
            WHERE earlier_event.incident_id = current_event.incident_id
              AND earlier_event.stream_seq < current_event.stream_seq
              AND earlier_intent.delivered_at IS NULL
       )
     ORDER BY current_event.emitted_at, current_event.incident_id, current_event.stream_seq
     FOR UPDATE OF intent SKIP LOCKED
     LIMIT 1
), claimed AS (
    UPDATE collaboration_event_intents AS intent
       SET lease_owner = $2,
           lease_expires_at = $3,
           attempt_count = attempt_count + 1,
           updated_at = $1
      FROM candidate
     WHERE intent.intent_id = candidate.intent_id
    RETURNING intent.intent_id, intent.sequenced_event_id
)
SELECT claimed.intent_id, replay.event_id, replay.incident_id, replay.event_family,
       replay.stream_seq, replay.canonical_payload, replay.emitted_at
  FROM claimed
  JOIN collaboration_replay_events AS replay
    ON replay.event_id = claimed.sequenced_event_id
`, now, d.workerID, now.Add(dispatchLease)).Scan(
		&intentID, &eventID, &incidentID, &family, &streamSeq, &payload, &emittedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return claimedEvent{}, false, nil
	}
	if err != nil {
		return claimedEvent{}, false, fmt.Errorf("claim sequenced collaboration event: %w", err)
	}
	return claimedEvent{
		IntentID: intentID,
		Message:  replayMessage(eventID, incidentID, family, streamSeq, payload, emittedAt),
	}, true, nil
}

func (d *Dispatcher) markDelivered(ctx context.Context, intentID uuid.UUID, now time.Time) error {
	tag, err := d.store.db.Exec(ctx, `
UPDATE collaboration_event_intents
   SET delivered_at = $3,
       lease_owner = NULL,
       lease_expires_at = NULL,
       last_error_code = NULL,
       updated_at = $3
 WHERE intent_id = $1
   AND lease_owner = $2
   AND delivered_at IS NULL
`, intentID, d.workerID, now)
	if err != nil {
		return fmt.Errorf("mark collaboration event delivered: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return errors.New("collaboration event delivery lease was lost")
	}
	return nil
}

func (d *Dispatcher) retryDelivery(ctx context.Context, intentID uuid.UUID, now time.Time, deliveryErr error) error {
	var attemptCount int
	if err := d.store.db.QueryRow(ctx, `
SELECT attempt_count
  FROM collaboration_event_intents
 WHERE intent_id = $1
   AND lease_owner = $2
`, intentID, d.workerID).Scan(&attemptCount); err != nil {
		return fmt.Errorf("read collaboration delivery attempt: %w", err)
	}
	backoff := 100 * time.Millisecond
	for index := 1; index < attemptCount && backoff < dispatchBackoffMax; index++ {
		backoff *= 2
	}
	if backoff > dispatchBackoffMax {
		backoff = dispatchBackoffMax
	}
	_, err := d.store.db.Exec(ctx, `
UPDATE collaboration_event_intents
   SET next_attempt_at = $3,
       lease_owner = NULL,
       lease_expires_at = NULL,
       last_error_code = 'broadcast_failed',
       updated_at = $4
 WHERE intent_id = $1
   AND lease_owner = $2
`, intentID, d.workerID, now.Add(backoff), now)
	if err != nil {
		return fmt.Errorf("schedule collaboration delivery retry after %v: %w", deliveryErr, err)
	}
	return nil
}

func (d *Dispatcher) pruneReplay(ctx context.Context, now time.Time) error {
	_, err := d.store.db.Exec(ctx, `
DELETE FROM collaboration_replay_events AS event
 USING collaboration_incident_stream_cursors AS cursor
 WHERE event.incident_id = cursor.incident_id
   AND event.emitted_at < $1
   AND event.stream_seq <= cursor.high_water_stream_seq - $2
`, now.Add(-platformws.ResumeWindow), platformws.MinimumReplayRetention)
	if err != nil {
		return fmt.Errorf("prune collaboration replay retention: %w", err)
	}
	return nil
}

func replayMessage(eventID uuid.UUID, incidentID uuid.UUID, family string, streamSeq int64, payload []byte, emittedAt time.Time) platformws.Message {
	return platformws.Message{
		Type:       family,
		IncidentID: incidentID.String(),
		EventID:    eventID.String(),
		EmittedAt:  emittedAt.UTC().Format(time.RFC3339Nano),
		StreamSeq:  &streamSeq,
		Payload:    append(json.RawMessage(nil), payload...),
	}
}

func canonicalObject(payload any) (json.RawMessage, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode collaboration event payload: %w", err)
	}
	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil || object == nil {
		return nil, errors.New("collaboration event payload must be an object")
	}
	canonical, err := json.Marshal(object)
	if err != nil {
		return nil, fmt.Errorf("canonicalize collaboration event payload: %w", err)
	}
	return canonical, nil
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
