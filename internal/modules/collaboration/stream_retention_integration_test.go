package collaboration_test

import (
	"bytes"
	"context"
	"crypto/sha256"
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

func runReplayRetentionScenarios(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	harness *appsupport.ServerHarness,
	admin flowtest.LoginResult,
	adminID uuid.UUID,
	dispatchIncidentUUID uuid.UUID,
	replay *privatestream.PostgresStream,
	intents privatestream.IntentWriter,
) {
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

		restartedReplay := newPostgresStreamForTest(t, pool)
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
		var (
			corruptEventID  uuid.UUID
			originalPayload []byte
		)
		if err := pool.QueryRow(ctx, `
SELECT event_id, canonical_payload
  FROM collaboration_replay_events
 WHERE incident_id = $1
   AND stream_seq = 1
`, dispatchIncidentUUID).Scan(&corruptEventID, &originalPayload); err != nil {
			t.Fatalf("load authenticated replay corruption fixture: %v", err)
		}
		intenttest.ReplacePersistedReplayPayload(t, pool, corruptEventID, []byte(`{"job_id":"corrupt-replay"}`))
		corruptReplay, err := restartedReplay.ReplayMessages(
			ctx,
			sessionID,
			dispatchIncidentUUID,
			"durable-restart-client",
			token,
			0,
			tokenNow,
		)
		if err == nil || len(corruptReplay.Messages) != 0 {
			t.Fatalf("corrupt authenticated replay result = %#v error=%v", corruptReplay, err)
		}
		intenttest.ReplacePersistedReplayPayload(t, pool, corruptEventID, originalPayload)
		repairedReplay, err := restartedReplay.ReplayMessages(
			ctx,
			sessionID,
			dispatchIncidentUUID,
			"durable-restart-client",
			token,
			0,
			tokenNow,
		)
		if err != nil || len(repairedReplay.Messages) != 2 ||
			repairedReplay.Messages[0].StreamSeq == nil || *repairedReplay.Messages[0].StreamSeq != 1 ||
			repairedReplay.Messages[1].StreamSeq == nil || *repairedReplay.Messages[1].StreamSeq != 2 {
			t.Fatalf("repaired authenticated replay result = %#v error=%v", repairedReplay, err)
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
		retentionPayload := collabtestprotocol.RawPayload(collabtestprotocol.NewIncidentJobProgressPayload(
			"retention",
			retentionIncidentUUID,
			collabprotocol.JobStatusSucceeded,
			collabprotocol.JobProgress{},
			retentionNow,
		))
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
       $3::jsonb,
       CASE
           WHEN sequence = 1 THEN $2::timestamptz - interval '24 hours 1 microsecond'
           WHEN sequence = 2 THEN $2::timestamptz - interval '24 hours'
           ELSE $2::timestamptz
       END
  FROM generate_series(1, 10002) AS sequence
`, retentionIncidentUUID, retentionNow, retentionPayload); err != nil {
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
