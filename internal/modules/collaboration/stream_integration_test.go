package collaboration_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	collabscenariotest "github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/scenariotest"
	incidentscenariotest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	timelinemodule "github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestDurableIncidentStream_Integration(t *testing.T) {
	runtime := collabscenariotest.StartRuntime(t)
	harness, admin, adminID, atomicIncidentID := setupSocketIncidentWithAdminID(
		t,
		runtime,
		"collaboration-durable-stream",
	)
	ctx := context.Background()
	closeCtx, cancelClose := context.WithTimeout(ctx, 5*time.Second)
	defer cancelClose()
	if err := harness.Server.Runtime.CollaborationDispatcher.Close(closeCtx); err != nil {
		t.Fatalf("stop runtime collaboration dispatcher: %v", err)
	}

	pool := harness.Server.Runtime.Postgres
	store := collaboration.NewStore(pool, nil)
	atomicIncidentUUID := uuid.MustParse(atomicIncidentID)

	t.Run("intent and source state commit or roll back together", func(t *testing.T) {
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `
DELETE FROM collaboration_event_intents WHERE incident_id = $1
`, atomicIncidentUUID)
		})
		var originalTitle string
		if err := pool.QueryRow(ctx, `SELECT title FROM incidents WHERE id = $1`, atomicIncidentUUID).Scan(&originalTitle); err != nil {
			t.Fatalf("load incident title: %v", err)
		}

		tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin failing source transaction: %v", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE incidents SET title = $2 WHERE id = $1`, atomicIncidentUUID, originalTitle+" changed"); err != nil {
			t.Fatalf("update source state: %v", err)
		}
		invalidIntent := requireJobIntent(t, uuid.New(), "atomic-failure", time.Now().UTC())
		if err := store.AppendIntentTx(ctx, tx, invalidIntent); err == nil {
			t.Fatal("intent with missing incident must fail its source transaction")
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("roll back failed intent transaction: %v", err)
		}

		var titleAfterRollback string
		if err := pool.QueryRow(ctx, `SELECT title FROM incidents WHERE id = $1`, atomicIncidentUUID).Scan(&titleAfterRollback); err != nil {
			t.Fatalf("reload incident title: %v", err)
		}
		if titleAfterRollback != originalTitle {
			t.Fatalf("source state survived failed intent insertion: got %q want %q", titleAfterRollback, originalTitle)
		}

		payload := map[string]any{
			"client_txn_id":                   "txn-collaboration-durable-exact-replay",
			"timeline.activity_synopsis_text": "Durable exact replay",
		}
		first := timelineroutetest.CreateRow(t, harness.Server, admin, atomicIncidentID, payload)
		replayResponse := httptestx.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+atomicIncidentID+"/views/"+timelinemodule.TimelineViewSchemaID+"/rows",
			payload,
			httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value),
		)
		replayed := httptestx.RequireSuccessEnvelope(t, replayResponse, http.StatusOK)["data"].(map[string]any)
		firstRecordID := first["row"].(map[string]any)["record_id"].(string)
		if replayed["row"].(map[string]any)["record_id"] != firstRecordID {
			t.Fatalf("exact replay changed record identity: first=%s replay=%v", firstRecordID, replayed["row"])
		}
		var intentCount int
		if err := pool.QueryRow(ctx, `
SELECT count(*)
  FROM collaboration_event_intents
 WHERE incident_id = $1
   AND source_record_id = $2
`, atomicIncidentUUID, uuid.MustParse(firstRecordID)).Scan(&intentCount); err != nil {
			t.Fatalf("count exact replay intents: %v", err)
		}
		if intentCount != 1 {
			t.Fatalf("exact replay intent count = %d want 1", intentCount)
		}
	})

	dispatchIncidentID := createDurableStreamIncident(t, harness, admin, "dispatcher")
	dispatchIncidentUUID := uuid.MustParse(dispatchIncidentID)
	clockNow := time.Now().UTC().Add(time.Second)
	appendCommittedIntent(t, pool, store, requireJobIntent(t, dispatchIncidentUUID, "dispatcher-outage", clockNow))

	t.Run("outage retry and restart reuse event identity and sequence", func(t *testing.T) {
		var replayCount int
		if err := pool.QueryRow(ctx, `
SELECT count(*) FROM collaboration_replay_events WHERE incident_id = $1
`, dispatchIncidentUUID).Scan(&replayCount); err != nil {
			t.Fatalf("count pre-dispatch replay events: %v", err)
		}
		if replayCount != 0 {
			t.Fatalf("dispatcher outage sequenced %d events without a dispatcher", replayCount)
		}

		broadcaster := &recordingBroadcaster{failRemaining: 1}
		failingDispatcher := collaboration.NewDispatcher(store, broadcaster, func() time.Time { return clockNow })
		if _, err := failingDispatcher.RunOnce(ctx); err != nil {
			t.Fatalf("run failing dispatcher: %v", err)
		}
		attempts := broadcaster.snapshot()
		if len(attempts) != 1 {
			t.Fatalf("broadcast attempts after injected failure = %d want 1", len(attempts))
		}
		firstAttempt := attempts[0]

		var (
			storedEventID uuid.UUID
			storedSeq     int64
			nextAttempt   time.Time
			errorCode     *string
		)
		if err := pool.QueryRow(ctx, `
SELECT replay.event_id, replay.stream_seq, intent.next_attempt_at, intent.last_error_code
  FROM collaboration_event_intents AS intent
  JOIN collaboration_replay_events AS replay
    ON replay.event_id = intent.sequenced_event_id
 WHERE intent.intent_key = $1
`, "job_progress:dispatcher-outage").Scan(&storedEventID, &storedSeq, &nextAttempt, &errorCode); err != nil {
			t.Fatalf("load failed durable delivery: %v", err)
		}
		if storedEventID.String() != firstAttempt.EventID || storedSeq != *firstAttempt.StreamSeq ||
			errorCode == nil || *errorCode != "broadcast_failed" {
			t.Fatalf("failed delivery was not retained canonically: event=%s seq=%d error=%v attempt=%#v", storedEventID, storedSeq, errorCode, firstAttempt)
		}

		clockNow = nextAttempt.Add(time.Millisecond)
		restartedDispatcher := collaboration.NewDispatcher(store, broadcaster, func() time.Time { return clockNow })
		if _, err := restartedDispatcher.RunOnce(ctx); err != nil {
			t.Fatalf("run restarted dispatcher: %v", err)
		}
		attempts = broadcaster.snapshot()
		if len(attempts) != 2 {
			t.Fatalf("broadcast attempts after restart = %d want 2", len(attempts))
		}
		if attempts[1].EventID != firstAttempt.EventID || *attempts[1].StreamSeq != *firstAttempt.StreamSeq {
			t.Fatalf("restart changed durable event identity: first=%#v retry=%#v", firstAttempt, attempts[1])
		}
		var delivered bool
		if err := pool.QueryRow(ctx, `
SELECT delivered_at IS NOT NULL
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, "job_progress:dispatcher-outage").Scan(&delivered); err != nil {
			t.Fatalf("load retried delivery state: %v", err)
		}
		if !delivered {
			t.Fatal("restarted dispatcher did not finish retained delivery")
		}

		clockNow = clockNow.Add(time.Second)
		appendCommittedIntent(t, pool, store, requireJobIntent(t, dispatchIncidentUUID, "no-subscribers", clockNow))
		noSubscriberDispatcher := collaboration.NewDispatcher(
			store,
			harness.Server.Runtime.WSHub,
			func() time.Time { return clockNow },
		)
		if _, err := noSubscriberDispatcher.RunOnce(ctx); err != nil {
			t.Fatalf("deliver with no subscribers: %v", err)
		}
		if err := pool.QueryRow(ctx, `
SELECT delivered_at IS NOT NULL
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, "job_progress:no-subscribers").Scan(&delivered); err != nil {
			t.Fatalf("load no-subscriber delivery state: %v", err)
		}
		if !delivered {
			t.Fatal("no-subscriber delivery remained pending")
		}
	})

	t.Run("duplicate claims retain one active delivery lease", func(t *testing.T) {
		incidentID := createDurableStreamIncident(t, harness, admin, "duplicate-claim")
		incidentUUID := uuid.MustParse(incidentID)
		claimNow := clockNow.Add(time.Second)
		appendCommittedIntent(t, pool, store, requireJobIntent(t, incidentUUID, "duplicate-claim", claimNow))

		broadcaster := newBlockingBroadcaster()
		firstDispatcher := collaboration.NewDispatcher(store, broadcaster, func() time.Time { return claimNow })
		secondDispatcher := collaboration.NewDispatcher(store, broadcaster, func() time.Time { return claimNow })
		firstResult := make(chan error, 1)
		go func() {
			_, err := firstDispatcher.RunOnce(ctx)
			firstResult <- err
		}()

		select {
		case <-broadcaster.entered:
		case <-time.After(5 * time.Second):
			t.Fatal("first dispatcher did not claim delivery")
		}
		type dispatcherResult struct {
			processed int
			err       error
		}
		secondResult := make(chan dispatcherResult, 1)
		go func() {
			processed, err := secondDispatcher.RunOnce(ctx)
			secondResult <- dispatcherResult{processed: processed, err: err}
		}()
		var (
			second           dispatcherResult
			duplicateBlocked bool
		)
		select {
		case second = <-secondResult:
		case <-time.After(500 * time.Millisecond):
			duplicateBlocked = true
		}
		close(broadcaster.release)
		if err := <-firstResult; err != nil {
			t.Fatalf("finish first dispatcher: %v", err)
		}
		if duplicateBlocked {
			second = <-secondResult
		}
		if second.err != nil {
			t.Fatalf("run competing dispatcher: %v", second.err)
		}
		if duplicateBlocked || second.processed != 0 {
			t.Fatalf("competing dispatcher acquired an active delivery lease: blocked=%v processed=%d", duplicateBlocked, second.processed)
		}
		if got := broadcaster.count(); got != 1 {
			t.Fatalf("duplicate claim broadcast count = %d want 1", got)
		}
	})

	t.Run("resume tokens are hashed durable and retention is conjunctive", func(t *testing.T) {
		var (
			sessionID        uuid.UUID
			sessionExpiresAt time.Time
		)
		if err := pool.QueryRow(ctx, `
SELECT id, session_expires_at
  FROM user_sessions
 WHERE user_id = $1
   AND revoked_at IS NULL
 ORDER BY created_at DESC
 LIMIT 1
`, adminID).Scan(&sessionID, &sessionExpiresAt); err != nil {
			t.Fatalf("load active session: %v", err)
		}
		tokenNow := time.Now().UTC()
		token, _, err := store.IssueResumeToken(
			ctx,
			sessionID,
			dispatchIncidentUUID,
			"durable-restart-client",
			sessionExpiresAt,
			tokenNow,
		)
		if err != nil {
			t.Fatalf("issue durable resume token: %v", err)
		}
		var storedHash []byte
		if err := pool.QueryRow(ctx, `
SELECT token_hash
  FROM collaboration_resume_tokens
 WHERE session_id = $1
   AND incident_id = $2
   AND client_instance_id = $3
`, sessionID, dispatchIncidentUUID, "durable-restart-client").Scan(&storedHash); err != nil {
			t.Fatalf("load stored resume-token hash: %v", err)
		}
		wantHash := sha256.Sum256([]byte(token))
		if !bytes.Equal(storedHash, wantHash[:]) || bytes.Equal(storedHash, []byte(token)) {
			t.Fatalf("resume token was not stored exclusively by hash")
		}

		restartedStore := collaboration.NewStore(pool, nil)
		replay, err := restartedStore.ReplayMessages(
			ctx,
			sessionID,
			dispatchIncidentUUID,
			"durable-restart-client",
			token,
			0,
			tokenNow,
		)
		if err != nil {
			t.Fatalf("replay after store restart: %v", err)
		}
		if replay.Status != platformws.ResumeStatusReplayed || len(replay.Messages) != 2 {
			t.Fatalf("restart replay result = status %q messages %d want replayed/2", replay.Status, len(replay.Messages))
		}
		if replay.Messages[0].StreamSeq == nil || replay.Messages[1].StreamSeq == nil ||
			*replay.Messages[0].StreamSeq != 1 || *replay.Messages[1].StreamSeq != 2 {
			t.Fatalf("restart replay order = %#v", replay.Messages)
		}
		reset, err := restartedStore.ReplayMessages(
			ctx,
			sessionID,
			dispatchIncidentUUID,
			"another-client",
			token,
			0,
			tokenNow,
		)
		if err != nil {
			t.Fatalf("validate mismatched token: %v", err)
		}
		if reset.Status != platformws.ResumeStatusResetNeeded || len(reset.Messages) != 0 {
			t.Fatalf("mismatched token replayed data: %#v", reset)
		}

		retentionIncidentID := createDurableStreamIncident(t, harness, admin, "retention")
		retentionIncidentUUID := uuid.MustParse(retentionIncidentID)
		retentionNow := time.Now().UTC().Add(time.Hour)
		if _, err := pool.Exec(ctx, `
INSERT INTO collaboration_incident_stream_cursors (
    incident_id, high_water_stream_seq, updated_at
) VALUES ($1, 10002, $2)
`, retentionIncidentUUID, retentionNow); err != nil {
			t.Fatalf("seed retention cursor: %v", err)
		}
		if _, err := pool.Exec(ctx, `
INSERT INTO collaboration_replay_events (
    event_id, incident_id, stream_seq, intent_key, event_family, canonical_payload, emitted_at
)
SELECT gen_random_uuid(),
       $1::uuid,
       sequence,
       'retention:' || $1::uuid::text || ':' || sequence::text,
       'job_progress',
       '{}'::jsonb,
       CASE
           WHEN sequence = 1 THEN $2::timestamptz - interval '5 minutes 1 microsecond'
           WHEN sequence = 2 THEN $2::timestamptz - interval '5 minutes'
           ELSE $2::timestamptz
       END
  FROM generate_series(1, 10002) AS sequence
`, retentionIncidentUUID, retentionNow); err != nil {
			t.Fatalf("seed retention events: %v", err)
		}
		retentionDispatcher := collaboration.NewDispatcher(
			store,
			harness.Server.Runtime.WSHub,
			func() time.Time { return retentionNow },
		)
		if _, err := retentionDispatcher.RunOnce(ctx); err != nil {
			t.Fatalf("prune retained replay events: %v", err)
		}
		var (
			retainedCount int
			firstSequence int64
		)
		if err := pool.QueryRow(ctx, `
SELECT count(*), min(stream_seq)
  FROM collaboration_replay_events
 WHERE incident_id = $1
`, retentionIncidentUUID).Scan(&retainedCount, &firstSequence); err != nil {
			t.Fatalf("load retained replay range: %v", err)
		}
		if retainedCount != platformws.MinimumReplayRetention+1 || firstSequence != 2 {
			t.Fatalf("retained replay range = count %d first %d want %d/2", retainedCount, firstSequence, platformws.MinimumReplayRetention+1)
		}
	})
}

func createDurableStreamIncident(
	t testing.TB,
	harness *collabscenariotest.ServerHarness,
	admin flowtest.LoginResult,
	suffix string,
) string {
	t.Helper()
	incident := incidentscenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-collaboration-durable-" + suffix,
		"incident_key":  "IR-COLLAB-DURABLE-" + suffix,
		"title":         "Collaboration durable " + suffix,
	})
	return incident["incident_id"].(string)
}

func requireJobIntent(t testing.TB, incidentID uuid.UUID, identity string, createdAt time.Time) collaboration.EventIntent {
	t.Helper()
	intent, err := collaboration.NewEventIntent(
		"job_progress:"+identity,
		incidentID,
		collaboration.EventFamilyJobProgress,
		map[string]any{"job_id": identity},
		"job:"+identity,
		0,
		createdAt,
	)
	if err != nil {
		t.Fatalf("create job intent: %v", err)
	}
	return intent
}

func appendCommittedIntent(t testing.TB, pool *pgxpool.Pool, store *collaboration.Store, intent collaboration.EventIntent) {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin intent transaction: %v", err)
	}
	if err := store.AppendIntentTx(ctx, tx, intent); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("append durable intent: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit durable intent: %v", err)
	}
}

type recordingBroadcaster struct {
	mu            sync.Mutex
	failRemaining int
	messages      []platformws.Message
}

func (b *recordingBroadcaster) DeliverReplayable(message platformws.Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = append(b.messages, message)
	if b.failRemaining > 0 {
		b.failRemaining--
		return errors.New("injected broadcast failure")
	}
	return nil
}

func (b *recordingBroadcaster) snapshot() []platformws.Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]platformws.Message(nil), b.messages...)
}

type blockingBroadcaster struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once

	mu       sync.Mutex
	messages []platformws.Message
}

func newBlockingBroadcaster() *blockingBroadcaster {
	return &blockingBroadcaster{
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (b *blockingBroadcaster) DeliverReplayable(message platformws.Message) error {
	b.mu.Lock()
	b.messages = append(b.messages, message)
	b.mu.Unlock()
	b.once.Do(func() { close(b.entered) })
	<-b.release
	return nil
}

func (b *blockingBroadcaster) count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.messages)
}
