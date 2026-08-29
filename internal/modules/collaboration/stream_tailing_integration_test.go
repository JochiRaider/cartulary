package collaboration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/intenttest"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
)

func runTailingScenarios(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	harness *appsupport.ServerHarness,
	admin flowtest.LoginResult,
	intents privatestream.IntentWriter,
) (uuid.UUID, time.Time) {
	dispatchIncidentID := createDurableStreamIncident(t, harness, admin, "dispatcher")
	dispatchIncidentUUID := uuid.MustParse(dispatchIncidentID)
	clockNow := time.Now().UTC().Add(time.Second)
	appendCommittedIntent(t, pool, intents, requireJobIntent(t, dispatchIncidentUUID, "dispatcher-outage", clockNow))

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
		failingDispatcher := newDispatcherForTest(pool, broadcaster, func() time.Time { return clockNow })
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
		restartedDispatcher := newDispatcherForTest(pool, broadcaster, func() time.Time { return clockNow })
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
		appendCommittedIntent(t, pool, intents, requireJobIntent(t, dispatchIncidentUUID, "no-subscribers", clockNow))
		noSubscriberDispatcher := newDispatcherForTest(pool, &recordingBroadcaster{}, func() time.Time { return clockNow })
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
		appendCommittedIntent(t, pool, intents, requireJobIntent(t, incidentUUID, "duplicate-claim", claimNow))

		firstBroadcaster := &recordingBroadcaster{}
		secondBroadcaster := &recordingBroadcaster{}
		firstDispatcher := newDispatcherForTest(pool, firstBroadcaster, func() time.Time { return claimNow })
		secondDispatcher := newDispatcherForTest(pool, secondBroadcaster, func() time.Time { return claimNow })
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

	t.Run("corrupt durable tail blocks later sequences without cursor advance or reorder", func(t *testing.T) {
		incidentID := uuid.MustParse(createDurableStreamIncident(t, harness, admin, "corrupt-tail"))
		fixtureTime := clockNow.Add(2 * time.Second)
		validPayload := collabtestprotocol.RawPayload(collabtestprotocol.NewIncidentJobProgressPayload(
			"tail-validation",
			incidentID,
			collabprotocol.JobStatusRunning,
			collabprotocol.JobProgress{},
			fixtureTime,
		))
		fixtures := make([]intenttest.PersistedReplayEventFixture, 0, 3)
		for sequence := int64(1); sequence <= 3; sequence++ {
			payload := validPayload
			if sequence == 2 {
				payload = []byte(`{"job_id":"corrupt-tail"}`)
			}
			fixtures = append(fixtures, intenttest.PersistedReplayEventFixture{
				EventID: uuid.New(), IncidentID: incidentID, StreamSeq: sequence,
				IntentKey:   "corrupt-tail:" + string(rune('0'+sequence)),
				EventFamily: privatestream.EventFamilyJobProgress, CanonicalPayload: payload,
				EmittedAt: fixtureTime.Add(time.Duration(sequence) * time.Millisecond),
			})
		}
		intenttest.InsertPersistedReplayEventFixtures(t, pool, fixtures...)

		broadcaster := &recordingBroadcaster{}
		dispatcher := newDispatcherForTest(pool, broadcaster, func() time.Time { return fixtureTime.Add(time.Second) })
		if _, err := dispatcher.RunOnce(ctx); err == nil {
			t.Fatal("corrupt durable tail was accepted")
		}
		if messages := messagesForIncident(broadcaster.snapshot(), incidentID.String()); len(messages) != 0 {
			t.Fatalf("corrupt tail partially fanned out messages: %#v", messages)
		}

		intenttest.ReplacePersistedReplayPayload(t, pool, fixtures[1].EventID, validPayload)
		if _, err := dispatcher.RunOnce(ctx); err != nil {
			t.Fatalf("deliver repaired durable tail: %v", err)
		}
		messages := messagesForIncident(broadcaster.snapshot(), incidentID.String())
		if len(messages) != 3 {
			t.Fatalf("repaired durable tail delivered %d messages want 3: %#v", len(messages), messages)
		}
		for index, message := range messages {
			if message.StreamSeq == nil || *message.StreamSeq != int64(index+1) {
				t.Fatalf("repaired durable tail order = %#v", messages)
			}
		}
	})
	return dispatchIncidentUUID, clockNow
}
