package collaboration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/intenttest"
)

func runSequencingRecoveryScenarios(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	harness *appsupport.ServerHarness,
	admin flowtest.LoginResult,
	intents privatestream.IntentWriter,
	recovery collaboration.RecoveryCapability,
	clockNow time.Time,
) {
	t.Run("deterministic payload failures quarantine only their incident and requeue explicitly", func(t *testing.T) {
		poisonIncidentID := createDurableStreamIncident(t, harness, admin, "poison")
		poisonIncidentUUID := uuid.MustParse(poisonIncidentID)
		quarantineNow := clockNow.Add(2 * time.Second)
		invalidIntent := intenttest.PersistedIntentFixture{
			IntentKey:        "record_changed:invalid-payload",
			IncidentID:       poisonIncidentUUID,
			EventFamily:      privatestream.EventFamilyRecordChanged,
			CanonicalPayload: json.RawMessage(`{"not":"a record change"}`),
			SourceIdentity:   "record:invalid-payload",
			CreatedAt:        quarantineNow,
		}
		intenttest.InsertPersistedIntentFixture(t, pool, invalidIntent)

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
		if _, err := harness.DB.ExecContext(ctx, `
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
		if _, err := harness.DB.ExecContext(ctx, `
CREATE TRIGGER cartulary_test_fail_requeue_intent_update
BEFORE UPDATE ON collaboration_event_intents
FOR EACH ROW EXECUTE FUNCTION cartulary_test_fail_requeue_intent_update()
`); err != nil {
			t.Fatalf("install requeue rollback trigger: %v", err)
		}
		defer func() {
			_, _ = harness.DB.ExecContext(context.Background(), `DROP TRIGGER IF EXISTS cartulary_test_fail_requeue_intent_update ON collaboration_event_intents`)
			_, _ = harness.DB.ExecContext(context.Background(), `DROP FUNCTION IF EXISTS cartulary_test_fail_requeue_intent_update()`)
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
		if _, err := harness.DB.ExecContext(ctx, `
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
		if _, err := harness.DB.ExecContext(ctx, `
CREATE TRIGGER cartulary_test_fail_requeue_journal_insert
BEFORE INSERT ON deployment_admin_audit_events
FOR EACH ROW EXECUTE FUNCTION cartulary_test_fail_requeue_journal_insert()
`); err != nil {
			t.Fatalf("install requeue journal rollback trigger: %v", err)
		}
		defer func() {
			_, _ = harness.DB.ExecContext(context.Background(), `DROP TRIGGER IF EXISTS cartulary_test_fail_requeue_journal_insert ON deployment_admin_audit_events`)
			_, _ = harness.DB.ExecContext(context.Background(), `DROP FUNCTION IF EXISTS cartulary_test_fail_requeue_journal_insert()`)
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

	t.Run("persisted corrupt job and extension payloads fail semantic validation before sequencing", func(t *testing.T) {
		cases := []struct {
			name    string
			family  string
			payload json.RawMessage
		}{
			{name: "job", family: privatestream.EventFamilyJobProgress, payload: json.RawMessage(`{"job_id":"persisted-corruption"}`)},
			{name: "extension", family: privatestream.EventFamilyExtensionResourceChange, payload: json.RawMessage(`{"extension_profile_id":"network_flow_activity","resource_kind":"network_flow_table","resource_id":"nft_invalid","change_kind":"remove","reason_code":"renamed"}`)},
		}
		for index, testCase := range cases {
			t.Run(testCase.name, func(t *testing.T) {
				incidentID := createDurableStreamIncident(t, harness, admin, "invalid-"+testCase.name)
				incidentUUID := uuid.MustParse(incidentID)
				attemptedAt := clockNow.Add(time.Duration(20+index) * time.Second)
				intent := intenttest.PersistedIntentFixture{
					IntentKey:        "persisted-corruption:" + testCase.name,
					IncidentID:       incidentUUID,
					EventFamily:      testCase.family,
					CanonicalPayload: testCase.payload,
					SourceIdentity:   "persisted-corruption:" + testCase.name,
					CreatedAt:        attemptedAt,
				}
				intenttest.InsertPersistedIntentFixture(t, pool, intent)
				dispatcher := newDispatcherForTest(pool, &recordingBroadcaster{}, func() time.Time { return attemptedAt.Add(time.Second) })
				if _, err := dispatcher.RunOnce(ctx); err != nil {
					t.Fatalf("attempt persisted corrupt sequencing: %v", err)
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
					t.Fatalf("load persisted corrupt sequencing result: %v", err)
				}
				if lastError == nil || *lastError != "invalid_event_payload" || replayCount != 0 {
					t.Fatalf("persisted corrupt sequencing = error %v replay_count %d", lastError, replayCount)
				}
				if _, err := pool.Exec(ctx, `
DELETE FROM collaboration_event_intents
 WHERE intent_key = $1
`, intent.IntentKey); err != nil {
					t.Fatalf("remove persisted corrupt fixture: %v", err)
				}
			})
		}
	})
}
