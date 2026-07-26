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

	t.Run("intent keys are immutable across exact and divergent replay", func(t *testing.T) {
		intent := requireJobIntent(t, atomicIncidentUUID, "immutable-intent-key", time.Now().UTC())
		appendCommittedIntent(t, pool, store, intent)
		appendCommittedIntent(t, pool, store, intent)

		divergent := intent
		divergent.SourceIdentity = divergent.SourceIdentity + ":divergent"
		tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin divergent intent transaction: %v", err)
		}
		err = store.AppendIntentTx(ctx, tx, divergent)
		if !errors.Is(err, collaboration.ErrIntentKeyCollision) {
			_ = tx.Rollback(ctx)
			t.Fatalf("divergent intent error = %v want ErrIntentKeyCollision", err)
		}
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			t.Fatalf("roll back divergent intent transaction: %v", rollbackErr)
		}

		var count int
		if err := pool.QueryRow(ctx, `
SELECT count(*)
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, intent.IntentKey).Scan(&count); err != nil {
			t.Fatalf("count immutable-key intents: %v", err)
		}
		if count != 1 {
			t.Fatalf("immutable-key intent count = %d want 1", count)
		}
		if _, err := pool.Exec(ctx, `
DELETE FROM collaboration_event_intents
 WHERE intent_key = $1
`, intent.IntentKey); err != nil {
			t.Fatalf("clean up immutable-key intent: %v", err)
		}
	})

	dispatchIncidentID := createDurableStreamIncident(t, harness, admin, "dispatcher")
	dispatchIncidentUUID := uuid.MustParse(dispatchIncidentID)
	clockNow := time.Now().UTC().Add(time.Second)
	appendCommittedIntent(t, pool, store, requireJobIntent(t, dispatchIncidentUUID, "dispatcher-outage", clockNow))

	t.Run("process-local tail retry and restart reuse durable event identity", func(t *testing.T) {
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
		if _, err := failingDispatcher.RunOnce(ctx); err == nil {
			t.Fatal("failing local tailer did not report injected delivery failure")
		}
		attempts := broadcaster.snapshot()
		if len(attempts) != 1 {
			t.Fatalf("broadcast attempts after injected failure = %d want 1", len(attempts))
		}
		firstAttempt := attempts[0]

		var (
			storedEventID uuid.UUID
			storedSeq     int64
			state         string
		)
		if err := pool.QueryRow(ctx, `
SELECT replay.event_id, replay.stream_seq, intent.dispatch_state
  FROM collaboration_event_intents AS intent
  JOIN collaboration_replay_events AS replay
    ON replay.event_id = intent.sequenced_event_id
 WHERE intent.intent_key = $1
`, "job_progress:dispatcher-outage").Scan(&storedEventID, &storedSeq, &state); err != nil {
			t.Fatalf("load durable sequenced event: %v", err)
		}
		if storedEventID.String() != firstAttempt.EventID || storedSeq != *firstAttempt.StreamSeq || state != "sequenced" {
			t.Fatalf("failed local delivery changed durable event: event=%s seq=%d state=%s attempt=%#v", storedEventID, storedSeq, state, firstAttempt)
		}

		clockNow = clockNow.Add(time.Millisecond)
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
		var sequenced bool
		if err := pool.QueryRow(ctx, `
SELECT dispatch_state = 'sequenced'
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, "job_progress:no-subscribers").Scan(&sequenced); err != nil {
			t.Fatalf("load no-subscriber sequence state: %v", err)
		}
		if !sequenced {
			t.Fatal("no-subscriber event was not durably sequenced")
		}
	})

	t.Run("every API process tails the same durable event", func(t *testing.T) {
		incidentID := createDurableStreamIncident(t, harness, admin, "duplicate-claim")
		incidentUUID := uuid.MustParse(incidentID)
		claimNow := clockNow.Add(time.Second)
		appendCommittedIntent(t, pool, store, requireJobIntent(t, incidentUUID, "duplicate-claim", claimNow))

		firstBroadcaster := &recordingBroadcaster{}
		secondBroadcaster := &recordingBroadcaster{}
		firstDispatcher := collaboration.NewDispatcher(store, firstBroadcaster, func() time.Time { return claimNow })
		secondDispatcher := collaboration.NewDispatcher(store, secondBroadcaster, func() time.Time { return claimNow })
		if _, err := firstDispatcher.RunOnce(ctx); err != nil {
			t.Fatalf("run first process tailer: %v", err)
		}
		if _, err := secondDispatcher.RunOnce(ctx); err != nil {
			t.Fatalf("run second process tailer: %v", err)
		}
		first := messagesForIncident(firstBroadcaster.snapshot(), incidentID)
		second := messagesForIncident(secondBroadcaster.snapshot(), incidentID)
		if len(first) != 1 || len(second) != 1 {
			t.Fatalf("process-local deliveries = first %d second %d want 1/1", len(first), len(second))
		}
		if first[0].EventID != second[0].EventID || *first[0].StreamSeq != *second[0].StreamSeq {
			t.Fatalf("process-local tailers observed different durable events: first=%#v second=%#v", first[0], second[0])
		}
	})

	t.Run("deterministic payload failures quarantine only their incident and requeue explicitly", func(t *testing.T) {
		poisonIncidentID := createDurableStreamIncident(t, harness, admin, "poison")
		poisonIncidentUUID := uuid.MustParse(poisonIncidentID)
		quarantineNow := clockNow.Add(2 * time.Second)
		invalidIntent, err := collaboration.NewEventIntent(
			"record_changed:invalid-payload",
			poisonIncidentUUID,
			collaboration.EventFamilyRecordChanged,
			map[string]any{"not": "a record change"},
			"record:invalid-payload",
			0,
			quarantineNow,
		)
		if err != nil {
			t.Fatalf("create invalid legacy intent fixture: %v", err)
		}
		appendCommittedIntent(t, pool, store, invalidIntent)

		dispatcher := collaboration.NewDispatcher(
			store,
			&recordingBroadcaster{},
			func() time.Time { return quarantineNow },
		)
		for attempt := 1; attempt <= 12; attempt++ {
			if _, err := dispatcher.RunOnce(ctx); err != nil {
				t.Fatalf("run deterministic sequencing attempt %d: %v", attempt, err)
			}
			var nextAttempt time.Time
			if err := pool.QueryRow(ctx, `
SELECT next_attempt_at
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, invalidIntent.IntentKey).Scan(&nextAttempt); err != nil {
				t.Fatalf("load deterministic sequencing retry %d: %v", attempt, err)
			}
			quarantineNow = nextAttempt.Add(time.Millisecond)
		}
		var (
			failureCount int
			reason       *string
		)
		if err := pool.QueryRow(ctx, `
SELECT failure_count, quarantine_reason
  FROM collaboration_incident_stream_cursors
 WHERE incident_id = $1
`, poisonIncidentUUID).Scan(&failureCount, &reason); err != nil {
			t.Fatalf("load quarantined incident cursor: %v", err)
		}
		if failureCount != 12 || reason == nil || *reason != "invalid_event_payload" {
			t.Fatalf("quarantine state = count %d reason %v want 12/invalid_event_payload", failureCount, reason)
		}

		healthyIncidentID := createDurableStreamIncident(t, harness, admin, "healthy-beside-poison")
		healthyIncidentUUID := uuid.MustParse(healthyIncidentID)
		appendCommittedIntent(t, pool, store, requireJobIntent(t, healthyIncidentUUID, "healthy-beside-poison", quarantineNow))
		if _, err := dispatcher.RunOnce(ctx); err != nil {
			t.Fatalf("sequence healthy incident beside quarantine: %v", err)
		}
		var healthyReplayCount int
		if err := pool.QueryRow(ctx, `
SELECT count(*)
  FROM collaboration_replay_events
 WHERE incident_id = $1
`, healthyIncidentUUID).Scan(&healthyReplayCount); err != nil {
			t.Fatalf("count healthy replay events: %v", err)
		}
		if healthyReplayCount != 1 {
			t.Fatalf("healthy incident replay count = %d want 1", healthyReplayCount)
		}

		if _, err := pool.Exec(ctx, `
DELETE FROM collaboration_event_intents
 WHERE intent_key = $1
`, invalidIntent.IntentKey); err != nil {
			t.Fatalf("remove repaired invalid intent fixture: %v", err)
		}
		if err := store.RequeueIncident(ctx, poisonIncidentUUID, quarantineNow.Add(time.Second)); err != nil {
			t.Fatalf("release repaired incident quarantine: %v", err)
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
           WHEN sequence = 1 THEN $2::timestamptz - interval '24 hours 1 microsecond'
           WHEN sequence = 2 THEN $2::timestamptz - interval '24 hours'
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
		if retainedCount != 10001 || firstSequence != 2 {
			t.Fatalf("retained replay range = count %d first %d want 10001/2", retainedCount, firstSequence)
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

func messagesForIncident(messages []platformws.Message, incidentID string) []platformws.Message {
	filtered := make([]platformws.Message, 0, len(messages))
	for _, message := range messages {
		if message.IncidentID == incidentID {
			filtered = append(filtered, message)
		}
	}
	return filtered
}
