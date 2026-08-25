package collaboration_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	collabprotocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	incidentscenariotest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	timelinemodule "github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestDurableIncidentStream_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness, admin, adminID, atomicIncidentID := setupSocketIncidentWithAdminID(
		t,
		runtime,
		"collaboration-durable-stream",
	)
	ctx := context.Background()
	closeCtx, cancelClose := context.WithTimeout(ctx, 5*time.Second)
	defer cancelClose()
	if err := harness.Collaboration.CloseDispatcher(closeCtx); err != nil {
		t.Fatalf("stop runtime collaboration dispatcher: %v", err)
	}

	pool := harness.Pool
	replay := privatestream.NewPostgresStream(pool, nil)
	intents := replay
	recovery := collaboration.NewRecoveryCapability(pool)
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
		if err := intents.AppendIntentTx(ctx, tx, invalidIntent); err == nil {
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
		appendCommittedIntent(t, pool, intents, intent)
		appendCommittedIntent(t, pool, intents, intent)

		divergent := intent
		divergent.SourceIdentity = divergent.SourceIdentity + ":divergent"
		tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin divergent intent transaction: %v", err)
		}
		err = intents.AppendIntentTx(ctx, tx, divergent)
		if !errors.Is(err, privatestream.ErrIntentKeyCollision) {
			_ = tx.Rollback(ctx)
			t.Fatalf("divergent intent error = %v want private stream collision", err)
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

	t.Run("deterministic payload failures quarantine only their incident and requeue explicitly", func(t *testing.T) {
		poisonIncidentID := createDurableStreamIncident(t, harness, admin, "poison")
		poisonIncidentUUID := uuid.MustParse(poisonIncidentID)
		quarantineNow := clockNow.Add(2 * time.Second)
		invalidIntent := privatestream.EventIntent{
			IntentKey:        "record_changed:invalid-payload",
			IncidentID:       poisonIncidentUUID,
			EventFamily:      privatestream.EventFamilyRecordChanged,
			CanonicalPayload: json.RawMessage(`{"not":"a record change"}`),
			SourceIdentity:   "record:invalid-payload",
			CreatedAt:        quarantineNow,
		}
		appendLegacyIntent(t, pool, invalidIntent)

		dispatcher := newDispatcherForTest(
			pool,
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
		appendCommittedIntent(t, pool, intents, requireJobIntent(t, healthyIncidentUUID, "healthy-beside-poison", quarantineNow))
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
		if _, err := requeueCollaborationIncident(recovery, ctx, poisonIncidentUUID, quarantineNow.Add(time.Second)); err != nil {
			t.Fatalf("release repaired incident quarantine: %v", err)
		}
	})

	t.Run("requeue preconditions collapse missing and non-quarantined incidents", func(t *testing.T) {
		notQuarantinedID := uuid.MustParse(createDurableStreamIncident(t, harness, admin, "requeue-not-quarantined"))
		for _, incidentID := range []uuid.UUID{notQuarantinedID, uuid.New()} {
			_, err := requeueCollaborationIncident(recovery, ctx, incidentID, clockNow.Add(30*time.Second))
			if requeueFailureKind(err) != collaboration.RequeueFailureIncidentNotQuarantined {
				t.Fatalf("requeue precondition for %s = %v", incidentID, err)
			}
		}
	})

	t.Run("requeue is single winner, attributed, and preserves pending event identity", func(t *testing.T) {
		incidentID, intentKey, payload := seedCurrentQuarantinedIncident(
			t,
			harness,
			admin,
			pool,
			"requeue-concurrent",
			false,
			clockNow.Add(31*time.Second),
		)
		operationIDs := []uuid.UUID{
			uuid.MustParse("30000000-0000-0000-0000-000000000001"),
			uuid.MustParse("30000000-0000-0000-0000-000000000002"),
		}
		start := make(chan struct{})
		type requeueAttempt struct {
			result      collaboration.RequeueResult
			error       error
			operationID uuid.UUID
		}
		attempts := make(chan requeueAttempt, 2)
		for _, operationID := range operationIDs {
			go func() {
				<-start
				result, err := collaboration.NewRecoveryCapability(pool).RequeueIncident(
					context.Background(),
					collaboration.RequeueRequest{
						OperationID: operationID,
						IncidentID:  incidentID,
						MutatedAt:   clockNow.Add(32 * time.Second),
					},
				)
				attempts <- requeueAttempt{result: result, error: err, operationID: operationID}
			}()
		}
		close(start)
		succeeded := 0
		rejected := 0
		committedOperationID := uuid.Nil
		for range 2 {
			attempt := <-attempts
			if attempt.error == nil {
				succeeded++
				committedOperationID = attempt.operationID
				if attempt.result.RequeuedIntentCount != 1 {
					t.Fatalf("committed requeue count = %d want 1", attempt.result.RequeuedIntentCount)
				}
			} else if requeueFailureKind(attempt.error) == collaboration.RequeueFailureIncidentNotQuarantined {
				rejected++
			} else {
				t.Fatalf("unexpected concurrent requeue error: %v", attempt.error)
			}
		}
		if succeeded != 1 || rejected != 1 {
			t.Fatalf("concurrent requeue outcomes = success %d rejected %d want 1/1", succeeded, rejected)
		}

		var (
			failureCount     int
			quarantinedAt    *time.Time
			quarantineReason *string
			attemptCount     int
			lastErrorCode    *string
			dispatchState    string
			storedPayload    []byte
		)
		if err := pool.QueryRow(ctx, `
SELECT cursor.failure_count,
       cursor.quarantined_at,
       cursor.quarantine_reason,
       intent.attempt_count,
       intent.last_error_code,
       intent.dispatch_state,
       intent.canonical_payload
  FROM collaboration_incident_stream_cursors AS cursor
  JOIN collaboration_event_intents AS intent
    ON intent.incident_id = cursor.incident_id
 WHERE cursor.incident_id = $1
   AND intent.intent_key = $2
`, incidentID, intentKey).Scan(
			&failureCount,
			&quarantinedAt,
			&quarantineReason,
			&attemptCount,
			&lastErrorCode,
			&dispatchState,
			&storedPayload,
		); err != nil {
			t.Fatalf("load concurrent requeue result: %v", err)
		}
		if failureCount != 0 || quarantinedAt != nil || quarantineReason != nil || attemptCount != 0 || lastErrorCode != nil {
			t.Fatalf("requeue retry state was not reset exactly: failure=%d quarantine=%v/%v attempt=%d last_error=%v", failureCount, quarantinedAt, quarantineReason, attemptCount, lastErrorCode)
		}
		if dispatchState != "pending" || !jsonValuesEqual(storedPayload, payload) {
			t.Fatalf("requeue changed event identity state: dispatch=%q payload=%s want pending/%s", dispatchState, storedPayload, payload)
		}
		var (
			journalCount       int
			journalOperationID string
			journalBefore      []byte
			journalAfter       []byte
		)
		if err := pool.QueryRow(ctx, `
SELECT count(*) OVER (), client_txn_id, before_json, after_json
  FROM deployment_admin_audit_events
 WHERE incident_id = $1
   AND event_source = 'operator'
   AND event_kind = 'collaboration_incident_requeued'
`, incidentID).Scan(&journalCount, &journalOperationID, &journalBefore, &journalAfter); err != nil {
			t.Fatalf("load collaboration requeue journal row: %v", err)
		}
		var beforeSummary map[string]any
		var afterSummary map[string]any
		if err := json.Unmarshal(journalBefore, &beforeSummary); err != nil {
			t.Fatalf("decode collaboration requeue before journal: %v", err)
		}
		if err := json.Unmarshal(journalAfter, &afterSummary); err != nil {
			t.Fatalf("decode collaboration requeue after journal: %v", err)
		}
		if journalCount != 1 || journalOperationID != committedOperationID.String() ||
			beforeSummary["failure_count"] != float64(12) || beforeSummary["quarantine_reason"] != "invalid_event_payload" ||
			afterSummary["quarantine_released"] != true || afterSummary["retry_state_reset"] != true || afterSummary["requeued_intent_count"] != float64(1) {
			t.Fatalf("collaboration requeue attribution changed: count=%d operation=%q before=%v after=%v", journalCount, journalOperationID, beforeSummary, afterSummary)
		}
		if _, err := requeueCollaborationIncident(recovery, ctx, incidentID, clockNow.Add(33*time.Second)); requeueFailureKind(err) != collaboration.RequeueFailureIncidentNotQuarantined {
			t.Fatalf("second committed requeue result = %v", err)
		}
	})

	t.Run("requeue rejects an unrepaired pending payload without changing state", func(t *testing.T) {
		incidentID, intentKey, payload := seedCurrentQuarantinedIncident(
			t,
			harness,
			admin,
			pool,
			"requeue-unrepaired-current-state",
			true,
			clockNow.Add(34*time.Second),
		)
		if _, err := requeueCollaborationIncident(recovery, ctx, incidentID, clockNow.Add(35*time.Second)); requeueFailureKind(err) != collaboration.RequeueFailureRepairNotVerified {
			t.Fatalf("unrepaired requeue result = %v", err)
		}
		var (
			dispatchState string
			storedPayload []byte
			intentCount   int
			failureCount  int
			quarantinedAt *time.Time
			attemptCount  int
		)
		if err := pool.QueryRow(ctx, `
SELECT max(intent.dispatch_state),
       max(intent.canonical_payload::text)::jsonb,
       count(*),
       max(cursor.failure_count),
       max(cursor.quarantined_at),
       max(intent.attempt_count)
  FROM collaboration_event_intents AS intent
  JOIN collaboration_incident_stream_cursors AS cursor
    ON cursor.incident_id = intent.incident_id
 WHERE intent.incident_id = $1
   AND intent.intent_key = $2
`, incidentID, intentKey).Scan(&dispatchState, &storedPayload, &intentCount, &failureCount, &quarantinedAt, &attemptCount); err != nil {
			t.Fatalf("load unrepaired intent: %v", err)
		}
		if intentCount != 1 || dispatchState != "pending" || !jsonValuesEqual(storedPayload, payload) || failureCount != 12 || quarantinedAt == nil || attemptCount != 12 {
			t.Fatalf("unrepaired requeue changed state: count=%d state=%q payload=%s failure=%d quarantine=%v attempt=%d", intentCount, dispatchState, storedPayload, failureCount, quarantinedAt, attemptCount)
		}
	})

	t.Run("requeue rolls back cursor release when intent reset fails", func(t *testing.T) {
		incidentID, intentKey, _ := seedCurrentQuarantinedIncident(
			t,
			harness,
			admin,
			pool,
			"requeue-rollback",
			false,
			clockNow.Add(36*time.Second),
		)
		if _, err := pool.Exec(ctx, `
CREATE OR REPLACE FUNCTION cartulary_test_fail_requeue_intent_update()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
BEGIN
    RAISE EXCEPTION 'injected requeue failure';
END
$function$
`); err != nil {
			t.Fatalf("install requeue rollback function: %v", err)
		}
		if _, err := pool.Exec(ctx, `
CREATE TRIGGER cartulary_test_fail_requeue_intent_update
BEFORE UPDATE ON collaboration_event_intents
FOR EACH ROW EXECUTE FUNCTION cartulary_test_fail_requeue_intent_update()
`); err != nil {
			t.Fatalf("install requeue rollback trigger: %v", err)
		}
		defer func() {
			_, _ = pool.Exec(context.Background(), `DROP TRIGGER IF EXISTS cartulary_test_fail_requeue_intent_update ON collaboration_event_intents`)
			_, _ = pool.Exec(context.Background(), `DROP FUNCTION IF EXISTS cartulary_test_fail_requeue_intent_update()`)
		}()
		if _, err := requeueCollaborationIncident(recovery, ctx, incidentID, clockNow.Add(37*time.Second)); requeueFailureKind(err) != collaboration.RequeueFailureTransaction {
			t.Fatal("injected intent reset failure unexpectedly committed")
		}
		var (
			failureCount  int
			quarantinedAt *time.Time
			attemptCount  int
		)
		if err := pool.QueryRow(ctx, `
SELECT cursor.failure_count, cursor.quarantined_at, intent.attempt_count
  FROM collaboration_incident_stream_cursors AS cursor
  JOIN collaboration_event_intents AS intent
    ON intent.incident_id = cursor.incident_id
 WHERE cursor.incident_id = $1
   AND intent.intent_key = $2
`, incidentID, intentKey).Scan(&failureCount, &quarantinedAt, &attemptCount); err != nil {
			t.Fatalf("load rolled-back requeue state: %v", err)
		}
		if failureCount != 12 || quarantinedAt == nil || attemptCount != 12 {
			t.Fatalf("partial requeue survived rollback: failure=%d quarantine=%v attempt=%d", failureCount, quarantinedAt, attemptCount)
		}
	})

	t.Run("requeue journal failure rolls back cursor and intent mutation", func(t *testing.T) {
		incidentID, intentKey, _ := seedCurrentQuarantinedIncident(
			t,
			harness,
			admin,
			pool,
			"requeue-journal-rollback",
			false,
			clockNow.Add(38*time.Second),
		)
		if _, err := pool.Exec(ctx, `
CREATE OR REPLACE FUNCTION cartulary_test_fail_requeue_journal_insert()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
BEGIN
    IF NEW.event_kind = 'collaboration_incident_requeued' THEN
        RAISE EXCEPTION 'injected requeue journal failure';
    END IF;
    RETURN NEW;
END
$function$
`); err != nil {
			t.Fatalf("install requeue journal rollback function: %v", err)
		}
		if _, err := pool.Exec(ctx, `
CREATE TRIGGER cartulary_test_fail_requeue_journal_insert
BEFORE INSERT ON deployment_admin_audit_events
FOR EACH ROW EXECUTE FUNCTION cartulary_test_fail_requeue_journal_insert()
`); err != nil {
			t.Fatalf("install requeue journal rollback trigger: %v", err)
		}
		defer func() {
			_, _ = pool.Exec(context.Background(), `DROP TRIGGER IF EXISTS cartulary_test_fail_requeue_journal_insert ON deployment_admin_audit_events`)
			_, _ = pool.Exec(context.Background(), `DROP FUNCTION IF EXISTS cartulary_test_fail_requeue_journal_insert()`)
		}()
		if _, err := requeueCollaborationIncident(recovery, ctx, incidentID, clockNow.Add(39*time.Second)); requeueFailureKind(err) != collaboration.RequeueFailureTransaction {
			t.Fatalf("injected journal failure result = %v", err)
		}
		var (
			failureCount  int
			quarantinedAt *time.Time
			attemptCount  int
		)
		if err := pool.QueryRow(ctx, `
SELECT cursor.failure_count, cursor.quarantined_at, intent.attempt_count
  FROM collaboration_incident_stream_cursors AS cursor
  JOIN collaboration_event_intents AS intent
    ON intent.incident_id = cursor.incident_id
 WHERE cursor.incident_id = $1
   AND intent.intent_key = $2
`, incidentID, intentKey).Scan(&failureCount, &quarantinedAt, &attemptCount); err != nil {
			t.Fatalf("load journal-rolled-back requeue state: %v", err)
		}
		if failureCount != 12 || quarantinedAt == nil || attemptCount != 12 {
			t.Fatalf("journal failure left partial requeue: failure=%d quarantine=%v attempt=%d", failureCount, quarantinedAt, attemptCount)
		}
	})

	t.Run("requeue commit failure reports unknown outcome and rolls back the uncommitted transaction", func(t *testing.T) {
		incidentID, intentKey, _ := seedCurrentQuarantinedIncident(
			t,
			harness,
			admin,
			pool,
			"requeue-commit-unknown",
			false,
			clockNow.Add(40*time.Second),
		)
		service := collaboration.NewRecoveryCapability(requeueCommitFailureDB{DB: pool})
		if _, err := requeueCollaborationIncident(service, ctx, incidentID, clockNow.Add(41*time.Second)); requeueFailureKind(err) != collaboration.RequeueFailureCommitOutcomeUnknown {
			t.Fatalf("injected commit failure result = %v", err)
		}
		var (
			failureCount  int
			quarantinedAt *time.Time
			attemptCount  int
			journalCount  int
		)
		if err := pool.QueryRow(ctx, `
SELECT cursor.failure_count,
       cursor.quarantined_at,
       intent.attempt_count,
       (
           SELECT count(*)
             FROM deployment_admin_audit_events AS audit
            WHERE audit.incident_id = cursor.incident_id
              AND audit.event_source = 'operator'
              AND audit.event_kind = 'collaboration_incident_requeued'
       )
  FROM collaboration_incident_stream_cursors AS cursor
  JOIN collaboration_event_intents AS intent
    ON intent.incident_id = cursor.incident_id
 WHERE cursor.incident_id = $1
   AND intent.intent_key = $2
`, incidentID, intentKey).Scan(&failureCount, &quarantinedAt, &attemptCount, &journalCount); err != nil {
			t.Fatalf("load commit-unknown requeue state: %v", err)
		}
		if failureCount != 12 || quarantinedAt == nil || attemptCount != 12 || journalCount != 0 {
			t.Fatalf("commit failure left partial state: failure=%d quarantine=%v attempt=%d journal=%d", failureCount, quarantinedAt, attemptCount, journalCount)
		}
	})

	t.Run("requeue honors a cancelled caller before transaction admission", func(t *testing.T) {
		incidentID, _, _ := seedCurrentQuarantinedIncident(
			t,
			harness,
			admin,
			pool,
			"requeue-cancelled",
			false,
			clockNow.Add(38*time.Second),
		)
		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := requeueCollaborationIncident(recovery, cancelled, incidentID, clockNow.Add(39*time.Second)); requeueFailureKind(err) != collaboration.RequeueFailureCancelled {
			t.Fatalf("cancelled requeue result = %v", err)
		}
		var quarantinedAt *time.Time
		if err := pool.QueryRow(ctx, `
SELECT quarantined_at
  FROM collaboration_incident_stream_cursors
 WHERE incident_id = $1
`, incidentID).Scan(&quarantinedAt); err != nil {
			t.Fatalf("load cancelled requeue cursor: %v", err)
		}
		if quarantinedAt == nil {
			t.Fatal("cancelled requeue released quarantine")
		}
	})

	t.Run("legacy invalid job and extension payloads fail semantic validation before sequencing", func(t *testing.T) {
		cases := []struct {
			name    string
			family  string
			payload json.RawMessage
		}{
			{name: "job", family: privatestream.EventFamilyJobProgress, payload: json.RawMessage(`{"job_id":"legacy-invalid"}`)},
			{name: "extension", family: privatestream.EventFamilyExtensionResourceChange, payload: json.RawMessage(`{"extension_profile_id":"network_flow_activity","resource_kind":"network_flow_table","resource_id":"nft_invalid","change_kind":"remove","reason_code":"renamed"}`)},
		}
		for index, testCase := range cases {
			t.Run(testCase.name, func(t *testing.T) {
				incidentID := createDurableStreamIncident(t, harness, admin, "invalid-"+testCase.name)
				incidentUUID := uuid.MustParse(incidentID)
				attemptedAt := clockNow.Add(time.Duration(20+index) * time.Second)
				intent := privatestream.EventIntent{
					IntentKey:        "legacy-invalid:" + testCase.name,
					IncidentID:       incidentUUID,
					EventFamily:      testCase.family,
					CanonicalPayload: testCase.payload,
					SourceIdentity:   "legacy:" + testCase.name,
					CreatedAt:        attemptedAt,
				}
				appendLegacyIntent(t, pool, intent)
				dispatcher := newDispatcherForTest(pool, &recordingBroadcaster{}, func() time.Time { return attemptedAt.Add(time.Second) })
				if _, err := dispatcher.RunOnce(ctx); err != nil {
					t.Fatalf("attempt legacy invalid sequencing: %v", err)
				}
				var (
					lastError   *string
					replayCount int
				)
				if err := pool.QueryRow(ctx, `
SELECT last_error_code,
       (SELECT count(*) FROM collaboration_replay_events WHERE incident_id = $2)
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, intent.IntentKey, incidentUUID).Scan(&lastError, &replayCount); err != nil {
					t.Fatalf("load legacy invalid sequencing result: %v", err)
				}
				if lastError == nil || *lastError != "invalid_event_payload" || replayCount != 0 {
					t.Fatalf("legacy invalid sequencing = error %v replay_count %d", lastError, replayCount)
				}
				if _, err := pool.Exec(ctx, `
DELETE FROM collaboration_event_intents
 WHERE intent_key = $1
`, intent.IntentKey); err != nil {
					t.Fatalf("remove legacy invalid fixture: %v", err)
				}
			})
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
		token, _, err := replay.IssueResumeToken(
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

		restartedReplay := privatestream.NewPostgresStream(pool, nil)
		replay, err := restartedReplay.ReplayMessages(
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
		if replay.Status != collabprotocol.ResumeStatusReplayed || len(replay.Messages) != 2 {
			t.Fatalf("restart replay result = status %q messages %d want replayed/2", replay.Status, len(replay.Messages))
		}
		if replay.Messages[0].StreamSeq == nil || replay.Messages[1].StreamSeq == nil ||
			*replay.Messages[0].StreamSeq != 1 || *replay.Messages[1].StreamSeq != 2 {
			t.Fatalf("restart replay order = %#v", replay.Messages)
		}
		reset, err := restartedReplay.ReplayMessages(
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
		if reset.Status != collabprotocol.ResumeStatusResetNeeded || len(reset.Messages) != 0 {
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
		retentionDispatcher := newDispatcherForTest(pool, &recordingBroadcaster{}, func() time.Time { return retentionNow })
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
	harness *appsupport.ServerHarness,
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

func seedCurrentQuarantinedIncident(
	t testing.TB,
	harness *appsupport.ServerHarness,
	admin flowtest.LoginResult,
	pool *pgxpool.Pool,
	suffix string,
	invalidPayload bool,
	seededAt time.Time,
) (uuid.UUID, string, []byte) {
	t.Helper()
	incidentID := uuid.MustParse(createDurableStreamIncident(t, harness, admin, suffix))
	intent := requireJobIntent(t, incidentID, suffix, seededAt)
	if invalidPayload {
		intent.EventFamily = privatestream.EventFamilyRecordChanged
		intent.CanonicalPayload = json.RawMessage(`{"not":"a record change"}`)
		intent.IntentKey = "record_changed:" + suffix
		intent.SourceIdentity = "record:" + suffix
	}
	appendLegacyIntent(t, pool, intent)
	if _, err := pool.Exec(context.Background(), `
INSERT INTO collaboration_incident_stream_cursors (
    incident_id, high_water_stream_seq, failure_count, quarantined_at,
    quarantine_reason, updated_at
) VALUES ($1, 0, 12, $2, 'invalid_event_payload', $2)
ON CONFLICT (incident_id) DO UPDATE
SET failure_count = 12,
    quarantined_at = EXCLUDED.quarantined_at,
    quarantine_reason = EXCLUDED.quarantine_reason,
	updated_at = EXCLUDED.updated_at
`, incidentID, seededAt); err != nil {
		t.Fatalf("seed quarantined collaboration cursor: %v", err)
	}
	if _, err := pool.Exec(context.Background(), `
UPDATE collaboration_event_intents
   SET attempt_count = 12,
       last_error_code = 'invalid_event_payload',
       next_attempt_at = $2,
       updated_at = $2
 WHERE incident_id = $1
   AND intent_key = $3
`, incidentID, seededAt, intent.IntentKey); err != nil {
		t.Fatalf("seed quarantined collaboration intent: %v", err)
	}
	return incidentID, intent.IntentKey, append([]byte(nil), intent.CanonicalPayload...)
}

func jsonValuesEqual(left []byte, right []byte) bool {
	var leftValue any
	var rightValue any
	if json.Unmarshal(left, &leftValue) != nil || json.Unmarshal(right, &rightValue) != nil {
		return false
	}
	return reflect.DeepEqual(leftValue, rightValue)
}

func requeueCollaborationIncident(
	service collaboration.RecoveryCapability,
	ctx context.Context,
	incidentID uuid.UUID,
	mutatedAt time.Time,
) (collaboration.RequeueResult, error) {
	return service.RequeueIncident(ctx, collaboration.RequeueRequest{
		OperationID: uuid.New(),
		IncidentID:  incidentID,
		MutatedAt:   mutatedAt.UTC(),
	})
}

func requeueFailureKind(err error) collaboration.RequeueFailureKind {
	var failure *collaboration.RequeueFailure
	if errors.As(err, &failure) {
		return failure.Kind
	}
	return ""
}

type requeueCommitFailureDB struct {
	postgres.DB
}

func (db requeueCommitFailureDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return requeueCommitFailureTx{Tx: tx}, nil
}

type requeueCommitFailureTx struct {
	pgx.Tx
}

func (requeueCommitFailureTx) Commit(context.Context) error {
	return errors.New("injected collaboration requeue commit failure")
}

func requireJobIntent(t testing.TB, incidentID uuid.UUID, identity string, createdAt time.Time) privatestream.EventIntent {
	t.Helper()
	intent, err := privatestream.NewEventIntent(
		"job_progress:"+identity,
		incidentID,
		privatestream.EventFamilyJobProgress,
		collabprotocol.NewIncidentJobProgressPayload(
			identity,
			incidentID,
			collabprotocol.JobStatusQueued,
			collabprotocol.JobProgress{Completed: 0},
			createdAt,
		),
		"job:"+identity,
		0,
		createdAt,
	)
	if err != nil {
		t.Fatalf("create job intent: %v", err)
	}
	return intent
}

func appendLegacyIntent(t testing.TB, pool *pgxpool.Pool, intent privatestream.EventIntent) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
INSERT INTO collaboration_event_intents (
    intent_key, incident_id, event_family, canonical_payload, source_identity,
    mutation_ordinal, next_attempt_at, created_at, updated_at
) VALUES ($1, $2, $3, $4::jsonb, $5, 0, $6, $6, $6)
`, intent.IntentKey, intent.IncidentID, intent.EventFamily, intent.CanonicalPayload, intent.SourceIdentity, intent.CreatedAt); err != nil {
		t.Fatalf("append legacy intent: %v", err)
	}
}

func appendCommittedIntent(t testing.TB, pool *pgxpool.Pool, intents *privatestream.PostgresStream, intent privatestream.EventIntent) {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin intent transaction: %v", err)
	}
	if err := intents.AppendIntentTx(ctx, tx, intent); err != nil {
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
	messages      []collabprotocol.Message
}

func newDispatcherForTest(
	db postgres.DB,
	broadcaster interface {
		DeliverReplayable(collabprotocol.Message) error
	},
	now func() time.Time,
) *privatestream.Dispatcher {
	return privatestream.NewDispatcher(privatestream.NewPostgresStream(db, now), broadcaster, now)
}

func (b *recordingBroadcaster) DeliverReplayable(message collabprotocol.Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = append(b.messages, message)
	if b.failRemaining > 0 {
		b.failRemaining--
		return errors.New("injected broadcast failure")
	}
	return nil
}

func (b *recordingBroadcaster) snapshot() []collabprotocol.Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]collabprotocol.Message(nil), b.messages...)
}

func messagesForIncident(messages []collabprotocol.Message, incidentID string) []collabprotocol.Message {
	filtered := make([]collabprotocol.Message, 0, len(messages))
	for _, message := range messages {
		if message.IncidentID == incidentID {
			filtered = append(filtered, message)
		}
	}
	return filtered
}
