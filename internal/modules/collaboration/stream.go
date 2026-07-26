package collaboration

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

const (
	EventFamilyRecordChanged           = "record_changed"
	EventFamilyJobProgress             = "job_progress"
	EventFamilyExtensionResourceChange = "extension_resource_changed"

	dispatchBatchLimit  = 100
	dispatchInterval    = 100 * time.Millisecond
	dispatchBackoffMax  = 30 * time.Second
	maxIntentPayload    = 256 * 1024
	replayPageLimit     = 250
	maxReplayEvents     = 10_000
	maxReplayBytes      = 16 * 1024 * 1024
	maxSequenceAttempts = 12
	maxRetainedEvents   = 100_000
	maxRetainedBytes    = 256 * 1024 * 1024
	maxRetentionAge     = 24 * time.Hour
)

var ErrIntentKeyCollision = errors.New("collaboration intent key collision")

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
	if s == nil || s.db == nil {
		return errors.New("collaboration intent store is not configured")
	}
	return AppendIntentTx(ctx, tx, intent)
}

// AppendIntentTx persists one validated intent through the caller's
// transaction. Source-owner producers use this operation when they already
// own the transaction and do not need any of Store's query or worker
// capabilities.
func AppendIntentTx(ctx context.Context, tx pgx.Tx, intent EventIntent) error {
	if tx == nil {
		return errors.New("collaboration intent transaction is not configured")
	}
	if err := validateEventIntent(intent); err != nil {
		return err
	}
	createdAt := intent.CreatedAt.UTC()
	tag, err := tx.Exec(ctx, `
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
ON CONFLICT (intent_key) DO NOTHING
`, intent.IntentKey, intent.IncidentID, intent.EventFamily, string(intent.CanonicalPayload),
		intent.SourceChangeSetID, intent.SourceRecordID, intent.SourceRowVersion,
		intent.SourceIdentity, intent.MutationOrdinal, createdAt)
	if err != nil {
		return fmt.Errorf("append collaboration event intent: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	var exactDuplicate bool
	if err := tx.QueryRow(ctx, `
SELECT incident_id = $2
   AND event_family = $3
   AND canonical_payload = $4::jsonb
   AND source_change_set_id IS NOT DISTINCT FROM $5
   AND source_record_id IS NOT DISTINCT FROM $6
   AND source_row_version IS NOT DISTINCT FROM $7
   AND source_identity = $8
   AND mutation_ordinal = $9
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, intent.IntentKey, intent.IncidentID, intent.EventFamily, string(intent.CanonicalPayload),
		intent.SourceChangeSetID, intent.SourceRecordID, intent.SourceRowVersion,
		intent.SourceIdentity, intent.MutationOrdinal).Scan(&exactDuplicate); err != nil {
		return fmt.Errorf("verify collaboration event intent replay: %w", err)
	}
	if exactDuplicate {
		return nil
	}
	var (
		existingPayload  []byte
		existingIdentity string
		existingOrdinal  int
	)
	if err := tx.QueryRow(ctx, `
SELECT canonical_payload::text, source_identity, mutation_ordinal
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, intent.IntentKey).Scan(&existingPayload, &existingIdentity, &existingOrdinal); err != nil {
		return fmt.Errorf("inspect collaboration intent collision: %w", err)
	}
	existingDigest := sha256.Sum256(existingPayload)
	incomingDigest := sha256.Sum256(intent.CanonicalPayload)
	return fmt.Errorf(
		"%w: %s existing_payload_sha256=%x incoming_payload_sha256=%x payload_mismatch_keys=%v existing_source_identity=%q incoming_source_identity=%q existing_ordinal=%d incoming_ordinal=%d",
		ErrIntentKeyCollision,
		intent.IntentKey,
		existingDigest,
		incomingDigest,
		payloadMismatchKeys(existingPayload, intent.CanonicalPayload),
		existingIdentity,
		intent.SourceIdentity,
		existingOrdinal,
		intent.MutationOrdinal,
	)
}

func payloadMismatchKeys(existing []byte, incoming []byte) []string {
	var existingObject map[string]json.RawMessage
	var incomingObject map[string]json.RawMessage
	if json.Unmarshal(existing, &existingObject) != nil || json.Unmarshal(incoming, &incomingObject) != nil {
		return []string{"<malformed>"}
	}
	keys := make(map[string]struct{}, len(existingObject)+len(incomingObject))
	for key := range existingObject {
		keys[key] = struct{}{}
	}
	for key := range incomingObject {
		keys[key] = struct{}{}
	}
	mismatches := make([]string, 0)
	for key := range keys {
		if !bytes.Equal(existingObject[key], incomingObject[key]) {
			mismatches = append(mismatches, key)
		}
	}
	slices.Sort(mismatches)
	return mismatches
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
	if len(intent.CanonicalPayload) > maxIntentPayload {
		return fmt.Errorf("collaboration event intent payload exceeds %d bytes", maxIntentPayload)
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
				result.Messages = []platformws.Message{}
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
	result.Status = platformws.ResumeStatusReplayed
	return result, nil
}

type replayBroadcaster interface {
	DeliverReplayable(platformws.Message) error
}

type notificationConnectionAcquirer interface {
	Acquire(context.Context) (*pgxpool.Conn, error)
}

type Dispatcher struct {
	store       *Store
	broadcaster replayBroadcaster
	now         func() time.Time
	interval    time.Duration

	mu              sync.Mutex
	runMu           sync.Mutex
	cancel          context.CancelFunc
	done            chan struct{}
	tailCursor      map[uuid.UUID]int64
	tailInitialized bool
}

func NewDispatcher(store *Store, broadcaster replayBroadcaster, now func() time.Time) *Dispatcher {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Dispatcher{
		store:       store,
		broadcaster: broadcaster,
		now:         now,
		interval:    dispatchInterval,
		tailCursor:  make(map[uuid.UUID]int64),
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
	d.runMu.Lock()
	if err := d.seedTailCursor(parent); err != nil {
		d.runMu.Unlock()
		return err
	}
	d.tailInitialized = true
	d.runMu.Unlock()
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
	notifications := make(chan struct{}, 1)
	listenerDone := make(chan struct{})
	if acquirer, ok := d.store.db.(notificationConnectionAcquirer); ok {
		go d.listenReplayNotifications(ctx, acquirer, notifications, listenerDone)
	} else {
		close(listenerDone)
	}
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
			<-listenerDone
			return
		case <-notifications:
			timer.Stop()
		case <-timer.C:
		}
	}
}

func (d *Dispatcher) listenReplayNotifications(
	ctx context.Context,
	acquirer notificationConnectionAcquirer,
	notifications chan<- struct{},
	done chan<- struct{},
) {
	defer close(done)
	connection, err := acquirer.Acquire(ctx)
	if err != nil {
		return
	}
	defer connection.Release()
	if _, err := connection.Exec(ctx, `LISTEN cartulary_collaboration_replay`); err != nil {
		return
	}
	for {
		if _, err := connection.Conn().WaitForNotification(ctx); err != nil {
			return
		}
		select {
		case notifications <- struct{}{}:
		default:
		}
	}
}

// seedTailCursor starts a newly launched API process at the current durable
// high-water mark. Existing events remain available through authenticated
// resume; only events committed after process startup are fanned out as live
// messages.
func (d *Dispatcher) seedTailCursor(ctx context.Context) error {
	rows, err := d.store.db.Query(ctx, `
SELECT incident_id, high_water_stream_seq
  FROM collaboration_incident_stream_cursors
`)
	if err != nil {
		return fmt.Errorf("seed collaboration replay tail cursor: %w", err)
	}
	defer rows.Close()
	cursor := make(map[uuid.UUID]int64)
	for rows.Next() {
		var incidentID uuid.UUID
		var highWater int64
		if err := rows.Scan(&incidentID, &highWater); err != nil {
			return fmt.Errorf("scan collaboration replay tail cursor: %w", err)
		}
		cursor[incidentID] = highWater
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate collaboration replay tail cursor: %w", err)
	}
	d.tailCursor = cursor
	return nil
}

func (d *Dispatcher) RunOnce(ctx context.Context) (processed int, runErr error) {
	if d == nil || d.store == nil || d.store.db == nil || d.broadcaster == nil {
		return 0, errors.New("collaboration dispatcher is not configured")
	}
	d.runMu.Lock()
	defer d.runMu.Unlock()
	defer func() {
		d.recordDispatcherRun(ctx, processed, runErr)
	}()
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	// Direct RunOnce callers are test and operator harnesses. Starting them at
	// zero makes durable replay behavior observable without changing the
	// production Start contract.
	if !d.tailInitialized {
		d.tailCursor = make(map[uuid.UUID]int64)
		d.tailInitialized = true
	}
	now := d.now().UTC()
	sequenced, err := d.sequencePending(ctx, now)
	if err != nil {
		return 0, err
	}
	tailed, err := d.tailReplay(ctx)
	if err != nil {
		return sequenced, err
	}
	if err := d.pruneReplay(ctx, now); err != nil {
		return sequenced + tailed, err
	}
	return sequenced + tailed, nil
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
`, dispatchBatchLimit, incidentID)
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
		if err := validateReplayPayload(intent); err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				return 0, fmt.Errorf("roll back invalid collaboration event sequencing: %w", rollbackErr)
			}
			if retryErr := d.retrySequencingFailure(ctx, intent, now); retryErr != nil {
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
	message := replayMessage(
		uuid.MustParse("00000000-0000-4000-8000-000000000001"),
		intent.IncidentID,
		intent.Family,
		1,
		intent.Payload,
		time.Unix(0, 0).UTC(),
	)
	if intent.Family == EventFamilyRecordChanged {
		if _, err := platformws.RecordChangeFromSequencedMessage(message); err != nil {
			return fmt.Errorf("invalid record-change payload: %w", err)
		}
	}
	return nil
}

func (d *Dispatcher) retrySequencingFailure(ctx context.Context, intent pendingIntent, now time.Time) error {
	tx, err := d.store.db.BeginTx(ctx, pgx.TxOptions{})
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

func (d *Dispatcher) tailReplay(ctx context.Context) (int, error) {
	cursorJSON, err := json.Marshal(d.tailCursor)
	if err != nil {
		return 0, fmt.Errorf("encode collaboration replay tail cursor: %w", err)
	}
	rows, err := d.store.db.Query(ctx, `
SELECT event_id, incident_id, event_family, stream_seq, canonical_payload, emitted_at
  FROM collaboration_replay_events AS event
 WHERE event.stream_seq > COALESCE(
       (($1::jsonb ->> event.incident_id::text)::bigint),
       0
 )
 ORDER BY event.emitted_at, event.incident_id, event.stream_seq
 LIMIT $2
`, cursorJSON, dispatchBatchLimit)
	if err != nil {
		return 0, fmt.Errorf("tail collaboration replay events: %w", err)
	}
	defer rows.Close()

	processed := 0
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
			return processed, fmt.Errorf("scan collaboration replay tail event: %w", err)
		}
		if err := d.broadcaster.DeliverReplayable(
			replayMessage(eventID, incidentID, family, streamSeq, payload, emittedAt),
		); err != nil {
			return processed, fmt.Errorf("deliver collaboration replay tail event: %w", err)
		}
		d.tailCursor[incidentID] = streamSeq
		processed++
	}
	if err := rows.Err(); err != nil {
		return processed, fmt.Errorf("iterate collaboration replay tail events: %w", err)
	}
	return processed, nil
}

func (d *Dispatcher) pruneReplay(ctx context.Context, now time.Time) error {
	_, err := d.store.db.Exec(ctx, `
DELETE FROM collaboration_replay_events
 WHERE emitted_at < $1
`, now.Add(-maxRetentionAge))
	if err != nil {
		return fmt.Errorf("prune collaboration replay maximum age: %w", err)
	}
	_, err = d.store.db.Exec(ctx, `
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
`, now.Add(-platformws.ResumeWindow), maxRetainedEvents, maxRetainedBytes)
	if err != nil {
		return fmt.Errorf("prune collaboration replay capacity: %w", err)
	}
	_, err = d.store.db.Exec(ctx, `
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
	_, err = d.store.db.Exec(ctx, `
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

func (s *Store) RequeueIncident(ctx context.Context, incidentID uuid.UUID, now time.Time) error {
	if s == nil || s.db == nil || incidentID == uuid.Nil {
		return errors.New("collaboration stream store is not configured")
	}
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin collaboration incident requeue: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `
UPDATE collaboration_incident_stream_cursors
   SET failure_count = 0,
       quarantined_at = NULL,
       quarantine_reason = NULL,
       updated_at = $2
 WHERE incident_id = $1
   AND quarantined_at IS NOT NULL
`, incidentID, now)
	if err != nil {
		return fmt.Errorf("release collaboration incident quarantine: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("collaboration incident is not quarantined")
	}
	if _, err := tx.Exec(ctx, `
UPDATE collaboration_event_intents AS intent
   SET attempt_count = 0,
       next_attempt_at = $2,
       last_error_code = NULL,
       updated_at = $2
 WHERE intent.incident_id = $1
   AND intent.dispatch_state = 'pending'
`, incidentID, now); err != nil {
		return fmt.Errorf("requeue collaboration incident: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit collaboration incident requeue: %w", err)
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
